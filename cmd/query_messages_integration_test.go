package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQueryUserMessagesCommand_Integration(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-user-messages-integration"
	sessionID := "test-session-query-user-messages-integration"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create fixture with user messages
	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"fix the bug in parser"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"I'll help"}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"add error handling"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"done"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	// Reset flags
	if err := queryUserMessagesCmd.Flags().Set("pattern", ""); err != nil {
		t.Fatalf("Failed to reset pattern flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "user-messages", "--session-only", "--pattern", ".*", "--output", "jsonl"})

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
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-user-messages-pattern"
	sessionID := "test-session-query-user-messages-pattern"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"debug this error"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"ok"}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"hello world"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"hi"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	// Reset flags
	if err := queryUserMessagesCmd.Flags().Set("pattern", ""); err != nil {
		t.Fatalf("Failed to reset pattern flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "user-messages", "--session-only", "--pattern", "debug", "--output", "jsonl"})

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
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-user-messages-no-matches"
	sessionID := "test-session-query-user-messages-no-matches"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"hello world"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"hi"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	// Reset flags
	if err := queryUserMessagesCmd.Flags().Set("pattern", ""); err != nil {
		t.Fatalf("Failed to reset pattern flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "user-messages", "--session-only", "--pattern", "nonexistent", "--output", "jsonl"})

	err := rootCmd.Execute()
	// Should complete without error but produce warning
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}
}
