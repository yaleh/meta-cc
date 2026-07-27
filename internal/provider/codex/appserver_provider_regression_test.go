package codex

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/provider/codex/appserver"
)

// TestConnectProcessKeepsChildAliveAfterConnectorReturns is a regression
// test for a bug found by adversarial audit: connectProcess used to derive
// its child process's exec.CommandContext lifetime from a
// context.WithTimeout whose cancel func was deferred *inside the connector
// closure itself*. exec.CommandContext's internal watcher kills the child
// the instant ctx.Done() fires, regardless of whether Wait() has been
// called — so that deferred cancel() fired the moment connect() returned,
// killing the freshly-started `codex app-server` process before any caller
// (listSessions/getSession) ever got to issue a request against it.
//
// This drives connectProcess against a real, separate OS process (this
// same test binary, re-exec'd as a fake app-server via
// runFakeAppServerHelperProcess in appserver_helper_process_test.go) —
// not the hand-rolled fakeThreadSource used by the rest of this file's
// tests — because the bug lives entirely in exec.CommandContext/context
// lifetime wiring, which a fake threadSource with no subprocess involved
// can never exercise.
func TestConnectProcessKeepsChildAliveAfterConnectorReturns(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process-group signal handling assumed unix in this test")
	}

	selfExe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}

	t.Setenv(appServerBinEnvVar, selfExe)
	t.Setenv(helperProcessEnvVar, "1")

	connect := connectProcess(nil)

	ctx := context.Background()
	src, closer, err := connect(ctx)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer closer.Close()

	// Give any (buggy) premature-cancellation watcher goroutine every
	// opportunity to have already killed the child by now — under the fix,
	// this delay is irrelevant since nothing should kill the process
	// regardless of how long we wait here.
	time.Sleep(100 * time.Millisecond)

	// This is the exact scenario that failed with a
	// "write |1: broken pipe"-shaped error against the buggy code: a call
	// issued strictly after connect() has already returned.
	if _, err := src.ThreadList(ctx, appserver.ThreadListParams{
		SourceKinds:    knownSourceKinds,
		ModelProviders: []string{},
	}); err != nil {
		t.Fatalf("ThreadList after connect() returned: %v (child process was likely killed by a premature context cancellation)", err)
	}
}
