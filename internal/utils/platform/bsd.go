//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package platform

import (
	"syscall"
)

const (
	LOCK_EX     = syscall.LOCK_EX
	LOCK_UN     = syscall.LOCK_UN
	LOCK_NB     = syscall.LOCK_NB
	EWOULDBLOCK = syscall.EWOULDBLOCK
)

var ErrNotSupported error = nil

// IsLockBusy checks if the error indicates the lock is busy
func IsLockBusy(err error) bool {
	return err == syscall.EWOULDBLOCK
}

// Flock locks the file descriptor
func Flock(fd int, how int) error {
	return syscall.Flock(fd, how)
}

// Statfs gets file system statistics
func Statfs(path string) (*Statfs_t, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}
	return &Statfs_t{
		Bsize:  uint64(stat.Bsize),
		Bavail: uint64(stat.Bavail),
	}, nil
}

// Statfs_t is the file system statistics structure
type Statfs_t struct {
	Bsize  uint64
	Bavail uint64
}

// BsizeField returns the block size
func (s *Statfs_t) BsizeField() uint64 {
	return s.Bsize
}

// BavailField returns the available blocks
func (s *Statfs_t) BavailField() uint64 {
	return s.Bavail
}
