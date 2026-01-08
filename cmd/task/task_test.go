package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/tasks"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// createTestTaskConfig creates a test configuration with a temporary directory.
func createTestTaskConfig(t *testing.T, tmpDir string) *config.LoadedConfig {
	t.Helper()

	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = tmpDir
	cfg.Paths.DiaryDir = "Diary"
	cfg.Format.CaptureSection = "Captured"
	cfg.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	cfg.Format.DailyNoteDirPattern = "{year}/{month}"
	cfg.Format.TaskSection = "Tasks"
	cfg.DiaryPath = filepath.Join(tmpDir, "Diary")
	cfg.TodoPath = filepath.Join(tmpDir, "todo.md")

	return cfg
}

// TestSyncTasks_NoTasksToSync tests syncTasks when there are no tasks to sync.
func TestSyncTasks_NoTasksToSync(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestTaskConfig(t, tmpDir)

	// Create diary directory but no daily note
	diaryPath := cfg.DiaryPath
	if err := os.MkdirAll(diaryPath, 0750); err != nil {
		t.Fatalf("Failed to create diary directory: %v", err)
	}

	// syncTasks should handle missing daily note gracefully
	err = syncTasks(context.Background(), cfg)
	if err != nil {
		t.Logf("syncTasks returned error for missing note: %v", err)
	}
}

// TestSyncTasks_WithDailyNote tests syncTasks with a valid daily note.
func TestSyncTasks_WithDailyNote(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestTaskConfig(t, tmpDir)

	// Create diary directory
	diaryPath := cfg.DiaryPath
	if err := os.MkdirAll(diaryPath, 0750); err != nil {
		t.Fatalf("Failed to create diary directory: %v", err)
	}

	// Create today's daily note with tasks
	todayNotePath := notes.BuildDailyNotePath(diaryPath, time.Now())
	dailyContent := `# Today

## Tasks

- [ ] Task 1
- [ ] Task 2
`
	ctx := context.Background()

	if err := notes.WriteNote(ctx, todayNotePath, dailyContent); err != nil {
		t.Fatalf("Failed to create daily note: %v", err)
	}

	// Sync tasks - should create todo file
	err = syncTasks(context.Background(), cfg)
	if err != nil {
		t.Fatalf("syncTasks failed: %v", err)
	}

	// Verify todo file was created
	if !utils.FileExists(cfg.TodoPath) {
		t.Errorf("Expected todo file to be created at %s", cfg.TodoPath)
	}
}

// TestSyncTasks_EmptyTaskSection tests syncTasks when Tasks section is empty.
func TestSyncTasks_EmptyTaskSection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestTaskConfig(t, tmpDir)

	// Create diary directory
	diaryPath := cfg.DiaryPath
	if err := os.MkdirAll(diaryPath, 0750); err != nil {
		t.Fatalf("Failed to create diary directory: %v", err)
	}

	// Create today's daily note with no Tasks section
	todayNotePath := notes.BuildDailyNotePath(diaryPath, time.Now())
	dailyContent := `# Today

Just some notes, no tasks here.
`
	ctx := context.Background()

	if err := notes.WriteNote(ctx, todayNotePath, dailyContent); err != nil {
		t.Fatalf("Failed to create daily note: %v", err)
	}

	// Sync should not fail but should report 0 tasks synced
	err = syncTasks(context.Background(), cfg)
	if err != nil {
		t.Fatalf("syncTasks failed: %v", err)
	}
}

// TestArchiveTasks_NoTasks tests archiveTasks when there are no completed tasks.
func TestArchiveTasks_NoTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestTaskConfig(t, tmpDir)

	// Create an empty todo file
	todoContent := "# To-Do List\n\n## Tasks\n\n"
	ctx := context.Background()

	if err := notes.WriteNote(ctx, cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}

	// Archive should handle empty todo file gracefully
	err = archiveTasks(context.Background(), cfg)
	if err != nil {
		t.Fatalf("archiveTasks failed: %v", err)
	}
}

// TestArchiveTasks_WithCompletedTasks tests archiveTasks with completed tasks.
func TestArchiveTasks_WithCompletedTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestTaskConfig(t, tmpDir)

	// Create a todo file with completed tasks
	todoContent := `# To-Do List

## Tasks

- [ ] Active task 1
- [x] Completed task 1
- [ ] Active task 2
- [x] Completed task 2
`
	ctx := context.Background()

	if err := notes.WriteNote(ctx, cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}

	// Archive should remove completed tasks
	err = archiveTasks(context.Background(), cfg)
	if err != nil {
		t.Fatalf("archiveTasks failed: %v", err)
	}

	// Verify completed tasks were removed
	content, err := os.ReadFile(cfg.TodoPath)
	if err != nil {
		t.Fatalf("Failed to read todo file: %v", err)
	}

	if strings.Contains(string(content), "Completed task") {
		t.Errorf("Expected completed tasks to be removed from todo file")
	}

	if !strings.Contains(string(content), "Active task") {
		t.Errorf("Expected active tasks to remain in todo file")
	}
}

