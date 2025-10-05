package output

import (
	"encoding/json"
	"fmt"
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
	// Marshal as compact JSON (no indentation)
	output, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSONL: %w", err)
	}
	return string(output), nil
}
