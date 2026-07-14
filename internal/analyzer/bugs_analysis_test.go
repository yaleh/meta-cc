package analyzer

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

func TestAnalyzeBugs_FixPair(t *testing.T) {
	toolCalls := []types.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "command not found"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "success"},
	}

	result, err := AnalyzeBugs([]types.SessionEntry{}, toolCalls, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.TotalPairs != 1 {
		t.Errorf("Expected TotalPairs=1, got %d", result.TotalPairs)
	}
	if len(result.Patterns) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(result.Patterns))
	}
}

func TestAnalyzeBugs_Recurrence(t *testing.T) {
	// Same tool+error appearing 3 times, each followed by a success
	toolCalls := []types.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "permission denied"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "success"},
		{UUID: "uuid-3", ToolName: "Bash", Status: "error", Error: "permission denied"},
		{UUID: "uuid-4", ToolName: "Bash", Status: "success"},
		{UUID: "uuid-5", ToolName: "Bash", Status: "error", Error: "permission denied"},
		{UUID: "uuid-6", ToolName: "Bash", Status: "success"},
	}

	result, err := AnalyzeBugs([]types.SessionEntry{}, toolCalls, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result.Patterns) == 0 {
		t.Fatal("Expected at least 1 pattern")
	}
	if result.Patterns[0].Recurrences != 3 {
		t.Errorf("Expected Recurrences=3, got %d", result.Patterns[0].Recurrences)
	}
	if result.TotalPairs != 3 {
		t.Errorf("Expected TotalPairs=3, got %d", result.TotalPairs)
	}
}

func TestAnalyzeBugs_SortedByRecurrence(t *testing.T) {
	// Create patterns: one error with 1 occurrence, one with 3, one with 2
	toolCalls := []types.ToolCall{
		// error-A appears once
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "error alpha unique"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "success"},
		// error-B appears 3 times
		{UUID: "uuid-3", ToolName: "Read", Status: "error", Error: "error beta repeated"},
		{UUID: "uuid-4", ToolName: "Read", Status: "success"},
		{UUID: "uuid-5", ToolName: "Read", Status: "error", Error: "error beta repeated"},
		{UUID: "uuid-6", ToolName: "Read", Status: "success"},
		{UUID: "uuid-7", ToolName: "Read", Status: "error", Error: "error beta repeated"},
		{UUID: "uuid-8", ToolName: "Read", Status: "success"},
		// error-C appears twice
		{UUID: "uuid-9", ToolName: "Grep", Status: "error", Error: "error gamma double"},
		{UUID: "uuid-10", ToolName: "Grep", Status: "success"},
		{UUID: "uuid-11", ToolName: "Grep", Status: "error", Error: "error gamma double"},
		{UUID: "uuid-12", ToolName: "Grep", Status: "success"},
	}

	result, err := AnalyzeBugs([]types.SessionEntry{}, toolCalls, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result.Patterns) != 3 {
		t.Fatalf("Expected 3 patterns, got %d", len(result.Patterns))
	}
	// Verify sorted descending by recurrences: 3, 2, 1
	if result.Patterns[0].Recurrences != 3 {
		t.Errorf("Expected first pattern Recurrences=3, got %d", result.Patterns[0].Recurrences)
	}
	if result.Patterns[1].Recurrences != 2 {
		t.Errorf("Expected second pattern Recurrences=2, got %d", result.Patterns[1].Recurrences)
	}
	if result.Patterns[2].Recurrences != 1 {
		t.Errorf("Expected third pattern Recurrences=1, got %d", result.Patterns[2].Recurrences)
	}
}

func TestAnalyzeBugs_EmptySession(t *testing.T) {
	result, err := AnalyzeBugs([]types.SessionEntry{}, []types.ToolCall{}, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result.Patterns) != 0 {
		t.Errorf("Expected 0 patterns, got %d", len(result.Patterns))
	}
	if result.TotalPairs != 0 {
		t.Errorf("Expected TotalPairs=0, got %d", result.TotalPairs)
	}
}

func TestAnalyzeBugs_DataSource(t *testing.T) {
	result, err := AnalyzeBugs([]types.SessionEntry{}, []types.ToolCall{}, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.DataSource != DataSourceMeasured {
		t.Errorf("Expected DataSource=%q, got %q", DataSourceMeasured, result.DataSource)
	}
}
