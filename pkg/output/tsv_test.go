package output

import (
	"encoding/json"
	"strings"
	"testing"

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

	tsv := FormatTSV(tools)

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
	tsv := FormatTSV(tools)

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

	tsv := FormatTSV(tools)

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
	tsvData := FormatTSV(tools)
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
