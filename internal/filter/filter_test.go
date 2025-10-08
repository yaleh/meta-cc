package filter

import (
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestParseFilter_SingleCondition(t *testing.T) {
	filter, err := ParseFilter("status=error")

	if err != nil {
		t.Fatalf("ParseFilter failed: %v", err)
	}

	if len(filter.Conditions) != 1 {
		t.Fatalf("Expected 1 condition, got %d", len(filter.Conditions))
	}

	cond := filter.Conditions[0]
	if cond.Field != "status" {
		t.Errorf("Expected field 'status', got '%s'", cond.Field)
	}

	if cond.Value != "error" {
		t.Errorf("Expected value 'error', got '%s'", cond.Value)
	}
}

func TestParseFilter_MultipleConditions(t *testing.T) {
	filter, err := ParseFilter("status=error,tool=Grep")

	if err != nil {
		t.Fatalf("ParseFilter failed: %v", err)
	}

	if len(filter.Conditions) != 2 {
		t.Fatalf("Expected 2 conditions, got %d", len(filter.Conditions))
	}
}

func TestParseFilter_InvalidFormat(t *testing.T) {
	_, err := ParseFilter("invalid_format")

	if err == nil {
		t.Error("Expected error for invalid filter format")
	}
}

func TestApplyFilter_ToolCalls_StatusError(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Grep",
			Status:   "success",
		},
		{
			UUID:     "uuid-2",
			ToolName: "Read",
			Status:   "error",
			Error:    "file not found",
		},
		{
			UUID:     "uuid-3",
			ToolName: "Bash",
			Status:   "error",
			Error:    "command failed",
		},
	}

	filter, _ := ParseFilter("status=error")
	result := ApplyFilter(toolCalls, filter)
	filtered, ok := result.([]parser.ToolCall)
	if !ok {
		t.Fatalf("Expected []parser.ToolCall, got %T", result)
	}

	if len(filtered) != 2 {
		t.Fatalf("Expected 2 filtered results, got %d", len(filtered))
	}

	// Verify all are error status
	for _, tc := range filtered {
		if tc.Status != "error" {
			t.Errorf("Expected status 'error', got '%s'", tc.Status)
		}
	}
}

func TestApplyFilter_ToolCalls_ToolName(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Grep"},
		{UUID: "uuid-2", ToolName: "Read"},
		{UUID: "uuid-3", ToolName: "Grep"},
	}

	filter, _ := ParseFilter("tool=Grep")
	result := ApplyFilter(toolCalls, filter)
	filtered, ok := result.([]parser.ToolCall)
	if !ok {
		t.Fatalf("Expected []parser.ToolCall, got %T", result)
	}

	if len(filtered) != 2 {
		t.Fatalf("Expected 2 filtered results, got %d", len(filtered))
	}

	for _, tc := range filtered {
		if tc.ToolName != "Grep" {
			t.Errorf("Expected tool name 'Grep', got '%s'", tc.ToolName)
		}
	}
}

func TestApplyFilter_SessionEntries_Type(t *testing.T) {
	entries := []parser.SessionEntry{
		{Type: "user", UUID: "uuid-1"},
		{Type: "assistant", UUID: "uuid-2"},
		{Type: "user", UUID: "uuid-3"},
	}

	filter, _ := ParseFilter("type=user")
	result := ApplyFilter(entries, filter)
	filtered, ok := result.([]parser.SessionEntry)
	if !ok {
		t.Fatalf("Expected []parser.SessionEntry, got %T", result)
	}

	if len(filtered) != 2 {
		t.Fatalf("Expected 2 filtered results, got %d", len(filtered))
	}

	for _, entry := range filtered {
		if entry.Type != "user" {
			t.Errorf("Expected type 'user', got '%s'", entry.Type)
		}
	}
}

func TestApplyFilter_EmptyFilter(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Grep"},
		{UUID: "uuid-2", ToolName: "Read"},
	}

	// Empty filter should return all data
	filter := &Filter{}
	result := ApplyFilter(toolCalls, filter)
	filtered, ok := result.([]parser.ToolCall)
	if !ok {
		t.Fatalf("Expected []parser.ToolCall, got %T", result)
	}

	if len(filtered) != len(toolCalls) {
		t.Errorf("Expected %d results with empty filter, got %d", len(toolCalls), len(filtered))
	}
}

