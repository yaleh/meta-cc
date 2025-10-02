package locator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yale/meta-cc/internal/testutil"
)

func TestFromSessionID_Success(t *testing.T) {
	// 准备测试环境
	homeDir, _ := os.UserHomeDir()
	projectHash := "-test-project-session-id"
	sessionID := "abc123-def456"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	os.MkdirAll(sessionDir, 0755)
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	os.WriteFile(sessionFile, []byte(`{"test":"data"}`), 0644)
	defer os.RemoveAll(sessionDir)

	locator := NewSessionLocator()
	path, err := locator.FromSessionID(sessionID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if path != sessionFile {
		t.Errorf("Expected path %s, got %s", sessionFile, path)
	}
}

func TestFromSessionID_NotFound(t *testing.T) {
	locator := NewSessionLocator()
	_, err := locator.FromSessionID("nonexistent-session-id")

	if err == nil {
		t.Error("Expected error for nonexistent session ID")
	}
}

func TestFromSessionID_MultipleProjects(t *testing.T) {
	// 准备：在多个项目目录中创建同名会话文件
	homeDir, _ := os.UserHomeDir()
	sessionID := "shared-session-id"

	// 项目1（旧）
	project1 := filepath.Join(homeDir, ".claude", "projects", "-project1")
	os.MkdirAll(project1, 0755)
	file1 := filepath.Join(project1, sessionID+".jsonl")
	os.WriteFile(file1, []byte("{}"), 0644)
	os.Chtimes(file1, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000))
	defer os.RemoveAll(project1)

	// 项目2（新）
	project2 := filepath.Join(homeDir, ".claude", "projects", "-project2")
	os.MkdirAll(project2, 0755)
	file2 := filepath.Join(project2, sessionID+".jsonl")
	os.WriteFile(file2, []byte("{}"), 0644)
	os.Chtimes(file2, testutil.TimeFromUnix(2000), testutil.TimeFromUnix(2000))
	defer os.RemoveAll(project2)

	locator := NewSessionLocator()
	path, err := locator.FromSessionID(sessionID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 应该返回最新的文件（project2）
	if path != file2 {
		t.Errorf("Expected newest file %s, got %s", file2, path)
	}
}

func TestFromProjectPath_Success(t *testing.T) {
	// 准备测试环境
	homeDir, _ := os.UserHomeDir()
	projectPath := "/home/yale/work/testproject"
	projectHash := "-home-yale-work-testproject"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	os.MkdirAll(sessionDir, 0755)

	// 创建多个会话文件
	oldSession := filepath.Join(sessionDir, "old-session.jsonl")
	newSession := filepath.Join(sessionDir, "new-session.jsonl")
	os.WriteFile(oldSession, []byte("{}"), 0644)
	os.WriteFile(newSession, []byte("{}"), 0644)
	os.Chtimes(oldSession, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000))
	os.Chtimes(newSession, testutil.TimeFromUnix(2000), testutil.TimeFromUnix(2000))
	defer os.RemoveAll(sessionDir)

	locator := NewSessionLocator()
	path, err := locator.FromProjectPath(projectPath)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if path != newSession {
		t.Errorf("Expected newest session %s, got %s", newSession, path)
	}
}

func TestFromProjectPath_NoSessions(t *testing.T) {
	locator := NewSessionLocator()
	_, err := locator.FromProjectPath("/nonexistent/project")

	if err == nil {
		t.Error("Expected error for project with no sessions")
	}
}
