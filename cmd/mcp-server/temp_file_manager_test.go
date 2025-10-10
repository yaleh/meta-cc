package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCreateTempFilePath(t *testing.T) {
	sessionHash := "abc12345"
	queryType := "query_tools"

	path := createTempFilePath(sessionHash, queryType)

	// Check pattern: {TempDir}/meta-cc-mcp-{session_hash}-{timestamp}-{query_type}.jsonl
	expectedPrefix := filepath.Join(os.TempDir(), "meta-cc-mcp-")
	if !strings.HasPrefix(path, expectedPrefix) {
		t.Errorf("expected path to start with %s, got %s", expectedPrefix, path)
	}
	if !strings.HasSuffix(path, ".jsonl") {
		t.Errorf("expected path to end with .jsonl, got %s", path)
	}
	if !strings.Contains(path, sessionHash) {
		t.Errorf("expected path to contain session hash %s, got %s", sessionHash, path)
	}
	if !strings.Contains(path, queryType) {
		t.Errorf("expected path to contain query type %s, got %s", queryType, path)
	}

	// Check uniqueness - two calls should produce different paths
	// Add small delay on Windows where nanosecond precision may be lower
	if runtime.GOOS == "windows" {
		time.Sleep(time.Millisecond)
	}
	path2 := createTempFilePath(sessionHash, queryType)
	if path == path2 {
		t.Errorf("expected unique paths, got same: %s", path)
	}
}

func TestWriteJSONLFile(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"id": 1, "name": "test1"},
		map[string]interface{}{"id": 2, "name": "test2"},
	}

	// Create temp file path
	tmpPath := filepath.Join(os.TempDir(), "test-write-jsonl.jsonl")
	defer os.Remove(tmpPath)

	// Write JSONL data
	err := writeJSONLFile(tmpPath, data)
	if err != nil {
		t.Fatalf("writeJSONLFile failed: %v", err)
	}

	// Read back and verify
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Verify JSON format
	var record1 map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &record1); err != nil {
		t.Fatalf("failed to parse line 1: %v", err)
	}
	if record1["id"].(float64) != 1 {
		t.Errorf("expected id=1, got %v", record1["id"])
	}
}

func TestCleanupOldFiles(t *testing.T) {
	// Create temp files with different ages
	now := time.Now()

	// Old file (8 days ago)
	oldPath := filepath.Join(os.TempDir(), "meta-cc-mcp-old-123-test.jsonl")
	if err := os.WriteFile(oldPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create old test file: %v", err)
	}
	oldTime := now.Add(-8 * 24 * time.Hour)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old file times: %v", err)
	}
	defer os.Remove(oldPath)

	// Recent file (5 days ago)
	recentPath := filepath.Join(os.TempDir(), "meta-cc-mcp-recent-456-test.jsonl")
	if err := os.WriteFile(recentPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create recent test file: %v", err)
	}
	recentTime := now.Add(-5 * 24 * time.Hour)
	if err := os.Chtimes(recentPath, recentTime, recentTime); err != nil {
		t.Fatalf("Failed to set recent file times: %v", err)
	}
	defer os.Remove(recentPath)

	// Cleanup files older than 7 days
	removed, freedBytes, err := cleanupOldFiles(7)
	if err != nil {
		t.Fatalf("cleanupOldFiles failed: %v", err)
	}

	// Verify old file was removed
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("expected old file to be removed, but it still exists")
	}

	// Verify recent file still exists
	if _, err := os.Stat(recentPath); err != nil {
		t.Errorf("expected recent file to exist, got error: %v", err)
	}

	// Verify removed count
	if len(removed) < 1 {
		t.Errorf("expected at least 1 removed file, got %d", len(removed))
	}

	// Verify freed bytes is positive
	if freedBytes <= 0 {
		t.Errorf("expected positive freed bytes, got %d", freedBytes)
	}
}

func BenchmarkFileWrite(b *testing.B) {
	// Generate 100KB of JSONL data
	data := make([]interface{}, 0)
	for i := 0; i < 500; i++ {
		data = append(data, map[string]interface{}{
			"id":        i,
			"name":      strings.Repeat("test", 10),
			"timestamp": time.Now().Unix(),
			"data":      strings.Repeat("x", 100),
		})
	}

	tmpPath := filepath.Join(os.TempDir(), "bench-write.jsonl")
	defer os.Remove(tmpPath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := writeJSONLFile(tmpPath, data); err != nil {
			b.Fatalf("Failed to write JSONL file: %v", err)
		}
	}
	b.StopTimer()

	// Verify performance requirement
	elapsed := b.Elapsed() / time.Duration(b.N)
	if elapsed > 200*time.Millisecond {
		b.Errorf("file write took %v, expected <200ms", elapsed)
	}
}
