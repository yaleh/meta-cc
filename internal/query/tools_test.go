package query

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

// mockSessionLoader is a test double implementing SessionLoader.
type mockSessionLoader struct {
	entries   []types.SessionEntry
	toolCalls []types.ToolCall
	turnIndex map[string]int
}

func (m *mockSessionLoader) Entries() []types.SessionEntry {
	return m.entries
}

func (m *mockSessionLoader) ExtractToolCalls() []types.ToolCall {
	return m.toolCalls
}

func (m *mockSessionLoader) BuildTurnIndex() map[string]int {
	return m.turnIndex
}

func TestRunToolsQuery_FiltersAndPagination(t *testing.T) {
	loader := &mockSessionLoader{
		toolCalls: []types.ToolCall{
			{ToolName: "Grep", UUID: "tool-1", Timestamp: "2025-10-02T10:00:00.000Z"},
			{ToolName: "Read", UUID: "tool-2", Timestamp: "2025-10-02T10:02:00.000Z", Status: "error", Error: "file not found"},
		},
	}

	opts := ToolsQueryOptions{
		Status: "error",
		Limit:  1,
	}

	calls, err := RunToolsQuery(loader, opts)
	if err != nil {
		t.Fatalf("RunToolsQuery failed: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 call after limit/filter, got %d", len(calls))
	}

	if calls[0].ToolName != "Read" {
		t.Fatalf("expected Read tool, got %s", calls[0].ToolName)
	}
}

func TestRunToolsQuery_FilterByToolName(t *testing.T) {
	loader := &mockSessionLoader{
		toolCalls: []types.ToolCall{
			{ToolName: "Bash", UUID: "tool-1", Timestamp: "2025-10-02T10:00:00.000Z"},
			{ToolName: "Read", UUID: "tool-2", Timestamp: "2025-10-02T10:01:00.000Z"},
		},
	}

	opts := ToolsQueryOptions{
		Tool: "Bash",
	}

	calls, err := RunToolsQuery(loader, opts)
	if err != nil {
		t.Fatalf("RunToolsQuery failed: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}

	if calls[0].ToolName != "Bash" {
		t.Fatalf("expected Bash tool, got %s", calls[0].ToolName)
	}
}

func TestRunToolsQuery_WhereLike(t *testing.T) {
	loader := &mockSessionLoader{
		toolCalls: []types.ToolCall{
			{ToolName: "meta-cc-run", UUID: "tool-1", Timestamp: "2025-10-02T10:00:00.000Z"},
		},
	}

	opts := ToolsQueryOptions{
		Where: "tool LIKE 'meta-cc%'",
	}

	calls, err := RunToolsQuery(loader, opts)
	if err != nil {
		t.Fatalf("RunToolsQuery failed: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].ToolName != "meta-cc-run" {
		t.Fatalf("expected meta-cc-run tool, got %s", calls[0].ToolName)
	}
}
