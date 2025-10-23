package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/parser"
)

// Helper function to create test entries
func createTestEntries() []parser.SessionEntry {
	return []parser.SessionEntry{
		{
			Type:       "user",
			UUID:       "user-1",
			Timestamp:  "2025-10-23T00:00:00Z",
			SessionID:  "session-1",
			ParentUUID: "parent-1",
			GitBranch:  "main",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "text",
						Text: "Read the file",
					},
				},
			},
		},
		{
			Type:       "assistant",
			UUID:       "assistant-1",
			Timestamp:  "2025-10-23T00:01:00Z",
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
								"file_path": "/test/file.go",
							},
						},
					},
				},
			},
		},
		{
			Type:       "user",
			UUID:       "user-2",
			Timestamp:  "2025-10-23T00:02:00Z",
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
							Content:   "file content",
							IsError:   false,
							Status:    "success",
						},
					},
				},
			},
		},
	}
}

func TestSelectResource(t *testing.T) {
	entries := createTestEntries()

	tests := []struct {
		name         string
		resource     string
		wantType     string
		wantMinCount int
		wantErr      bool
	}{
		{
			name:         "select_entries",
			resource:     "entries",
			wantType:     "[]parser.SessionEntry",
			wantMinCount: 3,
			wantErr:      false,
		},
		{
			name:         "select_messages",
			resource:     "messages",
			wantType:     "[]MessageView",
			wantMinCount: 2, // 2 user messages, 1 assistant message with tool_use
			wantErr:      false,
		},
		{
			name:         "select_tools",
			resource:     "tools",
			wantType:     "[]parser.ToolCall",
			wantMinCount: 1, // 1 tool call (Read)
			wantErr:      false,
		},
		{
			name:     "invalid_resource",
			resource: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SelectResource(entries, tt.resource)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Check result count based on type
			var count int
			switch tt.resource {
			case "entries":
				entries, ok := result.([]parser.SessionEntry)
				require.True(t, ok, "Result should be []parser.SessionEntry")
				count = len(entries)
			case "messages":
				messages, ok := result.([]MessageView)
				require.True(t, ok, "Result should be []MessageView")
				count = len(messages)
			case "tools":
				tools, ok := result.([]parser.ToolCall)
				require.True(t, ok, "Result should be []parser.ToolCall")
				count = len(tools)
			}

			assert.GreaterOrEqual(t, count, tt.wantMinCount,
				"Expected at least %d results, got %d", tt.wantMinCount, count)
		})
	}
}

func TestExtractMessages(t *testing.T) {
	entries := createTestEntries()

	messages := extractMessages(entries)

	require.NotEmpty(t, messages)

	// Verify message structure
	for _, msg := range messages {
		assert.NotEmpty(t, msg.UUID)
		assert.NotEmpty(t, msg.SessionID)
		assert.NotEmpty(t, msg.Timestamp)
		assert.Contains(t, []string{"user", "assistant"}, msg.Role)
		assert.NotNil(t, msg.ContentBlocks)
	}

	// Verify we have both user and assistant messages
	var hasUser, hasAssistant bool
	for _, msg := range messages {
		if msg.Role == "user" {
			hasUser = true
		}
		if msg.Role == "assistant" {
			hasAssistant = true
		}
	}
	assert.True(t, hasUser, "Should have at least one user message")
	assert.True(t, hasAssistant, "Should have at least one assistant message")
}

func TestExtractToolExecutions(t *testing.T) {
	entries := createTestEntries()

	tools := extractToolExecutions(entries)

	require.NotEmpty(t, tools)

	// Verify tool execution structure
	for _, tool := range tools {
		assert.NotEmpty(t, tool.UUID)
		assert.NotEmpty(t, tool.ToolName)
		assert.NotEmpty(t, tool.Timestamp)
		assert.NotNil(t, tool.Input)
	}

	// Verify we have the Read tool
	var hasRead bool
	for _, tool := range tools {
		if tool.ToolName == "Read" {
			hasRead = true
			assert.Equal(t, "success", tool.Status)
			assert.Equal(t, "file content", tool.Output)
		}
	}
	assert.True(t, hasRead, "Should have Read tool execution")
}

func TestMessageView(t *testing.T) {
	entries := createTestEntries()
	messages := extractMessages(entries)

	require.NotEmpty(t, messages)

	// Test MessageView fields
	msg := messages[0]

	// Verify all required fields are populated
	assert.NotEmpty(t, msg.UUID, "UUID should not be empty")
	assert.NotEmpty(t, msg.SessionID, "SessionID should not be empty")
	assert.NotEmpty(t, msg.Timestamp, "Timestamp should not be empty")
	assert.NotEmpty(t, msg.Role, "Role should not be empty")
	assert.NotEmpty(t, msg.ContentBlocks, "ContentBlocks should not be empty")

	// Verify content is extracted
	if msg.Role == "user" {
		foundText := false
		for _, block := range msg.ContentBlocks {
			if block.Type == "text" && block.Text != "" {
				foundText = true
				break
			}
		}
		assert.True(t, foundText || len(msg.ContentBlocks) > 0,
			"User message should have text content or at least one content block")
	}
}

func TestExtractMessagesEmptyInput(t *testing.T) {
	messages := extractMessages([]parser.SessionEntry{})
	assert.Empty(t, messages, "Should return empty slice for empty input")
}

func TestExtractToolExecutionsEmptyInput(t *testing.T) {
	tools := extractToolExecutions([]parser.SessionEntry{})
	assert.Empty(t, tools, "Should return empty slice for empty input")
}

func TestExtractMessagesNoMessages(t *testing.T) {
	entries := []parser.SessionEntry{
		{
			Type:      "summary",
			UUID:      "summary-1",
			Timestamp: "2025-10-23T00:00:00Z",
			// No Message field
		},
	}

	messages := extractMessages(entries)
	assert.Empty(t, messages, "Should return empty slice when no message entries")
}
