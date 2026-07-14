package analyzer

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

func makeToolCallWithOutput(toolName, filePath, output, status string) types.ToolCall {
	input := map[string]interface{}{}
	if filePath != "" {
		input["file_path"] = filePath
	}
	return types.ToolCall{
		UUID:      "test-uuid",
		ToolName:  toolName,
		Input:     input,
		Output:    output,
		Status:    status,
		Timestamp: "2025-10-02T10:00:00.000Z",
	}
}

func TestGetTechDebt_DetectsMarkers(t *testing.T) {
	toolCalls := []types.ToolCall{
		makeToolCallWithOutput("Read", "main.go", "func foo() {\n// TODO: fix this\n}", "success"),
	}
	result, err := GetTechDebt(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Find TODO marker count
	todoCount := 0
	for _, m := range result.Markers {
		if m.Label == "TODO" {
			todoCount = m.Count
		}
	}
	if todoCount <= 0 {
		t.Errorf("expected TODO count > 0, got %d", todoCount)
	}
	// main.go should appear in hotspot files
	found := false
	for _, f := range result.HotspotFiles {
		if f.File == "main.go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected main.go in HotspotFiles, got %v", result.HotspotFiles)
	}
}

func TestGetTechDebt_CountsPerFile(t *testing.T) {
	toolCalls := []types.ToolCall{
		makeToolCallWithOutput("Read", "a.go", "// TODO: first\n// TODO: second\n", "success"),
		makeToolCallWithOutput("Edit", "b.go", "// TODO: only one\n", "success"),
	}
	result, err := GetTechDebt(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.HotspotFiles) < 2 {
		t.Fatalf("expected at least 2 hotspot files, got %d", len(result.HotspotFiles))
	}
	if result.HotspotFiles[0].File != "a.go" {
		t.Errorf("expected a.go first (2 markers), got %s", result.HotspotFiles[0].File)
	}
	if result.HotspotFiles[0].MarkerCount < result.HotspotFiles[1].MarkerCount {
		t.Errorf("expected descending order: %v", result.HotspotFiles)
	}
}

func TestGetTechDebt_DetectsOpenIssues(t *testing.T) {
	toolCalls := []types.ToolCall{
		makeToolCallWithOutput("Bash", "", "error: build failed", "error"),
		makeToolCallWithOutput("Read", "", "more errors", "error"),
		makeToolCallWithOutput("Edit", "x.go", "stuff", "error"),
	}
	result, err := GetTechDebt(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OpenIssues <= 0 {
		t.Errorf("expected OpenIssues > 0, got %d", result.OpenIssues)
	}
}

func TestGetTechDebt_DataSource(t *testing.T) {
	result, err := GetTechDebt(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DataSource != DataSourceMeasured {
		t.Errorf("Expected DataSource=%q, got %q", DataSourceMeasured, result.DataSource)
	}
}

func TestGetTechDebt_EmptySession(t *testing.T) {
	result, err := GetTechDebt(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.OpenIssues != 0 {
		t.Errorf("expected 0 OpenIssues, got %d", result.OpenIssues)
	}
	if len(result.Markers) != 0 {
		t.Errorf("expected empty markers, got %v", result.Markers)
	}
	if len(result.HotspotFiles) != 0 {
		t.Errorf("expected empty hotspot files, got %v", result.HotspotFiles)
	}
}
