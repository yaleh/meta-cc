package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestQueryUserMessagesCommand_Integration(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-query-user-messages-integration"
	projectPath := "/home/yale/work/test-query-user-messages-integration"

	// Create fixture with user messages
	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"fix the bug in parser"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"I'll help"}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"add error handling"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"done"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	// Reset flags
	if err := queryUserMessagesCmd.Flags().Set("pattern", ""); err != nil {
		t.Fatalf("Failed to reset pattern flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "user-messages", "--session", sessionID, "--pattern", ".*", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains user messages
	if !strings.Contains(output, "fix") && !strings.Contains(output, "handling") {
		t.Errorf("Expected output to contain user messages, got: %s", output)
	}
}

func TestQueryUserMessagesCommand_PatternFilter(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-query-user-messages-pattern"
	projectPath := "/home/yale/work/test-query-user-messages-pattern"

	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"debug this error"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"ok"}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"hello world"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"hi"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	// Reset flags
	if err := queryUserMessagesCmd.Flags().Set("pattern", ""); err != nil {
		t.Fatalf("Failed to reset pattern flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "user-messages", "--session", sessionID, "--pattern", "debug", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains only matching messages
	if !strings.Contains(output, "debug") {
		t.Errorf("Expected output to contain 'debug', got: %s", output)
	}
	if strings.Contains(output, "hello world") {
		t.Errorf("Expected output to NOT contain 'hello world', got: %s", output)
	}
}

func TestQueryUserMessagesCommand_NoMatches(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-query-user-messages-no-matches"
	projectPath := "/home/yale/work/test-query-user-messages-no-matches"

	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"hello world"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"hi"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	// Reset flags
	if err := queryUserMessagesCmd.Flags().Set("pattern", ""); err != nil {
		t.Fatalf("Failed to reset pattern flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "user-messages", "--session", sessionID, "--pattern", "nonexistent", "--output", "jsonl"})

	err := rootCmd.Execute()
	// Should complete without error but produce warning
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}
}
