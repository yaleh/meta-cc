package stats

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestAggregate_GroupByTool(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "success"},
		{UUID: "2", ToolName: "Read", Status: "success"},
		{UUID: "3", ToolName: "Bash", Status: "error"},
		{UUID: "4", ToolName: "Read", Status: "success"},
		{UUID: "5", ToolName: "Edit", Status: "success"},
	}

	config := AggregateConfig{
		GroupBy: "tool",
		Metrics: []string{"count"},
	}

	results, err := Aggregate(tools, config)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(results))
	}

	// Verify Bash group
	bashFound := false
	for _, r := range results {
		if r.GroupValue == "Bash" {
			bashFound = true
			if count, ok := r.Metrics["count"].(int); !ok || count != 2 {
				t.Errorf("Expected Bash count=2, got %v", r.Metrics["count"])
			}
		}
	}
	if !bashFound {
		t.Error("Bash group not found")
	}
}

func TestAggregate_GroupByStatus(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "success"},
		{UUID: "2", ToolName: "Read", Status: "success"},
		{UUID: "3", ToolName: "Bash", Status: "error"},
	}

	config := AggregateConfig{
		GroupBy: "status",
		Metrics: []string{"count"},
	}

	results, err := Aggregate(tools, config)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(results))
	}

	// Check sorting by count (desc)
	if results[0].GroupValue != "success" {
		t.Errorf("Expected first group to be 'success', got %s", results[0].GroupValue)
	}
}

func TestAggregate_MultipleMetrics(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "success"},
		{UUID: "2", ToolName: "Bash", Status: "error"},
		{UUID: "3", ToolName: "Bash", Status: "success"},
	}

	config := AggregateConfig{
		GroupBy: "tool",
		Metrics: []string{"count", "error_rate"},
	}

	results, err := Aggregate(tools, config)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 group, got %d", len(results))
	}

	result := results[0]
	if count, ok := result.Metrics["count"].(int); !ok || count != 3 {
		t.Errorf("Expected count=3, got %v", result.Metrics["count"])
	}

	if errorRate, ok := result.Metrics["error_rate"].(float64); !ok || errorRate < 0.33 || errorRate > 0.34 {
		t.Errorf("Expected error_rate≈0.33, got %v", result.Metrics["error_rate"])
	}
}

func TestAggregate_EmptyInput(t *testing.T) {
	config := AggregateConfig{
		GroupBy: "tool",
		Metrics: []string{"count"},
	}

	results, err := Aggregate([]parser.ToolCall{}, config)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty input, got %d", len(results))
	}
}

func TestAggregate_InvalidGroupBy(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "success"},
	}

	config := AggregateConfig{
		GroupBy: "invalid_field",
		Metrics: []string{"count"},
	}

	_, err := Aggregate(tools, config)
	if err == nil {
		t.Error("Expected error for invalid group-by field")
	}
}

func TestAggregate_InvalidMetric(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "success"},
	}

	config := AggregateConfig{
		GroupBy: "tool",
		Metrics: []string{"invalid_metric"},
	}

	_, err := Aggregate(tools, config)
	if err == nil {
		t.Error("Expected error for invalid metric")
	}
}
