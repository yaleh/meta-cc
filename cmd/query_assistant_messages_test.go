package cmd

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/query"
)

func TestBuildAssistantMessagesBasic(t *testing.T) {
	entries := []parser.SessionEntry{
		{
			Type:      "assistant",
			UUID:      "a1",
			Timestamp: "2025-10-02T10:00:00Z",
			Message:   &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "Completed task"}}},
		},
	}

	messages, err := query.BuildAssistantMessages(entries, query.AssistantMessagesOptions{MinTools: -1, MaxTools: -1, MinTokens: -1, MinLength: -1, MaxLength: -1})
	if err != nil {
		t.Fatalf("BuildAssistantMessages failed: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
}

func TestBuildAssistantMessagesPattern(t *testing.T) {
	entries := []parser.SessionEntry{
		{Type: "assistant", UUID: "a1", Timestamp: "2025-10-02T10:00:00Z", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "error occurred"}}}},
		{Type: "assistant", UUID: "a2", Timestamp: "2025-10-02T10:01:00Z", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "success"}}}},
	}
	messages, err := query.BuildAssistantMessages(entries, query.AssistantMessagesOptions{Pattern: "error", MinTools: -1, MaxTools: -1, MinTokens: -1, MinLength: -1, MaxLength: -1})
	if err != nil {
		t.Fatalf("pattern filter failed: %v", err)
	}
	if len(messages) != 1 || messages[0].UUID != "a1" {
		t.Fatalf("expected only error message, got %+v", messages)
	}
}

func TestBuildAssistantMessagesToolCount(t *testing.T) {
	entries := []parser.SessionEntry{
		{Type: "assistant", UUID: "a1", Timestamp: "2025-10-02T10:00:00Z", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "tool_use", ToolUse: &parser.ToolUse{Name: "Read", Input: map[string]interface{}{}}}}}},
		{Type: "assistant", UUID: "a2", Timestamp: "2025-10-02T10:01:00Z", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "no tools"}}}},
	}
	messages, err := query.BuildAssistantMessages(entries, query.AssistantMessagesOptions{MinTools: 1, MaxTools: -1, MinTokens: -1, MinLength: -1, MaxLength: -1})
	if err != nil {
		t.Fatalf("tool filter failed: %v", err)
	}
	if len(messages) != 1 || messages[0].UUID != "a1" {
		t.Fatalf("expected only tool message, got %+v", messages)
	}
}

func TestBuildAssistantMessagesTokens(t *testing.T) {
	entries := []parser.SessionEntry{
		{Type: "assistant", UUID: "a1", Timestamp: "2025-10-02T10:00:00Z", Message: &parser.Message{Role: "assistant", Usage: map[string]interface{}{"output_tokens": 100.0}, Content: []parser.ContentBlock{{Type: "text", Text: "short"}}}},
		{Type: "assistant", UUID: "a2", Timestamp: "2025-10-02T10:01:00Z", Message: &parser.Message{Role: "assistant", Usage: map[string]interface{}{"output_tokens": 1000.0}, Content: []parser.ContentBlock{{Type: "text", Text: "long"}}}},
	}
	messages, err := query.BuildAssistantMessages(entries, query.AssistantMessagesOptions{MinTokens: 500, MinTools: -1, MaxTools: -1, MinLength: -1, MaxLength: -1})
	if err != nil {
		t.Fatalf("token filter failed: %v", err)
	}
	if len(messages) != 1 || messages[0].UUID != "a2" {
		t.Fatalf("expected only high token message, got %+v", messages)
	}
}
