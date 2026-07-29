package appserver

import (
	"encoding/json"
	"testing"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// TestMapTurnDefaultsToFullCompleteness is the DIR-032 "app-server
// thread/read is a confirmed-full surface" proof: every turn mapped from a
// real thread/read(includeTurns) response is HistoryCompletenessFull, not
// unspecified/summary/unloaded — this is the only turn-content surface this
// backend uses today.
func TestMapTurnDefaultsToFullCompleteness(t *testing.T) {
	turn := mapTurn(Turn{ID: "turn-1", Status: "completed"})
	if turn.Completeness != conversation.HistoryCompletenessFull {
		t.Fatalf("expected HistoryCompletenessFull, got %q", turn.Completeness)
	}
	if !turn.Completeness.IsFull() {
		t.Fatalf("expected IsFull() == true")
	}
}

func strPtr(s string) *string { return &s }

// TestMapThreadLineage covers the three DIR-032 lineage classifications
// MapThread can currently derive: a confirmed child (parentThreadId set), a
// confirmed root (an ordinary, non-subagent thread with no parent), and the
// explicit-unknown case (a subagent-sourced thread with no reported
// parent — spawn metadata was not confirmed, so it must not be presented as
// a root).
func TestMapThreadLineage(t *testing.T) {
	cliSource, _ := json.Marshal(map[string]string{"type": "cli"})
	subAgentSource, _ := json.Marshal(map[string]string{"type": "subAgent"})

	cases := []struct {
		name   string
		thread Thread
		want   conversation.LineageStatus
	}{
		{
			name:   "child",
			thread: Thread{ID: "t1", Source: cliSource, ParentThreadID: strPtr("parent-1")},
			want:   conversation.LineageStatusChild,
		},
		{
			name:   "root",
			thread: Thread{ID: "t2", Source: cliSource},
			want:   conversation.LineageStatusRoot,
		},
		{
			name:   "subagent with suppressed parent is unknown, not root",
			thread: Thread{ID: "t3", Source: subAgentSource},
			want:   conversation.LineageStatusUnknown,
		},
		{
			name:   "subagent with reported parent is still child",
			thread: Thread{ID: "t4", Source: subAgentSource, ParentThreadID: strPtr("parent-4")},
			want:   conversation.LineageStatusChild,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			session := MapThread(tc.thread)
			if session.Lineage != tc.want {
				t.Fatalf("Lineage = %q, want %q", session.Lineage, tc.want)
			}
		})
	}
}

// TestProjectLegacyFieldsExcludesSearchAndMetadataItems is the DIR-072
// no-duplication proof for the app-server projection half: the legacy
// UserText/AssistantText/ToolCalls derivation must ignore the DIR-072 item
// kinds (tool search call/output, world_state, settings_applied) exactly
// like the rollout adapter's projection does, so a turn that ever carries
// those items through this backend can never duplicate them as legacy
// tool_use content or message text. webSearch — the app-server's own
// search-shaped item type — likewise stays a distinct ItemKindWebSearch
// item, never a ToolCall.
func TestProjectLegacyFieldsExcludesSearchAndMetadataItems(t *testing.T) {
	now := unixPtrOrZero(nil, nil)
	turn := conversation.Turn{
		ID:        "turn-1",
		UserText:  "",
		Timestamp: now,
		Items: []conversation.Item{
			{Kind: conversation.ItemKindWorldState, WorldState: &conversation.WorldState{CWD: "/tmp"}, Timestamp: now},
			{Kind: conversation.ItemKindToolSearchCall, ID: "s1", ToolCallID: "s1", ToolName: "repo_search", Input: json.RawMessage(`{"query":"x"}`), Timestamp: now},
			{Kind: conversation.ItemKindToolSearchOutput, ToolCallID: "s1", Output: "1 match", Timestamp: now},
			{Kind: conversation.ItemKindSettingsApplied, Settings: &conversation.SettingsApplied{Keys: []string{"model"}}, Timestamp: now},
			{Kind: conversation.ItemKindWebSearch, Query: "golang generics", Timestamp: now},
		},
	}
	projectLegacyFields(&turn)
	if len(turn.ToolCalls) != 0 {
		t.Fatalf("search/web-search/metadata items must not duplicate into ToolCalls: %#v", turn.ToolCalls)
	}
	if turn.UserText != "" || turn.AssistantText != "" {
		t.Fatalf("metadata items must not leak into message projections: user=%q assistant=%q", turn.UserText, turn.AssistantText)
	}
}

// TestMapItemsContextCompactionBoundary proves the "contextCompaction" item
// type maps to a typed CompactionBoundary (reason/summary), best-effort
// decoded from whatever fields the payload carries, rather than an
// ItemKindUnknown fallback.
func TestMapItemsContextCompactionBoundary(t *testing.T) {
	raw := json.RawMessage(`{"type":"contextCompaction","id":"cc-1","reason":"context_window","summary":"Trimmed 4 earlier turns"}`)
	var item ThreadItem
	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	items := mapItems(item, unixPtrOrZero(nil, nil))
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	got := items[0]
	if got.Kind != conversation.ItemKindCompaction {
		t.Fatalf("expected ItemKindCompaction, got %q", got.Kind)
	}
	if got.Compaction == nil || got.Compaction.Reason != "context_window" || got.Compaction.Summary != "Trimmed 4 earlier turns" {
		t.Fatalf("unexpected compaction boundary: %#v", got.Compaction)
	}
}
