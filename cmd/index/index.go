package indexcmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/search"
)

var IndexCmd = &cobra.Command{
	Use:   "index [action]",
	Short: "Manage the search index",
	Long: `Manage the FTS5 search index for fast note searching.

Actions:
  rebuild   Rebuild the entire search index from scratch
  sync      Sync the index with current notes (incremental)
  status    Show index statistics

Examples:
  jotr index rebuild    # Full reindex of all notes
  jotr index sync       # Update index with changes only
  jotr index status     # Show index stats`,
	RunE: func(cmd *cobra.Command, args []string) error {
		action := "status"
		if len(args) > 0 {
			action = args[0]
		}

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		switch action {
		case "rebuild":
			return rebuildIndex(cmd.Context(), cfg)
		case "sync":
			return syncIndex(cmd.Context(), cfg)
		case "status":
			return indexStatus(cmd.Context(), cfg)
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
}

func rebuildIndex(ctx context.Context, cfg *config.LoadedConfig) error {
	indexPath := search.GetIndexPath(cfg.Paths.BaseDir)

	if err := os.Remove(indexPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old index: %w", err)
	}

	idx, err := search.Open(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	fmt.Println("Rebuilding search index...")
	start := time.Now()

	progress := func(indexed, total int) {
		if total > 0 {
			pct := (indexed * 100) / total
			fmt.Printf("\rIndexing... %d/%d (%d%%)", indexed, total, pct)
		}
	}

	result, err := idx.FullRebuild(ctx, cfg.Paths.BaseDir, &search.SyncOptions{
		ProgressCallback: progress,
	})

	fmt.Println()

	if err != nil {
		return fmt.Errorf("rebuild failed: %w", err)
	}

	duration := time.Since(start)
	fmt.Printf("\nIndex rebuilt successfully in %v\n", duration.Round(time.Second))
	fmt.Printf("  Indexed: %d notes\n", result.Indexed)
	if result.Errors > 0 {
		fmt.Printf("  Errors:  %d\n", result.Errors)
	}

	return nil
}

func syncIndex(ctx context.Context, cfg *config.LoadedConfig) error {
	indexPath := search.GetIndexPath(cfg.Paths.BaseDir)

	idx, err := search.Open(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	fmt.Println("Syncing search index...")
	start := time.Now()

	progress := func(indexed, total int) {
		if total > 0 {
			pct := (indexed * 100) / total
			fmt.Printf("\rSyncing... %d/%d (%d%%)", indexed, total, pct)
		}
	}

	result, err := idx.Sync(ctx, cfg.Paths.BaseDir, &search.SyncOptions{
		ProgressCallback: progress,
		DeleteMissing:    true,
	})

	fmt.Println()

	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	duration := time.Since(start)
	fmt.Printf("\nIndex synced successfully in %v\n", duration.Round(time.Millisecond*100))
	fmt.Printf("  Indexed: %d\n", result.Indexed)
	fmt.Printf("  Skipped: %d (unchanged)\n", result.Skipped)
	if result.Deleted > 0 {
		fmt.Printf("  Deleted: %d\n", result.Deleted)
	}
	if result.Errors > 0 {
		fmt.Printf("  Errors:  %d\n", result.Errors)
	}

	return nil
}

func indexStatus(ctx context.Context, cfg *config.LoadedConfig) error {
	indexPath := search.GetIndexPath(cfg.Paths.BaseDir)

	info, err := os.Stat(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Search index: Not created yet")
			fmt.Printf("\nRun 'jotr index rebuild' to create the index.\n")
			return nil
		}
		return fmt.Errorf("failed to stat index: %w", err)
	}

	idx, err := search.Open(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer idx.Close()

	stats, err := idx.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Println("Search Index Status")
	fmt.Println("==================")
	fmt.Printf("Location:    %s\n", indexPath)
	fmt.Printf("Size:        %s\n", formatBytes(info.Size()))
	fmt.Printf("Total notes: %d\n", stats["total_notes"])

	if lastIndexed, ok := stats["last_indexed"].(time.Time); ok {
		fmt.Printf("Last update: %s\n", lastIndexed.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("\nRun 'jotr index sync' to update with recent changes.\n")

	return nil
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
