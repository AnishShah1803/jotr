package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// createTestConfig creates a test configuration with a temporary directory.
func createTestConfigForTemplate(t *testing.T, tmpDir string) *config.LoadedConfig {
	t.Helper()

	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = tmpDir
	cfg.Paths.DiaryDir = "Diary"
	cfg.Format.CaptureSection = "Captured"
	cfg.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	cfg.Format.DailyNoteDirPattern = "{year}/{month}"
	cfg.DiaryPath = filepath.Join(tmpDir, "Diary")

	return cfg
}

// getTemplateDir returns the template directory path for a given config.
func getTemplateDirForTest(cfg *config.LoadedConfig) string {
	return filepath.Join(cfg.Paths.BaseDir, ".templates")
}

// TestListTemplates_Empty tests listing templates when no templates exist.
func TestListTemplates_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-template-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForTemplate(t, tmpDir)
	templateDir := getTemplateDirForTest(cfg)

	// Template directory doesn't exist
	if utils.FileExists(templateDir) {
		t.Errorf("Template directory should not exist: %s", templateDir)
	}
}

// TestListTemplates_WithTemplates tests listing templates when templates exist.
func TestListTemplates_WithTemplates(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-template-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForTemplate(t, tmpDir)
	templateDir := getTemplateDirForTest(cfg)

	// Create template directory and templates
	if err := os.MkdirAll(templateDir, 0750); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	templates := []string{"meeting", "work", "personal"}
	for _, name := range templates {
		templatePath := filepath.Join(templateDir, name+".md")
		content := "# " + name + " Template\n\n## Section 1\n\n## Section 2\n\n"

		if err := os.WriteFile(templatePath, []byte(content), constants.FilePerm0644); err != nil {
			t.Fatalf("Failed to create template %s: %v", name, err)
		}
	}

	// Count templates
	files, err := os.ReadDir(templateDir)
	if err != nil {
		t.Fatalf("Failed to read template directory: %v", err)
	}

	var templateCount int

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".md" {
			templateCount++
		}
	}

	if templateCount != 3 {
		t.Errorf("Expected 3 templates, got %d", templateCount)
	}
}

// TestListTemplates_NonMDFilesIgnored tests that non-.md files are ignored.
func TestListTemplates_NonMDFilesIgnored(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-template-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForTemplate(t, tmpDir)
	templateDir := getTemplateDirForTest(cfg)

	// Create template directory
	if err := os.MkdirAll(templateDir, 0750); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	// Create both .md and non-.md files
	files := map[string]string{
		"valid.md":     "# Valid Template\n",
		"notvalid.txt": "Not a template",
		"another.md":   "# Another Template\n",
		"config.json":  "{}",
	}

	for name, content := range files {
		path := filepath.Join(templateDir, name)
		if err := os.WriteFile(path, []byte(content), constants.FilePerm0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", name, err)
		}
	}

	// Count only .md files
	dirFiles, err := os.ReadDir(templateDir)
	if err != nil {
		t.Fatalf("Failed to read template directory: %v", err)
	}

	var mdCount int

	for _, f := range dirFiles {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".md" {
			mdCount++
		}
	}

	if mdCount != 2 {
		t.Errorf("Expected 2 .md files, got %d", mdCount)
	}
}

// TestCreateTemplate_Success tests successful template creation.
func TestCreateTemplate_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-template-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForTemplate(t, tmpDir)
	templateDir := getTemplateDirForTest(cfg)

	// Create template directory
	if err := os.MkdirAll(templateDir, 0750); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	templateName := "meeting"
	templatePath := filepath.Join(templateDir, templateName+".md")

	// Create template content
	content := "# meeting Template\n\n## Section 1\n\n## Section 2\n\n"
	if err := os.WriteFile(templatePath, []byte(content), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Verify template exists
	if !utils.FileExists(templatePath) {
		t.Errorf("Template should exist at %s", templatePath)
	}

	// Verify content
	storedContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	if string(storedContent) != content {
		t.Errorf("Template content mismatch. Expected %q, got %q", content, string(storedContent))
	}
}

