package analyzer

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

func TestAnalyzeBugs_FixPair(t *testing.T) {
	toolCalls := []types.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "command not found"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "success"},
	}

	result, err := AnalyzeBugs([]types.SessionEntry{}, toolCalls, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.TotalPairs != 1 {
		t.Errorf("Expected TotalPairs=1, got %d", result.TotalPairs)
	}
	if len(result.Patterns) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(result.Patterns))
	}
}

func TestAnalyzeBugs_Recurrence(t *testing.T) {
	// Same tool+error appearing 3 times, each followed by a success
	toolCalls := []types.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "permission denied"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "success"},
		{UUID: "uuid-3", ToolName: "Bash", Status: "error", Error: "permission denied"},
		{UUID: "uuid-4", ToolName: "Bash", Status: "success"},
		{UUID: "uuid-5", ToolName: "Bash", Status: "error", Error: "permission denied"},
		{UUID: "uuid-6", ToolName: "Bash", Status: "success"},
	}

	result, err := AnalyzeBugs([]types.SessionEntry{}, toolCalls, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result.Patterns) == 0 {
		t.Fatal("Expected at least 1 pattern")
	}
	if result.Patterns[0].Recurrences != 3 {
		t.Errorf("Expected Recurrences=3, got %d", result.Patterns[0].Recurrences)
	}
	if result.TotalPairs != 3 {
		t.Errorf("Expected TotalPairs=3, got %d", result.TotalPairs)
	}
}

func TestAnalyzeBugs_SortedByRecurrence(t *testing.T) {
	// Create patterns: one error with 1 occurrence, one with 3, one with 2
	toolCalls := []types.ToolCall{
		// error-A appears once
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "error alpha unique"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "success"},
		// error-B appears 3 times
		{UUID: "uuid-3", ToolName: "Read", Status: "error", Error: "error beta repeated"},
		{UUID: "uuid-4", ToolName: "Read", Status: "success"},
		{UUID: "uuid-5", ToolName: "Read", Status: "error", Error: "error beta repeated"},
		{UUID: "uuid-6", ToolName: "Read", Status: "success"},
		{UUID: "uuid-7", ToolName: "Read", Status: "error", Error: "error beta repeated"},
		{UUID: "uuid-8", ToolName: "Read", Status: "success"},
		// error-C appears twice
		{UUID: "uuid-9", ToolName: "Grep", Status: "error", Error: "error gamma double"},
		{UUID: "uuid-10", ToolName: "Grep", Status: "success"},
		{UUID: "uuid-11", ToolName: "Grep", Status: "error", Error: "error gamma double"},
		{UUID: "uuid-12", ToolName: "Grep", Status: "success"},
	}

	result, err := AnalyzeBugs([]types.SessionEntry{}, toolCalls, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result.Patterns) != 3 {
		t.Fatalf("Expected 3 patterns, got %d", len(result.Patterns))
	}
	// Verify sorted descending by recurrences: 3, 2, 1
	if result.Patterns[0].Recurrences != 3 {
		t.Errorf("Expected first pattern Recurrences=3, got %d", result.Patterns[0].Recurrences)
	}
	if result.Patterns[1].Recurrences != 2 {
		t.Errorf("Expected second pattern Recurrences=2, got %d", result.Patterns[1].Recurrences)
	}
	if result.Patterns[2].Recurrences != 1 {
		t.Errorf("Expected third pattern Recurrences=1, got %d", result.Patterns[2].Recurrences)
	}
}

func TestAnalyzeBugs_EmptySession(t *testing.T) {
	result, err := AnalyzeBugs([]types.SessionEntry{}, []types.ToolCall{}, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result.Patterns) != 0 {
		t.Errorf("Expected 0 patterns, got %d", len(result.Patterns))
	}
	if result.TotalPairs != 0 {
		t.Errorf("Expected TotalPairs=0, got %d", result.TotalPairs)
	}
}

