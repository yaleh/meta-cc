package response

import (
	"encoding/json"
	"os"
	"sort"
)

// FileReference provides metadata about a temporary JSONL file.
//
// This struct is deliberately LEAN. A pre-existing contract
// (cmd/mcp-server/file_reference_test.go TestFileReferenceSize) marshals it
// directly and requires ≤500 bytes, so the heavy self-describing metadata
// (shape / sample / recipes) is NOT stored here. DIR-080 / ADR-009 attaches
// that richer metadata to the caller-facing file_ref envelope inside
// BuildFileRefResponse (see adapter.go) instead of this struct.
type FileReference struct {
	Path      string                 `json:"path"`
	SizeBytes int64                  `json:"size_bytes"`
	LineCount int                    `json:"line_count"`
	Fields    []string               `json:"fields"`
	Summary   map[string]interface{} `json:"summary"`
}

// GenerateFileReference creates a FileReference with metadata for a JSONL file.
func GenerateFileReference(filePath string, data []interface{}) (*FileReference, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	fields := ExtractFields(data)
	summary := GenerateSummary(data)

	return &FileReference{
		Path:      filePath,
		SizeBytes: stat.Size(),
		LineCount: len(data),
		Fields:    fields,
		Summary:   summary,
	}, nil
}

// ExtractFields extracts unique field names from JSONL records.
func ExtractFields(records []interface{}) []string {
	if len(records) == 0 {
		return []string{}
	}

	fieldSet := make(map[string]bool)
	for _, record := range records {
		if recordMap, ok := record.(map[string]interface{}); ok {
			for field := range recordMap {
				fieldSet[field] = true
			}
		}
	}

	fields := make([]string, 0, len(fieldSet))
	for field := range fieldSet {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	return fields
}

// GenerateSummary creates summary statistics for JSONL records.
func GenerateSummary(records []interface{}) map[string]interface{} {
	summary := make(map[string]interface{})
	summary["record_count"] = len(records)

	if len(records) > 0 {
		firstJSON, err := json.Marshal(records[0])
		if err == nil {
			preview := string(firstJSON)
			if len(preview) > 100 {
				preview = preview[:97] + "..."
			}
			summary["preview"] = preview
		}
	}

	return summary
}
