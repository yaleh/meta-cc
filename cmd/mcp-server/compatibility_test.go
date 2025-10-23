package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/parser"
	querypkg "github.com/yaleh/meta-cc/internal/query"
)

// TestAdapterCompatibility tests that legacy tool adapters produce correct QueryParams
func TestAdapterCompatibility(t *testing.T) {
	t.Run("adaptQueryTools_basic", func(t *testing.T) {
		args := map[string]interface{}{
			"tool":   "Read",
			"status": "error",
		}

		params := adaptQueryTools(args)

		assert.Equal(t, "tools", params.Resource)
		assert.Equal(t, "Read", params.Filter.ToolName)
		assert.Equal(t, "error", params.Filter.ToolStatus)
	})

	t.Run("adaptQueryTools_with_limit", func(t *testing.T) {
		args := map[string]interface{}{
			"tool":   "Edit",
			"status": "success",
			"limit":  10,
		}

		params := adaptQueryTools(args)

		assert.Equal(t, "tools", params.Resource)
		assert.Equal(t, "Edit", params.Filter.ToolName)
		assert.Equal(t, "success", params.Filter.ToolStatus)
		assert.Equal(t, 10, params.Output.Limit)
	})

	t.Run("adaptQueryTools_empty_args", func(t *testing.T) {
		args := map[string]interface{}{}

		params := adaptQueryTools(args)

		assert.Equal(t, "tools", params.Resource)
		assert.Empty(t, params.Filter.ToolName)
		assert.Empty(t, params.Filter.ToolStatus)
	})

	t.Run("adaptQueryUserMessages_basic", func(t *testing.T) {
		args := map[string]interface{}{
			"pattern": "fix.*bug",
		}

		params := adaptQueryUserMessages(args)

		assert.Equal(t, "messages", params.Resource)
		assert.Equal(t, "user", params.Filter.Role)
		assert.Equal(t, "fix.*bug", params.Filter.ContentMatch)
	})

	t.Run("adaptQueryUserMessages_with_limit", func(t *testing.T) {
		args := map[string]interface{}{
			"pattern": "error",
			"limit":   5,
		}

		params := adaptQueryUserMessages(args)

		assert.Equal(t, "messages", params.Resource)
		assert.Equal(t, "user", params.Filter.Role)
		assert.Equal(t, "error", params.Filter.ContentMatch)
		assert.Equal(t, 5, params.Output.Limit)
	})

	t.Run("adaptQueryAssistantMessages_basic", func(t *testing.T) {
		args := map[string]interface{}{
			"pattern": ".*",
		}

		params := adaptQueryAssistantMessages(args)

		assert.Equal(t, "messages", params.Resource)
		assert.Equal(t, "assistant", params.Filter.Role)
		assert.Equal(t, ".*", params.Filter.ContentMatch)
	})

	t.Run("adaptQueryFiles_basic", func(t *testing.T) {
		args := map[string]interface{}{
			"threshold": 5,
		}

		params := adaptQueryFiles(args)

		assert.Equal(t, "tools", params.Resource)
		assert.Equal(t, "Read|Edit|Write", params.Filter.ToolName)
		assert.Equal(t, "count", params.Aggregate.Function)
		assert.Equal(t, "file_path", params.Aggregate.Field)
	})

	t.Run("adaptGetSessionStats_basic", func(t *testing.T) {
		args := map[string]interface{}{}

		params := adaptGetSessionStats(args)

		assert.Equal(t, "entries", params.Resource)
		// Session stats is complex and requires special handling
	})
}

