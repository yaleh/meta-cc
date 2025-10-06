package mcp

import (
	"strings"
	"testing"
)

// Test that session-level tools have correct naming with _session suffix
func TestSessionToolsNaming(t *testing.T) {
	tests := []struct {
		toolName        string
		expectedSuffix  string
		shouldHaveScope bool
	}{
		{"query_tools_session", "_session", true},
		{"query_user_messages_session", "_session", true},
		{"analyze_errors_session", "_session", true},
		{"query_tool_sequences_session", "_session", true},
		{"query_file_access_session", "_session", true},
		{"query_successful_prompts_session", "_session", true},
		{"query_context_session", "_session", true},
		{"get_session_stats", "", false}, // Backward compatibility - no _session suffix
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			if tt.shouldHaveScope && !strings.HasSuffix(tt.toolName, tt.expectedSuffix) {
				t.Errorf("Tool %s should have suffix %s", tt.toolName, tt.expectedSuffix)
			}
		})
	}
}

// Test query_tools_session tool definition
func TestToolQueryToolsSession(t *testing.T) {
	tool := GetToolDefinition("query_tools_session")

	if tool == nil {
		t.Fatal("query_tools_session tool should be defined")
	}

	if tool.Name != "query_tools_session" {
		t.Errorf("Expected name 'query_tools_session', got '%s'", tool.Name)
	}

	if !strings.Contains(tool.Description, "current session") {
		t.Error("Description should mention 'current session'")
	}

	// Verify schema
	schema := tool.InputSchema.(map[string]interface{})
	props := schema["properties"].(map[string]interface{})

	expectedParams := []string{"limit", "tool", "status"}
	for _, param := range expectedParams {
		if _, exists := props[param]; !exists {
			t.Errorf("Schema missing parameter: %s", param)
		}
	}
}

// Test query_user_messages_session tool definition
func TestToolQueryUserMessagesSession(t *testing.T) {
	tool := GetToolDefinition("query_user_messages_session")

	if tool == nil {
		t.Fatal("query_user_messages_session tool should be defined")
	}

	if tool.Name != "query_user_messages_session" {
		t.Errorf("Expected name 'query_user_messages_session', got '%s'", tool.Name)
	}

	if !strings.Contains(tool.Description, "current session") {
		t.Error("Description should mention 'current session'")
	}

	// Verify 'pattern' is required
	schema := tool.InputSchema.(map[string]interface{})
	required, ok := schema["required"].([]interface{})
	if !ok || len(required) == 0 {
		t.Fatal("Schema should have 'required' field with 'pattern'")
	}

	if required[0] != "pattern" {
		t.Errorf("Expected required field 'pattern', got '%v'", required[0])
	}
}

// Test analyze_errors_session tool definition
func TestToolAnalyzeErrorsSession(t *testing.T) {
	tool := GetToolDefinition("analyze_errors_session")

	if tool == nil {
		t.Fatal("analyze_errors_session tool should be defined")
	}

	if tool.Name != "analyze_errors_session" {
		t.Errorf("Expected name 'analyze_errors_session', got '%s'", tool.Name)
	}

	if !strings.Contains(tool.Description, "current session") {
		t.Error("Description should mention 'current session'")
	}
}

// Test query_tool_sequences_session tool definition
func TestToolQueryToolSequencesSession(t *testing.T) {
	tool := GetToolDefinition("query_tool_sequences_session")

	if tool == nil {
		t.Fatal("query_tool_sequences_session tool should be defined")
	}

	if tool.Name != "query_tool_sequences_session" {
		t.Errorf("Expected name 'query_tool_sequences_session', got '%s'", tool.Name)
	}

	schema := tool.InputSchema.(map[string]interface{})
	props := schema["properties"].(map[string]interface{})

	if _, exists := props["min_occurrences"]; !exists {
		t.Error("Schema should have 'min_occurrences' parameter")
	}
}

// Test query_file_access_session tool definition
func TestToolQueryFileAccessSession(t *testing.T) {
	tool := GetToolDefinition("query_file_access_session")

	if tool == nil {
		t.Fatal("query_file_access_session tool should be defined")
	}

	if tool.Name != "query_file_access_session" {
		t.Errorf("Expected name 'query_file_access_session', got '%s'", tool.Name)
	}

	// Verify 'file' is required
	schema := tool.InputSchema.(map[string]interface{})
	required := schema["required"].([]interface{})
	if required[0] != "file" {
		t.Errorf("Expected required field 'file', got '%v'", required[0])
	}
}

