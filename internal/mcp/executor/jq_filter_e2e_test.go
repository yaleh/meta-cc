package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
)

// jq_filter_e2e_test.go is the DIR-041 end-to-end regression: the four
// consolidated MCP query tools (query_session_content, query_session_signals,
// query_sessions, query_file_activity) all declare a "jq_filter" parameter,
// and it was read into PipelineConfig.JQFilter but never applied anywhere —
// BuildResponse silently returned the full unfiltered result regardless of
// what jq_filter was passed. These tests exercise the REAL ExecuteTool path
// (executor.go's NewToolPipelineConfig -> pipeline.BuildResponse), not just
// the pipeline package's own unit tests, against real fixture session files,
// proving the fix end-to-end for every one of the four affected tools.

// setupJQFilterFixtureProject wires a temporary Claude projects root with a
// single session JSONL containing one of each record shape the four
// consolidated tools care about: a user message, an assistant message with a
// tool_use block (feeds query_session_signals(type=tool_stats) and
// query_session_content(role=assistant/tool)), a tool_result block, a
// file-history-snapshot (feeds query_file_activity), and a system api_error
// (feeds query_session_signals(type=system_errors)).
func setupJQFilterFixtureProject(t *testing.T) (projectPath, sessionID string) {
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

	sessionID = "jq-filter-e2e-session"
	lines := []string{
		`{"type":"user","timestamp":"2026-01-01T00:00:00Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":"please run a command"}}`,
		`{"type":"assistant","timestamp":"2026-01-01T00:00:01Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_1","name":"OldTool","input":{"command":"echo old"}}],"usage":{"input_tokens":1}}}`,
		`{"type":"user","timestamp":"2026-01-01T00:00:02Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_1","is_error":true,"content":"old error"}]}}`,
		`{"type":"file-history-snapshot","timestamp":"2026-01-01T00:00:03Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","messageId":"msg-1","snapshot":{}}`,
		`{"type":"system","subtype":"api_error","timestamp":"2026-01-01T00:00:04Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","content":"old boom"}`,
		`{"type":"assistant","timestamp":"2026-01-01T00:01:01Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_2","name":"NewTool","input":{"command":"echo new"}}],"usage":{"input_tokens":2}}}`,
		`{"type":"user","timestamp":"2026-01-01T00:01:02Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_2","is_error":true,"content":"new error"}]}}`,
		`{"type":"system","subtype":"api_error","timestamp":"2026-01-01T00:01:04Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","content":"new boom"}`,
	}

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644))

	return resolvedProject, sessionID
}

// dataArrayLen parses an ExecuteTool JSON response (inline or file_ref mode)
// and returns len(data), reusing codex_context_turns_e2e_test.go's
// extractDataArray so both output modes are handled correctly.
func dataArrayLen(t *testing.T, output string) int {
	t.Helper()
	return len(extractDataArray(t, output))
}

func TestExecuteTool_QuerySessionTimeRangeFormerlyIgnoredPaths(t *testing.T) {
	projectPath, _ := setupJQFilterFixtureProject(t)
	e := NewToolExecutor()
	cfg := &config.Config{}

	cases := []struct {
		name, tool string
		args       map[string]interface{}
	}{
		{"assistant", "query_session_content", map[string]interface{}{"role": "assistant"}},
		{"tool", "query_session_content", map[string]interface{}{"role": "tool"}},
		{"errors", "query_session_signals", map[string]interface{}{"type": "errors"}},
		{"tokens", "query_session_signals", map[string]interface{}{"type": "tokens"}},
		{"system_errors", "query_session_signals", map[string]interface{}{"type": "system_errors"}},
		{"tool_stats", "query_session_signals", map[string]interface{}{"type": "tool_stats"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.args["working_dir"] = projectPath
			out, err := e.ExecuteTool(cfg, tc.tool, tc.args)
			require.NoError(t, err)
			require.Equal(t, 2, dataArrayLen(t, out), "baseline must include old and new records")
			tc.args["since"] = "2026-01-01T00:01:00Z"
			out, err = e.ExecuteTool(cfg, tc.tool, tc.args)
			require.NoError(t, err)
			require.Equal(t, 1, dataArrayLen(t, out), "since must exclude the old record")
		})
	}
}

