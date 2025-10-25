package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSessionFixture creates a Claude session file under the configured META_CC_PROJECTS_ROOT.
func writeSessionFixture(t *testing.T, projectPath, sessionID, content string) string {
	t.Helper()

	projectsRoot := os.Getenv("META_CC_PROJECTS_ROOT")
	if projectsRoot == "" {
		t.Fatal("META_CC_PROJECTS_ROOT must be set for tests")
	}

	// Resolve symlinks for consistent hashing on macOS (/var -> /private/var)
	resolvedPath, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		// If path doesn't exist yet, use original path
		resolvedPath = projectPath
	}

	hash := strings.ReplaceAll(resolvedPath, "\\", "-")
	hash = strings.ReplaceAll(hash, "/", "-")
	hash = strings.ReplaceAll(hash, ":", "-")

	sessionDir := filepath.Join(projectsRoot, hash)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write session fixture: %v", err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(sessionDir) })
	return sessionFile
}
