package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags [action]",
	Short: "Manage tags (list, find, stats)",
	Long: `Manage tags across all notes.
	
Actions:
  list              List all tags
  find [tag]        Find notes with tag
  stats             Show tag statistics
  
Examples:
  jotr tags list
  jotr tags find meeting
  jotr tags stats`,
	Aliases: []string{"tag"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		action := "list"
		if len(args) > 0 {
			action = args[0]
		}

		switch action {
		case "list":
			return listTags(cfg)
		case "find":
			if len(args) < 2 {
				return fmt.Errorf("tag name required")
			}
			return findByTag(cfg, args[1])
		case "stats":
			return tagStats(cfg)
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}

func extractTags(content string) []string {
	// Match #tag pattern
	re := regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	tagSet := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			tagSet[match[1]] = true
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags
}

func listTags(cfg *config.LoadedConfig) error {
	allNotes, err := notes.FindNotes(cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	tagSet := make(map[string]bool)

	for _, notePath := range allNotes {
		content, err := os.ReadFile(notePath)
		if err != nil {
			continue
		}

		tags := extractTags(string(content))
		for _, tag := range tags {
			tagSet[tag] = true
		}
	}

	if len(tagSet) == 0 {
		fmt.Println("No tags found")
		return nil
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	fmt.Printf("Found %d tags:\n\n", len(tags))
	for _, tag := range tags {
		fmt.Printf("  #%s\n", tag)
	}

	return nil
}

func findByTag(cfg *config.LoadedConfig, tag string) error {
	allNotes, err := notes.FindNotes(cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	// Remove # if provided
	tag = strings.TrimPrefix(tag, "#")

	var matches []string

	for _, notePath := range allNotes {
		content, err := os.ReadFile(notePath)
		if err != nil {
			continue
		}

		tags := extractTags(string(content))
		for _, t := range tags {
			if t == tag {
				matches = append(matches, notePath)
				break
			}
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No notes found with tag: #%s\n", tag)
		return nil
	}

	fmt.Printf("Found %d notes with #%s:\n\n", len(matches), tag)
	for _, match := range matches {
		relPath, _ := filepath.Rel(cfg.Paths.BaseDir, match)
		fmt.Printf("  %s\n", relPath)
	}

	return nil
}

func tagStats(cfg *config.LoadedConfig) error {
	allNotes, err := notes.FindNotes(cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	tagCounts := make(map[string]int)

	for _, notePath := range allNotes {
		content, err := os.ReadFile(notePath)
		if err != nil {
			continue
		}

		tags := extractTags(string(content))
		for _, tag := range tags {
			tagCounts[tag]++
		}
	}

	if len(tagCounts) == 0 {
		fmt.Println("No tags found")
		return nil
	}

	// Sort by count
	type tagCount struct {
		tag   string
		count int
	}
	var sorted []tagCount
	for tag, count := range tagCounts {
		sorted = append(sorted, tagCount{tag, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	fmt.Println("Tag Statistics:\n")
	for _, tc := range sorted {
		fmt.Printf("  #%-20s %d notes\n", tc.tag, tc.count)
	}

	return nil
}

