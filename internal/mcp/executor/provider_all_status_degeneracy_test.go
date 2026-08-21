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

// DIR-046: query_session_signals(type="tool_stats", status=...) and
// (type="errors") both degenerated badly under provider="all" — status
// filtering (and the plain errors handler) returned near-zero records even
// though errors existed across multiple sessions/providers. Root cause: a
// join-direction bug in internal/provider/claude/turns.go's
// buildTurns/joinToolCalls (see the DIR-046 regression test in
// internal/provider/claude/provider_test.go) correlated each tool_use with
// the WRONG neighboring tool_result, so records.Normalize's tool_result
// records — starved of their real Output/IsError — got dropped entirely
// (`if call.Output == "" && !call.IsError { continue }`), leaving
// handleQueryToolErrors and handleQueryTools' status join with virtually no
// is_error signal for any Claude-sourced record under provider="all" (the
// only path that routes Claude data through Normalize; provider="claude"
// alone uses the raw-JSONL path and never hit this bug). This test seeds
// TWO Claude sessions (each with one failed and one successful tool call)
// plus a Codex session (one failed and one successful tool call) under the
// same project, and proves provider="all" queries surface errors/successes
// spanning more than one session and more than one provider — not the
// silent near-zero collapse this task was filed to fix.

// setupProviderAllStatusDegeneracyFixture wires two Claude sessions (each
// with a failed Bash call and a successful Read call) and one Codex session
// (a failed exec_command call and a successful one) all scoped to the same
// project path, so provider="all" queries must aggregate across both
// providers and all three sessions.
func setupProviderAllStatusDegeneracyFixture(t *testing.T) (projectPath string) {
	t.Helper()
	// Pin the files backend so a real `codex app-server` never shadows the
	// hermetic fixture corpus (see tests/e2e/codex-e2e.sh).
	t.Setenv("META_CC_CODEX_BACKEND", "files")

	claudeRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", claudeRoot)
	t.Setenv("HOME", t.TempDir())

	rawProjectPath := t.TempDir()
	absProject, err := filepath.Abs(rawProjectPath)
	require.NoError(t, err)
	resolvedProject, err := filepath.EvalSymlinks(absProject)
	require.NoError(t, err)

	hash := strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolvedProject)
	sessionDir := filepath.Join(claudeRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	writeClaudeSession := func(sessionID string) {
		lines := []string{
			// Failed tool call (Bash).
			`{"type":"user","uuid":"` + sessionID + `-u1","parentUuid":null,"timestamp":"2026-01-01T00:00:00Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"text","text":"do stuff"}]}}`,
			`{"type":"assistant","uuid":"` + sessionID + `-a1","parentUuid":"` + sessionID + `-u1","timestamp":"2026-01-01T00:00:01Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"` + sessionID + `-toolu_err","name":"Bash","input":{"command":"false"}}]}}`,
			`{"type":"user","uuid":"` + sessionID + `-u2","parentUuid":"` + sessionID + `-a1","timestamp":"2026-01-01T00:00:02Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"` + sessionID + `-toolu_err","is_error":true,"content":"command failed"}]}}`,
			// Successful tool call (Read).
			`{"type":"assistant","uuid":"` + sessionID + `-a2","parentUuid":"` + sessionID + `-u2","timestamp":"2026-01-01T00:00:03Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"` + sessionID + `-toolu_ok","name":"Read","input":{"file_path":"/tmp/foo"}}]}}`,
			`{"type":"user","uuid":"` + sessionID + `-u3","parentUuid":"` + sessionID + `-a2","timestamp":"2026-01-01T00:00:04Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"` + sessionID + `-toolu_ok","is_error":false,"content":"file contents"}]}}`,
		}
		sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
		require.NoError(t, os.WriteFile(sessionFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644))
	}
	writeClaudeSession("claude-session-a")
	writeClaudeSession("claude-session-b")

	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("META_CC_CODEX_ROOT", codexHome)
	require.NoError(t, os.MkdirAll(codexHome, 0o755))

	// Two separate turns (one tool_use per assistant message) rather than
	// two tool calls in a single turn: gojq's `select(gen)` yields once per
	// TRUE value a generator condition produces, so an assistant message
	// with 2+ tool_use blocks legitimately appears more than once in
	// handleQueryTools' output (a pre-existing, provider-agnostic quirk of
	// the unfiltered/join jq shape, unrelated to the DIR-046 join-direction
	// bug this test targets) — keeping one tool call per turn here avoids
	// conflating that unrelated behavior with the fix under test.
	rolloutPath := filepath.Join(codexHome, "rollout.jsonl")
	codexLines := []string{
		`{"timestamp":"2026-01-01T00:01:00Z","type":"thread.started","payload":{"id":"codex-session"}}`,
		`{"timestamp":"2026-01-01T00:01:01Z","type":"turn.started","payload":{"id":"turn-1"}}`,
		`{"timestamp":"2026-01-01T00:01:02Z","type":"item.message","payload":{"role":"user","content":"do codex stuff"}}`,
		`{"timestamp":"2026-01-01T00:01:03Z","type":"item.tool_call","payload":{"id":"codex-toolu-err","name":"exec_command","input":{"cmd":"false"}}}`,
		`{"timestamp":"2026-01-01T00:01:04Z","type":"item.tool_result","payload":{"id":"codex-toolu-err","output":"command failed","is_error":true}}`,
		`{"timestamp":"2026-01-01T00:01:05Z","type":"turn.completed","payload":{"id":"turn-1"}}`,
		`{"timestamp":"2026-01-01T00:01:06Z","type":"turn.started","payload":{"id":"turn-2"}}`,
		`{"timestamp":"2026-01-01T00:01:07Z","type":"item.tool_call","payload":{"id":"codex-toolu-ok","name":"exec_command","input":{"cmd":"pwd"}}}`,
		`{"timestamp":"2026-01-01T00:01:08Z","type":"item.tool_result","payload":{"id":"codex-toolu-ok","output":"/tmp/project","is_error":false}}`,
		`{"timestamp":"2026-01-01T00:01:09Z","type":"turn.completed","payload":{"id":"turn-2"}}`,
	}
	require.NoError(t, os.WriteFile(rolloutPath, []byte(strings.Join(codexLines, "\n")+"\n"), 0o644))

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
		"codex-session", rolloutPath, resolvedProject, "provider-all degeneracy fixture", "gpt-5", "openai", 0, "cli", int64(1700000000))
	require.NoError(t, err)

	return resolvedProject
}

