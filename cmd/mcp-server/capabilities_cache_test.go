package main

import (
	"path/filepath"
	"testing"
)

// TestBuildCachePath tests cache path construction for GitHub sources
func TestBuildCachePath(t *testing.T) {
	tests := []struct {
		name     string
		source   GitHubSource
		filename string
		expected string
	}{
		{
			name: "basic source with branch",
			source: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "main",
				Subdir: "commands",
			},
			filename: "meta-errors.md",
			expected: filepath.Join(".capabilities-cache", "github", "yaleh", "meta-cc", "main", "commands", "meta-errors.md"),
		},
		{
			name: "source without subdir",
			source: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "develop",
				Subdir: "",
			},
			filename: "README.md",
			expected: filepath.Join(".capabilities-cache", "github", "yaleh", "meta-cc", "develop", "README.md"),
		},
		{
			name: "source with tag version",
			source: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "v1.0.0",
				Subdir: "capabilities/commands",
			},
			filename: "meta-quality-scan.md",
			expected: filepath.Join(".capabilities-cache", "github", "yaleh", "meta-cc", "v1.0.0", "capabilities/commands", "meta-quality-scan.md"),
		},
		{
			name: "different owner and repo",
			source: GitHubSource{
				Owner:  "community",
				Repo:   "extra-capabilities",
				Branch: "main",
				Subdir: "caps",
			},
			filename: "custom-capability.md",
			expected: filepath.Join(".capabilities-cache", "github", "community", "extra-capabilities", "main", "caps", "custom-capability.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCachePath(tt.source, tt.filename)
			if result != tt.expected {
				t.Errorf("buildCachePath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestLocalCapabilitySource tests the local capability source constant
func TestLocalCapabilitySource(t *testing.T) {
	if LocalCapabilitySource != "capabilities/commands" {
		t.Errorf("LocalCapabilitySource = %v, want %v", LocalCapabilitySource, "capabilities/commands")
	}
}

// TestCapabilityCacheDir tests the cache directory constant
func TestCapabilityCacheDir(t *testing.T) {
	if CapabilityCacheDir != ".capabilities-cache" {
		t.Errorf("CapabilityCacheDir = %v, want %v", CapabilityCacheDir, ".capabilities-cache")
	}
}
