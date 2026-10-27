//go:build !windows && !unix

package resourcecatalog

import "os"

func lockWriter(string) (*os.File, error) { return nil, ErrUnavailable }
