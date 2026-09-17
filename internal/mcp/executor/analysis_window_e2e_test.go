package executor

// DIR-095 end-to-end tests for the since/until window on the five analysis
// tools, exercised through the REAL MCP dispatch path (ExecuteTool →
// ValidateToolArgs → ExecuteSpecialTool → handler → internal/analysis.Service)
// rather than against the service directly.
//
// This layer is what proves the two things the service-level tests cannot:
//
//  1. since/until are actually *declarable* on these five tools. ExecuteTool
//     runs toolspkg.ValidateToolArgs first, and ValidateArgKeys rejects any key
//     the tool's schema does not declare — so before DIR-095 these tools could
//     not be asked a time-bounded question at all, and a schema-only fix
//     without a handler that reads the bounds would fail the same way.
//  2. A malformed bound reaches the caller as mcerrors.ErrInvalidInput from
//     the tool boundary, not as a string a caller has to pattern-match.

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/types"
)

// windowE2ETools are the five tools DIR-095 put behind the window.
var windowE2ETools = []string{
	"analyze_errors",
	"analyze_bugs",
	"quality_scan",
	"get_work_patterns",
	"get_tech_debt",
}

// The E2E window selects 2026-03-03 only, out of a two-day corpus.
const (
	windowE2ESince = "2026-03-03T00:00:00Z"
	windowE2EUntil = "2026-03-04T00:00:00Z"
	windowE2EFirst = 3
	windowE2ELast  = 4
)

// windowE2EPair builds one tool_use + tool_result pair.
func windowE2EPair(id, ts, tool, filePath, output, status, errText string) []types.SessionEntry {
	return []types.SessionEntry{
		{
			Type:      "assistant",
			UUID:      "u-" + id,
			Timestamp: ts,
			Message: &types.Message{
				Role: "assistant",
				Content: []types.ContentBlock{
					{Type: "tool_use", ToolUse: &types.ToolUse{
						ID:    id,
						Name:  tool,
						Input: map[string]interface{}{"file_path": filePath},
					}},
				},
			},
		},
		{
			Type:      "user",
			UUID:      "r-" + id,
			Timestamp: ts,
			Message: &types.Message{
				Role: "user",
				Content: []types.ContentBlock{
					{Type: "tool_result", ToolResult: &types.ToolResult{
						ToolUseID: id,
						Content:   output,
						Status:    status,
						Error:     errText,
					}},
				},
			},
		},
	}
}

// windowE2EDayEntries builds one corpus day: a failing Read followed by a
// succeeding Read whose output carries a TODO marker, so all five tools have
// something non-trivial to count on that day.
func windowE2EDayEntries(day int) []types.SessionEntry {
	ts := fmt.Sprintf("2026-03-%02dT10:00:00Z", day)
	filePath := fmt.Sprintf("pkg/day%d/file.go", day)

	var entries []types.SessionEntry
	entries = append(entries,
		windowE2EPair(fmt.Sprintf("d%d-err", day), ts, "Read", filePath, "ERROR: read failed", "error", "ERROR: read failed")...)
	entries = append(entries,
		windowE2EPair(fmt.Sprintf("d%d-ok", day), ts, "Read", filePath, fmt.Sprintf("// TODO: day%d cleanup\n", day), "success", "")...)
	return entries
}

// windowE2ECorpus is both days; windowE2ESubset is only the day the window
// keeps, built from the same function so the entries are identical.
func windowE2ECorpus() []types.SessionEntry {
	return append(windowE2EDayEntries(windowE2EFirst), windowE2EDayEntries(windowE2ELast)...)
}

func windowE2ESubset() []types.SessionEntry {
	return windowE2EDayEntries(windowE2EFirst)
}

// seedWindowE2EProject writes entries as one Claude session for a fresh project
// and returns the project path to pass as working_dir. It scans the corpus for
// a .jsonl timestamp rather than hard-coding one, so a future edit to the day
// builder cannot silently desynchronize the fixture from the window constants.
func seedWindowE2EProject(t *testing.T, projectsRoot string, entries []types.SessionEntry) string {
	t.Helper()

	absProject, err := filepath.Abs(t.TempDir())
	require.NoError(t, err)
	resolved, err := filepath.EvalSymlinks(absProject)
	require.NoError(t, err)

	var lines []string
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		require.NoError(t, err)
		lines = append(lines, string(data))
	}
	writeClaudeSessionFixtureAt(t, projectsRoot, resolved,
		"analysis-window-session", strings.Join(lines, "\n")+"\n")

	return resolved
}

// newWindowE2EProjects seeds the full corpus and the known time-bounded subset
// under one shared projects root, so both are addressable by working_dir.
func newWindowE2EProjects(t *testing.T) (full, subset string) {
	t.Helper()

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	return seedWindowE2EProject(t, projectsRoot, windowE2ECorpus()),
		seedWindowE2EProject(t, projectsRoot, windowE2ESubset())
}

