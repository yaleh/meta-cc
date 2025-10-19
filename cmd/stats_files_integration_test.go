package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatsFilesCommand_Integration(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-stats-files-integration"
	sessionID := "test-session-stats-files-integration"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create fixture with file operations
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Edit","input":{"file_path":"/test.txt","old_string":"a","new_string":"b"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"edited"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"new content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:06:00.000Z","uuid":"uuid-7","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-4","name":"Write","input":{"file_path":"/test2.txt","content":"data"}}]}}
{"type":"user","timestamp":"2025-10-02T10:07:00.000Z","uuid":"uuid-8","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-4","content":"written"}]}}
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
	rootCmd.SetArgs([]string{"stats", "files", "--session-only", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains file statistics
	if !strings.Contains(output, "test.txt") {
		t.Errorf("Expected output to contain 'test.txt', got: %s", output)
	}

	// Should have operation counts
	if !strings.Contains(output, "read_count") || !strings.Contains(output, "edit_count") {
		t.Errorf("Expected output to contain operation counts, got: %s", output)
	}
}

func TestStatsFilesCommand_SortByEditCount(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-stats-files-sort"
	sessionID := "test-session-stats-files-sort"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// test.txt: 2 edits
	// test2.txt: 1 edit
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Edit","input":{"file_path":"/test.txt","old_string":"a","new_string":"b"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"edited"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Edit","input":{"file_path":"/test.txt","old_string":"c","new_string":"d"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"edited"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Edit","input":{"file_path":"/test2.txt","old_string":"x","new_string":"y"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"edited"}]}}
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
	if err := statsFilesCmd.Flags().Set("sort-by", "total_ops"); err != nil {
		t.Fatalf("Failed to reset sort-by flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"stats", "files", "--session-only", "--sort-by", "edit_count", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify both files present (sorted by edit count)
	if !strings.Contains(output, "test.txt") || !strings.Contains(output, "test2.txt") {
		t.Errorf("Expected output to contain both files, got: %s", output)
	}
}

func TestStatsFilesCommand_TopN(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-stats-files-top"
	sessionID := "test-session-stats-files-top"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// 3 different files
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test1.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/test2.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Read","input":{"file_path":"/test3.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"content"}]}}
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
	if err := statsFilesCmd.Flags().Set("top", "0"); err != nil {
		t.Fatalf("Failed to reset top flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"stats", "files", "--session-only", "--top", "2", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Should only show top 2 files
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 2 {
		t.Errorf("Expected top 2 files, got %d lines: %s", len(lines), output)
	}
}
