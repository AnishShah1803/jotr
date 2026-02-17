package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AnishShah1803/jotr/internal/constants"
)

// TestCleanup_TempFileCleanup verifies that all temporary files are cleaned up.
func TestCleanup_TempFileCleanup(t *testing.T) {
	// Get the count of existing jotr test directories before test
	initialCount := countJotrTestDirs()

	// Run multiple operations that create temp files
	tmpDir1, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir 1: %v", err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir 2: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	// Create some test files in temp directories
	testFile1 := filepath.Join(tmpDir1, "test1.txt")
	testFile2 := filepath.Join(tmpDir2, "test2.txt")

	err = AtomicWriteFile(testFile1, []byte("test data 1"), constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Failed to write test file 1: %v", err)
	}

	err = AtomicWriteFile(testFile2, []byte("test data 2"), constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Failed to write test file 2: %v", err)
	}

	// Check that no additional temp directories were created by our operations
	// (AtomicWriteFile should only create temp files within the target directory)
	midTestCount := countJotrTestDirs()

	if midTestCount != initialCount+2 {
		t.Errorf("Expected temp dir count to increase by 2, got initial:%d, mid:%d",
			initialCount, midTestCount)
	}

	// Verify no leftover .tmp files in our temp directories
	tmpFiles1, err := filepath.Glob(filepath.Join(tmpDir1, "*.tmp*"))
	if err != nil {
		t.Fatalf("Failed to check for temp files: %v", err)
	}

	if len(tmpFiles1) > 0 {
		t.Errorf("Found leftover temp files in tmpDir1: %v", tmpFiles1)
	}

	tmpFiles2, err := filepath.Glob(filepath.Join(tmpDir2, "*.tmp*"))
	if err != nil {
		t.Fatalf("Failed to check for temp files: %v", err)
	}

	if len(tmpFiles2) > 0 {
		t.Errorf("Found leftover temp files in tmpDir2: %v", tmpFiles2)
	}
	// Test cleanup happens automatically when defer runs
	// We'll check this in the final count after defers execute
}

// TestCleanup_BackupFileManagement tests that backup files are managed properly.
func TestCleanup_BackupFileManagement(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create initial file
	initialData := []byte("initial content")

	err = os.WriteFile(testFile, initialData, constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Create backup
	backupPath, err := BackupFile(testFile)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Verify backup exists
	if backupPath == "" {
		t.Fatalf("Backup path is empty")
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup file was not created at %s", backupPath)
	}

	// Check that backup is in the same directory as original
	backupDir := filepath.Dir(backupPath)
	originalDir := filepath.Dir(testFile)

	if backupDir != originalDir {
		t.Errorf("Backup created in wrong directory. Expected %s, got %s", originalDir, backupDir)
	}

	// When tmpDir is removed by defer, both original and backup should be cleaned up
	// This happens automatically, but let's verify the backup follows expected naming
	expectedBackupName := testFile + ".backup"
	if backupPath != expectedBackupName {
		t.Errorf("Backup has unexpected name. Expected %s, got %s", expectedBackupName, backupPath)
	}
}

// countJotrTestDirs counts temporary directories created by our tests.
func countJotrTestDirs() int {
	entries, err := os.ReadDir("/tmp")
	if err != nil {
		return -1 // Can't count, but that's ok for this helper
	}

	count := 0

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "jotr-test-") {
			count++
		}
	}

	return count
}
