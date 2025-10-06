package main

import (
	"strings"
	"testing"
)

func TestNewToolExecutor(t *testing.T) {
	executor := NewToolExecutor()

	if executor == nil {
		t.Fatal("expected executor to be created")
	}

	if executor.metaCCPath == "" {
		t.Error("expected metaCCPath to be set")
	}
}

func TestGetStringParam(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		key      string
		defVal   string
		expected string
	}{
		{
			name:     "existing string parameter",
			args:     map[string]interface{}{"tool": "Bash"},
			key:      "tool",
			defVal:   "default",
			expected: "Bash",
		},
		{
			name:     "missing parameter uses default",
			args:     map[string]interface{}{},
			key:      "tool",
			defVal:   "default",
			expected: "default",
		},
		{
			name:     "non-string parameter uses default",
			args:     map[string]interface{}{"tool": 123},
			key:      "tool",
			defVal:   "default",
			expected: "default",
		},
		{
			name:     "nil args uses default",
			args:     nil,
			key:      "tool",
			defVal:   "default",
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringParam(tt.args, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetBoolParam(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		key      string
		defVal   bool
		expected bool
	}{
		{
			name:     "existing true parameter",
			args:     map[string]interface{}{"stats_only": true},
			key:      "stats_only",
			defVal:   false,
			expected: true,
		},
		{
			name:     "existing false parameter",
			args:     map[string]interface{}{"stats_only": false},
			key:      "stats_only",
			defVal:   true,
			expected: false,
		},
		{
			name:     "missing parameter uses default",
			args:     map[string]interface{}{},
			key:      "stats_only",
			defVal:   true,
			expected: true,
		},
		{
			name:     "non-bool parameter uses default",
			args:     map[string]interface{}{"stats_only": "true"},
			key:      "stats_only",
			defVal:   false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBoolParam(tt.args, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetIntParam(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		key      string
		defVal   int
		expected int
	}{
		{
			name:     "existing int parameter as float64",
			args:     map[string]interface{}{"limit": float64(10)},
			key:      "limit",
			defVal:   20,
			expected: 10,
		},
		{
			name:     "missing parameter uses default",
			args:     map[string]interface{}{},
			key:      "limit",
			defVal:   20,
			expected: 20,
		},
		{
			name:     "non-numeric parameter uses default",
			args:     map[string]interface{}{"limit": "10"},
			key:      "limit",
			defVal:   20,
			expected: 20,
		},
		{
			name:     "zero value",
			args:     map[string]interface{}{"limit": float64(0)},
			key:      "limit",
			defVal:   20,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIntParam(tt.args, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetFloatParam(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		key      string
		defVal   float64
		expected float64
	}{
		{
			name:     "existing float parameter",
			args:     map[string]interface{}{"threshold": 0.8},
			key:      "threshold",
			defVal:   0.5,
			expected: 0.8,
		},
		{
			name:     "missing parameter uses default",
			args:     map[string]interface{}{},
			key:      "threshold",
			defVal:   0.5,
			expected: 0.5,
		},
		{
			name:     "non-numeric parameter uses default",
			args:     map[string]interface{}{"threshold": "0.8"},
			key:      "threshold",
			defVal:   0.5,
			expected: 0.5,
		},
		{
			name:     "zero value",
			args:     map[string]interface{}{"threshold": 0.0},
			key:      "threshold",
			defVal:   0.5,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFloatParam(tt.args, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestBuildCommand(t *testing.T) {
	executor := NewToolExecutor()

	tests := []struct {
		name         string
		toolName     string
		args         map[string]interface{}
		scope        string
		outputFormat string
		expectNil    bool
		expectArgs   []string
	}{
		{
			name:         "query_tools with project scope",
			toolName:     "query_tools",
			args:         map[string]interface{}{"tool": "Bash", "status": "error"},
			scope:        "project",
			outputFormat: "jsonl",
			expectNil:    false,
			expectArgs:   []string{"--project", ".", "query", "tools", "--tool", "Bash", "--status", "error", "--output", "jsonl"},
		},
		{
			name:         "query_tools with session scope",
			toolName:     "query_tools",
			args:         map[string]interface{}{},
			scope:        "session",
			outputFormat: "jsonl",
			expectNil:    false,
			expectArgs:   []string{"query", "tools", "--output", "jsonl"},
		},
		{
			name:         "get_session_stats",
			toolName:     "get_session_stats",
			args:         map[string]interface{}{},
			scope:        "session",
			outputFormat: "jsonl",
			expectNil:    false,
			expectArgs:   []string{"parse", "stats", "--output", "jsonl"},
		},
		{
			name:         "query_user_messages with pattern",
			toolName:     "query_user_messages",
			args:         map[string]interface{}{"pattern": "test.*pattern", "limit": float64(10)},
			scope:        "project",
			outputFormat: "jsonl",
			expectNil:    false,
			expectArgs:   []string{"--project", ".", "query", "user-messages", "--match", "test.*pattern", "--limit", "10", "--output", "jsonl"},
		},
		{
			name:         "unknown tool",
			toolName:     "unknown_tool",
			args:         map[string]interface{}{},
			scope:        "project",
			outputFormat: "jsonl",
			expectNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.buildCommand(tt.toolName, tt.args, tt.scope, tt.outputFormat)

			if tt.expectNil {
				if result != nil {
					t.Errorf("expected nil for unknown tool, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected command args, got nil")
			}

			// Check that essential args are present
			resultStr := strings.Join(result, " ")
			for _, arg := range tt.expectArgs {
				if !strings.Contains(resultStr, arg) {
					t.Errorf("expected command to contain %s, got %v", arg, result)
				}
			}
		})
	}
}

func TestExecuteTool_OutputTruncation(t *testing.T) {
	// Test with max_output_bytes parameter
	args := map[string]interface{}{
		"max_output_bytes": float64(100), // Very small limit
		"jq_filter":        ".[]",
	}

	// The actual execution will fail since we don't have a real session,
	// but we're testing the parameter extraction
	maxBytes := getIntParam(args, "max_output_bytes", DefaultMaxOutputBytes)
	if maxBytes != 100 {
		t.Errorf("expected max_output_bytes=100, got %d", maxBytes)
	}
}

func TestExecuteTool_JQFilterParameter(t *testing.T) {
	args := map[string]interface{}{
		"jq_filter": ".[] | select(.Status == \"error\")",
	}

	jqFilter := getStringParam(args, "jq_filter", ".[]")
	if jqFilter != ".[] | select(.Status == \"error\")" {
		t.Errorf("expected jq_filter to be extracted correctly, got %s", jqFilter)
	}
}

func TestExecuteTool_StatsParameters(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		expectStats bool
		expectFirst bool
	}{
		{
			name:        "stats_only true",
			args:        map[string]interface{}{"stats_only": true},
			expectStats: true,
			expectFirst: false,
		},
		{
			name:        "stats_first true",
			args:        map[string]interface{}{"stats_first": true},
			expectStats: false,
			expectFirst: true,
		},
		{
			name:        "both false",
			args:        map[string]interface{}{},
			expectStats: false,
			expectFirst: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statsOnly := getBoolParam(tt.args, "stats_only", false)
			statsFirst := getBoolParam(tt.args, "stats_first", false)

			if statsOnly != tt.expectStats {
				t.Errorf("expected stats_only=%v, got %v", tt.expectStats, statsOnly)
			}
			if statsFirst != tt.expectFirst {
				t.Errorf("expected stats_first=%v, got %v", tt.expectFirst, statsFirst)
			}
		})
	}
}
