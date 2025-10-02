package locator

import (
	"fmt"
	"os"
)

// LocateOptions 定位选项
type LocateOptions struct {
	SessionID   string // 命令行参数 --session
	ProjectPath string // 命令行参数 --project
}

// Locate 统一的会话文件定位入口
// 按优先级尝试以下策略：
//
//	1. 环境变量 CC_SESSION_ID + CC_PROJECT_HASH
//	2. 命令行参数 --session
//	3. 命令行参数 --project
//	4. 自动检测（使用当前工作目录）
func (l *SessionLocator) Locate(opts LocateOptions) (string, error) {
	// 策略1: 环境变量
	if os.Getenv("CC_SESSION_ID") != "" {
		path, err := l.FromEnv()
		if err == nil {
			return path, nil
		}
		// 环境变量设置了但失败，记录警告但继续尝试其他策略
	}

	// 策略2: --session 参数
	if opts.SessionID != "" {
		path, err := l.FromSessionID(opts.SessionID)
		if err == nil {
			return path, nil
		}
		// 明确指定了 session 但找不到，直接返回错误
		return "", fmt.Errorf("session ID %q not found: %w", opts.SessionID, err)
	}

	// 策略3: --project 参数
	if opts.ProjectPath != "" {
		path, err := l.FromProjectPath(opts.ProjectPath)
		if err == nil {
			return path, nil
		}
		// 明确指定了 project 但找不到，直接返回错误
		return "", fmt.Errorf("no sessions found for project %q: %w", opts.ProjectPath, err)
	}

	// 策略4: 自动检测（使用当前工作目录）
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	path, err := l.FromProjectPath(cwd)
	if err == nil {
		return path, nil
	}

	return "", fmt.Errorf("failed to locate session file: tried env vars, session ID, project path, and auto-detection")
}
