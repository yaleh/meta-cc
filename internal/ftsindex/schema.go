// Package ftsindex implements DIR-031: an optional, local, incremental
// full-text index over the DIR-028 canonical conversation model (Session /
// Turn / Item). It is a derived cache, never an authoritative source: raw
// provider session stores (Claude JSONL, Codex rollout/SQLite/app-server)
// remain the source of truth, and every search hit carries enough
// provenance (provider, thread_id, turn_id, item_id) to re-fetch the exact
// canonical record rather than trusting the cached snippet.
//
// See docs/reference/fts-index.md for the on-disk location, lifecycle,
// privacy defaults, and cleanup operations.
package ftsindex

import (
	"context"
	"database/sql"
	"fmt"
)

// SchemaVersion identifies the index schema shape. A stored schema_version
// that doesn't match this constant is treated exactly like corruption (see
// corruption.go): the derived index is wiped and rebuilt from scratch,
// never partially migrated in place. This keeps the invalidation logic
// simple at the cost of a full reindex on schema changes, which is
// acceptable for a derived, rebuildable cache.
const SchemaVersion = 3

// DefaultBodyLimitBytes bounds how much text from a single Item is written
// into the index (both the metadata "body" column and the FTS index),
// mirroring the conversation.NewRawItem 4KB raw-provenance cap (DIR-028).
// Text beyond this limit is never indexed, so oversized tool output cannot
// be searched in full by default (see privacy.go).
const DefaultBodyLimitBytes = 4096

const schemaDDL = `
CREATE TABLE IF NOT EXISTS meta (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	session_key    TEXT PRIMARY KEY,
	provider       TEXT NOT NULL,
	session_id     TEXT NOT NULL,
	cwd            TEXT NOT NULL DEFAULT '',
	title          TEXT NOT NULL DEFAULT '',
	source_path    TEXT NOT NULL DEFAULT '',
	source_size    INTEGER NOT NULL DEFAULT -1,
	source_mtime   INTEGER NOT NULL DEFAULT 0,
	updated_at     INTEGER NOT NULL DEFAULT 0,
	indexed_at     INTEGER NOT NULL DEFAULT 0,
	item_count     INTEGER NOT NULL DEFAULT 0,
	complete       INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_sessions_provider_cwd ON sessions(provider, cwd);

CREATE TABLE IF NOT EXISTS items (
	rowid          INTEGER PRIMARY KEY,
	session_key    TEXT NOT NULL,
	provider       TEXT NOT NULL,
	thread_id      TEXT NOT NULL,
	turn_id        TEXT NOT NULL,
	item_id        TEXT NOT NULL,
	role           TEXT NOT NULL DEFAULT '',
	kind           TEXT NOT NULL DEFAULT '',
	cwd            TEXT NOT NULL DEFAULT '',
	title          TEXT NOT NULL DEFAULT '',
	tool_name      TEXT NOT NULL DEFAULT '',
	ts_unix        INTEGER NOT NULL DEFAULT 0,
	body             TEXT NOT NULL DEFAULT '',
	truncated        INTEGER NOT NULL DEFAULT 0,
	search_body      TEXT NOT NULL DEFAULT '',
	search_truncated INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_items_session_key ON items(session_key);
CREATE INDEX IF NOT EXISTS idx_items_provider_cwd ON items(provider, cwd);

CREATE VIRTUAL TABLE IF NOT EXISTS items_fts USING fts5(
	body,
	content='items',
	content_rowid='rowid',
	tokenize='porter unicode61'
);
`

// EnsureSchema creates every table/index this package needs (idempotent —
// safe to call on every Open) and records SchemaVersion in the meta table
// on first creation.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schemaDDL); err != nil {
		return fmt.Errorf("ftsindex: ensure schema: %w", err)
	}
	_, err := db.ExecContext(ctx,
		`INSERT INTO meta(key, value) VALUES('schema_version', ?)
		 ON CONFLICT(key) DO NOTHING`, fmt.Sprintf("%d", SchemaVersion))
	if err != nil {
		return fmt.Errorf("ftsindex: record schema version: %w", err)
	}
	return nil
}

// storedSchemaVersion reads meta.schema_version, returning 0 (never equal
// to any real SchemaVersion) if unset or unreadable.
func storedSchemaVersion(ctx context.Context, db *sql.DB) int {
	var raw string
	err := db.QueryRowContext(ctx, `SELECT value FROM meta WHERE key = 'schema_version'`).Scan(&raw)
	if err != nil {
		return 0
	}
	var v int
	if _, err := fmt.Sscan(raw, &v); err != nil {
		return 0
	}
	return v
}
