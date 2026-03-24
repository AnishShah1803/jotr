package cmd

import (
	"context"
	"io"
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

func TestRunTaskAdd_NormalizesNumericPriorityAndPersistsMarker(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestTaskConfig(t, tmpDir)
	cfg.StatePath = filepath.Join(tmpDir, "state.json")

	// Create diary directory and today's note (required for CreateTask)
	todayPath := notes.BuildDailyNotePath(cfg.DiaryPath, time.Now())
	if err := os.MkdirAll(filepath.Dir(todayPath), 0750); err != nil {
		t.Fatalf("failed to create diary directory: %v", err)
	}
	if err := notes.WriteNote(context.Background(), todayPath, "# Today\n\n## Tasks\n\n"); err != nil {
		t.Fatalf("failed to create today's note: %v", err)
	}

	output, err := RunTaskAdd(context.Background(), cfg, "Ship release", "1", []string{"ops"})
	if err != nil {
		t.Fatalf("RunTaskAdd returned error: %v", err)
	}
	if !strings.Contains(output, "Added task") {
		t.Fatalf("expected add confirmation output, got %q", output)
	}

	content, err := os.ReadFile(cfg.TodoPath)
	if err != nil {
		t.Fatalf("failed to read todo file: %v", err)
	}
	if !strings.Contains(string(content), "Ship release") || !strings.Contains(string(content), "[P1]") || !strings.Contains(string(content), "#ops") {
		t.Fatalf("expected persisted task line to contain task text, normalized priority marker, and tag, got:\n%s", string(content))
	}

	tasksOnTodo, err := tasks.ReadTasks(context.Background(), cfg.TodoPath)
	if err != nil {
		t.Fatalf("failed to read parsed tasks: %v", err)
	}
	if len(tasksOnTodo) != 1 {
		t.Fatalf("expected one task, got %d", len(tasksOnTodo))
	}
	if got := tasksOnTodo[0].Priority; got != "P1" {
		t.Fatalf("expected normalized parsed priority %q, got %q", "P1", got)
	}
}

func TestRunTaskList_DisplaysPrioritiesInOutput(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestTaskConfig(t, tmpDir)
	cfg.StatePath = filepath.Join(tmpDir, "state.json")

	todoContent := `# To-Do List

## Tasks

- [ ] Task with P0 [P0] #ops
- [ ] Task with P2 [P2]
- [ ] Task with P3 [P3]
- [ ] Task with no priority
`
	if err := notes.WriteNote(context.Background(), cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("failed to create todo file: %v", err)
	}

	output, err := withPatchedCLIIO(t, "", func() error {
		return RunTaskList(context.Background(), cfg, false)
	})
	if err != nil {
		t.Fatalf("RunTaskList returned error: %v", err)
	}

	if !strings.Contains(output, "[P0]") {
		t.Fatalf("expected P0 priority in output, got: %q", output)
	}
	if !strings.Contains(output, "[P2]") {
		t.Fatalf("expected P2 priority in output, got: %q", output)
	}
	if !strings.Contains(output, "[P3]") {
		t.Fatalf("expected P3 priority in output, got: %q", output)
	}
	if strings.Contains(output, "Task with no priority [P") {
		t.Fatalf("expected no priority marker for unprioritized task, got: %q", output)
	}
	if !strings.Contains(output, "#ops") {
		t.Fatalf("expected task tag metadata in output, got: %q", output)
	}
	if strings.Contains(output, "## Tasks") {
		t.Fatalf("expected task list output without markdown Tasks heading, got: %q", output)
	}
	if strings.Contains(output, "- [ ]") || strings.Contains(output, "- [x]") {
		t.Fatalf("expected task list output without markdown checkboxes, got: %q", output)
	}

	assertFragmentsInOrder(t, output, []string{
		"Task with P0",
		"Task with P2",
		"Task with P3",
		"Task with no priority",
	})
}

func assertFragmentsInOrder(t *testing.T, output string, fragments []string) {
	t.Helper()

	lastIdx := -1
	for _, fragment := range fragments {
		idx := strings.Index(output, fragment)
		if idx == -1 {
			t.Fatalf("expected output to contain fragment %q, got: %q", fragment, output)
		}
		if idx < lastIdx {
			t.Fatalf("expected fragment %q to appear after previous fragments, got: %q", fragment, output)
		}
		lastIdx = idx
	}
}

func withPatchedCLIIO(t *testing.T, input string, fn func() error) (string, error) {
	t.Helper()

	origStdin := os.Stdin
	origStdout := os.Stdout

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	os.Stdin = stdinR
	os.Stdout = stdoutW

	defer func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
		_ = stdinR.Close()
	}()

	if _, err := stdinW.WriteString(input); err != nil {
		t.Fatalf("failed to write stdin input: %v", err)
	}
	_ = stdinW.Close()

	runErr := fn()
	_ = stdoutW.Close()

	buf, readErr := io.ReadAll(stdoutR)
	_ = stdoutR.Close()
	if readErr != nil {
		t.Fatalf("failed to read captured stdout: %v", readErr)
	}

	return string(buf), runErr
}

