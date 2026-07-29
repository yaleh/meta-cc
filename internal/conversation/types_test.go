package conversation

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRoundTripSession(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	want := Session{
		ID:        "sess-1",
		Provider:  ProviderCodex,
		Title:     "title",
		CWD:       "/tmp/project",
		Model:     "gpt-5",
		CreatedAt: now,
		TokenUsage: TokenUsage{
			InputTokens:  10,
			OutputTokens: 20,
			CacheTokens:  3,
		},
		Turns: []Turn{{
			ID:        "turn-1",
			UserText:  "hello",
			Timestamp: now,
		}},
		Extensions: json.RawMessage(`{"rollout_path":"x"}`),
	}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal session: %v", err)
	}

	var got Session
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal session: %v", err)
	}

	if got.ID != want.ID || got.Provider != want.Provider || string(got.Extensions) != string(want.Extensions) {
		t.Fatalf("session mismatch: %#v", got)
	}
}

func TestRoundTripTurnAndToolCall(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	want := Turn{
		ID:            "turn-1",
		UserText:      "user",
		AssistantText: "assistant",
		Timestamp:     now,
		ToolCalls: []ToolCall{{
			ID:        "call-1",
			Name:      "exec_command",
			Input:     json.RawMessage(`{"cmd":"pwd"}`),
			Output:    "/tmp",
			IsError:   false,
			Timestamp: now,
		}},
		Extensions: json.RawMessage(`{"x":1}`),
	}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal turn: %v", err)
	}

	var got Turn
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal turn: %v", err)
	}

	if got.ID != want.ID || got.ToolCalls[0].Name != "exec_command" || string(got.Extensions) != `{"x":1}` {
		t.Fatalf("turn mismatch: %#v", got)
	}
}

func TestProviderConstantsAndOmitempty(t *testing.T) {
	if ProviderClaude != "claude" || ProviderCodex != "codex" {
		t.Fatalf("unexpected provider constants")
	}

	data, err := json.Marshal(Session{
		ID:        "sess-1",
		Provider:  ProviderClaude,
		CWD:       "/tmp",
		CreatedAt: time.Unix(1700000000, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("marshal session: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}

	if _, ok := raw["turns"]; ok {
		t.Fatalf("turns should be omitted")
	}
	if _, ok := raw["extensions"]; ok {
		t.Fatalf("extensions should be omitted")
	}
}

// TestTokenUsageHasAny covers the DIR-071 shared "does this usage carry
// anything?" predicate: every category — including reasoning output and an
// opaque aggregate — must count, so a usage that reports ONLY reasoning or
// ONLY an aggregate is never mistaken for empty.
func TestTokenUsageHasAny(t *testing.T) {
	if (TokenUsage{}).HasAny() {
		t.Fatalf("zero TokenUsage must report HasAny=false")
	}
	cases := map[string]TokenUsage{
		"input":     {InputTokens: 1},
		"output":    {OutputTokens: 1},
		"cache":     {CacheTokens: 1},
		"reasoning": {ReasoningOutputTokens: 1},
		"aggregate": {AggregateTokens: 1, AggregateSource: AggregateSourceCodexSQLite},
	}
	for name, usage := range cases {
		if !usage.HasAny() {
			t.Fatalf("HasAny=false for %s usage: %#v", name, usage)
		}
	}
}

// TestTokenUsageReasoningAndAggregateJSON verifies the DIR-071 fields survive a
// JSON round trip and that the reasoning/aggregate fields are omitted when
// unset (backward-compatible: legacy payloads gain no new zero-valued keys).
func TestTokenUsageReasoningAndAggregateJSON(t *testing.T) {
	data, err := json.Marshal(TokenUsage{InputTokens: 10, OutputTokens: 3})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, omitted := range []string{"reasoning_output_tokens", "aggregate_tokens", "aggregate_source"} {
		if _, ok := raw[omitted]; ok {
			t.Fatalf("%s should be omitted when unset: %#v", omitted, raw)
		}
	}

	full := TokenUsage{
		InputTokens:           10,
		OutputTokens:          3,
		ReasoningOutputTokens: 4,
		AggregateTokens:       99,
		AggregateSource:       AggregateSourceCodexSQLite,
	}
	data, err = json.Marshal(full)
	if err != nil {
		t.Fatalf("marshal full: %v", err)
	}
	var back TokenUsage
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal full: %v", err)
	}
	if back != full {
		t.Fatalf("round trip mismatch: got %#v, want %#v", back, full)
	}
}
