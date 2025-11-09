package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/spf13/cobra"
)

var shortcutCmd = &cobra.Command{
	Use:   "shortcut [action]",
	Short: "Manage command shortcuts",
	Long: `Create custom shortcuts for frequently used commands.
	
Actions:
  add [name] [command]    Add a shortcut
  remove [name]           Remove a shortcut
  list                    List all shortcuts
  
Examples:
  jotr shortcut add td "daily"
  jotr shortcut add ws "search work"
  jotr shortcut list
  jotr shortcut remove td`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("action required: add, remove, or list")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		action := args[0]

		switch action {
		case "add":
			if len(args) < 3 {
				return fmt.Errorf("usage: shortcut add [name] [command]")
			}
			return addShortcut(cfg, args[1], args[2])
		case "remove", "rm":
			if len(args) < 2 {
				return fmt.Errorf("usage: shortcut remove [name]")
			}
			return removeShortcut(cfg, args[1])
		case "list", "ls":
			return listShortcuts(cfg)
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
}

func init() {
	rootCmd.AddCommand(shortcutCmd)
}

func getShortcutFile(cfg *config.LoadedConfig) string {
	return filepath.Join(cfg.Paths.BaseDir, ".shortcuts.json")
}

func loadShortcuts(cfg *config.LoadedConfig) (map[string]string, error) {
	shortcutFile := getShortcutFile(cfg)
	
	if !notes.FileExists(shortcutFile) {
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

	return os.WriteFile(shortcutFile, data, 0644)
}

var reservedCommands = []string{
	"daily", "note", "search", "tags", "capture", "summary", "sync",
	"template", "streak", "calendar", "dashboard", "bulk", "graph",
	"alias", "check", "configure", "validate", "shortcut", "schedule",
	"help", "version", "list", "quick", "stats", "archive", "git",
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
	// Check if name is reserved
	if isReserved(name) {
		return fmt.Errorf("cannot use reserved command name: %s", name)
	}

	// Load existing shortcuts
	shortcuts, err := loadShortcuts(cfg)
	if err != nil {
		return err
	}

	// Add shortcut
	shortcuts[name] = command

	// Save
	if err := saveShortcuts(cfg, shortcuts); err != nil {
		return err
	}

	fmt.Printf("✓ Added shortcut: %s → %s\n", name, command)

	return nil
}

func removeShortcut(cfg *config.LoadedConfig, name string) error {
	// Load existing shortcuts
	shortcuts, err := loadShortcuts(cfg)
	if err != nil {
		return err
	}

	// Check if exists
	if _, exists := shortcuts[name]; !exists {
		return fmt.Errorf("shortcut not found: %s", name)
	}

	// Remove
	delete(shortcuts, name)

	// Save
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

	fmt.Println("Shortcuts:\n")
	for name, command := range shortcuts {
		fmt.Printf("  %s → %s\n", name, command)
	}

	return nil
}

