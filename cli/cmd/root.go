// Package cmd implements the clashon command line.
package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/ZeroMarker/clashon/sdk/app"
)

// env is the CLI's view of the app; methods live in core.go and sub.go.
type env struct{ *app.App }

func loadEnv() (*env, error) {
	a, err := app.Load()
	if err != nil {
		return nil, err
	}
	return &env{a}, nil
}

func requireCore() (*env, error) {
	a, err := app.RequireCore()
	if err != nil {
		return nil, err
	}
	return &env{a}, nil
}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

// Execute runs the root command.
func Execute() error {
	root := &cobra.Command{
		Use:          "clashon",
		Short:        "Clash (mihomo) client",
		SilenceUsage: true,
	}
	root.AddCommand(upCmd(), downCmd(), statusCmd(), subCmd(), proxyCmd(), modeCmd(), ruleCmd(), logCmd(), tuiCmd(), versionCmd(), coreCmd())
	return root.Execute()
}
