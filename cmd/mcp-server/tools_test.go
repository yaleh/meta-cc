package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStandardToolParameters(t *testing.T) {
	params := StandardToolParameters()

	// Verify all standard parameters exist
	requiredParams := []string{
		"scope", "jq_filter", "stats_only",
		"stats_first", "inline_threshold_bytes", "output_format",
	}

	for _, param := range requiredParams {
		if _, ok := params[param]; !ok {
			t.Errorf("missing standard parameter: %s", param)
		}
	}

	// Verify parameter types
	if params["scope"].Type != "string" {
		t.Errorf("scope should be string, got %s", params["scope"].Type)
	}
	if params["jq_filter"].Type != "string" {
		t.Errorf("jq_filter should be string, got %s", params["jq_filter"].Type)
	}
	if params["stats_only"].Type != "boolean" {
		t.Errorf("stats_only should be boolean, got %s", params["stats_only"].Type)
	}
	if params["stats_first"].Type != "boolean" {
		t.Errorf("stats_first should be boolean, got %s", params["stats_first"].Type)
	}
	if params["inline_threshold_bytes"].Type != "number" {
		t.Errorf("inline_threshold_bytes should be number, got %s", params["inline_threshold_bytes"].Type)
	}
	if params["output_format"].Type != "string" {
		t.Errorf("output_format should be string, got %s", params["output_format"].Type)
	}
}

func TestMergeParameters(t *testing.T) {
	specific := map[string]Property{
		"limit": {
			Type:        "number",
			Description: "Max results",
		},
		"scope": { // Override scope description
			Type:        "string",
			Description: "Custom scope description",
		},
	}

	merged := MergeParameters(specific)

	// Verify specific params are included
	if _, ok := merged["limit"]; !ok {
		t.Error("specific parameter 'limit' missing")
	}

	// Verify standard params are included
	if _, ok := merged["jq_filter"]; !ok {
		t.Error("standard parameter 'jq_filter' missing")
	}

	// Verify override works
	if merged["scope"].Description != "Custom scope description" {
		t.Errorf("parameter override failed, got: %s", merged["scope"].Description)
	}

	// Verify standard params that weren't overridden still exist
	if _, ok := merged["stats_only"]; !ok {
		t.Error("standard parameter 'stats_only' missing")
	}
	if _, ok := merged["inline_threshold_bytes"]; !ok {
		t.Error("standard parameter 'inline_threshold_bytes' missing")
	}
}

func TestAllToolsHaveStandardParameters(t *testing.T) {
	tools := getToolDefinitions()

	requiredParams := []string{
		"scope", "jq_filter", "stats_only", "stats_first", "inline_threshold_bytes", "output_format",
	}

	// Tools that should have message truncation parameters (Stage 15.1)
	messageTruncationTools := map[string]bool{
		"query_user_messages": true,
	}

	messageTruncationParams := []string{
		"max_message_length",
		"content_summary",
	}

	for _, tool := range tools {
		t.Run(tool.Name, func(t *testing.T) {
			// Skip deprecated tools (they still need standard params but we're phasing them out)
			if strings.Contains(tool.Description, "DEPRECATED") {
				t.Logf("Skipping deprecated tool: %s", tool.Name)
				return
			}

			// Skip utility tools that don't follow query tool patterns
			if tool.Name == "cleanup_temp_files" || tool.Name == "list_capabilities" || tool.Name == "get_capability" {
				t.Logf("Skipping utility tool: %s", tool.Name)
				return
			}

			// Check all standard parameters exist
			for _, param := range requiredParams {
				if _, ok := tool.InputSchema.Properties[param]; !ok {
					t.Errorf("tool %s missing standard parameter: %s", tool.Name, param)
				}
			}

			// Check message truncation parameters for specific tools
			if messageTruncationTools[tool.Name] {
				for _, param := range messageTruncationParams {
					if _, ok := tool.InputSchema.Properties[param]; !ok {
						t.Errorf("tool %s missing message truncation parameter: %s", tool.Name, param)
					}
				}
			}
		})
	}
}

