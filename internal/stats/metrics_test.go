package stats

import (
	"testing"

	"github.com/yale/meta-cc/internal/parser"
)

func TestCalculateMetric_Count(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Bash"},
		{UUID: "2", ToolName: "Read"},
		{UUID: "3", ToolName: "Edit"},
	}

	result, err := calculateMetric(tools, "count")
	if err != nil {
		t.Fatalf("calculateMetric failed: %v", err)
	}

	if count, ok := result.(int); !ok || count != 3 {
		t.Errorf("Expected count=3, got %v", result)
	}
}

func TestCalculateMetric_ErrorRate(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", Status: "success"},
		{UUID: "2", Status: "error"},
		{UUID: "3", Status: "success"},
		{UUID: "4", Status: "error"},
	}

	result, err := calculateMetric(tools, "error_rate")
	if err != nil {
		t.Fatalf("calculateMetric failed: %v", err)
	}

	if rate, ok := result.(float64); !ok || rate != 0.5 {
		t.Errorf("Expected error_rate=0.5, got %v", result)
	}
}

func TestCalculateMetric_ErrorRateAllSuccess(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", Status: "success"},
		{UUID: "2", Status: "success"},
	}

	result, err := calculateMetric(tools, "error_rate")
	if err != nil {
		t.Fatalf("calculateMetric failed: %v", err)
	}

	if rate, ok := result.(float64); !ok || rate != 0.0 {
		t.Errorf("Expected error_rate=0.0, got %v", result)
	}
}

func TestCalculateMetric_ErrorRateAllErrors(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", Status: "error"},
		{UUID: "2", Status: "error"},
	}

	result, err := calculateMetric(tools, "error_rate")
	if err != nil {
		t.Fatalf("calculateMetric failed: %v", err)
	}

	if rate, ok := result.(float64); !ok || rate != 1.0 {
		t.Errorf("Expected error_rate=1.0, got %v", result)
	}
}

func TestCalculateMetric_InvalidMetric(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Bash"},
	}

	_, err := calculateMetric(tools, "invalid")
	if err == nil {
		t.Error("Expected error for invalid metric")
	}
}

func TestCalculateMetric_EmptyTools(t *testing.T) {
	result, err := calculateMetric([]parser.ToolCall{}, "count")
	if err != nil {
		t.Fatalf("calculateMetric failed: %v", err)
	}

	if count, ok := result.(int); !ok || count != 0 {
		t.Errorf("Expected count=0 for empty tools, got %v", result)
	}
}

func TestGetFieldValue_Tool(t *testing.T) {
	tool := parser.ToolCall{ToolName: "Bash", Status: "success"}

	value := getFieldValue(tool, "tool")
	if value != "Bash" {
		t.Errorf("Expected 'Bash', got %s", value)
	}
}

func TestGetFieldValue_Status(t *testing.T) {
	tool := parser.ToolCall{ToolName: "Bash", Status: "error"}

	value := getFieldValue(tool, "status")
	if value != "error" {
		t.Errorf("Expected 'error', got %s", value)
	}
}

func TestGetFieldValue_UUID(t *testing.T) {
	tool := parser.ToolCall{UUID: "abc-123", ToolName: "Read"}

	value := getFieldValue(tool, "uuid")
	if value != "abc-123" {
		t.Errorf("Expected 'abc-123', got %s", value)
	}
}

func TestGetFieldValue_UnknownField(t *testing.T) {
	tool := parser.ToolCall{ToolName: "Bash"}

	value := getFieldValue(tool, "unknown")
	if value != "" {
		t.Errorf("Expected empty string for unknown field, got %s", value)
	}
}
