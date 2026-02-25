package search

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) (string, *Index) {
	t.Helper()
	dir, err := os.MkdirTemp("", "jotr-search-test-*")
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(dir, "search.db")
	idx, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	return dir, idx
}

func TestIndex_Lifecycle(t *testing.T) {
	dir, idx := setupTestDB(t)
	defer os.RemoveAll(dir)
	defer idx.Close()

	if idx.db == nil {
		t.Fatal("idx.db should not be nil")
	}
}

func TestIndex_CRUD(t *testing.T) {
	dir, idx := setupTestDB(t)
	defer os.RemoveAll(dir)
	defer idx.Close()

	ctx := context.Background()
	now := time.Now()

	err := idx.IndexNote(ctx, "/path/note1.md", "Note 1", "This is the first note with some content", now)
	if err != nil {
		t.Fatalf("Failed to index note: %v", err)
	}

	err = idx.IndexNote(ctx, "/path/note2.md", "Note 2", "Second note mentions apple", now)
	if err != nil {
		t.Fatalf("Failed to index note: %v", err)
	}

	results, err := idx.Search(ctx, "content", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Path != "/path/note1.md" {
		t.Errorf("Expected path /path/note1.md, got %s", results[0].Path)
	}

	err = idx.IndexNote(ctx, "/path/note1.md", "Note 1", "Updated content", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("Failed to update note: %v", err)
	}

	results, err = idx.Search(ctx, "first", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'first', got %d", len(results))
	}

	results, err = idx.Search(ctx, "Updated", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'Updated', got %d", len(results))
	}

	err = idx.DeleteNote(ctx, "/path/note1.md")
	if err != nil {
		t.Fatalf("Failed to delete note: %v", err)
	}

	results, err = idx.Search(ctx, "Updated", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'Updated' after delete, got %d", len(results))
	}
}

func TestIndex_EscapeQuery(t *testing.T) {
	escaped := escapeFTS5Query("something(with) *and^")
	if escaped != `"somethingwith and"` {
		t.Errorf("Unexpected escape result: %s", escaped)
	}
}

func TestIndex_GetStats(t *testing.T) {
	dir, idx := setupTestDB(t)
	defer os.RemoveAll(dir)
	defer idx.Close()

	ctx := context.Background()
	now := time.Now()

	idx.IndexNote(ctx, "/path/1.md", "1", "one", now)
	idx.IndexNote(ctx, "/path/2.md", "2", "two", now)

	stats, err := idx.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if total, ok := stats["total_notes"].(int); !ok || total != 2 {
		t.Errorf("Expected 2 total_notes, got %v", stats["total_notes"])
	}
}
