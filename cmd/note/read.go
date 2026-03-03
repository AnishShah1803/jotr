package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/output"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var ReadCmd = &cobra.Command{
	Use:   "read",
	Short: "Read note contents",
	Long: `Display the contents of a note in the terminal.

You can read notes by name (fuzzy match) or by exact path.

Examples:
  jotr read --file "My Note"
  jotr read --path "diary/2024-01-15.md"
  jotr read --file "Project" --lines 50
  jotr read --path "notes/ideas.md" --format raw`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileFlag, _ := cmd.Flags().GetString("file")
		pathFlag, _ := cmd.Flags().GetString("path")
		linesFlag, _ := cmd.Flags().GetInt("lines")
		formatFlag, _ := cmd.Flags().GetString("format")

		ctx := cmd.Context()
		cfg, err := config.LoadWithContext(ctx, "")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		var notePath string

		if fileFlag != "" {
			notePath, err = findNoteByName(ctx, cfg, fileFlag)
			if err != nil {
				return err
			}
		} else if pathFlag != "" {
			notePath = filepath.Join(cfg.Paths.BaseDir, pathFlag)
			if !utils.FileExists(notePath) {
				return fmt.Errorf("note not found: %s", pathFlag)
			}
		} else {
			return fmt.Errorf("either --file or --path must be specified")
		}

		content, err := os.ReadFile(notePath)
		if err != nil {
			return fmt.Errorf("failed to read note: %w", err)
		}

		contentStr := string(content)
		if linesFlag > 0 {
			lines := strings.Split(contentStr, "\n")
			if len(lines) > linesFlag {
				lines = lines[:linesFlag]
				contentStr = strings.Join(lines, "\n")
				contentStr += fmt.Sprintf("\n\n%s (%d more lines)", lipgloss.NewStyle().Foreground(output.SecondaryColor).Render("..."), len(strings.Split(string(content), "\n"))-linesFlag)
			}
		}

		switch formatFlag {
		case "raw":
			fmt.Print(contentStr)
		case "pretty":
			fmt.Println(formatPretty(contentStr, filepath.Base(notePath)))
		default:
			fmt.Println(formatPretty(contentStr, filepath.Base(notePath)))
		}

		return nil
	},
}

func findNoteByName(ctx context.Context, cfg *config.LoadedConfig, query string) (string, error) {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return "", fmt.Errorf("failed to find notes: %w", err)
	}

	if len(allNotes) == 0 {
		return "", fmt.Errorf("no notes found in vault")
	}

	query = strings.ToLower(query)
	var matches []string

	for _, notePath := range allNotes {
		name := strings.ToLower(filepath.Base(notePath))
		nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))

		if name == query || nameWithoutExt == query {
			return notePath, nil
		}
		if strings.Contains(name, query) {
			matches = append(matches, notePath)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no note found matching: %s", query)
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	fmt.Println(lipgloss.NewStyle().Foreground(output.SecondaryColor).Render("Multiple notes found:"))
	for i, notePath := range matches {
		relPath, _ := filepath.Rel(cfg.Paths.BaseDir, notePath)
		fmt.Printf("  %d. %s\n", i+1, relPath)
	}

	fmt.Print("\nSelect note (1-", len(matches), "): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}

	var selection int
	if _, err := fmt.Sscanf(strings.TrimSpace(input), "%d", &selection); err != nil {
		return "", fmt.Errorf("invalid selection")
	}

	if selection < 1 || selection > len(matches) {
		return "", fmt.Errorf("selection out of range")
	}

	return matches[selection-1], nil
}

func formatPretty(content, filename string) string {
	var result strings.Builder

	result.WriteString(lipgloss.NewStyle().Foreground(output.PrimaryColor).Render("─"))
	result.WriteString(lipgloss.NewStyle().Foreground(output.PrimaryColor).Render(strings.Repeat("─", len(filename)+2)))
	result.WriteString(lipgloss.NewStyle().Foreground(output.PrimaryColor).Render("─\n"))
	result.WriteString(lipgloss.NewStyle().Foreground(output.PrimaryColor).Render("│ "))
	result.WriteString(lipgloss.NewStyle().Foreground(output.SecondaryColor).Render(filename))
	result.WriteString(lipgloss.NewStyle().Foreground(output.PrimaryColor).Render(" │\n"))
	result.WriteString(lipgloss.NewStyle().Foreground(output.PrimaryColor).Render("─"))
	result.WriteString(lipgloss.NewStyle().Foreground(output.PrimaryColor).Render(strings.Repeat("─", len(filename)+2)))
	result.WriteString(lipgloss.NewStyle().Foreground(output.PrimaryColor).Render("─\n\n"))

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			result.WriteString(lipgloss.NewStyle().Foreground(output.AccentColor).Render(line) + "\n")
		} else if strings.HasPrefix(line, "## ") {
			result.WriteString(lipgloss.NewStyle().Foreground(output.SecondaryColor).Render(line) + "\n")
		} else if strings.HasPrefix(line, "- [ ]") {
			result.WriteString(lipgloss.NewStyle().Foreground(output.SecondaryColor).Render("- [ ]") + line[5:] + "\n")
		} else if strings.HasPrefix(line, "- [x]") {
			result.WriteString(lipgloss.NewStyle().Foreground(output.AccentColor).Render("- [x]") + line[5:] + "\n")
		} else if strings.HasPrefix(line, "-") {
			result.WriteString(lipgloss.NewStyle().Foreground(output.SecondaryColor).Render("-") + line[1:] + "\n")
		} else {
			result.WriteString(line + "\n")
		}
	}

	return result.String()
}

func init() {
	ReadCmd.Flags().String("file", "", "Note name to read (fuzzy match)")
	ReadCmd.Flags().String("path", "", "Exact path to note (relative to base dir)")
	ReadCmd.Flags().Int("lines", 0, "Limit output to N lines (0 = unlimited)")
	ReadCmd.Flags().String("format", "pretty", "Output format: pretty, raw")
}
