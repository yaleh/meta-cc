package locator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/testutil"
)

func setupProjectsRoot(t *testing.T) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv(codexHomeEnv, filepath.Join(t.TempDir(), "codex-home"))

	dir, err := os.MkdirTemp("", "meta-cc-projects-*")
	if err != nil {
		t.Fatalf("failed to create temp projects dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	t.Setenv(projectsRootEnv, dir)
	return dir
}

func TestFromSessionID_Success(t *testing.T) {
	// 准备测试环境
	projectsRoot := setupProjectsRoot(t)
	projectHash := "-test-project-session-id"
	sessionID := "abc123-def456"

	sessionDir := filepath.Join(projectsRoot, projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"test":"data"}`), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

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
	setupProjectsRoot(t)
	locator := NewSessionLocator()
	_, err := locator.FromSessionID("nonexistent-session-id")

	if err == nil {
		t.Error("Expected error for nonexistent session ID")
	}
}

func TestFromSessionID_CodexHomeSessions(t *testing.T) {
	t.Setenv(projectsRootEnv, "")
	home := t.TempDir()
	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("HOME", home)
	t.Setenv(codexHomeEnv, codexHome)

	sessionID := "codex-session-123"
	sessionDir := filepath.Join(codexHome, "sessions")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create Codex sessions dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"type":"codex-test"}`), 0644); err != nil {
		t.Fatalf("failed to write Codex session file: %v", err)
	}

	locator := NewSessionLocator()
	path, err := locator.FromSessionID(sessionID)
	if err != nil {
		t.Fatalf("Expected Codex session lookup to succeed, got: %v", err)
	}
	if path != sessionFile {
		t.Errorf("Expected Codex session %s, got %s", sessionFile, path)
	}
}

