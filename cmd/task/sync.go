package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/services"
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
  jotr s                       # Using alias`,
	Aliases: []string{"s"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		return syncTasks(cmd.Context(), cfg)
	},
}

func syncTasks(ctx context.Context, cfg *config.LoadedConfig) error {
	taskService := services.NewTaskService()

	result, err := taskService.SyncTasks(ctx, services.SyncOptions{
		DiaryPath:   cfg.DiaryPath,
		TodoPath:    cfg.TodoPath,
		StatePath:   cfg.StatePath,
		TaskSection: cfg.Format.TaskSection,
	})
	if err != nil {
		return err
	}

	if len(result.Conflicts) > 0 {
		fmt.Println("⚠ Conflicts detected:")
		for taskID, conflict := range result.Conflicts {
			fmt.Printf("  - Task %s: %s\n", taskID, conflict)
		}
		fmt.Println("\nResolve conflicts manually and run sync again.")
		return nil
	}

	totalChanges := result.TasksFromDaily + result.TasksFromTodo
	if totalChanges == 0 && result.DeletedTasks == 0 {
		fmt.Println("✓ Everything is in sync")
		return nil
	}

	if result.TasksFromDaily > 0 {
		fmt.Printf("✓ Synced %d task(s) from daily notes to todo list\n", result.TasksFromDaily)
	}
	if result.TasksFromTodo > 0 {
		fmt.Printf("✓ Synced %d task(s) from todo list to daily notes\n", result.TasksFromTodo)
	}
	if result.DeletedTasks > 0 {
		fmt.Printf("✓ Removed %d deleted task(s) from state\n", result.DeletedTasks)
	}

	return nil
}
