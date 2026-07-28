package types_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

// toolUseEntry builds an assistant SessionEntry carrying a single tool_use block.
func toolUseEntry(uuid, ts, toolUseID, toolName string, input map[string]interface{}) types.SessionEntry {
	if input == nil {
		input = map[string]interface{}{}
	}
	input["marker"] = toolUseID
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

// toolResultEntry builds a user SessionEntry carrying one tool_result block.
func toolResultEntry(uuid, ts, toolUseID, content string, isError bool) types.SessionEntry {
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
						Error:     errText(isError, content),
					},
				},
			},
		},
	}
}

func errText(isError bool, content string) string {
	if isError {
		return content
	}
	return ""
}

// buildDeterminismFixture returns a []SessionEntry with 12 tool_use blocks
// (23 content blocks total) across 5 distinct tools (Read, Bash, Edit, Write,
// Grep) with 3 error results, interleaved so that:
//   - two tool_use entries precede their results (tu-03/tu-04),
//   - one assistant entry carries two tool_use blocks (tu-05, tu-06),
//   - one user entry carries two tool_result blocks in reverse order
//     (result for tu-06 before result for tu-05),
//   - tu-12 has no result at all.
//
// The second return value is the expected deterministic emit order: the order
// in which the tool_use blocks appear when walking entries in slice order
// (within an entry, content-block order).
func buildDeterminismFixture() ([]types.SessionEntry, []string) {
	ts := func(m int) string { return fmt.Sprintf("2026-01-01T10:%02d:00Z", m) }
	entries := []types.SessionEntry{
		toolUseEntry("uuid-01", ts(0), "tu-01", "Read", map[string]interface{}{"file_path": "/src/a.go"}),
		toolResultEntry("uuid-r01", ts(0), "tu-01", "package a", false),
		toolUseEntry("uuid-02", ts(1), "tu-02", "Bash", map[string]interface{}{"command": "make test"}),
		toolResultEntry("uuid-r02", ts(1), "tu-02", "exit code 1", true), // error #1
		toolUseEntry("uuid-03", ts(2), "tu-03", "Edit", map[string]interface{}{"file_path": "/src/b.go"}),
		toolUseEntry("uuid-04", ts(3), "tu-04", "Write", map[string]interface{}{"file_path": "/src/c.go"}),
		toolResultEntry("uuid-r03", ts(3), "tu-03", "edited b.go", false),
		toolResultEntry("uuid-r04", ts(4), "tu-04", "wrote c.go", false),
		// One entry with TWO tool_use blocks: emit order must follow block order.
		{
			Type:      "assistant",
			UUID:      "uuid-05-06",
			Timestamp: ts(5),
			Message: &types.Message{
				Role: "assistant",
				Content: []types.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &types.ToolUse{
							ID:    "tu-05",
							Name:  "Grep",
							Input: map[string]interface{}{"pattern": "foo", "marker": "tu-05"},
						},
					},
					{
						Type: "tool_use",
						ToolUse: &types.ToolUse{
							ID:    "tu-06",
							Name:  "Read",
							Input: map[string]interface{}{"file_path": "/src/d.go", "marker": "tu-06"},
						},
					},
				},
			},
		},
		// One entry with TWO tool_result blocks, reversed relative to use order.
		{
			Type:      "user",
			UUID:      "uuid-r05-06",
			Timestamp: ts(6),
			Message: &types.Message{
				Role: "user",
				Content: []types.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &types.ToolResult{
							ToolUseID: "tu-06", Content: "package d", IsError: false,
						},
					},
					{
						Type: "tool_result",
						ToolResult: &types.ToolResult{
							ToolUseID: "tu-05", Content: "no matches", IsError: true, Error: "no matches", // error #2
						},
					},
				},
			},
		},
		toolUseEntry("uuid-07", ts(7), "tu-07", "Bash", map[string]interface{}{"command": "ls"}),
		toolResultEntry("uuid-r07", ts(7), "tu-07", "a.go b.go", false),
		toolUseEntry("uuid-08", ts(8), "tu-08", "Read", map[string]interface{}{"file_path": "/src/e.go"}),
		toolResultEntry("uuid-r08", ts(8), "tu-08", "package e", false),
		toolUseEntry("uuid-09", ts(9), "tu-09", "Edit", map[string]interface{}{"file_path": "/src/e.go"}),
		toolResultEntry("uuid-r09", ts(9), "tu-09", "old_string not found", true), // error #3
		toolUseEntry("uuid-10", ts(10), "tu-10", "Edit", map[string]interface{}{"file_path": "/src/e.go"}),
		toolResultEntry("uuid-r10", ts(10), "tu-10", "edited e.go", false),
		toolUseEntry("uuid-11", ts(11), "tu-11", "Grep", map[string]interface{}{"pattern": "bar"}),
		toolResultEntry("uuid-r11", ts(11), "tu-11", "e.go:1", false),
		toolUseEntry("uuid-12", ts(12), "tu-12", "Write", map[string]interface{}{"file_path": "/src/f.go"}),
		// tu-12 has no tool_result.
	}
	wantOrder := []string{
		"tu-01", "tu-02", "tu-03", "tu-04", "tu-05", "tu-06",
		"tu-07", "tu-08", "tu-09", "tu-10", "tu-11", "tu-12",
	}
	return entries, wantOrder
}

