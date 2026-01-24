package notes

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/testhelpers"
)

func TestCreateNote(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	notePath := "test-note.md"
	content := "# Test Note\n\nThis is a test note."
	ctx := context.Background()

	err := CreateNote(ctx, filepath.Join(fs.BaseDir, notePath), content)
	if err != nil {
		t.Fatalf("CreateNote() error = %v", err)
	}

	fs.AssertFileExists(t, notePath)
	fs.AssertFileContains(t, notePath, content)
}

func TestCreateNoteDirectory(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	ctx := context.Background()

	err := CreateNote(ctx, filepath.Join(fs.BaseDir, "subdir", "note.md"), "# Note")
	if err != nil {
		t.Fatalf("CreateNote() error = %v", err)
	}

	fs.AssertFileExists(t, filepath.Join("subdir", "note.md"))
}

func TestGetDailyNotePath(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	now := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	path, err := GetDailyNotePath(now)
	if err != nil {
		t.Fatalf("GetDailyNotePath() error = %v", err)
	}

	expected := filepath.Join(fs.BaseDir, "diary", "2026", "01-Jan", "2026-01-15-Thu.md")
	if path != expected {
		t.Errorf("GetDailyNotePath() = %q; want %q", path, expected)
	}
}

func TestGetNotesByPattern(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	fs.WriteFile(t, "note1.md", "# Note 1")
	fs.WriteFile(t, "note2.md", "# Note 2")
	fs.WriteFile(t, "other.txt", "Not a note")

	ctx := context.Background()

	notes, err := GetNotesByPattern(ctx, "*.md")
	if err != nil {
		t.Fatalf("GetNotesByPattern() error = %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("GetNotesByPattern() returned %d notes; want 2", len(notes))
	}
}

func TestGetNotesByTag(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	fs.WriteFile(t, "work.md", "# Work\nTags: #work #project")
	fs.WriteFile(t, "personal.md", "# Personal\nTags: #personal")
	fs.WriteFile(t, "idea.md", "# Idea\nTags: #work #idea")

	ctx := context.Background()

	workNotes, err := GetNotesByTag(ctx, "work")
	if err != nil {
		t.Fatalf("GetNotesByTag('work') error = %v", err)
	}

	if len(workNotes) != 2 {
		t.Errorf("GetNotesByTag('work') returned %d notes; want 2", len(workNotes))
	}
}

func TestUpdateLinks(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	configHelper := testhelpers.NewConfigHelper(fs)
	configHelper.CreateBasicConfig(t)

	configPath := filepath.Join(fs.BaseDir, ".config", "jotr", "config.json")
	os.Setenv("JOTR_CONFIG", configPath)

	fs.WriteFile(t, "note1.md", "# Note 1\nLink to [[note2]]")
	fs.WriteFile(t, "note2.md", "# Note 2")

	ctx := context.Background()

	err := UpdateLinks(ctx)
	if err != nil {
		t.Fatalf("UpdateLinks() error = %v", err)
	}

	content := fs.ReadFile(t, "note1.md")
	if !contains(content, "[[note2|Note 2]]") {
		t.Error("UpdateLinks() did not update link properly")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true

		for j := 0; j < len(substr); j++ {
			if i+j >= len(s) || s[i+j] != substr[j] {
				match = false
				break
			}
		}

		if match {
			return i
		}
	}

	return -1
}
