package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Phase 27 Stage 27.1: Tests for query and query_raw removed
// These tools were deleted to simplify the query interface
// Use the 10 shortcut query tools instead

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

// TestQueryWithTransform removed in Phase 27 Stage 27.1
// query tool with jq_transform feature was deleted

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