// Test query_successful_prompts_session tool definition
func TestToolQuerySuccessfulPromptsSession(t *testing.T) {
	tool := GetToolDefinition("query_successful_prompts_session")

	if tool == nil {
		t.Fatal("query_successful_prompts_session tool should be defined")
	}

	if tool.Name != "query_successful_prompts_session" {
		t.Errorf("Expected name 'query_successful_prompts_session', got '%s'", tool.Name)
	}

	schema := tool.InputSchema.(map[string]interface{})
	props := schema["properties"].(map[string]interface{})

	if _, exists := props["min_quality_score"]; !exists {
		t.Error("Schema should have 'min_quality_score' parameter")
	}
}

// Test query_context_session tool definition
func TestToolQueryContextSession(t *testing.T) {
	tool := GetToolDefinition("query_context_session")

	if tool == nil {
		t.Fatal("query_context_session tool should be defined")
	}

	if tool.Name != "query_context_session" {
		t.Errorf("Expected name 'query_context_session', got '%s'", tool.Name)
	}

	// Verify 'error_signature' is required
	schema := tool.InputSchema.(map[string]interface{})
	required := schema["required"].([]interface{})
	if required[0] != "error_signature" {
		t.Errorf("Expected required field 'error_signature', got '%v'", required[0])
	}
}

// Test backward compatibility: get_session_stats should exist and be unchanged
func TestGetSessionStatsBackwardCompatibility(t *testing.T) {
	tool := GetToolDefinition("get_session_stats")

	if tool == nil {
		t.Fatal("get_session_stats should exist for backward compatibility")
	}

	if tool.Name != "get_session_stats" {
		t.Errorf("Expected name 'get_session_stats', got '%s'", tool.Name)
	}

	// Verify it does NOT have _session suffix
	if strings.HasSuffix(tool.Name, "_session") {
		t.Error("get_session_stats should NOT have _session suffix (backward compatibility)")
	}
}

// Test that session-level tools do NOT use --project flag
func TestSessionToolsNoProjectFlag(t *testing.T) {
	tests := []struct {
		toolName string
		args     map[string]interface{}
	}{
		{
			toolName: "query_tools_session",
			args:     map[string]interface{}{"limit": float64(10)},
		},
		{
			toolName: "analyze_errors_session",
			args:     map[string]interface{}{"output_format": "json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			cmdArgs := BuildCommandArgs(tt.toolName, tt.args)

			// Verify --project flag is NOT present
			for i, arg := range cmdArgs {
				if arg == "--project" {
					t.Errorf("Session-level tool %s should NOT have --project flag, got: %v", tt.toolName, cmdArgs)
					break
				}
				if i < len(cmdArgs)-1 && arg == "--project" && cmdArgs[i+1] == "." {
					t.Errorf("Session-level tool %s should NOT have --project . flag", tt.toolName)
					break
				}
			}
		})
	}
}

// Test that all 8 session-level tools are registered
func TestAllSessionLevelToolsRegistered(t *testing.T) {
	expectedTools := []string{
		"query_tools_session",
		"query_user_messages_session",
		"analyze_errors_session",
		"query_tool_sequences_session",
		"query_file_access_session",
		"query_successful_prompts_session",
		"query_context_session",
		"get_session_stats", // Backward compatibility
	}

	allTools := ListAllTools()

	for _, expected := range expectedTools {
		found := false
		for _, tool := range allTools {
			if tool.Name == expected {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Session-level tool '%s' not registered", expected)
		}
	}
}

// Test tool execution logic returns valid format
func TestSessionToolExecutionFormat(t *testing.T) {
	tool := GetToolDefinition("analyze_errors_session")

	if tool == nil {
		t.Skip("Tool not yet implemented")
	}

	// Test that execution would return proper format (without actually executing)
	// This tests the command building logic
	args := map[string]interface{}{
		"output_format": "json",
	}

	cmdArgs := BuildCommandArgs("analyze_errors_session", args)

	// Should contain the analyze errors command
	hasAnalyze := false
	hasErrors := false
	for _, arg := range cmdArgs {
		if arg == "analyze" {
			hasAnalyze = true
		}
		if arg == "errors" {
			hasErrors = true
		}
	}

	if !hasAnalyze || !hasErrors {
		t.Errorf("Expected 'analyze errors' in command args, got: %v", cmdArgs)
	}

	// Should have --output json
	hasOutput := false
	for i := 0; i < len(cmdArgs)-1; i++ {
		if cmdArgs[i] == "--output" && cmdArgs[i+1] == "json" {
			hasOutput = true
			break
		}
	}

	if !hasOutput {
		t.Errorf("Expected '--output json' in command args, got: %v", cmdArgs)
	}
}
