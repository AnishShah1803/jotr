package notes

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// Note represents a note file.
type Note struct {
	Date     time.Time
	Metadata map[string]interface{}
	Path     string
	Name     string
	Content  string
}

// OpenInEditor opens a file in the user's preferred editor.
func OpenInEditor(path string) error {
	return OpenInEditorWithContext(context.Background(), path)
}

func OpenInEditorWithContext(ctx context.Context, path string) error {
	editor := config.GetEditorWithContext(ctx)

	// Check if editor is configured
	if editor == "" {
		return fmt.Errorf("no editor configured - set EDITOR environment variable or configure editor.default")
	}

	// Validate editor before execution
	if err := utils.ValidateEditor(editor); err != nil {
		return fmt.Errorf("invalid editor: %w", err)
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetEditorCmd returns a command to open a file in the editor.
func GetEditorCmd(path string) (*exec.Cmd, error) {
	return GetEditorCmdWithContext(context.Background(), path)
}

// GetEditorCmdWithContext returns a command to open a file in the editor, with context support.
func GetEditorCmdWithContext(ctx context.Context, path string) (*exec.Cmd, error) {
	editor := config.GetEditorWithContext(ctx)

	// Check if editor is configured
	if editor == "" {
		return nil, fmt.Errorf("no editor configured - set EDITOR environment variable or configure editor.default")
	}

	// Validate editor before execution
	if err := utils.ValidateEditor(editor); err != nil {
		return nil, fmt.Errorf("invalid editor: %w", err)
	}

	cmd := exec.Command(editor, path)

	return cmd, nil
}

func GetEditorCmdWithShellFallback(ctx context.Context, path string) (*exec.Cmd, error) {
	editor := config.GetEditorWithContext(ctx)

	if editor == "" {
		return nil, fmt.Errorf("no editor configured - set EDITOR environment variable or configure editor.default")
	}

	cmd := exec.Command(editor, path)

	return cmd, nil
}

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// ReadNote reads a note file with context support for cancellation.
func ReadNote(ctx context.Context, path string) (*Note, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	note := &Note{
		Path:    path,
		Name:    filepath.Base(path),
		Content: string(content),
	}

	return note, nil
}

// WriteNote writes content to a note file with context support.
func WriteNote(ctx context.Context, path string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// FindNotes finds all markdown files in a directory recursively with context support.
func FindNotes(ctx context.Context, dir string) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var notes []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation during traversal
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			notes = append(notes, path)
		}

		return nil
	})

	return notes, err
}

// SearchNotes searches for notes containing a query with context support.
func SearchNotes(ctx context.Context, dir string, query string) ([]string, error) {
	allNotes, err := FindNotes(ctx, dir)
	if err != nil {
		return nil, err
	}

	var matches []string

	query = strings.ToLower(query)

	for _, notePath := range allNotes {
		// Check context before reading each file
		select {
		case <-ctx.Done():
			return matches, ctx.Err()
		default:
		}

		content, err := os.ReadFile(notePath)
		if err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(string(content)), query) {
			matches = append(matches, notePath)
		}
	}

	return matches, nil
}

// BuildDailyNotePath builds the path for a daily note.
func BuildDailyNotePath(diaryDir string, date time.Time) string {
	year := date.Format("2006")
	monthNum := date.Format("01")
	monthAbbr := date.Format("Jan")
	month := date.Format("01")
	day := date.Format("02")
	weekday := date.Format("Mon")

	dirPath := filepath.Join(diaryDir, year, fmt.Sprintf("%s-%s", monthNum, monthAbbr))
	filename := fmt.Sprintf("%s-%s-%s-%s.md", year, month, day, weekday)

	return filepath.Join(dirPath, filename)
}

// CreateDailyNote creates a daily note with template with context support.
// It adds a Task section at the end if not already present in the sections.
func CreateDailyNote(ctx context.Context, notePath string, sections []string, date time.Time) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	dir := filepath.Dir(notePath)
	if err := EnsureDir(dir); err != nil {
		return err
	}

	content := fmt.Sprintf("# %s\n\n", date.Format("2006-01-02-Mon"))

	for _, section := range sections {
		content += fmt.Sprintf("## %s\n\n", section)
	}

	return os.WriteFile(notePath, []byte(content), 0644)
}

