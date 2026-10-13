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
)

// writeRuntimeConfig merges the current profile with settings.
func (e *env) writeRuntimeConfig() (string, error) {
	p, err := e.store.Current()
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(e.store.File(p.Name))
	if err != nil {
		return "", err
	}
	merged, err := config.Merge(raw, e.settings)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(e.layout.CoreHome(), 0o700); err != nil {
		return "", err
	}
	path := e.layout.RuntimeConfig()
	return path, os.WriteFile(path, merged, 0o600)
}

func (e *env) runnerOptions() runner.Options {
	return runner.Options{HomeDir: e.layout.CoreHome(), ConfigFile: e.layout.RuntimeConfig()}
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
				fmt.Printf("mixed proxy on :%d, controller %s (Ctrl-C to stop)\n", e.settings.MixedPort, e.settings.Controller)
				return runner.Run(ctx, e.runnerOptions())
			}
			exe, err := os.Executable()
			if err != nil {
				return err
			}
			pid, err := e.daemon.Start(exe, "core", "run")
			if err != nil {
				return err
			}
			if err := e.waitReady(15 * time.Second); err != nil {
				_ = e.daemon.Stop()
				return fmt.Errorf("%w\n%s", err, tail(e.layout.LogFile(), 20))
			}
			fmt.Printf("core started (pid %d): mixed proxy 127.0.0.1:%d, controller %s\n", pid, e.settings.MixedPort, e.settings.Controller)
			return nil
		},
	}
	c.Flags().BoolVarP(&foreground, "foreground", "f", false, "run in the foreground")
	return c
}

func (e *env) waitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := e.daemon.PID(); err != nil {
			return errors.New("core exited during startup")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := e.client.Version(ctx)
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
			if err := e.daemon.Stop(); err != nil {
				return err
			}
			fmt.Println("core stopped")
			return nil
		},
	}
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
			cur := "-"
			if p, err := e.store.Current(); err == nil {
				cur = p.Name
			}
			pid, err := e.daemon.PID()
			if errors.Is(err, daemon.ErrNotRunning) {
				fmt.Printf("core:    stopped\nprofile: %s\n", cur)
				return nil
			}
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			v, verr := e.client.Version(ctx)
			cfg, cerr := e.client.Configs(ctx)
			if verr != nil || cerr != nil {
				fmt.Printf("core:    running (pid %d), controller unreachable\nprofile: %s\n", pid, cur)
				return nil
			}
			fmt.Printf("core:    running (pid %d, mihomo %s)\nprofile: %s\nmode:    %s\nproxy:   127.0.0.1:%d\n", pid, v, cur, cfg.Mode, cfg.MixedPort)
			return nil
		},
	}
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
