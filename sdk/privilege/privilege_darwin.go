package privilege

import "os"

func check() error {
	if os.Geteuid() == 0 {
		return nil
	}
	return ErrMissing
}

func grant(string) error  { return ErrUnsupported }
func revoke(string) error { return ErrUnsupported }
