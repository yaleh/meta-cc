package codex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
	providerrecords "github.com/yaleh/meta-cc/internal/provider/records"
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

// TestLoadTurnsFromRolloutAccumulatesTokenUsageAcrossCalls is a regression
// test for DIR-065: a tool-using turn makes multiple model API calls, and
// Codex emits one token_count event per call. applyTokenUsage previously
// overwrote b.current.TokenUsage on every event instead of accumulating, so
// only the final call's per-call usage survived — undercounting the turn's
// real consumption by every earlier call. The fixture's single turn carries
// two token_count events reporting last_token_usage {in:100,out:50} then
// {in:80,out:40}; the per-turn usage must be their sum {in:180,out:90}, not
// the last event's {in:80,out:40}. The fixture's total_token_usage on the
// final event already is the cumulative {in:180,out:90}, so this also
// verifies internal consistency: for this single-turn session, sum(per-turn
// usage) must equal the session total.
func TestLoadTurnsFromRolloutAccumulatesTokenUsageAcrossCalls(t *testing.T) {
	turns, total, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-multi-tokencount-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected one turn, got %#v", turns)
	}
	turn := turns[0]
	if turn.TokenUsage.InputTokens != 180 || turn.TokenUsage.OutputTokens != 90 {
		t.Fatalf("per-turn usage not accumulated across token_count events: got %#v, want {in:180,out:90}", turn.TokenUsage)
	}
	if total.InputTokens != 180 || total.OutputTokens != 90 {
		t.Fatalf("unexpected session total: %#v", total)
	}
	// Single-turn session: sum(per-turn usage) must reconcile with the
	// session total.
	if turn.TokenUsage.InputTokens != total.InputTokens || turn.TokenUsage.OutputTokens != total.OutputTokens {
		t.Fatalf("per-turn usage %#v does not reconcile with session total %#v", turn.TokenUsage, total)
	}
}

// TestLoadTurnsFromRolloutRetainsReasoningOutputTokens is the DIR-071
// regression test for reasoning tokens being dropped. Codex reports
// reasoning_output_tokens on every token_count event, but applyTokenUsage
// previously read it ONLY as part of the non-zero guard and then discarded it,
// so a turn's reasoning cost vanished from every query and reasoning-inclusive
// totals could not be reconciled. The fixture's single turn carries two
// token_count events reporting reasoning 10 then 5; the per-turn usage must
// accumulate them to 15 (DIR-065 accumulation still holding for every
// category), and the cumulative session total must retain reasoning too.
func TestLoadTurnsFromRolloutRetainsReasoningOutputTokens(t *testing.T) {
	turns, total, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-reasoning-tokens-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected one turn, got %#v", turns)
	}
	turn := turns[0]

	// Per-turn usage accumulates BOTH calls across every category, reasoning
	// included: input 100+80, cache 20+10, output 50+40, reasoning 10+5.
	if turn.TokenUsage.InputTokens != 180 || turn.TokenUsage.CacheTokens != 30 || turn.TokenUsage.OutputTokens != 90 {
		t.Fatalf("per-turn in/cache/out not accumulated: %#v", turn.TokenUsage)
	}
	if turn.TokenUsage.ReasoningOutputTokens != 15 {
		t.Fatalf("per-turn reasoning not accumulated across calls: got %d, want 15 (%#v)", turn.TokenUsage.ReasoningOutputTokens, turn.TokenUsage)
	}

	// The cumulative session total (last total_token_usage) retains reasoning.
	if total.InputTokens != 180 || total.CacheTokens != 30 || total.OutputTokens != 90 || total.ReasoningOutputTokens != 15 {
		t.Fatalf("session total dropped reasoning: %#v", total)
	}

	// Single-turn reconciliation: sum(per-turn) must equal the cumulative
	// total, reasoning included — the difference a caller would otherwise be
	// unable to explain is zero here by construction.
	if turn.TokenUsage != total {
		t.Fatalf("per-turn %#v does not reconcile with session total %#v (reasoning must be on both sides)", turn.TokenUsage, total)
	}
}

