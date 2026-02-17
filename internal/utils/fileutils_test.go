package utils

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/constants"
)

func TestAtomicWriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	testData := []byte("Hello, atomic world!")

	// Test basic atomic write
	err = AtomicWriteFile(testFile, testData, constants.FilePerm0644)
	if err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	// Verify file was created with correct content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("Content mismatch. Expected %q, got %q", testData, content)
	}

	// Verify file permissions
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Mode().Perm() != constants.FilePerm0644 {
		t.Errorf("Permission mismatch. Expected 0644, got %o", info.Mode().Perm())
	}
}

func TestAtomicWriteFile_Overwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create initial file
	initialData := []byte("Initial content")

	err = AtomicWriteFile(testFile, initialData, constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Initial write failed: %v", err)
	}

	// Overwrite with new content
	newData := []byte("Updated content")

	err = AtomicWriteFile(testFile, newData, constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	// Verify new content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(newData) {
		t.Errorf("Content mismatch. Expected %q, got %q", newData, content)
	}
}

func TestAtomicWriteFile_ReadonlySource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "readonly.txt")

	// Create initial file and make it readonly
	initialData := []byte("Readonly content")

	err = AtomicWriteFile(testFile, initialData, constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Initial write failed: %v", err)
	}

	err = os.Chmod(testFile, 0444)
	if err != nil {
		t.Fatalf("Failed to make file readonly: %v", err)
	}

	// Atomic write should still work (creates new file, renames)
	newData := []byte("New content for readonly file")

	err = AtomicWriteFile(testFile, newData, constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Atomic write to readonly file failed: %v", err)
	}

	// Verify new content and permissions
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(newData) {
		t.Errorf("Content mismatch. Expected %q, got %q", newData, content)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Mode().Perm() != constants.FilePerm0644 {
		t.Errorf("Permission should be restored to 0644, got %o", info.Mode().Perm())
	}
}

func TestAtomicWriteFile_ReadonlyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	readonlyDir := filepath.Join(tmpDir, "readonly")

	err = os.Mkdir(readonlyDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create readonly dir: %v", err)
	}

	// Make directory readonly
	err = os.Chmod(readonlyDir, 0555)
	if err != nil {
		t.Fatalf("Failed to make directory readonly: %v", err)
	}

	// Restore permissions for cleanup
	defer os.Chmod(readonlyDir, 0755)

	testFile := filepath.Join(readonlyDir, "test.txt")
	testData := []byte("Test data")

	// Should fail with permission error
	err = AtomicWriteFile(testFile, testData, constants.FilePerm0644)
	if err == nil {
		t.Errorf("Expected permission error, but write succeeded")
	}
}

func TestCheckWritePermission(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test writable directory
	err = CheckWritePermission(tmpDir)
	if err != nil {
		t.Errorf("Expected writable directory to pass: %v", err)
	}

	// Test non-existent file in writable directory
	nonExistentFile := filepath.Join(tmpDir, "nonexistent.txt")

	err = CheckWritePermission(nonExistentFile)
	if err != nil {
		t.Errorf("Expected non-existent file in writable dir to pass: %v", err)
	}

	// Create a file and test write permission
	testFile := filepath.Join(tmpDir, "test.txt")

	err = os.WriteFile(testFile, []byte("test"), constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = CheckWritePermission(testFile)
	if err != nil {
		t.Errorf("Expected writable file to pass: %v", err)
	}

	// Test readonly file (should fail for direct write check)
	err = os.Chmod(testFile, 0444)
	if err != nil {
		t.Fatalf("Failed to make file readonly: %v", err)
	}

	err = CheckWritePermission(testFile)
	if err == nil {
		t.Errorf("Expected readonly file to fail write permission check")
	}
}

func TestBackupFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	testData := []byte("Original content")

	// Create original file
	err = os.WriteFile(testFile, testData, constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create backup
	backupPath, err := BackupFile(testFile)
	if err != nil {
		t.Fatalf("BackupFile failed: %v", err)
	}

	// Verify backup exists and has correct content
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupContent) != string(testData) {
		t.Errorf("Backup content mismatch. Expected %q, got %q", testData, backupContent)
	}

	// Verify original file still exists
	originalContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Original file missing after backup: %v", err)
	}

	if string(originalContent) != string(testData) {
		t.Errorf("Original content changed after backup. Expected %q, got %q", testData, originalContent)
	}
}

