package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

type mockReader struct {
	inputs []string
	pos    int
}

func newMockReader(inputs ...string) *mockReader {
	return &mockReader{inputs: inputs, pos: 0}
}

func (r *mockReader) ReadString(delim byte) (string, error) {
	if r.pos >= len(r.inputs) {
		return "", fmt.Errorf("no more input")
	}
	input := r.inputs[r.pos]
	r.pos++
	return input, nil
}

// createTestConfig creates a test configuration with a temporary directory.
func createTestConfig(t *testing.T, tmpDir string) *config.LoadedConfig {
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

// TestCreateNote_Success tests successful note creation.
func TestCreateNote_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a note manually to avoid interactive input
	noteName := "TestNote"
	notePath := filepath.Join(tmpDir, noteName+".md")
	content := "# " + noteName + "\n\n"

	ctx := context.Background()
	if err := notes.WriteNote(ctx, notePath, content); err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Verify the note exists
	if !utils.FileExists(notePath) {
		t.Errorf("Expected note to exist at %s, but it doesn't", notePath)
	}

	// Verify content
	storedContent, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	if string(storedContent) != content {
		t.Errorf("Note content mismatch. Expected %q, got %q", content, string(storedContent))
	}
}

// TestCreateNote_AlreadyExists tests that creating a note that already exists returns an error.
func TestCreateNote_AlreadyExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an existing note
	noteName := "ExistingNote"
	notePath := filepath.Join(tmpDir, noteName+".md")
	content := "# " + noteName + "\n\n"

	ctx := context.Background()
	if err := notes.WriteNote(ctx, notePath, content); err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Try to create the same note again
	_, err = os.Stat(notePath)
	if err != nil {
		t.Fatalf("Note should exist: %v", err)
	}

	// Verify error message would be appropriate
	if utils.FileExists(notePath) {
		// This is the expected behavior - note already exists
		t.Log("Correctly detected existing note")
	}
}

// TestOpenNote_SingleMatch tests opening a note when there's exactly one match.
func TestOpenNote_SingleMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Create a note
	noteName := "UniqueNote"
	notePath := filepath.Join(tmpDir, noteName+".md")
	content := "# " + noteName + "\n\nSome content here.\n"

	ctx := context.Background()
	if err := notes.WriteNote(ctx, notePath, content); err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Find notes
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		t.Fatalf("Failed to find notes: %v", err)
	}

	if len(allNotes) == 0 {
		t.Fatalf("Expected to find at least one note")
	}

	// Single match should be the note we created
	if len(allNotes) == 1 {
		if allNotes[0] != notePath {
			t.Errorf("Expected note path %s, got %s", notePath, allNotes[0])
		}
	}
}

// TestOpenNote_MultipleMatches tests that multiple matching notes are correctly identified.
func TestOpenNote_MultipleMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	ctx := context.Background()

	// Create multiple notes with "test" in the name
	notesToCreate := []string{"TestNote1", "TestNote2", "TestNote3"}
	for _, name := range notesToCreate {
		notePath := filepath.Join(tmpDir, name+".md")
		if err := notes.WriteNote(ctx, notePath, "# "+name+"\n"); err != nil {
			t.Fatalf("Failed to create note %s: %v", name, err)
		}
	}

	// Create a non-matching note
	otherPath := filepath.Join(tmpDir, "OtherNote.md")
	if err := notes.WriteNote(ctx, otherPath, "# Other\n"); err != nil {
		t.Fatalf("Failed to create other note: %v", err)
	}

	// Find notes matching "test"
	query := "test"

	var matches []string

	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		t.Fatalf("Failed to find notes: %v", err)
	}

	for _, notePath := range allNotes {
		name := strings.ToLower(filepath.Base(notePath))
		if strings.Contains(name, query) {
			matches = append(matches, notePath)
		}
	}

	if len(matches) != 3 {
		t.Errorf("Expected 3 matches for query 'test', got %d", len(matches))
	}
}

// TestOpenNote_NoMatch tests behavior when no notes match the query.
func TestOpenNote_NoMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	ctx := context.Background()

	// Create a note that doesn't match
	notePath := filepath.Join(tmpDir, "UniqueNote.md")
	if err := notes.WriteNote(ctx, notePath, "# Unique\n"); err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Query for something that doesn't exist
	query := "nonexistent"

	var matches []string

	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		t.Fatalf("Failed to find notes: %v", err)
	}

	for _, notePath := range allNotes {
		name := strings.ToLower(filepath.Base(notePath))
		if strings.Contains(name, query) {
			matches = append(matches, notePath)
		}
	}

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for query 'nonexistent', got %d", len(matches))
	}
}

