package output

import (
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestGenerateSummary(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Read", Status: "success", Error: ""},
		{UUID: "2", ToolName: "Read", Status: "success", Error: ""},
		{UUID: "3", ToolName: "Bash", Status: "success", Error: ""},
		{UUID: "4", ToolName: "Edit", Status: "error", Error: "file not found"},
		{UUID: "5", ToolName: "Read", Status: "success", Error: ""},
	}

	summary := GenerateSummary(tools)

	// Verify summary contains key statistics
	if !strings.Contains(summary, "Total Tools:") {
		t.Error("Summary missing 'Total Tools' line")
	}

	if !strings.Contains(summary, "5") {
		t.Error("Summary should contain total count of 5")
	}

	if !strings.Contains(summary, "Errors:") {
		t.Error("Summary missing 'Errors' line")
	}

	if !strings.Contains(summary, "1") {
		t.Error("Summary should contain error count of 1")
	}

	// Verify tool distribution
	if !strings.Contains(summary, "Top Tools:") {
		t.Error("Summary missing 'Top Tools' section")
	}

	if !strings.Contains(summary, "Read") {
		t.Error("Summary should mention 'Read' tool")
	}
}

func TestGenerateSummaryEmpty(t *testing.T) {
	tools := []parser.ToolCall{}
	summary := GenerateSummary(tools)

	if !strings.Contains(summary, "Total Tools: 0") {
		t.Error("Summary should handle empty tool list")
	}
}

func TestFormatSummaryFirst(t *testing.T) {
	tools := make([]parser.ToolCall, 100)
	for i := 0; i < 100; i++ {
		status := "success"
		errMsg := ""
		if i%10 == 0 {
			status = "error"
			errMsg = "test error"
		}
		tools[i] = parser.ToolCall{
			UUID:     "uuid-" + string(rune('0'+i%10)),
			ToolName: "Read",
			Status:   status,
			Error:    errMsg,
		}
	}

	output, err := FormatSummaryFirst(tools, 10, "tsv")
	if err != nil {
		t.Fatalf("FormatSummaryFirst failed: %v", err)
	}

	// Verify summary exists
	if !strings.Contains(output.Summary, "Session Summary") {
		t.Error("Output missing summary section")
	}

	if !strings.Contains(output.Summary, "Total Tools: 100") {
		t.Error("Summary should contain total count")
	}

	// Verify details contains only top 10
	detailLines := strings.Split(strings.TrimSpace(output.Details), "\n")
	// Should be header + 10 records = 11 lines
	if len(detailLines) != 11 {
		t.Errorf("expected 11 detail lines (header + 10 records), got %d", len(detailLines))
	}

	// Verify TSV format
	if !strings.Contains(detailLines[0], "\t") {
		t.Error("Details should be in TSV format")
	}
}

func TestFormatSummaryFirstJSON(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Read", Status: "success", Error: ""},
		{UUID: "2", ToolName: "Bash", Status: "success", Error: ""},
		{UUID: "3", ToolName: "Edit", Status: "error", Error: "error"},
	}

	output, err := FormatSummaryFirst(tools, 2, "jsonl")
	if err != nil {
		t.Fatalf("FormatSummaryFirst failed: %v", err)
	}

	// Verify summary
	if !strings.Contains(output.Summary, "Total Tools: 3") {
		t.Error("Summary incorrect")
	}

	// Verify JSON details
	if !strings.HasPrefix(output.Details, "[") {
		t.Error("Details should be JSON array")
	}

	// Verify only 2 records in details
	if strings.Count(output.Details, "\"UUID\"") != 2 {
		t.Error("Details should contain only 2 records")
	}
}

func TestFormatSummaryFirstAllRecords(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Read", Status: "success", Error: ""},
		{UUID: "2", ToolName: "Bash", Status: "success", Error: ""},
	}

	// topN = 0 means all records
	output, err := FormatSummaryFirst(tools, 0, "tsv")
	if err != nil {
		t.Fatalf("FormatSummaryFirst failed: %v", err)
	}

	// Verify all records included
	detailLines := strings.Split(strings.TrimSpace(output.Details), "\n")
	if len(detailLines) != 3 { // header + 2 records
		t.Errorf("expected 3 lines, got %d", len(detailLines))
	}
}

func TestFormatSummaryFirstTSV(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Read", Status: "success", Error: ""},
		{UUID: "2", ToolName: "Bash", Status: "success", Error: ""},
	}

	output, err := FormatSummaryFirst(tools, 5, "tsv")
	if err != nil {
		t.Fatalf("FormatSummaryFirst failed: %v", err)
	}

	// Verify TSV format (tab-separated)
	if !strings.Contains(output.Details, "\t") {
		t.Error("TSV details should contain tab separators")
	}
}

func TestFormatSummaryFirstInvalidFormat(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Read", Status: "success", Error: ""},
	}

	_, err := FormatSummaryFirst(tools, 10, "invalid")
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}
