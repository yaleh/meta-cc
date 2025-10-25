package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/config"
)

// TestQueryToolsDoNotApplySecondJQFilter tests that query and query_raw tools
// do NOT apply jq_filter a second time after their internal jq execution.
// This prevents the "expected an object but got: array" error.
func TestQueryToolsDoNotApplySecondJQFilter(t *testing.T) {
	cleanup := setupLibraryFixture(t)
	defer cleanup()

	cfg := &config.Config{}
	executor := NewToolExecutor()

	tests := []struct {
		name          string
		toolName      string
		args          map[string]interface{}
		shouldApplyJQ bool
		expectError   bool
		errorContains string
	}{
		{
			name:     "query tool with jq_filter should not apply filter twice",
			toolName: "query",
			args: map[string]interface{}{
				"jq_filter": "select(.type == \"user\")",
				"limit":     float64(2),
			},
			shouldApplyJQ: false, // Should NOT apply jq_filter again
			expectError:   false,
		},
		{
			name:     "query_raw tool should not apply jq_filter",
			toolName: "query_raw",
			args: map[string]interface{}{
				"jq_expression": "select(.type == \"user\")",
				"limit":         float64(2),
			},
			shouldApplyJQ: false, // Should NOT apply jq_filter
			expectError:   false,
		},
		{
			name:     "query_user_messages should apply jq_filter for post-processing",
			toolName: "query_user_messages",
			args: map[string]interface{}{
				"pattern": ".*",
				"limit":   float64(2),
			},
			shouldApplyJQ: true, // Should apply jq_filter (default: ".[]")
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executor.ExecuteTool(cfg, tt.toolName, tt.args)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing '%s', but got no error", tt.errorContains)
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Fatalf("expected error containing '%s', got: %v", tt.errorContains, err)
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

			// For query and query_raw, we expect a structured response with data field
			// For query_user_messages, we expect similar structure
			// The key test is that query/query_raw don't fail with "expected an object but got: array"
			t.Logf("Tool %s executed successfully, output length: %d", tt.toolName, len(output))
		})
	}
}

// TestQueryToolWithJQFilterParameter specifically tests the query tool
// to ensure it handles jq_filter parameter correctly without double application
func TestQueryToolWithJQFilterParameter(t *testing.T) {
	cleanup := setupLibraryFixture(t)
	defer cleanup()

	cfg := &config.Config{}
	executor := NewToolExecutor()

	// This should NOT fail with "expected an object but got: array"
	output, err := executor.ExecuteTool(cfg, "query", map[string]interface{}{
		"jq_filter": "select(.type == \"user\")",
		"limit":     float64(2),
	})

	if err != nil {
		// If error contains "expected an object but got: array", it means
		// jq_filter was applied twice (double application bug)
		if strings.Contains(err.Error(), "expected an object but got: array") {
			t.Fatalf("BUG: jq_filter was applied twice, causing: %v", err)
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify we got valid JSON output
	var result interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	t.Logf("query tool executed successfully with jq_filter parameter")
}

// TestQueryToolWithJQTransformParameter tests that query tool supports
// jq_transform parameter as specified in design documents
func TestQueryToolWithJQTransformParameter(t *testing.T) {
	cleanup := setupLibraryFixture(t)
	defer cleanup()

	cfg := &config.Config{}
	executor := NewToolExecutor()

	// Test jq_filter + jq_transform combination
	output, err := executor.ExecuteTool(cfg, "query", map[string]interface{}{
		"jq_filter":    "select(.type == \"user\")",
		"jq_transform": "{type, timestamp}",
		"limit":        float64(2),
	})

	if err != nil {
		t.Fatalf("query tool with jq_transform failed: %v", err)
	}

	// Verify output is valid JSON
	var result interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	t.Logf("query tool executed successfully with jq_transform parameter")
}
