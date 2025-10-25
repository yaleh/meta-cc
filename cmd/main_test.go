package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	overrideHome := filepath.Join(os.TempDir(), "meta-cc-test-home")
	if err := os.MkdirAll(overrideHome, 0o755); err != nil {
		panic(err)
	}

	// Resolve symlinks/short paths for cross-platform consistency
	// On Windows: C:\Users\RUNNER~1\... -> C:\Users\runneradmin\...
	// On macOS: /var/folders/... -> /private/var/folders/...
	resolvedHome, err := filepath.EvalSymlinks(overrideHome)
	if err != nil {
		// If resolution fails, use original path
		resolvedHome = overrideHome
	}

	projectsRoot := filepath.Join(resolvedHome, ".claude", "projects")
	if err := os.MkdirAll(projectsRoot, 0o755); err != nil {
		panic(err)
	}

	prevHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", resolvedHome)
	_ = os.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	exitCode := m.Run()

	_ = os.Setenv("HOME", prevHome)
	_ = os.RemoveAll(overrideHome)
	os.Exit(exitCode)
}
