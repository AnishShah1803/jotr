package templatecmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/templates"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var EditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit templates",
	RunE:  runEdit,
}

func init() {
	TemplateCmd.AddCommand(EditCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := config.LoadWithContext(ctx, "")
	if err != nil {
		return err
	}

	templateList, _ := templates.LoadTemplates(cfg)

	if len(templateList) == 0 {
		return fmt.Errorf("no templates found")
	}

	fmt.Println("\nTemplates:")
	for i, tmpl := range templateList {
		fmt.Printf("%d. %s\n", i+1, tmpl.Filename)
	}

	choice := utils.PromptChoice("Select template to edit", 1, len(templateList))

	selected := templateList[choice-1]
	templatePath := filepath.Join(cfg.TemplatesPath, selected.Filename)

	return notes.OpenInEditorWithContext(ctx, templatePath)
}
