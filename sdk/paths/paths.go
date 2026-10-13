// Package paths resolves ClashOn's on-disk layout.
package paths

import (
	"os"
	"path/filepath"
)

// Layout is the set of files ClashOn reads and writes.
type Layout struct {
	Root string
}

// Default returns $CLASHON_HOME, or <user config dir>/clashon.
func Default() (Layout, error) {
	if h := os.Getenv("CLASHON_HOME"); h != "" {
		return Layout{Root: h}, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return Layout{}, err
	}
	return Layout{Root: filepath.Join(dir, "clashon")}, nil
}

func (l Layout) Profiles() string     { return filepath.Join(l.Root, "profiles") }
func (l Layout) ProfileIndex() string { return filepath.Join(l.Root, "profiles.json") }
func (l Layout) Settings() string     { return filepath.Join(l.Root, "settings.json") }
func (l Layout) CoreHome() string     { return filepath.Join(l.Root, "core") }
func (l Layout) RuntimeConfig() string {
	return filepath.Join(l.Root, "core", "config.yaml")
}
func (l Layout) PIDFile() string { return filepath.Join(l.Root, "core.pid") }
func (l Layout) LogFile() string { return filepath.Join(l.Root, "core.log") }
