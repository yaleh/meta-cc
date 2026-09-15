package files

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestInspectFiles_LargeImageLine_NoError verifies that InspectFiles handles
// lines larger than 10MB (e.g. base64 image data) without error.
func TestInspectFiles_LargeImageLine_NoError(t *testing.T) {
	tmpDir := t.TempDir()
	sessionFile := filepath.Join(tmpDir, "session_large.jsonl")

	// Build a line with a 5MB base64 image payload using strings.Repeat
	imageData := strings.Repeat("A", 5*1024*1024)
	imageLine := `{"type":"user","timestamp":"2025-10-26T10:00:00Z","message":{"content":[{"type":"image","source":{"type":"base64","media_type":"image/png","data":"` + imageData + `"}}]}}`
	normalLine := `{"type":"assistant","timestamp":"2025-10-26T10:01:00Z","message":{"content":"hi"}}`

	content := imageLine + "\n" + normalLine + "\n"
	if err := os.WriteFile(sessionFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result := InspectFiles([]string{sessionFile}, false)

	file := result.Files[0]
	// Both lines should be counted (image line truncated to valid JSON by StrategyDefault)
	if file.LineCount < 1 {
		t.Errorf("Expected at least 1 line, got %d", file.LineCount)
	}
}

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
	result := InspectFiles([]string{sessionFile}, false)

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
	result := InspectFiles([]string{file1, file2}, false)

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
	result := InspectFiles([]string{sessionFile}, true)

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

	result := InspectFiles([]string{emptyFile}, false)

	file := result.Files[0]
	if file.LineCount != 0 {
		t.Errorf("Expected 0 lines for empty file, got %d", file.LineCount)
	}
	if len(file.RecordTypes) != 0 {
		t.Errorf("Expected 0 record types for empty file, got %d", len(file.RecordTypes))
	}

	// DIR-098: the 8eda8f4e shape (an empty session file) must be *named*, not
	// merely counted as zero lines. This is the whole point — a zero-line
	// result was previously indistinguishable from a thin-but-fine session.
	if !file.Empty {
		t.Error("expected empty=true for a zero-byte file")
	}
	if file.Entries != 0 {
		t.Errorf("expected 0 entries, got %d", file.Entries)
	}
	if len(result.MalformedFiles) != 1 {
		t.Fatalf("expected the empty file named in malformed_files, got %#v", result.MalformedFiles)
	}
	if result.MalformedFiles[0].File != emptyFile || result.MalformedFiles[0].Reason == "" {
		t.Errorf("expected %s named with a reason, got %#v", emptyFile, result.MalformedFiles[0])
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

	result := InspectFiles([]string{invalidFile}, false)

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

	// DIR-098: the corrupt line is named (not silently skipped), yet the file is
	// NOT excluded — it yielded message entries, so the corpus keeps it. Naming
	// the defect and excluding the file are two different judgments.
	if file.Error == "" {
		t.Error("expected the unparseable line to be reported in error")
	}
	if file.Entries != 2 {
		t.Errorf("Expected 2 message entries, got %d", file.Entries)
	}
	if len(result.MalformedFiles) != 0 {
		t.Errorf("a file that yielded entries must not be excluded, got %#v", result.MalformedFiles)
	}
}

// TestInspectFiles_NonExistentFile pins the DIR-098 contract: a file that
// cannot be read is reported as structured per-file data (and named in
// malformed_files), not raised as a whole-batch error. The earlier version of
// this test asserted the opposite — that InspectFiles returns an error — which
// is exactly the fail-fast shape that let one bad file erase the rest.
func TestInspectFiles_NonExistentFile(t *testing.T) {
	result := InspectFiles([]string{"/nonexistent/file.jsonl"}, false)

	if len(result.Files) != 1 {
		t.Fatalf("expected the unreadable file to still be reported, got %d entries", len(result.Files))
	}
	file := result.Files[0]
	if file.Parseable {
		t.Error("expected parseable=false for a non-existent file")
	}
	if file.Error == "" {
		t.Error("expected a non-empty error string naming the cause")
	}
	if len(result.MalformedFiles) != 1 || result.MalformedFiles[0].File != "/nonexistent/file.jsonl" {
		t.Errorf("expected the file named in malformed_files, got %#v", result.MalformedFiles)
	}
	if result.MalformedFiles[0].Reason == "" {
		t.Error("expected a non-empty reason for the malformed file")
	}
}

// TestInspectFiles_OneBadFileDoesNotEraseTheRest is the batch-tolerance half of
// the DIR-098 contract: a healthy file alongside an unreadable one is still
// inspected in full.
func TestInspectFiles_OneBadFileDoesNotEraseTheRest(t *testing.T) {
	tmpDir := t.TempDir()
	healthy := filepath.Join(tmpDir, "healthy.jsonl")
	content := `{"type":"user","timestamp":"2025-10-26T10:00:00Z"}` + "\n" +
		`{"type":"assistant","timestamp":"2025-10-26T10:01:00Z"}` + "\n"
	if err := os.WriteFile(healthy, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create healthy file: %v", err)
	}

	result := InspectFiles([]string{filepath.Join(tmpDir, "missing.jsonl"), healthy}, false)

	if len(result.Files) != 2 {
		t.Fatalf("expected both files reported, got %d", len(result.Files))
	}
	if !result.Files[1].Parseable || result.Files[1].Entries != 2 {
		t.Errorf("healthy file should still be inspected: %#v", result.Files[1])
	}
	if result.Summary.TotalRecords != 2 {
		t.Errorf("expected the healthy file's records to survive, got %d", result.Summary.TotalRecords)
	}
	if len(result.MalformedFiles) != 1 {
		t.Errorf("expected exactly the unreadable file named as malformed, got %#v", result.MalformedFiles)
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