func TestAnalyzeBugs_DataSource(t *testing.T) {
	result, err := AnalyzeBugs([]types.SessionEntry{}, []types.ToolCall{}, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.DataSource != DataSourceMeasured {
		t.Errorf("Expected DataSource=%q, got %q", DataSourceMeasured, result.DataSource)
	}
}

// TestAnalyzeBugsStats_OmitsExamples verifies stats_only's backing function
// (analyzer.AnalyzeBugsStats) produces the same pattern counts as AnalyzeBugs
// but with no per-pattern Examples text at all, mirroring
// TestGetTimelineStats_Basic (DIR-042).
func TestAnalyzeBugsStats_OmitsExamples(t *testing.T) {
	toolCalls := []types.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "permission denied: a very long diagnostic dump that would blow up response size if repeated verbatim"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "success"},
		{UUID: "uuid-3", ToolName: "Bash", Status: "error", Error: "permission denied: a very long diagnostic dump that would blow up response size if repeated verbatim"},
		{UUID: "uuid-4", ToolName: "Bash", Status: "success"},
	}

	stats, err := AnalyzeBugsStats([]types.SessionEntry{}, toolCalls)
	if err != nil {
		t.Fatalf("AnalyzeBugsStats returned error: %v", err)
	}
	if stats.TotalPairs != 2 {
		t.Errorf("Expected TotalPairs=2, got %d", stats.TotalPairs)
	}
	if stats.TotalPatterns != 1 {
		t.Errorf("Expected TotalPatterns=1, got %d", stats.TotalPatterns)
	}
	if len(stats.Patterns) != 1 {
		t.Fatalf("Expected 1 pattern, got %d", len(stats.Patterns))
	}
	if stats.Patterns[0].Recurrences != 2 {
		t.Errorf("Expected Recurrences=2, got %d", stats.Patterns[0].Recurrences)
	}

	// The BugPatternStat type has no Examples field at all -- verify this at
	// the JSON level so a future field addition would be caught.
	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("failed to marshal stats: %v", err)
	}
	if strings.Contains(string(data), "examples") {
		t.Errorf("expected no 'examples' field in BugAnalysisStats JSON, got: %s", data)
	}
	if strings.Contains(string(data), "very long diagnostic dump") {
		t.Errorf("expected no full-text error content in BugAnalysisStats JSON, got: %s", data)
	}

	if stats.DataSource != DataSourceMeasured {
		t.Errorf("Expected DataSource=%q, got %q", DataSourceMeasured, stats.DataSource)
	}
}

// sharedTestData builds [2 fixed + 98 unfixed] tool calls sharing one signature.
func fixTestData() []types.ToolCall {
	tc := make([]types.ToolCall, 0, 102)
	for i := 0; i < 2; i++ {
		tc = append(tc,
			types.ToolCall{UUID: "fixed-err", ToolName: "Bash", Status: "error", Error: "boom"},
			types.ToolCall{UUID: "fixed-ok", ToolName: "Bash", Status: "success"},
		)
	}
	for i := 0; i < 98; i++ {
		tc = append(tc, types.ToolCall{UUID: "unfixed", ToolName: "Bash", Status: "error", Error: "boom"})
	}
	return tc
}

