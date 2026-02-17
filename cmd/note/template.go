package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var TemplateCmd = &cobra.Command{
	Use:   "template [action]",
	Short: "Manage note templates",
	Long: `Manage note templates.
	
Actions:
  list              List all templates
  create [name]     Create a new template
  edit [name]       Edit a template
  delete [name]     Delete a template
  
Examples:
  jotr template list
  jotr template create meeting
  jotr template edit meeting`,
	Aliases: []string{"tmpl"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("action required: list, create, edit, or delete")
		}

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		action := args[0]

		switch action {
		case "list":
			return listTemplates(cfg)
		case "create":
			if len(args) < 2 {
				return fmt.Errorf("template name required")
			}
			return createTemplate(cfg, args[1])
		case "edit":
			if len(args) < 2 {
				return fmt.Errorf("template name required")
			}
			return editTemplate(cfg, args[1])
		case "delete":
			if len(args) < 2 {
				return fmt.Errorf("template name required")
			}
			return deleteTemplate(cfg, args[1])
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
}

func getTemplateDir(cfg *config.LoadedConfig) string {
	return filepath.Join(cfg.Paths.BaseDir, ".templates")
}

func listTemplates(cfg *config.LoadedConfig) error {
	templateDir := getTemplateDir(cfg)

	if !utils.FileExists(templateDir) {
		fmt.Println("No templates found")
		return nil
	}

	files, err := os.ReadDir(templateDir)
	if err != nil {
		return fmt.Errorf("failed to read templates: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No templates found")
		return nil
	}

	fmt.Println("Available templates:")
	fmt.Println()

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".md" {
			name := file.Name()[:len(file.Name())-3]
			fmt.Printf("  %s\n", name)
		}
	}

	return nil
}

func createTemplate(cfg *config.LoadedConfig, name string) error {
	templateDir := getTemplateDir(cfg)
	if err := notes.EnsureDir(templateDir); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	templatePath := filepath.Join(templateDir, name+".md")

	if utils.FileExists(templatePath) {
		return fmt.Errorf("template already exists: %s", name)
	}

	// Create basic template
	content := fmt.Sprintf("# %s Template\n\n## Section 1\n\n## Section 2\n\n", name)
	if err := os.WriteFile(templatePath, []byte(content), constants.FilePerm0644); err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	fmt.Printf("✓ Created template: %s\n", name)

	// Open in editor
	return notes.OpenInEditor(templatePath)
}

func editTemplate(cfg *config.LoadedConfig, name string) error {
	templateDir := getTemplateDir(cfg)
	templatePath := filepath.Join(templateDir, name+".md")

	if !utils.FileExists(templatePath) {
		return fmt.Errorf("template not found: %s", name)
	}

	return notes.OpenInEditor(templatePath)
}

func deleteTemplate(cfg *config.LoadedConfig, name string) error {
	templateDir := getTemplateDir(cfg)
	templatePath := filepath.Join(templateDir, name+".md")

	if !utils.FileExists(templatePath) {
		return fmt.Errorf("template not found: %s", name)
	}

	// Confirm deletion
	fmt.Printf("Delete template '%s'? (y/N): ", name)

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

	fmt.Printf("✓ Deleted template: %s\n", name)

	return nil
}
