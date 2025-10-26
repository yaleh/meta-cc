package query

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestInspectFiles_SingleFile tests inspection of a single session file
func TestInspectFiles_SingleFile(t *testing.T) {
	// Create test fixture
	tmpDir := t.TempDir()
	sessionFile := filepath.Join(tmpDir, "session1.jsonl")

	// Write test data with various record types
	records := []string{
		`{"type":"user","timestamp":"2025-10-26T10:00:00Z","message":{"content":"Hello"}}`,
		`{"type":"assistant","timestamp":"2025-10-26T10:01:00Z","message":{"content":"Hi"}}`,
		`{"type":"tool","timestamp":"2025-10-26T10:02:00Z","tool_name":"Read"}`,
		`{"type":"tool","timestamp":"2025-10-26T10:03:00Z","tool_name":"Write"}`,
		`{"type":"user","timestamp":"2025-10-26T10:04:00Z","message":{"content":"Goodbye"}}`,
	}

	content := ""
	for _, rec := range records {
		content += rec + "\n"
	}

	if err := os.WriteFile(sessionFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Execute inspection
	result, err := InspectFiles([]string{sessionFile}, false)
	if err != nil {
		t.Fatalf("InspectFiles failed: %v", err)
	}

	// Verify results
	if len(result.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(result.Files))
	}

	file := result.Files[0]
	if file.Path != sessionFile {
		t.Errorf("Expected path %s, got %s", sessionFile, file.Path)
	}

	if file.LineCount != 5 {
		t.Errorf("Expected 5 lines, got %d", file.LineCount)
	}

	// Check record types
	if file.RecordTypes["user"] != 2 {
		t.Errorf("Expected 2 user records, got %d", file.RecordTypes["user"])
	}
	if file.RecordTypes["assistant"] != 1 {
		t.Errorf("Expected 1 assistant record, got %d", file.RecordTypes["assistant"])
	}
	if file.RecordTypes["tool"] != 2 {
		t.Errorf("Expected 2 tool records, got %d", file.RecordTypes["tool"])
	}

	// Check time range
	if file.TimeRange.Start != "2025-10-26T10:00:00Z" {
		t.Errorf("Expected start time 2025-10-26T10:00:00Z, got %s", file.TimeRange.Start)
	}
	if file.TimeRange.End != "2025-10-26T10:04:00Z" {
		t.Errorf("Expected end time 2025-10-26T10:04:00Z, got %s", file.TimeRange.End)
	}

	// Verify summary
	if result.Summary.TotalFiles != 1 {
		t.Errorf("Expected 1 total file, got %d", result.Summary.TotalFiles)
	}
	if result.Summary.TotalRecords != 5 {
		t.Errorf("Expected 5 total records, got %d", result.Summary.TotalRecords)
	}

	// No samples should be included
	if len(file.Samples) != 0 {
		t.Errorf("Expected 0 samples, got %d", len(file.Samples))
	}
}

// TestInspectFiles_MultipleFiles tests inspection of multiple session files
func TestInspectFiles_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first file
	file1 := filepath.Join(tmpDir, "session1.jsonl")
	content1 := `{"type":"user","timestamp":"2025-10-26T10:00:00Z","message":{"content":"Test 1"}}
{"type":"assistant","timestamp":"2025-10-26T10:01:00Z","message":{"content":"Response 1"}}
`
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}

	// Create second file
	file2 := filepath.Join(tmpDir, "session2.jsonl")
	content2 := `{"type":"tool","timestamp":"2025-10-26T11:00:00Z","tool_name":"Read"}
{"type":"tool","timestamp":"2025-10-26T11:01:00Z","tool_name":"Write"}
{"type":"tool","timestamp":"2025-10-26T11:02:00Z","tool_name":"Bash"}
`
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	// Execute inspection
	result, err := InspectFiles([]string{file1, file2}, false)
	if err != nil {
		t.Fatalf("InspectFiles failed: %v", err)
	}

	// Verify results
	if len(result.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result.Files))
	}

	// Verify summary
	if result.Summary.TotalFiles != 2 {
		t.Errorf("Expected 2 total files, got %d", result.Summary.TotalFiles)
	}
	if result.Summary.TotalRecords != 5 {
		t.Errorf("Expected 5 total records, got %d", result.Summary.TotalRecords)
	}

	// Verify first file
	if result.Files[0].LineCount != 2 {
		t.Errorf("Expected 2 lines in file 1, got %d", result.Files[0].LineCount)
	}

	// Verify second file
	if result.Files[1].LineCount != 3 {
		t.Errorf("Expected 3 lines in file 2, got %d", result.Files[1].LineCount)
	}
}

