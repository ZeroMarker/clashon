//go:build !linux && !darwin && !windows

package sysproxy

func enable(string, int) error           { return ErrUnsupported }
func disable() error                     { return ErrUnsupported }
func enabled(string, int) (bool, error) { return false, ErrUnsupported }
