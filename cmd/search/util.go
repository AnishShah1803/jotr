package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// pickNoteByName finds a note matching query. If exactly one note matches,

// pickNoteByName finds a note matching query. If exactly one note matches,
// it is returned immediately. If multiple notes match, the user is prompted
// to select one interactively. Returns an error if no match is found.
func pickNoteByName(ctx context.Context, cfg *config.LoadedConfig, query string) (string, error) {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return "", fmt.Errorf("failed to find notes: %w", err)
	}

	if len(allNotes) == 0 {
		return "", fmt.Errorf("no notes found in vault")
	}

	ql := strings.ToLower(query)
	var matches []string

	for _, np := range allNotes {
		if strings.Contains(strings.ToLower(filepath.Base(np)), ql) {
			matches = append(matches, np)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("note not found: %s", query)
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	fmt.Println("Multiple notes found:")
	for i, np := range matches {
		rel := relPath(cfg.Paths.BaseDir, np)
		fmt.Printf("  %d. %s\n", i+1, rel)
	}

	if utils.IsReplMode() {
		return "", fmt.Errorf("cannot prompt in REPL mode: be more specific in your query")
	}

	fmt.Print("\nSelect note (1-", len(matches), "): ")

	input, err := utils.DefaultStdinReader.ReadString('\n')
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

// relPath returns target relative to base, falling back to target on error.
func relPath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}