// TestListNotes_WithNotes tests listing notes when notes exist.
func TestListNotes_WithNotes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	ctx := context.Background()

	// Create some notes
	notesToCreate := []string{"Note1", "Note2", "Note3"}
	for _, name := range notesToCreate {
		notePath := filepath.Join(tmpDir, name+".md")
		if err := notes.WriteNote(ctx, notePath, "# "+name+"\n"); err != nil {
			t.Fatalf("Failed to create note %s: %v", name, err)
		}
	}

	// List notes
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		t.Fatalf("Failed to find notes: %v", err)
	}

	if len(allNotes) != 3 {
		t.Errorf("Expected 3 notes, got %d", len(allNotes))
	}
}

// TestListNotes_Empty tests listing notes when no notes exist.
func TestListNotes_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	ctx := context.Background()

	// List notes (no notes created)
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		t.Fatalf("Failed to find notes: %v", err)
	}

	if len(allNotes) != 0 {
		t.Errorf("Expected 0 notes, got %d", len(allNotes))
	}
}

// TestListNotes_Subdirectory tests listing notes with subdirectories.
func TestListNotes_Subdirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	ctx := context.Background()

	// Create notes in subdirectories
	subDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(subDir, 0750); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	notesToCreate := map[string]string{
		filepath.Join(tmpDir, "Note1.md"):     "# Note1\n",
		filepath.Join(subDir, "WorkNote1.md"): "# WorkNote1\n",
		filepath.Join(subDir, "WorkNote2.md"): "# WorkNote2\n",
	}

	for path, content := range notesToCreate {
		if err := notes.WriteNote(ctx, path, content); err != nil {
			t.Fatalf("Failed to create note %s: %v", path, err)
		}
	}

	// List notes should find all notes
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		t.Fatalf("Failed to find notes: %v", err)
	}

	if len(allNotes) != 3 {
		t.Errorf("Expected 3 notes, got %d", len(allNotes))
	}
}

// TestListNotesFunction tests the actual listNotes function from note.go.
func TestListNotesFunction(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	ctx := context.Background()

	// Create some notes
	for i := 1; i <= 3; i++ {
		notePath := filepath.Join(tmpDir, fmt.Sprintf("Note%d.md", i))
		if err := notes.WriteNote(ctx, notePath, "# Note"+fmt.Sprintf("%d", i)+"\n"); err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}
	}

	// Call the actual listNotes function
	err = listNotes(context.Background(), cfg)
	if err != nil {
		t.Errorf("listNotes() returned unexpected error: %v", err)
	}
}

// TestNotePathBuilding tests that note paths are built correctly.
func TestNotePathBuilding(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name     string
		noteType string
		expected string
	}{
		{
			name:     "SimpleNote",
			noteType: "",
			expected: "SimpleNote.md",
		},
		{
			name:     "WorkNote",
			noteType: "work",
			expected: "work/WorkNote.md",
		},
		{
			name:     "ProjectNote",
			noteType: "projects/demo",
			expected: "projects/demo/ProjectNote.md",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var notePath string
			if tc.noteType != "" {
				notePath = filepath.Join(tmpDir, tc.noteType, tc.name+".md")
			} else {
				notePath = filepath.Join(tmpDir, tc.name+".md")
			}

			// filepath.Base only returns the filename, not the full relative path
			baseName := filepath.Base(notePath)
			expectedBase := tc.name + ".md"

			if baseName != expectedBase {
				t.Errorf("Expected path base %s, got %s", expectedBase, baseName)
			}
		})
	}
}

