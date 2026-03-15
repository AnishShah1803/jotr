package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var AliasCmd = &cobra.Command{
	Use:   "bookmark",
	Short: "Bookmark notes for quick access",
	Long: `Create bookmarks for notes for quick access.

Examples:
  jotr bookmark add work "Work/Projects.md"
  jotr bookmark add today "daily:0"
  jotr bookmark add yesterday "daily:-1"
  jotr bookmark list
  jotr bookmark resolve work`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var bookmarkAddCmd = &cobra.Command{
	Use:   "add [name] [target]",
	Short: "Add a bookmark",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return addAlias(cfg, args[0], args[1])
	},
}

var bookmarkRemoveCmd = &cobra.Command{
	Use:     "remove [name]",
	Short:   "Remove a bookmark",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return removeAlias(cfg, args[0])
	},
}

var bookmarkListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all bookmarks",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return listAliases(cfg)
	},
}

var bookmarkResolveCmd = &cobra.Command{
	Use:   "resolve [name]",
	Short: "Resolve a bookmark to its path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return resolveAlias(cfg, args[0])
	},
}

func init() {
	AliasCmd.AddCommand(bookmarkAddCmd, bookmarkRemoveCmd, bookmarkListCmd, bookmarkResolveCmd)
}

func getAliasFile(cfg *config.LoadedConfig) string {
	return filepath.Join(cfg.Paths.BaseDir, ".aliases.json")
}

func loadAliases(cfg *config.LoadedConfig) (map[string]string, error) {
	aliasFile := getAliasFile(cfg)

	if !utils.FileExists(aliasFile) {
		return make(map[string]string), nil
	}

	data, err := os.ReadFile(aliasFile)
	if err != nil {
		return nil, err
	}

	var aliases map[string]string
	if err := json.Unmarshal(data, &aliases); err != nil {
		return nil, err
	}

	return aliases, nil
}

func saveAliases(cfg *config.LoadedConfig, aliases map[string]string) error {
	aliasFile := getAliasFile(cfg)

	data, err := json.MarshalIndent(aliases, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(aliasFile, data, constants.FilePerm0644)
}

func addAlias(cfg *config.LoadedConfig, name, target string) error {
	aliases, err := loadAliases(cfg)
	if err != nil {
		return err
	}

	aliases[name] = target

	if err := saveAliases(cfg, aliases); err != nil {
		return err
	}

	fmt.Printf("✓ Added alias: %s → %s\n", name, target)

	return nil
}

func removeAlias(cfg *config.LoadedConfig, name string) error {
	aliases, err := loadAliases(cfg)
	if err != nil {
		return err
	}

	if _, exists := aliases[name]; !exists {
		return fmt.Errorf("alias not found: %s", name)
	}

	delete(aliases, name)

	if err := saveAliases(cfg, aliases); err != nil {
		return err
	}

	fmt.Printf("✓ Removed alias: %s\n", name)

	return nil
}

func listAliases(cfg *config.LoadedConfig) error {
	aliases, err := loadAliases(cfg)
	if err != nil {
		return err
	}

	if len(aliases) == 0 {
		fmt.Println("No aliases defined")
		return nil
	}

	fmt.Println("Aliases:")
	fmt.Println()

	for name, target := range aliases {
		resolved, _ := resolveAliasValue(cfg, target)
		if resolved != target {
			fmt.Printf("  %s → %s (%s)\n", name, target, resolved)
		} else {
			fmt.Printf("  %s → %s\n", name, target)
		}
	}

	return nil
}

func resolveAlias(cfg *config.LoadedConfig, name string) error {
	aliases, err := loadAliases(cfg)
	if err != nil {
		return err
	}

	target, exists := aliases[name]
	if !exists {
		return fmt.Errorf("alias not found: %s", name)
	}

	resolved, err := resolveAliasValue(cfg, target)
	if err != nil {
		return err
	}

	fmt.Printf("Alias: %s\n", name)
	fmt.Printf("Target: %s\n", target)
	fmt.Printf("Resolved: %s\n", resolved)

	return nil
}

func resolveAliasValue(cfg *config.LoadedConfig, value string) (string, error) {
	if strings.HasPrefix(value, "daily:") {
		offsetStr := strings.TrimPrefix(value, "daily:")

		var offset int

		fmt.Sscanf(offsetStr, "%d", &offset)

		date := time.Now().AddDate(0, 0, offset)
		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)

		return notePath, nil
	}

	if !filepath.IsAbs(value) {
		return filepath.Join(cfg.Paths.BaseDir, value), nil
	}

	return value, nil
}
