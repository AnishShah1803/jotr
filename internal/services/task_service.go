package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/state"
	"github.com/AnishShah1803/jotr/internal/tasks"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// TaskService provides task management operations.
type TaskService struct{}

// NewTaskService creates a new TaskService instance.
func NewTaskService() *TaskService {
	return &TaskService{}
}

// SyncOptions contains options for syncing tasks.
type SyncOptions struct {
	DiaryPath   string
	TodoPath    string
	StatePath   string
	TaskSection string
	LockTimeout time.Duration // Timeout for acquiring file locks (default: 10s)
}

// SyncResult contains the result of a sync operation.
type SyncResult struct {
	StatePath      string
	DailyPath      string
	TodoPath       string
	TasksRead      int
	TasksFromDaily int
	TasksFromTodo  int
	Conflicts      map[string]string
}

// acquireSyncLocks acquires locks on state, todo, and daily note files in the correct order.
// Returns a slice of file handles that must be released in reverse order.
// Lock order: state file → todo file → daily note
func (s *TaskService) acquireSyncLocks(statePath, todoPath, notePath string, timeout time.Duration) ([]*os.File, error) {
	var locks []*os.File

	// Acquire lock order: state file → todo file → daily note
	// This ordering must be consistent to prevent deadlocks

	// Lock state file
	if statePath != "" {
		lockFile, err := utils.LockFile(statePath, timeout)
		if err != nil {
			// Release any already acquired locks
			for _, l := range locks {
				utils.UnlockFile(l)
			}
			return nil, fmt.Errorf("failed to acquire lock on state file: %w", err)
		}
		locks = append(locks, lockFile)
	}

	// Lock todo file
	if todoPath != "" {
		lockFile, err := utils.LockFile(todoPath, timeout)
		if err != nil {
			// Release any already acquired locks
			for _, l := range locks {
				utils.UnlockFile(l)
			}
			return nil, fmt.Errorf("failed to acquire lock on todo file: %w", err)
		}
		locks = append(locks, lockFile)
	}

	// Lock daily note (if provided)
	if notePath != "" {
		lockFile, err := utils.LockFile(notePath, timeout)
		if err != nil {
			// Release any already acquired locks
			for _, l := range locks {
				utils.UnlockFile(l)
			}
			return nil, fmt.Errorf("failed to acquire lock on daily note: %w", err)
		}
		locks = append(locks, lockFile)
	}

	return locks, nil
}

// isLockTimeoutError checks if an error is a lock timeout error.
// This is used to provide user-friendly error messages for sync operations.
func (s *TaskService) isLockTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, utils.ErrLockTimeout)
}

