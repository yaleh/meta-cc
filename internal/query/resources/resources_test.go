package resources_test

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/query/resources"
	"github.com/yaleh/meta-cc/internal/types"
)

type mockLoader struct {
	entries   []types.SessionEntry
	toolCalls []types.ToolCall
	turnIndex map[string]int
}

func (m *mockLoader) Entries() []types.SessionEntry      { return m.entries }
func (m *mockLoader) ExtractToolCalls() []types.ToolCall { return m.toolCalls }
func (m *mockLoader) BuildTurnIndex() map[string]int     { return m.turnIndex }

func TestRunToolsQuery_InResourcesPackage(t *testing.T) {
	loader := &mockLoader{
		toolCalls: []types.ToolCall{
			{ToolName: "Bash", Status: "success", Timestamp: "2025-10-02T10:00:00.000Z"},
			{ToolName: "Read", Status: "error", Error: "file not found", Timestamp: "2025-10-02T10:01:00.000Z"},
		},
	}
	calls, err := resources.RunToolsQuery(loader, types.ToolsQueryOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 2 {
		t.Errorf("expected 2 calls, got %d", len(calls))
	}
}

func TestRunToolsQuery_FilterByStatus(t *testing.T) {
	loader := &mockLoader{
		toolCalls: []types.ToolCall{
			{ToolName: "Bash", Status: "success", Timestamp: "2025-10-02T10:00:00.000Z"},
			{ToolName: "Read", Status: "error", Error: "file not found", Timestamp: "2025-10-02T10:01:00.000Z"},
		},
	}
	calls, err := resources.RunToolsQuery(loader, types.ToolsQueryOptions{Status: "error"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].ToolName != "Read" {
		t.Errorf("expected Read, got %s", calls[0].ToolName)
	}
}

func TestRunToolsQuery_WhereLike(t *testing.T) {
	loader := &mockLoader{
		toolCalls: []types.ToolCall{
			{ToolName: "meta-cc-run", UUID: "tool-1", Timestamp: "2025-10-02T10:00:00.000Z"},
			{ToolName: "Bash", UUID: "tool-2", Timestamp: "2025-10-02T10:01:00.000Z"},
		},
	}
	calls, err := resources.RunToolsQuery(loader, types.ToolsQueryOptions{Where: "tool LIKE 'meta-cc%'"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].ToolName != "meta-cc-run" {
		t.Errorf("expected meta-cc-run, got %s", calls[0].ToolName)
	}
}

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

func TestRunUserMessagesQuery_InResourcesPackage(t *testing.T) {
	entries := []types.SessionEntry{
		makeTextEntry("uuid-1", "user", "Fix the bug", "2025-10-02T10:00:00.000Z"),
		makeTextEntry("uuid-2", "assistant", "Sure, I will fix it", "2025-10-02T10:01:00.000Z"),
	}
	turnIndex := map[string]int{
		"uuid-1": 0,
		"uuid-2": 1,
	}
	loader := &mockLoader{
		entries:   entries,
		turnIndex: turnIndex,
	}
	msgs, err := resources.RunUserMessagesQuery(loader, types.UserMessagesQueryOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 user message, got %d", len(msgs))
	}
	if msgs[0].Content != "Fix the bug" {
		t.Errorf("expected 'Fix the bug', got %q", msgs[0].Content)
	}
}

func TestRunUserMessagesQuery_PatternFilter(t *testing.T) {
	entries := []types.SessionEntry{
		makeTextEntry("uuid-1", "user", "Fix the bug", "2025-10-02T10:00:00.000Z"),
		makeTextEntry("uuid-2", "user", "Add new feature", "2025-10-02T10:01:00.000Z"),
	}
	turnIndex := map[string]int{
		"uuid-1": 0,
		"uuid-2": 1,
	}
	loader := &mockLoader{
		entries:   entries,
		turnIndex: turnIndex,
	}
	msgs, err := resources.RunUserMessagesQuery(loader, types.UserMessagesQueryOptions{Pattern: "Fix"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

func TestRunUserMessagesQuery_Context(t *testing.T) {
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
	loader := &mockLoader{
		entries:   entries,
		turnIndex: turnIndex,
	}
	msgs, err := resources.RunUserMessagesQuery(loader, types.UserMessagesQueryOptions{Pattern: "Fix", Context: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if len(msgs[0].ContextAfter) == 0 {
		t.Fatalf("expected context after message")
	}
	if msgs[0].ContextAfter[0].Role != "assistant" {
		t.Errorf("expected assistant context, got %s", msgs[0].ContextAfter[0].Role)
	}
}