func TestExecuteTool_QuerySessionTimeRangeInvalidDateConsistent(t *testing.T) {
	projectPath, _ := setupJQFilterFixtureProject(t)
	e := NewToolExecutor()
	cfg := &config.Config{}
	for _, args := range []map[string]interface{}{
		{"role": "assistant"}, {"role": "tool"},
		{"type": "errors"}, {"type": "tokens"}, {"type": "system_errors"}, {"type": "tool_stats"},
	} {
		args["working_dir"], args["since"] = projectPath, "not-a-date"
		tool := "query_session_signals"
		if _, ok := args["role"]; ok {
			tool = "query_session_content"
		}
		_, err := e.ExecuteTool(cfg, tool, args)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid since value")
	}
}

// TestExecuteTool_QuerySessionContent_JQFilterSelectFalseEmptiesResult is the
// DIR-041 regression for query_session_content: jq_filter="select(false)"
// must return zero records, not the full unfiltered set.
func TestExecuteTool_QuerySessionContent_JQFilterSelectFalseEmptiesResult(t *testing.T) {
	projectPath, _ := setupJQFilterFixtureProject(t)
	e := NewToolExecutor()
	cfg := &config.Config{}

	baseOut, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "assistant", "working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Greater(t, dataArrayLen(t, baseOut), 0, "baseline (no jq_filter) must return the assistant message")

	filteredOut, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "assistant", "jq_filter": "select(false)", "working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Equal(t, 0, dataArrayLen(t, filteredOut), "jq_filter=select(false) must empty the result, got: %s", filteredOut)
}

// TestExecuteTool_QuerySessionSignals_JQFilterSelectFalseEmptiesResult is the
// DIR-041 regression for query_session_signals.
func TestExecuteTool_QuerySessionSignals_JQFilterSelectFalseEmptiesResult(t *testing.T) {
	projectPath, _ := setupJQFilterFixtureProject(t)
	e := NewToolExecutor()
	cfg := &config.Config{}

	baseOut, err := e.ExecuteTool(cfg, "query_session_signals", map[string]interface{}{
		"type": "tool_stats", "working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Greater(t, dataArrayLen(t, baseOut), 0, "baseline (no jq_filter) must return the Bash tool_use")

	filteredOut, err := e.ExecuteTool(cfg, "query_session_signals", map[string]interface{}{
		"type": "tool_stats", "jq_filter": "select(false)", "working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Equal(t, 0, dataArrayLen(t, filteredOut), "jq_filter=select(false) must empty the result, got: %s", filteredOut)
}

// TestExecuteTool_QueryFileActivity_JQFilterSelectFalseEmptiesResult is the
// DIR-041 regression for query_file_activity.
func TestExecuteTool_QueryFileActivity_JQFilterSelectFalseEmptiesResult(t *testing.T) {
	projectPath, _ := setupJQFilterFixtureProject(t)
	e := NewToolExecutor()
	cfg := &config.Config{}

	baseOut, err := e.ExecuteTool(cfg, "query_file_activity", map[string]interface{}{
		"type": "snapshots", "working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Greater(t, dataArrayLen(t, baseOut), 0, "baseline (no jq_filter) must return the file-history-snapshot")

	filteredOut, err := e.ExecuteTool(cfg, "query_file_activity", map[string]interface{}{
		"type": "snapshots", "jq_filter": "select(false)", "working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Equal(t, 0, dataArrayLen(t, filteredOut), "jq_filter=select(false) must empty the result, got: %s", filteredOut)
}

// TestExecuteTool_QuerySessions_JQFilterSelectFalseEmptiesResult is the
// DIR-041 regression for query_sessions. query_sessions builds its entries
// via a Go-native SessionFilter (not an internal jq expression at all), so
// this also proves the fix is a genuine, tool-agnostic post-filter over
// BuildResponse's result set rather than something threaded through each
// tool's own internal jq composition.
func TestExecuteTool_QuerySessions_JQFilterSelectFalseEmptiesResult(t *testing.T) {
	projectPath, sessionID := setupClaudeSessionFixtureProject(t, "hello world")
	e := NewToolExecutor()
	cfg := &config.Config{}

	baseOut, err := e.ExecuteTool(cfg, "query_sessions", map[string]interface{}{
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	baseData := extractDataArray(t, baseOut)
	require.Len(t, baseData, 1, "baseline (no jq_filter) must return the one fixture session")
	require.Equal(t, sessionID, baseData[0]["session_id"])

	filteredOut, err := e.ExecuteTool(cfg, "query_sessions", map[string]interface{}{
		"jq_filter": "select(false)", "working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Equal(t, 0, dataArrayLen(t, filteredOut), "jq_filter=select(false) must empty the result, got: %s", filteredOut)
}
