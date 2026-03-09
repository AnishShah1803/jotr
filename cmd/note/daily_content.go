package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/options"
)

func appendDailyNote(_ context.Context, cfg *config.LoadedConfig, args []string, dateOpt *options.DateOption) error {
	notePath := notes.BuildDailyNotePath(cfg.DiaryPath, dateOpt.Date)

	content := ""
	if len(args) > 0 {
		content = strings.Join(args, " ")
	} else {
		if isReplMode() {
			return fmt.Errorf("content argument required in REPL mode: use 'daily append \"your content here\"'")
		}
		fmt.Print("Content to append: ")
		input, err := defaultReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read content: %w", err)
		}
		content = strings.TrimSpace(input)
	}

	if content == "" {
		return fmt.Errorf("content is required")
	}

	existing, err := os.ReadFile(notePath)
	if err != nil {
		return fmt.Errorf("failed to read daily note (does it exist?): %w", err)
	}

	newContent := strings.TrimRight(string(existing), "\n") + "\n" + content + "\n"
	if err := os.WriteFile(notePath, []byte(newContent), constants.FilePerm0644); err != nil {
		return fmt.Errorf("failed to write daily note: %w", err)
	}

	fmt.Printf("✓ Appended to: %s\n", notePath)
	return nil
}

func prependDailyNote(_ context.Context, cfg *config.LoadedConfig, args []string, dateOpt *options.DateOption) error {
	notePath := notes.BuildDailyNotePath(cfg.DiaryPath, dateOpt.Date)

	content := ""
	if len(args) > 0 {
		content = strings.Join(args, " ")
	} else {
		if isReplMode() {
			return fmt.Errorf("content argument required in REPL mode: use 'daily prepend \"your content here\"'")
		}
		fmt.Print("Content to prepend: ")
		input, err := defaultReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read content: %w", err)
		}
		content = strings.TrimSpace(input)
	}

	if content == "" {
		return fmt.Errorf("content is required")
	}

	existing, err := os.ReadFile(notePath)
	if err != nil {
		return fmt.Errorf("failed to read daily note (does it exist?): %w", err)
	}

	lines := strings.Split(string(existing), "\n")

	insertAt := 0
	if len(lines) > 0 && lines[0] == "---" {
		for i := 1; i < len(lines); i++ {
			if lines[i] == "---" {
				insertAt = i + 1
				break
			}
		}
	}

	result := make([]string, 0, len(lines)+2)
	result = append(result, lines[:insertAt]...)
	result = append(result, content)
	result = append(result, lines[insertAt:]...)

	newContent := strings.Join(result, "\n")
	if err := os.WriteFile(notePath, []byte(newContent), constants.FilePerm0644); err != nil {
		return fmt.Errorf("failed to write daily note: %w", err)
	}

	fmt.Printf("✓ Prepended to: %s\n", notePath)
	return nil
}

func readDailyNote(_ context.Context, cfg *config.LoadedConfig, dateOpt *options.DateOption) error {
	notePath := notes.BuildDailyNotePath(cfg.DiaryPath, dateOpt.Date)

	content, err := os.ReadFile(notePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("daily note does not exist for %s", dateOpt.Date.Format("2006-01-02"))
		}
		return fmt.Errorf("failed to read daily note: %w", err)
	}

	fmt.Print(string(content))
	return nil
}

func ensureDailyContentDate(dateOpt *options.DateOption) *options.DateOption {
	if dateOpt.Date.IsZero() {
		dateOpt.Date = time.Now()
	}
	return dateOpt
}
