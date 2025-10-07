package main

import (
	"os"
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
			expectArgs:   []string{"--project", ".", "query", "user-messages", "--pattern", "test.*pattern", "--limit", "10", "--output", "jsonl"},
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

func TestExecuteTool_InlineThresholdParameter(t *testing.T) {
	// Test with inline_threshold_bytes parameter
	args := map[string]interface{}{
		"inline_threshold_bytes": float64(4096), // Custom threshold
		"jq_filter":              ".[]",
	}

	// Test parameter extraction
	thresholdBytes := getIntParam(args, "inline_threshold_bytes", 8192)
	if thresholdBytes != 4096 {
		t.Errorf("expected inline_threshold_bytes=4096, got %d", thresholdBytes)
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

func TestExecuteTool_MessageTruncationParameters(t *testing.T) {
	tests := []struct {
		name          string
		args          map[string]interface{}
		expectMaxLen  int
		expectSummary bool
	}{
		{
			name:          "default max_message_length",
			args:          map[string]interface{}{},
			expectMaxLen:  0, // Changed from DefaultMaxMessageLength - rely on hybrid mode
			expectSummary: false,
		},
		{
			name: "custom max_message_length",
			args: map[string]interface{}{
				"max_message_length": float64(1000),
			},
			expectMaxLen:  1000,
			expectSummary: false,
		},
		{
			name: "content_summary enabled",
			args: map[string]interface{}{
				"content_summary": true,
			},
			expectMaxLen:  0, // Changed from DefaultMaxMessageLength
			expectSummary: true,
		},
		{
			name: "both parameters set",
			args: map[string]interface{}{
				"max_message_length": float64(200),
				"content_summary":    true,
			},
			expectMaxLen:  200,
			expectSummary: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Changed default from DefaultMaxMessageLength to 0 to match executor behavior
			maxLen := getIntParam(tt.args, "max_message_length", 0)
			summary := getBoolParam(tt.args, "content_summary", false)

			if maxLen != tt.expectMaxLen {
				t.Errorf("expected max_message_length=%d, got %d", tt.expectMaxLen, maxLen)
			}
			if summary != tt.expectSummary {
				t.Errorf("expected content_summary=%v, got %v", tt.expectSummary, summary)
			}
		})
	}
}

// Test parseJSONL function
func TestParseJSONL(t *testing.T) {
	executor := NewToolExecutor()
	tests := []struct {
		name      string
		jsonl     string
		expectLen int
		expectErr bool
	}{
		{
			name:      "single line",
			jsonl:     `{"id":1,"name":"test"}`,
			expectLen: 1,
			expectErr: false,
		},
		{
			name: "multiple lines",
			jsonl: `{"id":1,"name":"test1"}
{"id":2,"name":"test2"}
{"id":3,"name":"test3"}`,
			expectLen: 3,
			expectErr: false,
		},
		{
			name:      "empty string",
			jsonl:     "",
			expectLen: 0,
			expectErr: false,
		},
		{
			name:      "empty array (exit code 2 scenario)",
			jsonl:     "[]",
			expectLen: 0,
			expectErr: false, // Should handle [] as empty result
		},
		{
			name: "with empty lines",
			jsonl: `{"id":1}

{"id":2}`,
			expectLen: 2,
			expectErr: false,
		},
		{
			name:      "invalid JSON",
			jsonl:     `{"invalid": json}`,
			expectLen: 0,
			expectErr: true,
		},
		{
			name:      "mixed valid and invalid",
			jsonl:     `{"id":1}\ninvalid\n{"id":2}`,
			expectLen: 0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.parseJSONL(tt.jsonl)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != tt.expectLen {
				t.Errorf("expected %d items, got %d", tt.expectLen, len(result))
			}
		})
	}
}

// Test dataToJSONL function
func TestDataToJSONL(t *testing.T) {
	executor := NewToolExecutor()
	tests := []struct {
		name      string
		data      []interface{}
		expectLen int
		expectErr bool
	}{
		{
			name: "simple data",
			data: []interface{}{
				map[string]interface{}{"id": 1, "name": "test1"},
				map[string]interface{}{"id": 2, "name": "test2"},
			},
			expectLen: 2,
			expectErr: false,
		},
		{
			name:      "empty data",
			data:      []interface{}{},
			expectLen: 0,
			expectErr: false,
		},
		{
			name:      "nil data",
			data:      nil,
			expectLen: 0,
			expectErr: false,
		},
		{
			name: "complex nested data",
			data: []interface{}{
				map[string]interface{}{
					"id":   1,
					"meta": map[string]interface{}{"created": "2025-01-01"},
					"tags": []string{"a", "b"},
				},
			},
			expectLen: 1,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.dataToJSONL(tt.data)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Count lines in result
			lines := 0
			if result != "" {
				lines = strings.Count(result, "\n")
				// Add 1 if doesn't end with newline
				if !strings.HasSuffix(result, "\n") && result != "" {
					lines++
				}
			}

			if lines != tt.expectLen {
				t.Errorf("expected %d lines, got %d", tt.expectLen, lines)
			}

			// Verify it can be parsed back
			if result != "" {
				parsed, err := executor.parseJSONL(result)
				if err != nil {
					t.Errorf("generated JSONL cannot be parsed: %v", err)
				}
				if len(parsed) != tt.expectLen {
					t.Errorf("parsed data length mismatch: expected %d, got %d", tt.expectLen, len(parsed))
				}
			}
		})
	}
}

