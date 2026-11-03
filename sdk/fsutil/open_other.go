//go:build !windows

package fsutil

import "os"

// Open opens a file for reading.
func Open(path string) (*os.File, error) { return os.Open(path) }
