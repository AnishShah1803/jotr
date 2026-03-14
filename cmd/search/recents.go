package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
)

var recentsCmdFlags = struct {
	limit int
	total bool
}{}

func init() {
	RecentsCmd.Flags().IntVarP(&recentsCmdFlags.limit, "limit", "n", 10, "Maximum number of recent notes to show")
	RecentsCmd.Flags().BoolVarP(&recentsCmdFlags.total, "total", "t", false, "Show count only instead of listing")
}

var RecentsCmd = &cobra.Command{
	Use:   "recents",
	Short: "List recently modified notes",
	Long: `List notes sorted by modification time (most recent first).

Flags:
  --limit N   Show top N recent notes (default: 10)
  --total     Show count only instead of listing

Examples:
  jotr recents                 # Show 10 most recently modified notes
  jotr recents --limit 5       # Show 5 most recent
  jotr recents --total         # Show count of modified notes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		return listRecents(cmd.Context(), cfg, recentsCmdFlags.limit, recentsCmdFlags.total)
	},
}

type fileInfo struct {
	path    string
	modTime time.Time
}

func listRecents(ctx context.Context, cfg *config.LoadedConfig, limit int, showTotal bool) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	if len(allNotes) == 0 {
		fmt.Println("No notes found")
		return nil
	}

	var fileInfos []fileInfo
	for _, note := range allNotes {
		fullPath := filepath.Join(cfg.Paths.BaseDir, note)
		stat, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{
			path:    note,
			modTime: stat.ModTime(),
		})
	}

	if len(fileInfos) == 0 {
		fmt.Println("No notes found")
		return nil
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime.After(fileInfos[j].modTime)
	})

	if showTotal {
		fmt.Printf("%d\n", len(fileInfos))
		return nil
	}

	if limit > 0 && limit < len(fileInfos) {
		fileInfos = fileInfos[:limit]
	}

	for i, fi := range fileInfos {
		relPath := fi.path
		modTimeStr := fi.modTime.Format("2006-01-02 15:04:05")
		fmt.Printf("%d. %s (%s)\n", i+1, relPath, modTimeStr)
	}

	return nil
}
