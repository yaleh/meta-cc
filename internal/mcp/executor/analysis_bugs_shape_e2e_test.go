package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DIR-096 end-to-end regression, through the REAL analyze_bugs MCP tool
// handler: the tool's new max_patterns parameter must be reachable from the
// MCP call site (a schema parameter nothing reads is the "declared but
// ignored" failure class this repo already gates against), and the per-pattern
// examples a caller receives must be attributable objects rather than the bare
// error strings that used to be emitted.
//
// The fixture mirrors the live blowup at small scale: several distinct error
// signatures, each occurrence carrying a multi-kilobyte message, so an
// unbounded response could not stay inline.
func setupAnalyzeBugsShapeFixtureProject(t *testing.T, distinct int) (projectPath, sessionID string) {
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

	sessionID = "analyze-bugs-shape-fixture-session"
	var lines []string
	for p := 0; p < distinct; p++ {
		errID := fmt.Sprintf("toolu_err_%d", p)
		okID := fmt.Sprintf("toolu_ok_%d", p)
		ts := fmt.Sprintf("2026-01-01T00:00:%02dZ", p*4)
		okTS := fmt.Sprintf("2026-01-01T00:00:%02dZ", p*4+1)
		errText := fmt.Sprintf("boom pattern-%d: %s", p, strings.Repeat("e", 3000))
		lines = append(lines,
			`{"type":"assistant","timestamp":"`+ts+`","sessionId":"`+sessionID+`","cwd":"`+resolvedProject+`","message":{"role":"assistant","content":[{"type":"tool_use","id":"`+errID+`","name":"Bash","input":{"command":"boom"}}]}}`,
			`{"type":"user","timestamp":"`+ts+`","sessionId":"`+sessionID+`","cwd":"`+resolvedProject+`","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"`+errID+`","is_error":true,"content":"`+errText+`"}]}}`,
			`{"type":"assistant","timestamp":"`+okTS+`","sessionId":"`+sessionID+`","cwd":"`+resolvedProject+`","message":{"role":"assistant","content":[{"type":"tool_use","id":"`+okID+`","name":"Bash","input":{"command":"fixed"}}]}}`,
			`{"type":"user","timestamp":"`+okTS+`","sessionId":"`+sessionID+`","cwd":"`+resolvedProject+`","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"`+okID+`","is_error":false,"content":"the fix worked"}]}}`,
		)
	}

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644))

	return resolvedProject, sessionID
}

type e2eBugResult struct {
	Patterns []struct {
		ErrorSignature string `json:"error_signature"`
		Examples       []struct {
			SessionID string `json:"session_id"`
			Timestamp string `json:"timestamp"`
			ErrorText string `json:"error_text"`
			FixText   string `json:"fix_text"`
			Signature string `json:"signature"`
		} `json:"examples"`
	} `json:"patterns"`
	TotalPatterns     int `json:"total_patterns"`
	TruncatedPatterns int `json:"truncated_patterns"`
}

func TestAnalyzeBugs_EndToEnd_MaxPatternsAndStructuredExamples(t *testing.T) {
	const distinct = 6
	projectPath, sessionID := setupAnalyzeBugsShapeFixtureProject(t, distinct)
	e := NewToolExecutor()

	out, err := e.ExecuteTool(nil, "analyze_bugs", map[string]interface{}{
		"working_dir":  projectPath,
		"max_patterns": float64(2),
	})
	require.NoError(t, err)

	var res e2eBugResult
	require.NoError(t, json.Unmarshal([]byte(out), &res))
	assert.Len(t, res.Patterns, 2, "max_patterns=2 must be honored end-to-end, not just declared in the schema")
	assert.Equal(t, distinct, res.TotalPatterns, "the uncapped pattern count must still be reported")
	assert.Equal(t, distinct-2, res.TruncatedPatterns)
	assert.Less(t, len(out), 32768, "a capped response must stay inline")

	// Every example is an attributable object, not the bare error string the
	// tool emitted before DIR-096.
	noExamples := 0
	for _, p := range res.Patterns {
		if len(p.Examples) == 0 {
			noExamples++
			continue
		}
		for _, ex := range p.Examples {
			assert.Equal(t, sessionID, ex.SessionID, "example must name the session it was observed in")
			assert.NotEmpty(t, ex.Timestamp, "example must carry its timestamp")
			assert.Contains(t, ex.ErrorText, "boom pattern-", "example must carry the error text")
			assert.Equal(t, "the fix worked", ex.FixText, "example must carry the paired fix output")
			assert.Equal(t, p.ErrorSignature, ex.Signature, "example must carry its own signature")
		}
	}
	assert.Zero(t, noExamples, "each of the seeded patterns has an occurrence, so none may come back example-less")
}

func TestAnalyzeBugs_EndToEnd_DefaultFitsInlineThreshold(t *testing.T) {
	projectPath, _ := setupAnalyzeBugsShapeFixtureProject(t, 30)
	e := NewToolExecutor()

	out, err := e.ExecuteTool(nil, "analyze_bugs", map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	assert.Less(t, len(out), 32768,
		"a default analyze_bugs call must stay inline (<32768 bytes), got %d", len(out))

	var res e2eBugResult
	require.NoError(t, json.Unmarshal([]byte(out), &res))
	assert.LessOrEqual(t, len(res.Patterns), 20, "the default max_patterns must cap the pattern list")
	assert.NotEmpty(t, res.Patterns)
}
