package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/anish/jotr/internal/config"
	"github.com/spf13/cobra"
)

var (
	yesterday bool
	pathOnly  bool
)

var dailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Create or open daily note",
	Long: `Create or open today's daily note.
	
Examples:
  jotr daily              # Open today's note
  jotr daily --yesterday  # Open yesterday's note
  jotr daily --path       # Show path only`,
	Aliases: []string{"d"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Determine target date
		targetDate := time.Now()
		if yesterday {
			targetDate = targetDate.AddDate(0, 0, -1)
		}

		// Build note path
		notePath := buildDailyNotePath(cfg, targetDate)

		// If path only, just print and exit
		if pathOnly {
			fmt.Println(notePath)
			return nil
		}

		// Create note if it doesn't exist
		if _, err := os.Stat(notePath); os.IsNotExist(err) {
			if err := createDailyNote(notePath, cfg, targetDate); err != nil {
				return fmt.Errorf("failed to create daily note: %w", err)
			}
			fmt.Printf("✓ Created: %s\n", notePath)
		}

		// Open in editor
		return openInEditor(notePath)
	},
}

func init() {
	dailyCmd.Flags().BoolVar(&yesterday, "yesterday", false, "Open yesterday's note")
	dailyCmd.Flags().BoolVarP(&pathOnly, "path", "p", false, "Show path only")
}

func buildDailyNotePath(cfg *config.LoadedConfig, date time.Time) string {
	// Format: {year}/{month_num}-{month_abbr}/{year}-{month}-{day}-{weekday}.md
	year := date.Format("2006")
	monthNum := date.Format("01")
	monthAbbr := date.Format("Jan")
	month := date.Format("01")
	day := date.Format("02")
	weekday := date.Format("Mon")

	dirPath := filepath.Join(cfg.DiaryPath, year, fmt.Sprintf("%s-%s", monthNum, monthAbbr))
	filename := fmt.Sprintf("%s-%s-%s-%s.md", year, month, day, weekday)

	return filepath.Join(dirPath, filename)
}

func createDailyNote(notePath string, cfg *config.LoadedConfig, date time.Time) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(notePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create note with template
	content := fmt.Sprintf("# %s\n\n", date.Format("2006-01-02-Mon"))
	
	// Add sections from config
	for _, section := range cfg.Format.DailyNoteSections {
		content += fmt.Sprintf("## %s\n\n", section)
	}

	return os.WriteFile(notePath, []byte(content), 0644)
}

func openInEditor(path string) error {
	// Try to get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Default to vim
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

