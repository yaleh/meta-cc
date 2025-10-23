package parser

import (
	"encoding/json"
	"testing"
)

// TestToolCallJSONMarshaling verifies that ToolCall struct marshals to JSON with correct field names
func TestToolCallJSONMarshaling(t *testing.T) {
	toolCall := ToolCall{
		UUID:      "test-uuid-123",
		ToolName:  "TestTool",
		Input:     map[string]interface{}{"param": "value"},
		Output:    "test output",
		Status:    "success",
		Error:     "",
		Timestamp: "2025-01-01T00:00:00Z",
	}

	jsonBytes, err := json.Marshal(toolCall)
	if err != nil {
		t.Fatalf("Failed to marshal ToolCall: %v", err)
	}

	// Unmarshal to map to verify field names
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify all expected fields are present with correct names (snake_case)
	expectedFields := []string{"uuid", "tool_name", "input", "output", "status", "error", "timestamp"}
	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Expected field %q not found in JSON output", field)
		}
	}

	// Verify field values
	if result["uuid"] != "test-uuid-123" {
		t.Errorf("uuid mismatch: got %v, want test-uuid-123", result["uuid"])
	}
	if result["tool_name"] != "TestTool" {
		t.Errorf("tool_name mismatch: got %v, want TestTool", result["tool_name"])
	}
	if result["status"] != "success" {
		t.Errorf("status mismatch: got %v, want success", result["status"])
	}
}

// TestToolCallJSONFieldNamesExplicit verifies JSON tags are explicitly defined
func TestToolCallJSONFieldNamesExplicit(t *testing.T) {
	// This test ensures we have explicit JSON tags, not relying on Go defaults
	// After adding JSON tags, this will verify the tags match expected field names

	toolCall := ToolCall{
		ToolName: "Bash",
		Status:   "error",
	}

	jsonBytes, err := json.Marshal(toolCall)
	if err != nil {
		t.Fatalf("Failed to marshal ToolCall: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Check that JSON contains expected field names (snake_case)
	// This matches the schema documentation in cmd/mcp-server/tools.go
	expectedInJSON := []string{`"uuid"`, `"tool_name"`, `"input"`, `"output"`, `"status"`, `"error"`, `"timestamp"`}
	for _, expected := range expectedInJSON {
		if !containsString(jsonStr, expected) {
			t.Errorf("JSON output missing expected field name %s, got: %s", expected, jsonStr)
		}
	}
}

// TestToolCallJSONUnmarshaling verifies that JSON can be unmarshaled back to ToolCall
func TestToolCallJSONUnmarshaling(t *testing.T) {
	jsonStr := `{
		"uuid": "uuid-456",
		"tool_name": "Read",
		"input": {"file_path": "/tmp/test.txt"},
		"output": "file contents",
		"status": "success",
		"error": "",
		"timestamp": "2025-01-01T12:00:00Z"
	}`

	var toolCall ToolCall
	if err := json.Unmarshal([]byte(jsonStr), &toolCall); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify unmarshaled values
	if toolCall.UUID != "uuid-456" {
		t.Errorf("UUID mismatch: got %s, want uuid-456", toolCall.UUID)
	}
	if toolCall.ToolName != "Read" {
		t.Errorf("ToolName mismatch: got %s, want Read", toolCall.ToolName)
	}
	if toolCall.Status != "success" {
		t.Errorf("Status mismatch: got %s, want success", toolCall.Status)
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
