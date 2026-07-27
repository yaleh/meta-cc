package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupAnalysisBoundaryProject seeds one minimal Claude session under its
// own project-hash directory (rooted at projectsRoot) with the given
// sessionID, and returns the resolved project path to use as working_dir.
// Mirrors the seeding approach used by setupClaudeSessionFixtureProject in
// query_sessions_handler_test.go, but allows multiple distinct projects to
// be seeded under one shared projectsRoot (needed to exercise a genuine
// cross-project session_id combination in a single test).
func setupAnalysisBoundaryProject(t *testing.T, projectsRoot, sessionID string) (resolvedProject string) {
	t.Helper()

	project := t.TempDir()
	absProject, err := filepath.Abs(project)
	require.NoError(t, err)
	resolvedProject, err = filepath.EvalSymlinks(absProject)
	require.NoError(t, err)

	hash := strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolvedProject)
	sessionDir := filepath.Join(projectsRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	content := `{"type":"user","timestamp":"2026-01-01T00:00:00Z","sessionId":"` + sessionID +
		`","cwd":"` + resolvedProject + `","message":{"role":"user","content":"hello"}}` + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, sessionID+".jsonl"), []byte(content), 0o644))

	return resolvedProject
}

// TestAnalyzeErrors_EndToEnd_RejectsCrossProjectSessionID is the DIR-032
// end-to-end regression proof, through the REAL analyze_errors MCP tool
// handler (one of the six loadData-backed tools sharing this gap:
// analyze_errors, analyze_bugs, quality_scan, get_work_patterns,
// get_timeline, get_tech_debt), that a working_dir scoped to project A
// cannot read project B's session content merely by supplying project B's
// session_id. Before the internal/analysis/service.go fix, loadData's
// sessionID != "" branch called locator.FromSessionID with zero comparison
// against working_dir, so this combination would have succeeded and leaked
// project B's content.
func TestAnalyzeErrors_EndToEnd_RejectsCrossProjectSessionID(t *testing.T) {
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	projectA := setupAnalysisBoundaryProject(t, projectsRoot, "session-in-project-a")
	setupAnalysisBoundaryProject(t, projectsRoot, "session-in-project-b")

	_, err := NewToolExecutor().ExecuteTool(nil, "analyze_errors", map[string]interface{}{
		"working_dir": projectA,
		"session_id":  "session-in-project-b",
	})
	require.Error(t, err, "analyze_errors must reject a session_id outside the working_dir boundary")
}

// TestAnalyzeErrors_EndToEnd_AllowsSameProjectSessionID is the sanity
// counterpart proving the DIR-032 boundary fix does not spuriously break
// the legitimate same-project case for analyze_errors, exercised through
// the same real MCP tool handler used above (not just Service directly).
func TestAnalyzeErrors_EndToEnd_AllowsSameProjectSessionID(t *testing.T) {
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	projectA := setupAnalysisBoundaryProject(t, projectsRoot, "session-in-project-a")

	out, err := NewToolExecutor().ExecuteTool(nil, "analyze_errors", map[string]interface{}{
		"working_dir": projectA,
		"session_id":  "session-in-project-a",
	})
	require.NoError(t, err)
	require.NotEmpty(t, out)
}
