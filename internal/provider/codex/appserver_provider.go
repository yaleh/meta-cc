package codex

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/provider/codex/appserver"
)

// threadSource is the minimal app-server surface appServerBackend needs.
// *appserver.Client satisfies it directly (see connectProcess); tests
// inject a hand-rolled fake with no subprocess or pipes involved, keeping
// this package's own tests fast and hermetic on top of the appserver
// package's separate protocol-level (pipe-based) test suite.
type threadSource interface {
	ThreadList(ctx context.Context, params appserver.ThreadListParams) (appserver.ThreadListResult, error)
	ThreadRead(ctx context.Context, params appserver.ThreadReadParams) (appserver.ThreadReadResult, error)
}

// connectFunc establishes a fresh, initialized app-server connection and
// returns it alongside an io.Closer that MUST be called exactly once by the
// caller to guarantee no child process (or other resource) outlives the
// call — production wiring (connectProcess) closes a real
// *appserver.Process (terminating the child and reaping it, see
// appserver.Process.Close); tests use a no-op or call-tracking fake.
type connectFunc func(ctx context.Context) (threadSource, io.Closer, error)

// appServerBinEnvVar overrides the executable used to spawn `codex
// app-server`, for environments where the Codex CLI isn't on PATH as
// "codex".
const appServerBinEnvVar = "META_CC_CODEX_APP_SERVER_BIN"

func appServerBinary() string {
	if v := os.Getenv(appServerBinEnvVar); v != "" {
		return v
	}
	return "codex"
}

// startupTimeout bounds spawning the app-server process and completing the
// initialize handshake. callTimeout bounds each individual thread/list or
// thread/read exchange (across however many paginated requests one
// operation needs). Both are conservative but finite, per the Contract's
// "bounded timeouts" requirement.
const (
	startupTimeout = 10 * time.Second
	callTimeout    = 20 * time.Second
)

// connectProcess is the production connectFunc: it spawns a real `codex
// app-server` child pinned to loc's CODEX_HOME (so the external process
// reads the same state meta-cc is configured for, even when
// META_CC_CODEX_ROOT overrides the default ~/.codex) and performs the
// mandatory initialize handshake before handing back the client.
func connectProcess(loc *locator.CodexLocator) connectFunc {
	return func(ctx context.Context) (threadSource, io.Closer, error) {
		// procCtx bounds the spawned child's entire lifetime via
		// exec.CommandContext (see appserver.StartProcess): canceling it
		// kills the child immediately, even long after Start() has
		// returned successfully and regardless of whether Wait() has run
		// yet. It therefore MUST stay alive past this closure returning —
		// listSessions/getSession only issue thread/list/thread/read
		// *after* connect() returns — so it is only ever canceled
		// explicitly, via the returned closer's Close() (see
		// processCloser below), never by a defer in this function. (A
		// previous version deferred cancel() on a startupTimeout-bounded
		// context right here, which fired the instant connect() returned
		// and killed the freshly-started process before any caller could
		// use it — see appserver_provider_regression_test.go.)
		//
		// It's still derived from the caller's ctx (not
		// context.Background()) so an outer cancellation (e.g. the whole
		// meta-cc invocation being aborted) still tears the child down
		// even if the caller never reaches its own closer.Close().
		procCtx, procCancel := context.WithCancel(ctx)

		// initCtx separately bounds only the initialize handshake
		// (startupTimeout) and is safe to let expire/cancel as soon as
		// Initialize returns: unlike procCtx, its cancellation is never
		// wired to exec.CommandContext, so it cannot reach the child
		// process.
		initCtx, initCancel := context.WithTimeout(ctx, startupTimeout)
		defer initCancel()

		env := os.Environ()
		if loc != nil {
			env = append(env, "CODEX_HOME="+loc.Root())
		}
		proc, err := appserver.StartProcess(procCtx, appServerBinary(), []string{"app-server"}, env)
		if err != nil {
			procCancel()
			return nil, nil, fmt.Errorf("start codex app-server: %w", err)
		}
		if _, err := proc.Client.Initialize(initCtx, appserver.ClientInfo{Name: "meta-cc", Version: "dir-029"}); err != nil {
			stderr := proc.Stderr()
			_ = proc.Close()
			procCancel()
			if stderr != "" {
				return nil, nil, fmt.Errorf("codex app-server initialize: %w (stderr: %s)", err, stderr)
			}
			return nil, nil, fmt.Errorf("codex app-server initialize: %w", err)
		}
		return proc.Client, &processCloser{proc: proc, cancel: procCancel}, nil
	}
}

