package templatecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/templates"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all templates",
	RunE:  runList,
}

func init() {
	TemplateCmd.AddCommand(ListCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := config.LoadWithContext(ctx, "")
	if err != nil {
		return err
	}

	templateList, warnings := templates.LoadTemplates(cfg)

	for _, warn := range warnings {
		fmt.Printf("Warning: %s\n", warn)
	}

	if len(templateList) == 0 {
		fmt.Printf("No templates found in %s\n", cfg.TemplatesPath)
		return nil
	}

	fmt.Println("Available templates:")
	fmt.Println()

	categories := make(map[string][]*templates.Template)
	for _, tmpl := range templateList {
		categories[tmpl.Category] = append(categories[tmpl.Category], tmpl)
	}

	for cat, tmpls := range categories {
		fmt.Printf("%s:\n", cat)
		for _, tmpl := range tmpls {
			fmt.Printf("  %s\n", tmpl.Name)
		}
		fmt.Println()
	}

	return nil
}
