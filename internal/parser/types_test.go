package parser

import (
	"encoding/json"
	"testing"
)

func TestSessionEntryUnmarshal_UserEntry(t *testing.T) {
	jsonData := `{
		"type":"user",
		"timestamp":"2025-10-02T06:07:13.673Z",
		"message":{
			"role":"user",
			"content":[{"type":"text","text":"帮我修复这个认证 bug"}]
		},
		"uuid":"cfef2966-a593-4169-9956-ee24c804b717",
		"parentUuid":null,
		"sessionId":"6a32f273-191a-49c8-a5fc-a5dcba08531a",
		"cwd":"/home/yale/work/meta-cc",
		"version":"2.0.1",
		"gitBranch":"develop"
	}`

	var entry SessionEntry
	err := json.Unmarshal([]byte(jsonData), &entry)

	if err != nil {
		t.Fatalf("Failed to unmarshal SessionEntry: %v", err)
	}

	if entry.Type != "user" {
		t.Errorf("Expected type 'user', got '%s'", entry.Type)
	}

	if entry.Timestamp != "2025-10-02T06:07:13.673Z" {
		t.Errorf("Expected timestamp '2025-10-02T06:07:13.673Z', got '%s'", entry.Timestamp)
	}

	if entry.UUID != "cfef2966-a593-4169-9956-ee24c804b717" {
		t.Errorf("Unexpected UUID: %s", entry.UUID)
	}

	if entry.SessionID != "6a32f273-191a-49c8-a5fc-a5dcba08531a" {
		t.Errorf("Unexpected SessionID: %s", entry.SessionID)
	}

	if entry.Message == nil {
		t.Fatal("Expected Message to be non-nil")
	}

	if entry.Message.Role != "user" {
		t.Errorf("Expected message role 'user', got '%s'", entry.Message.Role)
	}

	if len(entry.Message.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(entry.Message.Content))
	}

	if entry.Message.Content[0].Type != "text" {
		t.Errorf("Expected content type 'text', got '%s'", entry.Message.Content[0].Type)
	}

	if entry.Message.Content[0].Text != "帮我修复这个认证 bug" {
		t.Errorf("Unexpected text content: %s", entry.Message.Content[0].Text)
	}
}

func TestSessionEntryUnmarshal_AssistantWithToolUse(t *testing.T) {
	jsonData := `{
		"type":"assistant",
		"timestamp":"2025-10-02T06:08:57.769Z",
		"message":{
			"id":"msg_01J73XtFeXqDHHQZhYXBiVSr",
			"type":"message",
			"role":"assistant",
			"model":"claude-sonnet-4-5-20250929",
			"content":[
				{"type":"text","text":"我来帮你检查代码"},
				{"type":"tool_use","id":"toolu_01","name":"Grep","input":{"pattern":"auth.*error","path":"."}}
			],
			"stop_reason":"tool_use",
			"usage":{"input_tokens":100,"output_tokens":50}
		},
		"uuid":"0606832a-4c37-494b-a7c4-a10693086b86",
		"parentUuid":"cfef2966-a593-4169-9956-ee24c804b717",
		"sessionId":"6a32f273-191a-49c8-a5fc-a5dcba08531a"
	}`

	var entry SessionEntry
	err := json.Unmarshal([]byte(jsonData), &entry)

	if err != nil {
		t.Fatalf("Failed to unmarshal SessionEntry: %v", err)
	}

	if entry.Type != "assistant" {
		t.Errorf("Expected type 'assistant', got '%s'", entry.Type)
	}

	if entry.Message == nil {
		t.Fatal("Expected Message to be non-nil")
	}

	if entry.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", entry.Message.Role)
	}

	if entry.Message.Model != "claude-sonnet-4-5-20250929" {
		t.Errorf("Expected model 'claude-sonnet-4-5-20250929', got '%s'", entry.Message.Model)
	}

	if len(entry.Message.Content) != 2 {
		t.Fatalf("Expected 2 content blocks, got %d", len(entry.Message.Content))
	}

	// 验证第二个 block 是 tool_use
	toolBlock := entry.Message.Content[1]
	if toolBlock.Type != "tool_use" {
		t.Errorf("Expected type 'tool_use', got '%s'", toolBlock.Type)
	}

	if toolBlock.ToolUse == nil {
		t.Fatal("Expected ToolUse to be non-nil")
	}

	if toolBlock.ToolUse.ID != "toolu_01" {
		t.Errorf("Expected tool ID 'toolu_01', got '%s'", toolBlock.ToolUse.ID)
	}

	if toolBlock.ToolUse.Name != "Grep" {
		t.Errorf("Expected tool name 'Grep', got '%s'", toolBlock.ToolUse.Name)
	}

	// 验证 input
	pattern, ok := toolBlock.ToolUse.Input["pattern"].(string)
	if !ok || pattern != "auth.*error" {
		t.Errorf("Expected pattern 'auth.*error', got '%v'", pattern)
	}
}

