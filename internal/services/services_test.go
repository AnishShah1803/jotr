package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/testhelpers"
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

	if result.TasksSynced != 0 {
		t.Errorf("SyncTasks().TasksSynced = %d; want 0", result.TasksSynced)
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
	fs.AssertFileExists(t, filepath.Join("Archive", "archive-2026-01.md"))
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
	if result.TasksSynced != 1 {
		t.Errorf("SyncTasks().TasksSynced = %d; want 1 (substring false positive prevention)", result.TasksSynced)
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
	if result.TasksSynced != 0 {
		t.Errorf("SyncTasks().TasksSynced = %d; want 0 (exact match should be deduplicated)", result.TasksSynced)
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
	if result.TasksSynced != 0 {
		t.Errorf("SyncTasks().TasksSynced = %d; want 0 (ID-based match should be deduplicated)", result.TasksSynced)
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
	if result.TasksSynced != 2 {
		t.Errorf("SyncTasks().TasksSynced = %d; want 2 (similar but different tasks should sync)", result.TasksSynced)
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