func TestBackupFile_NonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nonExistentFile := filepath.Join(tmpDir, "nonexistent.txt")

	// Backup of non-existent file should return empty string, no error
	backupPath, err := BackupFile(nonExistentFile)
	if err != nil {
		t.Errorf("BackupFile of non-existent file should not error: %v", err)
	}

	if backupPath != "" {
		t.Errorf("Expected empty backup path for non-existent file, got %q", backupPath)
	}
}

func TestWrapFileError(t *testing.T) {
	testErr := errors.New("original error")

	// Test with valid error
	result := WrapFileError("failed to read", "/path/to/file", testErr)
	if result == nil {
		t.Error("Expected non-nil result")
	}

	expectedMsg := "failed to read /path/to/file: original error"
	if result.Error() != expectedMsg {
		t.Errorf("Error message mismatch. Expected %q, got %q", expectedMsg, result.Error())
	}

	// Test with nil error
	nilResult := WrapFileError("failed to read", "/path/to/file", nil)
	if nilResult != nil {
		t.Errorf("Expected nil result for nil input, got %v", nilResult)
	}

	// Test error wrapping with errors.Is
	wrappedErr := WrapFileError("operation", "/path", testErr)
	if !errors.Is(wrappedErr, testErr) {
		t.Error("errors.Is should return true for wrapped error")
	}
}

func TestWrapFileErrorCtx(t *testing.T) {
	testErr := errors.New("original error")

	// Test with valid error
	ctx := context.Background()
	result := WrapFileErrorCtx(ctx, "failed to read", "/path/to/file", testErr)
	if result == nil {
		t.Error("Expected non-nil result")
	}

	expectedMsg := "failed to read /path/to/file: original error"
	if result.Error() != expectedMsg {
		t.Errorf("Error message mismatch. Expected %q, got %q", expectedMsg, result.Error())
	}

	// Test with nil error
	nilResult := WrapFileErrorCtx(ctx, "failed to read", "/path/to/file", nil)
	if nilResult != nil {
		t.Errorf("Expected nil result for nil input, got %v", nilResult)
	}

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	result = WrapFileErrorCtx(cancelledCtx, "operation", "/path", testErr)
	if result != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", result)
	}

	// Test error wrapping with errors.Is
	wrappedErr := WrapFileErrorCtx(ctx, "operation", "/path", testErr)
	if !errors.Is(wrappedErr, testErr) {
		t.Error("errors.Is should return true for wrapped error")
	}
}

func TestLockFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	err = os.WriteFile(testFile, []byte("test content"), constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Acquire lock
	lockFile, err := LockFile(testFile, 5*time.Second)
	if err != nil {
		t.Fatalf("LockFile failed: %v", err)
	}
	if lockFile == nil {
		t.Fatal("LockFile returned nil file handle")
	}

	// Verify we can read/write the lock file
	_, err = lockFile.Write([]byte("locked"))
	if err != nil {
		t.Errorf("Failed to write to lock file: %v", err)
	}

	// Release lock
	err = UnlockFile(lockFile)
	if err != nil {
		t.Errorf("UnlockFile failed: %v", err)
	}

	// Verify original file is still intact
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("Original file content changed. Expected %q, got %q", "test content", content)
	}
}

func TestLockFile_NilHandle(t *testing.T) {
	// UnlockFile with nil should not panic and should return nil
	err := UnlockFile(nil)
	if err != nil {
		t.Errorf("UnlockFile(nil) should return nil, got %v", err)
	}
}

