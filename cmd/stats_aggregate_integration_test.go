package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatsAggregateCommand_Integration(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-stats-aggregate-integration"
	sessionID := "test-session-stats-aggregate-integration"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create fixture with tool calls for aggregation
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/test2.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","is_error":true,"content":"not found"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Edit","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"edited"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

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
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-stats-aggregate-error"
	sessionID := "test-session-stats-aggregate-error"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"files"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Bash","input":{"command":"cat"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","is_error":true,"content":"error"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

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
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-stats-aggregate-filter"
	sessionID := "test-session-stats-aggregate-filter"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"files"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

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
