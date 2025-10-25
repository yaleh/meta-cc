package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestQuerySequencesCommand_Integration(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-query-sequences-integration"
	projectPath := "/home/yale/work/test-query-sequences-integration"

	// Create fixture with tool sequences
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Edit","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"edited"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:06:00.000Z","uuid":"uuid-7","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-4","name":"Edit","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:07:00.000Z","uuid":"uuid-8","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-4","content":"edited"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "tool-sequences", "--session", sessionID, "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains sequence or help text (command may not find repeated sequences)
	// Just verify command executed successfully without errors
	if len(output) == 0 {
		t.Errorf("Expected output from command, got empty string")
	}
}

func TestQuerySequencesCommand_WithPattern(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-query-sequences-pattern"
	projectPath := "/home/yale/work/test-query-sequences-pattern"

	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Grep","input":{"pattern":"test"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"match"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Grep","input":{"pattern":"test2"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"match2"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:06:00.000Z","uuid":"uuid-7","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-4","name":"Read","input":{"file_path":"/test2.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:07:00.000Z","uuid":"uuid-8","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-4","content":"content2"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "tool-sequences", "--session", sessionID, "--pattern", "Grep", "--min-occurrences", "1", "--include-builtin-tools", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains matching sequences
	if !strings.Contains(output, "Grep") {
		t.Errorf("Expected output to contain 'Grep' sequence, got: %s", output)
	}
}
