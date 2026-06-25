package tools_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/mcp/tools"
)

func TestStandardToolParameters(t *testing.T) {
	params := tools.StandardToolParameters()

	requiredParams := []string{
		"scope", "provider", "jq_filter", "stats_only",
		"stats_first", "inline_threshold_bytes", "output_format",
	}

	for _, param := range requiredParams {
		if _, ok := params[param]; !ok {
			t.Errorf("missing standard parameter: %s", param)
		}
	}
}

func TestMergeParameters(t *testing.T) {
	specific := map[string]tools.Property{
		"limit": {
			Type:        "number",
			Description: "Max results",
		},
		"scope": {
			Type:        "string",
			Description: "Custom scope description",
		},
	}

	merged := tools.MergeParameters(specific)

	if _, ok := merged["limit"]; !ok {
		t.Error("specific parameter 'limit' missing")
	}
	if _, ok := merged["jq_filter"]; !ok {
		t.Error("standard parameter 'jq_filter' missing")
	}
	if merged["scope"].Description != "Custom scope description" {
		t.Errorf("parameter override failed, got: %s", merged["scope"].Description)
	}
}

func TestGetToolDefinitions(t *testing.T) {
	defs := tools.GetToolDefinitions()
	if len(defs) == 0 {
		t.Error("expected non-empty tool definitions")
	}

	// Verify all tools serialize to JSON
	for _, tool := range defs {
		_, err := json.Marshal(tool)
		if err != nil {
			t.Errorf("tool %s failed to serialize: %v", tool.Name, err)
		}
	}
}

func TestBuildToolSchemaIndex(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	if len(index) == 0 {
		t.Error("expected non-empty schema index")
	}

	// Check a known tool (new consolidated tools)
	if _, ok := index["query_session_content"]; !ok {
		t.Error("expected query_session_content in index")
	}
}

func TestGetToolSchemaByName(t *testing.T) {
	index := tools.BuildToolSchemaIndex()

	schema, err := tools.GetToolSchemaByName(index, "query_session_content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "object" {
		t.Errorf("expected object type, got %s", schema.Type)
	}

	_, err = tools.GetToolSchemaByName(index, "nonexistent_tool")
	if err == nil {
		t.Fatal("expected error for nonexistent tool")
	}
	if !strings.Contains(err.Error(), "unknown tool") {
		t.Errorf("expected 'unknown tool' in error, got: %v", err)
	}
}

func TestValidateToolArgs_ValidTool_ValidArgs(t *testing.T) {
	err := tools.ValidateToolArgs("query_session_content", map[string]interface{}{
		"role":  "user",
		"limit": float64(10),
	})
	if err != nil {
		t.Fatalf("unexpected error for valid tool+args: %v", err)
	}
}

func TestValidateToolArgs_ValidTool_EmptyArgs(t *testing.T) {
	err := tools.ValidateToolArgs("query_session_signals", map[string]interface{}{
		"type": "errors",
	})
	if err != nil {
		t.Fatalf("unexpected error for valid args: %v", err)
	}
}

func TestValidateToolArgs_ValidTool_InvalidArgKey(t *testing.T) {
	err := tools.ValidateToolArgs("query_session_content", map[string]interface{}{
		"unknown_key": "value",
	})
	if err == nil {
		t.Fatal("expected error for invalid arg key")
	}
	if !strings.Contains(err.Error(), "unknown_key") {
		t.Errorf("expected 'unknown_key' in error, got: %v", err)
	}
}

