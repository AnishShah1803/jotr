package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/state"
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
		DiaryPath:   filepath.Join(fs.BaseDir, "diary"),
		TodoPath:    todoPath,
		StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
		TaskSection: "Tasks",
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
		DiaryPath:   filepath.Join(fs.BaseDir, "diary"),
		TodoPath:    todoPath,
		StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
		TaskSection: "Tasks",
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
		DiaryPath:   filepath.Join(fs.BaseDir, "diary"),
		TodoPath:    todoPath,
		StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
		TaskSection: "Tasks",
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
		DiaryPath:   filepath.Join(fs.BaseDir, "diary"),
		TodoPath:    todoPath,
		StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
		TaskSection: "Tasks",
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
		DiaryPath:   filepath.Join(fs.BaseDir, "diary"),
		TodoPath:    todoPath,
		StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
		TaskSection: "Tasks",
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

	initialState := state.NewTodoState()
	initialState.Tasks = map[string]state.TaskState{
		"task1": {ID: "task1", Text: "Task one", Completed: false, Source: "daily.md"},
		"task2": {ID: "task2", Text: "Task two", Completed: false, Source: "daily.md"},
	}
	if err := initialState.Write(statePath); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	todoContent := `# To-Do List

## Tasks

- [ ] Task one
- [ ] Task two
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
				DiaryPath:   filepath.Join(fs.BaseDir, "diary"),
				TodoPath:    todoPath,
				StatePath:   statePath,
				TaskSection: "Tasks",
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

	lockFile, err := utils.TryLockFile(statePath)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}
	if lockFile == nil {
		t.Fatal("TryLockFile returned nil file")
	}
	defer utils.UnlockFile(lockFile)

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

- [ ] Task one <!-- id: task1 -->
`
	fs.WriteFile(t, filepath.Join("diary", year, monthDir, dayFile), dailyNoteContent)

	statePath := filepath.Join(fs.BaseDir, ".todo_state.json")
	todoPath := filepath.Join(fs.BaseDir, "todo.md")

	initialState := state.NewTodoState()
	initialState.Tasks = map[string]state.TaskState{
		"task1": {ID: "task1", Text: "Task one", Completed: false, Source: "daily.md"},
	}
	if err := initialState.Write(statePath); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	todoContent := `# To-Do List

## Tasks

- [ ] Task one
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
				DiaryPath:   filepath.Join(fs.BaseDir, "diary"),
				TodoPath:    todoPath,
				StatePath:   statePath,
				TaskSection: "Tasks",
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
		t.Fatalf("Failed to read final state: %v", err)
	}

	if _, exists := finalState.Tasks["task1"]; !exists {
		t.Error("Task1 should still exist after concurrent syncs")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
