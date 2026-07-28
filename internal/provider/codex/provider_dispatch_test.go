package codex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
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

// buildManyPageThreadSource builds a fake threadSource with totalPages
// non-archived pages of one thread each, chained by synthetic cursors, so
// tests can prove a bounded fetch stops well short of the full corpus.
func buildManyPageThreadSource(totalPages int) *fakeThreadSource {
	pages := map[string]appserver.ThreadListResult{}
	cursor := ""
	for i := 0; i < totalPages; i++ {
		result := appserver.ThreadListResult{
			Data: []appserver.Thread{{ID: fmt.Sprintf("as-%d", i), CreatedAt: int64(i)}},
		}
		if i < totalPages-1 {
			next := fmt.Sprintf("cursor-%d", i+1)
			result.NextCursor = &next
		}
		pages["active:"+cursor] = result
		cursor = fmt.Sprintf("cursor-%d", i+1)
	}
	return &fakeThreadSource{pages: pages}
}

// TestProviderFetchSessionsBoundedStopsEarly is DIR-034's "wire
// ListSessionsPage to a real caller" proof: FetchSessionsBounded must
// satisfy a small limit without paging through the whole corpus (25 pages
// available here) — this is exactly the "hundreds of sessions" scaling
// concern query_sessions(provider="codex", limit=N) needed fixed.
func TestProviderFetchSessionsBoundedStopsEarly(t *testing.T) {
	const totalPages = 25

	src := buildManyPageThreadSource(totalPages)
	fake := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}
	p := filesFixtureProvider(t, ModeAppServer, fake)

	sessions, err := p.FetchSessionsBounded(context.Background(), conversation.SessionFilter{}, 1)
	if err != nil {
		t.Fatalf("FetchSessionsBounded: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected exactly 1 session, got %d", len(sessions))
	}
	if len(src.listCall) != 1 {
		t.Fatalf("expected exactly 1 thread/list call bounded by limit=1, got %d (of %d available pages)", len(src.listCall), totalPages)
	}

	// A larger limit spanning multiple pages should issue exactly that many
	// calls: bounded by the requested limit/page count, not by total corpus
	// size (still far short of totalPages).
	src2 := buildManyPageThreadSource(totalPages)
	fake2 := &appServerBackend{connect: connectFake(src2, &noopCloser{}, nil)}
	p2 := filesFixtureProvider(t, ModeAppServer, fake2)

	sessions2, err := p2.FetchSessionsBounded(context.Background(), conversation.SessionFilter{}, 3)
	if err != nil {
		t.Fatalf("FetchSessionsBounded (limit=3): %v", err)
	}
	if len(sessions2) != 3 {
		t.Fatalf("expected exactly 3 sessions, got %d", len(sessions2))
	}
	if len(src2.listCall) != 3 {
		t.Fatalf("expected exactly 3 thread/list calls, got %d (of %d available pages)", len(src2.listCall), totalPages)
	}
}

// TestProviderFetchSessionsBoundedFallsBackWhenLimitZero proves
// FetchSessionsBounded delegates to the unbounded ListSessionsFiltered path
// (paging to completion) when limit<=0 — "there's nothing to bound
// against" — so it never silently truncates an unbounded caller's results.
func TestProviderFetchSessionsBoundedFallsBackWhenLimitZero(t *testing.T) {
	const totalPages = 4
	src := buildManyPageThreadSource(totalPages)
	fake := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}
	p := filesFixtureProvider(t, ModeAppServer, fake)

	sessions, err := p.FetchSessionsBounded(context.Background(), conversation.SessionFilter{}, 0)
	if err != nil {
		t.Fatalf("FetchSessionsBounded: %v", err)
	}
	if len(sessions) != totalPages {
		t.Fatalf("expected all %d sessions from the full crawl, got %d", totalPages, len(sessions))
	}
}

// TestProviderFetchSessionsBoundedToleratesPageFailureMidPagination is the
// DIR-039 regression proof: page 1 of a 3-page fetch succeeds, page 2
// fails, and FetchSessionsBounded(ctx, filter, limit=3) must still return
// the 1 session collected from page 1 (not nil/empty), plus a non-nil
// warning (surfaced via Provider.Warnings(), mirroring how
// appServerListSessions/ListSessionsFiltered already fold app-server
// warnings in) describing the failure and naming the cursor to resume
// from — instead of discarding everything already collected, which is
// what the pre-fix code did (return nil, err).
func TestProviderFetchSessionsBoundedToleratesPageFailureMidPagination(t *testing.T) {
	cur1 := "cursor-1"
	cur2 := "cursor-2"
	failCursor := cur1
	src := &fakeThreadSource{
		pages: map[string]appserver.ThreadListResult{
			"active:":        {Data: []appserver.Thread{{ID: "t1", CreatedAt: 1}}, NextCursor: &cur1},
			"active:" + cur1: {Data: []appserver.Thread{{ID: "t2", CreatedAt: 2}}, NextCursor: &cur2},
			"active:" + cur2: {Data: []appserver.Thread{{ID: "t3", CreatedAt: 3}}},
		},
		failOnCursor: &failCursor,
		failErr:      errors.New("transient network error"),
	}
	fake := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}
	p := filesFixtureProvider(t, ModeAppServer, fake)

	sessions, err := p.FetchSessionsBounded(context.Background(), conversation.SessionFilter{}, 3)
	if err != nil {
		t.Fatalf("FetchSessionsBounded should tolerate a mid-pagination page failure (partial progress preserved), got error: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "t1" {
		t.Fatalf("expected only page 1's session (t1) to survive the page-2 failure, got %#v (must NOT be empty)", sessions)
	}

	warnings := p.Warnings()
	if len(warnings) == 0 {
		t.Fatalf("expected a non-empty warning describing the page failure, got none")
	}
	found := false
	for _, w := range warnings {
		if strings.Contains(w, cur1) {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a warning naming the resumable cursor %q, got %v", cur1, warnings)
	}
}

// TestProviderFetchSessionsBoundedFirstPageFailureIsStillFatal proves the
// DIR-039 fix preserves fail-fast behavior when the very FIRST page fails
// with zero sessions collected: this must still be a hard error (not a
// silent zero-result-plus-warning), matching
// appServerBackend.listSessionsFiltered's own first-page-vs-later-page
// distinction.
func TestProviderFetchSessionsBoundedFirstPageFailureIsStillFatal(t *testing.T) {
	src := &fakeThreadSource{listErr: errors.New("app-server totally unreachable")}
	fake := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}
	p := filesFixtureProvider(t, ModeAppServer, fake)

	sessions, err := p.FetchSessionsBounded(context.Background(), conversation.SessionFilter{}, 3)
	if err == nil {
		t.Fatalf("expected a hard error when the first page fails with zero sessions collected, got sessions=%#v, err=nil", sessions)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected no sessions on a first-page failure, got %#v", sessions)
	}
}
