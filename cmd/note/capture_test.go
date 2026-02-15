package cmd

import (
	"context"
	"fmt"
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

// createTestConfigForCapture creates a test configuration with a temporary directory.
func createTestConfigForCapture(t *testing.T, tmpDir string) *config.LoadedConfig {
	t.Helper()

	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = tmpDir
	cfg.Paths.DiaryDir = "Diary"
	cfg.Format.CaptureSection = "Captured"
	cfg.Format.DailyNoteSections = []string{"Notes", "Tasks"}
	cfg.DiaryPath = filepath.Join(tmpDir, "Diary")

	return cfg
}

// getDailyNotePath returns the expected path for today's daily note.
func getDailyNotePath(cfg *config.LoadedConfig) string {
	today := time.Now()
	return notes.BuildDailyNotePath(cfg.DiaryPath, today)
}

// saveAndRestoreCaptureTask saves the global captureTask flag and restores it after the test.
func saveAndRestoreCaptureTask(t *testing.T) {
	t.Helper()
	originalCaptureTask := captureTask
	t.Cleanup(func() { captureTask = originalCaptureTask })
	captureTask = false
}

// TestCaptureText_Success tests successful text capture to a new note.
func TestCaptureText_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-capture-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForCapture(t, tmpDir)
	saveAndRestoreCaptureTask(t)

	text := "Test capture message"
	err = captureText(context.Background(), cfg, text)
	if err != nil {
		t.Fatalf("captureText() returned error: %v", err)
	}

	// Verify note was created with correct path
	notePath := getDailyNotePath(cfg)
	if !utils.FileExists(notePath) {
		t.Errorf("Expected note to exist at %s, but it doesn't", notePath)
	}

	// Verify content contains captured text
	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, text) {
		t.Errorf("Note content should contain captured text %q, got:\n%s", text, contentStr)
	}

	// Verify Captured section exists
	if !strings.Contains(contentStr, "## Captured") {
		t.Errorf("Note content should contain '## Captured' section, got:\n%s", contentStr)
	}
}

// TestCaptureText_WithTaskFlag tests capture with the task flag set.
func TestCaptureText_WithTaskFlag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-capture-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForCapture(t, tmpDir)
	saveAndRestoreCaptureTask(t)
	captureTask = true

	text := "Important task to do"
	err = captureText(context.Background(), cfg, text)
	if err != nil {
		t.Fatalf("captureText() returned error: %v", err)
	}

	// Verify note was created
	notePath := getDailyNotePath(cfg)
	if !utils.FileExists(notePath) {
		t.Errorf("Expected note to exist at %s, but it doesn't", notePath)
	}

	// Verify content contains task format
	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	contentStr := string(content)

	// Task format should be "- [ ] text (timestamp)"
	if !strings.Contains(contentStr, "- [ ] "+text) {
		t.Errorf("Note content should contain task format '- [ ] %s', got:\n%s", text, contentStr)
	}
}

// TestCaptureText_AppendsToExistingSection tests that capture appends to existing section.
func TestCaptureText_AppendsToExistingSection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-capture-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForCapture(t, tmpDir)
	saveAndRestoreCaptureTask(t)

	// Create note with existing Captured section
	notePath := getDailyNotePath(cfg)
	initialContent := fmt.Sprintf("# %s\n\n## Captured\n\n- First entry\n", time.Now().Format("2006-01-02-Mon"))
	if err := os.WriteFile(notePath, []byte(initialContent), constants.FilePerm0644); err != nil {
		// Create directory structure if needed
		os.MkdirAll(filepath.Dir(notePath), 0755)
		os.WriteFile(notePath, []byte(initialContent), constants.FilePerm0644)
	}

	// Capture another entry
	text := "Second entry"
	err = captureText(context.Background(), cfg, text)
	if err != nil {
		t.Fatalf("captureText() returned error: %v", err)
	}

	// Verify both entries exist
	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	contentStr := string(content)

	// Both entries should be present
	if !strings.Contains(contentStr, "First entry") {
		t.Errorf("Note content should contain 'First entry', got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, text) {
		t.Errorf("Note content should contain new entry '%s', got:\n%s", text, contentStr)
	}
}

