package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/spf13/cobra"
)

var (
	listCount int
	listAll   bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent daily notes",
	Long:  `List recent daily notes.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return listRecentNotes(cfg)
	},
}

func init() {
	listCmd.Flags().IntVarP(&listCount, "count", "n", 10, "Number of notes to show")
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "Show all notes")
	rootCmd.AddCommand(listCmd)
}

func listRecentNotes(cfg *config.LoadedConfig) error {
	if listAll {
		allNotes, err := notes.FindNotes(cfg.Paths.BaseDir)
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

	fmt.Printf("Recent daily notes (last %d days):\n\n", listCount)

	today := time.Now()
	foundCount := 0

	for i := 0; i < listCount*2 && foundCount < listCount; i++ {
		date := today.AddDate(0, 0, -i)
		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)

		if notes.FileExists(notePath) {
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

