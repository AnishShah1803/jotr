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
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var AliasCmd = &cobra.Command{
	Use:   "alias [action]",
	Short: "Manage note aliases",
	Long: `Create aliases for notes for quick access.
	
Actions:
  add [name] [target]    Add an alias
  remove [name]          Remove an alias
  list                   List all aliases
  resolve [name]         Resolve an alias
  
Examples:
  jotr alias add work "Work/Projects.md"
  jotr alias add today "daily:0"
  jotr alias add yesterday "daily:-1"
  jotr alias list
  jotr alias resolve work`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("action required: add, remove, list, or resolve")
		}

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		action := args[0]

		switch action {
		case "add":
			if len(args) < 3 {
				return fmt.Errorf("usage: alias add [name] [target]")
			}
			return addAlias(cfg, args[1], args[2])
		case "remove", "rm":
			if len(args) < 2 {
				return fmt.Errorf("usage: alias remove [name]")
			}
			return removeAlias(cfg, args[1])
		case "list", "ls":
			return listAliases(cfg)
		case "resolve":
			if len(args) < 2 {
				return fmt.Errorf("usage: alias resolve [name]")
			}
			return resolveAlias(cfg, args[1])
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
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

	return os.WriteFile(aliasFile, data, 0644)
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
	// Handle dynamic daily aliases: daily:0, daily:-1, etc.
	if strings.HasPrefix(value, "daily:") {
		offsetStr := strings.TrimPrefix(value, "daily:")

		var offset int

		fmt.Sscanf(offsetStr, "%d", &offset)

		date := time.Now().AddDate(0, 0, offset)
		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)

		return notePath, nil
	}

	// Regular path - resolve relative to base
	if !filepath.IsAbs(value) {
		return filepath.Join(cfg.Paths.BaseDir, value), nil
	}

	return value, nil
}