// TestAnalysisTools_HonorWindow_EndToEnd is AC-1 proven through the real tool
// boundary: with since/until supplied, each of the five tools answers exactly
// as it would against a corpus holding only the in-window records.
func TestAnalysisTools_HonorWindow_EndToEnd(t *testing.T) {
	full, subset := newWindowE2EProjects(t)
	executor := NewToolExecutor()

	for _, tool := range windowE2ETools {
		t.Run(tool, func(t *testing.T) {
			unwindowed, err := executor.ExecuteTool(nil, tool, map[string]interface{}{"working_dir": full})
			require.NoError(t, err)

			windowed, err := executor.ExecuteTool(nil, tool, map[string]interface{}{
				"working_dir": full,
				"since":       windowE2ESince,
				"until":       windowE2EUntil,
			})
			require.NoError(t, err, "%s must accept since/until (ValidateToolArgs rejects undeclared keys)", tool)

			known, err := executor.ExecuteTool(nil, tool, map[string]interface{}{"working_dir": subset})
			require.NoError(t, err)

			assert.NotEqual(t, unwindowed, windowed,
				"%s must not answer a windowed question from the whole corpus", tool)
			assert.JSONEq(t, known, windowed,
				"%s windowed over the full corpus must match the known time-bounded subset", tool)
		})
	}
}

// TestAnalysisTools_RejectInvalidWindow_EndToEnd is AC-2 proven through the
// real tool boundary, and is deliberately run against a corpus the tools can
// otherwise read successfully: the point is that the invalid-input verdict
// comes from the parameter, not from a failed lookup.
func TestAnalysisTools_RejectInvalidWindow_EndToEnd(t *testing.T) {
	full, _ := newWindowE2EProjects(t)
	executor := NewToolExecutor()

	for _, tool := range windowE2ETools {
		for _, tc := range []struct {
			key   string
			value string
		}{
			{key: "since", value: "not-a-timestamp"},
			{key: "until", value: "2026-03-04"},
		} {
			t.Run(tool+"/"+tc.key, func(t *testing.T) {
				_, err := executor.ExecuteTool(nil, tool, map[string]interface{}{
					"working_dir": full,
					tc.key:        tc.value,
				})
				require.Error(t, err, "%s must reject %s=%q", tool, tc.key, tc.value)
				assert.True(t, errors.Is(err, mcerrors.ErrInvalidInput),
					"%s must report a malformed %s as ErrInvalidInput, got: %v", tool, tc.key, err)
			})
		}
	}
}

// TestAnalysisTools_HonorWindow_StatsOnly_EndToEnd is AC-3 through the tool
// boundary: the aggregate-only response is windowed too, so a stats_only
// caller cannot be handed whole-corpus numbers for a bounded question.
func TestAnalysisTools_HonorWindow_StatsOnly_EndToEnd(t *testing.T) {
	full, subset := newWindowE2EProjects(t)
	executor := NewToolExecutor()

	for _, tool := range windowE2ETools {
		t.Run(tool, func(t *testing.T) {
			unwindowed, err := executor.ExecuteTool(nil, tool, map[string]interface{}{
				"working_dir": full,
				"stats_only":  true,
			})
			require.NoError(t, err)

			windowed, err := executor.ExecuteTool(nil, tool, map[string]interface{}{
				"working_dir": full,
				"stats_only":  true,
				"since":       windowE2ESince,
				"until":       windowE2EUntil,
			})
			require.NoError(t, err)

			known, err := executor.ExecuteTool(nil, tool, map[string]interface{}{
				"working_dir": subset,
				"stats_only":  true,
			})
			require.NoError(t, err)

			assert.NotEqual(t, unwindowed, windowed, "%s stats_only must reflect the window", tool)
			assert.JSONEq(t, known, windowed,
				"%s stats_only windowed over the full corpus must match the known time-bounded subset", tool)
		})
	}
}

// TestAnalysisTools_RejectUndeclaredStatsFirst pins the DIR-048 boundary that
// AC-3's "stats_first" clause runs into on these five tools: stats_first is
// deliberately declared only on the four pipeline-routed query tools, because
// the six analysis tools return typed non-flat analyzer results and dispatch
// past the response pipeline that implements it. So it is not a second stats
// mode silently ignoring the window — it is rejected as an unknown parameter,
// which is the honest reading of that clause for these tools.
func TestAnalysisTools_RejectUndeclaredStatsFirst(t *testing.T) {
	full, _ := newWindowE2EProjects(t)
	executor := NewToolExecutor()

	for _, tool := range windowE2ETools {
		t.Run(tool, func(t *testing.T) {
			_, err := executor.ExecuteTool(nil, tool, map[string]interface{}{
				"working_dir": full,
				"stats_first": true,
			})
			require.Error(t, err, "%s must not silently ignore stats_first", tool)
			assert.Contains(t, err.Error(), "stats_first",
				"the rejection must name the unsupported parameter, got: %v", err)
		})
	}
}
