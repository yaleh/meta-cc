package query

import (
	"fmt"

	"github.com/yaleh/meta-cc/internal/parser"
)

// MessageView represents a flattened message view
type MessageView struct {
	UUID          string                `json:"uuid"`
	SessionID     string                `json:"session_id"`
	ParentUUID    string                `json:"parent_uuid"`
	Timestamp     string                `json:"timestamp"`
	Role          string                `json:"role"`
	Content       string                `json:"content,omitempty"` // Simplified text content
	ContentBlocks []parser.ContentBlock `json:"content_blocks"`    // Full content blocks
	GitBranch     string                `json:"git_branch,omitempty"`
}

// SelectResource selects the appropriate resource view based on resource type
// Returns interface{} slice where each element is of the appropriate type:
// - "entries": []parser.SessionEntry
// - "messages": []MessageView
// - "tools": []parser.ToolCall
func SelectResource(entries []parser.SessionEntry, resource string) (interface{}, error) {
	switch resource {
	case "entries":
		return entries, nil
	case "messages":
		return extractMessages(entries), nil
	case "tools":
		return extractToolExecutions(entries), nil
	default:
		return nil, fmt.Errorf("unknown resource type: %s", resource)
	}
}

// extractMessages extracts all messages (user/assistant entries) and returns MessageView slice
func extractMessages(entries []parser.SessionEntry) []MessageView {
	var messages []MessageView

	for _, entry := range entries {
		// Skip entries without Message
		if entry.Message == nil {
			continue
		}

		// Only process user and assistant messages
		if entry.Type != "user" && entry.Type != "assistant" {
			continue
		}

		msg := MessageView{
			UUID:          entry.UUID,
			SessionID:     entry.SessionID,
			ParentUUID:    entry.ParentUUID,
			Timestamp:     entry.Timestamp,
			Role:          entry.Message.Role,
			ContentBlocks: entry.Message.Content,
			GitBranch:     entry.GitBranch,
		}

		// Extract simplified text content
		msg.Content = extractTextContent(entry.Message.Content)

		messages = append(messages, msg)
	}

	return messages
}

// extractTextContent extracts text from content blocks
func extractTextContent(blocks []parser.ContentBlock) string {
	var text string
	for _, block := range blocks {
		if block.Type == "text" && block.Text != "" {
			if text != "" {
				text += "\n"
			}
			text += block.Text
		}
	}
	return text
}

// extractToolExecutions extracts all tool executions using the existing parser.ExtractToolCalls
func extractToolExecutions(entries []parser.SessionEntry) []parser.ToolCall {
	return parser.ExtractToolCalls(entries)
}
