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
		"jq_filter", "stats_only", "stats_first", "max_output_bytes",
	}

	for _, tool := range tools {
		// Skip deprecated tools
		if tool.Name == "analyze_errors" || tool.Name == "aggregate_stats" {
			continue
		}

		for _, param := range requiredParams {
			if _, ok := tool.InputSchema.Properties[param]; !ok {
				t.Errorf("tool %s missing parameter: %s", tool.Name, param)
			}
		}
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
