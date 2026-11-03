package fsutil

import (
	"os"

	"golang.org/x/sys/windows"
)

// Open opens a file for reading with FILE_SHARE_DELETE, so a writer in
// another process can still replace it while it is open. os.Open does not
// share delete access, which makes atomic replaces fail with
// ERROR_ACCESS_DENIED.
func Open(path string) (*os.File, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(h), path), nil
}
