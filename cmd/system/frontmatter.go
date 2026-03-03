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

var FrontmatterCmd = &cobra.Command{
	Use:   "frontmatter [action] [note-name]",
	Short: "Manage note frontmatter",
	Long: `View or edit frontmatter in notes.
	
Actions:
  list [note]         Show all frontmatter fields (default)
  get [note] [key]    Get a specific frontmatter value
  set [note] key=val  Set a frontmatter key to a value
  remove [note] [key] Remove a frontmatter key

Examples:
  jotr frontmatter MyNote              # Show frontmatter
  jotr frontmatter list MyNote         # Same as above
  jotr frontmatter get MyNote status   # Get 'status' field
  jotr frontmatter set MyNote status=done
  jotr frontmatter remove MyNote status
  jotr frontmatter MyNote --set status=done  # Legacy flag syntax`,
	Aliases: []string{"fm"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("note name required")
		}

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		action := args[0]
		switch action {
		case "list":
			if len(args) < 2 {
				return fmt.Errorf("note name required")
			}
			return showFrontmatter(cmd.Context(), cfg, args[1])
		case "get":
			if len(args) < 3 {
				return fmt.Errorf("usage: frontmatter get <note> <key>")
			}
			return getFrontmatter(cmd.Context(), cfg, args[1], args[2])
		case "set":
			if len(args) < 3 {
				return fmt.Errorf("usage: frontmatter set <note> key=value")
			}
			return setFrontmatter(cmd.Context(), cfg, args[1], args[2])
		case "remove":
			if len(args) < 3 {
				return fmt.Errorf("usage: frontmatter remove <note> <key>")
			}
			return removeFrontmatter(cmd.Context(), cfg, args[1], args[2])
		default:
			setValue, _ := cmd.Flags().GetString("set")
			if setValue != "" {
				return setFrontmatter(cmd.Context(), cfg, action, setValue)
			}
			return showFrontmatter(cmd.Context(), cfg, action)
		}
	},
}

func init() {
	FrontmatterCmd.Flags().String("set", "", "set a frontmatter key=value (legacy flag syntax)")
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
		newLines = append(newLines, "---")
		newLines = append(newLines, fmt.Sprintf("%s: %s", key, value))
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

	// Only strip the key within the frontmatter block.
	if len(lines) > 0 && lines[0] == "---" {
		newLines = append(newLines, lines[0])
		inFrontmatter := true
		for _, line := range lines[1:] {
			if inFrontmatter && line == "---" {
				inFrontmatter = false
				newLines = append(newLines, line)
			} else if inFrontmatter && strings.HasPrefix(line, key+":") {
				// drop this line
			} else {
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
