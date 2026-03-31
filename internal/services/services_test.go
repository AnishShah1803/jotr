package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/state"
	"github.com/AnishShah1803/jotr/internal/tasks"
	"github.com/AnishShah1803/jotr/internal/testhelpers"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// TaskService Tests

func TestTaskService_GetAllTasks(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	// Create a todo file with tasks
	todoContent := `# To-Do List

## Tasks

- [ ] Task one
- [x] Task two
- [ ] Task three
`
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	fs.WriteFile(t, "todo.md", todoContent)

	service := NewTaskService()
	ctx := context.Background()

	tasks, err := service.GetAllTasks(ctx, todoPath)
	if err != nil {
		t.Fatalf("GetAllTasks() error = %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("GetAllTasks() returned %d tasks; want 3", len(tasks))
	}
}

func TestTaskService_GetTaskStats(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	// Create a todo file with tasks
	todoContent := `# To-Do List

## Tasks

- [ ] Task one
- [x] Task two
- [ ] Task three
- [x] Task four
`
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	fs.WriteFile(t, "todo.md", todoContent)

	service := NewTaskService()
	ctx := context.Background()

	stats, err := service.GetTaskStats(ctx, todoPath)
	if err != nil {
		t.Fatalf("GetTaskStats() error = %v", err)
	}

	if stats.Total != 4 {
		t.Errorf("GetTaskStats().Total = %d; want 4", stats.Total)
	}

	if stats.Completed != 2 {
		t.Errorf("GetTaskStats().Completed = %d; want 2", stats.Completed)
	}

	if stats.Pending != 2 {
		t.Errorf("GetTaskStats().Pending = %d; want 2", stats.Pending)
	}
}

func TestTaskService_SyncTasks_NoTasksToSync(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	// Create a daily note without tasks
	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")
	fs.WriteFile(t, filepath.Join("diary", year, monthDir, dayFile), "# Daily Note\n\nNo tasks here")

	// Create empty todo file
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	fs.WriteFile(t, "todo.md", "# To-Do List\n\n## Tasks\n")

	service := NewTaskService()
	ctx := context.Background()

	result, err := service.SyncTasks(ctx, SyncOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: filepath.Join(fs.BaseDir, ".todo_state.json"),
	})
	if err != nil {
		t.Fatalf("SyncTasks() error = %v", err)
	}

	if result.TasksFromDaily != 0 && result.TasksFromTodo != 0 {
		t.Errorf("SyncTasks() synced %d from daily and %d from todo; want 0 changes", result.TasksFromDaily, result.TasksFromTodo)
	}
}

func TestTaskService_ArchiveTasks_NoCompleted(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	// Create todo file with no completed tasks
	todoContent := `# To-Do List

## Tasks

- [ ] Active task one
- [ ] Active task two
`
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	fs.WriteFile(t, "todo.md", todoContent)

	service := NewTaskService()
	ctx := context.Background()

	result, err := service.ArchiveTasks(ctx, ArchiveOptions{
		TodoPath: todoPath,
		BaseDir:  fs.BaseDir,
	})
	if err != nil {
		t.Fatalf("ArchiveTasks() error = %v", err)
	}

	if result.ArchivedCount != 0 {
		t.Errorf("ArchiveTasks().ArchivedCount = %d; want 0", result.ArchivedCount)
	}

	if result.RemainingCount != 2 {
		t.Errorf("ArchiveTasks().RemainingCount = %d; want 2", result.RemainingCount)
	}
}

func TestTaskService_ArchiveTasks_WithCompleted(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	// Create todo file with completed tasks
	todoContent := `# To-Do List

## Tasks

- [ ] Active task
- [x] Completed task one
- [x] Completed task two
`
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	fs.WriteFile(t, "todo.md", todoContent)

	service := NewTaskService()
	ctx := context.Background()

	result, err := service.ArchiveTasks(ctx, ArchiveOptions{
		TodoPath: todoPath,
		BaseDir:  fs.BaseDir,
	})
	if err != nil {
		t.Fatalf("ArchiveTasks() error = %v", err)
	}

	if result.ArchivedCount != 2 {
		t.Errorf("ArchiveTasks().ArchivedCount = %d; want 2", result.ArchivedCount)
	}

	if result.RemainingCount != 1 {
		t.Errorf("ArchiveTasks().RemainingCount = %d; want 1", result.RemainingCount)
	}

	// Verify archived file was created
	expectedArchive := fmt.Sprintf("archive-%s.md", time.Now().Format("2006-01"))
	fs.AssertFileExists(t, filepath.Join("Archive", expectedArchive))
}

func TestTaskService_GetTaskSummary(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	// Create a todo file with tasks
	todoContent := `# To-Do List

## Tasks

- [ ] Task one
- [x] Task two
`
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	fs.WriteFile(t, "todo.md", todoContent)

	service := NewTaskService()
	ctx := context.Background()

	summary, err := service.GetTaskSummary(ctx, todoPath)
	if err != nil {
		t.Fatalf("GetTaskSummary() error = %v", err)
	}

	if summary == nil {
		t.Error("GetTaskSummary() returned nil; expected a task")
	}
}

