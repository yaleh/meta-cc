package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// DIR-043: query_session_signals(type="tool_stats")'s "status" parameter was
// declared in the MCP tool schema ("filter by status (error/success)") but
// handleQueryTools never read args["status"] at all — status="error",
// status="success", and status="bogus-value-xyz" all returned the same
// unfiltered result. These tests exercise handleQueryTools directly against a
// fixture session containing both successful and failed tool calls, proving
// status="error" and status="success" now produce different, correctly
// scoped record sets, and that an unsupported status value fails closed with
// an error rather than silently passing through.

// setupToolStatsStatusFixtureProject wires a temporary Claude projects root
// with a single session JSONL containing three tool calls: two Bash
// tool_use/tool_result pairs that ended in error (is_error=true) and one
// Read tool_use/tool_result pair that succeeded (is_error=false).
func setupToolStatsStatusFixtureProject(t *testing.T) (projectPath string) {
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

	hash := strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolvedProject)
	sessionDir := filepath.Join(projectsRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	sessionID := "tool-stats-status-fixture-session"
	lines := []string{
		// Failed tool call #1 (Bash).
		`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_err_1","name":"Bash","input":{"command":"false"}}]}}`,
		`{"type":"user","timestamp":"2026-01-01T00:00:01Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_err_1","is_error":true,"content":"command failed"}]}}`,
		// Failed tool call #2 (Bash).
		`{"type":"assistant","timestamp":"2026-01-01T00:00:02Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_err_2","name":"Bash","input":{"command":"exit 1"}}]}}`,
		`{"type":"user","timestamp":"2026-01-01T00:00:03Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_err_2","is_error":true,"content":"command failed"}]}}`,
		// Successful tool call (Read).
		`{"type":"assistant","timestamp":"2026-01-01T00:00:04Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_ok_1","name":"Read","input":{"file_path":"/tmp/foo"}}]}}`,
		`{"type":"user","timestamp":"2026-01-01T00:00:05Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_ok_1","is_error":false,"content":"file contents"}]}}`,
	}

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644))

	return resolvedProject
}

// TestHandleQueryTools_StatusFilter_ErrorVsSuccessVsOmitted proves that
// status="error", status="success", and omitted status produce different,
// correctly-scoped record sets against a fixture with both outcomes present.
func TestHandleQueryTools_StatusFilter_ErrorVsSuccessVsOmitted(t *testing.T) {
	projectPath := setupToolStatsStatusFixtureProject(t)
	e := NewToolExecutor()

	unfiltered, err := handleQueryTools(e, "project", map[string]interface{}{
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Len(t, unfiltered.Entries, 3, "unfiltered tool_stats must return all 3 tool_use records")

	errResult, err := handleQueryTools(e, "project", map[string]interface{}{
		"working_dir": projectPath,
		"status":      "error",
	})
	require.NoError(t, err)
	require.Len(t, errResult.Entries, 2, "status=error must return exactly the 2 failed tool calls")
	for _, entry := range errResult.Entries {
		require.True(t, entryHasToolUseID(t, entry, "toolu_err_1") || entryHasToolUseID(t, entry, "toolu_err_2"),
			"status=error entry must be one of the failed tool calls, got: %#v", entry)
	}

	successResult, err := handleQueryTools(e, "project", map[string]interface{}{
		"working_dir": projectPath,
		"status":      "success",
	})
	require.NoError(t, err)
	require.Len(t, successResult.Entries, 1, "status=success must return exactly the 1 successful tool call")
	require.True(t, entryHasToolUseID(t, successResult.Entries[0], "toolu_ok_1"),
		"status=success entry must be the successful tool call, got: %#v", successResult.Entries[0])

	// The three results must be measurably different from each other, and
	// error+success must not double-count or exceed the unfiltered total.
	require.NotEqual(t, len(unfiltered.Entries), len(errResult.Entries))
	require.NotEqual(t, len(unfiltered.Entries), len(successResult.Entries))
	require.LessOrEqual(t, len(errResult.Entries)+len(successResult.Entries), len(unfiltered.Entries))
}

// TestHandleQueryTools_StatusFilter_InvalidValueFailsClosed proves that an
// unsupported status value returns an error instead of silently passing
// through the unfiltered result (the DIR-030 fail-closed contract).
func TestHandleQueryTools_StatusFilter_InvalidValueFailsClosed(t *testing.T) {
	projectPath := setupToolStatsStatusFixtureProject(t)
	e := NewToolExecutor()

	_, err := handleQueryTools(e, "project", map[string]interface{}{
		"working_dir": projectPath,
		"status":      "bogus-value-xyz",
	})
	require.Error(t, err, "status=bogus-value-xyz must fail closed with an error, not silently pass through")
}

// entryHasToolUseID checks whether a tool_stats result entry (a raw
// assistant record) contains a tool_use block with the given id.
func entryHasToolUseID(t *testing.T, entry interface{}, id string) bool {
	t.Helper()
	rec, ok := entry.(map[string]interface{})
	if !ok {
		return false
	}
	message, _ := rec["message"].(map[string]interface{})
	content, _ := message["content"].([]interface{})
	for _, block := range content {
		bm, ok := block.(map[string]interface{})
		if !ok {
			continue
		}
		if bm["type"] == "tool_use" && bm["id"] == id {
			return true
		}
	}
	return false
}