func TestFromSessionID_DefaultCodexSessions(t *testing.T) {
	t.Setenv(projectsRootEnv, "")
	t.Setenv(codexHomeEnv, "")
	home := t.TempDir()
	t.Setenv("HOME", home)

	sessionID := "default-codex-session"
	sessionDir := filepath.Join(home, ".codex", "sessions")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create default Codex sessions dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"type":"codex-test"}`), 0644); err != nil {
		t.Fatalf("failed to write Codex session file: %v", err)
	}

	locator := NewSessionLocator()
	path, err := locator.FromSessionID(sessionID)
	if err != nil {
		t.Fatalf("Expected default Codex session lookup to succeed, got: %v", err)
	}
	if path != sessionFile {
		t.Errorf("Expected Codex session %s, got %s", sessionFile, path)
	}
}

func TestFromSessionID_MissingErrorNamesHosts(t *testing.T) {
	t.Setenv(projectsRootEnv, "")
	t.Setenv(codexHomeEnv, "")
	home := t.TempDir()
	t.Setenv("HOME", home)

	locator := NewSessionLocator()
	_, err := locator.FromSessionID("missing-session")
	if err == nil {
		t.Fatal("Expected error for missing session")
	}
	msg := err.Error()
	for _, want := range []string{HostClaudeCode, HostCodex, ".claude", ".codex"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("missing-session error %q should mention %q", msg, want)
		}
	}
}

func TestFromSessionID_MultipleProjects(t *testing.T) {
	// 准备：在多个项目目录中创建同名会话文件
	projectsRoot := setupProjectsRoot(t)
	sessionID := "shared-session-id"

	// 项目1（旧）
	project1 := filepath.Join(projectsRoot, "-project1")
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
	// 项目2（新）
	project2 := filepath.Join(projectsRoot, "-project2")
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
	projectsRoot := setupProjectsRoot(t)
	// Use temp dir for cross-platform compatibility
	tempDir, err := os.MkdirTemp("", "testproject")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectPath := tempDir
	projectHash := PathToHash(projectPath)

	sessionDir := filepath.Join(projectsRoot, projectHash)
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
	setupProjectsRoot(t)
	locator := NewSessionLocator()
	_, err := locator.FromProjectPath("/nonexistent/project")

	if err == nil {
		t.Error("Expected error for project with no sessions")
	}
}

func TestFromProjectPath_CodexSessionsFallback(t *testing.T) {
	// The non-hashed Codex sessions content-scan fallback has been removed to prevent
	// O(n×file_size) hangs. Codex project sessions are now looked up via the explicit
	// provider path (provider="codex") rather than content-scanning all JSONL files.
	t.Setenv(projectsRootEnv, "")
	home := t.TempDir()
	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("HOME", home)
	t.Setenv(codexHomeEnv, codexHome)
	projectPath := t.TempDir()

	sessionDir := filepath.Join(codexHome, "sessions", "nested")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create Codex sessions dir: %v", err)
	}
	oldSession := filepath.Join(sessionDir, "old.jsonl")
	newSession := filepath.Join(sessionDir, "new.jsonl")
	if err := os.WriteFile(oldSession, []byte(`{"cwd":"`+projectPath+`"}`), 0644); err != nil {
		t.Fatalf("failed to write old session: %v", err)
	}
	if err := os.WriteFile(newSession, []byte(`{"cwd":"`+projectPath+`"}`), 0644); err != nil {
		t.Fatalf("failed to write new session: %v", err)
	}
	if err := os.Chtimes(oldSession, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set old session times: %v", err)
	}
	if err := os.Chtimes(newSession, testutil.TimeFromUnix(2000), testutil.TimeFromUnix(2000)); err != nil {
		t.Fatalf("failed to set new session times: %v", err)
	}

	locator := NewSessionLocator()
	// The content-scan fallback is removed; without a hashed Claude session dir,
	// FromProjectPath returns an error even if matching Codex JSONL files exist.
	// Use provider="codex" via the analysis service to query Codex sessions.
	_, err := locator.FromProjectPath(projectPath)
	if err == nil {
		t.Fatal("Expected error after content-scan fallback removal, but got success")
	}
}

func TestFromProjectPath_CodexSessionsFallbackIgnoresMessageMentions(t *testing.T) {
	// The non-hashed Codex sessions content-scan fallback has been removed.
	// This test verifies that FromProjectPath returns an error without scanning
	// Codex JSONL content when no hashed Claude session directory exists.
	t.Setenv(projectsRootEnv, "")
	home := t.TempDir()
	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("HOME", home)
	t.Setenv(codexHomeEnv, codexHome)
	projectPath := t.TempDir()
	otherProject := t.TempDir()

	sessionDir := filepath.Join(codexHome, "sessions", "nested")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create Codex sessions dir: %v", err)
	}
	mentionOnly := filepath.Join(sessionDir, "mention-only.jsonl")
	matching := filepath.Join(sessionDir, "matching.jsonl")
	if err := os.WriteFile(mentionOnly, []byte(`{"cwd":"`+otherProject+`","message":{"content":"please inspect `+projectPath+`"}}`), 0644); err != nil {
		t.Fatalf("failed to write mention-only session: %v", err)
	}
	if err := os.WriteFile(matching, []byte(`{"payload":{"cwd":"`+projectPath+`"}}`), 0644); err != nil {
		t.Fatalf("failed to write matching session: %v", err)
	}
	if err := os.Chtimes(mentionOnly, testutil.TimeFromUnix(3000), testutil.TimeFromUnix(3000)); err != nil {
		t.Fatalf("failed to set mention-only times: %v", err)
	}
	if err := os.Chtimes(matching, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set matching times: %v", err)
	}

	locator := NewSessionLocator()
	// After removing the content-scan fallback, no Codex session will be found
	// via FromProjectPath; users should use provider="codex" via analysis service.
	_, err := locator.FromProjectPath(projectPath)
	if err == nil {
		t.Fatal("Expected error after content-scan fallback removal, but got success")
	}
}

func TestFromProjectPath_RelativePath(t *testing.T) {
	// Test that relative paths like "." are resolved to absolute paths
	projectsRoot := setupProjectsRoot(t)

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	// Create session directory for current working directory
	projectHash := PathToHash(cwd)
	sessionDir := filepath.Join(projectsRoot, projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// Create a test session file
	testSession := filepath.Join(sessionDir, "test-session.jsonl")
	if err := os.WriteFile(testSession, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write test session: %v", err)
	}

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
	projectsRoot := setupProjectsRoot(t)
	// Use temp dir for cross-platform compatibility
	tempDir, err := os.MkdirTemp("", "testproject")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectPath := tempDir
	projectHash := PathToHash(projectPath)

	sessionDir := filepath.Join(projectsRoot, projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

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
	setupProjectsRoot(t)
	locator := NewSessionLocator()
	sessions, err := locator.AllSessionsFromProject("/nonexistent/project")

	if err == nil {
		t.Error("Expected error for project with no sessions")
	}

	if sessions != nil {
		t.Errorf("Expected nil sessions on error, got: %v", sessions)
	}
}

func TestSessionsFromProject_NoHashedSessions_ReturnsImmediately(t *testing.T) {
	// Set META_CC_PROJECTS_ROOT to a temp dir with no hashed subdir matching the probe path
	projectsRoot := t.TempDir()
	t.Setenv(projectsRootEnv, projectsRoot)
	t.Setenv("HOME", t.TempDir())

	// Set Codex root to a temp dir with 50 synthetic .jsonl files x 10 lines each
	// so findProjectJSONLFilesRecursive would take noticeable time if called
	codexHome := t.TempDir()
	t.Setenv(codexHomeEnv, codexHome)
	codexSessionsDir := filepath.Join(codexHome, "sessions")
	if err := os.MkdirAll(codexSessionsDir, 0755); err != nil {
		t.Fatalf("failed to create codex sessions dir: %v", err)
	}
	// Create 50 synthetic .jsonl files x 10 lines each
	syntheticContent := []byte(`{"cwd":"/some/other/project","type":"user"}` + "\n" +
		`{"cwd":"/some/other/project","type":"assistant"}` + "\n" +
		`{"cwd":"/some/other/project","type":"user"}` + "\n" +
		`{"cwd":"/some/other/project","type":"assistant"}` + "\n" +
		`{"cwd":"/some/other/project","type":"user"}` + "\n" +
		`{"cwd":"/some/other/project","type":"assistant"}` + "\n" +
		`{"cwd":"/some/other/project","type":"user"}` + "\n" +
		`{"cwd":"/some/other/project","type":"assistant"}` + "\n" +
		`{"cwd":"/some/other/project","type":"user"}` + "\n" +
		`{"cwd":"/some/other/project","type":"assistant"}` + "\n")
	for i := 0; i < 50; i++ {
		fname := filepath.Join(codexSessionsDir, fmt.Sprintf("synthetic-%d.jsonl", i))
		if err := os.WriteFile(fname, syntheticContent, 0644); err != nil {
			t.Fatalf("failed to write synthetic file: %v", err)
		}
	}

	// Use a project path that is not in any JSONL file and not in projectsRoot
	noSessionPath := t.TempDir()

	locator := NewSessionLocator()

	start := time.Now()
	_, err := locator.AllSessionsFromProject(noSessionPath)
	elapsed := time.Since(start)

	// Must return an error (no sessions found)
	if err == nil {
		t.Error("Expected error for project with no sessions")
	}
	// Must complete in under 200ms (fallback not called)
	if elapsed >= 200*time.Millisecond {
		t.Errorf("AllSessionsFromProject took %v, expected < 200ms (fallback must not be called)", elapsed)
	}
}

func TestAllSessionsFromProject_RelativePath(t *testing.T) {
	// Test that relative paths are resolved to absolute paths
	projectsRoot := setupProjectsRoot(t)
	cwd, _ := os.Getwd()

	projectHash := PathToHash(cwd)
	sessionDir := filepath.Join(projectsRoot, projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

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
