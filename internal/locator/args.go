package locator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FromSessionID 通过会话 ID 查找会话文件
// 遍历 ~/.claude/projects/*/，查找匹配的 {session-id}.jsonl
// 如果找到多个（跨项目同名会话），返回最新的
func (l *SessionLocator) FromSessionID(sessionID string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	projectsRoot := filepath.Join(homeDir, ".claude", "projects")
	if _, err := os.Stat(projectsRoot); os.IsNotExist(err) {
		return "", fmt.Errorf("Claude Code projects directory not found: %s", projectsRoot)
	}

	// 遍历所有项目目录
	projectDirs, err := os.ReadDir(projectsRoot)
	if err != nil {
		return "", fmt.Errorf("failed to read projects directory: %w", err)
	}

	var candidates []string
	sessionFilename := sessionID + ".jsonl"

	for _, projectDir := range projectDirs {
		if !projectDir.IsDir() {
			continue
		}

		sessionPath := filepath.Join(projectsRoot, projectDir.Name(), sessionFilename)
		if _, err := os.Stat(sessionPath); err == nil {
			candidates = append(candidates, sessionPath)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("session file not found for ID: %s", sessionID)
	}

	// 如果找到多个，返回最新的
	return findNewestFile(candidates)
}

// FromProjectPath 通过项目路径查找最新会话
// 1. 将项目路径转换为哈希（/ → -）
// 2. 定位 ~/.claude/projects/{hash}/
// 3. 返回该目录下最新的 .jsonl 文件
func (l *SessionLocator) FromProjectPath(projectPath string) (string, error) {
	// 计算项目哈希
	projectHash := pathToHash(projectPath)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		return "", fmt.Errorf("no sessions found for project: %s (hash: %s)", projectPath, projectHash)
	}

	// 查找所有 .jsonl 文件
	sessions, err := filepath.Glob(filepath.Join(sessionDir, "*.jsonl"))
	if err != nil {
		return "", fmt.Errorf("failed to search session files: %w", err)
	}

	if len(sessions) == 0 {
		return "", fmt.Errorf("no session files found in: %s", sessionDir)
	}

	// 返回最新的会话文件
	return findNewestFile(sessions)
}

// pathToHash 将项目路径转换为哈希目录名
// 例如：/home/yale/work/myproject → -home-yale-work-myproject
func pathToHash(path string) string {
	return strings.ReplaceAll(path, "/", "-")
}
