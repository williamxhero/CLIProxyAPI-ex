//go:build linux

package resourcecatalog

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func sourceFileID(path string) (string, error) {
	var info unix.Statx_t
	if err := unix.Statx(unix.AT_FDCWD, path, unix.AT_SYMLINK_NOFOLLOW, unix.STATX_INO|unix.STATX_BTIME|unix.STATX_TYPE, &info); err != nil {
		return "", ErrUnsupported
	}
	if info.Mask&unix.STATX_BTIME == 0 || info.Mode&unix.S_IFMT != unix.S_IFREG {
		return "", ErrUnsupported
	}
	return fmt.Sprintf("%d:%d:%d:%d:%d", info.Dev_major, info.Dev_minor, info.Ino, info.Btime.Sec, info.Btime.Nsec), nil
}
