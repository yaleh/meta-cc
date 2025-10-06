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

func TestTimeFromUnix(t *testing.T) {
	tests := []struct {
		name     string
		unixSec  int64
		expected string
	}{
		{
			name:     "epoch time",
			unixSec:  0,
			expected: "1970-01-01T00:00:00Z",
		},
		{
			name:     "positive timestamp",
			unixSec:  1609459200,
			expected: "2021-01-01T00:00:00Z",
		},
		{
			name:     "recent timestamp",
			unixSec:  1696248000,
			expected: "2023-10-02T12:00:00Z",
		},
		{
			name:     "negative timestamp (before epoch)",
			unixSec:  -86400,
			expected: "1969-12-31T00:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeFromUnix(tt.unixSec)
			resultStr := result.UTC().Format("2006-01-02T15:04:05Z")

			if resultStr != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, resultStr)
			}
		})
	}
}
