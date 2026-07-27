package ftsindex

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

func TestItemBody_TruncatesOversizedText(t *testing.T) {
	big := strings.Repeat("x", DefaultBodyLimitBytes+500)
	item := conversation.Item{Kind: conversation.ItemKindToolResult, Output: big}

	body, truncated := itemBody(item, DefaultBodyLimitBytes)
	if !truncated {
		t.Fatalf("expected truncated=true for oversized output")
	}
	if len(body) != DefaultBodyLimitBytes {
		t.Fatalf("body length = %d, want %d", len(body), DefaultBodyLimitBytes)
	}
}

func TestSearch_OversizedToolOutputNotFullyIndexed(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	// Construct a tool_result whose Output is bigger than the privacy cap:
	// a distinctive marker sits WITHIN the cap (should be findable) and
	// another sits well BEYOND the cap (must never be indexed/searchable).
	padding := strings.Repeat("z", DefaultBodyLimitBytes-100)
	output := "insidecapmarker " + padding + " " + strings.Repeat("y", 2000) + " beyondcapmarker"
	if len(output) <= DefaultBodyLimitBytes {
		t.Fatalf("test fixture sanity: output length %d must exceed cap %d", len(output), DefaultBodyLimitBytes)
	}

	sess := session(conversation.ProviderClaude, "s1", "/proj", "big output session", now)
	turns := []conversation.Turn{
		turn("t1", now, toolResultItem("i1", "Bash", output, now)),
	}
	loader := newCountingLoader(map[string][]conversation.Turn{sessionKey("claude", "s1"): turns})
	metaMap := map[string]SourceMeta{sessionKey("claude", "s1"): {Path: "/fake/s1.jsonl", Size: 1, ModTime: now}}

	if _, _, err := Refresh(ctx, db, []conversation.Session{sess}, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	if hits, err := Search(ctx, db, "insidecapmarker", SearchFilter{CWD: "/proj"}, 10); err != nil || len(hits) != 1 {
		t.Fatalf("in-cap marker search: hits=%v err=%v, want 1 hit", hits, err)
	} else if !hits[0].Truncated {
		t.Fatalf("hit for the oversized tool output does not report Truncated=true")
	}

	if hits, err := Search(ctx, db, "beyondcapmarker", SearchFilter{CWD: "/proj"}, 10); err != nil {
		t.Fatalf("beyond-cap marker search error: %v", err)
	} else if len(hits) != 0 {
		t.Fatalf("beyond-cap marker IS searchable (%d hits): oversized tool output was fully indexed, violating the privacy cap", len(hits))
	}

	var body string
	if err := db.QueryRow(`SELECT body FROM items WHERE session_key = ?`, sessionKey("claude", "s1")).Scan(&body); err != nil {
		t.Fatalf("read stored body: %v", err)
	}
	if len(body) > DefaultBodyLimitBytes {
		t.Fatalf("stored body length %d exceeds cap %d", len(body), DefaultBodyLimitBytes)
	}
	if strings.Contains(body, "beyondcapmarker") {
		t.Fatalf("stored body contains the beyond-cap marker: full oversized output was persisted")
	}
}
