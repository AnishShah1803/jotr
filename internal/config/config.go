package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// ValidationWarning represents a warning issued during configuration validation.
type ValidationWarning struct {
	Category string
	Message  string
}

// ValidateConfig validates the configuration for common issues.
func ValidateConfig(cfg *Config) ([]ValidationWarning, error) {
	var warnings []ValidationWarning
	var err error

	// Validate required fields
	if cfg.Paths.BaseDir == "" {
		return nil, fmt.Errorf("base_dir is required in config")
	}

	// Validate base directory exists or is creatable
	if err := validateDirectory(cfg.Paths.BaseDir); err != nil {
		return nil, fmt.Errorf("base_dir validation failed: %w", err)
	}

	// Validate diary directory path
	if cfg.Paths.DiaryDir == "" {
		return nil, fmt.Errorf("diary_dir is required in config")
	}

	// Validate todo file path
	if cfg.Paths.TodoFilePath == "" {
		return nil, fmt.Errorf("todo_file_path is required in config")
	}

	// Validate format settings
	if warnings, err = validateFormat(&cfg.Format, warnings); err != nil {
		return nil, fmt.Errorf("format validation failed: %w", err)
	}

	// Validate AI settings if enabled
	if cfg.AI.Enabled && cfg.AI.Command == "" {
		return nil, fmt.Errorf("AI is enabled but no command is configured")
	}

	// Validate editor configuration
	if warnings, err = validateEditor(&cfg.Editor, warnings); err != nil {
		return nil, fmt.Errorf("editor validation failed: %w", err)
	}

	// Validate frontmatter fields
	if err := validateFrontmatter(&cfg.Frontmatter); err != nil {
		return nil, fmt.Errorf("frontmatter validation failed: %w", err)
	}

	// Check for deprecated fields
	warnings = checkDeprecatedFields(cfg, warnings)

	// Validate paths are absolute or convertable
	warnings = validatePathAbsolutes(cfg, warnings)

	return warnings, nil
}

type configContextKey struct{}

var configKey = &configContextKey{}

func WithConfig(ctx context.Context, cfg interface{}) context.Context {
	return context.WithValue(ctx, configKey, cfg)
}

func GetConfigFromContext(ctx context.Context) (interface{}, bool) {
	cfg, ok := ctx.Value(configKey).(interface{})
	return cfg, ok
}

// PathsConfig holds path-related configuration settings.
type PathsConfig struct {
	BaseDir      string `json:"base_dir"`
	DiaryDir     string `json:"diary_dir"`
	TodoFilePath string `json:"todo_file_path"`
	PDPFilePath  string `json:"pdp_file_path"`
}

// FormatConfig holds formatting-related configuration settings.
type FormatConfig struct {
	TaskSection         string   `json:"task_section"`
	CaptureSection      string   `json:"capture_section"`
	DailyNotePattern    string   `json:"daily_note_pattern"`
	DailyNoteDirPattern string   `json:"daily_note_dir_pattern"`
	DailyNoteSections   []string `json:"daily_note_sections"`
}

// AIConfig holds AI-related configuration settings.
type AIConfig struct {
	Command string `json:"command"`
	Enabled bool   `json:"enabled"`
}

// StreaksConfig holds streak-related configuration settings.
type StreaksConfig struct {
	IncludeWeekends bool `json:"include_weekends"`
}

// DailyNoteTemplateConfig holds daily note template configuration.
type DailyNoteTemplateConfig struct {
	Sections        []TemplateSection `json:"sections"`
	Prompts         TemplatePrompts   `json:"prompts"`
	WeekdaySections []TemplateSection `json:"weekday_sections"`
	WeekendSections []TemplateSection `json:"weekend_sections"`
}

// GitConfig holds git-related configuration settings.
type GitConfig struct {
	AutoCommitFrequency string `json:"auto_commit_frequency"`
	Enabled             bool   `json:"enabled"`
	AutoCommit          bool   `json:"auto_commit"`
}

// SummaryConfig holds summary-related configuration settings.
type SummaryConfig struct {
	Sources                   []string `json:"sources"`
	DailyNotesExcludeSections []string `json:"daily_notes_exclude_sections"`
	IncludeNoteTypes          []string `json:"include_note_types"`
}

