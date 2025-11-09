package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/anish/jotr/internal/tasks"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive completed tasks",
	Long:  `Archive completed tasks from the to-do list.`,
	Aliases: []string{"arc"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return archiveTasks(cfg)
	},
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

func archiveTasks(cfg *config.LoadedConfig) error {
	// Read tasks from todo file
	allTasks, err := tasks.ReadTasks(cfg.TodoPath)
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	// Filter completed tasks
	var completedTasks []tasks.Task
	var activeTasks []tasks.Task

	for _, task := range allTasks {
		if task.Completed {
			completedTasks = append(completedTasks, task)
		} else {
			activeTasks = append(activeTasks, task)
		}
	}

	if len(completedTasks) == 0 {
		fmt.Println("No completed tasks to archive")
		return nil
	}

	// Create archive file path
	now := time.Now()
	archiveDir := filepath.Join(cfg.Paths.BaseDir, "Archive")
	if err := notes.EnsureDir(archiveDir); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	archiveFile := filepath.Join(archiveDir, fmt.Sprintf("archive-%s.md", now.Format("2006-01")))

	// Read existing archive or create new
	var archiveContent string
	if notes.FileExists(archiveFile) {
		content, err := os.ReadFile(archiveFile)
		if err != nil {
			return fmt.Errorf("failed to read archive: %w", err)
		}
		archiveContent = string(content)
	} else {
		archiveContent = fmt.Sprintf("# Archive - %s\n\n", now.Format("January 2006"))
	}

	// Add completed tasks to archive
	archiveContent += fmt.Sprintf("\n## Archived on %s\n\n", now.Format("2006-01-02"))
	for _, task := range completedTasks {
		archiveContent += fmt.Sprintf("- [x] %s\n", task.Text)
	}

	// Write archive
	if err := os.WriteFile(archiveFile, []byte(archiveContent), 0644); err != nil {
		return fmt.Errorf("failed to write archive: %w", err)
	}

	// Rebuild todo file with only active tasks
	content, err := os.ReadFile(cfg.TodoPath)
	if err != nil {
		return fmt.Errorf("failed to read todo file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	// Keep non-task lines and active tasks
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Keep headers and non-task lines
		if !strings.HasPrefix(trimmed, "- [x]") && !strings.HasPrefix(trimmed, "- [X]") {
			newLines = append(newLines, line)
		}
	}

	// Write back
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(cfg.TodoPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write todo file: %w", err)
	}

	fmt.Printf("✓ Archived %d completed tasks to: %s\n", len(completedTasks), archiveFile)
	fmt.Printf("✓ %d active tasks remaining\n", len(activeTasks))

	return nil
}

