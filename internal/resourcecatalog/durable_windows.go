//go:build windows

package resourcecatalog

import "golang.org/x/sys/windows"

func replaceFile(from, to string) error {
	src, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return ErrUnavailable
	}
	dst, err := windows.UTF16PtrFromString(to)
	if err != nil {
		return ErrUnavailable
	}
	if err = windows.MoveFileEx(src, dst, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH); err != nil {
		return ErrUnavailable
	}
	return nil
}

// Windows replacement requests write-through; directory fsync is not a Win32 API.
func syncDirectory(string) error { return nil }
