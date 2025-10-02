package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SessionEntry 表示 Claude Code 会话文件中的一个条目
// 可以是 user 消息、assistant 消息或其他类型（如 file-history-snapshot）
type SessionEntry struct {
	Type       string   `json:"type"`       // "user", "assistant", "file-history-snapshot", etc.
	Timestamp  string   `json:"timestamp"`  // ISO 8601 格式: "2025-10-02T06:07:13.673Z"
	UUID       string   `json:"uuid"`       // 条目唯一标识
	ParentUUID string   `json:"parentUuid"` // 父条目 UUID（构建消息链）
	SessionID  string   `json:"sessionId"`  // 会话 ID
	CWD        string   `json:"cwd"`        // 工作目录
	Version    string   `json:"version"`    // Claude Code 版本
	GitBranch  string   `json:"gitBranch"`  // Git 分支
	Message    *Message `json:"message"`    // 消息内容（仅 user/assistant 类型有值）
}

// IsMessage 判断条目是否为消息类型（user 或 assistant）
func (e *SessionEntry) IsMessage() bool {
	return e.Type == "user" || e.Type == "assistant"
}

// Message 表示消息的详细内容
type Message struct {
	ID         string                 `json:"id"`          // 消息 ID（assistant 消息有值）
	Role       string                 `json:"role"`        // "user" 或 "assistant"
	Model      string                 `json:"model"`       // 模型名称（assistant 消息有值）
	Content    []ContentBlock         `json:"-"`           // 内容块数组（手动处理）
	StopReason string                 `json:"stop_reason"` // 停止原因
	Usage      map[string]interface{} `json:"usage"`       // Token 使用统计
}

// UnmarshalJSON 自定义 JSON 反序列化
// content 字段可以是 string 或 []ContentBlock
func (m *Message) UnmarshalJSON(data []byte) error {
	// 首先尝试解析到临时结构
	type Alias Message
	aux := &struct {
		ContentRaw json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 处理 content 字段
	if len(aux.ContentRaw) == 0 {
		return nil
	}

	// 尝试作为字符串解析
	var contentStr string
	if err := json.Unmarshal(aux.ContentRaw, &contentStr); err == nil {
		// content 是字符串，转换为单个 text ContentBlock
		m.Content = []ContentBlock{
			{
				Type: "text",
				Text: contentStr,
			},
		}
		return nil
	}

	// 否则作为数组解析
	return json.Unmarshal(aux.ContentRaw, &m.Content)
}

// ContentBlock 表示消息中的一个内容块
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
	Content   string `json:"-"` // 手动处理（可以是 string 或 array）
	Status    string `json:"status,omitempty"`
	Error     string `json:"error,omitempty"`
}

// UnmarshalJSON 自定义 ToolResult 的反序列化逻辑
// content 字段可以是 string 或 array
func (tr *ToolResult) UnmarshalJSON(data []byte) error {
	type Alias ToolResult
	aux := &struct {
		ContentRaw json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(tr),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 处理 content 字段
	if len(aux.ContentRaw) == 0 {
		return nil
	}

	// 尝试作为字符串解析
	var contentStr string
	if err := json.Unmarshal(aux.ContentRaw, &contentStr); err == nil {
		tr.Content = contentStr
		return nil
	}

	// 否则作为数组解析（提取文本）
	var contentBlocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(aux.ContentRaw, &contentBlocks); err != nil {
		return fmt.Errorf("failed to unmarshal tool_result content: %w", err)
	}

	// 将所有文本块合并
	var texts []string
	for _, block := range contentBlocks {
		if block.Text != "" {
			texts = append(texts, block.Text)
		}
	}
	tr.Content = strings.Join(texts, "\n")

	return nil
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
		// 使用 ToolResult 的自定义 UnmarshalJSON 方法
		var toolResult ToolResult
		if err := json.Unmarshal(data, &toolResult); err != nil {
			return fmt.Errorf("failed to unmarshal tool_result: %w", err)
		}
		cb.ToolResult = &toolResult

	default:
		// 未知类型，保留原始数据但不报错
	}

	return nil
}
