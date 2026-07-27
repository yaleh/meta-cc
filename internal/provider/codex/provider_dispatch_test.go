package codex

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/provider/codex/appserver"
)

// filesFixtureProvider builds a Provider whose files (SQLite) backend has
// exactly one real session, "files-1", so tests can distinguish "app_server
// backend answered" from "fell back to files".
func filesFixtureProvider(t *testing.T, mode Mode, appServer *appServerBackend) *Provider {
	t.Helper()
	root := t.TempDir()
	t.Setenv("META_CC_CODEX_ROOT", root)
	loc := locator.NewCodexLocator()

	db, err := sql.Open("sqlite", loc.SQLiteDB())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY, rollout_path TEXT, cwd TEXT, title TEXT,
		model TEXT, model_provider TEXT, tokens_used INTEGER, source TEXT, created_at INTEGER
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at)
		VALUES ('files-1', '/tmp/r.jsonl', '/tmp', 'hi', 'gpt-5', 'openai', 1, 'cli', 1700000000)`); err != nil {
		t.Fatal(err)
	}

	return newProvider(loc, mode, appServer)
}

func TestProviderFilesModeNeverTouchesAppServer(t *testing.T) {
	called := false
	fake := &appServerBackend{connect: func(context.Context) (threadSource, io.Closer, error) {
		called = true
		return nil, nil, errors.New("should not be called")
	}}
	p := filesFixtureProvider(t, ModeFiles, fake)

	sessions, err := p.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "files-1" {
		t.Fatalf("unexpected sessions: %#v", sessions)
	}
	if called {
		t.Fatalf("ModeFiles must never invoke the app-server connector")
	}
	if p.Backend() != "files" {
		t.Fatalf("Backend() = %q, want files", p.Backend())
	}
}

func TestProviderAppServerModeFailsClearlyWithoutFallback(t *testing.T) {
	fake := &appServerBackend{connect: connectFake(nil, nil, errors.New("app-server unavailable"))}
	p := filesFixtureProvider(t, ModeAppServer, fake)

	_, err := p.ListSessions(context.Background())
	if err == nil {
		t.Fatalf("expected ModeAppServer to fail clearly, got nil error")
	}
	if p.Backend() == "files" {
		t.Fatalf("ModeAppServer must not fall back to files")
	}
}

func TestProviderAppServerModeSucceeds(t *testing.T) {
	src := &fakeThreadSource{pages: map[string]appserver.ThreadListResult{
		"active:": {Data: []appserver.Thread{{ID: "as-1", CreatedAt: 1}}},
	}}
	fake := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}
	p := filesFixtureProvider(t, ModeAppServer, fake)

	sessions, err := p.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "as-1" {
		t.Fatalf("expected app-server session, got %#v", sessions)
	}
	if p.Backend() != "app_server" {
		t.Fatalf("Backend() = %q, want app_server", p.Backend())
	}
}

func TestProviderAutoModeFallsBackAndReportsProvenance(t *testing.T) {
	fake := &appServerBackend{connect: connectFake(nil, nil, errors.New("no app-server"))}
	p := filesFixtureProvider(t, ModeAuto, fake)

	sessions, err := p.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "files-1" {
		t.Fatalf("expected fallback to files fixture, got %#v", sessions)
	}
	if p.Backend() != "files" {
		t.Fatalf("Backend() = %q, want files after fallback", p.Backend())
	}
	if len(p.Warnings()) == 0 {
		t.Fatalf("expected a warning recorded for the app-server failure")
	}
}

func TestProviderAutoModePrefersAppServerWhenHealthy(t *testing.T) {
	src := &fakeThreadSource{pages: map[string]appserver.ThreadListResult{
		"active:": {Data: []appserver.Thread{{ID: "as-1", CreatedAt: 1}}},
	}}
	fake := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}
	p := filesFixtureProvider(t, ModeAuto, fake)

	sessions, err := p.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "as-1" {
		t.Fatalf("expected app-server session preferred over files, got %#v", sessions)
	}
	if p.Backend() != "app_server" {
		t.Fatalf("Backend() = %q, want app_server", p.Backend())
	}
}

func TestProviderDetectAppServerReportsAbsentBinary(t *testing.T) {
	p := filesFixtureProvider(t, ModeFiles, newAppServerBackend(locator.NewCodexLocator()))
	t.Setenv(appServerBinEnvVar, "meta-cc-definitely-not-a-real-binary")
	result := p.DetectAppServer(context.Background())
	if result.Found {
		t.Fatalf("expected Found=false for a nonexistent binary, got %#v", result)
	}
}

// TestProviderListSessionsPageCapabilityPresent is the DIR-032 "capability
// present" path: when app-server is reachable (ModeAppServer here), a
// single ListSessionsPage call issues exactly one thread/list request and
// returns the server's real cursor for continuation, rather than loading
// everything up front.
func TestProviderListSessionsPageCapabilityPresent(t *testing.T) {
	cur1 := "cursor-1"
	src := &fakeThreadSource{pages: map[string]appserver.ThreadListResult{
		"active:":        {Data: []appserver.Thread{{ID: "as-1", CreatedAt: 1}}, NextCursor: &cur1},
		"active:" + cur1: {Data: []appserver.Thread{{ID: "as-2", CreatedAt: 2}}},
	}}
	fake := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}
	p := filesFixtureProvider(t, ModeAppServer, fake)

	page, err := p.ListSessionsPage(context.Background(), conversation.SessionFilter{}, "")
	if err != nil {
		t.Fatalf("ListSessionsPage: %v", err)
	}
	if page.Backend != "app_server" {
		t.Fatalf("expected Backend=app_server, got %q", page.Backend)
	}
	if len(page.Sessions) != 1 || page.Sessions[0].ID != "as-1" {
		t.Fatalf("unexpected page 1 sessions: %#v", page.Sessions)
	}
	if page.NextCursor != cur1 {
		t.Fatalf("expected NextCursor=%q, got %q", cur1, page.NextCursor)
	}
	if len(src.listCall) != 1 {
		t.Fatalf("expected exactly one thread/list call, got %d", len(src.listCall))
	}

	page, err = p.ListSessionsPage(context.Background(), conversation.SessionFilter{}, page.NextCursor)
	if err != nil {
		t.Fatalf("ListSessionsPage page 2: %v", err)
	}
	if len(page.Sessions) != 1 || page.Sessions[0].ID != "as-2" || page.NextCursor != "" {
		t.Fatalf("unexpected page 2: %#v", page)
	}
}

// TestProviderListSessionsPageCapabilityAbsentFallsBackSafely is the
// DIR-032 "capability absent" path: ModeFiles has no app-server pagination
// surface at all, so ListSessionsPage must fail safely to the existing
// non-paginated behavior — the full filtered result as a single page with
// an empty NextCursor — rather than erroring or returning a broken cursor.
func TestProviderListSessionsPageCapabilityAbsentFallsBackSafely(t *testing.T) {
	fake := &appServerBackend{connect: func(context.Context) (threadSource, io.Closer, error) {
		t.Fatalf("ModeFiles must never invoke the app-server connector")
		return nil, nil, nil
	}}
	p := filesFixtureProvider(t, ModeFiles, fake)

	page, err := p.ListSessionsPage(context.Background(), conversation.SessionFilter{}, "")
	if err != nil {
		t.Fatalf("ListSessionsPage: %v", err)
	}
	if page.Backend != "files" {
		t.Fatalf("expected Backend=files, got %q", page.Backend)
	}
	if len(page.Sessions) != 1 || page.Sessions[0].ID != "files-1" {
		t.Fatalf("unexpected sessions: %#v", page.Sessions)
	}
	if page.NextCursor != "" {
		t.Fatalf("expected empty NextCursor (no pagination capability), got %q", page.NextCursor)
	}
}

func TestProviderAutoModeCircuitBreakerOpensAfterRepeatedFailures(t *testing.T) {
	attempts := 0
	fake := &appServerBackend{connect: func(context.Context) (threadSource, io.Closer, error) {
		attempts++
		return nil, nil, errors.New("down")
	}}
	p := filesFixtureProvider(t, ModeAuto, fake)

	for i := 0; i < circuitBreakerThreshold; i++ {
		if _, err := p.ListSessions(context.Background()); err != nil {
			t.Fatalf("call %d: unexpected error (should fall back to files): %v", i, err)
		}
	}
	if attempts != circuitBreakerThreshold {
		t.Fatalf("expected %d app-server attempts, got %d", circuitBreakerThreshold, attempts)
	}

	// One more call: circuit should now be open, so the connector is not
	// invoked again (files fallback still succeeds).
	if _, err := p.ListSessions(context.Background()); err != nil {
		t.Fatalf("unexpected error while circuit open: %v", err)
	}
	if attempts != circuitBreakerThreshold {
		t.Fatalf("expected circuit breaker to skip app-server attempt, attempts = %d", attempts)
	}

	// Simulate cooldown expiry: the next call should retry app-server.
	p.mu.Lock()
	p.circuitOpenUntil = time.Now().Add(-time.Second)
	p.mu.Unlock()
	if _, err := p.ListSessions(context.Background()); err != nil {
		t.Fatalf("unexpected error after cooldown: %v", err)
	}
	if attempts != circuitBreakerThreshold+1 {
		t.Fatalf("expected app-server retry after cooldown, attempts = %d", attempts)
	}
}
