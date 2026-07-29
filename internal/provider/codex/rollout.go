package codex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/parser"
)

type schemaVersion int

const (
	schemaLegacy schemaVersion = iota
	schemaNew
)

func loadTurnsFromSession(session conversation.Session, maxLines int) ([]conversation.Turn, conversation.TokenUsage, error) {
	path, err := RolloutPath(session)
	if err != nil {
		return nil, conversation.TokenUsage{}, err
	}
	return loadTurnsFromRollout(path, maxLines)
}

// RolloutPath extracts the on-disk rollout file path recorded for a Codex
// session (stored in session.Extensions by the SQLite scan). Callers that
// need the raw file backing a Codex session — e.g. Stage 1 discovery tools —
// should use this instead of re-deriving the path themselves.
func RolloutPath(session conversation.Session) (string, error) {
	var ext struct {
		RolloutPath string `json:"rollout_path"`
	}
	if err := json.Unmarshal(session.Extensions, &ext); err != nil {
		return "", err
	}
	if ext.RolloutPath == "" {
		return "", fmt.Errorf("missing rollout_path for session %s", session.ID)
	}
	return ext.RolloutPath, nil
}

func loadTurnsFromRollout(path string, maxLines int) ([]conversation.Turn, conversation.TokenUsage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, conversation.TokenUsage{}, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	var (
		lineCount int
		version   schemaVersion
		detected  bool
		builder   = newTurnBuilder()
	)

	for {
		line, _, readErr := parser.ReadLineFiltered(reader, parser.StrategyDefault)
		if len(line) == 0 && readErr == io.EOF {
			break
		}
		if readErr != nil && readErr != io.EOF {
			return nil, conversation.TokenUsage{}, readErr
		}
		lineCount++
		if lineCount > maxLines {
			slog.Warn("codex rollout truncated", "path", path, "max_lines", maxLines)
			builder.truncated = true
			break
		}
		if !detected {
			version = detectSchemaVersion(line)
			detected = true
		}
		if version == schemaNew {
			builder.applyNew(line)
		} else {
			builder.applyLegacy(line)
		}
		if readErr == io.EOF {
			break
		}
	}
	builder.flush()
	return builder.turns, builder.totalTokenUsage, nil
}

func detectSchemaVersion(firstLine []byte) schemaVersion {
	var payload struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(firstLine, &payload); err == nil && strings.Contains(payload.Type, ".") {
		return schemaNew
	}
	return schemaLegacy
}

// Item-stream construction and the legacy-field compatibility projection.
//
// Codex rollouts report user/assistant text through two independent
// channels for the same turn: the legacy "event_msg" notification stream
// (e.g. agent_message / user_message) and the "response_item" transcript
// stream (message records with role user/assistant). Live Codex CLI 0.145
// rollouts record the same logical utterance through BOTH channels, so
// naively appending every occurrence doubles every message.
//
// Precedence rule (see docs/reference/jsonl-schema.md "Duplicate
// Assistant/User Text"): response_item is the richer, canonical
// representation and wins whenever both channels report a segment at the
// same position within the turn; event_msg is used only as a fallback for
// positions response_item never reported (legacy/event-only sessions).
// Observed Codex rollouts always emit the event_msg copy of a segment
// immediately before its response_item counterpart, so this is implemented
// as an in-place upgrade: an event_msg text is appended to the turn's Item
// stream as a placeholder, and the next response_item text for that role
// overwrites the oldest not-yet-upgraded placeholder in place (FIFO)
// instead of appending a duplicate. This keeps the Item stream itself
// already deduped, in true encounter order (interleaved with tool calls
// and everything else), so the legacy UserText/AssistantText/ToolCalls
// fields can be derived from Items by projection rather than by a second,
// independent parse. Matching is positional per role, never content-based,
// so two legitimately repeated messages (the same text sent again in a
// later turn, or two distinct segments in the same turn that happen to
// share text) are both preserved rather than collapsed.
type turnBuilder struct {
	current                  *conversation.Turn
	turns                    []conversation.Turn
	unknown                  []json.RawMessage
	totalTokenUsage          conversation.TokenUsage
	pendingUserEventMsg      []int
	pendingAssistantEventMsg []int

	// truncated is set once, immediately before the final flush() call that
	// follows loadTurnsFromRollout's maxLines-triggered break, so that
	// final (possibly partial) turn is marked HistoryCompletenessTruncated
	// instead of Full — a caller reading the returned turns then knows
	// "content after this point is missing", not merely absent from this
	// particular query (DIR-032).
	truncated bool
}