// markerOf extracts the tool_use marker stashed in Input by the fixture builder.
func markerOf(t *testing.T, tc types.ToolCall) string {
	t.Helper()
	m, _ := tc.Input["marker"].(string)
	if m == "" {
		t.Fatalf("tool call %+v has no marker in Input", tc)
	}
	return m
}

// TestExtractToolCalls_DeterministicEntryOrder locks in the DIR-058 fix:
// ExtractToolCalls must emit ToolCalls in entry (JSONL) order, identically on
// every call. Regression guard against map-iteration-order emit: with 12
// tool_use blocks, a map-range emit would match the expected order with
// probability 1/12! per iteration, so 50 iterations failing is a certainty
// for the buggy code.
func TestExtractToolCalls_DeterministicEntryOrder(t *testing.T) {
	entries, wantOrder := buildDeterminismFixture()

	var first []types.ToolCall
	for iter := 0; iter < 50; iter++ {
		got := types.ExtractToolCalls(entries)
		if len(got) != len(wantOrder) {
			t.Fatalf("iter %d: expected %d tool calls, got %d", iter, len(wantOrder), len(got))
		}
		for i, want := range wantOrder {
			if m := markerOf(t, got[i]); m != want {
				t.Fatalf("iter %d: output order not entry order at position %d: got %q, want %q (full order: %v)",
					iter, i, m, want, markers(got))
			}
		}
		if iter == 0 {
			first = got
			continue
		}
		if !reflect.DeepEqual(first, got) {
			t.Fatalf("iter %d: output not byte-identical to first iteration:\nfirst: %v\nthis:  %v",
				iter, markers(first), markers(got))
		}
	}
}

func markers(calls []types.ToolCall) []string {
	out := make([]string, len(calls))
	for i, tc := range calls {
		out[i], _ = tc.Input["marker"].(string)
	}
	return out
}

// TestExtractToolCalls_EntryOrderPairingIntact verifies that emitting in entry
// order preserves the existing tool_use↔tool_result pairing semantics
// (output, status normalization, unmatched calls).
func TestExtractToolCalls_EntryOrderPairingIntact(t *testing.T) {
	entries, _ := buildDeterminismFixture()
	calls := types.ExtractToolCalls(entries)
	byMarker := make(map[string]types.ToolCall, len(calls))
	for _, tc := range calls {
		byMarker[markerOf(t, tc)] = tc
	}

	// Paired success (is_error:false with no explicit status → normalized).
	if tc := byMarker["tu-01"]; tc.Status != "success" || tc.Output != "package a" {
		t.Errorf("tu-01: expected success/'package a', got %q/%q", tc.Status, tc.Output)
	}
	// Paired error (is_error:true → normalized to error, Error populated).
	if tc := byMarker["tu-02"]; tc.Status != "error" || tc.Error != "exit code 1" {
		t.Errorf("tu-02: expected error/'exit code 1', got %q/%q", tc.Status, tc.Error)
	}
	// Error from the multi-result user entry.
	if tc := byMarker["tu-05"]; tc.Status != "error" {
		t.Errorf("tu-05: expected error, got %q", tc.Status)
	}
	// Success from the same multi-result entry.
	if tc := byMarker["tu-06"]; tc.Status != "success" {
		t.Errorf("tu-06: expected success, got %q", tc.Status)
	}
	// Unmatched tool_use: no status, no output.
	if tc := byMarker["tu-12"]; tc.Status != "" || tc.Output != "" {
		t.Errorf("tu-12: expected empty status/output for unmatched call, got %q/%q", tc.Status, tc.Output)
	}
	// UUID must come from the entry containing the tool_use block.
	if tc := byMarker["tu-05"]; tc.UUID != "uuid-05-06" {
		t.Errorf("tu-05: expected UUID 'uuid-05-06', got %q", tc.UUID)
	}
}

// TestExtractToolCalls_DuplicateToolUseIDEmittedOnce locks in the dedup
// semantics: a tool_use ID appearing more than once yields exactly one
// ToolCall, at the position of its first occurrence.
func TestExtractToolCalls_DuplicateToolUseIDEmittedOnce(t *testing.T) {
	entries := []types.SessionEntry{
		toolUseEntry("u1", "2026-01-01T10:00:00Z", "tu-a", "Read", nil),
		toolUseEntry("u2", "2026-01-01T10:01:00Z", "tu-b", "Edit", nil),
		toolUseEntry("u3", "2026-01-01T10:02:00Z", "tu-a", "Read", nil), // duplicate ID
		toolResultEntry("u4", "2026-01-01T10:03:00Z", "tu-a", "ok", false),
		toolResultEntry("u5", "2026-01-01T10:04:00Z", "tu-b", "ok", false),
	}
	calls := types.ExtractToolCalls(entries)
	if len(calls) != 2 {
		t.Fatalf("expected 2 tool calls (dedup by tool_use ID), got %d: %v", len(calls), markers(calls))
	}
	if got := markers(calls); got[0] != "tu-a" || got[1] != "tu-b" {
		t.Fatalf("expected order [tu-a tu-b] (first occurrence), got %v", got)
	}
	if calls[0].UUID != "u1" {
		t.Errorf("expected first-occurrence UUID 'u1', got %q", calls[0].UUID)
	}
}
