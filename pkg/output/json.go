package output

import (
	"encoding/json"
	"fmt"
)

// FormatJSON formats any data as pretty-printed JSON
func FormatJSON(data interface{}) (string, error) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(output), nil
}
