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
	Short: "Sync tasks to to-do list",
	Long: `Sync tasks from daily notes to the main to-do list.

Tasks are extracted from the Tasks section and added to your todo file.

Examples:
  jotr sync                    # Sync tasks from today
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

	if result.TasksSynced == 0 {
		fmt.Println("No tasks to sync")
		return nil
	}

	fmt.Printf("✓ Synced %d tasks to: %s\n", result.TasksSynced, result.TodoPath)

	return nil
}
