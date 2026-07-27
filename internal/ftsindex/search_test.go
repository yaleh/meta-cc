package ftsindex

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// corpus builds a small multi-project, multi-session fixture used by
// several tests below: two sessions in /projA (one Claude, one Codex) and
// one session in /projB, each with a distinctive word in its content.
func corpus(now time.Time) ([]conversation.Session, map[string][]conversation.Turn, map[string]SourceMeta) {
	sessions := []conversation.Session{
		session(conversation.ProviderClaude, "a1", "/projA", "session a1", now),
		session(conversation.ProviderCodex, "a2", "/projA", "session a2", now),
		session(conversation.ProviderClaude, "b1", "/projB", "session b1", now),
	}
	turns := map[string][]conversation.Turn{
		sessionKey("claude", "a1"): {
			turn("t1", now,
				textItem("i1", conversation.ItemKindUserMessage, "user", "let's talk about giraffes today", now),
				textItem("i2", conversation.ItemKindAgentMessage, "assistant", "giraffes are tall mammals", now),
			),
		},
		sessionKey("codex", "a2"): {
			turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "unrelated octopus content", now)),
		},
		sessionKey("claude", "b1"): {
			turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "giraffes also live in project b", now)),
		},
	}
	meta := map[string]SourceMeta{
		sessionKey("claude", "a1"): {Path: "/fake/a1.jsonl", Size: 1, ModTime: now},
		sessionKey("codex", "a2"):  {Path: "/fake/a2.rollout", Size: 1, ModTime: now},
		sessionKey("claude", "b1"): {Path: "/fake/b1.jsonl", Size: 1, ModTime: now},
	}
	return sessions, turns, meta
}

// canonicalScan independently reproduces what Search should return by
// scanning the raw canonical turns/items directly (no index involved),
// applying the same predicate (substring match, cwd-scoped) Search's FTS
// query is expected to satisfy.
func canonicalScan(sessions []conversation.Session, turnsByKey map[string][]conversation.Turn, cwd, word string) []string {
	var got []string
	for _, s := range sessions {
		if s.CWD != cwd {
			continue
		}
		for _, tn := range turnsByKey[sessionKey(string(s.Provider), s.ID)] {
			for _, item := range tn.Items {
				if strings.Contains(strings.ToLower(rawSearchableText(item)), word) {
					got = append(got, s.ID+"/"+tn.ID+"/"+item.ID)
				}
			}
		}
	}
	sort.Strings(got)
	return got
}

func TestSearch_MatchesCanonicalScan(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	sessions, turnsMap, metaMap := corpus(now)
	loader := newCountingLoader(turnsMap)
	if _, _, err := Refresh(ctx, db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	hits, err := Search(ctx, db, "giraffes", SearchFilter{CWD: "/projA"}, 50)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	var got []string
	for _, h := range hits {
		got = append(got, h.ThreadID+"/"+h.TurnID+"/"+h.ItemID)
	}
	sort.Strings(got)

	want := canonicalScan(sessions, turnsMap, "/projA", "giraffes")
	if len(want) != 2 {
		t.Fatalf("test fixture sanity: canonical scan found %d giraffe items in /projA, want 2", len(want))
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("Search results %v do not match canonical scan %v", got, want)
	}
}

func TestSearch_RespectsCWDBoundary(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	sessions, turnsMap, metaMap := corpus(now)
	loader := newCountingLoader(turnsMap)
	if _, _, err := Refresh(ctx, db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	// "giraffes" appears in BOTH /projA (session a1) and /projB (session
	// b1). A search scoped to /projA must never leak the /projB hit —
	// this is the DIR-030-precedent cross-project boundary check.
	hits, err := Search(ctx, db, "giraffes", SearchFilter{CWD: "/projA"}, 50)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, h := range hits {
		if h.CWD != "/projA" {
			t.Fatalf("Search scoped to /projA returned a hit from cwd %q (session %s): cross-project leak", h.CWD, h.ThreadID)
		}
		if h.ThreadID == "b1" {
			t.Fatalf("Search scoped to /projA returned projB's session b1: cross-project leak")
		}
	}

	hitsB, err := Search(ctx, db, "giraffes", SearchFilter{CWD: "/projB"}, 50)
	if err != nil {
		t.Fatalf("Search projB: %v", err)
	}
	if len(hitsB) != 1 || hitsB[0].ThreadID != "b1" {
		t.Fatalf("Search scoped to /projB = %+v, want exactly session b1", hitsB)
	}
}

func TestSearch_RequiresCWD(t *testing.T) {
	db, _ := openTestDB(t)
	if _, err := Search(context.Background(), db, "anything", SearchFilter{}, 10); err == nil {
		t.Fatalf("Search with an empty CWD filter did not error (would search across every project)")
	}
}

func TestSearch_ProviderFilter(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	sessions, turnsMap, metaMap := corpus(now)
	loader := newCountingLoader(turnsMap)
	if _, _, err := Refresh(ctx, db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	hits, err := Search(ctx, db, "giraffes", SearchFilter{CWD: "/projA", Provider: conversation.ProviderCodex}, 50)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 0 {
		t.Fatalf("provider=codex filter returned %d hits for content only in the claude session", len(hits))
	}
}

func TestHydrate_ReturnsCanonicalItem(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	sessions, turnsMap, metaMap := corpus(now)
	loader := newCountingLoader(turnsMap)
	if _, _, err := Refresh(ctx, db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	hits, err := Search(ctx, db, "octopus", SearchFilter{CWD: "/projA"}, 10)
	if err != nil || len(hits) != 1 {
		t.Fatalf("Search: hits=%v err=%v", hits, err)
	}
	hit := hits[0]

	tn, item, ok := Hydrate(turnsMap[sessionKey(string(hit.Provider), hit.ThreadID)], hit)
	if !ok {
		t.Fatalf("Hydrate failed to resolve hit %+v against canonical turns", hit)
	}
	if tn.ID != "t1" || item.Text != "unrelated octopus content" {
		t.Fatalf("Hydrate returned unexpected turn/item: turn=%+v item=%+v", tn, item)
	}
}
