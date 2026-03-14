package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

func renameNote(ctx context.Context, cfg *config.LoadedConfig, args []string) error {
	return renameNoteWithReader(ctx, cfg, args, defaultReader)
}

func renameNoteWithReader(ctx context.Context, cfg *config.LoadedConfig, args []string, reader utils.StdinReader) error {
	if utils.IsReplMode() && len(args) < 2 {
		return fmt.Errorf("usage: note rename <query> <new-name>")
	}

	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	targetNote, err := pickNote(ctx, cfg, query, reader)
	if err != nil {
		return err
	}

	var newName string
	if len(args) > 1 {
		newName = strings.TrimSpace(args[1])
	} else {
		fmt.Printf("Renaming: %s\n", filepath.Base(targetNote))
		fmt.Print("New name (without extension): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read new name: %w", err)
		}
		newName = strings.TrimSpace(input)
	}
	if newName == "" {
		return fmt.Errorf("new name is required")
	}
	if !strings.HasSuffix(newName, ".md") {
		newName += ".md"
	}

	newPath := filepath.Join(filepath.Dir(targetNote), newName)
	if utils.FileExists(newPath) {
		return fmt.Errorf("a note already exists at: %s", newPath)
	}

	if err := os.Rename(targetNote, newPath); err != nil {
		return fmt.Errorf("failed to rename note: %w", err)
	}

	fmt.Printf("✓ Renamed to: %s\n", newPath)
	return nil
}

func deleteNote(ctx context.Context, cfg *config.LoadedConfig, args []string) error {
	return deleteNoteWithReader(ctx, cfg, args, defaultReader)
}

func deleteNoteWithReader(ctx context.Context, cfg *config.LoadedConfig, args []string, reader utils.StdinReader) error {
	force := false
	query := ""

	for _, arg := range args {
		if arg == "--force" {
			force = true
		} else if query == "" {
			query = arg
		}
	}

	if utils.IsReplMode() {
		if !force {
			return fmt.Errorf("note delete requires --force flag in REPL mode\nUsage: note delete <query> --force")
		}
		if query == "" {
			return fmt.Errorf("note delete requires a <query> argument\nUsage: note delete <query> --force")
		}
	}

	targetNote, err := pickNote(ctx, cfg, query, reader)
	if err != nil {
		return err
	}

	relPath, _ := filepath.Rel(cfg.Paths.BaseDir, targetNote)

	if !force {
		fmt.Printf("Delete %s? [y/N]: ", relPath)

		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		confirm := strings.TrimSpace(strings.ToLower(input))
		if confirm != "y" && confirm != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	if err := os.Remove(targetNote); err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	fmt.Printf("✓ Deleted: %s\n", relPath)
	return nil
}

func moveNote(ctx context.Context, cfg *config.LoadedConfig, args []string) error {
	return moveNoteWithReader(ctx, cfg, args, defaultReader)
}

func moveNoteWithReader(ctx context.Context, cfg *config.LoadedConfig, args []string, reader utils.StdinReader) error {
	if utils.IsReplMode() && len(args) < 2 {
		return fmt.Errorf("usage: note move <query> <destination>")
	}

	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	targetNote, err := pickNote(ctx, cfg, query, reader)
	if err != nil {
		return err
	}

	var dest string
	if len(args) > 1 {
		dest = strings.TrimSpace(args[1])
	} else {
		fmt.Printf("Moving: %s\n", filepath.Base(targetNote))
		fmt.Printf("Destination folder (relative to vault root, e.g. work): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read destination: %w", err)
		}
		dest = strings.TrimSpace(input)
	}
	if dest == "" {
		return fmt.Errorf("destination folder is required")
	}

	destDir := filepath.Join(cfg.Paths.BaseDir, dest)
	if err := notes.EnsureDir(destDir); err != nil {
		return fmt.Errorf("failed to create destination folder: %w", err)
	}

	newPath := filepath.Join(destDir, filepath.Base(targetNote))
	if utils.FileExists(newPath) {
		return fmt.Errorf("a note already exists at: %s", newPath)
	}

	if err := os.Rename(targetNote, newPath); err != nil {
		return fmt.Errorf("failed to move note: %w", err)
	}

	relNew, _ := filepath.Rel(cfg.Paths.BaseDir, newPath)
	fmt.Printf("✓ Moved to: %s\n", relNew)
	return nil
}

func pickNote(ctx context.Context, cfg *config.LoadedConfig, query string, reader utils.StdinReader) (string, error) {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return "", fmt.Errorf("failed to find notes: %w", err)
	}
	if len(allNotes) == 0 {
		return "", fmt.Errorf("no notes found")
	}

	var matches []string
	if query != "" {
		ql := strings.ToLower(query)
		for _, p := range allNotes {
			if strings.Contains(strings.ToLower(filepath.Base(p)), ql) {
				matches = append(matches, p)
			}
		}
	} else {
		matches = allNotes
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no notes found matching: %s", query)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}

	if utils.IsReplMode() {
		fmt.Println("Multiple notes found:")
		for i, p := range matches {
			rel, _ := filepath.Rel(cfg.Paths.BaseDir, p)
			fmt.Printf("%d. %s\n", i+1, rel)
		}
		return "", fmt.Errorf("multiple matches found, please be more specific")
	}

	fmt.Println("Multiple notes found:")
	for i, p := range matches {
		rel, _ := filepath.Rel(cfg.Paths.BaseDir, p)
		fmt.Printf("%d. %s\n", i+1, rel)
	}
	fmt.Print("\nSelect note (1-", len(matches), "): ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}
	var sel int
	if _, err := fmt.Sscanf(strings.TrimSpace(input), "%d", &sel); err != nil {
		return "", fmt.Errorf("invalid selection")
	}
	if sel < 1 || sel > len(matches) {
		return "", fmt.Errorf("selection out of range")
	}
	return matches[sel-1], nil
}
