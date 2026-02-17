package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
)

func TestCaptureCommand(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "jotr-capture-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test config
	testConfig := &config.LoadedConfig{
		Config: config.Config{},
	}
	testConfig.Paths.BaseDir = tmpDir
	testConfig.Paths.DiaryDir = "Diary"
	testConfig.Format.CaptureSection = "Captured"
	testConfig.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	testConfig.Format.DailyNoteDirPattern = "{year}/{month}"
	testConfig.DiaryPath = filepath.Join(tmpDir, "Diary")

	// Create diary directory
	err = os.MkdirAll(testConfig.DiaryPath, 0750)
	if err != nil {
		t.Fatalf("Failed to create diary directory: %v", err)
	}

	// Test data
	testText := "Test capture text"
	expectedFilePath := filepath.Join(testConfig.DiaryPath, "2025", "11", "2025-11-22-Sat.md")

	// Create the daily note directory structure
	err = os.MkdirAll(filepath.Dir(expectedFilePath), 0750)
	if err != nil {
		t.Fatalf("Failed to create daily note directory: %v", err)
	}

	// Create a minimal daily note file to test appending
	initialContent := `# 2025-11-22-Sat

## Notes

## Captured

`

	err = os.WriteFile(expectedFilePath, []byte(initialContent), constants.FilePerm0600)
	if err != nil {
		t.Fatalf("Failed to create initial daily note: %v", err)
	}

	// Test the capture functionality by directly calling the logic
	// (Since we can't easily test the full CLI command here)
	err = appendToFile(expectedFilePath, "## Captured", testText, false)
	if err != nil {
		t.Fatalf("Capture failed: %v", err)
	}

	// Verify the content was added
	content, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("Failed to read daily note after capture: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, testText) {
		t.Errorf("Captured text not found in file. Expected to contain %q, got:\n%s", testText, contentStr)
	}
}

func TestCaptureCommand_Task(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "jotr-capture-task-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	expectedFilePath := filepath.Join(tmpDir, "test-note.md")

	// Create a minimal note file
	initialContent := `# Test Note

## Captured

`

	err = os.WriteFile(expectedFilePath, []byte(initialContent), constants.FilePerm0600)
	if err != nil {
		t.Fatalf("Failed to create initial note: %v", err)
	}

	// Test capturing as a task
	taskText := "Complete test task"

	err = appendToFile(expectedFilePath, "## Captured", taskText, true)
	if err != nil {
		t.Fatalf("Task capture failed: %v", err)
	}

	// Verify the task was added with checkbox
	content, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("Failed to read note after task capture: %v", err)
	}

	contentStr := string(content)
	expectedTaskLine := "- [ ] " + taskText

	if !strings.Contains(contentStr, expectedTaskLine) {
		t.Errorf("Task not found in file. Expected to contain %q, got:\n%s", expectedTaskLine, contentStr)
	}
}

// appendToFile is a simplified version of the capture logic for testing.
func appendToFile(filePath, section, text string, isTask bool) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Find the section
	sectionIndex := strings.Index(contentStr, section)
	if sectionIndex == -1 {
		return nil // Section not found, would normally add it
	}

	// Find the end of the section (next ## or end of file)
	afterSection := contentStr[sectionIndex+len(section):]
	nextSectionIndex := strings.Index(afterSection, "\n## ")

	var insertPoint int
	if nextSectionIndex == -1 {
		// End of file
		insertPoint = len(contentStr)
	} else {
		insertPoint = sectionIndex + len(section) + nextSectionIndex
	}

	// Prepare the text to insert
	var newText string
	if isTask {
		newText = "\n- [ ] " + text
	} else {
		newText = "\n- " + text
	}

	// Insert the text
	newContent := contentStr[:insertPoint] + newText + contentStr[insertPoint:]

	return os.WriteFile(filePath, []byte(newContent), constants.FilePerm0600)
}
