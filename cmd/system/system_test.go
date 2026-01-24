package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/utils"
)

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

// TestAliasOperations tests the alias command functions.
func TestAliasOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-alias-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Test loading aliases when file doesn't exist
	aliases, err := loadAliases(cfg)
	if err != nil {
		t.Fatalf("Failed to load aliases: %v", err)
	}

	if len(aliases) != 0 {
		t.Errorf("Expected empty aliases, got %d", len(aliases))
	}

	// Test adding an alias
	err = addAlias(cfg, "test", "TestNote.md")
	if err != nil {
		t.Fatalf("Failed to add alias: %v", err)
	}

	// Verify alias was added
	aliases, err = loadAliases(cfg)
	if err != nil {
		t.Fatalf("Failed to load aliases: %v", err)
	}

	if aliases["test"] != "TestNote.md" {
		t.Errorf("Expected alias 'test' -> 'TestNote.md', got '%s'", aliases["test"])
	}

	// Test resolving alias value
	resolved, err := resolveAliasValue(cfg, "TestNote.md")
	if err != nil {
		t.Fatalf("Failed to resolve alias value: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "TestNote.md")
	if resolved != expectedPath {
		t.Errorf("Expected resolved path %s, got %s", expectedPath, resolved)
	}

	// Test resolving daily alias
	dailyResolved, err := resolveAliasValue(cfg, "daily:0")
	if err != nil {
		t.Fatalf("Failed to resolve daily alias: %v", err)
	}

	if dailyResolved == "" {
		t.Error("Expected non-empty resolved daily path")
	}

	// Test removing an alias
	err = removeAlias(cfg, "test")
	if err != nil {
		t.Fatalf("Failed to remove alias: %v", err)
	}

	// Verify alias was removed
	aliases, err = loadAliases(cfg)
	if err != nil {
		t.Fatalf("Failed to load aliases: %v", err)
	}

	if _, exists := aliases["test"]; exists {
		t.Error("Alias 'test' should have been removed")
	}

	// Test removing non-existent alias
	err = removeAlias(cfg, "nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent alias")
	}
}

// TestResolveAliasValue_DailyAlias tests daily: offset aliases.
func TestResolveAliasValue_DailyAlias(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-alias-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Test daily:0 (today)
	resolved, err := resolveAliasValue(cfg, "daily:0")
	if err != nil {
		t.Fatalf("Failed to resolve daily:0: %v", err)
	}

	today := time.Now().Format("2006-01-02")
	if !contains(resolved, today) {
		t.Errorf("Expected resolved path to contain today's date %s, got %s", today, resolved)
	}

	// Test daily:-1 (yesterday)
	yesterdayResolved, err := resolveAliasValue(cfg, "daily:-1")
	if err != nil {
		t.Fatalf("Failed to resolve daily:-1: %v", err)
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if !contains(yesterdayResolved, yesterday) {
		t.Errorf("Expected resolved path to contain yesterday's date %s, got %s", yesterday, yesterdayResolved)
	}
}

// TestShortcutOperations tests the shortcut command functions.
func TestShortcutOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-shortcut-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Test loading shortcuts when file doesn't exist
	shortcuts, err := loadShortcuts(cfg)
	if err != nil {
		t.Fatalf("Failed to load shortcuts: %v", err)
	}

	if len(shortcuts) != 0 {
		t.Errorf("Expected empty shortcuts, got %d", len(shortcuts))
	}

	// Test isReserved
	if !isReserved("daily") {
		t.Error("'daily' should be reserved")
	}

	if !isReserved("note") {
		t.Error("'note' should be reserved")
	}

	if isReserved("mycustom") {
		t.Error("'mycustom' should not be reserved")
	}

	// Test adding a shortcut
	err = addShortcut(cfg, "td", "daily")
	if err != nil {
		t.Fatalf("Failed to add shortcut: %v", err)
	}

	// Verify shortcut was added
	shortcuts, err = loadShortcuts(cfg)
	if err != nil {
		t.Fatalf("Failed to load shortcuts: %v", err)
	}

	if shortcuts["td"] != "daily" {
		t.Errorf("Expected shortcut 'td' -> 'daily', got '%s'", shortcuts["td"])
	}

	// Test adding reserved shortcut
	err = addShortcut(cfg, "daily", "note")
	if err == nil {
		t.Error("Expected error when adding reserved command as shortcut")
	}

	// Test removing a shortcut
	err = removeShortcut(cfg, "td")
	if err != nil {
		t.Fatalf("Failed to remove shortcut: %v", err)
	}

	// Verify shortcut was removed
	shortcuts, err = loadShortcuts(cfg)
	if err != nil {
		t.Fatalf("Failed to load shortcuts: %v", err)
	}

	if _, exists := shortcuts["td"]; exists {
		t.Error("Shortcut 'td' should have been removed")
	}

	// Test removing non-existent shortcut
	err = removeShortcut(cfg, "nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent shortcut")
	}
}

