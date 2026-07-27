//go:build !windows

package appserver

import (
	"os/exec"
	"syscall"
)

// processGroupAttr puts the child in its own process group so Close can
// signal the whole group (the app-server binary and any sandbox helper it
// spawns) instead of only the immediate child.
func processGroupAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

// killGroup signals the child's process group. force selects SIGKILL over
// SIGTERM for the escalation path. Errors are intentionally ignored: the
// process may have already exited, which is not a shutdown failure.
func killGroup(cmd *exec.Cmd, force bool) {
	if cmd.Process == nil {
		return
	}
	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}
	_ = syscall.Kill(-cmd.Process.Pid, sig)
}
