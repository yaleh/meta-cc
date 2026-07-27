package ftsindex

import (
	"context"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

func TestRefresh_IndexesNewSessions(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	sess := session(conversation.ProviderClaude, "s1", "/proj", "first session", now)
	turns := []conversation.Turn{
		turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "please add banana support", now)),
	}
	loader := newCountingLoader(map[string][]conversation.Turn{sessionKey("claude", "s1"): turns})
	metaMap := map[string]SourceMeta{sessionKey("claude", "s1"): {Path: "/fake/s1.jsonl", Size: 100, ModTime: now}}

	stats, warnings, err := Refresh(ctx, db, []conversation.Session{sess}, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if stats.SessionsIndexed != 1 || stats.ItemsIndexed != 1 {
		t.Fatalf("stats = %+v, want SessionsIndexed=1 ItemsIndexed=1", stats)
	}
	if loader.calls[sessionKey("claude", "s1")] != 1 {
		t.Fatalf("loader called %d times, want 1", loader.calls[sessionKey("claude", "s1")])
	}
}

func TestRefresh_SkipsUnchangedSessions_NoDeepReparse(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	sessions := []conversation.Session{
		session(conversation.ProviderClaude, "s1", "/proj", "session one", now),
		session(conversation.ProviderClaude, "s2", "/proj", "session two", now),
	}
	turnsMap := map[string][]conversation.Turn{
		sessionKey("claude", "s1"): {turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "alpha", now))},
		sessionKey("claude", "s2"): {turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "beta", now))},
	}
	metaMap := map[string]SourceMeta{
		sessionKey("claude", "s1"): {Path: "/fake/s1.jsonl", Size: 100, ModTime: now},
		sessionKey("claude", "s2"): {Path: "/fake/s2.jsonl", Size: 200, ModTime: now},
	}

	loader := newCountingLoader(turnsMap)
	if _, _, err := Refresh(ctx, db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		t.Fatalf("first Refresh: %v", err)
	}

	// Second Refresh over the SAME corpus with identical SourceMeta: neither
	// session's turns should be reparsed at all.
	stats, _, err := Refresh(ctx, db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes)
	if err != nil {
		t.Fatalf("second Refresh: %v", err)
	}
	if stats.SessionsSkipped != 2 || stats.SessionsIndexed != 0 {
		t.Fatalf("second Refresh stats = %+v, want SessionsSkipped=2 SessionsIndexed=0", stats)
	}
	if got := loader.calls[sessionKey("claude", "s1")]; got != 1 {
		t.Fatalf("s1 loader calls = %d, want exactly 1 (no deep reparse on unchanged corpus)", got)
	}
	if got := loader.calls[sessionKey("claude", "s2")]; got != 1 {
		t.Fatalf("s2 loader calls = %d, want exactly 1", got)
	}
}

func TestRefresh_ReindexesOnlyChangedSession(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	sessions := []conversation.Session{
		session(conversation.ProviderClaude, "s1", "/proj", "session one", now),
		session(conversation.ProviderClaude, "s2", "/proj", "session two", now),
	}
	turnsMap := map[string][]conversation.Turn{
		sessionKey("claude", "s1"): {turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "alpha original", now))},
		sessionKey("claude", "s2"): {turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "beta original", now))},
	}
	metaMap := map[string]SourceMeta{
		sessionKey("claude", "s1"): {Path: "/fake/s1.jsonl", Size: 100, ModTime: now},
		sessionKey("claude", "s2"): {Path: "/fake/s2.jsonl", Size: 200, ModTime: now},
	}
	loader := newCountingLoader(turnsMap)
	if _, _, err := Refresh(ctx, db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		t.Fatalf("first Refresh: %v", err)
	}

	// Mutate only s1's SourceMeta (simulating a rewrite: bigger file, later
	// mtime) and its content.
	metaMap[sessionKey("claude", "s1")] = SourceMeta{Path: "/fake/s1.jsonl", Size: 999, ModTime: now.Add(time.Hour)}
	turnsMap[sessionKey("claude", "s1")] = []conversation.Turn{
		turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "alpha rewritten content", now)),
	}

	stats, _, err := Refresh(ctx, db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes)
	if err != nil {
		t.Fatalf("second Refresh: %v", err)
	}
	if stats.SessionsIndexed != 1 || stats.SessionsSkipped != 1 {
		t.Fatalf("stats = %+v, want SessionsIndexed=1 (s1) SessionsSkipped=1 (s2)", stats)
	}
	if got := loader.calls[sessionKey("claude", "s1")]; got != 2 {
		t.Fatalf("s1 loader calls = %d, want 2 (reparsed after change)", got)
	}
	if got := loader.calls[sessionKey("claude", "s2")]; got != 1 {
		t.Fatalf("s2 loader calls = %d, want 1 (unaffected)", got)
	}

	// No duplicate rows: exactly one items row for s1's single item, with
	// the rewritten content, not two.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM items WHERE session_key = ?`, sessionKey("claude", "s1")).Scan(&count); err != nil {
		t.Fatalf("count items: %v", err)
	}
	if count != 1 {
		t.Fatalf("items row count for rewritten session = %d, want 1 (updated in place, not duplicated)", count)
	}

	hits, err := Search(ctx, db, "rewritten", SearchFilter{CWD: "/proj"}, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("Search for rewritten content returned %d hits, want 1", len(hits))
	}
	if hits, _ := Search(ctx, db, "original", SearchFilter{CWD: "/proj", ThreadID: "s1"}, 10); len(hits) != 0 {
		t.Fatalf("stale pre-rewrite content for s1 is still searchable: %d hits", len(hits))
	}
}

