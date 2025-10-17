package validation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTools_ValidFile(t *testing.T) {
	// Create temporary test file with sample tool definitions
	content := `package tools

func getToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "test_tool",
			Description: "A test tool",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"param1": map[string]interface{}{
						"type":        "string",
						"description": "First parameter",
					},
					"param2": map[string]interface{}{
						"type":        "number",
						"description": "Second parameter",
					},
				},
				"required": []string{"param1"},
			},
		},
	}
}`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	tools, err := ParseTools(tmpFile)
	if err != nil {
		t.Fatalf("ParseTools failed: %v", err)
	}

	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(tools))
	}

	tool := tools[0]
	if tool.Name != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", tool.Name)
	}

	if tool.Description != "A test tool" {
		t.Errorf("Expected description 'A test tool', got '%s'", tool.Description)
	}

	if len(tool.InputSchema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(tool.InputSchema.Properties))
	}

	if len(tool.InputSchema.Required) != 1 {
		t.Errorf("Expected 1 required parameter, got %d", len(tool.InputSchema.Required))
	}

	if tool.InputSchema.Required[0] != "param1" {
		t.Errorf("Expected required 'param1', got '%s'", tool.InputSchema.Required[0])
	}
}

func TestParseTools_FileNotFound(t *testing.T) {
	_, err := ParseTools("/nonexistent/file.go")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestParseTools_NoFunction(t *testing.T) {
	content := `package tools

// No getToolDefinitions function`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	_, err := ParseTools(tmpFile)
	if err == nil {
		t.Error("Expected error for missing function, got nil")
	}

	if err != nil && err.Error() != "could not find getToolDefinitions() function" {
		t.Errorf("Expected 'could not find getToolDefinitions() function', got '%s'", err.Error())
	}
}

func TestFindClosingBrace_Simple(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "simple brace",
			input:    "{abc}",
			expected: 4,
		},
		{
			name:     "nested braces",
			input:    "{a{b}c}",
			expected: 6,
		},
		{
			name:     "multiple nested",
			input:    "{a{b{c}d}e}",
			expected: 10,
		},
		{
			name:     "no closing",
			input:    "{abc",
			expected: -1,
		},
		{
			name:     "closing before opening",
			input:    "}abc{",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findClosingBrace(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d for input '%s'", tt.expected, result, tt.input)
			}
		})
	}
}

func TestParseProperties_MultipleParams(t *testing.T) {
	toolDef := `{
		Name:        "test_tool",
		Description: "Test",
		InputSchema: map[string]interface{}{
			"properties": map[string]interface{}{
				"filter": {
					Type:        "string",
					Description: "Filter criteria",
				},
				"limit": {
					Type:        "number",
					Description: "Result limit",
				},
			},
		},
	}`

	props := parseProperties(toolDef)

	if len(props) != 2 {
		t.Fatalf("Expected 2 properties, got %d", len(props))
	}

	if props["filter"].Type != "string" {
		t.Errorf("Expected filter type 'string', got '%s'", props["filter"].Type)
	}

	if props["filter"].Description != "Filter criteria" {
		t.Errorf("Expected filter description 'Filter criteria', got '%s'", props["filter"].Description)
	}

	if props["limit"].Type != "number" {
		t.Errorf("Expected limit type 'number', got '%s'", props["limit"].Type)
	}
}

func TestParseProperties_SkipsStandardParams(t *testing.T) {
	toolDef := `{
		properties: {
			"scope": {
				Type:        "string",
				Description: "Standard param",
			},
			"custom_param": {
				Type:        "string",
				Description: "Custom param",
			},
		},
	}`

	props := parseProperties(toolDef)

	// Should skip "scope" as it's a standard parameter
	if _, exists := props["scope"]; exists {
		t.Error("Standard parameter 'scope' should be skipped")
	}

	if _, exists := props["custom_param"]; !exists {
		t.Error("Custom parameter 'custom_param' should be included")
	}
}

func TestParseRequired_MultipleParams(t *testing.T) {
	toolDef := `{
		Required: []string{"param1", "param2", "param3"},
	}`

	required := parseRequired(toolDef)

	if len(required) != 3 {
		t.Fatalf("Expected 3 required params, got %d", len(required))
	}

	expectedParams := map[string]bool{
		"param1": false,
		"param2": false,
		"param3": false,
	}

	for _, param := range required {
		if _, exists := expectedParams[param]; exists {
			expectedParams[param] = true
		} else {
			t.Errorf("Unexpected required parameter: %s", param)
		}
	}

	for param, found := range expectedParams {
		if !found {
			t.Errorf("Expected required parameter not found: %s", param)
		}
	}
}

func TestParseRequired_NoRequired(t *testing.T) {
	toolDef := `{
		Name: "test_tool",
		// No Required field
	}`

	required := parseRequired(toolDef)

	if len(required) != 0 {
		t.Errorf("Expected no required params, got %d", len(required))
	}
}

func TestIsStandardParameter(t *testing.T) {
	tests := []struct {
		param    string
		expected bool
	}{
		{"scope", true},
		{"jq_filter", true},
		{"stats_only", true},
		{"stats_first", true},
		{"inline_threshold_bytes", true},
		{"output_format", true},
		{"custom_param", false},
		{"pattern", false},
		{"tool", false},
	}

	for _, tt := range tests {
		t.Run(tt.param, func(t *testing.T) {
			result := isStandardParameter(tt.param)
			if result != tt.expected {
				t.Errorf("isStandardParameter(%s) = %v, expected %v", tt.param, result, tt.expected)
			}
		})
	}
}

func TestParseToolsFromContent_MultipleTools(t *testing.T) {
	content := `package tools

func getToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "tool1",
			Description: "First tool",
			InputSchema: map[string]interface{}{
				"properties": map[string]interface{}{
					"param1": {
						Type:        "string",
						Description: "Parameter 1",
					},
				},
				"required": []string{"param1"},
			},
		},
		{
			Name:        "tool2",
			Description: "Second tool",
			InputSchema: map[string]interface{}{
				"properties": map[string]interface{}{
					"param2": {
						Type:        "number",
						Description: "Parameter 2",
					},
				},
			},
		},
	}
}`

	tools, err := parseToolsFromContent(content)
	if err != nil {
		t.Fatalf("parseToolsFromContent failed: %v", err)
	}

	if len(tools) != 2 {
		t.Fatalf("Expected 2 tools, got %d", len(tools))
	}

	// Verify first tool
	if tools[0].Name != "tool1" {
		t.Errorf("Expected first tool name 'tool1', got '%s'", tools[0].Name)
	}

	if len(tools[0].InputSchema.Required) != 1 {
		t.Errorf("Expected 1 required param for tool1, got %d", len(tools[0].InputSchema.Required))
	}

	// Verify second tool
	if tools[1].Name != "tool2" {
		t.Errorf("Expected second tool name 'tool2', got '%s'", tools[1].Name)
	}
}

// Helper function to create temporary test file
func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_tools.go")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tmpFile
}
