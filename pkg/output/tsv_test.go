package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/yale/meta-cc/internal/parser"
)

func TestFormatTSV(t *testing.T) {
	tools := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Read",
			Input:    map[string]interface{}{"file": "/path/to/file"},
			Output:   "File contents",
			Status:   "success",
			Error:    "",
		},
		{
			UUID:     "uuid-2",
			ToolName: "Bash",
			Input:    map[string]interface{}{"command": "ls -la"},
			Output:   "total 0",
			Status:   "success",
			Error:    "",
		},
	}

	tsv, err := FormatTSV(tools)
	if err != nil {
		t.Fatalf("FormatTSV failed: %v", err)
	}

	// Verify header exists
	if !strings.HasPrefix(tsv, "UUID\t") {
		t.Error("TSV missing header")
	}

	// Verify tab separators in header
	headerLine := strings.Split(tsv, "\n")[0]
	if !strings.Contains(headerLine, "\t") {
		t.Error("TSV header missing tab separators")
	}

	// Verify line count (header + 2 records)
	lines := strings.Split(strings.TrimSpace(tsv), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Verify each data line has tab separators
	for i := 1; i < len(lines); i++ {
		if !strings.Contains(lines[i], "\t") {
			t.Errorf("line %d missing tab separator", i)
		}
	}
}

func TestFormatTSVEmpty(t *testing.T) {
	tools := []parser.ToolCall{}
	tsv, err := FormatTSV(tools)
	if err != nil {
		t.Fatalf("FormatTSV failed: %v", err)
	}

	if tsv != "" {
		t.Errorf("expected empty string for empty input, got: %s", tsv)
	}
}

func TestFormatTSVEscaping(t *testing.T) {
	tools := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Bash",
			Input:    map[string]interface{}{"command": "echo\thello"},
			Output:   "Line 1\nLine 2",
			Status:   "error",
			Error:    "Error with\ttab and\nnewline",
		},
	}

	tsv, err := FormatTSV(tools)
	if err != nil {
		t.Fatalf("FormatTSV failed: %v", err)
	}

	// Verify tab characters in error are escaped
	if strings.Count(tsv, "\t") < 5 { // Header has 5 tabs minimum
		t.Error("TSV structure broken - not enough tabs")
	}

	// Verify the escaped tab appears as \\t in output
	if strings.Contains(tsv, "Error with\ttab") {
		t.Error("Tab character not escaped in error message")
	}

	// Verify escaped version exists
	if !strings.Contains(tsv, "\\t") {
		t.Error("Expected escaped tab character (\\\\t)")
	}

	// Verify newlines are escaped
	if strings.Contains(tsv, "Error with") && strings.Contains(tsv, "\nnewline") {
		lines := strings.Split(tsv, "\n")
		// Should be header + 1 data line = 2 lines total
		if len(lines) > 3 {
			t.Error("Newline in error message not properly escaped")
		}
	}
}

func TestTSVSizeReduction(t *testing.T) {
	// Generate 100 tool calls
	tools := make([]parser.ToolCall, 100)
	for i := 0; i < 100; i++ {
		tools[i] = parser.ToolCall{
			UUID:     "uuid-" + string(rune('a'+i%26)),
			ToolName: "Read",
			Input:    map[string]interface{}{"file": "/path/to/file"},
			Output:   "Sample output data for testing size reduction",
			Status:   "success",
			Error:    "",
		}
	}

	// JSON output
	jsonData, err := json.Marshal(tools)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	jsonSize := len(jsonData)

	// TSV output
	tsvData, err := FormatTSV(tools)
	if err != nil {
		t.Fatalf("FormatTSV failed: %v", err)
	}
	tsvSize := len(tsvData)

	// Verify TSV is at least 50% smaller than JSON
	reduction := 1.0 - float64(tsvSize)/float64(jsonSize)
	if reduction < 0.50 {
		t.Errorf("expected ≥50%% size reduction, got %.1f%% (JSON: %d bytes, TSV: %d bytes)",
			reduction*100, jsonSize, tsvSize)
	}

	t.Logf("Size reduction: %.1f%% (JSON: %d bytes, TSV: %d bytes)",
		reduction*100, jsonSize, tsvSize)
}

func TestFormatProjectedTSV(t *testing.T) {
	projected := []ProjectedToolCall{
		{
			"uuid":   "uuid-1",
			"tool":   "Read",
			"status": "success",
		},
		{
			"uuid":   "uuid-2",
			"tool":   "Bash",
			"status": "error",
		},
	}

	tsv := FormatProjectedTSV(projected)

	// Verify header
	lines := strings.Split(strings.TrimSpace(tsv), "\n")
	if len(lines) != 3 { // header + 2 records
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Verify all lines have tab separators
	for i, line := range lines {
		if !strings.Contains(line, "\t") {
			t.Errorf("line %d missing tab separator", i)
		}
	}
}

// TestFormatGenericTSV_SessionStats tests TSV formatting of SessionStats
func TestFormatGenericTSV_SessionStats(t *testing.T) {
	stats := SessionStats{
		TurnCount:          100,
		UserTurnCount:      50,
		AssistantTurnCount: 50,
		ToolCallCount:      75,
		ErrorCount:         5,
		ErrorRate:          6.67,
		DurationSeconds:    3600,
	}

	tsv, err := FormatGenericTSV(stats)
	if err != nil {
		t.Fatalf("FormatGenericTSV failed: %v", err)
	}

	// Verify vertical format (key\tvalue)
	lines := strings.Split(strings.TrimSpace(tsv), "\n")
	if len(lines) != 7 {
		t.Errorf("expected 7 lines (one per field), got %d", len(lines))
	}

	// Verify each line has exactly 2 columns
	for i, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			t.Errorf("line %d: expected 2 columns, got %d: %s", i, len(parts), line)
		}
	}

	// Verify specific fields exist
	allText := strings.Join(lines, "\n")
	expectedFields := []string{"TurnCount", "ToolCallCount", "ErrorRate"}
	for _, field := range expectedFields {
		if !strings.Contains(allText, field) {
			t.Errorf("missing field: %s", field)
		}
	}
}

