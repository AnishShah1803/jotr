package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anish/jotr/internal/utils"
)

// Config represents the jotr configuration
type Config struct {
	Paths struct {
		BaseDir      string `json:"base_dir"`
		DiaryDir     string `json:"diary_dir"`
		TodoFilePath string `json:"todo_file_path"`
		PDPFilePath  string `json:"pdp_file_path"`
	} `json:"paths"`
	Format struct {
		TaskSection         string   `json:"task_section"`
		CaptureSection      string   `json:"capture_section"`
		DailyNoteSections   []string `json:"daily_note_sections"`
		DailyNotePattern    string   `json:"daily_note_pattern"`
		DailyNoteDirPattern string   `json:"daily_note_dir_pattern"`
	} `json:"format"`
	AI struct {
		Enabled bool   `json:"enabled"`
		Command string `json:"command"`
	} `json:"ai"`
	Streaks struct {
		IncludeWeekends bool `json:"include_weekends"`
	} `json:"streaks"`
	DailyNoteTemplate struct {
		Sections        []TemplateSection `json:"sections"`
		Prompts         TemplatePrompts   `json:"prompts"`
		WeekdaySections []TemplateSection `json:"weekday_sections"`
		WeekendSections []TemplateSection `json:"weekend_sections"`
	} `json:"daily_note_template"`
	Git struct {
		Enabled             bool   `json:"enabled"`
		AutoCommit          bool   `json:"auto_commit"`
		AutoCommitFrequency string `json:"auto_commit_frequency"`
	} `json:"git"`
	Summary struct {
		Sources                   []string `json:"sources"`
		DailyNotesExcludeSections []string `json:"daily_notes_exclude_sections"`
		IncludeNoteTypes          []string `json:"include_note_types"`
	} `json:"summary"`
	Frontmatter struct {
		Fields map[string]FrontmatterField `json:"fields"`
	} `json:"frontmatter"`
	NoteTemplates map[string]interface{} `json:"note_templates"`
}

// TemplateSection represents a section in a template
type TemplateSection struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// TemplatePrompts represents prompts for different times of day
type TemplatePrompts struct {
	Morning []TemplatePrompt `json:"morning"`
	Evening []TemplatePrompt `json:"evening"`
}

// TemplatePrompt represents a single prompt
type TemplatePrompt struct {
	Section  string `json:"section"`
	Question string `json:"question"`
}

