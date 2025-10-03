package stats

import (
	"fmt"

	"github.com/yale/meta-cc/internal/parser"
)

// calculateMetric calculates a specific metric for a group of tools
func calculateMetric(tools []parser.ToolCall, metric string) (interface{}, error) {
	switch metric {
	case "count":
		return len(tools), nil

	case "error_rate":
		if len(tools) == 0 {
			return 0.0, nil
		}
		errorCount := 0
		for _, tool := range tools {
			if tool.Status == "error" {
				errorCount++
			}
		}
		return float64(errorCount) / float64(len(tools)), nil

	default:
		return nil, fmt.Errorf("unsupported metric: %s", metric)
	}
}

// getFieldValue extracts field value from ToolCall
func getFieldValue(tool parser.ToolCall, field string) string {
	switch field {
	case "tool":
		return tool.ToolName
	case "status":
		return tool.Status
	case "uuid":
		return tool.UUID
	default:
		return ""
	}
}
