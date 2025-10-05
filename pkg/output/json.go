package output

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatErrorJSON formats error objects as JSON for stderr output
func FormatErrorJSON(err error, code string) (string, error) {
	errObj := map[string]interface{}{
		"error": err.Error(),
		"code":  code,
	}
	output, marshalErr := json.Marshal(errObj)
	if marshalErr != nil {
		return "", fmt.Errorf("failed to marshal error JSON: %w", marshalErr)
	}
	return string(output), nil
}

// FormatJSONL formats data as JSON Lines (one JSON object per line)
func FormatJSONL(data interface{}) (string, error) {
	// Reflect to check if data is a slice/array
	// If it's a slice, marshal each item on a separate line
	// If it's a single object, just marshal it

	// Use type switch to handle slices
	switch v := data.(type) {
	case []interface{}:
		return formatSliceJSONL(v)
	default:
		// Try to handle as a slice via reflection
		// This handles typed slices like []ToolCall, []ErrorEntry, etc.
		return formatGenericJSONL(data)
	}
}

// formatSliceJSONL formats []interface{} as JSONL
func formatSliceJSONL(items []interface{}) (string, error) {
	if len(items) == 0 {
		return "", nil
	}

	var lines []string
	for _, item := range items {
		jsonBytes, err := json.Marshal(item)
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSONL item: %w", err)
		}
		lines = append(lines, string(jsonBytes))
	}

	return strings.Join(lines, "\n") + "\n", nil
}

// formatGenericJSONL uses reflection to handle any slice type
func formatGenericJSONL(data interface{}) (string, error) {
	// First try to marshal as-is (handles non-slice types and empty slices)
	// Check if it's a slice using type assertion patterns

	// Convert to []interface{} if it's a slice
	// Use json.Marshal + unmarshal trick to convert typed slice to []interface{}
	tempJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	// Check if it's a JSON array (starts with '[')
	if len(tempJSON) > 0 && tempJSON[0] == '[' {
		// It's an array, unmarshal to []interface{} and format as JSONL
		var items []interface{}
		if err := json.Unmarshal(tempJSON, &items); err != nil {
			return "", fmt.Errorf("failed to unmarshal array: %w", err)
		}

		if len(items) == 0 {
			return "", nil
		}

		var lines []string
		for _, item := range items {
			jsonBytes, err := json.Marshal(item)
			if err != nil {
				return "", fmt.Errorf("failed to marshal JSONL item: %w", err)
			}
			lines = append(lines, string(jsonBytes))
		}

		return strings.Join(lines, "\n") + "\n", nil
	}

	// It's a single object, just return it
	return string(tempJSON), nil
}
