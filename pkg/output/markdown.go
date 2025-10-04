package output

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/yale/meta-cc/internal/parser"
)

// FormatMarkdown formats data as Markdown
// Supports SessionEntry arrays, ToolCall arrays, and UserMessage arrays
func FormatMarkdown(data interface{}) (string, error) {
	switch v := data.(type) {
	case []parser.ToolCall:
		return formatToolCallsMarkdown(v), nil
	case []parser.SessionEntry:
		return formatSessionEntriesMarkdown(v), nil
	default:
		// Try to handle UserMessage type using reflection (defined in cmd package)
		if isUserMessageSlice(data) {
			return formatUserMessagesMarkdown(data), nil
		}
		return "", fmt.Errorf("unsupported data type for Markdown formatting")
	}
}

// isUserMessageSlice checks if data is a slice of structs with UserMessage fields
func isUserMessageSlice(data interface{}) bool {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return false
	}

	if v.Len() == 0 {
		// Empty slice - check the type
		t := v.Type().Elem()
		return t.Kind() == reflect.Struct && hasUserMessageFields(t)
	}

	// Check first element
	elem := v.Index(0)
	if elem.Kind() != reflect.Struct {
		return false
	}

	return hasUserMessageFields(elem.Type())
}

// hasUserMessageFields checks if a struct has the expected UserMessage fields
func hasUserMessageFields(t reflect.Type) bool {
	// UserMessage should have: TurnSequence, UUID, Timestamp, Content
	hasTurnSequence := false
	hasUUID := false
	hasTimestamp := false
	hasContent := false

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		switch field.Name {
		case "TurnSequence":
			hasTurnSequence = true
		case "UUID":
			hasUUID = true
		case "Timestamp":
			hasTimestamp = true
		case "Content":
			hasContent = true
		}
	}

	return hasTurnSequence && hasUUID && hasTimestamp && hasContent
}

func formatToolCallsMarkdown(toolCalls []parser.ToolCall) string {
	var sb strings.Builder

	sb.WriteString("# Tool Calls\n\n")
	sb.WriteString("| Tool | Input | Output | Status |\n")
	sb.WriteString("|------|-------|--------|--------|\n")

	for _, tc := range toolCalls {
		// Format Input (simplified display)
		inputStr := formatInputForTable(tc.Input)

		// Format Output (truncate long output)
		outputStr := tc.Output
		if len(outputStr) > 50 {
			outputStr = outputStr[:47] + "..."
		}

		// Status
		status := tc.Status
		if status == "" {
			status = "-"
		}
		if tc.Error != "" {
			status = fmt.Sprintf("error: %s", tc.Error)
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			tc.ToolName,
			inputStr,
			outputStr,
			status,
		))
	}

	return sb.String()
}

func formatSessionEntriesMarkdown(entries []parser.SessionEntry) string {
	var sb strings.Builder

	sb.WriteString("# Session Turns\n\n")

	for i, entry := range entries {
		sb.WriteString(fmt.Sprintf("## Turn %d\n\n", i+1))
		sb.WriteString(fmt.Sprintf("- **Type**: %s\n", entry.Type))
		sb.WriteString(fmt.Sprintf("- **UUID**: %s\n", entry.UUID))
		sb.WriteString(fmt.Sprintf("- **Timestamp**: %s\n", entry.Timestamp))

		if entry.Message != nil {
			sb.WriteString(fmt.Sprintf("- **Role**: %s\n", entry.Message.Role))

			// Extract text content
			for _, block := range entry.Message.Content {
				if block.Type == "text" && block.Text != "" {
					sb.WriteString(fmt.Sprintf("- **Content**: %s\n", block.Text))
				}
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func formatInputForTable(input map[string]interface{}) string {
	if len(input) == 0 {
		return "-"
	}

	var parts []string
	for k, v := range input {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	result := strings.Join(parts, ", ")
	if len(result) > 40 {
		result = result[:37] + "..."
	}

	return result
}

// formatUserMessagesMarkdown formats UserMessage slices as Markdown using reflection
func formatUserMessagesMarkdown(data interface{}) string {
	var sb strings.Builder

	sb.WriteString("# User Messages\n\n")
	sb.WriteString("| Turn | UUID | Timestamp | Content |\n")
	sb.WriteString("|------|------|-----------|----------|\n")

	v := reflect.ValueOf(data)
	for i := 0; i < v.Len(); i++ {
		msg := v.Index(i)

		// Extract fields using reflection
		turnSeq := getFieldValue(msg, "TurnSequence")
		uuid := getFieldValue(msg, "UUID")
		timestamp := getFieldValue(msg, "Timestamp")
		content := getFieldValue(msg, "Content")

		// Truncate content if too long (for table display)
		contentStr := fmt.Sprintf("%v", content)
		if len(contentStr) > 60 {
			contentStr = contentStr[:57] + "..."
		}

		// Escape pipe characters in content to avoid breaking table
		contentStr = strings.ReplaceAll(contentStr, "|", "\\|")

		sb.WriteString(fmt.Sprintf("| %v | %v | %v | %s |\n",
			turnSeq,
			uuid,
			timestamp,
			contentStr,
		))
	}

	return sb.String()
}

// getFieldValue gets a field value from a struct using reflection
func getFieldValue(v reflect.Value, fieldName string) interface{} {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return ""
	}
	return field.Interface()
}
