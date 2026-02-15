package templates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AnishShah1803/jotr/internal/constants"
)

func CreateTestTemplateDir(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	os.MkdirAll(templatesDir, 0755)
	return templatesDir, func() { os.RemoveAll(tmpDir) }
}

func CreateTestTemplate(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	os.WriteFile(path, []byte(content), constants.FilePerm0644)
}
