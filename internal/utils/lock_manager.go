package utils

import (
	"time"
)

// LockHandle represents an acquired file lock
type LockHandle interface {
	// Close releases the lock and closes the underlying file handle
	Close() error
}

// FileLockManager defines the interface for file-based locking
type FileLockManager interface {
	// LockFile acquires a blocking exclusive lock on a file
	LockFile(path string, timeout time.Duration) (LockHandle, error)

	// TryLockFile attempts to acquire a non-blocking exclusive lock on a file
	TryLockFile(path string) (LockHandle, error)

	// UnlockFile releases a lock
	UnlockFile(handle LockHandle) error
}
