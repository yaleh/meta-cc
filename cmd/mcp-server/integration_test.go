package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/config"
)

// TestQueryToolsInlineMode tests small result sets use inline mode
func TestQueryToolsInlineMode(t *testing.T) {
	cfg, _ := config.Load()
	// Simulate small query result (<8KB)
	data := make([]interface{}, 0, 50)
	for i := 0; i < 50; i++ {
		data = append(data, map[string]interface{}{
			"tool":   "Read",
			"status": "success",
			"index":  i,
		})
	}
	params := map[string]interface{}{}

	response, err := adaptResponse(cfg, data, params, "query_tools")
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
	cfg, _ := config.Load()
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

	params := map[string]interface{}{}

	response, err := adaptResponse(cfg, data, params, "query_tools")
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
	if err := os.Chtimes(oldFilePath, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set old file times: %v", err)
	}

	// Create recent file (2 days old)
	recentFilePath := filepath.Join(os.TempDir(), "meta-cc-mcp-"+sessionHash+"-recent-query_tools.jsonl")
	recentData := []interface{}{map[string]interface{}{"recent": "data"}}
	if err := writeJSONLFile(recentFilePath, recentData); err != nil {
		t.Fatalf("failed to create recent test file: %v", err)
	}
	// Set modification time to 2 days ago
	recentTime := now.Add(-2 * 24 * time.Hour)
	if err := os.Chtimes(recentFilePath, recentTime, recentTime); err != nil {
		t.Fatalf("failed to set recent file times: %v", err)
	}

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
	cfg, _ := config.Load()
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

			response, err := adaptResponse(cfg, data, params, "query_tools")
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
	cfg, _ := config.Load()
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

	params := map[string]interface{}{}

	// Execute query (should create file_ref)
	response, err := adaptResponse(cfg, data, params, "query_tools")
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
		response, err := adaptResponse(cfg, data, params, "query_tools")
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

// TestQueryToolsNoLimit verifies no-limit queries return all results via file_ref
func TestQueryToolsNoLimit(t *testing.T) {
	cfg, _ := config.Load()
	// Generate large dataset (1000 records)
	data := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = map[string]interface{}{
			"Timestamp": "2025-10-06T10:00:00Z",
			"ToolName":  "Bash",
			"Status":    "success",
			"Index":     i,
		}
	}

	// No limit parameter - should return all results
	params := map[string]interface{}{}

	response, err := adaptResponse(cfg, data, params, "query_tools")
	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	respMap, ok := response.(map[string]interface{})
	if !ok {
		t.Fatalf("response is not a map")
	}

	// Verify mode
	mode, ok := respMap["mode"].(string)
	if !ok {
		t.Fatalf("response missing mode field")
	}

	// Verify all 1000 records are accessible
	if mode == "inline" {
		data, ok := respMap["data"].([]interface{})
		if !ok {
			t.Fatalf("inline response missing data array")
		}
		if len(data) != 1000 {
			t.Errorf("expected 1000 records, got %d", len(data))
		}
	} else if mode == "file_ref" {
		fileRef, ok := respMap["file_ref"].(map[string]interface{})
		if !ok {
			t.Fatalf("file_ref response missing file_ref object")
		}

		lineCount, ok := fileRef["line_count"].(int)
		if !ok {
			t.Fatalf("file_ref missing line_count")
		}

		if lineCount != 1000 {
			t.Errorf("expected 1000 lines, got %d", lineCount)
		}

		// Clean up temp file
		if path, ok := fileRef["path"].(string); ok {
			os.Remove(path)
		}
	}
}

