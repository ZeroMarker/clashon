package privilege

import "golang.org/x/sys/windows"

func check() error {
	if windows.GetCurrentProcessToken().IsElevated() {
		return nil
	}
	return ErrMissing
}

func grant(string) error  { return ErrUnsupported }
func revoke(string) error { return ErrUnsupported }