func TestTaskService_SyncTasks_Deduplication_SubstringFalsePositive(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	// Create a daily note with task "Review proposal"
	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")
	dailyNoteContent := `# Daily Note

## Tasks

- [ ] Review proposal
`
	fs.WriteFile(t, filepath.Join("diary", year, monthDir, dayFile), dailyNoteContent)

	// Create todo file with a task that CONTAINS "Review proposal" as substring
	// This should NOT cause a false positive - "Review proposal document" is different from "Review proposal"
	todoContent := `# To-Do List

## Tasks

- [ ] Review proposal document
- [ ] Some other task
`
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	fs.WriteFile(t, "todo.md", todoContent)

	service := NewTaskService()
	ctx := context.Background()

	result, err := service.SyncTasks(ctx, SyncOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: filepath.Join(fs.BaseDir, ".todo_state.json"),
	})
	if err != nil {
		t.Fatalf("SyncTasks() error = %v", err)
	}

	// Should sync 1 task because "Review proposal" is not the same as "Review proposal document"
	if result.TasksFromDaily != 1 {
		t.Errorf("SyncTasks().TasksFromDaily = %d; want 1 (substring false positive prevention)", result.TasksFromDaily)
	}
}

func TestTaskService_SyncTasks_Deduplication_ExactMatch(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	// Create a daily note with task "Review proposal"
	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")
	dailyNoteContent := `# Daily Note

## Tasks

- [ ] Review proposal
`
	fs.WriteFile(t, filepath.Join("diary", year, monthDir, dayFile), dailyNoteContent)

	// Create todo file with EXACT same task
	todoContent := `# To-Do List

## Tasks

- [ ] Review proposal
- [ ] Some other task
`
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	fs.WriteFile(t, "todo.md", todoContent)

	service := NewTaskService()
	ctx := context.Background()

	result, err := service.SyncTasks(ctx, SyncOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: filepath.Join(fs.BaseDir, ".todo_state.json"),
	})
	if err != nil {
		t.Fatalf("SyncTasks() error = %v", err)
	}

	// Should NOT sync because exact match already exists
	if result.TasksFromDaily != 0 && result.TasksFromTodo != 0 {
		t.Errorf("SyncTasks() synced %d from daily and %d from todo; want 0 changes (exact match should be deduplicated)", result.TasksFromDaily, result.TasksFromTodo)
	}
}

func TestTaskService_SyncTasks_Deduplication_IDBased(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	// Create a daily note with task that has an ID
	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")
	dailyNoteContent := `# Daily Note

## Tasks

- [ ] Review proposal <!-- id: abc12345 -->
`
	fs.WriteFile(t, filepath.Join("diary", year, monthDir, dayFile), dailyNoteContent)

	// Create todo file with same task but different text formatting (same ID)
	todoContent := `# To-Do List

## Tasks

- [ ] Review proposal <!-- id: abc12345 -->
- [ ] Some other task
`
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	fs.WriteFile(t, "todo.md", todoContent)

	service := NewTaskService()
	ctx := context.Background()

	result, err := service.SyncTasks(ctx, SyncOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: filepath.Join(fs.BaseDir, ".todo_state.json"),
	})
	if err != nil {
		t.Fatalf("SyncTasks() error = %v", err)
	}

	// Should NOT sync because task with same ID already exists
	if result.TasksFromDaily != 0 && result.TasksFromTodo != 0 {
		t.Errorf("SyncTasks() synced %d from daily and %d from todo; want 0 changes (ID-based match should be deduplicated)", result.TasksFromDaily, result.TasksFromTodo)
	}
}

func TestTaskService_SyncTasks_Deduplication_MultipleSimilar(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	// Create a daily note with multiple similar-sounding but different tasks
	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")
	dailyNoteContent := `# Daily Note

## Tasks

- [ ] Update
- [ ] Update documentation
- [ ] Update config
`
	fs.WriteFile(t, filepath.Join("diary", year, monthDir, dayFile), dailyNoteContent)

	// Create todo file with only one of them
	todoContent := `# To-Do List

## Tasks

- [ ] Update
`
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	fs.WriteFile(t, "todo.md", todoContent)

	service := NewTaskService()
	ctx := context.Background()

	result, err := service.SyncTasks(ctx, SyncOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: filepath.Join(fs.BaseDir, ".todo_state.json"),
	})
	if err != nil {
		t.Fatalf("SyncTasks() error = %v", err)
	}

	// Should sync 2 tasks: "Update documentation" and "Update config"
	// Neither is a substring match of "Update" (exact) and "Update" is already in todo
	if result.TasksFromDaily != 2 {
		t.Errorf("SyncTasks().TasksFromDaily = %d; want 2 (similar but different tasks should sync)", result.TasksFromDaily)
	}
}

