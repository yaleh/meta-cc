package locator

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yaleh/meta-cc/internal/parser"
)

// FromSessionID 通过会话 ID 查找会话文件
// 遍历支持的 transcript roots，查找匹配的 {session-id}.jsonl
// 如果找到多个（跨项目同名会话），返回最新的
func (l *SessionLocator) FromSessionID(sessionID string) (string, error) {
	var candidates []string
	sessionFilename := sessionID + ".jsonl"
	var checked []string

	for _, root := range l.TranscriptRoots() {
		checked = append(checked, formatRoot(root))
		if _, err := os.Stat(root.Path); os.IsNotExist(err) {
			continue
		}

		if root.ProjectHashed {
			projectDirs, err := os.ReadDir(root.Path)
			if err != nil {
				continue
			}
			for _, projectDir := range projectDirs {
				if !projectDir.IsDir() {
					continue
				}

				sessionPath := filepath.Join(root.Path, projectDir.Name(), sessionFilename)
				if _, err := os.Stat(sessionPath); err == nil {
					candidates = append(candidates, sessionPath)
				}
			}
			continue
		}

		sessionPath := filepath.Join(root.Path, sessionFilename)
		if _, err := os.Stat(sessionPath); err == nil {
			candidates = append(candidates, sessionPath)
			continue
		}

		matches, err := findSessionFilesRecursive(root.Path, sessionFilename)
		if err == nil {
			candidates = append(candidates, matches...)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("session file not found for ID %q; checked transcript roots: %s",
			sessionID, strings.Join(checked, ", "))
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

	projectHash := PathToHash(absPath)

	sessions, err := l.sessionsFromProject(absPath, projectHash)
	if err != nil {
		return "", fmt.Errorf("no sessions found for project %q (hash: %s): %w",
			projectPath, projectHash, err)
	}

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

	projectHash := PathToHash(absPath)

	sessions, err := l.sessionsFromProject(absPath, projectHash)
	if err != nil {
		return nil, fmt.Errorf("no sessions found for project %q (hash: %s): %w",
			projectPath, projectHash, err)
	}

	// 返回所有会话文件
	return sessions, nil
}

func (l *SessionLocator) sessionsFromProject(projectPath, projectHash string) ([]string, error) {
	var sessions []string
	var checked []string

	for _, root := range l.TranscriptRoots() {
		checked = append(checked, formatRoot(root))
		if !root.ProjectHashed {
			continue
		}
		if _, err := os.Stat(root.Path); os.IsNotExist(err) {
			continue
		}

		sessionDir := filepath.Join(root.Path, projectHash)
		rootSessions, err := filepath.Glob(filepath.Join(sessionDir, "*.jsonl"))
		if err != nil {
			return nil, fmt.Errorf("failed to search session files in %s: %w", sessionDir, err)
		}
		sessions = append(sessions, rootSessions...)
	}
	if len(sessions) > 0 {
		return sessions, nil
	}

	return nil, fmt.Errorf("checked transcript roots: %s", strings.Join(checked, ", "))
}

func findSessionFilesRecursive(rootPath, filename string) ([]string, error) {
	var matches []string
	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == filename {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, errors.New("no matching session files")
	}
	return matches, nil
}

func findJSONLFilesRecursive(rootPath string) ([]string, error) {
	var matches []string
	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".jsonl" {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, errors.New("no jsonl session files")
	}
	return matches, nil
}

func findProjectJSONLFilesRecursive(rootPath, projectPath string) ([]string, error) {
	all, err := findJSONLFilesRecursive(rootPath)
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, path := range all {
		if fileContains(path, projectPath) {
			matches = append(matches, path)
		}
	}
	if len(matches) == 0 {
		return nil, errors.New("no project-matching jsonl session files")
	}
	return matches, nil
}

func fileContains(path, projectPath string) bool {
	if projectPath == "" {
		return false
	}
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	cleanProject := filepath.Clean(projectPath)
	reader := bufio.NewReader(file)
	for {
		line, _, readErr := parser.ReadLineFiltered(reader, parser.StrategyDefault)
		if len(line) == 0 && readErr == io.EOF {
			break
		}
		if readErr != nil && readErr != io.EOF {
			return false
		}
		var record map[string]interface{}
		if err := json.Unmarshal(line, &record); err != nil {
			if readErr == io.EOF {
				break
			}
			continue
		}
		if recordProjectPath(record, cleanProject) {
			return true
		}
		if readErr == io.EOF {
			break
		}
	}
	return false
}

func recordProjectPath(record map[string]interface{}, cleanProject string) bool {
	if cwd, ok := record["cwd"].(string); ok && filepath.Clean(cwd) == cleanProject {
		return true
	}
	payload, ok := record["payload"].(map[string]interface{})
	if !ok {
		return false
	}
	for _, key := range []string{"cwd", "working_dir", "workingDir"} {
		if cwd, ok := payload[key].(string); ok && filepath.Clean(cwd) == cleanProject {
			return true
		}
	}
	return false
}

func formatRoot(root SessionRoot) string {
	if root.ProjectHashed {
		return fmt.Sprintf("%s=%s (project-hash)", root.Host, root.Path)
	}
	return fmt.Sprintf("%s=%s", root.Host, root.Path)
}

// PathToHash converts a project path to the hashed directory name used by Claude Code.
// Example: /home/yale/work/myproject → -home-yale-work-myproject
// Windows: C:/Users/yale/work/myproject → C--Users-yale-work-myproject
//
// Resolves symlinks for consistent hashing across platforms
// (e.g., /var → /private/var on macOS).
func PathToHash(path string) string {
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
