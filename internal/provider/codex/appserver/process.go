package appserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// StderrCap bounds how much stderr output Process retains for diagnostics.
// Stderr is read on its own pipe, entirely separate from the stdout
// protocol stream, so app-server log noise can never corrupt an in-flight
// JSON-RPC frame.
const StderrCap = 64 * 1024

// shutdownGrace is how long Close waits for a graceful exit (SIGTERM to the
// process group) before escalating to SIGKILL.
const shutdownGrace = 2 * time.Second

// Process spawns and owns a `codex app-server` child, wiring its stdio into
// a Client. Callers MUST call Close exactly once (typically via defer right
// after a successful StartProcess) — Close guarantees the child and any
// process-group descendants (e.g. a sandbox helper) are terminated and
// reaped, so no zombie or orphaned process survives shutdown or request
// cancellation.
type Process struct {
	cmd    *exec.Cmd
	Client *Client

	stderrMu  sync.Mutex
	stderrBuf bytes.Buffer

	closeOnce sync.Once
	closeErr  error
}

// StartProcess spawns `command app-server [args...]` with the given
// environment (nil means "inherit the current process's environment", same
// as exec.Cmd's default) and returns a Process whose Client is ready for
// Initialize. ctx bounds the process's entire lifetime: cancellation (e.g. a
// caller-imposed startup timeout) tears the process down the same way Close
// does.
func StartProcess(ctx context.Context, command string, args []string, env []string) (*Process, error) {
	cmd := exec.CommandContext(ctx, command, args...) //nolint:gosec // command/args are caller-controlled configuration, not untrusted input
	if env != nil {
		cmd.Env = env
	}
	cmd.SysProcAttr = processGroupAttr()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("codex app-server: stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("codex app-server: stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("codex app-server: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("codex app-server: start %s app-server: %w", command, err)
	}

	p := &Process{cmd: cmd}
	go p.drainStderr(stderr)
	p.Client = NewClient(stdin, stdout)
	return p, nil
}

func (p *Process) drainStderr(r io.Reader) {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			p.stderrMu.Lock()
			if p.stderrBuf.Len() < StderrCap {
				p.stderrBuf.Write(buf[:n])
			}
			p.stderrMu.Unlock()
		}
		if err != nil {
			return
		}
	}
}

// Stderr returns a snapshot of captured stderr output (capped at
// StderrCap), useful for diagnosing a startup or request failure.
func (p *Process) Stderr() string {
	p.stderrMu.Lock()
	defer p.stderrMu.Unlock()
	return p.stderrBuf.String()
}

// Close terminates the child process group and waits for it to exit,
// escalating from SIGTERM to SIGKILL after shutdownGrace, then closes the
// protocol Client. Safe to call more than once; only the first call does
// work and its result is returned to every caller.
func (p *Process) Close() error {
	p.closeOnce.Do(func() {
		_ = p.Client.Close()
		p.closeErr = p.terminate()
	})
	return p.closeErr
}

func (p *Process) terminate() error {
	if p.cmd.Process == nil {
		return nil
	}
	killGroup(p.cmd, false)

	done := make(chan error, 1)
	go func() { done <- p.cmd.Wait() }()

	select {
	case err := <-done:
		return err
	case <-time.After(shutdownGrace):
		killGroup(p.cmd, true)
		return <-done
	}
}
