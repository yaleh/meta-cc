package executor

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"
)

// setupCodexMultiSessionFixtureProject wires a temporary Codex home whose
// state_5.sqlite threads table has goodCount valid sessions (each backed
// by the "rollout-legacy-sample.jsonl" fixture, rewritten to a real
// per-test cwd) plus one "corrupt" session whose rollout_path points at a
// file that does not exist on disk — simulating an unreadable/corrupt
// rollout without needing to hand-craft malformed JSONL. Returns the
// resolved project path to pass as working_dir.
func setupCodexMultiSessionFixtureProject(t *testing.T, goodCount int) string {
	t.Helper()
	// Pin the files backend so a real `codex app-server` never shadows the
	// hermetic fixture corpus (see tests/e2e/codex-e2e.sh).
	t.Setenv("META_CC_CODEX_BACKEND", "files")

	projectPath := t.TempDir()
	resolvedProject, err := filepath.EvalSymlinks(projectPath)
	require.NoError(t, err)

	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("META_CC_CODEX_ROOT", codexHome)
	require.NoError(t, os.MkdirAll(codexHome, 0o755))

	fixture, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-sample.jsonl"))
	require.NoError(t, err)
	fixture = []byte(strings.ReplaceAll(string(fixture), "/tmp/project", resolvedProject))

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

	for i := 0; i < goodCount; i++ {
		rolloutPath := filepath.Join(codexHome, "rollout-good-"+string(rune('a'+i))+".jsonl")
		require.NoError(t, os.WriteFile(rolloutPath, fixture, 0o644))
		_, err = db.Exec(`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"good-session-"+string(rune('a'+i)), rolloutPath, resolvedProject, "good", "gpt-5", "openai", 0, "cli", int64(1700000000+i))
		require.NoError(t, err)
	}

	// The corrupt session: rollout_path points at a file that was never
	// created, so LoadTurns (os.Open) fails for it specifically.
	_, err = db.Exec(`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"corrupt-session", filepath.Join(codexHome, "rollout-missing.jsonl"), resolvedProject, "corrupt", "gpt-5", "openai", 0, "cli", int64(1699999999))
	require.NoError(t, err)

	return resolvedProject
}

// TestQuerySessionContent_Codex_OneCorruptSessionDoesNotEraseOthers is the
// DIR-030 fixture-based proof of the "one corrupt rollout does not erase
// valid results from other sessions" contract: a project-scope query
// across 2 valid Codex sessions plus 1 session whose rollout file is
// missing must still return the 2 valid sessions' results, plus exactly
// one warning naming the corrupt session.
func TestQuerySessionContent_Codex_OneCorruptSessionDoesNotEraseOthers(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 2)

	e := NewToolExecutor()
	result, err := handleQuerySessionContent(e, "project", map[string]interface{}{
		"role":        "user",
		"provider":    "codex",
		"working_dir": projectPath,
	})
	require.NoError(t, err, "a corrupt session must not abort the whole project query")
	require.Len(t, result.Entries, 2, "expected one user-message record per valid session (2), corrupt session excluded")

	require.Len(t, result.Warnings, 1, "expected exactly one warning for the corrupt session")
	require.Contains(t, result.Warnings[0], "corrupt-session")
}
