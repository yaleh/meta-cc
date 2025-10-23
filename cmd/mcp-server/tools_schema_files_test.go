package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/stats"
)

// TestQueryFilesSchemaCompleteness verifies that the query_files schema
// documents all fields present in FileStats structure
func TestQueryFilesSchemaCompleteness(t *testing.T) {
	// Get the tool definition
	tools := getToolDefinitions()
	var queryFilesTool *Tool
	for _, tool := range tools {
		if tool.Name == "query_files" {
			queryFilesTool = &tool
			break
		}
	}

	if queryFilesTool == nil {
		t.Fatal("query_files tool not found")
	}

	// Get jq_filter parameter description
	jqFilterParam, ok := queryFilesTool.InputSchema.Properties["jq_filter"]
	if !ok {
		t.Fatal("jq_filter parameter not found in query_files")
	}

	desc := jqFilterParam.Description

	// Create a sample FileStats to verify all fields
	sample := stats.FileStats{
		FilePath:   "/tmp/test.go",
		ReadCount:  5,
		EditCount:  3,
		WriteCount: 2,
		ErrorCount: 1,
		TotalOps:   10,
		ErrorRate:  0.1,
	}

	jsonBytes, err := json.Marshal(sample)
	if err != nil {
		t.Fatalf("Failed to marshal FileStats: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify all fields from FileStats are documented in schema
	requiredFields := []string{"file_path", "read_count", "edit_count", "write_count", "error_count", "total_ops", "error_rate"}

	missingFields := []string{}
	for _, field := range requiredFields {
		// Check if field exists in actual JSON
		if _, exists := result[field]; !exists {
			t.Errorf("Field %q missing in FileStats JSON output", field)
		}

		// Check if field is documented in schema
		if !strings.Contains(desc, field) {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		t.Errorf("Schema missing documentation for fields: %v", missingFields)
	}
}

// TestQueryFilesExampleUsesCorrectFields verifies that the example jq filter
// uses fields that actually exist in FileStats
func TestQueryFilesExampleUsesCorrectFields(t *testing.T) {
	tools := getToolDefinitions()
	var queryFilesTool *Tool
	for _, tool := range tools {
		if tool.Name == "query_files" {
			queryFilesTool = &tool
			break
		}
	}

	if queryFilesTool == nil {
		t.Fatal("query_files tool not found")
	}

	jqFilterParam, ok := queryFilesTool.InputSchema.Properties["jq_filter"]
	if !ok {
		t.Fatal("jq_filter parameter not found")
	}

	desc := jqFilterParam.Description

	// Check if example mentions any invalid field patterns
	// (This is a basic check - we just verify it doesn't use obviously wrong patterns)
	if strings.Contains(desc, ".reads") || strings.Contains(desc, ".writes") || strings.Contains(desc, ".edits") {
		t.Error("Example should use correct field names: read_count, write_count, edit_count (not reads, writes, edits)")
	}

	// Log the description for manual verification
	t.Logf("query_files jq_filter description:\n%s", desc)
}

// TestFileStatsJSONFieldNames verifies FileStats JSON field naming
func TestFileStatsJSONFieldNames(t *testing.T) {
	sample := stats.FileStats{
		FilePath:   "/tmp/test.go",
		ReadCount:  5,
		EditCount:  3,
		WriteCount: 2,
		ErrorCount: 1,
		TotalOps:   10,
		ErrorRate:  0.1,
	}

	jsonBytes, err := json.Marshal(sample)
	if err != nil {
		t.Fatalf("Failed to marshal FileStats: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify snake_case JSON field names
	expectedFields := []string{
		`"file_path"`,
		`"read_count"`,
		`"edit_count"`,
		`"write_count"`,
		`"error_count"`,
		`"total_ops"`,
		`"error_rate"`,
	}

	for _, expected := range expectedFields {
		if !strings.Contains(jsonStr, expected) {
			t.Errorf("JSON output missing expected field %s, got: %s", expected, jsonStr)
		}
	}
}
