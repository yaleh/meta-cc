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
	// or trimmed).
	ItemKindCompaction ItemKind = "compaction"
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