// Test applyMessageFiltersToData function
func TestApplyMessageFiltersToData(t *testing.T) {
	executor := NewToolExecutor()
	tests := []struct {
		name                string
		data                []interface{}
		maxMessageLength    int
		contentSummary      bool
		expectTruncated     bool
		expectSummaryFields bool
	}{
		{
			name: "no truncation needed",
			data: []interface{}{
				map[string]interface{}{"content": "short", "turn": float64(1)},
			},
			maxMessageLength:    100,
			contentSummary:      false,
			expectTruncated:     false,
			expectSummaryFields: false,
		},
		{
			name: "truncation with long content",
			data: []interface{}{
				map[string]interface{}{"content": strings.Repeat("a", 200), "turn": float64(1)},
			},
			maxMessageLength:    50,
			contentSummary:      false,
			expectTruncated:     true,
			expectSummaryFields: false,
		},
		{
			name: "content summary mode",
			data: []interface{}{
				map[string]interface{}{
					"content":       "test content",
					"turn_sequence": float64(1),
					"timestamp":     "2025-01-01",
				},
			},
			maxMessageLength:    500,
			contentSummary:      true,
			expectTruncated:     false,
			expectSummaryFields: true,
		},
		{
			name:                "empty data",
			data:                []interface{}{},
			maxMessageLength:    500,
			contentSummary:      false,
			expectTruncated:     false,
			expectSummaryFields: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.applyMessageFiltersToData(tt.data, tt.maxMessageLength, tt.contentSummary)

			if len(result) != len(tt.data) {
				t.Errorf("expected %d items, got %d", len(tt.data), len(result))
				return
			}

			if len(result) > 0 {
				item := result[0].(map[string]interface{})

				if tt.expectTruncated {
					content := item["content"].(string)
					if len(content) > tt.maxMessageLength+20 { // Allow for truncation marker
						t.Errorf("content not truncated: length %d > max %d", len(content), tt.maxMessageLength)
					}
				}

				if tt.expectSummaryFields {
					if _, hasPreview := item["content_preview"]; !hasPreview {
						t.Error("expected content_preview field in summary mode")
					}
					if _, hasContent := item["content"]; hasContent {
						t.Error("should not have full content in summary mode")
					}
				}
			}
		})
	}
}

