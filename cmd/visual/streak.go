package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
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

// streakResult holds the calculated streak information.
type streakResult struct {
	currentStreak int
	longestStreak int
	totalNotes    int
}

// calculateStreak computes the current and longest streak for daily notes.
func calculateStreak(cfg *config.LoadedConfig) streakResult {
	today := time.Now()
	result := streakResult{}

	firstValidDay := true
	currentStreakSet := false
	tempStreak := 0

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

		if utils.FileExists(notePath) {
			tempStreak++
			result.totalNotes++

			// Set current streak on first valid day with a note
			if !currentStreakSet {
				result.currentStreak = tempStreak
				currentStreakSet = true
			} else {
				result.currentStreak = tempStreak
			}

			if tempStreak > result.longestStreak {
				result.longestStreak = tempStreak
			}
		} else {
			// Break in streak
			if firstValidDay {
				// No note on first valid day, current streak is 0
				result.currentStreak = 0
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

	return result
}

// displayStreakInfo displays the streak information with motivational messages.
func displayStreakInfo(result streakResult) {
	fmt.Println("🔥 Daily Note Streak")
	fmt.Println("====================")
	fmt.Println()

	if result.currentStreak == 0 {
		fmt.Println("Current Streak: 0 days")
		fmt.Println("💡 Create today's note to start a streak!")
	} else {
		fmt.Printf("Current Streak: %d days 🔥\n", result.currentStreak)

		if result.currentStreak == 1 {
			fmt.Println("Keep it up! Write tomorrow to continue.")
		} else if result.currentStreak < 7 {
			fmt.Println("Great start! Keep going!")
		} else if result.currentStreak < 30 {
			fmt.Println("Impressive! You're building a habit!")
		} else {
			fmt.Println("Amazing! You're on fire! 🔥🔥🔥")
		}
	}

	fmt.Printf("\nLongest Streak: %d days\n", result.longestStreak)
	fmt.Printf("Total Notes: %d\n", result.totalNotes)
}

// displayRecentActivity shows a 7-day calendar of note activity.
func displayRecentActivity(cfg *config.LoadedConfig) {
	today := time.Now()

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

// ShowStreak calculates and displays the daily note streak.
func ShowStreak(cfg *config.LoadedConfig) error {
	result := calculateStreak(cfg)
	displayStreakInfo(result)
	displayRecentActivity(cfg)

	return nil
}
