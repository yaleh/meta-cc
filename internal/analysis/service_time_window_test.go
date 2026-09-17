package analysis_test

// DIR-095 regression tests: analyze_errors, analyze_bugs, quality_scan,
// get_work_patterns, and get_tech_debt accepted no since/until parameters at
// all, so a "last 2 days" question was answered from the whole corpus and the
// report silently mixed time windows. query_session_signals and get_timeline
// already had the vocabulary; these tests prove it now reaches the analyzers.
//
// The oracle throughout is the AC's own wording: a windowed run over the full
// corpus must be indistinguishable from a run over a corpus that genuinely
// contains only the in-window records. Two independent projects are seeded
// under one META_CC_PROJECTS_ROOT and queried by working_dir, so the
// comparison never depends on the filter's own internals.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/analysis"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/types"
)

// windowedAnalysisTools are the five tools DIR-095 put behind the window.
var windowedAnalysisTools = []string{
	"analyze_errors",
	"analyze_bugs",
	"quality_scan",
	"get_work_patterns",
	"get_tech_debt",
}

// windowFixtureDayCalls / windowFixtureDayFailures describe the 1-indexed
// fixture days (2026-03-01 .. 2026-03-05). Both the call count and the failure
// count deliberately vary per day: a structurally uniform corpus would make
// every rate-based dimension (error rate, retry rate, diversity, completion)
// identical between the full corpus and the window, letting a filter that
// silently did nothing still pass an equality-only assertion.
var (
	windowFixtureDayCalls    = []int{2, 3, 4, 3, 5}
	windowFixtureDayFailures = []int{0, 1, 2, 1, 0}
)

// The window under test is the half-open [since, until) interval covering
// fixture days 3 and 4 — neither the first nor the last day, so a filter that
// drops a bound (or the whole filter) changes the answer.
const (
	windowFixtureSince = "2026-03-03T00:00:00Z"
	windowFixtureUntil = "2026-03-05T00:00:00Z"
	windowFixtureFirst = 3
	windowFixtureLast  = 4
)

// windowFixtureErrorSignatures are the two recurring failure texts the fixture
// alternates between. Two — not one per call — is deliberate: the analyzers
// break count ties by map iteration order, so a corpus where every error
// signature recurs exactly once makes `patterns`/`by_type` ordering
// non-deterministic between two otherwise-identical runs. Cycling two
// signatures gives each one a distinct recurrence count (alpha 3 vs beta 1
// over the full corpus, alpha 2 vs beta 1 over the window), which pins the
// sort order while still leaving the window visibly different from the whole.
var windowFixtureErrorSignatures = []string{"ERROR: alpha failure", "ERROR: beta failure"}

// windowFixtureDayEntries builds one fixture day. Every day carries a distinct
// hour (so get_work_patterns' 24-slot histogram is window-sensitive), a
// distinct .go path (so get_tech_debt's hotspot list is), and a TODO comment in
// the successful tool output (so tech-debt markers are actually present to be
// counted) — while the day's call and failure counts come from the plans above,
// so every rate-based dimension moves with the window.
func windowFixtureDayEntries(day int) []types.SessionEntry {
	calls := windowFixtureDayCalls[day-1]
	failures := windowFixtureDayFailures[day-1]
	hour := 8 + day

	var entries []types.SessionEntry
	for i := 0; i < calls; i++ {
		tool := "Read"
		if i%2 == 1 {
			tool = "Edit"
		}
		id := fmt.Sprintf("tu-d%d-c%d", day, i)
		ts := fmt.Sprintf("2026-03-%02dT%02d:%02d:00Z", day, hour, i)
		filePath := fmt.Sprintf("pkg/day%d/file%d.go", day, i)
		output := fmt.Sprintf("// TODO: day%d item%d follow-up\n", day, i)
		status, errText := "success", ""
		if i < failures {
			status = "error"
			errText = windowFixtureErrorSignatures[i%len(windowFixtureErrorSignatures)]
			output = errText
		}
		entries = append(entries,
			toolUseEntry("u-"+id, ts, id, tool, map[string]interface{}{"file_path": filePath}),
			toolResultEntry("r-"+id, ts, id, output, status, errText),
		)
	}
	return entries
}

