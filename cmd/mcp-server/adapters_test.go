package main

import (
	"testing"

	querypkg "github.com/yaleh/meta-cc/internal/query"
)

// TestAdaptQueryTools verifies query_tools adapter maps parameters correctly
func TestAdaptQueryTools(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		expected querypkg.QueryParams
	}{
		{
			name: "filter by tool name",
			args: map[string]interface{}{
				"tool": "Read",
			},
			expected: querypkg.QueryParams{
				Resource: "tools",
				Filter: querypkg.FilterSpec{
					ToolName: "Read",
				},
			},
		},
		{
			name: "filter by status",
			args: map[string]interface{}{
				"status": "error",
			},
			expected: querypkg.QueryParams{
				Resource: "tools",
				Filter: querypkg.FilterSpec{
					ToolStatus: "error",
				},
			},
		},
		{
			name: "filter by tool and status",
			args: map[string]interface{}{
				"tool":   "Read",
				"status": "error",
			},
			expected: querypkg.QueryParams{
				Resource: "tools",
				Filter: querypkg.FilterSpec{
					ToolName:   "Read",
					ToolStatus: "error",
				},
			},
		},
		{
			name: "with limit",
			args: map[string]interface{}{
				"tool":  "Bash",
				"limit": float64(10), // JSON unmarshals numbers as float64
			},
			expected: querypkg.QueryParams{
				Resource: "tools",
				Filter: querypkg.FilterSpec{
					ToolName: "Bash",
				},
				Output: querypkg.OutputSpec{
					Limit: 10,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adaptQueryTools(tt.args)

			if result.Resource != tt.expected.Resource {
				t.Errorf("Resource: got %s, want %s", result.Resource, tt.expected.Resource)
			}

			if result.Filter.ToolName != tt.expected.Filter.ToolName {
				t.Errorf("Filter.ToolName: got %s, want %s", result.Filter.ToolName, tt.expected.Filter.ToolName)
			}

			if result.Filter.ToolStatus != tt.expected.Filter.ToolStatus {
				t.Errorf("Filter.ToolStatus: got %s, want %s", result.Filter.ToolStatus, tt.expected.Filter.ToolStatus)
			}

			if result.Output.Limit != tt.expected.Output.Limit {
				t.Errorf("Output.Limit: got %d, want %d", result.Output.Limit, tt.expected.Output.Limit)
			}
		})
	}
}

// TestAdaptQueryUserMessages verifies query_user_messages adapter
func TestAdaptQueryUserMessages(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		expected querypkg.QueryParams
	}{
		{
			name: "search pattern",
			args: map[string]interface{}{
				"pattern": "error.*bug",
			},
			expected: querypkg.QueryParams{
				Resource: "messages",
				Filter: querypkg.FilterSpec{
					Role:         "user",
					ContentMatch: "error.*bug",
				},
			},
		},
		{
			name: "with limit",
			args: map[string]interface{}{
				"pattern": "fix",
				"limit":   float64(5),
			},
			expected: querypkg.QueryParams{
				Resource: "messages",
				Filter: querypkg.FilterSpec{
					Role:         "user",
					ContentMatch: "fix",
				},
				Output: querypkg.OutputSpec{
					Limit: 5,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adaptQueryUserMessages(tt.args)

			if result.Resource != tt.expected.Resource {
				t.Errorf("Resource: got %s, want %s", result.Resource, tt.expected.Resource)
			}

			if result.Filter.Role != tt.expected.Filter.Role {
				t.Errorf("Filter.Role: got %s, want %s", result.Filter.Role, tt.expected.Filter.Role)
			}

			if result.Filter.ContentMatch != tt.expected.Filter.ContentMatch {
				t.Errorf("Filter.ContentMatch: got %s, want %s", result.Filter.ContentMatch, tt.expected.Filter.ContentMatch)
			}

			if result.Output.Limit != tt.expected.Output.Limit {
				t.Errorf("Output.Limit: got %d, want %d", result.Output.Limit, tt.expected.Output.Limit)
			}
		})
	}
}

// TestAdaptQueryAssistantMessages verifies query_assistant_messages adapter
func TestAdaptQueryAssistantMessages(t *testing.T) {
	args := map[string]interface{}{
		"pattern":   ".*",
		"min_tools": float64(5),
		"limit":     float64(10),
	}

	result := adaptQueryAssistantMessages(args)

	if result.Resource != "messages" {
		t.Errorf("Resource: got %s, want messages", result.Resource)
	}

	if result.Filter.Role != "assistant" {
		t.Errorf("Filter.Role: got %s, want assistant", result.Filter.Role)
	}

	if result.Filter.ContentMatch != ".*" {
		t.Errorf("Filter.ContentMatch: got %s, want .*", result.Filter.ContentMatch)
	}

	// Note: min_tools is not mapped to FilterSpec (would need custom post-processing)
}

