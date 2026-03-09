package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/options"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var dateOption = options.NewDateOption()
var outputOption = options.NewOutputOption()

func init() {
	dateOption.AddFlags(DailyCmd)
	outputOption.AddFlags(DailyCmd)
}

var DailyCmd = &cobra.Command{
	Use:     "daily",
	Short:   "Create or open daily note",
	Long:    `Create or open today's daily note.`,
	Aliases: []string{"d"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		dateOption.SetTargetDate()
		opt := ensureDailyContentDate(&dateOption)

		if len(args) > 0 {
			switch args[0] {
			case "append":
				return appendDailyNote(cmd.Context(), cfg, args[1:], opt)
			case "prepend":
				return prependDailyNote(cmd.Context(), cfg, args[1:], opt)
			case "read":
				return readDailyNote(cmd.Context(), cfg, opt)
			default:
				return fmt.Errorf("unknown daily action: %s (use append, prepend, or read)", args[0])
			}
		}

		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, dateOption.Date)

		if outputOption.PathOnly {
			fmt.Println(notePath)
			return nil
		}

		if _, err := os.Stat(notePath); os.IsNotExist(err) {
			sections := notes.BuildDailyNoteSections(cfg)
			if err := notes.CreateDailyNote(cmd.Context(), notePath, sections, dateOption.Date); err != nil {
				return fmt.Errorf("failed to create daily note: %w", err)
			}
			fmt.Printf("✓ Created: %s\n", notePath)
		}

		return openInEditor(cmd.Context(), notePath)
	},
}

func openInEditor(ctx context.Context, path string) error {
	editor := config.GetEditorWithContext(ctx)
	if editor == "" {
		return fmt.Errorf("no editor configured - set EDITOR environment variable or configure editor.default")
	}

	if err := utils.ValidateEditor(editor); err != nil {
		return fmt.Errorf("invalid editor: %w", err)
	}

	execCmd := exec.Command(editor, path)
	if os.Getenv("JOTR_REPL_MODE") != "true" {
		execCmd.Stdin = os.Stdin
	}
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}
