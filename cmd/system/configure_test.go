package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/testhelpers"
)

func TestConfigureCommand_ConfigureCmdExists(t *testing.T) {
	if ConfigureCmd == nil {
		t.Error("ConfigureCmd should not be nil")
	}
	if ConfigureCmd.Use != "configure" {
		t.Errorf("Expected ConfigureCmd.Use to be 'configure', got %s", ConfigureCmd.Use)
	}
	if len(ConfigureCmd.Aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(ConfigureCmd.Aliases))
	}
	if ConfigureCmd.Aliases[0] != "config" {
		t.Errorf("Expected alias 'config', got %s", ConfigureCmd.Aliases[0])
	}
	if ConfigureCmd.Aliases[1] != "cfg" {
		t.Errorf("Expected alias 'cfg', got %s", ConfigureCmd.Aliases[1])
	}
}

func TestConfigureCommand_CommandIntegration(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	tmpDir := t.TempDir()
	input := fmt.Sprintf("%s\nDiary\ntodo\n\n", tmpDir)

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetArgs([]string{"configure"})

	output := testhelpers.CaptureStdout(func() {
		_ = rootCmd.Execute()
	})

	if !strings.Contains(output, "Configuration Wizard") {
		t.Error("Expected output to contain 'Configuration Wizard'")
	}
}

func TestConfigureCommand_ValidatesEmptyBaseDir(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	input := "\nDiary\ntodo\n\n"

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetOut(&strings.Builder{})
	rootCmd.SetErr(&strings.Builder{})
	rootCmd.SetArgs([]string{"configure"})

	err := rootCmd.Execute()

	if err == nil {
		t.Error("Expected error when base directory is empty")
	}
	if err != nil && !strings.Contains(err.Error(), "base directory cannot be empty") {
		t.Errorf("Expected error about empty base directory, got: %v", err)
	}
}