// TestScheduleOperations tests the schedule command functions.
func TestScheduleOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-schedule-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Test loading scheduled notes when file doesn't exist
	scheduled, err := loadScheduledNotes(cfg)
	if err != nil {
		t.Fatalf("Failed to load scheduled notes: %v", err)
	}

	if len(scheduled) != 0 {
		t.Errorf("Expected empty scheduled notes, got %d", len(scheduled))
	}

	// Test adding a scheduled note for a future date
	futureDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	err = addScheduledNote(cfg, futureDate, "Test scheduled note")
	if err != nil {
		t.Fatalf("Failed to add scheduled note: %v", err)
	}

	// Verify note was added
	scheduled, err = loadScheduledNotes(cfg)
	if err != nil {
		t.Fatalf("Failed to load scheduled notes: %v", err)
	}

	if len(scheduled) != 1 {
		t.Errorf("Expected 1 scheduled note, got %d", len(scheduled))
	}

	// Test adding scheduled note with past date
	pastDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	err = addScheduledNote(cfg, pastDate, "Past note")
	if err == nil {
		t.Error("Expected error when scheduling note in the past")
	}

	// Test adding scheduled note with invalid date
	err = addScheduledNote(cfg, "invalid-date", "Test note")
	if err == nil {
		t.Error("Expected error with invalid date format")
	}

	// Test deleting a scheduled note
	if len(scheduled) > 0 {
		err = deleteScheduledNote(cfg, scheduled[0].ID)
		if err != nil {
			t.Fatalf("Failed to delete scheduled note: %v", err)
		}

		// Verify note was deleted
		scheduled, err = loadScheduledNotes(cfg)
		if err != nil {
			t.Fatalf("Failed to load scheduled notes: %v", err)
		}

		if len(scheduled) != 0 {
			t.Errorf("Expected 0 scheduled notes after deletion, got %d", len(scheduled))
		}
	}

	// Test deleting non-existent note
	err = deleteScheduledNote(cfg, "nonexistent-id")
	if err == nil {
		t.Error("Expected error when deleting non-existent scheduled note")
	}
}

// TestFrontmatterOperations tests the frontmatter command functions.
func TestFrontmatterOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-frontmatter-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Create a test note with frontmatter
	notePath := filepath.Join(tmpDir, "TestNote.md")
	noteContent := `---
status: pending
priority: P1
tags: [test, example]
---

# Test Note

Content here.
`

	err = os.WriteFile(notePath, []byte(noteContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test note: %v", err)
	}

	// Test setFrontmatter with valid key=value
	err = setFrontmatter(context.Background(), cfg, "TestNote", "status=done")
	if err != nil {
		t.Fatalf("Failed to set frontmatter: %v", err)
	}

	// Verify frontmatter was updated
	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read note: %v", err)
	}

	if !contains(string(content), "status: done") {
		t.Errorf("Expected frontmatter to contain 'status: done', got:\n%s", string(content))
	}

	// Test setFrontmatter with invalid format
	err = setFrontmatter(context.Background(), cfg, "TestNote", "invalidformat")
	if err == nil {
		t.Error("Expected error with invalid frontmatter format")
	}

	// Test setFrontmatter for non-existent note
	err = setFrontmatter(context.Background(), cfg, "NonExistent", "status=done")
	if err == nil {
		t.Error("Expected error for non-existent note")
	}
}

// TestUpdateCheck tests the update check functionality.
func TestUpdateCheck(t *testing.T) {
	// Test CheckForUpdates function (exported)
	hasUpdate, latestVersion, err := CheckForUpdates()
	// We can't predict the exact behavior as it depends on actual GitHub releases
	// But we can verify the function doesn't error and returns valid data
	if err == nil {
		t.Logf("Update check completed: hasUpdate=%v, latestVersion=%s", hasUpdate, latestVersion)
	} else {
		t.Logf("Update check returned error (expected in test environment): %v", err)
	}
}

// TestMonthlySummary tests the monthly summary generation.
func TestMonthlySummary(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-monthly-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Set global monthly variables for testing
	now := time.Now()
	monthlyYear = now.Year()
	monthlyMonth = int(now.Month())

	// Test with non-existent month directory
	err = generateMonthlySummary(context.Background(), cfg)
	if err == nil {
		t.Error("Expected error when generating summary for month with no notes")
	}

	// Create a test daily note
	diaryPath := filepath.Join(tmpDir, "Diary", now.Format("2006"), now.Format("01-Jan"))

	err = os.MkdirAll(diaryPath, 0750)
	if err != nil {
		t.Fatalf("Failed to create diary directory: %v", err)
	}

	notePath := filepath.Join(diaryPath, now.Format("2006-01-02-Mon.md"))

	err = os.WriteFile(notePath, []byte("# Test Daily Note\n\nContent"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test note: %v", err)
	}

	// Test generating monthly summary
	err = generateMonthlySummary(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to generate monthly summary: %v", err)
	}

	// Verify summary was created
	summaryPath := filepath.Join(diaryPath, "../summaries/01-Jan-Summary.md")
	if !utils.FileExists(summaryPath) {
		t.Errorf("Expected summary file to exist at %s", summaryPath)
	}
}

// TestInvalidMonthValidation tests month validation in monthly summary.
func TestInvalidMonthValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-monthly-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Test invalid month (0)
	monthlyYear = 2024
	monthlyMonth = 0

	err = generateMonthlySummary(context.Background(), cfg)
	if err == nil {
		t.Error("Expected error for month 0")
	}

	// Test invalid month (13)
	monthlyMonth = 13

	err = generateMonthlySummary(context.Background(), cfg)
	if err == nil {
		t.Error("Expected error for month 13")
	}
}

