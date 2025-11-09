package cmd

import (
	"fmt"
	"time"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/spf13/cobra"
)

var streakCmd = &cobra.Command{
	Use:   "streak",
	Short: "Show daily note streak",
	Long:  `Show your current streak of consecutive daily notes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return showStreak(cfg)
	},
}

func init() {
	rootCmd.AddCommand(streakCmd)
}

func showStreak(cfg *config.LoadedConfig) error {
	today := time.Now()
	currentStreak := 0
	longestStreak := 0
	tempStreak := 0
	totalNotes := 0
	firstValidDay := true
	currentStreakSet := false

	// Check backwards from today
	for i := 0; i < 365; i++ {
		date := today.AddDate(0, 0, -i)

		// Skip weekends if configured
		if !cfg.Streaks.IncludeWeekends {
			weekday := date.Weekday()
			if weekday == time.Saturday || weekday == time.Sunday {
				continue
			}
		}

		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)

		if notes.FileExists(notePath) {
			tempStreak++
			totalNotes++

			// Set current streak on first valid day with a note
			if !currentStreakSet {
				currentStreak = tempStreak
				currentStreakSet = true
			} else {
				currentStreak = tempStreak
			}

			if tempStreak > longestStreak {
				longestStreak = tempStreak
			}
		} else {
			// Break in streak
			if firstValidDay {
				// No note on first valid day, current streak is 0
				currentStreak = 0
				break
			}
			// If we had a streak going, it's now broken
			if currentStreakSet {
				break
			}
			tempStreak = 0
		}

		firstValidDay = false
	}

	// Display streak
	fmt.Println("🔥 Daily Note Streak")
	fmt.Println("====================\n")

	if currentStreak == 0 {
		fmt.Println("Current Streak: 0 days")
		fmt.Println("💡 Create today's note to start a streak!")
	} else {
		fmt.Printf("Current Streak: %d days 🔥\n", currentStreak)
		
		if currentStreak == 1 {
			fmt.Println("Keep it up! Write tomorrow to continue.")
		} else if currentStreak < 7 {
			fmt.Println("Great start! Keep going!")
		} else if currentStreak < 30 {
			fmt.Println("Impressive! You're building a habit!")
		} else {
			fmt.Println("Amazing! You're on fire! 🔥🔥🔥")
		}
	}

	fmt.Printf("\nLongest Streak: %d days\n", longestStreak)
	fmt.Printf("Total Notes: %d\n", totalNotes)

	// Show recent activity
	fmt.Println("\nRecent Activity:")
	for i := 6; i >= 0; i-- {
		date := today.AddDate(0, 0, -i)
		
		// Skip weekends if configured
		if !cfg.Streaks.IncludeWeekends {
			weekday := date.Weekday()
			if weekday == time.Saturday || weekday == time.Sunday {
				continue
			}
		}

		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)
		exists := notes.FileExists(notePath)
		
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

	return nil
}

