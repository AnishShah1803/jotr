package cmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
)

func createTestUtilConfig(t *testing.T, tmpDir string) *config.LoadedConfig {
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

	return cfg
}

func TestBulkRename_FileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-bulk-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	ctx := context.Background()
	notePath := filepath.Join(tmpDir, "TestNote.md")
	notes.WriteNote(ctx, notePath, "# Test Note\nOld content here.\n")

	err = bulkRename(context.Background(), cfg, "Old", "New")
	if err != nil {
		t.Errorf("bulkRename should succeed: %v", err)
	}

	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	if !strings.Contains(string(content), "New content here") {
		t.Errorf("Expected content to be updated, got: %s", string(content))
	}
}

func TestBulkRename_NoMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-bulk-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	ctx := context.Background()
	notePath := filepath.Join(tmpDir, "TestNote.md")
	notes.WriteNote(ctx, notePath, "# Test Note\nUnique content.\n")

	err = bulkRename(context.Background(), cfg, "nonexistent", "replacement")
	if err != nil {
		t.Errorf("bulkRename should not error when no matches: %v", err)
	}
}

func TestBulkTag_Placeholder(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-bulk-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	err = bulkTag(context.Background(), cfg, "test-tag")
	if err != nil {
		t.Errorf("bulkTag should not error: %v", err)
	}
}

func TestRunHealthCheck_ValidConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-health-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	if err := os.MkdirAll(cfg.DiaryPath, 0750); err != nil {
		t.Fatalf("Failed to create diary dir: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.TodoPath), 0750); err != nil {
		t.Fatalf("Failed to create todo dir: %v", err)
	}

	origEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", "vim")
	defer os.Setenv("EDITOR", origEditor)

	err = runHealthCheck()
	if err != nil {
		t.Errorf("runHealthCheck should not error with valid config: %v", err)
	}
}

func TestGitCmd_GitAvailable(t *testing.T) {
	result := gitAvailable()
	if result {
		t.Log("Git is available in this environment")
	} else {
		t.Log("Git is not available in this environment")
	}
}

func TestGitCmd_GitStatus(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-git-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skip("Git not available, skipping test")
	}

	err = gitStatus(cfg)
	if err != nil && !strings.Contains(err.Error(), "No changes") && !strings.Contains(err.Error(), "git") {
		t.Errorf("gitStatus should not error on empty repo: %v", err)
	}
}

func TestGitCmd_GitHistory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-git-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skip("Git not available, skipping test")
	}

	gitCmd := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir

		return cmd.Run()
	}

	if err := gitCmd("config", "user.email", "test@test.com"); err != nil {
		t.Skip("Cannot configure git user, skipping test")
	}

	if err := gitCmd("config", "user.name", "Test"); err != nil {
		t.Skip("Cannot configure git user, skipping test")
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := gitCmd("add", "."); err != nil {
		t.Fatalf("Failed to add files: %v", err)
	}

	if err := gitCmd("commit", "-m", "Initial commit"); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	err = gitHistory(cfg)
	if err != nil {
		t.Errorf("gitHistory should not error: %v", err)
	}
}

func TestGitCmd_GitDiff(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-git-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skip("Git not available, skipping test")
	}

	err = gitDiff(cfg)
	if err != nil {
		t.Errorf("gitDiff should not error on empty repo: %v", err)
	}
}

func TestGitCmd_GitCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-git-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skip("Git not available, skipping test")
	}

	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = gitCommit(cfg)
	if err != nil && !strings.Contains(err.Error(), "nothing to commit") {
		t.Errorf("gitCommit should succeed or say nothing to commit: %v", err)
	}
}

func TestRunValidation_ValidConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-validate-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "jotr")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configContent := `{
		"paths": {
			"base_dir": "` + tmpDir + `",
			"diary_dir": "Diary",
			"todo_file_path": "todo.txt"
		},
		"format": {
			"task_section": "Tasks",
			"capture_section": "Captured",
			"daily_note_sections": ["Notes"]
		}
	}`

	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	origHome := os.Getenv("HOME")
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CONFIG_HOME", "")

	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("XDG_CONFIG_HOME", origXDG)
	}()

	err = runValidation(context.Background())
	if err != nil {
		t.Errorf("runValidation should not error: %v", err)
	}
}

func TestRunValidation_NoConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-validate-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)
}

func TestShowQuickMenu_Exit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-quick-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r
	defer w.Close()

	go func() {
		w.Write([]byte("0\n"))
		w.Close()
	}()

	err = showQuickMenu(context.Background(), cfg)
	if err != nil {
		t.Errorf("showQuickMenu with exit should not error: %v", err)
	}
}

func TestShowQuickMenu_OpenTodayNote_Existing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-quick-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	if err := os.MkdirAll(cfg.DiaryPath, 0755); err != nil {
		t.Fatalf("Failed to create diary dir: %v", err)
	}

	origEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", "true")
	defer os.Setenv("EDITOR", origEditor)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r
	defer w.Close()

	go func() {
		w.Write([]byte("1\n"))
		w.Close()
	}()

	err = showQuickMenu(context.Background(), cfg)
	if err != nil {
		t.Errorf("showQuickMenu opening today's note should not error: %v", err)
	}
}

func TestShowQuickMenu_OpenYesterdayNote_NotExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-quick-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r
	defer w.Close()

	go func() {
		w.Write([]byte("2\n"))
		w.Close()
	}()

	err = showQuickMenu(context.Background(), cfg)
	if err != nil && !strings.Contains(err.Error(), "doesn't exist") {
		t.Errorf("showQuickMenu should return error when yesterday's note doesn't exist: %v", err)
	}
}

func TestShowQuickMenu_QuickCapture_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-quick-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r
	defer w.Close()

	go func() {
		w.Write([]byte("3\n\n"))
		w.Close()
	}()

	err = showQuickMenu(context.Background(), cfg)
	if err != nil {
		t.Errorf("showQuickMenu with empty capture should not error: %v", err)
	}
}

func TestShowQuickMenu_QuickCapture_WithText(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-quick-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r
	defer w.Close()

	go func() {
		w.Write([]byte("3\nTest capture text\n"))
		w.Close()
	}()

	err = showQuickMenu(context.Background(), cfg)
	if err != nil {
		t.Errorf("showQuickMenu with capture text should not error: %v", err)
	}
}

func TestShowQuickMenu_Search_Cancelled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-quick-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestUtilConfig(t, tmpDir)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r
	defer w.Close()

	go func() {
		w.Write([]byte("4\n\n"))
		w.Close()
	}()

	err = showQuickMenu(context.Background(), cfg)
	if err != nil {
		t.Errorf("showQuickMenu with empty search should not error: %v", err)
	}
}