// TestCaptureText_CreatesMissingSection tests that capture creates section if missing.
func TestCaptureText_CreatesMissingSection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-capture-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForCapture(t, tmpDir)
	saveAndRestoreCaptureTask(t)

	// Create note without Captured section
	notePath := getDailyNotePath(cfg)
	initialContent := fmt.Sprintf("# %s\n\n## Notes\n\nSome notes\n", time.Now().Format("2006-01-02-Mon"))

	// Create directory structure if needed
	os.MkdirAll(filepath.Dir(notePath), 0755)
	if err := os.WriteFile(notePath, []byte(initialContent), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to create initial note: %v", err)
	}

	// Capture should create Captured section
	text := "New capture"
	err = captureText(context.Background(), cfg, text)
	if err != nil {
		t.Fatalf("captureText() returned error: %v", err)
	}

	// Verify section was created and entry added
	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "## Captured") {
		t.Errorf("Note content should contain '## Captured' section, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, text) {
		t.Errorf("Note content should contain captured text '%s', got:\n%s", text, contentStr)
	}
}

// TestCaptureText_MultipleCaptures tests multiple consecutive captures.
func TestCaptureText_MultipleCaptures(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-capture-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForCapture(t, tmpDir)
	saveAndRestoreCaptureTask(t)

	texts := []string{"First", "Second", "Third"}

	for _, text := range texts {
		if err := captureText(context.Background(), cfg, text); err != nil {
			t.Fatalf("captureText() for '%s' returned error: %v", text, err)
		}
	}

	// Verify all entries exist
	notePath := getDailyNotePath(cfg)
	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	contentStr := string(content)
	for _, text := range texts {
		if !strings.Contains(contentStr, text) {
			t.Errorf("Note content should contain '%s', got:\n%s", text, contentStr)
		}
	}
}

// TestCaptureText_CustomSection tests capture with custom capture section.
func TestCaptureText_CustomSection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-capture-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForCapture(t, tmpDir)
	cfg.Format.CaptureSection = "Ideas"
	saveAndRestoreCaptureTask(t)

	text := "A great idea"
	err = captureText(context.Background(), cfg, text)
	if err != nil {
		t.Fatalf("captureText() returned error: %v", err)
	}

	// Verify custom section was created
	notePath := getDailyNotePath(cfg)
	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	contentStr := string(content)

	// Should use custom section name
	if !strings.Contains(contentStr, "## Ideas") {
		t.Errorf("Note content should contain '## Ideas' section, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, text) {
		t.Errorf("Note content should contain '%s', got:\n%s", text, contentStr)
	}
}

// TestCaptureText_PreservesOtherSections tests that capture doesn't modify other sections.
func TestCaptureText_PreservesOtherSections(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-capture-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForCapture(t, tmpDir)
	saveAndRestoreCaptureTask(t)

	// Create note with multiple sections
	notePath := getDailyNotePath(cfg)
	initialContent := fmt.Sprintf("# %s\n\n## Notes\n\nExisting notes\n\n## Tasks\n\n- [ ] Existing task\n", time.Now().Format("2006-01-02-Mon"))

	// Create directory structure if needed
	os.MkdirAll(filepath.Dir(notePath), 0755)
	if err := os.WriteFile(notePath, []byte(initialContent), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to create initial note: %v", err)
	}

	// Capture should preserve existing content
	text := "New capture"
	err = captureText(context.Background(), cfg, text)
	if err != nil {
		t.Fatalf("captureText() returned error: %v", err)
	}

	// Verify existing content is preserved
	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "Existing notes") {
		t.Errorf("Note content should preserve 'Existing notes', got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "Existing task") {
		t.Errorf("Note content should preserve 'Existing task', got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, text) {
		t.Errorf("Note content should contain new capture '%s', got:\n%s", text, contentStr)
	}
}

// TestCaptureText_EmptyCaptureSection tests behavior with empty capture section.
func TestCaptureText_EmptyCaptureSection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-capture-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForCapture(t, tmpDir)
	cfg.Format.CaptureSection = "" // Empty should default to "Captured"
	saveAndRestoreCaptureTask(t)

	text := "Test with empty section config"
	err = captureText(context.Background(), cfg, text)
	if err != nil {
		t.Fatalf("captureText() returned error: %v", err)
	}

	// Should default to "Captured"
	notePath := getDailyNotePath(cfg)
	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "## Captured") {
		t.Errorf("Note content should contain '## Captured' section (default), got:\n%s", contentStr)
	}
}