func TestTaskService_SyncTasks_PromotesManualTodoOnlyTaskToDailyAndSurvivesPrune(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")
	dailyRelPath := filepath.Join("diary", year, monthDir, dayFile)
	dailyPath := filepath.Join(fs.BaseDir, dailyRelPath)

	const manualTaskText = "Manual todo-only task should sync into daily"

	fs.WriteFile(t, dailyRelPath, "# Daily Note\n\n## Tasks\n\n")
	fs.WriteFile(t, "todo.md", "# To-Do List\n\n## Tasks\n\n- [ ] "+manualTaskText+"\n")

	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	statePath := filepath.Join(fs.BaseDir, ".todo_state.json")

	service := NewTaskService()
	ctx := context.Background()

	if _, err := service.SyncTasks(ctx, SyncOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: statePath,
	}); err != nil {
		t.Fatalf("SyncTasks() error = %v", err)
	}

	dailyContent, err := os.ReadFile(dailyPath)
	if err != nil {
		t.Fatalf("failed to read daily note after sync: %v", err)
	}
	if !strings.Contains(string(dailyContent), "- [ ] "+manualTaskText) {
		t.Fatalf("expected manual todo-only task in daily note after sync, got:\n%s", dailyContent)
	}

	todoTasks, err := tasks.ReadTasks(ctx, todoPath)
	if err != nil {
		t.Fatalf("failed to read todo tasks after sync: %v", err)
	}

	foundInTodo := false
	for _, task := range todoTasks {
		if strings.TrimSpace(tasks.StripTaskID(task.Text)) == manualTaskText {
			foundInTodo = true
			break
		}
	}
	if !foundInTodo {
		t.Fatalf("expected manual task to remain in todo/state after sync")
	}

	todoState, err := state.Read(statePath)
	if err != nil {
		t.Fatalf("failed to read state after sync: %v", err)
	}

	var syncedTaskID string
	for id, taskState := range todoState.Tasks {
		if strings.TrimSpace(tasks.StripTaskID(taskState.Text)) == manualTaskText {
			syncedTaskID = id
			break
		}
	}
	if syncedTaskID == "" {
		t.Fatalf("expected manual task to be present in todo state after sync")
	}

	pruneResult, err := service.PruneTasks(ctx, PruneOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: statePath,
	})
	if err != nil {
		t.Fatalf("PruneTasks() error = %v", err)
	}
	if pruneResult.RemovedCount != 0 {
		t.Fatalf("PruneTasks().RemovedCount = %d; want 0 for synced manual todo-only task", pruneResult.RemovedCount)
	}
	if pruneResult.OrphanCount != 0 {
		t.Fatalf("PruneTasks().OrphanCount = %d; want 0 for synced manual todo-only task", pruneResult.OrphanCount)
	}

	prunedState, err := state.Read(statePath)
	if err != nil {
		t.Fatalf("failed to read state after prune: %v", err)
	}
	if _, exists := prunedState.Tasks[syncedTaskID]; !exists {
		t.Fatalf("expected synced manual task %q to remain in state after prune", syncedTaskID)
	}
}

func TestTaskService_LoadConfig(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	config, err := config.Load()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config == nil {
		t.Error("LoadConfig() returned nil")
	}
}

func TestTaskService_SyncTasks_ConcurrentFileLocking(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")

	dailyNoteContent := `# Daily Note

## Tasks

- [ ] Task one <!-- id: task1 -->
- [ ] Task two <!-- id: task2 -->
`
	fs.WriteFile(t, filepath.Join("diary", year, monthDir, dayFile), dailyNoteContent)

	statePath := filepath.Join(fs.BaseDir, ".todo_state.json")
	todoPath := filepath.Join(fs.BaseDir, "todo.md")

	dailyNotePath := filepath.Join(fs.BaseDir, "diary", year, monthDir, dayFile)

	initialState := state.NewTodoState()
	initialState.Tasks = map[string]state.TaskState{
		"task1": {ID: "task1", Text: "Task one", Completed: false, Source: dailyNotePath},
		"task2": {ID: "task2", Text: "Task two", Completed: false, Source: dailyNotePath},
	}
	if err := initialState.Write(statePath); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	todoContent := `# To-Do List

## Tasks

- [ ] Task one <!-- id: task1 -->
- [ ] Task two <!-- id: task2 -->
`
	fs.WriteFile(t, "todo.md", todoContent)

	service := NewTaskService()
	ctx := context.Background()

	numGoroutines := 5
	errors := make(chan error, numGoroutines)
	done := make(chan bool, numGoroutines)
	start := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer func() { done <- true }()

			<-start

			_, err := service.SyncTasks(ctx, SyncOptions{
				DiaryPath: filepath.Join(fs.BaseDir, "diary"),
				TodoPath:  todoPath,
				StatePath: statePath,
			})
			errors <- err
		}(i)
	}

	close(start)

	timeout := time.After(30 * time.Second)
	completed := 0

	for completed < numGoroutines {
		select {
		case <-done:
			completed++
		case err := <-errors:
			if err != nil {
				t.Errorf("Error in goroutine: %v", err)
			}
		case <-timeout:
			t.Fatalf("Test timed out after 30 seconds - possible deadlock (%d/%d completed)", completed, numGoroutines)
		}
	}

	finalState, err := state.Read(statePath)
	if err != nil {
		t.Fatalf("Failed to read final state: %v", err)
	}

	if len(finalState.Tasks) == 0 {
		t.Error("Final state has no tasks - data loss detected")
	}

	for id, task := range finalState.Tasks {
		if task.Text == "" {
			t.Errorf("Task %s has empty text (possible corruption)", id)
		}
		if task.ID == "" {
			t.Errorf("Task has empty ID (possible corruption): %v", task)
		}
	}
}

