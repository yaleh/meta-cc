package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

// TestQueryToolSequencesSchemaAccuracy verifies that the query_tool_sequences schema
// documentation matches the actual SequencePattern structure
func TestQueryToolSequencesSchemaAccuracy(t *testing.T) {
	// Get the tool definition
	tools := getToolDefinitions()
	var queryToolSequencesTool *Tool
	for _, tool := range tools {
		if tool.Name == "query_tool_sequences" {
			queryToolSequencesTool = &tool
			break
		}
	}

	if queryToolSequencesTool == nil {
		t.Fatal("query_tool_sequences tool not found")
	}

	// Get jq_filter parameter description
	jqFilterParam, ok := queryToolSequencesTool.InputSchema.Properties["jq_filter"]
	if !ok {
		t.Fatal("jq_filter parameter not found in query_tool_sequences")
	}

	desc := jqFilterParam.Description

	// Create a sample SequencePattern to verify field names
	sample := types.SequencePattern{
		Pattern: "Read → Edit → Write",
		Count:   5,
		Occurrences: []types.SequenceOccurrence{
			{StartTurn: 1, EndTurn: 3},
		},
		TimeSpanMin: 10,
	}

	jsonBytes, err := json.Marshal(sample)
	if err != nil {
		t.Fatalf("Failed to marshal SequencePattern: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify schema mentions "count" (not "occurrences" as a number)
	if !strings.Contains(desc, "count") {
		t.Error("Schema should mention 'count' field for the number of occurrences")
	}

	// Verify schema describes "occurrences" as an array (not a number)
	if strings.Contains(desc, `"occurrences": "number`) {
		t.Error("Schema incorrectly describes 'occurrences' as a number - it should be an array")
	}

	// The actual JSON should have these fields
	expectedFields := []string{"pattern", "count", "occurrences", "time_span_minutes"}
	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Expected field %q not found in marshaled JSON", field)
		}
	}

	// Verify that count is indeed a number and occurrences is an array
	if count, ok := result["count"].(float64); !ok || count != 5 {
		t.Errorf("count should be a number with value 5, got %v", result["count"])
	}

	if occurrences, ok := result["occurrences"].([]interface{}); !ok || len(occurrences) != 1 {
		t.Errorf("occurrences should be an array with 1 element, got %v", result["occurrences"])
	}
}

// TestQueryToolSequencesExampleUseCount verifies that the example jq filter
// uses .count instead of .occurrences for numeric comparison
func TestQueryToolSequencesExampleUseCount(t *testing.T) {
	tools := getToolDefinitions()
	var queryToolSequencesTool *Tool
	for _, tool := range tools {
		if tool.Name == "query_tool_sequences" {
			queryToolSequencesTool = &tool
			break
		}
	}

	if queryToolSequencesTool == nil {
		t.Fatal("query_tool_sequences tool not found")
	}

	jqFilterParam, ok := queryToolSequencesTool.InputSchema.Properties["jq_filter"]
	if !ok {
		t.Fatal("jq_filter parameter not found")
	}

	desc := jqFilterParam.Description

	// The example should use .count for numeric comparison, not .occurrences
	// Because occurrences is an array, not a number
	if strings.Contains(desc, ".occurrences >") || strings.Contains(desc, "select(.occurrences >") {
		t.Error("Example jq filter should use .count instead of .occurrences for numeric comparison")
	}

	// The example should ideally contain .count
	if !strings.Contains(desc, ".count") {
		t.Log("Warning: Example should demonstrate using .count field")
	}
}