// TestLoadTurnsFromRolloutMultiTurnReconcilesWithSessionTotal is the DIR-071
// turn-vs-session reconciliation proof for a MULTI-turn session: the cumulative
// total_token_usage (session total) must equal the sum of every turn's
// per-turn usage across ALL categories, reasoning included. Because reasoning
// tokens are now retained on both sides (not dropped on the per-turn side),
// the two reconcile exactly — the reasoning component that previously made
// them disagree is present on both. This is the invariant a caller relies on
// to explain any session-total-vs-sum-of-turns difference: when it is non-zero,
// the cause is a real signal (compaction / missing events), not dropped
// reasoning.
func TestLoadTurnsFromRolloutMultiTurnReconcilesWithSessionTotal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "multi-turn-reconcile.jsonl")
	content := `{"timestamp":"2026-06-14T09:00:00Z","type":"session_meta","payload":{"id":"sess-mt","cwd":"/tmp/project","model":"gpt-5"}}
{"timestamp":"2026-06-14T09:00:01Z","type":"turn_context","payload":{"turn_id":"turn-1"}}
{"timestamp":"2026-06-14T09:00:02Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"first"}]}}
{"timestamp":"2026-06-14T09:00:03Z","type":"event_msg","payload":{"type":"token_count","info":{"last_token_usage":{"input_tokens":100,"output_tokens":50,"reasoning_output_tokens":10},"total_token_usage":{"input_tokens":100,"output_tokens":50,"reasoning_output_tokens":10}}}}
{"timestamp":"2026-06-14T09:00:04Z","type":"turn_context","payload":{"turn_id":"turn-2"}}
{"timestamp":"2026-06-14T09:00:05Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"second"}]}}
{"timestamp":"2026-06-14T09:00:06Z","type":"event_msg","payload":{"type":"token_count","info":{"last_token_usage":{"input_tokens":60,"output_tokens":30,"reasoning_output_tokens":5},"total_token_usage":{"input_tokens":160,"output_tokens":80,"reasoning_output_tokens":15}}}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, total, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d: %#v", len(turns), turns)
	}

	// Sum every turn's per-turn usage across all categories.
	var sumIn, sumOut, sumCache, sumReasoning int
	for _, turn := range turns {
		sumIn += turn.TokenUsage.InputTokens
		sumOut += turn.TokenUsage.OutputTokens
		sumCache += turn.TokenUsage.CacheTokens
		sumReasoning += turn.TokenUsage.ReasoningOutputTokens
	}

	// Reconciliation: sum(per-turn) == cumulative session total, reasoning
	// included on both sides.
	if sumIn != total.InputTokens || sumOut != total.OutputTokens || sumCache != total.CacheTokens || sumReasoning != total.ReasoningOutputTokens {
		t.Fatalf("per-turn sums {in:%d out:%d cache:%d reasoning:%d} do not reconcile with session total %#v",
			sumIn, sumOut, sumCache, sumReasoning, total)
	}
	if total.ReasoningOutputTokens != 15 || sumReasoning != 15 {
		t.Fatalf("reasoning must reconcile to 15 on both sides: sum=%d total=%#v", sumReasoning, total)
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

// TestLoadTurnsFromRollout0145EventFamilies exercises the Codex 0.145 event
// families absent from earlier fixtures: world_state, compacted,
// tool_search_call, tool_search_output, thread_settings_applied,
// context_compacted, turn_aborted, and session_end. DIR-032 promoted four of
// these (compacted, context_compacted, turn_aborted, session_end) to typed
// handling (see TestLoadTurnsFromRollout0145TypedEventFamilies below), and
// DIR-072 promoted the remaining four (world_state, tool_search_call,
// tool_search_output, thread_settings_applied — see
// TestLoadTurnsFromRollout0145TypedSearchFamilies), so the fixture now
// produces typed items for ALL eight families with no raw passthrough left.
// All eight must never crash parsing, never silently disappear, and never
// interfere with the user/assistant text or tool calls that share their
// turn.
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

	// DIR-072: every 0.145 family in the fixture is now typed, so no raw
	// passthrough may remain in Extensions.codex_events (DIR-032 typed
	// compacted/context_compacted/turn_aborted/session_end; DIR-072 typed
	// world_state/tool_search_call/tool_search_output/
	// thread_settings_applied).
	if len(turn.Extensions) != 0 {
		t.Fatalf("expected no raw passthrough events to remain, got extensions: %s", turn.Extensions)
	}
}

// TestLoadTurnsFromRollout0145TypedEventFamilies is the DIR-032 "typed
// status/boundary rather than opaque unknown events" proof for the four
// event families promoted out of raw passthrough: the top-level "compacted"
// event and the event_msg "context_compacted" notification both become
// typed ItemKindCompaction items carrying CompactionBoundary metadata
// (never folded into UserText/AssistantText), "turn_aborted" sets
// TurnStatusAborted, and "session_end" becomes a typed ItemKindSessionEnd
// item.
func TestLoadTurnsFromRollout0145TypedEventFamilies(t *testing.T) {
	turns, _, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-0145-families-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("0145 families load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(turns))
	}
	turn := turns[0]

	if turn.Status != conversation.TurnStatusAborted {
		t.Fatalf("expected turn_aborted to set TurnStatusAborted, got %q", turn.Status)
	}

	var compactions []conversation.Item
	var sessionEnds []conversation.Item
	for _, item := range turn.Items {
		switch item.Kind {
		case conversation.ItemKindCompaction:
			compactions = append(compactions, item)
		case conversation.ItemKindSessionEnd:
			sessionEnds = append(sessionEnds, item)
		}
	}

	if len(compactions) != 2 {
		t.Fatalf("expected 2 typed compaction items (compacted + context_compacted), got %d: %#v", len(compactions), compactions)
	}
	var sawReason, sawSummary bool
	for _, c := range compactions {
		if c.Compaction == nil {
			t.Fatalf("compaction item missing CompactionBoundary: %#v", c)
		}
		if c.Compaction.Reason == "context_window" {
			sawReason = true
		}
		if c.Compaction.Summary == "Trimmed 4 earlier turns" {
			sawSummary = true
		}
	}
	if !sawReason || !sawSummary {
		t.Fatalf("expected one compaction item with Reason and one with Summary, got %#v", compactions)
	}

	// The compaction summary must never leak into the message projection —
	// otherwise a content query would present replaced/summarized text as
	// if it were an ordinary (and potentially duplicate) assistant message.
	if strings.Contains(turn.AssistantText, "Trimmed 4 earlier turns") {
		t.Fatalf("compaction summary must not be folded into AssistantText, got %q", turn.AssistantText)
	}

	if len(sessionEnds) != 1 || sessionEnds[0].Text != "completed" {
		t.Fatalf("expected 1 typed session_end item with reason %q, got %#v", "completed", sessionEnds)
	}
}

// TestLoadTurnsFromRolloutCompactionDoesNotDuplicateContent is the DIR-032
// compaction-dedup acceptance test (Contract: "Compaction preserves visible
// replacement history and boundary metadata without duplicating superseded
// content"). The fixture has a pre-compaction user/assistant exchange, a
// compaction boundary (event_msg context_compacted + top-level compacted),
// and a post-compaction user/assistant exchange, all within the same turn.
// Both exchanges must survive intact and exactly once each in the
// UserText/AssistantText content projection (the same fields
// query_session_content ultimately reads), and the compaction boundary's
// own summary text must never appear inside that projection — proving a
// content query cannot end up presenting the replaced text and its
// boundary/replacement as if they were both live, duplicate message
// content.
func TestLoadTurnsFromRolloutCompactionDoesNotDuplicateContent(t *testing.T) {
	turns, _, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-compaction-boundary-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("compaction boundary load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d: %#v", len(turns), turns)
	}
	turn := turns[0]

	wantUser := "What does main.go do?\nNow add a flag."
	if turn.UserText != wantUser {
		t.Fatalf("UserText = %q, want %q (pre- and post-compaction text joined exactly once each)", turn.UserText, wantUser)
	}
	wantAssistant := "It parses CLI args.\nAdded --verbose flag."
	if turn.AssistantText != wantAssistant {
		t.Fatalf("AssistantText = %q, want %q", turn.AssistantText, wantAssistant)
	}

	if strings.Count(turn.UserText, "What does main.go do?") != 1 {
		t.Fatalf("pre-compaction user text duplicated: %q", turn.UserText)
	}
	if strings.Contains(turn.UserText, "Summarized 1 earlier exchange") || strings.Contains(turn.AssistantText, "Summarized 1 earlier exchange") {
		t.Fatalf("compaction summary leaked into message projection: user=%q assistant=%q", turn.UserText, turn.AssistantText)
	}

	// The Item stream itself must retain both pre- and post-compaction
	// message items (the historical record) plus exactly one compaction
	// boundary item in between — nothing collapsed, nothing duplicated.
	var kinds []conversation.ItemKind
	compactionIdx := -1
	for i, item := range turn.Items {
		kinds = append(kinds, item.Kind)
		if item.Kind == conversation.ItemKindCompaction && item.Compaction != nil && item.Compaction.Summary != "" {
			compactionIdx = i
		}
	}
	if compactionIdx == -1 {
		t.Fatalf("expected a compaction item carrying the summary boundary, got items: %#v", kinds)
	}
	var preUser, postUser int
	for i, item := range turn.Items {
		if item.Kind == conversation.ItemKindUserMessage {
			if i < compactionIdx {
				preUser++
			} else {
				postUser++
			}
		}
	}
	if preUser != 1 || postUser != 1 {
		t.Fatalf("expected exactly one pre- and one post-compaction user message item, got pre=%d post=%d (items: %#v)", preUser, postUser, kinds)
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

// TestLoadTurnsFromRolloutPreservesItemOrderAndPhase is a DIR-028
// regression/acceptance test: a Codex turn with a commentary message, an
// interleaved tool call/result, and a final message must preserve the
// exact encounter order and each assistant message's phase (derived from
// the response_item "channel" field) in the canonical Items stream — not
// just a flattened, phase-less AssistantText string.
func TestLoadTurnsFromRolloutPreservesItemOrderAndPhase(t *testing.T) {
	turns, _, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-phased-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("phased load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d: %#v", len(turns), turns)
	}
	turn := turns[0]

	wantKinds := []conversation.ItemKind{
		conversation.ItemKindUserMessage,
		conversation.ItemKindAgentMessage,
		conversation.ItemKindToolCall,
		conversation.ItemKindToolResult,
		conversation.ItemKindAgentMessage,
	}
	if len(turn.Items) != len(wantKinds) {
		t.Fatalf("expected %d items in encounter order, got %d: %#v", len(wantKinds), len(turn.Items), turn.Items)
	}
	for i, kind := range wantKinds {
		if turn.Items[i].Kind != kind {
			t.Fatalf("item %d: expected kind %s, got %s (%#v)", i, kind, turn.Items[i].Kind, turn.Items[i])
		}
	}

	if turn.Items[1].Phase != conversation.PhaseCommentary || turn.Items[1].Text != "Investigating the failure..." {
		t.Fatalf("expected commentary phase item, got %#v", turn.Items[1])
	}
	if turn.Items[4].Phase != conversation.PhaseFinal || turn.Items[4].Text != "Fixed the bug." {
		t.Fatalf("expected final phase item, got %#v", turn.Items[4])
	}

	// Tool call/result retain linked identity via ToolCallID, and the turn
	// carries a stable ID.
	if turn.ID != "turn-phased" {
		t.Fatalf("expected stable turn ID, got %q", turn.ID)
	}
	if turn.Items[2].ID != "call-phase-1" || turn.Items[2].ToolCallID != "call-phase-1" {
		t.Fatalf("tool call item missing stable ID: %#v", turn.Items[2])
	}
	if turn.Items[3].ToolCallID != "call-phase-1" || turn.Items[3].Output != "found 2 matches" {
		t.Fatalf("tool result item not paired with call via ToolCallID: %#v", turn.Items[3])
	}

	// The legacy compatibility projection must still be correct: both
	// commentary and final text joined, and exactly one deduped tool call.
	if turn.AssistantText != "Investigating the failure...\nFixed the bug." {
		t.Fatalf("unexpected AssistantText projection: %q", turn.AssistantText)
	}
	if len(turn.ToolCalls) != 1 || turn.ToolCalls[0].Output != "found 2 matches" {
		t.Fatalf("unexpected ToolCalls projection: %#v", turn.ToolCalls)
	}
}

// TestLoadTurnsFromRollout0145EventFamiliesProducesTypedItems extends the
// 0145-event-families coverage (see TestLoadTurnsFromRollout0145EventFamilies
// above) to the item level: DIR-072's four promoted families must each
// produce exactly one typed canonical Item (never a generic ItemKindUnknown),
// and NO event in the fixture may remain as an unknown raw passthrough —
// the item stream classifies every 0.145 family the fixture carries.
// (Unknown items themselves remain fully supported for genuinely
// unrecognized events — see
// TestLoadTurnsFromRolloutUnknownFutureEventsRoundTripAsRaw.)
func TestLoadTurnsFromRollout0145EventFamiliesProducesTypedItems(t *testing.T) {
	turns, _, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-0145-families-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("0145 families load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(turns))
	}

	kinds := map[conversation.ItemKind]int{}
	for _, item := range turns[0].Items {
		kinds[item.Kind]++
		if item.Kind == conversation.ItemKindUnknown {
			t.Fatalf("no 0145 fixture event may remain an unknown raw passthrough: %#v", item)
		}
		if len(item.Raw) != 0 {
			t.Fatalf("typed items must not embed raw provenance: %#v", item)
		}
	}
	for _, kind := range []conversation.ItemKind{
		conversation.ItemKindWorldState,
		conversation.ItemKindToolSearchCall,
		conversation.ItemKindToolSearchOutput,
		conversation.ItemKindSettingsApplied,
	} {
		if kinds[kind] != 1 {
			t.Fatalf("expected exactly one typed %s item, got %d (kinds: %#v)", kind, kinds[kind], kinds)
		}
	}
}

// TestLoadTurnsFromRollout0145TypedSearchFamilies is the DIR-072 typed-item
// proof for the four promoted families, asserting each item's stable kind,
// source, timestamp, identity, and structured fields: world_state carries
// bounded WorldState metadata; tool_search_call/output correlate via
// ToolCallID; thread_settings_applied exposes setting KEYS only (never
// values). The search pair must never leak into the legacy ToolCalls
// projection (searches are not shell/function tools), and stream order is
// preserved (call before output).
func TestLoadTurnsFromRollout0145TypedSearchFamilies(t *testing.T) {
	turns, _, err := loadTurnsFromRollout(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-legacy-0145-families-sample.jsonl"), 100)
	if err != nil {
		t.Fatalf("0145 families load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(turns))
	}
	turn := turns[0]

	byKind := map[conversation.ItemKind][]int{}
	for i, item := range turn.Items {
		byKind[item.Kind] = append(byKind[item.Kind], i)
	}

	wsIdx := byKind[conversation.ItemKindWorldState]
	if len(wsIdx) != 1 {
		t.Fatalf("expected 1 world_state item, got %d", len(wsIdx))
	}
	ws := turn.Items[wsIdx[0]]
	if ws.WorldState == nil || ws.WorldState.CWD != "/tmp/project" || ws.WorldState.ApprovalPolicy != "on-request" {
		t.Fatalf("world_state metadata not typed/bounded: %#v", ws)
	}
	if ws.Source != "world_state" {
		t.Fatalf("world_state item must record its provenance source: %#v", ws)
	}
	if want := time.Date(2026, 7, 20, 10, 0, 2, 0, time.UTC); !ws.Timestamp.Equal(want) {
		t.Fatalf("world_state timestamp = %v, want %v", ws.Timestamp, want)
	}

	callIdx := byKind[conversation.ItemKindToolSearchCall]
	outIdx := byKind[conversation.ItemKindToolSearchOutput]
	if len(callIdx) != 1 || len(outIdx) != 1 {
		t.Fatalf("expected 1 tool_search_call and 1 tool_search_output item, got %#v", byKind)
	}
	call, output := turn.Items[callIdx[0]], turn.Items[outIdx[0]]
	if call.ID != "call-search-1" || call.ToolCallID != "call-search-1" || call.ToolName != "repo_search" {
		t.Fatalf("tool_search_call identity not typed: %#v", call)
	}
	if string(call.Input) != `{"query":"TODO"}` {
		t.Fatalf("tool_search_call input not retained: %s", call.Input)
	}
	if output.ToolCallID != call.ToolCallID {
		t.Fatalf("tool_search_output must correlate with its call via ToolCallID: call=%#v output=%#v", call, output)
	}
	if output.Output != "3 matches found" || output.IsError {
		t.Fatalf("tool_search_output fields wrong: %#v", output)
	}
	if callIdx[0] >= outIdx[0] {
		t.Fatalf("stream order not preserved: call at %d, output at %d", callIdx[0], outIdx[0])
	}
	// The search pair must never be confused with shell/function tools: the
	// legacy ToolCalls projection (what tool stats and normalized tool_use
	// records consume) must stay empty for this turn.
	if len(turn.ToolCalls) != 0 {
		t.Fatalf("tool searches must not leak into the legacy ToolCalls projection: %#v", turn.ToolCalls)
	}

	setIdx := byKind[conversation.ItemKindSettingsApplied]
	if len(setIdx) != 1 {
		t.Fatalf("expected 1 settings_applied item, got %d", len(setIdx))
	}
	settings := turn.Items[setIdx[0]]
	if settings.Settings == nil || settings.Source != "event_msg" {
		t.Fatalf("settings item not typed with provenance: %#v", settings)
	}
	wantKeys := []string{"approval_policy", "model"}
	if len(settings.Settings.Keys) != 2 || settings.Settings.Keys[0] != wantKeys[0] || settings.Settings.Keys[1] != wantKeys[1] {
		t.Fatalf("settings keys must be sorted and value-free: %#v", settings.Settings)
	}
}

// TestLoadTurnsFromRolloutToolSearchSurfacesFailures is the DIR-072
// correlation/failure proof: two search calls, one succeeding (status
// completed) and one failing (status failed + error string), must each pair
// with their output via ToolCallID, the failure must surface through
// IsError/Status/Output (error text standing in for absent output), and the
// search pair must never be confused with the shell/function tool call that
// shares the turn — the legacy ToolCalls projection contains ONLY the
// function call.
func TestLoadTurnsFromRolloutToolSearchSurfacesFailures(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tool-search-failures.jsonl")
	content := `{"timestamp":"2026-07-21T08:00:00Z","type":"session_meta","payload":{"id":"sess-sf","cwd":"/tmp/project"}}
{"timestamp":"2026-07-21T08:00:01Z","type":"turn_context","payload":{"turn_id":"turn-sf"}}
{"timestamp":"2026-07-21T08:00:02Z","type":"response_item","payload":{"type":"tool_search_call","call_id":"call-ok","name":"repo_search","arguments":"{\"query\":\"a\"}"}}
{"timestamp":"2026-07-21T08:00:03Z","type":"response_item","payload":{"type":"tool_search_output","call_id":"call-ok","output":"1 match","status":"completed"}}
{"timestamp":"2026-07-21T08:00:04Z","type":"response_item","payload":{"type":"tool_search_call","call_id":"call-bad","name":"web_search","arguments":"{\"query\":\"b\"}"}}
{"timestamp":"2026-07-21T08:00:05Z","type":"response_item","payload":{"type":"tool_search_output","call_id":"call-bad","status":"failed","error":"search backend unavailable"}}
{"timestamp":"2026-07-21T08:00:06Z","type":"response_item","payload":{"type":"function_call","name":"shell","call_id":"call-sh","arguments":"{\"cmd\":\"ls\"}"}}
{"timestamp":"2026-07-21T08:00:07Z","type":"response_item","payload":{"type":"function_call_output","call_id":"call-sh","output":"file.txt"}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, _, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(turns))
	}
	turn := turns[0]

	var calls, outputs []conversation.Item
	for _, item := range turn.Items {
		switch item.Kind {
		case conversation.ItemKindToolSearchCall:
			calls = append(calls, item)
		case conversation.ItemKindToolSearchOutput:
			outputs = append(outputs, item)
		}
	}
	if len(calls) != 2 || len(outputs) != 2 {
		t.Fatalf("expected 2 search calls and 2 search outputs, got %d/%d", len(calls), len(outputs))
	}
	byID := map[string]conversation.Item{}
	for _, output := range outputs {
		byID[output.ToolCallID] = output
	}
	ok := byID["call-ok"]
	if ok.Output != "1 match" || ok.IsError || ok.Status != conversation.StatusCompleted {
		t.Fatalf("successful search output wrong: %#v", ok)
	}
	bad := byID["call-bad"]
	if !bad.IsError || bad.Status != conversation.StatusFailed || bad.Output != "search backend unavailable" {
		t.Fatalf("failed search output must surface error via IsError/Status/Output: %#v", bad)
	}

	// Searches are not shell/function tools: the legacy projection must
	// carry ONLY the function call (no duplication, no conflation).
	if len(turn.ToolCalls) != 1 || turn.ToolCalls[0].Name != "shell" || turn.ToolCalls[0].Output != "file.txt" {
		t.Fatalf("legacy ToolCalls must contain only the function call: %#v", turn.ToolCalls)
	}
}

