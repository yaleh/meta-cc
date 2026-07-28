package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
)

// contains_cross_role_test.go is the DIR-062 regression: query_session_content's
// `contains` parameter is documented (tools.go) as a general literal-substring
// filter on message content, but the implementation honored it ONLY for
// role=assistant. For role=user, role=tool, and role=all the parameter was
// accepted and silently dropped — a never-matching `contains` returned the FULL
// result set instead of 0 records, silently lying to callers who believed they
// had filtered (tasks/DIR-062.md).
//
// The fix (option 1, preferred by the task) honors `contains` cross-role by
// routing every role through one shared clause builder (containsClause in
// handlers.go) that applies the same regexp.QuoteMeta + EscapeJQ + jq
// test(...; "i") path as the DIR-047 assistant fix, so escaping never diverges
// per branch and "." stays literal for every role ("main.go" must not match
// "mainXgo").

// setupContainsCrossRoleFixtureProject wires a temporary Claude projects root
// with a session covering every role:
//
//	R1 user string content: "alpha_token" + "main.go"
//	R2 user string content: "mainXgo" (regex-wildcard trap for "main.go")
//	R3 assistant tool_use blocks: Bash{command:"echo beta_token"}, Read{file_path:"src/mainXgo"}
//	R4 assistant tool_use block:  Read{file_path:"src/main.go"}
//	R5 user array content: tool_result{tool_use_id:"toolu_1", content:"output with gamma_token here"}
//	R6 user array content: tool_result{tool_use_id:"toolu_2", content:"plain output no tokens"}
//	R7 assistant string content: "delta_token"
func setupContainsCrossRoleFixtureProject(t *testing.T) (projectPath string) {
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

	sessionID := "contains-cross-role-session"
	lines := []string{
		`{"type":"user","timestamp":"2026-01-01T00:00:00Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":"first user message alpha_token open main.go now"}}`,
		`{"type":"user","timestamp":"2026-01-01T00:00:01Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":"second user message talks about mainXgo only"}}`,
		`{"type":"assistant","timestamp":"2026-01-01T00:00:02Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_1","name":"Bash","input":{"command":"echo beta_token"}},{"type":"tool_use","id":"toolu_2","name":"Read","input":{"file_path":"src/mainXgo"}}]}}`,
		`{"type":"assistant","timestamp":"2026-01-01T00:00:03Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_3","name":"Read","input":{"file_path":"src/main.go"}}]}}`,
		`{"type":"user","timestamp":"2026-01-01T00:00:04Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"output with gamma_token here"}]}}`,
		`{"type":"user","timestamp":"2026-01-01T00:00:05Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_2","content":"plain output no tokens"}]}}`,
		`{"type":"assistant","timestamp":"2026-01-01T00:00:06Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":"final assistant note delta_token"}}`,
	}

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644))

	return resolvedProject
}

// runContainsQuery executes query_session_content with the given args against
// the fixture project and returns the data array.
func runContainsQuery(t *testing.T, projectPath string, args map[string]interface{}) []map[string]interface{} {
	t.Helper()
	e := NewToolExecutor()
	cfg := &config.Config{}
	args["working_dir"] = projectPath
	out, err := e.ExecuteTool(cfg, "query_session_content", args)
	require.NoError(t, err)
	return extractDataArray(t, out)
}

// recordJSON marshals a result record so assertions can check which substrings
// it contains regardless of the per-role output shape.
func recordJSON(t *testing.T, rec map[string]interface{}) string {
	t.Helper()
	b, err := json.Marshal(rec)
	require.NoError(t, err)
	return string(b)
}

// TestContainsCrossRole_FixtureSanity proves the fixture actually contains
// unfiltered data for every role, so the 0-record assertions below are
// meaningful (a filter dropped everything vs. there was nothing to find).
func TestContainsCrossRole_FixtureSanity(t *testing.T) {
	projectPath := setupContainsCrossRoleFixtureProject(t)

	users := runContainsQuery(t, projectPath, map[string]interface{}{"role": "user"})
	require.Len(t, users, 2, "fixture must have 2 string-content user messages")

	toolUse := runContainsQuery(t, projectPath, map[string]interface{}{"role": "tool", "block_type": "tool_use"})
	require.Len(t, toolUse, 3, "fixture must have 3 tool_use blocks")

	toolResult := runContainsQuery(t, projectPath, map[string]interface{}{"role": "tool", "block_type": "tool_result"})
	require.Len(t, toolResult, 2, "fixture must have 2 tool_result blocks")

	all := runContainsQuery(t, projectPath, map[string]interface{}{"role": "all"})
	require.Len(t, all, 7, "fixture must have 7 user+assistant flow records")
}

// ─── role=user ────────────────────────────────────────────────────────────────

// DIR-062 AC1/AC2 (user half): a contains string present in no record must
// return 0 records, not the full set.
func TestContainsCrossRole_User_NeverMatchingReturnsZero(t *testing.T) {
	projectPath := setupContainsCrossRoleFixtureProject(t)

	data := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":     "user",
		"contains": "ZZZ_SHOULD_MATCH_NOTHING_999",
	})
	require.Len(t, data, 0,
		"role=user with a never-matching contains must return 0 records, got the full set: %d", len(data))
}

