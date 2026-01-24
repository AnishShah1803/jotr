package templates

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var variableRegex = regexp.MustCompile(`^(\w+)\s*=\s*<prompt>(.*?)</prompt>`)
var promptRegex = regexp.MustCompile(`<prompt>(.*?)</prompt>`)

func ParsePathComment(content string) (string, error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "<!-- Path:") {
			path := strings.TrimPrefix(line, "<!-- Path:")
			path = strings.TrimSpace(path)
			path = strings.TrimSuffix(path, "-->")
			path = strings.TrimSpace(path)
			if path == "" {
				return "", fmt.Errorf("%w: empty path", ErrInvalidPathSyntax)
			}
			return path, nil
		}
	}
	return "", nil
}

func ParsePrompts(content string) []Prompt {
	var prompts []Prompt
	matches := promptRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		prompts = append(prompts, Prompt{
			Question: match[1],
			Variable: nil,
		})
	}
	return prompts
}

func ParseVariables(content string) ([]Variable, error) {
	var vars []Variable
	seen := make(map[string]bool)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		matches := variableRegex.FindStringSubmatch(line)
		if matches != nil {
			name := matches[1]
			prompt := matches[2]

			if seen[name] {
				return nil, fmt.Errorf("%w: %s", ErrVariableConflict, name)
			}
			seen[name] = true

			vars = append(vars, Variable{
				Name:   name,
				Prompt: prompt,
			})
		}
	}
	return vars, nil
}

func ParseTemplate(filepathStr, content string) (*Template, error) {
	filename := filepath.Base(filepathStr)

	priority, category, name, err := ParseFilename(filename)
	if err != nil {
		return nil, err
	}

	targetPath, err := ParsePathComment(content)
	if err != nil {
		return nil, err
	}

	variables, err := ParseVariables(content)
	if err != nil {
		return nil, &TemplateError{Template: filename, Err: err}
	}

	prompts := ParsePrompts(content)

	return &Template{
		Filename:   filename,
		Priority:   priority,
		Category:   category,
		Name:       name,
		Content:    content,
		TargetPath: targetPath,
		Variables:  variables,
		Prompts:    prompts,
		BuiltIns:   make(map[string]string),
	}, nil
}

func ParseFilename(filename string) (priority int, category, name string, err error) {
	base := strings.TrimSuffix(filename, ".md")

	parts := strings.Split(base, "-")
	if len(parts) < 3 {
		return 0, "", "", fmt.Errorf("%w: %s", ErrInvalidFilename, filename)
	}

	priority, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", "", fmt.Errorf("%w: invalid priority", ErrInvalidFilename)
	}

	category = parts[1]
	name = strings.Join(parts[2:], "-")

	return priority, category, name, nil
}
