//go:build windows

package platform

import (
	"errors"
)

const (
	LOCK_EX     = 0
	LOCK_UN     = 0
	LOCK_NB     = 0
	EWOULDBLOCK = 0
)

var ErrNotSupported = errors.New("file locking not supported on this platform")

func IsLockBusy(err error) bool {
	return false
}

func Flock(fd int, how int) error {
	return ErrNotSupported
}

func Statfs(path string) (*Statfs_t, error) {
	return nil, ErrNotSupported
}

type Statfs_t struct {
	Bsize  uint64
	Bavail uint64
}

func (s *Statfs_t) BsizeField() uint64 {
	return s.Bsize
}

func (s *Statfs_t) BavailField() uint64 {
	return s.Bavail
}
