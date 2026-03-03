package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
)

var outlineCmdFlags = struct {
	file   string
	path   string
	format string
	total  bool
}{}

func init() {
	OutlineCmd.Flags().StringVar(&outlineCmdFlags.file, "file", "", "note name to outline")
	OutlineCmd.Flags().StringVar(&outlineCmdFlags.path, "path", "", "file path to outline")
	OutlineCmd.Flags().StringVar(&outlineCmdFlags.format, "format", "pretty", "output format: pretty, raw")
	OutlineCmd.Flags().BoolVar(&outlineCmdFlags.total, "total", false, "show only the heading count")
}

var OutlineCmd = &cobra.Command{
	Use:   "outline [note-name]",
	Short: "Show heading outline of a note",
	Long: `Extract and display the heading structure of a note.

Examples:
  jotr outline MyNote               # Show outline of MyNote
  jotr outline --file MyNote        # Same with flag
  jotr outline MyNote --format raw  # Raw heading text
  jotr outline MyNote --total       # Show only heading count`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		if outlineCmdFlags.path != "" {
			return outlineFile(outlineCmdFlags.path, outlineCmdFlags.format, outlineCmdFlags.total)
		}

		noteName := outlineCmdFlags.file
		if len(args) > 0 {
			noteName = args[0]
		}

		if noteName == "" {
			return fmt.Errorf("note name required")
		}

		allNotes, err := notes.FindNotes(cmd.Context(), cfg.Paths.BaseDir)
		if err != nil {
			return err
		}

		for _, np := range allNotes {
			if strings.Contains(strings.ToLower(filepath.Base(np)), strings.ToLower(noteName)) {
				return outlineFile(np, outlineCmdFlags.format, outlineCmdFlags.total)
			}
		}

		return fmt.Errorf("note not found: %s", noteName)
	},
}

func outlineFile(path, format string, totalOnly bool) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var headings []string

	for _, line := range lines {
		if trimmed := strings.TrimLeft(line, "#"); strings.HasPrefix(line, "#") && len(trimmed) > 0 && trimmed[0] == ' ' {
			headings = append(headings, line)
		}
	}

	if totalOnly {
		fmt.Printf("%d\n", len(headings))
		return nil
	}

	if len(headings) == 0 {
		fmt.Printf("No headings found in %s\n", filepath.Base(path))
		return nil
	}

	if format == "raw" {
		for _, h := range headings {
			fmt.Println(h)
		}
		return nil
	}

	fmt.Printf("Outline of %s:\n\n", filepath.Base(path))
	for _, h := range headings {
		level := 0
		for _, ch := range h {
			if ch == '#' {
				level++
			} else {
				break
			}
		}
		indent := strings.Repeat("  ", level-1)
		text := strings.TrimSpace(strings.TrimLeft(h, "#"))
		fmt.Printf("%s%s %s\n", indent, strings.Repeat("#", level), text)
	}

	return nil
}
