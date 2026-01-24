package templatecmd

import (
	"testing"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/templates"
)

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
