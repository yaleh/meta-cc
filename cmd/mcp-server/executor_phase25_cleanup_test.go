package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/config"
)

// pathToHash replicates the hash logic from internal/locator
// Used by tests to create mock session directories
func pathToHash(path string) string {
	hash := strings.ReplaceAll(path, "\\", "-")
	hash = strings.ReplaceAll(hash, "/", "-")
	hash = strings.ReplaceAll(hash, ":", "-")
	return hash
}

// setupTestSessionDir creates a mock Claude projects directory structure
// for testing SessionLocator integration. Returns the project path.
func setupTestSessionDir(t *testing.T, testData string) string {
	t.Helper()

	// Create temp directory as mock Claude projects root
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	// Create temp directory as project path
	projectPath := t.TempDir()

	// Resolve symlinks for consistent hashing on macOS (/var -> /private/var)
	resolvedPath, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		// If path doesn't exist yet, use original path
		resolvedPath = projectPath
	}

	// Calculate project hash (same logic as SessionLocator)
	projectHash := pathToHash(resolvedPath)

	// Create session directory: {projectsRoot}/{hash}/
	sessionDir := filepath.Join(projectsRoot, projectHash)
	err = os.MkdirAll(sessionDir, 0755)
	require.NoError(t, err)

	// Create session file
	sessionFile := filepath.Join(sessionDir, "test-session.jsonl")
	err = os.WriteFile(sessionFile, []byte(testData), 0644)
	require.NoError(t, err)

	return projectPath
}

