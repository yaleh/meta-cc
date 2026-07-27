package codex

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/provider"
	"github.com/yaleh/meta-cc/internal/provider/codex/appserver"
)

var _ provider.Provider = (*Provider)(nil)

// circuitBreakerThreshold/circuitBreakerCooldown: in ModeAuto, after this
// many consecutive app-server failures, skip attempting the app-server
// backend for circuitBreakerCooldown (falling straight through to files)
// rather than paying its full startup/handshake timeout on every call
// against a backend that's clearly not working right now.
const (
	circuitBreakerThreshold = 3
	circuitBreakerCooldown  = 60 * time.Second
)

// Provider is the Codex history provider. Depending on Mode it consults the
// app-server backend, the files (SQLite/rollout) backend, or both (auto,
// preferring app-server with fallback to files). See ResolveMode.
//
// IsAvailable deliberately reflects only the files backend's on-disk state
// (pre-DIR-029 behavior, unchanged): it answers "is there Codex session
// state to read at all", not "which backend would ModeAuto pick" — probing
// app-server availability requires either a real subprocess spawn or a
// bounded external `codex --version` call, and IsAvailable is called
// frequently (e.g. by rawfiles.NewRegistry) in contexts, including this
// package's own pre-existing tests, that must not depend on whether a
// `codex` binary happens to be installed on the host.
type Provider struct {
	locator  *locator.CodexLocator
	maxLines int
	mode     Mode

	appServer *appServerBackend

	mu                  sync.Mutex
	lastBackend         string
	warnings            []string
	consecutiveFailures int
	circuitOpenUntil    time.Time
}

// NewProvider builds a Codex provider using the backend mode from
// META_CC_CODEX_BACKEND (default ModeAuto). NewProvider has no error
// return (an existing, depended-upon signature — see
// internal/provider/rawfiles), so an invalid env value falls back to
// ModeAuto with a warning recorded (visible via Warnings()) rather than
// failing construction.
func NewProvider(loc *locator.CodexLocator) *Provider {
	mode, err := ResolveMode()
	p := newProvider(loc, mode, newAppServerBackend(loc))
	if err != nil {
		p.mode = ModeAuto
		p.warnings = append(p.warnings, err.Error())
	}
	return p
}

// NewProviderWithMode builds a Codex provider with an explicit mode,
// bypassing META_CC_CODEX_BACKEND. Exposed for callers that need
// programmatic control over backend selection.
func NewProviderWithMode(loc *locator.CodexLocator, mode Mode) *Provider {
	return newProvider(loc, mode, newAppServerBackend(loc))
}

func newProvider(loc *locator.CodexLocator, mode Mode, appServer *appServerBackend) *Provider {
	return &Provider{locator: loc, maxLines: 500_000, mode: mode, appServer: appServer}
}

func (p *Provider) ID() conversation.ProviderID {
	return conversation.ProviderCodex
}

func (p *Provider) IsAvailable(context.Context) bool {
	if p.locator == nil {
		return false
	}
	_, err := os.Stat(p.locator.SQLiteDB())
	return err == nil
}

// Backend reports which backend last supplied a ListSessions/GetSession/
// LoadTurns result ("app_server" or "files"), or "" if none has completed
// yet. This is the "reports which backend supplied results" provenance the
// Contract requires for ModeAuto, but is populated in every mode.
func (p *Provider) Backend() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastBackend
}

// Warnings returns structured (human-readable) warnings accumulated across
// this Provider's lifetime — app-server connection/call failures in
// ModeAuto, circuit-breaker state changes, and an invalid
// META_CC_CODEX_BACKEND value falling back to ModeAuto. Never cleared, so
// callers get the full diagnostic history for a Provider instance.
func (p *Provider) Warnings() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.warnings...)
}

// DetectAppServer runs a cheap, bounded `codex --version` check (see
// appserver.DetectCLIVersion) and reports whether an app-server backend is
// installed and satisfies appserver.MinSupportedVersion. This is an
// operational diagnostic — e.g. for a future "why is meta-cc using files
// instead of app_server" CLI command — and is independent of IsAvailable
// (which only reflects files-backend state) and of Mode (it never actually
// spawns the app-server or performs the initialize handshake, so success
// here doesn't guarantee ListSessions/GetSession/LoadTurns will succeed
// against app_server).
func (p *Provider) DetectAppServer(ctx context.Context) appserver.DetectResult {
	return p.appServer.available(ctx)
}

