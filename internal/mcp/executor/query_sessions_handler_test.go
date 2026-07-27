package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupClaudeSessionFixtureProject wires a temporary Claude projects root
// with one real session JSONL under the project's hashed directory, so the
// Claude provider's ListSessions/GetSession can read real metadata. Returns
// the resolved project path (to pass as working_dir) and the session ID.
func setupClaudeSessionFixtureProject(t *testing.T, title string) (projectPath, sessionID string) {
	t.Helper()

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	rawProjectPath := t.TempDir()
	absProject, err := filepath.Abs(rawProjectPath)
	require.NoError(t, err)
	resolvedProject, err := filepath.EvalSymlinks(absProject)
	require.NoError(t, err)

	hash := strings.ReplaceAll(resolvedProject, "\\", "-")
	hash = strings.ReplaceAll(hash, "/", "-")
	hash = strings.ReplaceAll(hash, ":", "-")
	sessionDir := filepath.Join(projectsRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	sessionID = "claude-session-1"
	entry := map[string]interface{}{
		"type":      "user",
		"timestamp": "2026-01-01T00:00:00Z",
		"sessionId": sessionID,
		"cwd":       resolvedProject,
		"message":   map[string]interface{}{"role": "user", "content": title},
	}
	line, err := json.Marshal(entry)
	require.NoError(t, err)

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, append(line, '\n'), 0o644))

	return resolvedProject, sessionID
}

// TestQuerySessions_Codex_MetadataOnlyDoesNotTouchRollout is the DIR-030
// "listing session metadata does not parse rollout contents" proof: it
// reuses setupCodexMultiSessionFixtureProject's "corrupt-session" thread
// (whose rollout_path points at a file that was never created) and proves
// query_sessions still returns metadata for it — something query_session_content
// cannot do (see TestQuerySessionContent_Codex_OneCorruptSessionDoesNotEraseOthers)
// because that path DOES need to open the rollout file.
func TestQuerySessions_Codex_MetadataOnlyDoesNotTouchRollout(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 2)

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 3, "expected all 3 threads (2 valid + 1 with a missing rollout file), since metadata listing never opens the rollout")

	var ids []string
	for _, e := range result.Entries {
		m := e.(map[string]interface{})
		ids = append(ids, m["session_id"].(string))
	}
	require.Contains(t, ids, "corrupt-session")
}

// TestQuerySessions_Codex_SessionIDFastPathIgnoresRolloutValidity proves the
// exact session_id fast path also never needs the rollout file to be valid.
func TestQuerySessions_Codex_SessionIDFastPathIgnoresRolloutValidity(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 1)

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
		"session_id":  "corrupt-session",
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)
	m := result.Entries[0].(map[string]interface{})
	require.Equal(t, "corrupt-session", m["session_id"])
}

// TestQuerySessions_Codex_SourceKindFilter proves source_kind filtering
// works (the fixture's threads all have source="cli").
func TestQuerySessions_Codex_SourceKindFilter(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 2)

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
		"source_kind": []interface{}{"cli"},
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 3)

	result, err = handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
		"source_kind": []interface{}{"vscode"},
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 0)
}

// TestQuerySessions_Codex_InvalidSourceKindFailsClosed proves an
// unrecognized source_kind value is an actionable validation error, not a
// silently-empty result.
func TestQuerySessions_Codex_InvalidSourceKindFailsClosed(t *testing.T) {
	_, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": t.TempDir(),
		"source_kind": []interface{}{"not-a-real-kind"},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not-a-real-kind")
}

// TestQuerySessions_ConflictingStatusAndArchivedFailsClosed proves a
// status/archived contradiction is rejected rather than silently resolved
// one way or the other.
func TestQuerySessions_ConflictingStatusAndArchivedFailsClosed(t *testing.T) {
	_, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": t.TempDir(),
		"status":      "active",
		"archived":    true,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "conflicting")
}

// TestQuerySessions_CodexOnlyFilterWithClaudeProviderFailsClosed proves
// that a Codex-only filter dimension combined with the (explicit or
// default) claude provider is rejected with an actionable error, since it
// can never match any Claude session — silently returning zero results
// would look identical to "no sessions matched" and mask the mistake.
func TestQuerySessions_CodexOnlyFilterWithClaudeProviderFailsClosed(t *testing.T) {
	_, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"working_dir": t.TempDir(),
		"source_kind": []interface{}{"cli"},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), `provider="codex"`)

	// Explicit provider="claude" is the same as the default and must also fail.
	_, err = handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "claude",
		"working_dir": t.TempDir(),
		"archived":    true,
	})
	require.Error(t, err)
}

// TestQuerySessions_Claude_DefaultBehaviorReturnsSessionMetadata is the
// DIR-030 "existing Claude-default schemas and results remain compatible"
// regression check for the new tool: with no provider argument (Claude
// default), query_sessions must still work and return metadata for a real
// Claude session fixture.
func TestQuerySessions_Claude_DefaultBehaviorReturnsSessionMetadata(t *testing.T) {
	projectPath, sessionID := setupClaudeSessionFixtureProject(t, "hello world")

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)

	m := result.Entries[0].(map[string]interface{})
	require.Equal(t, sessionID, m["session_id"])
	require.Equal(t, "claude", m["provider"])
	require.Equal(t, projectPath, m["cwd"])
}

// TestQuerySessions_Claude_ExactSessionID proves session_id also works on
// the Claude (default) path via GetSession.
func TestQuerySessions_Claude_ExactSessionID(t *testing.T) {
	projectPath, sessionID := setupClaudeSessionFixtureProject(t, "hello world")

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"working_dir": projectPath,
		"session_id":  sessionID,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)
	m := result.Entries[0].(map[string]interface{})
	require.Equal(t, sessionID, m["session_id"])
}

// TestQuerySessions_InvalidTimeValueFailsClosed proves an unparseable
// created_since/updated_since value is a validation error.
func TestQuerySessions_InvalidTimeValueFailsClosed(t *testing.T) {
	_, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"working_dir":   t.TempDir(),
		"created_since": "not-a-date",
	})
	require.Error(t, err)
}

// TestQuerySessions_InvalidProviderFailsClosed proves an unrecognized
// provider name is rejected (reusing rawfiles.ParseProviderFilter).
func TestQuerySessions_InvalidProviderFailsClosed(t *testing.T) {
	_, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "not-a-real-provider",
		"working_dir": t.TempDir(),
	})
	require.Error(t, err)
}