func TestToolDescriptionLength(t *testing.T) {
	tools := getToolDefinitions()

	for _, tool := range tools {
		if len(tool.Description) > 100 {
			t.Errorf("tool %s description too long: %d chars (max: 100)\nDescription: %s",
				tool.Name, len(tool.Description), tool.Description)
		}
	}
}

func TestToolsJSONSerialization(t *testing.T) {
	tools := getToolDefinitions()

	// Verify all tools can be serialized to JSON
	for _, tool := range tools {
		_, err := json.Marshal(tool)
		if err != nil {
			t.Errorf("tool %s failed to serialize: %v", tool.Name, err)
		}
	}
}

func TestToolDescriptionConsistency(t *testing.T) {
	tools := getToolDefinitions()

	for _, tool := range tools {
		if strings.Contains(tool.Description, "DEPRECATED") {
			continue
		}

		// Skip utility tools that don't follow "Default scope:" pattern
		if tool.Name == "cleanup_temp_files" || tool.Name == "list_capabilities" || tool.Name == "get_capability" {
			continue
		}

		// Should end with "Default scope: project.", "Default scope: session.", or "Default scope: none."
		validEndings := []string{
			"Default scope: project.",
			"Default scope: session.",
			"Default scope: none.",
		}

		hasValidEnding := false
		for _, ending := range validEndings {
			if strings.HasSuffix(tool.Description, ending) {
				hasValidEnding = true
				break
			}
		}

		if !hasValidEnding {
			t.Errorf("tool %s has inconsistent description ending: %s",
				tool.Name, tool.Description)
		}
	}
}

func TestQueryUserMessagesMessageTruncationParams(t *testing.T) {
	tools := getToolDefinitions()

	var queryUserMessages *Tool
	for i := range tools {
		if tools[i].Name == "query_user_messages" {
			queryUserMessages = &tools[i]
			break
		}
	}

	if queryUserMessages == nil {
		t.Fatal("query_user_messages tool not found")
	}

	props := queryUserMessages.InputSchema.Properties

	// Test max_message_length parameter
	t.Run("max_message_length", func(t *testing.T) {
		maxMsgLen, exists := props["max_message_length"]
		if !exists {
			t.Error("query_user_messages missing max_message_length parameter")
			return
		}

		if maxMsgLen.Type != "number" {
			t.Errorf("max_message_length should be number type, got %s", maxMsgLen.Type)
		}

		if !strings.Contains(maxMsgLen.Description, "default: 0") {
			t.Errorf("max_message_length should mention default value, got: %s", maxMsgLen.Description)
		}

		if !strings.Contains(maxMsgLen.Description, "Max chars per message") {
			t.Errorf("max_message_length description should be descriptive, got: %s", maxMsgLen.Description)
		}
	})

	// Test content_summary parameter
	t.Run("content_summary", func(t *testing.T) {
		contentSummary, exists := props["content_summary"]
		if !exists {
			t.Error("query_user_messages missing content_summary parameter")
			return
		}

		if contentSummary.Type != "boolean" {
			t.Errorf("content_summary should be boolean type, got %s", contentSummary.Type)
		}

		if !strings.Contains(contentSummary.Description, "preview") || !strings.Contains(contentSummary.Description, "100 chars") {
			t.Errorf("content_summary should mention preview and 100 chars, got: %s", contentSummary.Description)
		}
	})
}

// TestExtractToolsRemoved verifies that extract_tools has been removed (use query_tools instead)
func TestExtractToolsRemoved(t *testing.T) {
	tools := getToolDefinitions()

	for _, tool := range tools {
		if tool.Name == "extract_tools" {
			t.Error("extract_tools should be removed. Use query_tools instead.")
		}
	}
}

