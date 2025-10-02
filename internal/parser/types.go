package parser

import (
	"encoding/json"
	"fmt"
)

// Turn represents a conversation turn in a Claude Code session.
type Turn struct {
	Sequence  int            `json:"sequence"`
	Role      string         `json:"role"`
	Timestamp int64          `json:"timestamp"`
	Content   []ContentBlock `json:"content"`
}

// ContentBlock represents a content block within a Turn.
// It can be text, a tool use, or a tool result.
type ContentBlock struct {
	Type       string      `json:"type"`
	Text       string      `json:"text,omitempty"`
	ToolUse    *ToolUse    `json:"-"` // Manually handled during unmarshaling
	ToolResult *ToolResult `json:"-"` // Manually handled during unmarshaling
}

// ToolUse represents a tool invocation by the assistant.
type ToolUse struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ToolResult represents the result of a tool invocation.
type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	Status    string `json:"status,omitempty"`
	Error     string `json:"error,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for ContentBlock.
// It handles polymorphic content types: text, tool_use, and tool_result.
func (cb *ContentBlock) UnmarshalJSON(data []byte) error {
	// First, unmarshal common fields
	type Alias ContentBlock
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(cb),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal ContentBlock: %w", err)
	}

	// Parse type-specific fields based on the type
	switch cb.Type {
	case "text":
		// Text type is already handled by default unmarshaling

	case "tool_use":
		// Parse tool_use fields embedded directly in ContentBlock
		type ToolUseBlock struct {
			Type  string                 `json:"type"`
			ID    string                 `json:"id"`
			Name  string                 `json:"name"`
			Input map[string]interface{} `json:"input"`
		}
		var tub ToolUseBlock
		if err := json.Unmarshal(data, &tub); err != nil {
			return fmt.Errorf("failed to unmarshal tool_use: %w", err)
		}
		cb.ToolUse = &ToolUse{
			ID:    tub.ID,
			Name:  tub.Name,
			Input: tub.Input,
		}

	case "tool_result":
		// Parse tool_result fields embedded directly in ContentBlock
		type ToolResultBlock struct {
			Type      string `json:"type"`
			ToolUseID string `json:"tool_use_id"`
			Content   string `json:"content"`
			Status    string `json:"status,omitempty"`
			Error     string `json:"error,omitempty"`
		}
		var trb ToolResultBlock
		if err := json.Unmarshal(data, &trb); err != nil {
			return fmt.Errorf("failed to unmarshal tool_result: %w", err)
		}
		cb.ToolResult = &ToolResult{
			ToolUseID: trb.ToolUseID,
			Content:   trb.Content,
			Status:    trb.Status,
			Error:     trb.Error,
		}

	default:
		// Unknown type - preserve type field but don't error
	}

	return nil
}
