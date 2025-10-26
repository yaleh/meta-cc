package query

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test fixtures - sample JSONL data
const (
	testUser1 = `{"type":"user","timestamp":"2025-01-15T10:00:00Z","message":{"content":"fix bug"}}`
	testUser2 = `{"type":"user","timestamp":"2025-01-15T11:00:00Z","message":{"content":"add feature"}}`
	testUser3 = `{"type":"user","timestamp":"2025-01-15T12:00:00Z","message":{"content":"refactor code"}}`
	testAsst1 = `{"type":"assistant","timestamp":"2025-01-15T10:30:00Z","message":{"content":"fixing..."}}`
	testAsst2 = `{"type":"assistant","timestamp":"2025-01-15T11:30:00Z","message":{"content":"adding..."}}`
)

func TestExecuteStage2Query_BasicFilter(t *testing.T) {
	// Test case 1: Basic filter only
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test1.jsonl")

	// Create test file with mixed user and assistant messages
	testData := testUser1 + "\n" + testAsst1 + "\n" + testUser2 + "\n" + testAsst2 + "\n" + testUser3 + "\n"
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Execute query: filter for user messages only
	query := &Stage2Query{
		Files:  []string{testFile},
		Filter: `select(.type == "user")`,
	}

	result, err := ExecuteStage2Query(query)
	if err != nil {
		t.Fatalf("ExecuteStage2Query failed: %v", err)
	}

	// Verify results
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result.Results))
	}

	// Verify all results are user messages
	for i, res := range result.Results {
		resMap, ok := res.(map[string]interface{})
		if !ok {
			t.Errorf("Result %d is not a map", i)
			continue
		}
		if resMap["type"] != "user" {
			t.Errorf("Result %d has type %v, expected user", i, resMap["type"])
		}
	}

	// Verify metadata
	if result.Metadata.FilesProcessed != 1 {
		t.Errorf("Expected 1 file processed, got %d", result.Metadata.FilesProcessed)
	}
	if result.Metadata.ResultsReturned != 3 {
		t.Errorf("Expected 3 results returned, got %d", result.Metadata.ResultsReturned)
	}
	if result.Metadata.TotalRecordsScanned != 5 {
		t.Errorf("Expected 5 records scanned, got %d", result.Metadata.TotalRecordsScanned)
	}
}

func TestExecuteStage2Query_FilterAndSort(t *testing.T) {
	// Test case 2: Filter + sort
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test2.jsonl")

	// Create test file with user messages in non-chronological order
	testData := testUser2 + "\n" + testUser1 + "\n" + testUser3 + "\n"
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Execute query: filter for user messages and sort by timestamp
	query := &Stage2Query{
		Files:  []string{testFile},
		Filter: `select(.type == "user")`,
		Sort:   "sort_by(.timestamp)",
	}

	result, err := ExecuteStage2Query(query)
	if err != nil {
		t.Fatalf("ExecuteStage2Query failed: %v", err)
	}

	// Verify results are sorted
	if len(result.Results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(result.Results))
	}

	// Check timestamps are in ascending order
	timestamps := []string{
		"2025-01-15T10:00:00Z",
		"2025-01-15T11:00:00Z",
		"2025-01-15T12:00:00Z",
	}
	for i, res := range result.Results {
		resMap := res.(map[string]interface{})
		if resMap["timestamp"] != timestamps[i] {
			t.Errorf("Result %d has timestamp %v, expected %s", i, resMap["timestamp"], timestamps[i])
		}
	}
}

func TestExecuteStage2Query_FilterSortTransform(t *testing.T) {
	// Test case 3: Filter + sort + transform
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test3.jsonl")

	testData := testUser1 + "\n" + testUser2 + "\n"
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Execute query: filter, sort, and transform to include only type and timestamp
	query := &Stage2Query{
		Files:     []string{testFile},
		Filter:    `select(.type == "user")`,
		Sort:      "sort_by(.timestamp)",
		Transform: "{type, timestamp}",
	}

	result, err := ExecuteStage2Query(query)
	if err != nil {
		t.Fatalf("ExecuteStage2Query failed: %v", err)
	}

	// Verify results have only type and timestamp fields
	if len(result.Results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(result.Results))
	}

	for i, res := range result.Results {
		resMap := res.(map[string]interface{})
		if len(resMap) != 2 {
			t.Errorf("Result %d has %d fields, expected 2", i, len(resMap))
		}
		if _, ok := resMap["type"]; !ok {
			t.Errorf("Result %d missing type field", i)
		}
		if _, ok := resMap["timestamp"]; !ok {
			t.Errorf("Result %d missing timestamp field", i)
		}
		// Should NOT have message field
		if _, ok := resMap["message"]; ok {
			t.Errorf("Result %d has message field (should be excluded)", i)
		}
	}
}

func TestExecuteStage2Query_FilterAndLimit(t *testing.T) {
	// Test case 4: Filter + limit (verify truncation)
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test4.jsonl")

	testData := testUser1 + "\n" + testUser2 + "\n" + testUser3 + "\n"
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Execute query: filter for user messages with limit of 2
	query := &Stage2Query{
		Files:  []string{testFile},
		Filter: `select(.type == "user")`,
		Limit:  2,
	}

	result, err := ExecuteStage2Query(query)
	if err != nil {
		t.Fatalf("ExecuteStage2Query failed: %v", err)
	}

	// Verify only 2 results returned despite 3 matching
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results (limited), got %d", len(result.Results))
	}

	// Verify truncated flag is set
	if !result.Metadata.Truncated {
		t.Error("Expected truncated=true when limit is reached")
	}

	// Verify results_returned matches actual count
	if result.Metadata.ResultsReturned != 2 {
		t.Errorf("Expected results_returned=2, got %d", result.Metadata.ResultsReturned)
	}
}

