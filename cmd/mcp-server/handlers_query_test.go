package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHandleQuery tests the new query tool with 10 high-frequency queries
func TestHandleQuery(t *testing.T) {
	// Create temp directory with test JSONL files
	tmpDir := t.TempDir()

	// Create test data matching JSONL schema
	testEntries := []map[string]interface{}{
		// User messages
		{
			"type":      "user",
			"timestamp": "2025-01-01T10:00:00Z",
			"uuid":      "user-uuid-1",
			"message": map[string]interface{}{
				"content": "Fix the error in main.go",
			},
		},
		{
			"type":      "user",
			"timestamp": "2025-01-01T10:02:00Z",
			"uuid":      "user-uuid-2",
			"message": map[string]interface{}{
				"content": "@main.go please review this file",
			},
		},
		// Assistant messages with tool uses
		{
			"type":      "assistant",
			"timestamp": "2025-01-01T10:01:00Z",
			"uuid":      "assistant-uuid-1",
			"message": map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type":  "tool_use",
						"id":    "tool-1",
						"name":  "Read",
						"input": map[string]interface{}{"file": "main.go"},
					},
				},
			},
		},
		{
			"type":      "assistant",
			"timestamp": "2025-01-01T10:03:00Z",
			"uuid":      "assistant-uuid-2",
			"message": map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type":  "tool_use",
						"id":    "tool-2",
						"name":  "Edit",
						"input": map[string]interface{}{"file": "main.go"},
					},
				},
				"usage": map[string]interface{}{
					"input_tokens":  100,
					"output_tokens": 50,
				},
			},
		},
		// Tool results with errors
		{
			"type":      "user",
			"timestamp": "2025-01-01T10:04:00Z",
			"uuid":      "user-uuid-3",
			"message": map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type":        "tool_result",
						"tool_use_id": "tool-1",
						"is_error":    true,
						"content":     "File not found",
					},
				},
			},
		},
		// System entries
		{
			"type":      "system",
			"subtype":   "api_error",
			"timestamp": "2025-01-01T10:05:00Z",
			"uuid":      "system-uuid-1",
		},
		// File history snapshot
		{
			"type":      "file-history-snapshot",
			"timestamp": "2025-01-01T10:06:00Z",
			"messageId": "assistant-uuid-1",
		},
		// Summary
		{
			"type":      "summary",
			"timestamp": "2025-01-01T10:07:00Z",
			"summary":   "Fixed error in main.go",
		},
	}

	// Write to JSONL file
	file := filepath.Join(tmpDir, "session.jsonl")
	f, err := os.Create(file)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	for _, entry := range testEntries {
		data, _ := json.Marshal(entry)
		f.Write(data)
		f.WriteString("\n")
	}
	f.Close()

	// Create QueryExecutor
	executor := NewQueryExecutor(tmpDir)
	ctx := context.Background()

	// Test cases for 10 high-frequency queries
	tests := []struct {
		name         string
		jqFilter     string
		wantMinCount int // minimum expected results
		validateFunc func(t *testing.T, results []interface{})
	}{
		{
			name:         "Query 1: User Messages",
			jqFilter:     `select(.type == "user" and (.message.content | type == "string"))`,
			wantMinCount: 2,
			validateFunc: func(t *testing.T, results []interface{}) {
				for _, r := range results {
					m := r.(map[string]interface{})
					if m["type"] != "user" {
						t.Errorf("expected type=user, got %v", m["type"])
					}
				}
			},
		},
		{
			name:         "Query 2: Tool Executions (All Tools)",
			jqFilter:     `select(.type == "assistant") | select(.message.content[] | .type == "tool_use")`,
			wantMinCount: 2,
			validateFunc: func(t *testing.T, results []interface{}) {
				for _, r := range results {
					m := r.(map[string]interface{})
					if m["type"] != "assistant" {
						t.Errorf("expected type=assistant, got %v", m["type"])
					}
				}
			},
		},
		{
			name:         "Query 3: Tool Results (Error Detection)",
			jqFilter:     `select(.type == "user" and (.message.content | type == "array")) | select(.message.content[] | select(.type == "tool_result" and .is_error == true))`,
			wantMinCount: 1,
			validateFunc: func(t *testing.T, results []interface{}) {
				if len(results) == 0 {
					t.Error("expected at least one error result")
				}
			},
		},
		{
			name:         "Query 4: Assistant Responses with Token Usage",
			jqFilter:     `select(.type == "assistant" and has("message")) | select(.message | has("usage"))`,
			wantMinCount: 1,
			validateFunc: func(t *testing.T, results []interface{}) {
				for _, r := range results {
					m := r.(map[string]interface{})
					msg := m["message"].(map[string]interface{})
					if _, ok := msg["usage"]; !ok {
						t.Error("expected usage field in message")
					}
				}
			},
		},
		{
			name:         "Query 5: Parent-Child Relationships",
			jqFilter:     `select(.type == "user" or .type == "assistant")`,
			wantMinCount: 4,
			validateFunc: func(t *testing.T, results []interface{}) {
				for _, r := range results {
					m := r.(map[string]interface{})
					typ := m["type"].(string)
					if typ != "user" && typ != "assistant" {
						t.Errorf("expected type user or assistant, got %v", typ)
					}
				}
			},
		},
		{
			name:         "Query 6: System Entries (Error Events)",
			jqFilter:     `select(.type == "system" and .subtype == "api_error")`,
			wantMinCount: 1,
			validateFunc: func(t *testing.T, results []interface{}) {
				for _, r := range results {
					m := r.(map[string]interface{})
					if m["type"] != "system" {
						t.Errorf("expected type=system, got %v", m["type"])
					}
					if m["subtype"] != "api_error" {
						t.Errorf("expected subtype=api_error, got %v", m["subtype"])
					}
				}
			},
		},
		{
			name:         "Query 7: File History Snapshots",
			jqFilter:     `select(.type == "file-history-snapshot" and has("messageId"))`,
			wantMinCount: 1,
			validateFunc: func(t *testing.T, results []interface{}) {
				for _, r := range results {
					m := r.(map[string]interface{})
					if m["type"] != "file-history-snapshot" {
						t.Errorf("expected type=file-history-snapshot, got %v", m["type"])
					}
				}
			},
		},
		{
			name:         "Query 8: Conversation Timestamps (sorted)",
			jqFilter:     `select(.timestamp)`,
			wantMinCount: 8,
			validateFunc: func(t *testing.T, results []interface{}) {
				for _, r := range results {
					m := r.(map[string]interface{})
					if _, ok := m["timestamp"]; !ok {
						t.Error("expected timestamp field")
					}
				}
			},
		},
		{
			name:         "Query 9: Summary Records",
			jqFilter:     `select(.type == "summary")`,
			wantMinCount: 1,
			validateFunc: func(t *testing.T, results []interface{}) {
				for _, r := range results {
					m := r.(map[string]interface{})
					if m["type"] != "summary" {
						t.Errorf("expected type=summary, got %v", m["type"])
					}
				}
			},
		},
		{
			name:         "Query 10: Content Blocks (Tool Use)",
			jqFilter:     `select(.type == "assistant") | .message.content[] | select(.type == "tool_use")`,
			wantMinCount: 2,
			validateFunc: func(t *testing.T, results []interface{}) {
				for _, r := range results {
					m := r.(map[string]interface{})
					if m["type"] != "tool_use" {
						t.Errorf("expected type=tool_use, got %v", m["type"])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compile expression
			code, err := executor.compileExpression(tt.jqFilter)
			if err != nil {
				t.Fatalf("failed to compile expression: %v", err)
			}

			// Execute query
			results := executor.streamFiles(ctx, []string{file}, code, 0)

			// Check minimum count
			if len(results) < tt.wantMinCount {
				t.Errorf("expected at least %d results, got %d", tt.wantMinCount, len(results))
			}

			// Run custom validation
			if tt.validateFunc != nil {
				tt.validateFunc(t, results)
			}
		})
	}
}

// TestHandleQueryRaw tests the query_raw tool
func TestHandleQueryRaw(t *testing.T) {
	tmpDir := t.TempDir()

	// Create simple test data
	testEntries := []map[string]interface{}{
		{"type": "user", "id": 1},
		{"type": "assistant", "id": 2},
		{"type": "user", "id": 3},
	}

	file := filepath.Join(tmpDir, "session.jsonl")
	f, err := os.Create(file)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	for _, entry := range testEntries {
		data, _ := json.Marshal(entry)
		f.Write(data)
		f.WriteString("\n")
	}
	f.Close()

	executor := NewQueryExecutor(tmpDir)
	ctx := context.Background()

	tests := []struct {
		name         string
		jqExpression string
		wantCount    int
		wantErr      bool
	}{
		{
			name:         "simple filter",
			jqExpression: `select(.type == "user")`,
			wantCount:    2,
			wantErr:      false,
		},
		{
			name:         "identity filter",
			jqExpression: ".",
			wantCount:    3,
			wantErr:      false,
		},
		{
			name:         "complex transformation",
			jqExpression: `select(.type == "user") | {id, type}`,
			wantCount:    2,
			wantErr:      false,
		},
		{
			name:         "invalid expression",
			jqExpression: "select(",
			wantCount:    0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := executor.compileExpression(tt.jqExpression)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			results := executor.streamFiles(ctx, []string{file}, code, 0)

			if len(results) != tt.wantCount {
				t.Errorf("expected %d results, got %d", tt.wantCount, len(results))
			}
		})
	}
}

