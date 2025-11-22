package cmd

import (
	"fmt"
	"os"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/updater"
	"github.com/anish/jotr/internal/utils"
	"github.com/anish/jotr/internal/version"
	"github.com/spf13/cobra"
)

var updateFlag bool

var rootCmd = &cobra.Command{
	Use:   "jotr",
	Short: "A powerful journaling and note-taking tool",
	Long: `jotr is a command-line journaling and note-taking tool designed for daily use.
It supports daily notes, task management, templates, search, and much more.

When run without arguments, jotr launches the interactive dashboard.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set verbose mode based on flag
		verbose, _ := cmd.Flags().GetBool("verbose")
		utils.SetVerbose(verbose)
		utils.VerboseLog("Starting jotr with verbose mode enabled")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle --update flag
		if updateFlag {
			return updater.CheckAndUpdate(true)
		}

		// If no subcommand, launch dashboard
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			fmt.Fprintf(os.Stderr, "Run 'jotr configure' to set up your configuration\n")
			return fmt.Errorf("config load failed: %w", err)
		}

		if err := launchDashboard(cfg); err != nil {
			return fmt.Errorf("dashboard launch failed: %w", err)
		}
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("jotr version %s\n", version.Version)
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.Flags().BoolVar(&updateFlag, "update", false, "check for and install updates")

	// Core Note Management
	rootCmd.AddCommand(dailyCmd)
	rootCmd.AddCommand(noteCmd)
	rootCmd.AddCommand(captureCmd)
	rootCmd.AddCommand(templateCmd)

	// Task Management
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(summaryCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(archiveCmd)

	// Search and Navigation
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(tagsCmd)
	rootCmd.AddCommand(linksCmd)
	rootCmd.AddCommand(listCmd)

	// Visualization
	rootCmd.AddCommand(calendarCmd)
	rootCmd.AddCommand(streakCmd)
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(dashboardCmd)

	// Productivity Features
	rootCmd.AddCommand(aliasCmd)
	rootCmd.AddCommand(shortcutCmd)
	rootCmd.AddCommand(scheduleCmd)
	rootCmd.AddCommand(monthlyCmd)
	rootCmd.AddCommand(frontmatterCmd)

	// Utilities
	rootCmd.AddCommand(bulkCmd)
	rootCmd.AddCommand(gitCmd)
	rootCmd.AddCommand(quickCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(updateCmd)

	// Setup
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(versionCmd)
}

// Note: launchDashboard is defined in dashboard.go
