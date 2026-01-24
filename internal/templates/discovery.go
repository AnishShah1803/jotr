package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AnishShah1803/jotr/internal/config"
)

func GetTemplatesDir(cfg *config.LoadedConfig) string {
	return cfg.TemplatesPath
}

func SortTemplates(templates []*Template) []*Template {
	sorted := make([]*Template, len(templates))
	copy(sorted, templates)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority < sorted[j].Priority
		}
		if sorted[i].Category != sorted[j].Category {
			return sorted[i].Category < sorted[j].Category
		}
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}

func DiscoverTemplates(templatesDir string) ([]*Template, []error) {
	var templates []*Template
	var errors []error

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return templates, nil
		}
		return nil, []error{err}
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filepath := filepath.Join(templatesDir, entry.Name())
		content, err := os.ReadFile(filepath)
		if err != nil {
			errors = append(errors, fmt.Errorf("reading %s: %w", entry.Name(), err))
			continue
		}

		template, err := ParseTemplate(filepath, string(content))
		if err != nil {
			errors = append(errors, err)
			continue
		}

		templates = append(templates, template)
	}

	return templates, errors
}

func LoadTemplates(cfg *config.LoadedConfig) ([]*Template, []string) {
	templates, errs := DiscoverTemplates(cfg.TemplatesPath)

	var warnings []string
	for _, err := range errs {
		warnings = append(warnings, err.Error())
	}

	return SortTemplates(templates), warnings
}