func TestConfigureCommand_SavesConfigWithValidInput(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	tmpDir := t.TempDir()
	customDiary := "Journal"
	customTodo := "Tasks"
	customPdp := "PDP"

	input := fmt.Sprintf("%s\n%s\n%s\n%s\n", tmpDir, customDiary, customTodo, customPdp)

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetOut(&strings.Builder{})
	rootCmd.SetErr(&strings.Builder{})
	rootCmd.SetArgs([]string{"configure"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("ConfigureCmd failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "jotr", "config.json")

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if cfg.Paths.BaseDir != tmpDir {
		t.Errorf("Expected base_dir %s, got %s", tmpDir, cfg.Paths.BaseDir)
	}
	if cfg.Paths.DiaryDir != customDiary {
		t.Errorf("Expected diary_dir '%s', got %s", customDiary, cfg.Paths.DiaryDir)
	}
	if cfg.Paths.TodoFilePath != customTodo {
		t.Errorf("Expected todo_file_path '%s', got %s", customTodo, cfg.Paths.TodoFilePath)
	}
	if cfg.Paths.PDPFilePath != customPdp {
		t.Errorf("Expected pdp_file_path '%s', got %s", customPdp, cfg.Paths.PDPFilePath)
	}
}

func TestConfigureCommand_AppliesDefaultValues(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	tmpDir := t.TempDir()
	input := fmt.Sprintf("%s\n\n\n\n", tmpDir)

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetOut(&strings.Builder{})
	rootCmd.SetErr(&strings.Builder{})
	rootCmd.SetArgs([]string{"configure"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("ConfigureCmd failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "jotr", "config.json")

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if cfg.Paths.DiaryDir != "Diary" {
		t.Errorf("Expected default diary_dir 'Diary', got %s", cfg.Paths.DiaryDir)
	}
	if cfg.Paths.TodoFilePath != "todo" {
		t.Errorf("Expected default todo_file_path 'todo', got %s", cfg.Paths.TodoFilePath)
	}
	if cfg.Paths.PDPFilePath != "" {
		t.Errorf("Expected default pdp_file_path '', got %s", cfg.Paths.PDPFilePath)
	}
}

func TestConfigureCommand_ExpandsTildeInPath(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, "Documents", "Notes")

	input := "~/Documents/Notes\nDiary\ntodo\n\n"

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetOut(&strings.Builder{})
	rootCmd.SetErr(&strings.Builder{})
	rootCmd.SetArgs([]string{"configure"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("ConfigureCmd failed: %v", err)
	}

	configPath := filepath.Join(homeDir, ".config", "jotr", "config.json")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if cfg.Paths.BaseDir != expectedDir {
		t.Errorf("Expected base_dir with tilde expanded to %s, got %s", expectedDir, cfg.Paths.BaseDir)
	}
}

func TestConfigureCommand_TodoFileStripsMdExtension(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	tmpDir := t.TempDir()
	input := fmt.Sprintf("%s\nDiary\ntasks.md\n\n", tmpDir)

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetOut(&strings.Builder{})
	rootCmd.SetErr(&strings.Builder{})
	rootCmd.SetArgs([]string{"configure"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("ConfigureCmd failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "jotr", "config.json")

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if cfg.Paths.TodoFilePath != "tasks" {
		t.Errorf("Expected todo_file_path 'tasks' (without .md), got %s", cfg.Paths.TodoFilePath)
	}
}

func TestConfigureCommand_NestedDirectoryPath(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "work", "projects", "notes")

	input := fmt.Sprintf("%s\nDiary\ntodo\n\n", nestedPath)

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetOut(&strings.Builder{})
	rootCmd.SetErr(&strings.Builder{})
	rootCmd.SetArgs([]string{"configure"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("ConfigureCmd failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "jotr", "config.json")

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if cfg.Paths.BaseDir != nestedPath {
		t.Errorf("Expected base_dir %s, got %s", nestedPath, cfg.Paths.BaseDir)
	}
}

func TestConfigureCommand_OutputContainsAllSteps(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	tmpDir := t.TempDir()
	input := fmt.Sprintf("%s\nDiary\ntodo\n\n", tmpDir)

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetArgs([]string{"configure"})

	output := testhelpers.CaptureStdout(func() {
		_ = rootCmd.Execute()
	})

	requiredMessages := []string{
		"Configuration Wizard",
		"Base Directory",
		"Diary Directory",
		"Todo File Path",
		"PDP File",
		"Configuration complete",
	}

	for _, msg := range requiredMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput: %s", msg, output)
		}
	}
}

func TestConfigureCommand_AppliesFormatDefaults(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	tmpDir := t.TempDir()
	input := fmt.Sprintf("%s\nDiary\ntodo\n\n", tmpDir)

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetOut(&strings.Builder{})
	rootCmd.SetErr(&strings.Builder{})
	rootCmd.SetArgs([]string{"configure"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("ConfigureCmd failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "jotr", "config.json")

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if cfg.Format.TaskSection != "Important Things" {
		t.Errorf("Expected default TaskSection 'Important Things', got %s", cfg.Format.TaskSection)
	}
	if cfg.Format.CaptureSection != "Captured" {
		t.Errorf("Expected default CaptureSection 'Captured', got %s", cfg.Format.CaptureSection)
	}
	if len(cfg.Format.DailyNoteSections) != 2 {
		t.Errorf("Expected 2 default daily note sections, got %d", len(cfg.Format.DailyNoteSections))
	}
}

func TestConfigureCommand_AppliesAIAndStreakDefaults(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	tmpDir := t.TempDir()
	input := fmt.Sprintf("%s\nDiary\ntodo\n\n", tmpDir)

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetOut(&strings.Builder{})
	rootCmd.SetErr(&strings.Builder{})
	rootCmd.SetArgs([]string{"configure"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("ConfigureCmd failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "jotr", "config.json")

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if !cfg.AI.Enabled {
		t.Error("Expected AI.Enabled to be true by default")
	}
	if cfg.AI.Command != "auggie -p --quiet" {
		t.Errorf("Expected default AI.Command 'auggie -p --quiet', got %s", cfg.AI.Command)
	}
	if cfg.Streaks.IncludeWeekends {
		t.Error("Expected Streaks.IncludeWeekends to be false by default")
	}
}

func TestConfigureCommand_AllDefaultsApplied(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	tmpDir := t.TempDir()
	input := fmt.Sprintf("%s\n\n\n\n", tmpDir)

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetOut(&strings.Builder{})
	rootCmd.SetErr(&strings.Builder{})
	rootCmd.SetArgs([]string{"configure"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("ConfigureCmd failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "jotr", "config.json")

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if cfg.Paths.DiaryDir != "Diary" {
		t.Errorf("Expected DiaryDir 'Diary', got %s", cfg.Paths.DiaryDir)
	}
	if cfg.Paths.TodoFilePath != "todo" {
		t.Errorf("Expected TodoFilePath 'todo', got %s", cfg.Paths.TodoFilePath)
	}
	if cfg.Paths.PDPFilePath != "" {
		t.Errorf("Expected PDPFilePath '', got %s", cfg.Paths.PDPFilePath)
	}
}

func TestConfigureCommand_WritesValidJSON(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	tmpDir := t.TempDir()
	input := fmt.Sprintf("%s\nDiary\ntodo\n\n", tmpDir)

	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetOut(&strings.Builder{})
	rootCmd.SetErr(&strings.Builder{})
	rootCmd.SetArgs([]string{"configure"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("ConfigureCmd failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "jotr", "config.json")

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("Config file is not valid JSON: %v\nContent: %s", err, string(content))
	}

	if cfg.Version == "" {
		t.Error("Expected config to have a version field")
	}
}

func TestConfigureCommand_AliasConfig(t *testing.T) {
	rootCmd := &cobra.Command{Use: "jotr"}
	rootCmd.AddCommand(ConfigureCmd)

	aliases := ConfigureCmd.Aliases
	if len(aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(aliases))
	}
	if aliases[0] != "config" {
		t.Errorf("Expected alias 'config', got %s", aliases[0])
	}
	if aliases[1] != "cfg" {
		t.Errorf("Expected alias 'cfg', got %s", aliases[1])
	}
}

func TestConfigureCommand_AliasCommandsWork(t *testing.T) {
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	tmpDir := t.TempDir()
	input := fmt.Sprintf("%s\nDiary\ntodo\n\n", tmpDir)

	for _, alias := range []string{"config", "cfg"} {
		alias := alias
		t.Run(alias, func(t *testing.T) {
			rootCmd := &cobra.Command{Use: "jotr"}
			rootCmd.AddCommand(ConfigureCmd)

			rootCmd.SetIn(strings.NewReader(input))
			rootCmd.SetArgs([]string{alias})

			output := testhelpers.CaptureStdout(func() {
				_ = rootCmd.Execute()
			})

			if !strings.Contains(output, "Configuration Wizard") {
				t.Errorf("Expected alias %s to run configure wizard, but it didn't", alias)
			}
		})
	}
}
