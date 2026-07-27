package conversation

import (
	"encoding/json"
	"time"
)

// ItemKind identifies the semantic shape of an ordered Item within a Turn.
// This is the extension point for DIR-028's loss-minimizing model: provider
// adapters emit one Item per event they can classify, and anything they
// cannot classify yet is preserved as ItemKindUnknown rather than dropped.
type ItemKind string

const (
	// ItemKindUserMessage is a user-authored message segment.
	ItemKindUserMessage ItemKind = "user_message"
	// ItemKindAgentMessage is an assistant/agent-authored message segment.
	// Use Phase to distinguish intermediate commentary from the final
	// answer when the provider reports it.
	ItemKindAgentMessage ItemKind = "agent_message"
	// ItemKindToolCall is a request to invoke a tool (function call, MCP
	// call, etc.). Paired with a later ItemKindToolResult via ToolCallID.
	ItemKindToolCall ItemKind = "tool_call"
	// ItemKindToolResult is the output of a previously issued tool call.
	ItemKindToolResult ItemKind = "tool_result"
	// ItemKindCommandExecution is a shell/command execution distinct from a
	// generic tool call (e.g. Codex's exec sandbox events), when a provider
	// chooses to model it separately from ItemKindToolCall.
	ItemKindCommandExecution ItemKind = "command_execution"
	// ItemKindFileChange is a file create/edit/delete event.
	ItemKindFileChange ItemKind = "file_change"
	// ItemKindWebSearch is a web/tool search query and its result summary.
	ItemKindWebSearch ItemKind = "web_search"
	// ItemKindPlanUpdate is a structured plan/todo-list update.
	ItemKindPlanUpdate ItemKind = "plan_update"
	// ItemKindReasoning is model reasoning/thinking metadata (not shown as
	// a regular message).
	ItemKindReasoning ItemKind = "reasoning"
	// ItemKindCompaction marks a context-compaction event (history summarized
	// or trimmed). See CompactionBoundary (populated on the Compaction field)
	// for typed boundary metadata: what replaced the preceding context and
	// why, without re-embedding the superseded content itself.
	ItemKindCompaction ItemKind = "compaction"
	// ItemKindSessionEnd marks a session-lifecycle end event (e.g. Codex's
	// "session_end"), carrying an optional reason in Text. This is distinct
	// from TurnStatus because a session can end independent of any single
	// turn's own completion state (DIR-032).
	ItemKindSessionEnd ItemKind = "session_end"
	// ItemKindUnknown is a catch-all for events not yet modeled by a typed
	// kind above. It round-trips via the capped Raw field (see NewRawItem)
	// so nothing is silently dropped while still bounding memory use.
	ItemKindUnknown ItemKind = "unknown"
)

// AgentPhase distinguishes intermediate agent commentary from a turn's
// final output, when the source provider reports it (e.g. a Harmony-style
// "channel" of "commentary" vs "final" on a message item). Absent that
// signal, Phase is left unspecified rather than guessed.
type AgentPhase string

const (
	PhaseUnspecified AgentPhase = ""
	PhaseCommentary  AgentPhase = "commentary"
	PhaseFinal       AgentPhase = "final"
)

// ItemStatus captures the lifecycle status of an individual item (e.g. a
// tool call or command execution), when the source provider reports one.
type ItemStatus string

const (
	StatusUnspecified ItemStatus = ""
	StatusCompleted   ItemStatus = "completed"
	StatusFailed      ItemStatus = "failed"
	StatusInProgress  ItemStatus = "in_progress"
)

// TurnStatus captures the lifecycle status of a whole Turn, when the source
// provider reports one (e.g. Codex's turn.completed/turn.failed events).
type TurnStatus string

const (
	TurnStatusUnspecified TurnStatus = ""
	TurnStatusCompleted   TurnStatus = "completed"
	TurnStatusFailed      TurnStatus = "failed"
	TurnStatusInProgress  TurnStatus = "in_progress"
	// TurnStatusAborted marks a turn explicitly interrupted before
	// completion (e.g. Codex's "turn_aborted" event — a user interrupt or
	// similar), distinct from TurnStatusFailed (an error) and
	// TurnStatusUnspecified (no terminal signal observed at all, e.g. a
	// turn still open when the rollout stream ends — see the DIR-032
	// "live/ongoing" note in docs/reference/codex-history-model.md).
	TurnStatusAborted TurnStatus = "aborted"
)

