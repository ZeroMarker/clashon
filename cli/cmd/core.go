package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ZeroMarker/clashon/core/runner"
	"github.com/ZeroMarker/clashon/sdk/config"
	"github.com/ZeroMarker/clashon/sdk/daemon"
	"github.com/ZeroMarker/clashon/sdk/privilege"
)

// writeRuntimeConfig merges the current profile with settings.
func (e *env) writeRuntimeConfig() (string, error) {
	p, err := e.Store.Current()
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(e.Store.File(p.Name))
	if err != nil {
		return "", err
	}
	merged, err := config.Merge(raw, e.Settings)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(e.Layout.CoreHome(), 0o700); err != nil {
		return "", err
	}
	path := e.Layout.RuntimeConfig()
	return path, os.WriteFile(path, merged, 0o600)
}

func (e *env) runnerOptions() runner.Options {
	return runner.Options{HomeDir: e.Layout.CoreHome(), ConfigFile: e.Layout.RuntimeConfig()}
}

func upCmd() *cobra.Command {
	var foreground bool
	c := &cobra.Command{
		Use:   "up",
		Short: "Start the proxy core with the current profile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			if _, err := e.writeRuntimeConfig(); err != nil {
				return err
			}
			if foreground {
				ctx, stop := signalContext()
				defer stop()
				fmt.Printf("mixed proxy on :%d, controller %s (Ctrl-C to stop)\n", e.Settings.MixedPort, e.Settings.Controller)
				return runner.Run(ctx, e.runnerOptions())
			}
			exe, err := os.Executable()
			if err != nil {
				return err
			}
			pid, err := e.Daemon.Start(exe, "core", "run")
			if err != nil {
				return err
			}
			if err := e.waitReady(15 * time.Second); err != nil {
				_ = e.Daemon.Stop()
				return fmt.Errorf("%w\n%s", err, tail(e.Layout.LogFile(), 20))
			}
			fmt.Printf("core started (pid %d): mixed proxy 127.0.0.1:%d, controller %s\n", pid, e.Settings.MixedPort, e.Settings.Controller)
			return nil
		},
	}
	c.Flags().BoolVarP(&foreground, "foreground", "f", false, "run in the foreground")
	return c
}

func (e *env) waitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := e.Daemon.PID(); err != nil {
			return errors.New("core exited during startup")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := e.Client.Version(ctx)
		cancel()
		if err == nil {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return errors.New("core did not become ready in time")
}

func tail(path string, n int) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

func downCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Stop the background core",
		RunE: func(cmd *cobra.Command, _ []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			if err := e.Daemon.Stop(); err != nil {
				return err
			}
			fmt.Println("core stopped")
			return nil
		},
	}
}

// Status is the machine-readable output of `clashon status --json`.
type Status struct {
	Running    bool   `json:"running"`
	PID        int    `json:"pid,omitempty"`
	Reachable  bool   `json:"reachable"`
	Version    string `json:"version,omitempty"`
	Profile    string `json:"profile,omitempty"`
	Mode       string `json:"mode,omitempty"`
	MixedPort  int    `json:"mixedPort"`
	Controller string `json:"controller"`
	Secret     string `json:"secret"`
	TUN        bool   `json:"tun"`
	TUNReady   bool   `json:"tunReady"` // privileges for TUN are in place
}

func (e *env) status(ctx context.Context) (*Status, error) {
	st := &Status{
		MixedPort:  e.Settings.MixedPort,
		Controller: e.Settings.Controller,
		Secret:     e.Settings.Secret,
		TUN:        e.Settings.TUN,
		TUNReady:   privilege.Check() == nil,
	}
	if p, err := e.Store.Current(); err == nil {
		st.Profile = p.Name
	}
	pid, err := e.Daemon.PID()
	if errors.Is(err, daemon.ErrNotRunning) {
		return st, nil
	}
	if err != nil {
		return nil, err
	}
	st.Running, st.PID = true, pid
	v, verr := e.Client.Version(ctx)
	cfg, cerr := e.Client.Configs(ctx)
	if verr == nil && cerr == nil {
		st.Reachable, st.Version, st.Mode, st.MixedPort = true, v, cfg.Mode, cfg.MixedPort
	}
	return st, nil
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show core status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			st, err := e.status(cmd.Context())
			if err != nil {
				return err
			}
			if jsonOut {
				return printJSON(st)
			}
			switch {
			case !st.Running:
				fmt.Printf("core:    stopped\nprofile: %s\n", orDash(st.Profile))
			case !st.Reachable:
				fmt.Printf("core:    running (pid %d), controller unreachable\nprofile: %s\n", st.PID, orDash(st.Profile))
			default:
				fmt.Printf("core:    running (pid %d, mihomo %s)\nprofile: %s\nmode:    %s\nproxy:   127.0.0.1:%d\n", st.PID, st.Version, orDash(st.Profile), st.Mode, st.MixedPort)
			}
			return nil
		},
	}
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func coreCmd() *cobra.Command {
	c := &cobra.Command{Use: "core", Short: "Internal core commands", Hidden: true}
	c.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Run mihomo with the generated runtime config",
		RunE: func(cmd *cobra.Command, _ []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			ctx, stop := signalContext()
			defer stop()
			return runner.Run(ctx, e.runnerOptions())
		},
	})
	return c
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print versions",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Printf("clashon %s (mihomo %s)\n", Version, runner.Version())
		},
	}
}

// Version is set via -ldflags "-X github.com/ZeroMarker/clashon/cli/cmd.Version=...".
var Version = "dev"
