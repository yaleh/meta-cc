package conversation

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
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
		ItemKindSessionEnd,
		ItemKindWorldState,
		ItemKindToolSearchCall,
		ItemKindToolSearchOutput,
		ItemKindSettingsApplied,
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
	if TurnStatusAborted != "aborted" {
		t.Fatalf("unexpected TurnStatusAborted constant: %q", TurnStatusAborted)
	}
}

// TestHistoryCompletenessIsFull is the DIR-032 "a placeholder is never
// silently treated as complete" proof at the type level: only the zero
// value and HistoryCompletenessFull report IsFull()==true; every other
// state (summary/unloaded/truncated/unavailable) must not.
func TestHistoryCompletenessIsFull(t *testing.T) {
	full := map[HistoryCompleteness]bool{
		HistoryCompletenessUnspecified: true,
		HistoryCompletenessFull:        true,
		HistoryCompletenessSummary:     false,
		HistoryCompletenessUnloaded:    false,
		HistoryCompletenessTruncated:   false,
		HistoryCompletenessUnavailable: false,
	}
	for c, want := range full {
		if got := c.IsFull(); got != want {
			t.Fatalf("HistoryCompleteness(%q).IsFull() = %v, want %v", c, got, want)
		}
	}
}

// TestCompactionBoundaryRoundTrip proves an ItemKindCompaction item's typed
// boundary metadata (reason/summary) round-trips through JSON without being
// folded into Text (which would risk it being treated as ordinary message
// content by a naive consumer).
func TestCompactionBoundaryRoundTrip(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	item := Item{
		Kind:      ItemKindCompaction,
		Timestamp: now,
		Compaction: &CompactionBoundary{
			Reason:  "context_window",
			Summary: "Trimmed 4 earlier turns",
		},
	}
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Item
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Compaction == nil || got.Compaction.Reason != "context_window" || got.Compaction.Summary != "Trimmed 4 earlier turns" {
		t.Fatalf("compaction boundary mismatch: %#v", got.Compaction)
	}
	if got.Text != "" {
		t.Fatalf("compaction boundary must not leak into Text: %q", got.Text)
	}
}

