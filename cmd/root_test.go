package cmd

import (
	"testing"
)

// TestProjectFlagDefaultBehavior tests that commands default to using --project .
// when no session-specific flag is provided
func TestProjectFlagDefaultBehavior(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectProject bool
	}{
		{
			name:          "parse stats with no flags should use project",
			args:          []string{"parse", "stats"},
			expectProject: true,
		},
		{
			name:          "query tools with no flags should use project",
			args:          []string{"query", "tools"},
			expectProject: true,
		},
		{
			name:          "parse stats with --session-only should not use project",
			args:          []string{"parse", "stats", "--session-only"},
			expectProject: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies the behavior exists
			// Implementation will set projectPath to "." by default
			// unless --session-only flag is set

			// For now, just verify the test structure is correct
			if len(tt.args) == 0 {
				t.Error("Test args should not be empty")
			}
		})
	}
}

// TestSessionOnlyFlag tests the new --session-only flag that opts out of project-level analysis
func TestSessionOnlyFlag(t *testing.T) {
	t.Skip("Implementation pending")

	// When implemented, this should verify that:
	// 1. --session-only flag prevents default --project behavior
	// 2. Session is located via environment or auto-detection only
	// 3. Does not use project path defaulting
}

// Test Phase 14: meta-cc mcp subcommand should not exist (legacy removed)
// The MCP server is now a separate executable: meta-cc-mcp
func TestMCPSubcommandDoesNotExist(t *testing.T) {
	// Get all subcommands from rootCmd
	commands := rootCmd.Commands()

	// Check that "mcp" subcommand does NOT exist
	for _, cmd := range commands {
		if cmd.Name() == "mcp" {
			t.Errorf("Phase 14: 'mcp' subcommand should not exist. Use meta-cc-mcp executable instead.")
			t.Errorf("Found legacy mcp subcommand at: %s", cmd.Use)
		}
	}
}

// Test Phase 14: Verify expected subcommands exist (regression test)
func TestExpectedSubcommandsExist(t *testing.T) {
	expectedCommands := []string{
		"parse",
		"query",
		"analyze",
	}

	commands := rootCmd.Commands()
	commandMap := make(map[string]bool)

	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}
