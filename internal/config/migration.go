package config

import (
	"fmt"
	"strconv"
	"strings"
)

type MigrationFunc func(cfg *Config) error

var migrations = []MigrationFunc{
	migrateV0_9_0_toV1_0_0,
}

func migrateV0_9_0_toV1_0_0(cfg *Config) error {
	removeDeprecatedFields(cfg)
	setDefaultValues(cfg)
	normalizeFields(cfg)
	return nil
}

func removeDeprecatedFields(cfg *Config) {
	deleteDeprecatedField(cfg, "use_ai_beta")
	deleteDeprecatedField(cfg, "auto_backup_enabled")
	deleteDeprecatedField(cfg, "legacy_git_sync")
	deleteDeprecatedField(cfg, "old_template_format")
}

func deleteDeprecatedField(cfg *Config, fieldName string) {
	if cfg.NoteTemplates != nil {
		if _, exists := cfg.NoteTemplates[fieldName]; exists {
			delete(cfg.NoteTemplates, fieldName)
		}
	}
}

func setDefaultValues(cfg *Config) {
	if cfg.Version == "" {
		cfg.Version = ConfigVersion
	}

	if cfg.Format.TaskSection == "" {
		cfg.Format.TaskSection = "Tasks"
	}

	if cfg.Format.CaptureSection == "" {
		cfg.Format.CaptureSection = "Captured"
	}

	if len(cfg.Format.DailyNoteSections) == 0 {
		cfg.Format.DailyNoteSections = []string{"Notes", "Meetings"}
	}

	if cfg.Format.DailyNotePattern == "" {
		cfg.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	}

	if cfg.Format.DailyNoteDirPattern == "" {
		cfg.Format.DailyNoteDirPattern = "{year}/{month_num}-{month_abbr}"
	}

	// No longer setting default editor - leave empty if not configured

	if cfg.Streaks.IncludeWeekends {
		cfg.Streaks.IncludeWeekends = true
	}

	if len(cfg.Summary.Sources) == 0 {
		cfg.Summary.Sources = []string{"todo", "daily_notes"}
	}

	if cfg.Git.AutoCommitFrequency == "" {
		cfg.Git.AutoCommitFrequency = "manual"
	}

	if cfg.Frontmatter.Fields == nil {
		cfg.Frontmatter.Fields = map[string]FrontmatterField{}
	}
}

func normalizeFields(cfg *Config) {
	if cfg.NoteTemplates == nil {
		cfg.NoteTemplates = make(map[string]interface{})
	}

	if cfg.AI.Command == "" && cfg.AI.Enabled {
		cfg.AI.Enabled = false
	}

	cfg.Git.Enabled = false
}

// ParseVersion parses a semantic version string into its components.
func ParseVersion(v string) (major, minor, patch int, err error) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", v)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return major, minor, patch, nil
}

// RunMigrations runs any necessary configuration migrations.
func RunMigrations(cfg *Config) error {
	configVersion := cfg.Version
	if configVersion == "" {
		configVersion = "0.0.0"
	}

	currentMajor, currentMinor, currentPatch, err := ParseVersion(ConfigVersion)
	if err != nil {
		return fmt.Errorf("invalid current config version: %w", err)
	}

	loadedMajor, loadedMinor, loadedPatch, err := ParseVersion(configVersion)
	if err != nil {
		return fmt.Errorf("invalid loaded config version %q: %w", configVersion, err)
	}

	for i, migration := range migrations {
		migrationMajor := loadedMajor
		migrationMinor := loadedMinor
		migrationPatch := loadedPatch

		if migrationMajor < currentMajor ||
			(migrationMajor == currentMajor && migrationMinor < currentMinor) ||
			(migrationMajor == currentMajor && migrationMinor == currentMinor && migrationPatch < currentPatch) {
			if err := migration(cfg); err != nil {
				return fmt.Errorf("migration %d failed: %w", i+1, err)
			}
		}
	}

	if cfg.Version != ConfigVersion {
		cfg.Version = ConfigVersion
	}

	return nil
}
