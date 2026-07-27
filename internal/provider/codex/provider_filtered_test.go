package codex

import (
	"context"
	"errors"
	"testing"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/provider/codex/appserver"
)

// TestProviderListSessionsFilteredExactSessionIDReadsOnlyThatThread proves
// the "exact session lookup reads only the requested thread" contract at
// the Provider level: with filter.SessionID set, ListSessionsFiltered must
// call thread/read (getSession) and must NEVER call thread/list
// (listSessions) — listing every session would be strictly more work than
// necessary for a lookup the caller already has the exact ID for.
func TestProviderListSessionsFilteredExactSessionIDReadsOnlyThatThread(t *testing.T) {
	src := &fakeThreadSource{thread: appserver.Thread{ID: "files-1", CWD: "/tmp", CreatedAt: 1700000000}}
	fake := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}
	p := filesFixtureProvider(t, ModeAppServer, fake)

	sessions, err := p.ListSessionsFiltered(context.Background(), conversation.SessionFilter{SessionID: "files-1"})
	if err != nil {
		t.Fatalf("ListSessionsFiltered: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "files-1" {
		t.Fatalf("expected exactly the requested session, got %#v", sessions)
	}
	if len(src.listCall) != 0 {
		t.Fatalf("expected thread/list to never be called for an exact session_id lookup, got %d calls", len(src.listCall))
	}
}

// TestProviderListSessionsFilteredExactSessionIDNotFoundPropagatesError
// ensures a missing exact session_id surfaces as an error rather than a
// silently empty result, regardless of Mode.
func TestProviderListSessionsFilteredExactSessionIDNotFoundPropagatesError(t *testing.T) {
	fake := &appServerBackend{connect: connectFake(nil, nil, errors.New("app-server unavailable"))}
	p := filesFixtureProvider(t, ModeFiles, fake)

	_, err := p.ListSessionsFiltered(context.Background(), conversation.SessionFilter{SessionID: "does-not-exist"})
	if err == nil {
		t.Fatalf("expected an error for an unknown session_id")
	}
}

// TestProviderListSessionsFilteredExactSessionIDStillAppliesOtherFilters
// proves that when SessionID is set alongside another filter dimension
// (e.g. archived), that dimension is still enforced against the one
// fetched session (not bypassed just because a fast path was taken).
func TestProviderListSessionsFilteredExactSessionIDStillAppliesOtherFilters(t *testing.T) {
	fake := &appServerBackend{connect: connectFake(nil, nil, errors.New("app-server unavailable"))}
	p := filesFixtureProvider(t, ModeFiles, fake) // files-1 has no archived column -> Archived=false

	archivedTrue := true
	sessions, err := p.ListSessionsFiltered(context.Background(), conversation.SessionFilter{
		SessionID: "files-1",
		Archived:  &archivedTrue,
	})
	if err != nil {
		t.Fatalf("ListSessionsFiltered: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected the archived=true filter to exclude a non-archived session, got %#v", sessions)
	}
}

// TestProviderListSessionsFilteredCWDBoundaryAppliesAsCorrectnessBackstop
// proves that even when a backend doesn't push a filter dimension down
// (files backend only pushes cwd; this exercises source_kind, which it
// does not push into SQL), conversation.ApplyFilter still enforces it
// afterward so results are correct regardless of pushdown completeness.
func TestProviderListSessionsFilteredCWDBoundaryAppliesAsCorrectnessBackstop(t *testing.T) {
	fake := &appServerBackend{connect: connectFake(nil, nil, errors.New("app-server unavailable"))}
	p := filesFixtureProvider(t, ModeFiles, fake) // files-1 has source="cli"

	sessions, err := p.ListSessionsFiltered(context.Background(), conversation.SessionFilter{SourceKinds: []string{"vscode"}})
	if err != nil {
		t.Fatalf("ListSessionsFiltered: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected source_kind filter to exclude the cli-sourced session, got %#v", sessions)
	}

	sessions, err = p.ListSessionsFiltered(context.Background(), conversation.SessionFilter{SourceKinds: []string{"cli"}})
	if err != nil {
		t.Fatalf("ListSessionsFiltered: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "files-1" {
		t.Fatalf("expected the cli-sourced session to match, got %#v", sessions)
	}
}
