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
	"github.com/yaleh/meta-cc/internal/testutil"
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

// setupProjectDirWithMalformedCorpus seeds the checked-in DIR-094 regression
// corpus (one healthy session plus empty / metadata-only-stub / truncated-JSON
// siblings) into a temp project's transcript directory and returns the project
// path to pass as working_dir, the session directory the corpus was seeded
// into, and the file names that must be warned about.
func setupProjectDirWithMalformedCorpus(t *testing.T) (projectPath, sessionDir string, excluded []string) {
	t.Helper()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))
	projectPath = t.TempDir()
	absProject, err := filepath.Abs(projectPath)
	require.NoError(t, err)
	resolvedProject, err := filepath.EvalSymlinks(absProject)
	require.NoError(t, err)
	hash := strings.ReplaceAll(resolvedProject, "\\", "-")
	hash = strings.ReplaceAll(hash, "/", "-")
	hash = strings.ReplaceAll(hash, ":", "-")
	sessionDir = filepath.Join(projectsRoot, hash)
	excluded = testutil.SeedMalformedCorpus(t, sessionDir, resolvedProject)
	return projectPath, sessionDir, excluded
}

// compactJSON re-encodes a tool response so substring assertions are stable
// regardless of whether the response was marshaled compact or indented (Go
// sorts map keys, so the compact form is deterministic). Without this, a
// marker assertion would silently depend on the exact whitespace of the wire
// format, which is not what this test is about.
func compactJSON(t *testing.T, output string) string {
	t.Helper()
	var decoded interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &decoded), "tool output must be valid JSON")
	compact, err := json.Marshal(decoded)
	require.NoError(t, err)
	return string(compact)
}

// corpusTools is the shared per-tool table for the DIR-094 corpus tests: how
// to call each corpus-enumerating analysis path, and which session-derived
// markers must survive. Every wantContains entry is traceable to a specific
// entry of tests/fixtures/malformed-corpus/healthy.jsonl — none of the three
// excluded siblings can produce them (empty and stub carry no message entries
// at all; malformed does not parse).
var corpusTools = []struct {
	name         string
	call         func(*analysis.Service, map[string]interface{}) (string, error)
	wantContains []string
}{
	{
		// One is_error Read tool_result in the healthy session.
		name:         "analyze_errors",
		call:         (*analysis.Service).AnalyzeErrors,
		wantContains: []string{`"total_errors":1`, `"label":"file_not_found"`},
	},
	{
		// The failed Read followed by a successful Read → one fix pair.
		name:         "analyze_bugs",
		call:         (*analysis.Service).AnalyzeBugs,
		wantContains: []string{`"total_pairs":1`, `"fix_count":1`},
	},
	{
		// 3 tool calls, 1 of them errored; 2 distinct tool names.
		name:         "quality_scan",
		call:         (*analysis.Service).QualityScan,
		wantContains: []string{`"name":"error_rate","raw_value":"1/3"`, `"name":"tool_diversity","raw_value":"2/3"`},
	},
	{
		// Grep + 2×Read is the healthy session's whole tool surface.
		name:         "get_work_patterns",
		call:         (*analysis.Service).GetWorkPatterns,
		wantContains: []string{`"tool_name":"Grep"`, `"tool_name":"Read"`},
	},
	{
		// The successful Read's output carries a "// TODO:" comment, which is
		// both a marker and a hotspot attributed to that call's path.
		name:         "get_tech_debt",
		call:         (*analysis.Service).GetTechDebt,
		wantContains: []string{`"label":"TODO"`, `"file":"src/auth/token.js"`},
	},
	{
		// 8 entries spanning 00:00:00Z → 00:00:31Z.
		name:         "get_timeline",
		call:         (*analysis.Service).GetTimeline,
		wantContains: []string{`"total_span":"31s"`},
	},
}

// TestLoadDataWarnings_EveryCorpusPathReportsExcludedFiles is the DIR-094
// coverage proof for the analysis half of the corpus: get_timeline and all
// five analysis tools must return results AND name every excluded session
// file. DIR-018 had already made the unparseable case report; the empty /
// zero-message case — the actual 8eda8f4e shape — parsed cleanly and was
// therefore still dropped in silence, which is exactly the "silent tolerance
// is not the DIR-018 contract" gap this task closes.
//
// Both halves of the AC are asserted per path on purpose. The warning half is
// satisfied by any tool that merely notices the bad files; the results half —
// wantContains markers naming content only healthy.jsonl can contribute — is
// what proves the surviving corpus actually reached the analyzer instead of
// being dropped alongside the bad files.
//
// Driving every tool off one seeded corpus (rather than a per-tool fixture)
// is deliberate: the guarantee is about the shared loadData step, so a tool
// that diverges must fail here rather than quietly keep its own behavior.
func TestLoadDataWarnings_EveryCorpusPathReportsExcludedFiles(t *testing.T) {
	for _, tool := range corpusTools {
		t.Run(tool.name, func(t *testing.T) {
			projectPath, _, excluded := setupProjectDirWithMalformedCorpus(t)
			output, err := tool.call(analysis.New(), map[string]interface{}{"working_dir": projectPath})
			require.NoError(t, err, "a single empty or corrupt session file must not fail the whole tool")
			for _, name := range excluded {
				assertWarningsNameFile(t, output, name)
			}

			compact := compactJSON(t, output)
			for _, want := range tool.wantContains {
				assert.Contains(t, compact, want,
					"the healthy session's data must still reach %s despite the excluded siblings", tool.name)
			}
		})
	}
}

// TestCorpusMarkersRequireTheHealthySession is the non-vacuity guard for the
// wantContains half of TestLoadDataWarnings_EveryCorpusPathReportsExcludedFiles.
// A marker assertion only proves "results still flow" if the marker is
// genuinely absent when there are no results: without this, a marker that
// happens to match a zero-valued field ("total_errors":0 vs a lenient
// substring) would keep passing even after the healthy session stopped
// reaching the analyzer, and the DIR-094 regression would be silent again.
//
// Same corpus, same tool calls — the only difference is that healthy.jsonl is
// removed, leaving exactly the three excluded siblings. Every tool must still
// succeed (the file count is nonzero and all three are tolerated) while every
// marker is gone.
func TestCorpusMarkersRequireTheHealthySession(t *testing.T) {
	for _, tool := range corpusTools {
		t.Run(tool.name, func(t *testing.T) {
			projectPath, sessionDir, excluded := setupProjectDirWithMalformedCorpus(t)
			require.NoError(t, os.Remove(filepath.Join(sessionDir, "healthy.jsonl")))

			output, err := tool.call(analysis.New(), map[string]interface{}{"working_dir": projectPath})
			require.NoError(t, err, "an all-bad corpus is still a tolerable corpus, not a tool failure")
			for _, name := range excluded {
				assertWarningsNameFile(t, output, name)
			}

			compact := compactJSON(t, output)
			for _, want := range tool.wantContains {
				assert.NotContains(t, compact, want,
					"marker %q must be attributable to the healthy session alone, or it proves nothing", want)
			}
		})
	}
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