// entrySessionID extracts the normalized/raw session identifier off a
// query result entry, tolerating either field name (raw Claude JSONL uses
// "sessionId"; providerrecords.Normalize emits both "sessionId" and
// "session_id").
func entrySessionID(t *testing.T, entry interface{}) string {
	t.Helper()
	rec, ok := entry.(map[string]interface{})
	require.True(t, ok, "unexpected entry shape: %#v", entry)
	if sid, ok := rec["sessionId"].(string); ok && sid != "" {
		return sid
	}
	if sid, ok := rec["session_id"].(string); ok && sid != "" {
		return sid
	}
	t.Fatalf("entry has no sessionId/session_id: %#v", rec)
	return ""
}

// TestHandleQueryTools_ProviderAll_StatusFilterSpansMultipleSessions is the
// AC1 regression test: status="error" under provider="all" must return
// tool-call-error records from more than one session, not the pre-fix 0.
func TestHandleQueryTools_ProviderAll_StatusFilterSpansMultipleSessions(t *testing.T) {
	projectPath := setupProviderAllStatusDegeneracyFixture(t)
	e := NewToolExecutor()

	errResult, err := handleQueryTools(e, "project", map[string]interface{}{
		"working_dir": projectPath,
		"provider":    "all",
		"status":      "error",
	})
	require.NoError(t, err)
	require.NotEmpty(t, errResult.Entries, "status=error under provider=all must not silently collapse to 0 records")

	errSessions := map[string]bool{}
	for _, entry := range errResult.Entries {
		errSessions[entrySessionID(t, entry)] = true
	}
	require.Greater(t, len(errSessions), 1,
		"status=error under provider=all must span more than one session, got sessions: %v", errSessions)

	successResult, err := handleQueryTools(e, "project", map[string]interface{}{
		"working_dir": projectPath,
		"provider":    "all",
		"status":      "success",
	})
	require.NoError(t, err)
	require.NotEmpty(t, successResult.Entries, "status=success under provider=all must not silently collapse")

	successSessions := map[string]bool{}
	for _, entry := range successResult.Entries {
		successSessions[entrySessionID(t, entry)] = true
	}
	require.Greater(t, len(successSessions), 1,
		"status=success under provider=all must span more than one session, got sessions: %v", successSessions)

	// The three known providers/sessions (2 Claude + 1 Codex) each carry
	// exactly one error and one success tool call, so error+success counts
	// must reflect that — not the pre-fix near-total loss of Claude-sourced
	// signal (which left only the Codex session's 1 success visible).
	require.Len(t, errResult.Entries, 3, "expected exactly 3 error tool calls (1 per session across 2 Claude + 1 Codex)")
	require.Len(t, successResult.Entries, 3, "expected exactly 3 success tool calls (1 per session across 2 Claude + 1 Codex)")
}

// TestHandleQueryToolErrors_ProviderAll_SpansMultipleSessions is the AC2
// regression test: type="errors" under provider="all" must likewise return
// records spanning more than the single session it degenerated to.
func TestHandleQueryToolErrors_ProviderAll_SpansMultipleSessions(t *testing.T) {
	projectPath := setupProviderAllStatusDegeneracyFixture(t)
	e := NewToolExecutor()

	result, err := handleQueryToolErrors(e, "project", map[string]interface{}{
		"working_dir": projectPath,
		"provider":    "all",
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.Entries, "type=errors under provider=all must not silently collapse to 0 records")

	sessions := map[string]bool{}
	for _, entry := range result.Entries {
		sessions[entrySessionID(t, entry)] = true
	}
	require.Greater(t, len(sessions), 1,
		"type=errors under provider=all must span more than one session, got sessions: %v", sessions)
	require.Len(t, result.Entries, 3, "expected exactly 3 tool_result error records (1 per session across 2 Claude + 1 Codex)")
}
