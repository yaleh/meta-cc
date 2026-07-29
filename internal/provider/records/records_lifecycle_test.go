package records

import (
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// These tests are the DIR-053 regression coverage at the providerrecords layer.
// DIR-050 fixed internal/provider/codex/rollout.go's flush() so a lifecycle-only
// turn (session_end / turn_aborted / compaction) is retained in the
// []conversation.Turn slice. But Normalize — the only layer downstream of
// loadTurnsFromRollout that turns conversation.Turn into the query-time records
// MCP tools operate on — historically read only UserText/AssistantText/ToolCalls/
// TokenUsage, so a lifecycle-only turn still produced ZERO records and stayed
// invisible to every MCP tool. These tests pin the widening at the normalized
// record layer. They are deliberately DISTINCT from DIR-050's
// internal/provider/codex/rollout_test.go tests (which assert on conversation.Turn,
// not on normalized records).

func lifecycleSession() conversation.Session {
	return conversation.Session{
		ID:       "codex-lifecycle",
		Provider: conversation.ProviderCodex,
		CWD:      "/tmp/project",
		Model:    "gpt-5",
	}
}

// TestNormalizeLifecycleOnlySessionEnd proves a turn whose ONLY content is a
// typed ItemKindSessionEnd item (no user/assistant text, no tool call, no token
// usage) still yields exactly one normalized record — type "session_end" —
// carrying the lifecycle reason, rather than zero records.
func TestNormalizeLifecycleOnlySessionEnd(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	turns := []conversation.Turn{{
		ID:        "turn-end",
		Timestamp: now,
		Items: []conversation.Item{{
			Kind:      conversation.ItemKindSessionEnd,
			Text:      "completed",
			Timestamp: now,
			Source:    "session_end",
		}},
	}}

	got := Normalize(lifecycleSession(), turns)
	if len(got) != 1 {
		t.Fatalf("expected 1 lifecycle record for a session_end-only turn, got %d: %#v", len(got), got)
	}
	rec := got[0]
	if rec["type"] != "session_end" {
		t.Fatalf("expected type=session_end, got %#v", rec["type"])
	}
	if rec["reason"] != "completed" {
		t.Fatalf("expected reason=completed, got %#v", rec["reason"])
	}
	// Canonical identity (DIR-036) must be present on lifecycle records too.
	if rec["turn_id"] != "turn-end" || rec["turn_index"] != 0 || rec["seq"] != 0 {
		t.Fatalf("lifecycle record missing canonical identity: %#v", rec)
	}
	if rec["session_id"] != "codex-lifecycle" || rec["provider"] != conversation.ProviderCodex {
		t.Fatalf("lifecycle record missing identity fields: %#v", rec)
	}
}

// TestNormalizeLifecycleOnlyTurnAborted proves the status-only analog: a turn
// whose only signal is Status == TurnStatusAborted (Codex's "turn_aborted" sets
// a TurnStatus with NO accompanying Item) still yields one record — type
// "turn_aborted" carrying turn_status — rather than being dropped.
func TestNormalizeLifecycleOnlyTurnAborted(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	turns := []conversation.Turn{{
		ID:        "turn-abort",
		Status:    conversation.TurnStatusAborted,
		Timestamp: now,
	}}

	got := Normalize(lifecycleSession(), turns)
	if len(got) != 1 {
		t.Fatalf("expected 1 lifecycle record for a turn_aborted-only turn, got %d: %#v", len(got), got)
	}
	rec := got[0]
	if rec["type"] != "turn_aborted" {
		t.Fatalf("expected type=turn_aborted, got %#v", rec["type"])
	}
	if rec["turn_status"] != "aborted" {
		t.Fatalf("expected turn_status=aborted, got %#v", rec["turn_status"])
	}
}

// TestNormalizeLifecycleOnlyCompaction proves the compaction analog: a turn
// whose only content is a typed ItemKindCompaction item carrying a
// CompactionBoundary yields one record — type "compaction" — exposing the
// boundary reason/summary.
func TestNormalizeLifecycleOnlyCompaction(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	turns := []conversation.Turn{{
		ID:        "turn-compact",
		Timestamp: now,
		Items: []conversation.Item{{
			Kind:      conversation.ItemKindCompaction,
			Timestamp: now,
			Source:    "compaction_boundary",
			Compaction: &conversation.CompactionBoundary{
				Reason:  "context_window",
				Summary: "Summarized 1 earlier exchange",
			},
		}},
	}}

	got := Normalize(lifecycleSession(), turns)
	if len(got) != 1 {
		t.Fatalf("expected 1 lifecycle record for a compaction-only turn, got %d: %#v", len(got), got)
	}
	rec := got[0]
	if rec["type"] != "compaction" {
		t.Fatalf("expected type=compaction, got %#v", rec["type"])
	}
	boundary, ok := rec["compaction"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected compaction boundary map, got %#v", rec["compaction"])
	}
	if boundary["reason"] != "context_window" || boundary["summary"] != "Summarized 1 earlier exchange" {
		t.Fatalf("unexpected compaction boundary: %#v", boundary)
	}
}

// TestNormalizeLifecycleSignalOnContentTurnDoesNotAddRecord guards the scope of
// the DIR-053 widening: lifecycle records are emitted ONLY for turns that would
// otherwise produce zero records. A turn that already surfaces content
// (assistant text here) must NOT gain an extra lifecycle record just because it
// also carries a session_end item — that would change record counts for every
// ordinary session that happens to end with a session_end event.
func TestNormalizeLifecycleSignalOnContentTurnDoesNotAddRecord(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	turns := []conversation.Turn{{
		ID:            "turn-with-content",
		AssistantText: "done",
		Timestamp:     now,
		Items: []conversation.Item{{
			Kind:      conversation.ItemKindSessionEnd,
			Text:      "completed",
			Timestamp: now,
		}},
	}}

	got := Normalize(lifecycleSession(), turns)
	if len(got) != 1 {
		t.Fatalf("expected exactly the assistant record (no extra lifecycle record), got %d: %#v", len(got), got)
	}
	if got[0]["type"] != "assistant" {
		t.Fatalf("expected the single record to be the assistant record, got %#v", got[0]["type"])
	}
}

// TestNormalizeEmptyTurnWithoutLifecycleSignalEmitsNothing guards against the
// opposite regression: a turn with neither content nor any lifecycle signal must
// still emit zero records (we only widened emission for genuine lifecycle
// signals, not for every empty turn).
func TestNormalizeEmptyTurnWithoutLifecycleSignalEmitsNothing(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	turns := []conversation.Turn{{
		ID:        "turn-nothing",
		Timestamp: now,
	}}

	got := Normalize(lifecycleSession(), turns)
	if len(got) != 0 {
		t.Fatalf("expected 0 records for a content-less, signal-less turn, got %d: %#v", len(got), got)
	}
}

// TestNormalizeInProgressOnlyTurnEmitsNothing is the cross-provider guard for
// Claude's parser, which assigns TurnStatusInProgress to a user-only turn. That
// status means merely "no assistant observed yet"; it is not a lifecycle event
// and must not manufacture a turn_in_progress record.
func TestNormalizeInProgressOnlyTurnEmitsNothing(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	turns := []conversation.Turn{{
		ID:        "turn-open",
		Status:    conversation.TurnStatusInProgress,
		Timestamp: now,
	}}

	got := Normalize(conversation.Session{
		ID:       "claude-open",
		Provider: conversation.ProviderClaude,
	}, turns)
	if len(got) != 0 {
		t.Fatalf("expected no lifecycle record for TurnStatusInProgress alone, got %d: %#v", len(got), got)
	}
}
