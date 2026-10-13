// Package cmd implements the clashon command line.
package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/ZeroMarker/clashon/sdk/api"
	"github.com/ZeroMarker/clashon/sdk/config"
	"github.com/ZeroMarker/clashon/sdk/daemon"
	"github.com/ZeroMarker/clashon/sdk/paths"
	"github.com/ZeroMarker/clashon/sdk/profile"
)

// env bundles everything commands need.
type env struct {
	layout   paths.Layout
	settings config.Settings
	store    *profile.Store
	daemon   *daemon.Daemon
	client   *api.Client
}

func loadEnv() (*env, error) {
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
	return &env{
		layout:   l,
		settings: s,
		store:    &profile.Store{Dir: l.Profiles(), Index: l.ProfileIndex()},
		daemon:   &daemon.Daemon{PIDFile: l.PIDFile(), LogFile: l.LogFile()},
		client:   api.New(s.Controller, s.Secret),
	}, nil
}

// requireCore returns env for commands that talk to a running core.
func requireCore() (*env, error) {
	e, err := loadEnv()
	if err != nil {
		return nil, err
	}
	if _, err := e.daemon.PID(); errors.Is(err, daemon.ErrNotRunning) {
		return nil, errors.New("core is not running, start it with `clashon up`")
	}
	return e, nil
}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

// Execute runs the root command.
func Execute() error {
	root := &cobra.Command{
		Use:           "clashon",
		Short:         "Clash (mihomo) client",
		SilenceUsage: true,
	}
	root.AddCommand(upCmd(), downCmd(), statusCmd(), subCmd(), proxyCmd(), modeCmd(), ruleCmd(), logCmd(), versionCmd(), coreCmd())
	return root.Execute()
}
