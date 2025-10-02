package locator

import (
	"fmt"
	"os"
	"path/filepath"
)

// SessionLocator 负责定位会话文件
type SessionLocator struct{}

// NewSessionLocator 创建 SessionLocator 实例
func NewSessionLocator() *SessionLocator {
	return &SessionLocator{}
}

// FromEnv 从环境变量读取会话 ID 和项目哈希，构造文件路径
// 环境变量：
//   - CC_SESSION_ID: 会话 UUID
//   - CC_PROJECT_HASH: 项目路径哈希（已转换，如 -home-yale-work-myproject）
//
// 返回：
//   - 会话文件的完整路径
//   - 错误（如果环境变量缺失或文件不存在）
func (l *SessionLocator) FromEnv() (string, error) {
	sessionID := os.Getenv("CC_SESSION_ID")
	if sessionID == "" {
		return "", fmt.Errorf("CC_SESSION_ID environment variable not set")
	}

	projectHash := os.Getenv("CC_PROJECT_HASH")
	if projectHash == "" {
		return "", fmt.Errorf("CC_PROJECT_HASH environment variable not set")
	}

	// 构造会话文件路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	sessionPath := filepath.Join(
		homeDir,
		".claude",
		"projects",
		projectHash,
		fmt.Sprintf("%s.jsonl", sessionID),
	)

	// 验证文件存在
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return "", fmt.Errorf("session file not found: %s", sessionPath)
	} else if err != nil {
		return "", fmt.Errorf("failed to access session file: %w", err)
	}

	return sessionPath, nil
}
