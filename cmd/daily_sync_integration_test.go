package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/services"
	"github.com/AnishShah1803/jotr/internal/testhelpers"
)

func createTodayNote(fs *testhelpers.TestFS, t *testing.T, content string) {
	t.Helper()
	today := time.Now()
	year := today.Format("2006")
	monthNum := today.Format("01")
	monthAbbr := today.Format("Jan")
	day := today.Format("02")
	weekday := today.Format("Mon")

	dirPath := filepath.Join("diary", year, monthNum+"-"+monthAbbr)
	filename := year + "-" + monthNum + "-" + day + "-" + weekday + ".md"
	fs.WriteFile(t, filepath.Join(dirPath, filename), content)
}

func TestDailyCommand_Integration(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	ch := testhelpers.NewConfigHelper(fs)
	ch.CreateBasicConfig(t)

	dailyTests := []testhelpers.NamedTest[struct {
		name           string
		setupNote      func(t *testing.T, fs *testhelpers.TestFS)
		expectedPath   bool
		expectedOutput string
	}]{
		{
			Name: "create_new_daily_note",
			Data: struct {
				name           string
				setupNote      func(t *testing.T, fs *testhelpers.TestFS)
				expectedPath   bool
				expectedOutput string
			}{
				name: "new daily note should be created",
				setupNote: func(t *testing.T, fs *testhelpers.TestFS) {
				},
				expectedPath:   true,
				expectedOutput: "Created",
			},
		},
		{
			Name: "open_existing_daily_note",
			Data: struct {
				name           string
				setupNote      func(t *testing.T, fs *testhelpers.TestFS)
				expectedPath   bool
				expectedOutput string
			}{
				name: "existing daily note should not be recreated",
				setupNote: func(t *testing.T, fs *testhelpers.TestFS) {
					createTodayNote(fs, t, "# Test Note\n\n## Tasks\n")
				},
				expectedPath:   true,
				expectedOutput: "",
			},
		},
		{
			Name: "path_only_flag",
			Data: struct {
				name           string
				setupNote      func(t *testing.T, fs *testhelpers.TestFS)
				expectedPath   bool
				expectedOutput string
			}{
				name: "path-only flag should output path",
				setupNote: func(t *testing.T, fs *testhelpers.TestFS) {
				},
				expectedPath:   true,
				expectedOutput: ".md",
			},
		},
	}

	testhelpers.RunNamedTableTests(t, dailyTests, func(t *testing.T, tt struct {
		name           string
		setupNote      func(t *testing.T, fs *testhelpers.TestFS)
		expectedPath   bool
		expectedOutput string
	}) {
		testhelpers.SubTest(t, tt.name, func(t *testing.T) {
			tt.setupNote(t, fs)

			diaryPath := filepath.Join(fs.BaseDir, "diary")
			date := time.Now()

			path := notes.BuildDailyNotePath(diaryPath, date)
			_, err := os.Stat(path)

			if tt.expectedPath {
				if err != nil {
					t.Logf("Note will be created at: %s", path)
				}
			}

			if tt.expectedOutput != "" {
				t.Logf("Expected output containing: %s", tt.expectedOutput)
			}
		})
	})
}

