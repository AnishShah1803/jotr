package notes

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/search"
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
// Note: External edits via this function do not automatically update the search index.
// Run 'jotr index sync' to sync external modifications.
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
	return os.MkdirAll(path, constants.FilePermDir)
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

	if err := os.WriteFile(path, []byte(content), constants.FilePerm0644); err != nil {
		return err
	}

	updateIndex(ctx, path, content)

	return nil
}

func updateIndex(ctx context.Context, path string, content string) {
	cfg, err := config.LoadWithContext(ctx, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "updateIndex: failed to load config: %v\n", err)
		return
	}

	indexPath := search.GetIndexPath(cfg.Paths.BaseDir)

	if _, err := os.Stat(indexPath); err != nil {
		return
	}

	idx, err := search.Open(indexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "updateIndex: failed to open index: %v\n", err)
		return
	}
	defer idx.Close()

	title := search.ExtractTitle(content)
	info, _ := os.Stat(path)
	var modTime time.Time
	if info != nil {
		modTime = info.ModTime()
	} else {
		modTime = time.Now()
	}

	if err := idx.IndexNote(ctx, path, title, content, modTime); err != nil {
		fmt.Fprintf(os.Stderr, "updateIndex: failed to index note: %v\n", err)
	}
}

// FindNotes finds all markdown files in a directory recursively with context support.
func FindNotes(ctx context.Context, dir string) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var notes []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.IsDir() && d.Name() == ".jotr" {
			return filepath.SkipDir
		}

		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			notes = append(notes, path)
		}

		return nil
	})

	return notes, err
}

// SearchMatch represents a file that matched a search query, including its content.
type SearchMatch struct {
	Path    string
	Content []byte
}

func SearchNotes(ctx context.Context, dir string, query string) ([]SearchMatch, error) {
	indexPath := search.GetIndexPath(dir)

	if _, err := os.Stat(indexPath); err == nil {
		idx, err := search.Open(indexPath)
		if err == nil {
			defer idx.Close()
			results, err := idx.Search(ctx, query, 0)
			if err == nil && len(results) > 0 {
				matches := make([]SearchMatch, len(results))
				for i, r := range results {
					content, _ := os.ReadFile(r.Path)
					matches[i] = SearchMatch{
						Path:    r.Path,
						Content: content,
					}
				}
				return matches, nil
			}
		}
	}

	allNotes, err := FindNotes(ctx, dir)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)

	results, err := utils.ProcessFilesParallelWithContent(ctx, allNotes, 0, func(path string, content []byte) bool {
		return strings.Contains(strings.ToLower(string(content)), query)
	})
	if err != nil {
		return nil, err
	}

	matches := make([]SearchMatch, len(results))
	for i, r := range results {
		matches[i] = SearchMatch{
			Path:    r.Path,
			Content: r.Content,
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

	if err := os.WriteFile(notePath, []byte(content), constants.FilePerm0644); err != nil {
		return err
	}

	updateIndex(ctx, notePath, content)

	return nil
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
	contentMap := make(map[string][]byte)

	results, err := utils.ProcessFilesParallelWithContent(ctx, allNotes, 0, func(path string, content []byte) bool {
		return true
	})
	if err != nil {
		return err
	}

	for _, r := range results {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		name := strings.TrimSuffix(filepath.Base(r.Path), ".md")
		noteMap[name] = r.Path
		contentMap[r.Path] = r.Content

		title := search.ExtractTitle(string(r.Content))
		if title == "" {
			title = name
		}
		titleMap[name] = title
	}

	for _, notePath := range allNotes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		content, exists := contentMap[notePath]
		if !exists {
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
			if err := utils.AtomicWriteFileCtx(ctx, notePath, []byte(updatedContent), constants.FilePerm0644); err != nil {
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

	indexPath := search.GetIndexPath(cfg.Paths.BaseDir)

	if _, err := os.Stat(indexPath); err == nil {
		idx, err := search.Open(indexPath)
		if err == nil {
			defer idx.Close()
			results, err := idx.SearchByTag(ctx, tag, 0)
			if err == nil && len(results) > 0 {
				paths := make([]string, len(results))
				for i, r := range results {
					paths[i] = r.Path
				}
				return paths, nil
			}
		}
	}

	allNotes, err := FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return nil, err
	}

	tagPattern := "#" + tag

	matching, err := utils.ProcessFilesParallel(ctx, allNotes, 0, func(path string, content []byte) bool {
		return strings.Contains(string(content), tagPattern)
	})
	if err != nil {
		return nil, err
	}

	return matching, nil
}