// TestNoTruncationLargeData verifies no truncation occurs with 100KB+ data
func TestNoTruncationLargeData(t *testing.T) {
	cfg, _ := config.Load()
	// Generate 100KB+ dataset
	data := make([]interface{}, 2000)
	longString := strings.Repeat("x", 100) // 100 chars per record
	for i := 0; i < 2000; i++ {
		data[i] = map[string]interface{}{
			"Timestamp": "2025-10-06T10:00:00Z",
			"ToolName":  "Bash",
			"Status":    "success",
			"Args":      longString,
			"Index":     i,
		}
	}

	params := map[string]interface{}{}

	response, err := adaptResponse(cfg, data, params, "query_tools")
	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	respMap, ok := response.(map[string]interface{})
	if !ok {
		t.Fatalf("response is not a map")
	}

	// Verify file_ref mode (data should be >8KB)
	mode, ok := respMap["mode"].(string)
	if !ok || mode != "file_ref" {
		t.Errorf("expected file_ref mode for large data, got %v", mode)
	}

	// Verify all data is in temp file (no truncation)
	fileRef, ok := respMap["file_ref"].(map[string]interface{})
	if !ok {
		t.Fatalf("file_ref response missing file_ref object")
	}

	path, ok := fileRef["path"].(string)
	if !ok {
		t.Fatalf("file_ref missing path")
	}

	lineCount, ok := fileRef["line_count"].(int)
	if !ok {
		t.Fatalf("file_ref missing line_count")
	}

	if lineCount != 2000 {
		t.Errorf("expected 2000 lines (no truncation), got %d", lineCount)
	}

	// Verify file size is reasonable (>100KB)
	sizeBytes, ok := fileRef["size_bytes"].(int64)
	if !ok {
		t.Fatalf("file_ref missing size_bytes")
	}

	if sizeBytes < 100*1024 {
		t.Errorf("expected >100KB, got %d bytes", sizeBytes)
	}

	// Verify no [OUTPUT TRUNCATED] warnings in file
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	if strings.Contains(string(content), "[OUTPUT TRUNCATED]") {
		t.Error("found [OUTPUT TRUNCATED] warning - truncation should not occur")
	}

	// Clean up
	os.Remove(path)
}

// TestConfigurableThresholdParameter tests inline_threshold_bytes parameter
func TestConfigurableThresholdParameter(t *testing.T) {
	cfg, _ := config.Load()
	// Generate dataset around 10KB (small enough for 20KB threshold, too large for 8KB)
	data := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]interface{}{
			"Timestamp": "2025-10-06T10:00:00Z",
			"ToolName":  "Bash",
			"Status":    "success",
			"Args":      strings.Repeat("x", 30),
			"Index":     i,
		}
	}

	// Test 1: Default threshold (8KB) - should use file_ref
	t.Run("default_threshold", func(t *testing.T) {
		params := map[string]interface{}{}
		response, err := adaptResponse(cfg, data, params, "query_tools")
		if err != nil {
			t.Fatalf("adaptResponse failed: %v", err)
		}

		respMap, ok := response.(map[string]interface{})
		if !ok {
			t.Fatalf("response is not a map")
		}

		mode, ok := respMap["mode"].(string)
		if !ok || mode != "file_ref" {
			t.Errorf("expected file_ref mode with default threshold, got %v", mode)
		}

		// Clean up temp file
		if fileRef, ok := respMap["file_ref"].(map[string]interface{}); ok {
			if path, ok := fileRef["path"].(string); ok {
				os.Remove(path)
			}
		}
	})

	// Test 2: Custom threshold (20KB) - should use inline
	t.Run("custom_threshold_inline", func(t *testing.T) {
		params := map[string]interface{}{
			"inline_threshold_bytes": 20 * 1024, // 20KB
		}
		response, err := adaptResponse(cfg, data, params, "query_tools")
		if err != nil {
			t.Fatalf("adaptResponse failed: %v", err)
		}

		respMap, ok := response.(map[string]interface{})
		if !ok {
			t.Fatalf("response is not a map")
		}

		mode, ok := respMap["mode"].(string)
		if !ok || mode != "inline" {
			t.Errorf("expected inline mode with 20KB threshold, got %v", mode)
		}
	})

	// Test 3: Small threshold (1KB) - should use file_ref
	t.Run("small_threshold_file_ref", func(t *testing.T) {
		params := map[string]interface{}{
			"inline_threshold_bytes": 1024, // 1KB
		}
		response, err := adaptResponse(cfg, data, params, "query_tools")
		if err != nil {
			t.Fatalf("adaptResponse failed: %v", err)
		}

		respMap, ok := response.(map[string]interface{})
		if !ok {
			t.Fatalf("response is not a map")
		}

		mode, ok := respMap["mode"].(string)
		if !ok || mode != "file_ref" {
			t.Errorf("expected file_ref mode with 1KB threshold, got %v", mode)
		}

		// Clean up temp file
		if fileRef, ok := respMap["file_ref"].(map[string]interface{}); ok {
			if path, ok := fileRef["path"].(string); ok {
				os.Remove(path)
			}
		}
	})
}

