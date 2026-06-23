package query

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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
