package codex

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
)

func TestProviderAvailabilityWithRolloutsOnly(t *testing.T) {
	root := t.TempDir()
	t.Setenv("META_CC_CODEX_ROOT", root)
	rollout := filepath.Join(root, "sessions", "2026", "07", "29", "rollout-test.jsonl")
	if err := os.MkdirAll(filepath.Dir(rollout), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rollout, []byte(`{"timestamp":"2026-07-29T10:00:00Z","type":"session_meta","payload":{"id":"rollout-only","cwd":"/project","model":"gpt-5"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := NewProviderWithMode(locator.NewCodexLocator(), ModeFiles)
	if !p.IsAvailable(context.Background()) {
		t.Fatal("rollout-only Codex home should be available")
	}
	sessions, err := p.ListSessionsFiltered(context.Background(), conversation.SessionFilter{CWD: "/project"})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != "rollout-only" {
		t.Fatalf("unexpected rollout-only sessions: %#v", sessions)
	}
}

func TestRolloutFallbackArchivedAndCWDFiltering(t *testing.T) {
	root := t.TempDir()
	t.Setenv("META_CC_CODEX_ROOT", root)
	writeMeta := func(dir, id, cwd string) {
		path := filepath.Join(root, dir, id+".jsonl")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		line := fmt.Sprintf(`{"timestamp":"2026-07-29T10:00:00Z","type":"session_meta","payload":{"id":%q,"cwd":%q}}`+"\n", id, cwd)
		if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeMeta("sessions", "active-a", "/a")
	writeMeta("sessions", "active-b", "/b")
	writeMeta("archived_sessions", "archived-a", "/a")
	p := NewProviderWithMode(locator.NewCodexLocator(), ModeFiles)
	archived := true
	sessions, err := p.ListSessionsFiltered(context.Background(), conversation.SessionFilter{CWD: "/a", Archived: &archived})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != "archived-a" || !sessions[0].Archived {
		t.Fatalf("unexpected archived fallback result: %#v", sessions)
	}
}

func TestCorruptHighestDBFallsBackToCompatibleLowerDB(t *testing.T) {
	root := t.TempDir()
	t.Setenv("META_CC_CODEX_ROOT", root)
	if err := os.WriteFile(filepath.Join(root, "state_7.sqlite"), []byte("not a sqlite database"), 0o644); err != nil {
		t.Fatal(err)
	}
	good, err := sql.Open("sqlite", filepath.Join(root, "state_6.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := good.Exec(`CREATE TABLE threads (id TEXT, cwd TEXT, title TEXT, source TEXT, created_at INTEGER)`); err != nil {
		t.Fatal(err)
	}
	if _, err := good.Exec(`INSERT INTO threads VALUES ('lower', '/project', 'lower', 'cli', 1700000000)`); err != nil {
		t.Fatal(err)
	}
	good.Close()

	p := NewProviderWithMode(locator.NewCodexLocator(), ModeFiles)
	sessions, err := p.ListSessionsFiltered(context.Background(), conversation.SessionFilter{CWD: "/project"})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != "lower" {
		t.Fatalf("unexpected sessions: %#v", sessions)
	}
	warnings := p.Warnings()
	if len(warnings) != 1 || !strings.Contains(warnings[0], "state_7.sqlite") || !strings.Contains(warnings[0], "trying an older candidate") {
		t.Fatalf("expected one actionable corrupt-candidate warning, got %#v", warnings)
	}
}

func TestCorruptOnlyDBFallsBackToRollouts(t *testing.T) {
	root := t.TempDir()
	t.Setenv("META_CC_CODEX_ROOT", root)
	if err := os.WriteFile(filepath.Join(root, "state_9.sqlite"), []byte("not a sqlite database"), 0o644); err != nil {
		t.Fatal(err)
	}
	rollout := filepath.Join(root, "sessions", "rollout-safe.jsonl")
	if err := os.MkdirAll(filepath.Dir(rollout), 0o755); err != nil {
		t.Fatal(err)
	}
	meta := `{"timestamp":"2026-07-29T10:00:00Z","type":"session_meta","payload":{"id":"rollout-safe","cwd":"/safe"}}` + "\n"
	if err := os.WriteFile(rollout, []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewProviderWithMode(locator.NewCodexLocator(), ModeFiles)
	sessions, err := p.ListSessionsFiltered(context.Background(), conversation.SessionFilter{CWD: "/safe"})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != "rollout-safe" {
		t.Fatalf("unexpected rollout fallback sessions: %#v", sessions)
	}
	warnings := p.Warnings()
	if len(warnings) != 1 || !strings.Contains(warnings[0], "rollout fallback") || !strings.Contains(warnings[0], "state_9.sqlite") {
		t.Fatalf("expected bounded final-fallback warning, got %#v", warnings)
	}
}

func TestIncompatibleHighestDBFallsBackToCompatibleLowerDB(t *testing.T) {
	root := t.TempDir()
	t.Setenv("META_CC_CODEX_ROOT", root)
	bad, err := sql.Open("sqlite", filepath.Join(root, "state_6.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bad.Exec(`CREATE TABLE threads (id TEXT)`); err != nil {
		t.Fatal(err)
	}
	bad.Close()
	good, err := sql.Open("sqlite", filepath.Join(root, "state_5.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := good.Exec(`CREATE TABLE threads (id TEXT, cwd TEXT, title TEXT, source TEXT, created_at INTEGER)`); err != nil {
		t.Fatal(err)
	}
	if _, err := good.Exec(`INSERT INTO threads VALUES ('safe', '/project', 'safe', 'cli', 1700000000)`); err != nil {
		t.Fatal(err)
	}
	good.Close()
	p := NewProviderWithMode(locator.NewCodexLocator(), ModeFiles)
	sessions, err := p.ListSessionsFiltered(context.Background(), conversation.SessionFilter{CWD: "/project"})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != "safe" {
		t.Fatalf("unexpected sessions: %#v", sessions)
	}
}

func TestProviderAvailability(t *testing.T) {
	root := t.TempDir()
	loc := locator.NewCodexLocator()
	_ = loc

	t.Setenv("META_CC_CODEX_ROOT", root)
	p := NewProvider(locator.NewCodexLocator())
	if p.IsAvailable(context.Background()) {
		t.Fatalf("expected unavailable")
	}

	dbFile := filepath.Join(root, "state_5.sqlite")
	if err := os.WriteFile(dbFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !p.IsAvailable(context.Background()) {
		t.Fatalf("expected available")
	}
}
