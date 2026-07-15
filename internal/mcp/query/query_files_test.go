package query

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pathToHashQ mirrors locator.PathToHash for use in query package tests.
func pathToHashQ(path string) string {
	h := strings.ReplaceAll(path, "\\", "-")
	h = strings.ReplaceAll(h, "/", "-")
	h = strings.ReplaceAll(h, ":", "-")
	return h
}

// setupQueryTestDir creates a temp project dir, sets META_CC_PROJECTS_ROOT, and
// chdirs into the project path.  Returns (projectsRoot, sessionDir, projectPath).
func setupQueryTestDir(t *testing.T) (projectsRoot, sessionDir, projectPath string) {
	t.Helper()

	projectsRoot = t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	projectPath = t.TempDir()
	require.NoError(t, os.Chdir(projectPath))

	resolvedPath, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		resolvedPath = projectPath
	}
	projectHash := pathToHashQ(resolvedPath)

	sessionDir = filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	return projectsRoot, sessionDir, projectPath
}

// TestGetQueryFiles_SessionScope_ReturnsSingleFile verifies that scope="session"
// returns exactly 1 file even when the project dir contains multiple sessions.
func TestGetQueryFiles_SessionScope_ReturnsSingleFile(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Chdir(originalWd)) }()

	_, sessionDir, projectPath := setupQueryTestDir(t)

	// Create two session files so project scope would return 2
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "session1.jsonl"),
		[]byte(`{"type":"user"}`+"\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "session2.jsonl"),
		[]byte(`{"type":"user"}`+"\n"), 0644))

	files, err := GetQueryFiles("session", projectPath, false)
	require.NoError(t, err)

	// session scope must resolve to exactly 1 file (the latest session)
	assert.Len(t, files, 1, "session scope should return exactly 1 file")
}

// TestGetQueryFiles_SessionScope_IncludeSubagents_AppendsSubagentFiles verifies
// that scope="session" + includeSubagents=true returns the main session JSONL plus
// subagent files from <uuid>/subagents/.
func TestGetQueryFiles_SessionScope_IncludeSubagents_AppendsSubagentFiles(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Chdir(originalWd)) }()

	_, sessionDir, projectPath := setupQueryTestDir(t)

	sessionUUID := "abc123session"
	sessionFile := filepath.Join(sessionDir, sessionUUID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(`{"type":"user"}`+"\n"), 0644))

	// Create subagent files for this session
	subDir := filepath.Join(sessionDir, sessionUUID, "subagents")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	sub1 := filepath.Join(subDir, "sub1.jsonl")
	sub2 := filepath.Join(subDir, "sub2.jsonl")
	require.NoError(t, os.WriteFile(sub1, []byte(`{"type":"user"}`+"\n"), 0644))
	require.NoError(t, os.WriteFile(sub2, []byte(`{"type":"user"}`+"\n"), 0644))

	files, err := GetQueryFiles("session", projectPath, true)
	require.NoError(t, err)

	assert.Len(t, files, 3, "session scope + include_subagents=true should return main + 2 subagent files")
	assert.Contains(t, files, sessionFile)
	assert.Contains(t, files, sub1)
	assert.Contains(t, files, sub2)
}

// TestGetQueryFiles_ProjectScope_ExcludeSubagents verifies that scope="project" +
// includeSubagents=false returns only top-level JSONL files.
func TestGetQueryFiles_ProjectScope_ExcludeSubagents(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Chdir(originalWd)) }()

	_, sessionDir, projectPath := setupQueryTestDir(t)

	f1 := filepath.Join(sessionDir, "session1.jsonl")
	f2 := filepath.Join(sessionDir, "session2.jsonl")
	require.NoError(t, os.WriteFile(f1, []byte(`{"type":"user"}`+"\n"), 0644))
	require.NoError(t, os.WriteFile(f2, []byte(`{"type":"user"}`+"\n"), 0644))

	// Add a subagent file that must be excluded
	subDir := filepath.Join(sessionDir, "session1", "subagents")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	subFile := filepath.Join(subDir, "sub.jsonl")
	require.NoError(t, os.WriteFile(subFile, []byte(`{"type":"user"}`+"\n"), 0644))

	files, err := GetQueryFiles("project", projectPath, false)
	require.NoError(t, err)

	assert.Len(t, files, 2, "project scope + include_subagents=false should return only top-level files")
	assert.Contains(t, files, f1)
	assert.Contains(t, files, f2)
	assert.NotContains(t, files, subFile)
}

