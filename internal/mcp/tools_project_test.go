package mcp

import (
	"strings"
	"testing"
)

func TestToolQueryTools_Definition(t *testing.T) {
	tool := GetProjectLevelTool("query_tools")
	if tool == nil {
		t.Fatal("query_tools tool not found")
	}

	if tool.Name != "query_tools" {
		t.Errorf("Expected name 'query_tools', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Description should not be empty")
	}

	// Verify schema has properties
	schema, ok := tool.InputSchema.(map[string]interface{})
	if !ok {
		t.Fatal("InputSchema should be a map")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have 'properties' field")
	}

	// Verify expected parameters
	expectedParams := []string{"limit", "tool", "status", "where", "output_format"}
	for _, param := range expectedParams {
		if _, exists := props[param]; !exists {
			t.Errorf("Schema missing parameter: %s", param)
		}
	}
}

func TestToolQueryUserMessages_Definition(t *testing.T) {
	tool := GetProjectLevelTool("query_user_messages")
	if tool == nil {
		t.Fatal("query_user_messages tool not found")
	}

	if tool.Name != "query_user_messages" {
		t.Errorf("Expected name 'query_user_messages', got '%s'", tool.Name)
	}

	schema, ok := tool.InputSchema.(map[string]interface{})
	if !ok {
		t.Fatal("InputSchema should be a map")
	}

	// Verify 'pattern' is required
	required, ok := schema["required"].([]interface{})
	if !ok || len(required) == 0 {
		t.Error("Schema should have 'required' field with 'pattern'")
	}

	if required[0] != "pattern" {
		t.Errorf("Expected required field 'pattern', got '%v'", required[0])
	}
}

func TestToolGetStats_Definition(t *testing.T) {
	tool := GetProjectLevelTool("get_stats")
	if tool == nil {
		t.Fatal("get_stats tool not found")
	}

	if tool.Name != "get_stats" {
		t.Errorf("Expected name 'get_stats', got '%s'", tool.Name)
	}

	_, ok := tool.InputSchema.(map[string]interface{})
	if !ok {
		t.Fatal("InputSchema should be a map")
	}
}

func TestToolAnalyzeErrors_Definition(t *testing.T) {
	tool := GetProjectLevelTool("analyze_errors")
	if tool == nil {
		t.Fatal("analyze_errors tool not found")
	}

	if tool.Name != "analyze_errors" {
		t.Errorf("Expected name 'analyze_errors', got '%s'", tool.Name)
	}
}

func TestToolQueryToolSequences_Definition(t *testing.T) {
	tool := GetProjectLevelTool("query_tool_sequences")
	if tool == nil {
		t.Fatal("query_tool_sequences tool not found")
	}

	if tool.Name != "query_tool_sequences" {
		t.Errorf("Expected name 'query_tool_sequences', got '%s'", tool.Name)
	}

	schema, ok := tool.InputSchema.(map[string]interface{})
	if !ok {
		t.Fatal("InputSchema should be a map")
	}

	props := schema["properties"].(map[string]interface{})
	if _, exists := props["min_occurrences"]; !exists {
		t.Error("Schema should have 'min_occurrences' parameter")
	}
}

func TestToolQueryFileAccess_Definition(t *testing.T) {
	tool := GetProjectLevelTool("query_file_access")
	if tool == nil {
		t.Fatal("query_file_access tool not found")
	}

	if tool.Name != "query_file_access" {
		t.Errorf("Expected name 'query_file_access', got '%s'", tool.Name)
	}

	schema, ok := tool.InputSchema.(map[string]interface{})
	if !ok {
		t.Fatal("InputSchema should be a map")
	}

	// Verify 'file' is required
	required := schema["required"].([]interface{})
	if required[0] != "file" {
		t.Errorf("Expected required field 'file', got '%v'", required[0])
	}
}

func TestToolQuerySuccessfulPrompts_Definition(t *testing.T) {
	tool := GetProjectLevelTool("query_successful_prompts")
	if tool == nil {
		t.Fatal("query_successful_prompts tool not found")
	}

	if tool.Name != "query_successful_prompts" {
		t.Errorf("Expected name 'query_successful_prompts', got '%s'", tool.Name)
	}

	schema, ok := tool.InputSchema.(map[string]interface{})
	if !ok {
		t.Fatal("InputSchema should be a map")
	}

	props := schema["properties"].(map[string]interface{})
	if _, exists := props["min_quality_score"]; !exists {
		t.Error("Schema should have 'min_quality_score' parameter")
	}
}

