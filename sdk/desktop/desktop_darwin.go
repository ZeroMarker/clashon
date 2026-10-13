package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func agentFile() string {
	h, _ := os.UserHomeDir()
	return filepath.Join(h, "Library", "LaunchAgents", appID+".plist")
}

// appBundle maps .../ClashOn.app/Contents/MacOS/ClashOn to the .app so
// LaunchServices opens it normally.
func appBundle(exe string) string {
	if i := strings.Index(exe, ".app/Contents/MacOS/"); i >= 0 {
		return exe[:i+4]
	}
	return exe
}

func setAutostart(exe string, on bool) error {
	if !on {
		if err := os.Remove(agentFile()); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key><string>%s</string>
	<key>ProgramArguments</key>
	<array><string>/usr/bin/open</string><string>-g</string><string>%s</string><string>--args</string><string>--minimized</string></array>
	<key>RunAtLoad</key><true/>
</dict>
</plist>
`, appID, xmlEscape(appBundle(exe)))
	if err := os.MkdirAll(filepath.Dir(agentFile()), 0o755); err != nil {
		return err
	}
	return os.WriteFile(agentFile(), []byte(plist), 0o644)
}

func xmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

func autostart() (bool, error) {
	_, err := os.Stat(agentFile())
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func registerSchemes(string) error { return nil }
