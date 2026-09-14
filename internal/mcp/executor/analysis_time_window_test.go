package executor

// DIR-095, through the real MCP dispatch path: the five analysis tools that
// previously declared no time filter (analyze_errors, analyze_bugs,
// quality_scan, get_work_patterns, get_tech_debt) now accept the same RFC3339
// since/until window get_timeline already had.
//
// These tests cover the two things that only exist at this layer, which the
// internal/analysis tests cannot see:
//
//  1. The schema. A parameter that internal/analysis honors but no tool schema
//     declares is unreachable in practice: ExecuteTool's ValidateToolArgs
//     rejects undeclared keys, so every real call carrying since/until would
//     hard-error before the handler ever ran (the DIR-023 failure mode).
//  2. validateTimeWindow's fail-fast placement -- since/until are rejected
//     BEFORE the corpus is located and parsed, so a typo'd timestamp does not
//     first read every session file in the project.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	toolspkg "github.com/yaleh/meta-cc/internal/mcp/tools"
)

// analysisToolNames is the six analysis.Service-backed tools. get_timeline is
// included: it already declared since/until, and DIR-095 routed it through the
// same handler guard, so it is pinned here alongside the five it was added to.
var analysisToolNames = []string{
	"analyze_errors",
	"analyze_bugs",
	"quality_scan",
	"get_work_patterns",
	"get_tech_debt",
	"get_timeline",
}

// seedTimeWindowProject writes raw JSONL session lines under projectsRoot in
// the project-hash directory for a fresh project, and returns that project's
// path to use as working_dir.
func seedTimeWindowProject(t *testing.T, projectsRoot string, lines []string) (resolvedProject string) {
	t.Helper()

	project := t.TempDir()
	absProject, err := filepath.Abs(project)
	require.NoError(t, err)
	resolvedProject, err = filepath.EvalSymlinks(absProject)
	require.NoError(t, err)

	hash := strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolvedProject)
	sessionDir := filepath.Join(projectsRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, "session.jsonl"),
		[]byte(strings.Join(lines, "\n")+"\n"), 0o644))

	return resolvedProject
}

// timeWindowProject seeds one January Bash failure plus one February Bash
// success, and returns the working_dir. The window used by the tests below
// selects February only.
func timeWindowProject(t *testing.T) string {
	t.Helper()

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	lines := []string{
		`{"type":"assistant","uuid":"u-old","timestamp":"2026-01-15T10:00:00Z","sessionId":"s1","cwd":"/p",` +
			`"message":{"role":"assistant","content":[{"type":"tool_use","id":"tu-old","name":"Bash","input":{"command":"boom"}}]}}`,
		`{"type":"user","uuid":"r-old","timestamp":"2026-01-15T10:00:00Z","sessionId":"s1","cwd":"/p",` +
			`"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tu-old","content":"boom failure","status":"error","is_error":true}]}}`,
		`{"type":"assistant","uuid":"u-new","timestamp":"2026-02-15T10:00:00Z","sessionId":"s1","cwd":"/p",` +
			`"message":{"role":"assistant","content":[{"type":"tool_use","id":"tu-new","name":"Bash","input":{"command":"ok"}}]}}`,
		`{"type":"user","uuid":"r-new","timestamp":"2026-02-15T10:00:00Z","sessionId":"s1","cwd":"/p",` +
			`"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tu-new","content":"fine","status":"success"}]}}`,
	}
	return seedTimeWindowProject(t, projectsRoot, lines)
}

const (
	e2eWindowSince = "2026-02-01T00:00:00Z"
	e2eWindowUntil = "2026-03-01T00:00:00Z"
)

