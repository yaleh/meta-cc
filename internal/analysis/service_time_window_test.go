package analysis_test

// DIR-095: the five analysis tools that previously accepted NO time filter
// (analyze_errors, analyze_bugs, quality_scan, get_work_patterns,
// get_tech_debt) now honor the same RFC3339 since/until window get_timeline
// already had. Before this, a "last 2 days" question produced whole-corpus
// statistics while looking like a time-bounded answer.
//
// The central proof is an equivalence, not a hand-written expected value:
// running a tool WITH a window over the full corpus must produce exactly what
// running the same tool with NO window over a corpus containing only the
// in-window entries produces. That is the operational definition of "the
// window was honored", and it is why applyTimeWindow re-derives toolCalls
// from the filtered entries rather than filtering them independently -- a
// tool-call-counting tool (analyze_errors, quality_scan, get_work_patterns)
// would otherwise still see every out-of-window call.
//
// Each case additionally asserts the windowed result DIFFERS from the
// unwindowed one, so a tool that silently ignored its window could not pass by
// accident.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/analysis"
	"github.com/yaleh/meta-cc/internal/analyzer"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/types"
)

// The window under test: February 2026. outOfWindowEntries live in January,
// inWindowEntries in February.
const (
	windowSince = "2026-02-01T00:00:00Z"
	windowUntil = "2026-03-01T00:00:00Z"
)

// outOfWindowEntries is the January half of the fixture: one failing Bash call
// followed by a same-tool success (an analyze_bugs fix pair), plus a Read whose
// output carries a FIXME marker.
func outOfWindowEntries() []types.SessionEntry {
	return []types.SessionEntry{
		toolUseEntry("u-old-err", "2026-01-15T10:00:00Z", "tu-old-err", "Bash", map[string]interface{}{"command": "boom"}),
		toolResultEntry("r-old-err", "2026-01-15T10:00:00Z", "tu-old-err", "boom failure", "error", "boom failure"),
		toolUseEntry("u-old-fix", "2026-01-15T10:00:01Z", "tu-old-fix", "Bash", map[string]interface{}{"command": "fixed"}),
		toolResultEntry("r-old-fix", "2026-01-15T10:00:01Z", "tu-old-fix", "ok", "success", ""),
		toolUseEntry("u-old-read", "2026-01-15T10:00:02Z", "tu-old-read", "Read", map[string]interface{}{"file_path": "pkg/old.go"}),
		toolResultEntry("r-old-read", "2026-01-15T10:00:02Z", "tu-old-read", "// FIXME: stale marker\n", "success", ""),
	}
}

// inWindowEntries is the February half: two clean Reads, each with a TODO
// marker, on two distinct files.
func inWindowEntries() []types.SessionEntry {
	return []types.SessionEntry{
		toolUseEntry("u-new-a", "2026-02-15T10:00:00Z", "tu-new-a", "Read", map[string]interface{}{"file_path": "pkg/new_a.go"}),
		toolResultEntry("r-new-a", "2026-02-15T10:00:00Z", "tu-new-a", "// TODO: alpha\n", "success", ""),
		toolUseEntry("u-new-b", "2026-02-15T10:00:01Z", "tu-new-b", "Read", map[string]interface{}{"file_path": "pkg/new_b.go"}),
		toolResultEntry("r-new-b", "2026-02-15T10:00:01Z", "tu-new-b", "// TODO: beta\n", "success", ""),
	}
}

// seedWindowCorpus writes entries as one session file under projectsRoot, in
// the project-hash directory Claude Code itself would use for projectPath.
// Unlike setupProjectDirWithEntries it takes both roots as arguments, so a
// single test can seed several distinct projects (and therefore several
// distinct corpora) under one shared projectsRoot.
func seedWindowCorpus(t *testing.T, projectsRoot, projectPath string, entries []types.SessionEntry) string {
	t.Helper()

	absProject, err := filepath.Abs(projectPath)
	require.NoError(t, err)
	resolved, err := filepath.EvalSymlinks(absProject)
	require.NoError(t, err)
	hash := strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolved)
	sessionDir := filepath.Join(projectsRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	var lines []string
	for _, e := range entries {
		data, err := json.Marshal(e)
		require.NoError(t, err)
		lines = append(lines, string(data))
	}
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, "session.jsonl"),
		[]byte(strings.Join(lines, "\n")+"\n"), 0o644))

	return projectPath
}

// windowFixture seeds the full corpus and the in-window-only subset under one
// projectsRoot and returns the two projects' working_dir values.
func windowFixture(t *testing.T) (fullProject, subsetProject string) {
	t.Helper()

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	full := append(outOfWindowEntries(), inWindowEntries()...)
	fullProject = seedWindowCorpus(t, projectsRoot, t.TempDir(), full)
	subsetProject = seedWindowCorpus(t, projectsRoot, t.TempDir(), inWindowEntries())
	return fullProject, subsetProject
}

// analysisRun is one of the six analysis.Service entry points.
type analysisRun struct {
	name string
	run  func(map[string]interface{}) (string, error)
}

