package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestParseCapabilitySources tests parsing environment variable into source list
func TestParseCapabilitySources(t *testing.T) {
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
			envVar:   "/home/user/caps:yaleh/meta-cc-capabilities:./local",
			expected: 3,
			sources:  []string{"/home/user/caps", "yaleh/meta-cc-capabilities", "./local"},
		},
		{
			name:     "sources with whitespace",
			envVar:   " /home/user/caps : yaleh/meta-cc : ./local ",
			expected: 3,
			sources:  []string{"/home/user/caps", "yaleh/meta-cc", "./local"},
		},
		{
			name:     "empty segments ignored",
			envVar:   "/home/user/caps::yaleh/meta-cc",
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
func TestHasLocalSources(t *testing.T) {
	tests := []struct {
		name     string
		sources  []CapabilitySource
		expected bool
	}{
		{
			name:     "no sources",
			sources:  []CapabilitySource{},
			expected: false,
		},
		{
			name: "only local sources",
			sources: []CapabilitySource{
				{Type: SourceTypeLocal, Location: "/path/to/dir", Priority: 0},
			},
			expected: true,
		},
		{
			name: "only GitHub sources",
			sources: []CapabilitySource{
				{Type: SourceTypeGitHub, Location: "owner/repo", Priority: 0},
			},
			expected: false,
		},
		{
			name: "mixed sources with local",
			sources: []CapabilitySource{
				{Type: SourceTypeGitHub, Location: "owner/repo", Priority: 0},
				{Type: SourceTypeLocal, Location: "./local", Priority: 1},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasLocalSources(tt.sources)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

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
func TestGetCapabilityIndexDisableCache(t *testing.T) {
	// Use test fixtures
	fixturesPath := filepath.Join("..", "..", "tests", "fixtures", "capabilities")
	if _, err := os.Stat(fixturesPath); os.IsNotExist(err) {
		t.Skip("test fixtures not found")
	}

	// Use GitHub source (would normally be cached)
	sources := []CapabilitySource{
		{Type: SourceTypeGitHub, Location: "yaleh/meta-cc-capabilities", Priority: 0},
	}

	// Clear cache
	cacheMutex.Lock()
	globalCapabilityCache = nil
	cacheMutex.Unlock()

	// First call with cache disabled - should fail (GitHub not implemented)
	_, err := getCapabilityIndex(sources, true)
	if err == nil {
		t.Error("expected error for unimplemented GitHub source")
	}

	// Cache should still be nil
	cacheMutex.RLock()
	if globalCapabilityCache != nil {
		t.Error("cache should not be populated when GitHub source fails")
	}
	cacheMutex.RUnlock()
}

// TestIsCacheValid tests cache validation
func TestIsCacheValid(t *testing.T) {
	sources := []CapabilitySource{
		{Type: SourceTypeLocal, Location: "/path", Priority: 0},
	}

	// Test with no cache
	cacheMutex.Lock()
	globalCapabilityCache = nil
	cacheMutex.Unlock()

	if isCacheValid(sources) {
		t.Error("cache should be invalid when nil")
	}

	// Test with fresh cache
	cacheMutex.Lock()
	globalCapabilityCache = &CapabilityCache{
		Index:     make(CapabilityIndex),
		Timestamp: time.Now(),
		TTL:       1 * time.Hour,
	}
	cacheMutex.Unlock()

	if !isCacheValid(sources) {
		t.Error("cache should be valid when fresh")
	}

	// Test with expired cache
	cacheMutex.Lock()
	globalCapabilityCache = &CapabilityCache{
		Index:     make(CapabilityIndex),
		Timestamp: time.Now().Add(-2 * time.Hour),
		TTL:       1 * time.Hour,
	}
	cacheMutex.Unlock()

	if isCacheValid(sources) {
		t.Error("cache should be invalid when expired")
	}
}

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
