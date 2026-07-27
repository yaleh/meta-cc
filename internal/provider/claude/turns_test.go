package claude

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// TestItemsFromPairPreservesOrderAndPairing is a DIR-028 test for the thin
// Claude-side canonical-model adapter: it must map the already-computed
// flattened fields into Items in a stable order (user, then assistant,
// then each tool call immediately followed by its result), with tool
// call/result items linked via ToolCallID, without touching the
// established buildTurns/joinToolCalls parsing pipeline.
func TestItemsFromPairPreservesOrderAndPairing(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	calls := []conversation.ToolCall{{
		ID:        "call-1",
		Name:      "Grep",
		Input:     json.RawMessage(`{"pattern":"TODO"}`),
		Output:    "3 matches",
		Timestamp: ts,
	}}

	items := itemsFromPair("turn-1", "find TODOs", "found them", calls, ts)

	wantKinds := []conversation.ItemKind{
		conversation.ItemKindUserMessage,
		conversation.ItemKindAgentMessage,
		conversation.ItemKindToolCall,
		conversation.ItemKindToolResult,
	}
	if len(items) != len(wantKinds) {
		t.Fatalf("expected %d items, got %d: %#v", len(wantKinds), len(items), items)
	}
	for i, kind := range wantKinds {
		if items[i].Kind != kind {
			t.Fatalf("item %d: expected kind %s, got %s", i, kind, items[i].Kind)
		}
	}
	if items[0].Text != "find TODOs" || items[1].Text != "found them" {
		t.Fatalf("unexpected message text: %#v", items[:2])
	}
	if items[2].ToolCallID != "call-1" || items[3].ToolCallID != "call-1" {
		t.Fatalf("tool call/result not linked via ToolCallID: %#v", items[2:])
	}
	if items[3].Output != "3 matches" {
		t.Fatalf("tool result output not carried over: %#v", items[3])
	}
}

// TestItemsFromPairOmitsEmptySegments ensures a turn with no user text (or
// no tool calls) doesn't produce spurious empty Items.
func TestItemsFromPairOmitsEmptySegments(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	items := itemsFromPair("turn-2", "", "just an answer", nil, ts)
	if len(items) != 1 || items[0].Kind != conversation.ItemKindAgentMessage {
		t.Fatalf("expected exactly one agent message item, got %#v", items)
	}
}
