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
