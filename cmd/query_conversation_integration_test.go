package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestQueryConversationCommand_Integration(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-query-conversation-integration"
	projectPath := "/home/yale/work/test-query-conversation-integration"

	// Create fixture with conversation turns
	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"hello world"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"hi there"}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"how are you"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"doing great"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "conversation", "--session", sessionID, "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains conversation turns
	if !strings.Contains(output, "hello") {
		t.Errorf("Expected output to contain user message, got: %s", output)
	}
	if !strings.Contains(output, "hi there") {
		t.Errorf("Expected output to contain assistant response, got: %s", output)
	}
}

func TestQueryConversationCommand_PatternFilter(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-query-conversation-pattern"
	projectPath := "/home/yale/work/test-query-conversation-pattern"

	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"debug this error"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"let me help"}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"hello world"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"hi there"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	// Reset flags
	if err := queryConversationCmd.Flags().Set("pattern", ""); err != nil {
		t.Fatalf("Failed to reset pattern flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "conversation", "--session", sessionID, "--pattern", "debug", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains only matching conversation
	if !strings.Contains(output, "debug") {
		t.Errorf("Expected output to contain 'debug', got: %s", output)
	}
}