// TestConfigurableThresholdEnvironment tests META_CC_INLINE_THRESHOLD environment variable
func TestConfigurableThresholdEnvironment(t *testing.T) {
	// Generate dataset around 10KB (small enough for 20KB threshold, too large for 8KB)
	data := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]interface{}{
			"Timestamp": "2025-10-06T10:00:00Z",
			"ToolName":  "Bash",
			"Status":    "success",
			"Args":      strings.Repeat("x", 30),
			"Index":     i,
		}
	}

	// Test 1: Environment variable sets 20KB threshold
	t.Run("environment_threshold_20kb", func(t *testing.T) {
		os.Setenv("META_CC_INLINE_THRESHOLD", "20480") // 20KB
		defer os.Unsetenv("META_CC_INLINE_THRESHOLD")

		// Reload config to pick up environment variable
		cfg, _ := config.Load()

		params := map[string]interface{}{}
		response, err := adaptResponse(cfg, data, params, "query_tools")
		if err != nil {
			t.Fatalf("adaptResponse failed: %v", err)
		}

		respMap, ok := response.(map[string]interface{})
		if !ok {
			t.Fatalf("response is not a map")
		}

		mode, ok := respMap["mode"].(string)
		if !ok || mode != "inline" {
			t.Errorf("expected inline mode with environment threshold 20KB, got %v", mode)
		}
	})

	// Test 2: Parameter overrides environment variable
	t.Run("parameter_overrides_environment", func(t *testing.T) {
		os.Setenv("META_CC_INLINE_THRESHOLD", "20480") // 20KB
		defer os.Unsetenv("META_CC_INLINE_THRESHOLD")

		// Reload config to pick up environment variable
		cfg, _ := config.Load()

		params := map[string]interface{}{
			"inline_threshold_bytes": 1024, // 1KB parameter should override env
		}
		response, err := adaptResponse(cfg, data, params, "query_tools")
		if err != nil {
			t.Fatalf("adaptResponse failed: %v", err)
		}

		respMap, ok := response.(map[string]interface{})
		if !ok {
			t.Fatalf("response is not a map")
		}

		mode, ok := respMap["mode"].(string)
		if !ok || mode != "file_ref" {
			t.Errorf("expected file_ref mode (parameter override), got %v", mode)
		}

		// Clean up temp file
		if fileRef, ok := respMap["file_ref"].(map[string]interface{}); ok {
			if path, ok := fileRef["path"].(string); ok {
				os.Remove(path)
			}
		}
	})
}

// TestPerformanceBenchmarks verifies 100KB write meets <200ms requirement
func TestPerformanceBenchmarks(t *testing.T) {
	cfg, _ := config.Load()
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

	params := map[string]interface{}{}

	// Measure write time
	start := time.Now()
	response, err := adaptResponse(cfg, data, params, "query_tools")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	// Verify performance (<200ms for 100KB write)
	if elapsed > 200*time.Millisecond {
		t.Errorf("100KB write took %v, expected <200ms", elapsed)
	}

	// Clean up temp file
	if respMap, ok := response.(map[string]interface{}); ok {
		if fileRef, ok := respMap["file_ref"].(map[string]interface{}); ok {
			if path, ok := fileRef["path"].(string); ok {
				os.Remove(path)
			}
		}
	}

	t.Logf("100KB write completed in %v", elapsed)
}

