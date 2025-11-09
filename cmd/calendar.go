package cmd

import (
	"fmt"
	"time"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Show calendar view",
	Long:  `Show a calendar view of daily notes.`,
	Aliases: []string{"cal"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return showCalendar(cfg)
	},
}

func init() {
	rootCmd.AddCommand(calendarCmd)
}

func showCalendar(cfg *config.LoadedConfig) error {
	now := time.Now()
	year := now.Year()
	month := now.Month()

	fmt.Printf("📅 %s %d\n", month.String(), year)
	fmt.Println("====================\n")

	// Print header
	fmt.Println("Su Mo Tu We Th Fr Sa")

	// Get first day of month
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	
	// Print leading spaces
	weekday := int(firstDay.Weekday())
	for i := 0; i < weekday; i++ {
		fmt.Print("   ")
	}

	// Get last day of month
	lastDay := firstDay.AddDate(0, 1, -1).Day()

	// Print days
	for day := 1; day <= lastDay; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)
		
		exists := notes.FileExists(notePath)
		
		// Format day
		dayStr := fmt.Sprintf("%2d", day)
		
		if exists {
			// Highlight days with notes
			if day == now.Day() && month == now.Month() {
				fmt.Printf("\033[1;32m%s\033[0m ", dayStr) // Bold green for today
			} else {
				fmt.Printf("\033[32m%s\033[0m ", dayStr) // Green for notes
			}
		} else {
			if day == now.Day() && month == now.Month() {
				fmt.Printf("\033[1m%s\033[0m ", dayStr) // Bold for today
			} else {
				fmt.Printf("%s ", dayStr) // Normal
			}
		}

		// New line on Saturday
		if (weekday+day)%7 == 0 {
			fmt.Println()
		}
	}

	fmt.Println("\n")
	fmt.Println("Legend:")
	fmt.Println("  \033[32m##\033[0m - Day with note")
	fmt.Println("  \033[1m##\033[0m - Today")
	fmt.Println("  ## - No note")

	// Show stats for the month
	notesThisMonth := 0
	for day := 1; day <= lastDay; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)
		if notes.FileExists(notePath) {
			notesThisMonth++
		}
	}

	fmt.Printf("\nNotes this month: %d/%d days\n", notesThisMonth, lastDay)

	return nil
}

