package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestQuerySessionContent_Codex_LifecycleOnlyTurnObservableEndToEnd is the
// DIR-053 end-to-end regression test. DIR-050 made the Codex rollout parser
// RETAIN a lifecycle-only turn (session_end / turn_aborted / compaction) in the
// []conversation.Turn slice, but providerrecords.Normalize historically read
// only UserText/AssistantText/ToolCalls/TokenUsage, so such a turn still
// produced ZERO query-time records and stayed invisible to every MCP tool —
// the exact "session looks like it has no history" symptom DIR-050's Finding
// described. This test exercises the FULL path a real MCP client uses —
// query_session_content(role="all", provider="codex") ->
// handleQuerySessionContent -> handleQueryConversationFlow ->
// dispatchProviderQuery -> providerrecords.Build/Normalize -> jq filtering —
// against a fixture rollout ending in a bare session_end event, and asserts the
// lifecycle record actually surfaces in the tool response. It is deliberately
// DISTINCT from DIR-050's internal/provider/codex/rollout_test.go tests (which
// stop at conversation.Turn) and from the providerrecords-layer
// records_lifecycle_test.go (which stops at Normalize output): this one proves
// the signal survives all the way to an MCP tool response.
func TestQuerySessionContent_Codex_LifecycleOnlyTurnObservableEndToEnd(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-lifecycle-only-sample.jsonl")

	e := NewToolExecutor()
	result, err := handleQuerySessionContent(e, "project", map[string]interface{}{
		"role":        "all",
		"provider":    "codex",
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.Entries,
		"a lifecycle-only session must surface at least one record through role=all, not look empty")

	// Find the lifecycle record and verify its shape.
	var found bool
	for _, entry := range result.Entries {
		m, ok := entry.(map[string]interface{})
		require.True(t, ok, "unexpected entry shape: %#v", entry)
		if m["type"] != "session_end" {
			continue
		}
		found = true
		require.Equal(t, "completed", m["reason"], "session_end record must carry the lifecycle reason")
		// setupCodexRolloutFixtureProject registers the thread as
		// "codex-dedup-session"; the lifecycle record must carry that identity.
		require.Equal(t, "codex-dedup-session", m["session_id"], "record must carry session identity")
		require.NotNil(t, m["timestamp"], "record must carry a timestamp")
	}
	require.True(t, found,
		"role=all must surface the session_end lifecycle record; got entries: %#v", result.Entries)
}

// TestQuerySessionSignals_Codex_LifecycleOnlyTurnObservableViaTimestamps is a
// companion end-to-end check proving the lifecycle record is ALSO surfaced by a
// second, independent MCP tool whose existing filter already matches any record
// carrying a timestamp — query_session_signals(type="timestamps"). This shows
// the widened Normalize output is observable through more than one tool surface
// without any further handler change.
func TestQuerySessionSignals_Codex_LifecycleOnlyTurnObservableViaTimestamps(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-lifecycle-only-sample.jsonl")

	e := NewToolExecutor()
	result, err := handleQuerySessionSignals(e, "project", map[string]interface{}{
		"type":        "timestamps",
		"provider":    "codex",
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.Entries,
		"a lifecycle-only session must surface at least one timestamped record, not look empty")
}

// TestQuerySessionContent_Claude_UserOnlyTurnHasNoSpuriousLifecycleRecord
// exercises the Claude provider through the MCP path. Claude marks a user-only
// turn InProgress while awaiting an assistant; role=all must expose the user
// message only, not invent a lifecycle record from that bookkeeping status.
func TestQuerySessionContent_Claude_UserOnlyTurnHasNoSpuriousLifecycleRecord(t *testing.T) {
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	projectPath, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	projectHash := strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(projectPath)
	sessionDir := filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	entry := map[string]interface{}{
		"type": "user", "timestamp": "2026-06-14T06:00:00Z",
		"sessionId": "claude-in-progress", "cwd": projectPath,
		"message": map[string]interface{}{"role": "user", "content": "waiting for assistant"},
	}
	line, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "claude-in-progress.jsonl"), append(line, '\n'), 0o644))

	result, err := handleQuerySessionContent(NewToolExecutor(), "project", map[string]interface{}{
		"role": "all", "provider": "claude", "working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)
	record := result.Entries[0].(map[string]interface{})
	require.Equal(t, "user", record["type"])
}