// TestLegacyToolAdapters tests the adapter mapping and functions
func TestLegacyToolAdapters(t *testing.T) {
	t.Run("can_adapt_supported_tools", func(t *testing.T) {
		supportedTools := []string{
			"query_tools",
			"query_user_messages",
			"query_assistant_messages",
			"query_files",
			"get_session_stats",
		}

		for _, tool := range supportedTools {
			assert.True(t, canAdaptToUnifiedQuery(tool),
				"Tool %s should be adaptable", tool)
		}
	})

	t.Run("cannot_adapt_unsupported_tools", func(t *testing.T) {
		unsupportedTools := []string{
			"unknown_tool",
			"query_context",
			"query_conversation",
		}

		for _, tool := range unsupportedTools {
			assert.False(t, canAdaptToUnifiedQuery(tool),
				"Tool %s should not be adaptable yet", tool)
		}
	})

	t.Run("adaptLegacyTool_success", func(t *testing.T) {
		args := map[string]interface{}{
			"tool":   "Read",
			"status": "error",
		}

		params, err := adaptLegacyTool("query_tools", args)

		require.NoError(t, err)
		assert.Equal(t, "tools", params.Resource)
		assert.Equal(t, "Read", params.Filter.ToolName)
	})

	t.Run("adaptLegacyTool_error_unknown", func(t *testing.T) {
		args := map[string]interface{}{}

		_, err := adaptLegacyTool("unknown_tool", args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no adapter")
	})
}

// TestBackwardCompatibility_QueryResults tests that adapted queries produce
// equivalent results to legacy implementations (structure compatibility)
func TestBackwardCompatibility_QueryResults(t *testing.T) {
	// Create test entries
	entries := []parser.SessionEntry{
		{
			Type:       "user",
			UUID:       "user-1",
			Timestamp:  "2025-10-23T10:00:00Z",
			SessionID:  "session-1",
			ParentUUID: "",
			GitBranch:  "main",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "text",
						Text: "Read config file",
					},
				},
			},
		},
		{
			Type:       "assistant",
			UUID:       "assistant-1",
			Timestamp:  "2025-10-23T10:00:10Z",
			SessionID:  "session-1",
			ParentUUID: "user-1",
			GitBranch:  "main",
			Message: &parser.Message{
				Role: "assistant",
				Content: []parser.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &parser.ToolUse{
							ID:   "tool-1",
							Name: "Read",
							Input: map[string]interface{}{
								"file_path": "/config.yaml",
							},
						},
					},
				},
			},
		},
		{
			Type:       "user",
			UUID:       "tool-result-1",
			Timestamp:  "2025-10-23T10:00:15Z",
			SessionID:  "session-1",
			ParentUUID: "assistant-1",
			GitBranch:  "main",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &parser.ToolResult{
							ToolUseID: "tool-1",
							Content:   "config data",
							IsError:   false,
							Status:    "success",
						},
					},
				},
			},
		},
	}

	t.Run("query_tools_adapter_result_structure", func(t *testing.T) {
		// Adapt legacy tool params
		args := map[string]interface{}{
			"tool":   "Read",
			"status": "success",
		}
		params := adaptQueryTools(args)

		// Execute unified query
		result, err := querypkg.Query(entries, params)
		require.NoError(t, err)

		// Verify result structure matches legacy output
		tools, ok := result.([]parser.ToolCall)
		require.True(t, ok, "Result should be []parser.ToolCall for backward compatibility")
		assert.NotEmpty(t, tools)

		// Verify tool structure has expected fields
		for _, tool := range tools {
			assert.NotEmpty(t, tool.UUID, "Tool should have UUID")
			assert.NotEmpty(t, tool.ToolName, "Tool should have ToolName")
			assert.NotEmpty(t, tool.Timestamp, "Tool should have Timestamp")
			assert.NotEmpty(t, tool.Status, "Tool should have Status")
			assert.NotNil(t, tool.Input, "Tool should have Input")
		}
	})

	t.Run("query_user_messages_adapter_result_structure", func(t *testing.T) {
		// Adapt legacy tool params
		args := map[string]interface{}{
			"pattern": "config",
		}
		params := adaptQueryUserMessages(args)

		// Execute unified query
		result, err := querypkg.Query(entries, params)
		require.NoError(t, err)

		// Verify result structure
		messages, ok := result.([]querypkg.MessageView)
		require.True(t, ok, "Result should be []MessageView")
		assert.NotEmpty(t, messages)

		// Verify message structure has expected fields
		for _, msg := range messages {
			assert.NotEmpty(t, msg.UUID, "Message should have UUID")
			assert.NotEmpty(t, msg.SessionID, "Message should have SessionID")
			assert.NotEmpty(t, msg.Timestamp, "Message should have Timestamp")
			assert.Equal(t, "user", msg.Role, "Should only have user messages")
			assert.NotNil(t, msg.ContentBlocks, "Message should have ContentBlocks")
		}
	})

	t.Run("query_assistant_messages_adapter_result_structure", func(t *testing.T) {
		// Adapt legacy tool params
		args := map[string]interface{}{}
		params := adaptQueryAssistantMessages(args)

		// Execute unified query
		result, err := querypkg.Query(entries, params)
		require.NoError(t, err)

		// Verify result structure
		messages, ok := result.([]querypkg.MessageView)
		require.True(t, ok, "Result should be []MessageView")
		assert.NotEmpty(t, messages)

		// Verify only assistant messages
		for _, msg := range messages {
			assert.Equal(t, "assistant", msg.Role)
		}
	})

	t.Run("query_files_adapter_result_structure", func(t *testing.T) {
		// Adapt legacy tool params
		args := map[string]interface{}{
			"threshold": 1,
		}
		params := adaptQueryFiles(args)

		// Execute unified query
		result, err := querypkg.Query(entries, params)
		require.NoError(t, err)

		// Verify result structure for aggregation
		results, ok := result.([]map[string]interface{})
		require.True(t, ok, "Aggregated result should be []map[string]interface{}")
		assert.NotEmpty(t, results)

		// Verify aggregation structure
		for _, r := range results {
			assert.Contains(t, r, "file_path", "Should have file_path field")
			assert.Contains(t, r, "count", "Should have count field")
		}
	})
}