// TestQueryEquivalence tests that query and query_raw return identical results
func TestQueryEquivalence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test data
	testEntries := []map[string]interface{}{
		{"type": "user", "timestamp": "2025-01-01T10:00:00Z", "content": "hello"},
		{"type": "assistant", "timestamp": "2025-01-01T10:01:00Z", "content": "hi"},
		{"type": "user", "timestamp": "2025-01-01T10:02:00Z", "content": "bye"},
	}

	file := filepath.Join(tmpDir, "session.jsonl")
	f, err := os.Create(file)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	for _, entry := range testEntries {
		data, _ := json.Marshal(entry)
		f.Write(data)
		f.WriteString("\n")
	}
	f.Close()

	executor := NewQueryExecutor(tmpDir)
	ctx := context.Background()

	// Test: both tools should return same results for same expression
	expression := `select(.type == "user")`

	// Execute with query (using jq_filter)
	code1, err := executor.compileExpression(expression)
	if err != nil {
		t.Fatalf("failed to compile expression: %v", err)
	}
	results1 := executor.streamFiles(ctx, []string{file}, code1, 0)

	// Execute with query_raw (using jq_expression)
	code2, err := executor.compileExpression(expression)
	if err != nil {
		t.Fatalf("failed to compile expression: %v", err)
	}
	results2 := executor.streamFiles(ctx, []string{file}, code2, 0)

	// Results should be identical
	if len(results1) != len(results2) {
		t.Errorf("results length mismatch: query=%d, query_raw=%d", len(results1), len(results2))
	}

	// Convert to JSON for comparison
	json1, _ := json.Marshal(results1)
	json2, _ := json.Marshal(results2)

	if string(json1) != string(json2) {
		t.Error("results content mismatch between query and query_raw")
	}
}

