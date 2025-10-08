package locator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yaleh/meta-cc/internal/testutil"
)

func TestFromSessionID_Success(t *testing.T) {
	// 准备测试环境
	homeDir, _ := os.UserHomeDir()
	projectHash := "-test-project-session-id"
	sessionID := "abc123-def456"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"test":"data"}`), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}
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
	if err := os.MkdirAll(project1, 0755); err != nil {
		t.Fatalf("failed to create project1 dir: %v", err)
	}
	file1 := filepath.Join(project1, sessionID+".jsonl")
	if err := os.WriteFile(file1, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if err := os.Chtimes(file1, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set file1 times: %v", err)
	}
	defer os.RemoveAll(project1)

	// 项目2（新）
	project2 := filepath.Join(homeDir, ".claude", "projects", "-project2")
	if err := os.MkdirAll(project2, 0755); err != nil {
		t.Fatalf("failed to create project2 dir: %v", err)
	}
	file2 := filepath.Join(project2, sessionID+".jsonl")
	if err := os.WriteFile(file2, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}
	if err := os.Chtimes(file2, testutil.TimeFromUnix(2000), testutil.TimeFromUnix(2000)); err != nil {
		t.Fatalf("failed to set file2 times: %v", err)
	}
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
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// 创建多个会话文件
	oldSession := filepath.Join(sessionDir, "old-session.jsonl")
	newSession := filepath.Join(sessionDir, "new-session.jsonl")
	if err := os.WriteFile(oldSession, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write old session: %v", err)
	}
	if err := os.WriteFile(newSession, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write new session: %v", err)
	}
	if err := os.Chtimes(oldSession, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set old session times: %v", err)
	}
	if err := os.Chtimes(newSession, testutil.TimeFromUnix(2000), testutil.TimeFromUnix(2000)); err != nil {
		t.Fatalf("failed to set new session times: %v", err)
	}
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

func TestFromProjectPath_RelativePath(t *testing.T) {
	// Test that relative paths like "." are resolved to absolute paths
	homeDir, _ := os.UserHomeDir()

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	// Create session directory for current working directory
	projectHash := pathToHash(cwd)
	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// Create a test session file
	testSession := filepath.Join(sessionDir, "test-session.jsonl")
	if err := os.WriteFile(testSession, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write test session: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	locator := NewSessionLocator()

	// Test with relative path "."
	pathFromRelative, err := locator.FromProjectPath(".")
	if err != nil {
		t.Fatalf("Expected no error with relative path '.', got: %v", err)
	}

	// Test with absolute path
	pathFromAbsolute, err := locator.FromProjectPath(cwd)
	if err != nil {
		t.Fatalf("Expected no error with absolute path, got: %v", err)
	}

	// Both should return the same session file
	if pathFromRelative != pathFromAbsolute {
		t.Errorf("Relative path '.' and absolute path should resolve to same session.\nGot: %s\nExpected: %s",
			pathFromRelative, pathFromAbsolute)
	}

	if pathFromRelative != testSession {
		t.Errorf("Expected session %s, got %s", testSession, pathFromRelative)
	}
}

func TestAllSessionsFromProject_Success(t *testing.T) {
	// Test that AllSessionsFromProject returns all session files for a project
	homeDir, _ := os.UserHomeDir()
	projectPath := "/home/yale/work/testproject"
	projectHash := "-home-yale-work-testproject"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	// Create multiple session files
	session1 := filepath.Join(sessionDir, "session-1.jsonl")
	session2 := filepath.Join(sessionDir, "session-2.jsonl")
	session3 := filepath.Join(sessionDir, "session-3.jsonl")
	if err := os.WriteFile(session1, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write session1: %v", err)
	}
	if err := os.WriteFile(session2, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write session2: %v", err)
	}
	if err := os.WriteFile(session3, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write session3: %v", err)
	}

	locator := NewSessionLocator()
	sessions, err := locator.AllSessionsFromProject(projectPath)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(sessions))
	}

	// Verify all sessions are returned
	sessionMap := make(map[string]bool)
	for _, s := range sessions {
		sessionMap[s] = true
	}

	if !sessionMap[session1] {
		t.Errorf("Expected to find session1: %s", session1)
	}
	if !sessionMap[session2] {
		t.Errorf("Expected to find session2: %s", session2)
	}
	if !sessionMap[session3] {
		t.Errorf("Expected to find session3: %s", session3)
	}
}

func TestAllSessionsFromProject_NoSessions(t *testing.T) {
	locator := NewSessionLocator()
	sessions, err := locator.AllSessionsFromProject("/nonexistent/project")

	if err == nil {
		t.Error("Expected error for project with no sessions")
	}

	if sessions != nil {
		t.Errorf("Expected nil sessions on error, got: %v", sessions)
	}
}

func TestAllSessionsFromProject_RelativePath(t *testing.T) {
	// Test that relative paths are resolved to absolute paths
	homeDir, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()

	projectHash := pathToHash(cwd)
	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	// Create test sessions
	session1 := filepath.Join(sessionDir, "test1.jsonl")
	session2 := filepath.Join(sessionDir, "test2.jsonl")
	if err := os.WriteFile(session1, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write session1: %v", err)
	}
	if err := os.WriteFile(session2, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write session2: %v", err)
	}

	locator := NewSessionLocator()

	// Test with relative path "."
	sessionsFromRelative, err := locator.AllSessionsFromProject(".")
	if err != nil {
		t.Fatalf("Expected no error with relative path, got: %v", err)
	}

	// Test with absolute path
	sessionsFromAbsolute, err := locator.AllSessionsFromProject(cwd)
	if err != nil {
		t.Fatalf("Expected no error with absolute path, got: %v", err)
	}

	// Both should return the same sessions
	if len(sessionsFromRelative) != len(sessionsFromAbsolute) {
		t.Errorf("Relative and absolute paths should return same number of sessions. Got %d vs %d",
			len(sessionsFromRelative), len(sessionsFromAbsolute))
	}

	if len(sessionsFromRelative) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessionsFromRelative))
	}
}
