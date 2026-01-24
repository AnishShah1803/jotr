//go:build dev

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/utils"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev [command]",
	Short: "Run jotr in development mode with isolated config",
	Long: `Development mode uses an isolated configuration and data directory
to avoid interfering with your production jotr installation.

The dev mode uses './dev-config.json' for configuration and creates
notes in './dev-data/' by default.

Examples:
  jotr dev daily           # Open today's note in dev mode
  jotr dev configure       # Set up dev configuration
  jotr dev --help         # Show all dev commands`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("jotr dev requires a command")
			fmt.Println("Use 'jotr dev --help' for available commands")
			os.Exit(1)
		}

		// Set dev config path
		devConfigPath := "dev-config.json"
		if _, err := os.Stat(devConfigPath); os.IsNotExist(err) {
			// Create dev config if it doesn't exist
			if err := createDevConfig(devConfigPath); err != nil {
				utils.PrintError("Failed to create dev config: %v", err)
				os.Exit(1)
			}
			fmt.Printf("Created dev config at: %s\n", devConfigPath)
		}

		// Execute the subcommand with dev config
		rootCmd.SetArgs(append([]string{"--config", devConfigPath}, args...))
		if err := rootCmd.Execute(); err != nil {
			utils.PrintError("%v", err)
			os.Exit(1)
		}
	},
}

func createDevConfig(path string) error {
	// Create dev data directory
	devDataDir := "dev-data"
	if err := os.MkdirAll(filepath.Join(devDataDir, "Diary"), 0755); err != nil {
		return fmt.Errorf("failed to create dev data directory: %w", err)
	}

	// Create dev config directly
	cfg := &config.Config{}
	cfg.Paths.BaseDir = devDataDir
	cfg.Paths.DiaryDir = "Diary"
	cfg.Paths.TodoFilePath = "todo"
	cfg.Paths.PDPFilePath = ""

	cfg.Format.TaskSection = "Tasks"
	cfg.Format.CaptureSection = "Captured"
	cfg.Format.DailyNoteSections = []string{"Notes", "Development"}
	cfg.Format.DailyNotePattern = "{year}-{month}-{day}.md"
	cfg.Format.DailyNoteDirPattern = "{year}/{month}"

	cfg.AI.Enabled = false
	cfg.AI.Command = ""

	cfg.Streaks.IncludeWeekends = false

	return saveDevConfig(path, cfg)
}

func saveDevConfig(path string, cfg *config.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write dev config: %w", err)
	}

	return nil
}