// FrontmatterConfig holds frontmatter-related configuration settings.
type FrontmatterConfig struct {
	Fields map[string]FrontmatterField `json:"fields"`
}

const ConfigVersion = "1.0.0"

type Config struct {
	Version           string                 `json:"version"`
	Frontmatter       FrontmatterConfig      `json:"frontmatter"`
	NoteTemplates     map[string]interface{} `json:"note_templates"`
	Format            FormatConfig           `json:"format"`
	Paths             PathsConfig            `json:"paths"`
	AI                AIConfig               `json:"ai"`
	Git               GitConfig              `json:"git"`
	Editor            `json:"editor"`
	DailyNoteTemplate DailyNoteTemplateConfig `json:"daily_note_template"`
	Summary           SummaryConfig           `json:"summary"`
	Streaks           StreaksConfig           `json:"streaks"`
}

// TemplateSection represents a section in a template.
type TemplateSection struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// TemplatePrompts represents prompts for different times of day.
type TemplatePrompts struct {
	Morning []TemplatePrompt `json:"morning"`
	Evening []TemplatePrompt `json:"evening"`
}

// TemplatePrompt represents a single prompt.
type TemplatePrompt struct {
	Section  string `json:"section"`
	Question string `json:"question"`
}

// FrontmatterField represents a frontmatter field definition.
type FrontmatterField struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Values      []string `json:"values,omitempty"`
	Required    bool     `json:"required"`
}

// Editor holds editor configuration settings.
type Editor struct {
	Default string `json:"default"`
}

func (e *Editor) GetDefaultEditor() string {
	if e == nil || e.Default == "" {
		return ""
	}

	return e.Default
}

// LoadedConfig contains the loaded configuration with computed paths.
type LoadedConfig struct {
	Config
	DiaryPath     string
	TodoPath      string
	StatePath     string
	PDPPath       string
	TemplatesPath string
}

// Load reads and parses the jotr configuration file.
// It searches for the config file in the following order:
//  1. JOTR_CONFIG environment variable (if set)
//  2. XDG_CONFIG_HOME/jotr/config.json (if XDG_CONFIG_HOME is set)
//  3. ~/.config/jotr/config.json (default location)
//
// Returns a LoadedConfig with resolved paths and defaults applied.
// Use LoadWithContext for cancellation support.
func Load() (*LoadedConfig, error) {
	return LoadWithContext(context.Background(), "")
}

