package ftsindex

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

func TestRebuild_RecoversFromCorruptionAndRepopulates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fts.db")
	now := time.Now()

	sess := session(conversation.ProviderClaude, "s1", "/proj", "s1", now)
	turns := []conversation.Turn{turn("t1", now, textItem("i1", conversation.ItemKindUserMessage, "user", "dolphins are clever", now))}
	loader := newCountingLoader(map[string][]conversation.Turn{sessionKey("claude", "s1"): turns})
	metaMap := map[string]SourceMeta{sessionKey("claude", "s1"): {Path: "/fake/s1.jsonl", Size: 1, ModTime: now}}

	// Corrupt the index file before it's ever been created by this test.
	if err := os.WriteFile(path, []byte("garbage, not a real sqlite db"), 0o644); err != nil {
		t.Fatalf("write garbage: %v", err)
	}

	stats, warnings, err := Rebuild(context.Background(), path, []conversation.Session{sess}, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes)
	if err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if stats.SessionsIndexed != 1 {
		t.Fatalf("stats = %+v, want SessionsIndexed=1", stats)
	}

	db, degraded, err := Open(context.Background(), path)
	if err != nil || degraded {
		t.Fatalf("reopening rebuilt index: degraded=%v err=%v", degraded, err)
	}
	defer db.Close()

	hits, err := Search(context.Background(), db, "dolphins", SearchFilter{CWD: "/proj"}, 10)
	if err != nil || len(hits) != 1 {
		t.Fatalf("Search after Rebuild: hits=%v err=%v", hits, err)
	}
}

func TestClean_RemovesIndexFile(t *testing.T) {
	db, path := openTestDB(t)
	db.Close()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("index file missing before Clean: %v", err)
	}
	if err := Clean(path); err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("index file still present after Clean: err=%v", err)
	}
}

func TestInspect_ReportsCounts(t *testing.T) {
	db, _ := openTestDB(t)
	ctx := context.Background()
	now := time.Now()

	sess := session(conversation.ProviderClaude, "s1", "/proj", "s1", now)
	turns := []conversation.Turn{
		turn("t1", now,
			textItem("i1", conversation.ItemKindUserMessage, "user", "one", now),
			textItem("i2", conversation.ItemKindAgentMessage, "assistant", "two", now),
		),
	}
	loader := newCountingLoader(map[string][]conversation.Turn{sessionKey("claude", "s1"): turns})
	metaMap := map[string]SourceMeta{sessionKey("claude", "s1"): {Path: "/fake/s1.jsonl", Size: 1, ModTime: now}}
	if _, _, err := Refresh(ctx, db, []conversation.Session{sess}, staticSourceMeta(metaMap), loader.Load, DefaultBodyLimitBytes); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	result, err := Inspect(ctx, db)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if result.SchemaVersion != SchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", result.SchemaVersion, SchemaVersion)
	}
	if result.SessionCount != 1 {
		t.Fatalf("SessionCount = %d, want 1", result.SessionCount)
	}
	if result.ItemCount != 2 {
		t.Fatalf("ItemCount = %d, want 2", result.ItemCount)
	}
}

func TestIsDisabled(t *testing.T) {
	t.Setenv(disableEnvVar, "")
	if IsDisabled() {
		t.Fatalf("IsDisabled() = true with unset env var")
	}
	t.Setenv(disableEnvVar, "1")
	if !IsDisabled() {
		t.Fatalf("IsDisabled() = false with %s=1", disableEnvVar)
	}
	t.Setenv(disableEnvVar, "false")
	if IsDisabled() {
		t.Fatalf("IsDisabled() = true with %s=false", disableEnvVar)
	}
}

func TestDefaultPath_AndEnsureDir(t *testing.T) {
	proj := t.TempDir()
	path := DefaultPath(proj)
	want := filepath.Join(proj, ".meta-cc", "index", "fts.db")
	if path != want {
		t.Fatalf("DefaultPath = %q, want %q", path, want)
	}
	if err := EnsureDir(path); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	gitignore := filepath.Join(proj, ".meta-cc", "index", ".gitignore")
	if _, err := os.Stat(gitignore); err != nil {
		t.Fatalf(".gitignore not created: %v", err)
	}
}
