//go:build !linux && !darwin && !windows

package privilege

func check() error        { return ErrUnsupported }
func grant(string) error  { return ErrUnsupported }
func revoke(string) error { return ErrUnsupported }