// TestAnalysisTools_DeclareTimeWindowParameters is the DIR-023-style guard:
// since/until must be DECLARED on all five tools the analysis service now
// honors them on, or ValidateToolArgs rejects them before dispatch and the
// window is dead code in production.
func TestAnalysisTools_DeclareTimeWindowParameters(t *testing.T) {
	for _, name := range analysisToolNames {
		t.Run(name, func(t *testing.T) {
			require.NoError(t,
				toolspkg.ValidateToolArgs(name, map[string]interface{}{
					"since": e2eWindowSince,
					"until": e2eWindowUntil,
				}),
				"since/until must be declared in this tool's schema")

			schema, err := toolspkg.GetToolSchemaByName(toolspkg.BuildToolSchemaIndex(), name)
			require.NoError(t, err)
			for _, p := range []string{"since", "until"} {
				prop, ok := schema.Properties[p]
				require.True(t, ok, "%s must declare %q", name, p)
				assert.Equal(t, "string", prop.Type)
				assert.NotEmpty(t, prop.Description, "%s's %q needs a description", name, p)
			}
		})
	}
}

// TestAnalysisTools_DoNotDeclareStatsFirst pins the DIR-048 scoping that
// DIR-095's "stats_only and stats_first" acceptance criterion has to be read
// against: the six analysis tools return typed analyzer structs rather than
// the flat record array stats_first paginates, so they never declared it. With
// no stats_first output in existence for these tools, the window's obligation
// there is discharged by the stats_only path (covered in
// internal/analysis/service_time_window_test.go) plus this pin -- if a later
// change starts accepting stats_first here, this test fails and whoever made
// that change must decide what the window should mean for it.
func TestAnalysisTools_DoNotDeclareStatsFirst(t *testing.T) {
	for _, name := range analysisToolNames {
		t.Run(name, func(t *testing.T) {
			err := toolspkg.ValidateToolArgs(name, map[string]interface{}{"stats_first": true})
			require.Error(t, err, "stats_first is not a parameter of %s (DIR-048)", name)
			assert.Contains(t, err.Error(), "stats_first")
		})
	}
}

// TestAnalysisTools_EndToEnd_RejectsInvalidTimeWindow proves the sentinel
// reaches the caller through the real ExecuteTool dispatch, for every tool.
func TestAnalysisTools_EndToEnd_RejectsInvalidTimeWindow(t *testing.T) {
	project := timeWindowProject(t)

	for _, name := range analysisToolNames {
		for _, bad := range []map[string]interface{}{
			{"since": "last tuesday"},
			{"until": "2026-13-45T00:00:00Z"},
		} {
			t.Run(name, func(t *testing.T) {
				args := map[string]interface{}{"working_dir": project}
				for k, v := range bad {
					args[k] = v
				}
				out, err := NewToolExecutor().ExecuteTool(nil, name, args)
				require.Error(t, err)
				assert.True(t, errors.Is(err, mcerrors.ErrInvalidInput),
					"expected the invalid-input sentinel, got: %v", err)
				assert.Empty(t, out)
			})
		}
	}
}

// TestAnalyzeErrors_EndToEnd_HonorsTimeWindow is the user-visible acceptance
// check: the same tool call, differing only by the window, reports one error
// for the whole corpus and zero for the February-only window.
func TestAnalyzeErrors_EndToEnd_HonorsTimeWindow(t *testing.T) {
	project := timeWindowProject(t)
	e := NewToolExecutor()

	out, err := e.ExecuteTool(nil, "analyze_errors", map[string]interface{}{"working_dir": project})
	require.NoError(t, err)
	assert.Contains(t, out, `"total_errors":1`, "the unwindowed run sees the January failure")

	out, err = e.ExecuteTool(nil, "analyze_errors", map[string]interface{}{
		"working_dir": project,
		"since":       e2eWindowSince,
		"until":       e2eWindowUntil,
	})
	require.NoError(t, err)

	var res struct {
		TotalErrors int `json:"total_errors"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &res))
	assert.Zero(t, res.TotalErrors, "the January failure is outside the window")
}

// TestAnalyzeErrors_EndToEnd_AllowsOmittedWindow is the sanity counterpart: a
// call with neither since nor until must be unaffected by the new guard, so
// existing callers see no behavior change.
func TestAnalyzeErrors_EndToEnd_AllowsOmittedWindow(t *testing.T) {
	project := timeWindowProject(t)

	out, err := NewToolExecutor().ExecuteTool(nil, "analyze_errors", map[string]interface{}{
		"working_dir": project,
	})
	require.NoError(t, err)
	require.NotEmpty(t, out)
}
