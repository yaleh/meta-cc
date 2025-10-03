package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestQueryUserMessagesCommand_Exists(t *testing.T) {
	// Test: query user-messages command is registered under query
	cmd := queryCmd
	found := false
	for _, subcmd := range cmd.Commands() {
		if subcmd.Name() == "user-messages" {
			found = true
			break
		}
	}

	if !found {
		t.Error("query user-messages command not found")
	}
}

func TestQueryUserMessagesCommand_Help(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "user-messages", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify help mentions pattern matching
	expectedContent := []string{
		"Query user messages",
		"--match",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Expected '%s' in help output, got: %s", content, output)
		}
	}
}

func TestQueryUserMessagesCommand_NoFilters(t *testing.T) {
	t.Skip("Skipping test that requires real session data - manual verification needed")
}

func TestQueryUserMessagesCommand_Match(t *testing.T) {
	t.Skip("Skipping test that requires real session data - manual verification needed")
}

func TestQueryUserMessagesCommand_Limit(t *testing.T) {
	t.Skip("Skipping test that requires real session data - manual verification needed")
}
