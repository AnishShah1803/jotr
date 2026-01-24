package templates

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

func CreateFromTemplate(ctx context.Context, tmpl *Template, content, targetPath string) error {
	dir := filepath.Dir(targetPath)
	if err := utils.EnsureDir(dir); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if utils.FileExists(targetPath) {
		return fmt.Errorf("file already exists: %s", targetPath)
	}

	if err := utils.AtomicWriteFileCtx(ctx, targetPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

func CreateAndOpen(ctx context.Context, tmpl *Template, content, targetPath string) error {
	if err := CreateFromTemplate(ctx, tmpl, content, targetPath); err != nil {
		return err
	}

	return notes.OpenInEditorWithContext(ctx, targetPath)
}
