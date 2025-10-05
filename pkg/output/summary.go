package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/yale/meta-cc/internal/parser"
)

// SummaryOutput contains summary and detailed records
type SummaryOutput struct {
	Summary string // Summary statistics
	Details string // Detailed records (in specified format)
}

// GenerateSummary generates a compact summary of tool calls
func GenerateSummary(tools []parser.ToolCall) string {
	var sb strings.Builder

	sb.WriteString("=== Session Summary ===\n")
	sb.WriteString(fmt.Sprintf("Total Tools: %d\n", len(tools)))

	// Count errors
	errorCount := 0
	for _, tool := range tools {
		if tool.Status == "error" || tool.Error != "" {
			errorCount++
		}
	}

	errorRate := 0.0
	if len(tools) > 0 {
		errorRate = float64(errorCount) / float64(len(tools)) * 100
	}

	sb.WriteString(fmt.Sprintf("Errors: %d (%.1f%%)\n", errorCount, errorRate))

	// Calculate tool frequency
	toolFreq := make(map[string]int)
	for _, tool := range tools {
		toolFreq[tool.ToolName]++
	}

	// Sort by frequency
	type toolCount struct {
		name  string
		count int
	}
	var counts []toolCount
	for name, count := range toolFreq {
		counts = append(counts, toolCount{name, count})
	}
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].count == counts[j].count {
			return counts[i].name < counts[j].name
		}
		return counts[i].count > counts[j].count
	})

	// Show top 5 tools
	sb.WriteString("\nTop Tools:\n")
	for i, tc := range counts {
		if i >= 5 {
			break
		}
		sb.WriteString(fmt.Sprintf("  %d. %s (%d)\n", i+1, tc.name, tc.count))
	}

	return sb.String()
}

// FormatSummaryFirst outputs summary first, then top N detailed records
func FormatSummaryFirst(tools []parser.ToolCall, topN int, detailFormat string) (SummaryOutput, error) {
	// Generate summary
	summary := GenerateSummary(tools)

	// Select top N records (or all if topN <= 0 or topN >= len(tools))
	detailTools := tools
	if topN > 0 && topN < len(tools) {
		detailTools = tools[:topN]
	}

	// Format details
	var details string
	var err error

	switch detailFormat {
	case "jsonl":
		var data []byte
		data, err = json.Marshal(detailTools) // Compact JSON (JSONL)
		if err != nil {
			return SummaryOutput{}, err
		}
		details = string(data)

	case "tsv":
		var err error
		details, err = FormatTSV(detailTools)
		if err != nil {
			return SummaryOutput{}, fmt.Errorf("failed to format TSV: %w", err)
		}

	default:
		return SummaryOutput{}, fmt.Errorf("unsupported format: %s (supported: jsonl, tsv)", detailFormat)
	}

	return SummaryOutput{
		Summary: summary,
		Details: details,
	}, nil
}
