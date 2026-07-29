package projection

import (
	"encoding/json"
	"fmt"

	"github.com/yaleh/meta-cc/internal/conversation"
)

type ContentKind uint8

const (
	UserMessage ContentKind = iota
	AssistantMessage
	ToolResults
)

type Content struct {
	Kind  ContentKind
	Value interface{}
}

// Contents constructs the message.content values emitted for one canonical
// turn. records.Normalize and the index candidate projection both consume it.
func Contents(turn conversation.Turn) []Content {
	var out []Content
	if turn.UserText != "" {
		out = append(out, Content{Kind: UserMessage, Value: turn.UserText})
	}
	if turn.AssistantText != "" || len(turn.ToolCalls) > 0 || hasUsage(turn.TokenUsage) {
		content := make([]interface{}, 0, len(turn.ToolCalls)+1)
		if turn.AssistantText != "" {
			content = append(content, map[string]interface{}{"type": "text", "text": turn.AssistantText})
		}
		for _, call := range turn.ToolCalls {
			content = append(content, map[string]interface{}{
				"type": "tool_use", "id": call.ID, "name": call.Name, "input": toolInput(call.Input),
			})
		}
		out = append(out, Content{Kind: AssistantMessage, Value: content})
	}
	var results []interface{}
	for _, call := range turn.ToolCalls {
		if call.Output == "" && !call.IsError {
			continue
		}
		result := map[string]interface{}{
			"type": "tool_result", "tool_use_id": call.ID, "content": call.Output,
			"is_error": call.IsError,
		}
		if call.IsError {
			result["status"] = "error"
			result["error"] = call.Output
		} else {
			result["status"] = "success"
		}
		results = append(results, result)
	}
	if len(results) > 0 {
		out = append(out, Content{Kind: ToolResults, Value: results})
	}
	return out
}

// SearchStrings returns the exact message.content tostring values emitted by
// records.Normalize for one canonical turn.
func SearchStrings(turn conversation.Turn) ([]string, error) {
	contents := Contents(turn)
	out := make([]string, 0, len(contents))
	for _, content := range contents {
		text, err := Tostring(content.Value)
		if err != nil {
			return nil, err
		}
		out = append(out, text)
	}
	return out, nil
}

// Tostring is gojq/jq tostring for the JSON-compatible values used by
// normalized message.content: strings stay unquoted; structured values use
// compact JSON serialization.
func Tostring(value interface{}) (string, error) {
	if text, ok := value.(string); ok {
		return text, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("canonical content tostring: %w", err)
	}
	return string(encoded), nil
}

func toolInput(raw json.RawMessage) map[string]interface{} {
	input := map[string]interface{}{}
	if len(raw) == 0 {
		return input
	}
	if err := json.Unmarshal(raw, &input); err != nil || input == nil {
		return map[string]interface{}{}
	}
	return input
}

func hasUsage(usage conversation.TokenUsage) bool {
	return usage.InputTokens != 0 || usage.OutputTokens != 0 || usage.CacheTokens != 0
}