// SyncTasks performs bidirectional sync between daily notes and todo list.
func (s *TaskService) SyncTasks(ctx context.Context, opts SyncOptions) (*SyncResult, error) {
	result := &SyncResult{
		StatePath: opts.StatePath,
		TodoPath:  opts.TodoPath,
	}

	today := time.Now()
	notePath := notes.BuildDailyNotePath(opts.DiaryPath, today)
	result.DailyPath = notePath

	if !utils.FileExists(notePath) {
		return nil, fmt.Errorf("today's note doesn't exist: %s", notePath)
	}

	// Acquire locks on all three files before reading any data
	// Lock order: state file → todo file → daily note
	lockTimeout := opts.LockTimeout
	if lockTimeout <= 0 {
		lockTimeout = 10 * time.Second
	}
	locks, err := s.acquireSyncLocks(opts.StatePath, opts.TodoPath, notePath, lockTimeout)
	if err != nil {
		if s.isLockTimeoutError(err) {
			return nil, fmt.Errorf("another sync operation is in progress. Please try again in a few seconds")
		}
		return nil, err
	}
	// Release locks in reverse order when function returns
	defer func() {
		if locks == nil {
			return
		}
		for i := len(locks) - 1; i >= 0; i-- {
			utils.UnlockFile(locks[i])
		}
	}()

	// Read all data AFTER acquiring locks to prevent race conditions
	dailyTasks, err := tasks.ReadTasks(ctx, notePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read daily note: %w", err)
	}

	for i := range dailyTasks {
		tasks.EnsureTaskID(&dailyTasks[i])
	}

	result.TasksRead = len(dailyTasks)

	taskSection := opts.TaskSection
	if taskSection == "" {
		taskSection = "Tasks"
	}

	var activeDailyTasks []tasks.Task
	for _, task := range dailyTasks {
		if task.Section == taskSection && !task.Completed {
			activeDailyTasks = append(activeDailyTasks, task)
		}
	}

	todoState, err := state.Read(opts.StatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	if todoState.NeedsMigration() && utils.FileExists(opts.TodoPath) {
		existingTasks, _ := tasks.ReadTasks(ctx, opts.TodoPath)
		if len(existingTasks) > 0 {
			todoState.MigrateFromMarkdown(existingTasks, "migration")
		}
	}

	var todoTasks []tasks.Task
	if utils.FileExists(opts.TodoPath) {
		todoTasks, _ = tasks.ReadTasks(ctx, opts.TodoPath)
	}

	syncResult := todoState.BidirectionalSync(activeDailyTasks, todoTasks, notePath)

	result.Conflicts = syncResult.Conflicts
	if len(syncResult.Conflicts) > 0 {
		return result, nil
	}

	if syncResult.StateUpdated {
		if opts.StatePath != "" {
			if err := todoState.Write(opts.StatePath); err != nil {
				return nil, fmt.Errorf("failed to write state file: %w", err)
			}
		}
	}

	if syncResult.TodoChanged {
		if err := s.writeTodoFileFromState(opts.TodoPath, todoState, true); err != nil {
			return nil, fmt.Errorf("failed to write todo file: %w", err)
		}
	}

	if syncResult.DailyChanged {
		if err := s.updateDailyNoteFromState(notePath, dailyTasks, todoState, opts.TaskSection); err != nil {
			return nil, fmt.Errorf("failed to update daily note: %w", err)
		}
	}

	result.TasksFromDaily = syncResult.AppliedDaily
	result.TasksFromTodo = syncResult.AppliedTodo

	return result, nil
}

func (s *TaskService) updateDailyNoteFromState(notePath string, dailyTasks []tasks.Task, todoState *state.TodoState, taskSection string) error {
	if taskSection == "" {
		taskSection = "Tasks"
	}

	noteContent, err := os.ReadFile(notePath)
	if err != nil {
		return fmt.Errorf("failed to read daily note: %w", err)
	}

	lines := strings.Split(string(noteContent), "\n")
	var updatedLines []string
	var inTaskSection bool
	var sectionFound bool

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") {
			if sectionFound && inTaskSection {
				inTaskSection = false
				sectionName := strings.TrimPrefix(trimmed, "## ")
				if sectionName == taskSection {
					sectionFound = false
				}
			} else if !sectionFound {
				sectionName := strings.TrimPrefix(trimmed, "## ")
				if sectionName == taskSection {
					sectionFound = true
					inTaskSection = true
					updatedLines = append(updatedLines, line)
					updatedLines = append(updatedLines, "")

					for _, task := range dailyTasks {
						if task.Section == taskSection {
							if stateTask, exists := todoState.Tasks[task.ID]; exists {
								taskLine := s.formatTaskLine(stateTask)
								updatedLines = append(updatedLines, taskLine)
							}
						}
					}
					continue
				}
			}
		}

		if !inTaskSection {
			updatedLines = append(updatedLines, line)
		}
	}

	if !sectionFound {
		return nil
	}

	content := strings.Join(updatedLines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := utils.AtomicWriteFile(notePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write daily note: %w", err)
	}

	return nil
}

