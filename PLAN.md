# 📋 IMPLEMENTATION PLAN: File-Driven Template System for jotr

## Summary

**40 atomic tasks** organized into 9 phases, building from core data structures through CLI commands to full testing.

## Quick Reference

| Phase | Tasks | Focus |
|-------|-------|-------|
| 1 | 1-5 | Data structures & config ✓ |
| 2 | 6-10 | Template parsing ✓ |
| 3 | 11-13 | Template discovery ✓ |
| 4 | 14-16 | Prompt collection ✓ |
| 5 | 17-19 | Template rendering ✓ |
| 6 | 20-22 | File creation ✓ |
| 7 | 23-29 | CLI commands ✓ |
| 8 | 30-33 | Dashboard integration ✓ |
| 9 | 34-40 | Testing ✓ |

**STATUS: ALL PHASES COMPLETE ✓**

---

## Template Format Example

**Filename**: `1-1-meeting.md`

**Content**:
```markdown
<!-- Path: ~/Notes/Meetings/{$date}-{$topic}.md -->

# Meeting Notes: {$topic}

**Date:** {$date}
**Time:** {$time}

## Attendees
<prompt>Who attended this meeting?</prompt>

## Agenda
<prompt>What are the agenda items?</prompt>
```

---

## Key Design Decisions

1. **Directory**: `{base_dir}/templates/` (e.g., `~/Notes/templates/`)
2. **Filename format**: `{priority}-{category}-{name}.md`
3. **Variable syntax**: `name = <prompt>question?</prompt>` → creates `{$name}`
4. **Freeform prompts**: `<prompt>question?</prompt>` → replaced with user input
5. **Built-in vars**: `{$date}`, `{$datetime}`, `{$base_dir}`, `{$weekday}`, `{$time}`
6. **Error handling**: Warn and continue (skip invalid templates)
7. **Commands**: `jotr template`, `jotr template edit`

---

## Phase 1: Core Data Structures (Tasks 1-5) ✓

Create the foundational types for the template system.

### Task 1: Define Template core types ✓
**File**: `internal/templates/template.go`
**Steps**:
```go
package templates

import "time"

type Variable struct {
    Name   string
    Prompt string
    Value  string
}

type Prompt struct {
    Question string
    Variable *string  // nil if freeform
}

type Template struct {
    Filename      string
    Priority      int
    Category      string
    Name          string
    Content       string
    TargetPath    string
    Variables     []Variable
    Prompts       []Prompt
    BuiltIns      map[string]string
}
```
**Validation**: Compile check, verify struct fields match requirements
**Dependencies**: None

### Task 2: Define template-specific errors ✓
**File**: `internal/templates/errors.go`
**Steps**:
```go
package templates

import "errors"

var (
    ErrTemplateNotFound    = errors.New("template not found")
    ErrInvalidFilename     = errors.New("invalid template filename format")
    ErrParseFailed         = errors.New("failed to parse template")
    ErrVariableConflict    = errors.New("variable defined multiple times")
    ErrInvalidPathSyntax   = errors.New("invalid path syntax in <!-- Path: -->")
)

type TemplateError struct {
    Template string
    Err      error
}

func (e *TemplateError) Error() string {
    return fmt.Sprintf("template %s: %s", e.Template, e.Err.Error())
}

func (e *TemplateError) Unwrap() error {
    return e.Err
}
```
**Validation**: Compile check
**Dependencies**: Task 1

### Task 3: Add TemplatesDir to LoadedConfig ✓
**File**: `internal/config/config.go`
**Steps**:
- Add field to `LoadedConfig` struct:
```go
type LoadedConfig struct {
    Config
    DiaryPath    string
    TodoPath     string
    StatePath    string
    PDPPath      string
    TemplatesPath string  // NEW: BaseDir + "templates/"
}
```
- Update `LoadWithContext()` to compute path:
```go
loaded.TemplatesPath = filepath.Join(cfg.Paths.BaseDir, "templates")
```
**Validation**: Run `go build`, verify path computed correctly
**Dependencies**: None

### Task 4: Add template discovery helper ✓
**File**: `internal/templates/discovery.go` (stub)
**Steps**:
```go
func GetTemplatesDir(cfg *config.LoadedConfig) string {
    return cfg.TemplatesPath
}
```
**Validation**: Verify function returns correct path
**Dependencies**: Task 3

