package codex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/provider/codex/appserver"
	providerrecords "github.com/yaleh/meta-cc/internal/provider/records"
)

// gap-codex-appserver-sessions-missing-rollout-path:
//
// The rollout-path contract is: EVERY session the Codex provider lists
// carries the on-disk rollout file that backs it, in
// Session.Extensions["rollout_path"], whichever backend answered.
//
// Stage 1 discovery (rawfiles.SelectCodexFiles) resolves each listed session
// through codexprovider.RolloutPath and hard-failed the entire call when a
// session arrived without the field — which is exactly what the app-server
// backend produced, since thread/list and thread/read have no rollout-path
// field at all (see docs/reference/codex-app-server.md). In `auto` mode
// app_server is preferred, so the discovery tool returned an error instead of
// a file list on any project that had Codex data, even though the rollouts
// were on disk and the SQLite `threads.rollout_path` column was populated.
//
// These tests assert the contract per backend — files/sqlite,
// files/rollout-fallback, and app_server — against a generated fixture Codex
// root, so a backend that stops populating the field fails here rather than
// silently degrading at the consumer. They never read this host's ~/.codex.

const (
	contractSessionID = "01a0d3ec-517d-72c3-8976-14b980447d79"
	contractOtherID   = "01a0d3c6-85db-7480-92d5-515469a6fed5"
	contractProject   = "/fixture/project"
)

// contractFixture is a self-contained Codex root: two rollout files whose
// filenames embed their session IDs, plus (optionally) a state_N.sqlite
// threads table recording the same paths.
type contractFixture struct {
	root  string
	loc   *locator.CodexLocator
	paths map[string]string // session id -> on-disk rollout file
}

func newContractFixture(t *testing.T) contractFixture {
	t.Helper()
	root := t.TempDir()
	// The locator resolves META_CC_CODEX_ROOT first; CODEX_HOME is set too so
	// any code path that consults it agrees on the same fixture root.
	t.Setenv("META_CC_CODEX_ROOT", root)
	t.Setenv("CODEX_HOME", root)

	f := contractFixture{root: root, paths: map[string]string{}}
	for _, id := range []string{contractSessionID, contractOtherID} {
		path := filepath.Join(root, "sessions", "2026", "09", "24",
			"rollout-2026-09-24T22-57-55-"+id+".jsonl")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir rollout dir: %v", err)
		}
		meta := fmt.Sprintf(
			`{"timestamp":"2026-09-24T22:57:55Z","type":"session_meta","payload":{"id":%q,"cwd":%q,"model":"gpt-5","model_provider":"openai","source":"cli"}}`,
			id, contractProject)
		if err := os.WriteFile(path, []byte(meta+"\n"), 0o644); err != nil {
			t.Fatalf("write rollout: %v", err)
		}
		f.paths[id] = path
	}
	f.loc = locator.NewCodexLocator()
	return f
}

