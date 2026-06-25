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
	// TASK-7: query_session_content replaces query_user_messages
	messageTruncationTools := map[string]bool{
		"query_session_content": true,
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

			// Skip utility tools and Stage 1/2 tools that don't follow query tool patterns
			if tool.Name == "cleanup_temp_files" ||
				tool.Name == "get_session_directory" || tool.Name == "inspect_session_files" || tool.Name == "execute_stage2_query" {
				t.Logf("Skipping utility/two-stage tool: %s", tool.Name)
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

	// Tools with embedded usage hints may have longer descriptions (TASK-13)
	extendedDescTools := map[string]bool{
		"query_session_content": true,
		"query_session_signals": true,
		"execute_stage2_query":  true,
	}

	for _, tool := range tools {
		limit := 100
		if extendedDescTools[tool.Name] {
			limit = 300
		}
		if len(tool.Description) > limit {
			t.Errorf("tool %s description too long: %d chars (max: %d)\nDescription: %s",
				tool.Name, len(tool.Description), limit, tool.Description)
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

	// Tools with embedded usage hints (TASK-13) may have content after "Default scope:" sentence
	extendedDescTools := map[string]bool{
		"query_session_content": true,
		"query_session_signals": true,
	}

	for _, tool := range tools {
		if strings.Contains(tool.Description, "DEPRECATED") {
			continue
		}

		// Skip utility tools and two-stage tools (no scope param) that don't follow "Default scope:" pattern
		if tool.Name == "cleanup_temp_files" ||
			tool.Name == "get_session_directory" || tool.Name == "inspect_session_files" || tool.Name == "execute_stage2_query" {
			continue
		}

		// Tools with extended hints only need to contain "Default scope:" somewhere
		if extendedDescTools[tool.Name] {
			if !strings.Contains(tool.Description, "Default scope:") {
				t.Errorf("tool %s description must contain 'Default scope:', got: %s",
					tool.Name, tool.Description)
			}
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

func TestQuerySessionContentMessageTruncationParams(t *testing.T) {
	tools := getToolDefinitions()

	var querySessionContent *Tool
	for i := range tools {
		if tools[i].Name == "query_session_content" {
			querySessionContent = &tools[i]
			break
		}
	}

	if querySessionContent == nil {
		t.Fatal("query_session_content tool not found")
	}

	props := querySessionContent.InputSchema.Properties

	// Test max_message_length parameter
	t.Run("max_message_length", func(t *testing.T) {
		maxMsgLen, exists := props["max_message_length"]
		if !exists {
			t.Error("query_session_content missing max_message_length parameter")
			return
		}

		if maxMsgLen.Type != "number" {
			t.Errorf("max_message_length should be number type, got %s", maxMsgLen.Type)
		}
	})

	// Test content_summary parameter
	t.Run("content_summary", func(t *testing.T) {
		contentSummary, exists := props["content_summary"]
		if !exists {
			t.Error("query_session_content missing content_summary parameter")
			return
		}

		if contentSummary.Type != "boolean" {
			t.Errorf("content_summary should be boolean type, got %s", contentSummary.Type)
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
	// TASK-7: Updated to use new consolidated tool names
	limitTools := map[string]bool{
		"query_session_content":    true,
		"query_session_signals":    true,
		"query_file_activity":      true,
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

	// TASK-7: Updated to use new consolidated tool names
	limitTools := []string{
		"query_session_content",
		"query_session_signals",
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

// TestToolCountIncreasedTo14 verifies that the tool count is correct after TASK-7
func TestToolCountIncreasedTo14(t *testing.T) {
	tools := getToolDefinitions()

	// TASK-7: Removed 10 old query_* tools, added 3 consolidated tools = net -7
	// New target: 15 tools
	// - 3 consolidated query tools (query_session_content, query_session_signals, query_file_activity)
	// - 1 utility tool (cleanup_temp_files)
	// - 4 two-stage query tools (get_session_directory, inspect_session_files, execute_stage2_query, get_session_metadata)
	// - 6 analysis tools (analyze_errors, quality_scan, get_work_patterns, get_timeline, analyze_bugs, get_tech_debt)
	// - 1 doc session signals tool (query_edit_sequences)
	expectedCount := 15
	actualCount := len(tools)

	if actualCount != expectedCount {
		t.Errorf("expected %d tools after Phase 45.1, got %d", expectedCount, actualCount)

		// List all tool names for debugging
		t.Log("Current tools:")
		for _, tool := range tools {
			t.Logf("  - %s", tool.Name)
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
