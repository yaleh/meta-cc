package mcp

import (
	"fmt"
)

// ToolDefinition represents an MCP tool with its metadata
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema interface{}
}

// Session-level tool definitions (query current session only, with _session suffix)

var sessionLevelTools = map[string]*ToolDefinition{
	"query_tools_session": {
		Name:        "query_tools_session",
		Description: "Query tool calls in the current session only. For project-level queries, use query_tools.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of tool calls to return (default: 20)",
				},
				"tool": map[string]interface{}{
					"type":        "string",
					"description": "Filter by tool name (e.g., 'Bash', 'Read', 'Edit')",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"success", "error"},
					"description": "Filter by execution status",
				},
				"output_format": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"json", "md"},
					"default": "json",
				},
			},
		},
	},
	"query_user_messages_session": {
		Name:        "query_user_messages_session",
		Description: "Search user messages in the current session only using regex pattern matching.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Regex pattern to match in message content (required)",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"default":     10,
					"description": "Maximum number of results (default: 10)",
				},
				"output_format": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"json", "md"},
					"default": "json",
				},
			},
			"required": []interface{}{"pattern"},
		},
	},
	"analyze_errors_session": {
		Name:        "analyze_errors_session",
		Description: "Analyze error patterns in the current session only.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"output_format": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"json", "md"},
					"default": "json",
				},
			},
		},
	},
	"query_tool_sequences_session": {
		Name:        "query_tool_sequences_session",
		Description: "Query repeated tool call sequences (workflow patterns) in the current session only.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Specific sequence pattern to match (e.g., 'Read -> Edit -> Bash')",
				},
				"min_occurrences": map[string]interface{}{
					"type":        "integer",
					"default":     3,
					"description": "Minimum occurrences to report (default 3)",
				},
				"output_format": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"json", "md"},
					"default": "json",
				},
			},
		},
	},
	"query_file_access_session": {
		Name:        "query_file_access_session",
		Description: "Query file access history (read/edit/write operations) in the current session only.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file": map[string]interface{}{
					"type":        "string",
					"description": "File path to query (required)",
				},
				"output_format": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"json", "md"},
					"default": "json",
				},
			},
			"required": []interface{}{"file"},
		},
	},
	"query_successful_prompts_session": {
		Name:        "query_successful_prompts_session",
		Description: "Query successful prompt patterns in the current session only.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "integer",
					"default":     10,
					"description": "Maximum number of results (default 10)",
				},
				"min_quality_score": map[string]interface{}{
					"type":        "number",
					"default":     0.8,
					"description": "Minimum quality score (0.0-1.0, default 0.8)",
				},
				"output_format": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"json", "md"},
					"default": "json",
				},
			},
		},
	},
	"query_context_session": {
		Name:        "query_context_session",
		Description: "Query context around specific errors in the current session only.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"error_signature": map[string]interface{}{
					"type":        "string",
					"description": "Error pattern ID to query (required)",
				},
				"window": map[string]interface{}{
					"type":        "integer",
					"default":     3,
					"description": "Context window size in turns before/after (default 3)",
				},
				"output_format": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"json", "md"},
					"default": "json",
				},
			},
			"required": []interface{}{"error_signature"},
		},
	},
	"get_session_stats": {
		Name:        "get_session_stats",
		Description: "Get statistics for the current Claude Code session",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"output_format": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"json", "md"},
					"default": "json",
				},
			},
		},
	},
}

// GetToolDefinition returns a tool definition by name
func GetToolDefinition(name string) *ToolDefinition {
	return sessionLevelTools[name]
}

// ListAllTools returns all registered session-level tools
func ListAllTools() []*ToolDefinition {
	tools := make([]*ToolDefinition, 0, len(sessionLevelTools))
	for _, tool := range sessionLevelTools {
		tools = append(tools, tool)
	}
	return tools
}

// BuildCommandArgs builds CLI command arguments for a session-level tool
// Session-level tools do NOT include --project flag
func BuildCommandArgs(toolName string, args map[string]interface{}) []string {
	outputFormat := "json"
	if format, ok := args["output_format"].(string); ok {
		outputFormat = format
	}

	var cmdArgs []string

	switch toolName {
	case "get_session_stats":
		cmdArgs = []string{"parse", "stats", "--output", outputFormat}

	case "analyze_errors_session":
		cmdArgs = []string{"analyze", "errors", "--output", outputFormat}

	case "query_tools_session":
		cmdArgs = []string{"query", "tools", "--output", outputFormat}

		if tool, ok := args["tool"].(string); ok && tool != "" {
			cmdArgs = append(cmdArgs, "--tool", tool)
		}
		if status, ok := args["status"].(string); ok && status != "" {
			cmdArgs = append(cmdArgs, "--status", status)
		}
		if limit, ok := args["limit"].(float64); ok {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
		} else {
			cmdArgs = append(cmdArgs, "--limit", "20")
		}

	case "query_user_messages_session":
		pattern, ok := args["pattern"].(string)
		if !ok || pattern == "" {
			return []string{"error", "pattern parameter is required"}
		}

		cmdArgs = []string{"query", "user-messages", "--match", pattern, "--output", outputFormat}

		if limit, ok := args["limit"].(float64); ok {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
		} else {
			cmdArgs = append(cmdArgs, "--limit", "10")
		}

	case "query_context_session":
		errorSignature, ok := args["error_signature"].(string)
		if !ok || errorSignature == "" {
			return []string{"error", "error_signature parameter is required"}
		}

		cmdArgs = []string{"query", "context", "--error-signature", errorSignature, "--output", outputFormat}

		if window, ok := args["window"].(float64); ok {
			cmdArgs = append(cmdArgs, "--window", fmt.Sprintf("%.0f", window))
		} else {
			cmdArgs = append(cmdArgs, "--window", "3")
		}

	case "query_tool_sequences_session":
		cmdArgs = []string{"query", "tool-sequences", "--output", outputFormat}

		if minOccurrences, ok := args["min_occurrences"].(float64); ok {
			cmdArgs = append(cmdArgs, "--min-occurrences", fmt.Sprintf("%.0f", minOccurrences))
		} else {
			cmdArgs = append(cmdArgs, "--min-occurrences", "3")
		}
		if pattern, ok := args["pattern"].(string); ok && pattern != "" {
			cmdArgs = append(cmdArgs, "--pattern", pattern)
		}

	case "query_file_access_session":
		file, ok := args["file"].(string)
		if !ok || file == "" {
			return []string{"error", "file parameter is required"}
		}

		cmdArgs = []string{"query", "file-access", "--file", file, "--output", outputFormat}

	case "query_successful_prompts_session":
		cmdArgs = []string{"query", "successful-prompts", "--output", outputFormat}

		if minQualityScore, ok := args["min_quality_score"].(float64); ok {
			cmdArgs = append(cmdArgs, "--min-quality-score", fmt.Sprintf("%.2f", minQualityScore))
		}
		if limit, ok := args["limit"].(float64); ok {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
		} else {
			cmdArgs = append(cmdArgs, "--limit", "10")
		}

	default:
		return []string{"error", "unknown tool: " + toolName}
	}

	return cmdArgs
}
