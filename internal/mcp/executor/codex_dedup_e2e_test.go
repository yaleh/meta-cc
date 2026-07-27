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

// setupCodexRolloutFixtureProject wires a temporary Codex home (a
// state_5.sqlite pointing at a single thread) backed by the given rollout
// fixture under tests/fixtures/codex/. It rewrites the fixture's
// placeholder cwd ("/tmp/project") to a real per-test directory so that
// scope/project filtering (FilterSessionsForScope) matches, mirroring the
// pattern used by internal/analysis/service_test.go's
// setupCodexProviderProject and internal/provider/rawfiles/rawfiles_test.go's
// setupCodexHome. Returns the resolved project path to pass as working_dir.
func setupCodexRolloutFixtureProject(t *testing.T, fixtureName string) string {
	t.Helper()

	projectPath := t.TempDir()
	resolvedProject, err := filepath.EvalSymlinks(projectPath)
	require.NoError(t, err)

	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("META_CC_CODEX_ROOT", codexHome)
	require.NoError(t, os.MkdirAll(codexHome, 0o755))

	rolloutPath := filepath.Join(codexHome, "rollout.jsonl")
	fixture, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", fixtureName))
	require.NoError(t, err)
	fixture = []byte(strings.ReplaceAll(string(fixture), "/tmp/project", resolvedProject))
	require.NoError(t, os.WriteFile(rolloutPath, fixture, 0o644))

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
		"codex-dedup-session", rolloutPath, resolvedProject, "dedup e2e test", "gpt-5", "openai", 0, "cli", int64(1700000000))
	require.NoError(t, err)

	return resolvedProject
}

// TestQuerySessionContent_Codex_DedupesAssistantSegmentsEndToEnd is an
// end-to-end regression test for DIR-027 (duplicate Codex transcript
// messages): it exercises the full path a real MCP client uses —
// query_session_content(role=assistant, provider=codex) ->
// handleQuerySessionContent -> ExecuteQueryForProvider ->
// providerrecords.Build/Normalize -> jq filtering — rather than testing
// the rollout parser in isolation. This proves the fix is visible at the
// user-facing query surface, not just internally in conversation.Turn.
func TestQuerySessionContent_Codex_DedupesAssistantSegmentsEndToEnd(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-legacy-dedup-sample.jsonl")

	e := NewToolExecutor()
	result, err := handleQuerySessionContent(e, "project", map[string]interface{}{
		"role":        "assistant",
		"provider":    "codex",
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 3, "expected one assistant record per turn (turn-1, turn-2, turn-3), not doubled")

	var texts []string
	for _, entry := range result.Entries {
		m, ok := entry.(map[string]interface{})
		require.True(t, ok, "unexpected entry shape: %#v", entry)
		message, ok := m["message"].(map[string]interface{})
		require.True(t, ok, "missing message: %#v", m)
		content, ok := message["content"].([]interface{})
		require.True(t, ok, "missing content: %#v", message)

		var textBlocks []string
		for _, block := range content {
			b, ok := block.(map[string]interface{})
			require.True(t, ok)
			if b["type"] == "text" {
				textBlocks = append(textBlocks, b["text"].(string))
			}
		}
		require.Len(t, textBlocks, 1, "assistant record must carry exactly one text block per turn, not a duplicated pair: %#v", content)
		texts = append(texts, textBlocks[0])
	}

	require.Equal(t, []string{"Summary: done.", "Summary: done.", "Starting task\nTask complete"}, texts,
		"assistant segments duplicated across event_msg and response_item must collapse to one copy per turn/position, "+
			"while legitimately repeated text across distinct turns (turn-1 vs turn-2) and distinct segments "+
			"within the same turn (turn-3) must be preserved")
}

// TestQuerySessionContent_Codex_EventOnlyLegacyFixtureStillWorks is a
// companion end-to-end check that the pre-existing event_msg-only Codex
// fixture (no response_item messages at all) still surfaces user text
// through the same query surface after the dedup/precedence change —
// i.e. the event_msg fallback path is not broken by preferring
// response_item.
func TestQuerySessionContent_Codex_EventOnlyLegacyFixtureStillWorks(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-legacy-sample.jsonl")

	e := NewToolExecutor()
	result, err := handleQuerySessionContent(e, "project", map[string]interface{}{
		"role":        "user",
		"provider":    "codex",
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)

	m := result.Entries[0].(map[string]interface{})
	message := m["message"].(map[string]interface{})
	require.Equal(t, "show history", message["content"])
}

// TestQuerySessionContent_Codex_ResponseItemOnlyFixtureStillWorks is a
// companion end-to-end check that the pre-existing response_item-only rich
// Codex fixture (no event_msg agent_message/user_message at all) still
// surfaces user/assistant text correctly through the query surface.
func TestQuerySessionContent_Codex_ResponseItemOnlyFixtureStillWorks(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-legacy-rich-sample.jsonl")

	e := NewToolExecutor()
	result, err := handleQuerySessionContent(e, "project", map[string]interface{}{
		"role":        "assistant",
		"provider":    "codex",
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)

	m := result.Entries[0].(map[string]interface{})
	message := m["message"].(map[string]interface{})
	content := message["content"].([]interface{})
	textBlock := content[0].(map[string]interface{})
	require.Equal(t, "running tools", textBlock["text"])
}
