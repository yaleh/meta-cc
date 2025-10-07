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

	// Verify help mentions pattern matching and context
	expectedContent := []string{
		"Query user messages",
		"--pattern",
		"--with-context",
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

func TestQueryUserMessagesCommand_Pattern(t *testing.T) {
	t.Skip("Skipping test that requires real session data - manual verification needed")
}

func TestQueryUserMessagesCommand_Limit(t *testing.T) {
	t.Skip("Skipping test that requires real session data - manual verification needed")
}

func TestIsSystemMessage(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "Slash command trigger message",
			content:  "<command-message>meta-query-messages is running…</command-message>\n<command-name>/meta-query-messages</command-name>\n<command-args>project-planner</command-args>",
			expected: true,
		},
		{
			name:     "Slash command expanded content",
			content:  "# meta-query-messages: 用户消息搜索\n\nPhase 13 更新：默认分析当前项目的最新会话。",
			expected: true,
		},
		{
			name:     "Command message only",
			content:  "<command-message>meta-stats is running…</command-message>",
			expected: true,
		},
		{
			name:     "Command name only",
			content:  "<command-name>/meta-errors</command-name>",
			expected: true,
		},
		{
			name:     "Local command message",
			content:  "<local-command stdout>\nSome output\n</local-command>",
			expected: true,
		},
		{
			name:     "Caveat message",
			content:  "Caveat: This is a system warning",
			expected: true,
		},
		{
			name:     "Regular user message",
			content:  "Execute Phase 14 with agents project-planner, and then stage-executor for every stage.",
			expected: false,
		},
		{
			name:     "User message with question",
			content:  "使用 unix 命令行工具在 claude code 本项目历史中搜索包括 `project-planner` 的用户消息。",
			expected: false,
		},
		{
			name:     "Empty message",
			content:  "",
			expected: false,
		},
		{
			name:     "Whitespace only",
			content:  "   \n\t  ",
			expected: false,
		},
		{
			name:     "User message mentioning commands",
			content:  "I want to use /meta-stats to check the session statistics",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSystemMessage(tt.content)
			if result != tt.expected {
				t.Errorf("isSystemMessage(%q) = %v, expected %v", tt.content, result, tt.expected)
			}
		})
	}
}