// TestAdaptQueryFiles verifies query_files adapter
func TestAdaptQueryFiles(t *testing.T) {
	args := map[string]interface{}{
		"threshold": float64(5),
	}

	result := adaptQueryFiles(args)

	if result.Resource != "tools" {
		t.Errorf("Resource: got %s, want tools", result.Resource)
	}

	if result.Filter.ToolName != "Read|Edit|Write" {
		t.Errorf("Filter.ToolName: got %s, want Read|Edit|Write", result.Filter.ToolName)
	}

	if result.Aggregate.Function != "count" {
		t.Errorf("Aggregate.Function: got %s, want count", result.Aggregate.Function)
	}

	if result.Aggregate.Field != "file_path" {
		t.Errorf("Aggregate.Field: got %s, want file_path", result.Aggregate.Field)
	}
}

// TestCanAdaptToUnifiedQuery verifies adapter registry
func TestCanAdaptToUnifiedQuery(t *testing.T) {
	tests := []struct {
		toolName string
		canAdapt bool
	}{
		{"query_tools", true},
		{"query_user_messages", true},
		{"query_assistant_messages", true},
		{"query_files", true},
		{"get_session_stats", true},
		{"query_conversation", false},   // Not yet implemented
		{"query_tool_sequences", false}, // Not yet implemented
		{"cleanup_temp_files", false},   // Not adaptable (special tool)
		{"nonexistent_tool", false},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			result := canAdaptToUnifiedQuery(tt.toolName)
			if result != tt.canAdapt {
				t.Errorf("canAdaptToUnifiedQuery(%s): got %v, want %v", tt.toolName, result, tt.canAdapt)
			}
		})
	}
}

// TestAdaptLegacyTool verifies adaptLegacyTool function
func TestAdaptLegacyTool(t *testing.T) {
	t.Run("valid tool", func(t *testing.T) {
		args := map[string]interface{}{
			"tool":   "Read",
			"status": "error",
		}

		params, err := adaptLegacyTool("query_tools", args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if params.Resource != "tools" {
			t.Errorf("Resource: got %s, want tools", params.Resource)
		}

		if params.Filter.ToolName != "Read" {
			t.Errorf("Filter.ToolName: got %s, want Read", params.Filter.ToolName)
		}
	})

	t.Run("invalid tool", func(t *testing.T) {
		_, err := adaptLegacyTool("nonexistent_tool", map[string]interface{}{})
		if err == nil {
			t.Error("expected error for nonexistent tool, got nil")
		}
	})
}

// TestAdapterBackwardCompatibility verifies that adapters produce
// equivalent query parameters to legacy implementations
func TestAdapterBackwardCompatibility(t *testing.T) {
	t.Run("query_tools equivalence", func(t *testing.T) {
		// Legacy call: query_tools(tool="Read", status="error")
		legacyArgs := map[string]interface{}{
			"tool":   "Read",
			"status": "error",
		}

		// Adapted to unified query
		adapted := adaptQueryTools(legacyArgs)

		// Should select "tools" resource
		if adapted.Resource != "tools" {
			t.Errorf("expected resource 'tools', got %s", adapted.Resource)
		}

		// Should filter by tool_name and tool_status
		if adapted.Filter.ToolName != "Read" || adapted.Filter.ToolStatus != "error" {
			t.Errorf("filter mismatch: tool_name=%s, tool_status=%s",
				adapted.Filter.ToolName, adapted.Filter.ToolStatus)
		}
	})

	t.Run("query_user_messages equivalence", func(t *testing.T) {
		// Legacy call: query_user_messages(pattern="fix.*bug")
		legacyArgs := map[string]interface{}{
			"pattern": "fix.*bug",
		}

		// Adapted to unified query
		adapted := adaptQueryUserMessages(legacyArgs)

		// Should select "messages" resource with role="user"
		if adapted.Resource != "messages" {
			t.Errorf("expected resource 'messages', got %s", adapted.Resource)
		}

		if adapted.Filter.Role != "user" {
			t.Errorf("expected role 'user', got %s", adapted.Filter.Role)
		}

		if adapted.Filter.ContentMatch != "fix.*bug" {
			t.Errorf("expected content_match 'fix.*bug', got %s", adapted.Filter.ContentMatch)
		}
	})
}
