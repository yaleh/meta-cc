package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/query"
)

// TestQueryAssistantMessagesSchemaFieldNames verifies that the query_assistant_messages schema
// uses correct field names matching the AssistantMessage structure
func TestQueryAssistantMessagesSchemaFieldNames(t *testing.T) {
	// Get the tool definition
	tools := getToolDefinitions()
	var queryAssistantMessagesTool *Tool
	for _, tool := range tools {
		if tool.Name == "query_assistant_messages" {
			queryAssistantMessagesTool = &tool
			break
		}
	}

	if queryAssistantMessagesTool == nil {
		t.Fatal("query_assistant_messages tool not found")
	}

	// Get jq_filter parameter description
	jqFilterParam, ok := queryAssistantMessagesTool.InputSchema.Properties["jq_filter"]
	if !ok {
		t.Fatal("jq_filter parameter not found in query_assistant_messages")
	}

	desc := jqFilterParam.Description

	// Create a sample AssistantMessage to verify field names
	sample := query.AssistantMessage{
		TurnSequence:  1,
		UUID:          "test-uuid",
		Timestamp:     "2025-01-01T00:00:00Z",
		Model:         "claude-3",
		ContentBlocks: []query.AssistantContentBlock{},
		TextLength:    100,
		ToolUseCount:  5,
		TokensInput:   50,
		TokensOutput:  150,
	}

	jsonBytes, err := json.Marshal(sample)
	if err != nil {
		t.Fatalf("Failed to marshal AssistantMessage: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify correct field names are in actual JSON
	correctFields := map[string]bool{
		"turn_sequence":  true,
		"tool_use_count": true,
		"tokens_output":  true,
		"tokens_input":   true,
		"text_length":    true,
	}

	for field := range correctFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Expected field %q not found in AssistantMessage JSON", field)
		}
	}

	// Check if schema uses correct field names
	if !strings.Contains(desc, "turn_sequence") {
		t.Error("Schema should use 'turn_sequence' not 'turn'")
	}

	if !strings.Contains(desc, "tool_use_count") {
		t.Error("Schema should use 'tool_use_count' not 'tool_count'")
	}

	// Check that schema doesn't use incorrect short names
	if strings.Contains(desc, `"turn":`) && !strings.Contains(desc, `"turn_sequence":`) {
		t.Error("Schema incorrectly uses 'turn' instead of 'turn_sequence'")
	}

	if strings.Contains(desc, `"tool_count":`) && !strings.Contains(desc, `"tool_use_count":`) {
		t.Error("Schema incorrectly uses 'tool_count' instead of 'tool_use_count'")
	}
}

// TestQueryAssistantMessagesSchemaCompleteness verifies all fields are documented
func TestQueryAssistantMessagesSchemaCompleteness(t *testing.T) {
	tools := getToolDefinitions()
	var queryAssistantMessagesTool *Tool
	for _, tool := range tools {
		if tool.Name == "query_assistant_messages" {
			queryAssistantMessagesTool = &tool
			break
		}
	}

	if queryAssistantMessagesTool == nil {
		t.Fatal("query_assistant_messages tool not found")
	}

	jqFilterParam, ok := queryAssistantMessagesTool.InputSchema.Properties["jq_filter"]
	if !ok {
		t.Fatal("jq_filter parameter not found")
	}

	desc := jqFilterParam.Description

	// Key fields that should be documented
	expectedFields := []string{
		"turn_sequence",
		"timestamp",
		"tool_use_count",
		"tokens_output",
		"text_length",
		"tokens_input",
	}

	missingFields := []string{}
	for _, field := range expectedFields {
		if !strings.Contains(desc, field) {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		t.Errorf("Schema missing documentation for fields: %v", missingFields)
	}

	t.Logf("query_assistant_messages jq_filter description:\n%s", desc)
}

// TestQueryAssistantMessagesExampleUsesCorrectFields verifies example uses correct field names
func TestQueryAssistantMessagesExampleUsesCorrectFields(t *testing.T) {
	tools := getToolDefinitions()
	var queryAssistantMessagesTool *Tool
	for _, tool := range tools {
		if tool.Name == "query_assistant_messages" {
			queryAssistantMessagesTool = &tool
			break
		}
	}

	if queryAssistantMessagesTool == nil {
		t.Fatal("query_assistant_messages tool not found")
	}

	jqFilterParam, ok := queryAssistantMessagesTool.InputSchema.Properties["jq_filter"]
	if !ok {
		t.Fatal("jq_filter parameter not found")
	}

	desc := jqFilterParam.Description

	// Example should use tool_use_count, not tool_count
	if strings.Contains(desc, ".tool_count") {
		t.Error("Example should use .tool_use_count instead of .tool_count")
	}

	// If example filters by turn, it should use turn_sequence
	if strings.Contains(desc, "select(.turn ") {
		t.Error("Example should use .turn_sequence instead of .turn")
	}
}
