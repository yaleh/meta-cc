package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestGetQueryBaseDirSessionScope tests that session scope returns the directory
// containing the most recently modified session file
func TestGetQueryBaseDirSessionScope(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Create temporary test environment
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	// Change to project directory (simulate working in a project)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change to project directory: %v", err)
	}

	// Setup fake Claude Code session directory structure
	// ~/.claude/projects/{hash}/ contains session files
	homeDir := tempDir
	// Use actual path hashing logic (path.Join(...) -> "/-" replacement)
	absProjectDir, _ := filepath.Abs(projectDir)
	projectHash := pathToHash(absProjectDir)
	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session directory: %v", err)
	}

	// Override home directory for SessionLocator
	os.Setenv("META_CC_PROJECTS_ROOT", filepath.Join(homeDir, ".claude", "projects"))
	defer os.Unsetenv("META_CC_PROJECTS_ROOT")

	// Create multiple session files with different modification times
	session1 := filepath.Join(sessionDir, "session-old.jsonl")
	session2 := filepath.Join(sessionDir, "session-current.jsonl")
	session3 := filepath.Join(sessionDir, "session-older.jsonl")

	// Write files with staggered timestamps
	if err := os.WriteFile(session3, []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to write session3: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	if err := os.WriteFile(session1, []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to write session1: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	// session2 should be the newest
	if err := os.WriteFile(session2, []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to write session2: %v", err)
	}

	// Test session scope - should return directory of most recent session
	baseDir, err := getQueryBaseDir("session")
	if err != nil {
		t.Fatalf("getQueryBaseDir(session) failed: %v", err)
	}

	// Verify returned directory is the session directory
	if baseDir != sessionDir {
		t.Errorf("session scope: expected %s, got %s", sessionDir, baseDir)
	}

	// Verify that session scope returns different result than project scope
	// (session = single session dir, project = all sessions dir)
	projectBaseDir, err := getQueryBaseDir("project")
	if err != nil {
		t.Fatalf("getQueryBaseDir(project) failed: %v", err)
	}

	// Both should return the same directory (the session directory)
	// But the logic should be different (session uses newest file, project uses all files)
	if projectBaseDir != sessionDir {
		t.Errorf("project scope: expected %s, got %s", sessionDir, projectBaseDir)
	}
}

// TestGetQueryBaseDirSessionScopeNoSessions tests error handling when no sessions exist
func TestGetQueryBaseDirSessionScopeNoSessions(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Create temporary test environment with no sessions
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "empty-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	// Change to project directory
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change to project directory: %v", err)
	}

	// Override home directory
	os.Setenv("META_CC_PROJECTS_ROOT", filepath.Join(tempDir, ".claude", "projects"))
	defer os.Unsetenv("META_CC_PROJECTS_ROOT")

	// Test session scope with no sessions - should return error
	_, err = getQueryBaseDir("session")
	if err == nil {
		t.Error("expected error for session scope with no sessions, got nil")
	}
}

// TestGetQueryBaseDirSessionScopeSingleSession tests that session scope works with single session
func TestGetQueryBaseDirSessionScopeSingleSession(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Create temporary test environment
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "single-session-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	// Change to project directory
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change to project directory: %v", err)
	}

	// Setup session directory
	homeDir := tempDir
	absProjectDir, _ := filepath.Abs(projectDir)
	projectHash := pathToHash(absProjectDir)
	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session directory: %v", err)
	}

	// Override home directory
	os.Setenv("META_CC_PROJECTS_ROOT", filepath.Join(homeDir, ".claude", "projects"))
	defer os.Unsetenv("META_CC_PROJECTS_ROOT")

	// Create single session file
	session := filepath.Join(sessionDir, "only-session.jsonl")
	if err := os.WriteFile(session, []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to write session: %v", err)
	}

	// Test session scope - should succeed with single session
	baseDir, err := getQueryBaseDir("session")
	if err != nil {
		t.Fatalf("getQueryBaseDir(session) failed: %v", err)
	}

	if baseDir != sessionDir {
		t.Errorf("session scope: expected %s, got %s", sessionDir, baseDir)
	}
}
