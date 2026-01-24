package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AnishShah1803/jotr/internal/notes"
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
	TaskSection string
}

// SyncResult contains the result of a sync operation.
type SyncResult struct {
	TodoPath    string
	TasksRead   int
	TasksSynced int
}

// SyncTasks reads tasks from the daily note and syncs new ones to the todo file.
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

	// Filter tasks from the task section
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

	var todoContent string

	if utils.FileExists(opts.TodoPath) {
		content, err := os.ReadFile(opts.TodoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read todo file: %w", err)
		}

		todoContent = string(content)
	} else {
		todoContent = "# To-Do List\n\n## Tasks\n\n"
	}

	lines := strings.Split(todoContent, "\n")

	insertIndex := utils.FindSectionIndex(lines, "Tasks")
	if insertIndex == -1 {
		lines = append(lines, "", "## Tasks", "")
		insertIndex = len(lines) - 1
	}

	newLines := make([]string, 0, len(lines)+len(tasksToSync))
	newLines = append(newLines, lines[:insertIndex]...)

	syncedCount := 0

	for _, task := range tasksToSync {
		taskLine := fmt.Sprintf("- [ ] %s", task.Text)

		exists := false

		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			// Check for exact task line match first (most reliable for deduplication)
			// Strip ID from both sides to handle cases where EnsureTaskID added an ID
			baseTaskLine := tasks.StripTaskID(task.Text)
			expectedLine := fmt.Sprintf("- [ ] %s", strings.TrimSpace(baseTaskLine))
			if trimmedLine == expectedLine {
				exists = true
				break
			}
			// Fall back to ID-based matching if available (prevents false positives)
			if task.ID != "" && strings.Contains(trimmedLine, task.ID) {
				exists = true
				break
			}
		}

		if !exists {
			newLines = append(newLines, taskLine)
			syncedCount++
		}
	}

	newLines = append(newLines, lines[insertIndex:]...)

	newContent := strings.Join(newLines, "\n")
	if err := utils.AtomicWriteFile(opts.TodoPath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write todo file: %w", err)
	}

	result.TasksSynced = syncedCount
	result.TodoPath = opts.TodoPath

	return result, nil
}

// ArchiveOptions contains options for archiving tasks.
type ArchiveOptions struct {
	TodoPath string
	BaseDir  string
}

// ArchiveResult contains the result of an archive operation.
type ArchiveResult struct {
	ArchivePath    string
	ArchivedCount  int
	RemainingCount int
}

// ArchiveTasks moves completed tasks to an archive file.
func (s *TaskService) ArchiveTasks(ctx context.Context, opts ArchiveOptions) (*ArchiveResult, error) {
	result := &ArchiveResult{}

	allTasks, err := tasks.ReadTasks(ctx, opts.TodoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks: %w", err)
	}

	// Filter completed tasks
	var completedTasks []tasks.Task

	var activeTasks []tasks.Task

	for _, task := range allTasks {
		if task.Completed {
			completedTasks = append(completedTasks, task)
		} else {
			activeTasks = append(activeTasks, task)
		}
	}

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

	// Rebuild todo file with only active tasks - use file locking for safety
	lockFile, err := utils.LockFile(opts.TodoPath, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer utils.UnlockFile(lockFile)

	content, err := os.ReadFile(opts.TodoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read todo file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	var newLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Keep headers and non-task lines
		if !strings.HasPrefix(trimmed, "- [x]") && !strings.HasPrefix(trimmed, "- [X]") {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	if err := utils.AtomicWriteFile(opts.TodoPath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write todo file: %w", err)
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
