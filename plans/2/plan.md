# Phase 2: JSONL 解析器

## 概述

**目标**: 解析 Claude Code 会话文件的 JSONL 格式，提取 Turn 数据和工具调用信息

**代码量**: ~210 行（每个 Stage ≤ 200 行）

**依赖**: Phase 0（CLI 框架和测试工具）、Phase 1（会话文件定位）

**交付物**: 完整的 JSONL 解析器，支持 Turn 提取和 Tool Call 匹配

---

## Phase 目标

实现 JSONL 格式解析器，支持：

1. Turn 数据结构定义（Sequence, Role, Timestamp, Content）
2. ContentBlock 类型解析（text, tool_use, tool_result）
3. JSONL 文件逐行读取和解析
4. 工具调用提取（ToolUse + ToolResult 匹配）
5. 错误处理（非法 JSON、缺失字段）

**成功标准**:
- ✅ 能解析真实的 Claude Code 会话文件
- ✅ 正确提取所有 Turn 数据
- ✅ 正确匹配 ToolUse 和 ToolResult
- ✅ 处理边界情况（空行、非法 JSON、缺失字段）
- ✅ 所有单元测试通过
- ✅ README.md 包含解析器使用说明

---

## JSONL 数据结构

### Claude Code 会话文件格式

```jsonl
{"sequence":0,"role":"user","timestamp":1735689600,"content":[{"type":"text","text":"帮我修复这个认证 bug"}]}
{"sequence":1,"role":"assistant","timestamp":1735689605,"content":[{"type":"text","text":"我来帮你检查代码"},{"type":"tool_use","id":"toolu_01","name":"Grep","input":{"pattern":"auth.*error","path":"."}}]}
{"sequence":2,"role":"user","timestamp":1735689610,"content":[{"type":"tool_result","tool_use_id":"toolu_01","content":"src/auth.js:15: authError: token invalid"}]}
```

### 数据结构关系图

```plantuml
@startuml
!theme plain

package "解析流程" {
  [JSONL 文件] as File
  [逐行读取] as Reader
  [JSON 解析] as Parser
  [Turn 数据结构] as Turn
  [Tool Call 提取] as Tool

  File --> Reader
  Reader --> Parser
  Parser --> Turn
  Turn --> Tool
}

package "数据结构" {
  class Turn {
    Sequence int
    Role string
    Timestamp int64
    Content []ContentBlock
  }

  class ContentBlock {
    Type string
    Text string
    ToolUse *ToolUse
    ToolResult *ToolResult
  }

  class ToolUse {
    ID string
    Name string
    Input map[string]interface{}
  }

  class ToolResult {
    ToolUseID string
    Content string
    Status string
    Error string
  }

  class ToolCall {
    TurnSequence int
    ToolName string
    Input map[string]interface{}
    Output string
    Status string
    Error string
  }
}

Turn --> ContentBlock
ContentBlock --> ToolUse
ContentBlock --> ToolResult

@enduml
```

---

## Stage 2.1: 数据结构定义

### 目标

定义 Turn、ContentBlock、ToolUse、ToolResult 数据结构，支持 JSON 序列化和反序列化。

### TDD 工作流

**1. 准备阶段**

```bash
# 创建 parser 包目录
mkdir -p internal/parser
```

**2. 测试先行 - 编写测试**

#### `internal/parser/types_test.go` (~90 行)

```go
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
```

**3. 实现代码**

#### `internal/parser/types.go` (~100 行)

