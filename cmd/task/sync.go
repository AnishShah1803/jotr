package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/output"
	"github.com/AnishShah1803/jotr/internal/services"
)

var (
	syncDryRun  bool
	syncQuiet   bool
	syncJSON    bool
	syncVerbose bool
	syncNoColor bool
)

// Color configuration now centralized in internal/output
var (
	successColor = output.SuccessColor
	warningColor = output.WarningColor
	errorColor   = output.ErrorColor
	mutedColor   = output.MutedColor
)

func isColorEnabled() bool {
	if syncNoColor {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		// Log the error to stderr so users can notice issues with output redirection
		// This is important when stdout is redirected to a file or a pipe which
		// may cause Stat() to return an error. The warning is non-fatal.
		fmt.Fprintf(os.Stderr, "warning: failed to stat stdout: %v\n", err)
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) == os.ModeCharDevice
}

func colorize(text string, color lipgloss.Color) string {
	return output.Colorize(text, color, isColorEnabled())
}

func formatPrefix(prefix string) string {
	switch prefix {
	case "+":
		return colorize("+", successColor)
	case "~":
		return colorize("~", warningColor)
	case "!":
		return colorize("!", errorColor)
	case "-":
		return colorize("-", mutedColor)
	case "✓":
		return colorize("✓", successColor)
	case "⚠":
		return colorize("⚠", warningColor)
	default:
		return prefix
	}
}

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync tasks bidirectionally",
	Long: `Sync tasks between daily notes and todo list.

Changes in daily notes are propagated to the todo list.
Changes in the todo list are propagated to daily notes.
Conflicts are detected and reported.

Examples:
  jotr sync                    # Sync tasks bidirectionally
  jotr s                       # Using alias
  jotr sync --dry-run          # Preview changes without applying
  jotr sync --json             # Output in JSON format
  jotr sync --quiet            # Show only summary counts`,
	Aliases: []string{"s"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		return syncTasks(cmd.Context(), cfg)
	},
}

func init() {
	SyncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Show what would be done without making changes")
	SyncCmd.Flags().BoolVar(&syncQuiet, "quiet", false, "Suppress normal output, show only summary")
	SyncCmd.Flags().BoolVar(&syncJSON, "json", false, "Output in JSON format")
	SyncCmd.Flags().BoolVar(&syncVerbose, "verbose", false, "Enable verbose output with detailed task information")
	SyncCmd.Flags().BoolVar(&syncNoColor, "no-color", false, "Disable colored output")
}

func syncTasks(ctx context.Context, cfg *config.LoadedConfig) error {
	taskService := services.NewTaskService()

	opts := services.SyncOptions{
		DiaryPath:   cfg.DiaryPath,
		TodoPath:    cfg.TodoPath,
		StatePath:   cfg.StatePath,
		TaskSection: cfg.Format.TaskSection,
		DryRun:      syncDryRun,
	}

	result, err := taskService.SyncTasks(ctx, opts)
	if err != nil {
		return err
	}

	if syncJSON {
		return outputSyncJSON(result)
	}

	if syncQuiet {
		return outputSyncQuiet(result)
	}

	return outputSyncDefault(result, syncVerbose)
}