func TestApplyFilter_NoMatches(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Grep", Status: "success"},
	}

	filter, _ := ParseFilter("status=error")
	result := ApplyFilter(toolCalls, filter)
	filtered, ok := result.([]parser.ToolCall)
	if !ok {
		t.Fatalf("Expected []parser.ToolCall, got %T", result)
	}

	if len(filtered) != 0 {
		t.Errorf("Expected 0 results, got %d", len(filtered))
	}
}

// Stage 8.4 tests

func TestParseWhereCondition(t *testing.T) {
	// Test: ParseWhereCondition should work like ParseFilter
	filter, err := ParseWhereCondition("status=error")
	if err != nil {
		t.Fatalf("ParseWhereCondition failed: %v", err)
	}

	if len(filter.Conditions) != 1 {
		t.Fatalf("Expected 1 condition, got %d", len(filter.Conditions))
	}

	// Test multiple conditions
	filter2, err := ParseWhereCondition("tool=Bash,status=error")
	if err != nil {
		t.Fatalf("ParseWhereCondition failed: %v", err)
	}

	if len(filter2.Conditions) != 2 {
		t.Fatalf("Expected 2 conditions, got %d", len(filter2.Conditions))
	}
}

func TestValidateFilterField_ValidFields(t *testing.T) {
	// Test valid fields for tool_calls
	validToolFields := []string{"status", "tool", "uuid"}
	for _, field := range validToolFields {
		err := ValidateFilterField(field, "tool_calls")
		if err != nil {
			t.Errorf("Field '%s' should be valid for tool_calls, got error: %v", field, err)
		}
	}

	// Test valid fields for entries
	validEntryFields := []string{"type", "uuid", "role"}
	for _, field := range validEntryFields {
		err := ValidateFilterField(field, "entries")
		if err != nil {
			t.Errorf("Field '%s' should be valid for entries, got error: %v", field, err)
		}
	}
}

func TestValidateFilterField_InvalidFields(t *testing.T) {
	// Test invalid field
	err := ValidateFilterField("invalid_field", "tool_calls")
	if err == nil {
		t.Error("Expected error for invalid field, got nil")
	}

	// Test invalid data type
	err = ValidateFilterField("status", "invalid_type")
	if err == nil {
		t.Error("Expected error for invalid data type, got nil")
	}
}

func TestApplyWhere_ValidFilter(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Grep", Status: "success"},
		{UUID: "uuid-2", ToolName: "Read", Status: "error"},
	}

	result, err := ApplyWhere(toolCalls, "status=error", "tool_calls")
	if err != nil {
		t.Fatalf("ApplyWhere failed: %v", err)
	}

	filtered, ok := result.([]parser.ToolCall)
	if !ok {
		t.Fatalf("Expected []parser.ToolCall, got %T", result)
	}

	if len(filtered) != 1 {
		t.Fatalf("Expected 1 filtered result, got %d", len(filtered))
	}

	if filtered[0].Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", filtered[0].Status)
	}
}

func TestApplyWhere_InvalidField(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Grep", Status: "success"},
	}

	_, err := ApplyWhere(toolCalls, "invalid_field=value", "tool_calls")
	if err == nil {
		t.Error("Expected error for invalid field, got nil")
	}

	// Verify error message is helpful
	if !strings.Contains(err.Error(), "invalid field") {
		t.Errorf("Expected helpful error message, got: %v", err)
	}
}

func TestApplyWhere_MultipleConditions(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "success"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "error"},
		{UUID: "uuid-3", ToolName: "Read", Status: "error"},
	}

	result, err := ApplyWhere(toolCalls, "tool=Bash,status=error", "tool_calls")
	if err != nil {
		t.Fatalf("ApplyWhere failed: %v", err)
	}

	filtered, ok := result.([]parser.ToolCall)
	if !ok {
		t.Fatalf("Expected []parser.ToolCall, got %T", result)
	}

	if len(filtered) != 1 {
		t.Fatalf("Expected 1 filtered result (Bash AND error), got %d", len(filtered))
	}

	if filtered[0].ToolName != "Bash" || filtered[0].Status != "error" {
		t.Errorf("Expected Bash tool with error status, got %s/%s", filtered[0].ToolName, filtered[0].Status)
	}
}
