package analysis_test

// DIR-094 regression tests: every corpus-enumerating analysis path must
// tolerate one unusable session file by returning results PLUS a warning
// naming it — never a whole-batch error, and never a silent shorter result set.
//
// The gap this closes is the *second* half. DIR-018 (service_warnings_test.go)
// already covered a file that fails to PARSE. It did not cover a file that
// parses cleanly into zero message entries — the 0-byte / metadata-only
// session stub, which is exactly the shape of the 2026-07-30 dogfooding
// failure that made query_sessions hard-fail. ParseEntries returns no error
// for such a file, so loadData dropped it with no trace at all.
//
// The corpus is checked in (internal/analysis/testdata/dir094/) rather than
// built inline, so the tolerance cannot be lost by a future rewrite that
// quietly stops exercising these shapes.
//
// The central assertion is an equality one: because a zero-message file
// contributes nothing, the results from a corpus containing it must be
// identical to the results from the clean corpus — proving both that the bad
// file neither erased nor altered any result, and that the ONLY difference is
// the reported exclusion metadata.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/analysis"
)

const dir094FixtureDir = "testdata/dir094"

// dir094EmptySessionID is the session file name from the 2026-07-30
// dogfooding failure. The 0-byte fixture is installed under this name so the
// regression test reproduces the reported shape verbatim.
const dir094EmptySessionID = "8eda8f4e-2c74-4176-ba6b-8c45e890df42"

// dir094BadFixtures are the corpus files that contribute nothing and must be
// reported as skipped.
var dir094BadFixtures = []string{
	"empty-session.jsonl",
	"metadata-only-session.jsonl",
	"malformed-session.jsonl",
	"whitespace-session.jsonl",
}

// toolCall invokes one analysis tool through the Service, mirroring how the
// MCP handler reaches it.
type toolCall struct {
	name string
	call func(svc *analysis.Service, args map[string]interface{}) (string, error)
}

// dir094Tools enumerates every corpus-enumerating analysis path. get_timeline
// is listed alongside the five analysis tools because it is the one the
// dogfooding session fell back to, and it shares loadData with the rest —
// naming it explicitly is what makes "all paths" checkable rather than
// assumed.
func dir094Tools() []toolCall {
	return []toolCall{
		{"analyze_errors", (*analysis.Service).AnalyzeErrors},
		{"analyze_bugs", (*analysis.Service).AnalyzeBugs},
		{"quality_scan", (*analysis.Service).QualityScan},
		{"get_work_patterns", (*analysis.Service).GetWorkPatterns},
		{"get_tech_debt", (*analysis.Service).GetTechDebt},
		{"get_timeline", (*analysis.Service).GetTimeline},
	}
}

// setupDir094Corpus installs the checked-in fixture corpus into a fresh
// project-hash transcript directory and returns the project path.
//
// withValid=false installs only the zero-message files, giving the
// "nothing queryable here" baseline; withBadFiles=false installs only the
// valid session, giving the clean-corpus control. Both controls are what make
// the tolerance assertions meaningful rather than merely "the call returned".
func setupDir094Corpus(t *testing.T, withValid, withBadFiles bool) string {
	t.Helper()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	projectPath := t.TempDir()
	resolvedProject, err := filepath.EvalSymlinks(projectPath)
	require.NoError(t, err)
	sessionDir := filepath.Join(projectsRoot, strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolvedProject))
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	install := func(fixture, name string) {
		data, err := os.ReadFile(filepath.Join(dir094FixtureDir, fixture))
		require.NoError(t, err, "fixture %s must be checked in", fixture)
		require.NoError(t, os.WriteFile(filepath.Join(sessionDir, name+".jsonl"), data, 0o644))
	}
	if withValid {
		install("valid-session.jsonl", "valid-session")
	}
	if withBadFiles {
		install("empty-session.jsonl", dir094EmptySessionID)
		for _, fixture := range dir094BadFixtures[1:] {
			install(fixture, strings.TrimSuffix(fixture, ".jsonl"))
		}
	}
	return resolvedProject
}

// decodeResult unmarshals a tool response and strips the exclusion metadata,
// so the remainder can be compared against the clean-corpus control.
func decodeResult(t *testing.T, output string) map[string]interface{} {
	t.Helper()
	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &decoded), "tool output must be a JSON object; got: %s", output)
	return decoded
}

func withoutExclusionMetadata(decoded map[string]interface{}) map[string]interface{} {
	stripped := make(map[string]interface{}, len(decoded))
	for k, v := range decoded {
		if k == "warnings" || k == "skipped_files" {
			continue
		}
		stripped[k] = v
	}
	return stripped
}

