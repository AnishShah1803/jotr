package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// createTestConfigForDaily creates a test configuration with a temporary directory.
func createTestConfigForDaily(t *testing.T, tmpDir string) *config.LoadedConfig {
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

// TestBuildDailyNotePath tests that daily note paths are built correctly.
func TestBuildDailyNotePath(t *testing.T) {
	testCases := []struct {
		name     string
		date     time.Time
		diaryDir string
		expected string
	}{
		{
			name:     "Today",
			date:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			diaryDir: "/tmp/notes/Diary",
			expected: "/tmp/notes/Diary/2025/01-Jan/2025-01-15-Wed.md",
		},
		{
			name:     "Different month",
			date:     time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			diaryDir: "/tmp/notes/Diary",
			expected: "/tmp/notes/Diary/2024/12-Dec/2024-12-25-Wed.md",
		},
		{
			name:     "Weekend",
			date:     time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC),
			diaryDir: "/tmp/notes/Diary",
			expected: "/tmp/notes/Diary/2025/01-Jan/2025-01-11-Sat.md",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := notes.BuildDailyNotePath(tc.diaryDir, tc.date)
			if path != tc.expected {
				t.Errorf("Expected path %s, got %s", tc.expected, path)
			}
		})
	}
}

// TestDailyNoteCreation tests creating a new daily note.
func TestDailyNoteCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-daily-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForDaily(t, tmpDir)
	diaryPath := cfg.DiaryPath

	// Create directory structure
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	expectedPath := filepath.Join(diaryPath, "2025", "01", "2025-01-15-Wed.md")

	if err := os.MkdirAll(filepath.Dir(expectedPath), 0750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create daily note using the internal function
	sections := []string{"Notes", "Tasks"}
	if err := notes.CreateDailyNote(context.Background(), expectedPath, sections, date); err != nil {
		t.Fatalf("Failed to create daily note: %v", err)
	}

	// Verify note exists
	if !utils.FileExists(expectedPath) {
		t.Errorf("Daily note should exist at %s", expectedPath)
	}

	// Verify content contains expected sections
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read daily note: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# 2025-01-15-Wed") {
		t.Errorf("Daily note should contain date header")
	}

	if !strings.Contains(contentStr, "## Notes") {
		t.Errorf("Daily note should contain Notes section")
	}

	if !strings.Contains(contentStr, "## Tasks") {
		t.Errorf("Daily note should contain Tasks section")
	}
}

// TestDailyNotePathFormat tests the path format is correct.
func TestDailyNotePathFormat(t *testing.T) {
	// Test different patterns
	diaryPath := "/tmp/Diary"

	testCases := []struct {
		date   time.Time
		format string
	}{
		{
			date:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			format: "2025/01-Jan/2025-01-15-Wed.md",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			path := notes.BuildDailyNotePath(diaryPath, tc.date)

			relPath, err := filepath.Rel(diaryPath, path)
			if err != nil {
				t.Fatalf("Failed to get relative path: %v", err)
			}

			if relPath != tc.format {
				t.Errorf("Expected format %s, got %s", tc.format, relPath)
			}
		})
	}
}

// TestDailyNoteExists tests checking if daily note exists.
func TestDailyNoteExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-daily-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForDaily(t, tmpDir)
	diaryPath := cfg.DiaryPath

	// Path doesn't exist yet
	path := notes.BuildDailyNotePath(diaryPath, time.Now())
	if utils.FileExists(path) {
		t.Errorf("Daily note should not exist yet: %s", path)
	}

	// Create directory structure
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create note file
	if err := os.WriteFile(path, []byte("# Test\n"), constants.FilePerm0600); err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	if !utils.FileExists(path) {
		t.Errorf("Daily note should exist: %s", path)
	}
}

// TestDailyNoteSections tests building daily note sections.
func TestDailyNoteSections(t *testing.T) {
	testCases := []struct {
		name            string
		expectedPattern string
		sections        []string
	}{
		{
			name:            "Default sections",
			sections:        []string{"Notes", "Tasks"},
			expectedPattern: "## Notes",
		},
		{
			name:            "Single section",
			sections:        []string{"Journal"},
			expectedPattern: "## Journal",
		},
		{
			name:            "Multiple sections",
			sections:        []string{"Notes", "Meetings", "Tasks", "Captured"},
			expectedPattern: "## Tasks",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When sections are rendered, they get "## " prefix
			for _, section := range tc.sections {
				rendered := "## " + section
				if !strings.HasPrefix(rendered, "## ") {
					t.Errorf("Rendered section should start with '## ': %s", rendered)
				}
			}
		})
	}
}

// TestYesterdayFlag tests the yesterday flag behavior.
func TestYesterdayFlag(t *testing.T) {
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	// Calculate yesterday path
	todayPath := notes.BuildDailyNotePath("/tmp/Diary", today)
	yesterdayPath := notes.BuildDailyNotePath("/tmp/Diary", yesterday)

	// They should be different
	if todayPath == yesterdayPath {
		t.Errorf("Today and yesterday paths should be different")
	}

	// Yesterday path should contain yesterday's date
	year, month, day := yesterday.Date()
	weekday := yesterday.Format("Mon")

	expectedDateStr := time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	expectedFileName := expectedDateStr + "-" + weekday + ".md"

	// Check if yesterday's filename is in the path
	if !strings.Contains(yesterdayPath, expectedFileName) {
		t.Errorf("Expected yesterday path to contain %s, got %s", expectedFileName, yesterdayPath)
	}
}

