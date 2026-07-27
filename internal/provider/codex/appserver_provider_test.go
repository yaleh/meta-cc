package codex

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/provider/codex/appserver"
)

// fakeThreadSource is a hand-rolled threadSource: no subprocess, no pipes,
// just direct Go calls. This exercises appServerBackend's orchestration
// (pagination, archived-filter merging, error wrapping) independently of
// the appserver package's own protocol-level tests.
type fakeThreadSource struct {
	listErr  error
	pages    map[string]appserver.ThreadListResult // keyed by "archived=<bool>,cursor=<cursor>"
	readErr  error
	thread   appserver.Thread
	listCall []appserver.ThreadListParams

	// failOnCursor, when non-nil, makes exactly one ThreadList call whose
	// cursor matches *failOnCursor return failErr instead of the mapped
	// page — used to simulate a page failing partway through a multi-page
	// fetch (see TestAppServerBackendListAllToleratesPageFailureMidPagination).
	failOnCursor *string
	failErr      error
	failed       bool
}

func (f *fakeThreadSource) ThreadList(_ context.Context, p appserver.ThreadListParams) (appserver.ThreadListResult, error) {
	f.listCall = append(f.listCall, p)
	if f.listErr != nil {
		return appserver.ThreadListResult{}, f.listErr
	}
	if f.failOnCursor != nil && !f.failed && p.Cursor == *f.failOnCursor {
		f.failed = true
		return appserver.ThreadListResult{}, f.failErr
	}
	archived := false
	if p.Archived != nil {
		archived = *p.Archived
	}
	key := archivedCursorKey(archived, p.Cursor)
	return f.pages[key], nil
}

func (f *fakeThreadSource) ThreadRead(_ context.Context, p appserver.ThreadReadParams) (appserver.ThreadReadResult, error) {
	if f.readErr != nil {
		return appserver.ThreadReadResult{}, f.readErr
	}
	return appserver.ThreadReadResult{Thread: f.thread}, nil
}

func archivedCursorKey(archived bool, cursor string) string {
	if archived {
		return "archived:" + cursor
	}
	return "active:" + cursor
}

type noopCloser struct{ closed bool }

func (c *noopCloser) Close() error { c.closed = true; return nil }

func connectFake(src threadSource, closer io.Closer, err error) connectFunc {
	return func(context.Context) (threadSource, io.Closer, error) {
		if err != nil {
			return nil, nil, err
		}
		return src, closer, nil
	}
}

func TestAppServerBackendListSessionsMergesArchivedAndPaginates(t *testing.T) {
	cur := "cursor-1"
	src := &fakeThreadSource{pages: map[string]appserver.ThreadListResult{
		"active:":       {Data: []appserver.Thread{{ID: "t1", CreatedAt: 1}}, NextCursor: &cur},
		"active:" + cur: {Data: []appserver.Thread{{ID: "t2", CreatedAt: 2}}},
		"archived:":     {Data: []appserver.Thread{{ID: "t3", CreatedAt: 3}}},
	}}
	closer := &noopCloser{}
	b := &appServerBackend{connect: connectFake(src, closer, nil)}

	sessions, err := b.listSessions(context.Background())
	if err != nil {
		t.Fatalf("listSessions: %v", err)
	}
	if len(sessions) != 3 {
		t.Fatalf("expected 3 merged sessions (paginated active + archived), got %d: %#v", len(sessions), sessions)
	}
	if !closer.closed {
		t.Fatalf("expected connection to be closed after listSessions")
	}

	// Every request must explicitly set sourceKinds/modelProviders per the
	// Contract's "defaults must never silently omit eligible sessions" —
	// an empty/omitted sourceKinds would silently narrow to interactive-only.
	for _, p := range src.listCall {
		if len(p.SourceKinds) == 0 {
			t.Fatalf("thread/list called without explicit sourceKinds: %#v", p)
		}
		if p.ModelProviders == nil {
			t.Fatalf("thread/list called without explicit (non-nil) modelProviders: %#v", p)
		}
	}
}

