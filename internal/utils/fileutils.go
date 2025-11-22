package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// CheckWritePermission checks if we have write permission to a directory or file
func CheckWritePermission(path string) error {
	// If the file exists, check if we can write to it
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			// For directories, try to create a temp file
			tempFile := filepath.Join(path, ".write_test_temp")
			file, err := os.Create(tempFile)
			if err != nil {
				return fmt.Errorf("no write permission to directory %s: %w", path, err)
			}
			file.Close()
			os.Remove(tempFile)
			return nil
		} else {
			// For files, check if we can open for writing
			file, err := os.OpenFile(path, os.O_WRONLY, 0)
			if err != nil {
				return fmt.Errorf("no write permission to file %s: %w", path, err)
			}
			file.Close()
			return nil
		}
	}

	// File doesn't exist, check parent directory
	parentDir := filepath.Dir(path)
	return CheckWritePermission(parentDir)
}

// CheckDiskSpace checks if there's enough disk space for a write operation
func CheckDiskSpace(path string, requiredBytes int64) error {
	var stat syscall.Statfs_t

	// Get directory path if path is a file
	dir := path
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		dir = filepath.Dir(path)
	} else if os.IsNotExist(err) {
		dir = filepath.Dir(path)
	}

	err := syscall.Statfs(dir, &stat)
	if err != nil {
		// If we can't get disk space, don't fail the operation
		// Just log a warning and continue
		return nil
	}

	// Calculate available space
	availableBytes := int64(stat.Bavail) * int64(stat.Bsize)

	if availableBytes < requiredBytes {
		return fmt.Errorf("insufficient disk space: need %d bytes, have %d bytes available", requiredBytes, availableBytes)
	}

	return nil
}

// AtomicWriteFile writes data to a file atomically using a temporary file and rename
func AtomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	VerboseLog("Starting atomic write to: %s (%d bytes)", filename, len(data))

	// Check write permissions first
	if err := CheckWritePermission(filepath.Dir(filename)); err != nil {
		VerboseLogError("permission check", err)
		return fmt.Errorf("permission check failed: %w", err)
	}

	// Check disk space (with 10% buffer for safety)
	requiredSpace := int64(len(data)) + int64(len(data)/10) + 4096 // Add some buffer
	if err := CheckDiskSpace(filename, requiredSpace); err != nil {
		return fmt.Errorf("disk space check failed: %w", err)
	}

	// Create temp file in the same directory to ensure atomic rename
	dir := filepath.Dir(filename)
	tmpFile, err := os.CreateTemp(dir, "."+filepath.Base(filename)+".tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	tmpName := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		os.Remove(tmpName) // Clean up on any error
	}()

	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Ensure data is written to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close temp file
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Set correct permissions
	if err := os.Chmod(tmpName, perm); err != nil {
		return fmt.Errorf("failed to set permissions on temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpName, filename); err != nil {
		return fmt.Errorf("failed to rename temp file to target: %w", err)
	}

	VerboseLog("Atomic write completed successfully: %s", filename)
	return nil
}

// EnsureDir creates a directory with proper error handling
func EnsureDir(path string) error {
	if err := CheckWritePermission(filepath.Dir(path)); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", path, err)
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}

	return nil
}

// BackupFile creates a backup of an existing file
func BackupFile(filename string) (string, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// File doesn't exist, no backup needed
		return "", nil
	}

	backupName := filename + ".backup"

	// Read original file
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read original file for backup: %w", err)
	}

	// Write backup atomically
	if err := AtomicWriteFile(backupName, data, 0644); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupName, nil
}

// ValidateEditor checks if an editor command exists and is executable
func ValidateEditor(editorCmd string) error {
	if editorCmd == "" {
		return fmt.Errorf("editor command is empty")
	}

	// Split command to get just the executable name
	parts := filepath.SplitList(editorCmd)
	if len(parts) == 0 {
		return fmt.Errorf("invalid editor command: %s", editorCmd)
	}

	executable := parts[0]

	// Check if executable exists and is executable
	path, err := exec.LookPath(executable)
	if err != nil {
		return fmt.Errorf("editor '%s' not found in PATH: %w", executable, err)
	}

	// Check if it's actually executable
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot stat editor '%s': %w", path, err)
	}

	mode := info.Mode()
	if mode&0111 == 0 {
		return fmt.Errorf("editor '%s' is not executable", path)
	}

	return nil
}
