package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
)

// substring_not_regex_test.go is the DIR-047 regression: query_session_content's
// `contains` parameter (role=assistant) is documented as a literal substring
// filter ("Optional substring filter applied to message content
// (case-insensitive)"), but consolidated_handlers.go's handleQuerySessionContent
// passed the raw user-supplied string through EscapeJQ (which only escapes `\`
// and `"` for jq string-literal safety) and straight into jq's test() — a regex
// match function. Any regex metacharacter in the input (most commonly `.`) was
// therefore interpreted as a wildcard instead of a literal character, so a
// search for "main.go" also matched unrelated content like "mainXgo". This is
// the same defect DIR-005 (tasks/DIR-005.md) found and fixed, but whose fix
// never made it into main (unmerged milestone branch) before the task was
// marked done — see tasks/DIR-047.md.
//
// The fix wraps the raw substring in regexp.QuoteMeta before EscapeJQ, so the
// compiled jq pattern is a truly literal, regex-safe substring match.

// setupSubstringNotRegexFixtureProject wires a temporary Claude projects root
// with two assistant messages: one containing the literal substring
// "mainXgo" and one containing the literal substring "main.go". A correct
// literal-substring `contains: "main.go"` filter must match only the second.
func setupSubstringNotRegexFixtureProject(t *testing.T) (projectPath string) {
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

	sessionID := "substring-not-regex-session"
	lines := []string{
		`{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":"please rename mainXgo before continuing"}}`,
		`{"type":"assistant","timestamp":"2026-01-01T00:00:01Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":"please open main.go before continuing"}}`,
	}

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644))

	return resolvedProject
}

// TestSubstringNotRegex_QuerySessionContent_ContainsIsLiteralNotRegex is the
// DIR-047 regression proving query_session_content(role="assistant",
// contains="main.go") matches only the literal substring "main.go" and does
// NOT match "mainXgo" (which an unescaped-regex `.` wildcard would match).
func TestSubstringNotRegex_QuerySessionContent_ContainsIsLiteralNotRegex(t *testing.T) {
	projectPath := setupSubstringNotRegexFixtureProject(t)
	e := NewToolExecutor()
	cfg := &config.Config{}

	out, err := e.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":        "assistant",
		"contains":    "main.go",
		"working_dir": projectPath,
	})
	require.NoError(t, err)

	data := extractDataArray(t, out)
	require.Len(t, data, 1, "contains=\"main.go\" must match exactly the literal \"main.go\" record, not \"mainXgo\" too; got: %s", out)

	msg, ok := data[0]["message"].(map[string]interface{})
	require.True(t, ok, "expected message field in result: %#v", data[0])
	content, _ := msg["content"].(string)
	require.Contains(t, content, "main.go")
	require.NotContains(t, content, "mainXgo")
}
