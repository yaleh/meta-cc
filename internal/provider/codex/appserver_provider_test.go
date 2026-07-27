package codex

import (
	"context"
	"errors"
	"io"
	"testing"

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
}

func (f *fakeThreadSource) ThreadList(_ context.Context, p appserver.ThreadListParams) (appserver.ThreadListResult, error) {
	f.listCall = append(f.listCall, p)
	if f.listErr != nil {
		return appserver.ThreadListResult{}, f.listErr
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
