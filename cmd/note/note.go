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
	Use:     "note",
	Short:   "Create, open, or manage notes",
	Aliases: []string{"n"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var noteCreateCmd = &cobra.Command{
	Use:   "create [type]",
	Short: "Create a new note",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		noteType := ""
		if len(args) > 0 {
			noteType = args[0]
		}
		return createNote(cmd.Context(), cfg, noteType)
	},
}

var noteOpenCmd = &cobra.Command{
	Use:   "open [query]",
	Short: "Open an existing note",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		query := ""
		if len(args) > 0 {
			query = args[0]
		}
		return openNote(cmd.Context(), cfg, query)
	},
}

var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return listNotes(cmd.Context(), cfg)
	},
}

var noteRenameCmd = &cobra.Command{
	Use:   "rename [query] [new-name]",
	Short: "Rename an existing note",
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return renameNote(cmd.Context(), cfg, args)
	},
}

var noteDeleteCmd = &cobra.Command{
	Use:   "delete [query]",
	Short: "Delete a note",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		permanent, _ := cmd.Flags().GetBool("permanent")
		force, _ := cmd.Flags().GetBool("force")

		passArgs := make([]string, 0, len(args)+2)
		passArgs = append(passArgs, args...)
		if permanent {
			passArgs = append(passArgs, "--permanent")
		}
		if force {
			passArgs = append(passArgs, "--force")
		}
		return deleteNote(cmd.Context(), cfg, passArgs)
	},
}

var noteMoveCmd = &cobra.Command{
	Use:   "move [query] [destination]",
	Short: "Move a note to a subfolder",
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return moveNote(cmd.Context(), cfg, args)
	},
}

var noteRandomCmd = &cobra.Command{
	Use:   "random",
	Short: "Open a random note",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return randomNote(cmd.Context(), cfg)
	},
}

var noteUniqueCmd = &cobra.Command{
	Use:   "unique",
	Short: "List notes with unique content sizes",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return uniqueNotes(cmd.Context(), cfg)
	},
}

func init() {
	noteDeleteCmd.Flags().Bool("permanent", false, "Permanently delete instead of moving to trash")
	noteDeleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")

	NoteCmd.AddCommand(
		noteCreateCmd,
		noteOpenCmd,
		noteListCmd,
		noteRenameCmd,
		noteDeleteCmd,
		noteMoveCmd,
		noteRandomCmd,
		noteUniqueCmd,
	)
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
