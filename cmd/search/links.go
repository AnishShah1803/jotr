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
	Use:   "links",
	Short: "Show links and backlinks",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var linksOutgoingCmd = &cobra.Command{
	Use:   "outgoing <note-name>",
	Short: "Show outgoing wiki-links from a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return showLinks(cmd.Context(), cfg, args[0])
	},
}

var linksBacklinksCmd = &cobra.Command{
	Use:   "backlinks <note-name>",
	Short: "Show notes that link to a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return showBacklinks(cmd.Context(), cfg, args[0])
	},
}

var linksUnresolvedCmd = &cobra.Command{
	Use:   "unresolved",
	Short: "List all unresolved (broken) wiki-links",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return showUnresolvedLinks(cmd.Context(), cfg)
	},
}

var linksOrphansCmd = &cobra.Command{
	Use:   "orphans",
	Short: "List notes with no incoming links",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return showOrphanedNotes(cmd.Context(), cfg)
	},
}

var linksDeadendsCmd = &cobra.Command{
	Use:   "deadends",
	Short: "List notes with no outgoing links",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return showDeadends(cmd.Context(), cfg)
	},
}

func init() {
	LinksCmd.AddCommand(linksOutgoingCmd)
	LinksCmd.AddCommand(linksBacklinksCmd)
	LinksCmd.AddCommand(linksUnresolvedCmd)
	LinksCmd.AddCommand(linksOrphansCmd)
	LinksCmd.AddCommand(linksDeadendsCmd)
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

					relPath := relPath(cfg.Paths.BaseDir, note)
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
			relPath := relPath(cfg.Paths.BaseDir, note)
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

	// Build a map from lowercase note name → absolute path for O(1) lookup.
	noteByName := make(map[string]string, len(allNotes))
	for _, note := range allNotes {
		name := strings.ToLower(strings.TrimSuffix(filepath.Base(note), ".md"))
		noteByName[name] = note
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
				if targetNote, ok := noteByName[strings.ToLower(match[1])]; ok {
					backlinkedNotes[targetNote] = true
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
		relPath := relPath(cfg.Paths.BaseDir, note)
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
		relPath := relPath(cfg.Paths.BaseDir, note)
		fmt.Printf("  %s\n", relPath)
	}

	return nil
}