func newTurnBuilder() *turnBuilder {
	return &turnBuilder{}
}

// addEventMsgText appends a new message Item sourced from the event_msg
// channel and remembers its position as an upgrade candidate for a later
// response_item segment at the same (turn, role) position.
func (b *turnBuilder) addEventMsgText(kind conversation.ItemKind, role, text, timestamp string) {
	idx := b.appendTextItem(kind, role, "", "event_msg", text, timestamp)
	if kind == conversation.ItemKindUserMessage {
		b.pendingUserEventMsg = append(b.pendingUserEventMsg, idx)
	} else {
		b.pendingAssistantEventMsg = append(b.pendingAssistantEventMsg, idx)
	}
}

// addResponseItemText reconciles a response_item text segment: it upgrades
// the oldest pending event_msg placeholder for the same role in place, or
// (when there is no pending placeholder — a response_item-only rollout, or
// more response_item segments than event_msg ones) appends a new Item.
func (b *turnBuilder) addResponseItemText(kind conversation.ItemKind, role, phase, text, timestamp string) {
	pending := &b.pendingAssistantEventMsg
	if kind == conversation.ItemKindUserMessage {
		pending = &b.pendingUserEventMsg
	}
	if len(*pending) > 0 {
		idx := (*pending)[0]
		*pending = (*pending)[1:]
		b.current.Items[idx].Text = text
		b.current.Items[idx].Source = "response_item"
		b.current.Items[idx].Phase = agentPhase(phase)
		return
	}
	b.appendTextItem(kind, role, phase, "response_item", text, timestamp)
}

// The append*Item helpers (appendTextItem, appendToolCallItem,
// appendToolResultItem, appendCompactionItem) each begin with a defensive
// ensureTurn (DIR-061): every dispatch case already ensures a turn before
// appending, but the guard here means no future event case can reintroduce
// a nil-deref panic by omitting it. ensureTurn is a no-op when b.current is
// already set, so existing behavior is unchanged.
func (b *turnBuilder) appendTextItem(kind conversation.ItemKind, role, phase, source, text, timestamp string) int {
	b.ensureTurn("", timestamp)
	ts, _ := time.Parse(time.RFC3339, timestamp)
	b.current.Items = append(b.current.Items, conversation.Item{
		Kind:      kind,
		Role:      role,
		Phase:     agentPhase(phase),
		Text:      text,
		Timestamp: ts.UTC(),
		Source:    source,
	})
	return len(b.current.Items) - 1
}

// agentPhase maps a Harmony-style "channel" value ("commentary"/"final"),
// when a Codex rollout reports one, to the canonical AgentPhase. Absent or
// unrecognized values map to PhaseUnspecified rather than being guessed.
func agentPhase(channel string) conversation.AgentPhase {
	switch channel {
	case "commentary":
		return conversation.PhaseCommentary
	case "final":
		return conversation.PhaseFinal
	default:
		return conversation.PhaseUnspecified
	}
}

func (b *turnBuilder) appendToolCallItem(callID, name string, input json.RawMessage, timestamp string) {
	b.ensureTurn("", timestamp) // DIR-061 defensive guard, see appendTextItem
	ts, _ := time.Parse(time.RFC3339, timestamp)
	b.current.Items = append(b.current.Items, conversation.Item{
		ID:         callID,
		Kind:       conversation.ItemKindToolCall,
		ToolCallID: callID,
		ToolName:   name,
		Input:      input,
		Timestamp:  ts.UTC(),
	})
}

func (b *turnBuilder) appendToolResultItem(callID, output string, isError bool, timestamp string) {
	b.ensureTurn("", timestamp) // DIR-061 defensive guard, see appendTextItem
	ts, _ := time.Parse(time.RFC3339, timestamp)
	b.current.Items = append(b.current.Items, conversation.Item{
		Kind:       conversation.ItemKindToolResult,
		ToolCallID: callID,
		Output:     output,
		IsError:    isError,
		Timestamp:  ts.UTC(),
	})
}

