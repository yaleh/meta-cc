package rawfiles

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// setupCodexHome creates a temporary Codex home directory with a state_5.sqlite
// database containing a single thread whose rollout_path points at a rollout
// file under the same temp dir. Returns the project (cwd) path recorded for
// the session.
func setupCodexHome(t *testing.T, codexHome, sessionID, cwd string) string {
	t.Helper()
	t.Setenv("META_CC_CODEX_ROOT", codexHome)

	rolloutPath := filepath.Join(codexHome, sessionID+".jsonl")
	content := `{"timestamp":"2026-06-14T06:00:00Z","type":"session_meta","payload":{"id":"` + sessionID + `","cwd":"` + cwd + `"}}
{"timestamp":"2026-06-14T06:00:01Z","type":"event_msg","payload":{"type":"user_message","message":"hello from codex"}}
`
	require.NoError(t, os.WriteFile(rolloutPath, []byte(content), 0o644))

	db, err := sql.Open("sqlite", filepath.Join(codexHome, "state_5.sqlite"))
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY,
		rollout_path TEXT,
		cwd TEXT,
		title TEXT,
		model TEXT,
		model_provider TEXT,
		tokens_used INTEGER,
		source TEXT,
		created_at INTEGER
	)`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sessionID, rolloutPath, cwd, "rawfiles test", "gpt-5", "openai", 42, "cli", int64(1700000000))
	require.NoError(t, err)

	return rolloutPath
}

// setupCodexHomeMultiProject creates a single Codex home (one state_5.sqlite,
// shared like a real Codex install) with one thread per synthetic session,
// each recording its own cwd and created_at. Returns each session's rollout
// path keyed by session ID.
func setupCodexHomeMultiProject(t *testing.T, codexHome string, sessions []struct {
	id        string
	cwd       string
	createdAt int64
}) map[string]string {
	t.Helper()
	t.Setenv("META_CC_CODEX_ROOT", codexHome)

	db, err := sql.Open("sqlite", filepath.Join(codexHome, "state_5.sqlite"))
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY,
		rollout_path TEXT,
		cwd TEXT,
		title TEXT,
		model TEXT,
		model_provider TEXT,
		tokens_used INTEGER,
		source TEXT,
		created_at INTEGER
	)`)
	require.NoError(t, err)

	rolloutPaths := make(map[string]string, len(sessions))
	for _, s := range sessions {
		rolloutPath := filepath.Join(codexHome, s.id+".jsonl")
		content := `{"timestamp":"2026-06-14T06:00:00Z","type":"session_meta","payload":{"id":"` + s.id + `","cwd":"` + s.cwd + `"}}
{"timestamp":"2026-06-14T06:00:01Z","type":"event_msg","payload":{"type":"user_message","message":"hello from codex"}}
`
		require.NoError(t, os.WriteFile(rolloutPath, []byte(content), 0o644))

		_, err = db.Exec(`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			s.id, rolloutPath, s.cwd, "rawfiles multi-project test", "gpt-5", "openai", 42, "cli", s.createdAt)
		require.NoError(t, err)

		rolloutPaths[s.id] = rolloutPath
	}
	return rolloutPaths
}

// TestSelectCodexFiles_SessionScope_DoesNotCrossProjectBoundary is a
// regression test for a bug where scope=="session" ignored projectPath
// entirely: it sorted ALL Codex sessions across every project by CreatedAt
// and returned only the single globally most-recent one. Concretely, with
// two synthetic Codex projects A and B sharing one Codex home, and B's
// session created more recently than A's, a session-scope lookup for
// project A must return ONLY project A's rollout file — never project B's,
// even though B's session would win a naive global "most recent" sort.
func TestSelectCodexFiles_SessionScope_DoesNotCrossProjectBoundary(t *testing.T) {
	codexHome := t.TempDir()
	projectA := t.TempDir()
	projectB := t.TempDir()

	rolloutPaths := setupCodexHomeMultiProject(t, codexHome, []struct {
		id        string
		cwd       string
		createdAt int64
	}{
		{id: "sess-a", cwd: projectA, createdAt: 1_700_000_000}, // older
		{id: "sess-b", cwd: projectB, createdAt: 1_800_000_000}, // newer: would win a global sort
	})

	registryA := NewRegistry(projectA)
	filesA, err := SelectCodexFiles(context.Background(), registryA, "session", projectA)
	require.NoError(t, err)
	require.Len(t, filesA, 1, "project A's session-scope lookup must resolve to exactly one session")
	require.Equal(t, "sess-a", filesA[0].SessionID, "session-scope lookup for project A must not return project B's more recent session")
	require.Equal(t, rolloutPaths["sess-a"], filesA[0].Path)

	// Symmetric check: project B's own session-scope lookup must still work
	// and resolve to its own session.
	registryB := NewRegistry(projectB)
	filesB, err := SelectCodexFiles(context.Background(), registryB, "session", projectB)
	require.NoError(t, err)
	require.Len(t, filesB, 1)
	require.Equal(t, "sess-b", filesB[0].SessionID)
	require.Equal(t, rolloutPaths["sess-b"], filesB[0].Path)
}

func TestParseProviderFilter(t *testing.T) {
	cases := []struct {
		name    string
		want    []conversation.ProviderID
		wantErr bool
	}{
		{name: "", want: []conversation.ProviderID{conversation.ProviderClaude}},
		{name: "claude", want: []conversation.ProviderID{conversation.ProviderClaude}},
		{name: "codex", want: []conversation.ProviderID{conversation.ProviderCodex}},
		{name: "all", want: []conversation.ProviderID{conversation.ProviderClaude, conversation.ProviderCodex}},
		{name: "bogus", wantErr: true},
	}
	for _, tc := range cases {
		got, err := ParseProviderFilter(tc.name)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseProviderFilter(%q): expected error, got nil", tc.name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseProviderFilter(%q): unexpected error: %v", tc.name, err)
		}
		if len(got) != len(tc.want) {
			t.Fatalf("ParseProviderFilter(%q) = %v, want %v", tc.name, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("ParseProviderFilter(%q)[%d] = %v, want %v", tc.name, i, got[i], tc.want[i])
			}
		}
	}
}

func TestSelectCodexFiles_ReturnsRolloutPathForMatchingProject(t *testing.T) {
	codexHome := t.TempDir()
	projectPath := t.TempDir()
	rolloutPath := setupCodexHome(t, codexHome, "sess-1", projectPath)

	registry := NewRegistry(projectPath)
	files, err := SelectCodexFiles(context.Background(), registry, "project", projectPath)
	require.NoError(t, err)
	require.Len(t, files, 1)
	require.Equal(t, rolloutPath, files[0].Path)
	require.Equal(t, conversation.ProviderCodex, files[0].Provider)
	require.Equal(t, "sess-1", files[0].SessionID)
}

func TestSelectCodexFiles_UnavailableProvider_FailsClosed(t *testing.T) {
	// No state_5.sqlite present at all.
	t.Setenv("META_CC_CODEX_ROOT", t.TempDir())

	projectPath := t.TempDir()
	registry := NewRegistry(projectPath)
	_, err := SelectCodexFiles(context.Background(), registry, "project", projectPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "codex provider unavailable")
}

func TestSelectCodexFiles_NoMatchingSessions_FailsClosed(t *testing.T) {
	codexHome := t.TempDir()
	otherProject := t.TempDir()
	setupCodexHome(t, codexHome, "sess-1", otherProject)

	// Query a different project path than the one recorded in the DB.
	queryProject := t.TempDir()
	registry := NewRegistry(queryProject)
	_, err := SelectCodexFiles(context.Background(), registry, "project", queryProject)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no codex sessions found")
}