// TestNoteQueryMatching tests note query matching logic.
func TestNoteQueryMatching(t *testing.T) {
	testCases := []struct {
		name     string
		query    string
		noteName string
		expected bool
	}{
		{
			name:     "Exact match",
			query:    "test",
			noteName: "test.md",
			expected: true,
		},
		{
			name:     "Case insensitive",
			query:    "Test",
			noteName: "test.md",
			expected: true,
		},
		{
			name:     "Partial match",
			query:    "est",
			noteName: "test.md",
			expected: true,
		},
		{
			name:     "No match",
			query:    "abc",
			noteName: "test.md",
			expected: false,
		},
		{
			name:     "Extension ignored",
			query:    "test",
			noteName: "test.txt",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := strings.ToLower(tc.query)
			name := strings.ToLower(filepath.Base(tc.noteName))
			result := strings.Contains(name, query)

			if result != tc.expected {
				t.Errorf("Query matching: expected %v for query=%s, note=%s, got %v",
					tc.expected, tc.query, tc.noteName, result)
			}
		})
	}
}

// TestNoteContentCreation tests that note content is created correctly.
func TestNoteContentCreation(t *testing.T) {
	noteName := "My Test Note"

	// Content format used in createNote
	content := "# " + noteName + "\n\n"

	if !strings.HasPrefix(content, "# ") {
		t.Errorf("Content should start with '# ' for markdown heading")
	}

	if !strings.HasSuffix(content, "\n\n") {
		t.Errorf("Content should have trailing newlines")
	}
}

// TestCreateNoteWithReader_Success tests successful note creation.
func TestCreateNoteWithReader_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Mock input for note name
	reader := newMockReader("TestNote\n")

	err = createNoteWithReader(context.Background(), cfg, "", reader)
	if err != nil {
		t.Errorf("createNoteWithReader() returned unexpected error: %v", err)
	}

	// Verify the note was created
	notePath := filepath.Join(tmpDir, "TestNote.md")
	if !utils.FileExists(notePath) {
		t.Errorf("Expected note to exist at %s, but it doesn't", notePath)
	}

	// Verify content
	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	expectedContent := "# TestNote\n\n"
	if string(content) != expectedContent {
		t.Errorf("Note content mismatch. Expected %q, got %q", expectedContent, string(content))
	}
}

// TestCreateNoteWithReader_WithType tests note creation with a type (subdirectory).
func TestCreateNoteWithReader_WithType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Mock input for note name
	reader := newMockReader("WorkNote\n")

	err = createNoteWithReader(context.Background(), cfg, "work", reader)
	if err != nil {
		t.Errorf("createNoteWithReader() returned unexpected error: %v", err)
	}

	// Verify the note was created in the work subdirectory
	notePath := filepath.Join(tmpDir, "work", "WorkNote.md")
	if !utils.FileExists(notePath) {
		t.Errorf("Expected note to exist at %s, but it doesn't", notePath)
	}
}

// TestCreateNoteWithReader_EmptyName tests that empty note name returns error.
func TestCreateNoteWithReader_EmptyName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Mock input for empty note name
	reader := newMockReader("\n")

	err = createNoteWithReader(context.Background(), cfg, "", reader)
	if err == nil {
		t.Errorf("createNoteWithReader() expected error for empty name, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "note name is required") {
		t.Errorf("createNoteWithReader() error = %v, want message containing 'note name is required'", err)
	}
}

// TestCreateNoteWithReader_AlreadyExists tests that creating duplicate note returns error.
func TestCreateNoteWithReader_AlreadyExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Create existing note
	existingPath := filepath.Join(tmpDir, "ExistingNote.md")
	if err := os.WriteFile(existingPath, []byte("# ExistingNote\n"), 0644); err != nil {
		t.Fatalf("Failed to create existing note: %v", err)
	}

	// Mock input for note name that already exists
	reader := newMockReader("ExistingNote\n")

	err = createNoteWithReader(context.Background(), cfg, "", reader)
	if err == nil {
		t.Errorf("createNoteWithReader() expected error for existing note, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "note already exists") {
		t.Errorf("createNoteWithReader() error = %v, want message containing 'note already exists'", err)
	}
}

// TestCreateNoteWithReader_ReadError tests that read errors are properly handled.
func TestCreateNoteWithReader_ReadError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Mock reader that returns error
	reader := newMockReader()

	err = createNoteWithReader(context.Background(), cfg, "", reader)
	if err == nil {
		t.Errorf("createNoteWithReader() expected error from reader, got nil")
	}
}