func TestValidateToolArgs_UnknownTool(t *testing.T) {
	err := tools.ValidateToolArgs("no_such_tool", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
	if !strings.Contains(err.Error(), "no_such_tool") {
		t.Errorf("expected tool name in error, got: %v", err)
	}
}

// Phase A: New consolidated tool schema tests

func TestNewConsolidatedToolsPresent(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	newTools := []string{
		"query_session_content",
		"query_session_signals",
		"query_file_activity",
	}
	for _, name := range newTools {
		if _, ok := index[name]; !ok {
			t.Errorf("expected new tool %q in schema index", name)
		}
	}
}

func TestQuerySessionContentSchema(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	s, err := tools.GetToolSchemaByName(index, "query_session_content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Type != "object" {
		t.Errorf("expected object type, got %s", s.Type)
	}
	// must have role parameter
	if _, ok := s.Properties["role"]; !ok {
		t.Error("query_session_content must have 'role' parameter")
	}
	// role is required
	found := false
	for _, r := range s.Required {
		if r == "role" {
			found = true
		}
	}
	if !found {
		t.Error("query_session_content: 'role' must be required")
	}
}

func TestQuerySessionSignalsSchema(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	s, err := tools.GetToolSchemaByName(index, "query_session_signals")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.Properties["type"]; !ok {
		t.Error("query_session_signals must have 'type' parameter")
	}
	found := false
	for _, r := range s.Required {
		if r == "type" {
			found = true
		}
	}
	if !found {
		t.Error("query_session_signals: 'type' must be required")
	}
}

func TestQueryFileActivitySchema(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	s, err := tools.GetToolSchemaByName(index, "query_file_activity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.Properties["type"]; !ok {
		t.Error("query_file_activity must have 'type' parameter")
	}
	found := false
	for _, r := range s.Required {
		if r == "type" {
			found = true
		}
	}
	if !found {
		t.Error("query_file_activity: 'type' must be required")
	}
}

func TestValidateToolArgs_QuerySessionContent(t *testing.T) {
	err := tools.ValidateToolArgs("query_session_content", map[string]interface{}{
		"role":  "user",
		"limit": float64(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateToolArgs_QuerySessionSignals(t *testing.T) {
	err := tools.ValidateToolArgs("query_session_signals", map[string]interface{}{
		"type": "errors",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateToolArgs_QueryFileActivity(t *testing.T) {
	err := tools.ValidateToolArgs("query_file_activity", map[string]interface{}{
		"type": "snapshots",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestQuerySessionContentSchema_ToolRole_DescribesContextFields verifies that the
// block_type parameter description mentions context fields (timestamp/sessionId/turn).
func TestQuerySessionContentSchema_ToolRole_DescribesContextFields(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	s, err := tools.GetToolSchemaByName(index, "query_session_content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	prop, ok := s.Properties["block_type"]
	if !ok {
		t.Fatal("query_session_content must have 'block_type' parameter")
	}
	desc := prop.Description
	if !strings.Contains(desc, "timestamp") && !strings.Contains(desc, "sessionId") {
		t.Errorf("block_type description should mention context fields (timestamp/sessionId), got: %q", desc)
	}
}

// Phase D-description: Description hints tests (TASK-13)

func TestQuerySessionContentDescriptionHints(t *testing.T) {
	defs := tools.GetToolDefinitions()
	for _, tool := range defs {
		if tool.Name == "query_session_content" {
			if !strings.Contains(tool.Description, "timestamp") {
				t.Errorf("query_session_content description should contain 'timestamp', got: %q", tool.Description)
			}
			if !strings.Contains(tool.Description, "inspect_session_files") {
				t.Errorf("query_session_content description should contain 'inspect_session_files', got: %q", tool.Description)
			}
			return
		}
	}
	t.Fatal("query_session_content not found in tool definitions")
}

func TestQuerySessionSignalsDescriptionHints(t *testing.T) {
	defs := tools.GetToolDefinitions()
	for _, tool := range defs {
		if tool.Name == "query_session_signals" {
			if !strings.Contains(tool.Description, "since") {
				t.Errorf("query_session_signals description should contain 'since', got: %q", tool.Description)
			}
			if !strings.Contains(tool.Description, "until") {
				t.Errorf("query_session_signals description should contain 'until', got: %q", tool.Description)
			}
			return
		}
	}
	t.Fatal("query_session_signals not found in tool definitions")
}

func TestExecuteStage2QueryDescriptionHints(t *testing.T) {
	defs := tools.GetToolDefinitions()
	for _, tool := range defs {
		if tool.Name == "execute_stage2_query" {
			if !strings.Contains(tool.Description, "inspect_session_files") {
				t.Errorf("execute_stage2_query description should contain 'inspect_session_files', got: %q", tool.Description)
			}
			if !strings.Contains(tool.Description, "transform") {
				t.Errorf("execute_stage2_query description should contain 'transform', got: %q", tool.Description)
			}
			return
		}
	}
	t.Fatal("execute_stage2_query not found in tool definitions")
}

// Phase D: Old tools must be removed from tool definitions.
func TestOldToolsRemovedFromDefinitions(t *testing.T) {
	removedTools := []string{
		"query_tool_errors",
		"query_token_usage",
		"query_system_errors",
		"query_timestamps",
		"query_tools",
		"query_summaries",
		"query_tool_blocks",
		"query_user_messages",
		"query_conversation_flow",
		"query_file_snapshots",
	}
	index := tools.BuildToolSchemaIndex()
	for _, name := range removedTools {
		if _, ok := index[name]; ok {
			t.Errorf("old tool %q should have been removed from definitions in Phase D", name)
		}
	}
}
