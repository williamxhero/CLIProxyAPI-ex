//go:build !windows && !unix

package resourcecatalog

func replaceFile(string, string) error { return ErrUnavailable }
func syncDirectory(string) error       { return ErrUnavailable }
