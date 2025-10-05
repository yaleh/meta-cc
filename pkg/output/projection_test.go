package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yale/meta-cc/internal/parser"
)

// TestProjectToolCalls_BasicProjection tests basic field projection
func TestProjectToolCalls_BasicProjection(t *testing.T) {
	tools := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Bash",
			Input:    map[string]interface{}{"command": "ls"},
			Output:   "file1.txt\nfile2.txt",
			Status:   "success",
			Error:    "",
		},
		{
			UUID:     "uuid-2",
			ToolName: "Read",
			Input:    map[string]interface{}{"file_path": "/test/file.txt"},
			Output:   "file contents",
			Status:   "success",
			Error:    "",
		},
	}

	config := ProjectionConfig{
		Fields: []string{"UUID", "ToolName", "Status"},
	}

	projected, err := ProjectToolCalls(tools, config)
	if err != nil {
		t.Fatalf("ProjectToolCalls failed: %v", err)
	}

	if len(projected) != 2 {
		t.Fatalf("expected 2 projected records, got %d", len(projected))
	}

	// Verify each record has exactly 3 fields
	for i, p := range projected {
		if len(p) != 3 {
			t.Errorf("record %d: expected 3 fields, got %d", i, len(p))
		}

		// Verify expected fields exist
		if _, ok := p["UUID"]; !ok {
			t.Errorf("record %d: missing 'UUID' field", i)
		}
		if _, ok := p["ToolName"]; !ok {
			t.Errorf("record %d: missing 'ToolName' field", i)
		}
		if _, ok := p["Status"]; !ok {
			t.Errorf("record %d: missing 'Status' field", i)
		}

		// Verify unexpected fields don't exist
		if _, ok := p["Input"]; ok {
			t.Errorf("record %d: unexpected 'Input' field", i)
		}
		if _, ok := p["Output"]; ok {
			t.Errorf("record %d: unexpected 'Output' field", i)
		}
	}
}

// TestProjectToolCalls_WithErrorFields tests conditional error field inclusion
func TestProjectToolCalls_WithErrorFields(t *testing.T) {
	tools := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Bash",
			Status:   "success",
			Error:    "",
		},
		{
			UUID:     "uuid-2",
			ToolName: "Read",
			Status:   "error",
			Error:    "file not found",
		},
		{
			UUID:     "uuid-3",
			ToolName: "Edit",
			Status:   "error",
			Error:    "permission denied",
		},
	}

	config := ProjectionConfig{
		Fields:         []string{"UUID", "ToolName"},
		IfErrorInclude: []string{"Error", "Status"},
	}

	projected, err := ProjectToolCalls(tools, config)
	if err != nil {
		t.Fatalf("ProjectToolCalls failed: %v", err)
	}

	// Record 0: success - should have 2 fields (base only)
	if len(projected[0]) != 2 {
		t.Errorf("success record: expected 2 fields, got %d", len(projected[0]))
	}

	// Record 1: error - should have 4 fields (2 base + 2 error)
	if len(projected[1]) != 4 {
		t.Errorf("error record 1: expected 4 fields, got %d", len(projected[1]))
	}
	if _, ok := projected[1]["Error"]; !ok {
		t.Error("error record 1: missing 'Error' field")
	}
	if _, ok := projected[1]["Status"]; !ok {
		t.Error("error record 1: missing 'Status' field")
	}

	// Record 2: error - should have 4 fields (2 base + 2 error)
	if len(projected[2]) != 4 {
		t.Errorf("error record 2: expected 4 fields, got %d", len(projected[2]))
	}
}

// TestProjectToolCalls_NoProjection tests that no projection returns full objects
func TestProjectToolCalls_NoProjection(t *testing.T) {
	tools := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Bash",
			Input:    map[string]interface{}{"command": "ls"},
			Output:   "output",
			Status:   "success",
			Error:    "",
		},
	}

	config := ProjectionConfig{
		Fields: []string{}, // No projection
	}

	projected, err := ProjectToolCalls(tools, config)
	if err != nil {
		t.Fatalf("ProjectToolCalls failed: %v", err)
	}

	// Should return all fields
	if len(projected[0]) < 5 {
		t.Errorf("expected at least 5 fields (all ToolCall fields), got %d", len(projected[0]))
	}
}

// TestProjectionSizeReduction tests that projection reduces output size by ≥70%
func TestProjectionSizeReduction(t *testing.T) {
	// Generate 100 ToolCalls with realistic data
	tools := make([]parser.ToolCall, 100)
	for i := 0; i < 100; i++ {
		tools[i] = parser.ToolCall{
			UUID:     "00000000-0000-0000-0000-000000000000",
			ToolName: "Bash",
			Input: map[string]interface{}{
				"command":     "go test ./...",
				"description": "Run all unit tests",
				"timeout":     120000,
			},
			Output: "PASS\nok  \tgithub.com/yale/meta-cc/pkg/output\t0.023s\nPASS\nok  \tgithub.com/yale/meta-cc/internal/parser\t0.015s",
			Status: "success",
			Error:  "",
		}
	}

	// Full output (no projection)
	fullJSON, err := json.Marshal(tools)
	if err != nil {
		t.Fatalf("failed to marshal full JSON: %v", err)
	}
	fullSize := len(fullJSON)

	// Projected output (3 fields only)
	config := ProjectionConfig{
		Fields: []string{"UUID", "ToolName", "Status"},
	}
	projected, err := ProjectToolCalls(tools, config)
	if err != nil {
		t.Fatalf("ProjectToolCalls failed: %v", err)
	}

	projectedJSON, err := json.Marshal(projected)
	if err != nil {
		t.Fatalf("failed to marshal projected JSON: %v", err)
	}
	projectedSize := len(projectedJSON)

	// Calculate reduction percentage
	reduction := 1.0 - float64(projectedSize)/float64(fullSize)

	t.Logf("Full size: %d bytes, Projected size: %d bytes, Reduction: %.1f%%",
		fullSize, projectedSize, reduction*100)

	// Verify ≥70% size reduction
	if reduction < 0.70 {
		t.Errorf("expected ≥70%% size reduction, got %.1f%%", reduction*100)
	}
}

