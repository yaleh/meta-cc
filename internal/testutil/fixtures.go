package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// FixtureDir returns the fixtures directory path
func FixtureDir() string {
	// Try to find the fixtures directory by checking multiple potential locations
	candidates := []string{
		"../../tests/fixtures", // From internal/ packages
		"../tests/fixtures",    // From cmd/ package
		"tests/fixtures",       // From root
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Default to relative path from internal/
	return "../../tests/fixtures"
}

// LoadFixture loads test fixture file content
func LoadFixture(t *testing.T, filename string) []byte {
	t.Helper()

	path := filepath.Join(FixtureDir(), filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", filename, err)
	}

	return data
}

// TempSessionFile creates a temporary session file for testing
func TempSessionFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "session-*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	return tmpFile.Name()
}

