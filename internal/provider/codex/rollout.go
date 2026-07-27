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

// dedupSegments accumulates one role's text (user or assistant) as it is
// reported through two independent Codex rollout channels for the same
// turn: the legacy "event_msg" notification stream (e.g. agent_message /
// user_message) and the "response_item" transcript stream (message records
// with role user/assistant). Live Codex CLI 0.145 rollouts record the same
// logical utterance through BOTH channels, so naively concatenating every
// occurrence doubles every message.
//
// Precedence rule (see docs/reference/jsonl-schema.md "Duplicate
// Assistant/User Text"): response_item is the richer, canonical
// representation and wins whenever both channels report a segment at the
// same position within the turn; event_msg is used only as a fallback for
// positions response_item never reported (legacy/event-only sessions).
// Matching is positional and scoped to (turn, role) — content is never
// compared for equality — so two legitimately repeated messages (e.g. the
// same text sent again in a later turn, or two distinct segments in the
// same turn that happen to share text) are both preserved rather than
// collapsed.
type dedupSegments struct {
	eventMsg     []string
	responseItem []string
}

func (s *dedupSegments) addEventMsg(text string) {
	s.eventMsg = append(s.eventMsg, text)
}

func (s *dedupSegments) addResponseItem(text string) {
	s.responseItem = append(s.responseItem, text)
}

// merge reconciles the two channels position-by-position, preferring
// response_item and falling back to event_msg. Returns "" when neither
// channel reported anything for this role, so callers can distinguish "no
// legacy-schema segments recorded" from "segments recorded via the newer
// (dot-schema) path", which sets the turn's text directly and never
// populates dedupSegments.
func (s *dedupSegments) merge() string {
	n := len(s.eventMsg)
	if len(s.responseItem) > n {
		n = len(s.responseItem)
	}
	if n == 0 {
		return ""
	}
	parts := make([]string, 0, n)
	for i := 0; i < n; i++ {
		if i < len(s.responseItem) {
			parts = append(parts, s.responseItem[i])
		} else {
			parts = append(parts, s.eventMsg[i])
		}
	}
	return strings.Join(parts, "\n")
}

type turnBuilder struct {
	current           *conversation.Turn
	toolCallMap       map[string]int
	turns             []conversation.Turn
	unknown           []json.RawMessage
	totalTokenUsage   conversation.TokenUsage
	userSegments      dedupSegments
	assistantSegments dedupSegments
}

func newTurnBuilder() *turnBuilder {
	return &turnBuilder{toolCallMap: make(map[string]int)}
}

// finalizeSegments reconciles any legacy-schema user/assistant text
// accumulated for the current turn (see dedupSegments) into the turn's
// final UserText/AssistantText. It is a no-op when nothing was recorded
// through the legacy dedup path (e.g. the "new" dot-schema, which sets
// these fields directly and never touches dedupSegments).
func (b *turnBuilder) finalizeSegments() {
	if merged := b.userSegments.merge(); merged != "" {
		b.current.UserText = merged
	}
	if merged := b.assistantSegments.merge(); merged != "" {
		b.current.AssistantText = merged
	}
}

