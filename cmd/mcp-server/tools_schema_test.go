package main

import (
	"strings"
	"testing"
)

// TestQueryToolsHaveSchemaDocumentation verifies that all query tools include
// output schema documentation in their jq_filter parameter description.
// This helps users write correct jq filters by showing available field names.
func TestQueryToolsHaveSchemaDocumentation(t *testing.T) {
	tools := getToolDefinitions()

	// Tools that should have schema documentation in jq_filter (all fields are snake_case)
	schemaRequired := map[string][]string{
		"query_tools":              {"tool_name", "status", "timestamp", "error", "input", "output", "uuid"},
		"query_user_messages":      {"turn", "timestamp", "content"},
		"query_file_access":        {"file", "total_accesses", "operations", "timeline"},
		"query_tool_sequences":     {"pattern", "count", "occurrences", "time_span_minutes"},
		"query_successful_prompts": {"turn", "content", "quality_score"},
		"query_project_state":      {"timestamp", "type"},
	}

	for _, tool := range tools {
		expectedFields, shouldHaveSchema := schemaRequired[tool.Name]
		if !shouldHaveSchema {
			continue
		}

		t.Run(tool.Name, func(t *testing.T) {
			// Get jq_filter parameter from merged schema
			jqFilterProp, exists := tool.InputSchema.Properties["jq_filter"]
			if !exists {
				t.Fatalf("Tool %s missing jq_filter parameter", tool.Name)
			}

			desc := jqFilterProp.Description

			// Check for "Output schema:" section
			if !strings.Contains(desc, "Output schema:") && !strings.Contains(desc, "output schema:") {
				t.Errorf("Tool %s jq_filter description missing 'Output schema:' section.\nDescription: %s",
					tool.Name, desc)
			}

			// Check that all expected fields are documented
			for _, field := range expectedFields {
				if !strings.Contains(desc, field) {
					t.Errorf("Tool %s jq_filter description missing field '%s'.\nDescription: %s",
						tool.Name, field, desc)
				}
			}

			// Check for an example
			if !strings.Contains(desc, "Example:") && !strings.Contains(desc, "example:") {
				t.Errorf("Tool %s jq_filter description missing example.\nDescription: %s",
					tool.Name, desc)
			}
		})
	}
}

// TestJqFilterDescriptionFormat verifies the jq_filter description format
// across all tools to ensure consistency and usability
func TestJqFilterDescriptionFormat(t *testing.T) {
	tools := getToolDefinitions()

	for _, tool := range tools {
		// Skip tools without jq_filter (cleanup_temp_files, etc.)
		jqFilterProp, exists := tool.InputSchema.Properties["jq_filter"]
		if !exists {
			continue
		}

		t.Run(tool.Name, func(t *testing.T) {
			desc := jqFilterProp.Description

			// Should NOT contain "(default: '.[]')" format - this triggers Claude Code bug
			if strings.Contains(desc, "(default: '.[]')") {
				t.Errorf("Tool %s jq_filter uses problematic format '(default: ...)' which triggers quote escaping bug.\n"+
					"Use format: 'Defaults to ... when omitted' instead.\nDescription: %s",
					tool.Name, desc)
			}

			// Should mention default value somewhere
			if !strings.Contains(desc, "default") && !strings.Contains(desc, "Default") {
				t.Errorf("Tool %s jq_filter description should mention default value.\nDescription: %s",
					tool.Name, desc)
			}

			// Should warn about quoting
			if !strings.Contains(desc, "NOT wrap in quotes") && !strings.Contains(desc, "Do not quote") {
				t.Errorf("Tool %s jq_filter description should warn against wrapping in quotes.\nDescription: %s",
					tool.Name, desc)
			}
		})
	}
}

// TestQueryToolSchemaMatchesImplementation verifies that the 'query' tool schema
// matches the Phase 25 jq-based implementation and does NOT include obsolete
// object-based parameters that were removed.
func TestQueryToolSchemaMatchesImplementation(t *testing.T) {
	tools := getToolDefinitions()

	var queryTool *Tool
	for i := range tools {
		if tools[i].Name == "query" {
			queryTool = &tools[i]
			break
		}
	}

	if queryTool == nil {
		t.Fatal("query tool not found in tool definitions")
	}

	// Verify obsolete parameters are NOT in schema (Phase 25 breaking change)
	obsoleteParams := []string{"resource", "filter", "transform", "aggregate"}
	for _, param := range obsoleteParams {
		if _, exists := queryTool.InputSchema.Properties[param]; exists {
			t.Errorf("query tool schema contains obsolete parameter '%s' that was removed in Phase 25.\n"+
				"These object-based parameters were replaced with jq-based parameters (jq_filter, jq_transform).\n"+
				"See docs/plans/25/PHASE-25-IMPLEMENTATION-PLAN.md lines 127-148 for breaking change details.",
				param)
		}
	}

	// Verify jq-based parameters ARE in schema (Phase 25 new design)
	// Note: jq_filter and other standard parameters come from StandardToolParameters()
	// and are merged in via MergeParameters(), so they should exist
	requiredParams := []string{"jq_filter", "scope"}
	for _, param := range requiredParams {
		if _, exists := queryTool.InputSchema.Properties[param]; !exists {
			t.Errorf("query tool schema missing required parameter '%s' from Phase 25 jq-based design", param)
		}
	}

	// Verify description reflects jq-based interface
	if !strings.Contains(queryTool.Description, "query") && !strings.Contains(queryTool.Description, "jq") {
		t.Errorf("query tool description should mention jq-based querying.\nDescription: %s", queryTool.Description)
	}
}