```go
package parser

import (
	"encoding/json"
	"fmt"
)

// Turn 表示一个会话轮次
type Turn struct {
	Sequence  int            `json:"sequence"`
	Role      string         `json:"role"`
	Timestamp int64          `json:"timestamp"`
	Content   []ContentBlock `json:"content"`
}

// ContentBlock 表示 Turn 中的一个内容块
// 可以是文本、工具调用或工具结果
type ContentBlock struct {
	Type       string      `json:"type"`
	Text       string      `json:"text,omitempty"`
	ToolUse    *ToolUse    `json:"-"` // 手动处理序列化
	ToolResult *ToolResult `json:"-"` // 手动处理序列化
}

// ToolUse 表示一个工具调用
type ToolUse struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ToolResult 表示工具调用的结果
type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	Status    string `json:"status,omitempty"`
	Error     string `json:"error,omitempty"`
}

// UnmarshalJSON 自定义 ContentBlock 的反序列化逻辑
// 根据 type 字段，解析不同的内容到相应的字段
func (cb *ContentBlock) UnmarshalJSON(data []byte) error {
	// 先解析通用字段
	type Alias ContentBlock
	aux := &struct {
		*Alias
		RawToolUse    json.RawMessage `json:"tool_use,omitempty"`
		RawToolResult json.RawMessage `json:"tool_result,omitempty"`
	}{
		Alias: (*Alias)(cb),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal ContentBlock: %w", err)
	}

	// 根据 type 解析特定字段
	switch cb.Type {
	case "text":
		// text 类型已经由默认反序列化处理

	case "tool_use":
		// 解析 tool_use 字段
		var toolUse ToolUse
		// tool_use 数据直接嵌入在 ContentBlock 中（除了 type）
		// 需要重新解析整个 data
		type ToolUseBlock struct {
			Type  string                 `json:"type"`
			ID    string                 `json:"id"`
			Name  string                 `json:"name"`
			Input map[string]interface{} `json:"input"`
		}
		var tub ToolUseBlock
		if err := json.Unmarshal(data, &tub); err != nil {
			return fmt.Errorf("failed to unmarshal tool_use: %w", err)
		}
		toolUse.ID = tub.ID
		toolUse.Name = tub.Name
		toolUse.Input = tub.Input
		cb.ToolUse = &toolUse

	case "tool_result":
		// 解析 tool_result 字段
		var toolResult ToolResult
		type ToolResultBlock struct {
			Type      string `json:"type"`
			ToolUseID string `json:"tool_use_id"`
			Content   string `json:"content"`
			Status    string `json:"status,omitempty"`
			Error     string `json:"error,omitempty"`
		}
		var trb ToolResultBlock
		if err := json.Unmarshal(data, &trb); err != nil {
			return fmt.Errorf("failed to unmarshal tool_result: %w", err)
		}
		toolResult.ToolUseID = trb.ToolUseID
		toolResult.Content = trb.Content
		toolResult.Status = trb.Status
		toolResult.Error = trb.Error
		cb.ToolResult = &toolResult

	default:
		// 未知类型，保留原始数据但不报错
	}

	return nil
}
```

**4. 运行测试**

```bash
# 运行 parser 包测试
go test ./internal/parser -v

# 预期输出：
# === RUN   TestTurnUnmarshal_UserTurn
# --- PASS: TestTurnUnmarshal_UserTurn (0.00s)
# === RUN   TestTurnUnmarshal_AssistantWithToolUse
# --- PASS: TestTurnUnmarshal_AssistantWithToolUse (0.00s)
# === RUN   TestTurnUnmarshal_ToolResult
# --- PASS: TestTurnUnmarshal_ToolResult (0.00s)
# === RUN   TestContentBlockUnmarshal_CustomUnmarshaler
# --- PASS: TestContentBlockUnmarshal_CustomUnmarshaler (0.00s)
# PASS
```

### 交付物

**文件清单**:
```
meta-cc/
├── internal/
│   └── parser/
│       ├── types.go          # 数据结构定义 (~100 行)
│       └── types_test.go     # 单元测试 (~90 行)
```

**代码量**: ~190 行

### 验收标准

- ✅ `TestTurnUnmarshal_UserTurn` 测试通过（用户 Turn）
- ✅ `TestTurnUnmarshal_AssistantWithToolUse` 测试通过（助手 Turn + 工具调用）
- ✅ `TestTurnUnmarshal_ToolResult` 测试通过（工具结果）
- ✅ `TestContentBlockUnmarshal_CustomUnmarshaler` 测试通过（自定义反序列化）
- ✅ 所有测试无警告或失败
- ✅ 代码符合 Go 命名规范（导出类型有注释）

