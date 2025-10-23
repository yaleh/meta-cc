package query

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestBuildSessionStats(t *testing.T) {
	entries := []parser.SessionEntry{
		{Type: "user", Timestamp: "2025-10-02T10:00:00Z"},
		{Type: "assistant", Timestamp: "2025-10-02T10:01:00Z"},
	}
	toolCalls := []parser.ToolCall{{ToolName: "Bash"}}

	stats := BuildSessionStats(entries, toolCalls)
	if stats.TurnCount != 2 {
		t.Fatalf("expected 2 turns, got %d", stats.TurnCount)
	}
	if stats.ToolCallCount != 1 {
		t.Fatalf("expected 1 tool call, got %d", stats.ToolCallCount)
	}
}

func TestAnalyzeTimeSeries(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{ToolName: "Bash", Status: "success", Timestamp: "2025-10-02T10:00:00Z"},
		{ToolName: "Edit", Status: "error", Timestamp: "2025-10-02T10:05:00Z"},
	}

	points, err := AnalyzeTimeSeries(toolCalls, "tool-calls", "hour", "tool='Bash'")
	if err != nil {
		t.Fatalf("AnalyzeTimeSeries failed: %v", err)
	}
	if len(points) == 0 {
		t.Fatalf("expected at least one time series point")
	}
}
