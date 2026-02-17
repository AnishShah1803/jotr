package templates

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/utils"
)

func TestCompleteTemplateWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)

	templatesDir := filepath.Join(tempDir, "templates")
	baseDir := filepath.Join(tempDir, "notes")

	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	templateContent := `<!-- Path: {$base_dir}/Meetings/{$date}-{$topic}.md -->

# Meeting Notes: {$topic}

**Date:** {$date}
**Time:** {$time}

## Attendees
<prompt>Who attended this meeting?</prompt>

## Agenda
<prompt>What are the agenda items?</prompt>
`

	templatePath := filepath.Join(templatesDir, "1-1-meeting.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	templateList, errs := DiscoverTemplates(templatesDir)
	if len(errs) > 0 {
		t.Fatalf("Template discovery errors: %v", errs)
	}

	if len(templateList) != 1 {
		t.Fatalf("Expected 1 template, got %d", len(templateList))
	}

	tmpl := templateList[0]

	if tmpl.Name != "meeting" {
		t.Errorf("Expected name 'meeting', got '%s'", tmpl.Name)
	}

	if tmpl.Category != "1" {
		t.Errorf("Expected category '1', got '%s'", tmpl.Category)
	}

	if tmpl.Priority != 1 {
		t.Errorf("Expected priority 1, got %d", tmpl.Priority)
	}

	if len(tmpl.Variables) != 0 {
		t.Errorf("Expected 0 variables, got %d", len(tmpl.Variables))
	}

	if len(tmpl.Prompts) != 2 {
		t.Errorf("Expected 2 prompts, got %d", len(tmpl.Prompts))
	}

	cfg := &config.Config{
		Paths: struct {
			BaseDir      string `json:"base_dir"`
			DiaryDir     string `json:"diary_dir"`
			TodoFilePath string `json:"todo_file_path"`
			PDPFilePath  string `json:"pdp_file_path"`
		}{
			BaseDir: baseDir,
		},
	}

	ResolveBuiltIns(tmpl, cfg)

	if _, ok := tmpl.BuiltIns["{$date}"]; !ok {
		t.Error("Built-in variable {$date} not set")
	}

	userVars := map[string]string{
		"topic": "Project Planning",
	}

	renderedContent := RenderTemplate(tmpl, userVars, []string{"Alice, Bob", "Review timeline"}, cfg)

	if !contains(renderedContent, "Meeting Notes: Project Planning") {
		t.Error("Template content not rendered correctly")
	}

	targetPath, err := RenderTargetPath(tmpl, userVars, cfg)
	if err != nil {
		t.Fatalf("RenderTargetPath failed: %v", err)
	}

	if !contains(targetPath, "Meetings") {
		t.Error("Target path does not contain 'Meetings'")
	}

	if !contains(targetPath, "Project Planning") {
		t.Error("Target path does not contain topic")
	}

	ctx := context.Background()

	if err := CreateFromTemplate(ctx, tmpl, renderedContent, targetPath); err != nil {
		t.Fatalf("CreateFromTemplate failed: %v", err)
	}

	if !utils.FileExists(targetPath) {
		t.Fatal("File was not created")
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	fileContent := string(content)

	if !contains(fileContent, "Meeting Notes: Project Planning") {
		t.Error("File content does not contain expected text")
	}

	if contains(fileContent, "<!-- Path:") {
		t.Error("Path comment was not removed from rendered content")
	}

	if contains(fileContent, "<prompt>") {
		t.Error("Prompt tags were not replaced")
	}

	os.Remove(targetPath)

	if err := CreateFromTemplate(ctx, tmpl, renderedContent, targetPath); err == nil {
		t.Log("File creation succeeded (no error for existing file)")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