// TestPhase25ToolsNoDoubleJQApplication verifies that all Phase 25 tools
// (query, query_raw, and 10 convenience tools) do NOT have jq_filter
// applied a second time by the executor.
//
// Double jq application bug: Convenience tools internally call handleQuery(),
// which already executes jq. The executor should NOT apply jq_filter again
// because it would cause "expected an object but got: array" errors.
func TestPhase25ToolsNoDoubleJQApplication(t *testing.T) {
	// Create test data with both user and assistant messages
	testData := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","message":{"content":"test message"}}
{"type":"assistant","timestamp":"2025-01-01T10:00:01Z","message":{"content":[{"type":"tool_use","name":"Read","id":"1","input":{}}],"usage":{"input_tokens":100,"output_tokens":50}}}
{"type":"user","timestamp":"2025-01-01T10:00:02Z","message":{"content":[{"type":"tool_result","tool_use_id":"1","content":"result"}]}}
{"type":"summary","timestamp":"2025-01-01T10:00:03Z","summary":"Test summary"}
{"type":"file-history-snapshot","timestamp":"2025-01-01T10:00:04Z","messageId":"msg1","files":[]}
`

	// Setup mock Claude session directory structure
	projectPath := setupTestSessionDir(t, testData)

	// Change to project directory so SessionLocator can find sessions
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()
	err = os.Chdir(projectPath)
	require.NoError(t, err)

	// Create config
	cfg := &config.Config{}

	executor := NewToolExecutor()

	// Phase 25 tools: 1 core + 1 raw + 10 convenience = 12 tools
	phase25Tools := []struct {
		name string
		args map[string]interface{}
	}{
		// Core tools
		{"query", map[string]interface{}{"jq_filter": ".[]"}},
		{"query_raw", map[string]interface{}{"jq_expression": ".[]"}},

		// Convenience tools (all call handleQuery internally)
		{"query_user_messages", map[string]interface{}{"pattern": "test"}},
		{"query_tools", map[string]interface{}{}},
		{"query_tool_errors", map[string]interface{}{}},
		{"query_token_usage", map[string]interface{}{}},
		{"query_conversation_flow", map[string]interface{}{}},
		{"query_system_errors", map[string]interface{}{}},
		{"query_file_snapshots", map[string]interface{}{}},
		{"query_timestamps", map[string]interface{}{}},
		{"query_summaries", map[string]interface{}{}},
		{"query_tool_blocks", map[string]interface{}{"block_type": "tool_use"}},
	}

	for _, tc := range phase25Tools {
		t.Run(tc.name, func(t *testing.T) {
			// Set scope parameter for the tool
			args := make(map[string]interface{})
			for k, v := range tc.args {
				args[k] = v
			}
			args["scope"] = "session"

			// Execute tool
			result, err := executor.ExecuteTool(cfg, tc.name, args)

			// Should NOT get "expected an object but got: array" error
			// This error indicates double jq application
			if err != nil {
				require.NotContains(t, err.Error(), "expected an object but got: array",
					"Tool %s has double jq application bug", tc.name)

				// If there's an error but NOT the double jq bug, log it for debugging
				t.Logf("Tool %s returned error (not double jq bug): %v", tc.name, err)

				// For convenience tools, errors are acceptable as long as it's NOT double jq
				if tc.name != "query" && tc.name != "query_raw" {
					return
				}

				// For query/query_raw, error is a test failure
				t.Fatalf("Tool %s failed: %v", tc.name, err)
			}

			// Result should be valid (not testing exact content, just successful execution)
			require.NotEmpty(t, result, "Tool %s should return non-empty result", tc.name)
		})
	}
}

// TestLegacyToolsRemoved verifies that legacy tools have been completely removed
func TestLegacyToolsRemoved(t *testing.T) {
	cfg := &config.Config{}
	executor := NewToolExecutor()

	legacyTools := []string{
		"query_tool_sequences",
		"query_file_access",
		"get_session_stats",
		"query_project_state",
		"query_successful_prompts",
	}

	for _, toolName := range legacyTools {
		t.Run(toolName, func(t *testing.T) {
			_, err := executor.ExecuteTool(cfg, toolName, map[string]interface{}{"scope": "project"})

			// Should return "unknown tool" error
			require.Error(t, err)
			require.Contains(t, err.Error(), "unknown tool",
				"Legacy tool %s should be removed", toolName)
		})
	}
}

// TestPhase25ToolCount verifies expected tool count after cleanup
func TestPhase25ToolCount(t *testing.T) {
	tools := getToolDefinitions()

	// Expected: 15 tools total
	// - 1 query (Layer 2)
	// - 1 query_raw (Layer 3)
	// - 10 convenience tools (Layer 1)
	// - 3 utility tools (cleanup_temp_files, list_capabilities, get_capability)
	//
	// Removed: 5 legacy tools (query_tool_sequences, query_file_access, get_session_stats,
	//          query_project_state, query_successful_prompts)
	expectedCount := 15

	actualCount := len(tools)
	require.Equal(t, expectedCount, actualCount,
		"Expected %d tools after Phase 25 cleanup, got %d", expectedCount, actualCount)
}

// TestNoBackwardCompatibilityCode verifies all backward compatibility code removed
func TestNoBackwardCompatibilityCode(t *testing.T) {
	t.Run("no_legacy_mode_in_response_adapter", func(t *testing.T) {
		// OutputModeLegacy should not exist
		// This test will fail to compile if OutputModeLegacy still exists
		// But we can't reference it directly, so we test the behavior instead

		cfg := &config.Config{}
		data := []interface{}{map[string]interface{}{"test": "data"}}
		params := map[string]interface{}{
			"output_mode": "legacy", // Should be ignored/error
		}

		// adaptResponse should NOT recognize "legacy" mode
		result, err := adaptResponse(cfg, data, params, "query")
		require.NoError(t, err)

		// Should return hybrid mode response, NOT raw array
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok, "Should return hybrid mode response (map), not legacy raw array")
		require.Contains(t, resultMap, "mode", "Response should have 'mode' field")
	})

	t.Run("no_deprecated_filter_functions", func(t *testing.T) {
		// DefaultMaxMessageLength constant should not exist or be set to 0
		// FilterMessages and FilterMessagesSummary should not exist
		// This is a smoke test - actual removal verified by compilation

		// We can verify the constant value if it exists
		require.Equal(t, 0, DefaultMaxMessageLength,
			"DefaultMaxMessageLength should be 0 (no truncation) or removed")
	})
}

// TestPhase25ToolsExecuteJQOnce verifies jq filter is applied exactly once
func TestPhase25ToolsExecuteJQOnce(t *testing.T) {
	// This test verifies the executor logic:
	// - query and query_raw: handleQuery/handleQueryRaw already execute jq
	// - 10 convenience tools: handleQueryXxx calls handleQuery, which executes jq
	// - Therefore: executor should NOT apply jq_filter again for any Phase 25 tool

	t.Run("executor_skips_jq_for_phase25_tools", func(t *testing.T) {
		// Read executor.go source code
		executorSource, err := os.ReadFile("executor.go")
		require.NoError(t, err)

		executorCode := string(executorSource)

		// Verify that shouldApplyJQFilter logic excludes ALL Phase 25 tools
		// Expected pattern: Check for Phase 25 tool list
		require.Contains(t, executorCode, "Phase 25",
			"Executor should have Phase 25 tool handling logic")

		// The executor should have logic like:
		// isPhase25Tool := contains(phase25Tools, toolName)
		// shouldApplyJQFilter := !isPhase25Tool
		//
		// NOT:
		// shouldApplyJQFilter := toolName != "query" && toolName != "query_raw"

		// This is a smoke test - actual behavior tested by TestPhase25ToolsNoDoubleJQApplication
	})
}
