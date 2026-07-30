package analysis_test

// DIR-018 regression tests: loadData used to silently `continue` when
// ParseEntries failed on a session file, so analysis results were computed
// from whatever subset of the corpus happened to parse, with no evidence of
// the exclusion anywhere in the MCP tool response. These tests prove that:
//
//  1. A malformed session file in the corpus produces a "warnings" array in
//     the tool response naming that file (full and stats_only paths).
//  2. A clean corpus produces no "warnings" key at all (omitempty keeps the
//     wire format unchanged when nothing was excluded).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/analysis"
	"github.com/yaleh/meta-cc/internal/types"
)

// setupProjectDirWithMalformedFile seeds a project session directory with one
// valid session file (containing the given entries) and one malformed file
// whose contents cannot be parsed as JSONL. It returns the project path and
// the malformed file's name.
func setupProjectDirWithMalformedFile(t *testing.T, entries []types.SessionEntry) (string, string) {
	t.Helper()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))
	projectPath := t.TempDir()
	absProject, err := filepath.Abs(projectPath)
	require.NoError(t, err)
	resolvedProject, err := filepath.EvalSymlinks(absProject)
	require.NoError(t, err)
	hash := strings.ReplaceAll(resolvedProject, "\\", "-")
	hash = strings.ReplaceAll(hash, "/", "-")
	hash = strings.ReplaceAll(hash, ":", "-")
	sessionDir := filepath.Join(projectsRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	var lines []string
	for _, e := range entries {
		data, err := json.Marshal(e)
		require.NoError(t, err)
		lines = append(lines, string(data))
	}
	validFile := filepath.Join(sessionDir, "valid-session.jsonl")
	require.NoError(t, os.WriteFile(validFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644))

	malformedName := "corrupted-session.jsonl"
	malformedFile := filepath.Join(sessionDir, malformedName)
	require.NoError(t, os.WriteFile(malformedFile, []byte("{this is not valid json\n"), 0o644))
	return projectPath, malformedName
}

func validUserEntry() types.SessionEntry {
	return types.SessionEntry{
		Type:      "user",
		UUID:      "u1",
		Timestamp: "2026-01-01T00:00:00.000Z",
		Message: &types.Message{
			Role:    "user",
			Content: []types.ContentBlock{{Type: "text", Text: "hello"}},
		},
	}
}

// assertWarningsNameFile unmarshals a tool response and asserts it carries a
// warnings array with at least one entry naming the malformed file.
func assertWarningsNameFile(t *testing.T, output, malformedName string) {
	t.Helper()
	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &decoded), "tool output must be valid JSON")
	warnings, ok := decoded["warnings"].([]interface{})
	require.True(t, ok, "response must contain a warnings array when files were skipped; got: %s", output)
	require.NotEmpty(t, warnings, "warnings array must not be empty when a file was skipped")
	found := false
	for _, w := range warnings {
		if s, ok := w.(string); ok && strings.Contains(s, malformedName) {
			found = true
			break
		}
	}
	assert.True(t, found, "at least one warning must name the skipped file %s; got %v", malformedName, warnings)
}

func TestLoadDataWarnings_MalformedFileSurfacedInAnalyzeBugs(t *testing.T) {
	projectPath, malformedName := setupProjectDirWithMalformedFile(t, []types.SessionEntry{validUserEntry()})
	svc := analysis.New()
	output, err := svc.AnalyzeBugs(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err, "a malformed sibling file must not fail the whole analysis")
	assertWarningsNameFile(t, output, malformedName)
}

func TestLoadDataWarnings_MalformedFileSurfacedInStatsOnly(t *testing.T) {
	projectPath, malformedName := setupProjectDirWithMalformedFile(t, []types.SessionEntry{validUserEntry()})
	svc := analysis.New()
	output, err := svc.AnalyzeErrors(map[string]interface{}{"working_dir": projectPath, "stats_only": true})
	require.NoError(t, err)
	assertWarningsNameFile(t, output, malformedName)
}

func TestLoadDataWarnings_MalformedFileSurfacedInGetTimeline(t *testing.T) {
	projectPath, malformedName := setupProjectDirWithMalformedFile(t, []types.SessionEntry{validUserEntry()})
	svc := analysis.New()
	output, err := svc.GetTimeline(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	assertWarningsNameFile(t, output, malformedName)
}

func TestLoadDataWarnings_CleanCorpusOmitsWarnings(t *testing.T) {
	projectPath := setupProjectDirWithEntries(t, []types.SessionEntry{validUserEntry()})
	svc := analysis.New()
	output, err := svc.AnalyzeBugs(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &decoded))
	_, present := decoded["warnings"]
	assert.False(t, present, "clean corpus must not emit a warnings key (omitempty); got: %s", output)
}