func TestTaskService_SyncTasks_LockTimeoutError(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	statePath := filepath.Join(fs.BaseDir, ".todo_state.json")

	lm := utils.NewDefaultFileLockManager()
	lockHandle, err := lm.TryLockFile(statePath)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}
	if lockHandle == nil {
		t.Fatal("TryLockFile returned nil handle")
	}
	defer lm.UnlockFile(lockHandle)

	service := NewTaskService()

	timeoutErr := fmt.Errorf("%w: %s", utils.ErrLockTimeout, statePath)
	if !service.isLockTimeoutError(timeoutErr) {
		t.Error("isLockTimeoutError should return true for ErrLockTimeout wrapped error")
	}

	otherErr := fmt.Errorf("some other error")
	if service.isLockTimeoutError(otherErr) {
		t.Error("isLockTimeoutError should return false for non-timeout error")
	}

	if service.isLockTimeoutError(nil) {
		t.Error("isLockTimeoutError should return false for nil error")
	}
}

func TestTaskService_SyncTasks_UserFriendlyErrorMessage(t *testing.T) {
	service := NewTaskService()

	timeoutErr := fmt.Errorf("%w: some/path", utils.ErrLockTimeout)
	if !service.isLockTimeoutError(timeoutErr) {
		t.Error("isLockTimeoutError should return true for ErrLockTimeout wrapped error")
	}

	wrappedErr := fmt.Errorf("failed to acquire lock on state file: %w", timeoutErr)
	if !service.isLockTimeoutError(wrappedErr) {
		t.Error("isLockTimeoutError should detect nested ErrLockTimeout")
	}
}

func TestTaskService_WriteTodoFileFromState_CompletedDateSection(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	// Create a state with a task that:
	// - was created under section "2026-02-01"
	// - is completed with CompletedDate="2026-02-06"
	// - should appear under "## 2026-02-06" section in output, not "## 2026-02-01"
	todoState := state.NewTodoState()
	todoState.Tasks["task123"] = state.TaskState{
		ID:            "task123",
		Text:          "Review project proposal",
		Section:       "2026-02-01",
		Completed:     true,
		CompletedDate: "2026-02-06",
		CreatedDate:   "2026-02-01",
		Source:        "diary/2026-02-01.md",
	}

	// Also add an incomplete task under the same section to verify it stays there
	todoState.Tasks["task456"] = state.TaskState{
		ID:          "task456",
		Text:        "Ongoing task",
		Section:     "2026-02-01",
		Completed:   false,
		CreatedDate: "2026-02-01",
		Source:      "diary/2026-02-01.md",
	}

	todoPath := filepath.Join(fs.BaseDir, "todo.md")

	service := NewTaskService()
	if err := service.writeTodoFileFromState(todoPath, todoState, true); err != nil {
		t.Fatalf("writeTodoFileFromState() error = %v", err)
	}

	// Read the generated file
	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read generated todo file: %v", err)
	}

	contentStr := string(content)

	// Verify the completed task appears under "## 2026-02-06" section
	// Find the section boundaries
	idx0206 := strings.Index(contentStr, "## 2026-02-06")
	idx0201 := strings.Index(contentStr, "## 2026-02-01")

	if idx0206 == -1 {
		t.Error("Expected '## 2026-02-06' section header for completed task, not found")
	}
	if idx0201 == -1 {
		t.Error("Expected '## 2026-02-01' section header for incomplete task, not found")
	}

	// Find task positions
	idxCompleted := strings.Index(contentStr, "Review project proposal")
	idxIncomplete := strings.Index(contentStr, "Ongoing task")

	if idxCompleted == -1 {
		t.Error("Completed task 'Review project proposal' not found in output")
	}
	if idxIncomplete == -1 {
		t.Error("Incomplete task 'Ongoing task' not found in output")
	}

	// Verify completed task is in the 2026-02-06 section (after its header but before next section)
	if idx0206 != -1 && idxCompleted != -1 {
		if idxCompleted < idx0206 {
			t.Error("Completed task should appear after '## 2026-02-06' header")
		}
		// If there's a section after 2026-02-06, task should be before it
		if idx0201 > idx0206 && idxCompleted > idx0201 {
			t.Error("Completed task should be in '## 2026-02-06' section, not after '## 2026-02-01'")
		}
	}

	// Verify incomplete task is in the 2026-02-01 section
	if idx0201 != -1 && idxIncomplete != -1 {
		if idxIncomplete < idx0201 {
			t.Error("Incomplete task should appear after '## 2026-02-01' header")
		}
	}

	// Verify completed task has [x] checkbox
	if !strings.Contains(contentStr, "- [x] Review project proposal") {
		t.Error("Completed task should have [x] checkbox")
	}

	// Verify incomplete task has [ ] checkbox
	if !strings.Contains(contentStr, "- [ ] Ongoing task") {
		t.Error("Incomplete task should have [ ] checkbox")
	}
}

