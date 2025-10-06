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
			name: "project-level query with filter",
			builder: NewCommandBuilder("query", "tools").
				WithScope("project").
				WithFilter("tool", "Bash").
				WithLimit(20).
				WithOutputFormat("jsonl"),
			expected: []string{"--project", ".", "query", "tools", "--tool", "Bash", "--limit", "20", "--output", "jsonl"},
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
			name: "default scope is project",
			builder: NewCommandBuilder("query", "tools").
				WithOutputFormat("jsonl"),
			expected: []string{"--project", ".", "query", "tools", "--output", "jsonl"},
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
			expected: []string{"--project", ".", "query", "tools", "--output", "jsonl"},
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
				"limit":         float64(20),
				"output_format": "jsonl",
			},
			expected: []string{"--project", ".", "query", "tools", "--tool", "Bash", "--limit", "20", "--output", "jsonl"},
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

// TestWithExtraFlag tests adding extra flags to command builder
func TestWithExtraFlag(t *testing.T) {
	tests := []struct {
		name     string
		builder  *CommandBuilder
		expected []string
	}{
		{
			name: "extra flag with integer value",
			builder: NewCommandBuilder("query", "context").
				WithScope("project").
				WithExtraFlag("window", 5).
				WithOutputFormat("jsonl"),
			expected: []string{"--project", ".", "query", "context", "--window", "5", "--output", "jsonl"},
		},
		{
			name: "extra flag with string value",
			builder: NewCommandBuilder("stats", "aggregate").
				WithScope("project").
				WithExtraFlag("group-by", "tool").
				WithOutputFormat("jsonl"),
			expected: []string{"--project", ".", "stats", "aggregate", "--group-by", "tool", "--output", "jsonl"},
		},
		{
			name: "multiple extra flags",
			builder: NewCommandBuilder("query", "tool-sequences").
				WithScope("project").
				WithExtraFlag("min-occurrences", 3).
				WithExtraFlag("pattern", "Read -> Edit").
				WithOutputFormat("jsonl"),
			expected: []string{"--project", ".", "query", "tool-sequences", "--min-occurrences", "3", "--pattern", "Read -> Edit", "--output", "jsonl"},
		},
		{
			name: "extra flag with float value",
			builder: NewCommandBuilder("query", "successful-prompts").
				WithScope("project").
				WithExtraFlag("min-quality-score", 0.75).
				WithOutputFormat("jsonl"),
			expected: []string{"--project", ".", "query", "successful-prompts", "--min-quality-score", "0.75", "--output", "jsonl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder.Build()
			// Note: Map iteration order is not guaranteed, so we check that all expected flags are present
			// rather than exact order match for extra flags
			for _, flag := range tt.expected {
				assert.Contains(t, result, flag, "Expected flag %s to be in result", flag)
			}
		})
	}
}

// TestGetFloatArg tests float parameter extraction
func TestGetFloatArg(t *testing.T) {
	tests := []struct {
		name         string
		args         map[string]interface{}
		key          string
		defaultValue float64
		expected     float64
	}{
		{
			name: "existing float parameter",
			args: map[string]interface{}{
				"min_quality_score": 0.8,
			},
			key:          "min_quality_score",
			defaultValue: 0.5,
			expected:     0.8,
		},
		{
			name:         "missing parameter uses default",
			args:         map[string]interface{}{},
			key:          "min_quality_score",
			defaultValue: 0.5,
			expected:     0.5,
		},
		{
			name: "non-float parameter uses default",
			args: map[string]interface{}{
				"min_quality_score": "0.8",
			},
			key:          "min_quality_score",
			defaultValue: 0.5,
			expected:     0.5,
		},
		{
			name: "zero value",
			args: map[string]interface{}{
				"threshold": 0.0,
			},
			key:          "threshold",
			defaultValue: 0.5,
			expected:     0.0,
		},
		{
			name: "negative value",
			args: map[string]interface{}{
				"threshold": -0.5,
			},
			key:          "threshold",
			defaultValue: 0.0,
			expected:     -0.5,
		},
		{
			name: "value greater than 1",
			args: map[string]interface{}{
				"threshold": 1.5,
			},
			key:          "threshold",
			defaultValue: 1.0,
			expected:     1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFloatArg(tt.args, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
