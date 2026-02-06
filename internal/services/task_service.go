package services

import (
	"context"
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
}

// SyncResult contains the result of a sync operation.
type SyncResult struct {
	TodoPath    string
	TasksRead   int
	TasksSynced int
}

// SyncTasks reads tasks from the daily note and syncs new ones to the todo file using state as source of truth.
func (s *TaskService) SyncTasks(ctx context.Context, opts SyncOptions) (*SyncResult, error) {
	result := &SyncResult{}

	today := time.Now()
	notePath := notes.BuildDailyNotePath(opts.DiaryPath, today)

	if !utils.FileExists(notePath) {
		return nil, fmt.Errorf("today's note doesn't exist: %s", notePath)
	}

	dailyTasks, err := tasks.ReadTasks(ctx, notePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read daily note: %w", err)
	}

	result.TasksRead = len(dailyTasks)

	taskSection := opts.TaskSection
	if taskSection == "" {
		taskSection = "Tasks"
	}

	var tasksToSync []tasks.Task
	for _, task := range dailyTasks {
		if task.Section == taskSection && !task.Completed {
			tasks.EnsureTaskID(&task)
			tasksToSync = append(tasksToSync, task)
		}
	}

	if len(tasksToSync) == 0 {
		result.TasksSynced = 0
		result.TodoPath = opts.TodoPath
		return result, nil
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

	existingTasks := make(map[string]bool)
	if utils.FileExists(opts.TodoPath) {
		existingMarkdownTasks, _ := tasks.ReadTasks(ctx, opts.TodoPath)
		for _, t := range existingMarkdownTasks {
			existingTasks[t.ID] = true
			existingTasks[t.Text] = true
		}
	}

	sectionName := today.Format("2006-01-02")
	syncedCount := 0

	for _, task := range tasksToSync {
		if !todoState.HasTask(task.ID) && !existingTasks[task.ID] && !existingTasks[task.Text] {
			task.Section = sectionName
			todoState.AddTask(task, notePath)
			syncedCount++
		}
	}

	if opts.StatePath != "" {
		if err := todoState.Write(opts.StatePath); err != nil {
			return nil, fmt.Errorf("failed to write state file: %w", err)
		}
	}

	if err := s.writeTodoFileFromState(opts.TodoPath, todoState); err != nil {
		return nil, fmt.Errorf("failed to write todo file: %w", err)
	}

	result.TasksSynced = syncedCount
	result.TodoPath = opts.TodoPath

	return result, nil
}

// writeTodoFileFromState generates and writes the todo markdown file from state.
func (s *TaskService) writeTodoFileFromState(todoPath string, todoState *state.TodoState) error {
	var content strings.Builder
	content.WriteString("# To-Do List\n\n")

	activeTasks := todoState.GetActiveTasks()

	sections := make(map[string][]state.TaskState)
	for _, task := range activeTasks {
		section := task.Section
		if section == "" {
			section = "Tasks"
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
			content.WriteString(fmt.Sprintf("- [ ] %s\n", task.Text))
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
	TodoPath  string
	StatePath string
	BaseDir   string
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

	lockFile, err := utils.LockFile(opts.TodoPath, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer utils.UnlockFile(lockFile)

	if err := s.writeTodoFileFromState(opts.TodoPath, todoState); err != nil {
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

// findOptimalSectionInsertionPoint finds the best position for a new date section.
// It places newer dates before older dates to maintain chronological order.
func findOptimalSectionInsertionPoint(lines []string, sectionName string) int {
	dateRegex := regexp.MustCompile(`^## (\d{4}-\d{2}-\d{2})`)

	newDate, _ := time.Parse("2006-01-02", sectionName)

	for i, line := range lines {
		matches := dateRegex.FindStringSubmatch(line)
		if len(matches) > 0 {
			existingDate, _ := time.Parse("2006-01-02", matches[1])
			if newDate.After(existingDate) || newDate.Equal(existingDate) {
				return i
			}
		}
	}

	return -1
}

func findAfterTitleInsertionPoint(lines []string) int {
	for i, line := range lines {
		if strings.TrimSpace(line) == "" && i > 0 && strings.HasPrefix(lines[i-1], "# ") {
			return i + 1
		}
	}

	return 1
}

func insertAtPosition(lines []string, pos int, elements ...string) []string {
	result := make([]string, 0, len(lines)+len(elements))
	result = append(result, lines[:pos]...)
	result = append(result, elements...)
	result = append(result, lines[pos:]...)
	return result
}
