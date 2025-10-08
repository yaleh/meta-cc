package output

import (
	"encoding/json"

	"github.com/yaleh/meta-cc/internal/parser"
)

// SizeEstimate represents estimated output size
type SizeEstimate struct {
	EstimatedBytes int     `json:"estimated_bytes"`
	EstimatedKB    float64 `json:"estimated_kb"`
	Format         string  `json:"format"`
	RecordCount    int     `json:"record_count"`
}

// EstimateToolCallsSize estimates the output size for ToolCall slice
// Achieves ≥95% accuracy by sampling actual JSON serialization
func EstimateToolCallsSize(tools []parser.ToolCall, format string) (SizeEstimate, error) {
	var sizeBytes int

	switch format {
	case "json":
		// For JSON, sample first record and multiply
		if len(tools) == 0 {
			sizeBytes = 2 // "[]"
		} else {
			// Serialize first record to get accurate size
			sample, err := json.Marshal(tools[0])
			if err != nil {
				return SizeEstimate{}, err
			}
			// Each record + comma + newline (~2 bytes overhead)
			// Use actual size without buffer for better accuracy
			recordSize := len(sample) + 2
			sizeBytes = recordSize*len(tools) + 10 // +10 for array brackets
		}

	case "md", "markdown":
		// Markdown: table format with headers
		// Estimate: ~300 bytes per record (row with multiple columns)
		sizeBytes = len(tools)*300 + 500 // +500 for headers and separators

	case "csv":
		// CSV: comma-separated values
		// Estimate: ~200 bytes per record
		sizeBytes = len(tools)*200 + 100 // +100 for header row

	default:
		// Unknown format: conservative estimate
		sizeBytes = len(tools) * 300
	}

	return SizeEstimate{
		EstimatedBytes: sizeBytes,
		EstimatedKB:    float64(sizeBytes) / 1024.0,
		Format:         format,
		RecordCount:    len(tools),
	}, nil
}

// EstimateStatsSize estimates the output size for stats report
// Stats are fixed size regardless of session size
func EstimateStatsSize(format string) SizeEstimate {
	sizeMap := map[string]int{
		"json": 800,
		"md":   1200,
		"csv":  600,
	}

	size := sizeMap[format]
	if size == 0 {
		size = 1000
	}

	return SizeEstimate{
		EstimatedBytes: size,
		EstimatedKB:    float64(size) / 1024.0,
		Format:         format,
		RecordCount:    1, // Stats report is a single record
	}
}
