package analyzer

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

func TestGetWorkPatterns_ToolFrequency(t *testing.T) {
	toolCalls := []types.ToolCall{
		{ToolName: "Bash"},
		{ToolName: "Read"},
		{ToolName: "Bash"},
		{ToolName: "Bash"},
		{ToolName: "Read"},
	}

	result, err := GetWorkPatterns([]types.SessionEntry{}, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ToolFrequency) < 2 {
		t.Fatalf("expected at least 2 tools, got %d", len(result.ToolFrequency))
	}

	if result.ToolFrequency[0].ToolName != "Bash" || result.ToolFrequency[0].Count != 3 {
		t.Errorf("expected Bash count=3, got %s count=%d", result.ToolFrequency[0].ToolName, result.ToolFrequency[0].Count)
	}

	if result.ToolFrequency[1].ToolName != "Read" || result.ToolFrequency[1].Count != 2 {
		t.Errorf("expected Read count=2, got %s count=%d", result.ToolFrequency[1].ToolName, result.ToolFrequency[1].Count)
	}
}

func TestGetWorkPatterns_HourlyActivity(t *testing.T) {
	entries := []types.SessionEntry{
		{Timestamp: "2025-10-02T10:00:00Z"},
		{Timestamp: "2025-10-02T10:15:00Z"},
		{Timestamp: "2025-10-02T10:30:00Z"},
		{Timestamp: "2025-10-02T14:00:00Z"},
		{Timestamp: "2025-10-02T14:45:00Z"},
	}

	result, err := GetWorkPatterns(entries, []types.ToolCall{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.HourlyActivity) != 24 {
		t.Errorf("expected 24 elements in HourlyActivity, got %d", len(result.HourlyActivity))
	}

	if result.HourlyActivity[10] != 3 {
		t.Errorf("expected HourlyActivity[10]==3, got %d", result.HourlyActivity[10])
	}

	if result.HourlyActivity[14] != 2 {
		t.Errorf("expected HourlyActivity[14]==2, got %d", result.HourlyActivity[14])
	}
}

func TestGetWorkPatterns_ContextSwitches(t *testing.T) {
	// Alternating between two files within 5 minutes
	toolCalls := []types.ToolCall{
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "file_a.go"}, Timestamp: "2025-10-02T10:00:00Z"},
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "file_b.go"}, Timestamp: "2025-10-02T10:01:00Z"},
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "file_a.go"}, Timestamp: "2025-10-02T10:02:00Z"},
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "file_b.go"}, Timestamp: "2025-10-02T10:03:00Z"},
	}

	result, err := GetWorkPatterns([]types.SessionEntry{}, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ContextSwitches <= 0 {
		t.Errorf("expected ContextSwitches > 0, got %d", result.ContextSwitches)
	}
}

func TestGetWorkPatterns_DataSource(t *testing.T) {
	result, err := GetWorkPatterns([]types.SessionEntry{}, []types.ToolCall{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DataSource != DataSourceMeasured {
		t.Errorf("Expected DataSource=%q, got %q", DataSourceMeasured, result.DataSource)
	}
}

func TestGetWorkPatterns_EmptySession(t *testing.T) {
	result, err := GetWorkPatterns([]types.SessionEntry{}, []types.ToolCall{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.HourlyActivity) != 24 {
		t.Errorf("expected 24 elements in HourlyActivity, got %d", len(result.HourlyActivity))
	}

	for i, v := range result.HourlyActivity {
		if v != 0 {
			t.Errorf("expected HourlyActivity[%d]==0, got %d", i, v)
		}
	}

	if result.ContextSwitches != 0 {
		t.Errorf("expected ContextSwitches==0, got %d", result.ContextSwitches)
	}
}

// TestGetWorkPatternsStatsOnly_MatchesGetWorkPatterns verifies the
// stats_only backing function returns the same aggregate-only shape as
// GetWorkPatterns itself (GetWorkPatterns never carries per-item example
// text, so there is nothing further to omit) -- DIR-042.
func TestGetWorkPatternsStatsOnly_MatchesGetWorkPatterns(t *testing.T) {
	toolCalls := []types.ToolCall{
		{ToolName: "Bash"},
		{ToolName: "Bash"},
		{ToolName: "Read"},
	}

	full, err := GetWorkPatterns(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats, err := GetWorkPatternsStatsOnly(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stats.ToolFrequency) != len(full.ToolFrequency) {
		t.Fatalf("expected matching ToolFrequency lengths, got %d vs %d", len(stats.ToolFrequency), len(full.ToolFrequency))
	}
	for i := range full.ToolFrequency {
		if stats.ToolFrequency[i] != full.ToolFrequency[i] {
			t.Errorf("expected ToolFrequency[%d]=%v, got %v", i, full.ToolFrequency[i], stats.ToolFrequency[i])
		}
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("failed to marshal stats: %v", err)
	}
	if strings.Contains(string(data), "examples") {
		t.Errorf("expected no 'examples' field in WorkPatternsStats JSON, got: %s", data)
	}
}
