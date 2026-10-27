//go:build windows

package resourcecatalog

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func sourceFileID(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", ErrUnavailable
	}
	var info windows.ByHandleFileInformation
	err = windows.GetFileInformationByHandle(windows.Handle(f.Fd()), &info)
	errClose := f.Close()
	if err != nil || errClose != nil {
		return "", ErrUnsupported
	}
	if info.FileIndexHigh == 0 && info.FileIndexLow == 0 {
		return "", ErrUnsupported
	}
	return fmt.Sprintf("%d:%d:%d:%d:%d", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow, info.CreationTime.HighDateTime, info.CreationTime.LowDateTime), nil
}