func TestTaskService_WriteTodoFileFromState_DailyTaskUsesCreatedDateSection(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	todoState := state.NewTodoState()
	todoState.Tasks["task123"] = state.TaskState{
		ID:          "task123",
		Text:        "Draft release notes",
		Section:     "Tasks",
		Completed:   false,
		CreatedDate: "2026-02-01",
		Source:      "diary/2026-02-01.md",
	}

	todoPath := filepath.Join(fs.BaseDir, "todo.md")

	service := NewTaskService()
	if err := service.writeTodoFileFromState(todoPath, todoState, false); err != nil {
		t.Fatalf("writeTodoFileFromState() error = %v", err)
	}

	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read generated todo file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "## 2026-02-01") {
		t.Fatalf("expected daily task to be written under its created-date section, got:\n%s", contentStr)
	}
	if strings.Contains(contentStr, "## Tasks") {
		t.Fatalf("expected task not to be written under generic Tasks section, got:\n%s", contentStr)
	}
}

func TestTaskService_WriteTodoFileFromState_ActiveTaskMovesToCreatedDateSection(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	todoState := state.NewTodoState()
	todoState.Tasks["task123"] = state.TaskState{
		ID:          "task123",
		Text:        "Move me to my date section",
		Section:     "Tasks",
		Completed:   false,
		CreatedDate: "2026-02-07",
		Source:      "diary/2026-02-07.md",
	}

	todoPath := filepath.Join(fs.BaseDir, "todo.md")

	service := NewTaskService()
	if err := service.writeTodoFileFromState(todoPath, todoState, false); err != nil {
		t.Fatalf("writeTodoFileFromState() error = %v", err)
	}

	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read generated todo file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "## 2026-02-07") {
		t.Fatalf("expected active task to be written under created-date section, got:\n%s", contentStr)
	}
	if strings.Contains(contentStr, "## Tasks") {
		t.Fatalf("expected active task not to remain under generic Tasks section, got:\n%s", contentStr)
	}
}

func TestTaskService_WriteTodoFileFromState_ActiveTaskRehomesFromMismatchedSection(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	todoState := state.NewTodoState()
	todoState.Tasks["task123"] = state.TaskState{
		ID:          "task123",
		Text:        "Rehome me",
		Section:     "2026-01-01",
		Completed:   false,
		CreatedDate: "2026-02-07",
		Source:      "diary/2026-02-07.md",
	}

	todoPath := filepath.Join(fs.BaseDir, "todo.md")

	service := NewTaskService()
	if err := service.writeTodoFileFromState(todoPath, todoState, false); err != nil {
		t.Fatalf("writeTodoFileFromState() error = %v", err)
	}

	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read generated todo file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "## 2026-02-07") {
		t.Fatalf("expected task to be written under created-date section, got:\n%s", contentStr)
	}
	if strings.Contains(contentStr, "## 2026-01-01") {
		t.Fatalf("expected task to be moved out of mismatched section, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "- [ ] Rehome me") {
		t.Fatalf("expected task checkbox in output, got:\n%s", contentStr)
	}
}