func TestSessionEntryUnmarshal_ToolResult(t *testing.T) {
	jsonData := `{
		"type":"user",
		"timestamp":"2025-10-02T06:09:10.123Z",
		"message":{
			"role":"user",
			"content":[{"type":"tool_result","tool_use_id":"toolu_01","content":"src/auth.js:15: authError: token invalid"}]
		},
		"uuid":"abc123",
		"parentUuid":"0606832a-4c37-494b-a7c4-a10693086b86",
		"sessionId":"6a32f273-191a-49c8-a5fc-a5dcba08531a"
	}`

	var entry SessionEntry
	err := json.Unmarshal([]byte(jsonData), &entry)

	if err != nil {
		t.Fatalf("Failed to unmarshal SessionEntry: %v", err)
	}

	if entry.Message == nil {
		t.Fatal("Expected Message to be non-nil")
	}

	if len(entry.Message.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(entry.Message.Content))
	}

	resultBlock := entry.Message.Content[0]
	if resultBlock.Type != "tool_result" {
		t.Errorf("Expected type 'tool_result', got '%s'", resultBlock.Type)
	}

	if resultBlock.ToolResult == nil {
		t.Fatal("Expected ToolResult to be non-nil")
	}

	if resultBlock.ToolResult.ToolUseID != "toolu_01" {
		t.Errorf("Expected tool_use_id 'toolu_01', got '%s'", resultBlock.ToolResult.ToolUseID)
	}

	expectedContent := "src/auth.js:15: authError: token invalid"
	if resultBlock.ToolResult.Content != expectedContent {
		t.Errorf("Unexpected content: %s", resultBlock.ToolResult.Content)
	}
}

func TestSessionEntryUnmarshal_SkipNonMessageTypes(t *testing.T) {
	// 测试 file-history-snapshot 类型（应被忽略或标记）
	jsonData := `{
		"type":"file-history-snapshot",
		"messageId":"80d4a4d7-01c9-466f-83ea-f1f1498f1a6a",
		"snapshot":{"trackedFileBackups":{},"timestamp":"2025-10-02T06:07:13.675Z"},
		"isSnapshotUpdate":false
	}`

	var entry SessionEntry
	err := json.Unmarshal([]byte(jsonData), &entry)

	if err != nil {
		t.Fatalf("Failed to unmarshal SessionEntry: %v", err)
	}

	if entry.Type != "file-history-snapshot" {
		t.Errorf("Expected type 'file-history-snapshot', got '%s'", entry.Type)
	}

	// Message 应该为 nil，因为这不是消息类型
	if entry.Message != nil {
		t.Error("Expected Message to be nil for non-message type")
	}
}

func TestContentBlockUnmarshal_CustomUnmarshaler(t *testing.T) {
	// 测试自定义 UnmarshalJSON 是否正确处理不同类型
	testCases := []struct {
		name          string
		jsonData      string
		expectedType  string
		hasToolUse    bool
		hasToolResult bool
	}{
		{
			name:          "text content",
			jsonData:      `{"type":"text","text":"Hello"}`,
			expectedType:  "text",
			hasToolUse:    false,
			hasToolResult: false,
		},
		{
			name:          "tool_use content",
			jsonData:      `{"type":"tool_use","id":"t1","name":"Bash","input":{}}`,
			expectedType:  "tool_use",
			hasToolUse:    true,
			hasToolResult: false,
		},
		{
			name:          "tool_result content",
			jsonData:      `{"type":"tool_result","tool_use_id":"t1","content":"output"}`,
			expectedType:  "tool_result",
			hasToolUse:    false,
			hasToolResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var block ContentBlock
			err := json.Unmarshal([]byte(tc.jsonData), &block)

			if err != nil {
				t.Fatalf("Failed to unmarshal ContentBlock: %v", err)
			}

			if block.Type != tc.expectedType {
				t.Errorf("Expected type '%s', got '%s'", tc.expectedType, block.Type)
			}

			if tc.hasToolUse && block.ToolUse == nil {
				t.Error("Expected ToolUse to be non-nil")
			}

			if tc.hasToolResult && block.ToolResult == nil {
				t.Error("Expected ToolResult to be non-nil")
			}
		})
	}
}
