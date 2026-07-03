package query

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

const (
	queryWarnTestUser1 = `{"type":"user","timestamp":"2025-01-15T10:00:00Z","message":{"content":"fix bug"}}`
)

func TestRunQuery_TransformAllNull_Warning(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_runquery_warn.jsonl")
	if err := os.WriteFile(testFile, []byte(queryWarnTestUser1+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	executor := NewQueryExecutor(tempDir)
	result, err := executor.RunQuery(
		context.Background(),
		[]string{testFile},
		`select(.type == "user")`,
		".nonexistent",
		0,
	)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	if len(result.Warnings) == 0 {
		t.Error("expected non-empty Warnings when transform produces all-null results")
	}
	found := false
	for _, w := range result.Warnings {
		if containsStr(w, "inspect_session_files") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning to mention inspect_session_files, got: %v", result.Warnings)
	}
}

func TestRunQuery_TransformValidField_NoWarning(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_runquery_nowarn.jsonl")
	if err := os.WriteFile(testFile, []byte(queryWarnTestUser1+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	executor := NewQueryExecutor(tempDir)
	result, err := executor.RunQuery(
		context.Background(),
		[]string{testFile},
		`select(.type == "user")`,
		"{type: .type}",
		0,
	)
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	// Filter out any pre-existing warnings (e.g., file skip warnings) and check for transform warning
	for _, w := range result.Warnings {
		if containsStr(w, "inspect_session_files") {
			t.Errorf("unexpected transform warning for valid transform: %v", w)
		}
	}
}

func TestRunQueryWithTimeRange_TransformAllNull_Warning(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_rqwtr_warn.jsonl")
	if err := os.WriteFile(testFile, []byte(queryWarnTestUser1+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	executor := NewQueryExecutor(tempDir)
	result, err := executor.RunQueryWithTimeRange(
		context.Background(),
		[]string{testFile},
		`select(.type == "user")`,
		".nonexistent",
		0,
		ParsedTimeRange{},
	)
	if err != nil {
		t.Fatalf("RunQueryWithTimeRange failed: %v", err)
	}

	if len(result.Warnings) == 0 {
		t.Error("expected non-empty Warnings when transform produces all-null results")
	}
	found := false
	for _, w := range result.Warnings {
		if containsStr(w, "inspect_session_files") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning to mention inspect_session_files, got: %v", result.Warnings)
	}
}

func containsStr(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
