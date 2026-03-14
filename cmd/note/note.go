package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// defaultReader is the production reader using stdin.
var defaultReader = utils.DefaultStdinReader

var NoteCmd = &cobra.Command{
	Use:   "note [action]",
	Short: "Create, open, or manage notes",
	Long: `Manage notes with various actions.
	
Actions:
  create [type]     Create a new note
  open [query]      Open an existing note
  list              List all notes
  rename [query]    Rename an existing note
  delete [query]    Delete a note
  move [query]      Move a note to a subfolder
  random            Open a random note
  unique            List notes with unique properties
  
Examples:
  jotr note create           # Create new note
  jotr note create work      # Create note in work folder
  jotr note open MyNote      # Open note by name
  jotr note list             # List all notes
  jotr note rename MyNote    # Rename a note
  jotr note delete MyNote    # Delete a note
  jotr note move MyNote      # Move a note
  jotr note random           # Open a random note
  jotr note unique           # List notes with unique properties`,
	Aliases: []string{"n"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("action required: create, open, list, rename, delete, move, random, or unique")
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
		case "rename":
			return renameNote(cmd.Context(), cfg, args[1:])
		case "delete":
			return deleteNote(cmd.Context(), cfg, args[1:])
		case "move":
			return moveNote(cmd.Context(), cfg, args[1:])
		case "random":
			return randomNote(cmd.Context(), cfg)
		case "unique":
			return uniqueNotes(cmd.Context(), cfg)
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
}

func createNote(ctx context.Context, cfg *config.LoadedConfig, noteType string) error {
	return createNoteWithReader(ctx, cfg, noteType, defaultReader)
}

func createNoteWithReader(ctx context.Context, cfg *config.LoadedConfig, noteType string, reader utils.StdinReader) error {
	// Prompt for note name
	if utils.IsReplMode() {
		return fmt.Errorf("cannot prompt in REPL mode: provide note name as argument (e.g., 'note create [name]')")
	}

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
	return openNoteWithReader(ctx, cfg, query, utils.DefaultStdinReader)
}

func openNoteWithReader(ctx context.Context, cfg *config.LoadedConfig, query string, reader utils.StdinReader) error {
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

	// Multiple matches
	if utils.IsReplMode() {
		fmt.Println("Multiple notes found:")
		for i, notePath := range matches {
			relPath, _ := filepath.Rel(cfg.Paths.BaseDir, notePath)
			fmt.Printf("%d. %s\n", i+1, relPath)
		}
		return fmt.Errorf("cannot prompt in REPL mode: be more specific in your query")
	}

	// In interactive mode, show list and prompt
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

func randomNote(ctx context.Context, cfg *config.LoadedConfig) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return fmt.Errorf("failed to find notes: %w", err)
	}

	if len(allNotes) == 0 {
		return fmt.Errorf("no notes found")
	}

	randomIdx := rand.Intn(len(allNotes))
	return notes.OpenInEditor(allNotes[randomIdx])
}

func uniqueNotes(ctx context.Context, cfg *config.LoadedConfig) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	if len(allNotes) == 0 {
		fmt.Println("No notes found")
		return nil
	}

	sizeMap := make(map[int][]string)
	lineCounts := make(map[string]int)

	for _, note := range allNotes {
		content, err := os.ReadFile(note)
		if err != nil {
			continue
		}

		size := len(content)
		lineCount := strings.Count(string(content), "\n") + 1
		lineCounts[note] = lineCount

		sizeMap[size] = append(sizeMap[size], note)
	}

	var uniqueNotesList []string
	for _, notePaths := range sizeMap {
		if len(notePaths) == 1 {
			uniqueNotesList = append(uniqueNotesList, notePaths[0])
		}
	}

	if len(uniqueNotesList) == 0 {
		fmt.Println("No unique notes found (all notes have duplicate content sizes)")
		return nil
	}

	fmt.Printf("Found %d notes with unique properties:\n\n", len(uniqueNotesList))

	for _, note := range uniqueNotesList {
		relPath, _ := filepath.Rel(cfg.Paths.BaseDir, note)
		lineCount := lineCounts[note]
		fmt.Printf("  %s (%d lines)\n", relPath, lineCount)
	}

	return nil
}
