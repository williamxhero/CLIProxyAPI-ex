//go:build !windows && !linux

package resourcecatalog

func sourceFileID(string) (string, error) { return "", ErrUnsupported }
