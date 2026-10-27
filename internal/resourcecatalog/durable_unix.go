//go:build unix

package resourcecatalog

import (
	"os"
	"path/filepath"
)

func replaceFile(from, to string) error {
	if err := os.Rename(from, to); err != nil {
		return ErrUnavailable
	}
	return syncDirectory(filepath.Dir(to))
}
func syncDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return ErrUnavailable
	}
	err = f.Sync()
	errClose := f.Close()
	if err != nil || errClose != nil {
		return ErrUnavailable
	}
	return nil
}
