//go:build linux

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

func IsLockBusy(err error) bool {
	return err == syscall.EWOULDBLOCK
}

func Flock(fd int, how int) error {
	return syscall.Flock(fd, how)
}

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