// TestBackwardCompatibility_EdgeCases tests edge cases in adapter behavior
func TestBackwardCompatibility_EdgeCases(t *testing.T) {
	t.Run("empty_filter_produces_all_results", func(t *testing.T) {
		args := map[string]interface{}{}
		params := adaptQueryTools(args)

		// Empty filter should match all tools
		assert.Equal(t, "tools", params.Resource)
		assert.Empty(t, params.Filter.ToolName)
		assert.Empty(t, params.Filter.ToolStatus)
	})

	t.Run("nil_args_map_handled", func(t *testing.T) {
		// Should not panic with nil map
		var args map[string]interface{}
		params := adaptQueryTools(args)

		assert.Equal(t, "tools", params.Resource)
	})

	t.Run("missing_optional_params_use_defaults", func(t *testing.T) {
		args := map[string]interface{}{
			"tool": "Read",
			// status missing
		}
		params := adaptQueryTools(args)

		assert.Equal(t, "Read", params.Filter.ToolName)
		assert.Empty(t, params.Filter.ToolStatus)
		assert.Equal(t, 0, params.Output.Limit)
	})

	t.Run("zero_limit_means_unlimited", func(t *testing.T) {
		args := map[string]interface{}{
			"limit": 0,
		}
		params := adaptQueryTools(args)

		assert.Equal(t, 0, params.Output.Limit)
	})

	t.Run("pattern_empty_string_matches_all", func(t *testing.T) {
		args := map[string]interface{}{
			"pattern": "",
		}
		params := adaptQueryUserMessages(args)

		assert.Equal(t, "user", params.Filter.Role)
		assert.Empty(t, params.Filter.ContentMatch)
	})
}