// processCloser bundles a spawned appserver.Process with the cancel func
// for the context that bounds its exec.CommandContext lifetime (procCtx in
// connectProcess), so a caller's single Close() call both terminates the
// process (SIGTERM→SIGKILL, via Process.Close) and releases procCtx.
// Without this, procCtx would only ever be released when its parent ctx
// (typically the whole request's context) is itself done, rather than
// promptly when the connection is actually closed.
type processCloser struct {
	proc   *appserver.Process
	cancel context.CancelFunc
}

func (c *processCloser) Close() error {
	err := c.proc.Close()
	c.cancel()
	return err
}

// KnownSourceKinds is the full ThreadSourceKind enum (confirmed against
// `codex app-server generate-json-schema` for Codex CLI 0.145.0 — see
// docs/reference/codex-app-server.md), requested explicitly on every
// thread/list call so listing never silently narrows to the server's
// unqualified default (interactive sources only).
var KnownSourceKinds = []string{
	"cli", "vscode", "exec", "appServer",
	"subAgent", "subAgentReview", "subAgentCompact", "subAgentThreadSpawn", "subAgentOther",
	"unknown",
}

// appServerBackend lists/reads Codex threads via a connectFunc. It only
// ever issues thread/list and thread/read — never thread/start, resume,
// fork, archive, delete, or any other mutating method.
type appServerBackend struct {
	connect connectFunc

	mu       sync.Mutex
	warnings []string
}

func newAppServerBackend(loc *locator.CodexLocator) *appServerBackend {
	return &appServerBackend{connect: connectProcess(loc)}
}

// recordWarning appends a bounded diagnostic (e.g. a per-page thread/list
// failure — see listAll) without aborting the caller's in-flight listing.
func (b *appServerBackend) recordWarning(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.warnings = append(b.warnings, msg)
}

// drainWarnings returns and clears accumulated warnings, so a caller (see
// Provider.ListSessions/ListSessionsFiltered) can fold them into its own
// Warnings() exactly once per call rather than re-reporting stale entries.
func (b *appServerBackend) drainWarnings() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	w := b.warnings
	b.warnings = nil
	return w
}

// listSessions requests every session the app-server can report. Two
// concerns make this more than a single thread/list call:
//
//   - sourceKinds: an omitted/empty filter defaults to interactive sources
//     only, so KnownSourceKinds is always passed explicitly.
//   - archived: the server treats this as an exclusive boolean filter
//     (omitted/false returns only non-archived threads; true returns only
//     archived ones — there is no "both" value), so listSessions issues two
//     independently paginated passes and merges them, rather than silently
//     omitting archived threads from the default listing.
func (b *appServerBackend) listSessions(ctx context.Context) ([]conversation.Session, error) {
	return b.listSessionsFiltered(ctx, conversation.SessionFilter{})
}

