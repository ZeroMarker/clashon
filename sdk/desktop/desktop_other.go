//go:build !linux && !darwin && !windows

package desktop

func setAutostart(string, bool) error { return ErrUnsupported }
func autostart() (bool, error)        { return false, ErrUnsupported }
func registerSchemes(string) error    { return ErrUnsupported }
