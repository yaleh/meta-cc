package ftsindex

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"
)

// Open opens (creating if necessary) the FTS index database at path,
// verifying it is structurally healthy before returning it. If the file is
// corrupt/unreadable, or its schema_version doesn't match SchemaVersion, it
// is wiped and recreated from scratch — corruption recovery is "start over
// with a fresh, empty, correctly-shaped index", never a partial repair
// attempt, and the raw provider session stores it will be rebuilt from
// remain fully intact throughout (see docs/reference/fts-index.md).
//
// The returned degraded flag is true when this Open had to wipe/recreate
// the file, so callers can decide whether to trigger a full Rebuild
// afterward (an empty-but-healthy index degrades every search to "no
// results" rather than crashing or returning wrong results, but a caller
// that wants search to actually work again should reindex).
func Open(ctx context.Context, path string) (db *sql.DB, degraded bool, err error) {
	if err := EnsureDir(path); err != nil {
		return nil, false, err
	}

	db, openErr := sql.Open("sqlite", path)
	if openErr == nil {
		if healthy(ctx, db) {
			return db, false, nil
		}
		_ = db.Close()
	}

	// Unhealthy or unopenable: this is the corruption-recovery path. Never
	// propagate the failure to the caller as a hard error — degrade to a
	// freshly recreated, empty index instead, per the Contract ("index
	// corruption degrades to rebuild ... never crashes or silently returns
	// wrong/empty results without indication" — the `degraded` return value
	// is that indication).
	slog.Warn("ftsindex: index unhealthy or unreadable, recreating", "path", path, "open_err", openErr)
	removeIndexFiles(path)

	db, err = sql.Open("sqlite", path)
	if err != nil {
		return nil, true, err
	}
	if err := EnsureSchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, true, err
	}
	return db, true, nil
}

// healthy reports whether db is a structurally sound ftsindex database at
// the current SchemaVersion: openable, passes SQLite's own integrity check,
// and (once EnsureSchema has run) has core tables present with the expected
// schema version recorded.
func healthy(ctx context.Context, db *sql.DB) bool {
	if err := db.PingContext(ctx); err != nil {
		return false
	}
	var quick string
	if err := db.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&quick); err != nil || quick != "ok" {
		return false
	}
	if err := EnsureSchema(ctx, db); err != nil {
		return false
	}
	if v := storedSchemaVersion(ctx, db); v != SchemaVersion {
		// A version put there by IsHealthy's EnsureSchema call always equals
		// SchemaVersion for a fresh DB (INSERT ... ON CONFLICT DO NOTHING
		// only preserves a *pre-existing* different value), so this branch
		// only fires for a genuinely older/newer on-disk schema.
		return false
	}
	// A quick_check that passes on a file containing arbitrary garbage
	// bytes (not a valid SQLite file at all) never gets this far: SQLite
	// itself rejects it before quick_check runs. But an empty/zero-byte
	// file is treated by SQLite as "a new, empty database", which is a
	// legitimate healthy starting state, so no extra check is needed here.
	return true
}

// removeIndexFiles deletes the main db file plus any WAL/SHM sidecar files
// SQLite may have left behind, ignoring "not found" errors.
func removeIndexFiles(path string) {
	for _, suffix := range []string{"", "-wal", "-shm", "-journal"} {
		_ = os.Remove(path + suffix)
	}
}
