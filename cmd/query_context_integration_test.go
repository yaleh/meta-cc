package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQueryContextCommand_Integration(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-context-integration"
	sessionID := "test-session-query-context-integration"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create fixture with error and context
	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"before error"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","is_error":true,"content":"file not found"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"after error"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	// Reset flags
	if err := queryContextCmd.Flags().Set("error-signature", ""); err != nil {
		t.Fatalf("Failed to reset error-signature flag: %v", err)
	}
	if err := queryContextCmd.Flags().Set("window", "3"); err != nil {
		t.Fatalf("Failed to reset window flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "context", "--session", sessionID, "--error-signature", "Read:file", "--window", "1"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains context information
	if !strings.Contains(output, "Read") {
		t.Errorf("Expected output to contain error tool 'Read', got: %s", output)
	}
}

func TestQueryContextCommand_InvalidSignature(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-context-invalid"
	sessionID := "test-session-query-context-invalid"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"success"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	// Reset flags
	if err := queryContextCmd.Flags().Set("error-signature", ""); err != nil {
		t.Fatalf("Failed to reset error-signature flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "context", "--session", sessionID, "--error-signature", "NonExistent:error"})

	err := rootCmd.Execute()
	// Should complete without error but find no occurrences
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}
}
