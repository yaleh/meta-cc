package mcp

import (
	"fmt"
)

// Project-level tool definitions (query all sessions in project, NO _session suffix)

var projectLevelTools = map[string]*ToolDefinition{
	"query_tools": {
		Name:        "query_tools",
		Description: "Query tool calls across all sessions in the project. Returns tool execution history with fields: 'tool_name' (not 'tool'), 'status', 'error', 'timestamp'. See docs/guides/mcp-jq-quick-reference.md for jq syntax.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of tool calls to return (default: 20)",
				},
				"tool": map[string]interface{}{
					"type":        "string",
					"description": "Filter by tool name (e.g., 'Bash', 'Edit', 'Read')",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"success", "error"},
					"description": "Filter by execution status",
				},
				"where": map[string]interface{}{
					"type":        "string",
					"description": "SQL-like filter expression",
				},
				"output_format": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"json", "md"},
					"default": "json",
				},
			},
		},
	},
	"query_user_messages": {
		Name:        "query_user_messages",
		Description: "Search user messages across all sessions in the project using regex pattern matching. Returns messages with 'content' field (not 'message_content'). See docs/guides/mcp-jq-quick-reference.md for field names and jq syntax.",
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
	"get_stats": {
		Name:        "get_stats",
		Description: "Get statistics for all sessions in the project. Returns tool usage counts, error rates, and session metrics.",
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
	"analyze_errors": {
		Name:        "analyze_errors",
		Description: "Analyze error patterns across all sessions in the project. Detects repeated errors and common failure modes.",
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
	"query_tool_sequences": {
		Name:        "query_tool_sequences",
		Description: "Query repeated tool call sequences (workflow patterns) across all sessions in the project.",
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
	"query_file_access": {
		Name:        "query_file_access",
		Description: "Query file access history (read/edit/write operations) across all sessions in the project.",
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
	"query_successful_prompts": {
		Name:        "query_successful_prompts",
		Description: "Query successful prompt patterns across all sessions in the project (Stage 8.12).",
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
	"query_context": {
		Name:        "query_context",
		Description: "Query context around specific errors across all sessions in the project (Stage 8.10).",
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
}

// GetProjectLevelTool returns a project-level tool definition by name
func GetProjectLevelTool(name string) *ToolDefinition {
	return projectLevelTools[name]
}

// ListProjectLevelTools returns all registered project-level tools
func ListProjectLevelTools() []*ToolDefinition {
	tools := make([]*ToolDefinition, 0, len(projectLevelTools))
	for _, tool := range projectLevelTools {
		tools = append(tools, tool)
	}
	return tools
}

// BuildProjectLevelCommandArgs builds CLI command arguments for a project-level tool
// Project-level tools MUST include --project . flag
func BuildProjectLevelCommandArgs(toolName string, args map[string]interface{}) []string {
	outputFormat := "json"
	if format, ok := args["output_format"].(string); ok {
		outputFormat = format
	}

	var cmdArgs []string

	switch toolName {
	case "get_stats":
		cmdArgs = []string{"parse", "stats", "--project", ".", "--output", outputFormat}

	case "analyze_errors":
		cmdArgs = []string{"analyze", "errors", "--project", ".", "--output", outputFormat}

	case "query_tools":
		cmdArgs = []string{"query", "tools", "--project", ".", "--output", outputFormat}

		if tool, ok := args["tool"].(string); ok && tool != "" {
			cmdArgs = append(cmdArgs, "--tool", tool)
		}
		if status, ok := args["status"].(string); ok && status != "" {
			cmdArgs = append(cmdArgs, "--status", status)
		}
		if where, ok := args["where"].(string); ok && where != "" {
			cmdArgs = append(cmdArgs, "--where", where)
		}
		if limit, ok := args["limit"].(float64); ok {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
		} else {
			cmdArgs = append(cmdArgs, "--limit", "20")
		}

	case "query_user_messages":
		pattern, ok := args["pattern"].(string)
		if !ok || pattern == "" {
			return []string{"error", "pattern parameter is required"}
		}

		cmdArgs = []string{"query", "user-messages", "--project", ".", "--match", pattern, "--output", outputFormat}

		if limit, ok := args["limit"].(float64); ok {
			cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%.0f", limit))
		} else {
			cmdArgs = append(cmdArgs, "--limit", "10")
		}

	case "query_context":
		errorSignature, ok := args["error_signature"].(string)
		if !ok || errorSignature == "" {
			return []string{"error", "error_signature parameter is required"}
		}

		cmdArgs = []string{"query", "context", "--project", ".", "--error-signature", errorSignature, "--output", outputFormat}

		if window, ok := args["window"].(float64); ok {
			cmdArgs = append(cmdArgs, "--window", fmt.Sprintf("%.0f", window))
		} else {
			cmdArgs = append(cmdArgs, "--window", "3")
		}

	case "query_tool_sequences":
		cmdArgs = []string{"query", "tool-sequences", "--project", ".", "--output", outputFormat}

		if minOccurrences, ok := args["min_occurrences"].(float64); ok {
			cmdArgs = append(cmdArgs, "--min-occurrences", fmt.Sprintf("%.0f", minOccurrences))
		} else {
			cmdArgs = append(cmdArgs, "--min-occurrences", "3")
		}
		if pattern, ok := args["pattern"].(string); ok && pattern != "" {
			cmdArgs = append(cmdArgs, "--pattern", pattern)
		}

	case "query_file_access":
		file, ok := args["file"].(string)
		if !ok || file == "" {
			return []string{"error", "file parameter is required"}
		}

		cmdArgs = []string{"query", "file-access", "--project", ".", "--file", file, "--output", outputFormat}

	case "query_successful_prompts":
		cmdArgs = []string{"query", "successful-prompts", "--project", ".", "--output", outputFormat}

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
