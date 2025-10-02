package parser

import (
	"testing"

	"github.com/yale/meta-cc/internal/testutil"
)

func TestParseSession_ValidFile(t *testing.T) {
	// 使用测试 fixture（包含 4 行：1 个 file-history-snapshot + 3 个消息）
	filePath := testutil.FixtureDir() + "/sample-session.jsonl"

	parser := NewSessionParser(filePath)
	entries, err := parser.ParseEntries()

	if err != nil {
		t.Fatalf("Failed to parse session: %v", err)
	}

	// 应该只返回消息类型（过滤掉 file-history-snapshot）
	expectedEntries := 3
	if len(entries) != expectedEntries {
		t.Errorf("Expected %d message entries, got %d", expectedEntries, len(entries))
	}

	// 验证第一个条目（user）
	entry0 := entries[0]
	if entry0.Type != "user" {
		t.Errorf("Expected type 'user', got '%s'", entry0.Type)
	}
	if entry0.UUID != "cfef2966-a593-4169-9956-ee24c804b717" {
		t.Errorf("Unexpected UUID: %s", entry0.UUID)
	}
	if entry0.Message == nil {
		t.Fatal("Expected Message to be non-nil")
	}
	if entry0.Message.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", entry0.Message.Role)
	}

	// 验证第二个条目（assistant with tool）
	entry1 := entries[1]
	if entry1.Type != "assistant" {
		t.Errorf("Expected type 'assistant', got '%s'", entry1.Type)
	}
	if entry1.Message == nil {
		t.Fatal("Expected Message to be non-nil")
	}
	if len(entry1.Message.Content) != 2 {
		t.Errorf("Expected 2 content blocks, got %d", len(entry1.Message.Content))
	}

	// 验证工具调用
	hasToolUse := false
	for _, block := range entry1.Message.Content {
		if block.Type == "tool_use" && block.ToolUse != nil {
			hasToolUse = true
			if block.ToolUse.Name != "Grep" {
				t.Errorf("Expected tool name 'Grep', got '%s'", block.ToolUse.Name)
			}
		}
	}
	if !hasToolUse {
		t.Error("Expected tool_use in entry 1")
	}

	// 验证第三个条目（tool result）
	entry2 := entries[2]
	if entry2.Type != "user" {
		t.Errorf("Expected type 'user', got '%s'", entry2.Type)
	}
	if entry2.Message == nil {
		t.Fatal("Expected Message to be non-nil")
	}
	if len(entry2.Message.Content) < 1 {
		t.Fatal("Expected at least 1 content block")
	}
	if entry2.Message.Content[0].Type != "tool_result" {
		t.Errorf("Expected type 'tool_result', got '%s'", entry2.Message.Content[0].Type)
	}
}

func TestParseSession_EmptyFile(t *testing.T) {
	tempFile := testutil.TempSessionFile(t, "")

	parser := NewSessionParser(tempFile)
	entries, err := parser.ParseEntries()

	if err != nil {
		t.Fatalf("Expected no error for empty file, got: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected 0 entries for empty file, got %d", len(entries))
	}
}

func TestParseSession_InvalidJSON(t *testing.T) {
	content := `{"type":"user","timestamp":"2025-10-02T06:07:13.673Z","message":{"role":"user","content":[]},"uuid":"abc"}
invalid json line
{"type":"assistant","timestamp":"2025-10-02T06:08:57.769Z","message":{"role":"assistant","content":[]},"uuid":"def"}`

	tempFile := testutil.TempSessionFile(t, content)

	parser := NewSessionParser(tempFile)
	_, err := parser.ParseEntries()

	if err == nil {
		t.Error("Expected error for invalid JSON line")
	}
}

func TestParseSession_SkipEmptyLines(t *testing.T) {
	content := `{"type":"user","timestamp":"2025-10-02T06:07:13.673Z","message":{"role":"user","content":[]},"uuid":"abc"}

{"type":"assistant","timestamp":"2025-10-02T06:08:57.769Z","message":{"role":"assistant","content":[]},"uuid":"def"}

`

	tempFile := testutil.TempSessionFile(t, content)

	parser := NewSessionParser(tempFile)
	entries, err := parser.ParseEntries()

	if err != nil {
		t.Fatalf("Failed to parse session with empty lines: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries (empty lines skipped), got %d", len(entries))
	}
}

func TestParseSession_FileNotFound(t *testing.T) {
	parser := NewSessionParser("/nonexistent/file.jsonl")
	_, err := parser.ParseEntries()

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseSession_FilterNonMessageTypes(t *testing.T) {
	// 测试过滤非消息类型（如 file-history-snapshot）
	content := `{"type":"file-history-snapshot","messageId":"abc","snapshot":{}}
{"type":"user","timestamp":"2025-10-02T06:07:13.673Z","message":{"role":"user","content":[]},"uuid":"user1"}
{"type":"some-other-type","data":"ignored"}
{"type":"assistant","timestamp":"2025-10-02T06:08:57.769Z","message":{"role":"assistant","content":[]},"uuid":"asst1"}`

	tempFile := testutil.TempSessionFile(t, content)

	parser := NewSessionParser(tempFile)
	entries, err := parser.ParseEntries()

	// 应该只返回 user 和 assistant 类型
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 message entries (non-message types filtered), got %d", len(entries))
	}

	// 验证都是消息类型
	for _, entry := range entries {
		if !entry.IsMessage() {
			t.Errorf("Expected only message types, got '%s'", entry.Type)
		}
	}
}
