package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestGenerateFileReference verifies metadata generation accuracy
func TestGenerateFileReference(t *testing.T) {
	tests := []struct {
		name      string
		data      []interface{}
		wantError bool
	}{
		{
			name: "simple records",
			data: []interface{}{
				map[string]interface{}{"id": 1, "name": "test1"},
				map[string]interface{}{"id": 2, "name": "test2"},
			},
			wantError: false,
		},
		{
			name: "diverse schema",
			data: []interface{}{
				map[string]interface{}{"id": 1, "name": "test1"},
				map[string]interface{}{"id": 2, "email": "test@example.com"},
				map[string]interface{}{"name": "test3", "age": 30},
			},
			wantError: false,
		},
		{
			name:      "empty data",
			data:      []interface{}{},
			wantError: false,
		},
		{
			name: "large dataset",
			data: func() []interface{} {
				data := make([]interface{}, 1000)
				for i := 0; i < 1000; i++ {
					data[i] = map[string]interface{}{
						"id":   i,
						"name": "record" + string(rune(i)),
					}
				}
				return data
			}(),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.jsonl")

			// Write test data to file
			if err := writeTestJSONL(filePath, tt.data); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Generate file reference
			fileRef, err := generateFileReference(filePath, tt.data)

			if tt.wantError {
				if err == nil {
					t.Errorf("generateFileReference() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("generateFileReference() unexpected error: %v", err)
				return
			}

			// Validate structure
			if fileRef.Path != filePath {
				t.Errorf("FileReference.Path = %q, want %q", fileRef.Path, filePath)
			}

			if fileRef.LineCount != len(tt.data) {
				t.Errorf("FileReference.LineCount = %d, want %d", fileRef.LineCount, len(tt.data))
			}

			// Validate file size
			stat, err := os.Stat(filePath)
			if err != nil {
				t.Fatalf("Failed to stat file: %v", err)
			}

			if fileRef.SizeBytes != stat.Size() {
				t.Errorf("FileReference.SizeBytes = %d, want %d", fileRef.SizeBytes, stat.Size())
			}

			// Validate fields array is not nil
			if fileRef.Fields == nil {
				t.Errorf("FileReference.Fields is nil, want non-nil slice")
			}

			// Validate summary exists
			if fileRef.Summary == nil {
				t.Errorf("FileReference.Summary is nil, want non-nil map")
			}

			// Validate record count in summary
			if recordCount, ok := fileRef.Summary["record_count"].(int); ok {
				if recordCount != len(tt.data) {
					t.Errorf("Summary.record_count = %d, want %d", recordCount, len(tt.data))
				}
			} else if len(tt.data) > 0 {
				t.Errorf("Summary missing record_count field")
			}
		})
	}
}

// TestFileReferenceSize verifies <500 byte constraint
func TestFileReferenceSize(t *testing.T) {
	tests := []struct {
		name      string
		data      []interface{}
		maxFields int // Maximum number of unique fields to generate
	}{
		{
			name: "simple schema",
			data: []interface{}{
				map[string]interface{}{"id": 1, "name": "test"},
			},
			maxFields: 2,
		},
		{
			name: "complex schema with many fields",
			data: []interface{}{
				map[string]interface{}{
					"id":         1,
					"name":       "test",
					"email":      "test@example.com",
					"age":        30,
					"active":     true,
					"created":    "2025-01-01",
					"updated":    "2025-01-02",
					"status":     "active",
					"role":       "admin",
					"department": "engineering",
				},
			},
			maxFields: 10,
		},
		{
			name: "large dataset",
			data: func() []interface{} {
				data := make([]interface{}, 10000)
				for i := 0; i < 10000; i++ {
					data[i] = map[string]interface{}{
						"id":   i,
						"name": "record",
					}
				}
				return data
			}(),
			maxFields: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.jsonl")

			if err := writeTestJSONL(filePath, tt.data); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			fileRef, err := generateFileReference(filePath, tt.data)
			if err != nil {
				t.Fatalf("generateFileReference() error: %v", err)
			}

			// Serialize to JSON
			jsonBytes, err := json.Marshal(fileRef)
			if err != nil {
				t.Fatalf("Failed to marshal FileReference: %v", err)
			}

			size := len(jsonBytes)
			t.Logf("FileReference JSON size: %d bytes", size)
			t.Logf("FileReference JSON: %s", string(jsonBytes))

			// Verify size constraint
			if size > 500 {
				t.Errorf("FileReference JSON size = %d bytes, want ≤500 bytes", size)
			}
		})
	}
}

