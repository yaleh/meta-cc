package ftsindex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// SourceMetaFunc resolves a session's current on-disk invalidation
// fingerprint. Callers typically implement this via
// claudeprovider.FilePath/codexprovider.RolloutPath + os.Stat (see
// meta_source.go for a ready-made implementation covering both).
type SourceMetaFunc func(session conversation.Session) SourceMeta

// LoadTurnsFunc loads one session's full canonical Turn/Item content —
// typically provider.Provider.LoadTurns. Refresh only calls this for
// sessions whose SourceMeta changed, which is what makes "an unchanged
// corpus performs no deep reparse" true: callers can wrap this func to
// count invocations in tests.
type LoadTurnsFunc func(ctx context.Context, session conversation.Session) ([]conversation.Turn, error)

// Stats summarizes one Refresh call.
type Stats struct {
	SessionsSeen    int
	SessionsSkipped int // SourceMeta unchanged: no reparse, no reindex
	SessionsIndexed int // newly indexed or reindexed after a change
	SessionsFailed  int // LoadTurns or the transactional upsert failed
	ItemsIndexed    int
}

// testFailAfterNItems and errTestInjectedFailure are a test-only fault
// injection hook (set only by indexer_test.go, in this same package) used
// to simulate a mid-transaction failure without a full fault-injecting SQL
// driver: reindexSessionTx checks this after inserting each item, and
// returns errTestInjectedFailure once the Nth item has been inserted
// (within the still-open, not-yet-committed transaction), letting the test
// assert the transaction rolled back and the session's previous rows
// survived intact rather than ending up a mix of old and new. Zero value
// (0) disables injection — production code paths never set this.
var testFailAfterNItems int

var errTestInjectedFailure = errors.New("ftsindex: injected test failure")

// Refresh indexes sessions incrementally: a session whose SourceMeta
// (source path/size/mtime, or UpdatedAt when path-less) is unchanged from
// what's already stored is skipped without calling loadTurns at all. A
// session that is new or changed is fully reindexed via a single
// per-session transaction (reindexSessionTx) that replaces all of that
// session's rows atomically. One session's failure (LoadTurns error, or a
// transactional upsert failure) is recorded as a warning and does not abort
// the rest of the batch, mirroring providerrecords.Build's per-session
// tolerance (DIR-030).
func Refresh(ctx context.Context, db *sql.DB, sessions []conversation.Session, sourceMeta SourceMetaFunc, loadTurns LoadTurnsFunc, bodyLimit int) (Stats, []string, error) {
	var stats Stats
	var warnings []string

	for _, session := range sessions {
		stats.SessionsSeen++
		key := sessionKey(string(session.Provider), session.ID)
		meta := sourceMeta(session)

		row, ok, err := getSessionRow(ctx, db, key)
		if err != nil {
			return stats, warnings, fmt.Errorf("ftsindex: read session row %s: %w", key, err)
		}
		if ok && meta.unchanged(row) {
			stats.SessionsSkipped++
			continue
		}

		turns, err := loadTurns(ctx, session)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("ftsindex: session %s: failed to load turns, skipped: %v", key, err))
			stats.SessionsFailed++
			continue
		}

		n, err := reindexSessionTx(ctx, db, key, session, meta, turns, bodyLimit)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("ftsindex: session %s: reindex failed, previous index left intact: %v", key, err))
			stats.SessionsFailed++
			continue
		}
		stats.SessionsIndexed++
		stats.ItemsIndexed += n
	}

	return stats, warnings, nil
}

// reindexSessionTx replaces every indexed row for one session inside a
// single transaction: old rows (items_fts, items, sessions) are deleted and
// new rows inserted, then committed together. Any error before Commit
// (including the test-only injected failure above) triggers the deferred
// Rollback, so a failure partway through insertion leaves the PREVIOUS
// session state completely untouched rather than a mix of old and new rows
// — satisfying the Contract's "interrupted indexing ... never exposes
// partially replaced session data".
func reindexSessionTx(ctx context.Context, db *sql.DB, key string, session conversation.Session, meta SourceMeta, turns []conversation.Turn, bodyLimit int) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op once Commit succeeds

	if _, err := tx.ExecContext(ctx, `DELETE FROM items_fts WHERE rowid IN (SELECT rowid FROM items WHERE session_key = ?)`, key); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM items WHERE session_key = ?`, key); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE session_key = ?`, key); err != nil {
		return 0, err
	}

	count := 0
	for _, turn := range turns {
		for idx, item := range turn.Items {
			body, truncated := itemBody(item, bodyLimit)
			ts := item.Timestamp
			if ts.IsZero() {
				ts = turn.Timestamp
			}
			res, err := tx.ExecContext(ctx, `
				INSERT INTO items (session_key, provider, thread_id, turn_id, item_id, role, kind, cwd, title, tool_name, ts_unix, body, truncated)
				VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
				key, string(session.Provider), session.ID, turn.ID, itemID(turn.ID, idx, item),
				item.Role, string(item.Kind), session.CWD, session.Title, item.ToolName,
				ts.Unix(), body, boolToInt(truncated))
			if err != nil {
				return count, err
			}
			rowid, err := res.LastInsertId()
			if err != nil {
				return count, err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO items_fts(rowid, body) VALUES (?, ?)`, rowid, body); err != nil {
				return count, err
			}
			count++

			if testFailAfterNItems > 0 && count == testFailAfterNItems {
				return count, errTestInjectedFailure
			}
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO sessions (session_key, provider, session_id, cwd, title, source_path, source_size, source_mtime, updated_at, indexed_at, item_count)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		key, string(session.Provider), session.ID, session.CWD, session.Title,
		meta.Path, meta.Size, meta.ModTime.Unix(), meta.UpdatedAt.Unix(), time.Now().Unix(), count,
	); err != nil {
		return count, err
	}

	if err := tx.Commit(); err != nil {
		return count, err
	}
	return count, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Reconcile removes stale index rows for one (provider, cwd) scope: any
// currently-indexed session whose session_key is not among liveKeys is
// deleted (index rows for deleted/archived/moved-away sessions), reusing
// the same delete-fts-then-items-then-sessions sequence as
// reindexSessionTx, each in its own transaction. Callers should pass the
// full live session_key set for the SAME (provider, cwd) scope they just
// listed — passing a narrower set than what's really on disk would
// incorrectly treat still-live sessions as deleted.
func Reconcile(ctx context.Context, db *sql.DB, provider conversation.ProviderID, cwd string, liveKeys map[string]bool) (int, error) {
	indexed, err := currentSessionKeys(ctx, db, string(provider), cwd)
	if err != nil {
		return 0, err
	}

	removed := 0
	for key := range indexed {
		if liveKeys[key] {
			continue
		}
		if err := deleteSessionTx(ctx, db, key); err != nil {
			return removed, fmt.Errorf("ftsindex: reconcile: delete stale session %s: %w", key, err)
		}
		removed++
	}
	return removed, nil
}

func deleteSessionTx(ctx context.Context, db *sql.DB, key string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx, `DELETE FROM items_fts WHERE rowid IN (SELECT rowid FROM items WHERE session_key = ?)`, key); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM items WHERE session_key = ?`, key); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE session_key = ?`, key); err != nil {
		return err
	}
	return tx.Commit()
}
