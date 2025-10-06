package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestQueryToolsInlineMode tests small result sets use inline mode
func TestQueryToolsInlineMode(t *testing.T) {
	// Simulate small query result (<8KB)
	data := make([]interface{}, 0, 50)
	for i := 0; i < 50; i++ {
		data = append(data, map[string]interface{}{
			"Timestamp": "2025-10-06T10:00:00Z",
			"ToolName":  "Read",
			"Status":    "success",
			"Index":     i,
		})
	}

	params := map[string]interface{}{}

	response, err := adaptResponse(data, params, "query_tools")
	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	respMap, ok := response.(map[string]interface{})
	if !ok {
		t.Fatalf("response is not a map")
	}

	// Verify inline mode
	if mode, ok := respMap["mode"].(string); !ok || mode != "inline" {
		t.Errorf("expected mode=inline, got %v", respMap["mode"])
	}

	// Verify data is present
	if data, ok := respMap["data"].([]interface{}); !ok {
		t.Errorf("expected data array in inline response")
	} else if len(data) != 50 {
		t.Errorf("expected 50 records, got %d", len(data))
	}

	// Verify response serializes correctly
	serialized, err := serializeResponse(response)
	if err != nil {
		t.Fatalf("serializeResponse failed: %v", err)
	}

	if !strings.Contains(serialized, `"mode":"inline"`) {
		t.Errorf("serialized response missing mode field")
	}
}

// TestQueryToolsFileRefMode tests large result sets use file_ref mode
func TestQueryToolsFileRefMode(t *testing.T) {
	// Generate large dataset (>8KB, ~250KB)
	data := make([]interface{}, 0, 5000)
	longString := strings.Repeat("test-data-padding-", 20) // 360 chars to ensure >8KB
	for i := 0; i < 5000; i++ {
		data = append(data, map[string]interface{}{
			"Timestamp": "2025-10-06T10:00:00Z",
			"ToolName":  "Bash",
			"Status":    "success",
			"Duration":  123.45,
			"Args":      longString,
			"Index":     i,
		})
	}

	// Disable max_output_bytes to avoid truncation forcing inline mode
	params := map[string]interface{}{
		"max_output_bytes": 10 * 1024 * 1024, // 10MB
	}

	response, err := adaptResponse(data, params, "query_tools")
	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	respMap, ok := response.(map[string]interface{})
	if !ok {
		t.Fatalf("response is not a map")
	}

	// Verify file_ref mode
	if mode, ok := respMap["mode"].(string); !ok || mode != "file_ref" {
		t.Errorf("expected mode=file_ref, got %v", respMap["mode"])
	}

	// Verify file_ref structure
	fileRef, ok := respMap["file_ref"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected file_ref object in response")
	}

	// Verify required fields
	path, ok := fileRef["path"].(string)
	if !ok {
		t.Fatalf("file_ref missing path")
	}

	sizeBytes, ok := fileRef["size_bytes"].(int64)
	if !ok {
		t.Fatalf("file_ref missing size_bytes")
	}

	lineCount, ok := fileRef["line_count"].(int)
	if !ok {
		t.Fatalf("file_ref missing line_count")
	}

	// Verify file exists and is readable
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("temp file does not exist: %s", path)
	}

	// Verify file contents
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	if int64(len(content)) != sizeBytes {
		t.Errorf("file size mismatch: expected %d, got %d", sizeBytes, len(content))
	}

	// Verify line count
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != lineCount {
		t.Errorf("line count mismatch: expected %d, got %d", lineCount, len(lines))
	}

	// Verify JSONL format (first line should be valid JSON)
	var firstRecord map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &firstRecord); err != nil {
		t.Errorf("temp file is not valid JSONL: %v", err)
	}

	// Clean up temp file
	os.Remove(path)
}

