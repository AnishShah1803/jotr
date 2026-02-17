package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/utils/platform"
)

var ErrLockTimeout = errors.New("timeout waiting for file lock")

func LockFile(path string, timeout time.Duration) (*os.File, error) {
	lockPath := path + ".lock"

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, constants.FilePerm0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	deadline := time.Now().Add(timeout)

	for {
		err := platform.Flock(int(lockFile.Fd()), platform.LOCK_EX)
		if err == nil {
			return lockFile, nil
		}

		if time.Now().After(deadline) {
			lockFile.Close()
			return nil, fmt.Errorf("%w: %s", ErrLockTimeout, path)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// UnlockFile releases the lock acquired by LockFile and closes the file handle.
func UnlockFile(lockFile *os.File) error {
	if lockFile == nil {
		return nil
	}

	// Release the lock
	err := platform.Flock(int(lockFile.Fd()), platform.LOCK_UN)
	if err != nil {
		// Still try to close the file even if unlock fails
		closeErr := lockFile.Close()
		// Use errors.Join to preserve both errors in the chain
		return fmt.Errorf("failed to release file lock: %w", errors.Join(err, closeErr))
	}

	return lockFile.Close()
}

// TryLockFile attempts to acquire a non-blocking exclusive lock on a file.
// Returns nil, nil if lock is immediately available.
// Returns (nil, error) if lock cannot be acquired.
func TryLockFile(path string) (*os.File, error) {
	lockPath := path + ".lock"

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, constants.FilePerm0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	err = platform.Flock(int(lockFile.Fd()), platform.LOCK_EX|platform.LOCK_NB)
	if err != nil {
		lockFile.Close()

		if platform.IsLockBusy(err) {
			return nil, nil // Lock is held by another process
		}
		if err == platform.ErrNotSupported {
			// On platforms that don't support locking, assume success
			return lockFile, nil
		}

		return nil, fmt.Errorf("failed to acquire file lock: %w", err)
	}

	return lockFile, nil
}

// CheckWritePermission checks if we have write permission to a directory or file.
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
		}
		// For files, check if we can open for writing
		file, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			return fmt.Errorf("no write permission to file %s: %w", path, err)
		}

		file.Close()

		return nil
	}

	// File doesn't exist, check parent directory
	parentDir := filepath.Dir(path)

	return CheckWritePermission(parentDir)
}

// CheckDiskSpace checks if there's enough disk space for a write operation.
func CheckDiskSpace(path string, requiredBytes int64) error {
	// Get directory path if path is a file
	dir := path
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		dir = filepath.Dir(path)
	} else if os.IsNotExist(err) {
		dir = filepath.Dir(path)
	}

	stat, err := platform.Statfs(dir)
	if err != nil {
		// If we can't get disk space, don't fail the operation
		// Just log a warning and continue
		return nil
	}

	// Calculate available space
	availableBytes := int64(stat.BavailField()) * int64(stat.BsizeField())

	if availableBytes < requiredBytes {
		return fmt.Errorf("insufficient disk space: need %d bytes, have %d bytes available", requiredBytes, availableBytes)
	}

	return nil
}

// This is a wrapper around AtomicWriteFileCtx for backward compatibility.
func AtomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	return AtomicWriteFileCtx(context.Background(), filename, data, perm)
}

// AtomicWriteFileCtx writes data to a file atomically with context support.
func AtomicWriteFileCtx(ctx context.Context, filename string, data []byte, perm os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	VerboseLogWithContext(ctx, "Starting atomic write to: %s (%d bytes)", filename, len(data))

	// Check write permissions first
	if err := CheckWritePermission(filepath.Dir(filename)); err != nil {
		VerboseLogErrorWithContext(ctx, "permission check", err)
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

	VerboseLogWithContext(ctx, "Atomic write completed successfully: %s", filename)

	return nil
}

// EnsureDir creates a directory with proper error handling.
func EnsureDir(path string) error {
	if err := CheckWritePermission(filepath.Dir(path)); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", path, err)
	}

	if err := os.MkdirAll(path, constants.FilePermDir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}

	return nil
}

// BackupFile creates a backup of an existing file.
func BackupFile(filename string) (string, error) {
	return BackupFileCtx(context.Background(), filename)
}

// BackupFileCtx creates a backup of an existing file with context support.
func BackupFileCtx(ctx context.Context, filename string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

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
	if err := AtomicWriteFileCtx(ctx, backupName, data, constants.FilePerm0644); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupName, nil
}

// FileExists checks if a file or directory exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// FindSectionIndex finds the index of a markdown section header in a slice of lines.
// It looks for a line starting with "## " followed by the section name.
// Returns the index of the first non-empty line after the section header,
// or -1 if the section is not found.
func FindSectionIndex(lines []string, sectionName string) int {
	prefix := "## " + sectionName
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			insertIndex := i + 1
			for insertIndex < len(lines) && strings.TrimSpace(lines[insertIndex]) == "" {
				insertIndex++
			}

			return insertIndex
		}
	}

	return -1
}

// WrapFileError wraps file operation errors with operation and path context.
// This provides consistent error messages across the codebase for file operations.
// Returns nil if the input error is nil.
func WrapFileError(operation string, path string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s %s: %w", operation, path, err)
}

// WrapFileErrorCtx wraps file operation errors with context support and verbose logging.
// Use this for operations that may need to be cancelled or traced.
func WrapFileErrorCtx(ctx context.Context, operation string, path string, err error) error {
	if err == nil {
		return nil
	}

	VerboseLogErrorWithContext(ctx, operation+" "+path, err)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("%s %s: %w", operation, path, err)
	}
}

// ValidateEditor checks if an editor command exists and is executable.
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