---

## Stage 2.2: JSONL 读取器

### 目标

实现 JSONL 文件逐行读取，解析为 Turn 数组，处理空行和非法 JSON。

### TDD 工作流

**1. 测试先行 - 编写测试**

#### `internal/parser/reader_test.go` (~110 行)

```go
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
```

**2. 实现代码**

#### `internal/parser/reader.go` (~80 行)

```go
package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// SessionParser 负责解析 Claude Code 会话文件
type SessionParser struct {
	filePath string
}

// NewSessionParser 创建 SessionParser 实例
func NewSessionParser(filePath string) *SessionParser {
	return &SessionParser{
		filePath: filePath,
	}
}

// ParseTurns 解析 JSONL 文件，返回 Turn 数组
// JSONL 格式：每行一个 JSON 对象
// 处理规则：
//   - 跳过空行和空白行
//   - 非法 JSON 行返回错误
//   - 返回所有成功解析的 Turn
func (p *SessionParser) ParseTurns() ([]Turn, error) {
	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open session file: %w", err)
	}
	defer file.Close()

	var turns []Turn
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// 跳过空行和仅包含空白的行
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 解析 JSON 为 Turn
		var turn Turn
		if err := json.Unmarshal([]byte(line), &turn); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}

		turns = append(turns, turn)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading session file: %w", err)
	}

	return turns, nil
}

// ParseTurnsFromContent 从字符串内容解析 JSONL（用于测试）
func ParseTurnsFromContent(content string) ([]Turn, error) {
	var turns []Turn
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		// 跳过空行
		if strings.TrimSpace(line) == "" {
			continue
		}

		var turn Turn
		if err := json.Unmarshal([]byte(line), &turn); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum+1, err)
		}

		turns = append(turns, turn)
	}

	return turns, nil
}
```

**3. 运行测试**

```bash
# 运行所有 parser 测试
go test ./internal/parser -v

# 预期：所有测试通过
```

### 交付物

**文件清单**:
```
meta-cc/
├── internal/
│   └── parser/
│       ├── types.go          # 数据结构（Stage 2.1）
│       ├── types_test.go     # 数据结构测试（Stage 2.1）
│       ├── reader.go         # JSONL 读取器 (~80 行)
│       └── reader_test.go    # 读取器测试 (~110 行)
```

**代码量**: ~190 行（本 Stage）

### 验收标准

- ✅ `TestParseSession_ValidFile` 测试通过（解析真实 fixture）
- ✅ `TestParseSession_EmptyFile` 测试通过（空文件）
- ✅ `TestParseSession_InvalidJSON` 测试通过（非法 JSON 返回错误）
- ✅ `TestParseSession_SkipEmptyLines` 测试通过（跳过空行）
- ✅ `TestParseSession_FileNotFound` 测试通过（文件不存在）
- ✅ `TestParseSession_MissingRequiredFields` 测试通过（缺失字段处理）
- ✅ `go test ./internal/parser -v` 全部通过

---

## Stage 2.3: 工具调用提取

### 目标

从 Turn 数组中提取工具调用，匹配 ToolUse 和 ToolResult，生成 ToolCall 结构。

### TDD 工作流

**1. 测试先行 - 编写测试**

#### `internal/parser/tools_test.go` (~100 行)

