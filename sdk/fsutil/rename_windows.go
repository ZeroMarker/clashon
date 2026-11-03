package fsutil

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

// rename retries while another process has the target open: Windows refuses
// to replace such a file (ERROR_ACCESS_DENIED / ERROR_SHARING_VIOLATION),
// and readers here only hold files for a moment.
func rename(from, to string) error {
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
