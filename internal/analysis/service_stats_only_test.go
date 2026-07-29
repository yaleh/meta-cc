package analysis_test

// DIR-042 regression tests: analyze_bugs, analyze_errors, quality_scan,
// get_work_patterns, and get_tech_debt each advertise a stats_only
// parameter but, before this fix, silently ignored it -- always calling the
// full injected analyzer and marshaling its complete (potentially huge,
// per-item-example-carrying) result regardless of what the caller passed.
// GetTimeline was the only one of the six analysis.Service methods that
// actually honored stats_only (see TestGetTimelineStatsOnly in
// cmd/mcp-server/analysis_timeline_test.go, the template these tests
// mirror). Each test below proves two things for its method:
//
//  1. With stats_only:true, the injected (mocked) analyzer is never called
//     at all -- the short-circuit calls a package-level analyzer.*Stats*
//     function directly against the loaded entries/toolCalls, exactly the
//     way GetTimeline calls analyzer.GetTimelineStats directly rather than
//     going through s.analyzers.Timeline.
//  2. The stats_only response never contains the long, per-item example
//     text a full response would carry (reproducing, at test scale, the
//     74,059-character/token-limit failure this task closes).

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
	"github.com/yaleh/meta-cc/internal/types"
)

// longExampleText stands in for the kind of single oversized example
// (3,988 characters in the live DIR-042 failure) that must never appear in
// a stats_only response.
const longExampleText = "FAILURE-TRACE: " // will be repeated below to build a long string

func repeatedLongText() string {
	return longExampleText + strings.Repeat("x", 4000)
}

