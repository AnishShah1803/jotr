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
	"github.com/AnishShah1803/jotr/internal/search"
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
	if query == "" {
		return nil
	}

	indexPath := search.GetIndexPath(cfg.Paths.BaseDir)

	if _, err := os.Stat(indexPath); err == nil {
		idx, err := search.Open(indexPath)
		if err == nil {
			defer idx.Close()

			results, err := idx.Search(ctx, query, 50)
			if err == nil && len(results) > 0 {
				if GetSearchCountForTest() || searchOutputOption.CountOnly {
					fmt.Printf("%d matches found\n", len(results))
					return nil
				}

				if GetSearchFilesForTest() || searchOutputOption.FilesOnly {
					for _, r := range results {
						relPath, _ := filepath.Rel(cfg.Paths.BaseDir, r.Path)
						fmt.Println(relPath)
					}
					return nil
				}

				fmt.Printf("Found %d matches:\n\n", len(results))

				for _, r := range results {
					relPath, _ := filepath.Rel(cfg.Paths.BaseDir, r.Path)
					fmt.Printf("📄 %s", relPath)
					if r.Title != "" {
						fmt.Printf(" (%s)", r.Title)
					}
					fmt.Println()

					if r.Snippet != "" {
						fmt.Printf("   %s\n", r.Snippet)
					}
					fmt.Println()
				}
				return nil
			}
		}
	}

	matches, err := notes.SearchNotes(ctx, cfg.Paths.BaseDir, query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(matches) == 0 {
		fmt.Println("No matches found")
		return nil
	}

	if GetSearchCountForTest() || searchOutputOption.CountOnly {
		fmt.Printf("%d matches found\n", len(matches))
		return nil
	}

	if GetSearchFilesForTest() || searchOutputOption.FilesOnly {
		for _, match := range matches {
			relPath, _ := filepath.Rel(cfg.Paths.BaseDir, match.Path)
			fmt.Println(relPath)
		}
		return nil
	}

	fmt.Printf("Found %d matches:\n\n", len(matches))

	queryLower := strings.ToLower(query)

	for _, match := range matches {
		relPath, _ := filepath.Rel(cfg.Paths.BaseDir, match.Path)
		fmt.Printf("📄 %s\n", relPath)

		lines := strings.Split(string(match.Content), "\n")

		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), queryLower) {
				lineNum := i + 1
				highlighted := strings.ReplaceAll(line, query, fmt.Sprintf("**%s**", query))
				fmt.Printf("  %d: %s\n", lineNum, highlighted)
			}
		}

		fmt.Println()
	}

	return nil
}
