package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestStatsAggregateCommand_Integration(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-stats-aggregate-integration"
	projectPath := "/home/yale/work/test-stats-aggregate-integration"

	// Create fixture with tool calls for aggregation
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/test2.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","is_error":true,"content":"not found"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Edit","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"edited"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"stats", "aggregate", "--session", sessionID, "--group-by", "tool", "--metrics", "count", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains aggregated data by tool
	if !strings.Contains(output, "Read") || !strings.Contains(output, "Edit") {
		t.Errorf("Expected output to contain tool aggregation, got: %s", output)
	}

	// Should have count field
	if !strings.Contains(output, "count") {
		t.Errorf("Expected output to contain 'count' metric, got: %s", output)
	}
}

func TestStatsAggregateCommand_ErrorRate(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-stats-aggregate-error"
	projectPath := "/home/yale/work/test-stats-aggregate-error"

	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"files"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Bash","input":{"command":"cat"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","is_error":true,"content":"error"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	// Reset flags
	if err := statsAggregateCmd.Flags().Set("group-by", "tool"); err != nil {
		t.Fatalf("Failed to reset group-by flag: %v", err)
	}
	if err := statsAggregateCmd.Flags().Set("metrics", "count"); err != nil {
		t.Fatalf("Failed to reset metrics flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"stats", "aggregate", "--session", sessionID, "--group-by", "tool", "--metrics", "error_rate", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains error_rate metric
	if !strings.Contains(output, "error_rate") {
		t.Errorf("Expected output to contain 'error_rate' metric, got: %s", output)
	}
}

func TestStatsAggregateCommand_WithFilter(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-stats-aggregate-filter"
	projectPath := "/home/yale/work/test-stats-aggregate-filter"

	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"files"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	// Reset flags
	if err := statsAggregateCmd.Flags().Set("filter", ""); err != nil {
		t.Fatalf("Failed to reset filter flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"stats", "aggregate", "--session", sessionID, "--group-by", "tool", "--metrics", "count", "--filter", "tool='Read'", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Should only contain Read tool (filtered out Bash)
	if !strings.Contains(output, "Read") {
		t.Errorf("Expected output to contain 'Read', got: %s", output)
	}
}
