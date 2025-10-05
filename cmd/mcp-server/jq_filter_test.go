package main

import (
	"strings"
	"testing"
)

func TestApplyJQFilter_Simple(t *testing.T) {
	jsonlData := `{"tool":"Bash","status":"success"}
{"tool":"Read","status":"error"}
{"tool":"Edit","status":"success"}`

	jqExpr := `.[] | select(.status == "error")`

	result, err := ApplyJQFilter(jsonlData, jqExpr)
	if err != nil {
		t.Fatalf("ApplyJQFilter failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 result, got %d", len(lines))
	}

	if !strings.Contains(result, "Read") {
		t.Error("expected Read in result")
	}
}

func TestApplyJQFilter_Projection(t *testing.T) {
	jsonlData := `{"tool":"Bash","status":"success","duration":100}
{"tool":"Read","status":"error","duration":50}`

	jqExpr := `.[] | {tool: .tool, status: .status}`

	result, err := ApplyJQFilter(jsonlData, jqExpr)
	if err != nil {
		t.Fatalf("ApplyJQFilter failed: %v", err)
	}

	// Verify projection (no duration field)
	if strings.Contains(result, "duration") {
		t.Error("expected duration to be excluded")
	}
}

func TestApplyJQFilter_DefaultExpression(t *testing.T) {
	jsonlData := `{"tool":"Bash","status":"success"}
{"tool":"Read","status":"error"}`

	// Empty jq expression should default to ".[]"
	result, err := ApplyJQFilter(jsonlData, "")
	if err != nil {
		t.Fatalf("ApplyJQFilter failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 results, got %d", len(lines))
	}
}

func TestApplyJQFilter_InvalidExpression(t *testing.T) {
	jsonlData := `{"tool":"Bash","status":"success"}`

	// Invalid jq expression
	_, err := ApplyJQFilter(jsonlData, ".[ invalid syntax")
	if err == nil {
		t.Error("expected error for invalid jq expression")
	}
}

func TestApplyJQFilter_EmptyData(t *testing.T) {
	result, err := ApplyJQFilter("", ".[]")
	if err != nil {
		t.Fatalf("ApplyJQFilter failed: %v", err)
	}

	if strings.TrimSpace(result) != "" {
		t.Error("expected empty result for empty data")
	}
}

func TestGenerateStats(t *testing.T) {
	jsonlData := `{"tool":"Bash","status":"error"}
{"tool":"Bash","status":"error"}
{"tool":"Read","status":"error"}`

	stats, err := GenerateStats(jsonlData)
	if err != nil {
		t.Fatalf("GenerateStats failed: %v", err)
	}

	// Verify stats format
	if !strings.Contains(stats, "Bash") {
		t.Error("expected Bash in stats")
	}
	if !strings.Contains(stats, "count") {
		t.Error("expected count field")
	}

	// Verify count is correct (Bash should appear twice)
	lines := strings.Split(strings.TrimSpace(stats), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 stat entries, got %d", len(lines))
	}
}

func TestGenerateStats_AlternativeFieldNames(t *testing.T) {
	// Test with "ToolName" field instead of "tool"
	jsonlData := `{"ToolName":"Bash","Status":"error"}
{"ToolName":"Read","Status":"success"}`

	stats, err := GenerateStats(jsonlData)
	if err != nil {
		t.Fatalf("GenerateStats failed: %v", err)
	}

	if !strings.Contains(stats, "Bash") {
		t.Error("expected Bash in stats")
	}
	if !strings.Contains(stats, "Read") {
		t.Error("expected Read in stats")
	}
}

func TestGenerateStats_EmptyData(t *testing.T) {
	stats, err := GenerateStats("")
	if err != nil {
		t.Fatalf("GenerateStats failed: %v", err)
	}

	if strings.TrimSpace(stats) != "" {
		t.Error("expected empty stats for empty data")
	}
}
