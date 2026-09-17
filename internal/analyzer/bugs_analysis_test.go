package analyzer

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

func TestAnalyzeBugs_FixPair(t *testing.T) {
	toolCalls := []types.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "command not found"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "success"},
	}

	result, err := AnalyzeBugs([]types.SessionEntry{}, toolCalls, 0, 0)
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

	result, err := AnalyzeBugs([]types.SessionEntry{}, toolCalls, 0, 0)
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

	result, err := AnalyzeBugs([]types.SessionEntry{}, toolCalls, 0, 0)
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
	result, err := AnalyzeBugs([]types.SessionEntry{}, []types.ToolCall{}, 0, 0)
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
	result, err := AnalyzeBugs([]types.SessionEntry{}, []types.ToolCall{}, 0, 0)
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
		result, err := AnalyzeBugs([]types.SessionEntry{}, tc, 0, 0)
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
	result, err := AnalyzeBugs([]types.SessionEntry{}, tc, 0, 0)
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
	result, err := AnalyzeBugs([]types.SessionEntry{}, tc, 0, 0)
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

// --- DIR-096: structured examples and the max_patterns cap ---

// exampleEntries builds one SessionEntry per tool call, matching UUIDs so the
// session-id lookup in AnalyzeBugs resolves. SessionID is deliberately carried
// only on the entries (not on ToolCall) to mirror the real types.
func exampleEntries(sessionID string, toolCalls []types.ToolCall) []types.SessionEntry {
	entries := make([]types.SessionEntry, 0, len(toolCalls))
	for _, tc := range toolCalls {
		entries = append(entries, types.SessionEntry{
			Type:      "assistant",
			UUID:      tc.UUID,
			SessionID: sessionID,
			Timestamp: tc.Timestamp,
		})
	}
	return entries
}