// windowFixtureCorpus is every fixture day, in order.
func windowFixtureCorpus() []types.SessionEntry {
	var entries []types.SessionEntry
	for day := 1; day <= len(windowFixtureDayCalls); day++ {
		entries = append(entries, windowFixtureDayEntries(day)...)
	}
	return entries
}

// windowFixtureSubset is the corpus a correct filter must leave behind: exactly
// the days inside the window, byte-for-byte the same entries the full corpus
// carries for those days.
func windowFixtureSubset() []types.SessionEntry {
	var entries []types.SessionEntry
	for day := windowFixtureFirst; day <= windowFixtureLast; day++ {
		entries = append(entries, windowFixtureDayEntries(day)...)
	}
	return entries
}

// seedWindowCorpus writes entries as one Claude session under the
// META_CC_PROJECTS_ROOT-hashed directory for a fresh project, and returns that
// project's resolved path to use as working_dir.
//
// Unlike setupProjectDirWithEntries this does not set META_CC_PROJECTS_ROOT
// itself, so one test can seed both the full and the subset corpus under a
// single shared projects root and address them by working_dir — the whole
// point of the comparison.
func seedWindowCorpus(t *testing.T, projectsRoot string, entries []types.SessionEntry) string {
	t.Helper()

	absProject, err := filepath.Abs(t.TempDir())
	require.NoError(t, err)
	resolvedProject, err := filepath.EvalSymlinks(absProject)
	require.NoError(t, err)

	sessionDir := filepath.Join(projectsRoot, locator.PathToHash(resolvedProject))
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	var lines []string
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		require.NoError(t, err)
		lines = append(lines, string(data))
	}
	sessionFile := filepath.Join(sessionDir, "session.jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644))

	return resolvedProject
}

// newWindowFixtureProjects seeds the full corpus and the known time-bounded
// subset side by side and returns their working_dir paths.
func newWindowFixtureProjects(t *testing.T) (full, subset string) {
	t.Helper()

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	return seedWindowCorpus(t, projectsRoot, windowFixtureCorpus()),
		seedWindowCorpus(t, projectsRoot, windowFixtureSubset())
}

// runAnalysisTool invokes one of the five tools through the real
// *analysis.Service and returns its marshaled response.
func runAnalysisTool(t *testing.T, tool string, args map[string]interface{}) string {
	t.Helper()

	svc := analysis.New()
	var (
		out string
		err error
	)
	switch tool {
	case "analyze_errors":
		out, err = svc.AnalyzeErrors(args)
	case "analyze_bugs":
		out, err = svc.AnalyzeBugs(args)
	case "quality_scan":
		out, err = svc.QualityScan(args)
	case "get_work_patterns":
		out, err = svc.GetWorkPatterns(args)
	case "get_tech_debt":
		out, err = svc.GetTechDebt(args)
	default:
		t.Fatalf("unknown analysis tool %q", tool)
	}
	require.NoError(t, err, "%s returned an error", tool)
	return out
}

// TestAnalysisToolsHonorSinceUntilWindow is the AC-1 proof for all five tools:
// a windowed run over the full corpus equals a run over a corpus holding only
// the in-window records, and differs from the unwindowed run (so the equality
// is not satisfied by a filter that does nothing).
func TestAnalysisToolsHonorSinceUntilWindow(t *testing.T) {
	full, subset := newWindowFixtureProjects(t)

	for _, tool := range windowedAnalysisTools {
		t.Run(tool, func(t *testing.T) {
			unwindowed := runAnalysisTool(t, tool, map[string]interface{}{"working_dir": full})
			windowed := runAnalysisTool(t, tool, map[string]interface{}{
				"working_dir": full,
				"since":       windowFixtureSince,
				"until":       windowFixtureUntil,
			})
			known := runAnalysisTool(t, tool, map[string]interface{}{"working_dir": subset})

			assert.NotEqual(t, unwindowed, windowed,
				"%s must return a different result once the window excludes days outside it", tool)
			assert.JSONEq(t, known, windowed,
				"%s windowed over the full corpus must match the known time-bounded subset", tool)
		})
	}
}