```go
package parser

import (
	"testing"
)

func TestExtractToolCalls_SingleCall(t *testing.T) {
	turns := []Turn{
		{
			Sequence: 1,
			Role:     "assistant",
			Content: []ContentBlock{
				{Type: "text", Text: "检查代码"},
				{
					Type: "tool_use",
					ToolUse: &ToolUse{
						ID:   "toolu_01",
						Name: "Grep",
						Input: map[string]interface{}{
							"pattern": "auth.*error",
						},
					},
				},
			},
		},
		{
			Sequence: 2,
			Role:     "user",
			Content: []ContentBlock{
				{
					Type: "tool_result",
					ToolResult: &ToolResult{
						ToolUseID: "toolu_01",
						Content:   "auth.js:15: authError",
					},
				},
			},
		},
	}

	toolCalls := ExtractToolCalls(turns)

	if len(toolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(toolCalls))
	}

	tc := toolCalls[0]
	if tc.ToolName != "Grep" {
		t.Errorf("Expected tool name 'Grep', got '%s'", tc.ToolName)
	}

	if tc.TurnSequence != 1 {
		t.Errorf("Expected turn sequence 1, got %d", tc.TurnSequence)
	}

	if tc.Output != "auth.js:15: authError" {
		t.Errorf("Unexpected output: %s", tc.Output)
	}

	pattern, ok := tc.Input["pattern"].(string)
	if !ok || pattern != "auth.*error" {
		t.Errorf("Expected pattern 'auth.*error', got '%v'", pattern)
	}
}

func TestExtractToolCalls_MultipleCallsSameTurn(t *testing.T) {
	turns := []Turn{
		{
			Sequence: 1,
			Role:     "assistant",
			Content: []ContentBlock{
				{
					Type: "tool_use",
					ToolUse: &ToolUse{
						ID:   "tool_1",
						Name: "Read",
						Input: map[string]interface{}{"file": "a.txt"},
					},
				},
				{
					Type: "tool_use",
					ToolUse: &ToolUse{
						ID:   "tool_2",
						Name: "Grep",
						Input: map[string]interface{}{"pattern": "error"},
					},
				},
			},
		},
		{
			Sequence: 2,
			Role:     "user",
			Content: []ContentBlock{
				{
					Type: "tool_result",
					ToolResult: &ToolResult{
						ToolUseID: "tool_1",
						Content:   "file content",
					},
				},
				{
					Type: "tool_result",
					ToolResult: &ToolResult{
						ToolUseID: "tool_2",
						Content:   "match found",
					},
				},
			},
		},
	}

	toolCalls := ExtractToolCalls(turns)

	if len(toolCalls) != 2 {
		t.Fatalf("Expected 2 tool calls, got %d", len(toolCalls))
	}

	// 验证都被正确匹配
	for _, tc := range toolCalls {
		if tc.Output == "" {
			t.Errorf("Tool call %s has empty output", tc.ToolName)
		}
	}
}

func TestExtractToolCalls_UnmatchedToolUse(t *testing.T) {
	turns := []Turn{
		{
			Sequence: 1,
			Role:     "assistant",
			Content: []ContentBlock{
				{
					Type: "tool_use",
					ToolUse: &ToolUse{
						ID:   "orphan_tool",
						Name: "Bash",
						Input: map[string]interface{}{},
					},
				},
			},
		},
		// 没有对应的 tool_result
	}

	toolCalls := ExtractToolCalls(turns)

	if len(toolCalls) != 1 {
		t.Fatalf("Expected 1 tool call (unmatched), got %d", len(toolCalls))
	}

	tc := toolCalls[0]
	if tc.Output != "" {
		t.Errorf("Expected empty output for unmatched tool, got '%s'", tc.Output)
	}

	if tc.Status != "" {
		t.Errorf("Expected empty status for unmatched tool, got '%s'", tc.Status)
	}
}

func TestExtractToolCalls_NoToolCalls(t *testing.T) {
	turns := []Turn{
		{
			Sequence: 0,
			Role:     "user",
			Content: []ContentBlock{
				{Type: "text", Text: "Hello"},
			},
		},
		{
			Sequence: 1,
			Role:     "assistant",
			Content: []ContentBlock{
				{Type: "text", Text: "Hi there"},
			},
		},
	}

	toolCalls := ExtractToolCalls(turns)

	if len(toolCalls) != 0 {
		t.Errorf("Expected 0 tool calls, got %d", len(toolCalls))
	}
}
```

**2. 实现代码**

#### `internal/parser/tools.go` (~70 行)

