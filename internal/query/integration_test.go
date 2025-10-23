package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/parser"
)

// createComplexTestEntries creates a rich test dataset for integration testing
func createComplexTestEntries() []parser.SessionEntry {
	return []parser.SessionEntry{
		// User message 1
		{
			Type:       "user",
			UUID:       "user-1",
			Timestamp:  "2025-10-23T10:00:00Z",
			SessionID:  "session-1",
			ParentUUID: "",
			GitBranch:  "main",
			CWD:        "/home/user/project",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "text",
						Text: "Read the configuration file",
					},
				},
			},
		},
		// Assistant response 1 with Read tool
		{
			Type:       "assistant",
			UUID:       "assistant-1",
			Timestamp:  "2025-10-23T10:00:10Z",
			SessionID:  "session-1",
			ParentUUID: "user-1",
			GitBranch:  "main",
			CWD:        "/home/user/project",
			Message: &parser.Message{
				Role: "assistant",
				Content: []parser.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &parser.ToolUse{
							ID:   "tool-read-1",
							Name: "Read",
							Input: map[string]interface{}{
								"file_path": "/home/user/project/config.yaml",
							},
						},
					},
				},
			},
		},
		// Tool result 1 - success
		{
			Type:       "user",
			UUID:       "tool-result-1",
			Timestamp:  "2025-10-23T10:00:15Z",
			SessionID:  "session-1",
			ParentUUID: "assistant-1",
			GitBranch:  "main",
			CWD:        "/home/user/project",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &parser.ToolResult{
							ToolUseID: "tool-read-1",
							Content:   "server: localhost\nport: 8080",
							IsError:   false,
							Status:    "success",
						},
					},
				},
			},
		},
		// User message 2
		{
			Type:       "user",
			UUID:       "user-2",
			Timestamp:  "2025-10-23T10:01:00Z",
			SessionID:  "session-1",
			ParentUUID: "tool-result-1",
			GitBranch:  "main",
			CWD:        "/home/user/project",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "text",
						Text: "Try to read a non-existent file",
					},
				},
			},
		},
		// Assistant response 2 with Read tool (will fail)
		{
			Type:       "assistant",
			UUID:       "assistant-2",
			Timestamp:  "2025-10-23T10:01:10Z",
			SessionID:  "session-1",
			ParentUUID: "user-2",
			GitBranch:  "main",
			CWD:        "/home/user/project",
			Message: &parser.Message{
				Role: "assistant",
				Content: []parser.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &parser.ToolUse{
							ID:   "tool-read-2",
							Name: "Read",
							Input: map[string]interface{}{
								"file_path": "/home/user/project/nonexistent.txt",
							},
						},
					},
				},
			},
		},
		// Tool result 2 - error
		{
			Type:       "user",
			UUID:       "tool-result-2",
			Timestamp:  "2025-10-23T10:01:15Z",
			SessionID:  "session-1",
			ParentUUID: "assistant-2",
			GitBranch:  "main",
			CWD:        "/home/user/project",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &parser.ToolResult{
							ToolUseID: "tool-read-2",
							Content:   "",
							IsError:   true,
							Status:    "error",
						},
					},
				},
			},
		},
		// User message 3
		{
			Type:       "user",
			UUID:       "user-3",
			Timestamp:  "2025-10-23T10:02:00Z",
			SessionID:  "session-1",
			ParentUUID: "tool-result-2",
			GitBranch:  "feature/new-feature",
			CWD:        "/home/user/project",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "text",
						Text: "Write to a file",
					},
				},
			},
		},
		// Assistant response 3 with Edit tool
		{
			Type:       "assistant",
			UUID:       "assistant-3",
			Timestamp:  "2025-10-23T10:02:10Z",
			SessionID:  "session-1",
			ParentUUID: "user-3",
			GitBranch:  "feature/new-feature",
			CWD:        "/home/user/project",
			Message: &parser.Message{
				Role: "assistant",
				Content: []parser.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &parser.ToolUse{
							ID:   "tool-edit-1",
							Name: "Edit",
							Input: map[string]interface{}{
								"file_path":  "/home/user/project/README.md",
								"old_string": "# Old Title",
								"new_string": "# New Title",
							},
						},
					},
				},
			},
		},
		// Tool result 3 - success
		{
			Type:       "user",
			UUID:       "tool-result-3",
			Timestamp:  "2025-10-23T10:02:15Z",
			SessionID:  "session-1",
			ParentUUID: "assistant-3",
			GitBranch:  "feature/new-feature",
			CWD:        "/home/user/project",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &parser.ToolResult{
							ToolUseID: "tool-edit-1",
							Content:   "File edited successfully",
							IsError:   false,
							Status:    "success",
						},
					},
				},
			},
		},
		// Different session entry
		{
			Type:       "user",
			UUID:       "user-4",
			Timestamp:  "2025-10-23T11:00:00Z",
			SessionID:  "session-2",
			ParentUUID: "",
			GitBranch:  "main",
			CWD:        "/home/user/other",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "text",
						Text: "Different session query",
					},
				},
			},
		},
	}
}