// analysisRuns returns all six entry points. get_timeline is included even
// though it already supported since/until before DIR-095: applyTimeWindow
// replaced its private filtering block, so this pins that refactor too.
func analysisRuns(svc *analysis.Service) []analysisRun {
	return []analysisRun{
		{"analyze_errors", svc.AnalyzeErrors},
		{"analyze_bugs", svc.AnalyzeBugs},
		{"quality_scan", svc.QualityScan},
		{"get_work_patterns", svc.GetWorkPatterns},
		{"get_tech_debt", svc.GetTechDebt},
		{"get_timeline", svc.GetTimeline},
	}
}

// TestAnalysisTools_WindowedRunMatchesInWindowSubset is DIR-095's acceptance
// proof for every one of the six tools: a windowed run over the full corpus is
// identical to an unwindowed run over a corpus seeded with only the in-window
// entries, and different from the unwindowed run over the full corpus.
func TestAnalysisTools_WindowedRunMatchesInWindowSubset(t *testing.T) {
	fullProject, subsetProject := windowFixture(t)
	svc := analysis.New()

	windowedArgs := map[string]interface{}{
		"working_dir": fullProject,
		"since":       windowSince,
		"until":       windowUntil,
	}
	subsetArgs := map[string]interface{}{"working_dir": subsetProject}
	fullArgs := map[string]interface{}{"working_dir": fullProject}

	for _, tc := range analysisRuns(svc) {
		t.Run(tc.name, func(t *testing.T) {
			windowed, err := tc.run(windowedArgs)
			require.NoError(t, err)
			subset, err := tc.run(subsetArgs)
			require.NoError(t, err)
			full, err := tc.run(fullArgs)
			require.NoError(t, err)

			assert.JSONEq(t, subset, windowed,
				"a windowed run must equal an unwindowed run over the in-window subset")
			assert.NotEqual(t, full, windowed,
				"the window must actually change the result (otherwise it is silently ignored)")
		})
	}
}

// TestAnalysisTools_StatsOnlyHonorsWindow is DIR-095's aggregate-output proof:
// the stats_only short-circuit each of the five tools takes must describe the
// window, not the whole corpus. Filtering happens in applyTimeWindow, before
// either the full or the stats_only branch runs, so both see the same narrowed
// entries/toolCalls.
//
// On stats_first: the six analysis tools do not declare it (DIR-048 -- they
// return typed analyzer structs, not the flat record array stats_first
// paginates), and ValidateToolArgs rejects an undeclared key, so no
// stats_first output exists for these tools to reflect a window. The
// parameter's absence is pinned by
// TestAnalysisTools_DoNotDeclareStatsFirst in internal/mcp/executor.
func TestAnalysisTools_StatsOnlyHonorsWindow(t *testing.T) {
	fullProject, subsetProject := windowFixture(t)
	svc := analysis.New()

	windowedArgs := map[string]interface{}{
		"working_dir": fullProject,
		"since":       windowSince,
		"until":       windowUntil,
		"stats_only":  true,
	}
	subsetArgs := map[string]interface{}{"working_dir": subsetProject, "stats_only": true}
	fullArgs := map[string]interface{}{"working_dir": fullProject, "stats_only": true}

	for _, tc := range analysisRuns(svc) {
		t.Run(tc.name, func(t *testing.T) {
			windowed, err := tc.run(windowedArgs)
			require.NoError(t, err)
			subset, err := tc.run(subsetArgs)
			require.NoError(t, err)
			full, err := tc.run(fullArgs)
			require.NoError(t, err)

			assert.JSONEq(t, subset, windowed,
				"stats_only over a windowed corpus must equal stats_only over the in-window subset")
			assert.NotEqual(t, full, windowed,
				"stats_only must not report whole-corpus aggregates for a windowed call")
		})
	}
}

