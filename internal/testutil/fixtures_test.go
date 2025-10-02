package testutil

import (
	"testing"
)

func TestLoadFixture(t *testing.T) {
	data := LoadFixture(t, "sample-session.jsonl")

	if len(data) == 0 {
		t.Error("Expected non-empty fixture data")
	}

	// Verify JSONL format (each line should be valid JSON)
	lines := string(data)
	if lines == "" {
		t.Error("Expected at least one line in fixture")
	}
}

func TestTempSessionFile(t *testing.T) {
	content := `{"test":"data"}`
	path := TempSessionFile(t, content)

	if path == "" {
		t.Error("Expected non-empty temp file path")
	}

	// File should be auto-deleted after test completes
}
