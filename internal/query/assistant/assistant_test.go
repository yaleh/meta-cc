package assistant_test

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/query/assistant"
	"github.com/yaleh/meta-cc/internal/types"
)

func TestBuildAssistantMessagesPattern(t *testing.T) {
	entries := []types.SessionEntry{
		{Type: "assistant", UUID: "1", Timestamp: "2025-10-02T10:00:00Z", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "Completed task"}}}},
	}

	opts := assistant.AssistantMessagesOptions{
		MinTools:  -1,
		MaxTools:  -1,
		MinTokens: -1,
		MinLength: -1,
		MaxLength: -1,
	}
	messages, err := assistant.BuildAssistantMessages(entries, opts)
	if err != nil {
		t.Fatalf("BuildAssistantMessages failed: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	opts.Pattern = "Completed"
	messages, err = assistant.BuildAssistantMessages(entries, opts)
	if err != nil {
		t.Fatalf("BuildAssistantMessages failed with pattern: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message after pattern filter, got %d", len(messages))
	}
}

func TestBuildAssistantMessages_InvalidPattern(t *testing.T) {
	entries := []types.SessionEntry{
		{Type: "assistant", UUID: "1", Timestamp: "2025-10-02T10:00:00Z", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "Hello"}}}},
	}

	opts := assistant.AssistantMessagesOptions{
		Pattern:   "[invalid",
		MinTools:  -1,
		MaxTools:  -1,
		MinTokens: -1,
		MinLength: -1,
		MaxLength: -1,
	}
	_, err := assistant.BuildAssistantMessages(entries, opts)
	if err == nil {
		t.Fatal("Expected error for invalid regex pattern, got nil")
	}
}

func TestBuildConversationTurns(t *testing.T) {
	entries := []types.SessionEntry{
		{
			Type: "user", UUID: "u1", Timestamp: "2025-10-02T10:00:00Z",
			Message: &parser.Message{Role: "user", Content: []parser.ContentBlock{{Type: "text", Text: "Hello"}}},
		},
		{
			Type: "assistant", UUID: "a1", Timestamp: "2025-10-02T10:01:00Z",
			Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "Hi there"}}},
		},
	}

	opts := assistant.ConversationOptions{
		StartTurn:   -1,
		EndTurn:     -1,
		MinDuration: -1,
		MaxDuration: -1,
	}
	turns, err := assistant.BuildConversationTurns(entries, opts)
	if err != nil {
		t.Fatalf("BuildConversationTurns failed: %v", err)
	}
	if len(turns) == 0 {
		t.Fatal("expected at least one conversation turn")
	}
}

func TestBuildConversationTurns_InvalidPattern(t *testing.T) {
	entries := []types.SessionEntry{
		{
			Type: "user", UUID: "u1", Timestamp: "2025-10-02T10:00:00Z",
			Message: &parser.Message{Role: "user", Content: []parser.ContentBlock{{Type: "text", Text: "Hello"}}},
		},
	}

	opts := assistant.ConversationOptions{
		Pattern:     "[invalid",
		StartTurn:   -1,
		EndTurn:     -1,
		MinDuration: -1,
		MaxDuration: -1,
	}
	_, err := assistant.BuildConversationTurns(entries, opts)
	if err == nil {
		t.Fatal("Expected error for invalid regex pattern, got nil")
	}
}
