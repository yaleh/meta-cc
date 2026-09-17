package executor

// DIR-094: query_sessions is the path the 2026-07-30 dogfooding failure
// actually hit — a session file Claude Code had created but never written a
// message into (8eda8f4e-…jsonl) made the whole listing hard-fail, so EVERY
// other session disappeared with it. abc8135 made ListSessions skip such a
// file instead; DIR-094's job here is the other half of the DIR-018 contract:
// the skip must be REPORTED, not silent, and it must survive all the way to
// the tool response a caller actually sees.
//
// These tests run the whole response pipeline (ExecuteTool, not just
// handleQuerySessions) so "the response metadata carries the exclusion" is
// asserted against the serialized output rather than an intermediate struct.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/locator"
)

// dir094EmptySessionID is the exact session file name from the dogfooding
// failure, so the fixture reproduces the reported shape verbatim.
const dir094EmptySessionID = "8eda8f4e-2c74-4176-ba6b-8c45e890df42"

// setupClaudeCorpusFixture wires a Claude projects root holding one real
// session, plus (when withUnusable is set) a 0-byte 8eda8f4e-style stub. It
// returns the project path (for working_dir) and the valid session's ID.
func setupClaudeCorpusFixture(t *testing.T, withUnusable bool) (projectPath, validSessionID string) {
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

	validSessionID = "claude-session-1"
	writeClaudeSessionFixtureAt(t, projectsRoot, resolvedProject, validSessionID,
		`{"type":"user","timestamp":"2026-01-01T00:00:00Z","sessionId":"`+validSessionID+`","cwd":"`+resolvedProject+`","message":{"role":"user","content":"hello"}}`)

	if withUnusable {
		sessionDir := filepath.Join(projectsRoot, locator.PathToHash(resolvedProject))
		require.NoError(t, os.MkdirAll(sessionDir, 0o755))
		// 0 bytes: parses cleanly, yields zero message entries — the
		// silent-skip shape, as opposed to a parse error.
		require.NoError(t, os.WriteFile(filepath.Join(sessionDir, dir094EmptySessionID+".jsonl"), nil, 0o644))
	}

	return resolvedProject, validSessionID
}

// TestQuerySessions_ToleratesUnusableSessionFileAndReportsIt is AC1+AC2+AC3 on
// the query_sessions path: one bad file costs the caller exactly that file —
// every other session is still listed, and the response names what was
// excluded in both the prose and machine-readable channels.
func TestQuerySessions_ToleratesUnusableSessionFileAndReportsIt(t *testing.T) {
	projectPath, validSessionID := setupClaudeCorpusFixture(t, true)

	output, err := NewToolExecutor().ExecuteTool(&config.Config{}, "query_sessions", map[string]interface{}{
		"working_dir": projectPath,
	})
	require.NoError(t, err, "one unusable session file must not fail the whole listing")

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &resp), "response not valid JSON: %s", output)

	// Results PRESENT: the valid session survived the bad sibling.
	var ids []string
	for _, item := range extractDataArray(t, output) {
		if id, ok := item["session_id"].(string); ok {
			ids = append(ids, id)
		}
	}
	assert.Contains(t, ids, validSessionID, "the usable session must still be listed")
	assert.NotContains(t, ids, dir094EmptySessionID, "the 0-byte stub has nothing to list")

	// Reported, never silent: a warning that names the file...
	warnings, ok := resp["warnings"].([]interface{})
	require.True(t, ok, "response must carry a warnings array; got: %s", output)
	assert.True(t, anyStringContains(warnings, dir094EmptySessionID),
		"a warning must name the skipped file; got %v", warnings)

	// ...plus the machine-readable path list a caller reconciles against.
	skipped, ok := resp["skipped_files"].([]interface{})
	require.True(t, ok, "response must carry a skipped_files array; got: %s", output)
	assert.True(t, anyStringContains(skipped, dir094EmptySessionID),
		"skipped_files must list the excluded path; got %v", skipped)
}

// TestQuerySessions_CleanCorpusStaysWireCompatible pins the other direction:
// with nothing excluded, skipped_files is absent entirely, so a caller whose
// corpus is intact gets the pre-DIR-094 response shape.
func TestQuerySessions_CleanCorpusStaysWireCompatible(t *testing.T) {
	projectPath, validSessionID := setupClaudeCorpusFixture(t, false)

	output, err := NewToolExecutor().ExecuteTool(&config.Config{}, "query_sessions", map[string]interface{}{
		"working_dir": projectPath,
	})
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &resp), "response not valid JSON: %s", output)

	data := extractDataArray(t, output)
	require.Len(t, data, 1)
	assert.Equal(t, validSessionID, data[0]["session_id"])

	_, hasSkipped := resp["skipped_files"]
	assert.False(t, hasSkipped, "a clean corpus must not report skipped_files; got: %s", output)
	warnings, ok := resp["warnings"].([]interface{})
	require.True(t, ok, "warnings stays always-present; got: %s", output)
	assert.Empty(t, warnings, "a clean corpus has nothing to warn about; got %v", warnings)
}

func anyStringContains(values []interface{}, needle string) bool {
	for _, v := range values {
		if s, ok := v.(string); ok && strings.Contains(s, needle) {
			return true
		}
	}
	return false
}
