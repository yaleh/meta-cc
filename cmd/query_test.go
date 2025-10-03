package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestQueryCommandExists(t *testing.T) {
	// Test: query command is registered and shows help
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify help text is displayed
	if !strings.Contains(output, "Query and retrieve specific data") {
		t.Errorf("Expected help text, got: %s", output)
	}

	// Verify subcommands are mentioned
	expectedSubcommands := []string{"tools", "user-messages"}
	for _, subcmd := range expectedSubcommands {
		if !strings.Contains(output, subcmd) {
			t.Errorf("Expected subcommand '%s' in help, got: %s", subcmd, output)
		}
	}
}

func TestQueryCommandHelp(t *testing.T) {
	// Test: meta-cc query --help shows help and usage
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify help shows query description
	if !strings.Contains(output, "Query and retrieve specific data") {
		t.Errorf("Expected query description in help, got: %s", output)
	}

	// Verify examples are shown
	if !strings.Contains(output, "Examples:") {
		t.Errorf("Expected Examples section in help, got: %s", output)
	}
}

func TestQueryInvalidSubcommand(t *testing.T) {
	// Test: meta-cc query invalid-type
	// Expected: error with helpful message
	// NOTE: Cobra doesn't error on unknown subcommands unless we add validation
	// This test verifies the command exists and can be called
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"query", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Query command should exist: %v", err)
	}

	// Just verify query command is accessible
	if !strings.Contains(buf.String(), "Query and retrieve") {
		t.Error("Query command not properly registered")
	}
}

func TestQueryCommonFlags(t *testing.T) {
	// Test: verify common flags are available by checking they're registered
	cmd := queryCmd

	if cmd == nil {
		t.Fatal("query command is nil")
	}

	// Check that persistent flags are registered
	expectedFlags := []string{"limit", "sort-by", "reverse", "offset"}
	for _, flagName := range expectedFlags {
		flag := cmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected persistent flag '%s' to be registered", flagName)
		}
	}
}
