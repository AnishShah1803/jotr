package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/spf13/cobra"
)

type ScheduledNote struct {
	Date time.Time `json:"date"`
	Text string    `json:"text"`
	ID   string    `json:"id"`
}

var scheduleCmd = &cobra.Command{
	Use:   "schedule [action]",
	Short: "Schedule notes for future dates",
	Long: `Schedule notes to be created on future dates.
	
Actions:
  add [date] [text]    Schedule a note
  list                 List scheduled notes
  delete [id]          Delete scheduled note
  
Examples:
  jotr schedule add 2025-02-01 "Q1 Review"
  jotr schedule list
  jotr schedule delete abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("action required: add, list, or delete")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		action := args[0]

		switch action {
		case "add":
			if len(args) < 3 {
				return fmt.Errorf("usage: schedule add [date] [text]")
			}
			return addScheduledNote(cfg, args[1], args[2])
		case "list":
			return listScheduledNotes(cfg)
		case "delete":
			if len(args) < 2 {
				return fmt.Errorf("usage: schedule delete [id]")
			}
			return deleteScheduledNote(cfg, args[1])
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}

func getScheduleFile(cfg *config.LoadedConfig) string {
	return filepath.Join(cfg.Paths.BaseDir, ".scheduled_notes.json")
}

func loadScheduledNotes(cfg *config.LoadedConfig) ([]ScheduledNote, error) {
	scheduleFile := getScheduleFile(cfg)
	
	if !notes.FileExists(scheduleFile) {
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

	return os.WriteFile(scheduleFile, data, 0644)
}

func addScheduledNote(cfg *config.LoadedConfig, dateStr, text string) error {
	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
	}

	// Check if date is in the past
	if date.Before(time.Now().Truncate(24 * time.Hour)) {
		return fmt.Errorf("cannot schedule notes in the past")
	}

	// Load existing scheduled notes
	scheduled, err := loadScheduledNotes(cfg)
	if err != nil {
		return err
	}

	// Create new scheduled note
	newNote := ScheduledNote{
		Date: date,
		Text: text,
		ID:   fmt.Sprintf("%d", time.Now().Unix()),
	}

	scheduled = append(scheduled, newNote)

	// Save
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

	fmt.Println("Scheduled Notes:\n")
	for _, note := range scheduled {
		daysUntil := int(note.Date.Sub(time.Now()).Hours() / 24)
		fmt.Printf("  [%s] %s\n", note.ID, note.Date.Format("2006-01-02"))
		fmt.Printf("    %s\n", note.Text)
		if daysUntil == 0 {
			fmt.Println("    (Today!)")
		} else if daysUntil == 1 {
			fmt.Println("    (Tomorrow)")
		} else if daysUntil > 0 {
			fmt.Printf("    (in %d days)\n", daysUntil)
		} else {
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

	// Find and remove
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