func TestLockFile_Timeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	err = os.WriteFile(testFile, []byte("test content"), constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Acquire first lock using blocking LockFile
	lockFile1, err := LockFile(testFile, 5*time.Second)
	if err != nil {
		t.Fatalf("First LockFile failed: %v", err)
	}
	defer UnlockFile(lockFile1)

	// Try to acquire second lock with TryLockFile - should return nil immediately
	lockFile2, err := TryLockFile(testFile)
	if err != nil {
		t.Fatalf("TryLockFile returned unexpected error: %v", err)
	}
	if lockFile2 != nil {
		t.Error("TryLockFile should return nil when lock is held by another process")
		UnlockFile(lockFile2)
	}
}

func TestTryLockFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	err = os.WriteFile(testFile, []byte("test content"), constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// First tryLock should succeed
	lockFile1, err := TryLockFile(testFile)
	if err != nil {
		t.Fatalf("First TryLockFile failed: %v", err)
	}
	if lockFile1 == nil {
		t.Fatal("First TryLockFile returned nil, expected lock file")
	}

	// Second tryLock should return nil (lock already held)
	lockFile2, err := TryLockFile(testFile)
	if err != nil {
		t.Fatalf("Second TryLockFile returned error: %v", err)
	}
	if lockFile2 != nil {
		t.Error("Second TryLockFile should return nil when lock is held")
		UnlockFile(lockFile2)
	}

	// Release first lock
	err = UnlockFile(lockFile1)
	if err != nil {
		t.Errorf("UnlockFile failed: %v", err)
	}

	// Third tryLock should succeed after first is released
	lockFile3, err := TryLockFile(testFile)
	if err != nil {
		t.Fatalf("Third TryLockFile failed: %v", err)
	}
	if lockFile3 == nil {
		t.Fatal("Third TryLockFile returned nil, expected lock file")
	}

	// Clean up
	UnlockFile(lockFile3)
}

func TestAtomicWriteFileCtx_Cancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	testData := []byte("Test data for cancellation")

	// Test with already cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err = AtomicWriteFileCtx(cancelledCtx, testFile, testData, constants.FilePerm0644)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}

	// Test with valid context - should succeed
	validCtx := context.Background()
	err = AtomicWriteFileCtx(validCtx, testFile, testData, constants.FilePerm0644)
	if err != nil {
		t.Errorf("AtomicWriteFileCtx with valid context failed: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != string(testData) {
		t.Errorf("Content mismatch. Expected %q, got %q", testData, content)
	}
}

