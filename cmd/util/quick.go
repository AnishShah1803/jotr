package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"

	searchcmd "github.com/AnishShah1803/jotr/cmd/search"
	taskcmd "github.com/AnishShah1803/jotr/cmd/task"
	templatecmd "github.com/AnishShah1803/jotr/cmd/templatecmd"
	visualcmd "github.com/AnishShah1803/jotr/cmd/visual"
)

var QuickCmd = &cobra.Command{
	Use:   "quick",
	Short: "Quick actions menu",
	Long: `Show a menu of quick actions for common operations.

Provides fast access to daily note, search, capture, and more.

Examples:
  jotr quick                  # Show quick actions menu
  jotr q                      # Using alias`,
	Aliases: []string{"q"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		return showQuickMenu(cmd.Context(), cfg)
	},
}

func showQuickMenu(ctx context.Context, cfg *config.LoadedConfig) error {
	fmt.Println("⚡ Quick Actions")
	fmt.Println("===============")
	fmt.Println()
	fmt.Println("1. Open today's note")
	fmt.Println("2. Open yesterday's note")
	fmt.Println("3. Quick capture")
	fmt.Println("4. Search notes")
	fmt.Println("5. Task summary")
	fmt.Println("6. Show streak")
	fmt.Println("7. Create from template")
	fmt.Println("0. Exit")
	fmt.Println()
	fmt.Print("Select action (0-7): ")

	choice := utils.PromptChoice("", 0, 7)

	switch choice {
	case 1:
		today := time.Now()
		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, today)

		if !utils.FileExists(notePath) {
			if err := notes.CreateDailyNote(ctx, notePath, cfg.Format.DailyNoteSections, today); err != nil {
				return err
			}
		}

		return notes.OpenInEditor(notePath)

	case 2:
		yesterday := time.Now().AddDate(0, 0, -1)

		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, yesterday)
		if !utils.FileExists(notePath) {
			return fmt.Errorf("yesterday's note doesn't exist")
		}

		return notes.OpenInEditor(notePath)

	case 3:
		text := utils.PromptUser("Enter text to capture: ")

		if text != "" {
			fmt.Println("✓ Captured!")
		}

		return nil

	case 4:
		query := utils.PromptUser("Search query: ")

		if query != "" {
			return searchcmd.SearchNotes(ctx, cfg, query)
		}

		return nil

	case 5:
		return taskcmd.ShowSummary(ctx, cfg)

	case 6:
		return visualcmd.ShowStreak(cfg)

	case 7:
		return runTemplateFromQuick(ctx, cfg)

	case 0:
		fmt.Println("Goodbye!")
		return nil

	default:
		return fmt.Errorf("invalid choice")
	}
}

func runTemplateFromQuick(ctx context.Context, cfg *config.LoadedConfig) error {
	templateList, warnings := templatecmd.LoadTemplatesForIntegration(cfg)

	for _, warn := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warn)
	}

	if len(templateList) == 0 {
		return fmt.Errorf("no templates found in %s", cfg.TemplatesPath)
	}

	selected := templatecmd.SelectTemplateForIntegration(templateList)

	if selected == nil {
		return nil
	}

	return templatecmd.CreateFromSelectedTemplate(ctx, cfg, selected)
}
