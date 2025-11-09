package cmd

import (
	"fmt"
	"os"

	"github.com/anish/jotr/internal/config"
	"github.com/spf13/cobra"
)

const Version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "jotr",
	Short: "A powerful note-taking and task management CLI",
	Long: `jotr is a command-line tool for managing daily notes, tasks, and knowledge.
It helps you capture thoughts, track tasks, and organize your notes efficiently.

Run 'jotr' with no arguments to launch the interactive dashboard.`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand, launch dashboard
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			fmt.Println("\nRun 'jotr configure' to set up your configuration")
			os.Exit(1)
		}

		if err := launchDashboard(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error launching dashboard: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	
	// Add subcommands
	rootCmd.AddCommand(dailyCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configureCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("jotr version %s\n", Version)
	},
}

