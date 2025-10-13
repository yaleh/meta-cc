package main

import (
	"sync"
	"testing"
)

// TestLocalCapabilitySource tests the local capability source constant
func TestLocalCapabilitySource(t *testing.T) {
	if LocalCapabilitySource != "capabilities/commands" {
		t.Errorf("LocalCapabilitySource = %v, want %v", LocalCapabilitySource, "capabilities/commands")
	}
}

// TestSessionCacheDir tests that session cache directory is created properly
func TestSessionCacheDir(t *testing.T) {
	// Reset session cache for test
	sessionCacheDir = ""
	sessionCacheInitErr = nil
	sessionCacheOnce = sync.Once{}

	dir, err := getSessionCacheDir()
	if err != nil {
		t.Fatalf("getSessionCacheDir() error = %v", err)
	}

	if dir == "" {
		t.Error("getSessionCacheDir() returned empty directory")
	}

	// Should contain "claude-session" in path
	if !contains(dir, "claude-session") {
		t.Errorf("getSessionCacheDir() = %v, want to contain 'claude-session'", dir)
	}

	// Should contain ".meta-cc-capabilities" in path
	if !contains(dir, ".meta-cc-capabilities") {
		t.Errorf("getSessionCacheDir() = %v, want to contain '.meta-cc-capabilities'", dir)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
