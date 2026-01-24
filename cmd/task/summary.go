package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/services"
	"github.com/AnishShah1803/jotr/internal/tasks"
)

var SummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show task summary",
	Long: `Show a smart summary of pending tasks grouped by priority.

Shows overdue tasks, P0 critical tasks, and other priorities.

Examples:
  jotr summary                 # Show task summary
  jotr sum                     # Using alias`,
	Aliases: []string{"sum"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		return ShowSummary(cmd.Context(), cfg)
	},
}

// ShowSummary displays a summary of pending tasks grouped by priority and section.
// It reads tasks from the configured todo file and presents them in a formatted view.
func ShowSummary(ctx context.Context, cfg *config.LoadedConfig) error {

	taskService := services.NewTaskService()

	stats, err := taskService.GetTaskStats(ctx, cfg.TodoPath)
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	allTasks, err := taskService.GetAllTasks(ctx, cfg.TodoPath)
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	completed := false
	pendingTasks := tasks.FilterTasks(allTasks, &completed, "", "")

	if len(pendingTasks) == 0 {
		fmt.Println("✨ No pending tasks!")
		return nil
	}

	byPriority := tasks.GroupByPriority(pendingTasks)

	fmt.Println("📋 Task Summary")
	fmt.Println("===============")
	fmt.Println()

	var overdue []tasks.Task

	for _, task := range pendingTasks {
		if tasks.IsOverdue(task) {
			overdue = append(overdue, task)
		}
	}

	if len(overdue) > 0 {
		fmt.Printf("⚠️  Overdue (%d):\n", len(overdue))

		for _, task := range overdue {
			fmt.Printf("  %s\n", tasks.FormatTask(task))
		}

		fmt.Println()
	}

	if p0Tasks, ok := byPriority["P0"]; ok && len(p0Tasks) > 0 {
		fmt.Printf("🔴 P0 - Critical (%d):\n", len(p0Tasks))

		for _, task := range p0Tasks {
			fmt.Printf("  %s\n", tasks.FormatTask(task))
		}

		fmt.Println()
	}

	if p1Tasks, ok := byPriority["P1"]; ok && len(p1Tasks) > 0 {
		fmt.Printf("🟠 P1 - High (%d):\n", len(p1Tasks))

		for _, task := range p1Tasks {
			fmt.Printf("  %s\n", tasks.FormatTask(task))
		}

		fmt.Println()
	}

	if p2Tasks, ok := byPriority["P2"]; ok && len(p2Tasks) > 0 {
		fmt.Printf("🟡 P2 - Medium (%d):\n", len(p2Tasks))

		for _, task := range p2Tasks {
			fmt.Printf("  %s\n", tasks.FormatTask(task))
		}

		fmt.Println()
	}

	if p3Tasks, ok := byPriority["P3"]; ok && len(p3Tasks) > 0 {
		fmt.Printf("🟢 P3 - Low (%d):\n", len(p3Tasks))

		for _, task := range p3Tasks {
			fmt.Printf("  %s\n", tasks.FormatTask(task))
		}

		fmt.Println()
	}

	if noPriority, ok := byPriority["None"]; ok && len(noPriority) > 0 {
		fmt.Printf("⚪ No Priority (%d):\n", len(noPriority))

		for _, task := range noPriority {
			fmt.Printf("  %s\n", tasks.FormatTask(task))
		}

		fmt.Println()
	}

	fmt.Printf("Total: %d tasks (%d completed, %d pending)\n", stats.Total, stats.Completed, stats.Pending)

	return nil
}