// TestFormatGenericTSV_FileStats tests TSV formatting of FileStats slice
func TestFormatGenericTSV_FileStats(t *testing.T) {
	fileStats := []FileStats{
		{
			FilePath:   "/path/to/file1.go",
			ReadCount:  10,
			EditCount:  5,
			WriteCount: 2,
			ErrorCount: 1,
			TotalOps:   18,
			ErrorRate:  0.056,
		},
		{
			FilePath:   "/path/to/file2.go",
			ReadCount:  3,
			EditCount:  2,
			WriteCount: 1,
			ErrorCount: 0,
			TotalOps:   6,
			ErrorRate:  0.0,
		},
	}

	tsv, err := FormatGenericTSV(fileStats)
	if err != nil {
		t.Fatalf("FormatGenericTSV failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(tsv), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Verify header contains expected fields
	header := lines[0]
	expectedFields := []string{"FilePath", "ReadCount", "EditCount", "TotalOps", "ErrorRate"}
	for _, field := range expectedFields {
		if !strings.Contains(header, field) {
			t.Errorf("header missing field: %s", field)
		}
	}
}

// TestFormatGenericTSV_AggregateResult tests TSV formatting of aggregated stats
func TestFormatGenericTSV_AggregateResult(t *testing.T) {
	results := []AggregateResult{
		{
			GroupValue: "Bash",
			Metrics: map[string]interface{}{
				"count":      50,
				"error_rate": 0.04,
			},
		},
		{
			GroupValue: "Read",
			Metrics: map[string]interface{}{
				"count":      30,
				"error_rate": 0.0,
			},
		},
	}

	tsv, err := FormatGenericTSV(results)
	if err != nil {
		t.Fatalf("FormatGenericTSV failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(tsv), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Verify header
	header := lines[0]
	if !strings.Contains(header, "GroupValue") || !strings.Contains(header, "Metrics") {
		t.Errorf("header missing expected fields: %s", header)
	}
}

// TestFormatGenericTSV_TimeSeriesPoint tests TSV formatting of time series data
func TestFormatGenericTSV_TimeSeriesPoint(t *testing.T) {
	points := []TimeSeriesPoint{
		{
			Timestamp: mustParseTime("2025-10-04T10:00:00Z"),
			Value:     42.5,
		},
		{
			Timestamp: mustParseTime("2025-10-04T11:00:00Z"),
			Value:     38.2,
		},
	}

	tsv, err := FormatGenericTSV(points)
	if err != nil {
		t.Fatalf("FormatGenericTSV failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(tsv), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Verify header
	header := lines[0]
	if !strings.Contains(header, "Timestamp") || !strings.Contains(header, "Value") {
		t.Errorf("header missing expected fields: %s", header)
	}
}

// TestFormatGenericTSV_EmptySlice tests empty slice handling
func TestFormatGenericTSV_EmptySlice(t *testing.T) {
	var tools []parser.ToolCall

	tsv, err := FormatGenericTSV(tools)
	if err != nil {
		t.Fatalf("FormatGenericTSV should handle empty slice: %v", err)
	}

	if tsv != "" {
		t.Errorf("expected empty string for empty slice, got: %s", tsv)
	}
}

// TestFormatGenericTSV_UnsupportedType tests error handling for unsupported types
func TestFormatGenericTSV_UnsupportedType(t *testing.T) {
	invalidData := 42 // int is not supported

	_, err := FormatGenericTSV(invalidData)
	if err == nil {
		t.Error("expected error for unsupported type, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported data type") {
		t.Errorf("expected 'unsupported data type' error, got: %v", err)
	}
}

// TestFormatGenericTSV_NestedStructHandling tests complex nested fields
func TestFormatGenericTSV_NestedStructHandling(t *testing.T) {
	// ToolCall has nested Input/Output fields - should be JSON-serialized
	tools := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Read",
			Input:    map[string]interface{}{"file": "/path/to/file", "limit": 100},
			Output:   "File contents here",
			Status:   "success",
			Error:    "",
		},
	}

	tsv, err := FormatGenericTSV(tools)
	if err != nil {
		t.Fatalf("FormatGenericTSV failed: %v", err)
	}

	// Verify that nested Input field is present (JSON-serialized)
	if !strings.Contains(tsv, "{") && !strings.Contains(tsv, "file") {
		t.Error("nested Input field should be JSON-serialized in TSV output")
	}
}

// Helper types for testing (these should match actual types in production)
type SessionStats struct {
	TurnCount          int
	UserTurnCount      int
	AssistantTurnCount int
	ToolCallCount      int
	ErrorCount         int
	ErrorRate          float64
	DurationSeconds    int64
}

type FileStats struct {
	FilePath   string
	ReadCount  int
	EditCount  int
	WriteCount int
	ErrorCount int
	TotalOps   int
	ErrorRate  float64
}

type AggregateResult struct {
	GroupValue string
	Metrics    map[string]interface{}
}

type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float64
}

// Helper function to parse time for tests
func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(fmt.Sprintf("invalid time: %s", s))
	}
	return t
}
