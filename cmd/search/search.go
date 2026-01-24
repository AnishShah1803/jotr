package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/options"
)

var searchOutputOption = options.NewOutputOption()

var searchCmdFlags = struct {
	count bool
	files bool
}{}

func init() {
	searchOutputOption.AddFlags(SearchCmd)
}

func SetSearchCountForTest(count bool) {
	searchCmdFlags.count = count
}

func SetSearchFilesForTest(files bool) {
	searchCmdFlags.files = files
}

func GetSearchCountForTest() bool {
	return searchCmdFlags.count || searchOutputOption.CountOnly
}

func GetSearchFilesForTest() bool {
	return searchCmdFlags.files || searchOutputOption.FilesOnly
}

var SearchCmd = &cobra.Command{
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

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		query := strings.Join(args, " ")
		return SearchNotes(cmd.Context(), cfg, query)
	},
}

// SearchNotes performs a full-text search across all notes in the configured base directory.
// It displays matching files with highlighted context lines unless --count or --files flags are used.
func SearchNotes(ctx context.Context, cfg *config.LoadedConfig, query string) error {
	matches, err := notes.SearchNotes(ctx, cfg.Paths.BaseDir, query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(matches) == 0 {
		fmt.Println("No matches found")
		return nil
	}

	// Count only
	if GetSearchCountForTest() || searchOutputOption.CountOnly {
		fmt.Printf("%d matches found\n", len(matches))
		return nil
	}

	// Files only
	if GetSearchFilesForTest() || searchOutputOption.FilesOnly {
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