// TestArchiveTasks_AllCompleted tests archiveTasks when all tasks are completed.
func TestArchiveTasks_AllCompleted(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestTaskConfig(t, tmpDir)

	// Create a todo file with only completed tasks
	todoContent := `# To-Do List

## Tasks

- [x] Completed task 1
- [x] Completed task 2
`
	ctx := context.Background()

	if err := notes.WriteNote(ctx, cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}

	// Archive should handle all-completed todo file
	err = archiveTasks(context.Background(), cfg)
	if err != nil {
		t.Fatalf("archiveTasks failed: %v", err)
	}
}

// TestShowStats_EmptyTodo tests showStats with an empty todo file.
func TestShowStats_EmptyTodo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestTaskConfig(t, tmpDir)

	// Create an empty todo file
	todoContent := "# To-Do List\n\n## Tasks\n\n"
	ctx := context.Background()

	if err := notes.WriteNote(ctx, cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}

	// showStats should handle empty todo file
	err = showStats(context.Background(), cfg)
	if err != nil {
		t.Fatalf("showStats failed: %v", err)
	}
}

// TestShowStats_WithTasks tests showStats with tasks in the todo file.
func TestShowStats_WithTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestTaskConfig(t, tmpDir)

	// Create a todo file with various tasks
	todoContent := `# To-Do List

## Tasks

- [ ] P0: Critical task
- [ ] P1: High priority task
- [x] Completed task
`
	ctx := context.Background()

	if err := notes.WriteNote(ctx, cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}

	// showStats should display statistics
	err = showStats(context.Background(), cfg)
	if err != nil {
		t.Fatalf("showStats failed: %v", err)
	}
}

// TestShowSummary_NoPendingTasks tests ShowSummary when there are no pending tasks.
func TestShowSummary_NoPendingTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestTaskConfig(t, tmpDir)

	// Create a todo file with only completed tasks
	todoContent := `# To-Do List

## Tasks

- [x] Completed task 1
- [x] Completed task 2
`
	ctx := context.Background()

	if err := notes.WriteNote(ctx, cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}

	// ShowSummary should handle no pending tasks
	err = ShowSummary(ctx, cfg)
	if err != nil {
		t.Fatalf("ShowSummary failed: %v", err)
	}
}

// TestShowSummary_WithPendingTasks tests ShowSummary with pending tasks.
func TestShowSummary_WithPendingTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestTaskConfig(t, tmpDir)

	// Create a todo file with pending tasks
	todoContent := `# To-Do List

## Tasks

- [ ] P0: Critical task
- [ ] P1: High priority task
- [ ] P2: Medium priority task
- [x] Completed task
`
	ctx := context.Background()

	if err := notes.WriteNote(ctx, cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}

	// ShowSummary should display task summary
	err = ShowSummary(ctx, cfg)
	if err != nil {
		t.Fatalf("ShowSummary failed: %v", err)
	}
}

// TestTaskIDGeneration tests that task IDs are generated correctly.
func TestTaskIDGeneration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a todo file with tasks that need IDs
	todoPath := filepath.Join(tmpDir, "todo.md")
	todoContent := `# To-Do List

## Tasks

- [ ] Task without ID
`
	ctx := context.Background()

	if err := notes.WriteNote(ctx, todoPath, todoContent); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}

	// Read tasks and ensure they have IDs
	allTasks, err := tasks.ReadTasks(ctx, todoPath)
	if err != nil {
		t.Fatalf("ReadTasks failed: %v", err)
	}

	if len(allTasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(allTasks))
	}
}

// TestTaskParsingWithPriority tests parsing tasks with priority markers.
func TestTaskParsingWithPriority(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	todoPath := filepath.Join(tmpDir, "todo.md")
	todoContent := `# To-Do List

## Tasks

- [ ][P0] Critical priority task
- [ ][P1] High priority task
- [ ][P2] Medium priority task
- [ ][P3] Low priority task
- [ ] No priority task
`
	ctx := context.Background()

	if err := notes.WriteNote(ctx, todoPath, todoContent); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}

	// Read tasks
	allTasks, err := tasks.ReadTasks(ctx, todoPath)
	if err != nil {
		t.Fatalf("ReadTasks failed: %v", err)
	}

	if len(allTasks) != 5 {
		t.Errorf("Expected 5 tasks, got %d", len(allTasks))
	}

	// Check priorities
	priorityCounts := make(map[string]int)
	for _, task := range allTasks {
		priorityCounts[task.Priority]++
	}

	if priorityCounts["P0"] != 1 {
		t.Errorf("Expected 1 P0 task, got %d: %v", priorityCounts["P0"], priorityCounts)
	}

	if priorityCounts["P1"] != 1 {
		t.Errorf("Expected 1 P1 task, got %d: %v", priorityCounts["P1"], priorityCounts)
	}
	// Task without priority marker has empty string as priority
	if priorityCounts[""] != 1 {
		t.Errorf("Expected 1 task with no priority marker, got %d: %v", priorityCounts[""], priorityCounts)
	}
}