func TestTaskService_FormatTaskLine_CompletedWithDate(t *testing.T) {
	service := NewTaskService()

	tests := []struct {
		name     string
		task     state.TaskState
		expected string
	}{
		{
			name: "completed task with CompletedDate includes @completed tag",
			task: state.TaskState{
				ID:            "abc12345",
				Text:          "Review project proposal",
				Completed:     true,
				CompletedDate: "2026-02-06",
			},
			expected: "- [x] Review project proposal <!-- id: abc12345 --> @completed(2026-02-06)",
		},
		{
			name: "completed task without CompletedDate has no tag",
			task: state.TaskState{
				ID:        "def45678",
				Text:      "Another task",
				Completed: true,
			},
			expected: "- [x] Another task <!-- id: def45678 -->",
		},
		{
			name: "incomplete task never has @completed tag",
			task: state.TaskState{
				ID:            "ghi90123",
				Text:          "Pending task",
				Completed:     false,
				CompletedDate: "2026-02-06", // Should be ignored
			},
			expected: "- [ ] Pending task <!-- id: ghi90123 -->",
		},
		{
			name: "task without ID omits ID comment",
			task: state.TaskState{
				Text:          "No ID task",
				Completed:     true,
				CompletedDate: "2026-02-10",
			},
			expected: "- [x] No ID task @completed(2026-02-10)",
		},
		{
			name: "completed task with ID but no CompletedDate has no tag",
			task: state.TaskState{
				ID:        "jkl45678",
				Text:      "Just completed",
				Completed: true,
			},
			expected: "- [x] Just completed <!-- id: jkl45678 -->",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.formatTaskLine(tt.task)
			if result != tt.expected {
				t.Errorf("formatTaskLine() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTaskService_WriteTodoFileFromState_RoundTripIDPreservation(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	// Create a state with multiple tasks that have IDs
	todoState := state.NewTodoState()
	todoState.Tasks["abc12345"] = state.TaskState{
		ID:          "abc12345",
		Text:        "Review project proposal",
		Section:     "2026-02-01",
		Completed:   false,
		CreatedDate: "2026-02-01",
		Source:      "diary/2026-02-01.md",
	}
	todoState.Tasks["def67890"] = state.TaskState{
		ID:            "def67890",
		Text:          "Complete documentation",
		Section:       "2026-02-01",
		Completed:     true,
		CompletedDate: "2026-02-06",
		CreatedDate:   "2026-02-01",
		Source:        "diary/2026-02-01.md",
	}
	todoState.Tasks["cab24680"] = state.TaskState{
		ID:        "cab24680",
		Text:      "Write unit tests",
		Section:   "Tasks",
		Completed: false,
		Source:    "todo.md",
	}

	todoPath := filepath.Join(fs.BaseDir, "todo.md")

	service := NewTaskService()

	// Write the todo file from state
	if err := service.writeTodoFileFromState(todoPath, todoState, true); err != nil {
		t.Fatalf("writeTodoFileFromState() error = %v", err)
	}

	// Read the file back using tasks.ParseTasks
	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read generated todo file: %v", err)
	}

	parsedTasks := tasks.ParseTasks(string(content))

	// Verify all 3 tasks were parsed
	if len(parsedTasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(parsedTasks))
	}

	// Create a map for easy lookup by ID
	parsedByIDs := make(map[string]tasks.Task)
	for _, task := range parsedTasks {
		if task.ID != "" {
			parsedByIDs[task.ID] = task
		}
	}

	// Verify each task's ID was preserved
	testCases := []struct {
		expectedID       string
		expectedText     string
		expectedComplete bool
	}{
		{"abc12345", "Review project proposal", false},
		{"def67890", "Complete documentation", true},
		{"cab24680", "Write unit tests", false},
	}

	for _, tc := range testCases {
		task, exists := parsedByIDs[tc.expectedID]
		if !exists {
			t.Errorf("Task with ID %s not found in parsed output", tc.expectedID)
			continue
		}

		if task.Text != tc.expectedText {
			t.Errorf("Task %s: expected text %q, got %q", tc.expectedID, tc.expectedText, task.Text)
		}

		if task.Completed != tc.expectedComplete {
			t.Errorf("Task %s: expected Completed=%v, got %v", tc.expectedID, tc.expectedComplete, task.Completed)
		}
	}

	// Verify the file content actually contains the ID comments
	contentStr := string(content)
	expectedIDPatterns := []string{
		"<!-- id: abc12345 -->",
		"<!-- id: def67890 -->",
		"<!-- id: cab24680 -->",
	}

	for _, pattern := range expectedIDPatterns {
		if !strings.Contains(contentStr, pattern) {
			t.Errorf("Expected file to contain %q, but it was not found", pattern)
		}
	}

	// Verify completed task has @completed tag
	if !strings.Contains(contentStr, "@completed(2026-02-06)") {
		t.Error("Expected completed task to have @completed(2026-02-06) tag")
	}
}

func TestTaskService_SyncTasks_DeadlockPrevention(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")

	dailyNoteContent := `# Daily Note

## Tasks

- [ ] Task one <!-- id: abc12345 -->
`
	fs.WriteFile(t, filepath.Join("diary", year, monthDir, dayFile), dailyNoteContent)

	statePath := filepath.Join(fs.BaseDir, ".todo_state.json")
	todoPath := filepath.Join(fs.BaseDir, "todo.md")

	initialState := state.NewTodoState()
	initialState.Tasks = map[string]state.TaskState{
		"abc12345": {ID: "abc12345", Text: "Task one", Completed: false, Source: "daily.md"},
	}
	if err := initialState.Write(statePath); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	todoContent := `# To-Do List

## Tasks

- [ ] Task one <!-- id: abc12345 -->
`
	fs.WriteFile(t, "todo.md", todoContent)

	service := NewTaskService()
	ctx := context.Background()

	numOps := 10
	var wg sync.WaitGroup
	errChan := make(chan error, numOps)

	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.SyncTasks(ctx, SyncOptions{
				DiaryPath: filepath.Join(fs.BaseDir, "diary"),
				TodoPath:  todoPath,
				StatePath: statePath,
			})
			errChan <- err
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out - deadlock detected")
	}

	close(errChan)
	for err := range errChan {
		if err != nil && !errors.Is(err, utils.ErrLockTimeout) {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	finalState, err := state.Read(statePath)
	if err != nil {
		t.Fatalf("Failed to read state: %v", err)
	}

	if _, exists := finalState.Tasks["abc12345"]; !exists {
		t.Error("Task abc12345 should still exist after concurrent syncs")
	}
}

// TestTaskService_SyncTasks_WriteBackIDAndPropagate verifies the full lifecycle:
// 1. A task with no ID in the daily note gets an ID written back after the first sync.
// 2. On the second sync the completion status propagates from state to the daily note.
func TestTaskService_SyncTasks_WriteBackIDAndPropagate(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")
	dailyRelPath := filepath.Join("diary", year, monthDir, dayFile)

	// Daily note with a task that has NO id comment.
	fs.WriteFile(t, dailyRelPath, "# Daily Note\n\n- [ ] Review docs\n")

	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	statePath := filepath.Join(fs.BaseDir, ".todo_state.json")
	fs.WriteFile(t, "todo.md", "# To-Do List\n\n## Tasks\n")

	service := NewTaskService()
	ctx := context.Background()

	// --- First sync: task should get an ID and it should be written back to the daily note.
	_, err := service.SyncTasks(ctx, SyncOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: statePath,
	})
	if err != nil {
		t.Fatalf("first SyncTasks() error = %v", err)
	}

	// Read the daily note back and confirm an id comment was written.
	dailyNotePath := filepath.Join(fs.BaseDir, "diary", year, monthDir, dayFile)
	dailyContent, err := os.ReadFile(dailyNotePath)
	if err != nil {
		t.Fatalf("failed to read daily note after first sync: %v", err)
	}
	if !strings.Contains(string(dailyContent), "<!-- id:") {
		t.Fatalf("expected daily note to contain an id comment after first sync, got:\n%s", dailyContent)
	}

	// Extract the generated ID from state.
	todoState, err := state.Read(statePath)
	if err != nil {
		t.Fatalf("failed to read state after first sync: %v", err)
	}
	var taskID string
	for id := range todoState.Tasks {
		taskID = id
		break
	}
	if taskID == "" {
		t.Fatal("expected at least one task in state after first sync")
	}

	// Mark the task as completed, simulating the todo app writing a completed task.
	// Only update todo.md — do NOT touch the state file. The second sync should detect
	// the change via CompareWithTodoList (state=incomplete vs todo=[x]) and propagate
	// the completion back to the daily note.
	completedLine := fmt.Sprintf("- [x] Review docs <!-- id: %s --> @completed(%s)\n", taskID, now.Format("2006-01-02"))
	if err := os.WriteFile(todoPath, []byte("# To-Do List\n\n"+completedLine), 0644); err != nil {
		t.Fatalf("failed to write completed task to todo.md: %v", err)
	}

	// --- Second sync: completion should propagate back to the daily note.
	_, err = service.SyncTasks(ctx, SyncOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: statePath,
	})
	if err != nil {
		t.Fatalf("second SyncTasks() error = %v", err)
	}

	dailyContent, err = os.ReadFile(dailyNotePath)
	if err != nil {
		t.Fatalf("failed to read daily note after second sync: %v", err)
	}
	if !strings.Contains(string(dailyContent), "- [x]") {
		t.Errorf("expected daily note to contain '- [x]' after second sync, got:\n%s", dailyContent)
	}
}

func TestTaskService_UpdateTask_UsesTaskSourceNote(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	now := time.Now()
	year := now.AddDate(0, 0, -2).Format("2006")
	monthDir := now.AddDate(0, 0, -2).Format("01-Jan")
	dayFile := now.AddDate(0, 0, -2).Format("2006-01-02-Mon.md")
	dailyRelPath := filepath.Join("diary", year, monthDir, dayFile)
	dailyPath := filepath.Join(fs.BaseDir, dailyRelPath)

	fs.WriteFile(t, dailyRelPath, "# Daily Note\n\n## Tasks\n\n- [ ] Old note task <!-- id: abc12345 -->\n")

	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	statePath := filepath.Join(fs.BaseDir, ".todo_state.json")
	fs.WriteFile(t, "todo.md", "# To-Do List\n\n## Tasks\n\n- [ ] Old note task <!-- id: abc12345 -->\n")

	service := NewTaskService()
	ctx := context.Background()

	_, err := service.SyncTasks(ctx, SyncOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: statePath,
		DailyPath: dailyPath,
	})
	if err != nil {
		t.Fatalf("initial SyncTasks() error = %v", err)
	}

	_, err = service.UpdateTask(ctx, UpdateTaskOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: statePath,
		TaskID:    "abc12345",
		Text:      "Edited old note task",
		DailyPath: dailyPath,
	})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	dailyContent, err := os.ReadFile(dailyPath)
	if err != nil {
		t.Fatalf("failed to read daily note: %v", err)
	}
	if !strings.Contains(string(dailyContent), "Edited old note task") {
		t.Fatalf("expected old daily note to be updated, got:\n%s", dailyContent)
	}
	if strings.Contains(string(dailyContent), "Old note task <!-- id: abc12345 -->") {
		t.Fatalf("expected old task text to be replaced, got:\n%s", dailyContent)
	}
}

