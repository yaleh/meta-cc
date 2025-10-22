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
	prevHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", overrideHome)
	// Allow SessionLocator to pick up projects root automatically
	_ = os.Setenv("META_CC_PROJECTS_ROOT", filepath.Join(overrideHome, ".claude", "projects"))

	exitCode := m.Run()

	_ = os.Setenv("HOME", prevHome)
	_ = os.RemoveAll(overrideHome)
	os.Exit(exitCode)
}
