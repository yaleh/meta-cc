package cmd

import (
	"strings"
	"testing"
)

// Test Phase 12 Revision: Scope parameter replaces _session suffix
func TestExecuteTool_ScopeParameter(t *testing.T) {
	tests := []struct {
		name           string
		toolName       string
		args           map[string]interface{}
		expectsProject bool
		description    string
	}{
		{
			name:           "query_tools default scope is session",
			toolName:       "query_tools",
			args:           map[string]interface{}{"limit": float64(10)},
			expectsProject: false,
			description:    "Default scope is session for backward compatibility",
		},
		{
			name:           "query_tools with scope=project should include --project flag",
			toolName:       "query_tools",
			args:           map[string]interface{}{"limit": float64(10), "scope": "project"},
			expectsProject: true,
			description:    "Explicit project scope adds --project flag",
		},
		{
			name:           "query_tools with scope=session should NOT include --project flag",
			toolName:       "query_tools",
			args:           map[string]interface{}{"limit": float64(10), "scope": "session"},
			expectsProject: false,
			description:    "Explicit session scope without --project flag",
		},
		{
			name:           "analyze_errors default scope is session",
			toolName:       "analyze_errors",
			args:           map[string]interface{}{},
			expectsProject: false,
			description:    "Default scope is session",
		},
		{
			name:           "analyze_errors with scope=project",
			toolName:       "analyze_errors",
			args:           map[string]interface{}{"scope": "project"},
			expectsProject: true,
			description:    "Project scope adds --project flag",
		},
		{
			name:           "query_user_messages with scope=project",
			toolName:       "query_user_messages",
			args:           map[string]interface{}{"pattern": "test", "scope": "project"},
			expectsProject: true,
			description:    "Project scope for message search",
		},
		{
			name:           "query_user_messages default scope is session",
			toolName:       "query_user_messages",
			args:           map[string]interface{}{"pattern": "test"},
			expectsProject: false,
			description:    "Default to session scope",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build command args internally
			cmdArgs, err := buildToolCommandInternal(tt.toolName, tt.args)
			if err != nil {
				t.Fatalf("buildToolCommandInternal failed: %v", err)
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

// Test Phase 14: Command builder with scope parameter
func TestBuildToolCommand_HelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		args     map[string]interface{}
		expected []string
	}{
		{
			name:     "limit parameter extraction with default session scope",
			toolName: "query_tools",
			args:     map[string]interface{}{"limit": float64(50)},
			expected: []string{"query", "tools", "--limit", "50", "--output", "jsonl"},
		},
		{
			name:     "limit parameter with explicit project scope",
			toolName: "query_tools",
			args:     map[string]interface{}{"limit": float64(50), "scope": "project"},
			expected: []string{"--project", ".", "query", "tools", "--limit", "50", "--output", "jsonl"},
		},
		{
			name:     "limit with default value (session scope)",
			toolName: "query_tools",
			args:     map[string]interface{}{},
			expected: []string{"query", "tools", "--limit", "20", "--output", "jsonl"},
		},
		{
			name:     "pattern parameter with default session scope",
			toolName: "query_user_messages",
			args:     map[string]interface{}{"pattern": "test.*bug"},
			expected: []string{"query", "user-messages", "--match", "test.*bug", "--limit", "10", "--output", "jsonl"},
		},
		{
			name:     "pattern parameter with explicit project scope",
			toolName: "query_user_messages",
			args:     map[string]interface{}{"pattern": "test.*bug", "scope": "project"},
			expected: []string{"--project", ".", "query", "user-messages", "--match", "test.*bug", "--limit", "10", "--output", "jsonl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdArgs, err := buildToolCommandInternal(tt.toolName, tt.args)
			if err != nil {
				t.Fatalf("buildToolCommandInternal failed: %v", err)
			}

			// Compare slices
			if !equalSlices(cmdArgs, tt.expected) {
				t.Errorf("buildToolCommandInternal(%s, %v)\ngot:  %v\nwant: %v",
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
	cmdArgs, err := buildToolCommandInternal("get_session_stats", map[string]interface{}{})
	if err != nil {
		t.Fatalf("buildToolCommandInternal failed: %v", err)
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
	// Use the actual consolidated tool list from Phase 12 revision
	return getConsolidatedToolsList()
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
