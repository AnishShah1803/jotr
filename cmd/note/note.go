package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// Reader interface for testable input operations.
type Reader interface {
	ReadString(delim byte) (string, error)
}

// StdinReader wraps os.Stdin for production use.
type StdinReader struct{}

func (r StdinReader) ReadString(delim byte) (string, error) {
	return bufio.NewReader(os.Stdin).ReadString(delim)
}

// defaultReader is the production reader using stdin.
var defaultReader Reader = StdinReader{}

// SetReader allows replacing the reader (for testing).
func SetReader(r Reader) {
	defaultReader = r
}

var NoteCmd = &cobra.Command{
	Use:   "note [action]",
	Short: "Create, open, or manage notes",
	Long: `Manage notes with various actions.
	
Actions:
  create [type]     Create a new note
  open [query]      Open an existing note
  list              List all notes
  
Examples:
  jotr note create           # Create new note
  jotr note create work      # Create note in work folder
  jotr note open MyNote      # Open note by name
  jotr note list             # List all notes`,
	Aliases: []string{"n"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("action required: create, open, or list")
		}

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		action := args[0]

		switch action {
		case "create":
			noteType := ""
			if len(args) > 1 {
				noteType = args[1]
			}
			return createNote(cmd.Context(), cfg, noteType)
		case "open":
			query := ""
			if len(args) > 1 {
				query = args[1]
			}
			return openNote(cmd.Context(), cfg, query)
		case "list":
			return listNotes(cmd.Context(), cfg)
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
}

func createNote(ctx context.Context, cfg *config.LoadedConfig, noteType string) error {
	return createNoteWithReader(ctx, cfg, noteType, defaultReader)
}

func createNoteWithReader(ctx context.Context, cfg *config.LoadedConfig, noteType string, reader Reader) error {
	// Prompt for note name
	fmt.Print("Note name: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read note name: %w", err)
	}
	name := strings.TrimSpace(input)

	if name == "" {
		return fmt.Errorf("note name is required")
	}

	var notePath string
	if noteType != "" {
		notePath = filepath.Join(cfg.Paths.BaseDir, noteType, name+".md")
	} else {
		notePath = filepath.Join(cfg.Paths.BaseDir, name+".md")
	}

	if utils.FileExists(notePath) {
		return fmt.Errorf("note already exists: %s", notePath)
	}

	content := fmt.Sprintf("# %s\n\n", name)

	if err := notes.WriteNote(ctx, notePath, content); err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	fmt.Printf("✓ Created: %s\n", notePath)

	return notes.OpenInEditor(notePath)
}

func openNote(ctx context.Context, cfg *config.LoadedConfig, query string) error {
	return openNoteWithReader(ctx, cfg, query, defaultReader)
}

func openNoteWithReader(ctx context.Context, cfg *config.LoadedConfig, query string, reader Reader) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return fmt.Errorf("failed to find notes: %w", err)
	}

	if len(allNotes) == 0 {
		return fmt.Errorf("no notes found")
	}

	// If query provided, filter notes
	var matches []string

	if query != "" {
		query = strings.ToLower(query)

		for _, notePath := range allNotes {
			name := strings.ToLower(filepath.Base(notePath))
			if strings.Contains(name, query) {
				matches = append(matches, notePath)
			}
		}
	} else {
		matches = allNotes
	}

	if len(matches) == 0 {
		return fmt.Errorf("no notes found matching: %s", query)
	}

	// If single match, open it
	if len(matches) == 1 {
		return notes.OpenInEditor(matches[0])
	}

	// Multiple matches - show list and prompt
	fmt.Println("Multiple notes found:")

	for i, notePath := range matches {
		relPath, _ := filepath.Rel(cfg.Paths.BaseDir, notePath)
		fmt.Printf("%d. %s\n", i+1, relPath)
	}

	fmt.Print("\nSelect note (1-", len(matches), "): ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read selection: %w", err)
	}
	trimmed := strings.TrimSpace(input)

	var selection int
	if _, err := fmt.Sscanf(trimmed, "%d", &selection); err != nil {
		return fmt.Errorf("failed to parse selection: %w", err)
	}

	if selection < 1 || selection > len(matches) {
		return fmt.Errorf("invalid selection")
	}

	return notes.OpenInEditor(matches[selection-1])
}

func listNotes(ctx context.Context, cfg *config.LoadedConfig) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return fmt.Errorf("failed to find notes: %w", err)
	}

	if len(allNotes) == 0 {
		fmt.Println("No notes found")
		return nil
	}

	fmt.Printf("Found %d notes:\n\n", len(allNotes))

	for _, notePath := range allNotes {
		relPath, _ := filepath.Rel(cfg.Paths.BaseDir, notePath)
		fmt.Printf("  %s\n", relPath)
	}

	return nil
}
