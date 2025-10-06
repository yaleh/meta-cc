package output

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/yale/meta-cc/internal/parser"
)

// FormatTSV formats data as TSV (Tab-Separated Values)
// Supports any data type via reflection (generic formatting)
// TSV is ~50% smaller than JSON due to:
// - No quotes around values
// - No field names per record
// - No JSON structure overhead
//
// Supported types:
// - []parser.ToolCall (optimized)
// - Any struct (vertical format: key\tvalue)
// - Any []struct (table format with headers)
func FormatTSV(data interface{}) (string, error) {
	// Type switch for optimized formatting
	switch v := data.(type) {
	case []parser.ToolCall:
		// Optimized path for ToolCall (most common case)
		return formatToolCallsTSV(v), nil
	default:
		// Generic path using reflection
		return FormatGenericTSV(data)
	}
}

// formatToolCallsTSV formats ToolCall slice as TSV (optimized)
func formatToolCallsTSV(tools []parser.ToolCall) string {
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

// FormatGenericTSV formats any data type as TSV using reflection
// Supports: single structs (vertical format), slice of structs (table format)
func FormatGenericTSV(data interface{}) (string, error) {
	v := reflect.ValueOf(data)

	// Handle nil
	if !v.IsValid() {
		return "", nil
	}

	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}

	// Handle slice
	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			return "", nil
		}

		// Get headers from first element
		elem := v.Index(0)
		headers := getStructFields(elem)

		var sb strings.Builder
		sb.WriteString(strings.Join(headers, "\t"))
		sb.WriteString("\n")

		// Get rows
		for i := 0; i < v.Len(); i++ {
			row := getStructValues(v.Index(i))
			sb.WriteString(strings.Join(row, "\t"))
			sb.WriteString("\n")
		}

		return sb.String(), nil
	}

	// Handle single struct (vertical format: key\tvalue)
	if v.Kind() == reflect.Struct {
		headers := getStructFields(v)
		values := getStructValues(v)

		var sb strings.Builder
		for i, header := range headers {
			sb.WriteString(fmt.Sprintf("%s\t%s\n", header, values[i]))
		}

		return sb.String(), nil
	}

	return "", fmt.Errorf("unsupported data type for TSV formatting: %T", data)
}

// getStructFields extracts field names from a struct using reflection
func getStructFields(v reflect.Value) []string {
	var fields []string

	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fields
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fields
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Use JSON tag if available, otherwise use field name
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			// Parse JSON tag (format: "field_name,omitempty")
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			}
		}

		fields = append(fields, fieldName)
	}

	return fields
}

// getStructValues extracts field values from a struct using reflection
func getStructValues(v reflect.Value) []string {
	var values []string

	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return values
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return values
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		value := v.Field(i)
		values = append(values, formatTSVValue(value))
	}

	return values
}

// formatTSVValue formats a reflect.Value as string for TSV
func formatTSVValue(v reflect.Value) string {
	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		// Escape tabs and newlines
		s := v.String()
		return escapeTSV(s)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", v.Float())

	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())

	case reflect.Map, reflect.Slice, reflect.Struct:
		// Serialize complex types as JSON (compact)
		data, err := json.Marshal(v.Interface())
		if err != nil {
			return fmt.Sprintf("%v", v.Interface())
		}
		s := string(data)
		return escapeTSV(s)

	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