// TestGetQueryFiles_ProjectScope_IncludeSubagents verifies that scope="project" +
// includeSubagents=true returns all top-level + all subagent JSONL files.
func TestGetQueryFiles_ProjectScope_IncludeSubagents(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Chdir(originalWd)) }()

	_, sessionDir, projectPath := setupQueryTestDir(t)

	session1 := filepath.Join(sessionDir, "session1.jsonl")
	session2 := filepath.Join(sessionDir, "session2.jsonl")
	require.NoError(t, os.WriteFile(session1, []byte(`{"type":"user"}`+"\n"), 0644))
	require.NoError(t, os.WriteFile(session2, []byte(`{"type":"user"}`+"\n"), 0644))

	// Subagents for session1
	subDir1 := filepath.Join(sessionDir, "session1", "subagents")
	require.NoError(t, os.MkdirAll(subDir1, 0755))
	sub1 := filepath.Join(subDir1, "sub1.jsonl")
	require.NoError(t, os.WriteFile(sub1, []byte(`{"type":"user"}`+"\n"), 0644))

	// Subagents for session2
	subDir2 := filepath.Join(sessionDir, "session2", "subagents")
	require.NoError(t, os.MkdirAll(subDir2, 0755))
	sub2 := filepath.Join(subDir2, "sub2.jsonl")
	require.NoError(t, os.WriteFile(sub2, []byte(`{"type":"user"}`+"\n"), 0644))

	files, err := GetQueryFiles("project", projectPath, true)
	require.NoError(t, err)

	assert.Len(t, files, 4, "project scope + include_subagents=true should return 2 top-level + 2 subagent files")
	assert.Contains(t, files, session1)
	assert.Contains(t, files, session2)
	assert.Contains(t, files, sub1)
	assert.Contains(t, files, sub2)
}

// TestGetQueryFiles_SubagentDirScansOnlySubagentsNotToolResults confirms that
// tool-results/ directories are not scanned when include_subagents=true.
func TestGetQueryFiles_SubagentDirScansOnlySubagentsNotToolResults(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Chdir(originalWd)) }()

	_, sessionDir, projectPath := setupQueryTestDir(t)

	sessionUUID := "mysession"
	sessionFile := filepath.Join(sessionDir, sessionUUID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(`{"type":"user"}`+"\n"), 0644))

	// Create subagents dir (should be scanned)
	subDir := filepath.Join(sessionDir, sessionUUID, "subagents")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	subFile := filepath.Join(subDir, "agent.jsonl")
	require.NoError(t, os.WriteFile(subFile, []byte(`{"type":"user"}`+"\n"), 0644))

	// Create tool-results dir (should NOT be scanned)
	toolResultsDir := filepath.Join(sessionDir, sessionUUID, "tool-results")
	require.NoError(t, os.MkdirAll(toolResultsDir, 0755))
	toolFile := filepath.Join(toolResultsDir, "result.jsonl")
	require.NoError(t, os.WriteFile(toolFile, []byte(`{"type":"tool"}`+"\n"), 0644))

	files, err := GetQueryFiles("project", projectPath, true)
	require.NoError(t, err)

	assert.Contains(t, files, sessionFile, "top-level session file should be included")
	assert.Contains(t, files, subFile, "subagent file should be included")
	assert.NotContains(t, files, toolFile, "tool-results file should NOT be included")
}
