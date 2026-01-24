package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var captureTask bool

var CaptureCmd = &cobra.Command{
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

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		text := strings.Join(args, " ")
		return captureText(cmd.Context(), cfg, text)
	},
}

func captureText(ctx context.Context, cfg *config.LoadedConfig, text string) error {
	today := time.Now()
	notePath := notes.BuildDailyNotePath(cfg.DiaryPath, today)

	if !utils.FileExists(notePath) {
		if err := notes.CreateDailyNote(ctx, notePath, cfg.Format.DailyNoteSections, today); err != nil {
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

	insertIndex := utils.FindSectionIndex(lines, captureSection)

	// If section not found, add it at the end
	if insertIndex == -1 {
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