func outputSyncJSON(result *services.SyncResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result to JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputSyncQuiet(result *services.SyncResult) error {
	if len(result.Conflicts) > 0 {
		fmt.Printf("Conflicts: %d\n", len(result.Conflicts))
		return nil
	}

	totalChanges := result.TasksFromDaily + result.TasksFromTodo
	if totalChanges == 0 && result.DeletedTasks == 0 {
		fmt.Println("No changes")
		return nil
	}

	fmt.Printf("Daily: %d, Todo: %d, Deleted: %d\n", result.TasksFromDaily, result.TasksFromTodo, result.DeletedTasks)
	return nil
}

func outputSyncDefault(result *services.SyncResult, verbose bool) error {
	if syncDryRun {
		fmt.Printf("%s DRY RUN - No changes made\n", formatPrefix("⚠"))
		fmt.Println()
	}

	if len(result.Conflicts) > 0 {
		fmt.Printf("%s Conflicts detected:\n", formatPrefix("⚠"))
		for _, conflict := range result.ConflictsDetail {
			fmt.Printf("  %s \"%s\" - %s\n", formatPrefix("!"), conflict.TextDaily, conflict.Reason)
			if conflict.TextDaily != conflict.TextTodo {
				fmt.Printf("      Daily: \"%s\"\n", conflict.TextDaily)
				fmt.Printf("      Todo:  \"%s\"\n", conflict.TextTodo)
			}
		}
		fmt.Println("\nResolve conflicts manually and run sync again.")
		return nil
	}

	totalChanges := result.TasksFromDaily + result.TasksFromTodo
	if totalChanges == 0 && result.DeletedTasks == 0 {
		fmt.Printf("%s Everything is in sync\n", formatPrefix("✓"))
		return nil
	}

	if len(result.AddedFromDaily) > 0 || len(result.UpdatedFromDaily) > 0 {
		fmt.Println("From Daily Notes:")
		for _, task := range result.AddedFromDaily {
			if verbose {
				fmt.Printf("  %s Added: \"%s\" (id: %s)\n", formatPrefix("+"), task.Text, task.ID)
				if task.To != "" && task.To != task.Text {
					fmt.Printf("      To: \"%s\"\n", task.To)
				}
				if task.Details != "" {
					fmt.Printf("      Details: %s\n", task.Details)
				}
			} else {
				fmt.Printf("  %s Added: \"%s\" (id: %s)\n", formatPrefix("+"), task.Text, task.ID)
			}
		}
		for _, task := range result.UpdatedFromDaily {
			if verbose {
				fmt.Printf("  %s Updated: (id: %s)\n", formatPrefix("~"), task.ID)
				if task.From != "" {
					fmt.Printf("      From: \"%s\"\n", task.From)
				}
				if task.To != "" {
					fmt.Printf("      To:   \"%s\"\n", task.To)
				}
				if task.Details != "" {
					fmt.Printf("      Details: %s\n", task.Details)
				}
			} else {
				if task.Details != "" {
					fmt.Printf("  %s Updated: \"%s\" - %s (id: %s)\n", formatPrefix("~"), task.Text, task.Details, task.ID)
				} else {
					fmt.Printf("  %s Updated: \"%s\" (id: %s)\n", formatPrefix("~"), task.Text, task.ID)
				}
			}
		}
		fmt.Println()
	}

	if len(result.AddedFromTodo) > 0 || len(result.UpdatedFromTodo) > 0 {
		fmt.Println("From Todo List:")
		for _, task := range result.AddedFromTodo {
			if verbose {
				fmt.Printf("  %s Added: \"%s\" (id: %s)\n", formatPrefix("+"), task.Text, task.ID)
				if task.To != "" && task.To != task.Text {
					fmt.Printf("      To: \"%s\"\n", task.To)
				}
				if task.Details != "" {
					fmt.Printf("      Details: %s\n", task.Details)
				}
			} else {
				fmt.Printf("  %s Added: \"%s\" (id: %s)\n", formatPrefix("+"), task.Text, task.ID)
			}
		}
		for _, task := range result.UpdatedFromTodo {
			if verbose {
				fmt.Printf("  %s Updated: (id: %s)\n", formatPrefix("~"), task.ID)
				if task.From != "" {
					fmt.Printf("      From: \"%s\"\n", task.From)
				}
				if task.To != "" {
					fmt.Printf("      To:   \"%s\"\n", task.To)
				}
				if task.Details != "" {
					fmt.Printf("      Details: %s\n", task.Details)
				}
			} else {
				if task.Details != "" {
					fmt.Printf("  %s Updated: \"%s\" - %s (id: %s)\n", formatPrefix("~"), task.Text, task.Details, task.ID)
				} else {
					fmt.Printf("  %s Updated: \"%s\" (id: %s)\n", formatPrefix("~"), task.Text, task.ID)
				}
			}
		}
		fmt.Println()
	}

	if len(result.DeletedTasksDetail) > 0 {
		fmt.Println("Deleted:")
		for _, task := range result.DeletedTasksDetail {
			if verbose {
				fmt.Printf("  %s \"%s\" (id: %s)\n", formatPrefix("-"), task.Text, task.ID)
				if task.From != "" {
					fmt.Printf("      From: \"%s\"\n", task.From)
				}
				if task.Details != "" {
					fmt.Printf("      Details: %s\n", task.Details)
				}
			} else {
				fmt.Printf("  %s \"%s\" (id: %s)\n", formatPrefix("-"), task.Text, task.ID)
			}
		}
		fmt.Println()
	}

	fmt.Println("Summary:")
	fmt.Printf("  %d tasks checked\n", result.TasksRead)
	if result.TasksFromDaily > 0 {
		fmt.Printf("  %d task(s) added/updated from daily notes\n", result.TasksFromDaily)
	}
	if result.TasksFromTodo > 0 {
		fmt.Printf("  %d task(s) added/updated from todo list\n", result.TasksFromTodo)
	}
	if result.DeletedTasks > 0 {
		fmt.Printf("  %d task(s) deleted\n", result.DeletedTasks)
	}

	if verbose {
		if len(result.ChangedTaskIDs) > 0 {
			fmt.Println()
			fmt.Println("Changed task IDs:")
			for _, id := range result.ChangedTaskIDs {
				fmt.Printf("  - %s\n", id)
			}
		}
		if len(result.DeletedTaskIDs) > 0 {
			fmt.Println()
			fmt.Println("Deleted task IDs:")
			for _, id := range result.DeletedTaskIDs {
				fmt.Printf("  - %s\n", id)
			}
		}
	}
	if syncDryRun {
		fmt.Println("\n(No changes were made - dry run mode)")
	}

	return nil
}
