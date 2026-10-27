//go:build unix

package resourcecatalog

import (
	"os"

	"golang.org/x/sys/unix"
)

func lockWriter(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, ErrUnavailable
	}
	if err = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, ErrUnavailable
	}
	return f, nil
}