```go
package parser

// ToolCall 表示一个完整的工具调用（ToolUse + ToolResult）
type ToolCall struct {
	TurnSequence int                    // 工具调用所在的 Turn 序号
	ToolName     string                 // 工具名称
	Input        map[string]interface{} // 工具输入参数
	Output       string                 // 工具输出（ToolResult.Content）
	Status       string                 // 执行状态（success/error）
	Error        string                 // 错误信息（如果有）
}

// ExtractToolCalls 从 Turn 数组中提取所有工具调用
// 流程：
//  1. 遍历所有 Turn，收集 ToolUse（按 ID 索引）
//  2. 遍历所有 Turn，查找 ToolResult，匹配 tool_use_id
//  3. 生成 ToolCall 数组
func ExtractToolCalls(turns []Turn) []ToolCall {
	// Step 1: 收集所有 ToolUse（按 ID 索引）
	toolUseMap := make(map[string]struct {
		turnSeq int
		toolUse *ToolUse
	})

	for _, turn := range turns {
		for _, block := range turn.Content {
			if block.Type == "tool_use" && block.ToolUse != nil {
				toolUseMap[block.ToolUse.ID] = struct {
					turnSeq int
					toolUse *ToolUse
				}{
					turnSeq: turn.Sequence,
					toolUse: block.ToolUse,
				}
			}
		}
	}

	// Step 2: 收集所有 ToolResult（按 tool_use_id 索引）
	toolResultMap := make(map[string]*ToolResult)

	for _, turn := range turns {
		for _, block := range turn.Content {
			if block.Type == "tool_result" && block.ToolResult != nil {
				toolResultMap[block.ToolResult.ToolUseID] = block.ToolResult
			}
		}
	}

	// Step 3: 生成 ToolCall 数组
	var toolCalls []ToolCall

	for toolUseID, tu := range toolUseMap {
		toolCall := ToolCall{
			TurnSequence: tu.turnSeq,
			ToolName:     tu.toolUse.Name,
			Input:        tu.toolUse.Input,
		}

		// 查找匹配的 ToolResult
		if result, found := toolResultMap[toolUseID]; found {
			toolCall.Output = result.Content
			toolCall.Status = result.Status
			toolCall.Error = result.Error
		}

		toolCalls = append(toolCalls, toolCall)
	}

	return toolCalls
}
```

**3. 运行测试**

```bash
# 运行所有 parser 测试
go test ./internal/parser -v

# 预期：所有测试通过
```

### 交付物

**文件清单**:
```
meta-cc/
├── internal/
│   └── parser/
│       ├── types.go          # 数据结构（Stage 2.1）
│       ├── types_test.go     # 数据结构测试（Stage 2.1）
│       ├── reader.go         # JSONL 读取器（Stage 2.2）
│       ├── reader_test.go    # 读取器测试（Stage 2.2）
│       ├── tools.go          # 工具调用提取 (~70 行)
│       └── tools_test.go     # 工具提取测试 (~100 行)
```

**代码量**: ~170 行（本 Stage）

### 验收标准

- ✅ `TestExtractToolCalls_SingleCall` 测试通过（单个工具调用）
- ✅ `TestExtractToolCalls_MultipleCallsSameTurn` 测试通过（同一 Turn 多个工具）
- ✅ `TestExtractToolCalls_UnmatchedToolUse` 测试通过（未匹配的 ToolUse）
- ✅ `TestExtractToolCalls_NoToolCalls` 测试通过（无工具调用）
- ✅ `go test ./internal/parser -v` 全部通过
- ✅ 所有导出函数和类型有文档注释

---

## Phase 2 集成测试

### 端到端测试：完整解析流程

创建 `tests/integration/parser_test.go` 进行端到端测试：

#### `tests/integration/parser_test.go` (~80 行)