// BuildDailyNoteSections prepares the complete sections list for a daily note,
// including daily_note_sections from config and ensuring a Task section exists.
func BuildDailyNoteSections(cfg *config.LoadedConfig) []string {
	var allSections []string
	allSections = append(allSections, cfg.Format.DailyNoteSections...)

	taskSection := cfg.Format.TaskSection
	if taskSection == "" {
		taskSection = "Tasks"
	}

	hasTaskSection := false

	for _, section := range cfg.Format.DailyNoteSections {
		if section == taskSection {
			hasTaskSection = true
			break
		}
	}

	if !hasTaskSection {
		allSections = append(allSections, taskSection)
	}

	return allSections
}

// GetRecentDailyNotes gets the most recent daily notes with context support.
func GetRecentDailyNotes(ctx context.Context, diaryDir string, count int) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var notes []string

	today := time.Now()

	for i := 0; i < count; i++ {
		// Check context before processing each note
		select {
		case <-ctx.Done():
			return notes, ctx.Err()
		default:
		}

		date := today.AddDate(0, 0, -i)

		notePath := BuildDailyNotePath(diaryDir, date)
		if utils.FileExists(notePath) {
			notes = append(notes, notePath)
		}
	}

	return notes, nil
}

// UpdateLinks updates wiki-style links in all notes with context support.
func UpdateLinks(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cfg, err := config.LoadWithContext(ctx, "")
	if err != nil {
		return err
	}

	allNotes, err := FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	noteMap := make(map[string]string)
	titleMap := make(map[string]string)

	for _, notePath := range allNotes {
		// Check context during processing
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		name := strings.TrimSuffix(filepath.Base(notePath), ".md")
		noteMap[name] = notePath

		if content, err := os.ReadFile(notePath); err == nil {
			contentStr := string(content)

			lines := strings.Split(contentStr, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "# ") {
					title := strings.TrimSpace(strings.TrimPrefix(line, "# "))
					titleMap[name] = title

					break
				}
			}
		}

		if titleMap[name] == "" {
			titleMap[name] = name
		}
	}

	for _, notePath := range allNotes {
		// Check context before processing each note
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		content, err := os.ReadFile(notePath)
		if err != nil {
			continue
		}

		contentStr := string(content)
		updatedContent := contentStr

		for noteName := range noteMap {
			title := titleMap[noteName]
			linkPattern := "[[" + noteName + "]]"
			replacement := "[[" + noteName + "|" + title + "]]"
			updatedContent = strings.ReplaceAll(updatedContent, linkPattern, replacement)
		}

		if updatedContent != contentStr {
			if err := utils.AtomicWriteFileCtx(ctx, notePath, []byte(updatedContent), 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateNote creates a note file with the given content.
func CreateNote(ctx context.Context, path string, content string) error {
	return WriteNote(ctx, path, content)
}

// GetDailyNotePath returns the path for today's daily note.
func GetDailyNotePath(date time.Time) (string, error) {
	cfg, err := config.LoadWithContext(context.Background(), "")
	if err != nil {
		return "", err
	}

	return filepath.Join(cfg.Paths.BaseDir, BuildDailyNotePath(cfg.Paths.DiaryDir, date)), nil
}

// GetNotesByPattern finds notes matching the given glob pattern.
func GetNotesByPattern(ctx context.Context, pattern string) ([]string, error) {
	cfg, err := config.LoadWithContext(ctx, "")
	if err != nil {
		return nil, err
	}

	return FindNotes(ctx, cfg.Paths.BaseDir)
}

// GetNotesByTag finds notes containing the given tag with context support.
func GetNotesByTag(ctx context.Context, tag string) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cfg, err := config.LoadWithContext(ctx, "")
	if err != nil {
		return nil, err
	}

	notes, err := FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return nil, err
	}

	var matching []string

	tagPattern := "#" + tag

	for _, notePath := range notes {
		// Check context before reading each file
		select {
		case <-ctx.Done():
			return matching, ctx.Err()
		default:
		}

		content, err := os.ReadFile(notePath)
		if err != nil {
			continue
		}

		if strings.Contains(string(content), tagPattern) {
			matching = append(matching, notePath)
		}
	}

	return matching, nil
}