func TestLineageStatusValues(t *testing.T) {
	if LineageStatusRoot != "root" || LineageStatusChild != "child" || LineageStatusUnknown != "unknown" || LineageStatusUnspecified != "" {
		t.Fatalf("unexpected lineage status constants")
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

// TestWorldStateCapsValues is the DIR-072 size-bound proof for typed
// world-state metadata: every retained scalar field is capped so a
// malformed or hostile world_state snapshot (e.g. a multi-megabyte "cwd")
// can never embed unbounded data into a Turn, while ordinary values pass
// through untouched.
func TestWorldStateCapsValues(t *testing.T) {
	long := strings.Repeat("x", maxWorldStateValueLen*3)
	ws := NewWorldState(long, "on-request", "danger-full-access")
	if len(ws.CWD) > maxWorldStateValueLen {
		t.Fatalf("world-state cwd must be capped at %d bytes, got %d", maxWorldStateValueLen, len(ws.CWD))
	}
	if ws.ApprovalPolicy != "on-request" || ws.SandboxPolicy != "danger-full-access" {
		t.Fatalf("ordinary world-state fields must pass through: %#v", ws)
	}
	// The cap must not split a multi-byte rune (the result stays valid UTF-8).
	multiByte := strings.Repeat("世", maxWorldStateValueLen)
	if got := NewWorldState(multiByte, "", "").CWD; !utf8.ValidString(got) {
		t.Fatalf("capped world-state value must remain valid UTF-8")
	}
}

// TestSettingsAppliedRetainsKeysNeverValues is the DIR-072 privacy proof for
// typed thread-settings metadata: the item enumerates ONLY the applied
// setting keys (sorted, so callers can filter/summarize by them) and never
// embeds any setting value — values can carry credentials (env keys,
// tokens), so they must remain available only through explicit raw access
// to the source rollout, never through the typed item.
func TestSettingsAppliedRetainsKeysNeverValues(t *testing.T) {
	settings := map[string]json.RawMessage{
		"model":           json.RawMessage(`"gpt-5"`),
		"approval_policy": json.RawMessage(`"on-request"`),
		"env_key":         json.RawMessage(`"sk-super-secret-value"`),
	}
	s := NewSettingsApplied(settings)
	want := []string{"approval_policy", "env_key", "model"}
	if len(s.Keys) != len(want) || s.Truncated {
		t.Fatalf("unexpected settings keys: %#v", s)
	}
	for i, key := range want {
		if s.Keys[i] != key {
			t.Fatalf("settings keys must be sorted: got %#v, want %#v", s.Keys, want)
		}
	}

	item := Item{Kind: ItemKindSettingsApplied, Settings: s, Timestamp: time.Unix(1700000000, 0).UTC()}
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, leaked := range []string{"sk-super-secret-value", "gpt-5", "on-request"} {
		if strings.Contains(string(data), leaked) {
			t.Fatalf("setting value %q must never be embedded in the typed item: %s", leaked, data)
		}
	}
}

// TestSettingsAppliedCapsKeys proves the key enumeration itself is bounded:
// a settings object with more than maxSettingsKeys entries is truncated and
// flagged via Truncated, so a pathological settings payload cannot balloon
// the item.
func TestSettingsAppliedCapsKeys(t *testing.T) {
	big := map[string]json.RawMessage{}
	for i := 0; i < maxSettingsKeys+7; i++ {
		big[fmt.Sprintf("key-%03d", i)] = json.RawMessage(`"v"`)
	}
	s := NewSettingsApplied(big)
	if len(s.Keys) != maxSettingsKeys {
		t.Fatalf("settings keys must be capped at %d, got %d", maxSettingsKeys, len(s.Keys))
	}
	if !s.Truncated {
		t.Fatalf("overflowing settings must set Truncated: %#v", s)
	}
	if empty := NewSettingsApplied(nil); empty.Truncated || len(empty.Keys) != 0 {
		t.Fatalf("empty settings must yield an untruncated empty enumeration: %#v", empty)
	}
}

// TestTypedMetadataRoundTrip proves the DIR-072 typed payload fields
// (WorldState, Settings) survive a JSON round trip on Item, and that the
// tool-search correlation fields (ToolCallID/Status/IsError) do too, so
// downstream consumers can rely on them across serialization.
func TestTypedMetadataRoundTrip(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	items := []Item{
		{
			Kind:       ItemKindWorldState,
			WorldState: &WorldState{CWD: "/tmp/project", ApprovalPolicy: "on-request"},
			Timestamp:  now, Source: "world_state",
		},
		{
			Kind:      ItemKindSettingsApplied,
			Settings:  &SettingsApplied{Keys: []string{"model"}, Truncated: true},
			Timestamp: now, Source: "event_msg",
		},
		{
			Kind: ItemKindToolSearchCall, ID: "call-s1", ToolCallID: "call-s1",
			ToolName: "repo_search", Input: json.RawMessage(`{"query":"TODO"}`),
			Timestamp: now, Source: "response_item",
		},
		{
			Kind: ItemKindToolSearchOutput, ToolCallID: "call-s1", Output: "no matches",
			IsError: true, Status: StatusFailed, Timestamp: now, Source: "response_item",
		},
	}
	data, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got []Item
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got[0].WorldState == nil || got[0].WorldState.CWD != "/tmp/project" {
		t.Fatalf("world state metadata lost: %#v", got[0])
	}
	if got[1].Settings == nil || !got[1].Settings.Truncated || got[1].Settings.Keys[0] != "model" {
		t.Fatalf("settings metadata lost: %#v", got[1])
	}
	if got[3].ToolCallID != got[2].ToolCallID || got[3].Status != StatusFailed || !got[3].IsError {
		t.Fatalf("tool search correlation/status lost: %#v", got)
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
