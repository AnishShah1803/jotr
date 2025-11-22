package cmd

import (
	"fmt"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/tasks"
	"github.com/spf13/cobra"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show task summary",
	Long:  `Show a smart summary of tasks (overdue, P1, due soon).`,
	Aliases: []string{"sum"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return showSummary(cfg)
	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}

func showSummary(cfg *config.LoadedConfig) error {
	allTasks, err := tasks.ReadTasks(cfg.TodoPath)
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	// Filter pending tasks
	completed := false
	pendingTasks := tasks.FilterTasks(allTasks, &completed, "", "")

	if len(pendingTasks) == 0 {
		fmt.Println("✨ No pending tasks!")
		return nil
	}

	// Group by priority
	byPriority := tasks.GroupByPriority(pendingTasks)

	// Show summary
	fmt.Println("📋 Task Summary")
	fmt.Println("===============\n")

	// Show overdue tasks
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

	// Show P0 tasks
	if p0Tasks, ok := byPriority["P0"]; ok && len(p0Tasks) > 0 {
		fmt.Printf("🔴 P0 - Critical (%d):\n", len(p0Tasks))
		for _, task := range p0Tasks {
			fmt.Printf("  %s\n", tasks.FormatTask(task))
		}
		fmt.Println()
	}

	// Show P1 tasks
	if p1Tasks, ok := byPriority["P1"]; ok && len(p1Tasks) > 0 {
		fmt.Printf("🟠 P1 - High (%d):\n", len(p1Tasks))
		for _, task := range p1Tasks {
			fmt.Printf("  %s\n", tasks.FormatTask(task))
		}
		fmt.Println()
	}

	// Show P2 tasks
	if p2Tasks, ok := byPriority["P2"]; ok && len(p2Tasks) > 0 {
		fmt.Printf("🟡 P2 - Medium (%d):\n", len(p2Tasks))
		for _, task := range p2Tasks {
			fmt.Printf("  %s\n", tasks.FormatTask(task))
		}
		fmt.Println()
	}

	// Show P3 tasks
	if p3Tasks, ok := byPriority["P3"]; ok && len(p3Tasks) > 0 {
		fmt.Printf("🟢 P3 - Low (%d):\n", len(p3Tasks))
		for _, task := range p3Tasks {
			fmt.Printf("  %s\n", tasks.FormatTask(task))
		}
		fmt.Println()
	}

	// Show tasks without priority
	if noPriority, ok := byPriority["None"]; ok && len(noPriority) > 0 {
		fmt.Printf("⚪ No Priority (%d):\n", len(noPriority))
		for _, task := range noPriority {
			fmt.Printf("  %s\n", tasks.FormatTask(task))
		}
		fmt.Println()
	}

	// Show totals
	totalCount, completedCount, pendingCount := tasks.CountTasks(allTasks)
	fmt.Printf("Total: %d tasks (%d completed, %d pending)\n", totalCount, completedCount, pendingCount)

	return nil
}

