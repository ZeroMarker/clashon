// Package desktop integrates the GUI with the OS: login autostart and the
// clash:// URL scheme.
package desktop

import "errors"

// Schemes handled by the GUI.
var Schemes = []string{"clash", "clashmeta", "mihomo"}

const (
	appID   = "dev.clashon.clashon"
	appName = "ClashOn"
)

var ErrUnsupported = errors.New("not supported on this OS")

// SetAutostart starts exe (with args) at login when on is true.
func SetAutostart(exe string, on bool) error { return setAutostart(exe, on) }

// Autostart reports whether login autostart is enabled.
func Autostart() (bool, error) { return autostart() }

// RegisterSchemes makes exe the handler for clash:// links. macOS declares
// schemes in Info.plist instead, so this is a no-op there.
func RegisterSchemes(exe string) error { return registerSchemes(exe) }