func TestRunTaskComplete_ShowsListBeforePromptWhenSelectionMissing(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestTaskConfig(t, tmpDir)
	cfg.StatePath = filepath.Join(tmpDir, "state.json")

	todoContent := `# To-Do List

## Tasks

- [ ] First pending task
- [ ] Second pending task
`
	if err := notes.WriteNote(context.Background(), cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("failed to create todo file: %v", err)
	}

	output, err := withPatchedCLIIO(t, "1\n", func() error {
		return RunTaskComplete(context.Background(), cfg, nil)
	})
	if err != nil {
		t.Fatalf("RunTaskComplete returned error: %v", err)
	}

	listIdx := strings.Index(output, "First pending task")
	promptIdx := strings.Index(output, "Task numbers to complete:")
	if listIdx == -1 || promptIdx == -1 {
		t.Fatalf("expected both task list and prompt in output, got: %q", output)
	}
	if listIdx > promptIdx {
		t.Fatalf("expected task list to be shown before prompt; listIdx=%d promptIdx=%d output=%q", listIdx, promptIdx, output)
	}
}

func TestRunTaskComplete_IncludesCompletedTodoTaskTextInOutput(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestTaskConfig(t, tmpDir)
	cfg.StatePath = filepath.Join(tmpDir, "state.json")

	const taskText = "Todo-origin completion text"
	todoContent := `# To-Do List

## Tasks

- [ ] ` + taskText + `
`
	if err := notes.WriteNote(context.Background(), cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("failed to create todo file: %v", err)
	}

	output, err := withPatchedCLIIO(t, "", func() error {
		return RunTaskComplete(context.Background(), cfg, []string{"1"})
	})
	if err != nil {
		t.Fatalf("RunTaskComplete returned error: %v", err)
	}

	if !strings.Contains(output, taskText) {
		t.Fatalf("expected completion output to include completed task text %q, got: %q", taskText, output)
	}

	updatedTodo, err := os.ReadFile(cfg.TodoPath)
	if err != nil {
		t.Fatalf("failed to read updated todo file: %v", err)
	}
	if !strings.Contains(string(updatedTodo), "- [x] "+taskText) {
		t.Fatalf("expected completed marker in todo for task %q, got:\n%s", taskText, string(updatedTodo))
	}
}

