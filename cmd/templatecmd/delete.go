package templatecmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/templates"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a template",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	TemplateCmd.AddCommand(DeleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := config.LoadWithContext(ctx, "")
	if err != nil {
		return err
	}

	name := args[0]

	templateList, _ := templates.LoadTemplates(cfg)

	var selected *templates.Template
	for _, tmpl := range templateList {
		if tmpl.Name == name || tmpl.Filename == name || tmpl.Filename == name+".md" {
			selected = tmpl
			break
		}
	}

	if selected == nil {
		return fmt.Errorf("template not found: %s", name)
	}

	templatePath := filepath.Join(cfg.TemplatesPath, selected.Filename)

	if !utils.FileExists(templatePath) {
		return fmt.Errorf("template file not found: %s", templatePath)
	}

	fmt.Printf("Delete template '%s'? (y/N): ", selected.Name)

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		fmt.Println("Canceled")
		return nil
	}

	if response != "y" && response != "Y" {
		fmt.Println("Canceled")
		return nil
	}

	if err := os.Remove(templatePath); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	fmt.Printf("Deleted template: %s\n", selected.Name)

	return nil
}
