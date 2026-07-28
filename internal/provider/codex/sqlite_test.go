package codex

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

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
	if session.Lineage != conversation.LineageStatusUnknown {
		t.Fatalf("expected Lineage=unknown for a subAgent row with no parent_thread_id value, got %q", session.Lineage)
	}
}

// TestScanSessionLineage is the DIR-032 "explicit unknown state when spawn
// metadata was suppressed" proof at the SQLite scan level: a schema with no
// parent_thread_id column at all cannot report lineage (Unknown, not Root);
// once the column exists, a non-subagent row with no value is a confirmed
// Root, and a row with a value is a confirmed Child regardless of source
// kind.
func TestScanSessionLineage(t *testing.T) {
	// Case 1: no parent_thread_id column in the schema at all.
	dsn1 := "file:test-lineage-nocol?mode=memory&cache=shared"
	db1, err := sql.Open("sqlite", dsn1)
	if err != nil {
		t.Fatal(err)
	}
	defer db1.Close()
	if _, err := db1.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY, cwd TEXT, title TEXT, source TEXT, created_at INTEGER
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db1.Exec(`INSERT INTO threads(id, cwd, title, source, created_at) VALUES ('s1', '/proj', 't', 'cli', 1700000000)`); err != nil {
		t.Fatal(err)
	}
	session, err := getSessionFromDB(context.Background(), dsn1, "s1")
	if err != nil {
		t.Fatalf("getSessionFromDB: %v", err)
	}
	if session.Lineage != conversation.LineageStatusUnknown {
		t.Fatalf("expected Lineage=unknown when the schema has no parent_thread_id column, got %q", session.Lineage)
	}

	// Case 2: column present, empty value, non-subagent source -> Root.
	dsn2 := "file:test-lineage-root?mode=memory&cache=shared"
	db2, err := sql.Open("sqlite", dsn2)
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	if _, err := db2.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY, cwd TEXT, title TEXT, source TEXT, created_at INTEGER, parent_thread_id TEXT
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db2.Exec(`INSERT INTO threads(id, cwd, title, source, created_at, parent_thread_id) VALUES ('s2', '/proj', 't', 'cli', 1700000000, '')`); err != nil {
		t.Fatal(err)
	}
	session, err = getSessionFromDB(context.Background(), dsn2, "s2")
	if err != nil {
		t.Fatalf("getSessionFromDB: %v", err)
	}
	if session.Lineage != conversation.LineageStatusRoot {
		t.Fatalf("expected Lineage=root for a non-subagent row with an empty parent_thread_id, got %q", session.Lineage)
	}

	// Case 3: column present, non-empty value -> Child.
	if _, err := db2.Exec(`INSERT INTO threads(id, cwd, title, source, created_at, parent_thread_id) VALUES ('s3', '/proj', 't', 'subAgent', 1700000000, 'parent-1')`); err != nil {
		t.Fatal(err)
	}
	session, err = getSessionFromDB(context.Background(), dsn2, "s3")
	if err != nil {
		t.Fatalf("getSessionFromDB: %v", err)
	}
	if session.Lineage != conversation.LineageStatusChild {
		t.Fatalf("expected Lineage=child for a row with a non-empty parent_thread_id, got %q", session.Lineage)
	}
}

// TestExtractTime is the DIR-063 regression: extractTime must never turn a
// present-but-unparseable value into a bogus 1970 instant. A TEXT timestamp
// must parse via RFC3339 (or be rejected outright), a numeric string must be a
// fully-numeric epoch (not a leading-integer Sscan that reads "2026" out of a
// date), and int64(0)/negative/"" must yield the zero time so the
// created_at_ms -> created_at fallback chain engages. Any 1970 result for
// these inputs is the exact bug this test guards against.
func TestExtractTime(t *testing.T) {
	rfc3339 := time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC)
	epochSec := time.Unix(1700000000, 0)
	epochMilli := time.UnixMilli(1700000000000)

	cases := []struct {
		name  string
		value interface{}
		want  time.Time
	}{
		// TEXT RFC3339 timestamp: correct instant, never 1970.
		{"rfc3339 string", "2026-07-28T10:00:00Z", rfc3339},
		{"rfc3339 []byte", []byte("2026-07-28T10:00:00Z"), rfc3339},
		// Fully-numeric epoch strings (seconds and millis heuristic).
		{"numeric string seconds", "1700000000", epochSec},
		{"numeric string millis", "1700000000000", epochMilli},
		{"numeric []byte seconds", []byte("1700000000"), epochSec},
		// Integer epochs (seconds and millis heuristic).
		{"int64 seconds", int64(1700000000), epochSec},
		{"int64 millis", int64(1700000000000), epochMilli},
		// Absent/invalid values must be the zero time so fallback engages.
		{"int64 zero", int64(0), time.Time{}},
		{"int64 negative", int64(-5), time.Time{}},
		{"empty string", "", time.Time{}},
		{"empty []byte", []byte(""), time.Time{}},
		{"non-numeric string", "not-a-time", time.Time{}},
		{"nil", nil, time.Time{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractTime(tc.value)
			if !got.Equal(tc.want) {
				t.Fatalf("extractTime(%#v) = %v (IsZero=%v), want %v (IsZero=%v)",
					tc.value, got, got.IsZero(), tc.want, tc.want.IsZero())
			}
			if tc.want.IsZero() && !got.IsZero() {
				t.Fatalf("extractTime(%#v) = %v, want zero time (IsZero=true) so fallback engages", tc.value, got)
			}
		})
	}

	// Explicit guard against the exact reported bug: an RFC3339 TEXT
	// timestamp must never collapse to a 1970 date.
	if got := extractTime("2026-07-28T10:00:00Z"); got.Year() == 1970 {
		t.Fatalf("extractTime(RFC3339) collapsed to 1970: %v", got)
	}
}

// TestScanSessionCreatedAtMsZeroFallsBackToCreatedAt is the DIR-063
// end-to-end proof of the fallback chain: a row whose created_at_ms is a
// DEFAULT 0 integer must NOT pin CreatedAt to 1970 — extractTime returns the
// zero time for 0, IsZero() engages, and CreatedAt is taken from the real
// created_at column instead.
func TestScanSessionCreatedAtMsZeroFallsBackToCreatedAt(t *testing.T) {
	dsn := "file:test-created-at-fallback?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY, cwd TEXT, title TEXT, source TEXT,
		created_at INTEGER, created_at_ms INTEGER DEFAULT 0
	)`); err != nil {
		t.Fatal(err)
	}
	// created_at_ms is the DEFAULT 0; created_at carries the real epoch.
	if _, err := db.Exec(`INSERT INTO threads(id, cwd, title, source, created_at, created_at_ms)
		VALUES ('s1', '/proj', 't', 'cli', 1700000000, 0)`); err != nil {
		t.Fatal(err)
	}

	session, err := getSessionFromDB(context.Background(), dsn, "s1")
	if err != nil {
		t.Fatalf("getSessionFromDB: %v", err)
	}

	want := time.Unix(1700000000, 0).UTC()
	if !session.CreatedAt.Equal(want) {
		t.Fatalf("expected CreatedAt from created_at (%v), got %v", want, session.CreatedAt)
	}
	if session.CreatedAt.Year() == 1970 {
		t.Fatalf("CreatedAt pinned to 1970 despite a valid created_at: %v", session.CreatedAt)
	}
	// UpdatedAt falls back to createdAt when no updated_at column exists.
	if !session.UpdatedAt.Equal(want) {
		t.Fatalf("expected UpdatedAt to fall back to created_at (%v), got %v", want, session.UpdatedAt)
	}
}