// TestAnalysisToolsStatsOnlyHonorsSinceUntilWindow is the AC-3 proof: the
// stats_only short-circuit reads the same windowed slices as the full path, so
// its aggregates cover the window too. (stats_first is not declared on these
// tools at all — DIR-048 keeps it scoped to the four pipeline-routed tools —
// so stats_only is the stats mode that applies here; see
// TestAnalysisToolsRejectUndeclaredStatsFirst in internal/mcp/executor for the
// proof that it is rejected rather than silently ignored.)
func TestAnalysisToolsStatsOnlyHonorsSinceUntilWindow(t *testing.T) {
	full, subset := newWindowFixtureProjects(t)

	for _, tool := range windowedAnalysisTools {
		t.Run(tool, func(t *testing.T) {
			unwindowed := runAnalysisTool(t, tool, map[string]interface{}{
				"working_dir": full,
				"stats_only":  true,
			})
			windowed := runAnalysisTool(t, tool, map[string]interface{}{
				"working_dir": full,
				"stats_only":  true,
				"since":       windowFixtureSince,
				"until":       windowFixtureUntil,
			})
			known := runAnalysisTool(t, tool, map[string]interface{}{
				"working_dir": subset,
				"stats_only":  true,
			})

			assert.NotEqual(t, unwindowed, windowed,
				"%s stats_only must reflect the window", tool)
			assert.JSONEq(t, known, windowed,
				"%s stats_only windowed over the full corpus must match the known time-bounded subset", tool)
		})
	}
}

// TestAnalysisToolsRejectInvalidSinceUntil is the AC-2 proof at the service
// layer: a malformed bound is reported as mcerrors.ErrInvalidInput, so callers
// can branch on the sentinel (errors.Is) instead of matching message text, and
// no analyzer runs against a half-specified window.
func TestAnalysisToolsRejectInvalidSinceUntil(t *testing.T) {
	full, _ := newWindowFixtureProjects(t)

	for _, tool := range windowedAnalysisTools {
		for _, tc := range []struct {
			key   string
			value string
		}{
			{key: "since", value: "not-a-timestamp"},
			{key: "until", value: "2026-03-05"},
			{key: "since", value: "2026-03-03T00:00:00"},
		} {
			t.Run(tool+"/"+tc.key+"="+tc.value, func(t *testing.T) {
				svc := analysis.New()
				args := map[string]interface{}{"working_dir": full, tc.key: tc.value}

				var err error
				switch tool {
				case "analyze_errors":
					_, err = svc.AnalyzeErrors(args)
				case "analyze_bugs":
					_, err = svc.AnalyzeBugs(args)
				case "quality_scan":
					_, err = svc.QualityScan(args)
				case "get_work_patterns":
					_, err = svc.GetWorkPatterns(args)
				case "get_tech_debt":
					_, err = svc.GetTechDebt(args)
				}

				require.Error(t, err, "%s must reject %s=%q", tool, tc.key, tc.value)
				assert.ErrorIs(t, err, mcerrors.ErrInvalidInput,
					"%s must report a malformed %s as ErrInvalidInput, got: %v", tool, tc.key, err)
			})
		}
	}
}

// TestAnalysisToolsInvalidBoundOutranksMissingCorpus pins the ordering the
// executor's fail-fast validation depends on: a malformed bound is reported as
// ErrInvalidInput even pointed at a project with no session data at all. If
// validation lived only in the service's post-load filter, this case would
// surface as a corpus lookup failure instead and AC-2 would be untestable on a
// machine with no sessions.
func TestAnalysisToolsInvalidBoundOutranksMissingCorpus(t *testing.T) {
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	emptyProject := seedWindowCorpus(t, projectsRoot, nil)

	_, err := analysis.New().AnalyzeErrors(map[string]interface{}{
		"working_dir": emptyProject,
		"since":       "yesterday",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, mcerrors.ErrInvalidInput)
}

// TestAnalysisToolsUnboundedRunUnaffected guards the no-window path: omitting
// since/until must leave the corpus exactly as loaded, so the DIR-042/018
// behaviour these tools already had is not perturbed by the new parameter.
func TestAnalysisToolsUnboundedRunUnaffected(t *testing.T) {
	full, _ := newWindowFixtureProjects(t)

	baseline := runAnalysisTool(t, "analyze_errors", map[string]interface{}{"working_dir": full})
	explicitEmpty := runAnalysisTool(t, "analyze_errors", map[string]interface{}{
		"working_dir": full,
		"since":       "",
		"until":       "",
	})
	assert.JSONEq(t, baseline, explicitEmpty,
		"empty since/until must be treated as unbounded, not as a zero-width window")
}