### Task 5: Create template test helpers ✓
**File**: `internal/templates/testhelpers.go`
**Steps**:
```go
package templates

import "testing"

func CreateTestTemplateDir(t *testing.T) (string, func()) {
    t.Helper()
    tmpDir := t.TempDir()
    templatesDir := filepath.Join(tmpDir, "templates")
    os.MkdirAll(templatesDir, 0755)
    return templatesDir, func() { os.RemoveAll(tmpDir) }
}

func CreateTestTemplate(t *testing.T, dir, name, content string) {
    t.Helper()
    path := filepath.Join(dir, name)
    os.WriteFile(path, []byte(content), 0644)
}
```
**Validation**: Run `go test ./internal/templates/`
**Dependencies**: Task 1, Task 2

---

## Phase 2: Template Parsing (Tasks 6-10) ✓

Parse template files to extract metadata, variables, and prompts.

### Task 6: Parse filename components ✓
**File**: `internal/templates/parser.go`
**Steps**:
```go
func ParseFilename(filename string) (priority int, category, name string, err error) {
    // Remove .md extension
    base := strings.TrimSuffix(filename, ".md")

    // Split by dash
    parts := strings.Split(base, "-")
    if len(parts) < 3 {
        return 0, "", "", fmt.Errorf("%w: %s", ErrInvalidFilename, filename)
    }

    // Parse priority
    priority, err = strconv.Atoi(parts[0])
    if err != nil {
        return 0, "", "", fmt.Errorf("%w: invalid priority", ErrInvalidFilename)
    }

    // Join category and name (may contain dashes)
    category = parts[1]
    name = strings.Join(parts[2:], "-")

    return priority, category, name, nil
}
```
**Validation**: Add table-driven tests for valid/invalid filenames
**Dependencies**: Task 2

### Task 7: Parse path from frontmatter comment ✓
**File**: `internal/templates/parser.go`
**Steps**:
```go
func ParsePathComment(content string) (string, error) {
    lines := strings.Split(content, "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "<!-- Path:") {
            // Extract path between <!-- Path: and -->
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
    return "", nil  // No path comment is valid
}
```
**Validation**: Test with/without comment, valid/invalid syntax
**Dependencies**: Task 2

### Task 8: Parse variable definitions ✓
**File**: `internal/templates/parser.go`
**Steps**:
```go
var variableRegex = regexp.MustCompile(`^(\w+)\s*=\s*<prompt>(.*?)</prompt>`)

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
```
**Validation**: Test single/multiple variables, duplicates, edge cases
**Dependencies**: Task 2

### Task 9: Parse freeform prompts ✓
**File**: `internal/templates/parser.go`
**Steps**:
```go
var promptRegex = regexp.MustCompile(`<prompt>(.*?)</prompt>`)

func ParsePrompts(content string) []Prompt {
    var prompts []Prompt
    matches := promptRegex.FindAllStringSubmatch(content, -1)

    for _, match := range matches {
        prompts = append(prompts, Prompt{
            Question: match[1],
            Variable: nil,  // Freeform
        })
    }
    return prompts
}
```
**Validation**: Test with multiple prompts, no prompts
**Dependencies**: Task 2

### Task 10: Create complete template parser ✓
**File**: `internal/templates/parser.go`
**Steps**:
```go
func ParseTemplate(filepath, content string) (*Template, error) {
    filename := filepath.Base(filepath)

    // Parse filename
    priority, category, name, err := ParseFilename(filename)
    if err != nil {
        return nil, err
    }

    // Parse path comment
    targetPath, err := ParsePathComment(content)
    if err != nil {
        return nil, err
    }

    // Parse variables
    variables, err := ParseVariables(content)
    if err != nil {
        return nil, &TemplateError{Template: filename, Err: err}
    }

    // Parse prompts
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
```
**Validation**: Integration test with complete template file
**Dependencies**: Task 6, Task 7, Task 8, Task 9

---

---

---

## Phase 3: Template Discovery (Tasks 11-13) ✓

Scan templates directory and build template list.

