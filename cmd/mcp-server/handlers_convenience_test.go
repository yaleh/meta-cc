package main

import (
	"os"
	"testing"
)

// TestHandleQueryUserMessagesContentLengthFiltering tests min/max content length filtering
// Stage 30.3: Content length filtering for query_user_messages
func TestHandleQueryUserMessagesContentLengthFiltering(t *testing.T) {
	// Create test data with user messages of varying content lengths
	testData := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","message":{"content":"hi"}}
{"type":"user","timestamp":"2025-01-01T10:00:01Z","message":{"content":"medium length message here"}}
{"type":"user","timestamp":"2025-01-01T10:00:02Z","message":{"content":"this is a much longer message that exceeds fifty characters in total length for testing purposes"}}
{"type":"user","timestamp":"2025-01-01T10:00:03Z","message":{"content":[{"type":"tool_result","content":"array content"}]}}
`

	projectPath := setupTestSessionDir(t, testData)

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	executor := NewToolExecutor()

	t.Run("min_content_length_only", func(t *testing.T) {
		result, err := executor.ToolExecutor.ExecuteToolQuery("query_session_content", "session", map[string]interface{}{
			"role":               "user",
			"min_content_length": 10,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// "hi" (2 chars) should be excluded; "medium length message here" (25 chars) and the long message should remain
		if len(result.Entries) < 2 {
			t.Errorf("expected at least 2 results with min_content_length=10, got %d", len(result.Entries))
		}
		// "hi" should NOT be in results
		for _, r := range result.Entries {
			if m, ok := r.(map[string]interface{}); ok {
				if msg, ok := m["message"].(map[string]interface{}); ok {
					if content, ok := msg["content"].(string); ok && content == "hi" {
						t.Error("message 'hi' should be excluded by min_content_length=10")
					}
				}
			}
		}
	})

	t.Run("max_content_length_only", func(t *testing.T) {
		result, err := executor.ToolExecutor.ExecuteToolQuery("query_session_content", "session", map[string]interface{}{
			"role":               "user",
			"max_content_length": 30,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// "hi" (2 chars) and "medium length message here" (25 chars) should remain
		if len(result.Entries) != 2 {
			t.Errorf("expected 2 results with max_content_length=30, got %d", len(result.Entries))
		}
	})

	t.Run("both_min_and_max", func(t *testing.T) {
		result, err := executor.ToolExecutor.ExecuteToolQuery("query_session_content", "session", map[string]interface{}{
			"role":               "user",
			"min_content_length": 10,
			"max_content_length": 30,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Only "medium length message here" (25 chars) should match
		if len(result.Entries) != 1 {
			t.Errorf("expected 1 result with min=10, max=30, got %d", len(result.Entries))
		}
	})

	t.Run("neither_length_filter", func(t *testing.T) {
		result, err := executor.ToolExecutor.ExecuteToolQuery("query_session_content", "session", map[string]interface{}{
			"role": "user"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Default content_type is "string", so 3 string messages should be returned
		if len(result.Entries) != 3 {
			t.Errorf("expected 3 results with no length filters, got %d", len(result.Entries))
		}
	})

	t.Run("array_content_type_with_length_filter_returns_error", func(t *testing.T) {
		_, err := executor.ToolExecutor.ExecuteToolQuery("query_session_content", "session", map[string]interface{}{
			"role":               "user",
			"content_type":       "array",
			"min_content_length": 2,
		})
		if err == nil {
			t.Error("expected error when using content length filter with array content_type")
		}
	})

	t.Run("array_content_type_with_max_length_filter_returns_error", func(t *testing.T) {
		_, err := executor.ToolExecutor.ExecuteToolQuery("query_session_content", "session", map[string]interface{}{
			"role":               "user",
			"content_type":       "array",
			"max_content_length": 100,
		})
		if err == nil {
			t.Error("expected error when using content length filter with array content_type")
		}
	})
}

// TestQuerySessionContentSchemaHasContentLengthParams verifies schema includes content length parameters
// TASK-7: query_user_messages replaced by query_session_content
func TestQuerySessionContentSchemaHasContentLengthParams(t *testing.T) {
	tools := getToolDefinitions()

	var userMsgTool *Tool
	for i, tool := range tools {
		if tool.Name == "query_session_content" {
			userMsgTool = &tools[i]
			break
		}
	}

	if userMsgTool == nil {
		t.Fatal("query_session_content tool not found")
	}

	props := userMsgTool.InputSchema.Properties

	// Check min_content_length
	minProp, exists := props["min_content_length"]
	if !exists {
		t.Error("query_session_content schema missing min_content_length parameter")
	} else {
		if minProp.Type != "number" {
			t.Errorf("min_content_length type should be 'number', got '%s'", minProp.Type)
		}
	}

	// Check max_content_length
	maxProp, exists := props["max_content_length"]
	if !exists {
		t.Error("query_session_content schema missing max_content_length parameter")
	} else {
		if maxProp.Type != "number" {
			t.Errorf("max_content_length type should be 'number', got '%s'", maxProp.Type)
		}
	}

	// Check content_type is declared in schema
	contentTypeProp, exists := props["content_type"]
	if !exists {
		t.Error("query_session_content schema missing content_type parameter")
	} else {
		if contentTypeProp.Type != "string" {
			t.Errorf("content_type type should be 'string', got '%s'", contentTypeProp.Type)
		}
	}
}

// TestHandleQueryTools_ToolParamFilters verifies that the "tool" parameter
// (as declared in the schema) actually filters tool calls by name.
// Regression test for bug: schema used "tool" but handler read "tool_name".
func TestHandleQueryTools_ToolParamFilters(t *testing.T) {
	testData := `{"type":"assistant","timestamp":"2025-01-01T10:00:00Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"/foo"}}],"usage":{"input_tokens":10,"output_tokens":5}}}
{"type":"assistant","timestamp":"2025-01-01T10:00:01Z","message":{"content":[{"type":"tool_use","id":"t2","name":"Bash","input":{"command":"ls"}}],"usage":{"input_tokens":10,"output_tokens":5}}}
`
	projectPath := setupTestSessionDir(t, testData)

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if chErr := os.Chdir(originalWd); chErr != nil {
			t.Fatalf("failed to restore working directory: %v", chErr)
		}
	}()
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	executor := NewToolExecutor()

	t.Run("filter_by_tool_param_returns_only_matching", func(t *testing.T) {
		result, err := executor.ToolExecutor.ExecuteToolQuery("query_session_signals", "session", map[string]interface{}{
			"type": "tool_stats",
			"tool": "Read",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Entries) != 1 {
			t.Errorf("expected 1 result when filtering by tool=Read, got %d", len(result.Entries))
		}
	})

	t.Run("no_filter_returns_all", func(t *testing.T) {
		result, err := executor.ToolExecutor.ExecuteToolQuery("query_session_signals", "session", map[string]interface{}{
			"type": "tool_stats",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Entries) != 2 {
			t.Errorf("expected 2 results with no filter, got %d", len(result.Entries))
		}
	})
}
