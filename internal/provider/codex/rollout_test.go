package codex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDetectSchemaVersion(t *testing.T) {
	if got := detectSchemaVersion([]byte(`{"type":"session_meta"}`)); got != schemaLegacy {
		t.Fatalf("legacy detect failed")
	}
	if got := detectSchemaVersion([]byte(`{"type":"turn.started"}`)); got != schemaNew {
		t.Fatalf("new detect failed")
	}
}

func TestLoadTurnsFromRolloutLegacyAndNew(t *testing.T) {
	legacy, usage, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("legacy load: %v", err)
	}
	if len(legacy) != 1 || legacy[0].UserText == "" || len(legacy[0].ToolCalls) != 1 {
		t.Fatalf("unexpected legacy turns: %#v", legacy)
	}
	if usage.InputTokens != 0 {
		t.Fatalf("unexpected legacy usage: %#v", usage)
	}

	newTurns, _, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-new-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("new load: %v", err)
	}
	if len(newTurns) != 1 || newTurns[0].AssistantText == "" || len(newTurns[0].ToolCalls) != 1 {
		t.Fatalf("unexpected new turns: %#v", newTurns)
	}
}

func TestLoadTurnsFromRolloutLegacyCustomToolsAndTokenCount(t *testing.T) {
	turns, usage, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-rich-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("rich legacy load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected one turn, got %#v", turns)
	}
	turn := turns[0]
	if len(turn.ToolCalls) != 2 {
		t.Fatalf("expected function and custom tool calls, got %#v", turn.ToolCalls)
	}
	if turn.ToolCalls[1].Name != "apply_patch" || !turn.ToolCalls[1].IsError || turn.ToolCalls[1].Output != "patch failed" {
		t.Fatalf("custom tool output not normalized: %#v", turn.ToolCalls[1])
	}
	if turn.TokenUsage.InputTokens != 10 || turn.TokenUsage.OutputTokens != 3 || turn.TokenUsage.CacheTokens != 2 {
		t.Fatalf("turn usage mismatch: %#v", turn.TokenUsage)
	}
	if usage.InputTokens != 100 || usage.OutputTokens != 30 || usage.CacheTokens != 20 {
		t.Fatalf("total usage mismatch: %#v", usage)
	}
}

// TestLoadTurnsFromRolloutDedupesAssistantSegments is a regression test for
// the duplicate-assistant-text defect (DIR-027): live Codex CLI 0.145
// rollouts record the same assistant utterance through BOTH the legacy
// event_msg.agent_message channel and the response_item message(role
// assistant) channel. The fixture pairs both channels for the same segment
// within a turn (turn-1), repeats the identical text again in a distinct
// turn (turn-2, which must NOT be collapsed into turn-1 — legitimately
// repeated text across turns is preserved), and within a single turn
// (turn-3) emits two distinct assistant segments — separated by a tool
// call — each duplicated across both channels, verifying that dedup is
// scoped per-position and does not cross-collapse distinct segments.
func TestLoadTurnsFromRolloutDedupesAssistantSegments(t *testing.T) {
	turns, _, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-dedup-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("dedup load: %v", err)
	}
	if len(turns) != 3 {
		t.Fatalf("expected 3 turns, got %d: %#v", len(turns), turns)
	}

	if turns[0].AssistantText != "Summary: done." {
		t.Fatalf("turn-1: expected single deduped assistant segment, got %q", turns[0].AssistantText)
	}
	if turns[1].AssistantText != "Summary: done." {
		t.Fatalf("turn-2: expected identical text to turn-1 to be preserved (not collapsed across turns), got %q", turns[1].AssistantText)
	}
	if turns[2].AssistantText != "Starting task\nTask complete" {
		t.Fatalf("turn-3: expected two distinct deduped segments joined, got %q", turns[2].AssistantText)
	}
	if len(turns[2].ToolCalls) != 1 || turns[2].ToolCalls[0].Output != "file.txt" {
		t.Fatalf("turn-3: tool call unaffected by text dedup, got %#v", turns[2].ToolCalls)
	}

	for i, turn := range turns {
		if turn.UserText == "" {
			t.Fatalf("turn %d: expected non-empty user text, got %#v", i, turn)
		}
	}
}

// TestLoadTurnsFromRollout0145EventFamilies exercises the Codex 0.145
// event families absent from earlier fixtures: world_state, compacted,
// tool_search_call, tool_search_output, thread_settings_applied,
// context_compacted, turn_aborted, and session_end. None of these are
// semantically handled yet, but the contract is that they must not crash
// parsing, must not silently disappear, and must not interfere with the
// user/assistant text or tool calls that share their turn. They are
// expected to be preserved as raw events in the turn's Extensions
// (codex_events), the same fallback already used for any other unknown
// event/payload type.
//
// Fixture-refresh procedure: when a newer Codex CLI version changes or
// adds event families, capture a sanitized (no secrets/repo content),
// minimal rollout excerpt showing the new/changed shape, add it under
// tests/fixtures/codex/, and extend this test (or add a new one) to assert
// the parser still produces correct turns and does not drop data. See also
// docs/reference/jsonl-schema.md's "Newer Event Families (Codex 0.145+)"
// section.
func TestLoadTurnsFromRollout0145EventFamilies(t *testing.T) {
	turns, _, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-0145-families-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("0145 families load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected all events to stay within the single open turn, got %d turns: %#v", len(turns), turns)
	}

	turn := turns[0]
	if turn.UserText != "search the repo for TODOs" {
		t.Fatalf("unexpected user text: %q", turn.UserText)
	}
	if turn.AssistantText != "Found 3 TODOs." {
		t.Fatalf("expected deduped assistant text even with new event families interleaved, got %q", turn.AssistantText)
	}

	var ext struct {
		CodexEvents []json.RawMessage `json:"codex_events"`
	}
	if err := json.Unmarshal(turn.Extensions, &ext); err != nil {
		t.Fatalf("failed to unmarshal turn extensions: %v", err)
	}
	// world_state, tool_search_call, tool_search_output,
	// thread_settings_applied, context_compacted, compacted, turn_aborted,
	// session_end: 8 unrecognized events, all preserved raw.
	if len(ext.CodexEvents) != 8 {
		t.Fatalf("expected 8 preserved raw events for unrecognized 0145 families, got %d: %#v", len(ext.CodexEvents), ext.CodexEvents)
	}
}

func TestLoadTurnsFromRolloutTokenCountUsesEventTimestamp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token-first.jsonl")
	const eventTime = "2026-06-14T06:00:08Z"
	content := `{"timestamp":"` + eventTime + `","type":"event_msg","payload":{"type":"token_count","info":{"last_token_usage":{"input_tokens":10,"cached_input_tokens":2,"output_tokens":3},"total_token_usage":{"input_tokens":100,"cached_input_tokens":20,"output_tokens":30}}}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, _, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load rollout: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected one token usage turn, got %#v", turns)
	}
	want, _ := time.Parse(time.RFC3339, eventTime)
	if !turns[0].Timestamp.Equal(want) {
		t.Fatalf("token_count timestamp = %s, want %s", turns[0].Timestamp.Format(time.RFC3339), eventTime)
	}
}
