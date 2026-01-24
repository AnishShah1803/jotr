package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/options"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var outputOption = options.NewOutputOption()
var recentNotesLimit = 5

func init() {
	outputOption.AddFlags(ListCmd)
	ListCmd.Flags().IntVar(&recentNotesLimit, "limit", 5, "Number of recent notes to show")
}

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent daily notes",
	Long: `List recent daily notes from the last few days.

Shows daily notes with dates and status indicators.

Examples:
  jotr list                   # List last 5 daily notes
  jotr list --files           # List all notes
  jotr ls                     # Using alias`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		return listRecentNotes(cmd.Context(), cfg)
	},
}

func listRecentNotes(ctx context.Context, cfg *config.LoadedConfig) error {
	if outputOption.FilesOnly {
		allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
		if err != nil {
			return fmt.Errorf("failed to find notes: %w", err)
		}

		fmt.Printf("Found %d notes:\n\n", len(allNotes))

		for _, notePath := range allNotes {
			relPath, _ := filepath.Rel(cfg.Paths.BaseDir, notePath)
			fmt.Printf("  %s\n", relPath)
		}

		return nil
	}

	fmt.Printf("Recent daily notes (last %d days):\n\n", recentNotesLimit)

	today := time.Now()
	foundCount := 0

	for i := 0; i < recentNotesLimit*2 && foundCount < recentNotesLimit; i++ {
		date := today.AddDate(0, 0, -i)
		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)

		if utils.FileExists(notePath) {
			dateStr := date.Format("2006-01-02 Mon")
			status := "✓"

			if i == 0 {
				dateStr += " (today)"
			} else if i == 1 {
				dateStr += " (yesterday)"
			}

			fmt.Printf("  %s %s\n", status, dateStr)

			foundCount++
		}
	}

	if foundCount == 0 {
		fmt.Println("  No daily notes found")
	}

	return nil
}