// projectLegacyFields derives the backward-compatible
// UserText/AssistantText/ToolCalls fields from the turn's Items, rather
// than re-parsing the rollout. Message items are joined per role in
// encounter order (matching the historical dedupSegments.merge() join
// behavior); tool call/result items are paired by ToolCallID.
func (b *turnBuilder) projectLegacyFields() {
	var userParts, assistantParts []string
	var calls []conversation.ToolCall
	callIndex := make(map[string]int)
	for _, item := range b.current.Items {
		switch item.Kind {
		case conversation.ItemKindUserMessage:
			userParts = append(userParts, item.Text)
		case conversation.ItemKindAgentMessage:
			assistantParts = append(assistantParts, item.Text)
		case conversation.ItemKindToolCall:
			callIndex[item.ToolCallID] = len(calls)
			calls = append(calls, conversation.ToolCall{
				ID:        item.ToolCallID,
				Name:      item.ToolName,
				Input:     item.Input,
				Timestamp: item.Timestamp,
			})
		case conversation.ItemKindToolResult:
			if idx, ok := callIndex[item.ToolCallID]; ok {
				calls[idx].Output = item.Output
				calls[idx].IsError = item.IsError
			}
		}
	}
	if len(userParts) > 0 {
		b.current.UserText = strings.Join(userParts, "\n")
	}
	if len(assistantParts) > 0 {
		b.current.AssistantText = strings.Join(assistantParts, "\n")
	}
	b.current.ToolCalls = calls
}

func (b *turnBuilder) flush() {
	if b.current != nil {
		b.projectLegacyFields()
	}
	if b.current != nil && len(b.unknown) > 0 {
		ext, _ := json.Marshal(map[string][]json.RawMessage{
			"codex_events": b.unknown,
		})
		b.current.Extensions = ext
	}
	// DIR-050: a turn is retained whenever it carries ANY content at all —
	// not just the legacy-projected UserText/AssistantText/ToolCalls/
	// TokenUsage/Extensions fields. DIR-032 added lifecycle-only item kinds
	// (ItemKindSessionEnd, ItemKindCompaction) and a status-only signal
	// (TurnStatusAborted, set with no accompanying Item) that none of those
	// five checks can see, so a turn whose ONLY content is e.g. a bare
	// session_end or turn_aborted event was silently discarded. Checking
	// len(b.current.Items) > 0 (rather than enumerating specific item
	// kinds) also covers any future item kind added to the Items stream
	// without another handler-specific carve-out here.
	hasContent := b.current != nil && (b.current.UserText != "" || b.current.AssistantText != "" || len(b.current.ToolCalls) > 0 || b.current.TokenUsage.HasAny() || len(b.current.Extensions) > 0 || len(b.current.Items) > 0 || b.current.Status != conversation.TurnStatusUnspecified)
	if hasContent {
		if b.truncated {
			b.current.Completeness = conversation.HistoryCompletenessTruncated
		} else {
			b.current.Completeness = conversation.HistoryCompletenessFull
		}
		b.turns = append(b.turns, *b.current)
	}
	b.current = nil
	b.unknown = nil
	b.pendingUserEventMsg = nil
	b.pendingAssistantEventMsg = nil
}

func (b *turnBuilder) ensureTurn(id, timestamp string) {
	if b.current == nil {
		ts, _ := time.Parse(time.RFC3339, timestamp)
		b.current = &conversation.Turn{ID: id, Timestamp: ts.UTC()}
	}
}