func TestTaskService_UpdateTask_PersistsPriorityAndTagsOnTodoFile(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")
	if err := os.MkdirAll(filepath.Join(fs.BaseDir, "diary", year, monthDir), 0o750); err != nil {
		t.Fatalf("failed to create diary dir: %v", err)
	}

	fs.WriteFile(t, filepath.Join("diary", year, monthDir, dayFile), "# Daily Note\n\n## Tasks\n")

	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	statePath := filepath.Join(fs.BaseDir, ".todo_state.json")

	fs.WriteFile(t, "todo.md", `# To-Do List

## Tasks

- [ ] Existing tagged task [P2] #alpha #beta <!-- id: abc12345 -->
- [ ] Initially tagless task <!-- id: def67890 -->
`)

	service := NewTaskService()
	ctx := context.Background()

	_, err := service.SyncTasks(ctx, SyncOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: statePath,
	})
	if err != nil {
		t.Fatalf("initial SyncTasks() error = %v", err)
	}

	_, err = service.UpdateTask(ctx, UpdateTaskOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: statePath,
		TaskID:    "abc12345",
		Text:      "Existing tagged task updated",
		Priority:  "P1",
		Tags:      []string{"beta", "alpha"},
	})
	if err != nil {
		t.Fatalf("UpdateTask() for existing tagged task error = %v", err)
	}

	_, err = service.UpdateTask(ctx, UpdateTaskOptions{
		DiaryPath: filepath.Join(fs.BaseDir, "diary"),
		TodoPath:  todoPath,
		StatePath: statePath,
		TaskID:    "def67890",
		Text:      "Initially tagless task updated",
		Priority:  "P3",
		Tags:      []string{"newtag"},
	})
	if err != nil {
		t.Fatalf("UpdateTask() for initially tagless task error = %v", err)
	}

	tasksOnTodo, err := tasks.ReadTasks(ctx, todoPath)
	if err != nil {
		t.Fatalf("failed to read todo tasks after updates: %v", err)
	}

	findTask := func(id string) tasks.Task {
		t.Helper()
		for _, task := range tasksOnTodo {
			if task.ID == id {
				return task
			}
		}
		t.Fatalf("task with id %q not found in todo file", id)
		return tasks.Task{}
	}

	existingTagged := findTask("abc12345")
	if existingTagged.Priority != "P1" {
		t.Fatalf("expected updated priority P1 for existing tagged task, got %q", existingTagged.Priority)
	}
	tags := append([]string(nil), existingTagged.Tags...)
	sort.Strings(tags)
	if strings.Join(tags, ",") != "alpha,beta" {
		t.Fatalf("expected unchanged tags alpha,beta for existing tagged task, got %v", existingTagged.Tags)
	}

	initiallyTagless := findTask("def67890")
	if initiallyTagless.Priority != "P3" {
		t.Fatalf("expected updated priority P3 for initially tagless task, got %q", initiallyTagless.Priority)
	}
	if got := strings.Join(initiallyTagless.Tags, ","); got != "newtag" {
		t.Fatalf("expected newly added tag newtag for initially tagless task, got %v", initiallyTagless.Tags)
	}
}

