package locator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocate_WithSessionID(t *testing.T) {
	// 准备测试环境
	homeDir, _ := os.UserHomeDir()
	projectHash := "-test-locate-session"
	sessionID := "test-session-123"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"test":"data"}`), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	locator := NewSessionLocator()
	opts := LocateOptions{
		SessionID: sessionID,
	}

	path, err := locator.Locate(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if path != sessionFile {
		t.Errorf("Expected path %s, got %s", sessionFile, path)
	}
}

func TestLocate_SessionIDNotFound(t *testing.T) {
	locator := NewSessionLocator()
	opts := LocateOptions{
		SessionID: "nonexistent-session-id",
	}

	_, err := locator.Locate(opts)
	if err == nil {
		t.Error("Expected error for nonexistent session ID")
	}
}

func TestLocate_WithProjectPath(t *testing.T) {
	// 准备测试环境
	homeDir, _ := os.UserHomeDir()
	testProjectPath := "/test/project/path"
	projectHash := pathToHash(testProjectPath)

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, "session-abc.jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"test":"data"}`), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	locator := NewSessionLocator()
	opts := LocateOptions{
		ProjectPath: testProjectPath,
	}

	path, err := locator.Locate(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if path != sessionFile {
		t.Errorf("Expected path %s, got %s", sessionFile, path)
	}
}

func TestLocate_ProjectPathNotFound(t *testing.T) {
	locator := NewSessionLocator()
	opts := LocateOptions{
		ProjectPath: "/nonexistent/project/path",
	}

	_, err := locator.Locate(opts)
	if err == nil {
		t.Error("Expected error for nonexistent project path")
	}
}

func TestLocate_DefaultCWD(t *testing.T) {
	// 准备测试环境：使用当前工作目录
	cwd, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	projectHash := pathToHash(cwd)

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, "default-session.jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"test":"data"}`), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	locator := NewSessionLocator()
	opts := LocateOptions{}

	path, err := locator.Locate(opts)
	if err != nil {
		t.Fatalf("Expected no error with default CWD, got: %v", err)
	}

	if path != sessionFile {
		t.Errorf("Expected path %s, got %s", sessionFile, path)
	}
}

func TestLocate_SessionOnlyMode(t *testing.T) {
	locator := NewSessionLocator()
	opts := LocateOptions{
		SessionOnly: true,
	}

	// 在 session-only 模式下，如果没有环境变量或其他参数，应该失败
	_, err := locator.Locate(opts)
	if err == nil {
		t.Error("Expected error in session-only mode without session ID or env vars")
	}
}

func TestLocate_WithEnvVars(t *testing.T) {
	// 准备测试环境
	homeDir, _ := os.UserHomeDir()
	sessionID := "env-session-123"
	projectHash := "-test-env-project"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"test":"data"}`), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	// 设置环境变量
	os.Setenv("CC_SESSION_ID", sessionID)
	os.Setenv("CC_PROJECT_HASH", projectHash)
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	locator := NewSessionLocator()
	opts := LocateOptions{
		SessionOnly: true,
	}

	path, err := locator.Locate(opts)
	if err != nil {
		t.Fatalf("Expected no error with env vars, got: %v", err)
	}

	if path != sessionFile {
		t.Errorf("Expected path %s, got %s", sessionFile, path)
	}
}

func TestLocate_EnvVarsIgnoredInProjectMode(t *testing.T) {
	// 设置环境变量（应该被忽略）
	os.Setenv("CC_SESSION_ID", "should-be-ignored")
	defer os.Unsetenv("CC_SESSION_ID")

	// 准备项目路径的会话
	homeDir, _ := os.UserHomeDir()
	testProjectPath := "/test/project/for/env/test"
	projectHash := pathToHash(testProjectPath)

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, "project-session.jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"test":"data"}`), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	locator := NewSessionLocator()
	opts := LocateOptions{
		ProjectPath: testProjectPath,
		SessionOnly: false, // 项目模式
	}

	path, err := locator.Locate(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 应该使用项目路径的会话，而不是环境变量
	if path != sessionFile {
		t.Errorf("Expected path %s (from project), got %s", sessionFile, path)
	}
}

func TestLocate_SessionIDPriority(t *testing.T) {
	// 准备多个选项，验证 SessionID 优先级最高
	homeDir, _ := os.UserHomeDir()
	sessionID := "priority-test-session"
	projectHash := "-test-priority"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"test":"data"}`), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	// 设置环境变量（优先级更低）
	os.Setenv("CC_SESSION_ID", "env-session")
	defer os.Unsetenv("CC_SESSION_ID")

	locator := NewSessionLocator()
	opts := LocateOptions{
		SessionID:   sessionID,
		ProjectPath: "/some/other/path",
	}

	path, err := locator.Locate(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 应该使用 SessionID，而不是 ProjectPath 或环境变量
	if path != sessionFile {
		t.Errorf("Expected path from SessionID: %s, got %s", sessionFile, path)
	}
}