// TestLoadTurnsFromRolloutWorldStateAndSettingsBounded is the DIR-072
// size/privacy proof: an oversized world_state value is capped (never
// embedded in full), unknown/huge world_state fields and sensitive settings
// VALUES never appear anywhere in the marshaled turn, while settings KEYS
// remain available for filtering. The full payloads survive only through
// explicit raw access to the source rollout file.
func TestLoadTurnsFromRolloutWorldStateAndSettingsBounded(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bounded-payloads.jsonl")
	bigCWD := strings.Repeat("x", 5000)
	content := `{"timestamp":"2026-07-22T09:00:00Z","type":"session_meta","payload":{"id":"sess-bd","cwd":"/tmp/project"}}
{"timestamp":"2026-07-22T09:00:01Z","type":"turn_context","payload":{"turn_id":"turn-bd"}}
{"timestamp":"2026-07-22T09:00:02Z","type":"world_state","payload":{"cwd":"` + bigCWD + `","approval_policy":"never","env":{"TOKEN":"huge-environment-secret"}}}
{"timestamp":"2026-07-22T09:00:03Z","type":"event_msg","payload":{"type":"thread_settings_applied","settings":{"model":"\"gpt-5\"","api_key":"\"sk-live-abc123secret\""}}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, _, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(turns))
	}
	turn := turns[0]

	var ws, settings *conversation.Item
	for i := range turn.Items {
		switch turn.Items[i].Kind {
		case conversation.ItemKindWorldState:
			ws = &turn.Items[i]
		case conversation.ItemKindSettingsApplied:
			settings = &turn.Items[i]
		}
	}
	if ws == nil || ws.WorldState == nil {
		t.Fatalf("missing typed world_state item: %#v", turn.Items)
	}
	// conversation.maxWorldStateValueLen == 512: the oversized cwd is
	// capped, the short approval_policy survives.
	if len(ws.WorldState.CWD) != 512 {
		t.Fatalf("world_state cwd must be capped at 512 bytes, got %d", len(ws.WorldState.CWD))
	}
	if ws.WorldState.ApprovalPolicy != "never" {
		t.Fatalf("ordinary world_state field lost: %#v", ws.WorldState)
	}
	if settings == nil || settings.Settings == nil {
		t.Fatalf("missing typed settings item: %#v", turn.Items)
	}
	if len(settings.Settings.Keys) != 2 || settings.Settings.Keys[0] != "api_key" || settings.Settings.Keys[1] != "model" {
		t.Fatalf("settings keys must be retained (sorted): %#v", settings.Settings)
	}

	// Nothing sensitive or oversized may leak into the canonical turn:
	// marshal it the way downstream consumers see it and assert the secret
	// values, the unbounded cwd bytes, and the unknown env payload are all
	// absent.
	data, err := json.Marshal(turn)
	if err != nil {
		t.Fatalf("marshal turn: %v", err)
	}
	for _, leaked := range []string{"sk-live-abc123secret", "huge-environment-secret", bigCWD} {
		if strings.Contains(string(data), leaked) {
			t.Fatalf("sensitive/oversized payload leaked into the typed turn: %q", leaked)
		}
	}
}

// TestLoadTurnsFromRolloutUnknownFutureEventsRoundTripAsRaw is the DIR-072
// forward-compatibility proof: events from families the parser does NOT
// recognize (simulated future top-level, event_msg, and response_item
// shapes) must still survive as capped ItemKindUnknown items with raw
// provenance AND in the legacy Extensions.codex_events bag — promoting the
// four known families must not regress the unknown passthrough. A known
// typed family interleaved with them still classifies correctly.
func TestLoadTurnsFromRolloutUnknownFutureEventsRoundTripAsRaw(t *testing.T) {
	path := filepath.Join(t.TempDir(), "future-events.jsonl")
	content := `{"timestamp":"2027-01-01T00:00:00Z","type":"session_meta","payload":{"id":"sess-fut","cwd":"/tmp/project"}}
{"timestamp":"2027-01-01T00:00:01Z","type":"turn_context","payload":{"turn_id":"turn-fut"}}
{"timestamp":"2027-01-01T00:00:02Z","type":"quantum_state","payload":{"superposition":true}}
{"timestamp":"2027-01-01T00:00:03Z","type":"event_msg","payload":{"type":"future_notification","note":"not yet modeled"}}
{"timestamp":"2027-01-01T00:00:04Z","type":"response_item","payload":{"type":"hologram_call","call_id":"call-h"}}
{"timestamp":"2027-01-01T00:00:05Z","type":"world_state","payload":{"cwd":"/tmp/project","approval_policy":"never"}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, _, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(turns))
	}
	turn := turns[0]

	var unknown, typed int
	for _, item := range turn.Items {
		if item.Kind == conversation.ItemKindUnknown {
			unknown++
			if len(item.Raw) == 0 {
				t.Fatalf("unknown item must retain raw provenance: %#v", item)
			}
			if item.RawTruncated {
				t.Fatalf("small future events must not be truncated: %#v", item)
			}
			continue
		}
		if item.Kind == conversation.ItemKindWorldState {
			typed++
		}
	}
	if unknown != 3 {
		t.Fatalf("expected 3 unknown passthrough items for the future families, got %d", unknown)
	}
	if typed != 1 {
		t.Fatalf("known family interleaved with unknown events must still classify, got %d typed world_state items", typed)
	}

	var ext struct {
		CodexEvents []json.RawMessage `json:"codex_events"`
	}
	if err := json.Unmarshal(turn.Extensions, &ext); err != nil {
		t.Fatalf("unmarshal extensions: %v", err)
	}
	if len(ext.CodexEvents) != 3 {
		t.Fatalf("future events must also round-trip via Extensions.codex_events, got %d", len(ext.CodexEvents))
	}
}