// HistoryCompleteness describes how much of a Turn's content is actually
// present, distinguishing a fully materialized record from a placeholder.
// Backends that only ever return complete records (the Codex rollout/files
// adapter, and the stable app-server thread/read surface as currently
// confirmed — see docs/reference/codex-app-server.md) leave this at
// HistoryCompletenessUnspecified; IsFull treats that the same as
// HistoryCompletenessFull. Backends that CAN return partial/placeholder
// records (e.g. a future paginated or summary-view app-server response)
// must set one of the other values explicitly so a placeholder is never
// mistaken for complete content (DIR-032 Contract).
type HistoryCompleteness string

const (
	HistoryCompletenessUnspecified HistoryCompleteness = ""
	HistoryCompletenessFull        HistoryCompleteness = "full"
	// HistoryCompletenessSummary marks a turn whose content is a
	// server-provided summary/preview standing in for the full record —
	// e.g. a preview-only projection returned by a listing call that never
	// loaded full turn content.
	HistoryCompletenessSummary HistoryCompleteness = "summary"
	// HistoryCompletenessUnloaded marks a turn whose content was never
	// requested/fetched at all (a placeholder position is known, but no
	// content — not even a summary — has been retrieved for it).
	HistoryCompletenessUnloaded HistoryCompleteness = "unloaded"
	// HistoryCompletenessTruncated marks a turn at (or after) the point
	// where loading stopped early (e.g. a rollout's maxLines cap), so
	// callers know turns/content after this point are missing, not merely
	// absent from this particular query.
	HistoryCompletenessTruncated HistoryCompleteness = "truncated"
	// HistoryCompletenessUnavailable marks a turn known to exist (from
	// metadata) whose content could not be loaded at all (e.g. a page
	// fetch failed and no retry has succeeded yet).
	HistoryCompletenessUnavailable HistoryCompleteness = "unavailable"
)

// IsFull reports whether c represents fully materialized content: true for
// HistoryCompletenessFull and the zero value (a backend that doesn't
// distinguish completeness states but has always returned complete
// records). Every other value is a placeholder/partial state a caller must
// not treat as complete.
func (c HistoryCompleteness) IsFull() bool {
	return c == HistoryCompletenessUnspecified || c == HistoryCompletenessFull
}

// CompactionBoundary is typed metadata for an ItemKindCompaction item: it
// marks a point in a Turn's Item stream where preceding context was
// replaced/summarized. It deliberately does NOT re-embed the superseded
// content — the original pre-compaction Items remain in the stream, in
// their original position, as the historical record. CompactionBoundary
// only records that/why/what replaced them, so a caller reconstructing
// "current" context can skip everything before the boundary without a
// content query ever concatenating the original text and its replacement
// as if both were live simultaneously (DIR-032 Contract: "preserves visible
// replacement history and boundary metadata without duplicating superseded
// content").
type CompactionBoundary struct {
	// Reason is the provider-reported cause (e.g. "context_window"), when
	// reported.
	Reason string `json:"reason,omitempty"`
	// Summary is a human-readable note of what replaced the prior context,
	// when the provider reports one. It is boundary metadata, not a
	// duplicate message: callers must not fold it into UserText/
	// AssistantText message projections.
	Summary string `json:"summary,omitempty"`
}

// LineageStatus describes what is known about a session/thread's
// parent/child spawn relationship (see Session.ParentThreadID). Unlike a
// bare empty ParentThreadID (which is ambiguous between "confirmed no
// parent" and "spawn metadata unavailable"), LineageStatus makes that
// distinction explicit per the DIR-032 Contract: "explicit unknown state
// when spawn metadata was suppressed".
type LineageStatus string

const (
	LineageStatusUnspecified LineageStatus = ""
	// LineageStatusRoot means the provider positively confirmed this
	// session has no parent (a genuine top-level thread).
	LineageStatusRoot LineageStatus = "root"
	// LineageStatusChild means ParentThreadID is populated from a source
	// that reliably reports it.
	LineageStatusChild LineageStatus = "child"
	// LineageStatusUnknown means spawn metadata was not available (e.g. an
	// older threads-table schema with no parent_thread_id column, or a
	// subagent source kind whose spawn edge was suppressed) — this session
	// must NOT be presented as a confirmed root just because
	// ParentThreadID is empty.
	LineageStatusUnknown LineageStatus = "unknown"
)