// TestToolDescriptionsAccurate verifies tool descriptions match actual behavior (Stage 16.5)
func TestToolDescriptionsAccurate(t *testing.T) {
	tools := getToolDefinitions()

	// Tools with limit parameter should not have misleading "default: 20/10" descriptions
	limitTools := map[string]bool{
		"query_tools":              true,
		"query_user_messages":      true,
		"query_successful_prompts": true,
	}

	for _, tool := range tools {
		if !limitTools[tool.Name] {
			continue
		}

		t.Run(tool.Name, func(t *testing.T) {
			limitProp, exists := tool.InputSchema.Properties["limit"]
			if !exists {
				t.Errorf("tool %s missing limit parameter", tool.Name)
				return
			}

			// Check that description mentions "no limit by default" or similar
			if strings.Contains(limitProp.Description, "default: 20") || strings.Contains(limitProp.Description, "default: 10") {
				t.Errorf("tool %s still has misleading default limit in description: %s", tool.Name, limitProp.Description)
			}

			// Should mention hybrid output mode or no limit by default
			if !strings.Contains(limitProp.Description, "no limit by default") &&
				!strings.Contains(limitProp.Description, "rely on hybrid output mode") {
				t.Errorf("tool %s limit description should mention 'no limit by default' or 'hybrid output mode', got: %s",
					tool.Name, limitProp.Description)
			}
		})
	}
}

// TestLimitParameterBehavior verifies that limit parameter behavior is correctly documented
func TestLimitParameterBehavior(t *testing.T) {
	tools := getToolDefinitions()

	limitTools := []string{
		"query_tools",
		"query_user_messages",
		"query_successful_prompts",
	}

	for _, toolName := range limitTools {
		var tool *Tool
		for i := range tools {
			if tools[i].Name == toolName {
				tool = &tools[i]
				break
			}
		}

		if tool == nil {
			t.Errorf("tool %s not found", toolName)
			continue
		}

		t.Run(toolName, func(t *testing.T) {
			limitProp, exists := tool.InputSchema.Properties["limit"]
			if !exists {
				t.Errorf("tool %s missing limit parameter", toolName)
				return
			}

			if limitProp.Type != "number" {
				t.Errorf("tool %s limit should be number, got %s", toolName, limitProp.Type)
			}

			// Description should be informative about hybrid output mode
			desc := limitProp.Description
			if len(desc) < 20 {
				t.Errorf("tool %s limit description too short: %s", toolName, desc)
			}

			// Should not contain misleading default values
			if strings.Contains(desc, "(default: ") {
				t.Errorf("tool %s limit description contains misleading default value: %s", toolName, desc)
			}
		})
	}
}

// TestQueryToolSequencesToolDefinition verifies the query_tool_sequences tool is correctly defined
func TestQueryToolSequencesToolDefinition(t *testing.T) {
	tools := getToolDefinitions()

	var tool *Tool
	for i := range tools {
		if tools[i].Name == "query_tool_sequences" {
			tool = &tools[i]
			break
		}
	}

	if tool == nil {
		t.Fatal("query_tool_sequences tool not found")
	}

	// Verify description
	if !strings.Contains(tool.Description, "tool sequences") {
		t.Errorf("description should mention 'tool sequences', got: %s", tool.Description)
	}

	if !strings.HasSuffix(tool.Description, "Default scope: project.") {
		t.Errorf("description should end with 'Default scope: project.', got: %s", tool.Description)
	}

	// Verify required parameters exist
	props := tool.InputSchema.Properties
	requiredParams := []string{
		"pattern", "include_builtin_tools", "min_occurrences",
	}

	for _, param := range requiredParams {
		if _, exists := props[param]; !exists {
			t.Errorf("query_tool_sequences missing parameter: %s", param)
		}
	}

	// Verify parameter types
	if props["pattern"].Type != "string" {
		t.Errorf("pattern should be string, got %s", props["pattern"].Type)
	}
	if props["include_builtin_tools"].Type != "boolean" {
		t.Errorf("include_builtin_tools should be boolean, got %s", props["include_builtin_tools"].Type)
	}
	if props["min_occurrences"].Type != "number" {
		t.Errorf("min_occurrences should be number, got %s", props["min_occurrences"].Type)
	}

	// Verify standard parameters exist
	standardParams := []string{"scope", "jq_filter", "stats_only", "stats_first", "inline_threshold_bytes", "output_format"}
	for _, param := range standardParams {
		if _, exists := props[param]; !exists {
			t.Errorf("query_tool_sequences missing standard parameter: %s", param)
		}
	}
}

