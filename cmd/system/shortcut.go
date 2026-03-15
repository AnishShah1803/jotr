package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var ShortcutCmd = &cobra.Command{
	Use:   "shortcut",
	Short: "Manage command shortcuts",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var shortcutAddCmd = &cobra.Command{
	Use:   "add [name] [command]",
	Short: "Add a shortcut",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return addShortcut(cfg, args[0], args[1])
	},
}

var shortcutRemoveCmd = &cobra.Command{
	Use:     "remove [name]",
	Short:   "Remove a shortcut",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return removeShortcut(cfg, args[0])
	},
}

var shortcutListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all shortcuts",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return listShortcuts(cfg)
	},
}

func init() {
	ShortcutCmd.AddCommand(shortcutAddCmd, shortcutRemoveCmd, shortcutListCmd)
}

func getShortcutFile(cfg *config.LoadedConfig) string {
	return filepath.Join(cfg.Paths.BaseDir, ".shortcuts.json")
}

func loadShortcuts(cfg *config.LoadedConfig) (map[string]string, error) {
	shortcutFile := getShortcutFile(cfg)

	if !utils.FileExists(shortcutFile) {
		return make(map[string]string), nil
	}

	data, err := os.ReadFile(shortcutFile)
	if err != nil {
		return nil, err
	}

	var shortcuts map[string]string
	if err := json.Unmarshal(data, &shortcuts); err != nil {
		return nil, err
	}

	return shortcuts, nil
}

func saveShortcuts(cfg *config.LoadedConfig, shortcuts map[string]string) error {
	shortcutFile := getShortcutFile(cfg)

	data, err := json.MarshalIndent(shortcuts, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(shortcutFile, data, constants.FilePerm0644)
}

var reservedCommands = []string{
	"daily", "note", "search", "tags", "capture", "summary", "sync",
	"template", "streak", "calendar", "bulk", "graph",
	"alias", "check", "configure", "validate", "shortcut", "schedule",
	"help", "version", "list", "stats", "archive", "git",
	"links", "frontmatter", "monthly",
}

func isReserved(name string) bool {
	for _, cmd := range reservedCommands {
		if name == cmd {
			return true
		}
	}

	return false
}

func addShortcut(cfg *config.LoadedConfig, name, command string) error {
	if isReserved(name) {
		return fmt.Errorf("cannot use reserved command name: %s", name)
	}

	shortcuts, err := loadShortcuts(cfg)
	if err != nil {
		return err
	}

	shortcuts[name] = command

	if err := saveShortcuts(cfg, shortcuts); err != nil {
		return err
	}

	fmt.Printf("✓ Added shortcut: %s → %s\n", name, command)

	return nil
}

func removeShortcut(cfg *config.LoadedConfig, name string) error {
	shortcuts, err := loadShortcuts(cfg)
	if err != nil {
		return err
	}

	if _, exists := shortcuts[name]; !exists {
		return fmt.Errorf("shortcut not found: %s", name)
	}

	delete(shortcuts, name)

	if err := saveShortcuts(cfg, shortcuts); err != nil {
		return err
	}

	fmt.Printf("✓ Removed shortcut: %s\n", name)

	return nil
}

func listShortcuts(cfg *config.LoadedConfig) error {
	shortcuts, err := loadShortcuts(cfg)
	if err != nil {
		return err
	}

	if len(shortcuts) == 0 {
		fmt.Println("No shortcuts defined")
		return nil
	}

	fmt.Println("Shortcuts:")
	fmt.Println()

	for name, command := range shortcuts {
		fmt.Printf("  %s → %s\n", name, command)
	}

	return nil
}
