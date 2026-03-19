package services

import (
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// StreakResult holds the calculated streak information.
type StreakResult struct {
	CurrentStreak int
	LongestStreak int
	TotalNotes    int
}

// CalculateStreak computes the current and longest streak for daily notes.
func CalculateStreak(cfg *config.LoadedConfig) StreakResult {
	today := time.Now()
	result := StreakResult{}

	firstValidDay := true
	tempStreak := 0

	for i := 0; i < 365; i++ {
		date := today.AddDate(0, 0, -i)

		if !cfg.Streaks.IncludeWeekends {
			weekday := date.Weekday()
			if weekday == time.Saturday || weekday == time.Sunday {
				continue
			}
		}

		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)

		if utils.FileExists(notePath) {
			tempStreak++
			result.TotalNotes++

			result.CurrentStreak = tempStreak

			if tempStreak > result.LongestStreak {
				result.LongestStreak = tempStreak
			}
		} else {
			if firstValidDay {
				result.CurrentStreak = 0
				break
			}
			if tempStreak > 0 {
				break
			}

			tempStreak = 0
		}

		firstValidDay = false
	}

	return result
}
