package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		Sections         []TemplateSection `json:"sections"`
		Prompts          TemplatePrompts   `json:"prompts"`
		WeekdaySections  []TemplateSection `json:"weekday_sections"`
		WeekendSections  []TemplateSection `json:"weekend_sections"`
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
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s\nRun 'jotr configure' to create it", configPath)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

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

	return loaded, nil
}

// Save saves the configuration to ~/.config/jotr/config.json
func Save(cfg *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "jotr")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