```go
package integration

import (
	"testing"

	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/internal/testutil"
)

func TestIntegration_ParseRealSession(t *testing.T) {
	// 使用真实的测试 fixture
	filePath := testutil.FixtureDir() + "/sample-session.jsonl"

	// Step 1: 解析 JSONL 文件
	sessionParser := parser.NewSessionParser(filePath)
	turns, err := sessionParser.ParseTurns()

	if err != nil {
		t.Fatalf("Failed to parse session: %v", err)
	}

	if len(turns) != 3 {
		t.Fatalf("Expected 3 turns, got %d", len(turns))
	}

	// Step 2: 提取工具调用
	toolCalls := parser.ExtractToolCalls(turns)

	if len(toolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(toolCalls))
	}

	// Step 3: 验证工具调用内容
	tc := toolCalls[0]

	if tc.ToolName != "Grep" {
		t.Errorf("Expected tool name 'Grep', got '%s'", tc.ToolName)
	}

	if tc.TurnSequence != 1 {
		t.Errorf("Expected turn sequence 1, got %d", tc.TurnSequence)
	}

	expectedOutput := "src/auth.js:15: authError: token invalid"
	if tc.Output != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, tc.Output)
	}

	pattern, ok := tc.Input["pattern"].(string)
	if !ok || pattern != "auth.*error" {
		t.Errorf("Expected pattern 'auth.*error', got '%v'", pattern)
	}
}

func TestIntegration_ComplexSession(t *testing.T) {
	// 创建复杂的测试场景：多个工具调用、错误、嵌套等
	content := `{"sequence":0,"role":"user","timestamp":1000,"content":[{"type":"text","text":"test"}]}
{"sequence":1,"role":"assistant","timestamp":1001,"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"ls"}},{"type":"tool_use","id":"t2","name":"Read","input":{"file":"a.txt"}}]}
{"sequence":2,"role":"user","timestamp":1002,"content":[{"type":"tool_result","tool_use_id":"t1","content":"file1.txt\nfile2.txt","status":"success"},{"type":"tool_result","tool_use_id":"t2","content":"","status":"error","error":"file not found"}]}
{"sequence":3,"role":"assistant","timestamp":1003,"content":[{"type":"text","text":"found error"}]}`

	tempFile := testutil.TempSessionFile(t, content)

	// 解析
	sessionParser := parser.NewSessionParser(tempFile)
	turns, err := sessionParser.ParseTurns()

	if err != nil {
		t.Fatalf("Failed to parse complex session: %v", err)
	}

	if len(turns) != 4 {
		t.Fatalf("Expected 4 turns, got %d", len(turns))
	}

	// 提取工具调用
	toolCalls := parser.ExtractToolCalls(turns)

	if len(toolCalls) != 2 {
		t.Fatalf("Expected 2 tool calls, got %d", len(toolCalls))
	}

	// 验证错误状态
	errorToolFound := false
	for _, tc := range toolCalls {
		if tc.Status == "error" {
			errorToolFound = true
			if tc.Error != "file not found" {
				t.Errorf("Expected error 'file not found', got '%s'", tc.Error)
			}
		}
	}

	if !errorToolFound {
		t.Error("Expected to find a tool call with error status")
	}
}
```

### 运行集成测试

```bash
# 运行集成测试
go test ./tests/integration -run TestIntegration_Parse -v

# 运行所有测试（单元 + 集成）
go test ./... -v
```

---

## Phase 2 完成标准

### 功能验收

**必须满足所有条件**:

1. **数据结构定义**
   ```bash
   go test ./internal/parser -run TestTurnUnmarshal -v
   go test ./internal/parser -run TestContentBlockUnmarshal -v
   ```
   - ✅ Turn 结构正确反序列化
   - ✅ ContentBlock 根据 type 正确解析
   - ✅ ToolUse 和 ToolResult 正确提取

2. **JSONL 文件解析**
   ```bash
   go test ./internal/parser -run TestParseSession -v
   ```
   - ✅ 能解析真实的会话文件
   - ✅ 正确跳过空行
   - ✅ 非法 JSON 返回清晰的错误信息
   - ✅ 处理文件不存在的情况

3. **工具调用提取**
   ```bash
   go test ./internal/parser -run TestExtractToolCalls -v
   ```
   - ✅ 正确提取单个工具调用
   - ✅ 正确提取同一 Turn 的多个工具调用
   - ✅ 正确匹配 ToolUse 和 ToolResult
   - ✅ 处理未匹配的 ToolUse

4. **集成测试**
   ```bash
   go test ./tests/integration -run TestIntegration_Parse -v
   ```
   - ✅ 端到端解析真实会话文件
   - ✅ 提取工具调用并验证内容
   - ✅ 处理复杂场景（多工具、错误状态）

5. **所有测试通过**
   ```bash
   go test ./... -v
   ```
   - ✅ 所有单元测试通过
   - ✅ 所有集成测试通过
   - ✅ 无跳过或失败的测试

### 代码质量

- ✅ 总代码量 ≤ 500 行（Phase 2 约束）
  - Stage 2.1: ~190 行
  - Stage 2.2: ~190 行
  - Stage 2.3: ~170 行
  - 总计: ~550 行（包含测试）
  - 实现代码: ~250 行（符合约束）
- ✅ 每个 Stage 代码量 ≤ 200 行
- ✅ 无 Go 编译警告
- ✅ 所有导出函数、类型和方法有文档注释
- ✅ 测试覆盖率 > 80%

### 文档完整性

更新 `README.md`，添加 JSONL 解析器使用说明：

```markdown
## JSONL Parser

Parse Claude Code session files in JSONL format.

### Basic Usage

```go
import "github.com/yale/meta-cc/internal/parser"

// Parse session file
sessionParser := parser.NewSessionParser("/path/to/session.jsonl")
turns, err := sessionParser.ParseTurns()
if err != nil {
    log.Fatal(err)
}

// Extract tool calls
toolCalls := parser.ExtractToolCalls(turns)

for _, tc := range toolCalls {
    fmt.Printf("Tool: %s\n", tc.ToolName)
    fmt.Printf("Input: %v\n", tc.Input)
    fmt.Printf("Output: %s\n", tc.Output)
    if tc.Status == "error" {
        fmt.Printf("Error: %s\n", tc.Error)
    }
}
```

### Data Structures

- **Turn**: Represents a conversation turn (user or assistant)
- **ContentBlock**: A block within a turn (text, tool_use, or tool_result)
- **ToolUse**: A tool invocation with parameters
- **ToolResult**: The result of a tool execution
- **ToolCall**: Complete tool call (ToolUse + ToolResult matched)

### Session File Format

```jsonl
{"sequence":0,"role":"user","timestamp":1735689600,"content":[{"type":"text","text":"帮我修复这个认证 bug"}]}
{"sequence":1,"role":"assistant","timestamp":1735689605,"content":[{"type":"text","text":"我来帮你检查代码"},{"type":"tool_use","id":"toolu_01","name":"Grep","input":{"pattern":"auth.*error","path":"."}}]}
{"sequence":2,"role":"user","timestamp":1735689610,"content":[{"type":"tool_result","tool_use_id":"toolu_01","content":"src/auth.js:15: authError: token invalid"}]}
```
```

---

## 项目结构（Phase 2 完成后）

```
meta-cc/
├── go.mod
├── go.sum
├── Makefile
├── README.md                       # 更新：添加 JSONL 解析器说明
├── main.go
├── cmd/
│   └── root.go
├── internal/
│   ├── locator/                    # Phase 1
│   │   ├── env.go
│   │   ├── env_test.go
│   │   ├── args.go
│   │   ├── args_test.go
│   │   ├── helpers.go
│   │   ├── hash_test.go
│   │   └── locator.go
│   ├── parser/                     # Phase 2（新增）
│   │   ├── types.go               # 数据结构定义
│   │   ├── types_test.go
│   │   ├── reader.go              # JSONL 读取器
│   │   ├── reader_test.go
│   │   ├── tools.go               # 工具调用提取
│   │   └── tools_test.go
│   └── testutil/
│       ├── fixtures.go
│       ├── fixtures_test.go
│       └── time.go
└── tests/
    ├── fixtures/
    │   └── sample-session.jsonl
    └── integration/
        ├── locator_test.go
        └── parser_test.go          # 新增：解析器集成测试
```

---

## 依赖关系

**Phase 2 依赖**:
- Phase 0（CLI 框架、测试工具）
- Phase 1（会话文件定位）- 集成时使用

**后续 Phase 依赖于 Phase 2**:
- Phase 3（统计分析）依赖于 JSONL 解析器提供的 Turn 和 ToolCall 数据
- Phase 4（CLI 命令）依赖于解析器和定位器集成

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Claude Code 会话格式变化 | 高 | 充分测试真实会话文件；使用灵活的 JSON 解析（忽略未知字段） |
| 大型会话文件性能问题 | 中 | 使用流式解析（bufio.Scanner）；避免一次性加载整个文件 |
| 复杂的 ContentBlock 类型 | 中 | 使用自定义 UnmarshalJSON；处理未知类型时不报错 |
| ToolUse 和 ToolResult 不匹配 | 低 | 允许未匹配的 ToolUse（Output 为空）；记录警告 |

---

## 下一步行动

**Phase 2 完成后，进入 Phase 3: 统计分析**

Phase 3 将实现：
- 工具使用频率统计
- 错误模式检测
- 时间线分析
- 会话摘要生成

**准备工作**:
1. 确认 Phase 2 所有验收标准已满足
2. 运行完整测试套件（`make test`）
3. 测试与 Phase 1（locator）的集成
4. 提交代码到 git（使用 `feat:` 前缀）
5. 创建 Phase 3 规划文档 `plans/3/plan.md`

**集成验证**:
```bash
# 创建集成测试：locator + parser
go test ./tests/integration -run TestIntegration_LocatorAndParser -v
```

集成测试示例：
```go
func TestIntegration_LocatorAndParser(t *testing.T) {
    // 1. 使用 locator 定位会话文件
    loc := locator.NewSessionLocator()
    sessionPath, err := loc.Locate(locator.LocateOptions{
        ProjectPath: "/home/yale/work/myproject",
    })

    // 2. 使用 parser 解析会话文件
    p := parser.NewSessionParser(sessionPath)
    turns, err := p.ParseTurns()

    // 3. 提取工具调用
    toolCalls := parser.ExtractToolCalls(turns)

    // 验证完整流程
    if len(toolCalls) == 0 {
        t.Error("Expected tool calls from real session")
    }
}
```

---

## Phase 2 实现摘要

### 核心功能

1. **数据结构** (`types.go`)
   - Turn: 会话轮次
   - ContentBlock: 内容块（text/tool_use/tool_result）
   - ToolUse: 工具调用
   - ToolResult: 工具结果
   - 自定义 UnmarshalJSON 处理多态类型

2. **JSONL 读取器** (`reader.go`)
   - 逐行读取 JSONL 文件
   - 跳过空行
   - 错误处理和行号报告

3. **工具调用提取器** (`tools.go`)
   - 收集所有 ToolUse
   - 匹配 ToolResult
   - 生成完整的 ToolCall 数组

### 测试覆盖

- 单元测试: 290+ 行
- 集成测试: 80+ 行
- 覆盖场景:
  - 正常解析
  - 边界情况（空文件、空行、非法 JSON）
  - 复杂场景（多工具、错误状态、未匹配）
  - 端到端集成

### 代码行数统计

| 组件 | 实现代码 | 测试代码 | 总计 |
|------|---------|---------|------|
| Stage 2.1 (types) | ~100 | ~90 | ~190 |
| Stage 2.2 (reader) | ~80 | ~110 | ~190 |
| Stage 2.3 (tools) | ~70 | ~100 | ~170 |
| 集成测试 | - | ~80 | ~80 |
| **总计** | **~250** | **~380** | **~630** |

实现代码 ~250 行，符合 Phase 2 约束（≤ 500 行）。
