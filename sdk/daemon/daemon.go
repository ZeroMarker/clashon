// Package daemon starts and stops the background core process.
package daemon

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Daemon manages a detached child process tracked by a PID file.
type Daemon struct {
	PIDFile string
	LogFile string
}

var ErrNotRunning = errors.New("core is not running")

// Start launches exe with args detached from the terminal, logging to LogFile.
func (d *Daemon) Start(exe string, args ...string) (int, error) {
	if pid, err := d.PID(); err == nil {
		return pid, fmt.Errorf("core already running (pid %d)", pid)
	}
	logf, err := os.OpenFile(d.LogFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return 0, err
	}
	defer logf.Close()
	cmd := exec.Command(exe, args...)
	cmd.Stdout, cmd.Stderr = logf, logf
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	pid := cmd.Process.Pid
	_ = cmd.Process.Release()
	return pid, os.WriteFile(d.PIDFile, []byte(strconv.Itoa(pid)), 0o600)
}

// PID returns the running core's pid, cleaning up a stale PID file.
func (d *Daemon) PID() (int, error) {
	b, err := os.ReadFile(d.PIDFile)
	if os.IsNotExist(err) {
		return 0, ErrNotRunning
	}
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || !alive(pid) {
		_ = os.Remove(d.PIDFile)
		return 0, ErrNotRunning
	}
	return pid, nil
}

// Stop terminates the core and removes the PID file.
func (d *Daemon) Stop() error {
	pid, err := d.PID()
	if err != nil {
		return err
	}
	if err := terminate(pid); err != nil {
		return err
	}
	return os.Remove(d.PIDFile)
}
