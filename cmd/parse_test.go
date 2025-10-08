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
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Use test fixture content
	fixtureContent := `{"type":"user","timestamp":"2025-10-02T06:07:13.673Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"Hello"}]}}
{"type":"assistant","timestamp":"2025-10-02T06:07:14.673Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"Hi there"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
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
	rootCmd.SetArgs([]string{"parse", "extract", "--type", "turns", "--output", "jsonl"})

	// Execute command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output is valid JSONL (compact JSON)
	if !strings.Contains(output, `"type":"user"`) {
		t.Errorf("Expected JSONL output with user turn, got: %s", output)
	}

	if !strings.Contains(output, `"type":"assistant"`) {
		t.Errorf("Expected JSONL output with assistant turn, got: %s", output)
	}
}

func TestParseExtractCommand_TypeTools(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-tools"
	sessionID := "test-session-tools"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T06:07:13.673Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Grep","input":{"pattern":"error"}}]}}
{"type":"user","timestamp":"2025-10-02T06:07:14.673Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"match found"}]}}
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
	// Clear environment variables that might be set by previous tests
	os.Unsetenv("CC_SESSION_ID")
	os.Unsetenv("CC_PROJECT_HASH")

	// Change to a temporary directory that has no session
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(oldWd)

	// Reset rootCmd flags to clean state
	rootCmd.Flags().Set("session", "")
	rootCmd.Flags().Set("project", "")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"parse", "extract", "--type", "turns"})

	err := rootCmd.Execute()
	if err == nil {
		t.Errorf("Expected error when session file not found, but command succeeded. Output: %s", buf.String())
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

func TestParseStatsCommand_JSON(t *testing.T) {
	// Prepare test environment: create temporary session file
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-stats"
	sessionID := "test-session-stats"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create test fixture with tool calls
	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"Hello"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Grep","input":{"pattern":"test"}}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"match found"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","is_error":true,"content":"file not found"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	// Set environment variables
	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"parse", "stats", "--output", "jsonl"})

	// Execute command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains statistics
	expectedFields := []string{
		"TurnCount",
		"ToolCallCount",
		"ErrorCount",
		"DurationSeconds",
		"ErrorRate",
		"ToolFrequency",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected output to contain '%s', got: %s", field, output)
		}
	}
}

func TestParseStatsCommand_Markdown(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-stats-md"
	sessionID := "test-session-stats-md"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"Hello"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:05:30.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"Hi"}]}}
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
	rootCmd.SetArgs([]string{"parse", "stats", "--output", "tsv"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify we got some output (TSV/JSONL format)
	if output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestParseStatsCommand_MissingSession(t *testing.T) {
	// Clear environment variables that might be set by previous tests
	os.Unsetenv("CC_SESSION_ID")
	os.Unsetenv("CC_PROJECT_HASH")

	// Change to a temporary directory that has no session
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(oldWd)

	// Reset rootCmd flags to clean state
	rootCmd.Flags().Set("session", "")
	rootCmd.Flags().Set("project", "")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"parse", "stats"})

	err := rootCmd.Execute()
	if err == nil {
		t.Errorf("Expected error when session file not found, but command succeeded. Output: %s", buf.String())
	}
}