// TestCleanupTempFilesE2E tests the cleanup_temp_files tool end-to-end
func TestCleanupTempFilesE2E(t *testing.T) {
	// Create test temp files with different ages
	sessionHash := "test-session"
	now := time.Now()

	// Create old file (10 days old)
	oldFilePath := filepath.Join(os.TempDir(), "meta-cc-mcp-"+sessionHash+"-old-query_tools.jsonl")
	oldData := []interface{}{map[string]interface{}{"old": "data"}}
	if err := writeJSONLFile(oldFilePath, oldData); err != nil {
		t.Fatalf("failed to create old test file: %v", err)
	}
	// Set modification time to 10 days ago
	oldTime := now.Add(-10 * 24 * time.Hour)
	os.Chtimes(oldFilePath, oldTime, oldTime)

	// Create recent file (2 days old)
	recentFilePath := filepath.Join(os.TempDir(), "meta-cc-mcp-"+sessionHash+"-recent-query_tools.jsonl")
	recentData := []interface{}{map[string]interface{}{"recent": "data"}}
	if err := writeJSONLFile(recentFilePath, recentData); err != nil {
		t.Fatalf("failed to create recent test file: %v", err)
	}
	// Set modification time to 2 days ago
	recentTime := now.Add(-2 * 24 * time.Hour)
	os.Chtimes(recentFilePath, recentTime, recentTime)

	// Execute cleanup (7 day threshold)
	args := map[string]interface{}{
		"max_age_days": 7,
	}

	result, err := executeCleanupTool(args)
	if err != nil {
		t.Fatalf("executeCleanupTool failed: %v", err)
	}

	// Parse result
	var cleanupResult map[string]interface{}
	if err := json.Unmarshal([]byte(result), &cleanupResult); err != nil {
		t.Fatalf("failed to parse cleanup result: %v", err)
	}

	// Verify old file was removed
	if _, err := os.Stat(oldFilePath); !os.IsNotExist(err) {
		t.Errorf("old file should have been removed: %s", oldFilePath)
	}

	// Verify recent file still exists
	if _, err := os.Stat(recentFilePath); os.IsNotExist(err) {
		t.Errorf("recent file should not have been removed: %s", recentFilePath)
	}

	// Verify cleanup result
	removedCount, ok := cleanupResult["removed_count"].(float64)
	if !ok {
		t.Fatalf("cleanup result missing removed_count")
	}

	if int(removedCount) < 1 {
		t.Errorf("expected at least 1 file removed, got %d", int(removedCount))
	}

	// Clean up remaining test file
	os.Remove(recentFilePath)
}

// TestMultipleQueriesConcurrent tests concurrent file writes are race-free
func TestMultipleQueriesConcurrent(t *testing.T) {
	numGoroutines := 10
	data := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]interface{}{
			"index": i,
			"data":  strings.Repeat("x", 100),
		}
	}

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Launch concurrent queries
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			params := map[string]interface{}{
				"output_mode": "file_ref", // Force file_ref mode
			}

			response, err := adaptResponse(data, params, "query_tools")
			if err != nil {
				errors <- err
				return
			}

			// Verify response
			respMap, ok := response.(map[string]interface{})
			if !ok {
				errors <- nil
				return
			}

			// Clean up temp file
			if fileRef, ok := respMap["file_ref"].(map[string]interface{}); ok {
				if path, ok := fileRef["path"].(string); ok {
					os.Remove(path)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Errorf("concurrent query failed: %v", err)
		}
	}
}

// TestFileRefWithReadToolSimulation simulates Claude reading generated file
func TestFileRefWithReadToolSimulation(t *testing.T) {
	// Generate large dataset
	data := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = map[string]interface{}{
			"Timestamp": "2025-10-06T10:00:00Z",
			"ToolName":  "Bash",
			"Status":    "success",
			"Index":     i,
		}
	}

	// Disable max_output_bytes to ensure large result
	params := map[string]interface{}{
		"max_output_bytes": 10 * 1024 * 1024, // 10MB
	}

	// Execute query (should create file_ref)
	response, err := adaptResponse(data, params, "query_tools")
	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	respMap, ok := response.(map[string]interface{})
	if !ok {
		t.Fatalf("response is not a map")
	}

	fileRef, ok := respMap["file_ref"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected file_ref in response")
	}

	path, ok := fileRef["path"].(string)
	if !ok {
		t.Fatalf("file_ref missing path")
	}

	// Simulate Claude using Read tool to read the file
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Claude would fail to read temp file: %v", err)
	}

	// Verify JSONL parsing works
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	for i, line := range lines {
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}

	// Simulate Grep pattern matching
	matchCount := 0
	for _, line := range lines {
		if strings.Contains(line, `"Status":"success"`) {
			matchCount++
		}
	}

	if matchCount != 1000 {
		t.Errorf("expected 1000 matches, got %d", matchCount)
	}

	// Clean up
	os.Remove(path)
}

// BenchmarkLargeQueryFileWrite benchmarks 100KB file write performance
func BenchmarkLargeQueryFileWrite(b *testing.B) {
	// Generate 100KB dataset
	data := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = map[string]interface{}{
			"Timestamp": "2025-10-06T10:00:00Z",
			"ToolName":  "Bash",
			"Status":    "success",
			"Duration":  123.45,
			"Args":      strings.Repeat("a", 100),
		}
	}

	params := map[string]interface{}{
		"output_mode": "file_ref",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response, err := adaptResponse(data, params, "query_tools")
		if err != nil {
			b.Fatalf("adaptResponse failed: %v", err)
		}

		// Clean up temp file
		if respMap, ok := response.(map[string]interface{}); ok {
			if fileRef, ok := respMap["file_ref"].(map[string]interface{}); ok {
				if path, ok := fileRef["path"].(string); ok {
					os.Remove(path)
				}
			}
		}
	}
}

// BenchmarkModeSelection benchmarks mode selection logic
func BenchmarkModeSelection(b *testing.B) {
	testSizes := []int{1024, 8192, 16384, 102400}

	for _, size := range testSizes {
		b.Run(string(rune(size)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				selectOutputMode(size, "")
			}
		})
	}
}
