package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/search"
)

var linkRe = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

var LinksCmd = &cobra.Command{
	Use:   "links [note-name]",
	Short: "Show links and backlinks",
	Long: `Show links in a note and backlinks to a note.
	
Examples:
  jotr links MyNote          # Show links in MyNote
  jotr links --backlinks MyNote  # Show backlinks to MyNote`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("note name required")
		}

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		noteName := args[0]
		backlinks, _ := cmd.Flags().GetBool("backlinks")

		if backlinks {
			return showBacklinks(cmd.Context(), cfg, noteName)
		}
		return showLinks(cmd.Context(), cfg, noteName)
	},
}

func showLinks(ctx context.Context, cfg *config.LoadedConfig, noteName string) error {
	var allNotes []string
	indexPath := search.GetIndexPath(cfg.Paths.BaseDir)
	if _, err := os.Stat(indexPath); err == nil {
		if idx, err := search.Open(indexPath); err == nil {
			defer idx.Close()
			if paths, err := idx.GetIndexedPaths(ctx); err == nil {
				for p := range paths {
					allNotes = append(allNotes, p)
				}
			}
		}
	}

	if len(allNotes) == 0 {
		var err error
		allNotes, err = notes.FindNotes(ctx, cfg.Paths.BaseDir)
		if err != nil {
			return err
		}
	}

	var targetNote string

	for _, note := range allNotes {
		if strings.Contains(strings.ToLower(filepath.Base(note)), strings.ToLower(noteName)) {
			targetNote = note
			break
		}
	}

	if targetNote == "" {
		return fmt.Errorf("note not found: %s", noteName)
	}

	// Read note content
	content, err := os.ReadFile(targetNote)
	if err != nil {
		return err
	}

	matches := linkRe.FindAllStringSubmatch(string(content), -1)

	if len(matches) == 0 {
		fmt.Printf("No links found in %s\n", filepath.Base(targetNote))
		return nil
	}

	fmt.Printf("Links in %s:\n\n", filepath.Base(targetNote))

	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			link := match[1]
			if !seen[link] {
				fmt.Printf("  [[%s]]\n", link)

				seen[link] = true
			}
		}
	}

	return nil
}

func showBacklinks(ctx context.Context, cfg *config.LoadedConfig, noteName string) error {
	var allNotes []string
	indexPath := search.GetIndexPath(cfg.Paths.BaseDir)
	if _, err := os.Stat(indexPath); err == nil {
		if idx, err := search.Open(indexPath); err == nil {
			defer idx.Close()
			if results, err := idx.Search(ctx, noteName, 0); err == nil {
				for _, r := range results {
					allNotes = append(allNotes, r.Path)
				}
			}
		}
	}

	if len(allNotes) == 0 {
		var err error
		allNotes, err = notes.FindNotes(ctx, cfg.Paths.BaseDir)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Finding backlinks to '%s'...\n\n", noteName)

	found := false

	for _, note := range allNotes {
		content, err := os.ReadFile(note)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			matches := linkRe.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 && strings.Contains(strings.ToLower(match[1]), strings.ToLower(noteName)) {
					if !found {
						fmt.Println("Backlinks found:")

						found = true
					}

					relPath, _ := filepath.Rel(cfg.Paths.BaseDir, note)
					fmt.Printf("\n  %s:%d\n", relPath, i+1)
					fmt.Printf("    %s\n", strings.TrimSpace(line))
				}
			}
		}
	}

	if !found {
		fmt.Printf("No backlinks found for '%s'\n", noteName)
	}

	return nil
}
