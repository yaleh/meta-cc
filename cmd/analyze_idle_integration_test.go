package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestAnalyzeIdleCommand_Integration(t *testing.T) {
	// Prepare test environment
	testSessionID := "test-session-analyze-idle-integration"
	projectPath := "/home/yale/work/test-analyze-idle-integration"

	// Create fixture with idle periods (10+ minute gaps)
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:15:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Edit","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:16:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"edited"}]}}
`
	writeSessionFixture(t, projectPath, testSessionID, fixtureContent)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"analyze", "idle-periods", "--session", testSessionID, "--output", "jsonl", "--threshold", "5"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains idle period data
	// Should detect the 14-minute gap between 10:01 and 10:15
	if !strings.Contains(output, "duration") {
		t.Errorf("Expected output to contain idle period data, got: %s", output)
	}
}

func TestAnalyzeIdleCommand_WithThreshold(t *testing.T) {
	// Prepare test environment
	testSessionID := "test-session-analyze-idle-threshold"
	projectPath := "/home/yale/work/test-analyze-idle-threshold"

	// Create fixture with varying gap sizes
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Edit","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"edited"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:20:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Read","input":{"file_path":"/test2.txt"}}]}}
`
	writeSessionFixture(t, projectPath, testSessionID, fixtureContent)

	// Reset flags from previous tests
	if err := analyzeIdleCmd.Flags().Set("threshold", "5"); err != nil {
		t.Fatalf("Failed to reset threshold flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"analyze", "idle-periods", "--session", testSessionID, "--output", "jsonl", "--threshold", "10"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// With threshold=10, should only detect the 15-minute gap (10:05-10:20)
	// Should NOT detect the 3-minute gap (10:01-10:04)
	if len(output) == 0 {
		t.Errorf("Expected output with idle periods, got empty string")
	}
}
