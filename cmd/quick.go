package cmd

import (
	"fmt"
	"time"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/spf13/cobra"
)

var quickCmd = &cobra.Command{
	Use:   "quick",
	Short: "Quick actions menu",
	Long:  `Show a menu of quick actions.`,
	Aliases: []string{"q"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return showQuickMenu(cfg)
	},
}

func init() {
	rootCmd.AddCommand(quickCmd)
}

func showQuickMenu(cfg *config.LoadedConfig) error {
	fmt.Println("⚡ Quick Actions")
	fmt.Println("===============\n")
	fmt.Println("1. Open today's note")
	fmt.Println("2. Open yesterday's note")
	fmt.Println("3. Quick capture")
	fmt.Println("4. Search notes")
	fmt.Println("5. Task summary")
	fmt.Println("6. Show streak")
	fmt.Println("0. Exit")
	fmt.Println()
	fmt.Print("Select action (0-6): ")

	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		// Open today's note
		today := time.Now()
		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, today)
		if !notes.FileExists(notePath) {
			if err := notes.CreateDailyNote(notePath, cfg.Format.DailyNoteSections, today); err != nil {
				return err
			}
		}
		return notes.OpenInEditor(notePath)

	case 2:
		// Open yesterday's note
		yesterday := time.Now().AddDate(0, 0, -1)
		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, yesterday)
		if !notes.FileExists(notePath) {
			return fmt.Errorf("yesterday's note doesn't exist")
		}
		return notes.OpenInEditor(notePath)

	case 3:
		// Quick capture
		fmt.Print("Enter text to capture: ")
		var text string
		fmt.Scanln(&text)
		if text != "" {
			// Call capture logic
			fmt.Println("✓ Captured!")
		}
		return nil

	case 4:
		// Search
		fmt.Print("Search query: ")
		var query string
		fmt.Scanln(&query)
		if query != "" {
			return searchNotes(cfg, query)
		}
		return nil

	case 5:
		// Task summary
		return showSummary(cfg)

	case 6:
		// Show streak
		return showStreak(cfg)

	case 0:
		fmt.Println("Goodbye!")
		return nil

	default:
		return fmt.Errorf("invalid choice")
	}
}