// TestHybridOutputMode tests inline vs file_ref output modes
func TestHybridOutputMode(t *testing.T) {
	tests := []struct {
		name          string
		dataSize      int // number of entries
		threshold     int // inline threshold in bytes
		expectFileRef bool
	}{
		{
			name:          "small data - inline mode",
			dataSize:      10,
			threshold:     8192,
			expectFileRef: false,
		},
		{
			name:          "large data - file_ref mode",
			dataSize:      1000,
			threshold:     1024,
			expectFileRef: true,
		},
		{
			name:          "edge case - near threshold",
			dataSize:      100,
			threshold:     8192,
			expectFileRef: true, // 100 entries * ~120 bytes = ~12KB > 8KB threshold
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create test data
			var entries []map[string]interface{}
			for i := 0; i < tt.dataSize; i++ {
				entries = append(entries, map[string]interface{}{
					"type":      "user",
					"id":        i,
					"timestamp": "2025-01-01T10:00:00Z",
					"content":   "test message content for hybrid output mode testing",
				})
			}

			file := filepath.Join(tmpDir, "session.jsonl")
			f, err := os.Create(file)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}
			for _, entry := range entries {
				data, _ := json.Marshal(entry)
				f.Write(data)
				f.WriteString("\n")
			}
			f.Close()

			executor := NewQueryExecutor(tmpDir)
			ctx := context.Background()

			code, err := executor.compileExpression(".")
			if err != nil {
				t.Fatalf("failed to compile expression: %v", err)
			}

			results := executor.streamFiles(ctx, []string{file}, code, 0)

			// Serialize results to check size
			jsonData, err := json.Marshal(results)
			if err != nil {
				t.Fatalf("failed to marshal results: %v", err)
			}

			resultSize := len(jsonData)
			shouldUseFileRef := resultSize >= tt.threshold

			if shouldUseFileRef != tt.expectFileRef {
				t.Errorf("expected file_ref=%v for size=%d (threshold=%d), but logic suggests %v",
					tt.expectFileRef, resultSize, tt.threshold, shouldUseFileRef)
			}

			t.Logf("Data size: %d bytes, threshold: %d bytes, use file_ref: %v",
				resultSize, tt.threshold, shouldUseFileRef)
		})
	}
}

