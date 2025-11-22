package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Valid(t *testing.T) {
	// Create a temporary config file
	tmpDir, err := os.MkdirTemp("", "jotr-config-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "jotr")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "config.json")

	// Create valid config
	config := Config{}
	config.Paths.BaseDir = "/tmp/test-jotr"
	config.Paths.DiaryDir = "Diary"
	config.Paths.TodoFilePath = "todo.md"
	config.Format.TaskSection = "Important Things"
	config.Format.CaptureSection = "Captured"
	config.Format.DailyNoteSections = []string{"Notes", "Tasks"}
	config.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	config.Format.DailyNoteDirPattern = "{year}/{month}"
	config.Streaks.IncludeWeekends = false

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set HOME to our temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Load config
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	if loadedConfig.Paths.BaseDir != "/tmp/test-jotr" {
		t.Errorf("Expected base_dir '/tmp/test-jotr', got '%s'", loadedConfig.Paths.BaseDir)
	}

	if loadedConfig.Format.TaskSection != "Important Things" {
		t.Errorf("Expected task_section 'Important Things', got '%s'", loadedConfig.Format.TaskSection)
	}

	if len(loadedConfig.Format.DailyNoteSections) != 2 {
		t.Errorf("Expected 2 daily note sections, got %d", len(loadedConfig.Format.DailyNoteSections))
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-config-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "jotr")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "config.json")

	// Write invalid JSON
	err = os.WriteFile(configFile, []byte("{invalid json}"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Set HOME to our temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Load should fail with JSON error
	_, err = Load()
	if err == nil {
		t.Errorf("Expected JSON parsing error, but load succeeded")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-config-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set HOME to temp directory with no config
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Load should fail with file not found
	_, err = Load()
	if err == nil {
		t.Errorf("Expected file not found error, but load succeeded")
	}
}

func TestLoadConfig_MissingBaseDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-config-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "jotr")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "config.json")

	// Create config without base_dir
	config := Config{}
	config.Format.TaskSection = "Tasks"

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Set HOME to our temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Load should fail due to missing base_dir
	_, err = Load()
	if err == nil {
		t.Errorf("Expected base_dir validation error, but load succeeded")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-config-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set HOME to our temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create config to save
	config := Config{}
	config.Paths.BaseDir = "/tmp/test-save"
	config.Paths.DiaryDir = "Diary"
	config.Paths.TodoFilePath = "todo.md"
	config.Format.TaskSection = "Todo"
	config.Format.CaptureSection = "Quick"
	config.Format.DailyNoteSections = []string{"Notes", "Tasks", "Ideas"}
	config.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	config.Format.DailyNoteDirPattern = "{year}/{month}"
	config.Streaks.IncludeWeekends = true

	// Save config
	err = Save(&config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, ".config", "jotr", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created")
	}

	// Load and verify
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.Paths.BaseDir != "/tmp/test-save" {
		t.Errorf("Expected base_dir '/tmp/test-save', got '%s'", loadedConfig.Paths.BaseDir)
	}

	if loadedConfig.Format.TaskSection != "Todo" {
		t.Errorf("Expected task_section 'Todo', got '%s'", loadedConfig.Format.TaskSection)
	}

	if !loadedConfig.Streaks.IncludeWeekends {
		t.Errorf("Expected include_weekends true, got false")
	}
}

func TestSaveConfig_CreatesBackup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-config-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "jotr")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "config.json")

	// Create initial config
	initialConfig := Config{}
	initialConfig.Paths.BaseDir = "/initial/path"
	initialConfig.Paths.DiaryDir = "Diary"
	initialConfig.Paths.TodoFilePath = "todo.md"
	initialConfig.Format.TaskSection = "Initial Tasks"
	initialConfig.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	initialConfig.Format.DailyNoteDirPattern = "{year}/{month}"

	data, err := json.MarshalIndent(initialConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal initial config: %v", err)
	}

	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Set HOME to our temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Update config
	updatedConfig := Config{}
	updatedConfig.Paths.BaseDir = "/updated/path"
	updatedConfig.Paths.DiaryDir = "Diary"
	updatedConfig.Paths.TodoFilePath = "todo.md"
	updatedConfig.Format.TaskSection = "Updated Tasks"
	updatedConfig.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	updatedConfig.Format.DailyNoteDirPattern = "{year}/{month}"

	err = Save(&updatedConfig)
	if err != nil {
		t.Fatalf("Failed to save updated config: %v", err)
	}

	// Check that a backup file was created
	backupFile := configFile + ".backup"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Errorf("No backup file was created at %s", backupFile)
	} else {
		// Verify backup contains original content
		backupContent, err := os.ReadFile(backupFile)
		if err != nil {
			t.Fatalf("Failed to read backup file: %v", err)
		}

		var backupConfig Config
		err = json.Unmarshal(backupContent, &backupConfig)
		if err != nil {
			t.Fatalf("Failed to parse backup config: %v", err)
		}

		if backupConfig.Paths.BaseDir != "/initial/path" {
			t.Errorf("Backup doesn't contain original content")
		}
	}

	// Verify current config has updated content
	currentContent, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read current config: %v", err)
	}

	var currentConfig Config
	err = json.Unmarshal(currentContent, &currentConfig)
	if err != nil {
		t.Fatalf("Failed to parse current config: %v", err)
	}

	if currentConfig.Paths.BaseDir != "/updated/path" {
		t.Errorf("Current config doesn't contain updated content")
	}
}
