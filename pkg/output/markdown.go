package output

import (
	"fmt"
	"strings"

	"github.com/yale/meta-cc/internal/parser"
)

// FormatMarkdown formats data as Markdown
// Supports SessionEntry arrays and ToolCall arrays
func FormatMarkdown(data interface{}) (string, error) {
	switch v := data.(type) {
	case []parser.ToolCall:
		return formatToolCallsMarkdown(v), nil
	case []parser.SessionEntry:
		return formatSessionEntriesMarkdown(v), nil
	default:
		return "", fmt.Errorf("unsupported data type for Markdown formatting")
	}
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
