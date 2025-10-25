package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestStatsFilesCommand_Integration(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-stats-files-integration"
	projectPath := "/home/yale/work/test-stats-files-integration"

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
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"stats", "files", "--session", sessionID, "--output", "jsonl"})

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
	sessionID := "test-session-stats-files-sort"
	projectPath := "/home/yale/work/test-stats-files-sort"

	// test.txt: 2 edits
	// test2.txt: 1 edit
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Edit","input":{"file_path":"/test.txt","old_string":"a","new_string":"b"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"edited"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Edit","input":{"file_path":"/test.txt","old_string":"c","new_string":"d"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"edited"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Edit","input":{"file_path":"/test2.txt","old_string":"x","new_string":"y"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"edited"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	// Reset flags
	if err := statsFilesCmd.Flags().Set("sort-by", "total_ops"); err != nil {
		t.Fatalf("Failed to reset sort-by flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"stats", "files", "--session", sessionID, "--sort-by", "edit_count", "--output", "jsonl"})

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
	sessionID := "test-session-stats-files-top"
	projectPath := "/home/yale/work/test-stats-files-top"

	// 3 different files
	fixtureContent := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test1.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/test2.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"content"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:04:00.000Z","uuid":"uuid-5","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"Read","input":{"file_path":"/test3.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00.000Z","uuid":"uuid-6","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"content"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

	// Reset flags
	if err := statsFilesCmd.Flags().Set("top", "0"); err != nil {
		t.Fatalf("Failed to reset top flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"stats", "files", "--session", sessionID, "--top", "2", "--output", "jsonl"})

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
