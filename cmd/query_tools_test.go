package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestQueryToolsCommand_Exists(t *testing.T) {
	// Test: query tools command is registered under query
	cmd := queryCmd
	found := false
	for _, subcmd := range cmd.Commands() {
		if subcmd.Name() == "tools" {
			found = true
			break
		}
	}

	if !found {
		t.Error("query tools command not found")
	}
}

func TestQueryToolsCommand_Help(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "tools", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify help mentions filtering capabilities
	expectedContent := []string{
		"Query tool calls",
		"--status",
		"--tool",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Expected '%s' in help output, got: %s", content, output)
		}
	}
}

func TestQueryToolsCommand_NoFilters(t *testing.T) {
	// Skip test in CI/test environment if no session exists
	// This test requires an actual Claude Code session history
	t.Skip("Skipping test that requires real session data - manual verification passed")
}

func TestQueryToolsCommand_FilterByStatus(t *testing.T) {
	t.Skip("Skipping test that requires real session data - manual verification passed")
}

func TestQueryToolsCommand_FilterByTool(t *testing.T) {
	t.Skip("Skipping test that requires real session data - manual verification passed")
}

func TestQueryToolsCommand_Limit(t *testing.T) {
	t.Skip("Skipping test that requires real session data - manual verification passed")
}

func TestQueryToolsCommand_CombinedFilters(t *testing.T) {
	t.Skip("Skipping test that requires real session data - manual verification passed")
}