func TestAppServerBackendListSessionsWrapsError(t *testing.T) {
	src := &fakeThreadSource{listErr: errors.New("boom")}
	b := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}
	if _, err := b.listSessions(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAppServerBackendGetSessionMapsThread(t *testing.T) {
	src := &fakeThreadSource{thread: appserver.Thread{
		ID: "t1", CWD: "/repo", CreatedAt: 100,
		Turns: []appserver.Turn{{ID: "turn1", Status: "completed"}},
	}}
	closer := &noopCloser{}
	b := &appServerBackend{connect: connectFake(src, closer, nil)}

	session, err := b.getSession(context.Background(), "t1")
	if err != nil {
		t.Fatalf("getSession: %v", err)
	}
	if session.ID != "t1" || len(session.Turns) != 1 {
		t.Fatalf("unexpected session: %#v", session)
	}
	if !closer.closed {
		t.Fatalf("expected connection to be closed after getSession")
	}
}

func TestAppServerBackendConnectErrorPropagates(t *testing.T) {
	b := &appServerBackend{connect: connectFake(nil, nil, errors.New("no app-server"))}
	if _, err := b.listSessions(context.Background()); err == nil {
		t.Fatalf("expected connect error to propagate")
	}
}

// TestAppServerBackendListSessionsFilteredPushesParamsDownIntoThreadList is
// the DIR-030 "query planning demonstrates metadata filtering before deep
// loading" proof for the app_server backend: it asserts the CWD/Archived/
// ModelProviders/SourceKinds filter dimensions reach the thread/list
// request itself (real server-side narrowing), not just a post-hoc Go
// filter over an unfiltered response. thread/read (the only call that
// would parse turn/item content) is never invoked at all.
func TestAppServerBackendListSessionsFilteredPushesParamsDownIntoThreadList(t *testing.T) {
	src := &fakeThreadSource{pages: map[string]appserver.ThreadListResult{
		"active:": {Data: []appserver.Thread{{ID: "t1", CWD: "/proj", CreatedAt: 1}}},
	}}
	b := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}

	archivedFalse := false
	filter := conversation.SessionFilter{
		CWD:           "/proj",
		Archived:      &archivedFalse,
		ModelProvider: "openai",
		SourceKinds:   []string{"cli", "exec"},
	}
	sessions, err := b.listSessionsFiltered(context.Background(), filter)
	if err != nil {
		t.Fatalf("listSessionsFiltered: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "t1" {
		t.Fatalf("unexpected sessions: %#v", sessions)
	}

	if len(src.listCall) != 1 {
		t.Fatalf("expected exactly one thread/list call (archived filter is exclusive, so only one pass), got %d", len(src.listCall))
	}
	p := src.listCall[0]
	if len(p.CWD) != 1 || p.CWD[0] != "/proj" {
		t.Fatalf("expected cwd pushed into thread/list params, got %#v", p.CWD)
	}
	if p.Archived == nil || *p.Archived != false {
		t.Fatalf("expected archived=false pushed into thread/list params, got %#v", p.Archived)
	}
	if len(p.ModelProviders) != 1 || p.ModelProviders[0] != "openai" {
		t.Fatalf("expected modelProvider pushed into thread/list params, got %#v", p.ModelProviders)
	}
	if len(p.SourceKinds) != 2 {
		t.Fatalf("expected explicit sourceKinds pushed into thread/list params, got %#v", p.SourceKinds)
	}
}

// TestAppServerBackendListAllToleratesPageFailureMidPagination is the
// DIR-032 "per-page failures produce bounded warnings and resumable
// cursors rather than whole-corpus loss" proof: page 1 of a 3-page fetch
// succeeds, page 2 fails, and listAll must still return page 1's threads
// (not lose everything), plus a warning naming the exact cursor to retry
// from, rather than the whole listing returning zero sessions and no
// diagnostic.
func TestAppServerBackendListAllToleratesPageFailureMidPagination(t *testing.T) {
	cur1 := "cursor-1"
	cur2 := "cursor-2"
	failCursor := cur1
	src := &fakeThreadSource{
		pages: map[string]appserver.ThreadListResult{
			"active:":        {Data: []appserver.Thread{{ID: "t1", CreatedAt: 1}}, NextCursor: &cur1},
			"active:" + cur1: {Data: []appserver.Thread{{ID: "t2", CreatedAt: 2}}, NextCursor: &cur2},
			"active:" + cur2: {Data: []appserver.Thread{{ID: "t3", CreatedAt: 3}}},
			"archived:":      {}, // no archived threads
		},
		failOnCursor: &failCursor,
		failErr:      errors.New("transient network error"),
	}
	b := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}

	sessions, err := b.listSessionsFiltered(context.Background(), conversation.SessionFilter{})
	if err != nil {
		t.Fatalf("listSessionsFiltered should tolerate a mid-pagination page failure, got error: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "t1" {
		t.Fatalf("expected only page 1's thread (t1) to survive the page-2 failure, got %#v", sessions)
	}

	warnings := b.drainWarnings()
	if len(warnings) != 1 {
		t.Fatalf("expected exactly one warning for the failed page, got %v", warnings)
	}
	if !strings.Contains(warnings[0], cur1) {
		t.Fatalf("expected the warning to name the resumable cursor %q, got %q", cur1, warnings[0])
	}
}

// TestAppServerBackendListAllStopsSilentlyOnMidPaginationCancellation
// proves a context cancellation that occurs AFTER at least one page
// already succeeded (the caller's own shutdown mid-pagination, not a
// transient page error) returns the partial results without error and
// without a spurious warning — unlike a genuine page failure, "the caller
// gave up" doesn't need a resumable-cursor diagnostic.
func TestAppServerBackendListAllStopsSilentlyOnMidPaginationCancellation(t *testing.T) {
	cur1 := "cursor-1"
	failCursor := cur1
	src := &fakeThreadSource{
		pages: map[string]appserver.ThreadListResult{
			"active:": {Data: []appserver.Thread{{ID: "t1", CreatedAt: 1}}, NextCursor: &cur1},
		},
		failOnCursor: &failCursor,
		failErr:      context.Canceled,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	b := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}

	sessions, err := b.listSessionsFiltered(ctx, conversation.SessionFilter{Archived: boolPtr(false)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "t1" {
		t.Fatalf("expected page 1's thread to survive, got %#v", sessions)
	}
	if warnings := b.drainWarnings(); len(warnings) != 0 {
		t.Fatalf("expected no warnings for a mid-pagination cancellation, got %v", warnings)
	}
}

func boolPtr(b bool) *bool { return &b }

// TestAppServerBackendListSessionsPageFetchesOnePage is the DIR-032
// cursor-based continuation proof at the backend level: listSessionsPage
// issues exactly one thread/list call for the given cursor and returns the
// server's own nextCursor, rather than paging through to completion like
// listSessionsFiltered/listAll does.
func TestAppServerBackendListSessionsPageFetchesOnePage(t *testing.T) {
	cur1 := "cursor-1"
	src := &fakeThreadSource{pages: map[string]appserver.ThreadListResult{
		"active:":        {Data: []appserver.Thread{{ID: "t1", CreatedAt: 1}}, NextCursor: &cur1},
		"active:" + cur1: {Data: []appserver.Thread{{ID: "t2", CreatedAt: 2}}},
	}}
	b := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}

	sessions, next, err := b.listSessionsPage(context.Background(), conversation.SessionFilter{}, "")
	if err != nil {
		t.Fatalf("listSessionsPage: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "t1" {
		t.Fatalf("expected only page 1's thread, got %#v", sessions)
	}
	if next != cur1 {
		t.Fatalf("expected NextCursor=%q, got %q", cur1, next)
	}
	if len(src.listCall) != 1 {
		t.Fatalf("expected exactly one thread/list call for a single-page fetch, got %d", len(src.listCall))
	}

	sessions, next, err = b.listSessionsPage(context.Background(), conversation.SessionFilter{}, next)
	if err != nil {
		t.Fatalf("listSessionsPage page 2: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "t2" {
		t.Fatalf("expected page 2's thread, got %#v", sessions)
	}
	if next != "" {
		t.Fatalf("expected empty NextCursor at the end of pagination, got %q", next)
	}
}

// TestAppServerBackendListSessionsFilteredArchivedNilRunsBothPasses proves
// that an unset Archived filter still runs both the non-archived and
// archived thread/list passes (matching listSessions' existing behavior),
// while an explicit Archived value narrows to exactly one pass.
func TestAppServerBackendListSessionsFilteredArchivedNilRunsBothPasses(t *testing.T) {
	src := &fakeThreadSource{pages: map[string]appserver.ThreadListResult{
		"active:":   {Data: []appserver.Thread{{ID: "t1", CreatedAt: 1}}},
		"archived:": {Data: []appserver.Thread{{ID: "t2", CreatedAt: 2}}},
	}}
	b := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}

	sessions, err := b.listSessionsFiltered(context.Background(), conversation.SessionFilter{})
	if err != nil {
		t.Fatalf("listSessionsFiltered: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected both archived-state passes merged, got %#v", sessions)
	}
	for _, s := range sessions {
		wantArchived := s.ID == "t2"
		if s.Archived != wantArchived {
			t.Fatalf("session %s: expected Archived=%v, got %v", s.ID, wantArchived, s.Archived)
		}
		wantStatus := "active"
		if wantArchived {
			wantStatus = "archived"
		}
		if s.Status != wantStatus {
			t.Fatalf("session %s: expected Status=%q, got %q", s.ID, wantStatus, s.Status)
		}
	}
}
