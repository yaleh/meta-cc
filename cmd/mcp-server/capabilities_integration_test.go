package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIntegrationMultiSourceLocal tests loading capabilities from multiple local directories
func TestIntegrationMultiSourceLocal(t *testing.T) {
	// Create test fixtures
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	// Create capabilities in first directory
	cap1Path := filepath.Join(tmpDir1, "test-cap1.md")
	cap1Content := `---
name: test-cap1
description: Test capability 1.
keywords: test, cap1, first
category: diagnostics
---

# Test Capability 1

This is a test capability from the first source.
`
	if err := os.WriteFile(cap1Path, []byte(cap1Content), 0644); err != nil {
		t.Fatalf("Failed to create test capability 1: %v", err)
	}

	// Create capabilities in second directory
	cap2Path := filepath.Join(tmpDir2, "test-cap2.md")
	cap2Content := `---
name: test-cap2
description: Test capability 2.
keywords: test, cap2, second
category: assessment
---

# Test Capability 2

This is a test capability from the second source.
`
	if err := os.WriteFile(cap2Path, []byte(cap2Content), 0644); err != nil {
		t.Fatalf("Failed to create test capability 2: %v", err)
	}

	// Call list_capabilities via executeListCapabilitiesTool
	args := map[string]interface{}{
		"_sources":       tmpDir1 + string(os.PathListSeparator) + tmpDir2,
		"_disable_cache": true,
	}

	result, err := executeListCapabilitiesTool(args)
	if err != nil {
		t.Fatalf("Failed to list capabilities: %v", err)
	}

	// Parse JSON result
	var response struct {
		Capabilities []CapabilityMetadata `json:"capabilities"`
	}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Verify both capabilities loaded
	if len(response.Capabilities) != 2 {
		t.Errorf("Expected 2 capabilities, got %d", len(response.Capabilities))
	}

	// Verify cap1 and cap2 exist
	foundCap1 := false
	foundCap2 := false
	for _, cap := range response.Capabilities {
		if cap.Name == "test-cap1" {
			foundCap1 = true
		}
		if cap.Name == "test-cap2" {
			foundCap2 = true
		}
	}

	if !foundCap1 {
		t.Errorf("test-cap1 not found in capabilities")
	}
	if !foundCap2 {
		t.Errorf("test-cap2 not found in capabilities")
	}
}

// TestIntegrationPriorityOverride tests that same-name capabilities follow priority order
func TestIntegrationPriorityOverride(t *testing.T) {
	// Create test fixtures
	tmpDirHigh := t.TempDir()
	tmpDirLow := t.TempDir()

	// Create capability in high-priority directory
	capHighPath := filepath.Join(tmpDirHigh, "duplicate.md")
	capHighContent := `---
name: duplicate
description: High priority version.
keywords: test, duplicate
category: diagnostics
---

# High Priority

Content from high priority source.
`
	if err := os.WriteFile(capHighPath, []byte(capHighContent), 0644); err != nil {
		t.Fatalf("Failed to create high priority capability: %v", err)
	}

	// Create capability in low-priority directory (same name)
	capLowPath := filepath.Join(tmpDirLow, "duplicate.md")
	capLowContent := `---
name: duplicate
description: Low priority version.
keywords: test, duplicate
category: diagnostics
---

# Low Priority

Content from low priority source.
`
	if err := os.WriteFile(capLowPath, []byte(capLowContent), 0644); err != nil {
		t.Fatalf("Failed to create low priority capability: %v", err)
	}

	// Call list_capabilities (high priority first)
	args := map[string]interface{}{
		"_sources":       tmpDirHigh + string(os.PathListSeparator) + tmpDirLow,
		"_disable_cache": true,
	}

	result, err := executeListCapabilitiesTool(args)
	if err != nil {
		t.Fatalf("Failed to list capabilities: %v", err)
	}

	// Parse JSON result
	var response struct {
		Capabilities []CapabilityMetadata `json:"capabilities"`
	}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Verify only one capability loaded (after priority merge)
	if len(response.Capabilities) != 1 {
		t.Errorf("Expected 1 capability (after priority merge), got %d", len(response.Capabilities))
	}

	// Verify high-priority version is used
	if response.Capabilities[0].Description != "High priority version." {
		t.Errorf("Expected high priority description, got: %s", response.Capabilities[0].Description)
	}

	// Get full capability content
	getArgs := map[string]interface{}{
		"name":           "duplicate",
		"_sources":       tmpDirHigh + string(os.PathListSeparator) + tmpDirLow,
		"_disable_cache": true,
	}

	contentResult, err := executeGetCapabilityTool(getArgs)
	if err != nil {
		t.Fatalf("Failed to get capability: %v", err)
	}

	// Parse JSON response
	var getResponse struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(contentResult), &getResponse); err != nil {
		t.Fatalf("Failed to parse get_capability result: %v", err)
	}

	if !strings.Contains(getResponse.Content, "High Priority") {
		t.Errorf("Expected high priority content, got: %s", getResponse.Content)
	}
}

