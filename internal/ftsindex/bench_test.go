package ftsindex

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// syntheticCorpus builds n sessions of m items each, all under one cwd, for
// benchmarking. Every item's text embeds its session/item index so
// BenchmarkSearch can target a real, sparse hit.
func syntheticCorpus(n, m int) ([]conversation.Session, map[string][]conversation.Turn, map[string]SourceMeta) {
	now := time.Now()
	sessions := make([]conversation.Session, 0, n)
	turnsMap := make(map[string][]conversation.Turn, n)
	metaMap := make(map[string]SourceMeta, n)

	for i := 0; i < n; i++ {
		id := fmt.Sprintf("bench-session-%d", i)
		sessions = append(sessions, session(conversation.ProviderClaude, id, "/bench-proj", "bench session", now))

		items := make([]conversation.Item, 0, m)
		for j := 0; j < m; j++ {
			text := fmt.Sprintf("synthetic benchmark content session=%d item=%d filler words about oceans and mountains", i, j)
			items = append(items, textItem(fmt.Sprintf("i%d", j), conversation.ItemKindUserMessage, "user", text, now))
		}
		key := sessionKey("claude", id)
		turnsMap[key] = []conversation.Turn{turn("t1", now, items...)}
		metaMap[key] = SourceMeta{Path: "/fake/" + id + ".jsonl", Size: int64(m * 100), ModTime: now}
	}
	return sessions, turnsMap, metaMap
}

// BenchmarkColdIndex measures a full from-scratch index build over a
// synthetic corpus (every session is new, so every one gets fully parsed
// and inserted).
func BenchmarkColdIndex(b *testing.B) {
	sessions, turnsMap, metaMap := syntheticCorpus(200, 20)
	loader := newCountingLoader(turnsMap)

	for i := 0; i < b.N; i++ {
		path := filepath.Join(b.TempDir(), "fts.db")
		db, _, err := Open(context.Background(), path)
		if err != nil {
			b.Fatalf("Open: %v", err)
		}
		if _, _, err := Refresh(context.Background(), db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
			b.Fatalf("Refresh: %v", err)
		}
		db.Close()
	}
}

// BenchmarkIncrementalRefresh measures a Refresh call over a corpus that is
// already fully indexed and completely unchanged: this should be dominated
// by the (cheap) per-session SourceMeta comparison, not by reparsing, and
// should be dramatically faster per-session than BenchmarkColdIndex.
func BenchmarkIncrementalRefresh(b *testing.B) {
	sessions, turnsMap, metaMap := syntheticCorpus(200, 20)
	loader := newCountingLoader(turnsMap)

	path := filepath.Join(b.TempDir(), "fts.db")
	db, _, err := Open(context.Background(), path)
	if err != nil {
		b.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if _, _, err := Refresh(context.Background(), db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		b.Fatalf("initial Refresh: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := Refresh(context.Background(), db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
			b.Fatalf("Refresh: %v", err)
		}
	}
}

// BenchmarkSearch measures representative search latency against a fully
// indexed synthetic corpus.
func BenchmarkSearch(b *testing.B) {
	sessions, turnsMap, metaMap := syntheticCorpus(200, 20)
	loader := newCountingLoader(turnsMap)

	path := filepath.Join(b.TempDir(), "fts.db")
	db, _, err := Open(context.Background(), path)
	if err != nil {
		b.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if _, _, err := Refresh(context.Background(), db, sessions, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		b.Fatalf("Refresh: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Search(context.Background(), db, "oceans mountains", SearchFilter{CWD: "/bench-proj"}, 20); err != nil {
			b.Fatalf("Search: %v", err)
		}
	}
}
