package locator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFromEnv_Success(t *testing.T) {
	// 设置环境变量
	os.Setenv("CC_SESSION_ID", "5b57148c-89dc-4eb5-bc37-8122e194d90d")
	os.Setenv("CC_PROJECT_HASH", "-home-yale-work-myproject")
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	// 创建测试会话文件
	homeDir, _ := os.UserHomeDir()
	sessionDir := filepath.Join(homeDir, ".claude", "projects", "-home-yale-work-myproject")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, "5b57148c-89dc-4eb5-bc37-8122e194d90d.jsonl")
	if err := os.WriteFile(sessionFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}
	defer os.RemoveAll(filepath.Join(homeDir, ".claude", "projects", "-home-yale-work-myproject"))

	locator := NewSessionLocator()
	path, err := locator.FromEnv()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedPath := filepath.Join(homeDir, ".claude", "projects", "-home-yale-work-myproject", "5b57148c-89dc-4eb5-bc37-8122e194d90d.jsonl")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestFromEnv_MissingSessionID(t *testing.T) {
	// 确保环境变量不存在
	os.Unsetenv("CC_SESSION_ID")
	os.Unsetenv("CC_PROJECT_HASH")

	locator := NewSessionLocator()
	_, err := locator.FromEnv()

	if err == nil {
		t.Error("Expected error when CC_SESSION_ID is missing")
	}

	expectedMsg := "CC_SESSION_ID environment variable not set"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestFromEnv_MissingProjectHash(t *testing.T) {
	os.Setenv("CC_SESSION_ID", "test-session")
	os.Unsetenv("CC_PROJECT_HASH")
	defer os.Unsetenv("CC_SESSION_ID")

	locator := NewSessionLocator()
	_, err := locator.FromEnv()

	if err == nil {
		t.Error("Expected error when CC_PROJECT_HASH is missing")
	}

	expectedMsg := "CC_PROJECT_HASH environment variable not set"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestFromEnv_FileNotFound(t *testing.T) {
	os.Setenv("CC_SESSION_ID", "nonexistent-session")
	os.Setenv("CC_PROJECT_HASH", "-nonexistent-project")
	defer os.Unsetenv("CC_SESSION_ID")
	defer os.Unsetenv("CC_PROJECT_HASH")

	locator := NewSessionLocator()
	_, err := locator.FromEnv()

	if err == nil {
		t.Error("Expected error when session file does not exist")
	}
}
