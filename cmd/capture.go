package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/anish/jotr/internal/utils"
	"github.com/spf13/cobra"
)

var captureTask bool

var captureCmd = &cobra.Command{
	Use:   "capture [text]",
	Short: "Quick capture to daily note",
	Long: `Quickly capture text to today's daily note.
	
Examples:
  jotr capture "Meeting with team"
  jotr capture --task "Review PR #123"
  jotr cap "Quick thought"`,
	Aliases: []string{"cap"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("text to capture is required")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		text := strings.Join(args, " ")
		return captureText(cfg, text)
	},
}

func init() {
	captureCmd.Flags().BoolVar(&captureTask, "task", false, "Capture as task")
	rootCmd.AddCommand(captureCmd)
}

func captureText(cfg *config.LoadedConfig, text string) error {
	today := time.Now()
	notePath := notes.BuildDailyNotePath(cfg.DiaryPath, today)

	if !notes.FileExists(notePath) {
		if err := notes.CreateDailyNote(notePath, cfg.Format.DailyNoteSections, today); err != nil {
			return fmt.Errorf("failed to create daily note: %w", err)
		}
	}

	content, err := os.ReadFile(notePath)
	if err != nil {
		return fmt.Errorf("failed to read note: %w", err)
	}

	// Find the capture section
	captureSection := cfg.Format.CaptureSection
	if captureSection == "" {
		captureSection = "Captured"
	}

	lines := strings.Split(string(content), "\n")
	sectionFound := false
	insertIndex := -1

	for i, line := range lines {
		if strings.HasPrefix(line, "## "+captureSection) {
			sectionFound = true
			insertIndex = i + 1
			// Skip empty lines after section header
			for insertIndex < len(lines) && strings.TrimSpace(lines[insertIndex]) == "" {
				insertIndex++
			}
			break
		}
	}

	// If section not found, add it at the end
	if !sectionFound {
		lines = append(lines, "", fmt.Sprintf("## %s", captureSection), "")
		insertIndex = len(lines)
	}

	// Format the captured text
	timestamp := time.Now().Format("15:04")
	var capturedLine string
	if captureTask {
		capturedLine = fmt.Sprintf("- [ ] %s (%s)", text, timestamp)
	} else {
		capturedLine = fmt.Sprintf("- %s (%s)", text, timestamp)
	}

	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIndex]...)
	newLines = append(newLines, capturedLine)
	newLines = append(newLines, lines[insertIndex:]...)

	newContent := strings.Join(newLines, "\n")
	if err := utils.AtomicWriteFile(notePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write note: %w", err)
	}

	fmt.Printf("✓ Captured to: %s\n", notePath)
	fmt.Printf("  %s\n", capturedLine)

	return nil
}