func (b *turnBuilder) applyLegacy(line []byte) {
	var event struct {
		Timestamp string          `json:"timestamp"`
		Type      string          `json:"type"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(line, &event); err != nil {
		return
	}

	switch event.Type {
	case "session_meta":
		return
	case "turn_context":
		var payload struct {
			TurnID string `json:"turn_id"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		if b.current != nil && payload.TurnID != "" && b.current.ID != payload.TurnID {
			b.flush()
		}
		b.ensureTurn(payload.TurnID, event.Timestamp)
	case "event_msg":
		var payload struct {
			Type    string `json:"type"`
			Message string `json:"message"`
			TurnID  string `json:"turn_id"`
			Summary string `json:"summary"`
			Info    struct {
				LastTokenUsage  codexTokenUsage `json:"last_token_usage"`
				TotalTokenUsage codexTokenUsage `json:"total_token_usage"`
			} `json:"info"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		if payload.Type == "task_started" {
			b.flush()
			b.ensureTurn(payload.TurnID, event.Timestamp)
			return
		}
		b.ensureTurn(payload.TurnID, event.Timestamp)
		switch payload.Type {
		case "user_message":
			b.addEventMsgText(conversation.ItemKindUserMessage, "user", payload.Message, event.Timestamp)
		case "agent_message":
			b.addEventMsgText(conversation.ItemKindAgentMessage, "assistant", payload.Message, event.Timestamp)
		case "token_count":
			b.applyTokenUsage(event.Timestamp, payload.Info.LastTokenUsage, payload.Info.TotalTokenUsage)
			return
		case "context_compacted":
			// DIR-032: typed compaction boundary rather than raw passthrough
			// (see CompactionBoundary's doc comment) — the summary is
			// boundary metadata, never folded into UserText/AssistantText.
			b.appendCompactionItem("", payload.Summary, event.Timestamp)
		default:
			b.appendUnknown(line)
		}
	case "response_item":
		var envelope struct {
			Type      string          `json:"type"`
			Name      string          `json:"name"`
			Arguments string          `json:"arguments"`
			Input     json.RawMessage `json:"input"`
			CallID    string          `json:"call_id"`
			ID        string          `json:"id"`
			Output    string          `json:"output"`
			Status    string          `json:"status"`
			IsError   bool            `json:"is_error"`
			Error     string          `json:"error"`
			Role      string          `json:"role"`
			Channel   string          `json:"channel"`
			Content   json.RawMessage `json:"content"`
			Info      struct {
				LastTokenUsage  codexTokenUsage `json:"last_token_usage"`
				TotalTokenUsage codexTokenUsage `json:"total_token_usage"`
			} `json:"info"`
		}
		_ = json.Unmarshal(event.Payload, &envelope)
		b.ensureTurn("", event.Timestamp)
		switch envelope.Type {
		case "function_call", "custom_tool_call":
			callID := firstNonEmpty(envelope.CallID, envelope.ID)
			input := json.RawMessage(envelope.Arguments)
			if envelope.Type == "custom_tool_call" {
				input = encodeCustomToolInput(envelope.Input)
			}
			b.appendToolCallItem(callID, envelope.Name, input, event.Timestamp)
		case "function_call_output", "custom_tool_call_output":
			callID := firstNonEmpty(envelope.CallID, envelope.ID)
			isError := envelope.IsError || isErrorStatus(envelope.Status) || envelope.Error != ""
			output := envelope.Output
			if output == "" {
				output = envelope.Error
			}
			b.appendToolResultItem(callID, output, isError, event.Timestamp)
		case "message":
			text := extractResponseItemText(envelope.Content)
			switch envelope.Role {
			case "user":
				b.addResponseItemText(conversation.ItemKindUserMessage, "user", envelope.Channel, text, event.Timestamp)
			case "assistant":
				b.addResponseItemText(conversation.ItemKindAgentMessage, "assistant", envelope.Channel, text, event.Timestamp)
			case "developer", "system":
				return
			default:
				b.appendUnknown(line)
			}
		case "reasoning":
			b.current.Items = append(b.current.Items, conversation.NewRawItem(conversation.ItemKindReasoning, parseTimestamp(event.Timestamp), line))
			return
		case "token_count":
			b.applyTokenUsage(event.Timestamp, envelope.Info.LastTokenUsage, envelope.Info.TotalTokenUsage)
			return
		default:
			b.appendUnknown(line)
		}
	case "compacted":
		// DIR-032: typed compaction boundary (see CompactionBoundary) for
		// the top-level "compacted" event family, distinct from the
		// event_msg "context_compacted" notification handled above — Codex
		// 0.145 rollouts emit both for the same logical boundary, so each
		// becomes its own typed compaction Item rather than being merged.
		var payload struct {
			TurnID string `json:"turn_id"`
			Reason string `json:"reason"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		b.ensureTurn("", event.Timestamp)
		b.appendCompactionItem(payload.Reason, "", event.Timestamp)
	case "turn_aborted":
		// DIR-032: typed turn lifecycle status rather than raw passthrough —
		// see docs/reference/codex-history-model.md's turn-status coverage.
		var payload struct {
			TurnID string `json:"turn_id"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		b.ensureTurn(payload.TurnID, event.Timestamp)
		b.current.Status = conversation.TurnStatusAborted
	case "session_end":
		// DIR-032: typed session-lifecycle item rather than raw passthrough.
		var payload struct {
			Reason string `json:"reason"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		b.ensureTurn("", event.Timestamp)
		b.current.Items = append(b.current.Items, conversation.Item{
			Kind:      conversation.ItemKindSessionEnd,
			Text:      payload.Reason,
			Timestamp: parseTimestamp(event.Timestamp),
			Source:    "session_end",
		})
	default:
		b.appendUnknown(line)
	}
}

// appendCompactionItem appends a typed ItemKindCompaction item carrying a
// CompactionBoundary (DIR-032). reason/summary are whichever fields the
// triggering event reported (the legacy schema splits them across two
// separate event families — top-level "compacted" reports reason, event_msg
// "context_compacted" reports summary — so each call only ever sets one).
func (b *turnBuilder) appendCompactionItem(reason, summary, timestamp string) {
	b.ensureTurn("", timestamp) // DIR-061 defensive guard, see appendTextItem
	ts := parseTimestamp(timestamp)
	b.current.Items = append(b.current.Items, conversation.Item{
		Kind:      conversation.ItemKindCompaction,
		Timestamp: ts,
		Source:    "compaction_boundary",
		Compaction: &conversation.CompactionBoundary{
			Reason:  reason,
			Summary: summary,
		},
	})
}

type codexTokenUsage struct {
	InputTokens           int `json:"input_tokens"`
	CachedInputTokens     int `json:"cached_input_tokens"`
	OutputTokens          int `json:"output_tokens"`
	ReasoningOutputTokens int `json:"reasoning_output_tokens"`
}

// applyTokenUsage records one token_count event's usage against the current
// turn. Codex emits one token_count event per model API call, and a
// tool-using turn makes multiple calls in sequence — last_token_usage is
// per-call, not cumulative-within-turn (unlike total_token_usage, which is
// cumulative for the whole session). So the turn's TokenUsage accumulates
// each event's last_token_usage rather than being overwritten by it
// (DIR-065); otherwise only the final call's usage survives and every
// earlier call in the turn is silently discarded, undercounting the turn's
// real consumption and disagreeing with the (correct) cumulative total.
func (b *turnBuilder) applyTokenUsage(timestamp string, last, total codexTokenUsage) {
	b.ensureTurn("", timestamp)
	if last.InputTokens != 0 || last.OutputTokens != 0 || last.CachedInputTokens != 0 || last.ReasoningOutputTokens != 0 {
		b.current.TokenUsage.InputTokens += last.InputTokens
		b.current.TokenUsage.OutputTokens += last.OutputTokens
		b.current.TokenUsage.CacheTokens += last.CachedInputTokens
		// DIR-071: reasoning_output_tokens was previously read only as part of
		// the non-zero guard above and then discarded, so a turn's reasoning
		// cost vanished from every query. Accumulate it alongside the other
		// per-call categories so reasoning-inclusive turn totals reconcile.
		b.current.TokenUsage.ReasoningOutputTokens += last.ReasoningOutputTokens
	}
	if total.InputTokens != 0 || total.OutputTokens != 0 || total.CachedInputTokens != 0 || total.ReasoningOutputTokens != 0 {
		b.totalTokenUsage = conversation.TokenUsage{
			InputTokens:           total.InputTokens,
			OutputTokens:          total.OutputTokens,
			CacheTokens:           total.CachedInputTokens,
			ReasoningOutputTokens: total.ReasoningOutputTokens,
		}
	}
}

func (b *turnBuilder) applyNew(line []byte) {
	var event struct {
		Timestamp string          `json:"timestamp"`
		Type      string          `json:"type"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(line, &event); err != nil {
		return
	}

	switch event.Type {
	case "thread.started":
		return
	case "turn.started":
		var payload struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		b.flush()
		b.ensureTurn(payload.ID, event.Timestamp)
	case "item.message":
		var payload struct {
			Role    string `json:"role"`
			Content string `json:"content"`
			Channel string `json:"channel"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		b.ensureTurn("", event.Timestamp)
		// Unlike the legacy schema's dual event_msg/response_item channels,
		// the dot-schema reports each message exactly once, so items are
		// appended directly (no dedup upgrade needed) — but still appended
		// rather than overwritten, so multiple commentary/final messages in
		// one turn are all preserved in order instead of only the last one
		// surviving.
		if payload.Role == "user" {
			b.appendTextItem(conversation.ItemKindUserMessage, "user", payload.Channel, "item.message", payload.Content, event.Timestamp)
		} else if payload.Role == "assistant" {
			b.appendTextItem(conversation.ItemKindAgentMessage, "assistant", payload.Channel, "item.message", payload.Content, event.Timestamp)
		}
	case "item.tool_call":
		var payload struct {
			ID    string          `json:"id"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		b.ensureTurn("", event.Timestamp)
		b.appendToolCallItem(payload.ID, payload.Name, payload.Input, event.Timestamp)
	case "item.tool_result":
		var payload struct {
			ID      string `json:"id"`
			Output  string `json:"output"`
			IsError bool   `json:"is_error"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		// DIR-061: guard like the item.message/item.tool_call siblings — an
		// item.tool_result can be the first record of a head-truncated
		// rollout or arrive after turn.completed (which flushes, setting
		// b.current = nil); without this guard appendToolResultItem
		// nil-derefs instead of degrading gracefully.
		b.ensureTurn("", event.Timestamp)
		b.appendToolResultItem(payload.ID, payload.Output, payload.IsError, event.Timestamp)
	case "turn.completed":
		if b.current != nil {
			b.current.Status = conversation.TurnStatusCompleted
		}
		b.flush()
	case "turn.failed", "error":
		b.appendUnknown(line)
		b.current.Status = conversation.TurnStatusFailed
	default:
		b.appendUnknown(line)
	}
}

// appendUnknown preserves an unrecognized event both in the legacy
// Extensions.codex_events bag (unchanged, for backward compatibility) and
// as a capped ItemKindUnknown Item in the ordered item stream, so
// unrecognized event families are round-trippable at the item level too
// without embedding unbounded raw payloads.
//
// DIR-064: the turn this event opens is stamped with the EVENT's own
// timestamp, parsed once and shared by both the Turn (via ensureTurn) and the
// Item so the two always agree. time.Now() is used only as a documented last
// resort when the event carries no parseable timestamp. Previously the turn
// was stamped with the query-time wall clock and the event time was parsed
// only for the Item; because ensureTurn is a no-op once a turn is open, every
// later event in the turn inherited that now(), and Normalize emitted it as
// the record timestamp — so historical sessions carried today's date, skewing
// since/until filtering and recency sorting.
func (b *turnBuilder) appendUnknown(line []byte) {
	var stamped struct {
		Timestamp string `json:"timestamp"`
	}
	_ = json.Unmarshal(line, &stamped)
	ts := parseTimestamp(stamped.Timestamp)

	turnTs := ts
	if turnTs.IsZero() {
		turnTs = time.Now().UTC() // last resort: event has no usable timestamp
	}
	b.ensureTurn("", turnTs.Format(time.RFC3339))
	b.unknown = append(b.unknown, append(json.RawMessage(nil), line...))

	itemTs := ts
	if itemTs.IsZero() {
		itemTs = turnTs // inherit the turn time (event time, or now() if none)
	}
	b.current.Items = append(b.current.Items, conversation.NewRawItem(conversation.ItemKindUnknown, itemTs, line))
}

func parseTimestamp(timestamp string) time.Time {
	t, _ := time.Parse(time.RFC3339, timestamp)
	return t.UTC()
}

func encodeCustomToolInput(input json.RawMessage) json.RawMessage {
	if len(input) == 0 || string(input) == "null" {
		return json.RawMessage(`{}`)
	}
	var decoded interface{}
	if err := json.Unmarshal(input, &decoded); err == nil {
		if _, ok := decoded.(map[string]interface{}); ok {
			return input
		}
	}
	data, err := json.Marshal(map[string]json.RawMessage{"input": input})
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return data
}

func isErrorStatus(status string) bool {
	switch status {
	case "", "completed", "success", "ok":
		return false
	default:
		return true
	}
}

func extractResponseItemText(content json.RawMessage) string {
	var items []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(content, &items); err != nil {
		return ""
	}
	var parts []string
	for _, item := range items {
		if item.Text != "" {
			parts = append(parts, item.Text)
		}
	}
	return strings.Join(parts, "\n")
}
