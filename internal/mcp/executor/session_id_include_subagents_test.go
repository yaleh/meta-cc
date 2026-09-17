package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupClaudeSubagentNeedleFixture builds a project whose session has a
// subagents/ directory, and places the needle EXCLUSIVELY inside that
// subagent file. The main session JSONL never contains it.
//
// AC2 is enforced here, by POSITION and not by claim: the fixture reads both
// files back off disk after writing them and asserts the needle occurs in the
// subagent file and zero times in the main session file. A regression that
// moved the needle into the main file (which would make the test pass for the
// wrong reason — "found" without ever expanding the subagent directory) fails
// at fixture construction instead of silently green-lighting the probe.
func setupClaudeSubagentNeedleFixture(t *testing.T, needle string) (projectPath, sessionID, subagentFile string) {
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

	projectHash := strings.ReplaceAll(resolvedProject, "\\", "-")
	projectHash = strings.ReplaceAll(projectHash, "/", "-")
	projectHash = strings.ReplaceAll(projectHash, ":", "-")
	sessionDir := filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	sessionID = "claude-subagent-needle-session"

	// The main session file deliberately does NOT contain the needle.
	mainEntry := map[string]interface{}{
		"type":      "user",
		"timestamp": "2026-01-01T00:00:00Z",
		"sessionId": sessionID,
		"cwd":       resolvedProject,
		"message":   map[string]interface{}{"role": "user", "content": "main session body without the probe"},
	}
	mainLine, err := json.Marshal(mainEntry)
	require.NoError(t, err)

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, append(mainLine, '\n'), 0o644))

	// The subagent file is the ONLY place the needle lives.
	subDir := filepath.Join(sessionDir, sessionID, "subagents")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	subEntry := map[string]interface{}{
		"type":      "user",
		"timestamp": "2026-01-01T00:00:01Z",
		"sessionId": sessionID,
		"cwd":       resolvedProject,
		"message":   map[string]interface{}{"role": "user", "content": needle},
	}
	subLine, err := json.Marshal(subEntry)
	require.NoError(t, err)

	subagentFile = filepath.Join(subDir, "agent-needle.jsonl")
	require.NoError(t, os.WriteFile(subagentFile, append(subLine, '\n'), 0o644))

	// ── AC2: assert the probe's POSITION from disk, not from intent ──────────
	mainBytes, err := os.ReadFile(sessionFile)
	require.NoError(t, err)
	require.Zero(t, strings.Count(string(mainBytes), needle),
		"fixture invariant: the needle must not occur in the main session file, or this test cannot distinguish 'found via subagent expansion' from 'found in the main file'")

	subBytes, err := os.ReadFile(subagentFile)
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(string(subBytes), needle),
		"fixture invariant: the needle must occur exactly once in the subagent file")

	return resolvedProject, sessionID, subagentFile
}

// TestQuerySessionContent_ExplicitSessionID_IncludeSubagents_FindsSubagentOnlyContent
// is the regression proof for the silent no-op: with an explicit session_id,
// include_subagents=true used to be dropped on the floor — dispatchProviderQuery
// routed to ExecuteQueryForSession, which streams exactly one file (the main
// session JSONL) and never consults the flag. The probe lives only in
// <session>/subagents/*.jsonl, so returning 0 entries and returning "not found
// anywhere" were the same observable — the exact "queried and clean" vs "never
// queried" conflation this task is about.
func TestQuerySessionContent_ExplicitSessionID_IncludeSubagents_FindsSubagentOnlyContent(t *testing.T) {
	const needle = "NEEDLE_ONLY_IN_SUBAGENT_7f3a9c"
	projectPath, sessionID, _ := setupClaudeSubagentNeedleFixture(t, needle)

	result, err := handleQuerySessionContent(NewToolExecutor(), "project", map[string]interface{}{
		"role":              "all",
		"working_dir":       projectPath,
		"session_id":        sessionID,
		"contains":          needle,
		"include_subagents": true,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1,
		"explicit session_id + include_subagents=true must reach <session>/subagents/*.jsonl, where the needle is the only occurrence")
}

// TestQuerySessionContent_ExplicitSessionID_ExcludeSubagents_DoesNotExpand is
// the paired negative: the fix must not turn subagent expansion into an
// unconditional behavior. With include_subagents=false the same query over the
// same fixture must still find nothing — otherwise the flag has simply been
// replaced by a hard-coded true.
func TestQuerySessionContent_ExplicitSessionID_ExcludeSubagents_DoesNotExpand(t *testing.T) {
	const needle = "NEEDLE_ONLY_IN_SUBAGENT_7f3a9c"
	projectPath, sessionID, _ := setupClaudeSubagentNeedleFixture(t, needle)

	result, err := handleQuerySessionContent(NewToolExecutor(), "project", map[string]interface{}{
		"role":              "all",
		"working_dir":       projectPath,
		"session_id":        sessionID,
		"contains":          needle,
		"include_subagents": false,
	})
	require.NoError(t, err)
	require.Empty(t, result.Entries,
		"include_subagents=false must keep the subagent directory unexpanded, so needle-only-in-subagent stays invisible")
}
