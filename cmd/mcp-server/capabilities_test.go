package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseCapabilitySources tests parsing environment variable into source list
func TestParseCapabilitySources(t *testing.T) {
	// Use platform-specific path separator for test data
	sep := string(os.PathListSeparator)

	tests := []struct {
		name     string
		envVar   string
		expected int
		sources  []string
	}{
		{
			name:     "empty env var",
			envVar:   "",
			expected: 0,
		},
		{
			name:     "single local source",
			envVar:   "/home/user/.config/meta-cc/capabilities",
			expected: 1,
			sources:  []string{"/home/user/.config/meta-cc/capabilities"},
		},
		{
			name:     "single GitHub source",
			envVar:   "yaleh/meta-cc-capabilities",
			expected: 1,
			sources:  []string{"yaleh/meta-cc-capabilities"},
		},
		{
			name:     "multiple sources with colon separator",
			envVar:   "/home/user/caps" + sep + "yaleh/meta-cc-capabilities" + sep + "./local",
			expected: 3,
			sources:  []string{"/home/user/caps", "yaleh/meta-cc-capabilities", "./local"},
		},
		{
			name:     "sources with whitespace",
			envVar:   " /home/user/caps " + sep + " yaleh/meta-cc " + sep + " ./local ",
			expected: 3,
			sources:  []string{"/home/user/caps", "yaleh/meta-cc", "./local"},
		},
		{
			name:     "empty segments ignored",
			envVar:   "/home/user/caps" + sep + sep + "yaleh/meta-cc",
			expected: 2,
			sources:  []string{"/home/user/caps", "yaleh/meta-cc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCapabilitySources(tt.envVar)
			if len(result) != tt.expected {
				t.Errorf("expected %d sources, got %d", tt.expected, len(result))
			}

			if tt.sources != nil {
				for i, expected := range tt.sources {
					if result[i].Location != expected {
						t.Errorf("source[%d]: expected %q, got %q", i, expected, result[i].Location)
					}
					if result[i].Priority != i {
						t.Errorf("source[%d]: expected priority %d, got %d", i, i, result[i].Priority)
					}
				}
			}
		})
	}
}

