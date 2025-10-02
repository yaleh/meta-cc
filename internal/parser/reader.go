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

// ParseEntries 解析 JSONL 文件，返回 SessionEntry 数组
// JSONL 格式：每行一个 JSON 对象
// 处理规则：
//   - 跳过空行和空白行
//   - 非法 JSON 行返回错误
//   - 仅返回消息类型（type == "user" 或 "assistant"）
//   - 过滤掉 file-history-snapshot 等非消息类型
func (p *SessionParser) ParseEntries() ([]SessionEntry, error) {
	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open session file: %w", err)
	}
	defer file.Close()

	var entries []SessionEntry
	scanner := bufio.NewScanner(file)

	// Increase buffer size for large lines (Claude Code sessions can have very long lines)
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// 跳过空行和仅包含空白的行
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 解析 JSON 为 SessionEntry
		var entry SessionEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}

		// 仅保留消息类型
		if entry.IsMessage() {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading session file: %w", err)
	}

	return entries, nil
}

// ParseEntriesFromContent 从字符串内容解析 JSONL（用于测试）
func ParseEntriesFromContent(content string) ([]SessionEntry, error) {
	var entries []SessionEntry
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		// 跳过空行
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry SessionEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum+1, err)
		}

		// 仅保留消息类型
		if entry.IsMessage() {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}
