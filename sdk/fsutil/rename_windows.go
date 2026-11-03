package fsutil

import (
	"errors"
	"os"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// rename replaces to with from even while other processes have to open.
//
// MoveFileEx (os.Rename) uses legacy semantics and fails with
// ERROR_ACCESS_DENIED whenever the target is open. POSIX-semantics renames
// (Windows 10 1809+, NTFS) succeed as long as the readers share delete
// access, which fsutil.Open does. Elsewhere fall back to os.Rename with a
// short retry for readers that only hold the file for a moment.
func rename(from, to string) error {
	if err := posixRename(from, to); err == nil {
		return nil
	}
	deadline := time.Now().Add(2 * time.Second)
	delay := 2 * time.Millisecond
	for {
		err := os.Rename(from, to)
		if err == nil || !retryable(err) || time.Now().After(deadline) {
			return err
		}
		time.Sleep(delay)
		delay = min(delay*2, 50*time.Millisecond)
	}
}

func retryable(err error) bool {
	return errors.Is(err, windows.ERROR_ACCESS_DENIED) || errors.Is(err, windows.ERROR_SHARING_VIOLATION)
}

const (
	fileRenameInfoEx              = 22 // FILE_INFO_BY_HANDLE_CLASS
	fileRenameFlagReplaceIfExists = 0x1
	fileRenameFlagPosixSemantics  = 0x2
	fileRenameInfoFileNameOffset  = unsafe.Offsetof(fileRenameInfo{}.FileName
)
	fileRenameInfoHeaderSize       = int(fileRenameInfoFileNameOffset)
)

// fileRenameInfo mirrors FILE_RENAME_INFO with the Flags member of the union.
type fileRenameInfo struct {
	Flags          uint32
	RootDirectory  windows.Handle
	FileNameLength uint32
	FileName       [1]uint16
}

func posixRename(from, to string) error {
	src, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	dst, err := windows.UTF16FromString(to)
	if err != nil {
		return err
	}
	h, err := windows.CreateFile(src, windows.DELETE|windows.SYNCHRONIZE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)

	name := dst[:len(dst)-1] // without the terminating NUL
	size := fileRenameInfoHeaderSize + (len(name)+1)*2
	buf := make([]uint64, (size+7)/8) // 8-byte aligned backing store
	info := (*fileRenameInfo)(unsafe.Pointer(&buf[0]))
	info.Flags = fileRenameFlagReplaceIfExists | fileRenameFlagPosixSemantics
	info.FileNameLength = uint32(len(name) * 2)
	copy(unsafe.Slice(&info.FileName[0], len(name)+1), dst)
	return windows.SetFileInformationByHandle(h, fileRenameInfoEx, (*byte)(unsafe.Pointer(info)), uint32(size))
}