// TestDetectSourceType tests source type detection
func TestDetectSourceType(t *testing.T) {
	tests := []struct {
		name     string
		location string
		expected SourceType
	}{
		{
			name:     "absolute path",
			location: "/home/user/.config/meta-cc/capabilities",
			expected: SourceTypeLocal,
		},
		{
			name:     "relative path with dot",
			location: "./capabilities",
			expected: SourceTypeLocal,
		},
		{
			name:     "relative path with double dot",
			location: "../capabilities",
			expected: SourceTypeLocal,
		},
		{
			name:     "GitHub repo format",
			location: "yaleh/meta-cc-capabilities",
			expected: SourceTypeGitHub,
		},
		{
			name:     "GitHub repo with subdirectory",
			location: "yaleh/meta-cc-capabilities/commands",
			expected: SourceTypeGitHub,
		},
		{
			name:     "simple directory name (local)",
			location: "capabilities",
			expected: SourceTypeLocal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectSourceType(tt.location)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestParseFrontmatter tests frontmatter extraction from markdown files
func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expected    CapabilityMetadata
	}{
		{
			name: "valid frontmatter",
			content: `---
name: meta-errors
description: Analyze error patterns and prevention recommendations.
keywords: error, debug, troubleshooting, diagnostics
category: diagnostics
---

# Content here`,
			expectError: false,
			expected: CapabilityMetadata{
				Name:        "meta-errors",
				Description: "Analyze error patterns and prevention recommendations.",
				Keywords:    []string{"error", "debug", "troubleshooting", "diagnostics"},
				Category:    "diagnostics",
			},
		},
		{
			name: "frontmatter without keywords",
			content: `---
name: meta-coach
description: Get workflow optimization recommendations.
category: workflow
---

# Content`,
			expectError: false,
			expected: CapabilityMetadata{
				Name:        "meta-coach",
				Description: "Get workflow optimization recommendations.",
				Keywords:    []string{},
				Category:    "workflow",
			},
		},
		{
			name:        "no frontmatter",
			content:     "# Just markdown content",
			expectError: true,
		},
		{
			name: "missing name field",
			content: `---
description: Missing name
category: test
---`,
			expectError: true,
		},
		{
			name: "malformed YAML",
			content: `---
name: broken
keywords: [ unclosed
---`,
			expectError: true,
		},
		{
			name:    "CRLF line endings (Windows)",
			content: "---\r\nname: test-crlf\r\ndescription: Test CRLF line endings.\r\nkeywords: test, crlf, windows\r\ncategory: test\r\n---\r\n\r\n# Test Content",
			expected: CapabilityMetadata{
				Name:        "test-crlf",
				Description: "Test CRLF line endings.",
				Category:    "test",
				Keywords:    []string{"test", "crlf", "windows"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFrontmatter(tt.content)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Name != tt.expected.Name {
				t.Errorf("name: expected %q, got %q", tt.expected.Name, result.Name)
			}
			if result.Description != tt.expected.Description {
				t.Errorf("description: expected %q, got %q", tt.expected.Description, result.Description)
			}
			if result.Category != tt.expected.Category {
				t.Errorf("category: expected %q, got %q", tt.expected.Category, result.Category)
			}
			if len(result.Keywords) != len(tt.expected.Keywords) {
				t.Errorf("keywords length: expected %d, got %d", len(tt.expected.Keywords), len(result.Keywords))
			} else {
				for i, kw := range tt.expected.Keywords {
					if result.Keywords[i] != kw {
						t.Errorf("keywords[%d]: expected %q, got %q", i, kw, result.Keywords[i])
					}
				}
			}
		})
	}
}

// TestLoadLocalCapabilities tests loading capabilities from local directory
func TestLoadLocalCapabilities(t *testing.T) {
	// Use test fixtures
	fixturesPath := filepath.Join("..", "..", "tests", "fixtures", "capabilities")
	if _, err := os.Stat(fixturesPath); os.IsNotExist(err) {
		t.Skip("test fixtures not found")
	}

	capabilities, err := loadLocalCapabilities(fixturesPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find valid-capability.md and another-capability.md
	// Should skip invalid-frontmatter.md, no-frontmatter.md, missing-name.md
	if len(capabilities) < 2 {
		t.Errorf("expected at least 2 valid capabilities, got %d", len(capabilities))
	}

	// Verify one of the capabilities
	found := false
	for _, cap := range capabilities {
		if cap.Name == "meta-errors" {
			found = true
			if cap.Category != "diagnostics" {
				t.Errorf("expected category 'diagnostics', got %q", cap.Category)
			}
			if len(cap.Keywords) != 4 {
				t.Errorf("expected 4 keywords, got %d", len(cap.Keywords))
			}
		}
	}

	if !found {
		t.Error("expected to find 'meta-errors' capability")
	}
}

// TestSourcePriorityMerging tests priority-based merging of capabilities
func TestSourcePriorityMerging(t *testing.T) {
	// Create temp directories with overlapping capabilities
	tempDir1, err := os.MkdirTemp("", "cap1-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir1)

	tempDir2, err := os.MkdirTemp("", "cap2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir2)

	// Create capability with same name but different description in both dirs
	cap1Content := `---
name: shared-cap
description: Description from source 1
category: test
---
# Content`

	cap2Content := `---
name: shared-cap
description: Description from source 2
category: test
---
# Content`

	if err := os.WriteFile(filepath.Join(tempDir1, "cap.md"), []byte(cap1Content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir2, "cap.md"), []byte(cap2Content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test priority: source1 should override source2
	sources := []CapabilitySource{
		{Type: SourceTypeLocal, Location: tempDir1, Priority: 0},
		{Type: SourceTypeLocal, Location: tempDir2, Priority: 1},
	}

	index, err := mergeSources(sources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cap, exists := index["shared-cap"]
	if !exists {
		t.Fatal("expected 'shared-cap' to exist in index")
	}

	// Should have description from source 1 (higher priority)
	if cap.Description != "Description from source 1" {
		t.Errorf("expected description from source 1, got %q", cap.Description)
	}

	if cap.Source != tempDir1 {
		t.Errorf("expected source to be %q, got %q", tempDir1, cap.Source)
	}
}

// TestInvalidSourceHandling tests error handling for invalid sources
func TestInvalidSourceHandling(t *testing.T) {
	tests := []struct {
		name      string
		sources   []CapabilitySource
		expectErr bool
	}{
		{
			name: "nonexistent local directory",
			sources: []CapabilitySource{
				{Type: SourceTypeLocal, Location: "/nonexistent/path", Priority: 0},
			},
			expectErr: true,
		},
		{
			name: "valid source",
			sources: []CapabilitySource{
				{Type: SourceTypeLocal, Location: filepath.Join("..", "..", "tests", "fixtures", "capabilities"), Priority: 0},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mergeSources(tt.sources)
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestEmptySourcesHandling tests handling of empty source list
func TestEmptySourcesHandling(t *testing.T) {
	sources := []CapabilitySource{}
	index, err := mergeSources(sources)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(index) != 0 {
		t.Errorf("expected empty index, got %d capabilities", len(index))
	}
}

// TestHasLocalSources tests detection of local sources
// DISABLED: Function removed due to session-scoped cache refactoring
// func TestHasLocalSources(t *testing.T) { ... }

// TestGetCapabilityIndexCaching tests caching behavior
func TestGetCapabilityIndexCaching(t *testing.T) {
	// Use test fixtures
	fixturesPath := filepath.Join("..", "..", "tests", "fixtures", "capabilities")
	if _, err := os.Stat(fixturesPath); os.IsNotExist(err) {
		t.Skip("test fixtures not found")
	}

	sources := []CapabilitySource{
		{Type: SourceTypeLocal, Location: fixturesPath, Priority: 0},
	}

	// First call - should populate cache (but not use it because local source)
	index1, err := getCapabilityIndex(sources, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(index1) == 0 {
		t.Error("expected non-empty capability index")
	}

	// Second call - should still bypass cache (local sources)
	index2, err := getCapabilityIndex(sources, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(index2) != len(index1) {
		t.Errorf("expected same number of capabilities, got %d vs %d", len(index1), len(index2))
	}
}

// TestGetCapabilityIndexDisableCache tests cache bypass
// DISABLED: Function removed due to session-scoped cache refactoring
// func TestGetCapabilityIndexDisableCache(t *testing.T) { ... }

// TestIsCacheValid tests cache validation
// DISABLED: Function removed due to session-scoped cache refactoring
// func TestIsCacheValid(t *testing.T) { ... }

// TestGetCapabilityContent tests retrieval of capability content
func TestGetCapabilityContent(t *testing.T) {
	// Use test fixtures
	fixturesPath := filepath.Join("..", "..", "tests", "fixtures", "capabilities")
	if _, err := os.Stat(fixturesPath); os.IsNotExist(err) {
		t.Skip("test fixtures not found")
	}

	tests := []struct {
		name        string
		capName     string
		sources     []CapabilitySource
		expectError bool
		contains    string
	}{
		{
			name:    "existing capability from local source",
			capName: "meta-errors",
			sources: []CapabilitySource{
				{Type: SourceTypeLocal, Location: fixturesPath, Priority: 0},
			},
			expectError: false,
			contains:    "name: meta-errors",
		},
		{
			name:    "nonexistent capability",
			capName: "nonexistent-cap",
			sources: []CapabilitySource{
				{Type: SourceTypeLocal, Location: fixturesPath, Priority: 0},
			},
			expectError: true,
		},
		{
			name:        "empty sources",
			capName:     "meta-errors",
			sources:     []CapabilitySource{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := getCapabilityContent(tt.capName, tt.sources)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.contains != "" && !strings.Contains(content, tt.contains) {
				t.Errorf("expected content to contain %q", tt.contains)
			}
		})
	}
}

// TestGetCapabilityContentPriority tests priority-based capability retrieval
func TestGetCapabilityContentPriority(t *testing.T) {
	// Create temp directories with overlapping capabilities
	tempDir1, err := os.MkdirTemp("", "cap1-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir1)

	tempDir2, err := os.MkdirTemp("", "cap2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir2)

	// Create capability with same name but different content in both dirs
	cap1Content := `---
name: test-cap
description: From source 1
category: test
---
# Source 1 Content`

	cap2Content := `---
name: test-cap
description: From source 2
category: test
---
# Source 2 Content`

	if err := os.WriteFile(filepath.Join(tempDir1, "test-cap.md"), []byte(cap1Content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir2, "test-cap.md"), []byte(cap2Content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test priority: source1 (priority 0) should be returned
	sources := []CapabilitySource{
		{Type: SourceTypeLocal, Location: tempDir1, Priority: 0},
		{Type: SourceTypeLocal, Location: tempDir2, Priority: 1},
	}

	content, err := getCapabilityContent("test-cap", sources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(content, "Source 1 Content") {
		t.Error("expected content from source 1 (higher priority)")
	}

	if strings.Contains(content, "Source 2 Content") {
		t.Error("unexpected content from source 2 (lower priority)")
	}
}

// TestReadLocalCapability tests reading capability from local filesystem
func TestReadLocalCapability(t *testing.T) {
	// Use test fixtures
	fixturesPath := filepath.Join("..", "..", "tests", "fixtures", "capabilities")
	if _, err := os.Stat(fixturesPath); os.IsNotExist(err) {
		t.Skip("test fixtures not found")
	}

	tests := []struct {
		name        string
		capName     string
		path        string
		expectError bool
		contains    string
	}{
		{
			name:        "existing capability",
			capName:     "meta-errors",
			path:        fixturesPath,
			expectError: false,
			contains:    "name: meta-errors",
		},
		{
			name:        "nonexistent capability",
			capName:     "nonexistent",
			path:        fixturesPath,
			expectError: true,
		},
		{
			name:        "invalid path",
			capName:     "test",
			path:        "/nonexistent/path",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := readLocalCapability(tt.capName, tt.path)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.contains != "" && !strings.Contains(content, tt.contains) {
				t.Errorf("expected content to contain %q", tt.contains)
			}
		})
	}
}

// TestExecuteGetCapabilityTool tests the get_capability MCP tool
func TestExecuteGetCapabilityTool(t *testing.T) {
	// Use test fixtures
	fixturesPath := filepath.Join("..", "..", "tests", "fixtures", "capabilities")
	if _, err := os.Stat(fixturesPath); os.IsNotExist(err) {
		t.Skip("test fixtures not found")
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		contains    string
	}{
		{
			name: "retrieve existing capability",
			args: map[string]interface{}{
				"name":     "meta-errors",
				"_sources": fixturesPath,
			},
			expectError: false,
			contains:    "meta-errors",
		},
		{
			name: "nonexistent capability",
			args: map[string]interface{}{
				"name":     "nonexistent",
				"_sources": fixturesPath,
			},
			expectError: true,
		},
		{
			name: "missing name parameter",
			args: map[string]interface{}{
				"_sources": fixturesPath,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executeGetCapabilityTool(tt.args)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("expected result to contain %q", tt.contains)
			}

			// Verify JSON structure
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(result), &response); err != nil {
				t.Errorf("failed to parse JSON response: %v", err)
				return
			}

			// Check required fields
			if mode, ok := response["mode"].(string); !ok || mode != "inline" {
				t.Errorf("expected mode 'inline', got %v", response["mode"])
			}

			if name, ok := response["name"].(string); !ok {
				t.Error("expected 'name' field in response")
			} else if name != tt.args["name"] {
				t.Errorf("expected name %q, got %q", tt.args["name"], name)
			}

			if _, ok := response["content"].(string); !ok {
				t.Error("expected 'content' field in response")
			}
		})
	}
}

// TestParseGitHubSource tests parsing of GitHub source strings with @ symbol
func TestParseGitHubSource(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    GitHubSource
		expectError bool
	}{
		{
			name:  "branch with subdirectory",
			input: "yaleh/meta-cc@main/commands",
			expected: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "main",
				Subdir: "commands",
			},
			expectError: false,
		},
		{
			name:  "tag with subdirectory",
			input: "yaleh/meta-cc@v1.0.0/commands",
			expected: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "v1.0.0",
				Subdir: "commands",
			},
			expectError: false,
		},
		{
			name:  "branch without subdirectory",
			input: "yaleh/meta-cc@develop",
			expected: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "develop",
				Subdir: "",
			},
			expectError: false,
		},
		{
			name:  "no @ symbol with subdirectory (defaults to main)",
			input: "yaleh/meta-cc/commands",
			expected: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "main",
				Subdir: "commands",
			},
			expectError: false,
		},
		{
			name:  "no @ symbol without subdirectory (defaults to main)",
			input: "yaleh/meta-cc",
			expected: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "main",
				Subdir: "",
			},
			expectError: false,
		},
		{
			name:  "commit hash",
			input: "yaleh/meta-cc@abc123def",
			expected: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "abc123def",
				Subdir: "",
			},
			expectError: false,
		},
		{
			name:        "invalid format - missing repo",
			input:       "yaleh",
			expectError: true,
		},
		{
			name:        "invalid format - only slash",
			input:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGitHubSource(tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Owner != tt.expected.Owner {
				t.Errorf("Owner: expected %q, got %q", tt.expected.Owner, result.Owner)
			}
			if result.Repo != tt.expected.Repo {
				t.Errorf("Repo: expected %q, got %q", tt.expected.Repo, result.Repo)
			}
			if result.Branch != tt.expected.Branch {
				t.Errorf("Branch: expected %q, got %q", tt.expected.Branch, result.Branch)
			}
			if result.Subdir != tt.expected.Subdir {
				t.Errorf("Subdir: expected %q, got %q", tt.expected.Subdir, result.Subdir)
			}
		})
	}
}

// TestBuildJsDelivrURL tests jsDelivr URL generation
func TestBuildJsDelivrURL(t *testing.T) {
	tests := []struct {
		name     string
		source   GitHubSource
		filename string
		expected string
	}{
		{
			name: "main branch with subdirectory",
			source: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "main",
				Subdir: "commands",
			},
			filename: "meta-errors.md",
			expected: "https://cdn.jsdelivr.net/gh/yaleh/meta-cc@main/commands/meta-errors.md",
		},
		{
			name: "tag with subdirectory",
			source: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "v1.0.0",
				Subdir: "commands",
			},
			filename: "meta-errors.md",
			expected: "https://cdn.jsdelivr.net/gh/yaleh/meta-cc@v1.0.0/commands/meta-errors.md",
		},
		{
			name: "branch without subdirectory",
			source: GitHubSource{
				Owner:  "yaleh",
				Repo:   "meta-cc",
				Branch: "develop",
				Subdir: "",
			},
			filename: "README.md",
			expected: "https://cdn.jsdelivr.net/gh/yaleh/meta-cc@develop/README.md",
		},
		{
			name: "commit hash without subdirectory",
			source: GitHubSource{
				Owner:  "community",
				Repo:   "extras",
				Branch: "abc123def",
				Subdir: "",
			},
			filename: "capability.md",
			expected: "https://cdn.jsdelivr.net/gh/community/extras@abc123def/capability.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildJsDelivrURL(tt.source, tt.filename)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestDetectVersionType tests version type detection (branch vs tag)
// DISABLED: Function removed due to session-scoped cache refactoring (no TTL needed)
// func TestDetectVersionType(t *testing.T) { ... }

// TestGetCacheTTL tests cache TTL based on version type
// DISABLED: Function removed due to session-scoped cache refactoring (no TTL needed)
// func TestGetCacheTTL(t *testing.T) { ... }

// TestDefaultSourceIsPackage verifies that the default source is package URL when no env var is set
func TestDefaultSourceIsPackage(t *testing.T) {
	// Clear environment variable
	oldEnv := os.Getenv("META_CC_CAPABILITY_SOURCES")
	os.Unsetenv("META_CC_CAPABILITY_SOURCES")
	defer func() {
		if oldEnv != "" {
			os.Setenv("META_CC_CAPABILITY_SOURCES", oldEnv)
		}
	}()

	// Test list_capabilities - should use package default
	// This will actually try to download from GitHub releases
	args := map[string]interface{}{}
	result, err := executeListCapabilitiesTool(args)

	// Since we're actually connecting to GitHub releases in the test, check the result
	if err != nil {
		// If it fails (e.g., network issue), that's okay for a test
		t.Logf("Package download test failed (expected in CI): %v", err)
	} else {
		// If it succeeds, verify we got JSON output with capabilities
		if !strings.Contains(result, "capabilities") {
			t.Error("expected capabilities in result")
		}
	}
}

// TestDefaultSourceConstant verifies the default source constant value
func TestDefaultSourceConstant(t *testing.T) {
	expected := "https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz"
	if DefaultCapabilitySource != expected {
		t.Errorf("expected DefaultCapabilitySource to be %q, got %q", expected, DefaultCapabilitySource)
	}
}

// TestIsServerError tests detection of 5xx server errors
func TestIsServerError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "500 error",
			err:      fmt.Errorf("jsDelivr returned status 500"),
			expected: true,
		},
		{
			name:     "503 service unavailable",
			err:      fmt.Errorf("jsDelivr returned status 503 (server error)"),
			expected: true,
		},
		{
			name:     "404 not found",
			err:      fmt.Errorf("jsDelivr returned status 404"),
			expected: false,
		},
		{
			name:     "200 success",
			err:      fmt.Errorf("jsDelivr returned status 200"),
			expected: false,
		},
		{
			name:     "network error",
			err:      fmt.Errorf("no such host"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isServerError(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsNetworkUnreachableError tests detection of network errors
func TestIsNetworkUnreachableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "no such host",
			err:      fmt.Errorf("no such host"),
			expected: true,
		},
		{
			name:     "connection refused",
			err:      fmt.Errorf("connection refused"),
			expected: true,
		},
		{
			name:     "network is unreachable",
			err:      fmt.Errorf("network is unreachable"),
			expected: true,
		},
		{
			name:     "404 error",
			err:      fmt.Errorf("jsDelivr returned status 404"),
			expected: false,
		},
		{
			name:     "500 error",
			err:      fmt.Errorf("jsDelivr returned status 500"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNetworkUnreachableError(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsCacheStale tests stale cache detection
// DISABLED: Function removed due to session-scoped cache refactoring (no TTL/staleness concept)
// func TestIsCacheStale(t *testing.T) { ... }

// Test package source detection
func TestDetectSourceTypePackage(t *testing.T) {
	tests := []struct {
		location string
		expected SourceType
	}{
		// Package sources
		{"https://example.com/caps.tar.gz", SourceTypePackage},
		{"/path/to/caps.tar.gz", SourceTypePackage},
		{"~/downloads/caps.tgz", SourceTypePackage},
		{"./local/caps.tar.gz", SourceTypePackage},

		// GitHub sources
		{"yaleh/meta-cc", SourceTypeGitHub},
		{"yaleh/meta-cc@main/commands", SourceTypeGitHub},

		// Local sources
		{"/path/to/dir", SourceTypeLocal},
		{"~/config/caps", SourceTypeLocal},
		{"./relative/path", SourceTypeLocal},
	}

	for _, tt := range tests {
		t.Run(tt.location, func(t *testing.T) {
			result := detectSourceType(tt.location)
			if result != tt.expected {
				t.Errorf("detectSourceType(%q) = %v, want %v",
					tt.location, result, tt.expected)
			}
		})
	}
}

// Test package cache directory generation
func TestGetPackageCacheDir(t *testing.T) {
	tests := []struct {
		location string
		wantErr  bool
	}{
		{"https://example.com/caps.tar.gz", false},
		{"/local/path/caps.tar.gz", false},
	}

	for _, tt := range tests {
		t.Run(tt.location, func(t *testing.T) {
			cacheDir, err := getPackageCacheDir(tt.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPackageCacheDir(%q) error = %v, wantErr %v",
					tt.location, err, tt.wantErr)
			}
			if !tt.wantErr && cacheDir == "" {
				t.Errorf("getPackageCacheDir(%q) returned empty path", tt.location)
			}
			if !tt.wantErr {
				// Verify cache dir contains "packages" and hash
				if !strings.Contains(cacheDir, "packages") {
					t.Errorf("cache dir should contain 'packages': %s", cacheDir)
				}
			}
		})
	}
}

// Helper: create test package
func createTestPackage(t *testing.T, path string, files map[string]string) {
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create package: %v", err)
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	for name, content := range files {
		// Write header
		header := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("failed to write header: %v", err)
		}

		// Write content
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write content: %v", err)
		}
	}
}

// Test package extraction
func TestExtractPackage(t *testing.T) {
	// Create test package
	tmpDir := t.TempDir()
	packagePath := filepath.Join(tmpDir, "test.tar.gz")

	createTestPackage(t, packagePath, map[string]string{
		"commands/test.md": "# Test Capability",
		"agents/test.md":   "# Test Agent",
	})

	// Extract package
	extractDir := filepath.Join(tmpDir, "extract")
	err := extractPackage(packagePath, extractDir)
	if err != nil {
		t.Fatalf("extractPackage failed: %v", err)
	}

	// Verify extracted files
	testFile := filepath.Join(extractDir, "commands", "test.md")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("expected file not found: %s", testFile)
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != "# Test Capability" {
		t.Errorf("unexpected content: %s", string(content))
	}
}

// Test loading capabilities from package
func TestLoadPackageCapabilities(t *testing.T) {
	// Create test package
	tmpDir := t.TempDir()
	packagePath := filepath.Join(tmpDir, "test.tar.gz")

	createTestPackage(t, packagePath, map[string]string{
		"commands/meta-test.md": `---
name: meta-test
description: Test capability.
keywords: test
category: test
---
# Test`,
	})

	// Load capabilities from package
	caps, err := loadPackageCapabilities(packagePath)
	if err != nil {
		t.Fatalf("loadPackageCapabilities failed: %v", err)
	}

	// Verify loaded capabilities
	if len(caps) != 1 {
		t.Errorf("expected 1 capability, got %d", len(caps))
	}
	if caps[0].Name != "meta-test" {
		t.Errorf("expected name 'meta-test', got %q", caps[0].Name)
	}
}

// Test package source integration with mergeSources
func TestMergeSourcesWithPackage(t *testing.T) {
	// Create test package
	tmpDir := t.TempDir()
	packagePath := filepath.Join(tmpDir, "test.tar.gz")

	createTestPackage(t, packagePath, map[string]string{
		"commands/meta-package-test.md": `---
name: meta-package-test
description: Package test capability.
keywords: package, test
category: test
---
# Package Test`,
	})

	// Test merging with package source
	sources := []CapabilitySource{
		{Type: SourceTypePackage, Location: packagePath, Priority: 0},
	}

	index, err := mergeSources(sources)
	if err != nil {
		t.Fatalf("mergeSources failed: %v", err)
	}

	// Verify capability loaded
	if len(index) != 1 {
		t.Errorf("expected 1 capability, got %d", len(index))
	}

	cap, exists := index["meta-package-test"]
	if !exists {
		t.Error("expected capability 'meta-package-test' not found")
	}
	if cap.Description != "Package test capability." {
		t.Errorf("unexpected description: %s", cap.Description)
	}
}

// TestGetPackageTTL tests TTL calculation for packages
// DISABLED: Function removed due to session-scoped cache refactoring (no TTL needed)
// func TestGetPackageTTL(t *testing.T) { ... }

// TestIsReleasePackage tests release package detection
// DISABLED: Function removed due to session-scoped cache refactoring (no TTL needed)
// func TestIsReleasePackage(t *testing.T) { ... }

// TestPackageCacheMetadata tests cache metadata persistence
// DISABLED: Function removed due to session-scoped cache refactoring (no metadata persistence)
// func TestPackageCacheMetadata(t *testing.T) { ... }

// TestCacheValidation tests cache validation logic
// DISABLED: Function removed due to session-scoped cache refactoring (no validation/TTL logic)
// func TestCacheValidation(t *testing.T) { ... }