// TestParseProjectionConfig tests parsing field specs from strings
func TestParseProjectionConfig(t *testing.T) {
	tests := []struct {
		name                string
		fieldsStr           string
		ifErrorIncludeStr   string
		expectedFields      int
		expectedErrorFields int
	}{
		{
			name:                "basic fields",
			fieldsStr:           "UUID,ToolName,Status",
			ifErrorIncludeStr:   "",
			expectedFields:      3,
			expectedErrorFields: 0,
		},
		{
			name:                "fields with spaces",
			fieldsStr:           "UUID, ToolName, Status",
			ifErrorIncludeStr:   "",
			expectedFields:      3,
			expectedErrorFields: 0,
		},
		{
			name:                "with error fields",
			fieldsStr:           "UUID,ToolName",
			ifErrorIncludeStr:   "Error,Output",
			expectedFields:      2,
			expectedErrorFields: 2,
		},
		{
			name:                "empty config",
			fieldsStr:           "",
			ifErrorIncludeStr:   "",
			expectedFields:      0,
			expectedErrorFields: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseProjectionConfig(tt.fieldsStr, tt.ifErrorIncludeStr)

			if len(config.Fields) != tt.expectedFields {
				t.Errorf("expected %d fields, got %d", tt.expectedFields, len(config.Fields))
			}

			if len(config.IfErrorInclude) != tt.expectedErrorFields {
				t.Errorf("expected %d error fields, got %d", tt.expectedErrorFields, len(config.IfErrorInclude))
			}

			// Verify trimming
			for _, field := range config.Fields {
				if field != trimSpace(field) {
					t.Errorf("field '%s' was not trimmed", field)
				}
			}
		})
	}
}

// Helper function to verify trimming
func trimSpace(s string) string {
	// Simple trim implementation for testing
	start := 0
	end := len(s)

	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}

	return s[start:end]
}

// TestProjectToolCalls_InvalidFields tests handling of non-existent fields
func TestProjectToolCalls_InvalidFields(t *testing.T) {
	tools := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Bash",
			Status:   "success",
		},
	}

	config := ProjectionConfig{
		Fields: []string{"UUID", "NonExistentField", "Status"},
	}

	projected, err := ProjectToolCalls(tools, config)
	if err != nil {
		t.Fatalf("ProjectToolCalls failed: %v", err)
	}

	// Should only include fields that exist (UUID and Status)
	// NonExistentField should be silently ignored
	if len(projected[0]) > 3 {
		t.Errorf("expected at most 3 fields, got %d", len(projected[0]))
	}
}

// TestFormatProjectedOutput_JSON tests JSON output formatting
func TestFormatProjectedOutput_JSON(t *testing.T) {
	projected := []ProjectedToolCall{
		{"UUID": "test-1", "ToolName": "Bash", "Status": "success"},
		{"UUID": "test-2", "ToolName": "Read", "Status": "error"},
	}

	output, err := FormatProjectedOutput(projected, "json")
	if err != nil {
		t.Fatalf("FormatProjectedOutput failed: %v", err)
	}

	// Verify valid JSON
	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("expected 2 records, got %d", len(parsed))
	}
}

// TestFormatProjectedOutput_Markdown tests Markdown output formatting
func TestFormatProjectedOutput_Markdown(t *testing.T) {
	projected := []ProjectedToolCall{
		{"UUID": "test-1", "ToolName": "Bash"},
		{"UUID": "test-2", "ToolName": "Read"},
	}

	output, err := FormatProjectedOutput(projected, "markdown")
	if err != nil {
		t.Fatalf("FormatProjectedOutput failed: %v", err)
	}

	// Verify table structure
	if !strings.Contains(output, "| ") {
		t.Error("markdown output missing table structure")
	}
	if !strings.Contains(output, "---") {
		t.Error("markdown output missing separator row")
	}
	if !strings.Contains(output, "ToolName") {
		t.Error("markdown output missing ToolName header")
	}
}

// TestFormatProjectedOutput_CSV tests CSV output formatting
func TestFormatProjectedOutput_CSV(t *testing.T) {
	projected := []ProjectedToolCall{
		{"UUID": "test-1", "ToolName": "Bash"},
		{"UUID": "test-2", "ToolName": "Read"},
	}

	output, err := FormatProjectedOutput(projected, "csv")
	if err != nil {
		t.Fatalf("FormatProjectedOutput failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Verify CSV structure
	if !strings.Contains(lines[0], ",") {
		t.Error("CSV header missing comma separator")
	}
}

// TestFormatProjectedOutput_Empty tests empty data handling
func TestFormatProjectedOutput_Empty(t *testing.T) {
	projected := []ProjectedToolCall{}

	output, err := FormatProjectedOutput(projected, "markdown")
	if err != nil {
		t.Fatalf("FormatProjectedOutput failed: %v", err)
	}

	if output != "No data" {
		t.Errorf("expected 'No data', got '%s'", output)
	}
}

// TestEscapeCSV tests CSV field escaping
func TestEscapeCSV(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"has,comma", "\"has,comma\""},
		{"has\"quote", "\"has\"\"quote\""},
		{"has\nnewline", "\"has\nnewline\""},
		{"normal text", "normal text"},
	}

	for _, tt := range tests {
		result := escapeCSV(tt.input)
		if result != tt.expected {
			t.Errorf("escapeCSV(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
