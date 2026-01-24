package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/options"
	"github.com/AnishShah1803/jotr/internal/services"
)

var statsTimeRange = options.NewTimeRangeOption()

func init() {
	statsTimeRange.AddFlags(StatsCmd)
}

var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show task statistics",
	Long: `Show detailed task statistics including completion rates and priorities.

Examples:
  jotr stats                   # Show all-time stats
  jotr stats --week           # Show stats for last 7 days
  jotr stats --month          # Show stats for last 30 days
  jotr st                     # Using alias`,
	Aliases: []string{"st"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		return showStats(cmd.Context(), cfg)
	},
}

func showStats(ctx context.Context, cfg *config.LoadedConfig) error {
	taskService := services.NewTaskService()

	stats, err := taskService.GetTaskStats(ctx, cfg.TodoPath)
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	fmt.Println("📊 Task Statistics")
	fmt.Println("==================")
	fmt.Println()

	days := statsTimeRange.GetDays()
	if days == 7 {
		fmt.Println("📅 Last 7 Days")
	} else if days == 30 {
		fmt.Println("📅 Last 30 Days")
	} else if days == 365 {
		fmt.Println("📅 Last 365 Days")
	} else if days > 0 {
		fmt.Printf("📅 Last %d Days\n", days)
	} else {
		fmt.Println("📅 All Time")
	}
	fmt.Println()

	fmt.Printf("Total Tasks:      %d\n", stats.Total)
	fmt.Printf("Completed:        %d (%.1f%%)\n", stats.Completed, stats.CompletionRate)
	fmt.Printf("Pending:          %d\n", stats.Pending)
	fmt.Println()

	fmt.Println("By Priority:")

	priorities := []string{"P0", "P1", "P2", "P3", "None"}
	for _, priority := range priorities {
		if taskList, ok := stats.ByPriority[priority]; ok {
			completedCount := 0

			for _, task := range taskList {
				if task.Completed {
					completedCount++
				}
			}

			fmt.Printf("  %-10s %d tasks (%d completed)\n", priority+":", len(taskList), completedCount)
		}
	}

	fmt.Println()

	fmt.Println("By Section:")

	for section, taskList := range stats.BySection {
		completedCount := 0

		for _, task := range taskList {
			if task.Completed {
				completedCount++
			}
		}

		fmt.Printf("  %-20s %d tasks (%d completed)\n", section+":", len(taskList), completedCount)
	}

	fmt.Println()

	if stats.Overdue > 0 {
		fmt.Printf("⚠️  Overdue: %d tasks\n", stats.Overdue)
	}

	return nil
}
