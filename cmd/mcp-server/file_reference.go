package main

import (
	"encoding/json"
	"os"
	"sort"
)

// FileReference provides metadata about a temporary JSONL file.
// It helps Claude understand the structure and content of large query results
// without loading the entire file into memory.
type FileReference struct {
	// Path is the absolute file path to the JSONL file
	Path string `json:"path"`

	// SizeBytes is the file size in bytes
	SizeBytes int64 `json:"size_bytes"`

	// LineCount is the number of JSONL records
	LineCount int `json:"line_count"`

	// Fields is an alphabetically sorted list of unique field names
	Fields []string `json:"fields"`

	// Summary contains statistics and preview data
	Summary map[string]interface{} `json:"summary"`
}

// generateFileReference creates a FileReference with metadata for a JSONL file.
//
// Parameters:
//   - filePath: Absolute path to the JSONL file
//   - data: Array of records written to the file
//
// Returns:
//   - FileReference with path, size, line count, fields, and summary
//   - Error if file stat fails
//
// The function extracts metadata without re-reading the file, using the provided
// data array to derive fields and summary statistics.
func generateFileReference(filePath string, data []interface{}) (*FileReference, error) {
	// Get file size
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// Extract fields and summary
	fields := extractFields(data)
	summary := generateSummary(data)

	return &FileReference{
		Path:      filePath,
		SizeBytes: stat.Size(),
		LineCount: len(data),
		Fields:    fields,
		Summary:   summary,
	}, nil
}

// extractFields extracts unique field names from JSONL records.
//
// Parameters:
//   - records: Array of records (typically []map[string]interface{})
//
// Returns:
//   - Alphabetically sorted list of unique field names
//
// The function handles diverse schemas where different records may have
// different fields. It collects all unique field names across all records.
func extractFields(records []interface{}) []string {
	if len(records) == 0 {
		return []string{}
	}

	// Collect unique fields
	fieldSet := make(map[string]bool)

	for _, record := range records {
		// Try to cast to map
		if recordMap, ok := record.(map[string]interface{}); ok {
			for field := range recordMap {
				fieldSet[field] = true
			}
		}
	}

	// Convert to sorted slice
	fields := make([]string, 0, len(fieldSet))
	for field := range fieldSet {
		fields = append(fields, field)
	}

	sort.Strings(fields)

	return fields
}

// generateSummary creates summary statistics for JSONL records.
//
// Parameters:
//   - records: Array of records
//
// Returns:
//   - Map containing:
//   - record_count: Total number of records
//   - preview: Compact JSON string of first record (max 100 chars)
//
// The summary provides Claude with context about the data without loading
// the entire dataset. It's designed to serialize to <500 bytes total.
// Preview is truncated to keep the FileReference size under constraint.
func generateSummary(records []interface{}) map[string]interface{} {
	summary := make(map[string]interface{})

	// Always include record count
	summary["record_count"] = len(records)

	// Add compact preview of first record (if records exist)
	if len(records) > 0 {
		// Serialize first record to compact JSON
		firstJSON, err := json.Marshal(records[0])
		if err == nil {
			preview := string(firstJSON)
			// Truncate if too long to stay under 500 byte constraint
			if len(preview) > 100 {
				preview = preview[:97] + "..."
			}
			summary["preview"] = preview
		}
	}

	return summary
}