func TestFindSectionIndex(t *testing.T) {
	tests := []struct {
		name        string
		lines       []string
		sectionName string
		wantIndex   int
	}{
		{
			name:        "section exists at beginning",
			lines:       []string{"## Tasks", "- [ ] Task 1", "- [ ] Task 2"},
			sectionName: "Tasks",
			wantIndex:   1,
		},
		{
			name:        "section exists with leading whitespace",
			lines:       []string{"## Tasks", "", "- [ ] Task 1"},
			sectionName: "Tasks",
			wantIndex:   2,
		},
		{
			name:        "section exists with trailing whitespace",
			lines:       []string{"## Tasks", "   ", "- [ ] Task 1"},
			sectionName: "Tasks",
			wantIndex:   2,
		},
		{
			name:        "section does not exist",
			lines:       []string{"## Other", "- [ ] Task 1"},
			sectionName: "Tasks",
			wantIndex:   -1,
		},
		{
			name:        "empty lines slice",
			lines:       []string{},
			sectionName: "Tasks",
			wantIndex:   -1,
		},
		{
			name:        "only header exists",
			lines:       []string{"## Tasks"},
			sectionName: "Tasks",
			wantIndex:   1,
		},
		{
			name:        "section with multiple headers - return first match",
			lines:       []string{"## Tasks", "- [ ] Task 1", "## Tasks", "- [ ] Task 2"},
			sectionName: "Tasks",
			wantIndex:   1,
		},
		{
			name:        "partial match not found",
			lines:       []string{"## Tasks", "- [ ] Task 1"},
			sectionName: "task",
			wantIndex:   -1,
		},
		{
			name:        "section at end of file",
			lines:       []string{"## Notes", "- [ ] Note 1", "## Tasks", "- [ ] Task 1"},
			sectionName: "Tasks",
			wantIndex:   3, // First non-empty line after "## Tasks" header
		},
		{
			name:        "section with content before and after",
			lines:       []string{"# Title", "", "## Tasks", "- [ ] Task 1", "", "## Notes", "- [ ] Note 1"},
			sectionName: "Tasks",
			wantIndex:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndex := FindSectionIndex(tt.lines, tt.sectionName)
			if gotIndex != tt.wantIndex {
				t.Errorf("FindSectionIndex(%q) = %d, want %d", tt.sectionName, gotIndex, tt.wantIndex)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test existing file
	existingFile := filepath.Join(tmpDir, "existing.txt")
	err = os.WriteFile(existingFile, []byte("content"), constants.FilePerm0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !FileExists(existingFile) {
		t.Error("FileExists should return true for existing file")
	}

	// Test non-existing file
	nonExistingFile := filepath.Join(tmpDir, "non-existing.txt")
	if FileExists(nonExistingFile) {
		t.Error("FileExists should return false for non-existing file")
	}

	// Test existing directory
	if !FileExists(tmpDir) {
		t.Error("FileExists should return true for existing directory")
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test creating new directory
	newDir := filepath.Join(tmpDir, "newsubdir")
	err = EnsureDir(newDir)
	if err != nil {
		t.Errorf("EnsureDir failed for new directory: %v", err)
	}

	if !FileExists(newDir) {
		t.Error("Directory should exist after EnsureDir")
	}

	// Test existing directory (should not fail)
	err = EnsureDir(newDir)
	if err != nil {
		t.Errorf("EnsureDir should not fail for existing directory: %v", err)
	}

	// Test nested directory creation
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")
	err = EnsureDir(nestedDir)
	if err != nil {
		t.Errorf("EnsureDir failed for nested directory: %v", err)
	}

	if !FileExists(nestedDir) {
		t.Error("Nested directory should exist after EnsureDir")
	}
}

func TestValidateEditor(t *testing.T) {
	// Test empty editor
	err := ValidateEditor("")
	if err == nil {
		t.Error("ValidateEditor should fail for empty editor")
	}

	// Test non-existent editor
	err = ValidateEditor("/nonexistent/editor")
	if err == nil {
		t.Error("ValidateEditor should fail for non-existent editor")
	}

	// Test valid editor (sh should exist on most systems)
	err = ValidateEditor("/bin/sh")
	if err != nil {
		t.Errorf("ValidateEditor should succeed for /bin/sh: %v", err)
	}
}

func TestCheckDiskSpace(t *testing.T) {
	// Test with sufficient disk space (1 byte required, any path has at least that much)
	err := CheckDiskSpace(t.TempDir(), 1)
	if err != nil {
		t.Errorf("CheckDiskSpace should succeed with sufficient space: %v", err)
	}

	// Test with file path instead of directory
	tmpFile := filepath.Join(t.TempDir(), "testfile.txt")
	err = CheckDiskSpace(tmpFile, 1)
	if err != nil {
		t.Errorf("CheckDiskSpace should succeed for file path: %v", err)
	}

	// Test with non-existent path (should use parent directory)
	err = CheckDiskSpace("/nonexistent/path/to/file.txt", 1)
	if err != nil {
		t.Errorf("CheckDiskSpace should succeed for non-existent path: %v", err)
	}

	// Test with zero bytes required
	err = CheckDiskSpace(t.TempDir(), 0)
	if err != nil {
		t.Errorf("CheckDiskSpace should succeed with zero required bytes: %v", err)
	}
}

func TestCheckDiskSpace_InsufficientSpace(t *testing.T) {
	// Test with a very large required bytes value that exceeds any normal disk
	// This should trigger the "insufficient disk space" error
	err := CheckDiskSpace(t.TempDir(), 1<<60) // 1 exabyte
	if err == nil {
		t.Error("CheckDiskSpace should fail with insufficient space")
	}
	if err != nil && !strings.Contains(err.Error(), "insufficient disk space") {
		t.Errorf("Error message should mention 'insufficient disk space', got: %v", err)
	}
}
