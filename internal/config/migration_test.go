package config

import (
	"encoding/json"
	"testing"
)

func TestMigrateV0_9_0_toV1_0_0_RemovesDeprecatedFields(t *testing.T) {
	cfg := &Config{}
	cfg.Version = "0.9.0"
	cfg.NoteTemplates = map[string]interface{}{
		"use_ai_beta":         true,
		"auto_backup_enabled": false,
		"legacy_git_sync":     true,
		"old_template_format": "v1",
		"valid_template":      "keep this",
	}

	err := migrateV0_9_0_toV1_0_0(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if _, exists := cfg.NoteTemplates["use_ai_beta"]; exists {
		t.Error("Expected use_ai_beta to be removed")
	}
	if _, exists := cfg.NoteTemplates["auto_backup_enabled"]; exists {
		t.Error("Expected auto_backup_enabled to be removed")
	}
	if _, exists := cfg.NoteTemplates["legacy_git_sync"]; exists {
		t.Error("Expected legacy_git_sync to be removed")
	}
	if _, exists := cfg.NoteTemplates["old_template_format"]; exists {
		t.Error("Expected old_template_format to be removed")
	}
	if _, exists := cfg.NoteTemplates["valid_template"]; !exists {
		t.Error("Expected valid_template to be preserved")
	}
}

func TestMigrateV0_9_0_toV1_0_0_SetsDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.Version = "0.9.0"

	err := migrateV0_9_0_toV1_0_0(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Format.TaskSection != "Tasks" {
		t.Errorf("Expected default task_section 'Tasks', got '%s'", cfg.Format.TaskSection)
	}
	if cfg.Format.CaptureSection != "Captured" {
		t.Errorf("Expected default capture_section 'Captured', got '%s'", cfg.Format.CaptureSection)
	}
	if len(cfg.Format.DailyNoteSections) != 2 {
		t.Errorf("Expected 2 daily note sections, got %d", len(cfg.Format.DailyNoteSections))
	}
	if cfg.Format.DailyNotePattern != "{year}-{month}-{day}-{weekday}" {
		t.Errorf("Expected default daily_note_pattern, got '%s'", cfg.Format.DailyNotePattern)
	}
	if cfg.Format.DailyNoteDirPattern != "{year}/{month_num}-{month_abbr}" {
		t.Errorf("Expected default daily_note_dir_pattern, got '%s'", cfg.Format.DailyNoteDirPattern)
	}
	if cfg.Editor.Default != "" {
		t.Errorf("Expected empty editor default, got '%s'", cfg.Editor.Default)
	}
	if len(cfg.Summary.Sources) != 2 {
		t.Errorf("Expected 2 summary sources, got %d", len(cfg.Summary.Sources))
	}
	if cfg.Git.AutoCommitFrequency != "manual" {
		t.Errorf("Expected default git auto_commit_frequency 'manual', got '%s'", cfg.Git.AutoCommitFrequency)
	}
}

func TestMigrateV0_9_0_toV1_0_0_NormalizesFields(t *testing.T) {
	cfg := &Config{}
	cfg.Version = "0.9.0"
	cfg.AI.Enabled = true
	cfg.AI.Command = ""

	err := migrateV0_9_0_toV1_0_0(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.AI.Enabled {
		t.Error("Expected AI.Enabled to be false when command is empty")
	}
}

func TestMigrateV0_9_0_toV1_0_0_PreservesExistingValues(t *testing.T) {
	cfg := &Config{}
	cfg.Version = "0.9.0"
	cfg.Paths.BaseDir = "/custom/path"
	cfg.Paths.DiaryDir = "MyDiary"
	cfg.Format.TaskSection = "CustomTasks"
	cfg.NoteTemplates = map[string]interface{}{
		"custom": "value",
	}

	err := migrateV0_9_0_toV1_0_0(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Paths.BaseDir != "/custom/path" {
		t.Errorf("Expected base_dir to be preserved, got '%s'", cfg.Paths.BaseDir)
	}
	if cfg.Paths.DiaryDir != "MyDiary" {
		t.Errorf("Expected diary_dir to be preserved, got '%s'", cfg.Paths.DiaryDir)
	}
	if cfg.Format.TaskSection != "CustomTasks" {
		t.Errorf("Expected task_section to be preserved, got '%s'", cfg.Format.TaskSection)
	}
	if cfg.NoteTemplates["custom"] != "value" {
		t.Error("Expected custom template to be preserved")
	}
}

func TestRunMigrations_WithEmptyVersion(t *testing.T) {
	cfg := &Config{}
	cfg.Version = ""

	err := RunMigrations(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Version != ConfigVersion {
		t.Errorf("Expected version %s after migration, got %s", ConfigVersion, cfg.Version)
	}
}

func TestRunMigrations_WithOldVersion(t *testing.T) {
	cfg := &Config{}
	cfg.Version = "0.8.0"

	err := RunMigrations(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Version != ConfigVersion {
		t.Errorf("Expected version %s after migration, got %s", ConfigVersion, cfg.Version)
	}
}

func TestRunMigrations_WithCurrentVersion(t *testing.T) {
	cfg := &Config{}
	cfg.Version = ConfigVersion

	err := RunMigrations(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Version != ConfigVersion {
		t.Errorf("Expected version %s, got %s", ConfigVersion, cfg.Version)
	}
}

func TestDeleteDeprecatedField(t *testing.T) {
	cfg := &Config{}
	cfg.NoteTemplates = map[string]interface{}{
		"old_field":  "value",
		"good_field": "keep",
	}

	deleteDeprecatedField(cfg, "old_field")

	if _, exists := cfg.NoteTemplates["old_field"]; exists {
		t.Error("Expected old_field to be deleted")
	}
	if _, exists := cfg.NoteTemplates["good_field"]; !exists {
		t.Error("Expected good_field to remain")
	}
}

func TestDeleteDeprecatedField_NilTemplates(t *testing.T) {
	cfg := &Config{}

	deleteDeprecatedField(cfg, "some_field")
}

func TestSetDefaultValues_NilConfig(t *testing.T) {
	cfg := &Config{}

	setDefaultValues(cfg)

	if cfg.Frontmatter.Fields == nil {
		t.Error("Expected Frontmatter.Fields to be initialized")
	}
}

func TestNormalizeFields_NilTemplates(t *testing.T) {
	cfg := &Config{}
	cfg.AI.Enabled = true
	cfg.AI.Command = ""

	normalizeFields(cfg)

	if cfg.NoteTemplates == nil {
		t.Error("Expected NoteTemplates to be initialized")
	}
}

func TestMigrateConfig_CompleteFlow(t *testing.T) {
	oldConfig := `{
		"version": "0.7.0",
		"paths": {
			"base_dir": "/test/path",
			"diary_dir": "Diary",
			"todo_file_path": "todo"
		},
		"format": {
			"task_section": "",
			"capture_section": "",
			"daily_note_pattern": "",
			"daily_note_dir_pattern": "",
			"daily_note_sections": []
		},
		"note_templates": {
			"use_ai_beta": true,
			"old_format": true
		}
	}`

	var cfg Config
	err := json.Unmarshal([]byte(oldConfig), &cfg)
	if err != nil {
		t.Fatalf("Failed to unmarshal old config: %v", err)
	}

	err = RunMigrations(&cfg)
	if err != nil {
		t.Fatalf("Expected no error during migration, got: %v", err)
	}

	if cfg.Version != ConfigVersion {
		t.Errorf("Expected migrated config version to be %s, got %s", ConfigVersion, cfg.Version)
	}

	if cfg.Format.TaskSection != "Tasks" {
		t.Errorf("Expected default TaskSection after migration, got '%s'", cfg.Format.TaskSection)
	}

	if _, exists := cfg.NoteTemplates["use_ai_beta"]; exists {
		t.Error("Expected deprecated field to be removed after migration")
	}
}
