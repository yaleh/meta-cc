package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yale/meta-cc/internal/parser"
)

// FormatTSV formats ToolCall slice as TSV (Tab-Separated Values)
// TSV is ~50% smaller than JSON due to:
// - No quotes around values
// - No field names per record
// - No JSON structure overhead
func FormatTSV(tools []parser.ToolCall) string {
	if len(tools) == 0 {
		return ""
	}

	var sb strings.Builder

	// Header row
	sb.WriteString("UUID\tToolName\tStatus\tError\n")

	// Data rows
	for _, tool := range tools {
		sb.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\n",
			tool.UUID,
			tool.ToolName,
			tool.Status,
			escapeTSV(tool.Error),
		))
	}

	return sb.String()
}

// FormatProjectedTSV formats projected data as TSV
func FormatProjectedTSV(projected []ProjectedToolCall) string {
	if len(projected) == 0 {
		return ""
	}

	// Extract field names from first record (sorted for consistency)
	var fields []string
	for field := range projected[0] {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	var sb strings.Builder

	// Header row
	for i, field := range fields {
		if i > 0 {
			sb.WriteString("\t")
		}
		sb.WriteString(field)
	}
	sb.WriteString("\n")

	// Data rows
	for _, record := range projected {
		for i, field := range fields {
			if i > 0 {
				sb.WriteString("\t")
			}
			value := fmt.Sprintf("%v", record[field])
			sb.WriteString(escapeTSV(value))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// escapeTSV escapes tab and newline characters to prevent breaking TSV format
func escapeTSV(s string) string {
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
