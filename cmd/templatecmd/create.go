package templatecmd

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

var CreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new template",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func init() {
	TemplateCmd.AddCommand(CreateCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := config.LoadWithContext(ctx, "")
	if err != nil {
		return err
	}

	name := args[0]

	if err := notes.EnsureDir(cfg.TemplatesPath); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	templatePath := filepath.Join(cfg.TemplatesPath, name+".md")

	if utils.FileExists(templatePath) {
		return fmt.Errorf("template already exists: %s", name)
	}

	content := fmt.Sprintf(`---
name: %s
category: General
priority: 10
---

# %s

## Section 1

## Section 2
`, name, name)

	if err := os.WriteFile(templatePath, []byte(content), constants.FilePerm0644); err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	fmt.Printf("Created template: %s\n", name)

	return notes.OpenInEditorWithContext(ctx, templatePath)
}