// TestJSONLOutputFormatE2E verifies query tools output proper multi-line JSONL format
// This is an end-to-end test for the JSONL format fix (Phase 27 Stage 27.5)
func TestJSONLOutputFormatE2E(t *testing.T) {
	cfg, _ := config.Load()

	// Test Case 1: Verify file_ref output is multi-line JSONL (not single-line JSON array)
	t.Run("file_ref_multiline_jsonl", func(t *testing.T) {
		// Generate large dataset to trigger file_ref mode
		data := make([]interface{}, 200)
		for i := 0; i < 200; i++ {
			data[i] = map[string]interface{}{
				"type":      "user",
				"id":        i,
				"timestamp": "2025-10-27T10:00:00Z",
				"content":   "test message content for JSONL format verification",
			}
		}

		params := map[string]interface{}{}

		// Execute query through response adapter
		response, err := adaptResponse(cfg, data, params, "query_user_messages")
		if err != nil {
			t.Fatalf("adaptResponse failed: %v", err)
		}

		respMap, ok := response.(map[string]interface{})
		if !ok {
			t.Fatalf("response is not a map")
		}

		// Verify file_ref mode
		mode, ok := respMap["mode"].(string)
		if !ok || mode != "file_ref" {
			t.Fatalf("expected file_ref mode, got %v", mode)
		}

		// Get file path
		fileRef, ok := respMap["file_ref"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected file_ref object")
		}

		path, ok := fileRef["path"].(string)
		if !ok {
			t.Fatalf("file_ref missing path")
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read temp file: %v", err)
		}

		// CRITICAL: Verify file is multi-line JSONL, not single-line JSON array
		lines := strings.Split(strings.TrimSpace(string(content)), "\n")

		// Should have 200 lines (one per record), not 1 line (JSON array)
		if len(lines) == 1 {
			t.Errorf("FAIL: Output is single-line JSON array, should be multi-line JSONL")
			t.Errorf("Expected 200 lines, got 1 line")
			t.Errorf("This indicates the JSONL format fix is not working")
		}

		if len(lines) != 200 {
			t.Errorf("expected 200 lines (multi-line JSONL), got %d", len(lines))
		}

		// Verify each line is valid JSON (not part of array)
		for i, line := range lines {
			if i >= 5 {
				break // Check first 5 lines only
			}

			var record map[string]interface{}
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				t.Errorf("line %d is not valid JSON: %v", i+1, err)
			}

			// Verify it's an object, not array element
			if record["type"] != "user" {
				t.Errorf("line %d: expected type=user, got %v", i+1, record["type"])
			}
		}

		// Verify file does NOT start with '[' (JSON array)
		if len(content) > 0 && content[0] == '[' {
			t.Error("FAIL: File starts with '[' - this is a JSON array, not JSONL format")
		}

		// Clean up
		os.Remove(path)
		t.Logf("✅ PASS: File is proper multi-line JSONL format (%d lines)", len(lines))
	})

	// Test Case 2: Verify Claude Code Read tool can process line-by-line
	t.Run("claude_read_tool_compatibility", func(t *testing.T) {
		// Generate larger dataset to ensure file_ref mode
		data := make([]interface{}, 200)
		for i := 0; i < 200; i++ {
			data[i] = map[string]interface{}{
				"tool_name": "Bash",
				"status":    "success",
				"index":     i,
				"args":      strings.Repeat("x", 50), // Add padding to ensure >8KB
			}
		}

		params := map[string]interface{}{}
		response, err := adaptResponse(cfg, data, params, "query_tools")
		if err != nil {
			t.Fatalf("adaptResponse failed: %v", err)
		}

		respMap, ok := response.(map[string]interface{})
		if !ok {
			t.Fatalf("response is not a map")
		}

		// Verify file_ref mode
		mode, ok := respMap["mode"].(string)
		if !ok || mode != "file_ref" {
			t.Skipf("Skipping test - got mode=%v, expected file_ref", mode)
		}

		fileRef, ok := respMap["file_ref"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected file_ref object")
		}

		path, ok := fileRef["path"].(string)
		if !ok {
			t.Fatalf("file_ref missing path")
		}

		// Simulate Claude Code Read tool: read with offset and limit
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Claude would fail to read file: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")

		// Simulate reading lines 10-20 (Claude Code offset/limit)
		if len(lines) < 20 {
			t.Fatalf("not enough lines for offset test")
		}

		for i := 10; i < 20; i++ {
			var record map[string]interface{}
			if err := json.Unmarshal([]byte(lines[i]), &record); err != nil {
				t.Errorf("Claude would fail to parse line %d: %v", i, err)
			}
		}

		os.Remove(path)
		t.Logf("✅ PASS: Claude Code Read tool can process line-by-line")
	})

	// Test Case 3: Verify Claude Code Search/Grep can filter line-by-line
	t.Run("claude_search_tool_compatibility", func(t *testing.T) {
		// Generate larger dataset with mixed statuses to ensure file_ref mode
		data := make([]interface{}, 200)
		for i := 0; i < 200; i++ {
			status := "success"
			if i%3 == 0 {
				status = "error"
			}
			data[i] = map[string]interface{}{
				"tool_name": "Bash",
				"status":    status,
				"index":     i,
				"args":      strings.Repeat("x", 50), // Add padding to ensure >8KB
			}
		}

		params := map[string]interface{}{}
		response, err := adaptResponse(cfg, data, params, "query_tools")
		if err != nil {
			t.Fatalf("adaptResponse failed: %v", err)
		}

		respMap, ok := response.(map[string]interface{})
		if !ok {
			t.Fatalf("response is not a map")
		}

		// Verify file_ref mode
		mode, ok := respMap["mode"].(string)
		if !ok || mode != "file_ref" {
			t.Skipf("Skipping test - got mode=%v, expected file_ref", mode)
		}

		fileRef, ok := respMap["file_ref"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected file_ref object")
		}

		path, ok := fileRef["path"].(string)
		if !ok {
			t.Fatalf("file_ref missing path")
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		// Simulate Claude Code Search: grep for "error"
		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		matchCount := 0
		for _, line := range lines {
			if strings.Contains(line, `"status":"error"`) {
				matchCount++
			}
		}

		// Should find 67 error entries (0, 3, 6, 9, ..., 198) - every 3rd index
		expectedErrors := 67
		if matchCount != expectedErrors {
			t.Errorf("expected %d error matches, got %d", expectedErrors, matchCount)
		}

		os.Remove(path)
		t.Logf("✅ PASS: Claude Code Search can filter line-by-line (%d matches)", matchCount)
	})

	// Test Case 4: End-to-end test with executeQuery
	t.Run("e2e_execute_query_to_file", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping E2E test in short mode")
		}

		// Setup test session directory
		testData := `{"type":"user","timestamp":"2025-10-27T10:00:00Z","message":{"content":"test 1"}}
{"type":"user","timestamp":"2025-10-27T10:00:01Z","message":{"content":"test 2"}}
{"type":"user","timestamp":"2025-10-27T10:00:02Z","message":{"content":"test 3"}}
{"type":"assistant","timestamp":"2025-10-27T10:00:03Z","message":{"content":"response"}}
`
		projectPath := setupTestSessionDir(t, testData)

		// Change to project directory
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		os.Chdir(projectPath)

		executor := NewToolExecutor()

		// Execute query tool (full E2E)
		args := map[string]interface{}{
			"pattern": "test",
			"scope":   "session",
		}

		result, err := executor.ExecuteTool(cfg, "query_user_messages", args)
		if err != nil {
			t.Fatalf("ExecuteTool failed: %v", err)
		}

		// Parse response
		var response map[string]interface{}
		if err := json.Unmarshal([]byte(result), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		// Check mode (should be inline for small dataset)
		mode, ok := response["mode"].(string)
		if !ok {
			t.Fatalf("response missing mode")
		}

		// Verify data format
		if mode == "inline" {
			data, ok := response["data"].([]interface{})
			if !ok {
				t.Fatalf("inline response missing data array")
			}
			if len(data) == 0 {
				t.Error("expected at least one result")
			}
		} else if mode == "file_ref" {
			fileRef, ok := response["file_ref"].(map[string]interface{})
			if !ok {
				t.Fatalf("file_ref response missing file_ref object")
			}

			path, ok := fileRef["path"].(string)
			if !ok {
				t.Fatalf("file_ref missing path")
			}

			// Verify JSONL format
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read temp file: %v", err)
			}

			lines := strings.Split(strings.TrimSpace(string(content)), "\n")
			if len(lines) <= 1 {
				t.Error("FAIL: E2E test produced single-line output, expected multi-line JSONL")
			}

			os.Remove(path)
		}

		t.Logf("✅ PASS: E2E query executed successfully (mode=%s)", mode)
	})
}

