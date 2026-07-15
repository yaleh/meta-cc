package query

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	stageTestUser1 = `{"type":"user","timestamp":"2025-01-15T10:00:00Z","message":{"content":"fix bug"}}`
)

func TestHandleExecuteStage2Query_TransformAllNull_WarningsInResponse(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_stage_warn.jsonl")
	if err := os.WriteFile(testFile, []byte(stageTestUser1+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	args := map[string]interface{}{
		"files":     []interface{}{testFile},
		"filter":    `select(.type == "user")`,
		"transform": ".nonexistent",
	}

	result, err := HandleExecuteStage2Query(context.Background(), args)
	if err != nil {
		t.Fatalf("HandleExecuteStage2Query failed: %v", err)
	}

	// Marshal to JSON and unmarshal to check warnings
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	warningsRaw, ok := parsed["warnings"]
	if !ok {
		t.Fatal("expected 'warnings' key in response")
	}

	warnings, ok := warningsRaw.([]interface{})
	if !ok {
		t.Fatalf("expected warnings to be an array, got %T", warningsRaw)
	}

	if len(warnings) == 0 {
		t.Error("expected non-empty warnings array when transform produces all-null results")
	}
}

func TestHandleExecuteStage2Query_TransformValidField_EmptyWarnings(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_stage_nowarn.jsonl")
	if err := os.WriteFile(testFile, []byte(stageTestUser1+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	args := map[string]interface{}{
		"files":     []interface{}{testFile},
		"filter":    `select(.type == "user")`,
		"transform": "{type: .type}",
	}

	result, err := HandleExecuteStage2Query(context.Background(), args)
	if err != nil {
		t.Fatalf("HandleExecuteStage2Query failed: %v", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	warningsRaw, ok := parsed["warnings"]
	if !ok {
		t.Fatal("expected 'warnings' key in response")
	}

	warnings, ok := warningsRaw.([]interface{})
	if !ok {
		t.Fatalf("expected warnings to be an array, got %T", warningsRaw)
	}

	if len(warnings) != 0 {
		t.Errorf("expected empty warnings for valid transform, got: %v", warnings)
	}
}

// TestHandleGetSessionMetadata_SessionScope_SingleFile verifies that HandleGetSessionMetadata
// with scope=session returns metadata for exactly the single current session file.
func TestHandleGetSessionMetadata_SessionScope_SingleFile(t *testing.T) {
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore cwd: %v", err)
		}
	}()

	// Set up a fake project session directory
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	projectPath := t.TempDir()
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	resolvedPath, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		resolvedPath = projectPath
	}
	// Compute project hash
	h := strings.ReplaceAll(resolvedPath, "\\", "-")
	h = strings.ReplaceAll(h, "/", "-")
	h = strings.ReplaceAll(h, ":", "-")

	sessionDir := filepath.Join(projectsRoot, h)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// Create two session files so project scope would return 2
	f1 := filepath.Join(sessionDir, "sess1.jsonl")
	f2 := filepath.Join(sessionDir, "sess2.jsonl")
	if err := os.WriteFile(f1, []byte(`{"type":"user"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to write sess1: %v", err)
	}
	if err := os.WriteFile(f2, []byte(`{"type":"user"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to write sess2: %v", err)
	}

	args := map[string]interface{}{
		"scope": "session",
	}

	result, err := HandleGetSessionMetadata(context.Background(), args)
	if err != nil {
		t.Fatalf("HandleGetSessionMetadata failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	fileCount, ok := resultMap["file_count"].(int)
	if !ok {
		t.Fatalf("expected file_count to be int, got %T: %v", resultMap["file_count"], resultMap["file_count"])
	}

	if fileCount != 1 {
		t.Errorf("expected session scope to return file_count=1, got %d", fileCount)
	}
}
