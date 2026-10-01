package executor

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/locator"
)

// gap-include-subagents-dropped-on-session-id-path:
//
// include_subagents is structurally unreachable whenever the caller supplies
// session_id. dispatchProviderQuery's session_id branch called
// ExecuteQueryForSession WITHOUT forwarding includeSubagents, and
// ExecuteQueryForSession's claude branch streamed only the single parent
// transcript — so a needle that exists solely in that session's subagent
// transcripts returned total_records:0 for BOTH include_subagents:true and
// include_subagents:false. The flag was accepted, defaulted to true, and
// discarded before any resolver saw it.
//
// These tests rebuild a generated fixture (never a real session needle) on
// every run with a fresh nonce, so a leftover from a previous run cannot
// satisfy them, and pass a symlink-resolved working_dir so they exercise the
// physical project directory the locator hashes.

// freshFixtureNonce returns a per-run random token. A stale artifact from an
// earlier run carries a different nonce and therefore cannot satisfy the
// "FIXTURE-SUB-<nonce>" assertion.
func freshFixtureNonce(t *testing.T) string {
	t.Helper()
	b := make([]byte, 12)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return hex.EncodeToString(b)
}

// setupIncludeSubagentsFixtureProject builds a temporary Claude projects root
// holding ONE session whose subagent transcript contains a needle the parent
// transcript does not — the exact on-disk shape the bug hid.
//
//	<projectsRoot>/<hash>/<sid>.jsonl                     -> assistant "FIXTURE-MAIN-<nonce>"
//	<projectsRoot>/<hash>/<sid>/subagents/agent-fix.jsonl -> assistant "FIXTURE-SUB-<nonce>"
//	                                                          (sessionId=<sid>, isSidechain:true)
//
// It returns the symlink-resolved project path (to pass as working_dir), the
// session id, and both needles.
func setupIncludeSubagentsFixtureProject(t *testing.T) (projectPath, sessionID, mainNeedle, subNeedle string) {
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

	hash := locator.PathToHash(resolvedProject)
	sessionDir := filepath.Join(projectsRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	sessionID = "include-subagents-fixture-session"

	nonce := freshFixtureNonce(t)
	mainNeedle = "FIXTURE-MAIN-" + nonce
	subNeedle = "FIXTURE-SUB-" + nonce

	// Parent transcript: carries only the MAIN needle.
	mainLine := fmt.Sprintf(
		`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":%q,"cwd":%q,"message":{"role":"assistant","content":%q}}`,
		sessionID, resolvedProject, mainNeedle)
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, sessionID+".jsonl"), []byte(mainLine+"\n"), 0o644))

	// Subagent transcript: the SUB needle exists ONLY here. Its sessionId field
	// carries the PARENT uuid (matching real transcripts) and isSidechain marks
	// it as a sidechain.
	subDir := filepath.Join(sessionDir, sessionID, locator.SubagentDirName)
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	subLine := fmt.Sprintf(
		`{"type":"assistant","timestamp":"2026-01-01T00:00:01Z","sessionId":%q,"isSidechain":true,"cwd":%q,"message":{"role":"assistant","content":%q}}`,
		sessionID, resolvedProject, subNeedle)
	require.NoError(t, os.WriteFile(
		filepath.Join(subDir, "agent-fix.jsonl"), []byte(subLine+"\n"), 0o644))

	return resolvedProject, sessionID, mainNeedle, subNeedle
}

// totalRecordsFromToolOutput parses an ExecuteTool response envelope and
// returns pagination.total_records (the full pre-page result count).
func totalRecordsFromToolOutput(t *testing.T, output string) int {
	t.Helper()
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &resp), "response not valid JSON: %s", output)
	pagination, ok := resp["pagination"].(map[string]interface{})
	require.True(t, ok, "expected pagination metadata in response: %#v", resp)
	total, _ := pagination["total_records"].(float64)
	return int(total)
}

// TestQuerySessionContent_Claude_SessionIDIncludeSubagents covers AC 1-4: with
// an exact session_id, a needle that exists only in a subagent transcript is
// reachable when include_subagents=true and unreachable when false — the two
// arms are no longer byte-identical. Before the fix both arms returned
// total_records:0.
func TestQuerySessionContent_Claude_SessionIDIncludeSubagents(t *testing.T) {
	projectPath, sessionID, _, subNeedle := setupIncludeSubagentsFixtureProject(t)

	e := NewToolExecutor()
	cfg := &config.Config{}

	withSubs, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":              "all",
		"contains":          subNeedle,
		"session_id":        sessionID,
		"working_dir":       projectPath,
		"include_subagents": true,
	})
	require.NoError(t, err)

	withoutSubs, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":              "all",
		"contains":          subNeedle,
		"session_id":        sessionID,
		"working_dir":       projectPath,
		"include_subagents": false,
	})
	require.NoError(t, err)

	totalWith := totalRecordsFromToolOutput(t, withSubs)
	totalWithout := totalRecordsFromToolOutput(t, withoutSubs)

	require.GreaterOrEqual(t, totalWith, 1,
		"include_subagents=true must reach the subagent transcript holding %q", subNeedle)
	require.Equal(t, 0, totalWithout,
		"include_subagents=false must not search subagent transcripts")
	require.NotEqual(t, withSubs, withoutSubs,
		"the two arms must no longer be byte-identical")
}

// TestQuerySessionContent_Claude_SessionIDIncludeSubagents_MainStillMatches
// proves the fix does not displace the parent transcript: a needle in the
// parent file is found with include_subagents=false too.
func TestQuerySessionContent_Claude_SessionIDIncludeSubagents_MainStillMatches(t *testing.T) {
	projectPath, sessionID, mainNeedle, _ := setupIncludeSubagentsFixtureProject(t)

	e := NewToolExecutor()
	cfg := &config.Config{}

	out, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":              "all",
		"contains":          mainNeedle,
		"session_id":        sessionID,
		"working_dir":       projectPath,
		"include_subagents": false,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, totalRecordsFromToolOutput(t, out), 1,
		"a needle in the parent transcript must still be reachable with include_subagents=false")
}