// maxRawBytes bounds how much raw provenance an unknown/raw Item may embed.
// Payloads larger than this are truncated (see NewRawItem) so a single
// oversized or malformed event can never balloon a Turn's memory footprint
// merely by being unrecognized.
const maxRawBytes = 4096

// Item is one entry in a Turn's ordered, loss-minimizing event stream. Not
// every field applies to every Kind; each provider adapter populates the
// subset relevant to the events it emits. Item deliberately mirrors (rather
// than replaces) the legacy Turn.UserText/AssistantText/ToolCalls fields:
// those remain a derived compatibility projection over Items, computed by
// the provider adapter, so existing MCP query output is unaffected while
// new item-level fidelity becomes available for future queries.
type Item struct {
	// ID is the provider-native identifier for this item, when the source
	// event carries one (e.g. a Codex tool call_id or dot-schema item id).
	// Empty when the source has no stable identity for the event (most
	// plain message segments).
	ID string `json:"id,omitempty"`

	Kind   ItemKind   `json:"kind"`
	Role   string     `json:"role,omitempty"`
	Phase  AgentPhase `json:"phase,omitempty"`
	Status ItemStatus `json:"status,omitempty"`

	// Text carries message content for ItemKindUserMessage/AgentMessage.
	Text string `json:"text,omitempty"`

	// ToolCallID links an ItemKindToolCall to its ItemKindToolResult (and
	// is also set on the ItemKindToolCall item itself, equal to ID, so
	// both items can be joined on the same key).
	ToolCallID string          `json:"tool_call_id,omitempty"`
	ToolName   string          `json:"tool_name,omitempty"`
	Input      json.RawMessage `json:"input,omitempty"`
	Output     string          `json:"output,omitempty"`
	IsError    bool            `json:"is_error,omitempty"`

	// Command/ExitCode are populated for ItemKindCommandExecution.
	Command  string `json:"command,omitempty"`
	ExitCode *int   `json:"exit_code,omitempty"`

	// Paths is populated for ItemKindFileChange (files touched).
	Paths []string `json:"paths,omitempty"`

	// Query is populated for ItemKindWebSearch.
	Query string `json:"query,omitempty"`

	// PlanSteps is populated for ItemKindPlanUpdate.
	PlanSteps []string `json:"plan_steps,omitempty"`

	// Compaction carries typed boundary metadata for ItemKindCompaction
	// items (DIR-032). Nil for every other Kind.
	Compaction *CompactionBoundary `json:"compaction,omitempty"`

	Timestamp time.Time `json:"timestamp"`

	// Source records the provenance channel that produced this item (e.g.
	// Codex's "event_msg" vs "response_item"), useful for debugging
	// dedup/precedence decisions without re-parsing the raw rollout.
	Source string `json:"source,omitempty"`

	// Raw preserves the provider's original event for ItemKindUnknown (and
	// optionally other kinds needing extra provenance), capped at
	// maxRawBytes via NewRawItem. When the payload fits under the cap, Raw
	// is the original JSON value; when truncated, Raw is instead a JSON
	// string of the first maxRawBytes bytes (since a mid-token cut is not
	// guaranteed to be valid JSON on its own) and RawTruncated is set, so
	// callers can distinguish "this is the whole event" from "this was cut
	// short".
	Raw          json.RawMessage `json:"raw,omitempty"`
	RawTruncated bool            `json:"raw_truncated,omitempty"`
}

// NewRawItem builds an Item of the given kind that preserves payload as its
// Raw provenance, capping it at maxRawBytes so unrecognized or oversized
// provider events can never embed unbounded data into a Turn. Truncation
// sets RawTruncated rather than attempting to produce truncated-but-valid
// JSON.
func NewRawItem(kind ItemKind, timestamp time.Time, payload []byte) Item {
	item := Item{
		Kind:      kind,
		Timestamp: timestamp,
	}
	if len(payload) <= maxRawBytes {
		item.Raw = append(json.RawMessage(nil), payload...)
		return item
	}
	// Truncated bytes are not guaranteed to be valid JSON on their own (the
	// cut point may fall mid-token), so wrap them as a JSON string literal
	// rather than embedding them as a raw (potentially malformed) JSON
	// value. This keeps the Item itself always marshal/unmarshal-safe.
	truncated, err := json.Marshal(string(payload[:maxRawBytes]))
	if err != nil {
		truncated = []byte(`""`)
	}
	item.Raw = truncated
	item.RawTruncated = true
	return item
}
