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

type turnBuilder struct {
	current         *conversation.Turn
	toolCallMap     map[string]int
	turns           []conversation.Turn
	unknown         []json.RawMessage
	totalTokenUsage conversation.TokenUsage
}

func newTurnBuilder() *turnBuilder {
	return &turnBuilder{toolCallMap: make(map[string]int)}
}

func (b *turnBuilder) flush() {
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
			b.current.UserText = payload.Message
		case "agent_message":
			if b.current.AssistantText != "" {
				b.current.AssistantText += "\n"
			}
			b.current.AssistantText += payload.Message
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
				if b.current.UserText == "" {
					b.current.UserText = text
				}
			case "assistant":
				if b.current.AssistantText != "" && text != "" {
					b.current.AssistantText += "\n"
				}
				b.current.AssistantText += text
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