// TestAnalyzeBugs_ExamplesAreStructuredObjects pins AC #2: examples are objects
// addressing each occurrence by session/time/error rather than bare strings,
// and carry the paired fix excerpt plus the signature when available.
func TestAnalyzeBugs_ExamplesAreStructuredObjects(t *testing.T) {
	toolCalls := []types.ToolCall{
		{UUID: "u1", ToolName: "Bash", Status: "error", Error: "command not found", Timestamp: "2026-01-01T00:00:00Z"},
		{UUID: "u2", ToolName: "Bash", Status: "success", Output: "ok", Timestamp: "2026-01-01T00:00:01Z"},
		// A second occurrence of the same signature that is never fixed.
		{UUID: "u3", ToolName: "Bash", Status: "error", Error: "command not found", Timestamp: "2026-01-01T00:00:02Z"},
	}
	result, err := AnalyzeBugs(exampleEntries("sess-abc", toolCalls), toolCalls, 0, 0)
	if err != nil {
		t.Fatalf("AnalyzeBugs: %v", err)
	}
	if len(result.Patterns) != 1 {
		t.Fatalf("want 1 pattern, got %d", len(result.Patterns))
	}
	examples := result.Patterns[0].Examples
	if len(examples) != 2 {
		t.Fatalf("want 2 examples, got %d", len(examples))
	}

	// JSON-level assertion too: the consumer that hit an AttributeError was
	// parsing this as JSON, so pin the wire shape, not just the Go struct.
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var wire struct {
		Patterns []struct {
			Examples []map[string]interface{} `json:"examples"`
		} `json:"patterns"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(wire.Patterns) != 1 || len(wire.Patterns[0].Examples) != 2 {
		t.Fatalf("wire shape: want 1 pattern with 2 examples, got %s", data)
	}

	fixed := examples[0]
	if fixed.SessionID != "sess-abc" {
		t.Errorf("want session_id sess-abc, got %q", fixed.SessionID)
	}
	if fixed.Timestamp != "2026-01-01T00:00:00Z" {
		t.Errorf("unexpected timestamp %q", fixed.Timestamp)
	}
	if fixed.ErrorText != "command not found" {
		t.Errorf("unexpected error_text %q", fixed.ErrorText)
	}
	if fixed.FixText != "ok" {
		t.Errorf("want fix_text \"ok\" on the paired example, got %q", fixed.FixText)
	}
	if fixed.Signature == "" {
		t.Error("want non-empty signature")
	}
	if fixed.Signature != result.Patterns[0].ErrorSignature {
		t.Errorf("example signature %q must match pattern signature %q",
			fixed.Signature, result.Patterns[0].ErrorSignature)
	}

	unfixed := examples[1]
	if unfixed.FixText != "" {
		t.Errorf("unfixed occurrence must omit fix_text, got %q", unfixed.FixText)
	}
	if unfixed.Timestamp != "2026-01-01T00:00:02Z" {
		t.Errorf("unexpected timestamp for unfixed example: %q", unfixed.Timestamp)
	}
	if unfixed.SessionID != "sess-abc" {
		t.Errorf("want session_id on unfixed example too, got %q", unfixed.SessionID)
	}
}

// TestAnalyzeBugs_MaxPatternsCapsAndRanks pins AC #1 and the ranking rule:
// the cap keeps the most recurrent patterns, and the pre-cap count stays
// visible as total_patterns so truncation is never silent.
func TestAnalyzeBugs_MaxPatternsCapsAndRanks(t *testing.T) {
	// 8 distinct signatures with descending recurrence (8,7,...,1).
	var toolCalls []types.ToolCall
	for p := 0; p < 8; p++ {
		for r := 0; r < 8-p; r++ {
			toolCalls = append(toolCalls, types.ToolCall{
				UUID:     fmt.Sprintf("u-%d-%d", p, r),
				ToolName: "Bash",
				Status:   "error",
				Error:    fmt.Sprintf("distinct error %d", p),
			})
		}
	}

	uncapped, err := AnalyzeBugs(nil, toolCalls, 0, 0)
	if err != nil {
		t.Fatalf("AnalyzeBugs: %v", err)
	}
	if len(uncapped.Patterns) != 8 {
		t.Fatalf("want 8 patterns uncapped, got %d", len(uncapped.Patterns))
	}

	capped, err := AnalyzeBugs(nil, toolCalls, 0, 5)
	if err != nil {
		t.Fatalf("AnalyzeBugs: %v", err)
	}
	if len(capped.Patterns) > 5 {
		t.Errorf("AC #1: max_patterns=5 returned %d patterns", len(capped.Patterns))
	}
	if len(capped.Patterns) != 5 {
		t.Errorf("want 5 patterns, got %d", len(capped.Patterns))
	}
	if capped.TotalPatterns != 8 {
		t.Errorf("want total_patterns=8 (pre-cap), got %d", capped.TotalPatterns)
	}
	// The cap must select the same top-N the uncapped ranking produced.
	for i := range capped.Patterns {
		if capped.Patterns[i].ErrorSignature != uncapped.Patterns[i].ErrorSignature {
			t.Errorf("pattern %d: capped %q != uncapped %q",
				i, capped.Patterns[i].ErrorSignature, uncapped.Patterns[i].ErrorSignature)
		}
	}
	// Aggregate totals describe the whole corpus, not the returned slice.
	if capped.TotalErrors != uncapped.TotalErrors {
		t.Errorf("total_errors must be corpus-wide: %d != %d", capped.TotalErrors, uncapped.TotalErrors)
	}

	// A cap larger than the corpus is a no-op, and a corpus smaller than the
	// cap returns everything (AC #1 is "at most", not "exactly").
	small, err := AnalyzeBugs(nil, toolCalls, 0, 100)
	if err != nil {
		t.Fatalf("AnalyzeBugs: %v", err)
	}
	if len(small.Patterns) != 8 {
		t.Errorf("cap above corpus size must be a no-op, got %d patterns", len(small.Patterns))
	}
}

// TestAnalyzeBugs_MaxPatternsIsDeterministic guards the ranking's total order:
// patterns are gathered by ranging over a map, so without the signature
// tiebreak a cap would return a different subset on each call.
func TestAnalyzeBugs_MaxPatternsIsDeterministic(t *testing.T) {
	// Every pattern has identical recurrence, forcing the tiebreak.
	var toolCalls []types.ToolCall
	for p := 0; p < 12; p++ {
		toolCalls = append(toolCalls, types.ToolCall{
			UUID:     fmt.Sprintf("u-%d", p),
			ToolName: "Bash",
			Status:   "error",
			Error:    fmt.Sprintf("tied error %d", p),
		})
	}

	var first []string
	for i := 0; i < 25; i++ {
		result, err := AnalyzeBugs(nil, toolCalls, 0, 4)
		if err != nil {
			t.Fatalf("AnalyzeBugs: %v", err)
		}
		got := make([]string, len(result.Patterns))
		for j, p := range result.Patterns {
			got[j] = p.ErrorSignature
		}
		if i == 0 {
			first = got
			continue
		}
		if strings.Join(got, ",") != strings.Join(first, ",") {
			t.Fatalf("iteration %d returned %v, first returned %v", i, got, first)
		}
	}
}

// TestAnalyzeBugs_FixTextIsBounded verifies one oversized success output cannot
// dominate the payload — the fix excerpt is capped regardless of output size.
func TestAnalyzeBugs_FixTextIsBounded(t *testing.T) {
	huge := strings.Repeat("y", 50_000)
	toolCalls := []types.ToolCall{
		{UUID: "u1", ToolName: "Bash", Status: "error", Error: "boom"},
		{UUID: "u2", ToolName: "Bash", Status: "success", Output: huge},
	}
	result, err := AnalyzeBugs(nil, toolCalls, 0, 0)
	if err != nil {
		t.Fatalf("AnalyzeBugs: %v", err)
	}
	got := result.Patterns[0].Examples[0].FixText
	if len(got) > maxExampleFixTextBytes {
		t.Errorf("fix_text must be capped at %d bytes, got %d", maxExampleFixTextBytes, len(got))
	}
	if got == "" {
		t.Error("fix_text must be present when a fix was paired")
	}
}

// TestAnalyzeBugs_FixTextNotAttachedWhenExampleBudgetFull verifies the fix is
// still counted when the per-pattern example budget is exhausted.
func TestAnalyzeBugs_FixTextNotAttachedWhenExampleBudgetFull(t *testing.T) {
	toolCalls := []types.ToolCall{
		{UUID: "u1", ToolName: "Bash", Status: "error", Error: "boom"},
		{UUID: "u2", ToolName: "Bash", Status: "success", Output: "first fix"},
		{UUID: "u3", ToolName: "Bash", Status: "error", Error: "boom"},
		{UUID: "u4", ToolName: "Bash", Status: "success", Output: "second fix"},
	}
	result, err := AnalyzeBugs(nil, toolCalls, 1, 0) // limit 1 example per pattern
	if err != nil {
		t.Fatalf("AnalyzeBugs: %v", err)
	}
	p := result.Patterns[0]
	if len(p.Examples) != 1 {
		t.Fatalf("want 1 example, got %d", len(p.Examples))
	}
	if p.FixCount != 2 {
		t.Errorf("both fixes must still be counted, got %d", p.FixCount)
	}
	if p.Examples[0].FixText != "first fix" {
		t.Errorf("want the first fix attached, got %q", p.Examples[0].FixText)
	}
}
