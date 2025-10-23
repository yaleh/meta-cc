package output

import (
	"strings"
	"testing"
)

// TestFormatJSONLEmptySlice verifies that empty slices return valid JSON array
// instead of empty string. This prevents jq errors when query returns no results.
func TestFormatJSONLEmptySlice(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "empty []interface{}",
			input: []interface{}{},
			want:  "[]",
		},
		{
			name:  "empty slice via generic handling",
			input: []string{},
			want:  "[]",
		},
		{
			name:  "nil slice",
			input: []interface{}(nil),
			want:  "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatJSONL(tt.input)
			if err != nil {
				t.Errorf("FormatJSONL() error = %v", err)
				return
			}

			// Must return valid JSON array, not empty string
			if got != tt.want {
				t.Errorf("FormatJSONL() = %q, want %q", got, tt.want)
			}

			// Empty string would cause jq errors
			if got == "" {
				t.Error("FormatJSONL() returned empty string - this causes 'unexpected end of JSON input' in jq")
			}
		})
	}
}

// TestFormatJSONLEmptyResultsJqCompatibility verifies that empty results
// can be piped to jq without errors
func TestFormatJSONLEmptyResultsJqCompatibility(t *testing.T) {
	emptySlice := []interface{}{}
	output, err := FormatJSONL(emptySlice)
	if err != nil {
		t.Fatalf("FormatJSONL() error = %v", err)
	}

	// Simulate what jq sees
	if output == "" {
		t.Error("Empty output will cause jq error: 'unexpected end of JSON input'")
	}

	// Should be valid JSON that jq can parse
	if output != "[]" {
		t.Errorf("Expected valid empty array '[]', got %q", output)
	}

	// Test that it can be parsed as JSON
	if !strings.HasPrefix(output, "[") || !strings.HasSuffix(output, "]") {
		t.Errorf("Output should be valid JSON array: %q", output)
	}
}

// TestFormatJSONLNonEmptySlice verifies that non-empty slices still work correctly
func TestFormatJSONLNonEmptySlice(t *testing.T) {
	type TestRecord struct {
		Field string `json:"field"`
	}

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "single item",
			input: []interface{}{map[string]string{"field": "value"}},
		},
		{
			name:  "multiple items",
			input: []TestRecord{{Field: "a"}, {Field: "b"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatJSONL(tt.input)
			if err != nil {
				t.Errorf("FormatJSONL() error = %v", err)
				return
			}

			// Non-empty results should produce JSONL output (not empty, not [])
			if got == "" {
				t.Error("FormatJSONL() should not return empty string for non-empty slice")
			}
			if got == "[]" {
				t.Error("FormatJSONL() should return JSONL format for non-empty slice, not JSON array")
			}

			// Should contain actual data
			if !strings.Contains(got, "field") {
				t.Errorf("FormatJSONL() should contain data: %q", got)
			}
		})
	}
}
