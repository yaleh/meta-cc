package parser

import (
	"testing"
)

func TestExtractToolCalls_SingleCall(t *testing.T) {
	entries := []SessionEntry{
		{
			Type: "assistant",
			UUID: "entry1",
			Message: &Message{
				Role: "assistant",
				Content: []ContentBlock{
					{Type: "text", Text: "检查代码"},
					{
						Type: "tool_use",
						ToolUse: &ToolUse{
							ID:   "toolu_01",
							Name: "Grep",
							Input: map[string]interface{}{
								"pattern": "auth.*error",
							},
						},
					},
				},
			},
		},
		{
			Type: "user",
			UUID: "entry2",
			Message: &Message{
				Role: "user",
				Content: []ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &ToolResult{
							ToolUseID: "toolu_01",
							Content:   "auth.js:15: authError",
						},
					},
				},
			},
		},
	}

	toolCalls := ExtractToolCalls(entries)

	if len(toolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(toolCalls))
	}

	tc := toolCalls[0]
	if tc.ToolName != "Grep" {
		t.Errorf("Expected tool name 'Grep', got '%s'", tc.ToolName)
	}

	if tc.UUID != "entry1" {
		t.Errorf("Expected UUID 'entry1', got '%s'", tc.UUID)
	}

	if tc.Output != "auth.js:15: authError" {
		t.Errorf("Unexpected output: %s", tc.Output)
	}

	pattern, ok := tc.Input["pattern"].(string)
	if !ok || pattern != "auth.*error" {
		t.Errorf("Expected pattern 'auth.*error', got '%v'", pattern)
	}
}

func TestExtractToolCalls_MultipleCallsSameEntry(t *testing.T) {
	entries := []SessionEntry{
		{
			Type: "assistant",
			UUID: "entry1",
			Message: &Message{
				Role: "assistant",
				Content: []ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &ToolUse{
							ID:    "tool_1",
							Name:  "Read",
							Input: map[string]interface{}{"file": "a.txt"},
						},
					},
					{
						Type: "tool_use",
						ToolUse: &ToolUse{
							ID:    "tool_2",
							Name:  "Grep",
							Input: map[string]interface{}{"pattern": "error"},
						},
					},
				},
			},
		},
		{
			Type: "user",
			UUID: "entry2",
			Message: &Message{
				Role: "user",
				Content: []ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &ToolResult{
							ToolUseID: "tool_1",
							Content:   "file content",
						},
					},
					{
						Type: "tool_result",
						ToolResult: &ToolResult{
							ToolUseID: "tool_2",
							Content:   "match found",
						},
					},
				},
			},
		},
	}

	toolCalls := ExtractToolCalls(entries)

	if len(toolCalls) != 2 {
		t.Fatalf("Expected 2 tool calls, got %d", len(toolCalls))
	}

	// 验证都被正确匹配
	for _, tc := range toolCalls {
		if tc.Output == "" {
			t.Errorf("Tool call %s has empty output", tc.ToolName)
		}
	}
}

func TestExtractToolCalls_UnmatchedToolUse(t *testing.T) {
	entries := []SessionEntry{
		{
			Type: "assistant",
			UUID: "entry1",
			Message: &Message{
				Role: "assistant",
				Content: []ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &ToolUse{
							ID:    "orphan_tool",
							Name:  "Bash",
							Input: map[string]interface{}{},
						},
					},
				},
			},
		},
		// 没有对应的 tool_result
	}

	toolCalls := ExtractToolCalls(entries)

	if len(toolCalls) != 1 {
		t.Fatalf("Expected 1 tool call (unmatched), got %d", len(toolCalls))
	}

	tc := toolCalls[0]
	if tc.Output != "" {
		t.Errorf("Expected empty output for unmatched tool, got '%s'", tc.Output)
	}

	if tc.Status != "" {
		t.Errorf("Expected empty status for unmatched tool, got '%s'", tc.Status)
	}
}

func TestExtractToolCalls_NoToolCalls(t *testing.T) {
	entries := []SessionEntry{
		{
			Type: "user",
			UUID: "entry1",
			Message: &Message{
				Role: "user",
				Content: []ContentBlock{
					{Type: "text", Text: "Hello"},
				},
			},
		},
		{
			Type: "assistant",
			UUID: "entry2",
			Message: &Message{
				Role: "assistant",
				Content: []ContentBlock{
					{Type: "text", Text: "Hi there"},
				},
			},
		},
	}

	toolCalls := ExtractToolCalls(entries)

	if len(toolCalls) != 0 {
		t.Errorf("Expected 0 tool calls, got %d", len(toolCalls))
	}
}

// TestExtractToolCallsWithIsError tests that tool calls with is_error=true are correctly extracted
// Verifies that the Error field is properly populated when IsError=true
func TestExtractToolCallsWithIsError(t *testing.T) {
	// Simulating a real session where MCP tool failed with is_error=true
	entries := []SessionEntry{
		{
			Type:      "assistant",
			UUID:      "adac9d46-e2a9-4318-8faa-4b90e8883d00",
			Timestamp: "2025-10-05T00:59:13.857Z",
			Message: &Message{
				Role: "assistant",
				Content: []ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &ToolUse{
							ID:   "toolu_123",
							Name: "mcp__meta_cc__query_user_messages_session",
							Input: map[string]interface{}{
								"limit":         5,
								"output_format": "json",
								"pattern":       ".",
							},
						},
					},
				},
			},
		},
		{
			Type:      "user",
			UUID:      "4634e9c4-5804-4c1e-904d-52cec719e08f",
			Timestamp: "2025-10-05T00:59:14.756Z",
			Message: &Message{
				Role: "user",
				Content: []ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &ToolResult{
							ToolUseID: "toolu_123",
							Content:   "MCP error -32603: Tool execution failed",
							IsError:   true,                                      // Simulating is_error=true from JSONL
							Error:     "MCP error -32603: Tool execution failed", // Should be populated by UnmarshalJSON
						},
					},
				},
			},
		},
	}

	toolCalls := ExtractToolCalls(entries)

	if len(toolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(toolCalls))
	}

	tc := toolCalls[0]

	// Verify basic extraction
	if tc.ToolName != "mcp__meta_cc__query_user_messages_session" {
		t.Errorf("Expected tool name 'mcp__meta_cc__query_user_messages_session', got %q", tc.ToolName)
	}

	if tc.Output != "MCP error -32603: Tool execution failed" {
		t.Errorf("Expected Output to contain error message, got %q", tc.Output)
	}

	// Verify that when is_error=true, Error field is correctly populated
	if tc.Error == "" {
		t.Errorf("Expected Error field to be populated when is_error=true in JSONL, got empty string")
		t.Logf("Current ToolCall: %+v", tc)
		t.FailNow()
	}
}