// Additional buildCommand tests for tools not covered above
func TestBuildCommandAdditional(t *testing.T) {
	executor := NewToolExecutor()

	tests := []struct {
		name         string
		toolName     string
		args         map[string]interface{}
		scope        string
		outputFormat string
		wantArgs     []string
	}{
		{
			name:         "query_context",
			toolName:     "query_context",
			args:         map[string]interface{}{"error_signature": "file_not_found", "window": float64(3)},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "query", "context", "--error-signature", "file_not_found", "--window", "3", "--output", "jsonl"},
		},
		{
			name:         "query_tool_sequences",
			toolName:     "query_tool_sequences",
			args:         map[string]interface{}{"pattern": "Read->Edit", "min_occurrences": float64(5)},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "analyze", "sequences", "--pattern", "Read->Edit", "--min-occurrences", "5", "--output", "jsonl"},
		},
		{
			name:         "query_tool_sequences_with_builtin_filter",
			toolName:     "query_tool_sequences",
			args:         map[string]interface{}{"min_occurrences": float64(3), "include_builtin_tools": false},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "analyze", "sequences", "--min-occurrences", "3", "--output", "jsonl"},
		},
		{
			name:         "query_tool_sequences_include_builtin",
			toolName:     "query_tool_sequences",
			args:         map[string]interface{}{"min_occurrences": float64(3), "include_builtin_tools": true},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "analyze", "sequences", "--min-occurrences", "3", "--include-builtin-tools", "--output", "jsonl"},
		},
		{
			name:         "query_file_access",
			toolName:     "query_file_access",
			args:         map[string]interface{}{"file": "main.go"},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "query", "file-access", "--file", "main.go", "--output", "jsonl"},
		},
		{
			name:         "query_project_state",
			toolName:     "query_project_state",
			args:         map[string]interface{}{},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "query", "project-state", "--output", "jsonl"},
		},
		{
			name:         "query_successful_prompts",
			toolName:     "query_successful_prompts",
			args:         map[string]interface{}{"limit": float64(10), "min_quality_score": 0.85},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "query", "successful-prompts", "--limit", "10", "--min-quality-score", "0.85", "--output", "jsonl"},
		},
		{
			name:         "query_tools_advanced",
			toolName:     "query_tools_advanced",
			args:         map[string]interface{}{"where": "tool='Read'", "limit": float64(20)},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "query", "tools", "--where", "tool='Read'", "--limit", "20", "--output", "jsonl"},
		},
		{
			name:         "query_time_series",
			toolName:     "query_time_series",
			args:         map[string]interface{}{"interval": "hour", "metric": "tool-calls", "where": "status='error'"},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "stats", "timeseries", "--interval", "hour", "--metric", "tool-calls", "--where", "status='error'", "--output", "jsonl"},
		},
		{
			name:         "query_files with default threshold",
			toolName:     "query_files",
			args:         map[string]interface{}{},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "analyze", "file-churn", "--output", "jsonl"},
		},
		{
			name:         "query_files with custom threshold",
			toolName:     "query_files",
			args:         map[string]interface{}{"threshold": float64(10)},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "analyze", "file-churn", "--threshold", "10", "--output", "jsonl"},
		},
		{
			name:         "query_files ignores unsupported parameters",
			toolName:     "query_files",
			args:         map[string]interface{}{"sort_by": "total_ops", "top": float64(10), "where": "ext='go'", "threshold": float64(3)},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     []string{"--project", ".", "analyze", "file-churn", "--threshold", "3", "--output", "jsonl"},
		},
		{
			name:         "cleanup_temp_files",
			toolName:     "cleanup_temp_files",
			args:         map[string]interface{}{},
			scope:        "project",
			outputFormat: "jsonl",
			wantArgs:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.buildCommand(tt.toolName, tt.args, tt.scope, tt.outputFormat)

			if tt.wantArgs == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.wantArgs) {
				t.Errorf("expected %d args, got %d\nExpected: %v\nGot: %v", len(tt.wantArgs), len(result), tt.wantArgs, result)
				return
			}

			for i, arg := range tt.wantArgs {
				if result[i] != arg {
					t.Errorf("arg %d mismatch: expected %q, got %q", i, arg, result[i])
				}
			}
		})
	}
}

