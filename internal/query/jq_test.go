package query

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

func TestParseJQExpressionQuotedError(t *testing.T) {
	_, err := parseJQExpression(`'.[]'`)
	if err == nil {
		t.Fatal("expected quoted expression to return error")
	}
	if !strings.Contains(err.Error(), "appears to be quoted") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestParseJSONLRecordsInvalidJSON(t *testing.T) {
	_, err := parseJSONLRecords("not-json\n")
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
	if !strings.Contains(err.Error(), "invalid JSON at line 1") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestEncodeJQResultsMarshalError(t *testing.T) {
	result, err := encodeJQResults([]interface{}{make(chan int)})
	if err == nil {
		t.Fatal("expected marshal error for channel value")
	}
	if result != "" {
		t.Fatalf("expected empty result string, got %q", result)
	}
}

// TestApplyJQFilter_QuotedExpressionError verifies improved error message for quoted expressions
func TestApplyJQFilter_QuotedExpressionError(t *testing.T) {
	jsonlData := `{"tool":"Bash","status":"success"}
{"tool":"Read","status":"error"}`

	// Test common mistake: wrapping jq expression in quotes
	testCases := []struct {
		name     string
		badExpr  string
		expected string
	}{
		{
			name:     "single quoted expression",
			badExpr:  `'.[] | {tool: .tool}'`,
			expected: "appears to be quoted",
		},
		{
			name:     "single quoted complex expression",
			badExpr:  `'.[] | {turn: .turn, content: .content[0:100]}'`,
			expected: "appears to be quoted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ApplyJQFilter(jsonlData, tc.badExpr)
			if err == nil {
				t.Errorf("expected error for quoted expression: %s", tc.badExpr)
				return
			}

			// Verify error message contains helpful guidance
			if !strings.Contains(err.Error(), tc.expected) {
				t.Errorf("error message should contain '%s' for expression '%s', got: %v",
					tc.expected, tc.badExpr, err)
			}

			// Verify error message suggests correct syntax
			if !strings.Contains(err.Error(), ".[] | {field: .field}") {
				t.Errorf("error message should suggest correct syntax for expression '%s', got: %v",
					tc.badExpr, err)
			}

			t.Logf("Error for '%s': %v", tc.badExpr, err)
		})
	}
}

// TestApplyJQFilter_GenuineSyntaxStillReportsOriginalError verifies that genuine syntax errors still get appropriate error messages
func TestApplyJQFilter_GenuineSyntaxStillReportsOriginalError(t *testing.T) {
	jsonlData := `{"tool":"Bash","status":"success"}`

	// Test genuine syntax errors (not quote-related)
	testCases := []struct {
		name     string
		badExpr  string
		expected string
	}{
		{
			name:     "invalid bracket syntax",
			badExpr:  `. [ invalid syntax`,
			expected: "invalid jq expression",
		},
		{
			name:     "missing closing brace",
			badExpr:  `.[] | select(.tool == "Bash"`,
			expected: "invalid jq expression",
		},
		{
			name:     "invalid function",
			badExpr:  `.[] | invalid_function()`,
			expected: "invalid jq expression",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ApplyJQFilter(jsonlData, tc.badExpr)
			if err == nil {
				t.Errorf("expected error for invalid expression: %s", tc.badExpr)
				return
			}

			// Verify error message doesn't incorrectly suggest quote issues
			if strings.Contains(err.Error(), "appears to be quoted") {
				t.Errorf("genuine syntax error should not suggest quote issues for expression '%s', got: %v",
					tc.badExpr, err)
			}

			// Should still indicate invalid jq expression
			if !strings.Contains(err.Error(), tc.expected) {
				t.Errorf("error message should contain '%s' for expression '%s', got: %v",
					tc.expected, tc.badExpr, err)
			}

			t.Logf("Error for '%s': %v", tc.badExpr, err)
		})
	}
}