// TestAnalysisTools_WindowedCountsExcludeOutOfWindowData pins the concrete
// numbers, so a failure says *what* leaked rather than only that two JSON
// blobs differ.
func TestAnalysisTools_WindowedCountsExcludeOutOfWindowData(t *testing.T) {
	fullProject, _ := windowFixture(t)
	svc := analysis.New()

	windowed := func(args map[string]interface{}) map[string]interface{} {
		if args == nil {
			args = map[string]interface{}{}
		}
		args["working_dir"] = fullProject
		args["since"] = windowSince
		args["until"] = windowUntil
		return args
	}
	full := func(args map[string]interface{}) map[string]interface{} {
		if args == nil {
			args = map[string]interface{}{}
		}
		args["working_dir"] = fullProject
		return args
	}

	t.Run("analyze_errors", func(t *testing.T) {
		var fullRes analyzer.ErrorAnalysisResult
		out, err := svc.AnalyzeErrors(full(nil))
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal([]byte(out), &fullRes))
		assert.Equal(t, 1, fullRes.TotalErrors, "the January Bash failure is the only error in the corpus")

		var windowedRes analyzer.ErrorAnalysisResult
		out, err = svc.AnalyzeErrors(windowed(nil))
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal([]byte(out), &windowedRes))
		assert.Zero(t, windowedRes.TotalErrors, "no error occurred inside the February window")
		assert.Empty(t, windowedRes.ByTool)
		assert.Equal(t, "2026-02-15T10:00:00Z", windowedRes.TimeRange.Start,
			"the reported time range must describe the window, not the whole corpus")
	})

	t.Run("analyze_bugs", func(t *testing.T) {
		var fullRes analyzer.BugAnalysisResult
		out, err := svc.AnalyzeBugs(full(nil))
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal([]byte(out), &fullRes))
		assert.Equal(t, 1, fullRes.TotalErrors)
		assert.Equal(t, 1, fullRes.TotalPairs, "the January error->success Bash pair")

		var windowedRes analyzer.BugAnalysisResult
		out, err = svc.AnalyzeBugs(windowed(nil))
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal([]byte(out), &windowedRes))
		assert.Zero(t, windowedRes.TotalErrors)
		assert.Zero(t, windowedRes.TotalPairs, "the fix pair lies outside the window")
		assert.Empty(t, windowedRes.Patterns)
	})

	t.Run("quality_scan", func(t *testing.T) {
		assert.Equal(t, "1/5", dimensionRawValue(t, svc, full(nil), "error_rate"))
		assert.Equal(t, "0/2", dimensionRawValue(t, svc, windowed(nil), "error_rate"))
	})

	t.Run("get_work_patterns", func(t *testing.T) {
		var fullRes analyzer.WorkPatternsResult
		out, err := svc.GetWorkPatterns(full(nil))
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal([]byte(out), &fullRes))
		assert.Equal(t, map[string]int{"Bash": 2, "Read": 3}, toolFrequency(fullRes))

		var windowedRes analyzer.WorkPatternsResult
		out, err = svc.GetWorkPatterns(windowed(nil))
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal([]byte(out), &windowedRes))
		assert.Equal(t, map[string]int{"Read": 2}, toolFrequency(windowedRes),
			"the January Bash calls must not survive the window")
	})

	t.Run("get_tech_debt", func(t *testing.T) {
		var fullRes analyzer.TechDebtResult
		out, err := svc.GetTechDebt(full(nil))
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal([]byte(out), &fullRes))
		assert.Equal(t, map[string]int{"FIXME": 1, "TODO": 2}, markerCounts(fullRes))

		var windowedRes analyzer.TechDebtResult
		out, err = svc.GetTechDebt(windowed(nil))
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal([]byte(out), &windowedRes))
		assert.Equal(t, map[string]int{"TODO": 2}, markerCounts(windowedRes),
			"the January FIXME was read outside the window")

		var windowedStats analyzer.TechDebtStats
		statsArgs := windowed(map[string]interface{}{"stats_only": true})
		out, err = svc.GetTechDebt(statsArgs)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal([]byte(out), &windowedStats))
		assert.Equal(t, 2, windowedStats.MarkerCount, "stats_only must count only in-window markers")
	})
}

// TestAnalysisTools_RejectInvalidTimeWindow proves an unparseable bound is a
// typed invalid-input error, not a crash and not a silent whole-corpus result.
// It runs against a seeded corpus because the analysis Service validates the
// window after loading (the executor's validateTimeWindow is the fail-fast,
// pre-load guard -- see internal/mcp/executor/analysis_time_window_test.go).
func TestAnalysisTools_RejectInvalidTimeWindow(t *testing.T) {
	fullProject, _ := windowFixture(t)
	svc := analysis.New()

	for _, tc := range analysisRuns(svc) {
		for _, bad := range []map[string]interface{}{
			{"since": "2026-02-01"},
			{"until": "02/01/2026"},
		} {
			t.Run(tc.name+" "+mapKeys(bad), func(t *testing.T) {
				args := map[string]interface{}{"working_dir": fullProject}
				for k, v := range bad {
					args[k] = v
				}
				out, err := tc.run(args)
				require.Error(t, err)
				assert.ErrorIs(t, err, mcerrors.ErrInvalidInput,
					"an unparseable timestamp must surface the invalid-input sentinel")
				assert.Empty(t, out)
			})
		}
	}
}

// --- small decoders -------------------------------------------------------

func dimensionRawValue(t *testing.T, svc *analysis.Service, args map[string]interface{}, name string) string {
	t.Helper()
	out, err := svc.QualityScan(args)
	require.NoError(t, err)
	var res analyzer.QualityScanResult
	require.NoError(t, json.Unmarshal([]byte(out), &res))
	for _, d := range res.Dimensions {
		if d.Name == name {
			return d.RawValue
		}
	}
	t.Fatalf("quality_scan returned no %q dimension", name)
	return ""
}

func toolFrequency(res analyzer.WorkPatternsResult) map[string]int {
	freq := make(map[string]int, len(res.ToolFrequency))
	for _, tc := range res.ToolFrequency {
		freq[tc.ToolName] = tc.Count
	}
	return freq
}

func markerCounts(res analyzer.TechDebtResult) map[string]int {
	counts := make(map[string]int, len(res.Markers))
	for _, m := range res.Markers {
		counts[m.Label] = m.Count
	}
	return counts
}

func mapKeys(m map[string]interface{}) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ",")
}
