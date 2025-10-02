package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseExtractCommand_TypeTurns(t *testing.T) {
	// Prepare test environment: create temporary session file
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-parse"
	sessionID := "test-session-extract"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	os.MkdirAll(sessionDir, 0755)
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Use test fixture content
	fixtureContent := `{"type":"user","timestamp":"2025-10-02T06:07:13.673Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"Hello"}]}}
{"type":"assistant","timestamp":"2025-10-02T06:07:14.673Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"Hi there"}]}}
`
	os.WriteFile(sessionFile, []byte(fixtureContent), 0644)
	defer os.RemoveAll(sessionDir)

	// Set environment variables
	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"parse", "extract", "--type", "turns", "--output", "json"})

	// Execute command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output is valid JSON
	if !strings.Contains(output, `"type": "user"`) {
		t.Errorf("Expected JSON output with user turn, got: %s", output)
	}

	if !strings.Contains(output, `"type": "assistant"`) {
		t.Errorf("Expected JSON output with assistant turn, got: %s", output)
	}
}

func TestParseExtractCommand_TypeTools(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-tools"
	sessionID := "test-session-tools"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	os.MkdirAll(sessionDir, 0755)
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T06:07:13.673Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Grep","input":{"pattern":"error"}}]}}
{"type":"user","timestamp":"2025-10-02T06:07:14.673Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"match found"}]}}
`
	os.WriteFile(sessionFile, []byte(fixtureContent), 0644)
	defer os.RemoveAll(sessionDir)

	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"parse", "extract", "--type", "tools"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify contains tool name
	if !strings.Contains(output, "Grep") {
		t.Errorf("Expected tool name 'Grep' in output, got: %s", output)
	}
}

func TestParseExtractCommand_MissingSessionFile(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("CC_SESSION_ID")
	os.Unsetenv("CC_PROJECT_HASH")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"parse", "extract", "--type", "turns"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when session file not found")
	}
}

func TestParseExtractCommand_InvalidType(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"parse", "extract", "--type", "invalid"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid type")
	}

	errOutput := err.Error()
	if !strings.Contains(errOutput, "invalid type") {
		t.Errorf("Expected error message about invalid type, got: %s", errOutput)
	}
}