// DIR-062 AC3 (user half): a known-present substring returns ONLY matching
// records.
func TestContainsCrossRole_User_PositiveControl(t *testing.T) {
	projectPath := setupContainsCrossRoleFixtureProject(t)

	data := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":     "user",
		"contains": "alpha_token",
	})
	require.Len(t, data, 1, "exactly one user message contains alpha_token")
	msg, ok := data[0]["message"].(map[string]interface{})
	require.True(t, ok, "expected message field: %#v", data[0])
	content, _ := msg["content"].(string)
	require.Contains(t, content, "alpha_token")
}

// DIR-062 AC4 (user half): the new role uses the identical QuoteMeta+EscapeJQ
// path as the assistant branch — "." is literal, so "main.go" must NOT match
// "mainXgo".
func TestContainsCrossRole_User_EscapingIsLiteral(t *testing.T) {
	projectPath := setupContainsCrossRoleFixtureProject(t)

	data := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":     "user",
		"contains": "main.go",
	})
	require.Len(t, data, 1,
		"contains=\"main.go\" must match only the literal \"main.go\" user message, not \"mainXgo\" too")
	msg := data[0]["message"].(map[string]interface{})
	content, _ := msg["content"].(string)
	require.Contains(t, content, "main.go")
	require.NotContains(t, content, "mainXgo")
}

// ─── role=tool, block_type=tool_use ──────────────────────────────────────────

func TestContainsCrossRole_ToolUse_NeverMatchingReturnsZero(t *testing.T) {
	projectPath := setupContainsCrossRoleFixtureProject(t)

	data := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":       "tool",
		"block_type": "tool_use",
		"contains":   "ZZZ_SHOULD_MATCH_NOTHING_999",
	})
	require.Len(t, data, 0,
		"role=tool (tool_use) with a never-matching contains must return 0 records, got: %d", len(data))
}

func TestContainsCrossRole_ToolUse_PositiveControl(t *testing.T) {
	projectPath := setupContainsCrossRoleFixtureProject(t)

	// Match on the block's input content.
	data := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":       "tool",
		"block_type": "tool_use",
		"contains":   "beta_token",
	})
	require.Len(t, data, 1, "exactly one tool_use block has beta_token in its input")
	require.Equal(t, "Bash", data[0]["name"])

	// Match on the block's name field.
	byName := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":       "tool",
		"block_type": "tool_use",
		"contains":   "Read",
	})
	require.Len(t, byName, 2, "exactly two tool_use blocks are named Read")
	for _, rec := range byName {
		require.Equal(t, "Read", rec["name"])
	}
}

func TestContainsCrossRole_ToolUse_EscapingIsLiteral(t *testing.T) {
	projectPath := setupContainsCrossRoleFixtureProject(t)

	data := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":       "tool",
		"block_type": "tool_use",
		"contains":   "main.go",
	})
	require.Len(t, data, 1,
		"contains=\"main.go\" must match only the tool_use whose input has \"main.go\", not \"mainXgo\"")
	rec := recordJSON(t, data[0])
	require.Contains(t, rec, "main.go")
	require.NotContains(t, rec, "mainXgo")
}

// ─── role=tool, block_type=tool_result ───────────────────────────────────────

func TestContainsCrossRole_ToolResult_NeverMatchingReturnsZero(t *testing.T) {
	projectPath := setupContainsCrossRoleFixtureProject(t)

	data := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":       "tool",
		"block_type": "tool_result",
		"contains":   "ZZZ_SHOULD_MATCH_NOTHING_999",
	})
	require.Len(t, data, 0,
		"role=tool (tool_result) with a never-matching contains must return 0 records, got: %d", len(data))
}

func TestContainsCrossRole_ToolResult_PositiveControl(t *testing.T) {
	projectPath := setupContainsCrossRoleFixtureProject(t)

	data := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":       "tool",
		"block_type": "tool_result",
		"contains":   "gamma_token",
	})
	require.Len(t, data, 1, "exactly one tool_result block has gamma_token in its content")
	require.Equal(t, "toolu_1", data[0]["tool_use_id"])
	content, _ := data[0]["content"].(string)
	require.Contains(t, content, "gamma_token")
}

// ─── role=all ─────────────────────────────────────────────────────────────────

func TestContainsCrossRole_All_NeverMatchingReturnsZero(t *testing.T) {
	projectPath := setupContainsCrossRoleFixtureProject(t)

	data := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":     "all",
		"contains": "ZZZ_SHOULD_MATCH_NOTHING_999",
	})
	require.Len(t, data, 0,
		"role=all with a never-matching contains must return 0 records, got: %d", len(data))
}

func TestContainsCrossRole_All_PositiveControl(t *testing.T) {
	projectPath := setupContainsCrossRoleFixtureProject(t)

	// Assistant string content is matched.
	data := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":     "all",
		"contains": "delta_token",
	})
	require.Len(t, data, 1, "exactly one flow record contains delta_token")
	require.Equal(t, "assistant", data[0]["type"])

	// User string content is matched, and ONLY matching records come back.
	userHit := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":     "all",
		"contains": "alpha_token",
	})
	require.Len(t, userHit, 1, "exactly one flow record contains alpha_token")
	require.Equal(t, "user", userHit[0]["type"])

	// Array content (tool_result blocks inside user records) is matched via its
	// JSON text representation.
	arrHit := runContainsQuery(t, projectPath, map[string]interface{}{
		"role":     "all",
		"contains": "gamma_token",
	})
	require.Len(t, arrHit, 1, "exactly one flow record contains gamma_token")
	require.Equal(t, "user", arrHit[0]["type"])
}