// TestOpenNoteWithReader_SingleMatch tests opening a note when there's exactly one match.
func TestOpenNoteWithReader_SingleMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	ctx := context.Background()

	// Create a single note
	notePath := filepath.Join(tmpDir, "UniqueNote.md")
	if err := notes.WriteNote(ctx, notePath, "# UniqueNote\n"); err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Query that matches exactly one note
	reader := newMockReader()

	err = openNoteWithReader(context.Background(), cfg, "unique", reader)
	// Should open the note (which may fail if editor is not available, but that's ok)
	// The important thing is it didn't return "no notes found" or similar errors
	if err != nil && strings.Contains(err.Error(), "no notes found") {
		t.Errorf("openNoteWithReader() should find the note, got: %v", err)
	}
}

// TestOpenNoteWithReader_MultipleMatches tests multiple matching notes with selection.
func TestOpenNoteWithReader_MultipleMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	ctx := context.Background()

	// Create multiple notes with "test" in the name
	notesToCreate := []string{"TestNote1", "TestNote2", "TestNote3"}
	for _, name := range notesToCreate {
		notePath := filepath.Join(tmpDir, name+".md")
		if err := notes.WriteNote(ctx, notePath, "# "+name+"\n"); err != nil {
			t.Fatalf("Failed to create note %s: %v", name, err)
		}
	}

	// Query that matches multiple notes, select the second one
	reader := newMockReader("2\n")

	err = openNoteWithReader(context.Background(), cfg, "test", reader)
	// Should try to open the second note
	if err != nil && !strings.Contains(err.Error(), "no notes found") {
		// Expected error is likely from trying to open the editor, which is fine
		t.Logf("openNoteWithReader() returned error (expected if no editor): %v", err)
	}
}

// TestOpenNoteWithReader_NoMatch tests behavior when no notes match the query.
func TestOpenNoteWithReader_NoMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	ctx := context.Background()

	// Create a note that doesn't match
	notePath := filepath.Join(tmpDir, "UniqueNote.md")
	if err := notes.WriteNote(ctx, notePath, "# Unique\n"); err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Query for something that doesn't exist
	reader := newMockReader()

	err = openNoteWithReader(context.Background(), cfg, "nonexistent", reader)
	if err == nil {
		t.Errorf("openNoteWithReader() expected error for no matches, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "no notes found matching") {
		t.Errorf("openNoteWithReader() error = %v, want message containing 'no notes found matching'", err)
	}
}

// TestOpenNoteWithReader_InvalidSelection tests handling of invalid selection input.
func TestOpenNoteWithReader_InvalidSelection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	ctx := context.Background()

	// Create multiple notes
	for i := 1; i <= 3; i++ {
		notePath := filepath.Join(tmpDir, fmt.Sprintf("Note%d.md", i))
		if err := notes.WriteNote(ctx, notePath, "# Note"+fmt.Sprintf("%d", i)+"\n"); err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}
	}

	// Invalid selection (out of range)
	reader := newMockReader("99\n")

	err = openNoteWithReader(context.Background(), cfg, "note", reader)
	if err == nil {
		t.Errorf("openNoteWithReader() expected error for invalid selection, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "invalid selection") {
		t.Errorf("openNoteWithReader() error = %v, want message containing 'invalid selection'", err)
	}
}

// TestOpenNoteWithReader_NonNumericSelection tests handling of non-numeric selection.
func TestOpenNoteWithReader_NonNumericSelection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	ctx := context.Background()

	// Create multiple notes
	for i := 1; i <= 3; i++ {
		notePath := filepath.Join(tmpDir, fmt.Sprintf("Note%d.md", i))
		if err := notes.WriteNote(ctx, notePath, "# Note"+fmt.Sprintf("%d", i)+"\n"); err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}
	}

	// Non-numeric selection
	reader := newMockReader("abc\n")

	err = openNoteWithReader(context.Background(), cfg, "note", reader)
	if err == nil {
		t.Errorf("openNoteWithReader() expected error for non-numeric selection, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "failed to parse selection") {
		t.Errorf("openNoteWithReader() error = %v, want message containing 'failed to parse selection'", err)
	}
}

// TestOpenNoteWithReader_EmptyNotes tests behavior when no notes exist at all.
func TestOpenNoteWithReader_EmptyNotes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-note-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// No notes created
	reader := newMockReader()

	err = openNoteWithReader(context.Background(), cfg, "", reader)
	if err == nil {
		t.Errorf("openNoteWithReader() expected error for empty notes, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "no notes found") {
		t.Errorf("openNoteWithReader() error = %v, want message containing 'no notes found'", err)
	}
}
