package notes

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/anish/jotr/internal/utils"
)

// Note represents a note file
type Note struct {
	Path     string
	Name     string
	Date     time.Time
	Content  string
	Metadata map[string]interface{}
}

// OpenInEditor opens a file in the user's preferred editor
func OpenInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}

	// Validate editor exists and is executable
	if err := utils.ValidateEditor(editor); err != nil {
		return fmt.Errorf("editor validation failed: %w", err)
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetEditorCmd returns a command to open a file in the editor
func GetEditorCmd(path string) *exec.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}

	// Note: We can't validate here since this is used in TUI context
	// Validation should be done at startup or in a separate function
	cmd := exec.Command(editor, path)
	return cmd
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return utils.EnsureDir(path)
}

// ReadNote reads a note file
func ReadNote(path string) (*Note, error) {
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

// WriteNote writes content to a note file
func WriteNote(path string, content string) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return err
	}
	return utils.AtomicWriteFile(path, []byte(content), 0644)
}

// FindNotes finds all markdown files in a directory recursively
func FindNotes(dir string) ([]string, error) {
	var notes []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			notes = append(notes, path)
		}
		return nil
	})

	return notes, err
}

// SearchNotes searches for notes containing a query
func SearchNotes(dir string, query string) ([]string, error) {
	allNotes, err := FindNotes(dir)
	if err != nil {
		return nil, err
	}

	var matches []string
	query = strings.ToLower(query)

	for _, notePath := range allNotes {
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

// BuildDailyNotePath builds the path for a daily note
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

// CreateDailyNote creates a daily note with template
func CreateDailyNote(notePath string, sections []string, date time.Time) error {
	dir := filepath.Dir(notePath)
	if err := EnsureDir(dir); err != nil {
		return err
	}

	content := fmt.Sprintf("# %s\n\n", date.Format("2006-01-02-Mon"))

	for _, section := range sections {
		content += fmt.Sprintf("## %s\n\n", section)
	}

	return utils.AtomicWriteFile(notePath, []byte(content), 0644)
}

// GetRecentDailyNotes gets the most recent daily notes
func GetRecentDailyNotes(diaryDir string, count int) ([]string, error) {
	var notes []string
	today := time.Now()

	for i := 0; i < count; i++ {
		date := today.AddDate(0, 0, -i)
		notePath := BuildDailyNotePath(diaryDir, date)
		if FileExists(notePath) {
			notes = append(notes, notePath)
		}
	}

	return notes, nil
}

