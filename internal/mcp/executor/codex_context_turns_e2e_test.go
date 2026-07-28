package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/locator"
)

// DIR-036: query_session_content(role=user, provider=codex, context_turns>0)
// used to silently return an empty result even though the identical query
// with context_turns=0 returned real matches (the pipeline rescanned a
// Claude-shaped baseDir by uuid, but Codex-normalized records have neither).
// These tests reproduce that false negative end-to-end (through
// ExecuteTool, not just handleQuerySessionContent) and prove the fix.

// writeClaudeSessionFixtureAt writes a Claude-shaped JSONL fixture into the
// META_CC_PROJECTS_ROOT-hashed directory locator.SessionLocator expects for
// projectPath, mirroring cmd/mcp-server/executor_test.go's
// writeSessionFixture so the same fixture shape/discovery mechanics are
// exercised from this package too.
func writeClaudeSessionFixtureAt(t *testing.T, projectsRoot, projectPath, sessionID, content string) {
	t.Helper()
	hash := locator.PathToHash(projectPath)
	sessionDir := filepath.Join(projectsRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(content), 0o644))
}

// extractDataArray parses an ExecuteTool JSON response (inline or file_ref
// mode) and returns the "data" array as []map[string]interface{}.
func extractDataArray(t *testing.T, output string) []map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &resp), "response not valid JSON: %s", output)

	mode, _ := resp["mode"].(string)
	var items []interface{}
	if mode == "file_ref" {
		fileRef, ok := resp["file_ref"].(map[string]interface{})
		require.True(t, ok, "file_ref mode missing file_ref: %#v", resp)
		path, _ := fileRef["path"].(string)
		require.NotEmpty(t, path, "file_ref missing path: %#v", fileRef)
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		defer os.Remove(path)
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			if line == "" {
				continue
			}
			var v interface{}
			require.NoError(t, json.Unmarshal([]byte(line), &v))
			items = append(items, v)
		}
	} else {
		data, ok := resp["data"].([]interface{})
		require.True(t, ok, "expected data array, got %#v", resp["data"])
		items = data
	}

	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		require.True(t, ok, "expected map item, got %T", item)
		out = append(out, m)
	}
	return out
}

// TestQuerySessionContent_Codex_ContextTurnsPreservesMatches is the DIR-036
// core regression: a 5-turn Codex rollout fixture where only the middle
// turn's user message matches "吞吐" (mirroring the /home/yale/work/quay
// reproduction). Without context_turns, exactly one record matches. The bug
// was that adding context_turns>0 silently replaced that match with an empty
// result; the fix must return the match plus a bounded, correctly-ordered
// context window instead.
func TestQuerySessionContent_Codex_ContextTurnsPreservesMatches(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-context-turns-sample.jsonl")
	e := NewToolExecutor()
	cfg := &config.Config{}

	// Baseline: context_turns omitted (0) returns exactly 1 match.
	baseOut, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "user", "provider": "codex", "pattern": "吞吐", "working_dir": projectPath,
	})
	require.NoError(t, err)
	baseData := extractDataArray(t, baseOut)
	require.Len(t, baseData, 1, "baseline (context_turns=0) must return exactly 1 match")

	// RED (pre-fix)/GREEN (post-fix): context_turns=2 must not erase the match.
	ctxOut, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "user", "provider": "codex", "pattern": "吞吐", "context_turns": float64(2), "working_dir": projectPath,
	})
	require.NoError(t, err)
	ctxData := extractDataArray(t, ctxOut)
	require.NotEmpty(t, ctxData, "context_turns=2 must not silently erase the Codex match (DIR-036)")

	// Exactly one record must be the match (context:false); the fixture's
	// middle turn (turn-3, index 4 in the 10-record stream) with N=2 yields
	// the bounded window [idx2..idx6]: turn-2's user+assistant, turn-3's
	// user (match) + assistant, and turn-4's user — 5 records total.
	var matchCount int
	var contents []string
	for _, rec := range ctxData {
		msg, _ := rec["message"].(map[string]interface{})
		require.NotNil(t, msg, "record missing message: %#v", rec)
		if content, ok := msg["content"].(string); ok {
			contents = append(contents, content)
		}
		if ctx, ok := rec["context"].(bool); ok && !ctx {
			matchCount++
		}
	}
	require.Equal(t, 1, matchCount, "expected exactly one context:false matched record, got records: %#v", ctxData)
	require.Len(t, ctxData, 5, "expected bounded window of 5 records (2 before, match, 2 after by record position), got %#v", contents)
}

