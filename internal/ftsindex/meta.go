package ftsindex

import (
	"context"
	"database/sql"
	"time"
)

// SourceMeta is the invalidation fingerprint for one session's authoritative
// backing store, per the Contract ("rollout path, size, mtime, and
// backend/schema version where available"). Path/Size/ModTime are populated
// whenever the session is backed by a real on-disk file (Claude JSONL,
// Codex rollout via the files backend); for sessions with no resolvable
// local file (e.g. Codex app-server-only threads), Path is left empty and
// UpdatedAt — the provider's own authoritative last-modified signal (DIR-030
// Session.UpdatedAt) — is used as the invalidation fingerprint instead.
type SourceMeta struct {
	Path      string
	Size      int64
	ModTime   time.Time
	UpdatedAt time.Time
}

// unchanged reports whether meta still matches a previously stored session
// row, i.e. whether reindexing that session can be skipped entirely. A
// path-backed session compares (size, mtime); a path-less session compares
// UpdatedAt. Switching between the two modes (e.g. a session that used to
// resolve a path and now doesn't) is always treated as "changed" so it
// never silently skips a real reindex.
func (m SourceMeta) unchanged(row sessionRow) bool {
	if m.Path != "" {
		return row.sourcePath == m.Path && row.sourceSize == m.Size && row.sourceMtime == m.ModTime.Unix()
	}
	if row.sourcePath != "" {
		return false
	}
	return row.updatedAt == m.UpdatedAt.Unix()
}

type sessionRow struct {
	sessionKey  string
	provider    string
	sessionID   string
	cwd         string
	sourcePath  string
	sourceSize  int64
	sourceMtime int64
	updatedAt   int64
}

func sessionKey(provider, sessionID string) string {
	return provider + ":" + sessionID
}

func getSessionRow(ctx context.Context, db *sql.DB, key string) (sessionRow, bool, error) {
	var row sessionRow
	err := db.QueryRowContext(ctx,
		`SELECT session_key, provider, session_id, cwd, source_path, source_size, source_mtime, updated_at
		 FROM sessions WHERE session_key = ?`, key,
	).Scan(&row.sessionKey, &row.provider, &row.sessionID, &row.cwd, &row.sourcePath, &row.sourceSize, &row.sourceMtime, &row.updatedAt)
	if err == sql.ErrNoRows {
		return sessionRow{}, false, nil
	}
	if err != nil {
		return sessionRow{}, false, err
	}
	return row, true, nil
}

// currentSessionKeys returns every session_key currently indexed for a
// given provider+cwd scope, used by Reconcile to detect stale rows.
func currentSessionKeys(ctx context.Context, db *sql.DB, provider, cwd string) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT session_key FROM sessions WHERE provider = ? AND cwd = ?`, provider, cwd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]bool{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		out[key] = true
	}
	return out, rows.Err()
}
