package templates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AnishShah1803/jotr/internal/config"
)

func TestSortTemplates(t *testing.T) {
	templates := []*Template{
		{Filename: "3-2-c.md", Priority: 3, Category: "2", Name: "c"},
		{Filename: "1-1-a.md", Priority: 1, Category: "1", Name: "a"},
		{Filename: "2-1-b.md", Priority: 2, Category: "1", Name: "b"},
		{Filename: "1-2-d.md", Priority: 1, Category: "2", Name: "d"},
		{Filename: "2-2-e.md", Priority: 2, Category: "2", Name: "e"},
	}

	sorted := SortTemplates(templates)

	expectedOrder := []string{
		"1-1-a.md",
		"1-2-d.md",
		"2-1-b.md",
		"2-2-e.md",
		"3-2-c.md",
	}

	if len(sorted) != len(expectedOrder) {
		t.Fatalf("SortTemplates() got %d templates, want %d", len(sorted), len(expectedOrder))
	}

	for i, expected := range expectedOrder {
		if sorted[i].Filename != expected {
			t.Errorf("SortTemplates()[%d] = %s, want %s", i, sorted[i].Filename, expected)
		}
	}
}

func TestDiscoverTemplates(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		templates, errs := DiscoverTemplates(tmpDir)

		if len(templates) != 0 {
			t.Errorf("DiscoverTemplates() got %d templates, want 0", len(templates))
		}
		if len(errs) != 0 {
			t.Errorf("DiscoverTemplates() got %d errors, want 0", len(errs))
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		templates, errs := DiscoverTemplates("/non/existent/path")

		if len(templates) != 0 {
			t.Errorf("DiscoverTemplates() got %d templates, want 0", len(templates))
		}
		if len(errs) != 0 {
			t.Errorf("DiscoverTemplates() got %d errors, want 0", len(errs))
		}
	})

	t.Run("valid templates only", func(t *testing.T) {
		tmpDir := t.TempDir()

		os.WriteFile(filepath.Join(tmpDir, "1-1-meeting.md"), []byte(`# Meeting`), 0644)
		os.WriteFile(filepath.Join(tmpDir, "2-2-journal.md"), []byte(`# Journal`), 0644)

		templates, errs := DiscoverTemplates(tmpDir)

		if len(errs) != 0 {
			t.Errorf("DiscoverTemplates() got %d errors, want 0: %v", len(errs), errs)
		}
		if len(templates) != 2 {
			t.Errorf("DiscoverTemplates() got %d templates, want 2", len(templates))
		}
	})

	t.Run("mixed valid and invalid templates", func(t *testing.T) {
		tmpDir := t.TempDir()

		os.WriteFile(filepath.Join(tmpDir, "1-1-valid.md"), []byte(`# Valid`), 0644)
		os.WriteFile(filepath.Join(tmpDir, "invalid.md"), []byte(`# Invalid`), 0644)
		os.WriteFile(filepath.Join(tmpDir, "2-2-another.md"), []byte(`# Another`), 0644)

		templates, errs := DiscoverTemplates(tmpDir)

		if len(templates) != 2 {
			t.Errorf("DiscoverTemplates() got %d templates, want 2", len(templates))
		}
		if len(errs) != 1 {
			t.Errorf("DiscoverTemplates() got %d errors, want 1", len(errs))
		}
	})

	t.Run("ignores non-md files", func(t *testing.T) {
		tmpDir := t.TempDir()

		os.WriteFile(filepath.Join(tmpDir, "1-1-test.md"), []byte(`# Test`), 0644)
		os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte(`text`), 0644)

		templates, errs := DiscoverTemplates(tmpDir)

		if len(errs) != 0 {
			t.Errorf("DiscoverTemplates() got %d errors, want 0", len(errs))
		}
		if len(templates) != 1 {
			t.Errorf("DiscoverTemplates() got %d templates, want 1", len(templates))
		}
	})

	t.Run("ignores subdirectories", func(t *testing.T) {
		tmpDir := t.TempDir()

		os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
		os.WriteFile(filepath.Join(tmpDir, "1-1-test.md"), []byte(`# Test`), 0644)

		templates, errs := DiscoverTemplates(tmpDir)

		if len(errs) != 0 {
			t.Errorf("DiscoverTemplates() got %d errors, want 0", len(errs))
		}
		if len(templates) != 1 {
			t.Errorf("DiscoverTemplates() got %d templates, want 1", len(templates))
		}
	})
}

func TestLoadTemplates(t *testing.T) {
	t.Run("valid templates only", func(t *testing.T) {
		tmpDir := t.TempDir()

		cfg := &config.LoadedConfig{}
		cfg.TemplatesPath = tmpDir

		os.WriteFile(filepath.Join(tmpDir, "1-1-first.md"), []byte(`# First`), 0644)
		os.WriteFile(filepath.Join(tmpDir, "2-2-second.md"), []byte(`# Second`), 0644)

		templates, warnings := LoadTemplates(cfg)

		if len(warnings) != 0 {
			t.Errorf("LoadTemplates() got %d warnings, want 0: %v", len(warnings), warnings)
		}
		if len(templates) != 2 {
			t.Errorf("LoadTemplates() got %d templates, want 2", len(templates))
		}
		if templates[0].Filename != "1-1-first.md" {
			t.Errorf("LoadTemplates()[0] = %s, want 1-1-first.md (should be sorted)", templates[0].Filename)
		}
	})

	t.Run("with warnings for invalid templates", func(t *testing.T) {
		tmpDir := t.TempDir()

		cfg := &config.LoadedConfig{}
		cfg.TemplatesPath = tmpDir

		os.WriteFile(filepath.Join(tmpDir, "1-1-valid.md"), []byte(`# Valid`), 0644)
		os.WriteFile(filepath.Join(tmpDir, "invalid.md"), []byte(`# Invalid`), 0644)

		templates, warnings := LoadTemplates(cfg)

		if len(templates) != 1 {
			t.Errorf("LoadTemplates() got %d templates, want 1", len(templates))
		}
		if len(warnings) != 1 {
			t.Errorf("LoadTemplates() got %d warnings, want 1", len(warnings))
		}
	})
}