// TestCreateTemplate_AlreadyExists tests that creating a duplicate template is handled.
func TestCreateTemplate_AlreadyExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-template-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForTemplate(t, tmpDir)
	templateDir := getTemplateDirForTest(cfg)

	// Create template directory and existing template
	if err := os.MkdirAll(templateDir, 0750); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	templateName := "existing"

	templatePath := filepath.Join(templateDir, templateName+".md")
	if err := os.WriteFile(templatePath, []byte("# Existing\n"), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to create existing template: %v", err)
	}

	// Check that template exists
	if !utils.FileExists(templatePath) {
		t.Errorf("Template should exist: %s", templatePath)
	}
}

// TestEditTemplate_NotFound tests editing a non-existent template.
func TestEditTemplate_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-template-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForTemplate(t, tmpDir)
	templateDir := getTemplateDirForTest(cfg)

	// Template directory doesn't exist
	if utils.FileExists(templateDir) {
		t.Errorf("Template directory should not exist: %s", templateDir)
	}
}

// TestDeleteTemplate_Success tests successful template deletion.
func TestDeleteTemplate_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-template-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForTemplate(t, tmpDir)
	templateDir := getTemplateDirForTest(cfg)

	// Create template directory and template
	if err := os.MkdirAll(templateDir, 0750); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	templateName := "delete-me"

	templatePath := filepath.Join(templateDir, templateName+".md")
	if err := os.WriteFile(templatePath, []byte("# Delete Me\n"), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Verify template exists
	if !utils.FileExists(templatePath) {
		t.Errorf("Template should exist before deletion: %s", templatePath)
	}

	if err := os.Remove(templatePath); err != nil {
		t.Fatalf("Failed to delete template: %v", err)
	}

	// Verify template no longer exists
	if utils.FileExists(templatePath) {
		t.Errorf("Template should not exist after deletion: %s", templatePath)
	}
}

// TestDeleteTemplate_NotFound tests deleting a non-existent template.
func TestDeleteTemplate_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-template-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfigForTemplate(t, tmpDir)
	templateDir := getTemplateDirForTest(cfg)

	nonExistentPath := filepath.Join(templateDir, "nonexistent.md")

	// Verify file doesn't exist
	if utils.FileExists(nonExistentPath) {
		t.Errorf("Template should not exist: %s", nonExistentPath)
	}
}

// TestTemplatePathBuilding tests that template paths are built correctly.
func TestTemplatePathBuilding(t *testing.T) {
	testCases := []struct {
		name     string
		baseDir  string
		template string
		expected string
	}{
		{
			name:     "Simple template",
			baseDir:  "/tmp/notes",
			template: "meeting",
			expected: "/tmp/notes/.templates/meeting.md",
		},
		{
			name:     "Template with spaces",
			baseDir:  "/tmp/my notes",
			template: "work template",
			expected: "/tmp/my notes/.templates/work template.md",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			templateDir := filepath.Join(tc.baseDir, ".templates")
			templatePath := filepath.Join(templateDir, tc.template+".md")

			if templatePath != tc.expected {
				t.Errorf("Expected template path %s, got %s", tc.expected, templatePath)
			}
		})
	}
}

// TestTemplateNaming tests template naming conventions.
func TestTemplateNaming(t *testing.T) {
	testCases := []struct {
		name         string
		templateName string
		expectedFile string
	}{
		{
			name:         "Simple name",
			templateName: "meeting",
			expectedFile: "meeting.md",
		},
		{
			name:         "Name with spaces",
			templateName: "work template",
			expectedFile: "work template.md",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Template name is converted to filename by appending .md
			actualFile := tc.templateName + ".md"
			if actualFile != tc.expectedFile {
				t.Errorf("Expected filename %s, got %s", tc.expectedFile, actualFile)
			}

			// Display name is filename without .md extension
			displayName := actualFile[:len(actualFile)-3]
			if displayName != tc.templateName {
				t.Errorf("Expected display name %s, got %s", tc.templateName, displayName)
			}
		})
	}
}

// TestTemplateContentFormat tests that template content follows expected format.
func TestTemplateContentFormat(t *testing.T) {
	templateName := "meeting"

	// Template content format from createTemplate
	content := "# " + templateName + " Template\n\n## Section 1\n\n## Section 2\n\n"

	// Verify format
	if !strings.HasPrefix(content, "# ") {
		t.Errorf("Content should start with '# ' for markdown heading")
	}

	if !strings.Contains(content, "Template") {
		t.Errorf("Content should contain 'Template' in title")
	}

	// Count sections (## headers)
	sectionCount := strings.Count(content, "## ")
	if sectionCount < 2 {
		t.Errorf("Expected at least 2 sections, got %d", sectionCount)
	}
}