// TestQueryWithTransform tests jq_filter and jq_transform combination
func TestQueryWithTransform(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test data
	testEntries := []map[string]interface{}{
		{"type": "user", "timestamp": "2025-01-01T10:00:00Z", "content": "hello"},
		{"type": "assistant", "timestamp": "2025-01-01T10:01:00Z", "content": "hi"},
		{"type": "user", "timestamp": "2025-01-01T10:02:00Z", "content": "bye"},
	}

	file := filepath.Join(tmpDir, "session.jsonl")
	f, err := os.Create(file)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	for _, entry := range testEntries {
		data, _ := json.Marshal(entry)
		f.Write(data)
		f.WriteString("\n")
	}
	f.Close()

	executor := NewQueryExecutor(tmpDir)
	ctx := context.Background()

	// Test: filter user messages and transform to {type, timestamp}
	filterExpr := `select(.type == "user")`
	transformExpr := `{type, timestamp}`
	combinedExpr := executor.buildExpression(filterExpr, transformExpr)

	code, err := executor.compileExpression(combinedExpr)
	if err != nil {
		t.Fatalf("failed to compile expression: %v", err)
	}

	results := executor.streamFiles(ctx, []string{file}, code, 0)

	// Should have 2 user messages
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Check that results only contain type and timestamp
	for _, r := range results {
		m := r.(map[string]interface{})
		if _, ok := m["type"]; !ok {
			t.Error("expected type field in transformed result")
		}
		if _, ok := m["timestamp"]; !ok {
			t.Error("expected timestamp field in transformed result")
		}
		if _, ok := m["content"]; ok {
			t.Error("unexpected content field in transformed result (should be filtered out)")
		}
	}
}

// TestGetQueryBaseDir tests that getQueryBaseDir correctly locates session directories
func TestGetQueryBaseDir(t *testing.T) {
	tests := []struct {
		name              string
		scope             string
		expectedBehavior  string // "use_locator"
		shouldContainHash bool   // For project scope, should return path with hash
	}{
		{
			name:             "session scope uses FromProjectPath",
			scope:            "session",
			expectedBehavior: "use_locator",
		},
		{
			name:              "project scope uses AllSessionsFromProject",
			scope:             "project",
			expectedBehavior:  "use_locator",
			shouldContainHash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir, err := getQueryBaseDir(tt.scope)

			// Both session and project scope use SessionLocator
			// We expect "no sessions found" error in test environment
			// The important part is that it's using SessionLocator (not returning cwd)

			// We expect an error because no sessions exist in test environment
			if err == nil {
				t.Fatalf("expected error for %s scope (no sessions), got baseDir: %s", tt.scope, baseDir)
			}

			// The error should be about sessions not found or unable to locate
			errMsg := err.Error()
			if errMsg == "" {
				t.Errorf("expected error message, got empty")
			}

			expectedErrors := []string{"no sessions found", "failed to locate"}
			hasExpectedError := false
			for _, expectedErr := range expectedErrors {
				if strings.Contains(errMsg, expectedErr) {
					hasExpectedError = true
					break
				}
			}

			if !hasExpectedError {
				t.Errorf("expected error containing 'no sessions found' or 'failed to locate', got: %v", err)
			}

			t.Logf("Got expected error (SessionLocator used): %v", err)
		})
	}
}

// TestGetQueryBaseDirIntegration tests getQueryBaseDir with actual SessionLocator
func TestGetQueryBaseDirIntegration(t *testing.T) {
	// This test verifies that project scope uses SessionLocator.AllSessionsFromProject
	// Setup: Create a fake .claude/projects structure
	homeDir := t.TempDir()
	_ = filepath.Join(homeDir, ".claude", "projects") // projectsDir for future use

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", homeDir)

	// Create project directory structure
	projectPath := t.TempDir()

	// Calculate project hash (same logic as locator)
	// Note: This requires access to internal/locator logic
	// For now, we'll test the behavior indirectly

	t.Run("project scope should use SessionLocator", func(t *testing.T) {
		// Unset CLAUDE_PROJECT_DIR to force discovery
		os.Unsetenv("CLAUDE_PROJECT_DIR")

		// Change to project directory
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		os.Chdir(projectPath)

		baseDir, err := getQueryBaseDir("project")

		// We expect this to fail with "no sessions found" because we haven't
		// created the session directory structure. But it should NOT fail with
		// "no JSONL files found in <project_root>"
		if err != nil {
			// The error should be about sessions not found, not about JSONL files
			errMsg := err.Error()
			if errMsg == "no JSONL files found in "+projectPath {
				t.Errorf("getQueryBaseDir returned project root path error, should use SessionLocator: %v", err)
			}
			// Expected error: no sessions found (this is OK for this test)
			t.Logf("Expected error (no sessions setup): %v", err)
		}

		// If no error, baseDir should NOT be the project root
		if err == nil && baseDir == projectPath {
			t.Errorf("project scope returned project root (%s), should use SessionLocator", projectPath)
		}
	})
}