### Task 11: Implement template directory scanner ✓
**File**: `internal/templates/discovery.go`
**Steps**:
```go
func DiscoverTemplates(templatesDir string) ([]*Template, []error) {
    var templates []*Template
    var errors []error

    entries, err := os.ReadDir(templatesDir)
    if err != nil {
        if os.IsNotExist(err) {
            return templates, nil  // No templates dir is OK
        }
        return nil, err
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
            continue  // Skip invalid templates
        }

        templates = append(templates, template)
    }

    return templates, errors
}
```
**Validation**: Test with empty dir, valid templates, invalid templates
**Dependencies**: Task 10

### Task 12: Sort templates by priority and category ✓
**File**: `internal/templates/discovery.go`
**Steps**:
```go
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
```
**Validation**: Test sorting with various priorities
**Dependencies**: Task 11

### Task 13: Add template discovery wrapper with warnings ✓
**File**: `internal/templates/discovery.go`
**Steps**:
```go
func LoadTemplates(cfg *config.LoadedConfig) ([]*Template, []string) {
    templates, errs := DiscoverTemplates(cfg.TemplatesPath)

    var warnings []string
    for _, err := range errs {
        warnings = append(warnings, err.Error())
    }

    return SortTemplates(templates), warnings
}
```
**Validation**: Test warning output for invalid templates
**Dependencies**: Task 11, Task 12

---

---

## Phase 4: Prompt Collection (Tasks 14-16) ✓

Collect user input for variables and prompts.

### Task 14: Implement built-in variable resolution ✓
**File**: `internal/templates/renderer.go`
**Steps**:
```go
func ResolveBuiltIns(tmpl *Template) {
    now := time.Now()

    tmpl.BuiltIns = map[string]string{
        "{$date}":     now.Format("2006-01-02"),
        "{$datetime}": now.Format("2006-01-02 15:04"),
        "{$base_dir}": config.Get().Paths.BaseDir,  // Or pass in
        "{$weekday}":  now.Format("Monday"),
        "{$time}":     now.Format("15:04"),
    }
}
```
**Validation**: Test all built-ins resolve correctly
**Dependencies**: Task 1, Task 3

### Task 15: Collect user input for variables ✓
**File**: `internal/templates/collector.go`
**Steps**:
```go
import "github.com/AnishShah1803/jotr/internal/utils"

func CollectVariableValues(vars []Variable) (map[string]string, error) {
    values := make(map[string]string)

    for _, v := range vars {
        value, err := utils.PromptUserRequired(v.Prompt)
        if err != nil {
            return nil, err
        }
        values[v.Name] = value
    }

    return values, nil
}
```
**Validation**: Mock input and test collection
**Dependencies**: Task 1

### Task 16: Collect user input for freeform prompts ✓
**File**: `internal/templates/collector.go`
**Steps**:
```go
func CollectPromptValues(prompts []Prompt) ([]string, error) {
    var values []string

    for _, p := range prompts {
        value, err := utils.PromptUserRequired(p.Question)
        if err != nil {
            return nil, err
        }
        values = append(values, value)
    }

    return values, nil
}
```
**Validation**: Test with multiple prompts
**Dependencies**: Task 1

---

---

## Phase 5: Template Rendering (Tasks 17-19) ✓

Substitute variables and prompts to create final content.

### Task 17: Substitute variables in content ✓
**File**: `internal/templates/renderer.go`
**Steps**:
```go
func SubstituteVariables(content string, vars, builtins map[string]string) string {
    result := content

    // Substitute user variables
    for name, value := range vars {
        placeholder := fmt.Sprintf("{$%s}", name)
        result = strings.ReplaceAll(result, placeholder, value)
    }

    // Substitute built-ins
    for name, value := range builtins {
        result = strings.ReplaceAll(result, name, value)
    }

    return result
}
```
**Validation**: Test substitution with various patterns
**Dependencies**: Task 14

### Task 18: Replace freeform prompts with values ✓
**File**: `internal/templates/renderer.go`
**Steps**:
```go
func SubstitutePrompts(content string, values []string) string {
    result := content
    promptRegex := regexp.MustCompile(`<prompt>.*?</prompt>`)

    matches := promptRegex.FindAllStringIndex(content, -1)
    for i, match := range matches {
        if i < len(values) {
            start := match[0]
            end := match[1]
            result = result[:start] + values[i] + result[end:]
        }
    }

    return result
}
```
**Validation**: Test prompt replacement
**Dependencies**: Task 16

