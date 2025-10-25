package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// pathToHash replicates the hash logic from internal/locator
// Used by tests to create mock session directories
func pathToHash(path string) string {
	hash := strings.ReplaceAll(path, "\\", "-")
	hash = strings.ReplaceAll(hash, "/", "-")
	hash = strings.ReplaceAll(hash, ":", "-")
	return hash
}

// setupTestSessionDir creates a mock Claude projects directory structure
// for testing SessionLocator integration. Returns the project path.
func setupTestSessionDir(t *testing.T, testData string) string {
	t.Helper()

	// Create temp directory as mock Claude projects root
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	// Create temp directory as project path
	projectPath := t.TempDir()

	// Calculate project hash (same logic as SessionLocator)
	projectHash := pathToHash(projectPath)

	// Create session directory: {projectsRoot}/{hash}/
	sessionDir := filepath.Join(projectsRoot, projectHash)
	err := os.MkdirAll(sessionDir, 0755)
	require.NoError(t, err)

	// Create session file
	sessionFile := filepath.Join(sessionDir, "test-session.jsonl")
	err = os.WriteFile(sessionFile, []byte(testData), 0644)
	require.NoError(t, err)

	return projectPath
}