func LoadWithContext(ctx context.Context, configPathOverride string) (*LoadedConfig, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var configPath string
	if configPathOverride != "" {
		configPath = configPathOverride
		utils.VerboseLogWithContext(ctx, "Using config from override: %s", configPath)
	} else if cfgFromCtx, ok := GetConfigFromContext(ctx); ok && cfgFromCtx != "" {
		// Config path from context (string)
		configPath = cfgFromCtx.(string)
		utils.VerboseLogWithContext(ctx, "Using config from context: %s", configPath)
	} else if envConfig := os.Getenv("JOTR_CONFIG"); envConfig != "" {
		configPath = envConfig
		utils.VerboseLogWithContext(ctx, "Using config from JOTR_CONFIG env: %s", configPath)
	} else {
		// Use dev config in dev builds
		if _, err := os.Stat("dev-config.json"); err == nil {
			configPath = "dev-config.json"
		} else {
			// Use standard default path for production builds
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}

			configPath = filepath.Join(homeDir, ".config", "jotr", "config.json")
		}
		utils.VerboseLogWithContext(ctx, "Using config path: %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s\nRun 'jotr configure' to create it", configPath)
		}

		utils.VerboseLogErrorWithContext(ctx, "reading config file", err)

		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		utils.VerboseLogErrorWithContext(ctx, "parsing config JSON", err)
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	utils.VerboseLogWithContext(ctx, "Config loaded successfully, base_dir: %s", cfg.Paths.BaseDir)

	if err := RunMigrations(&cfg); err != nil {
		return nil, fmt.Errorf("config migration failed: %w", err)
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

	loaded.TemplatesPath = filepath.Join(cfg.Paths.BaseDir, "templates")

	// Set default editor if not configured

	// Validate the loaded configuration
	if _, err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return loaded, nil
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

func validateFormat(format *FormatConfig, warnings []ValidationWarning) ([]ValidationWarning, error) {
	// Validate required patterns
	if format.DailyNotePattern == "" {
		return nil, fmt.Errorf("daily_note_pattern is required")
	}

	if format.DailyNoteDirPattern == "" {
		return nil, fmt.Errorf("daily_note_dir_pattern is required")
	}

	// Validate pattern placeholders
	requiredPlaceholders := []string{"{year}", "{month}", "{day}"}
	for _, placeholder := range requiredPlaceholders {
		if !strings.Contains(format.DailyNotePattern, placeholder) {
			return nil, fmt.Errorf("daily_note_pattern must contain %s", placeholder)
		}
	}

	// Validate all placeholders are valid (detect typos)
	validPlaceholders := map[string]bool{
		"{year}": true, "{month}": true, "{month_num}": true,
		"{month_abbr}": true, "{day}": true, "{weekday}": true,
	}
	placeholderPattern := `{[a-z_]+}`
	re := regexp.MustCompile(placeholderPattern)
	matches := re.FindAllString(format.DailyNotePattern, -1)
	for _, match := range matches {
		if !validPlaceholders[match] {
			warnings = append(warnings, ValidationWarning{
				Category: "format",
				Message:  fmt.Sprintf("unknown placeholder '%s' in daily_note_pattern (valid: year, month, month_num, month_abbr, day, weekday)", match),
			})
		}
	}

	return warnings, nil
}

func validateFrontmatter(frontmatter *FrontmatterConfig) error {
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

func validateEditor(editor *Editor, warnings []ValidationWarning) ([]ValidationWarning, error) {
	if editor.Default != "" {
		if err := utils.ValidateEditor(editor.Default); err != nil {
			return nil, fmt.Errorf("invalid editor configuration: %w", err)
		}
	}

	return warnings, nil
}

func checkDeprecatedFields(cfg *Config, warnings []ValidationWarning) []ValidationWarning {
	deprecatedFieldWarnings := map[string]string{
		"use_ai_beta":         "use_ai_beta is deprecated, use ai.enabled instead",
		"auto_backup_enabled": "auto_backup_enabled is no longer supported",
		"legacy_git_sync":     "legacy_git_sync has been replaced by git configuration",
		"old_template_format": "old_template_format is deprecated, use the new template system",
	}

	for fieldName, message := range deprecatedFieldWarnings {
		warnings = append(warnings, ValidationWarning{
			Category: "deprecated",
			Message:  fmt.Sprintf("deprecated field '%s': %s", fieldName, message),
		})
	}

	return warnings
}

func validatePathAbsolutes(cfg *Config, warnings []ValidationWarning) []ValidationWarning {
	if !filepath.IsAbs(cfg.Paths.BaseDir) {
		warnings = append(warnings, ValidationWarning{
			Category: "paths",
			Message:  "base_dir is not an absolute path - relative paths may cause issues",
		})
	}

	return warnings
}

// Save saves the configuration to ~/.config/jotr/config.json.
func Save(cfg *Config) error {
	if cfg.Version != ConfigVersion {
		cfg.Version = ConfigVersion
	}

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
	if err := utils.AtomicWriteFile(configPath, data, constants.FilePerm0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetEditor returns the editor to use for opening files.
// It checks the following sources in order:
//  1. EDITOR environment variable (if set)
//  2. Configured editor in the config file

// The configured editor can be set using "jotr configure" or by editing
// the config file directly.
func GetEditor() string {
	return GetEditorWithContext(context.Background())
}

// GetEditorWithContext returns the editor to use for opening files, with context support.
// It checks the following sources in order:
//  1. EDITOR environment variable (if set)
//  2. Configured editor in the config file

func GetEditorWithContext(ctx context.Context) string {
	// First check EDITOR environment variable
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// Fall back to configured default
	cfg, err := LoadWithContext(ctx, "")
	if err != nil {
		return ""
	}

	return cfg.GetDefaultEditor()
}

// IsEditorConfigured checks if an editor is configured.
// Returns true if EDITOR env var is set or editor.default is configured.
func IsEditorConfigured() bool {
	return IsEditorConfiguredWithContext(context.Background())
}

// IsEditorConfiguredWithContext checks if an editor is configured, with context support.
func IsEditorConfiguredWithContext(ctx context.Context) bool {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return true
	}

	cfg, err := LoadWithContext(ctx, "")
	if err != nil {
		return false
	}

	return cfg.Editor.Default != ""
}
