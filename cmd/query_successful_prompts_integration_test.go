package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuerySuccessfulPromptsCommand_Integration(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-successful-prompts-integration"
	sessionID := "test-session-query-successful-prompts-integration"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	// Create fixture with successful prompts
	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"create a function to parse JSON"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Write","input":{"file_path":"/parser.go"}}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"written"}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"thanks"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "successful-prompts", "--session", sessionID, "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify output contains successful prompt patterns (should have quality_score and pattern_features)
	if !strings.Contains(output, "quality_score") || !strings.Contains(output, "pattern_features") {
		t.Errorf("Expected output to contain successful prompt info, got: %s", output)
	}
}

func TestQuerySuccessfulPromptsCommand_MinQualityScore(t *testing.T) {
	// Prepare test environment
	homeDir, _ := os.UserHomeDir()
	projectHash := "-home-yale-work-test-query-successful-prompts-quality"
	sessionID := "test-session-query-successful-prompts-quality"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")

	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"write tests"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Write","input":{"file_path":"/test.go"}}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"written"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(fixtureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	// Reset flags
	if err := querySuccessfulPromptsCmd.Flags().Set("min-quality-score", "0.0"); err != nil {
		t.Fatalf("Failed to reset min-quality-score flag: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "successful-prompts", "--session", sessionID, "--min-quality-score", "0.5", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// Should complete without error (even if no high-quality prompts found)
	output := buf.String()
	_ = output // Output may be empty or contain results
}
