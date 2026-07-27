//go:build windows

package appserver

import (
	"os/exec"
	"syscall"
)

// processGroupAttr is a no-op on Windows: there is no direct SysProcAttr
// equivalent used here, so Close falls back to killing the immediate
// process only (see killGroup).
func processGroupAttr() *syscall.SysProcAttr { return nil }

// killGroup terminates the immediate child process. Windows process-group
// semantics differ enough from POSIX (job objects, not pgid signals) that a
// full process-tree kill is out of scope for this release; the app-server
// child itself is always terminated and reaped.
func killGroup(cmd *exec.Cmd, _ bool) {
	if cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
}