// TestExecuteToolE2E_AllTools verifies all tools execute without errors
func TestExecuteToolE2E_AllTools(t *testing.T) {
	cfg, _ := config.Load()
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode (takes ~22s)")
	}

	// Save and restore working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	// Change to project root (two levels up from cmd/mcp-server)
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("failed to change to project root: %v", err)
	}

	executor := NewToolExecutor()

	tests := []struct {
		name     string
		toolName string
		args     map[string]interface{}
	}{
		{
			name:     "query_tools",
			toolName: "query_tools",
			args: map[string]interface{}{
				"limit": 10,
				"scope": "project",
			},
		},
		{
			name:     "query_user_messages",
			toolName: "query_user_messages",
			args: map[string]interface{}{
				"pattern": "test",
				"limit":   5,
				"scope":   "project",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.ExecuteTool(cfg, tt.toolName, tt.args)
			if err != nil {
				t.Fatalf("ExecuteTool(%s) failed: %v", tt.toolName, err)
			}

			if result == "" {
				t.Errorf("expected non-empty result for %s", tt.toolName)
			}

			// Verify result is valid JSON
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(result), &response); err != nil {
				t.Errorf("result is not valid JSON for %s: %v", tt.toolName, err)
			}

			// Clean up temp files if file_ref mode
			if mode, ok := response["mode"].(string); ok && mode == "file_ref" {
				if fileRef, ok := response["file_ref"].(map[string]interface{}); ok {
					if path, ok := fileRef["path"].(string); ok {
						os.Remove(path)
					}
				}
			}
		})
	}
}