func (b *turnBuilder) flush() {
	if b.current != nil {
		b.finalizeSegments()
	}
	if b.current != nil && len(b.unknown) > 0 {
		ext, _ := json.Marshal(map[string][]json.RawMessage{
			"codex_events": b.unknown,
		})
		b.current.Extensions = ext
	}
	if b.current != nil && (b.current.UserText != "" || b.current.AssistantText != "" || len(b.current.ToolCalls) > 0 || hasUsage(b.current.TokenUsage) || len(b.current.Extensions) > 0) {
		b.turns = append(b.turns, *b.current)
	}
	b.current = nil
	b.toolCallMap = make(map[string]int)
	b.unknown = nil
	b.userSegments = dedupSegments{}
	b.assistantSegments = dedupSegments{}
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
			b.userSegments.addEventMsg(payload.Message)
		case "agent_message":
			b.assistantSegments.addEventMsg(payload.Message)
		case "token_count":
			b.applyTokenUsage(event.Timestamp, payload.Info.LastTokenUsage, payload.Info.TotalTokenUsage)
			return
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
			ts, _ := time.Parse(time.RFC3339, event.Timestamp)
			callID := firstNonEmpty(envelope.CallID, envelope.ID)
			input := json.RawMessage(envelope.Arguments)
			if envelope.Type == "custom_tool_call" {
				input = encodeCustomToolInput(envelope.Input)
			}
			call := conversation.ToolCall{
				ID:        callID,
				Name:      envelope.Name,
				Input:     input,
				Timestamp: ts.UTC(),
			}
			b.toolCallMap[callID] = len(b.current.ToolCalls)
			b.current.ToolCalls = append(b.current.ToolCalls, call)
		case "function_call_output", "custom_tool_call_output":
			callID := firstNonEmpty(envelope.CallID, envelope.ID)
			if idx, ok := b.toolCallMap[callID]; ok {
				b.current.ToolCalls[idx].Output = envelope.Output
				b.current.ToolCalls[idx].IsError = envelope.IsError || isErrorStatus(envelope.Status) || envelope.Error != ""
				if b.current.ToolCalls[idx].Output == "" {
					b.current.ToolCalls[idx].Output = envelope.Error
				}
			}
		case "message":
			text := extractResponseItemText(envelope.Content)
			switch envelope.Role {
			case "user":
				b.userSegments.addResponseItem(text)
			case "assistant":
				b.assistantSegments.addResponseItem(text)
			case "developer", "system":
				return
			default:
				b.appendUnknown(line)
			}
		case "reasoning":
			return
		case "token_count":
			b.applyTokenUsage(event.Timestamp, envelope.Info.LastTokenUsage, envelope.Info.TotalTokenUsage)
			return
		default:
			b.appendUnknown(line)
		}
	default:
		b.appendUnknown(line)
	}
}

type codexTokenUsage struct {
	InputTokens           int `json:"input_tokens"`
	CachedInputTokens     int `json:"cached_input_tokens"`
	OutputTokens          int `json:"output_tokens"`
	ReasoningOutputTokens int `json:"reasoning_output_tokens"`
}

func (b *turnBuilder) applyTokenUsage(timestamp string, last, total codexTokenUsage) {
	b.ensureTurn("", timestamp)
	if last.InputTokens != 0 || last.OutputTokens != 0 || last.CachedInputTokens != 0 || last.ReasoningOutputTokens != 0 {
		b.current.TokenUsage = conversation.TokenUsage{
			InputTokens:  last.InputTokens,
			OutputTokens: last.OutputTokens,
			CacheTokens:  last.CachedInputTokens,
		}
	}
	if total.InputTokens != 0 || total.OutputTokens != 0 || total.CachedInputTokens != 0 || total.ReasoningOutputTokens != 0 {
		b.totalTokenUsage = conversation.TokenUsage{
			InputTokens:  total.InputTokens,
			OutputTokens: total.OutputTokens,
			CacheTokens:  total.CachedInputTokens,
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
		}
		_ = json.Unmarshal(event.Payload, &payload)
		b.ensureTurn("", event.Timestamp)
		if payload.Role == "user" {
			b.current.UserText = payload.Content
		} else if payload.Role == "assistant" {
			b.current.AssistantText = payload.Content
		}
	case "item.tool_call":
		var payload struct {
			ID    string          `json:"id"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		b.ensureTurn("", event.Timestamp)
		ts, _ := time.Parse(time.RFC3339, event.Timestamp)
		b.toolCallMap[payload.ID] = len(b.current.ToolCalls)
		b.current.ToolCalls = append(b.current.ToolCalls, conversation.ToolCall{
			ID:        payload.ID,
			Name:      payload.Name,
			Input:     payload.Input,
			Timestamp: ts.UTC(),
		})
	case "item.tool_result":
		var payload struct {
			ID      string `json:"id"`
			Output  string `json:"output"`
			IsError bool   `json:"is_error"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		if idx, ok := b.toolCallMap[payload.ID]; ok {
			b.current.ToolCalls[idx].Output = payload.Output
			b.current.ToolCalls[idx].IsError = payload.IsError
		}
	case "turn.completed":
		b.flush()
	case "turn.failed", "error":
		b.appendUnknown(line)
	default:
		b.appendUnknown(line)
	}
}

func (b *turnBuilder) appendUnknown(line []byte) {
	b.ensureTurn("", time.Now().UTC().Format(time.RFC3339))
	b.unknown = append(b.unknown, append(json.RawMessage(nil), line...))
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

func hasUsage(usage conversation.TokenUsage) bool {
	return usage.InputTokens != 0 || usage.OutputTokens != 0 || usage.CacheTokens != 0
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
