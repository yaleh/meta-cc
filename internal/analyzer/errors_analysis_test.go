package analyzer

import (
	"encoding/json"
	"strings"
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

	// Same error label ("connection_error") across different tools merges into 1 group
	if len(result.ByType) != 1 {
		t.Errorf("Expected 1 error type group (same label), got %d", len(result.ByType))
	}
	if result.ByType[0].Count != 2 {
		t.Errorf("Expected count 2 for merged group, got %d", result.ByType[0].Count)
	}
	if result.ByType[0].Label != "connection_error" {
		t.Errorf("Expected label 'connection_error', got %q", result.ByType[0].Label)
	}
	if result.ByType[0].Signature == "" {
		t.Error("Expected Signature field to be populated for backward compatibility")
	}
}

func TestAnalyzeErrors_DifferentLabelsStaySeparate(t *testing.T) {
	toolCalls := []types.ToolCall{
		{UUID: "1", ToolName: "Bash", Status: "error", Error: "command not found: foo"},
		{UUID: "2", ToolName: "Bash", Status: "error", Error: "permission denied"},
	}

	result, err := AnalyzeErrors([]types.SessionEntry{}, toolCalls, 10)
	if err != nil {
		t.Fatalf("AnalyzeErrors returned error: %v", err)
	}

	if len(result.ByType) != 2 {
		t.Errorf("Expected 2 error type groups for different labels, got %d", len(result.ByType))
	}

	labels := make(map[string]int)
	for _, g := range result.ByType {
		if g.Label == "" {
			t.Error("Expected Label field to be populated in every group")
		}
		labels[g.Label] = g.Count
	}
	if labels["command_not_found"] != 1 {
		t.Errorf("Expected count 1 for command_not_found, got %d", labels["command_not_found"])
	}
	if labels["permission_denied"] != 1 {
		t.Errorf("Expected count 1 for permission_denied, got %d", labels["permission_denied"])
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

// TestAnalyzeErrorsStats_OmitsExamples reproduces (at the analyzer level) the
// DIR-042 finding: a single very long error string, repeated many times,
// must not appear anywhere in the stats_only-backing result. This mirrors
// TestGetTimelineStats_Basic as the template.
func TestAnalyzeErrorsStats_OmitsExamples(t *testing.T) {
	longErr := strings.Repeat("x", 4000) // stand-in for the 3,988-char example from the live failure
	var toolCalls []types.ToolCall
	for i := 0; i < 21; i++ {
		toolCalls = append(toolCalls, types.ToolCall{UUID: "u", ToolName: "Bash", Status: "error", Error: longErr})
	}
	toolCalls = append(toolCalls, types.ToolCall{UUID: "u2", ToolName: "Read", Status: "error", Error: "not found"})

	stats, err := AnalyzeErrorsStats([]types.SessionEntry{}, toolCalls)
	if err != nil {
		t.Fatalf("AnalyzeErrorsStats returned error: %v", err)
	}
	if stats.TotalErrors != 22 {
		t.Errorf("Expected TotalErrors=22, got %d", stats.TotalErrors)
	}
	if len(stats.ByTool) != 2 {
		t.Fatalf("Expected 2 tool groups, got %d", len(stats.ByTool))
	}

	toolCounts := make(map[string]int)
	for _, g := range stats.ByTool {
		toolCounts[g.ToolName] = g.Count
	}
	if toolCounts["Bash"] != 21 {
		t.Errorf("Expected Bash count 21, got %d", toolCounts["Bash"])
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("failed to marshal stats: %v", err)
	}
	out := string(data)
	if strings.Contains(out, "examples") {
		t.Errorf("expected no 'examples' field in ErrorAnalysisStats JSON, got: %s", out)
	}
	if strings.Contains(out, longErr) {
		t.Errorf("expected no full-text error content in ErrorAnalysisStats JSON")
	}
	if len(out) > 2000 {
		t.Errorf("expected a small bounded stats JSON, got %d bytes", len(out))
	}
}

func TestAnalyzeErrorsStats_DataSource(t *testing.T) {
	stats, err := AnalyzeErrorsStats([]types.SessionEntry{}, []types.ToolCall{})
	if err != nil {
		t.Fatalf("AnalyzeErrorsStats returned error: %v", err)
	}
	if stats.DataSource != DataSourceMeasured {
		t.Errorf("Expected DataSource=%q, got %q", DataSourceMeasured, stats.DataSource)
	}
	if stats.TotalErrors != 0 {
		t.Errorf("Expected TotalErrors=0, got %d", stats.TotalErrors)
	}
}
