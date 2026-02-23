package search

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SyncOptions struct {
	ProgressCallback func(indexed, total int)
	DeleteMissing    bool
}

type SyncResult struct {
	Indexed  int
	Deleted  int
	Errors   int
	Skipped  int
	Duration time.Duration
}

func (idx *Index) Sync(ctx context.Context, notesDir string, opts *SyncOptions) (*SyncResult, error) {
	if opts == nil {
		opts = &SyncOptions{}
	}

	start := time.Now()
	result := &SyncResult{}

	indexedPaths, err := idx.GetIndexedPaths(ctx)
	if err != nil {
		return nil, err
	}

	var filesToIndex []string
	var allFiles []string

	err = filepath.WalkDir(notesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.IsDir() && d.Name() == ".jotr" {
			return filepath.SkipDir
		}

		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			allFiles = append(allFiles, path)

			info, err := d.Info()
			if err != nil {
				return nil
			}

			modTime := info.ModTime().Unix()
			if lastMod, exists := indexedPaths[path]; !exists || lastMod != modTime {
				filesToIndex = append(filesToIndex, path)
			} else {
				result.Skipped++
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if opts.DeleteMissing {
		currentFiles := make(map[string]bool)
		for _, f := range allFiles {
			currentFiles[f] = true
		}

		for path := range indexedPaths {
			if !currentFiles[path] {
				if err := idx.DeleteNote(ctx, path); err != nil {
					result.Errors++
				} else {
					result.Deleted++
				}
			}
		}
	}

	total := len(filesToIndex)
	for i, path := range filesToIndex {
		select {
		case <-ctx.Done():
			result.Duration = time.Since(start)
			return result, ctx.Err()
		default:
		}

		if err := idx.indexFile(ctx, path); err != nil {
			result.Errors++
			continue
		}

		result.Indexed++

		if opts.ProgressCallback != nil {
			opts.ProgressCallback(i+1, total)
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

func (idx *Index) indexFile(ctx context.Context, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	title := ExtractTitle(string(content))

	return idx.IndexNote(ctx, path, title, string(content), info.ModTime())
}

func ExtractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func (idx *Index) FullRebuild(ctx context.Context, notesDir string, opts *SyncOptions) (*SyncResult, error) {
	if opts == nil {
		opts = &SyncOptions{}
	}

	start := time.Now()
	result := &SyncResult{}

	if _, err := idx.db.ExecContext(ctx, "DELETE FROM notes_fts"); err != nil {
		return nil, err
	}
	if _, err := idx.db.ExecContext(ctx, "DELETE FROM notes"); err != nil {
		return nil, err
	}

	var filesToIndex []string

	err := filepath.WalkDir(notesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.IsDir() && d.Name() == ".jotr" {
			return filepath.SkipDir
		}

		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			filesToIndex = append(filesToIndex, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	total := len(filesToIndex)
	for i, path := range filesToIndex {
		select {
		case <-ctx.Done():
			result.Duration = time.Since(start)
			return result, ctx.Err()
		default:
		}

		if err := idx.indexFile(ctx, path); err != nil {
			result.Errors++
			continue
		}

		result.Indexed++

		if opts.ProgressCallback != nil {
			opts.ProgressCallback(i+1, total)
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}
