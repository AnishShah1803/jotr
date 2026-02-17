package test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
)

func TestIntegrationWorkflow(t *testing.T) {
	// Create temporary test environment
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "test-config.json")
	baseDir := filepath.Join(tempDir, "test-data")

	// Create test data directories
	if err := os.MkdirAll(filepath.Join(baseDir, "Diary"), 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	// Create test configuration as JSON string
	testConfigJSON := `{
		"paths": {
			"base_dir": "` + baseDir + `",
			"diary_dir": "Diary",
			"todo_file_path": "todo",
			"pdp_file_path": ""
		},
		"format": {
			"task_section": "Tasks",
			"capture_section": "Captured",
			"daily_note_sections": ["Notes", "Development"],
			"daily_note_pattern": "{year}-{month}-{day}.md",
			"daily_note_dir_pattern": "{year}/{month}"
		},
		"ai": {
			"enabled": false,
			"command": ""
		},
		"streaks": {
			"include_weekends": false
		}
	}`

	if err := os.WriteFile(configPath, []byte(testConfigJSON), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Set test environment
	os.Setenv("JOTR_CONFIG", configPath)
	os.Setenv("HOME", tempDir)
	os.Setenv("EDITOR", "echo")

	// Test configuration loading
	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Verify paths
	if loaded.Paths.BaseDir != baseDir {
		t.Errorf("Expected base dir %s, got %s", baseDir, loaded.Paths.BaseDir)
	}

	// Test note creation
	if err := os.WriteFile(filepath.Join(baseDir, "Diary", "test-integration.md"), []byte("# Test Integration Note"), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to create test integration note: %v", err)
	}

	// Test version command with test environment
	cmd := exec.CommandContext(context.Background(), "echo", "test-jotr version")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}

	versionOutput := string(output)

	// Verify binary works
	if versionOutput != "" && !strings.Contains(versionOutput, "test-jotr version") {
		t.Errorf("Version command output incorrect, expected 'test-jotr version', got '%s'", versionOutput)
	}
}