// TestLoadTurnsFromRolloutDotSchemaPreservesMultipleAssistantMessages is a
// DIR-028 regression test for a real gap in the pre-existing dot-schema
// (item.message) handling: it used to directly overwrite
// Turn.AssistantText on every item.message event, so a turn with more than
// one assistant message silently lost every message but the last. Items
// must now preserve all of them, in order, and the legacy projection must
// join them (matching the legacy-schema join behavior) instead of losing
// data.
func TestLoadTurnsFromRolloutDotSchemaPreservesMultipleAssistantMessages(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dot-schema-multi.jsonl")
	content := `{"timestamp":"2026-07-27T09:00:00Z","type":"thread.started","payload":{"id":"sess-multi"}}
{"timestamp":"2026-07-27T09:00:01Z","type":"turn.started","payload":{"id":"turn-multi"}}
{"timestamp":"2026-07-27T09:00:02Z","type":"item.message","payload":{"role":"user","content":"do the thing"}}
{"timestamp":"2026-07-27T09:00:03Z","type":"item.message","payload":{"role":"assistant","content":"first commentary"}}
{"timestamp":"2026-07-27T09:00:04Z","type":"item.message","payload":{"role":"assistant","content":"second and final"}}
{"timestamp":"2026-07-27T09:00:05Z","type":"turn.completed","payload":{"id":"turn-multi"}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, _, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(turns))
	}
	turn := turns[0]
	if turn.Status != conversation.TurnStatusCompleted {
		t.Fatalf("expected completed status, got %q", turn.Status)
	}
	if turn.AssistantText != "first commentary\nsecond and final" {
		t.Fatalf("expected both assistant messages preserved in order, got %q", turn.AssistantText)
	}

	var assistantItems int
	for _, item := range turn.Items {
		if item.Kind == conversation.ItemKindAgentMessage {
			assistantItems++
		}
	}
	if assistantItems != 2 {
		t.Fatalf("expected 2 distinct assistant message items, got %d: %#v", assistantItems, turn.Items)
	}
}

// TestLoadTurnsFromRolloutLifecycleOnlySessionEnd is the DIR-050 regression
// test: a turn opened by task_started whose ONLY content before EOF is a
// bare session_end event (no user/assistant message, no tool call, no token
// usage) must still be retained as a Turn carrying the typed
// ItemKindSessionEnd item -- flush()'s retention condition must not discard
// a turn just because none of UserText/AssistantText/ToolCalls/TokenUsage/
// Extensions ended up populated, when the turn's Items stream is non-empty.
func TestLoadTurnsFromRolloutLifecycleOnlySessionEnd(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session-end-only.jsonl")
	content := `{"timestamp":"2026-06-14T06:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"t1"}}
{"timestamp":"2026-06-14T06:00:05Z","type":"session_end","payload":{"reason":"completed"}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, _, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn (not discarded), got %d: %#v", len(turns), turns)
	}
	turn := turns[0]

	var sessionEnds []conversation.Item
	for _, item := range turn.Items {
		if item.Kind == conversation.ItemKindSessionEnd {
			sessionEnds = append(sessionEnds, item)
		}
	}
	if len(sessionEnds) != 1 || sessionEnds[0].Text != "completed" {
		t.Fatalf("expected 1 typed session_end item with reason %q, got %#v", "completed", sessionEnds)
	}
}

// TestLoadTurnsFromRolloutLifecycleOnlyTurnAborted is the DIR-050 regression
// test for the turn_aborted analog: a turn opened by task_started then
// immediately aborted (e.g. the user hits Ctrl-C right after the model
// starts thinking), with no message/tool activity at all, must still be
// retained as a Turn with Status == TurnStatusAborted rather than silently
// discarded.
func TestLoadTurnsFromRolloutLifecycleOnlyTurnAborted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "turn-aborted-only.jsonl")
	content := `{"timestamp":"2026-06-14T06:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"t1"}}
{"timestamp":"2026-06-14T06:00:02Z","type":"turn_aborted","payload":{"turn_id":"t1"}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, _, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn (not discarded), got %d: %#v", len(turns), turns)
	}
	if turns[0].Status != conversation.TurnStatusAborted {
		t.Fatalf("expected TurnStatusAborted, got %q", turns[0].Status)
	}
}

// TestLoadTurnsFromRolloutLifecycleOnlyCompaction is the DIR-050 regression
// test for the compaction analog: a turn whose only content is a
// compacted/context_compacted event (no other message/tool/usage content in
// that turn) must still be retained as a Turn carrying the typed
// ItemKindCompaction item with its CompactionBoundary, not discarded.
func TestLoadTurnsFromRolloutLifecycleOnlyCompaction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "compaction-only.jsonl")
	content := `{"timestamp":"2026-06-14T06:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"t1"}}
{"timestamp":"2026-06-14T06:00:03Z","type":"compacted","payload":{"turn_id":"t1","reason":"context_window"}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, _, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn (not discarded), got %d: %#v", len(turns), turns)
	}
	turn := turns[0]

	var compactions []conversation.Item
	for _, item := range turn.Items {
		if item.Kind == conversation.ItemKindCompaction {
			compactions = append(compactions, item)
		}
	}
	if len(compactions) != 1 || compactions[0].Compaction == nil || compactions[0].Compaction.Reason != "context_window" {
		t.Fatalf("expected 1 typed compaction item with Reason %q, got %#v", "context_window", compactions)
	}
}

