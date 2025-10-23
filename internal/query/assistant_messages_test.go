package query

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestBuildAssistantMessagesPattern(t *testing.T) {
	entries := []parser.SessionEntry{
		{Type: "assistant", UUID: "1", Timestamp: "2025-10-02T10:00:00Z", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "Completed task"}}}},
	}

	opts := AssistantMessagesOptions{
		MinTools:  -1,
		MaxTools:  -1,
		MinTokens: -1,
		MinLength: -1,
		MaxLength: -1,
	}
	messages, err := BuildAssistantMessages(entries, opts)
	if err != nil {
		t.Fatalf("BuildAssistantMessages failed: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	opts.Pattern = "Completed"
	messages, err = BuildAssistantMessages(entries, opts)
	if err != nil {
		t.Fatalf("BuildAssistantMessages failed with pattern: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message after pattern filter, got %d", len(messages))
	}
}
