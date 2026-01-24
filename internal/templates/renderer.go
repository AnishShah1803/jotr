package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
)

func ResolveBuiltIns(tmpl *Template, cfg *config.Config) {
	now := time.Now()

	tmpl.BuiltIns = map[string]string{
		"{$date}":     now.Format("2006-01-02"),
		"{$datetime}": now.Format("2006-01-02 15:04"),
		"{$base_dir}": cfg.Paths.BaseDir,
		"{$weekday}":  now.Format("Monday"),
		"{$time}":     now.Format("15:04"),
	}
}

func SubstituteVariables(content string, vars, builtins map[string]string) string {
	result := content

	for name, value := range vars {
		placeholder := "{$" + name + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	for name, value := range builtins {
		result = strings.ReplaceAll(result, name, value)
	}

	return result
}

func SubstitutePrompts(content string, values []string) string {
	result := content
	promptRegex := regexp.MustCompile(`<prompt>.*?</prompt>`)

	matches := promptRegex.FindAllStringIndex(content, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		if i < len(values) {
			match := matches[i]
			start := match[0]
			end := match[1]
			result = result[:start] + values[i] + result[end:]
		}
	}

	return result
}

func RenderTargetPath(tmpl *Template, vars map[string]string, cfg *config.Config) (string, error) {
	if tmpl.TargetPath == "" {
		return "", fmt.Errorf("template has no target path")
	}

	ResolveBuiltIns(tmpl, cfg)

	path := SubstituteVariables(tmpl.TargetPath, vars, tmpl.BuiltIns)

	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}

	return path, nil
}

func RenderTemplate(tmpl *Template, userVars map[string]string, promptValues []string, cfg *config.Config) string {
	ResolveBuiltIns(tmpl, cfg)

	content := SubstituteVariables(tmpl.Content, userVars, tmpl.BuiltIns)

	content = SubstitutePrompts(content, promptValues)

	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "<!-- Path:") {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}
