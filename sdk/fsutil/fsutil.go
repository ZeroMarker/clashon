// Package fsutil writes files atomically, safely against concurrent readers
// in other processes (the GUI, TUI and CLI all read the same files).
package fsutil

import (
	"os"
	"path/filepath"
)

// WriteFile writes data to path via a temp file and rename, so readers see
// either the old or the new content, never a partial file.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return rename(tmp.Name(), path)
}
