package templatecmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/templates"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var TemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "Create note from template",
	RunE:  runTemplate,
}

func init() {
}

func runTemplate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := config.LoadWithContext(ctx, "")
	if err != nil {
		return err
	}

	templateList, warnings := templates.LoadTemplates(cfg)

	for _, warn := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warn)
	}

	if len(templateList) == 0 {
		return fmt.Errorf("no templates found in %s", cfg.TemplatesPath)
	}

	selected := selectTemplateInteractive(templateList)

	if selected == nil {
		return nil
	}

	return createFromSelectedTemplate(ctx, cfg, selected)
}

func selectTemplateInteractive(templateList []*templates.Template) *templates.Template {
	return SelectTemplateForIntegration(templateList)
}

func SelectTemplateForIntegration(templateList []*templates.Template) *templates.Template {
	categories := make(map[string][]*templates.Template)
	for _, tmpl := range templateList {
		categories[tmpl.Category] = append(categories[tmpl.Category], tmpl)
	}

	fmt.Println("\nAvailable Templates:")
	index := 1
	for cat, tmpls := range categories {
		fmt.Printf("\n%s:\n", cat)
		for _, tmpl := range tmpls {
			fmt.Printf("  %d. %s\n", index, tmpl.Name)
			index++
		}
	}

	choice := utils.PromptChoice("Select template", 1, index-1)

	var flat []*templates.Template
	for _, tmpls := range categories {
		flat = append(flat, tmpls...)
	}

	return flat[choice-1]
}

func createFromSelectedTemplate(ctx context.Context, cfg *config.LoadedConfig, tmpl *templates.Template) error {
	return CreateFromSelectedTemplate(ctx, cfg, tmpl)
}

func CreateFromSelectedTemplate(ctx context.Context, cfg *config.LoadedConfig, tmpl *templates.Template) error {
	varValues, err := templates.CollectVariableValues(tmpl.Variables)
	if err != nil {
		return err
	}

	promptValues, err := templates.CollectPromptValues(tmpl.Prompts)
	if err != nil {
		return err
	}

	content := templates.RenderTemplate(tmpl, varValues, promptValues, &cfg.Config)

	targetPath, err := templates.RenderTargetPath(tmpl, varValues, &cfg.Config)
	if err != nil {
		return err
	}

	return templates.CreateAndOpen(ctx, tmpl, content, targetPath)
}

func LoadTemplatesForIntegration(cfg *config.LoadedConfig) ([]*templates.Template, []string) {
	return templates.LoadTemplates(cfg)
}