// TestExtractFields verifies field detection from JSONL records
func TestExtractFields(t *testing.T) {
	tests := []struct {
		name       string
		records    []interface{}
		wantFields []string
	}{
		{
			name:       "empty records",
			records:    []interface{}{},
			wantFields: []string{},
		},
		{
			name: "single record",
			records: []interface{}{
				map[string]interface{}{"id": 1, "name": "test"},
			},
			wantFields: []string{"id", "name"},
		},
		{
			name: "uniform schema",
			records: []interface{}{
				map[string]interface{}{"id": 1, "name": "test1"},
				map[string]interface{}{"id": 2, "name": "test2"},
			},
			wantFields: []string{"id", "name"},
		},
		{
			name: "diverse schema",
			records: []interface{}{
				map[string]interface{}{"id": 1, "name": "test1"},
				map[string]interface{}{"id": 2, "email": "test@example.com"},
				map[string]interface{}{"name": "test3", "age": 30},
			},
			wantFields: []string{"age", "email", "id", "name"}, // Alphabetically sorted
		},
		{
			name: "nested objects",
			records: []interface{}{
				map[string]interface{}{
					"id":   1,
					"user": map[string]interface{}{"name": "john"},
				},
			},
			wantFields: []string{"id", "user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFields(tt.records)

			// Check length
			if len(got) != len(tt.wantFields) {
				t.Errorf("extractFields() length = %d, want %d", len(got), len(tt.wantFields))
				t.Errorf("  got: %v", got)
				t.Errorf("  want: %v", tt.wantFields)
				return
			}

			// Check each field (should be sorted)
			for i, field := range tt.wantFields {
				if got[i] != field {
					t.Errorf("extractFields()[%d] = %q, want %q", i, got[i], field)
				}
			}
		})
	}
}

// TestSummaryStatistics verifies summary accuracy
func TestSummaryStatistics(t *testing.T) {
	tests := []struct {
		name    string
		records []interface{}
		checks  func(t *testing.T, summary map[string]interface{})
	}{
		{
			name:    "empty records",
			records: []interface{}{},
			checks: func(t *testing.T, summary map[string]interface{}) {
				if recordCount, ok := summary["record_count"].(int); !ok || recordCount != 0 {
					t.Errorf("Summary.record_count = %v, want 0", summary["record_count"])
				}
				// Empty records should not have preview
				if _, ok := summary["preview"]; ok {
					t.Errorf("Summary should not have preview for empty records")
				}
			},
		},
		{
			name: "single record",
			records: []interface{}{
				map[string]interface{}{"id": 1, "name": "test"},
			},
			checks: func(t *testing.T, summary map[string]interface{}) {
				if recordCount, ok := summary["record_count"].(int); !ok || recordCount != 1 {
					t.Errorf("Summary.record_count = %v, want 1", summary["record_count"])
				}
				if _, ok := summary["preview"]; !ok {
					t.Errorf("Summary missing preview")
				}
			},
		},
		{
			name: "multiple records",
			records: []interface{}{
				map[string]interface{}{"id": 1},
				map[string]interface{}{"id": 2},
				map[string]interface{}{"id": 3},
			},
			checks: func(t *testing.T, summary map[string]interface{}) {
				if recordCount, ok := summary["record_count"].(int); !ok || recordCount != 3 {
					t.Errorf("Summary.record_count = %v, want 3", summary["record_count"])
				}

				// Check preview exists and contains data from first record
				if preview, ok := summary["preview"].(string); ok {
					if len(preview) == 0 {
						t.Errorf("Summary.preview is empty")
					}
					// Preview should contain first record's id
					if preview != `{"id":1}` {
						t.Logf("Preview: %s", preview)
					}
				} else {
					t.Errorf("Summary.preview is not a string")
				}
			},
		},
		{
			name: "large record with truncation",
			records: []interface{}{
				map[string]interface{}{
					"id":         1,
					"long_field": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
				},
			},
			checks: func(t *testing.T, summary map[string]interface{}) {
				if recordCount, ok := summary["record_count"].(int); !ok || recordCount != 1 {
					t.Errorf("Summary.record_count = %v, want 1", summary["record_count"])
				}

				// Check preview is truncated to ≤100 chars
				if preview, ok := summary["preview"].(string); ok {
					if len(preview) > 100 {
						t.Errorf("Summary.preview length = %d, want ≤100", len(preview))
					}
					// Should end with "..." if truncated
					if len(preview) == 100 && preview[97:] != "..." {
						t.Errorf("Truncated preview should end with '...'")
					}
				} else {
					t.Errorf("Summary.preview is not a string")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := generateSummary(tt.records)

			if summary == nil {
				t.Fatalf("generateSummary() returned nil")
			}

			tt.checks(t, summary)
		})
	}
}

// Helper functions

// writeTestJSONL writes test data to a JSONL file
func writeTestJSONL(path string, data []interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range data {
		if err := encoder.Encode(record); err != nil {
			return err
		}
	}

	return nil
}
