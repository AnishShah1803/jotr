package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
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
	permanent := false
	force := false
	query := ""

	for _, arg := range args {
		if arg == "--permanent" {
			permanent = true
		} else if arg == "--force" {
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

	relPath := relPath(cfg.Paths.BaseDir, targetNote)

	if !force {
		if permanent {
			fmt.Printf("Permanently delete %s? [y/N]: ", relPath)
		} else {
			fmt.Printf("Move %s to trash? [y/N]: ", relPath)
		}

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

	if permanent {
		if err := os.Remove(targetNote); err != nil {
			return fmt.Errorf("failed to delete note: %w", err)
		}
		fmt.Printf("✓ Permanently deleted: %s\n", relPath)
	} else {
		trashDir := filepath.Join(cfg.Paths.BaseDir, ".trash")
		if err := notes.EnsureDir(trashDir); err != nil {
			return fmt.Errorf("failed to create trash directory: %w", err)
		}

		trashPath := filepath.Join(trashDir, filepath.Base(targetNote))
		if utils.FileExists(trashPath) {
			return fmt.Errorf("note already exists in trash: %s", filepath.Base(targetNote))
		}

		if err := os.Rename(targetNote, trashPath); err != nil {
			return fmt.Errorf("failed to move note to trash: %w", err)
		}
		fmt.Printf("✓ Moved to trash: %s\n", relPath)
	}
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

	if err := updateLinksAfterMove(ctx, cfg, targetNote, newPath); err != nil {
		fmt.Printf("Warning: failed to update links: %v\n", err)
	}

	relNew := relPath(cfg.Paths.BaseDir, newPath)
	fmt.Printf("✓ Moved to: %s\n", relNew)
	return nil
}

func updateLinksAfterMove(ctx context.Context, cfg *config.LoadedConfig, oldPath string, newPath string) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	oldName := strings.TrimSuffix(filepath.Base(oldPath), ".md")
	newName := strings.TrimSuffix(filepath.Base(newPath), ".md")
	oldRelDir, _ := filepath.Rel(cfg.Paths.BaseDir, filepath.Dir(oldPath))
	newRelDir, _ := filepath.Rel(cfg.Paths.BaseDir, filepath.Dir(newPath))

	wikiLinkPattern := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	mdLinkPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	for _, notePath := range allNotes {
		if notePath == newPath {
			continue
		}

		content, err := os.ReadFile(notePath)
		if err != nil {
			continue
		}

		updated := false
		text := string(content)

		text = wikiLinkPattern.ReplaceAllStringFunc(text, func(match string) string {
			link := strings.TrimPrefix(strings.TrimSuffix(match, "]]"), "[[")
			parts := strings.Split(link, "|")
			target := strings.TrimSpace(parts[0])

			if target == oldName || target == oldRelDir+"/"+oldName {
				updated = true
				newLink := newRelDir + "/" + newName
				if strings.Contains(link, "|") {
					return "[[" + newLink + "|" + strings.TrimSpace(parts[1]) + "]]"
				}
				return "[[" + newLink + "]]"
			}
			return match
		})

		text = mdLinkPattern.ReplaceAllStringFunc(text, func(match string) string {
			submatch := mdLinkPattern.FindStringSubmatch(match)
			if len(submatch) < 3 {
				return match
			}
			linkText := submatch[1]
			linkURL := submatch[2]

			if filepath.Base(linkURL) == oldName+".md" || filepath.Base(linkURL) == oldName {
				updated = true
				newURL := newRelDir + "/" + newName + ".md"
				return "[" + linkText + "](" + newURL + ")"
			}
			return match
		})

		if updated {
			if err := os.WriteFile(notePath, []byte(text), constants.FilePerm0644); err != nil {
				fmt.Printf("Warning: failed to update links in %s: %v\n", notePath, err)
			}
		}
	}

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
			rel := relPath(cfg.Paths.BaseDir, p)
			fmt.Printf("%d. %s\n", i+1, rel)
		}
		return "", fmt.Errorf("multiple matches found, please be more specific")
	}

	fmt.Println("Multiple notes found:")
	for i, p := range matches {
		rel := relPath(cfg.Paths.BaseDir, p)
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