// TestLoadDataTolerance_EveryPathReturnsResultsPlusWarning is the AC1 test: on
// each corpus-enumerating path, the 8eda8f4e-style empty file (plus the other
// zero-message shapes) yields the same results as a clean corpus, together
// with a warning that names the excluded file.
func TestLoadDataTolerance_EveryPathReturnsResultsPlusWarning(t *testing.T) {
	for _, tool := range dir094Tools() {
		t.Run(tool.name, func(t *testing.T) {
			cleanProject := setupDir094Corpus(t, true, false)
			cleanOutput, err := tool.call(analysis.New(), map[string]interface{}{"working_dir": cleanProject})
			require.NoError(t, err, "%s on a clean corpus must succeed", tool.name)

			// Baseline: the same tool over a corpus of ONLY zero-message
			// files. If a tool's output equals this, it returned nothing —
			// which is precisely what "the bad file erased the results"
			// would look like.
			badOnlyProject := setupDir094Corpus(t, false, true)
			badOnlyOutput, err := tool.call(analysis.New(), map[string]interface{}{"working_dir": badOnlyProject})
			require.NoError(t, err, "%s must not fail a batch of only zero-message files either", tool.name)

			dirtyProject := setupDir094Corpus(t, true, true)
			dirtyOutput, err := tool.call(analysis.New(), map[string]interface{}{"working_dir": dirtyProject})
			require.NoError(t, err, "%s must tolerate a zero-message sibling file, not fail the whole batch", tool.name)

			cleanDecoded := decodeResult(t, cleanOutput)
			dirtyDecoded := decodeResult(t, dirtyOutput)
			badOnlyDecoded := decodeResult(t, badOnlyOutput)

			// Results PRESENT: the valid session genuinely reaches the
			// result, so the output is distinguishable from the
			// zero-message-only baseline.
			assert.NotEqual(t, withoutExclusionMetadata(badOnlyDecoded), withoutExclusionMetadata(cleanDecoded),
				"%s must reflect the valid session's data", tool.name)

			// Results UNCHANGED: the zero-message files contributed nothing,
			// so every non-metadata field must match the clean-corpus control
			// exactly — the bad file neither altered nor shortened anything.
			assert.Equal(t,
				withoutExclusionMetadata(cleanDecoded),
				withoutExclusionMetadata(dirtyDecoded),
				"%s results must be identical with and without the zero-message files", tool.name)

			// Warning naming the file, plus the machine-readable path list.
			warnings, ok := dirtyDecoded["warnings"].([]interface{})
			require.True(t, ok, "%s response must carry a warnings array; got: %s", tool.name, dirtyOutput)
			skipped, ok := dirtyDecoded["skipped_files"].([]interface{})
			require.True(t, ok, "%s response must carry a skipped_files array; got: %s", tool.name, dirtyOutput)

			for _, fixture := range dir094BadFixtures {
				name := strings.TrimSuffix(fixture, ".jsonl")
				if fixture == "empty-session.jsonl" {
					name = dir094EmptySessionID
				}
				assert.True(t, anyStringContains(warnings, name),
					"%s: a warning must name the skipped file %s; got %v", tool.name, name, warnings)
				assert.True(t, anyStringContains(skipped, name),
					"%s: skipped_files must list %s; got %v", tool.name, name, skipped)
			}
		})
	}
}

// TestLoadDataTolerance_StatsOnlyPathsAlsoReport is the same contract on the
// stats_only short-circuit, which runs a different analyzer path but shares
// loadData.
func TestLoadDataTolerance_StatsOnlyPathsAlsoReport(t *testing.T) {
	for _, tool := range dir094Tools() {
		t.Run(tool.name, func(t *testing.T) {
			project := setupDir094Corpus(t, true, true)
			output, err := tool.call(analysis.New(), map[string]interface{}{
				"working_dir": project,
				"stats_only":  true,
			})
			require.NoError(t, err, "%s stats_only must tolerate a zero-message sibling file", tool.name)

			decoded := decodeResult(t, output)
			skipped, ok := decoded["skipped_files"].([]interface{})
			require.True(t, ok, "%s stats_only response must carry skipped_files; got: %s", tool.name, output)
			assert.True(t, anyStringContains(skipped, dir094EmptySessionID),
				"%s stats_only must list the 8eda8f4e-style empty file; got %v", tool.name, skipped)
		})
	}
}

// TestLoadDataTolerance_CleanCorpusOmitsExclusionMetadata keeps the wire format
// unchanged when nothing was excluded: no warnings, no skipped_files key.
func TestLoadDataTolerance_CleanCorpusOmitsExclusionMetadata(t *testing.T) {
	for _, tool := range dir094Tools() {
		t.Run(tool.name, func(t *testing.T) {
			project := setupDir094Corpus(t, true, false)
			output, err := tool.call(analysis.New(), map[string]interface{}{"working_dir": project})
			require.NoError(t, err)

			decoded := decodeResult(t, output)
			_, hasWarnings := decoded["warnings"]
			_, hasSkipped := decoded["skipped_files"]
			assert.False(t, hasWarnings, "%s: a clean corpus must not emit warnings; got %s", tool.name, output)
			assert.False(t, hasSkipped, "%s: a clean corpus must not emit skipped_files; got %s", tool.name, output)
		})
	}
}

// TestLoadDataTolerance_SessionScopeNamedEmptyFile covers the exact-session
// selector: asking for the one session that happens to be empty must report
// that fact rather than quietly returning an empty result.
func TestLoadDataTolerance_SessionScopeNamedEmptyFile(t *testing.T) {
	project := setupDir094Corpus(t, true, true)
	output, err := analysis.New().AnalyzeBugs(map[string]interface{}{
		"working_dir": project,
		"session_id":  dir094EmptySessionID,
	})
	require.NoError(t, err)

	decoded := decodeResult(t, output)
	skipped, ok := decoded["skipped_files"].([]interface{})
	require.True(t, ok, "an exact-session read of an empty file must report the exclusion; got: %s", output)
	assert.True(t, anyStringContains(skipped, dir094EmptySessionID), "skipped_files = %v", skipped)
}

func anyStringContains(values []interface{}, needle string) bool {
	for _, v := range values {
		if s, ok := v.(string); ok && strings.Contains(s, needle) {
			return true
		}
	}
	return false
}