### Task 19: Render complete template ✓
**File**: `internal/templates/renderer.go`
**Steps**:
```go
func RenderTemplate(tmpl *Template, userVars map[string]string, promptValues []string) string {
    // Resolve built-ins
    ResolveBuiltIns(tmpl)

    // Substitute variables
    content := SubstituteVariables(tmpl.Content, userVars, tmpl.BuiltIns)

    // Substitute prompts
    content = SubstitutePrompts(content, promptValues)

    // Remove path comment if present
    lines := strings.Split(content, "\n")
    var filtered []string
    for _, line := range lines {
        if !strings.HasPrefix(strings.TrimSpace(line), "<!-- Path:") {
            filtered = append(filtered, line)
        }
    }

    return strings.Join(filtered, "\n")
}
```
**Validation**: Integration test with full template
**Dependencies**: Task 17, Task 18

---

---

## Phase 6: File Creation (Tasks 20-22)

Validate paths and create files from rendered templates.

### Task 20: Render target path with variables ✓
**File**: `internal/templates/renderer.go`
**Steps**:
```go
func RenderTargetPath(tmpl *Template, vars map[string]string) (string, error) {
    if tmpl.TargetPath == "" {
        return "", fmt.Errorf("template has no target path")
    }

    ResolveBuiltIns(tmpl)

    path := SubstituteVariables(tmpl.TargetPath, vars, tmpl.BuiltIns)

    // Expand ~ if present
    if strings.HasPrefix(path, "~/") {
        home, _ := os.UserHomeDir()
        path = filepath.Join(home, path[2:])
    }

    return path, nil
}
```
**Validation**: Test path expansion and variable substitution
**Dependencies**: Task 17

### Task 21: Create file from rendered template ✓
**File**: `internal/templates/creator.go`
**Steps**:
```go
import "github.com/AnishShah1803/jotr/internal/utils"

func CreateFromTemplate(ctx context.Context, tmpl *Template, content, targetPath string) error {
    // Ensure directory exists
    dir := filepath.Dir(targetPath)
    if err := utils.EnsureDir(dir); err != nil {
        return fmt.Errorf("creating directory: %w", err)
    }

    // Check if file exists
    if utils.FileExists(targetPath) {
        return fmt.Errorf("file already exists: %s", targetPath)
    }

    // Atomic write
    if err := utils.AtomicWriteFileCtx(ctx, targetPath, []byte(content), 0644); err != nil {
        return fmt.Errorf("writing file: %w", err)
    }

    return nil
}
```
**Validation**: Test file creation, directory creation, error handling
**Dependencies**: None (uses existing utils) ✓

### Task 22: Open created file in editor ✓
**File**: `internal/templates/creator.go`
**Steps**:
```go
import "github.com/AnishShah1803/jotr/internal/notes"

func CreateAndOpen(ctx context.Context, tmpl *Template, content, targetPath string) error {
    if err := CreateFromTemplate(ctx, tmpl, content, targetPath); err != nil {
        return err
    }

    return notes.OpenNoteInEditor(ctx, targetPath)
}
```
**Validation**: Test that editor is called after creation
**Dependencies**: Task 21 ✓

---

## Phase 7: CLI Commands (Tasks 23-29)

Implement `jotr template` and `jotr template edit` commands.

### Task 23: Create template command structure
**File**: `cmd/template/template.go`
**Steps**:
```go
package templatecmd

import "github.com/spf13/cobra"

var TemplateCmd = &cobra.Command{
    Use:   "template",
    Short: "Create note from template",
    RunE:  runTemplate,
}

func init() {
    // Add flags if needed
}
```
**Validation**: Run `go build`, command registers
**Dependencies**: None

### Task 24: Implement interactive template selector
**File**: `cmd/template/template.go`
**Steps**:
```go
func runTemplate(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    cfg, err := config.LoadWithContext(ctx, "")
    if err != nil {
        return err
    }

    // Load templates
    templates, warnings := templates.LoadTemplates(cfg)

    // Print warnings
    for _, warn := range warnings {
        fmt.Fprintf(os.Stderr, "Warning: %s\n", warn)
    }

    if len(templates) == 0 {
        return fmt.Errorf("no templates found in %s", cfg.TemplatesPath)
    }

    // Display templates grouped by category
    selected := selectTemplateInteractive(templates)

    return createFromSelectedTemplate(ctx, cfg, selected)
}
```
**Validation**: Test with no templates, one template, multiple templates
**Dependencies**: Task 13, Task 22

