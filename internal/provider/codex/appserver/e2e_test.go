package appserver

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestRealAppServerE2E is the version-gated, isolated E2E smoke test DIR-029
// calls for: it drives an actual installed `codex app-server` child process
// through the initialize handshake and one thread/list call.
//
// It is gated, not required, for two reasons:
//  1. `codex` may not be installed at all in a given build/CI environment.
//  2. Even when installed, this package only claims support for
//     MinSupportedVersion+ (empirically verified against 0.145.0 — see
//     docs/reference/codex-app-server.md); an older or incompatible CLI
//     should not fail `make commit`, it should be skipped.
//
// CODEX_HOME is pinned to a fresh t.TempDir() for the whole test, so this
// never reads or writes the developer's real ~/.codex.
func TestRealAppServerE2E(t *testing.T) {
	if testing.Short() {
		// `make commit`/`make test` run `go test -short`, which must never
		// depend on a real external `codex` process's behavior (per the
		// Contract) — this test only runs opt-in via a non-short `go test`
		// invocation, regardless of what happens to be on PATH.
		t.Skip("skipping real app-server E2E test in -short mode")
	}

	ctx := context.Background()

	detect := DetectCLIVersion(ctx, "codex")
	if !detect.Found {
		t.Skip("codex CLI not found on PATH; skipping real app-server E2E test")
	}
	if detect.Err != nil {
		t.Skipf("codex CLI found but --version could not be parsed (%v); skipping", detect.Err)
	}
	if !detect.Supported {
		t.Skipf("codex CLI %s is older than MinSupportedVersion %s; skipping", detect.Version, MinSupportedVersion)
	}

	codexHome := t.TempDir()
	env := envWithCodexHome(codexHome)

	startCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	proc, err := StartProcess(startCtx, "codex", []string{"app-server"}, env)
	if err != nil {
		t.Fatalf("StartProcess: %v", err)
	}
	defer func() {
		if err := proc.Close(); err != nil {
			t.Logf("Process.Close: %v (stderr: %s)", err, proc.Stderr())
		}
	}()

	initCtx, initCancel := context.WithTimeout(ctx, 10*time.Second)
	defer initCancel()
	initResult, err := proc.Client.Initialize(initCtx, ClientInfo{Name: "meta-cc-e2e-test", Version: "dir-029"})
	if err != nil {
		t.Fatalf("Initialize: %v (stderr: %s)", err, proc.Stderr())
	}
	if initResult.CodexHome == "" {
		t.Fatalf("Initialize: expected non-empty CodexHome, got %#v", initResult)
	}

	listCtx, listCancel := context.WithTimeout(ctx, 10*time.Second)
	defer listCancel()
	listResult, err := proc.Client.ThreadList(listCtx, ThreadListParams{
		SourceKinds:    []string{"cli", "vscode", "exec", "appServer", "subAgent", "subAgentReview", "subAgentCompact", "subAgentThreadSpawn", "subAgentOther", "unknown"},
		ModelProviders: []string{},
	})
	if err != nil {
		t.Fatalf("ThreadList: %v (stderr: %s)", err, proc.Stderr())
	}
	// A brand-new, isolated CODEX_HOME has no threads yet.
	if len(listResult.Data) != 0 {
		t.Fatalf("expected no threads in a fresh isolated CODEX_HOME, got %d: %#v", len(listResult.Data), listResult.Data)
	}
}

// envWithCodexHome returns the current process's environment with
// CODEX_HOME overridden to home. exec.Cmd resolves duplicate keys by using
// the last occurrence, so a simple append is sufficient (see os/exec's
// Cmd.Env doc); no need to filter out any pre-existing CODEX_HOME entry.
func envWithCodexHome(home string) []string {
	return append(os.Environ(), "CODEX_HOME="+home)
}
