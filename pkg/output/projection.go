package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/yale/meta-cc/internal/parser"
)

// ProjectionConfig defines which fields to include in output
type ProjectionConfig struct {
	Fields         []string // Base fields to include
	IfErrorInclude []string // Additional fields to include for error records
}

// ProjectedToolCall represents a ToolCall with only selected fields
type ProjectedToolCall map[string]interface{}

// ProjectToolCalls applies field projection to ToolCall slice
// Returns a slice of maps containing only the specified fields
func ProjectToolCalls(tools []parser.ToolCall, config ProjectionConfig) ([]ProjectedToolCall, error) {
	if len(config.Fields) == 0 {
		// No projection - return full objects as maps
		return convertToMaps(tools), nil
	}

	projected := make([]ProjectedToolCall, 0, len(tools))

	for _, tool := range tools {
		// Convert ToolCall to map for field access
		toolMap := toolCallToMap(tool)

		// Build projected object with base fields
		projectedTool := make(ProjectedToolCall)

		// Include specified base fields
		for _, field := range config.Fields {
			if value, ok := toolMap[field]; ok {
				projectedTool[field] = value
			}
		}

		// If this is an error record, include additional error fields
		if tool.Status == "error" && len(config.IfErrorInclude) > 0 {
			for _, field := range config.IfErrorInclude {
				if value, ok := toolMap[field]; ok {
					projectedTool[field] = value
				}
			}
		}

		projected = append(projected, projectedTool)
	}

	return projected, nil
}

// ParseProjectionConfig parses field specification strings
// Format: "field1,field2,field3" (comma-separated)
func ParseProjectionConfig(fieldsStr, ifErrorIncludeStr string) ProjectionConfig {
	config := ProjectionConfig{}

	if fieldsStr != "" {
		fields := strings.Split(fieldsStr, ",")
		config.Fields = make([]string, 0, len(fields))
		for _, f := range fields {
			trimmed := strings.TrimSpace(f)
			if trimmed != "" {
				config.Fields = append(config.Fields, trimmed)
			}
		}
	}

	if ifErrorIncludeStr != "" {
		errorFields := strings.Split(ifErrorIncludeStr, ",")
		config.IfErrorInclude = make([]string, 0, len(errorFields))
		for _, f := range errorFields {
			trimmed := strings.TrimSpace(f)
			if trimmed != "" {
				config.IfErrorInclude = append(config.IfErrorInclude, trimmed)
			}
		}
	}

	return config
}

// toolCallToMap converts a ToolCall struct to a map for field access
// Uses JSON marshaling/unmarshaling to handle struct-to-map conversion
func toolCallToMap(tool parser.ToolCall) map[string]interface{} {
	// Use JSON round-trip for accurate conversion
	data, _ := json.Marshal(tool)
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	return m
}

// convertToMaps converts ToolCall slice to ProjectedToolCall slice (no projection)
func convertToMaps(tools []parser.ToolCall) []ProjectedToolCall {
	result := make([]ProjectedToolCall, len(tools))
	for i, tool := range tools {
		result[i] = toolCallToMap(tool)
	}
	return result
}

// FormatProjectedOutput formats projected output in the specified format
func FormatProjectedOutput(projected []ProjectedToolCall, format string) (string, error) {
	switch format {
	case "json":
		data, err := json.MarshalIndent(projected, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "md", "markdown":
		return formatProjectedMarkdown(projected), nil

	case "csv":
		return formatProjectedCSV(projected), nil

	case "tsv":
		return FormatProjectedTSV(projected), nil

	default:
		return "", fmt.Errorf("unsupported format for projection: %s", format)
	}
}

// formatProjectedMarkdown formats projected data as Markdown table
func formatProjectedMarkdown(projected []ProjectedToolCall) string {
	if len(projected) == 0 {
		return "No data"
	}

	// Extract field names from first record (sorted for consistency)
	var fields []string
	for field := range projected[0] {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	var sb strings.Builder

	// Header row
	sb.WriteString("| ")
	for _, field := range fields {
		sb.WriteString(field)
		sb.WriteString(" | ")
	}
	sb.WriteString("\n")

	// Separator row
	sb.WriteString("| ")
	for range fields {
		sb.WriteString("--- | ")
	}
	sb.WriteString("\n")

	// Data rows
	for _, record := range projected {
		sb.WriteString("| ")
		for _, field := range fields {
			value := fmt.Sprintf("%v", record[field])
			// Escape pipe characters to prevent breaking table
			value = strings.ReplaceAll(value, "|", "\\|")
			sb.WriteString(value)
			sb.WriteString(" | ")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatProjectedCSV formats projected data as CSV
func formatProjectedCSV(projected []ProjectedToolCall) string {
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
			sb.WriteString(",")
		}
		sb.WriteString(escapeCSV(field))
	}
	sb.WriteString("\n")

	// Data rows
	for _, record := range projected {
		for i, field := range fields {
			if i > 0 {
				sb.WriteString(",")
			}
			value := fmt.Sprintf("%v", record[field])
			sb.WriteString(escapeCSV(value))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// escapeCSV escapes a CSV field value
func escapeCSV(s string) string {
	// If the string contains comma, quotes, or newlines, wrap in quotes and escape quotes
	if strings.ContainsAny(s, ",\"\n") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + s + "\""
	}
	return s
}