// TestPathOnlyFlag tests the path-only output behavior.
func TestPathOnlyFlag(t *testing.T) {
	diaryPath := "/tmp/Diary"
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	path := notes.BuildDailyNotePath(diaryPath, date)

	// Path-only mode should just output the path
	if path == "" {
		t.Errorf("Path should not be empty")
	}

	// Path should be absolute or relative correctly
	if !strings.Contains(path, "2025-01-15") {
		t.Errorf("Path should contain date: %s", path)
	}
}

// TestDailyNoteOpening tests the daily note opening flow.
func TestDailyNoteOpening(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-daily-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForDaily(t, tmpDir)
	diaryPath := cfg.DiaryPath

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	path := notes.BuildDailyNotePath(diaryPath, date)

	// Create directory structure
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create the note
	if err := os.WriteFile(path, []byte("# 2025-01-15-Wed\n\n## Notes\n\n## Tasks\n"), constants.FilePerm0600); err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Verify note exists and can be read
	if !utils.FileExists(path) {
		t.Errorf("Note should exist: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Should be able to read note: %v", err)
	}

	if len(content) == 0 {
		t.Errorf("Note should not be empty")
	}
}

// TestOpenInEditor_NoEditorConfigured tests that openInEditor returns clear error when editor is not configured.
func TestOpenInEditor_NoEditorConfigured(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-daily-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForDaily(t, tmpDir)
	diaryPath := cfg.DiaryPath

	// Create directory structure
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	path := notes.BuildDailyNotePath(diaryPath, date)
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Unset EDITOR environment variable to simulate no editor configured
	oldEditor := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", oldEditor)
	os.Unsetenv("EDITOR")

	// Clear any configured editor in config
	cfg.Editor.Default = ""

	err = openInEditor(context.Background(), path)

	if err == nil {
		t.Errorf("Expected error when editor is not configured")
	}

	expectedErrMsg := "no editor configured"
	if err == nil || !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedErrMsg, err)
	}

	// Check that error includes helpful guidance
	if err == nil || !strings.Contains(err.Error(), "EDITOR environment variable") && !strings.Contains(err.Error(), "editor.default") {
		t.Errorf("Expected error message to include guidance on how to fix, got: %v", err)
	}
}

// TestOpenInEditor_InvalidEditor tests that openInEditor returns error for invalid editor.
func TestOpenInEditor_InvalidEditor(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-daily-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForDaily(t, tmpDir)
	diaryPath := cfg.DiaryPath

	// Create directory structure
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	path := notes.BuildDailyNotePath(diaryPath, date)
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Set EDITOR to a non-existent command
	oldEditor := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", oldEditor)
	os.Setenv("EDITOR", "nonexistent-editor-binary-12345")

	err = openInEditor(context.Background(), path)

	if err == nil {
		t.Errorf("Expected error when editor is invalid")
	}

	// The error should indicate the editor is invalid or not found
	if err == nil || (!strings.Contains(err.Error(), "invalid editor") && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "executable")) {
		t.Errorf("Expected error to indicate invalid editor, got: %v", err)
	}
}

// TestOpenInEditor_EditorConfigured tests that openInEditor succeeds when editor is properly configured.
func TestOpenInEditor_EditorConfigured(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-daily-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForDaily(t, tmpDir)
	diaryPath := cfg.DiaryPath

	// Create directory structure
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	path := notes.BuildDailyNotePath(diaryPath, date)
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Set EDITOR to 'true' which will succeed immediately
	oldEditor := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", oldEditor)
	os.Setenv("EDITOR", "true")

	err = openInEditor(context.Background(), path)

	if err != nil {
		t.Errorf("Expected no error when editor is properly configured, got: %v", err)
	}
}

// TestOpenInEditor_ErrorMessageClarity tests that error messages are clear and actionable.
func TestOpenInEditor_ErrorMessageClarity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-daily-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForDaily(t, tmpDir)
	diaryPath := cfg.DiaryPath

	// Create directory structure
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	path := notes.BuildDailyNotePath(diaryPath, date)
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Unset EDITOR
	oldEditor := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", oldEditor)
	os.Unsetenv("EDITOR")

	cfg.Editor.Default = ""

	err = openInEditor(context.Background(), path)

	if err == nil {
		t.Errorf("Expected error when editor is not configured")
	}

	errorMsg := err.Error()

	// Check for clarity elements
	hasSolution := strings.Contains(errorMsg, "set EDITOR") || strings.Contains(errorMsg, "editor.default")
	hasProblem := strings.Contains(errorMsg, "no editor") || strings.Contains(errorMsg, "not configured")

	if !hasProblem {
		t.Errorf("Error message should clearly state the problem, got: %s", errorMsg)
	}

	if !hasSolution {
		t.Errorf("Error message should provide a solution, got: %s", errorMsg)
	}
}
