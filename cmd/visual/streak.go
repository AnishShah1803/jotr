package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/services"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var StreakCmd = &cobra.Command{
	Use:   "streak",
	Short: "Show daily note streak",
	Long: `Show your current streak of consecutive daily notes.

Displays current streak, longest streak, and a 7-day activity calendar.

Examples:
  jotr streak                 # Show streak information`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		return ShowStreak(cfg)
	},
}

func displayStreakInfo(result services.StreakResult) {
	fmt.Println("🔥 Daily Note Streak")
	fmt.Println("====================")
	fmt.Println()

	if result.CurrentStreak == 0 {
		fmt.Println("Current Streak: 0 days")
		fmt.Println("💡 Create today's note to start a streak!")
	} else {
		fmt.Printf("Current Streak: %d days 🔥\n", result.CurrentStreak)

		if result.CurrentStreak == 1 {
			fmt.Println("Keep it up! Write tomorrow to continue.")
		} else if result.CurrentStreak < 7 {
			fmt.Println("Great start! Keep going!")
		} else if result.CurrentStreak < 30 {
			fmt.Println("Impressive! You're building a habit!")
		} else {
			fmt.Println("Amazing! You're on fire! 🔥🔥🔥")
		}
	}

	fmt.Printf("\nLongest Streak: %d days\n", result.LongestStreak)
	fmt.Printf("Total Notes: %d\n", result.TotalNotes)
}

func displayRecentActivity(cfg *config.LoadedConfig) {
	today := time.Now()

	fmt.Println("\nRecent Activity:")

	for i := 6; i >= 0; i-- {
		date := today.AddDate(0, 0, -i)

		if !cfg.Streaks.IncludeWeekends {
			weekday := date.Weekday()
			if weekday == time.Saturday || weekday == time.Sunday {
				continue
			}
		}

		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)
		exists := utils.FileExists(notePath)

		dateStr := date.Format("Mon Jan 02")
		if i == 0 {
			dateStr += " (today)"
		}

		status := "✓"
		if !exists {
			status = "○"
		}

		fmt.Printf("  %s %s\n", status, dateStr)
	}
}

func ShowStreak(cfg *config.LoadedConfig) error {
	result := services.CalculateStreak(cfg)
	displayStreakInfo(result)
	displayRecentActivity(cfg)

	return nil
}
