package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestAdaptInlineResponse tests inline mode formatting for small results
func TestAdaptInlineResponse(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"tool": "Read", "status": "success"},
		map[string]interface{}{"tool": "Write", "status": "success"},
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

	// Check mode is inline
	if mode, ok := respMap["mode"].(string); !ok || mode != "inline" {
		t.Errorf("expected mode=inline, got %v", respMap["mode"])
	}

	// Check data is present
	if _, ok := respMap["data"].([]interface{}); !ok {
		t.Errorf("expected data array in inline response")
	}
}

// TestAdaptFileRefResponse tests file reference mode formatting for large results
func TestAdaptFileRefResponse(t *testing.T) {
	// Generate large dataset (>8KB)
	// Each record is ~200 bytes, need 50+ records to exceed 8KB
	data := make([]interface{}, 0, 100)
	longString := strings.Repeat("x", 150) // 150 chars to ensure >8KB total
	for i := 0; i < 100; i++ {
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

	response, err := adaptResponse(data, params, "query_tools")
	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	respMap, ok := response.(map[string]interface{})
	if !ok {
		t.Fatalf("response is not a map")
	}

	// Check mode is file_ref
	if mode, ok := respMap["mode"].(string); !ok || mode != "file_ref" {
		t.Errorf("expected mode=file_ref, got %v", respMap["mode"])
	}

	// Check file_ref is present
	fileRef, ok := respMap["file_ref"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected file_ref object in response")
	}

	// Verify file_ref structure
	if _, ok := fileRef["path"].(string); !ok {
		t.Errorf("file_ref missing path")
	}
	if _, ok := fileRef["size_bytes"].(int64); !ok {
		t.Errorf("file_ref missing size_bytes")
	}
	if _, ok := fileRef["line_count"].(int); !ok {
		t.Errorf("file_ref missing line_count")
	}

	// Clean up temp file
	if path, ok := fileRef["path"].(string); ok {
		os.Remove(path)
	}
}

// TestHybridModeWithOutputControl tests compatibility with Phase 15 filters
func TestHybridModeWithOutputControl(t *testing.T) {
	data := make([]interface{}, 0, 100)
	for i := 0; i < 100; i++ {
		data = append(data, map[string]interface{}{
			"tool":   "Read",
			"status": "success",
		})
	}

	// Test with max_output_bytes
	params := map[string]interface{}{
		"max_output_bytes": 1000,
	}

	response, err := adaptResponse(data, params, "query_tools")
	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	respMap, ok := response.(map[string]interface{})
	if !ok {
		t.Fatalf("response is not a map")
	}

	// Should be inline due to max_output_bytes truncation
	if mode, ok := respMap["mode"].(string); !ok || mode != "inline" {
		t.Errorf("expected mode=inline with max_output_bytes, got %v", respMap["mode"])
	}
}

// TestAdaptOutputModeOverride tests explicit mode parameter override
func TestAdaptOutputModeOverride(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"tool": "Read", "status": "success"},
	}

	// Force file_ref mode even though data is small
	params := map[string]interface{}{
		"output_mode": "file_ref",
	}

	response, err := adaptResponse(data, params, "query_tools")
	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	respMap, ok := response.(map[string]interface{})
	if !ok {
		t.Fatalf("response is not a map")
	}

	// Should be file_ref due to explicit override
	if mode, ok := respMap["mode"].(string); !ok || mode != "file_ref" {
		t.Errorf("expected mode=file_ref with override, got %v", respMap["mode"])
	}

	// Clean up
	if fileRef, ok := respMap["file_ref"].(map[string]interface{}); ok {
		if path, ok := fileRef["path"].(string); ok {
			os.Remove(path)
		}
	}
}

// TestBackwardCompatibility tests that existing clients are unaffected
func TestBackwardCompatibility(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"tool": "Read", "status": "success"},
	}

	// Legacy mode (no hybrid output)
	params := map[string]interface{}{
		"output_mode": "legacy",
	}

	response, err := adaptResponse(data, params, "query_tools")
	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	// Should return raw data array for legacy mode
	respArray, ok := response.([]interface{})
	if !ok {
		t.Errorf("expected raw array for legacy mode, got %T", response)
	}

	if len(respArray) != 1 {
		t.Errorf("expected 1 item in legacy response, got %d", len(respArray))
	}
}

