package search

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIndex_Sync(t *testing.T) {
	dir, idx := setupTestDB(t)
	defer os.RemoveAll(dir)
	defer idx.Close()

	ctx := context.Background()
	notesDir := filepath.Join(dir, "notes")
	os.MkdirAll(notesDir, 0755)

	os.WriteFile(filepath.Join(notesDir, "note1.md"), []byte("# Note One\nContent one"), 0644)
	os.WriteFile(filepath.Join(notesDir, "note2.md"), []byte("# Note Two\nContent two"), 0644)

	res, err := idx.Sync(ctx, notesDir, &SyncOptions{DeleteMissing: true})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if res.Indexed != 2 {
		t.Errorf("Expected 2 indexed notes, got %d", res.Indexed)
	}

	res, err = idx.Sync(ctx, notesDir, &SyncOptions{DeleteMissing: true})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if res.Skipped != 2 {
		t.Errorf("Expected 2 skipped notes, got %d", res.Skipped)
	}

	os.WriteFile(filepath.Join(notesDir, "note1.md"), []byte("# Note One Updated\nContent one updated"), 0644)
	os.Chtimes(filepath.Join(notesDir, "note1.md"), time.Now(), time.Now().Add(2*time.Second))
	os.Remove(filepath.Join(notesDir, "note2.md"))
	os.WriteFile(filepath.Join(notesDir, "note3.md"), []byte("# Note Three\nContent three"), 0644)

	res, err = idx.Sync(ctx, notesDir, &SyncOptions{DeleteMissing: true})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if res.Indexed != 2 {
		t.Errorf("Expected 2 indexed notes, got %d", res.Indexed)
	}
	if res.Deleted != 1 {
		t.Errorf("Expected 1 deleted note, got %d", res.Deleted)
	}

	stats, _ := idx.GetStats(ctx)
	if total, ok := stats["total_notes"].(int); !ok || total != 2 {
		t.Errorf("Expected 2 total_notes after sync, got %v", stats["total_notes"])
	}
}

func TestIndex_ExtractTitle(t *testing.T) {
	tests := []struct {
		content string
		want    string
	}{
		{"# Hello World\nSome content", "Hello World"},
		{"\n\n  #   Spaced Title  \n", "Spaced Title"},
		{"No title here\nJust content", ""},
		{"## Subtitle\nShould not match", ""},
	}

	for _, tt := range tests {
		got := ExtractTitle(tt.content)
		if got != tt.want {
			t.Errorf("ExtractTitle(%q) = %q, want %q", tt.content, got, tt.want)
		}
	}
}
