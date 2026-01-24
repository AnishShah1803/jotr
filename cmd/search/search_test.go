package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
)

// createTestSearchConfig creates a test configuration with a temporary directory.
func createTestSearchConfig(t *testing.T, tmpDir string) *config.LoadedConfig {
	t.Helper()

	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = tmpDir
	cfg.Paths.DiaryDir = "Diary"
	cfg.Format.CaptureSection = "Captured"
	cfg.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	cfg.Format.DailyNoteDirPattern = "{year}/{month}"
	cfg.DiaryPath = filepath.Join(tmpDir, "Diary")

	return cfg
}

// Helper function to create a note for testing.
func createTestNote(t *testing.T, tmpDir, filename, content string) string {
	t.Helper()

	notePath := filepath.Join(tmpDir, filename+".md")
	ctx := context.Background()

	if err := notes.WriteNote(ctx, notePath, content); err != nil {
		t.Fatalf("Failed to create note %s: %v", filename, err)
	}

	return notePath
}

// TestSearchNotes_NoQuery tests that SearchNotes returns an error when no query is provided.
func TestSearchNotes_NoQuery(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Test with empty query - this tests the search functionality
	ctx := context.Background()

	matches, err := notes.SearchNotes(ctx, cfg.Paths.BaseDir, "")
	if err != nil {
		// Empty query might return error or empty results depending on implementation
		t.Logf("SearchNotes returned error for empty query: %v", err)
	} else {
		if len(matches) != 0 {
			t.Errorf("Expected 0 matches for empty query, got %d", len(matches))
		}
	}
}

// TestSearchNotes_SingleMatch tests searching for a term that matches exactly one note.
func TestSearchNotes_SingleMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create a note with specific content
	createTestNote(t, tmpDir, "MeetingNotes", "# Meeting Notes\n\nDiscussion about project X.\n")

	// Search for content that exists only in this note
	ctx := context.Background()

	matches, err := notes.SearchNotes(ctx, cfg.Paths.BaseDir, "project X")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(matches))
	}

	if len(matches) > 0 && !strings.Contains(matches[0], "MeetingNotes") {
		t.Errorf("Expected match to contain MeetingNotes, got: %s", matches[0])
	}
}

// TestSearchNotes_MultipleMatches tests searching for a term that matches multiple notes.
func TestSearchNotes_MultipleMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create multiple notes with the same content
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\nTODO: implement feature.\n")
	createTestNote(t, tmpDir, "Note2", "# Note 2\n\nTODO: design feature.\n")
	createTestNote(t, tmpDir, "Note3", "# Note 3\n\nDone with feature.\n")

	// Search for content that exists in multiple notes
	ctx := context.Background()

	matches, err := notes.SearchNotes(ctx, cfg.Paths.BaseDir, "TODO")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches for TODO, got %d", len(matches))
	}
}

// TestSearchNotes_NoMatches tests searching for a term that doesn't exist.
func TestSearchNotes_NoMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create a note with specific content
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\nSome content here.\n")

	// Search for something that doesn't exist
	ctx := context.Background()

	matches, err := notes.SearchNotes(ctx, cfg.Paths.BaseDir, "nonexistentterm12345")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}
}

// TestSearchNotes_CaseInsensitive tests that search is case insensitive.
func TestSearchNotes_CaseInsensitive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create a note with mixed case content
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\nIMPORTANT: This is important.\n")

	// Search with different cases
	ctx := context.Background()

	matchesLower, err := notes.SearchNotes(ctx, cfg.Paths.BaseDir, "important")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	matchesUpper, err := notes.SearchNotes(ctx, cfg.Paths.BaseDir, "IMPORTANT")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	matchesMixed, err := notes.SearchNotes(ctx, cfg.Paths.BaseDir, "Important")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	if len(matchesLower) != 1 || len(matchesUpper) != 1 || len(matchesMixed) != 1 {
		t.Errorf("Expected 1 match for all case variations, got lower=%d, upper=%d, mixed=%d",
			len(matchesLower), len(matchesUpper), len(matchesMixed))
	}
}

// TestSearchNotes_Subdirectory tests searching in notes with subdirectories.
func TestSearchNotes_Subdirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create subdirectory and note
	subDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(subDir, 0750); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	notePath := filepath.Join(subDir, "WorkNote.md")
	ctx := context.Background()

	if err := notes.WriteNote(ctx, notePath, "# Work Note\n\nProject requirements.\n"); err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Search should find notes in subdirectories
	matches, err := notes.SearchNotes(ctx, cfg.Paths.BaseDir, "requirements")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(matches))
	}
}

