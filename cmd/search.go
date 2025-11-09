package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/spf13/cobra"
)

var (
	searchCount bool
	searchFiles bool
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search across all notes",
	Long: `Search for text across all notes.
	
Examples:
  jotr search "meeting notes"    # Search for text
  jotr search --count "TODO"     # Count matches
  jotr search --files "project"  # Show only filenames`,
	Aliases: []string{"find", "grep"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("search query required")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		query := strings.Join(args, " ")
		return searchNotes(cfg, query)
	},
}

func init() {
	searchCmd.Flags().BoolVar(&searchCount, "count", false, "Show count only")
	searchCmd.Flags().BoolVar(&searchFiles, "files", false, "Show filenames only")
	rootCmd.AddCommand(searchCmd)
}

func searchNotes(cfg *config.LoadedConfig, query string) error {
	matches, err := notes.SearchNotes(cfg.Paths.BaseDir, query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(matches) == 0 {
		fmt.Println("No matches found")
		return nil
	}

	// Count only
	if searchCount {
		fmt.Printf("%d matches found\n", len(matches))
		return nil
	}

	// Files only
	if searchFiles {
		for _, match := range matches {
			relPath, _ := filepath.Rel(cfg.Paths.BaseDir, match)
			fmt.Println(relPath)
		}
		return nil
	}

	// Full output with context
	fmt.Printf("Found %d matches:\n\n", len(matches))
	
	for _, match := range matches {
		relPath, _ := filepath.Rel(cfg.Paths.BaseDir, match)
		fmt.Printf("📄 %s\n", relPath)

		// Read file and show matching lines
		content, err := os.ReadFile(match)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		queryLower := strings.ToLower(query)

		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), queryLower) {
				lineNum := i + 1
				// Highlight the match (simple version)
				highlighted := strings.ReplaceAll(line, query, fmt.Sprintf("**%s**", query))
				fmt.Printf("  %d: %s\n", lineNum, highlighted)
			}
		}
		fmt.Println()
	}

	return nil
}