// TestIntegrationCacheBehavior tests that local sources bypass cache
func TestIntegrationCacheBehavior(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial capability
	capPath := filepath.Join(tmpDir, "cached-test.md")
	capContent := `---
name: cached-test
description: Initial version.
keywords: test, cache
category: diagnostics
---

# Initial Version
`
	if err := os.WriteFile(capPath, []byte(capContent), 0644); err != nil {
		t.Fatalf("Failed to create capability: %v", err)
	}

	// First load
	args1 := map[string]interface{}{
		"_sources":       tmpDir,
		"_disable_cache": false, // Enable cache to test local bypass
	}

	result1, err := executeListCapabilitiesTool(args1)
	if err != nil {
		t.Fatalf("Failed to list capabilities (first): %v", err)
	}

	var response1 struct {
		Capabilities []CapabilityMetadata `json:"capabilities"`
	}
	if err := json.Unmarshal([]byte(result1), &response1); err != nil {
		t.Fatalf("Failed to parse result (first): %v", err)
	}

	if len(response1.Capabilities) != 1 {
		t.Fatalf("Expected 1 capability, got %d", len(response1.Capabilities))
	}

	if response1.Capabilities[0].Description != "Initial version." {
		t.Errorf("Expected initial description, got: %s", response1.Capabilities[0].Description)
	}

	// Update capability (simulate local development)
	capContentUpdated := `---
name: cached-test
description: Updated version.
keywords: test, cache
category: diagnostics
---

# Updated Version
`
	if err := os.WriteFile(capPath, []byte(capContentUpdated), 0644); err != nil {
		t.Fatalf("Failed to update capability: %v", err)
	}

	// Second load (should reflect changes immediately for local sources, even with cache enabled)
	args2 := map[string]interface{}{
		"_sources":       tmpDir,
		"_disable_cache": false, // Cache enabled, but local sources bypass
	}

	result2, err := executeListCapabilitiesTool(args2)
	if err != nil {
		t.Fatalf("Failed to list capabilities (second): %v", err)
	}

	var response2 struct {
		Capabilities []CapabilityMetadata `json:"capabilities"`
	}
	if err := json.Unmarshal([]byte(result2), &response2); err != nil {
		t.Fatalf("Failed to parse result (second): %v", err)
	}

	// Local sources should bypass cache, so changes reflect immediately
	if response2.Capabilities[0].Description != "Updated version." {
		t.Errorf("Expected updated description (local sources bypass cache), got: %s", response2.Capabilities[0].Description)
	}
}

// TestIntegrationEndToEnd tests complete workflow: list + get + content verification
func TestIntegrationEndToEnd(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple capabilities
	caps := []struct {
		name        string
		description string
		keywords    string
		category    string
		content     string
	}{
		{"e2e-cap1", "End-to-end test 1.", "test, e2e, first", "diagnostics", "# E2E Test 1\n\nContent 1"},
		{"e2e-cap2", "End-to-end test 2.", "test, e2e, second", "assessment", "# E2E Test 2\n\nContent 2"},
		{"e2e-cap3", "End-to-end test 3.", "test, e2e, third", "visualization", "# E2E Test 3\n\nContent 3"},
	}

	for _, cap := range caps {
		capPath := filepath.Join(tmpDir, cap.name+".md")
		capContent := "---\n"
		capContent += "name: " + cap.name + "\n"
		capContent += "description: " + cap.description + "\n"
		capContent += "keywords: " + cap.keywords + "\n"
		capContent += "category: " + cap.category + "\n"
		capContent += "---\n\n"
		capContent += cap.content + "\n"

		if err := os.WriteFile(capPath, []byte(capContent), 0644); err != nil {
			t.Fatalf("Failed to create capability %s: %v", cap.name, err)
		}
	}

	// Step 1: List capabilities
	listArgs := map[string]interface{}{
		"_sources":       tmpDir,
		"_disable_cache": true,
	}

	listResult, err := executeListCapabilitiesTool(listArgs)
	if err != nil {
		t.Fatalf("Failed to list capabilities: %v", err)
	}

	var listResponse struct {
		Capabilities []CapabilityMetadata `json:"capabilities"`
	}
	if err := json.Unmarshal([]byte(listResult), &listResponse); err != nil {
		t.Fatalf("Failed to parse list result: %v", err)
	}

	if len(listResponse.Capabilities) != 3 {
		t.Errorf("Expected 3 capabilities, got %d", len(listResponse.Capabilities))
	}

	// Step 2: Get each capability and verify
	for _, capData := range caps {
		// Find capability in list
		var foundCap *CapabilityMetadata
		for i := range listResponse.Capabilities {
			if listResponse.Capabilities[i].Name == capData.name {
				foundCap = &listResponse.Capabilities[i]
				break
			}
		}

		if foundCap == nil {
			t.Errorf("Capability %s not found in list", capData.name)
			continue
		}

		// Verify metadata
		if foundCap.Description != capData.description {
			t.Errorf("Expected description %s, got %s", capData.description, foundCap.Description)
		}

		if foundCap.Category != capData.category {
			t.Errorf("Expected category %s, got %s", capData.category, foundCap.Category)
		}

		// Get full capability content
		getArgs := map[string]interface{}{
			"name":           capData.name,
			"_sources":       tmpDir,
			"_disable_cache": true,
		}

		contentResult, err := executeGetCapabilityTool(getArgs)
		if err != nil {
			t.Errorf("Failed to get capability %s: %v", capData.name, err)
			continue
		}

		// Parse JSON response
		var getResponse struct {
			Mode    string `json:"mode"`
			Name    string `json:"name"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(contentResult), &getResponse); err != nil {
			t.Errorf("Failed to parse get_capability result: %v", err)
			continue
		}

		// Verify content
		if !strings.Contains(getResponse.Content, capData.content) {
			t.Errorf("Expected content to contain '%s', got: %s", capData.content, getResponse.Content)
		}
	}
}