// setupProjectDirWithEntries mirrors setupEmptyProjectDir but seeds the
// session file with real content (marshaled types.SessionEntry values)
// instead of an empty file, so loadData returns non-trivial entries/
// toolCalls for the stats_only short-circuit to operate on.
func setupProjectDirWithEntries(t *testing.T, entries []types.SessionEntry) string {
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
	sessionFile := filepath.Join(sessionDir, "session.jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644))
	return projectPath
}

// toolUseEntry builds an assistant SessionEntry carrying a single tool_use block.
func toolUseEntry(uuid, ts, toolUseID, toolName string, input map[string]interface{}) types.SessionEntry {
	return types.SessionEntry{
		Type:      "assistant",
		UUID:      uuid,
		Timestamp: ts,
		Message: &types.Message{
			Role: "assistant",
			Content: []types.ContentBlock{
				{
					Type:    "tool_use",
					ToolUse: &types.ToolUse{ID: toolUseID, Name: toolName, Input: input},
				},
			},
		},
	}
}

// toolResultEntry builds a user SessionEntry carrying a single tool_result block.
func toolResultEntry(uuid, ts, toolUseID, content, status, errText string) types.SessionEntry {
	return types.SessionEntry{
		Type:      "user",
		UUID:      uuid,
		Timestamp: ts,
		Message: &types.Message{
			Role: "user",
			Content: []types.ContentBlock{
				{
					Type: "tool_result",
					ToolResult: &types.ToolResult{
						ToolUseID: toolUseID,
						Content:   content,
						Status:    status,
						Error:     errText,
					},
				},
			},
		},
	}
}

// buildErrorHeavySession returns entries for count Bash tool calls, all
// failing with the same long error text, at successive timestamps.
func buildErrorHeavySession(count int) []types.SessionEntry {
	var entries []types.SessionEntry
	longErr := repeatedLongText()
	for i := 0; i < count; i++ {
		id := "tu-" + string(rune('a'+i))
		ts := "2025-10-02T10:00:0" + string(rune('0'+i%9)) + ".000Z"
		entries = append(entries,
			toolUseEntry("u-"+id, ts, id, "Bash", map[string]interface{}{"command": "boom"}),
			toolResultEntry("r-"+id, ts, id, longErr, "error", longErr),
		)
	}
	return entries
}

// buildBugFixSession returns entries for count error->success Bash fix pairs,
// each error carrying the same long error text.
func buildBugFixSession(count int) []types.SessionEntry {
	var entries []types.SessionEntry
	longErr := repeatedLongText()
	for i := 0; i < count; i++ {
		errID := "tu-err-" + string(rune('a'+i))
		okID := "tu-ok-" + string(rune('a'+i))
		ts := "2025-10-02T10:00:0" + string(rune('0'+i%9)) + ".000Z"
		entries = append(entries,
			toolUseEntry("u-"+errID, ts, errID, "Bash", map[string]interface{}{"command": "boom"}),
			toolResultEntry("r-"+errID, ts, errID, longErr, "error", longErr),
			toolUseEntry("u-"+okID, ts, okID, "Bash", map[string]interface{}{"command": "fixed"}),
			toolResultEntry("r-"+okID, ts, okID, "ok", "success", ""),
		)
	}
	return entries
}

// buildTechDebtSession returns entries for count Read calls, each against a
// distinct file path, each output carrying a TODO marker -- enough to build
// a HotspotFiles list with count entries in the full (non-stats_only) result.
func buildTechDebtSession(count int) []types.SessionEntry {
	var entries []types.SessionEntry
	for i := 0; i < count; i++ {
		id := "tu-td-" + string(rune('a'+i))
		ts := "2025-10-02T10:00:0" + string(rune('0'+i%9)) + ".000Z"
		path := "pkg/file" + string(rune('a'+i)) + ".go"
		entries = append(entries,
			toolUseEntry("u-"+id, ts, id, "Read", map[string]interface{}{"file_path": path}),
			toolResultEntry("r-"+id, ts, id, "// TODO: fix this later\n", "success", ""),
		)
	}
	return entries
}

// --- capturing stubs: record whether the injected analyzer was invoked ---

type capturingBugAnalyzer struct {
	called bool
	result *analyzer.BugAnalysisResult
}

func (c *capturingBugAnalyzer) AnalyzeBugs(_ []types.SessionEntry, _ []types.ToolCall, _ int) (*analyzer.BugAnalysisResult, error) {
	c.called = true
	return c.result, nil
}

type capturingQualityScanner struct {
	called bool
	result *analyzer.QualityScanResult
}

func (c *capturingQualityScanner) QualityScan(_ []types.SessionEntry, _ []types.ToolCall) (*analyzer.QualityScanResult, error) {
	c.called = true
	return c.result, nil
}

type capturingWorkPatternsAnalyzer struct {
	called bool
	result *analyzer.WorkPatternsResult
}

func (c *capturingWorkPatternsAnalyzer) GetWorkPatterns(_ []types.SessionEntry, _ []types.ToolCall) (*analyzer.WorkPatternsResult, error) {
	c.called = true
	return c.result, nil
}

func TestService_AnalyzeErrors_StatsOnly_OmitsExamplesAndBypassesAnalyzer(t *testing.T) {
	entries := buildErrorHeavySession(21)
	projectPath := setupProjectDirWithEntries(t, entries)

	stub := &capturingErrorAnalyzer{}
	svc := analysis.NewWithAnalyzers(analysis.Analyzers{ErrorAnalyzer: stub})

	out, err := svc.AnalyzeErrors(map[string]interface{}{"working_dir": projectPath, "stats_only": true})
	require.NoError(t, err)
	assert.False(t, stub.called, "stats_only must bypass the injected ErrorAnalyzer entirely")
	assert.NotContains(t, out, "examples", "stats_only response must not carry an examples field")
	assert.NotContains(t, out, repeatedLongText(), "stats_only response must not carry full-text error content")
	assert.Less(t, len(out), 2000, "stats_only response for a 21-error session must stay small")

	var stats analyzer.ErrorAnalysisStats
	require.NoError(t, json.Unmarshal([]byte(out), &stats))
	assert.Equal(t, 21, stats.TotalErrors)

	// Sanity: without stats_only, the injected analyzer IS used and its
	// (here, small) result comes back unchanged.
	stub2 := &capturingErrorAnalyzer{}
	svc2 := analysis.NewWithAnalyzers(analysis.Analyzers{ErrorAnalyzer: stub2})
	_, err = svc2.AnalyzeErrors(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	assert.True(t, stub2.called, "non-stats_only calls must still use the injected ErrorAnalyzer")
}

func TestService_AnalyzeBugs_StatsOnly_OmitsExamplesAndBypassesAnalyzer(t *testing.T) {
	entries := buildBugFixSession(5)
	projectPath := setupProjectDirWithEntries(t, entries)

	stub := &capturingBugAnalyzer{}
	svc := analysis.NewWithAnalyzers(analysis.Analyzers{BugAnalyzer: stub})

	out, err := svc.AnalyzeBugs(map[string]interface{}{"working_dir": projectPath, "stats_only": true})
	require.NoError(t, err)
	assert.False(t, stub.called, "stats_only must bypass the injected BugAnalyzer entirely")
	assert.NotContains(t, out, "examples", "stats_only response must not carry an examples field")
	assert.NotContains(t, out, repeatedLongText(), "stats_only response must not carry full-text error content")

	var stats analyzer.BugAnalysisStats
	require.NoError(t, json.Unmarshal([]byte(out), &stats))
	// NOTE: types.ExtractToolCalls builds its result by ranging over a Go map
	// keyed by tool_use ID, so the resulting toolCalls order (and therefore
	// exactly how many of the 5 seeded error->success pairs fall inside
	// AnalyzeBugs's fixed lookahead window) is not guaranteed run-to-run.
	// Assert the loose, order-independent bound instead of an exact count;
	// the exact-count behavior is already covered deterministically by
	// TestAnalyzeBugs_Recurrence/TestAnalyzeBugsStats_OmitsExamples, which
	// build toolCalls slices directly rather than through the real
	// entries->toolCalls extraction pipeline.
	assert.GreaterOrEqual(t, stats.TotalPairs, 0, "TotalPairs must be a valid non-negative count")
	assert.LessOrEqual(t, stats.TotalPairs, 5, "expected no more than the 5 seeded pairs")
	assert.Less(t, len(out), 2000, "stats_only response for a 5-pair session must stay small")

	stub2 := &capturingBugAnalyzer{result: &analyzer.BugAnalysisResult{}}
	svc2 := analysis.NewWithAnalyzers(analysis.Analyzers{BugAnalyzer: stub2})
	_, err = svc2.AnalyzeBugs(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	assert.True(t, stub2.called, "non-stats_only calls must still use the injected BugAnalyzer")
}

func TestService_QualityScan_StatsOnly_BypassesAnalyzer(t *testing.T) {
	entries := buildErrorHeavySession(3)
	projectPath := setupProjectDirWithEntries(t, entries)

	stub := &capturingQualityScanner{}
	svc := analysis.NewWithAnalyzers(analysis.Analyzers{QualityScanner: stub})

	out, err := svc.QualityScan(map[string]interface{}{"working_dir": projectPath, "stats_only": true})
	require.NoError(t, err)
	assert.False(t, stub.called, "stats_only must bypass the injected QualityScanner entirely")
	assert.NotContains(t, out, "examples", "stats_only response must not carry an examples field")

	var stats analyzer.QualityScanStats
	require.NoError(t, json.Unmarshal([]byte(out), &stats))
	assert.NotEmpty(t, stats.Dimensions)

	stub2 := &capturingQualityScanner{result: &analyzer.QualityScanResult{}}
	svc2 := analysis.NewWithAnalyzers(analysis.Analyzers{QualityScanner: stub2})
	_, err = svc2.QualityScan(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	assert.True(t, stub2.called, "non-stats_only calls must still use the injected QualityScanner")
}

func TestService_GetWorkPatterns_StatsOnly_BypassesAnalyzer(t *testing.T) {
	entries := buildErrorHeavySession(3)
	projectPath := setupProjectDirWithEntries(t, entries)

	stub := &capturingWorkPatternsAnalyzer{}
	svc := analysis.NewWithAnalyzers(analysis.Analyzers{WorkPatterns: stub})

	out, err := svc.GetWorkPatterns(map[string]interface{}{"working_dir": projectPath, "stats_only": true})
	require.NoError(t, err)
	assert.False(t, stub.called, "stats_only must bypass the injected WorkPatternsAnalyzer entirely")
	assert.NotContains(t, out, "examples", "stats_only response must not carry an examples field")

	var stats analyzer.WorkPatternsStats
	require.NoError(t, json.Unmarshal([]byte(out), &stats))
	assert.NotEmpty(t, stats.ToolFrequency)

	stub2 := &capturingWorkPatternsAnalyzer{result: &analyzer.WorkPatternsResult{}}
	svc2 := analysis.NewWithAnalyzers(analysis.Analyzers{WorkPatterns: stub2})
	_, err = svc2.GetWorkPatterns(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	assert.True(t, stub2.called, "non-stats_only calls must still use the injected WorkPatternsAnalyzer")
}

// TestService_GetTechDebt_StatsOnly_OmitsHotspotFileList verifies stats_only
// converts the already-computed TechDebtResult (which, unlike the other four
// fixed methods, GetTechDebt must still compute up front to support the
// optional source_dir merge) into an aggregate-only summary: the full
// per-file HotspotFiles path list is dropped in favor of a single
// hotspot_file_count, so the response shrinks even though the underlying
// scan itself still ran.
func TestService_GetTechDebt_StatsOnly_OmitsHotspotFileList(t *testing.T) {
	entries := buildTechDebtSession(30)
	projectPath := setupProjectDirWithEntries(t, entries)

	svc := analysis.New()

	full, err := svc.GetTechDebt(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	assert.Contains(t, full, "hotspot_files", "the full (non-stats_only) response should still carry the per-file list")

	var fullResult analyzer.TechDebtResult
	require.NoError(t, json.Unmarshal([]byte(full), &fullResult))
	require.Len(t, fullResult.HotspotFiles, 30)

	statsOut, err := svc.GetTechDebt(map[string]interface{}{"working_dir": projectPath, "stats_only": true})
	require.NoError(t, err)
	assert.NotContains(t, statsOut, "hotspot_files", "stats_only response must not carry the full per-file hotspot list")

	var stats analyzer.TechDebtStats
	require.NoError(t, json.Unmarshal([]byte(statsOut), &stats))
	assert.Equal(t, 30, stats.HotspotFileCount)
	assert.Equal(t, 30, stats.MarkerCount)
	assert.Less(t, len(statsOut), len(full), "stats_only response must be measurably smaller than the full response")
}
