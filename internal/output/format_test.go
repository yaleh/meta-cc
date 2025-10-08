package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestFormatOutput_JSONL(t *testing.T) {
	tools := []parser.ToolCall{
		{
			UUID:     "test-uuid-1",
			ToolName: "Bash",
			Status:   "success",
			Error:    "",
		},
		{
			UUID:     "test-uuid-2",
			ToolName: "Edit",
			Status:   "error",
			Error:    "file not found",
		},
	}

	output, err := FormatOutput(tools, "jsonl")
	if err != nil {
		t.Fatalf("FormatOutput failed: %v", err)
	}

	// Verify JSONL format (one object per line)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (JSONL), got %d", len(lines))
	}

	// Verify first line
	var first parser.ToolCall
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("line 1 is not valid JSON: %v", err)
	}

	if first.UUID != "test-uuid-1" {
		t.Errorf("expected UUID 'test-uuid-1', got '%s'", first.UUID)
	}
}

func TestFormatOutput_TSV(t *testing.T) {
	tools := []parser.ToolCall{
		{
			UUID:     "test-uuid-1",
			ToolName: "Bash",
			Status:   "success",
			Error:    "",
		},
		{
			UUID:     "test-uuid-2",
			ToolName: "Edit",
			Status:   "error",
			Error:    "file not found",
		},
	}

	output, err := FormatOutput(tools, "tsv")
	if err != nil {
		t.Fatalf("FormatOutput failed: %v", err)
	}

	// Verify TSV format (tab-separated values with header)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("expected 3 lines (1 header + 2 data), got %d", len(lines))
	}

	// Verify header
	header := lines[0]
	if !strings.Contains(header, "\t") {
		t.Error("header should be tab-separated")
	}

	if !strings.Contains(header, "UUID") {
		t.Error("header should contain 'UUID'")
	}

	if !strings.Contains(header, "ToolName") {
		t.Error("header should contain 'ToolName'")
	}

	// Verify data row
	dataRow := lines[1]
	fields := strings.Split(dataRow, "\t")
	if len(fields) < 2 {
		t.Errorf("expected at least 2 fields in data row, got %d", len(fields))
	}

	if fields[0] != "test-uuid-1" {
		t.Errorf("expected first field to be 'test-uuid-1', got '%s'", fields[0])
	}

	if fields[1] != "Bash" {
		t.Errorf("expected second field to be 'Bash', got '%s'", fields[1])
	}
}

func TestFormatOutput_UnsupportedFormat(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "test-uuid", ToolName: "Bash", Status: "success"},
	}

	// Test removed format: json (should fail)
	_, err := FormatOutput(tools, "json")
	if err == nil {
		t.Error("expected error for unsupported format 'json'")
	}

	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Errorf("expected unsupported format error, got: %v", err)
	}

	// Test removed format: md (should fail)
	_, err = FormatOutput(tools, "md")
	if err == nil {
		t.Error("expected error for unsupported format 'md'")
	}

	// Test removed format: csv (should fail)
	_, err = FormatOutput(tools, "csv")
	if err == nil {
		t.Error("expected error for unsupported format 'csv'")
	}
}

func TestFormatOutput_TSV_NonToolCalls(t *testing.T) {
	// TSV now supports all data types via reflection
	nonToolData := struct {
		Name  string
		Value int
	}{
		Name:  "test",
		Value: 42,
	}

	output, err := FormatOutput(nonToolData, "tsv")
	if err != nil {
		t.Errorf("TSV should support non-ToolCall data, got error: %v", err)
	}

	// Verify output is valid TSV (vertical format for single struct)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 { // Name and Value fields
		t.Errorf("expected 2 lines for single struct, got %d", len(lines))
	}

	// Verify tab-separated format
	for i, line := range lines {
		if !strings.Contains(line, "\t") {
			t.Errorf("line %d missing tab separator: %s", i, line)
		}
	}
}

func TestFormatOutput_EmptyArray(t *testing.T) {
	tools := []parser.ToolCall{}

	// JSONL should handle empty arrays (returns empty string)
	output, err := FormatOutput(tools, "jsonl")
	if err != nil {
		t.Fatalf("FormatOutput failed for empty array: %v", err)
	}

	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty string for empty array, got '%s'", output)
	}

	// TSV should handle empty arrays (returns empty string)
	output, err = FormatOutput(tools, "tsv")
	if err != nil {
		t.Fatalf("FormatOutput failed for empty array: %v", err)
	}

	if output != "" {
		t.Errorf("expected empty string for empty TSV, got '%s'", output)
	}
}