// listSessionsFiltered is listSessions extended with DIR-030 metadata
// filters pushed directly into the thread/list request: cwd, archived,
// modelProvider, and sourceKinds are all fields ThreadListParams already
// supports server-side (see docs/reference/codex-app-server.md), so
// sending them narrows what the app-server itself returns — no thread/read
// (and therefore no turn loading) happens here at all. Filter dimensions
// the server doesn't support (title, parent thread, updated/created time
// range) are NOT applied here; callers apply conversation.ApplyFilter on
// the result for those, which is still metadata-only (no deep parse).
func (b *appServerBackend) listSessionsFiltered(ctx context.Context, filter conversation.SessionFilter) ([]conversation.Session, error) {
	src, closer, err := b.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	base := buildThreadListParams(filter)

	archivedPasses := []bool{false, true}
	if filter.Archived != nil {
		archivedPasses = []bool{*filter.Archived}
	}

	var sessions []conversation.Session
	for _, archived := range archivedPasses {
		// DIR-032: a failure on the very FIRST page of a pass (zero threads
		// fetched) is still treated as a real pass failure and aborts the
		// whole listing, preserving pre-DIR-032 semantics (ModeAuto's
		// ordinary fallback/circuit-breaker behavior — see dispatch — still
		// applies when app-server appears to not be working at all). A
		// failure AFTER at least one page already succeeded is different:
		// listAll returns the threads fetched so far plus a recorded
		// warning (drained by the caller via drainWarnings) instead of
		// discarding them, so a genuine mid-pagination blip degrades
		// gracefully rather than causing whole-corpus loss.
		threads, err := b.listAll(ctx, src, base, archived)
		if err != nil {
			return nil, err
		}
		for _, t := range threads {
			session := appserver.MapThread(t)
			session.Archived = archived
			if archived {
				session.Status = "archived"
			} else {
				session.Status = "active"
			}
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

// listAll follows thread/list pagination (cursor/nextCursor) to completion
// for one archived-state pass. See listSessionsFiltered's comment for the
// first-page-vs-later-page failure distinction.
func (b *appServerBackend) listAll(ctx context.Context, src threadSource, base appserver.ThreadListParams, archived bool) ([]appserver.Thread, error) {
	var threads []appserver.Thread
	cursor := ""
	for {
		params := base
		params.Cursor = cursor
		params.Archived = &archived
		result, err := src.ThreadList(ctx, params)
		if err != nil {
			if len(threads) == 0 {
				return nil, fmt.Errorf("thread/list: %w", err)
			}
			if ctx.Err() != nil {
				// The caller's own shutdown, not a page-specific problem —
				// stop silently rather than recording a spurious warning.
				return threads, nil
			}
			b.recordWarning(fmt.Sprintf(
				"thread/list page failed (archived=%v, cursor=%q): %v; returning %d thread(s) fetched so far; resume from cursor=%q",
				archived, cursor, err, len(threads), cursor))
			return threads, nil
		}
		threads = append(threads, result.Data...)
		if result.NextCursor == nil || *result.NextCursor == "" {
			return threads, nil
		}
		cursor = *result.NextCursor
	}
}

// buildThreadListParams derives the base ThreadListParams shared by
// listSessionsFiltered (which pages through every result) and
// listSessionsPage (DIR-032 — which fetches exactly one page). Centralized
// so both honor the same "explicit sourceKinds/modelProviders, never the
// server's narrower unqualified default" rule (see docs/reference/
// codex-app-server.md's "avoiding the default-filter pitfall").
func buildThreadListParams(filter conversation.SessionFilter) appserver.ThreadListParams {
	base := appserver.ThreadListParams{
		SourceKinds:    KnownSourceKinds,
		ModelProviders: []string{},
	}
	if len(filter.SourceKinds) > 0 {
		base.SourceKinds = filter.SourceKinds
	}
	if filter.CWD != "" {
		base.CWD = []string{filter.CWD}
	}
	if filter.ModelProvider != "" {
		base.ModelProviders = []string{filter.ModelProvider}
	}
	return base
}

// listSessionsPage fetches exactly ONE thread/list page (DIR-032's
// cursor-based continuation API — see codex.Provider.ListSessionsPage): the
// server's "archived" filter is exclusive (see listSessionsFiltered's
// comment), so a caller that leaves filter.Archived unset gets the
// non-archived pass; pass filter.Archived explicitly to page through
// archived threads instead. Unlike listSessionsFiltered/listAll (which
// tolerate a mid-pagination page failure by returning partial results plus
// a warning), a single-page fetch that fails simply returns the error: the
// caller already holds the cursor it passed in and can retry it directly,
// so there is no "partial progress to preserve" within this one call.
func (b *appServerBackend) listSessionsPage(ctx context.Context, filter conversation.SessionFilter, cursor string) (sessions []conversation.Session, nextCursor string, err error) {
	src, closer, err := b.connect(ctx)
	if err != nil {
		return nil, "", err
	}
	defer closer.Close()

	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	archived := false
	if filter.Archived != nil {
		archived = *filter.Archived
	}

	params := buildThreadListParams(filter)
	params.Cursor = cursor
	params.Archived = &archived

	result, err := src.ThreadList(ctx, params)
	if err != nil {
		return nil, "", fmt.Errorf("thread/list: %w", err)
	}
	for _, t := range result.Data {
		session := appserver.MapThread(t)
		session.Archived = archived
		if archived {
			session.Status = "archived"
		} else {
			session.Status = "active"
		}
		sessions = append(sessions, session)
	}
	if result.NextCursor != nil {
		nextCursor = *result.NextCursor
	}
	return sessions, nextCursor, nil
}

func (b *appServerBackend) getSession(ctx context.Context, sessionID string) (conversation.Session, error) {
	src, closer, err := b.connect(ctx)
	if err != nil {
		return conversation.Session{}, err
	}
	defer closer.Close()

	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	result, err := src.ThreadRead(ctx, appserver.ThreadReadParams{ThreadID: sessionID, IncludeTurns: true})
	if err != nil {
		return conversation.Session{}, fmt.Errorf("thread/read: %w", err)
	}
	return appserver.MapThread(result.Thread), nil
}

func (b *appServerBackend) loadTurns(ctx context.Context, sessionID string) ([]conversation.Turn, error) {
	session, err := b.getSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return session.Turns, nil
}

// available reports whether an app-server backend is worth attempting:
// a `codex` binary is on PATH (or META_CC_CODEX_APP_SERVER_BIN points at
// one) and its --version satisfies appserver.MinSupportedVersion. This is
// a cheap, bounded pre-check (see appserver.DetectCLIVersion) used by
// ModeAuto/ModeAppServer diagnostics; it does not by itself guarantee a
// thread/list call will succeed (the process could still fail to start or
// the handshake could still fail), which is why callers must still handle
// listSessions/getSession errors regardless of this check.
func (b *appServerBackend) available(ctx context.Context) appserver.DetectResult {
	return appserver.DetectCLIVersion(ctx, appServerBinary())
}
