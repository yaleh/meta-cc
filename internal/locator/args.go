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
	projectsRoot := l.projectsRoot
	if projectsRoot == "" {
		return "", fmt.Errorf("Claude Code projects directory not configured")
	}

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
	// 解析相对路径为绝对路径（如 "." -> "/home/yale/work/meta-cc"）
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve project path: %w", err)
	}

	// 计算项目哈希 (pathToHash now handles symlink resolution)
	projectHash := pathToHash(absPath)

	projectsRoot := l.projectsRoot
	if projectsRoot == "" {
		return "", fmt.Errorf("Claude Code projects directory not configured")
	}

	sessionDir := filepath.Join(projectsRoot, projectHash)
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

// AllSessionsFromProject 通过项目路径查找所有会话文件
// 1. 将项目路径转换为哈希（/ → -）
// 2. 定位 ~/.claude/projects/{hash}/
// 3. 返回该目录下所有 .jsonl 文件的路径
func (l *SessionLocator) AllSessionsFromProject(projectPath string) ([]string, error) {
	// 解析相对路径为绝对路径（如 "." -> "/home/yale/work/meta-cc"）
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project path: %w", err)
	}

	// 计算项目哈希 (pathToHash now handles symlink resolution)
	projectHash := pathToHash(absPath)

	projectsRoot := l.projectsRoot
	if projectsRoot == "" {
		return nil, fmt.Errorf("Claude Code projects directory not configured")
	}

	sessionDir := filepath.Join(projectsRoot, projectHash)
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no sessions found for project: %s (hash: %s)", projectPath, projectHash)
	}

	// 查找所有 .jsonl 文件
	sessions, err := filepath.Glob(filepath.Join(sessionDir, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("failed to search session files: %w", err)
	}

	if len(sessions) == 0 {
		return nil, fmt.Errorf("no session files found in: %s", sessionDir)
	}

	// 返回所有会话文件
	return sessions, nil
}

// pathToHash 将项目路径转换为哈希目录名
// 例如：/home/yale/work/myproject → -home-yale-work-myproject
// Windows: C:/Users/yale/work/myproject → C--Users-yale-work-myproject
//
// Note: Resolves symlinks to ensure consistent hashing across platforms.
// On macOS, /var is a symlink to /private/var, so we resolve it before hashing.
func pathToHash(path string) string {
	// Handle empty path edge case
	if path == "" {
		return ""
	}

	// Resolve symlinks for consistent hashing (e.g., /var -> /private/var on macOS)
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		// If resolution fails (e.g., path doesn't exist), use original path
		resolved = path
	}

	// Normalize path separators (both forward slash and backslash) to -
	// First replace backslashes (Windows paths)
	hash := strings.ReplaceAll(resolved, "\\", "-")
	// Then replace forward slashes (Unix paths and normalized Windows paths)
	hash = strings.ReplaceAll(hash, "/", "-")
	// Finally replace colons (Windows drive letters like C:)
	hash = strings.ReplaceAll(hash, ":", "-")
	return hash
}
