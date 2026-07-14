package analyzer

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

func TestAnalyzeErrors_GroupsByTool(t *testing.T) {
	toolCalls := []types.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "error", Error: "exit 1"},
		{UUID: "2", ToolName: "Bash", Status: "error", Error: "exit 2"},
		{UUID: "3", ToolName: "Bash", Status: "error", Error: "exit 3"},
		{UUID: "4", ToolName: "Read", Status: "error", Error: "not found"},
		{UUID: "5", ToolName: "Read", Status: "error", Error: "permission denied"},
	}

	result, err := AnalyzeErrors([]types.SessionEntry{}, toolCalls, 10)
	if err != nil {
		t.Fatalf("AnalyzeErrors returned error: %v", err)
	}

	if result.TotalErrors != 5 {
		t.Errorf("Expected TotalErrors 5, got %d", result.TotalErrors)
	}

	if len(result.ByTool) != 2 {
		t.Fatalf("Expected 2 tool groups, got %d", len(result.ByTool))
	}

	toolCounts := make(map[string]int)
	for _, g := range result.ByTool {
		toolCounts[g.ToolName] = g.Count
	}

	if toolCounts["Bash"] != 3 {
		t.Errorf("Expected Bash count 3, got %d", toolCounts["Bash"])
	}
	if toolCounts["Read"] != 2 {
		t.Errorf("Expected Read count 2, got %d", toolCounts["Read"])
	}
}

func TestAnalyzeErrors_GroupsByErrorType(t *testing.T) {
	sharedErr := "connection refused"
	toolCalls := []types.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "error", Error: sharedErr},
		{UUID: "2", ToolName: "Read", Status: "error", Error: sharedErr},
	}

	result, err := AnalyzeErrors([]types.SessionEntry{}, toolCalls, 10)
	if err != nil {
		t.Fatalf("AnalyzeErrors returned error: %v", err)
	}

	// Different tools with same error text produce different signatures (tool name is part of signature)
	// So we expect 2 error type groups
	if len(result.ByType) != 2 {
		t.Errorf("Expected 2 error type groups, got %d", len(result.ByType))
	}
}

func TestAnalyzeErrors_SurfacesExamples(t *testing.T) {
	toolCalls := []types.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "error", Error: "same error"},
		{UUID: "2", ToolName: "Bash", Status: "error", Error: "same error"},
		{UUID: "3", ToolName: "Bash", Status: "error", Error: "same error"},
		{UUID: "4", ToolName: "Bash", Status: "error", Error: "same error"},
		{UUID: "5", ToolName: "Bash", Status: "error", Error: "same error"},
	}

	result, err := AnalyzeErrors([]types.SessionEntry{}, toolCalls, 3)
	if err != nil {
		t.Fatalf("AnalyzeErrors returned error: %v", err)
	}

	for _, g := range result.ByTool {
		if len(g.Examples) > 3 {
			t.Errorf("Expected at most 3 examples per group, got %d", len(g.Examples))
		}
	}
	for _, g := range result.ByType {
		if len(g.Examples) > 3 {
			t.Errorf("Expected at most 3 examples per type group, got %d", len(g.Examples))
		}
	}
}

func TestAnalyzeErrors_TimeRange(t *testing.T) {
	entries := []types.SessionEntry{
		makeEntry("uuid-1", "2025-10-02T10:00:00.000Z"),
		makeEntry("uuid-2", "2025-10-02T10:30:00.000Z"),
	}
	toolCalls := []types.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "error", Error: "fail", Timestamp: "2025-10-02T10:00:00.000Z"},
	}

	result, err := AnalyzeErrors(entries, toolCalls, 10)
	if err != nil {
		t.Fatalf("AnalyzeErrors returned error: %v", err)
	}

	if result.TimeRange.Start == "" {
		t.Error("Expected TimeRange.Start to be set")
	}
	if result.TimeRange.End == "" {
		t.Error("Expected TimeRange.End to be set")
	}
	if result.TimeRange.End <= result.TimeRange.Start {
		t.Error("Expected TimeRange.End to be after TimeRange.Start")
	}
}

func TestAnalyzeErrors_DataSource(t *testing.T) {
	result, err := AnalyzeErrors([]types.SessionEntry{}, []types.ToolCall{}, 0)
	if err != nil {
		t.Fatalf("AnalyzeErrors returned error: %v", err)
	}
	if result.DataSource != DataSourceMeasured {
		t.Errorf("Expected DataSource=%q, got %q", DataSourceMeasured, result.DataSource)
	}
}

func TestAnalyzeErrors_EmptySession(t *testing.T) {
	result, err := AnalyzeErrors([]types.SessionEntry{}, []types.ToolCall{}, 10)
	if err != nil {
		t.Fatalf("AnalyzeErrors returned error for empty input: %v", err)
	}
	if result.TotalErrors != 0 {
		t.Errorf("Expected TotalErrors 0, got %d", result.TotalErrors)
	}
	if len(result.ByTool) != 0 {
		t.Errorf("Expected empty ByTool, got %d entries", len(result.ByTool))
	}
	if len(result.ByType) != 0 {
		t.Errorf("Expected empty ByType, got %d entries", len(result.ByType))
	}
}
