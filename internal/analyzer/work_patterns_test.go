package analyzer

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestGetWorkPatterns_ToolFrequency(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{ToolName: "Bash"},
		{ToolName: "Read"},
		{ToolName: "Bash"},
		{ToolName: "Bash"},
		{ToolName: "Read"},
	}

	result, err := GetWorkPatterns([]parser.SessionEntry{}, toolCalls)
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
	entries := []parser.SessionEntry{
		{Timestamp: "2025-10-02T10:00:00Z"},
		{Timestamp: "2025-10-02T10:15:00Z"},
		{Timestamp: "2025-10-02T10:30:00Z"},
		{Timestamp: "2025-10-02T14:00:00Z"},
		{Timestamp: "2025-10-02T14:45:00Z"},
	}

	result, err := GetWorkPatterns(entries, []parser.ToolCall{})
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
	toolCalls := []parser.ToolCall{
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "file_a.go"}, Timestamp: "2025-10-02T10:00:00Z"},
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "file_b.go"}, Timestamp: "2025-10-02T10:01:00Z"},
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "file_a.go"}, Timestamp: "2025-10-02T10:02:00Z"},
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "file_b.go"}, Timestamp: "2025-10-02T10:03:00Z"},
	}

	result, err := GetWorkPatterns([]parser.SessionEntry{}, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ContextSwitches <= 0 {
		t.Errorf("expected ContextSwitches > 0, got %d", result.ContextSwitches)
	}
}

func TestGetWorkPatterns_DataSource(t *testing.T) {
	result, err := GetWorkPatterns([]parser.SessionEntry{}, []parser.ToolCall{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DataSource != DataSourceMeasured {
		t.Errorf("Expected DataSource=%q, got %q", DataSourceMeasured, result.DataSource)
	}
}

func TestGetWorkPatterns_EmptySession(t *testing.T) {
	result, err := GetWorkPatterns([]parser.SessionEntry{}, []parser.ToolCall{})
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
