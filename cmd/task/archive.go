package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/services"
)

var ArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive completed tasks",
	Long: `Archive completed tasks from the to-do list to an archive file.

Archived tasks are removed from the active todo list.

Examples:
  jotr archive                 # Archive completed tasks
  jotr arc                     # Using alias`,
	Aliases: []string{"arc"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		return archiveTasks(cmd.Context(), cfg)
	},
}

func archiveTasks(ctx context.Context, cfg *config.LoadedConfig) error {
	taskService := services.NewTaskService()

	result, err := taskService.ArchiveTasks(ctx, services.ArchiveOptions{
		TodoPath: cfg.TodoPath,
		BaseDir:  cfg.Paths.BaseDir,
	})
	if err != nil {
		return err
	}

	if result.ArchivedCount == 0 {
		fmt.Println("No completed tasks to archive")
		return nil
	}

	fmt.Printf("✓ Archived %d completed tasks to: %s\n", result.ArchivedCount, result.ArchivePath)
	fmt.Printf("✓ %d active tasks remaining\n", result.RemainingCount)

	return nil
}