func TestToolQueryContext_Definition(t *testing.T) {
	tool := GetProjectLevelTool("query_context")
	if tool == nil {
		t.Fatal("query_context tool not found")
	}

	if tool.Name != "query_context" {
		t.Errorf("Expected name 'query_context', got '%s'", tool.Name)
	}

	schema, ok := tool.InputSchema.(map[string]interface{})
	if !ok {
		t.Fatal("InputSchema should be a map")
	}

	// Verify 'error_signature' is required
	required := schema["required"].([]interface{})
	if required[0] != "error_signature" {
		t.Errorf("Expected required field 'error_signature', got '%v'", required[0])
	}
}

func TestListProjectLevelTools(t *testing.T) {
	tools := ListProjectLevelTools()

	expectedTools := []string{
		"query_tools",
		"query_user_messages",
		"get_stats",
		"analyze_errors",
		"query_tool_sequences",
		"query_file_access",
		"query_successful_prompts",
		"query_context",
	}

	// Build a map of tool names for quick lookup
	toolMap := make(map[string]bool)
	for _, tool := range tools {
		toolMap[tool.Name] = true
	}

	for _, toolName := range expectedTools {
		if !toolMap[toolName] {
			t.Errorf("Tool '%s' not registered", toolName)
		}
	}

	// Verify total count
	if len(tools) < len(expectedTools) {
		t.Errorf("Expected at least %d tools, got %d", len(expectedTools), len(tools))
	}
}

func TestProjectToolsHaveNoSessionSuffix(t *testing.T) {
	tools := ListProjectLevelTools()

	for _, tool := range tools {
		if strings.HasSuffix(tool.Name, "_session") {
			t.Errorf("Project-level tool '%s' should NOT have '_session' suffix", tool.Name)
		}
	}
}

func TestProjectToolsHaveProperDescriptions(t *testing.T) {
	tools := ListProjectLevelTools()

	for _, tool := range tools {
		if tool.Description == "" {
			t.Errorf("Tool '%s' has empty description", tool.Name)
		}

		// Project-level tools should mention cross-session or project scope
		lowerDesc := strings.ToLower(tool.Description)
		hasProjectScope := strings.Contains(lowerDesc, "all sessions") ||
			strings.Contains(lowerDesc, "project") ||
			strings.Contains(lowerDesc, "across sessions")

		if !hasProjectScope {
			t.Errorf("Tool '%s' description should indicate project-level scope: %s", tool.Name, tool.Description)
		}
	}
}

func TestBuildProjectLevelCommandArgs(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		args     map[string]interface{}
		expected []string
	}{
		{
			name:     "get_stats with default output",
			toolName: "get_stats",
			args:     map[string]interface{}{},
			expected: []string{"parse", "stats", "--project", ".", "--output", "json"},
		},
		{
			name:     "query_tools with filters",
			toolName: "query_tools",
			args: map[string]interface{}{
				"tool":   "Bash",
				"status": "error",
				"limit":  float64(10),
			},
			expected: []string{"query", "tools", "--project", ".", "--output", "json", "--tool", "Bash", "--status", "error", "--limit", "10"},
		},
		{
			name:     "query_user_messages with pattern",
			toolName: "query_user_messages",
			args: map[string]interface{}{
				"pattern": "test.*pattern",
				"limit":   float64(5),
			},
			expected: []string{"query", "user-messages", "--project", ".", "--match", "test.*pattern", "--output", "json", "--limit", "5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildProjectLevelCommandArgs(tt.toolName, tt.args)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d args, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, arg := range tt.expected {
				if result[i] != arg {
					t.Errorf("Arg %d: expected '%s', got '%s'", i, arg, result[i])
				}
			}
		})
	}
}

func TestBuildProjectLevelCommandArgsIncludesProjectFlag(t *testing.T) {
	toolNames := []string{
		"get_stats",
		"analyze_errors",
		"query_tools",
		"query_user_messages",
		"query_context",
		"query_tool_sequences",
		"query_file_access",
		"query_successful_prompts",
	}

	for _, toolName := range toolNames {
		args := map[string]interface{}{
			"output_format": "json",
		}

		// Add required parameters for specific tools
		switch toolName {
		case "query_user_messages":
			args["pattern"] = "test"
		case "query_context":
			args["error_signature"] = "test-sig"
		case "query_file_access":
			args["file"] = "test.go"
		}

		result := BuildProjectLevelCommandArgs(toolName, args)

		// Verify --project . flag is present
		hasProjectFlag := false
		for i := 0; i < len(result)-1; i++ {
			if result[i] == "--project" && result[i+1] == "." {
				hasProjectFlag = true
				break
			}
		}

		if !hasProjectFlag {
			t.Errorf("Tool '%s' command args missing '--project .' flag: %v", toolName, result)
		}
	}
}