// Test getSessionHash with different environment variables
func TestGetSessionHash(t *testing.T) {
	// Save original env vars
	origSessionID := os.Getenv("CC_SESSION_ID")
	origProjectHash := os.Getenv("CC_PROJECT_HASH")
	defer func() {
		os.Setenv("CC_SESSION_ID", origSessionID)
		os.Setenv("CC_PROJECT_HASH", origProjectHash)
	}()

	tests := []struct {
		name           string
		sessionID      string
		projectHash    string
		expectNotEmpty bool
	}{
		{
			name:           "with session ID",
			sessionID:      "abc123-def456-ghi789",
			projectHash:    "",
			expectNotEmpty: true,
		},
		{
			name:           "with project hash",
			sessionID:      "",
			projectHash:    "project-hash-123",
			expectNotEmpty: true,
		},
		{
			name:           "with both",
			sessionID:      "session-abc",
			projectHash:    "project-xyz",
			expectNotEmpty: true,
		},
		{
			name:           "with neither",
			sessionID:      "",
			projectHash:    "",
			expectNotEmpty: true, // Falls back to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CC_SESSION_ID", tt.sessionID)
			os.Setenv("CC_PROJECT_HASH", tt.projectHash)

			result := getSessionHash()

			if tt.expectNotEmpty && result == "" {
				t.Error("expected non-empty session hash")
			}

			// Check that result is reasonable length (8 chars for session hash prefix)
			if tt.sessionID != "" && len(result) < 8 {
				t.Errorf("session hash too short: %s", result)
			}
		})
	}
}

// TestExecuteMetaCC tests meta-cc command execution with a mock binary
func TestExecuteMetaCC(t *testing.T) {
	// Create a temporary test script that simulates meta-cc
	testScript := `#!/bin/bash
if [[ "$1" == "parse" && "$2" == "stats" ]]; then
	echo '{"total_turns":10,"tool_calls":25}'
	exit 0
elif [[ "$1" == "query" && "$2" == "tools" ]]; then
	if [[ "$3" == "--status" && "$4" == "error" ]]; then
		# Simulate no results scenario (exit code 2)
		echo '[]'
		echo "Warning: No results found" >&2
		exit 2
	else
		echo '{"tool":"Bash","count":5}'
		echo '{"tool":"Read","count":3}'
		exit 0
	fi
else
	echo "unknown command" >&2
	exit 1
fi
`
	// Write test script to temporary file
	tmpFile, err := os.CreateTemp("", "mock-meta-cc-*.sh")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testScript); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	// Make it executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}

	tests := []struct {
		name        string
		cmdArgs     []string
		expectError bool
		expectOut   string
	}{
		{
			name:        "successful parse stats command",
			cmdArgs:     []string{"parse", "stats", "--output", "jsonl"},
			expectError: false,
			expectOut:   "total_turns",
		},
		{
			name:        "successful query tools command",
			cmdArgs:     []string{"query", "tools", "--output", "jsonl"},
			expectError: false,
			expectOut:   "Bash",
		},
		{
			name:        "exit code 2 (no results) should not be an error",
			cmdArgs:     []string{"query", "tools", "--status", "error"},
			expectError: false, // Exit code 2 should NOT be treated as error
			expectOut:   "[]",  // Should return stdout content (empty array)
		},
		{
			name:        "unknown command returns error",
			cmdArgs:     []string{"unknown", "command"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &ToolExecutor{metaCCPath: tmpFile.Name()}
			output, err := executor.executeMetaCC(tt.cmdArgs)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if !strings.Contains(output, tt.expectOut) {
					t.Errorf("expected output to contain %q, got %q", tt.expectOut, output)
				}
			}
		})
	}
}

