package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCommandBuilder tests the command builder abstraction
func TestCommandBuilder(t *testing.T) {
	tests := []struct {
		name     string
		builder  *CommandBuilder
		expected []string
	}{
		{
			name: "basic query tools command",
			builder: NewCommandBuilder("query", "tools").
				WithScope("session").
				WithOutputFormat("jsonl"),
			expected: []string{"query", "tools", "--output", "jsonl"},
		},
		{
			name: "project-level query with filters",
			builder: NewCommandBuilder("query", "tools").
				WithScope("project").
				WithFilter("tool", "Bash").
				WithFilter("status", "error").
				WithLimit(20).
				WithOutputFormat("jsonl"),
			expected: []string{"--project", ".", "query", "tools", "--tool", "Bash", "--status", "error", "--limit", "20", "--output", "jsonl"},
		},
		{
			name: "query user messages with pattern",
			builder: NewCommandBuilder("query", "user-messages").
				WithScope("project").
				WithRequiredParam("match", "fix.*bug").
				WithLimit(10).
				WithOutputFormat("tsv"),
			expected: []string{"--project", ".", "query", "user-messages", "--match", "fix.*bug", "--limit", "10", "--output", "tsv"},
		},
		{
			name: "stats command with session scope",
			builder: NewCommandBuilder("parse", "stats").
				WithScope("session").
				WithOutputFormat("jsonl"),
			expected: []string{"parse", "stats", "--output", "jsonl"},
		},
		{
			name: "analyze errors with project scope",
			builder: NewCommandBuilder("analyze", "errors").
				WithScope("project").
				WithOutputFormat("jsonl"),
			expected: []string{"--project", ".", "analyze", "errors", "--output", "jsonl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder.Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCommandBuilderDefaults tests default value handling
func TestCommandBuilderDefaults(t *testing.T) {
	tests := []struct {
		name     string
		builder  *CommandBuilder
		expected []string
	}{
		{
			name: "default scope is session",
			builder: NewCommandBuilder("query", "tools").
				WithOutputFormat("jsonl"),
			expected: []string{"query", "tools", "--output", "jsonl"},
		},
		{
			name: "default output format is jsonl",
			builder: NewCommandBuilder("query", "tools").
				WithScope("session"),
			expected: []string{"query", "tools", "--output", "jsonl"},
		},
		{
			name: "limit of 0 means no limit flag",
			builder: NewCommandBuilder("query", "tools").
				WithLimit(0).
				WithOutputFormat("jsonl"),
			expected: []string{"query", "tools", "--output", "jsonl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder.Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSemanticDefaults tests semantic default constants
func TestSemanticDefaults(t *testing.T) {
	assert.Equal(t, 10, DefaultLimitSmall, "Small limit should be 10")
	assert.Equal(t, 20, DefaultLimitMedium, "Medium limit should be 20")
	assert.Equal(t, 100, DefaultLimitLarge, "Large limit should be 100")
}

// TestBuildToolCommandFromArgs tests building commands from MCP arguments
func TestBuildToolCommandFromArgs(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		args        map[string]interface{}
		expected    []string
		expectError bool
	}{
		{
			name:     "query_tools with scope parameter",
			toolName: "query_tools",
			args: map[string]interface{}{
				"scope":         "project",
				"tool":          "Bash",
				"status":        "error",
				"limit":         float64(20),
				"output_format": "jsonl",
			},
			expected: []string{"--project", ".", "query", "tools", "--tool", "Bash", "--status", "error", "--limit", "20", "--output", "jsonl"},
		},
		{
			name:     "query_tools with session scope",
			toolName: "query_tools",
			args: map[string]interface{}{
				"scope":         "session",
				"limit":         float64(20),
				"output_format": "jsonl",
			},
			expected: []string{"query", "tools", "--limit", "20", "--output", "jsonl"},
		},
		{
			name:     "get_session_stats (backward compatibility)",
			toolName: "get_session_stats",
			args: map[string]interface{}{
				"output_format": "jsonl",
			},
			expected: []string{"parse", "stats", "--output", "jsonl"},
		},
		{
			name:     "query_user_messages missing required pattern",
			toolName: "query_user_messages",
			args: map[string]interface{}{
				"output_format": "jsonl",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildToolCommand(tt.toolName, tt.args)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
