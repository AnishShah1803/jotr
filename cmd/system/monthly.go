package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/tasks"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var (
	monthlyYear  int
	monthlyMonth int
)

var MonthlyCmd = &cobra.Command{
	Use:   "monthly",
	Short: "Generate monthly summary",
	Long: `Generate a summary of the month including notes and tasks.

Creates a summary markdown file in the summaries directory.

Examples:
  jotr monthly --year 2025 --month 1  # Generate January 2025 summary
  jotr month                          # Using alias`,
	Aliases: []string{"month"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		return generateMonthlySummary(cmd.Context(), cfg)
	},
}

func generateMonthlySummary(ctx context.Context, cfg *config.LoadedConfig) error {
	// Validate month
	if monthlyMonth < 1 || monthlyMonth > 12 {
		return fmt.Errorf("month must be between 1 and 12")
	}

	// Create target date
	targetDate := time.Date(monthlyYear, time.Month(monthlyMonth), 1, 0, 0, 0, 0, time.Local)
	monthStr := targetDate.Format("01")
	monthAbbr := targetDate.Format("Jan")
	monthName := targetDate.Format("January")

	// Build month directory path
	monthDir := filepath.Join(cfg.DiaryPath, fmt.Sprintf("%d", monthlyYear), fmt.Sprintf("%s-%s", monthStr, monthAbbr))

	// Check if directory exists
	if !utils.FileExists(monthDir) {
		return fmt.Errorf("no daily notes found for %s %d", monthName, monthlyYear)
	}

	// Find all daily notes in the month
	dailyNotes, err := filepath.Glob(filepath.Join(monthDir, "*.md"))
	if err != nil {
		return fmt.Errorf("failed to find notes: %w", err)
	}

	// Filter out summary files
	var validNotes []string

	for _, note := range dailyNotes {
		if !strings.HasSuffix(note, "-Summary.md") {
			validNotes = append(validNotes, note)
		}
	}

	if len(validNotes) == 0 {
		return fmt.Errorf("no daily notes found for %s %d", monthName, monthlyYear)
	}

	// Create summaries directory
	yearDir := filepath.Join(cfg.DiaryPath, fmt.Sprintf("%d", monthlyYear))

	summariesDir := filepath.Join(yearDir, "summaries")
	if err := notes.EnsureDir(summariesDir); err != nil {
		return fmt.Errorf("failed to create summaries directory: %w", err)
	}

	summaryPath := filepath.Join(summariesDir, fmt.Sprintf("%s-%s-Summary.md", monthStr, monthAbbr))

	// Generate summary content
	content := fmt.Sprintf("# %s %d Summary\n\n", monthName, monthlyYear)
	content += fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02"))
	content += fmt.Sprintf("## Overview\n\n")
	content += fmt.Sprintf("- Total daily notes: %d\n", len(validNotes))
	content += fmt.Sprintf("- Month: %s %d\n\n", monthName, monthlyYear)

	// Count tasks from todo file
	if utils.FileExists(cfg.TodoPath) {
		allTasks, err := tasks.ReadTasks(ctx, cfg.TodoPath)
		if err == nil {
			total, completed, pending := tasks.CountTasks(allTasks)
			content += fmt.Sprintf("## Tasks\n\n")
			content += fmt.Sprintf("- Total tasks: %d\n", total)
			content += fmt.Sprintf("- Completed: %d\n", completed)
			content += fmt.Sprintf("- Pending: %d\n\n", pending)
		}
	}

	// List daily notes
	content += fmt.Sprintf("## Daily Notes\n\n")

	for _, notePath := range validNotes {
		basename := filepath.Base(notePath)
		content += fmt.Sprintf("- %s\n", basename)
	}

	// Write summary
	if err := os.WriteFile(summaryPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write summary: %w", err)
	}

	fmt.Printf("✓ Monthly summary created: %s\n", summaryPath)
	fmt.Printf("  %d daily notes\n", len(validNotes))

	// Open in editor
	return notes.OpenInEditor(summaryPath)
}
