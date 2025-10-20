package locator

import (
	"fmt"
	"os"
)

// LocateOptions 定位选项
type LocateOptions struct {
	SessionID   string // 命令行参数 --session
	ProjectPath string // 命令行参数 --project
	SessionOnly bool   // Phase 13: 强制仅分析当前会话（禁用项目级默认行为）
}

// Locate 统一的会话文件定位入口
// Phase 13: 默认使用项目级分析（--project .），除非设置 --session-only
// 按优先级尝试以下策略：
//
//  1. 命令行参数 --session
//  2. 命令行参数 --project
//  3. 默认：当前工作目录（Phase 13: 项目级默认）
func (l *SessionLocator) Locate(opts LocateOptions) (string, error) {
	// 策略1: --session 参数
	if opts.SessionID != "" {
		path, err := l.FromSessionID(opts.SessionID)
		if err == nil {
			return path, nil
		}
		// 明确指定了 session 但找不到，直接返回错误
		return "", fmt.Errorf("session ID %q not found: %w", opts.SessionID, err)
	}

	// Phase 13: 默认使用当前工作目录作为项目路径（除非明确指定 --project）
	projectPath := opts.ProjectPath
	if projectPath == "" && !opts.SessionOnly {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		projectPath = cwd
	}

	// 策略2: --project 参数或默认项目路径
	if projectPath != "" {
		path, err := l.FromProjectPath(projectPath)
		if err == nil {
			return path, nil
		}
		// 明确指定了 project 但找不到，直接返回错误
		return "", fmt.Errorf("no sessions found for project %q: %w", projectPath, err)
	}

	return "", fmt.Errorf("failed to locate session file: no session specified")
}
