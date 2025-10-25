package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestQueryProjectStateCommand_Integration(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-query-project-state-integration"
	projectPath := "/home/yale/work/test-query-project-state-integration"

	// Create fixture with various tool calls
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"file content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Edit","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"edited"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"output"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "project-state", "--session", sessionID, "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains project state information
	if !strings.Contains(output, "session_id") || !strings.Contains(output, "recent_files") {
		t.Errorf("Expected output to contain project state fields, got: %s", output)
	}
}

func TestQueryProjectStateCommand_EmptySession(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-query-project-state-empty"
	projectPath := "/home/yale/work/test-query-project-state-empty"

	// Create empty fixture
	fixtureContent := ""
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "project-state", "--session", sessionID, "--output", "jsonl"})

	err := rootCmd.Execute()
	// Should complete without error but produce warning
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}
}