// FrontmatterField represents a frontmatter field definition
type FrontmatterField struct {
	Type        string   `json:"type"`
	Values      []string `json:"values,omitempty"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
}

// LoadedConfig contains the loaded configuration with computed paths
type LoadedConfig struct {
	Config
	DiaryPath string
	TodoPath  string
	StatePath string
	PDPPath   string
}

// Load loads the configuration from ~/.config/jotr/config.json
func Load() (*LoadedConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "jotr", "config.json")
	utils.VerboseLog("Loading config from: %s", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s\nRun 'jotr configure' to create it", configPath)
		}
		utils.VerboseLogError("reading config file", err)
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		utils.VerboseLogError("parsing config JSON", err)
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	utils.VerboseLog("Config loaded successfully, base_dir: %s", cfg.Paths.BaseDir)

	// Validate required fields
	if cfg.Paths.BaseDir == "" {
		return nil, fmt.Errorf("base_dir is required in config")
	}

	// Build computed paths
	loaded := &LoadedConfig{Config: cfg}

	// Diary directory
	loaded.DiaryPath = filepath.Join(cfg.Paths.BaseDir, cfg.Paths.DiaryDir)

	// Todo file path (without .md in config)
	todoFilePath := cfg.Paths.TodoFilePath
	if todoFilePath == "" {
		todoFilePath = "todo"
	}
	todoFilePath = strings.TrimSuffix(todoFilePath, ".md")
	loaded.TodoPath = filepath.Join(cfg.Paths.BaseDir, todoFilePath+".md")

	// State file (auto-generated from todo file name, in same directory)
	todoDir := filepath.Dir(loaded.TodoPath)
	todoBasename := filepath.Base(strings.TrimSuffix(todoFilePath, ".md"))
	stateFile := fmt.Sprintf(".%s_state.json", todoBasename)
	loaded.StatePath = filepath.Join(todoDir, stateFile)

	// PDP file path (optional)
	if cfg.Paths.PDPFilePath != "" {
		pdpFilePath := strings.TrimSuffix(cfg.Paths.PDPFilePath, ".md")
		loaded.PDPPath = filepath.Join(cfg.Paths.BaseDir, pdpFilePath+".md")
	}

	// Validate the loaded configuration
	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return loaded, nil
}

// ValidateConfig validates the configuration for common issues
func ValidateConfig(cfg *Config) error {
	// Validate required fields
	if cfg.Paths.BaseDir == "" {
		return fmt.Errorf("base_dir is required in config")
	}

	// Validate base directory exists or is creatable
	if err := validateDirectory(cfg.Paths.BaseDir); err != nil {
		return fmt.Errorf("base_dir validation failed: %w", err)
	}

	// Validate diary directory path
	if cfg.Paths.DiaryDir == "" {
		return fmt.Errorf("diary_dir is required in config")
	}

	// Validate todo file path
	if cfg.Paths.TodoFilePath == "" {
		return fmt.Errorf("todo_file_path is required in config")
	}

	// Validate format settings
	if err := validateFormat(&cfg.Format); err != nil {
		return fmt.Errorf("format validation failed: %w", err)
	}

	// Validate AI settings if enabled
	if cfg.AI.Enabled && cfg.AI.Command == "" {
		return fmt.Errorf("AI is enabled but no command is configured")
	}

	// Validate frontmatter fields
	if err := validateFrontmatter(&cfg.Frontmatter); err != nil {
		return fmt.Errorf("frontmatter validation failed: %w", err)
	}

	return nil
}

func validateDirectory(path string) error {
	// Check if directory exists
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", path)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("cannot access directory: %w", err)
	}

	// Directory doesn't exist, check if we can create it
	parent := filepath.Dir(path)
	if parent == path {
		return fmt.Errorf("cannot create directory at root: %s", path)
	}

	// Check parent directory
	if err := validateDirectory(parent); err != nil {
		return fmt.Errorf("cannot create parent directory: %w", err)
	}

	return nil
}

func validateFormat(format *struct {
	TaskSection         string   `json:"task_section"`
	CaptureSection      string   `json:"capture_section"`
	DailyNoteSections   []string `json:"daily_note_sections"`
	DailyNotePattern    string   `json:"daily_note_pattern"`
	DailyNoteDirPattern string   `json:"daily_note_dir_pattern"`
}) error {
	// Validate required patterns
	if format.DailyNotePattern == "" {
		return fmt.Errorf("daily_note_pattern is required")
	}
	if format.DailyNoteDirPattern == "" {
		return fmt.Errorf("daily_note_dir_pattern is required")
	}

	// Validate pattern placeholders
	requiredPlaceholders := []string{"{year}", "{month}", "{day}"}
	for _, placeholder := range requiredPlaceholders {
		if !strings.Contains(format.DailyNotePattern, placeholder) {
			return fmt.Errorf("daily_note_pattern must contain %s", placeholder)
		}
	}

	return nil
}

func validateFrontmatter(frontmatter *struct {
	Fields map[string]FrontmatterField `json:"fields"`
}) error {
	for fieldName, field := range frontmatter.Fields {
		if fieldName == "" {
			return fmt.Errorf("frontmatter field name cannot be empty")
		}

		// Validate field type
		validTypes := []string{"string", "enum", "list", "boolean", "number"}
		validType := false
		for _, validT := range validTypes {
			if field.Type == validT {
				validType = true
				break
			}
		}
		if !validType {
			return fmt.Errorf("frontmatter field '%s' has invalid type: %s", fieldName, field.Type)
		}

		// Validate enum values
		if field.Type == "enum" && len(field.Values) == 0 {
			return fmt.Errorf("frontmatter field '%s' is enum but has no values", fieldName)
		}
	}

	return nil
}

// Save saves the configuration to ~/.config/jotr/config.json
func Save(cfg *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "jotr")
	configPath := filepath.Join(configDir, "config.json")

	// Create backup of existing config if it exists
	if _, err := utils.BackupFile(configPath); err != nil {
		return fmt.Errorf("failed to backup existing config: %w", err)
	}

	// Ensure config directory exists
	if err := utils.EnsureDir(configDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file atomically
	if err := utils.AtomicWriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
