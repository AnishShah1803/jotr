package templatecmd

import (
	"testing"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/templates"
)

func TestLoadTemplatesForIntegration(t *testing.T) {
	templates := []*templates.Template{
		{
			Name:     "test1",
			Category: "cat1",
		},
		{
			Name:     "test2",
			Category: "cat1",
		},
	}

	selected := SelectTemplateForIntegration(templates)

	if selected == nil {
		t.Fatal("SelectTemplateForIntegration returned nil")
	}

	if selected.Name != "test1" && selected.Name != "test2" {
		t.Errorf("Unexpected selection: %s", selected.Name)
	}
}

func TestCreateFromSelectedTemplate(t *testing.T) {
	tmpl := &templates.Template{
		Name:       "test",
		Category:   "cat",
		Priority:   1,
		Content:    "# Test",
		TargetPath: "/tmp/test.md",
		Variables:  []templates.Variable{},
		Prompts:    []templates.Prompt{},
	}

	cfg := &config.LoadedConfig{}

	if err := CreateFromSelectedTemplate(nil, cfg, tmpl); err == nil {
		t.Error("Expected error with nil context, got nil")
	}
}
