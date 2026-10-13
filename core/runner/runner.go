// Package runner embeds mihomo and runs it in-process.
package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/metacubex/mihomo/config"
	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/hub"
	"github.com/metacubex/mihomo/hub/executor"
)

// Options controls an embedded mihomo instance.
type Options struct {
	HomeDir    string // mihomo home: geodata, providers, cache
	ConfigFile string // full config, already merged with overrides
}

// Validate parses the config without starting anything.
func Validate(opts Options) error {
	if err := setup(opts); err != nil {
		return err
	}
	_, err := executor.Parse()
	return err
}

// Run starts mihomo and blocks until ctx is cancelled.
func Run(ctx context.Context, opts Options) error {
	if err := setup(opts); err != nil {
		return err
	}
	if err := hub.Parse(nil); err != nil {
		return fmt.Errorf("start core: %w", err)
	}
	<-ctx.Done()
	executor.Shutdown()
	return nil
}

func setup(opts Options) error {
	home, err := filepath.Abs(opts.HomeDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(home, 0o755); err != nil {
		return err
	}
	cfg, err := filepath.Abs(opts.ConfigFile)
	if err != nil {
		return err
	}
	C.SetHomeDir(home)
	C.SetConfig(cfg)
	return config.Init(home)
}

// Version returns the embedded mihomo version.
func Version() string { return C.Version }
