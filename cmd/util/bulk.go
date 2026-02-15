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
	Use:   "bulk [action]",
	Short: "Bulk operations on notes",
	Long: `Perform bulk operations on multiple notes.
	
Actions:
  rename [old] [new]    Rename text across all notes
  tag [tag]             Add tag to all notes matching query
  
Examples:
  jotr bulk rename "old text" "new text"
  jotr bulk tag meeting --query "team sync"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("action required: rename or tag")
		}

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		action := args[0]

		switch action {
		case "rename":
			if len(args) < 3 {
				return fmt.Errorf("usage: bulk rename [old] [new]")
			}
			return bulkRename(cmd.Context(), cfg, args[1], args[2])
		case "tag":
			if len(args) < 2 {
				return fmt.Errorf("usage: bulk tag [tag]")
			}
			return bulkTag(cmd.Context(), cfg, args[1])
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
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
	// For now, just show what would be tagged
	// In a real implementation, you'd prompt for confirmation
	fmt.Printf("Bulk tagging with #%s\n", tag)
	fmt.Println("(This is a placeholder - implement with query filter)")

	return nil
}
