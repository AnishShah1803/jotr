package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/notes"
)

var BulkCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Bulk operations on notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var bulkRenameCmd = &cobra.Command{
	Use:   "rename [old] [new]",
	Short: "Rename text across all notes",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return bulkRename(cmd.Context(), cfg, args[0], args[1])
	},
}

var bulkTagCmd = &cobra.Command{
	Use:   "tag [tag]",
	Short: "Add tag to all notes matching query",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return bulkTag(cmd.Context(), cfg, args[0])
	},
}

func init() {
	BulkCmd.AddCommand(bulkRenameCmd, bulkTagCmd)
}

func bulkRename(ctx context.Context, cfg *config.LoadedConfig, oldText, newText string) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return fmt.Errorf("failed to find notes: %w", err)
	}

	modifiedCount := 0

	for _, notePath := range allNotes {
		content, err := os.ReadFile(notePath)
		if err != nil {
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, oldText) {
			newContent := strings.ReplaceAll(contentStr, oldText, newText)
			if err := os.WriteFile(notePath, []byte(newContent), constants.FilePerm0644); err != nil {
				fmt.Printf("⚠️  Failed to update: %s\n", notePath)
				continue
			}

			relPath, _ := filepath.Rel(cfg.Paths.BaseDir, notePath)
			fmt.Printf("✓ Updated: %s\n", relPath)

			modifiedCount++
		}
	}

	fmt.Printf("\n✓ Modified %d notes\n", modifiedCount)

	return nil
}

func bulkTag(ctx context.Context, cfg *config.LoadedConfig, tag string) error {
	fmt.Printf("Bulk tagging with #%s\n", tag)
	fmt.Println("(This is a placeholder - implement with query filter)")

	return nil
}
