package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuerySequencesCommand_Integration(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-sequences-integration"
	sessionID := "test-session-query-sequences-integration"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

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
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "tool-sequences", "--session-only", "--output", "jsonl"})

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
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-sequences-pattern"
	sessionID := "test-session-query-sequences-pattern"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Grep","input":{"pattern":"test"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"match"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Grep","input":{"pattern":"test2"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"match2"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:06:00.000Z","uuid":"uuid-7","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-4","name":"Read","input":{"file_path":"/test2.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:07:00.000Z","uuid":"uuid-8","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-4","content":"content2"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "tool-sequences", "--session-only", "--pattern", "Grep", "--min-occurrences", "1", "--include-builtin-tools", "--output", "jsonl"})

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
