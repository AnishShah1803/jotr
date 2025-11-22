package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/anish/jotr/internal/tasks"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync tasks to to-do list",
	Long:  `Sync tasks from daily note to the main to-do list.`,
	Aliases: []string{"s"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return syncTasks(cfg)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func syncTasks(cfg *config.LoadedConfig) error {
	// Get today's note
	today := time.Now()
	notePath := notes.BuildDailyNotePath(cfg.DiaryPath, today)

	if !notes.FileExists(notePath) {
		return fmt.Errorf("today's note doesn't exist: %s", notePath)
	}

	// Read tasks from daily note
	dailyTasks, err := tasks.ReadTasks(notePath)
	if err != nil {
		return fmt.Errorf("failed to read daily note: %w", err)
	}

	// Filter tasks from the task section
	// Use task_section from config, default to "Tasks" if not set
	taskSection := cfg.Format.TaskSection
	if taskSection == "" {
		taskSection = "Tasks"
	}

	var tasksToSync []tasks.Task
	for _, task := range dailyTasks {
		if task.Section == taskSection && !task.Completed {
			// Ensure task has an ID
			tasks.EnsureTaskID(&task)
			tasksToSync = append(tasksToSync, task)
		}
	}

	if len(tasksToSync) == 0 {
		fmt.Println("No tasks to sync")
		return nil
	}

	// Read existing todo file
	var todoContent string
	if notes.FileExists(cfg.TodoPath) {
		content, err := os.ReadFile(cfg.TodoPath)
		if err != nil {
			return fmt.Errorf("failed to read todo file: %w", err)
		}
		todoContent = string(content)
	} else {
		// Create new todo file
		todoContent = "# To-Do List\n\n## Tasks\n\n"
	}

	// Add tasks to todo file
	lines := strings.Split(todoContent, "\n")
	
	// Find the Tasks section or create it
	tasksSectionIndex := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "## Tasks") || strings.HasPrefix(line, "## "+taskSection) {
			tasksSectionIndex = i
			break
		}
	}

	if tasksSectionIndex == -1 {
		// Add Tasks section at the end
		lines = append(lines, "", "## Tasks", "")
		tasksSectionIndex = len(lines) - 1
	}

	// Insert tasks after the section header
	insertIndex := tasksSectionIndex + 1
	// Skip empty lines
	for insertIndex < len(lines) && strings.TrimSpace(lines[insertIndex]) == "" {
		insertIndex++
	}

	// Add synced tasks
	newLines := make([]string, 0, len(lines)+len(tasksToSync))
	newLines = append(newLines, lines[:insertIndex]...)
	
	syncedCount := 0
	for _, task := range tasksToSync {
		taskLine := fmt.Sprintf("- [ ] %s", task.Text)
		
		// Check if task already exists
		exists := false
		for _, line := range lines {
			if strings.Contains(line, task.Text) {
				exists = true
				break
			}
		}

		if !exists {
			newLines = append(newLines, taskLine)
			syncedCount++
		}
	}

	newLines = append(newLines, lines[insertIndex:]...)

	// Write back
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(cfg.TodoPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write todo file: %w", err)
	}

	fmt.Printf("✓ Synced %d tasks to: %s\n", syncedCount, cfg.TodoPath)

	return nil
}

