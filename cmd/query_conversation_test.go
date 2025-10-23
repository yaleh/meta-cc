package cmd

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/query"
)

func TestBuildConversationTurnsBasic(t *testing.T) {
	entries := []parser.SessionEntry{
		{Type: "user", UUID: "u1", Timestamp: "2025-10-02T10:00:00Z", Message: &parser.Message{Role: "user", Content: []parser.ContentBlock{{Type: "text", Text: "Question"}}}},
		{Type: "assistant", UUID: "a1", Timestamp: "2025-10-02T10:00:01Z", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "Answer"}}}},
	}

	turns, err := query.BuildConversationTurns(entries, query.ConversationOptions{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("BuildConversationTurns failed: %v", err)
	}
	if len(turns) == 0 {
		t.Fatal("expected at least one conversation turn")
	}
}

func TestBuildConversationTurnsPattern(t *testing.T) {
	entries := []parser.SessionEntry{
		{Type: "user", UUID: "u1", Timestamp: "2025-10-02T10:00:00Z", Message: &parser.Message{Role: "user", Content: []parser.ContentBlock{{Type: "text", Text: "error happened"}}}},
		{Type: "assistant", UUID: "a1", Timestamp: "2025-10-02T10:00:01Z", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "Acknowledged"}}}},
	}
	turns, err := query.BuildConversationTurns(entries, query.ConversationOptions{Pattern: "error", PatternTarget: "user"})
	if err != nil {
		t.Fatalf("pattern filter failed: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(turns))
	}
}

func TestBuildConversationTurnsScope(t *testing.T) {
	entries := []parser.SessionEntry{
		{Type: "user", UUID: "u1", Timestamp: "2025-10-02T10:00:00Z", Message: &parser.Message{Role: "user", Content: []parser.ContentBlock{{Type: "text", Text: "Q"}}}},
		{Type: "assistant", UUID: "a1", Timestamp: "2025-10-02T10:05:00Z", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "A"}}}},
	}
	turns, err := query.BuildConversationTurns(entries, query.ConversationOptions{Limit: 10})
	if err != nil {
		t.Fatalf("conversation build failed: %v", err)
	}
	if len(turns) == 0 {
		t.Fatal("expected conversation turns")
	}
}