// TestLoadTurnsFromRolloutDotSchemaToolResultFirstLine is the DIR-061
// regression test for a dot-schema rollout whose FIRST record is an
// item.tool_result event (head-truncated/rotated rollout file). No
// turn.started precedes it, so b.current is nil when the item.tool_result
// case runs; the parser must ensureTurn rather than nil-deref panic, and
// must yield a turn containing that tool-result item.
func TestLoadTurnsFromRolloutDotSchemaToolResultFirstLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dot-schema-tool-result-first.jsonl")
	content := `{"timestamp":"2026-07-27T10:00:00Z","type":"item.tool_result","payload":{"id":"call-1","output":"late result","is_error":false}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, _, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn (not discarded, no panic), got %d: %#v", len(turns), turns)
	}

	var results []conversation.Item
	for _, item := range turns[0].Items {
		if item.Kind == conversation.ItemKindToolResult {
			results = append(results, item)
		}
	}
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 tool-result item, got %d: %#v", len(results), turns[0].Items)
	}
	if results[0].ToolCallID != "call-1" || results[0].Output != "late result" || results[0].IsError {
		t.Fatalf("unexpected tool-result item: %#v", results[0])
	}
}

// TestLoadTurnsFromRolloutDotSchemaToolResultAfterTurnCompleted is the
// DIR-061 regression test for a stray item.tool_result that arrives AFTER a
// turn.completed (which flushes -> b.current = nil), e.g. a late/duplicate
// notification line or a resumed stream. The parser must not panic; the
// stray result is folded into a fresh turn rather than crashing the request.
func TestLoadTurnsFromRolloutDotSchemaToolResultAfterTurnCompleted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dot-schema-tool-result-after-completed.jsonl")
	content := `{"timestamp":"2026-07-27T10:00:00Z","type":"thread.started","payload":{"id":"sess-stray"}}
{"timestamp":"2026-07-27T10:00:01Z","type":"turn.started","payload":{"id":"turn-1"}}
{"timestamp":"2026-07-27T10:00:02Z","type":"item.message","payload":{"role":"user","content":"do the thing"}}
{"timestamp":"2026-07-27T10:00:03Z","type":"turn.completed","payload":{"id":"turn-1"}}
{"timestamp":"2026-07-27T10:00:04Z","type":"item.tool_result","payload":{"id":"call-9","output":"stray result","is_error":true}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, _, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns (completed turn + stray-result turn), got %d: %#v", len(turns), turns)
	}
	if turns[0].Status != conversation.TurnStatusCompleted {
		t.Fatalf("expected first turn completed, got %q", turns[0].Status)
	}

	var results []conversation.Item
	for _, item := range turns[1].Items {
		if item.Kind == conversation.ItemKindToolResult {
			results = append(results, item)
		}
	}
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 tool-result item in the stray turn, got %d: %#v", len(results), turns[1].Items)
	}
	if results[0].ToolCallID != "call-9" || results[0].Output != "stray result" || !results[0].IsError {
		t.Fatalf("unexpected stray tool-result item: %#v", results[0])
	}
}