// writeStateDB creates a state_5.sqlite whose threads table records each
// session's rollout path — the "DB column is populated" half of the repro.
func (f contractFixture) writeStateDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(f.root, "state_5.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open fixture db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY,
		rollout_path TEXT,
		thread_source TEXT,
		cwd TEXT,
		title TEXT,
		model TEXT,
		model_provider TEXT,
		tokens_used INTEGER,
		source TEXT,
		created_at INTEGER,
		parent_thread_id TEXT
	)`); err != nil {
		t.Fatalf("create threads table: %v", err)
	}
	created := int64(1758754675)
	for _, id := range []string{contractSessionID, contractOtherID} {
		if _, err := db.Exec(
			`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, source, created_at)
			 VALUES (?, ?, ?, ?, 'gpt-5', 'openai', 'cli', ?)`,
			id, f.paths[id], contractProject, id, created); err != nil {
			t.Fatalf("insert fixture thread %s: %v", id, err)
		}
	}
}

// assertContractHolds requires every listed session to resolve a rollout path
// that exists on disk, and (for the fixture's own sessions) to resolve to the
// exact file the fixture wrote.
func assertContractHolds(t *testing.T, sessions []conversation.Session, want map[string]string) {
	t.Helper()
	if len(sessions) == 0 {
		t.Fatal("expected at least one listed session")
	}
	for _, session := range sessions {
		path, err := RolloutPath(session)
		if err != nil {
			t.Fatalf("backend violated the rollout-path contract for session %s: %v", session.ID, err)
		}
		if want[session.ID] != "" && path != want[session.ID] {
			t.Fatalf("session %s resolved to %q, want %q", session.ID, path, want[session.ID])
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("session %s resolved a rollout path that does not exist on disk: %v", session.ID, err)
		}
	}
}

// appServerListing drives the real Provider over the app-server backend,
// substituting a fake threadSource for the codex subprocess (see
// appserver_provider_test.go) so no CLI is required.
func appServerListing(t *testing.T, f contractFixture, threads ...appserver.Thread) []conversation.Session {
	t.Helper()
	src := &fakeThreadSource{pages: map[string]appserver.ThreadListResult{
		archivedCursorKey(false, ""): {Data: threads},
	}}
	p := newProvider(f.loc, ModeAppServer, &appServerBackend{
		connect: connectFake(src, &noopCloser{}, nil),
		locator: f.loc,
	})
	sessions, err := p.ListSessionsFiltered(context.Background(), conversation.SessionFilter{})
	if err != nil {
		t.Fatalf("ListSessionsFiltered over app_server: %v", err)
	}
	return sessions
}

func contractThread(id string) appserver.Thread {
	return appserver.Thread{ID: id, CWD: contractProject, CreatedAt: 1758754675}
}

// TestRolloutPathContractPerBackend is the per-backend assertion the contract
// requires: each of the three Codex backends, driven over the same fixture,
// returns sessions that all satisfy RolloutPath.
func TestRolloutPathContractPerBackend(t *testing.T) {
	f := newContractFixture(t)
	f.writeStateDB(t)

	t.Run("files/sqlite", func(t *testing.T) {
		sessions, _, err := listSessionsFromDBFiltered(context.Background(), f.loc.SQLiteDB(), conversation.SessionFilter{})
		if err != nil {
			t.Fatalf("listSessionsFromDBFiltered: %v", err)
		}
		assertContractHolds(t, sessions, f.paths)
	})

	t.Run("files/rollout-fallback", func(t *testing.T) {
		sessions, _, err := discoverRolloutSessions(
			[]rolloutRoot{{path: f.loc.SessionsRoot()}}, conversation.SessionFilter{})
		if err != nil {
			t.Fatalf("discoverRolloutSessions: %v", err)
		}
		assertContractHolds(t, sessions, f.paths)
	})

	t.Run("app_server", func(t *testing.T) {
		// The protocol mapping on its own cannot supply the field — this is
		// the root cause, pinned so a future MapThread change is noticed.
		if path := rolloutPathOf(appserver.MapThread(contractThread(contractSessionID))); path != "" {
			t.Fatalf("app-server mapping unexpectedly carries a rollout path: %q", path)
		}
		sessions := appServerListing(t, f, contractThread(contractSessionID), contractThread(contractOtherID))
		assertContractHolds(t, sessions, f.paths)
	})
}

// TestRolloutPathContractSingleSession covers the one-session lookup the
// GetSession/LoadTurns path uses (resolveRolloutPath), in both the
// threads-table and the rollout-tree-only variants.
func TestRolloutPathContractSingleSession(t *testing.T) {
	t.Run("with state database", func(t *testing.T) {
		f := newContractFixture(t)
		f.writeStateDB(t)
		assertContractHolds(t, []conversation.Session{appServerGetSession(t, f, contractSessionID)}, f.paths)
	})

	t.Run("rollout tree only", func(t *testing.T) {
		f := newContractFixture(t)
		assertContractHolds(t, []conversation.Session{appServerGetSession(t, f, contractSessionID)}, f.paths)
	})
}

// appServerGetSession drives Provider.GetSession over the app-server backend.
func appServerGetSession(t *testing.T, f contractFixture, sessionID string) conversation.Session {
	t.Helper()
	src := &fakeThreadSource{thread: contractThread(sessionID)}
	p := newProvider(f.loc, ModeAppServer, &appServerBackend{
		connect: connectFake(src, &noopCloser{}, nil),
		locator: f.loc,
	})
	session, err := p.GetSession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("GetSession over app_server: %v", err)
	}
	return session
}

// TestRolloutPathContractWithoutStateDatabase covers the app-server backend
// on an install with no usable threads database: the rollout trees are the
// only source, and they still satisfy the contract. This is the path a
// SQLite-column-only "fix" would have missed.
func TestRolloutPathContractWithoutStateDatabase(t *testing.T) {
	f := newContractFixture(t)
	if len(f.loc.SQLiteDBCandidates()) != 0 {
		t.Fatalf("fixture unexpectedly has a state database: %v", f.loc.SQLiteDBCandidates())
	}
	sessions := appServerListing(t, f, contractThread(contractSessionID), contractThread(contractOtherID))
	assertContractHolds(t, sessions, f.paths)
}

// TestDiscoveryOverAppServerListingDoesNotHardFail walks the exact sequence
// rawfiles.SelectCodexFiles performs — list, scope-filter, then resolve each
// session's rollout path — and requires it to produce a file list instead of
// the reported hard error.
func TestDiscoveryOverAppServerListingDoesNotHardFail(t *testing.T) {
	f := newContractFixture(t)
	f.writeStateDB(t)

	sessions := appServerListing(t, f, contractThread(contractSessionID), contractThread(contractOtherID))
	sessions = providerrecords.FilterSessionsForScope(sessions, "project", contractProject, conversation.ProviderCodex)

	var files []string
	for _, session := range sessions {
		path, err := RolloutPath(session)
		if err != nil {
			t.Fatalf("failed to resolve rollout path for codex session %s: %v", session.ID, err)
		}
		files = append(files, path)
	}
	if len(files) != 2 {
		t.Fatalf("expected both fixture sessions to yield files, got %v", files)
	}
}

// TestRolloutPathUnavailableIsTyped pins the consumer-side degradation: a
// session that genuinely cannot be located reports ErrRolloutPathUnavailable
// so a caller enumerating many sessions can skip/report that one instead of
// treating it as an opaque failure.
func TestRolloutPathUnavailableIsTyped(t *testing.T) {
	if _, err := RolloutPath(conversation.Session{ID: "no-such-session"}); !errors.Is(err, ErrRolloutPathUnavailable) {
		t.Fatalf("RolloutPath error = %v, want ErrRolloutPathUnavailable", err)
	}
}
