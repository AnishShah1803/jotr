package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
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

		cfg, err := config.Load()
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
			return createNote(cfg, noteType)
		case "open":
			query := ""
			if len(args) > 1 {
				query = args[1]
			}
			return openNote(cfg, query)
		case "list":
			return listNotes(cfg)
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
}

func init() {
	rootCmd.AddCommand(noteCmd)
}

func createNote(cfg *config.LoadedConfig, noteType string) error {
	// Prompt for note name
	fmt.Print("Note name: ")
	var name string
	fmt.Scanln(&name)
	
	if name == "" {
		return fmt.Errorf("note name is required")
	}

	// Build note path
	var notePath string
	if noteType != "" {
		notePath = filepath.Join(cfg.Paths.BaseDir, noteType, name+".md")
	} else {
		notePath = filepath.Join(cfg.Paths.BaseDir, name+".md")
	}

	// Check if note already exists
	if notes.FileExists(notePath) {
		return fmt.Errorf("note already exists: %s", notePath)
	}

	// Create note with basic template
	content := fmt.Sprintf("# %s\n\n", name)
	if err := notes.WriteNote(notePath, content); err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	fmt.Printf("✓ Created: %s\n", notePath)

	// Open in editor
	return notes.OpenInEditor(notePath)
}

func openNote(cfg *config.LoadedConfig, query string) error {
	// Find all notes
	allNotes, err := notes.FindNotes(cfg.Paths.BaseDir)
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
	var selection int
	fmt.Scanln(&selection)

	if selection < 1 || selection > len(matches) {
		return fmt.Errorf("invalid selection")
	}

	return notes.OpenInEditor(matches[selection-1])
}

func listNotes(cfg *config.LoadedConfig) error {
	allNotes, err := notes.FindNotes(cfg.Paths.BaseDir)
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

