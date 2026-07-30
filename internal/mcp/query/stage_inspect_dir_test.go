package query

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC3: passing a directory (which get_session_directory naturally returns) to
// inspect_session_files must yield a structured correction naming the exact
// supported files workflow, instead of a low-level "is a directory" error.
func TestHandleInspectSessionFiles_DirectoryInput_StructuredCorrection(t *testing.T) {
	dir := t.TempDir()
	// A session file inside the directory — the caller meant to pass this.
	if err := os.WriteFile(filepath.Join(dir, "s1.jsonl"), []byte(`{"type":"user","timestamp":"2025-01-15T10:00:00Z"}`+"\n"), 0o644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	_, err := HandleInspectSessionFiles(context.Background(), map[string]interface{}{
		"files": []interface{}{dir},
	})
	if err == nil {
		t.Fatal("expected a structured correction for a directory input, got nil error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "directory") {
		t.Errorf("correction should identify the directory input:\n%s", msg)
	}
	if !strings.Contains(msg, "get_session_directory") {
		t.Errorf("correction should point to the get_session_directory workflow:\n%s", msg)
	}
	if !strings.Contains(msg, "files") {
		t.Errorf("correction should reference the exact .files value to pass:\n%s", msg)
	}
}

// AC3 (negative): a real file path still inspects successfully — the directory
// guard must not change valid behavior.
func TestHandleInspectSessionFiles_FileInput_StillWorks(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "s1.jsonl")
	if err := os.WriteFile(file, []byte(`{"type":"user","timestamp":"2025-01-15T10:00:00Z"}`+"\n"), 0o644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	result, err := HandleInspectSessionFiles(context.Background(), map[string]interface{}{
		"files": []interface{}{file},
	})
	if err != nil {
		t.Fatalf("valid file input should inspect successfully, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected a non-nil inspection result")
	}
}

// AC3: a mix of a valid file and a directory still surfaces the structured
// correction (the directory entry is the actionable problem).
func TestHandleInspectSessionFiles_MixedFileAndDirectory_Correction(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "s1.jsonl")
	if err := os.WriteFile(file, []byte(`{"type":"user"}`+"\n"), 0o644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}
	subDir := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to make subdir: %v", err)
	}

	_, err := HandleInspectSessionFiles(context.Background(), map[string]interface{}{
		"files": []interface{}{file, subDir},
	})
	if err == nil {
		t.Fatal("expected a structured correction when any entry is a directory")
	}
	if !strings.Contains(err.Error(), "get_session_directory") {
		t.Errorf("correction should point to the get_session_directory workflow:\n%s", err.Error())
	}
}