// TestBackwardCompatibility_OutputFormat tests that output format is consistent
func TestBackwardCompatibility_OutputFormat(t *testing.T) {
	// Test that adapted queries produce results in expected JSON format

	t.Run("tools_output_has_snake_case", func(t *testing.T) {
		entries := []parser.SessionEntry{
			{
				Type:       "assistant",
				UUID:       "assistant-1",
				Timestamp:  "2025-10-23T10:00:00Z",
				SessionID:  "session-1",
				ParentUUID: "user-1",
				Message: &parser.Message{
					Role: "assistant",
					Content: []parser.ContentBlock{
						{
							Type: "tool_use",
							ToolUse: &parser.ToolUse{
								ID:   "tool-1",
								Name: "Read",
								Input: map[string]interface{}{
									"file_path": "/test.txt",
								},
							},
						},
					},
				},
			},
			{
				Type:       "user",
				UUID:       "tool-result-1",
				Timestamp:  "2025-10-23T10:00:05Z",
				SessionID:  "session-1",
				ParentUUID: "assistant-1",
				Message: &parser.Message{
					Role: "user",
					Content: []parser.ContentBlock{
						{
							Type: "tool_result",
							ToolResult: &parser.ToolResult{
								ToolUseID: "tool-1",
								Content:   "content",
								Status:    "success",
							},
						},
					},
				},
			},
		}

		params := adaptQueryTools(map[string]interface{}{})
		result, err := querypkg.Query(entries, params)
		require.NoError(t, err)

		tools, ok := result.([]parser.ToolCall)
		require.True(t, ok)
		assert.NotEmpty(t, tools)

		// Verify structure uses snake_case compatible fields
		tool := tools[0]
		assert.NotEmpty(t, tool.ToolName)
		assert.NotEmpty(t, tool.Status)
		assert.NotEmpty(t, tool.Timestamp)
		assert.NotEmpty(t, tool.UUID)
	})

	t.Run("messages_output_has_required_fields", func(t *testing.T) {
		entries := []parser.SessionEntry{
			{
				Type:       "user",
				UUID:       "user-1",
				Timestamp:  "2025-10-23T10:00:00Z",
				SessionID:  "session-1",
				ParentUUID: "",
				Message: &parser.Message{
					Role: "user",
					Content: []parser.ContentBlock{
						{
							Type: "text",
							Text: "Hello",
						},
					},
				},
			},
		}

		params := adaptQueryUserMessages(map[string]interface{}{})
		result, err := querypkg.Query(entries, params)
		require.NoError(t, err)

		messages, ok := result.([]querypkg.MessageView)
		require.True(t, ok)
		assert.NotEmpty(t, messages)

		// Verify MessageView structure
		msg := messages[0]
		assert.NotEmpty(t, msg.UUID)
		assert.NotEmpty(t, msg.SessionID)
		assert.NotEmpty(t, msg.Timestamp)
		assert.NotEmpty(t, msg.Role)
		assert.NotNil(t, msg.ContentBlocks)
	})

	t.Run("aggregation_output_has_count_field", func(t *testing.T) {
		entries := []parser.SessionEntry{
			{
				Type:       "assistant",
				UUID:       "assistant-1",
				Timestamp:  "2025-10-23T10:00:00Z",
				SessionID:  "session-1",
				ParentUUID: "user-1",
				Message: &parser.Message{
					Role: "assistant",
					Content: []parser.ContentBlock{
						{
							Type: "tool_use",
							ToolUse: &parser.ToolUse{
								ID:   "tool-1",
								Name: "Read",
								Input: map[string]interface{}{
									"file_path": "/test.txt",
								},
							},
						},
					},
				},
			},
			{
				Type:       "user",
				UUID:       "tool-result-1",
				Timestamp:  "2025-10-23T10:00:05Z",
				SessionID:  "session-1",
				ParentUUID: "assistant-1",
				Message: &parser.Message{
					Role: "user",
					Content: []parser.ContentBlock{
						{
							Type: "tool_result",
							ToolResult: &parser.ToolResult{
								ToolUseID: "tool-1",
								Content:   "content",
								Status:    "success",
							},
						},
					},
				},
			},
		}

		params := adaptQueryFiles(map[string]interface{}{})
		result, err := querypkg.Query(entries, params)
		require.NoError(t, err)

		results, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, results)

		// Verify aggregation has expected structure
		for _, r := range results {
			assert.Contains(t, r, "count")
			count, ok := r["count"].(int)
			require.True(t, ok)
			assert.Greater(t, count, 0)
		}
	})
}
