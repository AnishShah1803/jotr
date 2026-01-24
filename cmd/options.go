package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
)

func LoadConfig(ctx context.Context) (*config.LoadedConfig, error) {
	cfg, err := config.LoadWithContext(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func AddCommonExample(cmd *cobra.Command, examples ...string) {
	if len(examples) == 0 {
		return
	}
	exampleText := ""
	for _, example := range examples {
		exampleText += fmt.Sprintf("  %s\n", example)
	}
	if cmd.Long != "" {
		cmd.Long += "\n\nExamples:\n" + exampleText
	} else {
		cmd.Example = exampleText
	}
}

func AddStandardAliases(cmd *cobra.Command, aliases ...string) {
	cmd.Aliases = append(cmd.Aliases, aliases...)
}

func AddFlagDescriptions(cmd *cobra.Command, descriptions map[string]string) {
	for flag, desc := range descriptions {
		if flagCmd := cmd.Flags().Lookup(flag); flagCmd != nil {
			flagCmd.Usage = desc
		}
	}
}

func AddSeeAlso(cmd *cobra.Command, seeAlso ...string) {
	seeAlsoText := ""
	for _, cmdName := range seeAlso {
		seeAlsoText += fmt.Sprintf("  jotr %s\n", cmdName)
	}
	cmd.SetUsageTemplate(cmd.UsageTemplate() + "\nSee Also:\n" + seeAlsoText)
}

func AddDryRunFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
}

func AddVerboseFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("verbose", false, "Enable verbose output")
}

func SetCommandLong(cmd *cobra.Command, description, examples string) {
	cmd.Long = description
	if examples != "" {
		cmd.Long += "\n\nExamples:\n" + examples
	}
}