// TestAnalyzeBugs_UnfixedErrorsCounted verifies (DIR-019): 100 errors + 2 fixes
// → Recurrences=100, FixCount=2, UnfixedErrors=98. Covers both full and stats.
func TestAnalyzeBugs_UnfixedErrorsCounted(t *testing.T) {
	tc := fixTestData()

	t.Run("full", func(t *testing.T) {
		result, err := AnalyzeBugs([]types.SessionEntry{}, tc, 0)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if len(result.Patterns) != 1 {
			t.Fatalf("want 1 pattern, got %d", len(result.Patterns))
		}
		p := result.Patterns[0]
		if p.Recurrences != 100 || p.FixCount != 2 || p.UnfixedErrors != 98 {
			t.Errorf("want R=100 F=2 U=98, got R=%d F=%d U=%d", p.Recurrences, p.FixCount, p.UnfixedErrors)
		}
		if result.TotalErrors != 100 || result.TotalPairs != 2 || result.UnfixedErrors != 98 {
			t.Errorf("want TE=100 TP=2 UE=98, got TE=%d TP=%d UE=%d",
				result.TotalErrors, result.TotalPairs, result.UnfixedErrors)
		}
	})

	t.Run("stats", func(t *testing.T) {
		stats, err := AnalyzeBugsStats([]types.SessionEntry{}, tc)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if len(stats.Patterns) != 1 {
			t.Fatalf("want 1 pattern, got %d", len(stats.Patterns))
		}
		sp := stats.Patterns[0]
		if sp.Recurrences != 100 || sp.FixCount != 2 || sp.UnfixedErrors != 98 {
			t.Errorf("want R=100 F=2 U=98, got R=%d F=%d U=%d", sp.Recurrences, sp.FixCount, sp.UnfixedErrors)
		}
		if stats.TotalErrors != 100 || stats.TotalPairs != 2 || stats.UnfixedErrors != 98 {
			t.Errorf("want TE=100 TP=2 UE=98, got TE=%d TP=%d UE=%d",
				stats.TotalErrors, stats.TotalPairs, stats.UnfixedErrors)
		}
	})
}

// TestAnalyzeBugs_FixNotDoubleCounted verifies one-to-one fix matching (DIR-019):
// two errors sharing one success → FixCount=1, UnfixedErrors=1.
func TestAnalyzeBugs_FixNotDoubleCounted(t *testing.T) {
	tc := []types.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "error", Error: "boom"},
		{UUID: "2", ToolName: "Bash", Status: "error", Error: "boom"},
		{UUID: "3", ToolName: "Bash", Status: "success"},
	}
	result, err := AnalyzeBugs([]types.SessionEntry{}, tc, 0)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(result.Patterns) != 1 {
		t.Fatalf("want 1 pattern, got %d", len(result.Patterns))
	}
	p := result.Patterns[0]
	if p.Recurrences != 2 || p.FixCount != 1 || p.UnfixedErrors != 1 {
		t.Errorf("want R=2 F=1 U=1, got R=%d F=%d U=%d", p.Recurrences, p.FixCount, p.UnfixedErrors)
	}
	if result.TotalErrors != 2 || result.TotalPairs != 1 || result.UnfixedErrors != 1 {
		t.Errorf("want TE=2 TP=1 UE=1, got TE=%d TP=%d UE=%d",
			result.TotalErrors, result.TotalPairs, result.UnfixedErrors)
	}
}

// TestAnalyzeBugs_UnfixedOnlyPatternVisible verifies (DIR-019) error patterns
// with zero fixes produce visible output.
func TestAnalyzeBugs_UnfixedOnlyPatternVisible(t *testing.T) {
	tc := []types.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "error", Error: "never fixed"},
		{UUID: "2", ToolName: "Bash", Status: "error", Error: "never fixed"},
	}
	result, err := AnalyzeBugs([]types.SessionEntry{}, tc, 0)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(result.Patterns) != 1 {
		t.Fatalf("want 1 pattern, got %d", len(result.Patterns))
	}
	p := result.Patterns[0]
	if p.FixCount != 0 || p.Recurrences != 2 || p.UnfixedErrors != 2 {
		t.Errorf("want F=0 R=2 U=2, got F=%d R=%d U=%d", p.FixCount, p.Recurrences, p.UnfixedErrors)
	}
	if len(p.Examples) != 2 {
		t.Errorf("want 2 examples, got %d", len(p.Examples))
	}
}

func TestAnalyzeBugsStats_Empty(t *testing.T) {
	stats, err := AnalyzeBugsStats([]types.SessionEntry{}, []types.ToolCall{})
	if err != nil {
		t.Fatalf("AnalyzeBugsStats returned error: %v", err)
	}
	if stats.TotalPairs != 0 {
		t.Errorf("Expected TotalPairs=0, got %d", stats.TotalPairs)
	}
	if len(stats.Patterns) != 0 {
		t.Errorf("Expected 0 patterns, got %d", len(stats.Patterns))
	}
}