// TestQueryFileAccessToolDefinition verifies the query_file_access tool is correctly defined
func TestQueryFileAccessToolDefinition(t *testing.T) {
	tools := getToolDefinitions()

	var tool *Tool
	for i := range tools {
		if tools[i].Name == "query_file_access" {
			tool = &tools[i]
			break
		}
	}

	if tool == nil {
		t.Fatal("query_file_access tool not found")
	}

	// Verify description
	if !strings.Contains(tool.Description, "file") {
		t.Errorf("description should mention 'file', got: %s", tool.Description)
	}

	if !strings.HasSuffix(tool.Description, "Default scope: project.") {
		t.Errorf("description should end with 'Default scope: project.', got: %s", tool.Description)
	}

	// Verify required parameters exist
	props := tool.InputSchema.Properties
	requiredParams := []string{
		"file",
	}

	for _, param := range requiredParams {
		if _, exists := props[param]; !exists {
			t.Errorf("query_file_access missing parameter: %s", param)
		}
	}

	// Verify parameter types
	if props["file"].Type != "string" {
		t.Errorf("file should be string, got %s", props["file"].Type)
	}

	// Verify required field
	if len(tool.InputSchema.Required) < 1 || tool.InputSchema.Required[0] != "file" {
		t.Errorf("file should be required parameter, got: %v", tool.InputSchema.Required)
	}

	// Verify standard parameters exist
	standardParams := []string{"scope", "jq_filter", "stats_only", "stats_first", "inline_threshold_bytes", "output_format"}
	for _, param := range standardParams {
		if _, exists := props[param]; !exists {
			t.Errorf("query_file_access missing standard parameter: %s", param)
		}
	}
}

// TestQueryProjectStateToolDefinition verifies the query_project_state tool is correctly defined
func TestQueryProjectStateToolDefinition(t *testing.T) {
	tools := getToolDefinitions()

	var tool *Tool
	for i := range tools {
		if tools[i].Name == "query_project_state" {
			tool = &tools[i]
			break
		}
	}

	if tool == nil {
		t.Fatal("query_project_state tool not found")
	}

	// Verify description
	if !strings.Contains(tool.Description, "project state") {
		t.Errorf("description should mention 'project state', got: %s", tool.Description)
	}

	if !strings.HasSuffix(tool.Description, "Default scope: project.") {
		t.Errorf("description should end with 'Default scope: project.', got: %s", tool.Description)
	}

	// Verify standard parameters exist (this tool has no specific parameters)
	props := tool.InputSchema.Properties
	standardParams := []string{"scope", "jq_filter", "stats_only", "stats_first", "inline_threshold_bytes", "output_format"}
	for _, param := range standardParams {
		if _, exists := props[param]; !exists {
			t.Errorf("query_project_state missing standard parameter: %s", param)
		}
	}
}

// TestToolCountIncreasedTo14 verifies that the tool count has increased from 12 to 14
func TestToolCountIncreasedTo14(t *testing.T) {
	tools := getToolDefinitions()

	// Phase 19 adds 2 new tools: query_assistant_messages, query_conversation
	// Phase 22 Stage 22.2 adds 1 new tool: list_capabilities
	// Phase 22 Stage 22.3 adds 1 new tool: get_capability
	// Phase 24 Stage 24.3 adds 1 new tool: query (unified query interface)
	// Phase 25 Stage 25.2 adds 1 new tool: query_raw (raw jq interface)
	// Phase 25 Stage 25.3 adds 8 convenience tools (18 -> 26)
	// Phase 25 Stage 25.4 removes 6 deprecated tools (26 -> 20)
	// Removed: query_context, query_tools_advanced, query_time_series,
	// query_assistant_messages, query_conversation, query_files
	// New target: 20 tools (1 core + 1 raw + 8 convenience + 7 legacy + 3 utility)
	expectedCount := 20
	actualCount := len(tools)

	if actualCount != expectedCount {
		t.Errorf("expected %d tools after Phase 25 Stage 25.3, got %d", expectedCount, actualCount)

		// List all tool names for debugging
		t.Log("Current tools:")
		for _, tool := range tools {
			t.Logf("  - %s", tool.Name)
		}
	}
}

