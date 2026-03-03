package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
)

var filesCmdFlags = struct {
	folder string
	ext    string
	total  bool
}{}

func init() {
	FilesCmd.Flags().StringVar(&filesCmdFlags.folder, "folder", "", "filter by folder path")
	FilesCmd.Flags().StringVar(&filesCmdFlags.ext, "ext", "", "filter by file extension (e.g. md, txt)")
	FilesCmd.Flags().BoolVar(&filesCmdFlags.total, "total", false, "show only the total count")
}

var FilesCmd = &cobra.Command{
	Use:   "files",
	Short: "List files in the vault",
	Long: `List all files in the vault with optional filtering.

Examples:
  jotr files                        # List all files
  jotr files --folder Projects      # List files in Projects folder
  jotr files --ext md               # List only markdown files
  jotr files --total                # Show only the count`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		allNotes, err := notes.FindNotes(cmd.Context(), cfg.Paths.BaseDir)
		if err != nil {
			return fmt.Errorf("failed to list files: %w", err)
		}

		var filtered []string
		for _, notePath := range allNotes {
			relPath, err := filepath.Rel(cfg.Paths.BaseDir, notePath)
			if err != nil {
				relPath = notePath
			}

			if filesCmdFlags.folder != "" {
				if !strings.HasPrefix(relPath, filesCmdFlags.folder+"/") &&
					!strings.HasPrefix(relPath, filesCmdFlags.folder+string(filepath.Separator)) {
					continue
				}
			}

			if filesCmdFlags.ext != "" {
				ext := strings.TrimPrefix(filesCmdFlags.ext, ".")
				if strings.TrimPrefix(filepath.Ext(notePath), ".") != ext {
					continue
				}
			}

			filtered = append(filtered, relPath)
		}

		if filesCmdFlags.total {
			fmt.Printf("%d\n", len(filtered))
			return nil
		}

		if len(filtered) == 0 {
			fmt.Println("No files found")
			return nil
		}

		for _, f := range filtered {
			fmt.Println(f)
		}

		fmt.Printf("\n%d file(s)\n", len(filtered))
		return nil
	},
}
