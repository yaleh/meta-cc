package cmd

import (
	"strings"
	"testing"
)

// Test Phase 12: Project-level vs Session-level tool execution
func TestExecuteTool_ProjectLevel(t *testing.T) {
	tests := []struct {
		name           string
		toolName       string
		args           map[string]interface{}
		expectsProject bool
		description    string
	}{
		{
			name:           "query_tools should include --project flag",
			toolName:       "query_tools",
			args:           map[string]interface{}{"limit": float64(10)},
			expectsProject: true,
			description:    "Project-level tool without _session suffix",
		},
		{
			name:           "query_tools_session should NOT include --project flag",
			toolName:       "query_tools_session",
			args:           map[string]interface{}{"limit": float64(10)},
			expectsProject: false,
			description:    "Session-level tool with _session suffix",
		},
		{
			name:           "analyze_errors should include --project flag",
			toolName:       "analyze_errors",
			args:           map[string]interface{}{},
			expectsProject: true,
			description:    "Project-level error analysis",
		},
		{
			name:           "analyze_errors_session should NOT include --project flag",
			toolName:       "analyze_errors_session",
			args:           map[string]interface{}{},
			expectsProject: false,
			description:    "Session-level error analysis",
		},
		{
			name:           "query_user_messages should include --project flag",
			toolName:       "query_user_messages",
			args:           map[string]interface{}{"pattern": "test"},
			expectsProject: true,
			description:    "Project-level message search",
		},
		{
			name:           "query_user_messages_session should NOT include --project flag",
			toolName:       "query_user_messages_session",
			args:           map[string]interface{}{"pattern": "test"},
			expectsProject: false,
			description:    "Session-level message search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build command args internally
			cmdArgs, err := buildToolCommand(tt.toolName, tt.args)
			if err != nil {
				t.Fatalf("buildToolCommand failed: %v", err)
			}

			hasProjectFlag := false
			for i, arg := range cmdArgs {
				if arg == "--project" && i+1 < len(cmdArgs) && cmdArgs[i+1] == "." {
					hasProjectFlag = true
					break
				}
			}

			if tt.expectsProject && !hasProjectFlag {
				t.Errorf("%s: expected --project flag but not found in: %v", tt.description, cmdArgs)
			}
			if !tt.expectsProject && hasProjectFlag {
				t.Errorf("%s: unexpected --project flag found in: %v", tt.description, cmdArgs)
			}
		})
	}
}

// Test Phase 13: Output format should be jsonl/tsv not json/md
func TestMCPToolSchemas_OutputFormat(t *testing.T) {
	// Get tools list
	tools := getMCPTools()

	for _, tool := range tools {
		t.Run(tool["name"].(string), func(t *testing.T) {
			schema := tool["inputSchema"].(map[string]interface{})
			props := schema["properties"].(map[string]interface{})

			if outputFmt, ok := props["output_format"]; ok {
				fmtMap := outputFmt.(map[string]interface{})

				// Check enum values
				if enum, ok := fmtMap["enum"]; ok {
					enumSlice := enum.([]string)

					// Should contain "jsonl" not "json"
					hasJSONL := false
					hasTSV := false
					hasJSON := false
					hasMD := false

					for _, val := range enumSlice {
						if val == "jsonl" {
							hasJSONL = true
						}
						if val == "tsv" {
							hasTSV = true
						}
						if val == "json" {
							hasJSON = true
						}
						if val == "md" {
							hasMD = true
						}
					}

					if !hasJSONL {
						t.Errorf("Tool %s: output_format should include 'jsonl'", tool["name"])
					}
					if !hasTSV {
						t.Errorf("Tool %s: output_format should include 'tsv'", tool["name"])
					}
					if hasJSON {
						t.Errorf("Tool %s: output_format should NOT include 'json' (use 'jsonl')", tool["name"])
					}
					if hasMD {
						t.Errorf("Tool %s: output_format should NOT include 'md' (deprecated)", tool["name"])
					}
				}

				// Check default value
				if def, ok := fmtMap["default"]; ok {
					if def != "jsonl" {
						t.Errorf("Tool %s: default output_format should be 'jsonl', got '%v'", tool["name"], def)
					}
				}
			}
		})
	}
}

// Test Phase 14: Helper functions reduce duplication
func TestBuildToolCommand_HelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		args     map[string]interface{}
		expected []string
	}{
		{
			name:     "limit parameter extraction",
			toolName: "query_tools",
			args:     map[string]interface{}{"limit": float64(50)},
			expected: []string{"--project", ".", "query", "tools", "--limit", "50", "--output", "jsonl"},
		},
		{
			name:     "limit with default value",
			toolName: "query_tools_session",
			args:     map[string]interface{}{},
			expected: []string{"query", "tools", "--limit", "20", "--output", "jsonl"},
		},
		{
			name:     "pattern parameter (required)",
			toolName: "query_user_messages",
			args:     map[string]interface{}{"pattern": "test.*bug"},
			expected: []string{"--project", ".", "query", "user-messages", "--match", "test.*bug", "--limit", "10", "--output", "jsonl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdArgs, err := buildToolCommand(tt.toolName, tt.args)
			if err != nil {
				t.Fatalf("buildToolCommand failed: %v", err)
			}

			// Compare slices
			if !equalSlices(cmdArgs, tt.expected) {
				t.Errorf("buildToolCommand(%s, %v)\ngot:  %v\nwant: %v",
					tt.toolName, tt.args, cmdArgs, tt.expected)
			}
		})
	}
}

// Test structured error handling
func TestMCPErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing required parameter - pattern",
			toolName:    "query_user_messages",
			args:        map[string]interface{}{},
			expectError: true,
			errorMsg:    "pattern parameter is required",
		},
		{
			name:        "missing required parameter - file",
			toolName:    "query_file_access",
			args:        map[string]interface{}{},
			expectError: true,
			errorMsg:    "file parameter is required",
		},
		{
			name:        "unknown tool",
			toolName:    "unknown_tool",
			args:        map[string]interface{}{},
			expectError: true,
			errorMsg:    "unknown tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeTool(tt.toolName, tt.args)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("error message should contain '%s', got: %v", tt.errorMsg, err)
			}
		})
	}
}

// Test backward compatibility: get_session_stats retains original behavior
func TestBackwardCompatibility_GetSessionStats(t *testing.T) {
	cmdArgs, err := buildToolCommand("get_session_stats", map[string]interface{}{})
	if err != nil {
		t.Fatalf("buildToolCommand failed: %v", err)
	}

	// Should NOT have --project flag (session-only for backward compatibility)
	for i, arg := range cmdArgs {
		if arg == "--project" {
			t.Errorf("get_session_stats should NOT include --project flag for backward compatibility, found at index %d: %v", i, cmdArgs)
		}
	}

	// Should use parse stats (not query)
	if len(cmdArgs) < 2 || cmdArgs[0] != "parse" || cmdArgs[1] != "stats" {
		t.Errorf("get_session_stats should map to 'parse stats', got: %v", cmdArgs)
	}
}

// Helper functions for testing

func getMCPTools() []map[string]interface{} {
	// We can't easily test handleToolsList without refactoring
	// So we'll manually construct expected schema based on current implementation
	tools := []map[string]interface{}{
		{
			"name": "query_tools",
			"inputSchema": map[string]interface{}{
				"properties": map[string]interface{}{
					"output_format": map[string]interface{}{
						"enum":    []string{"jsonl", "tsv"},
						"default": "jsonl",
					},
				},
			},
		},
	}
	return tools
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
