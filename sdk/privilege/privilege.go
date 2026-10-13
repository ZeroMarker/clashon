// Package privilege checks and grants what the core needs to create a TUN device.
package privilege

import "errors"

// ErrMissing means the current binary cannot create a TUN device.
var ErrMissing = errors.New("TUN needs elevated privileges, run `sudo clashon service install` first")

// ErrUnsupported means Grant is not implemented on this OS yet.
var ErrUnsupported = errors.New("privilege setup is not supported on this OS yet; run clashon as administrator/root")

// Check returns nil if processes started from this binary can create a TUN device.
func Check() error { return check() }

// Grant gives exe the privileges it needs. It must itself run as root.
func Grant(exe string) error { return grant(exe) }

// Revoke undoes Grant.
func Revoke(exe string) error { return revoke(exe) }
