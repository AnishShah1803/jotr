package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/search"
)

var linkRe = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

var LinksCmd = &cobra.Command{
	Use:   "links [action] [note-name]",
	Short: "Show links and backlinks",
	Long: `Show links in a note and backlinks to a note.
	
Actions:
  outgoing [note]    Show outgoing wiki-links from a note (default)
  backlinks [note]   Show notes that link to the given note
  unresolved         List all unresolved (broken) wiki-links
  orphans            List all orphaned notes (notes with no incoming links)
  deadends           List all dead-end notes (notes with no outgoing links)
  
Examples:
  jotr links MyNote              # Show outgoing links in MyNote
  jotr links outgoing MyNote     # Explicit outgoing links
  jotr links backlinks MyNote    # Show backlinks to MyNote
  jotr links unresolved          # Show broken links
  jotr links orphans             # Show orphaned notes
  jotr links deadends            # Show notes with no outgoing links
  jotr links --backlinks MyNote  # Flag-based backlinks (legacy)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return fmt.Errorf("action required: outgoing, backlinks, unresolved, orphans, or deadends")
		}

		action := args[0]
		switch action {
		case "outgoing":
			if len(args) < 2 {
				return fmt.Errorf("note name required")
			}
			return showLinks(cmd.Context(), cfg, args[1])
		case "backlinks":
			if len(args) < 2 {
				return fmt.Errorf("note name required")
			}
			return showBacklinks(cmd.Context(), cfg, args[1])
		case "unresolved":
			return showUnresolvedLinks(cmd.Context(), cfg)
		case "orphans":
			return showOrphanedNotes(cmd.Context(), cfg)
		case "deadends":
			return showDeadends(cmd.Context(), cfg)
		default:
			backlinks, _ := cmd.Flags().GetBool("backlinks")
			if backlinks {
				return showBacklinks(cmd.Context(), cfg, action)
			}
			return showLinks(cmd.Context(), cfg, action)
		}
	},
}

func init() {
	LinksCmd.Flags().Bool("backlinks", false, "show backlinks instead of outgoing links")
}

func showLinks(ctx context.Context, cfg *config.LoadedConfig, noteName string) error {
	targetNote, err := pickNoteByName(ctx, cfg, noteName)
	if err != nil {
		return err
	}

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

func showUnresolvedLinks(ctx context.Context, cfg *config.LoadedConfig) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	noteNames := make(map[string]bool)
	for _, note := range allNotes {
		base := strings.TrimSuffix(filepath.Base(note), ".md")
		noteNames[strings.ToLower(base)] = true
	}

	unresolvedLinks := make(map[string][]string)

	for _, note := range allNotes {
		content, err := os.ReadFile(note)
		if err != nil {
			continue
		}

		matches := linkRe.FindAllStringSubmatch(string(content), -1)
		seen := make(map[string]bool)

		for _, match := range matches {
			if len(match) > 1 {
				link := match[1]
				if !seen[link] {
					linkLower := strings.ToLower(link)
					if !noteNames[linkLower] {
						unresolvedLinks[link] = append(unresolvedLinks[link], note)
					}
					seen[link] = true
				}
			}
		}
	}

	if len(unresolvedLinks) == 0 {
		fmt.Println("All links are resolved!")
		return nil
	}

	keys := make([]string, 0, len(unresolvedLinks))
	for k := range unresolvedLinks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("Found %d unresolved links:\n\n", len(unresolvedLinks))

	for _, link := range keys {
		notePaths := unresolvedLinks[link]
		fmt.Printf("  [[%s]] - referenced in %d note(s):\n", link, len(notePaths))
		for _, note := range notePaths {
			relPath, _ := filepath.Rel(cfg.Paths.BaseDir, note)
			fmt.Printf("    - %s\n", relPath)
		}
	}

	return nil
}

func showOrphanedNotes(ctx context.Context, cfg *config.LoadedConfig) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	backlinkedNotes := make(map[string]bool)

	for _, note := range allNotes {
		content, err := os.ReadFile(note)
		if err != nil {
			continue
		}

		matches := linkRe.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) > 1 {
				for _, targetNote := range allNotes {
					targetName := strings.TrimSuffix(filepath.Base(targetNote), ".md")
					if strings.EqualFold(match[1], targetName) {
						backlinkedNotes[targetNote] = true
						break
					}
				}
			}
		}
	}

	var orphans []string
	for _, note := range allNotes {
		if !backlinkedNotes[note] {
			orphans = append(orphans, note)
		}
	}

	if len(orphans) == 0 {
		fmt.Println("No orphaned notes found!")
		return nil
	}

	sort.Strings(orphans)

	fmt.Printf("Found %d orphaned notes (no incoming links):\n\n", len(orphans))

	for _, note := range orphans {
		relPath, _ := filepath.Rel(cfg.Paths.BaseDir, note)
		fmt.Printf("  %s\n", relPath)
	}

	return nil
}

func showDeadends(ctx context.Context, cfg *config.LoadedConfig) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	var deadends []string

	for _, note := range allNotes {
		content, err := os.ReadFile(note)
		if err != nil {
			continue
		}

		matches := linkRe.FindAllStringSubmatch(string(content), -1)
		if len(matches) == 0 {
			deadends = append(deadends, note)
		}
	}

	if len(deadends) == 0 {
		fmt.Println("No dead-end notes found!")
		return nil
	}

	sort.Strings(deadends)

	fmt.Printf("Found %d notes with no outgoing links:\n\n", len(deadends))

	for _, note := range deadends {
		relPath, _ := filepath.Rel(cfg.Paths.BaseDir, note)
		fmt.Printf("  %s\n", relPath)
	}

	return nil
}
