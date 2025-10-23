package query

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestBuildSuccessfulPrompts(t *testing.T) {
	entries := []parser.SessionEntry{
		{Type: "user", UUID: "1", Message: &parser.Message{Role: "user", Content: []parser.ContentBlock{{Type: "text", Text: "Please fix bug"}}}},
		{Type: "assistant", UUID: "2", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "tool_use", ToolUse: &parser.ToolUse{Name: "Bash", Input: map[string]interface{}{"command": "ls"}}}}}},
		{Type: "user", UUID: "3", Message: &parser.Message{Role: "user", Content: []parser.ContentBlock{{Type: "tool_result", ToolResult: &parser.ToolResult{ToolUseID: "tool-1", Content: "output"}}}}},
		{Type: "assistant", UUID: "4", Message: &parser.Message{Role: "assistant", Content: []parser.ContentBlock{{Type: "text", Text: "Completed task"}}}},
	}

	result := BuildSuccessfulPrompts(entries, 0.0, 10)
	if len(result.Prompts) == 0 {
		t.Fatal("expected at least one successful prompt")
	}
}