func (p *Provider) ListSessions(ctx context.Context) ([]conversation.Session, error) {
	return dispatch(p, ctx, p.appServer.listSessions, p.filesListSessions)
}

func (p *Provider) GetSession(ctx context.Context, sessionID string) (conversation.Session, error) {
	return dispatch(p, ctx,
		func(ctx context.Context) (conversation.Session, error) { return p.appServer.getSession(ctx, sessionID) },
		func(ctx context.Context) (conversation.Session, error) { return p.filesGetSession(ctx, sessionID) },
	)
}

func (p *Provider) LoadTurns(ctx context.Context, sessionID string) ([]conversation.Turn, error) {
	return dispatch(p, ctx,
		func(ctx context.Context) ([]conversation.Turn, error) { return p.appServer.loadTurns(ctx, sessionID) },
		func(ctx context.Context) ([]conversation.Turn, error) { return p.filesLoadTurns(ctx, sessionID) },
	)
}

func (p *Provider) filesListSessions(ctx context.Context) ([]conversation.Session, error) {
	return listSessionsFromDB(ctx, p.locator.SQLiteDB())
}

func (p *Provider) filesGetSession(ctx context.Context, sessionID string) (conversation.Session, error) {
	return getSessionFromDB(ctx, p.locator.SQLiteDB(), sessionID)
}

func (p *Provider) filesLoadTurns(ctx context.Context, sessionID string) ([]conversation.Turn, error) {
	session, err := p.filesGetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	turns, _, err := loadTurnsFromSession(session, p.maxLines)
	return turns, err
}

// dispatch runs appCall/filesCall according to p.mode:
//
//   - ModeFiles: filesCall only, appCall is never invoked.
//   - ModeAppServer: appCall only; a failure is returned as-is (wrapped),
//     never silently substituted with files.
//   - ModeAuto: appCall first (unless the circuit breaker is open), falling
//     back to filesCall on any appCall error; every fallback and circuit
//     breaker transition is recorded via Warnings().
//
// T is the per-call result type (sessions slice, one session, or turns
// slice), so ListSessions/GetSession/LoadTurns share one dispatch path
// instead of three hand-duplicated copies of this mode/fallback logic.
func dispatch[T any](p *Provider, ctx context.Context, appCall, filesCall func(context.Context) (T, error)) (T, error) {
	switch p.mode {
	case ModeFiles:
		result, err := filesCall(ctx)
		p.recordBackend("files", err)
		return result, err

	case ModeAppServer:
		result, err := appCall(ctx)
		if err != nil {
			p.recordBackend("app_server", err)
			var zero T
			return zero, fmt.Errorf("codex app_server backend: %w", err)
		}
		p.recordBackend("app_server", nil)
		return result, nil

	default: // ModeAuto
		if !p.circuitOpen() {
			result, err := appCall(ctx)
			if err == nil {
				p.recordSuccess()
				return result, nil
			}
			p.recordFailure(err)
		}
		result, err := filesCall(ctx)
		p.recordBackend("files", err)
		return result, err
	}
}

func (p *Provider) circuitOpen() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return time.Now().Before(p.circuitOpenUntil)
}

func (p *Provider) recordFailure(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consecutiveFailures++
	p.warnings = append(p.warnings, fmt.Sprintf(
		"codex app_server backend failed (%d/%d consecutive), falling back to files: %v",
		p.consecutiveFailures, circuitBreakerThreshold, err))
	if p.consecutiveFailures >= circuitBreakerThreshold {
		p.circuitOpenUntil = time.Now().Add(circuitBreakerCooldown)
		p.warnings = append(p.warnings, fmt.Sprintf(
			"codex app_server circuit breaker open for %s after %d consecutive failures",
			circuitBreakerCooldown, p.consecutiveFailures))
	}
}

func (p *Provider) recordSuccess() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consecutiveFailures = 0
	p.circuitOpenUntil = time.Time{}
	p.lastBackend = "app_server"
}

func (p *Provider) recordBackend(backend string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lastBackend = backend
	if err != nil {
		p.warnings = append(p.warnings, fmt.Sprintf("codex %s backend error: %v", backend, err))
	}
}
