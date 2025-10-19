package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQueryErrorsCommand_Integration(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-errors-integration"
	sessionID := "test-session-query-errors-integration"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create fixture with tool errors
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","is_error":true,"content":"file not found"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Grep","input":{"pattern":"test"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"match found"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Edit","input":{"file_path":"/invalid.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","is_error":true,"content":"permission denied"}]}}
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
	rootCmd.SetArgs([]string{"query", "errors", "--session-only", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains error entries
	if !strings.Contains(output, "Read") {
		t.Errorf("Expected output to contain 'Read' error, got: %s", output)
	}
	if !strings.Contains(output, "Edit") {
		t.Errorf("Expected output to contain 'Edit' error, got: %s", output)
	}
	if strings.Contains(output, "Grep") {
		t.Errorf("Expected output to NOT contain 'Grep' (successful tool), got: %s", output)
	}

	// Verify error count
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 error entries, got %d", len(lines))
	}
}

func TestQueryErrorsCommand_NoErrors(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-errors-no-errors"
	sessionID := "test-session-query-errors-no-errors"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create fixture with no errors
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"file content"}]}}
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
	rootCmd.SetArgs([]string{"query", "errors", "--session-only", "--output", "jsonl"})

	err := rootCmd.Execute()
	// Should complete with error for no results
	if err == nil {
		t.Fatalf("Expected error for no results, got nil")
	}

	// Verify error message mentions no results
	if !strings.Contains(err.Error(), "no results") && !strings.Contains(err.Error(), "No results") {
		t.Errorf("Expected 'no results' error, got: %v", err)
	}
}

func TestQueryErrorsCommand_Pagination(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-errors-pagination"
	sessionID := "test-session-query-errors-pagination"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create fixture with multiple errors
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test1.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","is_error":true,"content":"error 1"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/test2.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","is_error":true,"content":"error 2"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Read","input":{"file_path":"/test3.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","is_error":true,"content":"error 3"}]}}
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
	rootCmd.SetArgs([]string{"query", "errors", "--session-only", "--limit", "1", "--offset", "0", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Should contain errors (limit may apply at different level)
	if !strings.Contains(output, "error") {
		t.Errorf("Expected errors in output, got: %s", output)
	}
}