func TestExecuteStage2Query_EmptyResultSet(t *testing.T) {
	// Test case 5: Empty result set
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test5.jsonl")

	// Create file with only assistant messages
	testData := testAsst1 + "\n" + testAsst2 + "\n"
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Execute query: filter for user messages (should find none)
	query := &Stage2Query{
		Files:  []string{testFile},
		Filter: `select(.type == "user")`,
	}

	result, err := ExecuteStage2Query(query)
	if err != nil {
		t.Fatalf("ExecuteStage2Query failed: %v", err)
	}

	// Verify empty results
	if len(result.Results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(result.Results))
	}

	// Verify metadata
	if result.Metadata.ResultsReturned != 0 {
		t.Errorf("Expected results_returned=0, got %d", result.Metadata.ResultsReturned)
	}
	if result.Metadata.Truncated {
		t.Error("Expected truncated=false for empty result set")
	}
}

func TestExecuteStage2Query_InvalidJQExpression(t *testing.T) {
	// Test case 6: Invalid jq expression
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test6.jsonl")

	testData := testUser1 + "\n"
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Execute query with invalid jq filter
	query := &Stage2Query{
		Files:  []string{testFile},
		Filter: `select(invalid syntax here)`, // Invalid jq
	}

	_, err := ExecuteStage2Query(query)
	if err == nil {
		t.Error("Expected error for invalid jq expression, got nil")
	}
}

func TestExecuteStage2Query_NonExistentFile(t *testing.T) {
	// Test case 7: Non-existent file
	query := &Stage2Query{
		Files:  []string{"/nonexistent/file.jsonl"},
		Filter: `select(.type == "user")`,
	}

	_, err := ExecuteStage2Query(query)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestExecuteStage2Query_MultipleFiles(t *testing.T) {
	// Test multiple files processing
	tempDir := t.TempDir()
	testFile1 := filepath.Join(tempDir, "test_a.jsonl")
	testFile2 := filepath.Join(tempDir, "test_b.jsonl")

	// Create two files with user messages
	if err := os.WriteFile(testFile1, []byte(testUser1+"\n"+testUser2+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte(testUser3+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	// Execute query across both files
	query := &Stage2Query{
		Files:  []string{testFile1, testFile2},
		Filter: `select(.type == "user")`,
	}

	result, err := ExecuteStage2Query(query)
	if err != nil {
		t.Fatalf("ExecuteStage2Query failed: %v", err)
	}

	// Verify results from both files
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 results from 2 files, got %d", len(result.Results))
	}

	// Verify metadata
	if result.Metadata.FilesProcessed != 2 {
		t.Errorf("Expected 2 files processed, got %d", result.Metadata.FilesProcessed)
	}
}

func TestBuildJQExpression(t *testing.T) {
	tests := []struct {
		name      string
		filter    string
		sort      string
		transform string
		expected  string
	}{
		{
			name:      "filter only",
			filter:    `select(.type == "user")`,
			sort:      "",
			transform: "",
			expected:  `.[] | select(.type == "user")`,
		},
		{
			name:      "filter and sort",
			filter:    `select(.type == "user")`,
			sort:      "sort_by(.timestamp)",
			transform: "",
			expected:  `[.[] | select(.type == "user")] | sort_by(.timestamp) | .[]`,
		},
		{
			name:      "filter, sort, and transform",
			filter:    `select(.type == "user")`,
			sort:      "sort_by(.timestamp)",
			transform: "{type, timestamp}",
			expected:  `[.[] | select(.type == "user")] | sort_by(.timestamp) | .[] | {type, timestamp}`,
		},
		{
			name:      "filter and transform",
			filter:    `select(.type == "user")`,
			sort:      "",
			transform: "{type}",
			expected:  `.[] | select(.type == "user") | {type}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildJQExpression(tt.filter, tt.sort, tt.transform)
			if result != tt.expected {
				t.Errorf("buildJQExpression() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// Benchmark test to verify performance < 100ms for 3MB of data
func BenchmarkExecuteStage2Query_3MB(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "large.jsonl")

	// Generate ~3MB of JSONL data (approximately 3000 records)
	var testData strings.Builder
	for i := 0; i < 3000; i++ {
		record := map[string]interface{}{
			"type":      "user",
			"timestamp": "2025-01-15T10:00:00Z",
			"message": map[string]interface{}{
				"content": "This is a test message with some content to make it realistic sized",
			},
			"metadata": map[string]interface{}{
				"index": i,
				"tags":  []string{"test", "benchmark", "performance"},
			},
		}
		jsonBytes, _ := json.Marshal(record)
		testData.Write(jsonBytes)
		testData.WriteString("\n")
	}

	if err := os.WriteFile(testFile, []byte(testData.String()), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	// Verify file is approximately 3MB
	info, _ := os.Stat(testFile)
	b.Logf("Test file size: %d bytes (~%.2f MB)", info.Size(), float64(info.Size())/1024/1024)

	// Execute query
	query := &Stage2Query{
		Files:  []string{testFile},
		Filter: `select(.type == "user")`,
		Limit:  100, // Limit results to avoid memory issues in benchmark
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := ExecuteStage2Query(query)
		if err != nil {
			b.Fatalf("ExecuteStage2Query failed: %v", err)
		}
		if result.Metadata.ExecutionTimeMs >= 100 {
			b.Errorf("Execution took %dms, expected < 100ms", result.Metadata.ExecutionTimeMs)
		}
	}
}
