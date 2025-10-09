package parser

import (
	"encoding/json"
	"testing"
)

// TestMessageMarshalJSON tests basic message serialization
func TestMessageMarshalJSON(t *testing.T) {
	msg := &Message{
		ID:    "msg_123",
		Role:  "assistant",
		Model: "claude-3-5-sonnet",
		Content: []ContentBlock{
			{Type: "text", Text: "Hello world"},
		},
		StopReason: "end_turn",
		Usage:      map[string]interface{}{"input_tokens": 10, "output_tokens": 5},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal back to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify content field exists
	content, ok := result["content"]
	if !ok {
		t.Fatal("content field missing")
	}

	// Verify content is an array
	contentArr, ok := content.([]interface{})
	if !ok {
		t.Fatalf("content should be array, got %T", content)
	}

	if len(contentArr) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(contentArr))
	}
}

// TestMessageMarshalJSONEmpty tests empty content handling
func TestMessageMarshalJSONEmpty(t *testing.T) {
	msg := &Message{
		ID:      "msg_456",
		Role:    "user",
		Content: []ContentBlock{},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	content := result["content"]
	contentArr, ok := content.([]interface{})
	if !ok {
		t.Fatalf("content should be array, got %T", content)
	}

	if len(contentArr) != 0 {
		t.Fatalf("expected empty array, got %d items", len(contentArr))
	}
}

// TestContentBlockMarshalJSON tests all content block types
func TestContentBlockMarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		block ContentBlock
		check func(*testing.T, map[string]interface{})
	}{
		{
			name:  "text block",
			block: ContentBlock{Type: "text", Text: "Hello"},
			check: func(t *testing.T, result map[string]interface{}) {
				if result["type"] != "text" {
					t.Errorf("expected type=text, got %v", result["type"])
				}
				if result["text"] != "Hello" {
					t.Errorf("expected text=Hello, got %v", result["text"])
				}
			},
		},
		{
			name: "tool_use block",
			block: ContentBlock{
				Type: "tool_use",
				ToolUse: &ToolUse{
					ID:    "tool_123",
					Name:  "bash",
					Input: map[string]interface{}{"command": "ls"},
				},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				if result["type"] != "tool_use" {
					t.Errorf("expected type=tool_use, got %v", result["type"])
				}
				if result["id"] != "tool_123" {
					t.Errorf("expected id=tool_123, got %v", result["id"])
				}
				if result["name"] != "bash" {
					t.Errorf("expected name=bash, got %v", result["name"])
				}
			},
		},
		{
			name: "tool_result block",
			block: ContentBlock{
				Type: "tool_result",
				ToolResult: &ToolResult{
					ToolUseID: "tool_123",
					Content:   "output text",
					IsError:   false,
					Status:    "success",
				},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				if result["type"] != "tool_result" {
					t.Errorf("expected type=tool_result, got %v", result["type"])
				}
				if result["tool_use_id"] != "tool_123" {
					t.Errorf("expected tool_use_id=tool_123, got %v", result["tool_use_id"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.block)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			tt.check(t, result)
		})
	}
}

// TestContentBlockMarshalJSONNil tests nil field handling
func TestContentBlockMarshalJSONNil(t *testing.T) {
	// Tool use with nil ToolUse pointer should return error
	block := ContentBlock{Type: "tool_use", ToolUse: nil}
	_, err := json.Marshal(block)
	if err == nil {
		t.Error("expected error for nil ToolUse, got none")
	}

	// Tool result with nil ToolResult pointer should return error
	block2 := ContentBlock{Type: "tool_result", ToolResult: nil}
	_, err = json.Marshal(block2)
	if err == nil {
		t.Error("expected error for nil ToolResult, got none")
	}
}

// TestContentBlockMarshalJSONUnknownType tests unknown type handling
func TestContentBlockMarshalJSONUnknownType(t *testing.T) {
	// Unknown types should be marshaled with just the type field
	block := ContentBlock{Type: "thinking"}
	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("unexpected error for unknown type: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result["type"] != "thinking" {
		t.Errorf("expected type=thinking, got %v", result["type"])
	}

	// Should only have the type field
	if len(result) != 1 {
		t.Errorf("expected 1 field, got %d: %+v", len(result), result)
	}
}

// TestSerializationRoundTrip tests unmarshal → marshal → unmarshal consistency
func TestSerializationRoundTrip(t *testing.T) {
	original := &Message{
		ID:    "msg_789",
		Role:  "assistant",
		Model: "claude-3-5-sonnet",
		Content: []ContentBlock{
			{Type: "text", Text: "First text"},
			{
				Type: "tool_use",
				ToolUse: &ToolUse{
					ID:    "tool_456",
					Name:  "read",
					Input: map[string]interface{}{"file": "test.txt"},
				},
			},
			{
				Type: "tool_result",
				ToolResult: &ToolResult{
					ToolUseID: "tool_456",
					Content:   "file contents",
					IsError:   false,
				},
			},
			{Type: "text", Text: "Second text"},
		},
		StopReason: "end_turn",
		Usage:      map[string]interface{}{"input_tokens": 100, "output_tokens": 50},
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify key fields match
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %s, want %s", decoded.ID, original.ID)
	}
	if decoded.Role != original.Role {
		t.Errorf("Role mismatch: got %s, want %s", decoded.Role, original.Role)
	}
	if len(decoded.Content) != len(original.Content) {
		t.Fatalf("Content length mismatch: got %d, want %d", len(decoded.Content), len(original.Content))
	}

	// Verify content blocks
	for i, block := range decoded.Content {
		origBlock := original.Content[i]
		if block.Type != origBlock.Type {
			t.Errorf("Block %d type mismatch: got %s, want %s", i, block.Type, origBlock.Type)
		}

		switch block.Type {
		case "text":
			if block.Text != origBlock.Text {
				t.Errorf("Block %d text mismatch: got %s, want %s", i, block.Text, origBlock.Text)
			}
		case "tool_use":
			if block.ToolUse == nil {
				t.Errorf("Block %d ToolUse is nil", i)
				continue
			}
			if block.ToolUse.ID != origBlock.ToolUse.ID {
				t.Errorf("Block %d tool_use ID mismatch: got %s, want %s", i, block.ToolUse.ID, origBlock.ToolUse.ID)
			}
			if block.ToolUse.Name != origBlock.ToolUse.Name {
				t.Errorf("Block %d tool_use Name mismatch: got %s, want %s", i, block.ToolUse.Name, origBlock.ToolUse.Name)
			}
		case "tool_result":
			if block.ToolResult == nil {
				t.Errorf("Block %d ToolResult is nil", i)
				continue
			}
			if block.ToolResult.ToolUseID != origBlock.ToolResult.ToolUseID {
				t.Errorf("Block %d tool_result ToolUseID mismatch: got %s, want %s", i, block.ToolResult.ToolUseID, origBlock.ToolResult.ToolUseID)
			}
			if block.ToolResult.Content != origBlock.ToolResult.Content {
				t.Errorf("Block %d tool_result Content mismatch: got %s, want %s", i, block.ToolResult.Content, origBlock.ToolResult.Content)
			}
		}
	}
}
