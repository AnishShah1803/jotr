package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var AliasesCmd = &cobra.Command{
	Use:   "aliases [action]",
	Short: "Manage note aliases",
	Long: `Manage and track note aliases for cross-referencing.
	
Actions:
  list              List all aliases and their notes
  find [alias]      Find notes using a specific alias
  stats             Show alias statistics
  
Aliases can be defined in frontmatter as 'aliases: [alias1, alias2]' or 'alias: alias1'.
  
Examples:
  jotr aliases list
  jotr aliases find shortname
  jotr aliases stats`,
	Aliases: []string{"alias"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		action := "list"
		if len(args) > 0 {
			action = args[0]
		}

		switch action {
		case "list":
			return listAliases(cmd.Context(), cfg)
		case "find":
			if len(args) < 2 {
				return fmt.Errorf("alias name required")
			}
			return findByAlias(cmd.Context(), cfg, args[1])
		case "stats":
			return aliasStats(cmd.Context(), cfg)
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
}

func extractAliases(content string) []string {
	lines := strings.Split(content, "\n")

	if len(lines) < 3 || lines[0] != "---" {
		return nil
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return nil
	}

	var aliases []string
	aliasRegex := regexp.MustCompile(`^\s*aliases?\s*:\s*(.+)$`)
	arrayRegex := regexp.MustCompile(`\[([^\]]+)\]`)

	for i := 1; i < endIdx; i++ {
		matches := aliasRegex.FindStringSubmatch(lines[i])
		if len(matches) > 1 {
			value := strings.TrimSpace(matches[1])

			// Handle array format: [alias1, alias2]
			if arrayMatches := arrayRegex.FindStringSubmatch(value); len(arrayMatches) > 1 {
				items := strings.Split(arrayMatches[1], ",")
				for _, item := range items {
					cleaned := strings.TrimSpace(item)
					cleaned = strings.Trim(cleaned, `"'`)
					if cleaned != "" {
						aliases = append(aliases, cleaned)
					}
				}
			} else {
				// Single value
				value = strings.Trim(value, `"'`)
				if value != "" {
					aliases = append(aliases, value)
				}
			}
		}
	}

	return aliases
}

func listAliases(ctx context.Context, cfg *config.LoadedConfig) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	aliasMap := make(map[string][]string) // alias -> [note paths]

	results, err := utils.ProcessFilesParallelWithContent(ctx, allNotes, 0, func(path string, content []byte) bool {
		return true
	})
	if err != nil {
		return err
	}

	for _, r := range results {
		aliases := extractAliases(string(r.Content))
		for _, alias := range aliases {
			aliasMap[alias] = append(aliasMap[alias], r.Path)
		}
	}

	if len(aliasMap) == 0 {
		fmt.Println("No aliases found")
		return nil
	}

	keys := make([]string, 0, len(aliasMap))
	for k := range aliasMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("Found %d aliases:\n\n", len(aliasMap))

	for _, alias := range keys {
		paths := aliasMap[alias]
		fmt.Printf("  %s -> ", alias)

		if len(paths) == 1 {
			relPath, _ := filepath.Rel(cfg.Paths.BaseDir, paths[0])
			fmt.Printf("%s\n", relPath)
		} else {
			fmt.Printf("%d notes\n", len(paths))
			for _, path := range paths {
				relPath, _ := filepath.Rel(cfg.Paths.BaseDir, path)
				fmt.Printf("    - %s\n", relPath)
			}
		}
	}

	return nil
}

func findByAlias(ctx context.Context, cfg *config.LoadedConfig, alias string) error {
	alias = strings.TrimSpace(alias)

	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	var matches []string

	results, err := utils.ProcessFilesParallelWithContent(ctx, allNotes, 0, func(path string, content []byte) bool {
		return true
	})
	if err != nil {
		return err
	}

	for _, r := range results {
		aliases := extractAliases(string(r.Content))
		for _, a := range aliases {
			if strings.EqualFold(a, alias) {
				matches = append(matches, r.Path)
				break
			}
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No notes found with alias: %s\n", alias)
		return nil
	}

	fmt.Printf("Found %d notes with alias '%s':\n\n", len(matches), alias)

	for _, match := range matches {
		relPath, _ := filepath.Rel(cfg.Paths.BaseDir, match)
		fmt.Printf("  %s\n", relPath)
	}

	return nil
}

func aliasStats(ctx context.Context, cfg *config.LoadedConfig) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	aliasCounts := make(map[string]int)
	notesWithAliases := 0

	results, err := utils.ProcessFilesParallelWithContent(ctx, allNotes, 0, func(path string, content []byte) bool {
		return true
	})
	if err != nil {
		return err
	}

	for _, r := range results {
		aliases := extractAliases(string(r.Content))
		if len(aliases) > 0 {
			notesWithAliases++
			for _, alias := range aliases {
				aliasCounts[alias]++
			}
		}
	}

	if len(aliasCounts) == 0 {
		fmt.Println("No aliases found")
		return nil
	}

	type aliasCount struct {
		alias string
		count int
	}

	var sorted []aliasCount
	for alias, count := range aliasCounts {
		sorted = append(sorted, aliasCount{alias, count})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	fmt.Println("Alias Statistics:")
	fmt.Printf("\nNotes with aliases: %d\n", notesWithAliases)
	fmt.Printf("Total unique aliases: %d\n\n", len(aliasCounts))

	for _, ac := range sorted {
		fmt.Printf("  %-20s %d note(s)\n", ac.alias, ac.count)
	}

	return nil
}
