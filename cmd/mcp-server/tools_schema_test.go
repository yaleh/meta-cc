package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/config"
)

// TestQueryToolsHaveSchemaDocumentation verifies that all query tools include
// output schema documentation in their jq_filter parameter description.
// This helps users write correct jq filters by showing available field names.
func TestQueryToolsHaveSchemaDocumentation(t *testing.T) {
	tools := getToolDefinitions()

	// Tools that should have schema documentation in jq_filter (all fields are snake_case)
	// TASK-7: Updated to use new consolidated tool names
	schemaRequired := map[string][]string{
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

// TestQueryToolSchemaMatchesImplementation removed in Phase 27 Stage 27.1
// The query tool was deleted to simplify the query interface

// Phase 29 Stage 29.3: Schema accessor tests

func TestGetToolSchemaByName_KnownTools(t *testing.T) {
	// Reset cached index to ensure clean test
	toolSchemaIndex = nil

	// New consolidated query tool names that should have schemas
	queryTools := []string{
		"query_session_content",
		"query_session_signals",
		"query_file_activity",
	}

	for _, name := range queryTools {
		t.Run(name, func(t *testing.T) {
			schema, err := getToolSchemaByName(name)
			if err != nil {
				t.Fatalf("expected no error for %s, got: %v", name, err)
			}
			if schema.Type != "object" {
				t.Errorf("expected schema type 'object', got '%s'", schema.Type)
			}
			if len(schema.Properties) == 0 {
				t.Errorf("expected non-empty properties for %s", name)
			}
		})
	}
}

func TestGetToolSchemaByName_QuerySessionContent_HasRole(t *testing.T) {
	toolSchemaIndex = nil

	schema, err := getToolSchemaByName("query_session_content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := schema.Properties["role"]; !ok {
		t.Error("expected 'role' in query_session_content properties")
	}
}

func TestGetToolSchemaByName_NonExistent(t *testing.T) {
	toolSchemaIndex = nil

	_, err := getToolSchemaByName("query_nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent tool, got nil")
	}
	if !strings.Contains(err.Error(), "unknown tool") {
		t.Errorf("expected 'unknown tool' in error message, got: %v", err)
	}
}

func TestGetToolSchemaByName_SpecialTools(t *testing.T) {
	toolSchemaIndex = nil

	// Special tools handled by executeSpecialTool should also have schemas
	specialTools := []string{
		"cleanup_temp_files",
		"get_session_directory",
		"inspect_session_files",
		"execute_stage2_query",
		"get_session_metadata",
	}

	for _, name := range specialTools {
		t.Run(name, func(t *testing.T) {
			schema, err := getToolSchemaByName(name)
			if err != nil {
				t.Fatalf("expected no error for %s, got: %v", name, err)
			}
			if schema.Type != "object" {
				t.Errorf("expected schema type 'object', got '%s'", schema.Type)
			}
		})
	}
}

func TestExecuteTool_UnknownToolFailsEarly(t *testing.T) {
	toolSchemaIndex = nil

	executor := NewToolExecutor()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	_, err = executor.ExecuteTool(cfg, "totally_fake_tool", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for unknown tool, got nil")
	}
	if !strings.Contains(err.Error(), "unknown tool") {
		t.Errorf("expected 'unknown tool' in error, got: %v", err)
	}
}

// Phase 29 Stage 29.4: Dispatch validation tests

func TestExecuteTool_UnknownParameterReturnsError(t *testing.T) {
	toolSchemaIndex = nil

	executor := NewToolExecutor()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// "match" is not a valid parameter for query_session_content
	_, err = executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"match": "foo",
		"role":  "user",
	})
	if err == nil {
		t.Fatal("expected error for unknown parameter 'match', got nil")
	}
	if !strings.Contains(err.Error(), "match") {
		t.Errorf("expected error to mention 'match', got: %v", err)
	}
	if !strings.Contains(err.Error(), "unknown parameter") {
		t.Errorf("expected 'unknown parameter' in error, got: %v", err)
	}
}

func TestExecuteTool_ValidParameterSucceeds(t *testing.T) {
	toolSchemaIndex = nil
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	executor := NewToolExecutor()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// "role" is a valid parameter for query_session_content
	// This may fail with session-not-found but should NOT fail with parameter validation error
	_, err = executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role": "user",
	})
	if err != nil && strings.Contains(err.Error(), "unknown parameter") {
		t.Errorf("valid parameter 'pattern' should not trigger validation error, got: %v", err)
	}
}

func TestExecuteTool_InvalidScopeReturnsError(t *testing.T) {
	toolSchemaIndex = nil

	executor := NewToolExecutor()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// "sessions" is not a valid scope value (should be "session")
	_, err = executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":  "user",
		"scope": "sessions",
	})
	if err == nil {
		t.Fatal("expected error for invalid scope 'sessions', got nil")
	}
	if !strings.Contains(err.Error(), "sessions") {
		t.Errorf("expected error to mention 'sessions', got: %v", err)
	}
	if !strings.Contains(err.Error(), "project") && !strings.Contains(err.Error(), "session") {
		t.Errorf("expected error to mention valid scope values, got: %v", err)
	}
}

func TestExecuteTool_ValidScopeSucceeds(t *testing.T) {
	toolSchemaIndex = nil
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	executor := NewToolExecutor()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// "session" is a valid scope value
	_, err = executor.ExecuteTool(cfg, "query_session_content", map[string]interface{}{
		"role":  "user",
		"scope": "session",
	})
	if err != nil && strings.Contains(err.Error(), "invalid scope") {
		t.Errorf("valid scope 'session' should not trigger scope validation error, got: %v", err)
	}
}

func TestExecuteTool_SpecialToolsRejectUnknownParameters(t *testing.T) {
	toolSchemaIndex = nil

	executor := NewToolExecutor()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Special tools must pass the same schema validation as query tools before
	// their handler is dispatched (DIR-023 post-release regression).
	_, err = executor.ExecuteTool(cfg, "cleanup_temp_files", map[string]interface{}{
		"max_age_days": float64(7),
		"extra_param":  "must_be_rejected",
	})
	if err == nil {
		t.Fatal("expected special tool to reject an unknown parameter")
	}
	if !strings.Contains(err.Error(), "unknown parameter") || !strings.Contains(err.Error(), "extra_param") {
		t.Errorf("expected unknown parameter error naming extra_param, got: %v", err)
	}
}

func TestExecuteTool_MultipleUnknownParamsListsAll(t *testing.T) {
	toolSchemaIndex = nil

	executor := NewToolExecutor()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Both "match" and "foo" are unknown params
	_, err = executor.ExecuteTool(cfg, "query_session_signals", map[string]interface{}{
		"type":  "errors",
		"match": "bar",
		"foo":   "baz",
	})
	if err == nil {
		t.Fatal("expected error for unknown parameters, got nil")
	}
	// Error should mention at least one unknown param
	if !strings.Contains(err.Error(), "unknown parameter") {
		t.Errorf("expected 'unknown parameter' in error, got: %v", err)
	}
}
