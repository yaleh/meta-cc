package analyzer

import "github.com/yaleh/meta-cc/internal/types"

// makeToolCalls creates ToolCall fixtures directly
func makeToolCalls(toolName string, status string, errMsg string) []types.ToolCall {
	return []types.ToolCall{
		{
			UUID:      "test-uuid-1",
			ToolName:  toolName,
			Status:    status,
			Error:     errMsg,
			Timestamp: "2025-10-02T10:00:00.000Z",
		},
	}
}

// makeEntry creates a single SessionEntry fixture for testing
func makeEntry(uuid string, timestamp string) types.SessionEntry {
	return types.SessionEntry{
		Type:      "assistant",
		UUID:      uuid,
		Timestamp: timestamp,
	}
}
