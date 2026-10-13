//go:build !windows

package daemon

import (
	"os/exec"
	"syscall"
	"time"
)

func detach(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }

func alive(pid int) bool { return syscall.Kill(pid, 0) == nil }

func terminate(pid int) error {
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return err
	}
	for i := 0; i < 50 && alive(pid); i++ {
		time.Sleep(100 * time.Millisecond)
	}
	if alive(pid) {
		return syscall.Kill(pid, syscall.SIGKILL)
	}
	return nil
}
