package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/notes"
)

func quoteFrontmatterValue(v string) string {
	needsQuoting := strings.ContainsAny(v, ":\n#") ||
		strings.HasPrefix(v, " ") || strings.HasSuffix(v, " ") ||
		len(v) > 0 && strings.ContainsAny(string(v[0]), "[]{},|>&*!?@`'\"")
	if !needsQuoting {
		return v
	}
	return `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
}

var FrontmatterCmd = &cobra.Command{
	Use:     "frontmatter",
	Short:   "Manage note frontmatter",
	Aliases: []string{"fm"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var frontmatterListCmd = &cobra.Command{
	Use:   "list [note]",
	Short: "Show all frontmatter fields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return showFrontmatter(cmd.Context(), cfg, args[0])
	},
}

var frontmatterGetCmd = &cobra.Command{
	Use:   "get [note] [key]",
	Short: "Get a specific frontmatter value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return getFrontmatter(cmd.Context(), cfg, args[0], args[1])
	},
}

var frontmatterSetCmd = &cobra.Command{
	Use:   "set [note] key=value",
	Short: "Set or update a frontmatter field",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return setFrontmatter(cmd.Context(), cfg, args[0], args[1])
	},
}

var frontmatterRemoveCmd = &cobra.Command{
	Use:   "remove [note] [key]",
	Short: "Remove a frontmatter field",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return removeFrontmatter(cmd.Context(), cfg, args[0], args[1])
	},
}

func init() {
	FrontmatterCmd.AddCommand(frontmatterListCmd, frontmatterGetCmd, frontmatterSetCmd, frontmatterRemoveCmd)
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
		newLines = append(newLines, "---")
		updated := false

		for i := 1; i < len(lines); i++ {
			if lines[i] == "---" {
				if !updated {
					newLines = append(newLines, fmt.Sprintf("%s: %s", key, quoteFrontmatterValue(value)))
				}

				newLines = append(newLines, lines[i:]...)

				break
			}

			if strings.HasPrefix(lines[i], key+":") {
				newLines = append(newLines, fmt.Sprintf("%s: %s", key, quoteFrontmatterValue(value)))
				updated = true
			} else {
				newLines = append(newLines, lines[i])
			}
		}
	} else {
		newLines = append(newLines, "---")
		newLines = append(newLines, fmt.Sprintf("%s: %s", key, quoteFrontmatterValue(value)))
		newLines = append(newLines, "---")
		newLines = append(newLines, "")
		newLines = append(newLines, lines...)
	}

	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(targetNote, []byte(newContent), constants.FilePerm0644); err != nil {
		return err
	}

	fmt.Printf("✓ Updated %s: %s = %s\n", filepath.Base(targetNote), key, value)

	return nil
}

func getFrontmatter(ctx context.Context, cfg *config.LoadedConfig, noteName string, key string) error {
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
		return fmt.Errorf("no frontmatter in %s", filepath.Base(targetNote))
	}

	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			break
		}
		if strings.HasPrefix(lines[i], key+":") {
			parts := strings.SplitN(lines[i], ":", 2)
			fmt.Printf("%s\n", strings.TrimSpace(parts[1]))
			return nil
		}
	}

	return fmt.Errorf("key not found: %s", key)
}

func removeFrontmatter(ctx context.Context, cfg *config.LoadedConfig, noteName string, key string) error {
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
	var newLines []string

	if len(lines) > 0 && lines[0] == "---" {
		newLines = append(newLines, lines[0])
		inFrontmatter := true
		for _, line := range lines[1:] {
			if inFrontmatter && line == "---" {
				inFrontmatter = false
				newLines = append(newLines, line)
			} else if !inFrontmatter || !strings.HasPrefix(line, key+":") {
				newLines = append(newLines, line)
			}
		}
	} else {
		newLines = lines
	}

	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(targetNote, []byte(newContent), constants.FilePerm0644); err != nil {
		return err
	}

	fmt.Printf("✓ Removed %s from %s\n", key, filepath.Base(targetNote))
	return nil
}
