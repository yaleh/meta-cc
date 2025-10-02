package output

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/yale/meta-cc/internal/parser"
)

// FormatCSV formats data as CSV
// Currently only supports ToolCall arrays
func FormatCSV(data interface{}) (string, error) {
	switch v := data.(type) {
	case []parser.ToolCall:
		return formatToolCallsCSV(v)
	default:
		return "", fmt.Errorf("unsupported data type for CSV formatting")
	}
}

func formatToolCallsCSV(toolCalls []parser.ToolCall) (string, error) {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// Write header
	header := []string{"UUID", "Tool", "Input", "Output", "Status", "Error"}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, tc := range toolCalls {
		inputStr := formatInputForCSV(tc.Input)

		row := []string{
			tc.UUID,
			tc.ToolName,
			inputStr,
			tc.Output,
			tc.Status,
			tc.Error,
		}

		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %w", err)
	}

	return sb.String(), nil
}

func formatInputForCSV(input map[string]interface{}) string {
	if len(input) == 0 {
		return ""
	}

	var parts []string
	for k, v := range input {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	return strings.Join(parts, "; ")
}
