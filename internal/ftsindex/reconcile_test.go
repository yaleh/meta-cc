package ftsindex

import (
	"context"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

func TestReconcile_RemovesDeletedSession(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	sessions := []conversation.Session{
		session(conversation.ProviderClaude, "s1", "/proj", "keep me", now),
		session(conversation.ProviderClaude, "s2", "/proj", "delete me", now),
	}
	turnsMap := map[string][]conversation.Turn{
		sessionKey("claude", "s1"): {turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "alpha", now))},
		sessionKey("claude", "s2"): {turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "gamma", now))},
	}
	metaMap := map[string]SourceMeta{
		sessionKey("claude", "s1"): {Path: "/fake/s1.jsonl", Size: 100, ModTime: now},
		sessionKey("claude", "s2"): {Path: "/fake/s2.jsonl", Size: 100, ModTime: now},
	}
	loader := newCountingLoader(turnsMap)
	if _, _, err := Refresh(ctx, db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	// s2 is "deleted from disk": it no longer appears in the live listing.
	liveKeys := map[string]bool{sessionKey("claude", "s1"): true}
	removed, err := Reconcile(ctx, db, conversation.ProviderClaude, "/proj", liveKeys)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if removed != 1 {
		t.Fatalf("Reconcile removed = %d, want 1", removed)
	}

	var sessionCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE session_key = ?`, sessionKey("claude", "s2")).Scan(&sessionCount); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessionCount != 0 {
		t.Fatalf("deleted session's row still present in sessions table")
	}
	var itemCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM items WHERE session_key = ?`, sessionKey("claude", "s2")).Scan(&itemCount); err != nil {
		t.Fatalf("count items: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("deleted session's items still present")
	}

	if hits, _ := Search(ctx, db, "gamma", SearchFilter{CWD: "/proj"}, 10); len(hits) != 0 {
		t.Fatalf("deleted session's content still searchable: %d hits", len(hits))
	}
	if hits, _ := Search(ctx, db, "alpha", SearchFilter{CWD: "/proj"}, 10); len(hits) != 1 {
		t.Fatalf("kept session's content no longer searchable: %d hits", len(hits))
	}
}

func TestReconcile_DoesNotTouchOtherProjects(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	sessions := []conversation.Session{
		session(conversation.ProviderClaude, "s1", "/projA", "a", now),
		session(conversation.ProviderClaude, "s1-b", "/projB", "b", now),
	}
	turnsMap := map[string][]conversation.Turn{
		sessionKey("claude", "s1"):   {turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "in project a", now))},
		sessionKey("claude", "s1-b"): {turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "in project b", now))},
	}
	metaMap := map[string]SourceMeta{
		sessionKey("claude", "s1"):   {Path: "/fake/a.jsonl", Size: 1, ModTime: now},
		sessionKey("claude", "s1-b"): {Path: "/fake/b.jsonl", Size: 1, ModTime: now},
	}
	loader := newCountingLoader(turnsMap)
	if _, _, err := Refresh(ctx, db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	// Reconcile projA with an EMPTY live set: projA's session should be
	// removed, but projB's must be entirely unaffected.
	removed, err := Reconcile(ctx, db, conversation.ProviderClaude, "/projA", map[string]bool{})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}
	if hits, _ := Search(ctx, db, "project", SearchFilter{CWD: "/projB"}, 10); len(hits) != 1 {
		t.Fatalf("projB content affected by projA reconcile: %d hits", len(hits))
	}
}
