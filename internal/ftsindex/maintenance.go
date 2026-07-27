package ftsindex

import (
	"context"
	"database/sql"
	"os"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// disableEnvVar lets a user opt out of the derived FTS index entirely
// (Contract: "users can disable ... the derived index"). Callers that want
// to honor it should check IsDisabled() before ever calling Open/Refresh —
// this package never reads it internally, so a caller can also always
// force-enable/bypass it for tests or explicit rebuild/inspect operations.
const disableEnvVar = "META_CC_DISABLE_FTS_INDEX"

// IsDisabled reports whether the user has opted out of the derived FTS
// index via META_CC_DISABLE_FTS_INDEX=1 (or any non-empty value other than
// "0"/"false").
func IsDisabled() bool {
	v := os.Getenv(disableEnvVar)
	return v != "" && v != "0" && v != "false"
}

// Rebuild discards any existing index at path (corrupt or not) and
// reindexes every session in sessions from scratch. It is the explicit
// counterpart to Open's automatic corruption-recovery path: a caller (CLI
// flag, MCP tool, or a test) can always force a full rebuild on demand,
// e.g. after a schema/privacy-policy change or to recover from a corrupt
// index without waiting for the next incidental Open.
func Rebuild(ctx context.Context, path string, sessions []conversation.Session, sourceMeta SourceMetaFunc, loadTurns LoadTurnsFunc, bodyLimit int) (Stats, []string, error) {
	removeIndexFiles(path)
	db, _, err := Open(ctx, path)
	if err != nil {
		return Stats{}, nil, err
	}
	defer db.Close()

	return Refresh(ctx, db, sessions, sourceMeta, loadTurns, bodyLimit)
}

// Clean permanently deletes the index database (and any WAL/SHM sidecar
// files) at path. This is the explicit "clean the derived index" operation
// — safe at any time, since the index is purely a derived cache: the next
// Open recreates an empty, healthy index, and the next Refresh/Rebuild
// repopulates it from the still-fully-intact raw session stores.
func Clean(path string) error {
	removeIndexFiles(path)
	return nil
}

// InspectResult summarizes an index's current contents for the "inspect
// the derived index" operation.
type InspectResult struct {
	SchemaVersion int
	SessionCount  int
	ItemCount     int
}

// Inspect reports the current schema version and row counts of db.
func Inspect(ctx context.Context, db *sql.DB) (InspectResult, error) {
	result := InspectResult{SchemaVersion: storedSchemaVersion(ctx, db)}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions`).Scan(&result.SessionCount); err != nil {
		return result, err
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM items`).Scan(&result.ItemCount); err != nil {
		return result, err
	}
	return result, nil
}
