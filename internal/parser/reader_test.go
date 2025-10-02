package parser

import (
	"testing"

	"github.com/yale/meta-cc/internal/testutil"
)

func TestParseSession_ValidFile(t *testing.T) {
	// 使用测试 fixture
	filePath := testutil.FixtureDir() + "/sample-session.jsonl"

	parser := NewSessionParser(filePath)
	turns, err := parser.ParseTurns()

	if err != nil {
		t.Fatalf("Failed to parse session: %v", err)
	}

	expectedTurns := 3
	if len(turns) != expectedTurns {
		t.Errorf("Expected %d turns, got %d", expectedTurns, len(turns))
	}

	// 验证第一个 turn（user）
	turn0 := turns[0]
	if turn0.Role != "user" {
		t.Errorf("Expected turn 0 role 'user', got '%s'", turn0.Role)
	}
	if turn0.Sequence != 0 {
		t.Errorf("Expected turn 0 sequence 0, got %d", turn0.Sequence)
	}

	// 验证第二个 turn（assistant with tool）
	turn1 := turns[1]
	if turn1.Role != "assistant" {
		t.Errorf("Expected turn 1 role 'assistant', got '%s'", turn1.Role)
	}
	if len(turn1.Content) != 2 {
		t.Errorf("Expected 2 content blocks in turn 1, got %d", len(turn1.Content))
	}

	// 验证工具调用
	hasToolUse := false
	for _, block := range turn1.Content {
		if block.Type == "tool_use" && block.ToolUse != nil {
			hasToolUse = true
			if block.ToolUse.Name != "Grep" {
				t.Errorf("Expected tool name 'Grep', got '%s'", block.ToolUse.Name)
			}
		}
	}
	if !hasToolUse {
		t.Error("Expected tool_use in turn 1")
	}

	// 验证第三个 turn（tool result）
	turn2 := turns[2]
	if turn2.Role != "user" {
		t.Errorf("Expected turn 2 role 'user', got '%s'", turn2.Role)
	}
	if len(turn2.Content) < 1 {
		t.Fatal("Expected at least 1 content block in turn 2")
	}
	if turn2.Content[0].Type != "tool_result" {
		t.Errorf("Expected type 'tool_result', got '%s'", turn2.Content[0].Type)
	}
}

func TestParseSession_EmptyFile(t *testing.T) {
	tempFile := testutil.TempSessionFile(t, "")

	parser := NewSessionParser(tempFile)
	turns, err := parser.ParseTurns()

	if err != nil {
		t.Fatalf("Expected no error for empty file, got: %v", err)
	}

	if len(turns) != 0 {
		t.Errorf("Expected 0 turns for empty file, got %d", len(turns))
	}
}

func TestParseSession_InvalidJSON(t *testing.T) {
	content := `{"sequence":0,"role":"user","timestamp":1735689600,"content":[]}
invalid json line
{"sequence":1,"role":"assistant","timestamp":1735689605,"content":[]}`

	tempFile := testutil.TempSessionFile(t, content)

	parser := NewSessionParser(tempFile)
	_, err := parser.ParseTurns()

	if err == nil {
		t.Error("Expected error for invalid JSON line")
	}
}

func TestParseSession_SkipEmptyLines(t *testing.T) {
	content := `{"sequence":0,"role":"user","timestamp":1735689600,"content":[]}

{"sequence":1,"role":"assistant","timestamp":1735689605,"content":[]}

`

	tempFile := testutil.TempSessionFile(t, content)

	parser := NewSessionParser(tempFile)
	turns, err := parser.ParseTurns()

	if err != nil {
		t.Fatalf("Failed to parse session with empty lines: %v", err)
	}

	if len(turns) != 2 {
		t.Errorf("Expected 2 turns (empty lines skipped), got %d", len(turns))
	}
}

func TestParseSession_FileNotFound(t *testing.T) {
	parser := NewSessionParser("/nonexistent/file.jsonl")
	_, err := parser.ParseTurns()

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseSession_MissingRequiredFields(t *testing.T) {
	// 测试缺少必需字段的 JSON（如缺少 sequence）
	content := `{"role":"user","timestamp":1735689600,"content":[]}`

	tempFile := testutil.TempSessionFile(t, content)

	parser := NewSessionParser(tempFile)
	turns, err := parser.ParseTurns()

	// 应该能解析，但 sequence 会是零值
	if err != nil {
		t.Fatalf("Expected no error (zero value for missing field), got: %v", err)
	}

	if len(turns) != 1 {
		t.Fatalf("Expected 1 turn, got %d", len(turns))
	}

	if turns[0].Sequence != 0 {
		t.Errorf("Expected sequence 0 (zero value), got %d", turns[0].Sequence)
	}
}