### Task 25: Implement template selection UI
**File**: `cmd/template/template.go`
**Steps**:
```go
func selectTemplateInteractive(templates []*Template) *Template {
    // Group by category
    categories := make(map[string][]*Template)
    for _, tmpl := range templates {
        categories[tmpl.Category] = append(categories[tmpl.Category], tmpl)
    }

    // Display numbered menu
    fmt.Println("\nAvailable Templates:")
    index := 1
    for cat, tmpls := range categories {
        fmt.Printf("\n%s:\n", cat)
        for _, tmpl := range tmpls {
            fmt.Printf("  %d. %s\n", index, tmpl.Name)
            index++
        }
    }

    // Get user choice
    choice, err := utils.PromptChoice("Select template", 1, index-1)
    if err != nil {
        return nil
    }

    // Find selected template
    flat := flattenTemplates(categories)
    return flat[choice-1]
}
```
**Validation**: Test selection with multiple categories
**Dependencies**: Task 24

### Task 26: Implement template creation workflow
**File**: `cmd/template/template.go`
**Steps**:
```go
func createFromSelectedTemplate(ctx context.Context, cfg *config.LoadedConfig, tmpl *Template) error {
    // Collect variable values
    varValues, err := templates.CollectVariableValues(tmpl.Variables)
    if err != nil {
        return err
    }

    // Collect prompt values
    promptValues, err := templates.CollectPromptValues(tmpl.Prompts)
    if err != nil {
        return err
    }

    // Render template
    content := templates.RenderTemplate(tmpl, varValues, promptValues)

    // Get target path
    targetPath, err := templates.RenderTargetPath(tmpl, varValues)
    if err != nil {
        return err
    }

    // Create and open
    return templates.CreateAndOpen(ctx, tmpl, content, targetPath)
}
```
**Validation**: End-to-end test with real template
**Dependencies**: Task 19, Task 22

### Task 27: Register template command in root
**File**: `cmd/root.go`
**Steps**:
- Add import: `"github.com/AnishShah1803/jotr/cmd/template"`
- Add to rootCmd: `rootCmd.AddCommand(templatecmd.TemplateCmd)`
**Validation**: Run `jotr template`
**Dependencies**: Task 23

### Task 28: Create edit command structure
**File**: `cmd/template/edit.go`
**Steps**:
```go
package templatecmd

import "github.com/spf13/cobra"

var EditCmd = &cobra.Command{
    Use:   "edit",
    Short: "Edit templates",
    RunE:  runEdit,
}

func init() {
    TemplateCmd.AddCommand(EditCmd)
}
```
**Validation**: Run `jotr template edit`
**Dependencies**: Task 23

### Task 29: Implement template edit workflow
**File**: `cmd/template/edit.go`
**Steps**:
```go
func runEdit(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    cfg, err := config.LoadWithContext(ctx, "")
    if err != nil {
        return err
    }

    // List templates
    templates, _ := templates.LoadTemplates(cfg)

    if len(templates) == 0 {
        return fmt.Errorf("no templates found")
    }

    // Simple numbered list
    fmt.Println("\nTemplates:")
    for i, tmpl := range templates {
        fmt.Printf("%d. %s\n", i+1, tmpl.Filename)
    }

    // Select template to edit
    choice, err := utils.PromptChoice("Select template to edit", 1, len(templates))
    if err != nil {
        return err
    }

    selected := templates[choice-1]
    templatePath := filepath.Join(cfg.TemplatesPath, selected.Filename)

    // Open in editor
    return notes.OpenNoteInEditor(ctx, templatePath)
}
```
**Validation**: Test edit workflow
**Dependencies**: Task 13

---

## Phase 8: Dashboard Integration (Tasks 30-33)

Add template option to quick menu.

### Task 30: Add template action to dashboard
**File**: `cmd/util/quick.go`
**Steps**:
- Add template option to quick menu
- Bind to template creation workflow
**Validation**: Run quick menu, verify template option
**Dependencies**: Task 26

### Task 31: Add keyboard shortcut for templates
**File**: `cmd/util/quick.go`
**Steps**:
- Bind `t` key to template action
**Validation**: Test keybinding in quick menu
**Dependencies**: Task 30

