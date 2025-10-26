package main

import (
	"encoding/json"
	"testing"

	"github.com/yaleh/meta-cc/internal/config"
)

// Phase 27 Stage 27.1: Tests for query and query_raw removed
// These tools were deleted to simplify the query interface
// The 10 convenience tools now call executeQuery() directly

// TestConvenienceToolsExecuteCorrectly verifies that convenience tools
// execute jq queries correctly without double jq application
func TestConvenienceToolsExecuteCorrectly(t *testing.T) {
	cleanup := setupLibraryFixture(t)
	defer cleanup()

	cfg := &config.Config{}
	executor := NewToolExecutor()

	tests := []struct {
		name        string
		toolName    string
		args        map[string]interface{}
		expectError bool
	}{
		{
			name:     "query_user_messages should execute correctly",
			toolName: "query_user_messages",
			args: map[string]interface{}{
				"pattern": ".*",
				"limit":   float64(2),
			},
			expectError: false,
		},
		{
			name:     "query_tools should execute correctly",
			toolName: "query_tools",
			args: map[string]interface{}{
				"limit": float64(2),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executor.ExecuteTool(cfg, tt.toolName, tt.args)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, but got no error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Parse output to verify structure
			var result interface{}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("failed to parse output as JSON: %v\nOutput: %s", err, output)
			}

			t.Logf("Tool %s executed successfully, output length: %d", tt.toolName, len(output))
		})
	}
}
