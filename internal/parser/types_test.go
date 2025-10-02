package parser

import (
	"encoding/json"
	"testing"
)

func TestTurnUnmarshal_UserTurn(t *testing.T) {
	jsonData := `{"sequence":0,"role":"user","timestamp":1735689600,"content":[{"type":"text","text":"帮我修复这个认证 bug"}]}`

	var turn Turn
	err := json.Unmarshal([]byte(jsonData), &turn)

	if err != nil {
		t.Fatalf("Failed to unmarshal Turn: %v", err)
	}

	if turn.Sequence != 0 {
		t.Errorf("Expected sequence 0, got %d", turn.Sequence)
	}

	if turn.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", turn.Role)
	}

	if turn.Timestamp != 1735689600 {
		t.Errorf("Expected timestamp 1735689600, got %d", turn.Timestamp)
	}

	if len(turn.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(turn.Content))
	}

	if turn.Content[0].Type != "text" {
		t.Errorf("Expected content type 'text', got '%s'", turn.Content[0].Type)
	}

	if turn.Content[0].Text != "帮我修复这个认证 bug" {
		t.Errorf("Unexpected text content: %s", turn.Content[0].Text)
	}
}

func TestTurnUnmarshal_AssistantWithToolUse(t *testing.T) {
	jsonData := `{"sequence":1,"role":"assistant","timestamp":1735689605,"content":[{"type":"text","text":"我来帮你检查代码"},{"type":"tool_use","id":"toolu_01","name":"Grep","input":{"pattern":"auth.*error","path":"."}}]}`

	var turn Turn
	err := json.Unmarshal([]byte(jsonData), &turn)

	if err != nil {
		t.Fatalf("Failed to unmarshal Turn: %v", err)
	}

	if turn.Sequence != 1 {
		t.Errorf("Expected sequence 1, got %d", turn.Sequence)
	}

	if turn.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", turn.Role)
	}

	if len(turn.Content) != 2 {
		t.Fatalf("Expected 2 content blocks, got %d", len(turn.Content))
	}

	// 验证第二个 block 是 tool_use
	toolBlock := turn.Content[1]
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

func TestTurnUnmarshal_ToolResult(t *testing.T) {
	jsonData := `{"sequence":2,"role":"user","timestamp":1735689610,"content":[{"type":"tool_result","tool_use_id":"toolu_01","content":"src/auth.js:15: authError: token invalid"}]}`

	var turn Turn
	err := json.Unmarshal([]byte(jsonData), &turn)

	if err != nil {
		t.Fatalf("Failed to unmarshal Turn: %v", err)
	}

	if len(turn.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(turn.Content))
	}

	resultBlock := turn.Content[0]
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

func TestContentBlockUnmarshal_CustomUnmarshaler(t *testing.T) {
	// 测试自定义 UnmarshalJSON 是否正确处理不同类型
	testCases := []struct {
		name        string
		jsonData    string
		expectedType string
		hasToolUse  bool
		hasToolResult bool
	}{
		{
			name:        "text content",
			jsonData:    `{"type":"text","text":"Hello"}`,
			expectedType: "text",
			hasToolUse:  false,
			hasToolResult: false,
		},
		{
			name:        "tool_use content",
			jsonData:    `{"type":"tool_use","id":"t1","name":"Bash","input":{}}`,
			expectedType: "tool_use",
			hasToolUse:  true,
			hasToolResult: false,
		},
		{
			name:        "tool_result content",
			jsonData:    `{"type":"tool_result","tool_use_id":"t1","content":"output"}`,
			expectedType: "tool_result",
			hasToolUse:  false,
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