### Task 32: Create template from dashboard
**File**: `cmd/util/quick.go`
**Steps**:
- Reuse template selection UI from template command
- Integrate with quick menu workflow
**Validation**: Test end-to-end from quick menu
**Dependencies**: Task 25, Task 30

### Task 33: Refresh dashboard after template creation
**File**: `cmd/util/quick.go`
**Steps**:
- Add refresh logic after file creation
- Ensure new file appears in quick menu
**Validation**: Create template, verify it appears
**Dependencies**: Task 32

---

## Phase 9: Testing (Tasks 34-40)

Comprehensive unit and integration tests.

### Task 34: Test filename parsing
**File**: `internal/templates/parser_test.go`
**Steps**: Create table-driven tests for:
- Valid filenames: `1-1-meeting.md`, `10-2-journal-entry.md`
- Invalid filenames: `meeting.md`, `a-b-c.md`, `1-name.md`
- Edge cases: `1--name.md` (empty category)
**Validation**: `go test -v ./internal/templates/`
**Dependencies**: Task 6

### Task 35: Test variable parsing
**File**: `internal/templates/parser_test.go`
**Steps**: Test:
- Single variable
- Multiple variables
- Duplicate detection
- Invalid syntax
**Validation**: All tests pass
**Dependencies**: Task 8

### Task 36: Test prompt parsing
**File**: `internal/templates/parser_test.go`
**Steps**: Test:
- Single prompt
- Multiple prompts
- Mixed with variables
- Nested prompts (should handle)
**Validation**: All tests pass
**Dependencies**: Task 9

### Task 37: Test template discovery
**File**: `internal/templates/discovery_test.go`
**Steps**: Test:
- Empty directory
- Valid templates only
- Mixed valid/invalid templates
- Non-existent directory
**Validation**: All tests pass
**Dependencies**: Task 11

### Task 38: Test variable substitution
**File**: `internal/templates/renderer_test.go`
**Steps**: Test:
- Single variable
- Multiple variables
- Built-in variables
- Missing variables (should not crash)
- Overlapping names
**Validation**: All tests pass
**Dependencies**: Task 17

### Task 39: Test complete template workflow
**File**: `internal/templates/integration_test.go`
**Steps**: End-to-end test:
1. Create test template file
2. Parse template
3. Collect input (mocked)
4. Render template
5. Verify output
6. Create file
7. Verify file contents
**Validation**: Integration test passes
**Dependencies**: Task 22

### Task 40: Test CLI commands
**File**: `cmd/template/template_test.go`
**Steps**: Test:
- `jotr template` with no templates
- `jotr template` with templates
- `jotr template edit`
- Error handling
**Validation**: All command tests pass
**Dependencies**: Task 29

---

## Implementation Summary

| Phase | Tasks | Focus Area |
|-------|-------|------------|
| 1 | 1-5 | Data structures & config |
| 2 | 6-10 | Parsing logic |
| 3 | 11-13 | Template discovery |
| 4 | 14-16 | User input collection |
| 5 | 17-19 | Variable substitution |
| 6 | 20-22 | File creation |
| 7 | 23-29 | CLI commands |
| 8 | 30-33 | Dashboard integration |
| 9 | 34-40 | Testing |

**Total: 40 atomic tasks**

---

## Key Design Decisions

1. **No External Template Engine**: Use regex-based parsing (simpler, fewer dependencies)
2. **Error Recovery**: Warn + continue for invalid templates
3. **Filename Convention**: Priority-based sorting for UX
4. **Built-in Variables**: Pre-populated map for easy access
5. **Atomic Writes**: Reuse existing `AtomicWriteFileCtx()`
6. **Editor Integration**: Reuse `OpenNoteInEditor()`
7. **Config Integration**: Extend `LoadedConfig`, avoid breaking changes

---

## Validation Checklist

- [x] All 40 tasks defined with dependencies
- [x] Follows existing codebase patterns (context, utils, errors)
- [x] Reuses existing infrastructure (no reinventing)
- [x] Comprehensive test coverage (unit + integration)
- [x] CLI commands follow cobra patterns
- [x] Error handling matches jotr conventions
- [x] No breaking changes to existing functionality

**This plan is ready for implementation.**
