// Package app wires the SDK pieces together for front ends (CLI, TUI, GUI).
package app

import (
	"errors"
	"os"

	"github.com/ZeroMarker/clashon/sdk/api"
	"github.com/ZeroMarker/clashon/sdk/config"
	"github.com/ZeroMarker/clashon/sdk/daemon"
	"github.com/ZeroMarker/clashon/sdk/paths"
	"github.com/ZeroMarker/clashon/sdk/profile"
)

// App holds everything a front end needs.
type App struct {
	Layout   paths.Layout
	Settings config.Settings
	Store    *profile.Store
	Daemon   *daemon.Daemon
	Client   *api.Client
}

// ErrCoreNotRunning is returned by RequireCore.
var ErrCoreNotRunning = errors.New("core is not running, start it with `clashon up`")

// Load reads settings from the default layout, creating it if needed.
func Load() (*App, error) {
	l, err := paths.Default()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(l.Root, 0o700); err != nil {
		return nil, err
	}
	s, err := config.LoadSettings(l.Settings())
	if err != nil {
		return nil, err
	}
	return &App{
		Layout:   l,
		Settings: s,
		Store:    &profile.Store{Dir: l.Profiles(), Index: l.ProfileIndex()},
		Daemon:   &daemon.Daemon{PIDFile: l.PIDFile(), LogFile: l.LogFile()},
		Client:   api.New(s.Controller, s.Secret),
	}, nil
}

// RequireCore is Load plus a check that the background core is running.
func RequireCore() (*App, error) {
	a, err := Load()
	if err != nil {
		return nil, err
	}
	if _, err := a.Daemon.PID(); errors.Is(err, daemon.ErrNotRunning) {
		return nil, ErrCoreNotRunning
	}
	return a, nil
}
