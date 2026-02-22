package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/utils/platform"
)

type DefaultFileLockManager struct{}

func NewDefaultFileLockManager() *DefaultFileLockManager {
	return &DefaultFileLockManager{}
}

func (m *DefaultFileLockManager) LockFile(path string, timeout time.Duration) (LockHandle, error) {
	lockPath := path + ".lock"

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, constants.FilePerm0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	deadline := time.Now().Add(timeout)

	for {
		err := platform.Flock(int(lockFile.Fd()), platform.LOCK_EX)
		if err == nil {
			return &lockHandleFile{file: lockFile}, nil
		}

		if time.Now().After(deadline) {
			lockFile.Close()
			return nil, fmt.Errorf("%w: %s", ErrLockTimeout, path)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func (m *DefaultFileLockManager) TryLockFile(path string) (LockHandle, error) {
	lockPath := path + ".lock"

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, constants.FilePerm0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	err = platform.Flock(int(lockFile.Fd()), platform.LOCK_EX|platform.LOCK_NB)
	if err != nil {
		lockFile.Close()

		if platform.IsLockBusy(err) {
			return nil, nil
		}
		if err == platform.ErrNotSupported {
			return &lockHandleFile{file: lockFile}, nil
		}

		return nil, fmt.Errorf("failed to acquire file lock: %w", err)
	}

	return &lockHandleFile{file: lockFile}, nil
}

func (m *DefaultFileLockManager) UnlockFile(handle LockHandle) error {
	if handle == nil {
		return nil
	}
	return handle.Close()
}

type lockHandleFile struct {
	file *os.File
}

func (h *lockHandleFile) Close() error {
	if h.file == nil {
		return nil
	}

	err := platform.Flock(int(h.file.Fd()), platform.LOCK_UN)
	if err != nil {
		closeErr := h.file.Close()
		return fmt.Errorf("failed to release file lock: %w", errors.Join(err, closeErr))
	}

	return h.file.Close()
}
