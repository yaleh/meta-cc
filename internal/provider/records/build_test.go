package records

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/conversation"
	providerpkg "github.com/yaleh/meta-cc/internal/provider"
)

// fakeProvider is a minimal providerpkg.Provider used to test Build/
// BuildForSession's per-session tolerance and exact-session-id fast path
// without needing real Codex/Claude fixtures on disk.
type fakeProvider struct {
	id             conversation.ProviderID
	available      bool
	sessions       []conversation.Session
	turnsBySession map[string][]conversation.Turn
	turnsErr       map[string]error // sessionID -> LoadTurns error
	getSessionErr  map[string]error // sessionID -> GetSession error

	loadTurnsCalls  []string
	getSessionCalls []string
}

func (f *fakeProvider) ID() conversation.ProviderID      { return f.id }
func (f *fakeProvider) IsAvailable(context.Context) bool { return f.available }

func (f *fakeProvider) ListSessions(context.Context) ([]conversation.Session, error) {
	return f.sessions, nil
}

func (f *fakeProvider) GetSession(_ context.Context, sessionID string) (conversation.Session, error) {
	f.getSessionCalls = append(f.getSessionCalls, sessionID)
	if err, ok := f.getSessionErr[sessionID]; ok {
		return conversation.Session{}, err
	}
	for _, s := range f.sessions {
		if s.ID == sessionID {
			return s, nil
		}
	}
	return conversation.Session{}, errors.New("session not found")
}

func (f *fakeProvider) LoadTurns(_ context.Context, sessionID string) ([]conversation.Turn, error) {
	f.loadTurnsCalls = append(f.loadTurnsCalls, sessionID)
	if err, ok := f.turnsErr[sessionID]; ok {
		return nil, err
	}
	return f.turnsBySession[sessionID], nil
}

// TestBuild_OneSessionLoadTurnsFailureDoesNotAbortOthers is a fake-provider
// unit-level companion to the fixture-based end-to-end test in
// internal/mcp/executor: it proves Build (used by every content/signal
// query and by the analysis-tools loadData path) keeps the results from
// every session whose LoadTurns succeeded, recording a warning for the one
// that failed instead of returning a hard error.
func TestBuild_OneSessionLoadTurnsFailureDoesNotAbortOthers(t *testing.T) {
	p := &fakeProvider{
		id:        conversation.ProviderCodex,
		available: true,
		sessions: []conversation.Session{
			{ID: "good-1", Provider: conversation.ProviderCodex, CWD: "/proj"},
			{ID: "bad", Provider: conversation.ProviderCodex, CWD: "/proj"},
			{ID: "good-2", Provider: conversation.ProviderCodex, CWD: "/proj"},
		},
		turnsBySession: map[string][]conversation.Turn{
			"good-1": {{ID: "t1", UserText: "hi"}},
			"good-2": {{ID: "t2", UserText: "hey"}},
		},
		turnsErr: map[string]error{"bad": errors.New("corrupt rollout")},
	}
	registry := providerpkg.NewRegistry(p)

	records, warnings, err := Build(context.Background(), registry, []conversation.ProviderID{conversation.ProviderCodex}, "project", "/proj")
	if err != nil {
		t.Fatalf("Build should tolerate one session's LoadTurns failure, got error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected records from both good sessions, got %d: %#v", len(records), records)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "bad") {
		t.Fatalf("expected exactly one warning naming the failed session, got %#v", warnings)
	}
}

// TestBuildForSession_ExactLookupNeverListsAllSessions proves the "exact
// session lookup reads only the requested thread" contract: BuildForSession
// must call GetSession/LoadTurns for the requested ID only, and must never
// call ListSessions at all.
func TestBuildForSession_ExactLookupNeverListsAllSessions(t *testing.T) {
	p := &fakeProvider{
		id:        conversation.ProviderCodex,
		available: true,
		sessions: []conversation.Session{
			{ID: "target", Provider: conversation.ProviderCodex, CWD: "/proj"},
			{ID: "other", Provider: conversation.ProviderCodex, CWD: "/proj"},
		},
		turnsBySession: map[string][]conversation.Turn{
			"target": {{ID: "t1", UserText: "hi"}},
		},
	}
	registry := providerpkg.NewRegistry(p)

	records, warnings, err := BuildForSession(context.Background(), registry, []conversation.ProviderID{conversation.ProviderCodex}, "target", "/proj")
	if err != nil {
		t.Fatalf("BuildForSession: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(records) != 1 {
		t.Fatalf("expected exactly one record, got %#v", records)
	}
	if len(p.getSessionCalls) != 1 || p.getSessionCalls[0] != "target" {
		t.Fatalf("expected exactly one GetSession(target) call, got %#v", p.getSessionCalls)
	}
	if len(p.loadTurnsCalls) != 1 || p.loadTurnsCalls[0] != "target" {
		t.Fatalf("expected exactly one LoadTurns(target) call, got %#v", p.loadTurnsCalls)
	}
}

// TestBuildForSession_CWDBoundaryExcludesCrossProjectSession proves a
// session_id lookup cannot cross project boundaries: a session that exists
// but belongs to a different cwd must be excluded (with a warning), not
// returned just because the ID matched.
func TestBuildForSession_CWDBoundaryExcludesCrossProjectSession(t *testing.T) {
	p := &fakeProvider{
		id:        conversation.ProviderCodex,
		available: true,
		sessions: []conversation.Session{
			{ID: "target", Provider: conversation.ProviderCodex, CWD: "/other-project"},
		},
	}
	registry := providerpkg.NewRegistry(p)

	_, warnings, err := BuildForSession(context.Background(), registry, []conversation.ProviderID{conversation.ProviderCodex}, "target", "/proj")
	if err == nil {
		t.Fatalf("expected an error when the only match is outside the project boundary")
	}
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "different project") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a cwd-boundary warning, got %#v", warnings)
	}
}
