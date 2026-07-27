package conversation

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestItemRoundTrip(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	exitCode := 0
	want := Item{
		ID:         "item-1",
		Kind:       ItemKindToolCall,
		Role:       "assistant",
		Phase:      PhaseCommentary,
		Status:     StatusCompleted,
		Text:       "",
		ToolCallID: "call-1",
		ToolName:   "exec_command",
		Input:      json.RawMessage(`{"cmd":"pwd"}`),
		Output:     "/tmp",
		IsError:    false,
		Command:    "pwd",
		ExitCode:   &exitCode,
		Paths:      []string{"a.go", "b.go"},
		Query:      "",
		PlanSteps:  []string{"step 1", "step 2"},
		Timestamp:  now,
		Source:     "response_item",
	}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal item: %v", err)
	}

	var got Item
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal item: %v", err)
	}

	if got.ID != want.ID || got.Kind != want.Kind || got.Phase != want.Phase || got.Status != want.Status {
		t.Fatalf("item mismatch: %#v", got)
	}
	if got.ExitCode == nil || *got.ExitCode != 0 {
		t.Fatalf("exit code mismatch: %#v", got.ExitCode)
	}
	if len(got.Paths) != 2 || got.Paths[1] != "b.go" {
		t.Fatalf("paths mismatch: %#v", got.Paths)
	}
	if len(got.PlanSteps) != 2 {
		t.Fatalf("plan steps mismatch: %#v", got.PlanSteps)
	}
}

func TestItemKindCoverage(t *testing.T) {
	kinds := []ItemKind{
		ItemKindUserMessage,
		ItemKindAgentMessage,
		ItemKindToolCall,
		ItemKindToolResult,
		ItemKindCommandExecution,
		ItemKindFileChange,
		ItemKindWebSearch,
		ItemKindPlanUpdate,
		ItemKindReasoning,
		ItemKindCompaction,
		ItemKindUnknown,
	}
	seen := make(map[ItemKind]bool)
	for _, k := range kinds {
		if k == "" {
			t.Fatalf("item kind must not be empty: %#v", kinds)
		}
		if seen[k] {
			t.Fatalf("duplicate item kind constant: %s", k)
		}
		seen[k] = true
	}
}

func TestAgentPhaseValues(t *testing.T) {
	if PhaseCommentary != "commentary" || PhaseFinal != "final" || PhaseUnspecified != "" {
		t.Fatalf("unexpected phase constants: commentary=%q final=%q unspecified=%q", PhaseCommentary, PhaseFinal, PhaseUnspecified)
	}
}

func TestTurnStatusValues(t *testing.T) {
	if TurnStatusCompleted != "completed" || TurnStatusFailed != "failed" || TurnStatusInProgress != "in_progress" || TurnStatusUnspecified != "" {
		t.Fatalf("unexpected turn status constants")
	}
}

// TestNewRawItemCapsPayload proves that unknown/raw items never embed
// unbounded raw payloads: oversized input is truncated and flagged via
// RawTruncated, rather than round-tripping megabytes of raw JSON into every
// Turn that contains an unrecognized event.
func TestNewRawItemCapsPayload(t *testing.T) {
	small := []byte(`{"type":"world_state","cwd":"/tmp"}`)
	item := NewRawItem(ItemKindUnknown, time.Unix(1700000000, 0).UTC(), small)
	if item.RawTruncated {
		t.Fatalf("small payload should not be truncated: %#v", item)
	}
	if string(item.Raw) != string(small) {
		t.Fatalf("small payload should round-trip verbatim: got %s", item.Raw)
	}

	big := []byte(`{"type":"blob","data":"` + strings.Repeat("x", maxRawBytes*2) + `"}`)
	bigItem := NewRawItem(ItemKindUnknown, time.Unix(1700000000, 0).UTC(), big)
	if !bigItem.RawTruncated {
		t.Fatalf("oversized payload must be marked truncated")
	}
	// Raw is JSON-string-quoted when truncated (a mid-token cut is not
	// guaranteed to be valid JSON by itself), so allow a small amount of
	// quoting/escaping overhead above the byte cap rather than requiring
	// an exact bound.
	if len(bigItem.Raw) > maxRawBytes+16 {
		t.Fatalf("truncated raw payload exceeds cap: got %d bytes, want <= %d", len(bigItem.Raw), maxRawBytes+16)
	}

	// Round-trip: the capped item itself must still marshal/unmarshal
	// cleanly (it's the Item envelope that must round-trip, not the
	// truncated raw bytes as valid JSON).
	data, err := json.Marshal(bigItem)
	if err != nil {
		t.Fatalf("marshal truncated item: %v", err)
	}
	var got Item
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal truncated item: %v", err)
	}
	if !got.RawTruncated {
		t.Fatalf("truncation flag lost across round trip")
	}
}

func TestTurnItemsRoundTrip(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	turn := Turn{
		ID:        "turn-1",
		Status:    TurnStatusCompleted,
		Timestamp: now,
		Items: []Item{
			{Kind: ItemKindUserMessage, Text: "hi", Role: "user", Timestamp: now},
			{Kind: ItemKindAgentMessage, Text: "hello", Role: "assistant", Phase: PhaseFinal, Timestamp: now},
		},
	}

	data, err := json.Marshal(turn)
	if err != nil {
		t.Fatalf("marshal turn: %v", err)
	}
	var got Turn
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal turn: %v", err)
	}
	if got.Status != TurnStatusCompleted {
		t.Fatalf("status mismatch: %#v", got.Status)
	}
	if len(got.Items) != 2 || got.Items[0].Kind != ItemKindUserMessage || got.Items[1].Phase != PhaseFinal {
		t.Fatalf("items mismatch: %#v", got.Items)
	}
}