// TestTaskService_CreateTask_NoDeadlock verifies that CreateTask -> SyncTasks
// doesn't deadlock by reacquiring the same locks.
func TestTaskService_CreateTask_NoDeadlock(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")
	dailyRelPath := filepath.Join("diary", year, monthDir, dayFile)

	diaryPath := filepath.Join(fs.BaseDir, "diary")
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	statePath := filepath.Join(fs.BaseDir, ".todo_state.json")

	fs.WriteFile(t, dailyRelPath, "# Daily Note\n")
	fs.WriteFile(t, "todo.md", "# To-Do List\n\n## Tasks\n")

	service := NewTaskService()
	ctx := context.Background()

	done := make(chan bool, 1)
	var result *CreateTaskResult
	var err error

	go func() {
		result, err = service.CreateTask(ctx, CreateTaskOptions{
			DiaryPath:   diaryPath,
			TodoPath:    todoPath,
			StatePath:   statePath,
			Text:        "Test task",
			LockTimeout: 5 * time.Second,
		})
		done <- true
	}()

	select {
	case <-done:
		if err != nil {
			t.Fatalf("CreateTask() error = %v", err)
		}
		if result == nil || result.TaskID == "" {
			t.Fatal("expected CreateTask to return a task with an ID")
		}

		dailyNotePath := filepath.Join(fs.BaseDir, dailyRelPath)
		dailyContent, err := os.ReadFile(dailyNotePath)
		if err != nil {
			t.Fatalf("failed to read daily note: %v", err)
		}

		if !strings.Contains(string(dailyContent), result.TaskID) {
			t.Errorf("expected daily note to contain task ID %s, got:\n%s", result.TaskID, dailyContent)
		}

	case <-time.After(10 * time.Second):
		t.Fatal("CreateTask() deadlocked or timed out")
	}
}

// TestTaskService_UpdateTask_NoDeadlock verifies that UpdateTask -> SyncTasks
// doesn't deadlock by reacquiring the same locks.
func TestTaskService_UpdateTask_NoDeadlock(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	now := time.Now()
	year := now.Format("2006")
	monthDir := now.Format("01-Jan")
	dayFile := now.Format("2006-01-02-Mon.md")
	dailyRelPath := filepath.Join("diary", year, monthDir, dayFile)

	diaryPath := filepath.Join(fs.BaseDir, "diary")
	todoPath := filepath.Join(fs.BaseDir, "todo.md")
	statePath := filepath.Join(fs.BaseDir, ".todo_state.json")

	fs.WriteFile(t, dailyRelPath, "# Daily Note\n")
	fs.WriteFile(t, "todo.md", "# To-Do List\n\n## Tasks\n")

	service := NewTaskService()
	ctx := context.Background()

	createResult, err := service.CreateTask(ctx, CreateTaskOptions{
		DiaryPath:   diaryPath,
		TodoPath:    todoPath,
		StatePath:   statePath,
		Text:        "Original task",
		LockTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	done := make(chan bool, 1)
	var updateErr error

	go func() {
		_, updateErr = service.UpdateTask(ctx, UpdateTaskOptions{
			DiaryPath:   diaryPath,
			TodoPath:    todoPath,
			StatePath:   statePath,
			TaskID:      createResult.TaskID,
			Text:        "Updated task text",
			LockTimeout: 5 * time.Second,
		})
		done <- true
	}()

	select {
	case <-done:
		if updateErr != nil {
			t.Fatalf("UpdateTask() error = %v", updateErr)
		}

	case <-time.After(10 * time.Second):
		t.Fatal("UpdateTask() deadlocked or timed out")
	}
}