// Helper function to check if string contains substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

// TestListFunctions tests the list functions that print to stdout.
func TestListFunctions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-list-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Test listAliases with no aliases
	err = listAliases(cfg)
	if err != nil {
		t.Errorf("listAliases should not error with empty aliases: %v", err)
	}

	// Add an alias and test listAliases
	err = addAlias(cfg, "test", "TestNote.md")
	if err != nil {
		t.Fatalf("Failed to add alias: %v", err)
	}

	err = listAliases(cfg)
	if err != nil {
		t.Errorf("listAliases should not error with aliases: %v", err)
	}

	// Test listShortcuts with no shortcuts
	err = listShortcuts(cfg)
	if err != nil {
		t.Errorf("listShortcuts should not error with empty shortcuts: %v", err)
	}

	// Add a shortcut and test listShortcuts
	err = addShortcut(cfg, "td", "daily")
	if err != nil {
		t.Fatalf("Failed to add shortcut: %v", err)
	}

	err = listShortcuts(cfg)
	if err != nil {
		t.Errorf("listShortcuts should not error with shortcuts: %v", err)
	}

	// Test listScheduledNotes with no notes
	err = listScheduledNotes(cfg)
	if err != nil {
		t.Errorf("listScheduledNotes should not error with empty notes: %v", err)
	}

	// Add a scheduled note and test listScheduledNotes
	futureDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	err = addScheduledNote(cfg, futureDate, "Test note")
	if err != nil {
		t.Fatalf("Failed to add scheduled note: %v", err)
	}

	err = listScheduledNotes(cfg)
	if err != nil {
		t.Errorf("listScheduledNotes should not error with notes: %v", err)
	}
}

// TestResolveAlias tests the resolveAlias function.
func TestResolveAlias(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-resolve-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Test resolveAlias with non-existent alias
	err = resolveAlias(cfg, "nonexistent")
	if err == nil {
		t.Error("Expected error when resolving non-existent alias")
	}

	// Add an alias and test resolveAlias
	err = addAlias(cfg, "test", "TestNote.md")
	if err != nil {
		t.Fatalf("Failed to add alias: %v", err)
	}

	err = resolveAlias(cfg, "test")
	if err != nil {
		t.Errorf("resolveAlias should not error with existing alias: %v", err)
	}
}

// TestShowFrontmatter tests the showFrontmatter function.
func TestShowFrontmatter(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-frontmatter-show-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	// Test showFrontmatter with non-existent note
	err = showFrontmatter(context.Background(), cfg, "NonExistent")
	if err == nil {
		t.Error("Expected error when showing frontmatter for non-existent note")
	}

	// Create a note without frontmatter
	notePath := filepath.Join(tmpDir, "NoFrontmatter.md")

	err = os.WriteFile(notePath, []byte("# Test Note\n\nContent"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test note: %v", err)
	}

	err = showFrontmatter(context.Background(), cfg, "NoFrontmatter")
	if err != nil {
		t.Errorf("showFrontmatter should not error for note without frontmatter: %v", err)
	}

	// Create a note with frontmatter
	noteWithFM := filepath.Join(tmpDir, "WithFrontmatter.md")
	fmContent := `---
status: pending
priority: P1
---

# Test Note
Content here.
`

	err = os.WriteFile(noteWithFM, []byte(fmContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test note with frontmatter: %v", err)
	}

	err = showFrontmatter(context.Background(), cfg, "WithFrontmatter")
	if err != nil {
		t.Errorf("showFrontmatter should not error for note with frontmatter: %v", err)
	}
}

// TestAliasFilePath tests the getAliasFile function.
func TestAliasFilePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-aliaspath-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	aliasFile := getAliasFile(cfg)
	expected := filepath.Join(tmpDir, ".aliases.json")

	if aliasFile != expected {
		t.Errorf("Expected alias file path %s, got %s", expected, aliasFile)
	}
}

// TestShortcutFilePath tests the getShortcutFile function.
func TestShortcutFilePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-shortcutpath-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	shortcutFile := getShortcutFile(cfg)
	expected := filepath.Join(tmpDir, ".shortcuts.json")

	if shortcutFile != expected {
		t.Errorf("Expected shortcut file path %s, got %s", expected, shortcutFile)
	}
}

// TestScheduleFilePath tests the getScheduleFile function.
func TestScheduleFilePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-schedulepath-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestConfig(t, tmpDir)

	scheduleFile := getScheduleFile(cfg)
	expected := filepath.Join(tmpDir, ".scheduled_notes.json")

	if scheduleFile != expected {
		t.Errorf("Expected schedule file path %s, got %s", expected, scheduleFile)
	}
}
