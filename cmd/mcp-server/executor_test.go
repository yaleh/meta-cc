package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/config"
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
