package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestQuerySuccessfulPromptsCommand_Integration(t *testing.T) {
	// Prepare test environment
	sessionID := "test-session-query-successful-prompts-integration"
	projectPath := "/home/yale/work/test-query-successful-prompts-integration"

	// Create fixture with successful prompts
	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"create a function to parse JSON"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Write","input":{"file_path":"/parser.go"}}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"written"}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"thanks"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

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
	sessionID := "test-session-query-successful-prompts-quality"
	projectPath := "/home/yale/work/test-query-successful-prompts-quality"

	fixtureContent := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":[{"type":"text","text":"write tests"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Write","input":{"file_path":"/test.go"}}]}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"written"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixtureContent)

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