func TestRefresh_TransactionalOnInterruptedIndexing(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	sess := session(conversation.ProviderClaude, "s1", "/proj", "session one", now)
	originalTurns := []conversation.Turn{
		turn("t1", now,
			textItem("i1", conversation.ItemKindUserMessage, "user", "original stable content", now),
			textItem("i2", conversation.ItemKindAgentMessage, "assistant", "original reply", now),
		),
	}
	metaMap := map[string]SourceMeta{sessionKey("claude", "s1"): {Path: "/fake/s1.jsonl", Size: 100, ModTime: now}}
	loader := newCountingLoader(map[string][]conversation.Turn{sessionKey("claude", "s1"): originalTurns})

	if _, _, err := Refresh(ctx, db, []conversation.Session{sess}, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		t.Fatalf("initial Refresh: %v", err)
	}

	// Now simulate an interrupted reindex: SourceMeta changes (forcing a
	// real reindex attempt), new content is supplied, but the transaction
	// fails partway through (after the first item insert) via the
	// package-private test-only injection hook.
	metaMap[sessionKey("claude", "s1")] = SourceMeta{Path: "/fake/s1.jsonl", Size: 500, ModTime: now.Add(time.Hour)}
	loader.turns[sessionKey("claude", "s1")] = []conversation.Turn{
		turn("t1", now,
			textItem("i1", conversation.ItemKindUserMessage, "user", "corrupted partial write attempt", now),
			textItem("i2", conversation.ItemKindAgentMessage, "assistant", "should never land", now),
		),
	}
	testFailAfterNItems = 1
	defer func() { testFailAfterNItems = 0 }()

	stats, warnings, err := Refresh(ctx, db, []conversation.Session{sess}, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes)
	if err != nil {
		t.Fatalf("Refresh returned a hard error instead of a per-session warning: %v", err)
	}
	if stats.SessionsFailed != 1 {
		t.Fatalf("stats.SessionsFailed = %d, want 1", stats.SessionsFailed)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %v, want exactly 1", warnings)
	}

	// The OLD rows must survive completely intact: neither a mix of old+new
	// nor empty.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM items WHERE session_key = ?`, sessionKey("claude", "s1")).Scan(&count); err != nil {
		t.Fatalf("count items: %v", err)
	}
	if count != 2 {
		t.Fatalf("items row count after interrupted reindex = %d, want 2 (old rows intact)", count)
	}

	if hits, _ := Search(ctx, db, "corrupted", SearchFilter{CWD: "/proj"}, 10); len(hits) != 0 {
		t.Fatalf("interrupted (never-committed) new content is searchable: %d hits", len(hits))
	}
	if hits, _ := Search(ctx, db, "stable", SearchFilter{CWD: "/proj"}, 10); len(hits) != 1 {
		t.Fatalf("old content did not survive the interrupted reindex: %d hits", len(hits))
	}

	// FTS and items tables must also stay in sync (no orphaned fts rows).
	var ftsCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM items_fts`).Scan(&ftsCount); err != nil {
		t.Fatalf("count items_fts: %v", err)
	}
	if ftsCount != count {
		t.Fatalf("items_fts count = %d, items count = %d: out of sync after rollback", ftsCount, count)
	}
}