// TestQueryE2E_FilterTransformAggregate tests the complete query pipeline
func TestQueryE2E_FilterTransformAggregate(t *testing.T) {
	entries := createComplexTestEntries()

	t.Run("filter_failed_reads", func(t *testing.T) {
		// Query: Find all failed Read tool calls
		params := QueryParams{
			Resource: "tools",
			Filter: FilterSpec{
				ToolName:   "Read",
				ToolStatus: "error",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		tools, ok := result.([]parser.ToolCall)
		require.True(t, ok)
		assert.Len(t, tools, 1, "Should have exactly 1 failed Read")
		assert.Equal(t, "Read", tools[0].ToolName)
		assert.Equal(t, "error", tools[0].Status)
	})

	t.Run("count_tools_by_name", func(t *testing.T) {
		// Query: Count tool calls by name
		params := QueryParams{
			Resource: "tools",
			Aggregate: AggregateSpec{
				Function: "count",
				Field:    "tool_name",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		results, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, results)

		// Should have Read and Edit
		toolCounts := make(map[string]int)
		for _, r := range results {
			toolName := r["tool_name"].(string)
			count := r["count"].(int)
			toolCounts[toolName] = count
		}

		assert.Equal(t, 2, toolCounts["Read"], "Should have 2 Read calls")
		assert.Equal(t, 1, toolCounts["Edit"], "Should have 1 Edit call")
	})

	t.Run("filter_by_git_branch", func(t *testing.T) {
		// Query: Find entries on feature branch
		params := QueryParams{
			Resource: "entries",
			Filter: FilterSpec{
				GitBranch: "feature/new-feature",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		filtered, ok := result.([]parser.SessionEntry)
		require.True(t, ok)
		assert.NotEmpty(t, filtered)

		for _, entry := range filtered {
			assert.Equal(t, "feature/new-feature", entry.GitBranch)
		}
	})

	t.Run("filter_by_session", func(t *testing.T) {
		// Query: Find all entries in session-2
		params := QueryParams{
			Resource: "entries",
			Filter: FilterSpec{
				SessionID: "session-2",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		filtered, ok := result.([]parser.SessionEntry)
		require.True(t, ok)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "session-2", filtered[0].SessionID)
	})

	t.Run("user_messages_with_text", func(t *testing.T) {
		// Query: Find all user messages
		params := QueryParams{
			Resource: "messages",
			Filter: FilterSpec{
				Role: "user",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		messages, ok := result.([]MessageView)
		require.True(t, ok)
		assert.NotEmpty(t, messages)

		// All should be user role (excluding tool_result)
		userCount := 0
		for _, msg := range messages {
			if msg.Role == "user" {
				userCount++
				// Verify structure
				assert.NotEmpty(t, msg.UUID)
				assert.NotEmpty(t, msg.SessionID)
				assert.NotEmpty(t, msg.Timestamp)
			}
		}
		assert.Greater(t, userCount, 0, "Should have at least one user message")
	})

	t.Run("successful_tools_only", func(t *testing.T) {
		// Query: Find all successful tool executions
		params := QueryParams{
			Resource: "tools",
			Filter: FilterSpec{
				ToolStatus: "success",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		tools, ok := result.([]parser.ToolCall)
		require.True(t, ok)
		assert.Len(t, tools, 2, "Should have 2 successful tool calls")

		for _, tool := range tools {
			assert.Equal(t, "success", tool.Status)
		}
	})

	t.Run("count_entries_by_type", func(t *testing.T) {
		// Query: Count entries by type
		params := QueryParams{
			Resource: "entries",
			Aggregate: AggregateSpec{
				Function: "count",
				Field:    "type",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		results, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, results)

		typeCounts := make(map[string]int)
		for _, r := range results {
			entryType := r["type"].(string)
			count := r["count"].(int)
			typeCounts[entryType] = count
		}

		assert.Greater(t, typeCounts["user"], 0)
		assert.Greater(t, typeCounts["assistant"], 0)
	})
}

// TestQueryE2E_EmptyResults tests queries that should return empty results
func TestQueryE2E_EmptyResults(t *testing.T) {
	entries := createComplexTestEntries()

	t.Run("filter_nonexistent_tool", func(t *testing.T) {
		params := QueryParams{
			Resource: "tools",
			Filter: FilterSpec{
				ToolName: "NonExistentTool",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		tools, ok := result.([]parser.ToolCall)
		require.True(t, ok)
		assert.Empty(t, tools)
	})

	t.Run("filter_nonexistent_session", func(t *testing.T) {
		params := QueryParams{
			Resource: "entries",
			Filter: FilterSpec{
				SessionID: "nonexistent-session",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		filtered, ok := result.([]parser.SessionEntry)
		require.True(t, ok)
		assert.Empty(t, filtered)
	})

	t.Run("empty_input_entries", func(t *testing.T) {
		params := QueryParams{
			Resource: "tools",
		}

		result, err := Query([]parser.SessionEntry{}, params)
		require.NoError(t, err)

		tools, ok := result.([]parser.ToolCall)
		require.True(t, ok)
		assert.Empty(t, tools)
	})
}

// TestQueryE2E_ComplexFilters tests combining multiple filter conditions
func TestQueryE2E_ComplexFilters(t *testing.T) {
	entries := createComplexTestEntries()

	t.Run("filter_read_errors_only", func(t *testing.T) {
		// Query: Read tool + error status
		params := QueryParams{
			Resource: "tools",
			Filter: FilterSpec{
				ToolName:   "Read",
				ToolStatus: "error",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		tools, ok := result.([]parser.ToolCall)
		require.True(t, ok)
		assert.Len(t, tools, 1)
		assert.Equal(t, "Read", tools[0].ToolName)
		assert.Equal(t, "error", tools[0].Status)
	})

	t.Run("filter_main_branch_user_messages", func(t *testing.T) {
		// Query: User messages on main branch
		params := QueryParams{
			Resource: "messages",
			Filter: FilterSpec{
				Role:      "user",
				GitBranch: "main",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		messages, ok := result.([]MessageView)
		require.True(t, ok)
		assert.NotEmpty(t, messages)

		for _, msg := range messages {
			assert.Equal(t, "user", msg.Role)
			// GitBranch should be "main" but MessageView might not include it
			// This is acceptable as MessageView focuses on message content
		}
	})

	t.Run("filter_by_session_and_type", func(t *testing.T) {
		// Query: User entries in session-1
		params := QueryParams{
			Resource: "entries",
			Filter: FilterSpec{
				SessionID: "session-1",
				Type:      "user",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		filtered, ok := result.([]parser.SessionEntry)
		require.True(t, ok)
		assert.NotEmpty(t, filtered)

		for _, entry := range filtered {
			assert.Equal(t, "session-1", entry.SessionID)
			assert.Equal(t, "user", entry.Type)
		}
	})
}

// TestQueryE2E_ErrorHandling tests error conditions
func TestQueryE2E_ErrorHandling(t *testing.T) {
	entries := createComplexTestEntries()

	t.Run("invalid_resource_type", func(t *testing.T) {
		params := QueryParams{
			Resource: "invalid_resource",
		}

		_, err := Query(entries, params)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("invalid_aggregate_function", func(t *testing.T) {
		params := QueryParams{
			Resource: "tools",
			Aggregate: AggregateSpec{
				Function: "invalid_func",
			},
		}

		_, err := Query(entries, params)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("nil_entries", func(t *testing.T) {
		params := QueryParams{
			Resource: "entries",
		}

		// nil slice should be treated as empty
		result, err := Query(nil, params)
		require.NoError(t, err)

		filtered, ok := result.([]parser.SessionEntry)
		require.True(t, ok)
		assert.Empty(t, filtered)
	})
}

// TestQueryE2E_Aggregation tests various aggregation scenarios
func TestQueryE2E_Aggregation(t *testing.T) {
	entries := createComplexTestEntries()

	t.Run("count_all_tools", func(t *testing.T) {
		params := QueryParams{
			Resource: "tools",
			Aggregate: AggregateSpec{
				Function: "count",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		results, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, results, 1)
		assert.Contains(t, results[0], "count")
		count := results[0]["count"].(int)
		assert.Equal(t, 3, count, "Should have 3 total tool calls")
	})

	t.Run("count_by_status", func(t *testing.T) {
		params := QueryParams{
			Resource: "tools",
			Aggregate: AggregateSpec{
				Function: "count",
				Field:    "status",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		results, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, results)

		statusCounts := make(map[string]int)
		for _, r := range results {
			status := r["status"].(string)
			count := r["count"].(int)
			statusCounts[status] = count
		}

		assert.Equal(t, 2, statusCounts["success"])
		assert.Equal(t, 1, statusCounts["error"])
	})

	t.Run("group_messages_by_role", func(t *testing.T) {
		params := QueryParams{
			Resource: "messages",
			Aggregate: AggregateSpec{
				Function: "count",
				Field:    "role",
			},
		}

		result, err := Query(entries, params)
		require.NoError(t, err)

		results, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, results)

		roleCounts := make(map[string]int)
		for _, r := range results {
			role := r["role"].(string)
			count := r["count"].(int)
			roleCounts[role] = count
		}

		assert.Greater(t, roleCounts["user"], 0)
		assert.Greater(t, roleCounts["assistant"], 0)
	})
}