// TestInspectFiles_WithSamples tests sample collection
func TestInspectFiles_WithSamples(t *testing.T) {
	tmpDir := t.TempDir()
	sessionFile := filepath.Join(tmpDir, "session1.jsonl")

	records := []string{
		`{"type":"user","timestamp":"2025-10-26T10:00:00Z","message":{"content":"This is a user message that is longer than 100 characters to test the preview truncation functionality in the sample collector"}}`,
		`{"type":"assistant","timestamp":"2025-10-26T10:01:00Z","message":{"content":"Assistant response"}}`,
		`{"type":"tool","timestamp":"2025-10-26T10:02:00Z","tool_name":"Read"}`,
	}

	content := ""
	for _, rec := range records {
		content += rec + "\n"
	}

	if err := os.WriteFile(sessionFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Execute inspection with samples
	result, err := InspectFiles([]string{sessionFile}, true)
	if err != nil {
		t.Fatalf("InspectFiles failed: %v", err)
	}

	file := result.Files[0]

	// Should have samples for each type
	if len(file.Samples) == 0 {
		t.Errorf("Expected samples, got 0")
	}

	// Verify samples contain required fields
	foundTypes := make(map[string]bool)
	for _, sample := range file.Samples {
		if sample.Type == "" {
			t.Errorf("Sample missing type")
		}
		if sample.Timestamp == "" {
			t.Errorf("Sample missing timestamp")
		}
		if sample.Preview == "" {
			t.Errorf("Sample missing preview")
		}
		if len(sample.Preview) > 100 {
			t.Errorf("Sample preview should be truncated to 100 chars, got %d", len(sample.Preview))
		}
		foundTypes[sample.Type] = true
	}

	// Should have at least one sample
	if len(foundTypes) == 0 {
		t.Errorf("Expected at least one sample type")
	}
}

// TestInspectFiles_EmptyFile tests handling of empty files
func TestInspectFiles_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.jsonl")

	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	result, err := InspectFiles([]string{emptyFile}, false)
	if err != nil {
		t.Fatalf("InspectFiles failed: %v", err)
	}

	file := result.Files[0]
	if file.LineCount != 0 {
		t.Errorf("Expected 0 lines for empty file, got %d", file.LineCount)
	}
	if len(file.RecordTypes) != 0 {
		t.Errorf("Expected 0 record types for empty file, got %d", len(file.RecordTypes))
	}
}

// TestInspectFiles_InvalidJSON tests handling of invalid JSON lines
func TestInspectFiles_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.jsonl")

	content := `{"type":"user","timestamp":"2025-10-26T10:00:00Z"}
invalid json line
{"type":"assistant","timestamp":"2025-10-26T10:01:00Z"}
`
	if err := os.WriteFile(invalidFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	result, err := InspectFiles([]string{invalidFile}, false)
	if err != nil {
		t.Fatalf("InspectFiles failed: %v", err)
	}

	file := result.Files[0]
	// Should process valid lines and skip invalid ones
	if file.LineCount != 3 {
		t.Errorf("Expected 3 lines, got %d", file.LineCount)
	}

	// Should have counted only valid records
	totalValidRecords := 0
	for _, count := range file.RecordTypes {
		totalValidRecords += count
	}
	if totalValidRecords != 2 {
		t.Errorf("Expected 2 valid records, got %d", totalValidRecords)
	}
}

// TestInspectFiles_NonExistentFile tests error handling for missing files
func TestInspectFiles_NonExistentFile(t *testing.T) {
	_, err := InspectFiles([]string{"/nonexistent/file.jsonl"}, false)
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}
}

// TestParseRecordType tests the record type extraction
func TestParseRecordType(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "user type",
			line:     `{"type":"user","timestamp":"2025-10-26T10:00:00Z"}`,
			expected: "user",
		},
		{
			name:     "assistant type",
			line:     `{"type":"assistant","timestamp":"2025-10-26T10:00:00Z"}`,
			expected: "assistant",
		},
		{
			name:     "tool type",
			line:     `{"type":"tool","timestamp":"2025-10-26T10:00:00Z"}`,
			expected: "tool",
		},
		{
			name:     "invalid json",
			line:     `invalid json`,
			expected: "unknown",
		},
		{
			name:     "missing type field",
			line:     `{"timestamp":"2025-10-26T10:00:00Z"}`,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRecordType(tt.line)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestExtractTimeRange tests time range extraction from records
func TestExtractTimeRange(t *testing.T) {
	records := []string{
		`{"type":"user","timestamp":"2025-10-26T10:00:00Z"}`,
		`{"type":"assistant","timestamp":"2025-10-26T10:05:00Z"}`,
		`{"type":"tool","timestamp":"2025-10-26T10:03:00Z"}`,
		`invalid json`,
		`{"type":"user","timestamp":"2025-10-26T10:10:00Z"}`,
	}

	var minTime, maxTime time.Time
	for _, line := range records {
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err == nil {
			if ts, ok := record["timestamp"].(string); ok {
				if t, err := time.Parse(time.RFC3339, ts); err == nil {
					if minTime.IsZero() || t.Before(minTime) {
						minTime = t
					}
					if maxTime.IsZero() || t.After(maxTime) {
						maxTime = t
					}
				}
			}
		}
	}

	expectedStart := "2025-10-26T10:00:00Z"
	expectedEnd := "2025-10-26T10:10:00Z"

	if minTime.Format(time.RFC3339) != expectedStart {
		t.Errorf("Expected start time %s, got %s", expectedStart, minTime.Format(time.RFC3339))
	}
	if maxTime.Format(time.RFC3339) != expectedEnd {
		t.Errorf("Expected end time %s, got %s", expectedEnd, maxTime.Format(time.RFC3339))
	}
}
