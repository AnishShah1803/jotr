package cmd

import (
	"fmt"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/tasks"
	"github.com/spf13/cobra"
)

var (
	statsWeek  bool
	statsMonth bool
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show task statistics",
	Long:  `Show detailed task statistics.`,
	Aliases: []string{"st"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return showStats(cfg)
	},
}

func init() {
	statsCmd.Flags().BoolVar(&statsWeek, "week", false, "Show only this week")
	statsCmd.Flags().BoolVar(&statsMonth, "month", false, "Show only this month")
	rootCmd.AddCommand(statsCmd)
}

func showStats(cfg *config.LoadedConfig) error {
	// Read tasks from todo file
	allTasks, err := tasks.ReadTasks(cfg.TodoPath)
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	fmt.Println("📊 Task Statistics")
	fmt.Println("==================\n")

	// Overall stats
	total, completed, pending := tasks.CountTasks(allTasks)
	completionRate := 0.0
	if total > 0 {
		completionRate = float64(completed) / float64(total) * 100
	}

	fmt.Printf("Total Tasks:      %d\n", total)
	fmt.Printf("Completed:        %d (%.1f%%)\n", completed, completionRate)
	fmt.Printf("Pending:          %d\n", pending)
	fmt.Println()

	// By priority
	byPriority := tasks.GroupByPriority(allTasks)
	fmt.Println("By Priority:")
	priorities := []string{"P0", "P1", "P2", "P3", "None"}
	for _, priority := range priorities {
		if taskList, ok := byPriority[priority]; ok {
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

	// By section
	bySection := tasks.GroupBySection(allTasks)
	fmt.Println("By Section:")
	for section, taskList := range bySection {
		completedCount := 0
		for _, task := range taskList {
			if task.Completed {
				completedCount++
			}
		}
		fmt.Printf("  %-20s %d tasks (%d completed)\n", section+":", len(taskList), completedCount)
	}
	fmt.Println()

	// Overdue tasks
	var overdue []tasks.Task
	for _, task := range allTasks {
		if tasks.IsOverdue(task) {
			overdue = append(overdue, task)
		}
	}

	if len(overdue) > 0 {
		fmt.Printf("⚠️  Overdue: %d tasks\n", len(overdue))
	}

	return nil
}

