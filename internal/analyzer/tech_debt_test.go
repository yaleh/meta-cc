package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/types"
)

func makeToolCallWithOutput(toolName, filePath, output, status string) types.ToolCall {
	input := map[string]interface{}{}
	if filePath != "" {
		input["file_path"] = filePath
	}
	return types.ToolCall{
		UUID:      "test-uuid",
		ToolName:  toolName,
		Input:     input,
		Output:    output,
		Status:    status,
		Timestamp: "2025-10-02T10:00:00.000Z",
	}
}

func TestGetTechDebt_DetectsMarkers(t *testing.T) {
	toolCalls := []types.ToolCall{
		makeToolCallWithOutput("Read", "main.go", "func foo() {\n// TODO: fix this\n}", "success"),
	}
	result, err := GetTechDebt(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Find TODO marker count
	todoCount := 0
	for _, m := range result.Markers {
		if m.Label == "TODO" {
			todoCount = m.Count
		}
	}
	if todoCount <= 0 {
		t.Errorf("expected TODO count > 0, got %d", todoCount)
	}
	// main.go should appear in hotspot files
	found := false
	for _, f := range result.HotspotFiles {
		if f.File == "main.go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected main.go in HotspotFiles, got %v", result.HotspotFiles)
	}
}

func TestGetTechDebt_CountsPerFile(t *testing.T) {
	toolCalls := []types.ToolCall{
		makeToolCallWithOutput("Read", "a.go", "// TODO: first\n// TODO: second\n", "success"),
		makeToolCallWithOutput("Edit", "b.go", "// TODO: only one\n", "success"),
	}
	result, err := GetTechDebt(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.HotspotFiles) < 2 {
		t.Fatalf("expected at least 2 hotspot files, got %d", len(result.HotspotFiles))
	}
	if result.HotspotFiles[0].File != "a.go" {
		t.Errorf("expected a.go first (2 markers), got %s", result.HotspotFiles[0].File)
	}
	if result.HotspotFiles[0].MarkerCount < result.HotspotFiles[1].MarkerCount {
		t.Errorf("expected descending order: %v", result.HotspotFiles)
	}
}

func TestGetTechDebt_DetectsOpenIssues(t *testing.T) {
	toolCalls := []types.ToolCall{
		makeToolCallWithOutput("Bash", "", "error: build failed", "error"),
		makeToolCallWithOutput("Read", "", "more errors", "error"),
		makeToolCallWithOutput("Edit", "x.go", "stuff", "error"),
	}
	result, err := GetTechDebt(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OpenIssues <= 0 {
		t.Errorf("expected OpenIssues > 0, got %d", result.OpenIssues)
	}
}

func TestGetTechDebt_DataSource(t *testing.T) {
	result, err := GetTechDebt(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DataSource != DataSourceMeasured {
		t.Errorf("Expected DataSource=%q, got %q", DataSourceMeasured, result.DataSource)
	}
}

func TestScanSourceDir_DetectsMarkers(t *testing.T) {
	// Create temp directory with source files containing known markers
	tmpDir := t.TempDir()

	// Create a known code file
	writeFile(t, filepath.Join(tmpDir, "main.go"), "package main\n\n// TODO: fix this later\nfunc foo() {\n\t// FIXME: broken\n}\n")
	writeFile(t, filepath.Join(tmpDir, "lib.py"), "# HACK: workaround\ndef bar():\n\t# XXX: clean up\n\tpass\n")

	// Create hidden dir with code files (should be skipped)
	hiddenDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(hiddenDir, 0755)
	require.NoError(t, err)
	writeFile(t, filepath.Join(hiddenDir, "config.go"), "// TODO: should not be seen\n")

	result, err := ScanSourceDir(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check marker counts: TODO=1, FIXME=1, HACK=1, XXX=1
	labelMap := make(map[string]int)
	for _, m := range result.Markers {
		labelMap[m.Label] = m.Count
	}
	assert.Equal(t, 1, labelMap["TODO"], "TODO count mismatch")
	assert.Equal(t, 1, labelMap["FIXME"], "FIXME count mismatch")
	assert.Equal(t, 1, labelMap["HACK"], "HACK count mismatch")
	assert.Equal(t, 1, labelMap["XXX"], "XXX count mismatch")

	// Check hotspot files: expect 2 files
	assert.Equal(t, 2, len(result.HotspotFiles), "expected 2 hotspot files")

	// Hidden dir files should NOT appear
	for _, f := range result.HotspotFiles {
		assert.NotContains(t, f.File, ".git", "hidden dir files should be skipped")
	}

	// DataSource should be measured
	assert.Equal(t, DataSourceMeasured, result.DataSource)

	// OpenIssues should be 0 (source scan has no open issues)
	assert.Equal(t, 0, result.OpenIssues)
}

func TestScanSourceDir_SkipsNodeModules(t *testing.T) {
	tmpDir := t.TempDir()
	nmDir := filepath.Join(tmpDir, "node_modules")
	err := os.MkdirAll(nmDir, 0755)
	require.NoError(t, err)
	writeFile(t, filepath.Join(nmDir, "index.js"), "// TODO: in node_modules\n")

	result, err := ScanSourceDir(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 0, len(result.HotspotFiles), "node_modules should be skipped entirely")
}

func TestMergeTechDebtResults(t *testing.T) {
	a := &TechDebtResult{
		Markers:      []MarkerCount{{Label: "TODO", Count: 3}, {Label: "FIXME", Count: 1}},
		HotspotFiles: []FileDebt{{File: "a.go", MarkerCount: 2}, {File: "b.go", MarkerCount: 1}},
		OpenIssues:   2,
		DataSource:   DataSourceMeasured,
	}
	b := &TechDebtResult{
		Markers:      []MarkerCount{{Label: "TODO", Count: 1}, {Label: "HACK", Count: 4}},
		HotspotFiles: []FileDebt{{File: "a.go", MarkerCount: 3}, {File: "c.go", MarkerCount: 4}},
		OpenIssues:   1,
		DataSource:   DataSourceMeasured,
	}

	merged := MergeTechDebtResults(a, b, DataSourceMeasured)

	// Check markers: TODO=4, FIXME=1, HACK=4
	labelMap := make(map[string]int)
	for _, m := range merged.Markers {
		labelMap[m.Label] = m.Count
	}
	assert.Equal(t, 4, labelMap["TODO"])
	assert.Equal(t, 1, labelMap["FIXME"])
	assert.Equal(t, 4, labelMap["HACK"])

	// Check hotspot files: a.go=5, c.go=4, b.go=1 (sorted desc)
	assert.Equal(t, 3, len(merged.HotspotFiles))
	assert.Equal(t, "a.go", merged.HotspotFiles[0].File)
	assert.Equal(t, 5, merged.HotspotFiles[0].MarkerCount)

	// OpenIssues uses max
	assert.Equal(t, 2, merged.OpenIssues)

	// DataSource passed through
	assert.Equal(t, DataSourceMeasured, merged.DataSource)
}

func TestGetTechDebt_EmptySession(t *testing.T) {
	result, err := GetTechDebt(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.OpenIssues != 0 {
		t.Errorf("expected 0 OpenIssues, got %d", result.OpenIssues)
	}
	if len(result.Markers) != 0 {
		t.Errorf("expected empty markers, got %v", result.Markers)
	}
	if len(result.HotspotFiles) != 0 {
		t.Errorf("expected empty hotspot files, got %v", result.HotspotFiles)
	}
}

// writeFile is a test helper that writes content to a file path, failing the test on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}
