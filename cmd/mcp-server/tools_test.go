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
		"stats_first", "max_output_bytes", "output_format",
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
	if params["max_output_bytes"].Type != "number" {
		t.Errorf("max_output_bytes should be number, got %s", params["max_output_bytes"].Type)
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
	if _, ok := merged["max_output_bytes"]; !ok {
		t.Error("standard parameter 'max_output_bytes' missing")
	}
}

func TestAllToolsHaveStandardParameters(t *testing.T) {
	tools := getToolDefinitions()

	requiredParams := []string{
		"scope", "jq_filter", "stats_only", "stats_first", "max_output_bytes", "output_format",
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
		// Skip aggregate_stats (will be removed in Stage 15.2)
		if tool.Name == "aggregate_stats" {
			continue
		}

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

		// Skip aggregate_stats (will be removed in Stage 15.2)
		if tool.Name == "aggregate_stats" {
			continue
		}

		// Should end with "Default scope: project." or "Default scope: session."
		validEndings := []string{
			"Default scope: project.",
			"Default scope: session.",
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

		if !strings.Contains(maxMsgLen.Description, "default: 500") {
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
