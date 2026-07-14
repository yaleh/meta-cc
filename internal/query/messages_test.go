package query

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/types"
)

func makeTextEntry(uuid, role, text, timestamp string) types.SessionEntry {
	return types.SessionEntry{
		Type:      role,
		UUID:      uuid,
		Timestamp: timestamp,
		Message: &parser.Message{
			Role:    role,
			Content: []parser.ContentBlock{{Type: "text", Text: text}},
		},
	}
}

func TestRunUserMessagesQuery_PatternAndContext(t *testing.T) {
	entries := []types.SessionEntry{
		makeTextEntry("uuid-1", "user", "Fix bug in parser", "2025-10-02T10:00:00.000Z"),
		makeTextEntry("uuid-2", "assistant", "Sure", "2025-10-02T10:00:10.000Z"),
		makeTextEntry("uuid-3", "user", "Add new feature", "2025-10-02T10:01:00.000Z"),
	}
	turnIndex := map[string]int{
		"uuid-1": 0,
		"uuid-2": 1,
		"uuid-3": 2,
	}

	loader := &mockSessionLoader{
		entries:   entries,
		turnIndex: turnIndex,
	}

	opts := UserMessagesQueryOptions{
		Pattern: "Fix",
		Context: 1,
	}

	msgs, err := RunUserMessagesQuery(loader, opts)
	if err != nil {
		t.Fatalf("RunUserMessagesQuery failed: %v", err)
	}

	if len(msgs) != 1 {
		t.Fatalf("expected 1 user message, got %d", len(msgs))
	}

	if len(msgs[0].ContextAfter) == 0 {
		t.Fatalf("expected context after message")
	}

	if msgs[0].ContextAfter[0].Role != "assistant" {
		t.Fatalf("expected assistant context after, got %s", msgs[0].ContextAfter[0].Role)
	}
}

func TestRunUserMessagesQuery_LimitOffset(t *testing.T) {
	entries := []types.SessionEntry{
		makeTextEntry("uuid-1", "user", "Message 1", "2025-10-02T10:00:00.000Z"),
		makeTextEntry("uuid-2", "user", "Message 2", "2025-10-02T10:01:00.000Z"),
		makeTextEntry("uuid-3", "user", "Message 3", "2025-10-02T10:02:00.000Z"),
	}
	turnIndex := map[string]int{
		"uuid-1": 0,
		"uuid-2": 1,
		"uuid-3": 2,
	}

	loader := &mockSessionLoader{
		entries:   entries,
		turnIndex: turnIndex,
	}

	opts := UserMessagesQueryOptions{
		Offset: 1,
		Limit:  1,
	}

	msgs, err := RunUserMessagesQuery(loader, opts)
	if err != nil {
		t.Fatalf("RunUserMessagesQuery failed: %v", err)
	}

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message after pagination, got %d", len(msgs))
	}

	if msgs[0].Content != "Message 2" {
		t.Fatalf("expected Message 2, got %s", msgs[0].Content)
	}
}