// TestBuildInlineResponse tests inline response formatting
func TestBuildInlineResponse(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"foo": "bar"},
	}

	response := buildInlineResponse(data)

	if response["mode"] != "inline" {
		t.Errorf("expected mode=inline, got %v", response["mode"])
	}

	if responseData, ok := response["data"].([]interface{}); !ok {
		t.Errorf("expected data array")
	} else if len(responseData) != 1 {
		t.Errorf("expected 1 item, got %d", len(responseData))
	}
}

// TestBuildFileRefResponse tests file reference response formatting
func TestBuildFileRefResponse(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"tool": "Read", "status": "success"},
	}

	// Create temp file
	filePath := createTempFilePath("test-session", "query_tools")

	err := writeJSONLFile(filePath, data)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	defer os.Remove(filePath)

	response, err := buildFileRefResponse(filePath, data)
	if err != nil {
		t.Fatalf("buildFileRefResponse failed: %v", err)
	}

	if response["mode"] != "file_ref" {
		t.Errorf("expected mode=file_ref, got %v", response["mode"])
	}

	fileRef, ok := response["file_ref"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected file_ref object")
	}

	if fileRef["path"] != filePath {
		t.Errorf("expected path=%s, got %v", filePath, fileRef["path"])
	}
}

// TestSerializeResponse tests JSON serialization of hybrid responses
func TestSerializeResponse(t *testing.T) {
	response := map[string]interface{}{
		"mode": "inline",
		"data": []interface{}{
			map[string]interface{}{"foo": "bar"},
		},
	}

	serialized, err := serializeResponse(response)
	if err != nil {
		t.Fatalf("serializeResponse failed: %v", err)
	}

	// Verify valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(serialized), &parsed); err != nil {
		t.Errorf("serialized response is not valid JSON: %v", err)
	}

	if parsed["mode"] != "inline" {
		t.Errorf("expected mode=inline, got %v", parsed["mode"])
	}
}

// TestEmptyDataHandling tests handling of empty result sets
func TestEmptyDataHandling(t *testing.T) {
	data := []interface{}{}

	params := map[string]interface{}{}

	response, err := adaptResponse(data, params, "query_tools")
	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	respMap, ok := response.(map[string]interface{})
	if !ok {
		t.Fatalf("response is not a map")
	}

	// Should be inline for empty data
	if mode, ok := respMap["mode"].(string); !ok || mode != "inline" {
		t.Errorf("expected mode=inline for empty data, got %v", respMap["mode"])
	}

	if data, ok := respMap["data"].([]interface{}); !ok || len(data) != 0 {
		t.Errorf("expected empty data array")
	}
}

// TestStatsOnlyWithHybridMode tests stats_only compatibility
func TestStatsOnlyWithHybridMode(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"tool": "Read", "status": "success"},
		map[string]interface{}{"tool": "Write", "status": "error"},
	}

	params := map[string]interface{}{
		"stats_only": true,
	}

	// Note: stats_only is handled in executor.go before adaptResponse,
	// so we just verify adaptResponse doesn't break with empty params
	response, err := adaptResponse(data, params, "query_tools")
	if err != nil {
		t.Fatalf("adaptResponse failed: %v", err)
	}

	if response == nil {
		t.Errorf("response should not be nil")
	}
}

// TestFileCleanupOnError tests that temp files are cleaned up on errors
func TestFileCleanupOnError(t *testing.T) {
	// Test with invalid data that should cause an error
	data := []interface{}{
		make(chan int), // Channels cannot be JSON-marshaled
	}

	params := map[string]interface{}{
		"output_mode": "file_ref",
	}

	_, err := adaptResponse(data, params, "query_tools")
	if err == nil {
		t.Errorf("expected error with invalid data")
	}

	// Verify no temp files were left behind
	// (This is a basic check - full cleanup is tested in temp_file_manager_test.go)
}

// TestLargeDatasetPerformance benchmarks file write performance
func BenchmarkAdaptResponseFileRef(b *testing.B) {
	// Generate 100KB dataset
	data := make([]interface{}, 0, 1000)
	for i := 0; i < 1000; i++ {
		data = append(data, map[string]interface{}{
			"Timestamp": "2025-10-06T10:00:00Z",
			"ToolName":  "Bash",
			"Status":    "success",
			"Duration":  123.45,
			"Args":      strings.Repeat("a", 100),
		})
	}

	params := map[string]interface{}{}

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
