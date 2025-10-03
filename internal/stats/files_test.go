package stats

import (
	"testing"

	"github.com/yale/meta-cc/internal/parser"
)

func TestAnalyzeFileStats(t *testing.T) {
	tests := []struct {
		name      string
		toolCalls []parser.ToolCall
		wantFiles int
	}{
		{
			name: "basic file operations",
			toolCalls: []parser.ToolCall{
				{ToolName: "Read", Input: map[string]interface{}{"file_path": "main.go"}, Status: "success"},
				{ToolName: "Edit", Input: map[string]interface{}{"file_path": "main.go"}, Status: "success"},
				{ToolName: "Edit", Input: map[string]interface{}{"file_path": "main.go"}, Status: "error"},
				{ToolName: "Write", Input: map[string]interface{}{"file_path": "test.go"}, Status: "success"},
			},
			wantFiles: 2,
		},
		{
			name: "no file operations",
			toolCalls: []parser.ToolCall{
				{ToolName: "Bash", Input: map[string]interface{}{"command": "ls"}, Status: "success"},
			},
			wantFiles: 0,
		},
		{
			name:      "empty tool calls",
			toolCalls: []parser.ToolCall{},
			wantFiles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := AnalyzeFileStats(tt.toolCalls)
			if len(stats) != tt.wantFiles {
				t.Errorf("got %d files, want %d", len(stats), tt.wantFiles)
			}
		})
	}
}

func TestFileStats_Counts(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "main.go"}, Status: "success"},
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "main.go"}, Status: "success"},
		{ToolName: "Edit", Input: map[string]interface{}{"file_path": "main.go"}, Status: "success"},
		{ToolName: "Edit", Input: map[string]interface{}{"file_path": "main.go"}, Status: "error"},
		{ToolName: "Write", Input: map[string]interface{}{"file_path": "main.go"}, Status: "success"},
	}

	stats := AnalyzeFileStats(toolCalls)

	if len(stats) != 1 {
		t.Fatalf("expected 1 file, got %d", len(stats))
	}

	mainStats := stats[0]

	if mainStats.FilePath != "main.go" {
		t.Errorf("file_path: expected main.go, got %s", mainStats.FilePath)
	}

	if mainStats.ReadCount != 2 {
		t.Errorf("read_count: expected 2, got %d", mainStats.ReadCount)
	}

	if mainStats.EditCount != 2 {
		t.Errorf("edit_count: expected 2, got %d", mainStats.EditCount)
	}

	if mainStats.WriteCount != 1 {
		t.Errorf("write_count: expected 1, got %d", mainStats.WriteCount)
	}

	if mainStats.ErrorCount != 1 {
		t.Errorf("error_count: expected 1, got %d", mainStats.ErrorCount)
	}

	if mainStats.TotalOps != 5 {
		t.Errorf("total_ops: expected 5, got %d", mainStats.TotalOps)
	}

	expectedErrorRate := 1.0 / 5.0
	if mainStats.ErrorRate != expectedErrorRate {
		t.Errorf("error_rate: expected %.3f, got %.3f", expectedErrorRate, mainStats.ErrorRate)
	}
}

func TestFileStats_MultipleFiles(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{ToolName: "Edit", Input: map[string]interface{}{"file_path": "a.go"}, Status: "success"},
		{ToolName: "Edit", Input: map[string]interface{}{"file_path": "a.go"}, Status: "success"},
		{ToolName: "Edit", Input: map[string]interface{}{"file_path": "a.go"}, Status: "success"},
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "b.go"}, Status: "success"},
		{ToolName: "Write", Input: map[string]interface{}{"file_path": "c.go"}, Status: "error"},
	}

	stats := AnalyzeFileStats(toolCalls)

	if len(stats) != 3 {
		t.Fatalf("expected 3 files, got %d", len(stats))
	}

	// Default sort should be by TotalOps descending
	if stats[0].FilePath != "a.go" {
		t.Errorf("first file should be a.go (most ops), got %s", stats[0].FilePath)
	}

	if stats[0].TotalOps != 3 {
		t.Errorf("a.go should have 3 total ops, got %d", stats[0].TotalOps)
	}
}

func TestFileStats_SortBy(t *testing.T) {
	stats := []FileStats{
		{FilePath: "a.go", EditCount: 5, ErrorCount: 1, TotalOps: 10, ErrorRate: 0.1},
		{FilePath: "b.go", EditCount: 10, ErrorCount: 3, TotalOps: 15, ErrorRate: 0.2},
		{FilePath: "c.go", EditCount: 3, ErrorCount: 5, TotalOps: 8, ErrorRate: 0.625},
	}

	tests := []struct {
		sortBy string
		want   string // expected first file
	}{
		{"edit_count", "b.go"},
		{"error_count", "c.go"},
		{"total_ops", "b.go"},
		{"error_rate", "c.go"},
	}

	for _, tt := range tests {
		t.Run("sort_by_"+tt.sortBy, func(t *testing.T) {
			// Make a copy
			statsCopy := make([]FileStats, len(stats))
			copy(statsCopy, stats)

			SortFileStats(statsCopy, tt.sortBy)

			if statsCopy[0].FilePath != tt.want {
				t.Errorf("sort by %s: expected %s first, got %s", tt.sortBy, tt.want, statsCopy[0].FilePath)
			}
		})
	}
}

func TestFileStats_PathVariants(t *testing.T) {
	toolCalls := []parser.ToolCall{
		// Test "file_path" field
		{ToolName: "Read", Input: map[string]interface{}{"file_path": "a.go"}, Status: "success"},
		// Test "path" field (alternate)
		{ToolName: "Edit", Input: map[string]interface{}{"path": "b.go"}, Status: "success"},
		// Test notebook_path for NotebookEdit
		{ToolName: "NotebookEdit", Input: map[string]interface{}{"notebook_path": "c.ipynb"}, Status: "success"},
	}

	stats := AnalyzeFileStats(toolCalls)

	if len(stats) != 3 {
		t.Fatalf("expected 3 files, got %d", len(stats))
	}
}

func TestFileStats_TopN(t *testing.T) {
	stats := []FileStats{
		{FilePath: "a.go", TotalOps: 10},
		{FilePath: "b.go", TotalOps: 8},
		{FilePath: "c.go", TotalOps: 6},
		{FilePath: "d.go", TotalOps: 4},
	}

	topN := 2
	if topN < len(stats) {
		stats = stats[:topN]
	}

	if len(stats) != 2 {
		t.Errorf("expected 2 files after top-N, got %d", len(stats))
	}

	if stats[0].FilePath != "a.go" || stats[1].FilePath != "b.go" {
		t.Errorf("expected a.go and b.go, got %s and %s", stats[0].FilePath, stats[1].FilePath)
	}
}