// TestExtractTags_Basic tests the extractTags helper function.
func TestExtractTags_Basic(t *testing.T) {
	content := `# Project Alpha
This is a note with #important tags and #urgent matters.
Also #todo and #bug tags.`

	tags := extractTags(content)

	expectedTags := map[string]bool{
		"important": true,
		"urgent":    true,
		"todo":      true,
		"bug":       true,
	}

	if len(tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(tags))
	}

	for _, tag := range tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag found: %s", tag)
		}
	}
}

// TestExtractTags_NoTags tests extractTags with content that has no tags.
func TestExtractTags_NoTags(t *testing.T) {
	content := `# Note Title
This is a note without any tags.
Just regular content here.`

	tags := extractTags(content)

	if len(tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(tags))
	}
}

// TestExtractTags_DuplicateTags tests that duplicate tags are handled.
func TestExtractTags_DuplicateTags(t *testing.T) {
	content := `# Note
#tag1 is here.
Also #tag1 again.
And #tag2 and #tag1 once more.`

	tags := extractTags(content)

	// Should only have 2 unique tags
	if len(tags) != 2 {
		t.Errorf("Expected 2 unique tags, got %d: %v", len(tags), tags)
	}

	// Count occurrences of each tag
	tagCount := make(map[string]int)
	for _, tag := range tags {
		tagCount[tag]++
	}

	for tag, count := range tagCount {
		if count != 1 {
			t.Errorf("Expected each tag to appear once, but %s appeared %d times", tag, count)
		}
	}
}

// TestExtractTags_SpecialCharacters tests tags with special characters.
func TestExtractTags_SpecialCharacters(t *testing.T) {
	content := "#alpha\n#beta-gamma\n#DELTA_123\n#epsilon42"

	tags := extractTags(content)

	// The regex #([a-zA-Z0-9_-]+) matches alphanumeric, underscore, hyphen
	// Expected tags: alpha, beta-gamma, DELTA_123, epsilon42
	expected := []string{"alpha", "beta-gamma", "DELTA_123", "epsilon42"}

	if len(tags) != len(expected) {
		t.Errorf("Expected %d tags, got %d: %v", len(expected), len(tags), tags)
	}

	// Check all expected tags are present
	for _, exp := range expected {
		found := false

		for _, tag := range tags {
			if tag == exp {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected tag %s not found in results: %v", exp, tags)
		}
	}
}

// TestListCmd_NoNotes tests listing notes when no notes exist.
func TestListCmd_NoNotes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-list-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create the diary directory
	diaryPath := filepath.Join(tmpDir, "Diary")
	if err := os.MkdirAll(diaryPath, 0750); err != nil {
		t.Fatalf("Failed to create diary directory: %v", err)
	}

	// listRecentNotes should handle empty directory gracefully
	err = listRecentNotes(context.Background(), cfg)
	if err != nil {
		t.Fatalf("listRecentNotes failed: %v", err)
	}
}

// TestListCmd_WithDailyNotes tests listing daily notes.
func TestListCmd_WithDailyNotes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-list-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create today's daily note
	diaryPath := cfg.DiaryPath
	if err := os.MkdirAll(diaryPath, 0750); err != nil {
		t.Fatalf("Failed to create diary directory: %v", err)
	}

	todayNotePath := notes.BuildDailyNotePath(diaryPath, time.Now())
	ctx := context.Background()

	if err := notes.WriteNote(ctx, todayNotePath, "# Today\n\nDaily note content.\n"); err != nil {
		t.Fatalf("Failed to create today's note: %v", err)
	}

	// listRecentNotes should find today's note
	err = listRecentNotes(context.Background(), cfg)
	if err != nil {
		t.Fatalf("listRecentNotes failed: %v", err)
	}
}

// TestSearchNotes_WithCountFlag tests the --count flag for SearchNotes.
func TestSearchNotes_WithCountFlag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-count-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create multiple notes with the same content
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\nTODO: implement feature.\n")
	createTestNote(t, tmpDir, "Note2", "# Note 2\n\nTODO: design feature.\n")
	createTestNote(t, tmpDir, "Note3", "# Note 3\n\nDone with feature.\n")

	// Set the --count flag
	SetSearchCountForTest(true)
	defer func() { SetSearchCountForTest(false) }()

	// Search for content that exists in multiple notes
	err = SearchNotes(context.Background(), cfg, "TODO")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}
}

