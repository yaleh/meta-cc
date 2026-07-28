package analyzer

import (
	"fmt"
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

// detToolUseEntry builds an assistant SessionEntry carrying one tool_use block.
func detToolUseEntry(uuid, ts, toolUseID, toolName string, input map[string]interface{}) types.SessionEntry {
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

// detToolResultEntry builds a user SessionEntry carrying one tool_result block.
// Status is left empty to mimic real JSONL (normalized from IsError by
// ExtractToolCalls).
func detToolResultEntry(uuid, ts, toolUseID, content string, isError bool) types.SessionEntry {
	errText := ""
	if isError {
		errText = content
	}
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
						IsError:   isError,
						Error:     errText,
					},
				},
			},
		},
	}
}

// detFixtureEntries builds 12 tool calls (interleaved assistant tool_use and
// user tool_result entries) across 5 tools with 3 errors, where each error is
// followed by a success of the same tool within 3 positions, and file-path
// tools alternate files 1 minute apart. In entry (JSONL) order the expected
// metrics are:
//
//	AnalyzeBugs TotalPairs        = 3 (tu-03→tu-04, tu-07→tu-08, tu-09→tu-10)
//	GetWorkPatterns ContextSwitches = 7
//	QualityScan retry_rate raw    = "3/12"
//
// Before the DIR-058 fix, ExtractToolCalls emitted in randomized map order,
// so these adjacency-based metrics varied between identical calls.
func detFixtureEntries() []types.SessionEntry {
	ts := func(m int) string { return fmt.Sprintf("2026-01-01T10:%02d:00Z", m) }
	fp := func(p string) map[string]interface{} { return map[string]interface{}{"file_path": p} }
	return []types.SessionEntry{
		detToolUseEntry("u01", ts(0), "tu-01", "Read", fp("/src/a.go")),
		detToolResultEntry("r01", ts(0), "tu-01", "package a", false),
		detToolUseEntry("u02", ts(1), "tu-02", "Edit", fp("/src/b.go")),
		detToolResultEntry("r02", ts(1), "tu-02", "edited b.go", false),
		detToolUseEntry("u03", ts(2), "tu-03", "Bash", map[string]interface{}{"command": "make test"}),
		detToolResultEntry("r03", ts(2), "tu-03", "make: *** boom", true), // error #1
		detToolUseEntry("u04", ts(3), "tu-04", "Bash", map[string]interface{}{"command": "make test"}),
		detToolResultEntry("r04", ts(3), "tu-04", "ok", false), // fix for #1
		detToolUseEntry("u05", ts(4), "tu-05", "Read", fp("/src/c.go")),
		detToolResultEntry("r05", ts(4), "tu-05", "package c", false),
		detToolUseEntry("u06", ts(5), "tu-06", "Write", fp("/src/a.go")),
		detToolResultEntry("r06", ts(5), "tu-06", "wrote a.go", false),
		detToolUseEntry("u07", ts(6), "tu-07", "Grep", map[string]interface{}{"path": "/src/d"}),
		detToolResultEntry("r07", ts(6), "tu-07", "no matches", true), // error #2
		detToolUseEntry("u08", ts(7), "tu-08", "Grep", map[string]interface{}{"path": "/src/d"}),
		detToolResultEntry("r08", ts(7), "tu-08", "d.go:1", false), // fix for #2
		detToolUseEntry("u09", ts(8), "tu-09", "Edit", fp("/src/e.go")),
		detToolResultEntry("r09", ts(8), "tu-09", "old_string not found", true), // error #3
		detToolUseEntry("u10", ts(9), "tu-10", "Edit", fp("/src/e.go")),
		detToolResultEntry("r10", ts(9), "tu-10", "edited e.go", false), // fix for #3
		detToolUseEntry("u11", ts(10), "tu-11", "Read", fp("/src/f.go")),
		detToolResultEntry("r11", ts(10), "tu-11", "package f", false),
		detToolUseEntry("u12", ts(11), "tu-12", "Write", fp("/src/g.go")),
		detToolResultEntry("r12", ts(11), "tu-12", "wrote g.go", false),
	}
}

// TestAnalyzers_DeterministicAcrossRepeatedCalls locks in the DIR-058 fix at
// the analyzer layer: for identical input, the order-sensitive metrics
// (AnalyzeBugs TotalPairs, GetWorkPatterns ContextSwitches, QualityScan
// retry_rate raw value) must be identical across repeated calls, and must
// match the temporal (entry-order) expectation. Each iteration extracts tool
// calls fresh via types.ExtractToolCalls — the call that was nondeterministic
// before the fix.
func TestAnalyzers_DeterministicAcrossRepeatedCalls(t *testing.T) {
	entries := detFixtureEntries()

	const (
		wantPairs       = 3
		wantSwitches    = 7
		wantRetryRaw    = "3/12"
		repeatIteration = 10
	)

	for iter := 0; iter < repeatIteration; iter++ {
		toolCalls := types.ExtractToolCalls(entries)
		if len(toolCalls) != 12 {
			t.Fatalf("iter %d: expected 12 tool calls, got %d", iter, len(toolCalls))
		}

		bugs, err := AnalyzeBugs(entries, toolCalls, 0)
		if err != nil {
			t.Fatalf("iter %d: AnalyzeBugs: %v", iter, err)
		}
		if bugs.TotalPairs != wantPairs {
			t.Fatalf("iter %d: AnalyzeBugs TotalPairs = %d, want %d (order-dependent metric drifted)",
				iter, bugs.TotalPairs, wantPairs)
		}

		wp, err := GetWorkPatterns(entries, toolCalls)
		if err != nil {
			t.Fatalf("iter %d: GetWorkPatterns: %v", iter, err)
		}
		if wp.ContextSwitches != wantSwitches {
			t.Fatalf("iter %d: GetWorkPatterns ContextSwitches = %d, want %d (order-dependent metric drifted)",
				iter, wp.ContextSwitches, wantSwitches)
		}

		qs, err := QualityScan(entries, toolCalls)
		if err != nil {
			t.Fatalf("iter %d: QualityScan: %v", iter, err)
		}
		retryRaw := retryRawValue(t, qs)
		if retryRaw != wantRetryRaw {
			t.Fatalf("iter %d: QualityScan retry_rate raw = %q, want %q (order-dependent metric drifted)",
				iter, retryRaw, wantRetryRaw)
		}
	}
}

func retryRawValue(t *testing.T, res *QualityScanResult) string {
	t.Helper()
	for _, dim := range res.Dimensions {
		if dim.Name == "retry_rate" {
			return dim.RawValue
		}
	}
	t.Fatalf("retry_rate dimension not found in %+v", res.Dimensions)
	return ""
}