func TestRunTaskComplete_DailyTaskCompletionUpdatesSourceNoteAndTodo(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestTaskConfig(t, tmpDir)
	cfg.StatePath = filepath.Join(tmpDir, "state.json")

	todayNotePath := notes.BuildDailyNotePath(cfg.DiaryPath, time.Now())
	if err := os.MkdirAll(filepath.Dir(todayNotePath), 0750); err != nil {
		t.Fatalf("failed to create diary directory: %v", err)
	}

	const taskText = "Daily-source completion text"
	dailyContent := `# Today

## Tasks

- [ ] ` + taskText + `
`
	if err := notes.WriteNote(context.Background(), todayNotePath, dailyContent); err != nil {
		t.Fatalf("failed to create daily note: %v", err)
	}

	origDryRun, origQuiet, origJSON, origVerbose, origNoColor := syncDryRun, syncQuiet, syncJSON, syncVerbose, syncNoColor
	t.Cleanup(func() {
		syncDryRun, syncQuiet, syncJSON, syncVerbose, syncNoColor = origDryRun, origQuiet, origJSON, origVerbose, origNoColor
	})
	syncDryRun = false
	syncQuiet = false
	syncJSON = false
	syncVerbose = false
	syncNoColor = true

	if err := syncTasks(context.Background(), cfg); err != nil {
		t.Fatalf("syncTasks failed: %v", err)
	}

	output, err := withPatchedCLIIO(t, "", func() error {
		return RunTaskComplete(context.Background(), cfg, []string{"1"})
	})
	if err != nil {
		t.Fatalf("RunTaskComplete returned error: %v", err)
	}

	if !strings.Contains(output, taskText) {
		t.Fatalf("expected completion output to include completed task text %q, got: %q", taskText, output)
	}

	updatedDaily, err := os.ReadFile(todayNotePath)
	if err != nil {
		t.Fatalf("failed to read updated daily note: %v", err)
	}
	if !strings.Contains(string(updatedDaily), "- [x] "+taskText) {
		t.Fatalf("expected completed marker in daily note for task %q, got:\n%s", taskText, string(updatedDaily))
	}

	updatedTodo, err := os.ReadFile(cfg.TodoPath)
	if err != nil {
		t.Fatalf("failed to read updated todo file: %v", err)
	}
	if !strings.Contains(string(updatedTodo), "- [x] "+taskText) {
		t.Fatalf("expected todo list to stay in sync with completed task %q, got:\n%s", taskText, string(updatedTodo))
	}
}

func TestRunTaskEdit_ShowsListBeforePromptWhenSelectionMissing(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestTaskConfig(t, tmpDir)
	cfg.StatePath = filepath.Join(tmpDir, "state.json")

	todoContent := `# To-Do List

## Tasks

- [ ] Editable task
`
	if err := notes.WriteNote(context.Background(), cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("failed to create todo file: %v", err)
	}

	origEditor := os.Getenv("EDITOR")
	t.Cleanup(func() {
		_ = os.Setenv("EDITOR", origEditor)
	})
	if err := os.Setenv("EDITOR", "true"); err != nil {
		t.Fatalf("failed to set EDITOR: %v", err)
	}

	output, err := withPatchedCLIIO(t, "1\n", func() error {
		return RunTaskEdit(context.Background(), cfg, nil)
	})
	if err != nil {
		t.Fatalf("RunTaskEdit returned error: %v", err)
	}

	listIdx := strings.Index(output, "Editable task")
	promptIdx := strings.Index(output, "Task number to edit:")
	if listIdx == -1 || promptIdx == -1 {
		t.Fatalf("expected both task list and prompt in output, got: %q", output)
	}
	if listIdx > promptIdx {
		t.Fatalf("expected task list to be shown before prompt; listIdx=%d promptIdx=%d output=%q", listIdx, promptIdx, output)
	}
}

func TestTaskEditFreeTextSanitizationHelpersStripMetadata(t *testing.T) {
	raw := "Plan release [P2] #ops #urgent"

	clean := stripPriorityMarker(stripTaskTags(raw))
	if clean != "Plan release" {
		t.Fatalf("expected cleaned task text %q, got %q", "Plan release", clean)
	}
}

func TestDisplayTaskLine_KeepsPriorityAndTagsStructuredAfterTextSanitization(t *testing.T) {
	task := tasks.Task{
		Text:     "Plan release [P2] #ops #urgent",
		Priority: "P2",
		Tags:     []string{"ops", "urgent"},
	}

	line := displayTaskLine(task)

	if !strings.Contains(line, "[P2]") {
		t.Fatalf("expected display line to include structured priority, got %q", line)
	}
	if !strings.Contains(line, "#ops #urgent") {
		t.Fatalf("expected display line to include structured tags, got %q", line)
	}
	if strings.Contains(line, "Plan release [P2]") {
		t.Fatalf("expected free-text portion to exclude inline priority marker, got %q", line)
	}
	if strings.Contains(line, "tags:") {
		t.Fatalf("expected display line to omit the literal tags label, got %q", line)
	}
	if strings.Count(line, "#ops") != 1 || strings.Count(line, "#urgent") != 1 {
		t.Fatalf("expected tags to appear once via structured metadata, got %q", line)
	}
}
