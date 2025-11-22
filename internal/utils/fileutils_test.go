package utils

import (
	"os"
	"path/filepath"
	"testing"
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
	err = AtomicWriteFile(testFile, testData, 0644)
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

	if info.Mode().Perm() != 0644 {
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
	err = AtomicWriteFile(testFile, initialData, 0644)
	if err != nil {
		t.Fatalf("Initial write failed: %v", err)
	}

	// Overwrite with new content
	newData := []byte("Updated content")
	err = AtomicWriteFile(testFile, newData, 0644)
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
	err = AtomicWriteFile(testFile, initialData, 0644)
	if err != nil {
		t.Fatalf("Initial write failed: %v", err)
	}

	err = os.Chmod(testFile, 0444)
	if err != nil {
		t.Fatalf("Failed to make file readonly: %v", err)
	}

	// Atomic write should still work (creates new file, renames)
	newData := []byte("New content for readonly file")
	err = AtomicWriteFile(testFile, newData, 0644)
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

	if info.Mode().Perm() != 0644 {
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
	err = AtomicWriteFile(testFile, testData, 0644)
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
	err = os.WriteFile(testFile, []byte("test"), 0644)
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
	err = os.WriteFile(testFile, testData, 0644)
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