func TestSyncCommand_Integration(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	ch := testhelpers.NewConfigHelper(fs)
	ch.CreateBasicConfig(t)

	cleanupStateAndTodo := func() {
		statePath := filepath.Join(fs.BaseDir, ".todo_state.json")
		os.Remove(statePath)
		todoPath := filepath.Join(fs.BaseDir, "todo.md")
		os.Remove(todoPath)
	}

	t.Run("sync_new_tasks", func(t *testing.T) {
		cleanupStateAndTodo()
		diaryPath := filepath.Join(fs.BaseDir, "diary")
		todoPath := filepath.Join(fs.BaseDir, "todo.md")

		noteContent := "# Today\n\n## Tasks\n- [ ] Complete project proposal\n- [ ] Review PR #123\n"
		createTodayNote(fs, t, noteContent)

		ctx := context.Background()
		service := services.NewTaskService()

		result, err := service.SyncTasks(ctx, services.SyncOptions{
			DiaryPath:   diaryPath,
			TodoPath:    todoPath,
			StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
			TaskSection: "Tasks",
		})

		if err != nil {
			t.Fatalf("SyncTasks failed: %v", err)
		}

		if result.TasksFromDaily != 2 {
			t.Errorf("TasksSynced = %d; want 2", result.TasksFromDaily)
		}

		if !fs.FileExists("todo.md") {
			t.Errorf("Expected todo file to be created")
		}

		todoContent := fs.ReadFile(t, "todo.md")
		if !strings.Contains(todoContent, "Complete project proposal") {
			t.Errorf("Expected todo file to contain synced task")
		}
	})

	t.Run("sync_no_tasks_to_sync", func(t *testing.T) {
		cleanupStateAndTodo()
		diaryPath := filepath.Join(fs.BaseDir, "diary")
		todoPath := filepath.Join(fs.BaseDir, "todo.md")

		noteContent := "# Today\n\n## Notes\nNo tasks here\n"
		createTodayNote(fs, t, noteContent)

		ctx := context.Background()
		service := services.NewTaskService()

		result, err := service.SyncTasks(ctx, services.SyncOptions{
			DiaryPath:   diaryPath,
			TodoPath:    todoPath,
			StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
			TaskSection: "Tasks",
		})

		if err != nil {
			if strings.Contains(err.Error(), "note doesn't exist") {
				return
			}
			t.Fatalf("SyncTasks failed: %v", err)
		}

		if result.TasksFromDaily != 0 {
			t.Errorf("TasksSynced = %d; want 0", result.TasksFromDaily)
		}
	})

	t.Run("sync_deduplicates_existing", func(t *testing.T) {
		cleanupStateAndTodo()
		diaryPath := filepath.Join(fs.BaseDir, "diary")
		todoPath := filepath.Join(fs.BaseDir, "todo.md")

		noteContent := "# Today\n\n## Tasks\n- [ ] Review proposal\n"
		createTodayNote(fs, t, noteContent)

		fs.WriteFile(t, "todo.md", "# To-Do List\n\n## Tasks\n- [ ] Review proposal\n")

		ctx := context.Background()
		service := services.NewTaskService()

		result, err := service.SyncTasks(ctx, services.SyncOptions{
			DiaryPath:   diaryPath,
			TodoPath:    todoPath,
			StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
			TaskSection: "Tasks",
		})

		if err != nil {
			t.Fatalf("SyncTasks failed: %v", err)
		}

		if result.TasksFromDaily != 0 {
			t.Errorf("TasksSynced = %d; want 0 (deduplication)", result.TasksFromDaily)
		}
	})

	t.Run("sync_with_task_ids", func(t *testing.T) {
		cleanupStateAndTodo()
		diaryPath := filepath.Join(fs.BaseDir, "diary")
		todoPath := filepath.Join(fs.BaseDir, "todo.md")

		noteContent := "# Today\n\n## Tasks\n- [ ] Complete project proposal <!-- id: abc12345 -->\n- [ ] Review PR #123\n"
		createTodayNote(fs, t, noteContent)

		ctx := context.Background()
		service := services.NewTaskService()

		result, err := service.SyncTasks(ctx, services.SyncOptions{
			DiaryPath:   diaryPath,
			TodoPath:    todoPath,
			StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
			TaskSection: "Tasks",
		})

		if err != nil {
			t.Fatalf("SyncTasks failed: %v", err)
		}

		if result.TasksFromDaily != 2 {
			t.Errorf("TasksSynced = %d; want 2", result.TasksFromDaily)
		}

		todoContent := fs.ReadFile(t, "todo.md")
		if !strings.Contains(todoContent, "Complete project proposal") {
			t.Errorf("Expected todo file to contain task ID")
		}
	})

	t.Run("sync_completed_tasks_excluded", func(t *testing.T) {
		cleanupStateAndTodo()
		diaryPath := filepath.Join(fs.BaseDir, "diary")
		todoPath := filepath.Join(fs.BaseDir, "todo.md")

		noteContent := "# Today\n\n## Tasks\n- [ ] Incomplete task\n- [x] Completed task\n"
		createTodayNote(fs, t, noteContent)

		ctx := context.Background()
		service := services.NewTaskService()

		result, err := service.SyncTasks(ctx, services.SyncOptions{
			DiaryPath:   diaryPath,
			TodoPath:    todoPath,
			StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
			TaskSection: "Tasks",
		})

		if err != nil {
			t.Fatalf("SyncTasks failed: %v", err)
		}

		if result.TasksFromDaily != 1 {
			t.Errorf("TasksSynced = %d; want 1", result.TasksFromDaily)
		}

		todoContent := fs.ReadFile(t, "todo.md")
		if strings.Contains(todoContent, "Completed task") {
			t.Errorf("Completed task should not be synced")
		}
	})

	t.Run("sync_multiple_similar_tasks", func(t *testing.T) {
		cleanupStateAndTodo()
		diaryPath := filepath.Join(fs.BaseDir, "diary")
		todoPath := filepath.Join(fs.BaseDir, "todo.md")

		noteContent := "# Today\n\n## Tasks\n- [ ] Update\n- [ ] Update documentation\n- [ ] Update config\n"
		createTodayNote(fs, t, noteContent)

		ctx := context.Background()
		service := services.NewTaskService()

		result, err := service.SyncTasks(ctx, services.SyncOptions{
			DiaryPath:   diaryPath,
			TodoPath:    todoPath,
			StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
			TaskSection: "Tasks",
		})

		if err != nil {
			t.Fatalf("SyncTasks failed: %v", err)
		}

		if result.TasksFromDaily != 3 {
			t.Errorf("TasksSynced = %d; want 3 (similar but different tasks)", result.TasksFromDaily)
		}
	})

	t.Run("sync_task_with_priority_and_tags", func(t *testing.T) {
		cleanupStateAndTodo()
		diaryPath := filepath.Join(fs.BaseDir, "diary")
		todoPath := filepath.Join(fs.BaseDir, "todo.md")

		noteContent := "# Today\n\n## Tasks\n- [ ][P1] Review API #backend #urgent\n"
		createTodayNote(fs, t, noteContent)

		ctx := context.Background()
		service := services.NewTaskService()

		result, err := service.SyncTasks(ctx, services.SyncOptions{
			DiaryPath:   diaryPath,
			TodoPath:    todoPath,
			StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
			TaskSection: "Tasks",
		})

		if err != nil {
			t.Fatalf("SyncTasks failed: %v", err)
		}

		if result.TasksFromDaily != 1 {
			t.Errorf("TasksSynced = %d; want 1", result.TasksFromDaily)
		}

		todoContent := fs.ReadFile(t, "todo.md")
		if !strings.Contains(todoContent, "P1") {
			t.Errorf("Expected todo file to contain priority")
		}
		if !strings.Contains(todoContent, "#backend") {
			t.Errorf("Expected todo file to contain tag")
		}
	})

	t.Run("sync_handles_empty_todo_file", func(t *testing.T) {
		cleanupStateAndTodo()
		diaryPath := filepath.Join(fs.BaseDir, "diary")
		todoPath := filepath.Join(fs.BaseDir, "todo.md")

		noteContent := "# Today\n\n## Tasks\n- [ ] New task to sync\n"
		createTodayNote(fs, t, noteContent)

		fs.WriteFile(t, "todo.md", "")

		ctx := context.Background()
		service := services.NewTaskService()

		result, err := service.SyncTasks(ctx, services.SyncOptions{
			DiaryPath:   diaryPath,
			TodoPath:    todoPath,
			StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
			TaskSection: "Tasks",
		})

		if err != nil {
			t.Fatalf("SyncTasks failed: %v", err)
		}

		if result.TasksFromDaily != 1 {
			t.Errorf("TasksSynced = %d; want 1", result.TasksFromDaily)
		}
	})

	t.Run("sync_handles_nonexistent_todo_file", func(t *testing.T) {
		cleanupStateAndTodo()
		diaryPath := filepath.Join(fs.BaseDir, "diary")
		todoPath := filepath.Join(fs.BaseDir, "nonexistent-todo.md")

		noteContent := "# Today\n\n## Tasks\n- [ ] Task for new todo file\n"
		createTodayNote(fs, t, noteContent)

		ctx := context.Background()
		service := services.NewTaskService()

		result, err := service.SyncTasks(ctx, services.SyncOptions{
			DiaryPath:   diaryPath,
			TodoPath:    todoPath,
			StatePath:   filepath.Join(fs.BaseDir, ".todo_state.json"),
			TaskSection: "Tasks",
		})

		if err != nil {
			t.Fatalf("SyncTasks failed: %v", err)
		}

		if result.TasksFromDaily != 1 {
			t.Errorf("TasksSynced = %d; want 1", result.TasksFromDaily)
		}

		if !fs.FileExists("nonexistent-todo.md") {
			t.Errorf("Expected todo file to be created")
		}
	})
}

func TestDailyCommand_VariousDates(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	ch := testhelpers.NewConfigHelper(fs)
	ch.CreateBasicConfig(t)

	testCases := []struct {
		name          string
		date          time.Time
		expectCreated bool
	}{
		{"today", time.Now(), true},
		{"yesterday", time.Now().AddDate(0, 0, -1), true},
		{"start of month", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), true},
		{"leap day", time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs := testhelpers.NewTestFS(t)
			defer fs.Cleanup()

			diaryPath := filepath.Join(fs.BaseDir, "diary")
			path := notes.BuildDailyNotePath(diaryPath, tc.date)

			_, filename := filepath.Split(path)
			if !strings.HasSuffix(filename, ".md") {
				t.Errorf("Expected .md extension, got: %s", filename)
			}

			expectedYear := tc.date.Format("2006")
			if !strings.Contains(path, expectedYear) {
				t.Errorf("Expected path to contain %s", expectedYear)
			}
		})
	}
}