// TestLoadTurnsFromRolloutUnknownFirstEventUsesEventTimestamp is the DIR-064
// regression test. When a rollout's FIRST event is an unrecognized dot-schema
// event (the exact schema-drift case appendUnknown exists to tolerate), the
// turn it opens must be stamped with the EVENT's own timestamp, not the
// query-time wall clock. Before the fix, appendUnknown called
// ensureTurn("", time.Now()...) so a 2025 session queried in any later year
// carried today's date; because ensureTurn is a no-op once a turn is open,
// every subsequent event in that turn inherited the bogus now(), and
// Normalize emitted it verbatim as the record timestamp — skewing since/until
// filtering and recency sorting.
//
// The assertion is run-time-independent: it pins the year to 2025 (the event's
// year) rather than comparing against time.Now(), so it fails on the old code
// regardless of when the suite runs and stays green after the fix.
func TestLoadTurnsFromRolloutUnknownFirstEventUsesEventTimestamp(t *testing.T) {
	const eventTS = "2025-03-01T00:00:00Z"
	path := filepath.Join(t.TempDir(), "unknown-first-event.jsonl")
	// First line is an unrecognized dot-schema event (contains "." so
	// detectSchemaVersion routes to the new-schema dispatcher, whose default
	// case calls appendUnknown). It carries a 2025 timestamp. The following
	// item.message opens no new turn (ensureTurn is a no-op once one exists),
	// so the whole turn is governed by the timestamp appendUnknown set.
	content := `{"timestamp":"` + eventTS + `","type":"mystery.unknown_event","payload":{"note":"schema drift"}}
{"timestamp":"2025-03-01T00:00:01Z","type":"item.message","payload":{"role":"user","content":"hello"}}
{"timestamp":"2025-03-01T00:00:02Z","type":"turn.completed","payload":{"id":"t-unknown"}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	turns, _, err := loadTurnsFromRollout(path, 100)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d: %#v", len(turns), turns)
	}
	turn := turns[0]

	// AC: the turn timestamp must be the event's 2025 timestamp, not the
	// current date. Run-time-independent: pin the year to 2025.
	wantTS := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	if !turn.Timestamp.Equal(wantTS) {
		t.Fatalf("turn timestamp = %v (year %d), want event timestamp %v (year 2025); appendUnknown likely used time.Now()",
			turn.Timestamp, turn.Timestamp.Year(), wantTS)
	}
	if turn.Timestamp.Year() != 2025 {
		t.Fatalf("turn timestamp year = %d, want 2025 (run-time-independent)", turn.Timestamp.Year())
	}

	// AC: the unknown Item and its enclosing Turn carry the same event-derived
	// timestamp when the event has one.
	var unknownItems []conversation.Item
	for _, item := range turn.Items {
		if item.Kind == conversation.ItemKindUnknown {
			unknownItems = append(unknownItems, item)
		}
	}
	if len(unknownItems) != 1 {
		t.Fatalf("expected exactly 1 unknown item, got %d: %#v", len(unknownItems), turn.Items)
	}
	if !unknownItems[0].Timestamp.Equal(turn.Timestamp) {
		t.Fatalf("unknown item timestamp = %v, want equal to turn timestamp %v",
			unknownItems[0].Timestamp, turn.Timestamp)
	}

	// AC: the emitted record timestamp equals the event timestamp. Normalize
	// copies turn.Timestamp verbatim into each record's "timestamp" field,
	// which is what since/until filtering and recency sorts consume.
	records := providerrecords.Normalize(conversation.Session{ID: "sess-unknown", Provider: conversation.ProviderCodex}, turns)
	if len(records) == 0 {
		t.Fatalf("expected Normalize to emit at least one record for the turn")
	}
	for _, rec := range records {
		if got, _ := rec["timestamp"].(string); got != eventTS {
			t.Fatalf("record timestamp = %q, want event timestamp %q (run-time-independent)", got, eventTS)
		}
	}
}