func (s *TaskService) formatTaskLine(stateTask state.TaskState) string {
	var sb strings.Builder
	if stateTask.Completed {
		sb.WriteString("- [x] ")
	} else {
		sb.WriteString("- [ ] ")
	}

	sb.WriteString(stateTask.Text)

	if stateTask.ID != "" {
		sb.WriteString(fmt.Sprintf(" <!-- id: %s -->", stateTask.ID))
	}

	return sb.String()
}

// writeTodoFileFromState generates and writes the todo markdown file from state.
func (s *TaskService) writeTodoFileFromState(todoPath string, todoState *state.TodoState, includeCompleted bool) error {
	var content strings.Builder
	content.WriteString("# To-Do List\n\n")

	var tasksToWrite []state.TaskState
	if includeCompleted {
		// Get all tasks including completed ones
		allTasks := todoState.ToTasks()
		for _, task := range allTasks {
			tasksToWrite = append(tasksToWrite, state.TaskState{
				Text:      task.Text,
				Section:   task.Section,
				Priority:  task.Priority,
				Tags:      task.Tags,
				ID:        task.ID,
				Completed: task.Completed,
			})
		}
	} else {
		tasksToWrite = todoState.GetActiveTasks()
	}

	sections := make(map[string][]state.TaskState)
	for _, task := range tasksToWrite {
		var section string
		// If task is completed and has a CompletedDate, use that as the section
		if task.Completed && task.CompletedDate != "" {
			section = task.CompletedDate
		} else {
			section = task.Section
			if section == "" {
				section = "Tasks"
			}
		}
		sections[section] = append(sections[section], task)
	}

	var sectionNames []string
	for name := range sections {
		sectionNames = append(sectionNames, name)
	}

	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	for i := 0; i < len(sectionNames)-1; i++ {
		for j := i + 1; j < len(sectionNames); j++ {
			dateI := dateRegex.MatchString(sectionNames[i])
			dateJ := dateRegex.MatchString(sectionNames[j])

			if dateI && dateJ {
				if sectionNames[i] < sectionNames[j] {
					sectionNames[i], sectionNames[j] = sectionNames[j], sectionNames[i]
				}
			} else if !dateI && dateJ {
				sectionNames[i], sectionNames[j] = sectionNames[j], sectionNames[i]
			}
		}
	}

	for _, sectionName := range sectionNames {
		content.WriteString(fmt.Sprintf("## %s\n\n", sectionName))
		for _, task := range sections[sectionName] {
			checkbox := "[ ]"
			if task.Completed {
				checkbox = "[x]"
			}
			content.WriteString(fmt.Sprintf("- %s %s\n", checkbox, task.Text))
		}
		content.WriteString("\n")
	}

	if err := utils.AtomicWriteFile(todoPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write todo file: %w", err)
	}

	return nil
}

// ArchiveOptions contains options for archiving tasks.
type ArchiveOptions struct {
	TodoPath    string
	StatePath   string
	BaseDir     string
	LockTimeout time.Duration // Timeout for acquiring file locks (default: 10s)
}

// ArchiveResult contains the result of an archive operation.
type ArchiveResult struct {
	ArchivePath    string
	ArchivedCount  int
	RemainingCount int
}