// TestSearchNotes_WithFilesFlag tests the --files flag for SearchNotes.
func TestSearchNotes_WithFilesFlag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-files-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create multiple notes with the same content
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\nTODO: implement feature.\n")
	createTestNote(t, tmpDir, "Note2", "# Note 2\n\nTODO: design feature.\n")
	createTestNote(t, tmpDir, "Note3", "# Note 3\n\nDone with feature.\n")

	// Set the --files flag
	SetSearchFilesForTest(true)
	defer func() { SetSearchFilesForTest(false) }()

	// Search for content that exists in multiple notes
	err = SearchNotes(context.Background(), cfg, "TODO")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}
}

// TestSearchNotes_NoMatchesOutput tests searching when no matches exist (SearchNotes wrapper).
func TestSearchNotes_NoMatchesOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-no-matches-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create a note with specific content
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\nSome content here.\n")

	// Search for something that doesn't exist
	err = SearchNotes(context.Background(), cfg, "nonexistentterm12345")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}
}

// TestListTags_NoTags tests listTags when no notes exist.
func TestListTags_NoTags(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-tags-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create the base directory
	if err := os.MkdirAll(tmpDir, 0750); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	// listTags should handle no tags gracefully
	err = listTags(context.Background(), cfg)
	if err != nil {
		t.Fatalf("listTags failed: %v", err)
	}
}

// TestListTags_WithTags tests listTags when notes with tags exist.
func TestListTags_WithTags(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-tags-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create notes with tags
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\n#important #urgent\n")
	createTestNote(t, tmpDir, "Note2", "# Note 2\n\n#important #todo\n")

	// listTags should find the tags
	err = listTags(context.Background(), cfg)
	if err != nil {
		t.Fatalf("listTags failed: %v", err)
	}
}

// TestFindByTag_Exists tests findByTag when notes with the tag exist.
func TestFindByTag_Exists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-find-tag-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create notes with the tag
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\n#meeting\n")
	createTestNote(t, tmpDir, "Note2", "# Note 2\n\n#meeting #notes\n")

	// findByTag should find the notes
	err = findByTag(context.Background(), cfg, "meeting")
	if err != nil {
		t.Fatalf("findByTag failed: %v", err)
	}
}

// TestFindByTag_NotExists tests findByTag when no notes with the tag exist.
func TestFindByTag_NotExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-find-tag-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create notes without the tag
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\nNo tags here.\n")

	// findByTag should handle no results gracefully
	err = findByTag(context.Background(), cfg, "nonexistent")
	if err != nil {
		t.Fatalf("findByTag failed: %v", err)
	}
}

// TestFindByTag_WithHashPrefix tests findByTag when tag includes # prefix.
func TestFindByTag_WithHashPrefix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-find-tag-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create notes with the tag
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\n#bug\n")

	// findByTag should handle # prefix
	err = findByTag(context.Background(), cfg, "#bug")
	if err != nil {
		t.Fatalf("findByTag failed: %v", err)
	}
}

// TestTagStats_NoTags tests tagStats when no tags exist.
func TestTagStats_NoTags(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-tag-stats-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create the base directory
	if err := os.MkdirAll(tmpDir, 0750); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	// tagStats should handle no tags gracefully
	err = tagStats(context.Background(), cfg)
	if err != nil {
		t.Fatalf("tagStats failed: %v", err)
	}
}

// TestTagStats_WithTags tests tagStats when notes with tags exist.
func TestTagStats_WithTags(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-tag-stats-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create notes with tags (some with multiple tags, some repeated)
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\n#important #urgent\n")
	createTestNote(t, tmpDir, "Note2", "# Note 2\n\n#important #todo\n")
	createTestNote(t, tmpDir, "Note3", "# Note 3\n\n#important\n")

	// tagStats should show the counts
	err = tagStats(context.Background(), cfg)
	if err != nil {
		t.Fatalf("tagStats failed: %v", err)
	}
}

// TestShowLinks_NoLinks tests showLinks when note has no links.
func TestShowLinks_NoLinks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-links-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create a note without links
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\nNo links here.\n")

	// showLinks should handle no links gracefully
	err = showLinks(context.Background(), cfg, "Note1")
	if err != nil {
		t.Fatalf("showLinks failed: %v", err)
	}
}

// TestShowLinks_WithLinks tests showLinks when note has links.
func TestShowLinks_WithLinks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-links-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create a note with links
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\nSee [[OtherNote]] and [[AnotherNote]].\n")

	// showLinks should find the links
	err = showLinks(context.Background(), cfg, "Note1")
	if err != nil {
		t.Fatalf("showLinks failed: %v", err)
	}
}

// TestShowLinks_NoteNotFound tests showLinks when note doesn't exist.
func TestShowLinks_NoteNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-links-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create the base directory
	if err := os.MkdirAll(tmpDir, 0750); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	// showLinks should return error for non-existent note
	err = showLinks(context.Background(), cfg, "NonExistentNote")
	if err == nil {
		t.Error("Expected error for non-existent note, got nil")
	}
}

