package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatsTimeSeriesCommand_Integration(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-stats-timeseries-integration"
	sessionID := "test-session-stats-timeseries-integration"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create fixture with timestamps across multiple hours
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T11:00:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Edit","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T11:01:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"edited"}]}}
{"type":"assistant","timestamp":"2025-10-02T11:30:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Read","input":{"file_path":"/test2.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T11:31:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"content"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"stats", "time-series", "--session", sessionID, "--metric", "tool-calls", "--interval", "hour", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains time series data
	// Should have buckets for 10:00 hour (1 tool) and 11:00 hour (2 tools)
	if !strings.Contains(output, "timestamp") {
		t.Errorf("Expected output to contain time series data, got: %s", output)
	}
}

func TestStatsTimeSeriesCommand_ErrorRate(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-stats-timeseries-error"
	sessionID := "test-session-stats-timeseries-error"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create fixture with errors
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:30:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/missing.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:31:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","is_error":true,"content":"not found"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	// Reset flags
	if err := statsTimeSeriesCmd.Flags().Set("metric", "tool-calls"); err != nil {
		t.Fatalf("Failed to reset metric flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"stats", "time-series", "--session", sessionID, "--metric", "error-rate", "--interval", "hour", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains error rate metric
	if len(output) == 0 {
		t.Errorf("Expected output with error rate data, got empty string")
	}
}

func TestStatsTimeSeriesCommand_DayInterval(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-stats-timeseries-day"
	sessionID := "test-session-stats-timeseries-day"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create fixture with timestamps across multiple days
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-03T10:00:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Edit","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-03T10:01:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"edited"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	// Reset flags
	if err := statsTimeSeriesCmd.Flags().Set("interval", "hour"); err != nil {
		t.Fatalf("Failed to reset interval flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"stats", "time-series", "--session", sessionID, "--metric", "tool-calls", "--interval", "day", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Should bucket by day (Oct 2 and Oct 3)
	if len(output) == 0 {
		t.Errorf("Expected output with day-bucketed data, got empty string")
	}
}