// ArchiveTasks moves completed tasks to an archive file using state as source of truth.
func (s *TaskService) ArchiveTasks(ctx context.Context, opts ArchiveOptions) (*ArchiveResult, error) {
	result := &ArchiveResult{}

	lockTimeout := opts.LockTimeout
	if lockTimeout <= 0 {
		lockTimeout = 10 * time.Second
	}
	locks, err := s.acquireSyncLocks(opts.StatePath, opts.TodoPath, "", lockTimeout)
	if err != nil {
		if s.isLockTimeoutError(err) {
			return nil, fmt.Errorf("another archive operation is in progress. Please try again in a few seconds")
		}
		return nil, err
	}
	defer func() {
		if locks == nil {
			return
		}
		for i := len(locks) - 1; i >= 0; i-- {
			utils.UnlockFile(locks[i])
		}
	}()

	todoState, err := state.Read(opts.StatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	if todoState.NeedsMigration() && utils.FileExists(opts.TodoPath) {
		existingTasks, _ := tasks.ReadTasks(ctx, opts.TodoPath)
		if len(existingTasks) > 0 {
			todoState.MigrateFromMarkdown(existingTasks, "migration")
		}
	}

	completedTasks := todoState.GetCompletedTasks()
	activeTasks := todoState.GetActiveTasks()

	if len(completedTasks) == 0 {
		result.ArchivedCount = 0
		result.RemainingCount = len(activeTasks)
		return result, nil
	}

	now := time.Now()

	archiveDir := filepath.Join(opts.BaseDir, "Archive")
	if err := notes.EnsureDir(archiveDir); err != nil {
		return nil, fmt.Errorf("failed to create archive directory: %w", err)
	}

	archiveFile := filepath.Join(archiveDir, fmt.Sprintf("archive-%s.md", now.Format("2006-01")))

	var archiveContent string
	if utils.FileExists(archiveFile) {
		content, err := os.ReadFile(archiveFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read archive: %w", err)
		}
		archiveContent = string(content)
	} else {
		archiveContent = fmt.Sprintf("# Archive - %s\n\n", now.Format("January 2006"))
	}

	archiveContent += fmt.Sprintf("\n## Archived on %s\n\n", now.Format("2006-01-02"))
	for _, task := range completedTasks {
		archiveContent += fmt.Sprintf("- [x] %s\n", task.Text)
	}

	if err := utils.AtomicWriteFile(archiveFile, []byte(archiveContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write archive: %w", err)
	}

	if err := s.writeTodoFileFromState(opts.TodoPath, todoState, false); err != nil {
		return nil, fmt.Errorf("failed to write todo file: %w", err)
	}

	todoState.MarkArchived()
	if opts.StatePath != "" {
		if err := todoState.Write(opts.StatePath); err != nil {
			return nil, fmt.Errorf("failed to write state file: %w", err)
		}
	}

	result.ArchivedCount = len(completedTasks)
	result.RemainingCount = len(activeTasks)
	result.ArchivePath = archiveFile

	return result, nil
}

// GetAllTasks reads all tasks from a file.
func (s *TaskService) GetAllTasks(ctx context.Context, todoPath string) ([]tasks.Task, error) {
	return tasks.ReadTasks(ctx, todoPath)
}

// GetTaskSummary returns a summary of tasks grouped by priority.
func (s *TaskService) GetTaskSummary(ctx context.Context, todoPath string) (*tasks.Task, error) {
	allTasks, err := tasks.ReadTasks(ctx, todoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks: %w", err)
	}

	completed := false
	pendingTasks := tasks.FilterTasks(allTasks, &completed, "", "")

	_ = tasks.GroupByPriority(pendingTasks)

	if len(pendingTasks) > 0 {
		return &pendingTasks[0], nil
	}

	return nil, nil
}

// GetTaskStats returns statistics about tasks.
func (s *TaskService) GetTaskStats(ctx context.Context, todoPath string) (*TaskStats, error) {
	allTasks, err := tasks.ReadTasks(ctx, todoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks: %w", err)
	}

	stats := &TaskStats{
		Total:      len(allTasks),
		ByPriority: tasks.GroupByPriority(allTasks),
		BySection:  tasks.GroupBySection(allTasks),
	}

	_, completed, pending := tasks.CountTasks(allTasks)
	stats.Completed = completed
	stats.Pending = pending

	if stats.Total > 0 {
		stats.CompletionRate = float64(stats.Completed) / float64(stats.Total) * 100
	}

	// Count overdue tasks
	for _, task := range allTasks {
		if tasks.IsOverdue(task) {
			stats.Overdue++
		}
	}

	return stats, nil
}

// TaskStats contains task statistics.
type TaskStats struct {
	ByPriority     map[string][]tasks.Task
	BySection      map[string][]tasks.Task
	Total          int
	Completed      int
	Pending        int
	Overdue        int
	CompletionRate float64
}
