package ftsindex

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// openTestDB opens a fresh, schema-initialized index database backed by a
// real file in t.TempDir() (not :memory:, so tests can exercise Open's
// corruption-recovery path against a real file on disk).
func openTestDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fts.db")
	db, degraded, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if degraded {
		t.Fatalf("Open reported degraded on a brand-new path")
	}
	t.Cleanup(func() { db.Close() })
	return db, path
}

func session(provider conversation.ProviderID, id, cwd, title string, updatedAt time.Time) conversation.Session {
	return conversation.Session{
		ID:        id,
		Provider:  provider,
		CWD:       cwd,
		Title:     title,
		CreatedAt: updatedAt,
		UpdatedAt: updatedAt,
	}
}

func textItem(id string, kind conversation.ItemKind, role, text string, ts time.Time) conversation.Item {
	return conversation.Item{ID: id, Kind: kind, Role: role, Text: text, Timestamp: ts}
}

func toolResultItem(id, toolName, output string, ts time.Time) conversation.Item {
	return conversation.Item{ID: id, Kind: conversation.ItemKindToolResult, ToolName: toolName, Output: output, Timestamp: ts}
}

func turn(id string, ts time.Time, items ...conversation.Item) conversation.Turn {
	return conversation.Turn{ID: id, Timestamp: ts, Items: items}
}

// staticSourceMeta returns a SourceMetaFunc backed by an in-memory,
// per-session map the test can mutate between Refresh calls to simulate a
// file changing (or a session gaining/losing UpdatedAt-only invalidation).
func staticSourceMeta(m map[string]SourceMeta) SourceMetaFunc {
	return func(s conversation.Session) SourceMeta {
		return m[sessionKey(string(s.Provider), s.ID)]
	}
}

// countingLoader wraps a fixed session->turns map with an invocation
// counter per session key, so tests can assert exactly which sessions were
// (not) reparsed by a given Refresh call.
type countingLoader struct {
	turns map[string][]conversation.Turn
	calls map[string]int
}

func newCountingLoader(turns map[string][]conversation.Turn) *countingLoader {
	return &countingLoader{turns: turns, calls: map[string]int{}}
}

func (c *countingLoader) Load(_ context.Context, s conversation.Session) ([]conversation.Turn, error) {
	key := sessionKey(string(s.Provider), s.ID)
	c.calls[key]++
	return c.turns[key], nil
}