// TestExecuteTool tests the full ExecuteTool flow with mock binary
func TestExecuteTool(t *testing.T) {
	// Create mock meta-cc script
	testScript := `#!/bin/bash
# Handle both session and project scopes
if [[ "$1" == "--project" ]]; then
	shift 2  # Skip --project and path
fi

if [[ "$1" == "parse" && "$2" == "stats" ]]; then
	echo '{"total_turns":10,"tool_calls":25,"errors":2}'
	exit 0
elif [[ "$1" == "query" && "$2" == "tools" ]]; then
	echo '{"tool":"Bash","status":"success","count":5}'
	echo '{"tool":"Read","status":"success","count":3}'
	exit 0
elif [[ "$1" == "query" && "$2" == "user-messages" ]]; then
	echo '{"turn":1,"timestamp":"2025-01-01T00:00:00Z","content":"test message with long content that should be truncated if max_message_length is set"}'
	exit 0
else
	echo "command not implemented" >&2
	exit 1
fi
`
	tmpFile, err := os.CreateTemp("", "mock-meta-cc-*.sh")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testScript); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}
	tmpFile.Close()
	os.Chmod(tmpFile.Name(), 0755)

	executor := &ToolExecutor{metaCCPath: tmpFile.Name()}

	tests := []struct {
		name        string
		toolName    string
		args        map[string]interface{}
		expectError bool
		expectOut   string
	}{
		{
			name:     "get_session_stats",
			toolName: "get_session_stats",
			args: map[string]interface{}{
				"scope":         "session",
				"output_format": "jsonl",
			},
			expectError: false,
			expectOut:   "total_turns",
		},
		{
			name:     "query_tools with jq filter",
			toolName: "query_tools",
			args: map[string]interface{}{
				"scope":         "project",
				"jq_filter":     ".[]",
				"output_format": "jsonl",
			},
			expectError: false,
			expectOut:   "Bash",
		},
		{
			name:     "query_user_messages with content summary",
			toolName: "query_user_messages",
			args: map[string]interface{}{
				"scope":           "session",
				"pattern":         "test",
				"content_summary": true,
				"output_format":   "jsonl",
			},
			expectError: false,
			expectOut:   "turn",
		},
		{
			name:     "query_user_messages with max_message_length",
			toolName: "query_user_messages",
			args: map[string]interface{}{
				"scope":              "session",
				"pattern":            "test",
				"max_message_length": 50,
				"output_format":      "jsonl",
			},
			expectError: false,
			expectOut:   "content",
		},
		{
			name:     "stats_only mode",
			toolName: "get_session_stats",
			args: map[string]interface{}{
				"scope":         "session",
				"stats_only":    true,
				"output_format": "jsonl",
			},
			expectError: false,
			expectOut:   "count",
		},
		{
			name:     "stats_first mode",
			toolName: "query_tools",
			args: map[string]interface{}{
				"scope":         "project",
				"stats_first":   true,
				"output_format": "jsonl",
			},
			expectError: false,
			expectOut:   "---",
		},
		{
			name:     "unknown tool returns error",
			toolName: "unknown_tool",
			args: map[string]interface{}{
				"output_format": "jsonl",
			},
			expectError: true,
		},
		{
			name:     "cleanup_temp_files tool",
			toolName: "cleanup_temp_files",
			args: map[string]interface{}{
				"max_age_hours": float64(24),
			},
			expectError: false,
			expectOut:   "freed_bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executor.ExecuteTool(tt.toolName, tt.args)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if tt.expectOut != "" && !strings.Contains(output, tt.expectOut) {
					t.Errorf("expected output to contain %q, got %q", tt.expectOut, output)
				}
			}
		})
	}
}
