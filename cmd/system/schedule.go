package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/utils"
)

type ScheduledNote struct {
	Date time.Time `json:"date"`
	Text string    `json:"text"`
	ID   string    `json:"id"`
}

const (
	hoursPerDay = 24
)

var ScheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Schedule notes for future dates",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var scheduleAddCmd = &cobra.Command{
	Use:   "add [date] [text]",
	Short: "Schedule a note for a future date",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return addScheduledNote(cfg, args[0], args[1])
	},
}

var scheduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scheduled notes",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return listScheduledNotes(cfg)
	},
}

var scheduleDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a scheduled note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return deleteScheduledNote(cfg, args[0])
	},
}

func init() {
	ScheduleCmd.AddCommand(scheduleAddCmd, scheduleListCmd, scheduleDeleteCmd)
}

func getScheduleFile(cfg *config.LoadedConfig) string {
	return filepath.Join(cfg.Paths.BaseDir, ".scheduled_notes.json")
}

func loadScheduledNotes(cfg *config.LoadedConfig) ([]ScheduledNote, error) {
	scheduleFile := getScheduleFile(cfg)

	if !utils.FileExists(scheduleFile) {
		return []ScheduledNote{}, nil
	}

	data, err := os.ReadFile(scheduleFile)
	if err != nil {
		return nil, err
	}

	var scheduled []ScheduledNote
	if err := json.Unmarshal(data, &scheduled); err != nil {
		return nil, err
	}

	return scheduled, nil
}

func saveScheduledNotes(cfg *config.LoadedConfig, scheduled []ScheduledNote) error {
	scheduleFile := getScheduleFile(cfg)

	data, err := json.MarshalIndent(scheduled, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(scheduleFile, data, constants.FilePerm0600)
}

func addScheduledNote(cfg *config.LoadedConfig, dateStr, text string) error {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
	}

	if date.Before(time.Now().Truncate(hoursPerDay * time.Hour)) {
		return fmt.Errorf("cannot schedule notes in the past")
	}

	scheduled, err := loadScheduledNotes(cfg)
	if err != nil {
		return err
	}

	newNote := ScheduledNote{
		Date: date,
		Text: text,
		ID:   fmt.Sprintf("%d", time.Now().Unix()),
	}

	scheduled = append(scheduled, newNote)

	if err := saveScheduledNotes(cfg, scheduled); err != nil {
		return err
	}

	fmt.Printf("✓ Scheduled note for %s\n", date.Format("2006-01-02"))
	fmt.Printf("  %s\n", text)
	fmt.Printf("  ID: %s\n", newNote.ID)

	return nil
}

func listScheduledNotes(cfg *config.LoadedConfig) error {
	scheduled, err := loadScheduledNotes(cfg)
	if err != nil {
		return err
	}

	if len(scheduled) == 0 {
		fmt.Println("No scheduled notes")
		return nil
	}

	fmt.Println("Scheduled Notes:")
	fmt.Println()

	for _, note := range scheduled {
		daysUntil := int(time.Until(note.Date).Hours() / hoursPerDay)
		fmt.Printf("  [%s] %s\n", note.ID, note.Date.Format("2006-01-02"))
		fmt.Printf("    %s\n", note.Text)

		switch {
		case daysUntil == 0:
			fmt.Println("    (Today!)")
		case daysUntil == 1:
			fmt.Println("    (Tomorrow)")
		case daysUntil > 0:
			fmt.Printf("    (in %d days)\n", daysUntil)
		default:
			fmt.Printf("    (overdue by %d days)\n", -daysUntil)
		}

		fmt.Println()
	}

	return nil
}

func deleteScheduledNote(cfg *config.LoadedConfig, id string) error {
	scheduled, err := loadScheduledNotes(cfg)
	if err != nil {
		return err
	}

	found := false
	newScheduled := []ScheduledNote{}

	for _, note := range scheduled {
		if note.ID == id {
			found = true

			fmt.Printf("✓ Deleted scheduled note: %s\n", note.Text)
		} else {
			newScheduled = append(newScheduled, note)
		}
	}

	if !found {
		return fmt.Errorf("scheduled note not found: %s", id)
	}

	return saveScheduledNotes(cfg, newScheduled)
}
