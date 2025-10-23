package parser

import (
	"encoding/json"
	"strings"
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

// TestToolCallJSONSchema verifies that ToolCall uses snake_case for all JSON fields
// This ensures consistency with Claude Code JSONL schema (Phase 24, Stage 24.1)
func TestToolCallJSONSchema(t *testing.T) {
	toolCall := ToolCall{
		UUID:      "test-uuid-123",
		ToolName:  "Read",
		Input:     map[string]interface{}{"file_path": "/test/file.go"},
		Output:    "file content here",
		Status:    "success",
		Error:     "",
		Timestamp: "2025-10-23T00:00:00Z",
	}

	// Serialize to JSON
	data, err := json.Marshal(toolCall)
	if err != nil {
		t.Fatalf("Failed to marshal ToolCall: %v", err)
	}

	jsonStr := string(data)

	// Verify all fields use snake_case (lowercase with underscores)
	expectedFields := []string{
		"uuid",
		"tool_name",
		"input",
		"output",
		"status",
		"error",
		"timestamp",
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, `"`+field+`"`) {
			t.Errorf("Expected JSON to contain field %q with snake_case, but not found. JSON: %s", field, jsonStr)
		}
	}

	// Verify NO PascalCase fields exist
	forbiddenFields := []string{
		"UUID",
		"ToolName",
		"Input",
		"Output",
		"Status",
		"Error",
		"Timestamp",
	}

	for _, field := range forbiddenFields {
		if strings.Contains(jsonStr, `"`+field+`"`) {
			t.Errorf("JSON should NOT contain PascalCase field %q. Found in: %s", field, jsonStr)
		}
	}

	// Verify deserialization works with snake_case
	var decoded ToolCall
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal ToolCall: %v", err)
	}

	// Verify values are preserved
	if decoded.UUID != toolCall.UUID {
		t.Errorf("UUID mismatch after unmarshal: got %q, want %q", decoded.UUID, toolCall.UUID)
	}
	if decoded.ToolName != toolCall.ToolName {
		t.Errorf("ToolName mismatch after unmarshal: got %q, want %q", decoded.ToolName, toolCall.ToolName)
	}
	if decoded.Status != toolCall.Status {
		t.Errorf("Status mismatch after unmarshal: got %q, want %q", decoded.Status, toolCall.Status)
	}
}
