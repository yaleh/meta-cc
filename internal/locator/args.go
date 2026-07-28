package locator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// FromSessionIDScoped resolves sessionID to a file path exactly like
// FromSessionID, but additionally enforces the caller's working_dir/cwd
// boundary before returning it.
//
// FromSessionID (above) is a GLOBAL search: it walks every project-hash
// directory on disk looking for a matching {session_id}.jsonl and returns
// whatever it finds, without ever comparing the result against the
// caller's working directory. Used directly, that shape is a cross-project
// leak: any caller who learns a session_id can read that session's content
// regardless of which project it claims to be scoped to.
//
// This exact defect was independently introduced and independently caught
// three separate times in this repo's history — internal/mcp/executor's
// ExecuteQueryForSession (DIR-030), internal/provider/claude's
// findSessionFile (found during the DIR-032 build), and
// internal/analysis/service.go's loadData (found by a DIR-032 adversarial
// audit AFTER the bug class was believed closed). Each fix hand-wrote the
// same boundary comparison: resolve workingDir to the project-hash
// directory name Claude Code itself uses (PathToHash) and reject the match
// if the resolved session file does not live under it. DIR-033
// crystallizes that one comparison here so no future caller has to
// remember to reimplement it — FromSessionID itself stays unscoped and is
// only ever called from within this package; every external caller must
// go through FromSessionIDScoped instead.
//
// An empty workingDir is a no-op (matches the pre-existing per-callsite
// behavior of skipping the boundary check when no project scope was
// requested/available).
func (l *SessionLocator) FromSessionIDScoped(sessionID, workingDir string) (string, error) {
	file, err := l.FromSessionID(sessionID)
	if err != nil {
		return "", fmt.Errorf("session_id %q not found: %w", sessionID, err)
	}

	boundaryDir := workingDir
	if abs, absErr := filepath.Abs(boundaryDir); absErr == nil {
		boundaryDir = abs
	}
	if expectedHash := PathToHash(boundaryDir); expectedHash != "" {
		actualHash := filepath.Base(filepath.Dir(file))
		if actualHash != expectedHash {
			return "", fmt.Errorf("session_id %q not found for project %q", sessionID, boundaryDir)
		}
	}

	return file, nil
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
