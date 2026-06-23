package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/config"
	pipelinepkg "github.com/yaleh/meta-cc/internal/mcp/pipeline"
)

const testSessionID = "test-session"

func writeSessionFixture(t *testing.T, projectPath, sessionID, content string) {
	t.Helper()

	projectsRoot := os.Getenv("META_CC_PROJECTS_ROOT")
	if projectsRoot == "" {
		t.Fatal("META_CC_PROJECTS_ROOT must be set for tests")
	}

	// Resolve symlinks for consistent hashing on macOS (/var -> /private/var)
	resolvedPath, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		// If path doesn't exist yet, use original path
		resolvedPath = projectPath
	}

	hash := strings.ReplaceAll(resolvedPath, "\\", "-")
	hash = strings.ReplaceAll(hash, "/", "-")
	hash = strings.ReplaceAll(hash, ":", "-")

	sessionDir := filepath.Join(projectsRoot, hash)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write session fixture: %v", err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(sessionDir) })
}

func setupLibraryFixture(t *testing.T) func() {
	projectDir := t.TempDir()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	fixture := `{"type":"user","timestamp":"2025-10-02T09:59:59Z","uuid":"uuid-0","sessionId":"` + testSessionID + `","message":{"role":"user","content":[{"type":"text","text":"run analysis"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:00:00Z","uuid":"uuid-1","sessionId":"` + testSessionID + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","timestamp":"2025-10-02T10:00:01Z","uuid":"uuid-2","sessionId":"` + testSessionID + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"file.txt"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:00:02Z","uuid":"uuid-3","sessionId":"` + testSessionID + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/tmp/file.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:00:03Z","uuid":"uuid-4","sessionId":"` + testSessionID + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"file contents"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:00:04Z","uuid":"uuid-5","sessionId":"` + testSessionID + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-3","name":"meta-cc-run","input":{"command":"meta"}}]}}
{"type":"user","timestamp":"2025-10-02T10:00:05Z","uuid":"uuid-6","sessionId":"` + testSessionID + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-3","content":"ok"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:00:06Z","uuid":"uuid-7","sessionId":"` + testSessionID + `","message":{"role":"assistant","content":[{"type":"text","text":"Completed task"}]}}
	{"type":"user","timestamp":"2025-10-02T10:00:07Z","uuid":"uuid-8","sessionId":"` + testSessionID + `","message":{"role":"user","content":[{"type":"text","text":"test message with long content that should be truncated if max_message_length is set"}]}}
`

	writeSessionFixture(t, projectDir, testSessionID, fixture)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to chdir to project dir: %v", err)
	}

	return func() {
		_ = os.Chdir(oldWd)
	}
}

func TestNewToolExecutor(t *testing.T) {
	executor := NewToolExecutor()

	if executor == nil {
		t.Fatal("expected executor to be created")
	}

	// Note: metaCCPath removed - all tools now use internal/query library
}

func TestNewToolPipelineConfig(t *testing.T) {
	args := map[string]interface{}{
		"jq_filter":          ".[] | .name",
		"stats_only":         true,
		"stats_first":        false,
		"output_format":      "json",
		"max_message_length": float64(120),
		"content_summary":    true,
	}

	config := newToolPipelineConfig(args)

	if config.jqFilter != ".[] | .name" {
		t.Fatalf("unexpected jqFilter: %s", config.jqFilter)
	}
	if !config.statsOnly {
		t.Fatal("expected statsOnly to be true")
	}
	if config.statsFirst {
		t.Fatal("expected statsFirst to be false")
	}
	if config.outputFormat != "json" {
		t.Fatalf("unexpected outputFormat: %s", config.outputFormat)
	}
	if config.maxMessageLength != 120 {
		t.Fatalf("expected maxMessageLength to be 120, got %d", config.maxMessageLength)
	}
	if !config.contentSummary {
		t.Fatal("expected contentSummary to be true")
	}

	defaults := newToolPipelineConfig(map[string]interface{}{})
	if defaults.jqFilter != ".[]" {
		t.Fatalf("unexpected default jqFilter: %s", defaults.jqFilter)
	}
	if defaults.outputFormat != "jsonl" {
		t.Fatalf("unexpected default outputFormat: %s", defaults.outputFormat)
	}
	if defaults.maxMessageLength != 0 {
		t.Fatalf("expected default maxMessageLength to be 0, got %d", defaults.maxMessageLength)
	}
}

func TestToolPipelineConfigRequiresMessageFilters(t *testing.T) {
	cases := []struct {
		name   string
		cfg    toolPipelineConfig
		expect bool
	}{
		{
			name:   "no filters",
			cfg:    toolPipelineConfig{},
			expect: false,
		},
		{
			name: "max length",
			cfg: toolPipelineConfig{
				maxMessageLength: 80,
			},
			expect: true,
		},
		{
			name: "content summary",
			cfg: toolPipelineConfig{
				contentSummary: true,
			},
			expect: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.requiresMessageFilters(); got != tc.expect {
				t.Fatalf("requiresMessageFilters() = %v, expect %v", got, tc.expect)
			}
		})
	}
}

// TestScopeArgs removed - scopeArgs function deleted as part of Phase 23 CLI removal

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

// TestBuildCommand removed - buildCommand function deleted as part of Phase 23 CLI removal
// All query tools now use internal/query library directly instead of spawning CLI subprocess

func TestExecuteTool_InlineThresholdParameter(t *testing.T) {
	// Test with inline_threshold_bytes parameter
	args := map[string]interface{}{
		"inline_threshold_bytes": float64(4096), // Custom threshold
		"jq_filter":              ".[]",
	}

	// Test parameter extraction
	thresholdBytes := getIntParam(args, "inline_threshold_bytes", 32768)
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

// Test DataToJSONL function (now in pipeline package)
func TestDataToJSONL(t *testing.T) {
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
			result, err := pipelinepkg.DataToJSONL(tt.data)

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
			result := executor.applyMessageFiltersToData(tt.data, tt.maxMessageLength, tt.contentSummary, DefaultPreviewLength)

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

// TestBuildCommandAdditional removed - buildCommand function deleted as part of Phase 23 CLI removal

// Test getSessionHash fallback behavior (env vars no longer used)
func TestGetSessionHash(t *testing.T) {
	// Save original env vars
	origSessionID := os.Getenv("CC_SESSION_ID")
	origProjectHash := os.Getenv("CC_PROJECT_HASH")
	defer func() {
		if origSessionID != "" {
			os.Setenv("CC_SESSION_ID", origSessionID)
		} else {
			os.Unsetenv("CC_SESSION_ID")
		}
		if origProjectHash != "" {
			os.Setenv("CC_PROJECT_HASH", origProjectHash)
		} else {
			os.Unsetenv("CC_PROJECT_HASH")
		}
	}()

	// Clear env vars
	os.Unsetenv("CC_SESSION_ID")
	os.Unsetenv("CC_PROJECT_HASH")

	cfg, _ := config.Load()

	tests := []struct {
		name         string
		expectedHash string
	}{
		{
			name:         "should return unknown when env vars not set",
			expectedHash: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSessionHash(cfg)

			if result != tt.expectedHash {
				t.Errorf("expected session hash '%s', got '%s'", tt.expectedHash, result)
			}
		})
	}
}

// TestExecuteMetaCC removed - executeMetaCC function deleted as part of Phase 23 CLI removal
// All query tools now use internal/query library directly. See executor_no_cli_test.go for
// tests verifying that tools don't attempt CLI execution.

// TestStatsDispatch verifies that BuildStatsOnlyResponse uses timestamp-based stats for
// user-message tools and tool-name stats for tool-record tools (Phase 49).
func TestStatsDispatch(t *testing.T) {
	// User message records (no tool field, have timestamp + sessionId)
	userRecords := []interface{}{
		map[string]interface{}{
			"type":      "user",
			"timestamp": "2026-03-09T06:10:00Z",
			"sessionId": "sess-A",
		},
		map[string]interface{}{
			"type":      "user",
			"timestamp": "2026-03-09T07:20:00Z",
			"sessionId": "sess-B",
		},
	}

	// Tool records (have tool field)
	toolRecords := []interface{}{
		map[string]interface{}{
			"tool":   "Bash",
			"status": "error",
		},
		map[string]interface{}{
			"tool":   "Bash",
			"status": "error",
		},
		map[string]interface{}{
			"tool":   "Read",
			"status": "success",
		},
	}

	// User-message / content tools should use timestamp stats (output should have "hour" key)
	for _, toolName := range []string{"query_session_content", "query_session_signals"} {
		t.Run(toolName+"_uses_timestamp_stats", func(t *testing.T) {
			output, err := pipelinepkg.BuildStatsOnlyResponse(userRecords, true, "turn")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(output, `"hour"`) {
				t.Errorf("tool %s should use timestamp stats (containing 'hour'), got: %s", toolName, output)
			}
			if strings.Contains(output, `"key":"unknown"`) {
				t.Errorf("tool %s should NOT use tool stats (containing 'key:unknown'), got: %s", toolName, output)
			}
		})
	}

	// Tool-record tools should use tool-name stats (output should have "key" field, not "hour")
	for _, toolName := range []string{"query_file_activity"} {
		t.Run(toolName+"_uses_tool_stats", func(t *testing.T) {
			output, err := pipelinepkg.BuildStatsOnlyResponse(toolRecords, false, "turn")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(output, `"key"`) {
				t.Errorf("tool %s should use tool stats (containing 'key'), got: %s", toolName, output)
			}
			if strings.Contains(output, `"hour"`) {
				t.Errorf("tool %s should NOT use timestamp stats (containing 'hour'), got: %s", toolName, output)
			}
		})
	}
}

// TestQueryToolsNotRegistered verifies that query and query_raw tools are NOT registered
// Phase 27 Stage 27.1: Delete old query interfaces
func TestQueryToolsNotRegistered(t *testing.T) {
	executor := NewToolExecutor()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Test that query tool returns unknown tool error
	_, err = executor.ExecuteTool(cfg, "query", map[string]interface{}{})
	if err == nil {
		t.Error("expected error for query tool, got nil")
	}
	if !strings.Contains(err.Error(), "unknown tool") {
		t.Errorf("expected 'unknown tool' error for query, got: %v", err)
	}

	// Test that query_raw tool returns unknown tool error
	_, err = executor.ExecuteTool(cfg, "query_raw", map[string]interface{}{"jq_expression": "."})
	if err == nil {
		t.Error("expected error for query_raw tool, got nil")
	}
	if !strings.Contains(err.Error(), "unknown tool") {
		t.Errorf("expected 'unknown tool' error for query_raw, got: %v", err)
	}
}

func setupTwoSessionFixture(t *testing.T) func() {
	t.Helper()
	projectDir := t.TempDir()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	fixture := "{\"type\":\"user\",\"timestamp\":\"2026-03-09T06:00:00Z\",\"uuid\":\"u1\",\"sessionId\":\"sess-A\",\"message\":{\"role\":\"user\",\"content\":\"hello\"}}\n" +
		"{\"type\":\"user\",\"timestamp\":\"2026-03-09T07:00:00Z\",\"uuid\":\"u2\",\"sessionId\":\"sess-B\",\"message\":{\"role\":\"user\",\"content\":\"world\"}}\n"

	writeSessionFixture(t, projectDir, "two-sessions", fixture)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	return func() { _ = os.Chdir(oldWd) }
}

func TestStatsFirstWithContentSummary(t *testing.T) {
	cleanup := setupTwoSessionFixture(t)
	defer cleanup()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":            "user",
		"pattern":         ".",
		"stats_first":     true,
		"content_summary": true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// stats_first output: "<stats lines>\n---\n<detail>"
	parts := strings.SplitN(output, "\n---\n", 2)
	if len(parts) < 2 {
		t.Fatalf("expected stats separator in output, got: %s", output)
	}

	// Parse first line of stats as JSON
	firstLine := strings.SplitN(strings.TrimSpace(parts[0]), "\n", 2)[0]
	var summary map[string]interface{}
	if err := json.Unmarshal([]byte(firstLine), &summary); err != nil {
		t.Fatalf("failed to parse stats summary: %v", err)
	}

	sessionCount := int(summary["session_count"].(float64))
	if sessionCount != 2 {
		t.Errorf("session_count = %d, want 2", sessionCount)
	}

	// Detail section should have content_preview (content_summary applied)
	if !strings.Contains(parts[1], "content_preview") {
		t.Errorf("detail section should contain content_preview, got: %s", parts[1])
	}
}

func TestStatsOnlyWithContentSummary(t *testing.T) {
	cleanup := setupTwoSessionFixture(t)
	defer cleanup()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":            "user",
		"pattern":         ".",
		"stats_only":      true,
		"content_summary": true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	firstLine := strings.SplitN(strings.TrimSpace(output), "\n", 2)[0]
	var summary map[string]interface{}
	if err := json.Unmarshal([]byte(firstLine), &summary); err != nil {
		t.Fatalf("failed to parse stats: %v", err)
	}

	sessionCount := int(summary["session_count"].(float64))
	if sessionCount != 2 {
		t.Errorf("session_count = %d, want 2", sessionCount)
	}
}

func TestStatsFirstWithoutContentSummary(t *testing.T) {
	cleanup := setupTwoSessionFixture(t)
	defer cleanup()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":        "user",
		"pattern":     ".",
		"stats_first": true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parts := strings.SplitN(output, "\n---\n", 2)
	firstLine := strings.SplitN(strings.TrimSpace(parts[0]), "\n", 2)[0]
	var summary map[string]interface{}
	if err := json.Unmarshal([]byte(firstLine), &summary); err != nil {
		t.Fatalf("failed to parse stats: %v", err)
	}

	sessionCount := int(summary["session_count"].(float64))
	if sessionCount != 2 {
		t.Errorf("session_count = %d, want 2", sessionCount)
	}
}

// TestPreviewLengthParameter tests the preview_length parameter for query_user_messages
func TestPreviewLengthParameter(t *testing.T) {
	// Build a fixture with content longer than 20 chars
	projectDir := t.TempDir()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	longContent := strings.Repeat("x", 200)
	fixture := "{\"type\":\"user\",\"timestamp\":\"2026-03-09T06:00:00Z\",\"uuid\":\"u1\",\"sessionId\":\"sess-A\",\"message\":{\"role\":\"user\",\"content\":\"" + longContent + "\"}}\n"

	writeSessionFixture(t, projectDir, "preview-session", fixture)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	t.Run("content_summary=true, preview_length=20: all previews ≤ 20 runes", func(t *testing.T) {
		output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
			"role":            "user",
			"pattern":         ".",
			"content_summary": true,
			"preview_length":  float64(20),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			var item map[string]interface{}
			if err := json.Unmarshal([]byte(line), &item); err != nil {
				continue
			}
			preview, ok := item["content_preview"].(string)
			if !ok {
				continue
			}
			runeCount := len([]rune(preview))
			// Account for "..." suffix (3 runes)
			contentRunes := runeCount
			if strings.HasSuffix(preview, "...") {
				contentRunes = len([]rune(strings.TrimSuffix(preview, "...")))
			}
			if contentRunes > 20 {
				t.Errorf("content_preview has %d runes, expected ≤ 20: %q", contentRunes, preview)
			}
		}
	})

	t.Run("content_summary=false, preview_length=20: no error, full content returned", func(t *testing.T) {
		output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
			"role":           "user",
			"pattern":        ".",
			"preview_length": float64(20),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output == "" {
			t.Error("expected non-empty output")
		}
	})
}

// TestConsolidatedQueryToolsRegistered verifies that the 3 consolidated query tools are registered
// TASK-7: Replaced 10 old query_* tools with 3 consolidated tools
func TestConsolidatedQueryToolsRegistered(t *testing.T) {
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	consolidatedTools := []struct {
		name string
		args map[string]interface{}
	}{
		{"query_session_content", map[string]interface{}{"role": "user"}},
		{"query_session_signals", map[string]interface{}{"type": "errors"}},
		{"query_file_activity", map[string]interface{}{"type": "snapshots"}},
	}

	executor := NewToolExecutor()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	for _, tc := range consolidatedTools {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// These tools should be registered, but will fail with "no sessions found"
			// in test environment. We just need to verify they don't return "unknown tool" error.
			_, err := executor.ExecuteTool(cfg, tc.name, tc.args)

			// The error should NOT be "unknown tool"
			if err != nil && strings.Contains(err.Error(), "unknown tool") {
				t.Errorf("tool %s should be registered but got 'unknown tool' error", tc.name)
			}
		})
	}
}

// TestShortcutQueryToolsRegistered was renamed to TestConsolidatedQueryToolsRegistered in TASK-7.
// The old 10 shortcut tools have been removed; this is kept as a placeholder.
func TestShortcutQueryToolsRegistered(t *testing.T) {
	t.Skip("TASK-7: Old shortcut tools removed; see TestConsolidatedQueryToolsRegistered")
}

// setupGroupBySessionFixture creates two sessions with 2 turns each
func setupGroupBySessionFixture(t *testing.T) func() {
	t.Helper()
	projectDir := t.TempDir()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	fixture := "{\"type\":\"user\",\"timestamp\":\"2026-03-09T06:00:00Z\",\"uuid\":\"u1\",\"sessionId\":\"grp-sess-A\",\"message\":{\"role\":\"user\",\"content\":\"hello from A\"}}\n" +
		"{\"type\":\"user\",\"timestamp\":\"2026-03-09T06:01:00Z\",\"uuid\":\"u2\",\"sessionId\":\"grp-sess-A\",\"message\":{\"role\":\"user\",\"content\":\"second from A\"}}\n" +
		"{\"type\":\"user\",\"timestamp\":\"2026-03-09T07:00:00Z\",\"uuid\":\"u3\",\"sessionId\":\"grp-sess-B\",\"message\":{\"role\":\"user\",\"content\":\"hello from B\"}}\n" +
		"{\"type\":\"user\",\"timestamp\":\"2026-03-09T07:01:00Z\",\"uuid\":\"u4\",\"sessionId\":\"grp-sess-B\",\"message\":{\"role\":\"user\",\"content\":\"second from B\"}}\n"

	writeSessionFixture(t, projectDir, "group-sessions", fixture)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	return func() { _ = os.Chdir(oldWd) }
}

// extractGroupedResults extracts the slice of session objects from an ExecuteTool output.
// Handles both "inline" mode (data key) and "file_ref" mode (reads temp file).
func extractGroupedResults(t *testing.T, output string) []interface{} {
	t.Helper()

	var responseObj map[string]interface{}
	if err := json.Unmarshal([]byte(output), &responseObj); err != nil {
		t.Fatalf("failed to parse response JSON: %v\noutput: %s", err, output)
	}

	mode, _ := responseObj["mode"].(string)
	if mode == "inline" {
		results, ok := responseObj["data"].([]interface{})
		if !ok {
			t.Fatalf("inline mode: expected data array, got %T in: %s", responseObj["data"], output)
		}
		return results
	}

	// file_ref mode: read the temp file
	fileRefObj, ok := responseObj["file_ref"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected file_ref object or inline data, got mode=%q in: %s", mode, output)
	}
	path, _ := fileRefObj["path"].(string)
	if path == "" {
		t.Fatalf("file_ref missing path in: %s", output)
	}
	defer os.Remove(path)

	rawBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file_ref temp file %s: %v", path, err)
	}

	var results []interface{}
	for _, line := range strings.Split(strings.TrimSpace(string(rawBytes)), "\n") {
		if line == "" {
			continue
		}
		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("failed to parse JSONL line: %v", err)
		}
		results = append(results, obj)
	}
	return results
}

// TestGroupBySession_Integration tests group_by_session via ExecuteTool
func TestGroupBySession_Integration(t *testing.T) {
	cleanup := setupGroupBySessionFixture(t)
	defer cleanup()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":             "user",
		"pattern":          ".",
		"group_by_session": true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := extractGroupedResults(t, output)

	if len(results) == 0 {
		t.Fatal("expected at least one session object in results")
	}

	firstResult, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", results[0])
	}

	if _, has := firstResult["session_id"]; !has {
		t.Errorf("expected session_id key in grouped result, got: %v", firstResult)
	}
}

// TestGroupBySession_MutualExclusionWithStatsOnly tests that group_by_session and stats_only are mutually exclusive
func TestGroupBySession_MutualExclusionWithStatsOnly(t *testing.T) {
	cleanup := setupGroupBySessionFixture(t)
	defer cleanup()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	_, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":             "user",
		"pattern":          ".",
		"group_by_session": true,
		"stats_only":       true,
	})

	if err == nil {
		t.Fatal("expected error for mutually exclusive params, got nil")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected error to contain 'mutually exclusive', got: %v", err)
	}
}

// TestGroupBySession_WithContentSummary tests group_by_session combined with content_summary
func TestGroupBySession_WithContentSummary(t *testing.T) {
	cleanup := setupGroupBySessionFixture(t)
	defer cleanup()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":             "user",
		"pattern":          ".",
		"group_by_session": true,
		"content_summary":  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := extractGroupedResults(t, output)

	if len(results) == 0 {
		t.Fatal("expected session objects in results")
	}

	// Each session object should have a turns array
	sessionObj, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map for session object")
	}

	turns, ok := sessionObj["turns"].([]interface{})
	if !ok {
		t.Fatalf("expected turns array in session object, got %T", sessionObj["turns"])
	}
	if len(turns) == 0 {
		t.Fatal("expected turns to be non-empty")
	}

	// Turns should contain content_preview (content_summary applied before grouping)
	firstTurn, ok := turns[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map for turn, got %T", turns[0])
	}
	if _, has := firstTurn["content_preview"]; !has {
		t.Errorf("expected content_preview in turn after content_summary, got keys: %v", firstTurn)
	}
}

// TestGroupBySession_WithStatsFirst tests group_by_session combined with stats_first
func TestGroupBySession_WithStatsFirst(t *testing.T) {
	cleanup := setupGroupBySessionFixture(t)
	defer cleanup()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":             "user",
		"pattern":          ".",
		"group_by_session": true,
		"stats_first":      true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// stats_first output format: "<stats lines>\n---\n<detail>"
	if !strings.Contains(output, "---") {
		t.Fatalf("expected stats separator '---' in stats_first output, got: %s", output)
	}

	parts := strings.SplitN(output, "\n---\n", 2)
	if len(parts) < 2 {
		t.Fatalf("expected 2 parts separated by '---', got: %s", output)
	}

	// Stats header should contain session_count and total
	statsSection := parts[0]
	firstStatLine := strings.SplitN(strings.TrimSpace(statsSection), "\n", 2)[0]
	var stats map[string]interface{}
	if err := json.Unmarshal([]byte(firstStatLine), &stats); err != nil {
		t.Fatalf("failed to parse stats header: %v", err)
	}
	if _, has := stats["session_count"]; !has {
		t.Errorf("expected session_count in stats header, got: %v", stats)
	}
	if _, has := stats["total"]; !has {
		t.Errorf("expected total in stats header, got: %v", stats)
	}

	// Detail section should contain session_id (grouped detail)
	detailSection := parts[1]
	if !strings.Contains(detailSection, "session_id") {
		t.Errorf("expected session_id in grouped detail section, got: %s", detailSection)
	}
}

// setupSessionStatsFixture creates a fixture with 2 sessions for stats_level=session tests.
// sess-A: 3 turns 10 min apart; sess-B: 2 turns 5 min apart.
func setupSessionStatsFixture(t *testing.T) func() {
	t.Helper()
	projectDir := t.TempDir()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	fixture := "{\"type\":\"user\",\"timestamp\":\"2026-03-09T06:00:00Z\",\"uuid\":\"s1\",\"sessionId\":\"stats-sess-A\",\"message\":{\"role\":\"user\",\"content\":\"turn 1 A\"}}\n" +
		"{\"type\":\"user\",\"timestamp\":\"2026-03-09T06:05:00Z\",\"uuid\":\"s2\",\"sessionId\":\"stats-sess-A\",\"message\":{\"role\":\"user\",\"content\":\"turn 2 A\"}}\n" +
		"{\"type\":\"user\",\"timestamp\":\"2026-03-09T06:10:00Z\",\"uuid\":\"s3\",\"sessionId\":\"stats-sess-A\",\"message\":{\"role\":\"user\",\"content\":\"turn 3 A\"}}\n" +
		"{\"type\":\"user\",\"timestamp\":\"2026-03-09T07:00:00Z\",\"uuid\":\"s4\",\"sessionId\":\"stats-sess-B\",\"message\":{\"role\":\"user\",\"content\":\"turn 1 B\"}}\n" +
		"{\"type\":\"user\",\"timestamp\":\"2026-03-09T07:05:00Z\",\"uuid\":\"s5\",\"sessionId\":\"stats-sess-B\",\"message\":{\"role\":\"user\",\"content\":\"turn 2 B\"}}\n"

	writeSessionFixture(t, projectDir, "session-stats", fixture)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	return func() { _ = os.Chdir(oldWd) }
}

// TestStatsLevelSession_StatsOnly tests stats_level="session" with stats_only=true
func TestStatsLevelSession_StatsOnly(t *testing.T) {
	cleanup := setupSessionStatsFixture(t)
	defer cleanup()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":        "user",
		"pattern":     ".",
		"stats_only":  true,
		"stats_level": "session",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines (summary + sessions), got %d: %s", len(lines), output)
	}

	// First line: summary with total_sessions (not hour buckets)
	var summary map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &summary); err != nil {
		t.Fatalf("failed to parse summary line: %v", err)
	}
	if _, has := summary["total_sessions"]; !has {
		t.Errorf("expected total_sessions key in summary, got: %v", summary)
	}
	if _, has := summary["hour"]; has {
		t.Errorf("should NOT have 'hour' key in session-level stats, got: %v", summary)
	}

	// Per-session lines should have session_id, match_count, duration_minutes
	var sessLine map[string]interface{}
	if err := json.Unmarshal([]byte(lines[1]), &sessLine); err != nil {
		t.Fatalf("failed to parse session line: %v", err)
	}
	if _, has := sessLine["session_id"]; !has {
		t.Errorf("expected session_id in per-session line, got: %v", sessLine)
	}
	if _, has := sessLine["match_count"]; !has {
		t.Errorf("expected match_count in per-session line, got: %v", sessLine)
	}
	if _, has := sessLine["duration_minutes"]; !has {
		t.Errorf("expected duration_minutes in per-session line, got: %v", sessLine)
	}
}

// TestStatsLevelSession_StatsFirst tests stats_level="session" with stats_first=true
func TestStatsLevelSession_StatsFirst(t *testing.T) {
	cleanup := setupSessionStatsFixture(t)
	defer cleanup()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":        "user",
		"pattern":     ".",
		"stats_first": true,
		"stats_level": "session",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// stats_first output: "<stats lines>\n---\n<detail>"
	if !strings.Contains(output, "---") {
		t.Fatalf("expected stats separator '---' in stats_first output, got: %s", output)
	}

	parts := strings.SplitN(output, "\n---\n", 2)
	if len(parts) < 2 {
		t.Fatalf("expected 2 parts separated by '---', got: %s", output)
	}

	// Stats header should contain total_sessions (session-level aggregation)
	firstStatLine := strings.SplitN(strings.TrimSpace(parts[0]), "\n", 2)[0]
	var stats map[string]interface{}
	if err := json.Unmarshal([]byte(firstStatLine), &stats); err != nil {
		t.Fatalf("failed to parse stats header: %v", err)
	}
	if _, has := stats["total_sessions"]; !has {
		t.Errorf("expected total_sessions in stats header (session-level), got: %v", stats)
	}
	if _, has := stats["hour"]; has {
		t.Errorf("should NOT have 'hour' key in session-level stats header, got: %v", stats)
	}

	// Detail records should follow after "---"
	if parts[1] == "" {
		t.Error("expected detail records after '---' separator")
	}
}

// TestStatsLevelTurn_Regression tests that omitting stats_level still produces hour-bucket output
func TestStatsLevelTurn_Regression(t *testing.T) {
	cleanup := setupSessionStatsFixture(t)
	defer cleanup()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":       "user",
		"pattern":    ".",
		"stats_only": true,
		// no stats_level — should default to "turn" (hourly buckets)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have hour buckets (original behavior)
	if !strings.Contains(output, `"hour"`) {
		t.Errorf("expected 'hour' key in default stats_only output (turn-level), got: %s", output)
	}
	if strings.Contains(output, `"total_sessions"`) {
		t.Errorf("should NOT have 'total_sessions' when stats_level is default (turn), got: %s", output)
	}
}

// TestStatsLevelInvalid tests that an invalid stats_level value returns an error
func TestStatsLevelInvalid(t *testing.T) {
	cleanup := setupSessionStatsFixture(t)
	defer cleanup()

	executor := NewToolExecutor()
	cfg := &config.Config{}

	_, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":        "user",
		"pattern":     ".",
		"stats_level": "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid stats_level, got nil")
	}
	if !strings.Contains(err.Error(), "must be 'turn' or 'session'") {
		t.Errorf("expected error to contain \"must be 'turn' or 'session'\", got: %v", err)
	}
}

// setupContextTurnsFixture creates a fixture for context_turns tests.
// Returns (projectDir, executor, cleanup). The session file is placed in
// a META_CC_PROJECTS_ROOT hash directory so ExecuteTool can find it.
// Fixture: 5 turns for "ctx-sess-A", all matching pattern ".".
func setupContextTurnsFixture(t *testing.T) (string, *ToolExecutor, func()) {
	t.Helper()
	projectDir := t.TempDir()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	fixture := "" +
		`{"type":"user","uuid":"ct0","sessionId":"ctx-sess-A","timestamp":"2026-03-09T06:00:00Z","message":{"role":"user","content":"turn 0"}}` + "\n" +
		`{"type":"user","uuid":"ct1","sessionId":"ctx-sess-A","timestamp":"2026-03-09T06:01:00Z","message":{"role":"user","content":"turn 1"}}` + "\n" +
		`{"type":"user","uuid":"ct2","sessionId":"ctx-sess-A","timestamp":"2026-03-09T06:02:00Z","message":{"role":"user","content":"turn 2"}}` + "\n" +
		`{"type":"user","uuid":"ct3","sessionId":"ctx-sess-A","timestamp":"2026-03-09T06:03:00Z","message":{"role":"user","content":"turn 3"}}` + "\n" +
		`{"type":"user","uuid":"ct4","sessionId":"ctx-sess-A","timestamp":"2026-03-09T06:04:00Z","message":{"role":"user","content":"turn 4"}}` + "\n"

	writeSessionFixture(t, projectDir, "ctx-session", fixture)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	return projectDir, NewToolExecutor(), func() { _ = os.Chdir(oldWd) }
}

// setupContextTurnsOverlapFixture creates a fixture with 10 turns for overlap tests.
func setupContextTurnsOverlapFixture(t *testing.T) (string, *ToolExecutor, func()) {
	t.Helper()
	projectDir := t.TempDir()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	var lines []string
	for i := 0; i < 10; i++ {
		ts := fmt.Sprintf("2026-03-09T06:%02d:00Z", i)
		// turns 2 and 4 contain "match-me"; others contain "turn N"
		var content string
		if i == 2 || i == 4 {
			content = fmt.Sprintf("match-me turn %d", i)
		} else {
			content = fmt.Sprintf("turn %d", i)
		}
		lines = append(lines, fmt.Sprintf(
			`{"type":"user","uuid":"co%d","sessionId":"overlap-sess","timestamp":"%s","message":{"role":"user","content":"%s"}}`,
			i, ts, content,
		))
	}

	fixture := strings.Join(lines, "\n") + "\n"
	writeSessionFixture(t, projectDir, "overlap-session", fixture)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	return projectDir, NewToolExecutor(), func() { _ = os.Chdir(oldWd) }
}

// extractContextTurnsResults reads inline or file_ref output and returns []map[string]interface{}.
func extractContextTurnsResults(t *testing.T, output string) []map[string]interface{} {
	t.Helper()
	var responseObj map[string]interface{}
	if err := json.Unmarshal([]byte(output), &responseObj); err != nil {
		t.Fatalf("failed to parse response JSON: %v\noutput: %s", err, output)
	}

	mode, _ := responseObj["mode"].(string)
	var items []interface{}
	if mode == "inline" {
		var ok bool
		items, ok = responseObj["data"].([]interface{})
		if !ok {
			t.Fatalf("inline mode: expected data array, got %T", responseObj["data"])
		}
	} else {
		fileRefObj, ok := responseObj["file_ref"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected file_ref or inline, got mode=%q output: %s", mode, output)
		}
		path, _ := fileRefObj["path"].(string)
		rawBytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file_ref %s: %v", path, err)
		}
		defer os.Remove(path)
		for _, line := range strings.Split(strings.TrimSpace(string(rawBytes)), "\n") {
			if line == "" {
				continue
			}
			var obj interface{}
			if err := json.Unmarshal([]byte(line), &obj); err != nil {
				t.Fatalf("failed to parse JSONL line: %v", err)
			}
			items = append(items, obj)
		}
	}

	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map item, got %T", item)
		}
		result = append(result, m)
	}
	return result
}

// TestContextTurns_Basic: 5-turn session, match turn at index 2, context_turns=1 →
// indices 1,2,3 returned; index 2 has "context":false, others "context":true.
func TestContextTurns_Basic(t *testing.T) {
	_, executor, cleanup := setupContextTurnsFixture(t)
	defer cleanup()

	cfg := &config.Config{}
	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":          "user",
		"pattern":       "turn 2",
		"context_turns": float64(1),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := extractContextTurnsResults(t, output)
	if len(results) != 3 {
		t.Fatalf("expected 3 turns (index 1,2,3), got %d: %v", len(results), results)
	}

	// Find the matched turn (uuid=ct2) and verify context field
	for _, r := range results {
		uuid, _ := r["uuid"].(string)
		ctxVal, hasCtx := r["context"]
		if uuid == "ct2" {
			if hasCtx && ctxVal.(bool) == true {
				t.Errorf("matched turn (ct2) should have context:false, got context:true")
			}
		} else {
			if !hasCtx || ctxVal.(bool) != true {
				t.Errorf("context turn (%s) should have context:true, got %v", uuid, ctxVal)
			}
		}
	}
}

// TestContextTurns_BoundaryStart: match turn 0, context_turns=2 → turns 0,1,2 (no negative index).
func TestContextTurns_BoundaryStart(t *testing.T) {
	_, executor, cleanup := setupContextTurnsFixture(t)
	defer cleanup()

	cfg := &config.Config{}
	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":          "user",
		"pattern":       "turn 0",
		"context_turns": float64(2),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := extractContextTurnsResults(t, output)
	if len(results) != 3 {
		t.Fatalf("expected 3 turns (index 0,1,2), got %d: %v", len(results), results)
	}

	// Verify first turn (uuid=ct0) is the matched one
	first := results[0]
	if uuid, _ := first["uuid"].(string); uuid != "ct0" {
		t.Errorf("expected first turn uuid=ct0, got %s", uuid)
	}
}

// TestContextTurns_BoundaryEnd: match last turn (index 4), context_turns=2 → turns 2,3,4.
func TestContextTurns_BoundaryEnd(t *testing.T) {
	_, executor, cleanup := setupContextTurnsFixture(t)
	defer cleanup()

	cfg := &config.Config{}
	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":          "user",
		"pattern":       "turn 4",
		"context_turns": float64(2),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := extractContextTurnsResults(t, output)
	if len(results) != 3 {
		t.Fatalf("expected 3 turns (index 2,3,4), got %d: %v", len(results), results)
	}

	// Verify last turn (uuid=ct4) is the matched one
	last := results[len(results)-1]
	if uuid, _ := last["uuid"].(string); uuid != "ct4" {
		t.Errorf("expected last turn uuid=ct4, got %s", uuid)
	}
}

// TestContextTurns_OverlappingWindows: 10-turn session, matches at indices 2 and 4,
// context_turns=2 → indices 0–6 returned (windows merged), no duplicates.
// Turns 2 and 4 have context:false; others have context:true.
func TestContextTurns_OverlappingWindows(t *testing.T) {
	_, executor, cleanup := setupContextTurnsOverlapFixture(t)
	defer cleanup()

	cfg := &config.Config{}
	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":          "user",
		"pattern":       "match-me",
		"context_turns": float64(2),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := extractContextTurnsResults(t, output)
	// Window for turn 2: [0,1,2,3,4]; window for turn 4: [2,3,4,5,6]
	// Union: [0,1,2,3,4,5,6] = 7 items
	if len(results) != 7 {
		t.Fatalf("expected 7 turns (indices 0-6), got %d: %v", len(results), results)
	}

	// Verify no duplicate uuids
	seen := make(map[string]bool)
	for _, r := range results {
		uuid, _ := r["uuid"].(string)
		if seen[uuid] {
			t.Errorf("duplicate uuid found: %s", uuid)
		}
		seen[uuid] = true
	}

	// Turns co2 and co4 should have context:false; others context:true
	for _, r := range results {
		uuid, _ := r["uuid"].(string)
		ctxVal, hasCtx := r["context"]
		if uuid == "co2" || uuid == "co4" {
			if hasCtx && ctxVal.(bool) == true {
				t.Errorf("matched turn %s should have context:false, got context:true", uuid)
			}
		} else {
			if !hasCtx || ctxVal.(bool) != true {
				t.Errorf("context turn %s should have context:true, got %v", uuid, ctxVal)
			}
		}
	}
}

// TestContextTurns_ArrayContentType: content_type=array with context_turns → no error,
// no "context" field in output (silently ignored).
func TestContextTurns_ArrayContentType(t *testing.T) {
	projectDir := t.TempDir()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	fixture := `{"type":"user","uuid":"ar1","sessionId":"arr-sess","timestamp":"2026-03-09T06:00:00Z","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":"ok"}]}}` + "\n" +
		`{"type":"user","uuid":"ar2","sessionId":"arr-sess","timestamp":"2026-03-09T06:01:00Z","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t2","content":"done"}]}}` + "\n"
	writeSessionFixture(t, projectDir, "arr-session", fixture)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	executor := NewToolExecutor()
	cfg := &config.Config{}
	output, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":          "user",
		"content_type":  "array",
		"context_turns": float64(2),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return results without "context" field
	if strings.Contains(output, `"context"`) {
		t.Errorf("array content_type: should not add context field, but got: %s", output)
	}
}

// TestContextTurns_Zero: context_turns=0 → output identical to not specifying context_turns.
func TestContextTurns_Zero(t *testing.T) {
	_, executor, cleanup := setupContextTurnsFixture(t)
	defer cleanup()

	cfg := &config.Config{}

	// Without context_turns
	output1, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":    "user",
		"pattern": "turn 2",
	})
	if err != nil {
		t.Fatalf("without context_turns: unexpected error: %v", err)
	}

	// With context_turns=0
	output2, err := executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":          "user",
		"pattern":       "turn 2",
		"context_turns": float64(0),
	})
	if err != nil {
		t.Fatalf("context_turns=0: unexpected error: %v", err)
	}

	// Both should have the same number of results (just 1 match, no context)
	results1 := extractContextTurnsResults(t, output1)
	results2 := extractContextTurnsResults(t, output2)
	if len(results1) != len(results2) {
		t.Errorf("context_turns=0 should produce same count as no context_turns: got %d vs %d", len(results2), len(results1))
	}
}