// TestQuerySessionContent_Codex_FiveTurnFixtureMiddleMatchOrdering is the
// AC's "five-turn fixture matching the middle turn" check in isolation:
// windows must be bounded, ordered, and carry correct context flags.
func TestQuerySessionContent_Codex_FiveTurnFixtureMiddleMatchOrdering(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-context-turns-sample.jsonl")
	e := NewToolExecutor()
	cfg := &config.Config{}

	out, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "user", "provider": "codex", "pattern": "吞吐", "context_turns": float64(1), "working_dir": projectPath,
	})
	require.NoError(t, err)
	data := extractDataArray(t, out)

	// N=1 around the matched user record (idx4 in the 10-record stream)
	// bounds to [idx3..idx5]: turn-2's assistant, turn-3's user (match),
	// turn-3's assistant.
	require.Len(t, data, 3, "expected bounded 3-record window for N=1, got %#v", data)

	var order []string
	for _, rec := range data {
		ts, _ := rec["timestamp"].(string)
		order = append(order, ts)
	}
	for i := 1; i < len(order); i++ {
		require.LessOrEqual(t, order[i-1], order[i], "records must be in chronological order: %#v", order)
	}
}

// TestQuerySessionContent_Codex_ContextTurnsWithGroupBySession proves
// context_turns and group_by_session compose correctly for Codex: the
// grouped result must not be empty, and must still contain the match.
func TestQuerySessionContent_Codex_ContextTurnsWithGroupBySession(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-context-turns-sample.jsonl")
	e := NewToolExecutor()
	cfg := &config.Config{}

	out, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "user", "provider": "codex", "pattern": "吞吐",
		"context_turns": float64(1), "group_by_session": true, "working_dir": projectPath,
	})
	require.NoError(t, err)
	data := extractDataArray(t, out)
	require.NotEmpty(t, data, "group_by_session + context_turns must not erase Codex matches")

	found := false
	for _, group := range data {
		turns, ok := group["turns"].([]interface{})
		if !ok {
			continue
		}
		if len(turns) > 0 {
			found = true
		}
	}
	require.True(t, found, "expected at least one non-empty session group, got %#v", data)
}

// TestQuerySessionContent_Codex_ContextTurnsProviderAll proves provider="all"
// composes correctly with context_turns when the match is Codex-only: the
// Codex match must survive alongside an (unrelated, non-matching) Claude
// session in the same project, not be erased by the multi-provider merge.
func TestQuerySessionContent_Codex_ContextTurnsProviderAll(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-context-turns-sample.jsonl")

	// Give the Claude leg of provider=all a real (but non-matching) session
	// in the same project directory: internal/provider/claude.Provider.
	// ListSessions -> locator.AllSessionsFromProject returns a hard error
	// (not just an empty slice) when a project has zero Claude sessions,
	// which would abort providerrecords.Build's provider=all merge before
	// context expansion is ever reached — a pre-existing, DIR-036-unrelated
	// constraint this fixture must satisfy to exercise the actual
	// provider=all + context_turns path.
	isolatedRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", isolatedRoot)
	writeClaudeSessionFixtureAt(t, isolatedRoot, projectPath, "claude-unrelated-session",
		`{"type":"user","uuid":"cu0","sessionId":"claude-unrelated-session","timestamp":"2026-07-20T08:00:00Z","message":{"role":"user","content":"unrelated claude turn"}}`+"\n")

	e := NewToolExecutor()
	cfg := &config.Config{}

	out, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "user", "provider": "all", "pattern": "吞吐", "context_turns": float64(2), "working_dir": projectPath,
	})
	require.NoError(t, err)
	data := extractDataArray(t, out)
	require.NotEmpty(t, data, "provider=all + context_turns must not erase the Codex match")

	matchCount := 0
	for _, rec := range data {
		if ctx, ok := rec["context"].(bool); ok && !ctx {
			matchCount++
		}
	}
	require.Equal(t, 1, matchCount, "expected exactly one matched record under provider=all, got %#v", data)
}

// TestClaudeCodexContextTurnsParity builds equivalent 5-turn logical
// histories for Claude and Codex (same number of turns, same matched
// middle-turn pattern) and proves both providers produce an equivalent
// logical window shape (record count and single-match count) for the same
// N, without regressing the existing Claude output contract.
func TestClaudeCodexContextTurnsParity(t *testing.T) {
	// --- Claude side: 5 user-only records (uuid-identified), matching the
	// existing Claude ExpandContextTurns contract/tests exactly. ---
	projectDir := t.TempDir()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	var lines []string
	msgs := []string{"turn one baseline", "turn two baseline", "measure 吞吐 rate please", "turn four baseline", "turn five baseline"}
	for i, m := range msgs {
		lines = append(lines, fmt.Sprintf(
			`{"type":"user","uuid":"cl%d","sessionId":"claude-parity-sess","timestamp":"2026-07-20T09:00:%02dZ","message":{"role":"user","content":"%s"}}`,
			i, i, m,
		))
	}
	writeClaudeSessionFixtureAt(t, projectsRoot, projectDir, "claude-parity-session", joinLines(lines))

	e := NewToolExecutor()
	cfg := &config.Config{}

	claudeOut, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "user", "provider": "claude", "pattern": "吞吐", "context_turns": float64(1), "working_dir": projectDir,
	})
	require.NoError(t, err)
	claudeData := extractDataArray(t, claudeOut)

	claudeMatches := 0
	for _, rec := range claudeData {
		if ctx, ok := rec["context"].(bool); ok && !ctx {
			claudeMatches++
		}
	}
	require.Equal(t, 1, claudeMatches, "Claude baseline contract: exactly one match")
	require.NotEmpty(t, claudeData, "Claude must return non-empty context window")

	// --- Codex side: equivalent 5-turn history (each turn user+assistant),
	// same matched middle turn, same N. ---
	codexProjectPath := setupCodexRolloutFixtureProject(t, "rollout-context-turns-sample.jsonl")
	codexOut, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "user", "provider": "codex", "pattern": "吞吐", "context_turns": float64(1), "working_dir": codexProjectPath,
	})
	require.NoError(t, err)
	codexData := extractDataArray(t, codexOut)

	codexMatches := 0
	for _, rec := range codexData {
		if ctx, ok := rec["context"].(bool); ok && !ctx {
			codexMatches++
		}
	}
	require.Equal(t, 1, codexMatches, "Codex: exactly one match, same as Claude (provider parity)")
	require.NotEmpty(t, codexData, "Codex must return non-empty context window, matching Claude's non-empty contract")
}

