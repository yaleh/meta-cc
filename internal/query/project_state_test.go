package query

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestBuildProjectState(t *testing.T) {
	entries := []parser.SessionEntry{
		{Type: "file-history-snapshot", UUID: "snapshot"},
		{Type: "assistant", UUID: "1", Timestamp: "2025-10-02T10:00:00Z", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "tool_use", ToolUse: &parser.ToolUse{Name: "Read", Input: map[string]interface{}{"file_path": "/tmp/file.txt"}}}}}},
		{Type: "assistant", UUID: "3", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "Completed task"}}}},
	}

	state := BuildProjectState(entries, ProjectStateOptions{IncludeIncomplete: true})
	if state == nil {
		t.Fatal("expected state")
	}
	if len(state.RecentFiles) != 1 {
		t.Fatalf("expected 1 recent file, got %d", len(state.RecentFiles))
	}
	if state.CurrentFocus == "" {
		t.Fatal("expected current focus")
	}
}
