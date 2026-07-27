package codex

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/yaleh/meta-cc/internal/conversation"
)

func TestSQLiteListAndGetSession(t *testing.T) {
	db, err := sql.Open("sqlite", "file:test-sqlite?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY,
		rollout_path TEXT,
		cwd TEXT,
		title TEXT,
		model TEXT,
		model_provider TEXT,
		tokens_used INTEGER,
		source TEXT,
		created_at INTEGER
	)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at)
	VALUES ('s1', '/tmp/rollout.jsonl', '/tmp', 'hello', 'gpt-5', 'openai', 7, 'cli', 1700000000)`)
	if err != nil {
		t.Fatal(err)
	}

	sessions, err := listSessionsFromDB(context.Background(), "file:test-sqlite?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("listSessionsFromDB: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "s1" {
		t.Fatalf("unexpected sessions: %#v", sessions)
	}

	session, err := getSessionFromDB(context.Background(), "file:test-sqlite?mode=memory&cache=shared", "s1")
	if err != nil {
		t.Fatalf("getSessionFromDB: %v", err)
	}
	if session.Title != "hello" {
		t.Fatalf("unexpected session: %#v", session)
	}
}

func TestSQLiteMissingTableAndUnknownSession(t *testing.T) {
	db, err := sql.Open("sqlite", "file:test-missing?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = listSessionsFromDB(context.Background(), "file:test-missing?mode=memory&cache=shared")
	var schemaErr *SQLiteSchemaError
	if !errors.As(err, &schemaErr) {
		t.Fatalf("expected SQLiteSchemaError, got %v", err)
	}

	_, err = db.Exec(`CREATE TABLE threads (id TEXT PRIMARY KEY, cwd TEXT, title TEXT, source TEXT)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = getSessionFromDB(context.Background(), "file:test-missing?mode=memory&cache=shared", "missing")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

// TestListSessionsFromDBFilteredPushesCWDIntoWhereClause is the DIR-030
// "query planning demonstrates metadata filtering before deep loading"
// proof for the files (SQLite) backend: cwd is part of every tested/
// supported threads schema, so it is pushed into a SQL WHERE clause rather
// than fetched-then-filtered in Go — this test proves the pushdown by
// seeding two projects and asserting only the matching-cwd row's session
// is returned from a query scoped to one cwd.
func TestListSessionsFromDBFilteredPushesCWDIntoWhereClause(t *testing.T) {
	dsn := "file:test-filtered-cwd?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dsn)
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
	if _, err := db.Exec(`INSERT INTO threads(id, cwd, title, source, created_at) VALUES
		('s1', '/proj-a', 'a', 'cli', 1700000000),
		('s2', '/proj-b', 'b', 'cli', 1700000001)`); err != nil {
		t.Fatal(err)
	}

	sessions, warnings, err := listSessionsFromDBFiltered(context.Background(), dsn, conversation.SessionFilter{CWD: "/proj-a"})
	if err != nil {
		t.Fatalf("listSessionsFromDBFiltered: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(sessions) != 1 || sessions[0].ID != "s1" {
		t.Fatalf("expected only the matching-cwd session, got %#v", sessions)
	}
}

// TestListSessionsFromDBFilteredToleratesCorruptRow proves the DIR-030
// "one corrupt rollout does not erase valid results from other sessions"
// contract at the threads-table scan level: a row whose required columns
// are absent/malformed is skipped with a warning, not fatal to the whole
// listing.
func TestListSessionsFromDBFilteredToleratesCorruptRow(t *testing.T) {
	dsn := "file:test-filtered-corrupt?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// A minimal schema missing the "title" column required by scanSession,
	// so every row fails to scan — this simulates a corrupted/incompatible
	// threads table without needing to hand-craft a type-mismatched value
	// (modernc.org/sqlite is quite permissive about type coercion). created_at
	// is kept so the ORDER BY clause itself still succeeds.
	if _, err := db.Exec(`CREATE TABLE threads (id TEXT PRIMARY KEY, cwd TEXT, source TEXT, created_at INTEGER)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO threads(id, cwd, source, created_at) VALUES ('broken', '/proj', 'cli', 1700000000)`); err != nil {
		t.Fatal(err)
	}

	sessions, warnings, err := listSessionsFromDBFiltered(context.Background(), dsn, conversation.SessionFilter{})
	if err != nil {
		t.Fatalf("listSessionsFromDBFiltered should tolerate a bad row, got error: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected zero sessions from an all-broken table, got %#v", sessions)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected exactly one warning for the unreadable row, got %v", warnings)
	}
}

func TestScanSessionPopulatesDIR030Metadata(t *testing.T) {
	dsn := "file:test-metadata?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dsn)
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
	if _, err := db.Exec(`INSERT INTO threads(id, cwd, title, model, model_provider, source, created_at)
		VALUES ('sub-1', '/proj', 'sub', 'gpt-5', 'openai', 'subAgent', 1700000000)`); err != nil {
		t.Fatal(err)
	}

	session, err := getSessionFromDB(context.Background(), dsn, "sub-1")
	if err != nil {
		t.Fatalf("getSessionFromDB: %v", err)
	}
	if session.SourceKind != "subAgent" {
		t.Fatalf("expected SourceKind=subAgent, got %q", session.SourceKind)
	}
	if !session.IsSubagent {
		t.Fatalf("expected IsSubagent=true for a subAgent source kind")
	}
	if session.ModelProvider != "openai" {
		t.Fatalf("expected ModelProvider=openai, got %q", session.ModelProvider)
	}
	if session.Status != "active" || session.Archived {
		t.Fatalf("expected default Status=active/Archived=false when no archived column exists, got %q/%v", session.Status, session.Archived)
	}
}