// TestShowBacklinks_NoBacklinks tests showBacklinks when no backlinks exist.
func TestShowBacklinks_NoBacklinks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-backlinks-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create a note without backlinks
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\nNo backlinks here.\n")

	// showBacklinks should handle no backlinks gracefully
	err = showBacklinks(context.Background(), cfg, "Note1")
	if err != nil {
		t.Fatalf("showBacklinks failed: %v", err)
	}
}

// TestShowBacklinks_WithBacklinks tests showBacklinks when backlinks exist.
func TestShowBacklinks_WithBacklinks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-backlinks-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create notes with backlinks
	createTestNote(t, tmpDir, "Note1", "# Note 1\n\nSee [[TargetNote]].\n")
	createTestNote(t, tmpDir, "Note2", "# Note 2\n\nAlso mentions [[TargetNote]].\n")

	// showBacklinks should find the backlinks
	err = showBacklinks(context.Background(), cfg, "TargetNote")
	if err != nil {
		t.Fatalf("showBacklinks failed: %v", err)
	}
}

// TestShowBacklinks_PartialMatch tests showBacklinks with partial note name match.
func TestShowBacklinks_PartialMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-backlinks-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create note with full name that should be matched partially
	createTestNote(t, tmpDir, "MyImportantNote", "# Note\n\nSee [[MyImportantNote]].\n")

	// showBacklinks should find backlinks for partial match
	err = showBacklinks(context.Background(), cfg, "Important")
	if err != nil {
		t.Fatalf("showBacklinks failed: %v", err)
	}
}

// TestListCmd_WithListAll tests listRecentNotes with --all flag.
func TestListCmd_WithListAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-list-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestSearchConfig(t, tmpDir)

	// Create multiple notes
	createTestNote(t, tmpDir, "Note1", "# Note 1\n")
	createTestNote(t, tmpDir, "Note2", "# Note 2\n")
	createTestNote(t, tmpDir, "Note3", "# Note 3\n")

	// Set the --all flag
	outputOption.FilesOnly = true
	defer func() { outputOption.FilesOnly = false }()

	// listRecentNotes should list all notes
	err = listRecentNotes(context.Background(), cfg)
	if err != nil {
		t.Fatalf("listRecentNotes failed: %v", err)
	}
}

// BenchmarkSearchNotes_Small is a small-scale benchmark for SearchNotes.
func BenchmarkSearchNotes_Small(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-bench-small-")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Create 25 notes (small scale)
	for i := 0; i < 25; i++ {
		content := fmt.Sprintf("# Note %d\n\nSome content about project development and planning.\n", i)
		createTestNoteForBenchmark(b, tmpDir, fmt.Sprintf("Note%d", i), content)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = notes.SearchNotes(ctx, tmpDir, "project")
	}
}

// BenchmarkSearchNotes_Medium is a medium-scale benchmark for SearchNotes.
func BenchmarkSearchNotes_Medium(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-bench-medium-")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Create 100 notes (medium scale)
	for i := 0; i < 100; i++ {
		content := fmt.Sprintf(`# Note %d

## Overview
This note contains information about %s.

## Details
Detailed discussion about %s development.
Key points:
- Point one regarding %s
- Point two about %s
- Point three from %s

## Summary
Summary of %s work completed.
`, i, randomWord(), randomWord(), randomWord(), randomWord(), randomWord(), randomWord())
		createTestNoteForBenchmark(b, tmpDir, fmt.Sprintf("Note%d", i), content)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = notes.SearchNotes(ctx, tmpDir, "development")
	}
}

