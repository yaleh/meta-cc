package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/analysis"
)

// TestGetTechDebtToolRegistered verifies get_tech_debt appears in getToolDefinitions()
func TestGetTechDebtToolRegistered(t *testing.T) {
	tools := getToolDefinitions()
	for _, tool := range tools {
		if tool.Name == "get_tech_debt" {
			return
		}
	}
	t.Fatal("get_tech_debt not found in tool definitions")
}

// TestGetTechDebtToolExecution loads test.jsonl and verifies valid JSON output
func TestGetTechDebtToolExecution(t *testing.T) {
	testJSONL, err := filepath.Abs("test.jsonl")
	require.NoError(t, err)
	_, err = os.Stat(testJSONL)
	require.NoError(t, err, "test.jsonl must exist")

	projectPath := setupAnalysisTestProjectDir(t, testJSONL)

	args := map[string]interface{}{
		"working_dir": projectPath,
	}

	output, err := analysis.New().GetTechDebt(args)
	require.NoError(t, err, "executeGetTechDebtTool should not return error")
	require.NotEmpty(t, output, "output should not be empty")

	// Verify output is valid JSON with expected fields
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err, "output should be valid JSON")

	_, hasMarkers := result["markers"]
	assert.True(t, hasMarkers, "result should have 'markers' field")

	_, hasHotspotFiles := result["hotspot_files"]
	assert.True(t, hasHotspotFiles, "result should have 'hotspot_files' field")

	_, hasOpenIssues := result["open_issues"]
	assert.True(t, hasOpenIssues, "result should have 'open_issues' field")
}

// TestGetTechDebtWithSourceDir verifies get_tech_debt merges source_dir markers
func TestGetTechDebtWithSourceDir(t *testing.T) {
	// Create a source directory with known tech debt markers
	srcDir := t.TempDir()
	err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n\n// TODO: fix this\n// HACK: workaround\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(srcDir, "util.ts"), []byte("// FIXME: broken\n"), 0644)
	require.NoError(t, err)

	testJSONL, err := filepath.Abs("test.jsonl")
	require.NoError(t, err)
	_, err = os.Stat(testJSONL)
	require.NoError(t, err, "test.jsonl must exist")

	projectPath := setupAnalysisTestProjectDir(t, testJSONL)

	args := map[string]interface{}{
		"working_dir": projectPath,
		"source_dir":  srcDir,
	}

	output, err := analysis.New().GetTechDebt(args)
	require.NoError(t, err)
	require.NotEmpty(t, output)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	// markers should contain at least the source-dir markers
	markersRaw, ok := result["markers"].([]interface{})
	require.True(t, ok, "markers should be an array")

	// Build a label map from the markers
	labelMap := make(map[string]float64)
	for _, m := range markersRaw {
		entry := m.(map[string]interface{})
		labelMap[entry["label"].(string)] = entry["count"].(float64)
	}

	// TODO and HACK and FIXME should be found from source scan
	assert.GreaterOrEqual(t, labelMap["TODO"], float64(1), "TODO should be found from source scan")
	assert.GreaterOrEqual(t, labelMap["HACK"], float64(1), "HACK should be found from source scan")
	assert.GreaterOrEqual(t, labelMap["FIXME"], float64(1), "FIXME should be found from source scan")

	// hotspot_files should include source files
	hotspotsRaw, ok := result["hotspot_files"].([]interface{})
	require.True(t, ok, "hotspot_files should be an array")
	assert.GreaterOrEqual(t, len(hotspotsRaw), 2, "should have at least 2 hotspot files (source scan + transcript)")

	// data_source should be measured
	assert.Equal(t, "measured", result["data_source"])
}
