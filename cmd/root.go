package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/repl"
	"github.com/AnishShah1803/jotr/internal/updater"
	"github.com/AnishShah1803/jotr/internal/utils"
	"github.com/AnishShah1803/jotr/internal/version"

	indexcmd "github.com/AnishShah1803/jotr/cmd/index"
	notecmd "github.com/AnishShah1803/jotr/cmd/note"
	searchcmd "github.com/AnishShah1803/jotr/cmd/search"
	systemcmd "github.com/AnishShah1803/jotr/cmd/system"
	taskcmd "github.com/AnishShah1803/jotr/cmd/task"
	templatecmd "github.com/AnishShah1803/jotr/cmd/templatecmd"
	utilcmd "github.com/AnishShah1803/jotr/cmd/util"
	visualcmd "github.com/AnishShah1803/jotr/cmd/visual"
)

var updateFlag bool
var configPath string
var timeout time.Duration

var rootCmd = &cobra.Command{
	Use:   "jotr",
	Short: "A powerful journaling and note-taking tool",
	Long: `jotr is a command-line journaling and note-taking tool designed for daily use.
It supports daily notes, task management, templates, search, and much more.

When run without arguments, jotr launches an interactive REPL interface.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		timeout, _ := cmd.Flags().GetDuration("timeout")
		configPath, _ := cmd.Flags().GetString("config")

		var ctx context.Context
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), timeout)
			defer cancel()
			utils.VerboseLogWithContext(ctx, "Context created with timeout: %v", timeout)
		} else {
			ctx = context.Background()
		}

		if verbose {
			ctx = utils.WithVerboseContext(ctx, true)
			utils.VerboseLogWithContext(ctx, "Starting jotr with verbose mode enabled")
		}

		ctx = config.WithConfig(ctx, configPath)
		if configPath != "" {
			utils.VerboseLog("Config path set to: %s", configPath)
		}

		cmd.SetContext(ctx)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateFlag {
			return updater.CheckAndUpdate(true)
		}

		ctx := cmd.Context()
		cfg, err := config.LoadWithContext(ctx, "")
		if err != nil {
			utils.PrintError("loading config: %v", err)
			utils.PrintError("Run 'jotr configure' to set up your configuration")
			return fmt.Errorf("config load failed: %w", err)
		}

		return repl.LaunchREPL(ctx, cfg, cmd.Root())
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("jotr version %s\n", version.GetVersion())
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 0, "command timeout (e.g., 30s, 5m)")
	rootCmd.Flags().BoolVar(&updateFlag, "update", false, "check for and install updates")

	// Core Note Management
	rootCmd.AddCommand(notecmd.DailyCmd)
	rootCmd.AddCommand(notecmd.NoteCmd)
	rootCmd.AddCommand(notecmd.CaptureCmd)
	rootCmd.AddCommand(notecmd.ReadCmd)

	// Task Management
	rootCmd.AddCommand(taskcmd.TaskCmd)
	rootCmd.AddCommand(taskcmd.SyncCmd)
	rootCmd.AddCommand(taskcmd.SummaryCmd)
	rootCmd.AddCommand(taskcmd.StatsCmd)
	rootCmd.AddCommand(taskcmd.ArchiveCmd)

	// Search and Navigation
	rootCmd.AddCommand(searchcmd.SearchCmd)
	rootCmd.AddCommand(searchcmd.TagsCmd)
	rootCmd.AddCommand(searchcmd.LinksCmd)
	rootCmd.AddCommand(searchcmd.AliasesCmd)
	rootCmd.AddCommand(searchcmd.RecentsCmd)
	rootCmd.AddCommand(searchcmd.ListCmd)
	rootCmd.AddCommand(searchcmd.FilesCmd)
	rootCmd.AddCommand(searchcmd.WordcountCmd)
	rootCmd.AddCommand(searchcmd.OutlineCmd)

	// Visualization
	rootCmd.AddCommand(visualcmd.CalendarCmd)
	rootCmd.AddCommand(visualcmd.StreakCmd)
	rootCmd.AddCommand(visualcmd.GraphCmd)

	// Productivity Features
	rootCmd.AddCommand(systemcmd.AliasCmd)
	rootCmd.AddCommand(systemcmd.ShortcutCmd)
	rootCmd.AddCommand(systemcmd.ScheduleCmd)
	rootCmd.AddCommand(systemcmd.MonthlyCmd)
	rootCmd.AddCommand(systemcmd.FrontmatterCmd)
	rootCmd.AddCommand(systemcmd.PropertiesCmd)

	// Utilities
	rootCmd.AddCommand(utilcmd.BulkCmd)
	rootCmd.AddCommand(utilcmd.GitCmd)

	rootCmd.AddCommand(utilcmd.CheckCmd)
	rootCmd.AddCommand(utilcmd.ValidateCmd)

	// Templates
	rootCmd.AddCommand(templatecmd.TemplateCmd)
	rootCmd.AddCommand(systemcmd.UpdateCmd)

	// Setup
	rootCmd.AddCommand(systemcmd.ConfigureCmd)
	rootCmd.AddCommand(versionCmd)

	// Search Index
	rootCmd.AddCommand(indexcmd.IndexCmd)
}
