package main

import (
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/config"
)

// TestQueryToolsDoNotUseCLI verifies that query tools use the internal/query library
// instead of spawning meta-cc subprocess. This test ensures Phase 23 completion:
// all MCP query tools must execute without CLI dependency.
func TestQueryToolsDoNotUseCLI(t *testing.T) {
	// Create executor - no CLI path needed since all tools use library
	executor := &ToolExecutor{}

	cfg := &config.Config{}

	// All query tools that must use library (not CLI)
	queryTools := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "query_tools",
			args: map[string]interface{}{
				"scope": "session",
			},
		},
		{
			name: "query_user_messages",
			args: map[string]interface{}{
				"scope":   "session",
				"pattern": "test",
			},
		},
		{
			name: "query_tool_sequences",
			args: map[string]interface{}{
				"scope": "session",
			},
		},
		{
			name: "query_file_access",
			args: map[string]interface{}{
				"scope": "session",
				"file":  "test.go",
			},
		},
		{
			name: "get_session_stats",
			args: map[string]interface{}{
				"scope": "session",
			},
		},
		{
			name: "query_project_state",
			args: map[string]interface{}{
				"scope": "session",
			},
		},
		{
			name: "query_successful_prompts",
			args: map[string]interface{}{
				"scope": "session",
			},
		},
	}

	for _, tc := range queryTools {
		t.Run(tc.name, func(t *testing.T) {
			_, err := executor.ExecuteTool(cfg, tc.name, tc.args)

			// The tool may fail for legitimate reasons (e.g., session not found),
			// but it must NOT fail because it tried to execute the non-existent CLI binary.
			if err != nil {
				errMsg := err.Error()

				// Check for indicators that CLI was attempted
				cliIndicators := []string{
					"no such file",
					"executable file not found",
					"cannot find",
					"/nonexistent/path/to/meta-cc",
				}

				for _, indicator := range cliIndicators {
					if strings.Contains(errMsg, indicator) {
						t.Fatalf("Tool %s attempted to use CLI (error: %v). "+
							"All query tools must use internal/query library, not spawn meta-cc subprocess.",
							tc.name, err)
					}
				}
			}
		})
	}
}

// TestUnknownToolReturnsError verifies that unknown tools return proper error
// without attempting CLI execution (no fallback to buildCommand/executeMetaCC)
func TestUnknownToolReturnsError(t *testing.T) {
	executor := &ToolExecutor{}

	cfg := &config.Config{}

	unknownTools := []string{
		"unknown_tool",
		"query_nonexistent",
		"invalid_command",
	}

	for _, toolName := range unknownTools {
		t.Run(toolName, func(t *testing.T) {
			args := map[string]interface{}{
				"scope": "session",
			}

			_, err := executor.ExecuteTool(cfg, toolName, args)

			if err == nil {
				t.Errorf("Expected error for unknown tool %s, got nil", toolName)
				return
			}

			errMsg := err.Error()

			// Should get "unknown tool" error, not CLI execution error
			if !strings.Contains(errMsg, "unknown tool") {
				t.Errorf("Expected 'unknown tool' error for %s, got: %v", toolName, err)
			}

			// Must NOT attempt to execute CLI
			cliIndicators := []string{
				"no such file",
				"executable file not found",
				"/nonexistent/path/to/meta-cc",
			}

			for _, indicator := range cliIndicators {
				if strings.Contains(errMsg, indicator) {
					t.Fatalf("Unknown tool %s attempted CLI execution (should return error immediately). "+
						"Error: %v", toolName, err)
				}
			}
		})
	}
}

// TestSpecialToolsDoNotUseCLI verifies that special tools (cleanup, capabilities)
// also don't use CLI subprocess
func TestSpecialToolsDoNotUseCLI(t *testing.T) {
	executor := &ToolExecutor{}

	cfg := &config.Config{}

	specialTools := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "cleanup_temp_files",
			args: map[string]interface{}{
				"max_age_days": 7,
			},
		},
		{
			name: "list_capabilities",
			args: map[string]interface{}{},
		},
		{
			name: "get_capability",
			args: map[string]interface{}{
				"capability_name": "meta-errors",
			},
		},
	}

	for _, tc := range specialTools {
		t.Run(tc.name, func(t *testing.T) {
			_, err := executor.ExecuteTool(cfg, tc.name, tc.args)

			// May fail for other reasons, but not CLI execution
			if err != nil {
				errMsg := err.Error()

				cliIndicators := []string{
					"no such file",
					"executable file not found",
					"/nonexistent/path/to/meta-cc",
				}

				for _, indicator := range cliIndicators {
					if strings.Contains(errMsg, indicator) {
						t.Fatalf("Special tool %s attempted to use CLI (error: %v). "+
							"Should use direct implementation.", tc.name, err)
					}
				}
			}
		})
	}
}
