package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/services"
)

var (
	syncDryRun  bool
	syncQuiet   bool
	syncJSON    bool
	syncVerbose bool
)

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
}

func syncTasks(ctx context.Context, cfg *config.LoadedConfig) error {
	taskService := services.NewTaskService()

	opts := services.SyncOptions{
		DiaryPath:   cfg.DiaryPath,
		TodoPath:    cfg.TodoPath,
		StatePath:   cfg.StatePath,
		TaskSection: cfg.Format.TaskSection,
		DryRun:      syncDryRun,
		Quiet:       syncQuiet,
		JSON:        syncJSON,
		Verbose:     syncVerbose,
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
		fmt.Println("0")
		return nil
	}

	fmt.Printf("Daily: %d, Todo: %d, Deleted: %d\n", result.TasksFromDaily, result.TasksFromTodo, result.DeletedTasks)
	return nil
}

func outputSyncDefault(result *services.SyncResult, verbose bool) error {
	if syncDryRun {
		fmt.Println("⚠ DRY RUN - No changes made")
		fmt.Println()
	}

	if len(result.Conflicts) > 0 {
		fmt.Println("⚠ Conflicts detected:")
		for _, conflict := range result.ConflictsDetail {
			fmt.Printf("  ! \"%s\" - %s\n", conflict.TextDaily, conflict.Reason)
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
		fmt.Println("✓ Everything is in sync")
		return nil
	}

	if len(result.AddedFromDaily) > 0 || len(result.UpdatedFromDaily) > 0 {
		fmt.Println("From Daily Notes:")
		for _, task := range result.AddedFromDaily {
			fmt.Printf("  + Added: \"%s\" (id: %s)\n", task.Text, task.ID)
		}
		for _, task := range result.UpdatedFromDaily {
			if task.Details != "" {
				fmt.Printf("  ~ Updated: \"%s\" - %s (id: %s)\n", task.Text, task.Details, task.ID)
			} else {
				fmt.Printf("  ~ Updated: \"%s\" (id: %s)\n", task.Text, task.ID)
			}
		}
		fmt.Println()
	}

	if len(result.AddedFromTodo) > 0 || len(result.UpdatedFromTodo) > 0 {
		fmt.Println("From Todo List:")
		for _, task := range result.AddedFromTodo {
			fmt.Printf("  + Added: \"%s\" (id: %s)\n", task.Text, task.ID)
		}
		for _, task := range result.UpdatedFromTodo {
			if task.Details != "" {
				fmt.Printf("  ~ Updated: \"%s\" - %s (id: %s)\n", task.Text, task.Details, task.ID)
			} else {
				fmt.Printf("  ~ Updated: \"%s\" (id: %s)\n", task.Text, task.ID)
			}
		}
		fmt.Println()
	}

	if len(result.DeletedTasksDetail) > 0 {
		fmt.Println("Deleted:")
		for _, task := range result.DeletedTasksDetail {
			fmt.Printf("  - \"%s\" (id: %s)\n", task.Text, task.ID)
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

	if syncDryRun {
		fmt.Println("\n(No changes were made - dry run mode)")
	}

	return nil
}
