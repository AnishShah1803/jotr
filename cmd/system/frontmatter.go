package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
)

var FrontmatterCmd = &cobra.Command{
	Use:   "frontmatter [note-name]",
	Short: "Manage note frontmatter",
	Long: `View or edit frontmatter in notes.
	
Examples:
  jotr frontmatter MyNote        # Show frontmatter
  jotr frontmatter MyNote --set status=done`,
	Aliases: []string{"fm"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("note name required")
		}

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		noteName := args[0]
		setValue, _ := cmd.Flags().GetString("set")

		if setValue != "" {
			return setFrontmatter(cmd.Context(), cfg, noteName, setValue)
		}
		return showFrontmatter(cmd.Context(), cfg, noteName)
	},
}

func showFrontmatter(ctx context.Context, cfg *config.LoadedConfig, noteName string) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	var targetNote string

	for _, note := range allNotes {
		if strings.Contains(strings.ToLower(filepath.Base(note)), strings.ToLower(noteName)) {
			targetNote = note
			break
		}
	}

	if targetNote == "" {
		return fmt.Errorf("note not found: %s", noteName)
	}

	content, err := os.ReadFile(targetNote)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	if len(lines) < 3 || lines[0] != "---" {
		fmt.Printf("No frontmatter in %s\n", filepath.Base(targetNote))
		return nil
	}

	endIdx := -1

	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		fmt.Printf("Invalid frontmatter in %s\n", filepath.Base(targetNote))
		return nil
	}

	fmt.Printf("Frontmatter in %s:\n\n", filepath.Base(targetNote))

	for i := 1; i < endIdx; i++ {
		fmt.Printf("  %s\n", lines[i])
	}

	return nil
}

func setFrontmatter(ctx context.Context, cfg *config.LoadedConfig, noteName string, setValue string) error {
	parts := strings.SplitN(setValue, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, use: key=value")
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	var targetNote string

	for _, note := range allNotes {
		if strings.Contains(strings.ToLower(filepath.Base(note)), strings.ToLower(noteName)) {
			targetNote = note
			break
		}
	}

	if targetNote == "" {
		return fmt.Errorf("note not found: %s", noteName)
	}

	content, err := os.ReadFile(targetNote)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	newLines := []string{}
	if len(lines) > 0 && lines[0] == "---" {
		// Has frontmatter, update it
		newLines = append(newLines, "---")
		updated := false

		for i := 1; i < len(lines); i++ {
			if lines[i] == "---" {
				if !updated {
					newLines = append(newLines, fmt.Sprintf("%s: %s", key, value))
				}

				newLines = append(newLines, lines[i:]...)

				break
			}

			if strings.HasPrefix(lines[i], key+":") {
				newLines = append(newLines, fmt.Sprintf("%s: %s", key, value))
				updated = true
			} else {
				newLines = append(newLines, lines[i])
			}
		}
	} else {
		// No frontmatter, add it
		newLines = append(newLines, "---")
		newLines = append(newLines, fmt.Sprintf("%s: %s", key, value))
		newLines = append(newLines, "---")
		newLines = append(newLines, "")
		newLines = append(newLines, lines...)
	}

	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(targetNote, []byte(newContent), 0644); err != nil {
		return err
	}

	fmt.Printf("✓ Updated %s: %s = %s\n", filepath.Base(targetNote), key, value)

	return nil
}