// TestListCapabilitiesToolRegistration verifies that list_capabilities tool is registered
func TestListCapabilitiesToolRegistration(t *testing.T) {
	tools := getToolDefinitions()

	var listCapTool *Tool
	for i := range tools {
		if tools[i].Name == "list_capabilities" {
			listCapTool = &tools[i]
			break
		}
	}

	if listCapTool == nil {
		t.Fatal("list_capabilities tool not found")
	}

	// Verify description
	if !strings.Contains(listCapTool.Description, "capabilities") {
		t.Errorf("description should mention 'capabilities', got: %s", listCapTool.Description)
	}

	// list_capabilities is a utility tool, not a query tool, so it shouldn't have standard params
	// It should NOT have "Default scope:" suffix
	if strings.Contains(listCapTool.Description, "Default scope:") {
		t.Errorf("list_capabilities should not have 'Default scope:' (it's a utility tool), got: %s", listCapTool.Description)
	}
}

// TestListCapabilitiesToolSchema verifies the list_capabilities tool schema
func TestListCapabilitiesToolSchema(t *testing.T) {
	tools := getToolDefinitions()

	var tool *Tool
	for i := range tools {
		if tools[i].Name == "list_capabilities" {
			tool = &tools[i]
			break
		}
	}

	if tool == nil {
		t.Fatal("list_capabilities tool not found")
	}

	props := tool.InputSchema.Properties

	// list_capabilities should NOT have standard parameters
	// It's a utility tool like cleanup_temp_files
	standardParams := []string{"scope", "jq_filter", "stats_only", "stats_first", "inline_threshold_bytes", "output_format"}
	for _, param := range standardParams {
		if _, exists := props[param]; exists {
			t.Errorf("list_capabilities should NOT have standard parameter: %s (it's a utility tool)", param)
		}
	}

	// list_capabilities has no public parameters (all are hidden test parameters)
	// Hidden test parameters (_sources, _disable_cache) should NOT be in the schema
	if _, exists := props["_sources"]; exists {
		t.Error("_sources is a hidden test parameter and should NOT be in schema")
	}
	if _, exists := props["_disable_cache"]; exists {
		t.Error("_disable_cache is a hidden test parameter and should NOT be in schema")
	}

	// Verify schema is valid (should be empty properties object)
	if len(props) > 0 {
		t.Errorf("list_capabilities should have no public parameters, got %d parameters", len(props))
		for name := range props {
			t.Logf("  - %s", name)
		}
	}
}

// TestJqFilterDescriptionImproved verifies that jq_filter parameter description includes quote escaping guidance
func TestJqFilterDescriptionImproved(t *testing.T) {
	params := StandardToolParameters()

	jqFilterParam := params["jq_filter"]

	// Verify description contains important guidance
	desc := jqFilterParam.Description

	// Should mention "IMPORTANT" or "Do NOT" to highlight quote escaping rule
	if !strings.Contains(desc, "IMPORTANT") && !strings.Contains(desc, "Do NOT") {
		t.Errorf("jq_filter description should highlight quote escaping with 'IMPORTANT' or 'Do NOT', got: %s", desc)
	}

	// Should contain example of correct syntax
	if !strings.Contains(desc, ".[] | {field: .field}") {
		t.Errorf("jq_filter description should include correct syntax example, got: %s", desc)
	}

	// Should warn about quotes
	if !strings.Contains(desc, "quotes") {
		t.Errorf("jq_filter description should warn about quotes, got: %s", desc)
	}

	// Should still mention default value (case-insensitive)
	descLower := strings.ToLower(desc)
	if !strings.Contains(descLower, "default") {
		t.Errorf("jq_filter description should still mention default value, got: %s", desc)
	}

	t.Logf("jq_filter description: %s", desc)
}
