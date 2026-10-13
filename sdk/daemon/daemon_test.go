//go:build !windows

package daemon

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestStartStop(t *testing.T) {
	dir := t.TempDir()
	d := &Daemon{PIDFile: filepath.Join(dir, "pid"), LogFile: filepath.Join(dir, "log")}
	if _, err := d.PID(); !errors.Is(err, ErrNotRunning) {
		t.Fatalf("PID before start: %v", err)
	}
	pid, err := d.Start("sleep", "30")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := d.PID(); err != nil || got != pid {
		t.Fatalf("PID = %d, %v", got, err)
	}
	if _, err := d.Start("sleep", "30"); err == nil {
		t.Fatal("second start should fail")
	}
	if err := d.Stop(); err != nil {
		t.Fatal(err)
	}
	if _, err := d.PID(); !errors.Is(err, ErrNotRunning) {
		t.Fatalf("PID after stop: %v", err)
	}
}
