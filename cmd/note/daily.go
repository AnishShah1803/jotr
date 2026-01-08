package cmd

import (
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
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		dateOption.SetTargetDate()
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

		return openInEditor(notePath)
	},
}

func openInEditor(path string) error {
	editor := config.GetEditor()

	if err := utils.ValidateEditor(editor); err != nil {
		return fmt.Errorf("invalid editor: %w", err)
	}

	execCmd := exec.Command(editor, path)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}