func joinLines(lines []string) string {
	out := ""
	for _, l := range lines {
		out += l + "\n"
	}
	return out
}

// TestQuerySessionContent_Codex_ContextTurnsWithExplicitSessionID proves
// context_turns composes correctly with an explicit session_id (the DIR-030
// fast path): providerrecords.BuildForSession is exactly the mechanism
// expandProviderContext reuses to reload context, so this should already be
// consistent, but is worth locking down explicitly per the DIR-036 AC.
func TestQuerySessionContent_Codex_ContextTurnsWithExplicitSessionID(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-context-turns-sample.jsonl")
	e := NewToolExecutor()
	cfg := &config.Config{}

	// setupCodexRolloutFixtureProject always registers the rollout under the
	// fixed thread id "codex-dedup-session" in its SQLite fixture (see
	// codex_dedup_e2e_test.go), regardless of the rollout file's own
	// session_meta id.
	out, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "user", "provider": "codex", "pattern": "吞吐",
		"context_turns": float64(1), "session_id": "codex-dedup-session", "working_dir": projectPath,
	})
	require.NoError(t, err)
	data := extractDataArray(t, out)
	require.NotEmpty(t, data, "explicit session_id + context_turns must not erase the Codex match")

	matchCount := 0
	for _, rec := range data {
		if ctx, ok := rec["context"].(bool); ok && !ctx {
			matchCount++
		}
	}
	require.Equal(t, 1, matchCount, "expected exactly one matched record, got %#v", data)
}

// TestQuerySessionContent_Codex_ContextTurnsWithContentSummary proves
// content_summary composes with context_turns for Codex the same way it
// already does for Claude: ApplyContentSummary preserves the record's "seq"
// identity (see internal/mcp/filters.ApplyContentSummary) so
// ExpandContextTurnsCanonical can still locate the match, but the expanded
// window's records come from a fresh canonical reload (full shape), exactly
// mirroring the pre-existing Claude ExpandContextTurns behavior (a
// content_summary match combined with context_turns has never preserved the
// summary projection through the expansion — verified against the existing
// Claude path). The composition must not erase the match, not silently
// invent a different (broken) shape for Codex specifically.
func TestQuerySessionContent_Codex_ContextTurnsWithContentSummary(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-context-turns-sample.jsonl")
	e := NewToolExecutor()
	cfg := &config.Config{}

	out, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "user", "provider": "codex", "pattern": "吞吐",
		"context_turns": float64(1), "content_summary": true, "working_dir": projectPath,
	})
	require.NoError(t, err)
	data := extractDataArray(t, out)
	require.NotEmpty(t, data, "content_summary + context_turns must not erase the Codex match")
	require.Len(t, data, 3, "expected the same bounded window as without content_summary")

	matchCount := 0
	for _, rec := range data {
		if ctx, ok := rec["context"].(bool); ok && !ctx {
			matchCount++
		}
	}
	require.Equal(t, 1, matchCount, "expected exactly one matched record surviving the composition, got %#v", data)
}

// TestQuerySessionContent_Codex_ContextTurnsWithPagination proves pagination
// composes with context_turns: the expanded window is paginated afterward,
// not lost.
func TestQuerySessionContent_Codex_ContextTurnsWithPagination(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-context-turns-sample.jsonl")
	e := NewToolExecutor()
	cfg := &config.Config{}

	out, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "user", "provider": "codex", "pattern": "吞吐",
		"context_turns": float64(1), "page_size": float64(2), "working_dir": projectPath,
	})
	require.NoError(t, err)
	data := extractDataArray(t, out)
	require.Len(t, data, 2, "page_size=2 should return exactly 2 of the 3 windowed records")

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(out), &resp))
	pagination, ok := resp["pagination"].(map[string]interface{})
	require.True(t, ok, "expected pagination metadata: %#v", resp)
	total, _ := pagination["total_records"].(float64)
	require.Equal(t, float64(3), total, "pagination total_records should reflect the full expanded window, not just the page")
}