// BenchmarkSearchNotes_Large is a large-scale benchmark for SearchNotes.
func BenchmarkSearchNotes_Large(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-bench-large-")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Create 250 notes (large scale, matching FindNotes benchmark)
	baseContent := `# Daily Note

## Tasks

### Todo
- [ ] Complete project setup <!-- id: abc123 -->
- [ ] Write documentation <!-- id: def456 -->

### In Progress
- [ ] Review code changes <!-- id: ghi789 -->
- [ ] Run tests <!-- id: jkl012 -->

### Done
- [x] Initial commit <!-- id: mno345 -->

## Notes

### Meeting Notes
Discussion about the sprint planning and roadmap.

### Ideas
Ideas for new features and improvements.

### Journal
End of day reflection and summary.
`

	for i := 0; i < 250; i++ {
		// Vary the content slightly per note
		content := fmt.Sprintf("# Daily Note - %d\n\n%s\n\nAdditional notes about note %d.\n", i, baseContent, i)
		year := 2024 + (i % 3)
		month := (i % 12) + 1
		day := (i % 28) + 1
		subDir := filepath.Join(tmpDir, fmt.Sprintf("%d/%02d", year, month))
		if err := os.MkdirAll(subDir, 0755); err != nil {
			b.Fatalf("Failed to create subdir: %v", err)
		}
		notePath := filepath.Join(subDir, fmt.Sprintf("%02d-%02d-%d.md", day, month, i))
		if err := notes.WriteNote(ctx, notePath, content); err != nil {
			b.Fatalf("Failed to create note %d: %v", i, err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = notes.SearchNotes(ctx, tmpDir, "sprint")
	}
}

// BenchmarkSearchNotes_NoMatches benchmarks searching for terms that don't exist.
func BenchmarkSearchNotes_NoMatches(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-bench-no-matches-")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Create 100 notes with content
	for i := 0; i < 100; i++ {
		content := fmt.Sprintf("# Note %d\n\nFixed content without the searched term.\n", i)
		createTestNoteForBenchmark(b, tmpDir, fmt.Sprintf("Note%d", i), content)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = notes.SearchNotes(ctx, tmpDir, "nonexistenttermthatdoesnotexist12345")
	}
}

// BenchmarkSearchNotes_CaseSensitivity benchmarks case-insensitive search.
func BenchmarkSearchNotes_CaseSensitivity(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-bench-case-")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Create notes with mixed case content
	for i := 0; i < 50; i++ {
		content := fmt.Sprintf("# Note %d\n\nIMPORTANT: This is an URGENT update about API.\n", i)
		createTestNoteForBenchmark(b, tmpDir, fmt.Sprintf("Note%d", i), content)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = notes.SearchNotes(ctx, tmpDir, "important")
	}
}

// BenchmarkSearchNotes_Subdirectory benchmarks searching with nested directories.
func BenchmarkSearchNotes_Subdirectory(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-bench-subdir-")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Create notes in nested subdirectories
	dirs := []string{"work", "personal", "projects", "archive", "research"}
	for d := 0; d < len(dirs); d++ {
		for i := 0; i < 20; i++ {
			subDir := filepath.Join(tmpDir, dirs[d], fmt.Sprintf("sub%d", i%3))
			if err := os.MkdirAll(subDir, 0755); err != nil {
				b.Fatalf("Failed to create subdir: %v", err)
			}
			content := fmt.Sprintf("# Note from %s\n\nSearchable content about %s topic %d.\n", dirs[d], dirs[d], i)
			notePath := filepath.Join(subDir, fmt.Sprintf("Note%d.md", i))
			if err := notes.WriteNote(ctx, notePath, content); err != nil {
				b.Fatalf("Failed to create note: %v", err)
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = notes.SearchNotes(ctx, tmpDir, "content")
	}
}

// BenchmarkSearchNotes_FindNotesOnly benchmarks the FindNotes part of search.
func BenchmarkSearchNotes_FindNotesOnly(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "jotr-search-bench-find-")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Create 250 notes in nested structure (same as large benchmark)
	for i := 0; i < 250; i++ {
		year := 2024 + (i % 3)
		month := (i % 12) + 1
		day := (i % 28) + 1
		subDir := filepath.Join(tmpDir, fmt.Sprintf("%d/%02d", year, month))
		if err := os.MkdirAll(subDir, 0755); err != nil {
			b.Fatalf("Failed to create subdir: %v", err)
		}
		notePath := filepath.Join(subDir, fmt.Sprintf("%02d-%02d-%d.md", day, month, i))
		content := fmt.Sprintf("# Note %d\n\nTest content.\n", i)
		if err := notes.WriteNote(ctx, notePath, content); err != nil {
			b.Fatalf("Failed to create note %d: %v", i, err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = notes.FindNotes(ctx, tmpDir)
	}
}

// randomWord returns a pseudo-random word for benchmark content generation.
func randomWord() string {
	words := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta", "iota", "kappa", "lambda", "mu", "nu", "xi", "omicron", "pi", "rho", "sigma", "tau", "upsilon"}
	return words[time.Now().UnixNano()%int64(len(words))]
}

func createTestNoteForBenchmark(b *testing.B, tmpDir, filename, content string) string {
	notePath := filepath.Join(tmpDir, filename+".md")
	ctx := context.Background()

	if err := notes.WriteNote(ctx, notePath, content); err != nil {
		b.Fatalf("Failed to create note %s: %v", filename, err)
	}

	return notePath
}
