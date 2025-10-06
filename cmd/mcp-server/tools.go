package main

// Tool Description Template:
// Format: "<action> <object>. Default scope: <project/session>."
// Requirements:
//   - Maximum length: 100 characters
//   - Must include "Default scope:" suffix
//   - Use active voice and imperative form
//   - Focus on "what" not "how" or "why"
//
// Examples:
//   - Good: "Query tool calls with filters. Default scope: project."
//   - Bad:  "Query tool call history across project with filters (tool name, status). Default project-level scope reveals cross-session usage patterns and trends."

// StandardToolParameters returns the standard set of parameters for all MCP tools
func StandardToolParameters() map[string]Property {
	return map[string]Property{
		"scope": {
			Type:        "string",
			Description: "Query scope: 'project' (default) or 'session'",
		},
		"jq_filter": {
			Type:        "string",
			Description: "jq expression for filtering (default: '.[]')",
		},
		"stats_only": {
			Type:        "boolean",
			Description: "Return only statistics (default: false)",
		},
		"stats_first": {
			Type:        "boolean",
			Description: "Return stats first, then details (default: false)",
		},
		"max_output_bytes": {
			Type:        "number",
			Description: "Max output size in bytes (default: 51200)",
		},
		"output_format": {
			Type:        "string",
			Description: "Output format: jsonl or tsv (default: jsonl)",
		},
	}
}

// MergeParameters merges tool-specific params with standard params
func MergeParameters(specific map[string]Property) map[string]Property {
	result := make(map[string]Property)

	// Add standard parameters first
	for k, v := range StandardToolParameters() {
		result[k] = v
	}

	// Override/add specific parameters
	for k, v := range specific {
		result[k] = v
	}

	return result
}

func getToolDefinitions() []Tool {
	return []Tool{
		{
			Name:        "get_session_stats",
			Description: "Get session statistics. Default scope: session.",
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: MergeParameters(map[string]Property{}),
			},
		},
		{
			Name:        "query_tools",
			Description: "Query tool calls with filters. Default scope: project.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: MergeParameters(map[string]Property{
					"limit": {
						Type:        "number",
						Description: "Max results (default: 20)",
					},
					"tool": {
						Type:        "string",
						Description: "Filter by tool name",
					},
					"status": {
						Type:        "string",
						Description: "Filter by status (error/success)",
					},
				}),
			},
		},
		{
			Name:        "query_user_messages",
			Description: "Search user messages with regex. May contain large outputs. Default scope: project.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: MergeParameters(map[string]Property{
					"pattern": {
						Type:        "string",
						Description: "Regex pattern to match (required)",
					},
					"limit": {
						Type:        "number",
						Description: "Max results (default: 10)",
					},
					"max_message_length": {
						Type:        "number",
						Description: "Max chars per message content (default: 500, prevents huge summaries)",
					},
					"content_summary": {
						Type:        "boolean",
						Description: "Return only turn/timestamp/preview (100 chars), skip full content",
					},
				}),
				Required: []string{"pattern"},
			},
		},
		{
			Name:        "query_context",
			Description: "Query error context. Default scope: project.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: MergeParameters(map[string]Property{
					"error_signature": {
						Type:        "string",
						Description: "Error pattern ID (required)",
					},
					"window": {
						Type:        "number",
						Description: "Context window size (default: 3)",
					},
				}),
				Required: []string{"error_signature"},
			},
		},
		{
			Name:        "query_tool_sequences",
			Description: "Query workflow patterns. Default scope: project.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: MergeParameters(map[string]Property{
					"pattern": {
						Type:        "string",
						Description: "Sequence pattern to match",
					},
					"min_occurrences": {
						Type:        "number",
						Description: "Min occurrences (default: 3)",
					},
				}),
			},
		},
		{
			Name:        "query_file_access",
			Description: "Query file operation history. Default scope: project.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: MergeParameters(map[string]Property{
					"file": {
						Type:        "string",
						Description: "File path (required)",
					},
				}),
				Required: []string{"file"},
			},
		},
		{
			Name:        "query_project_state",
			Description: "Query project state evolution. Default scope: project.",
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: MergeParameters(map[string]Property{}),
			},
		},
		{
			Name:        "query_successful_prompts",
			Description: "Query successful prompt patterns. Default scope: project.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: MergeParameters(map[string]Property{
					"limit": {
						Type:        "number",
						Description: "Max results (default: 10)",
					},
					"min_quality_score": {
						Type:        "number",
						Description: "Min quality score (default: 0.8)",
					},
				}),
			},
		},
		{
			Name:        "query_tools_advanced",
			Description: "Query tools with SQL-like filters. Default scope: project.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: MergeParameters(map[string]Property{
					"where": {
						Type:        "string",
						Description: "SQL-like filter expression (required)",
					},
					"limit": {
						Type:        "number",
						Description: "Max results (default: 20)",
					},
				}),
				Required: []string{"where"},
			},
		},
		{
			Name:        "query_time_series",
			Description: "Analyze metrics over time. Default scope: project.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: MergeParameters(map[string]Property{
					"interval": {
						Type:        "string",
						Description: "Time interval (hour/day/week, default: hour)",
					},
					"metric": {
						Type:        "string",
						Description: "Metric to analyze (default: tool-calls)",
					},
					"where": {
						Type:        "string",
						Description: "Optional filter expression",
					},
				}),
			},
		},
		{
			Name:        "query_files",
			Description: "File-level operation stats. Default scope: project.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: MergeParameters(map[string]Property{
					"sort_by": {
						Type:        "string",
						Description: "Sort field (default: total_ops)",
					},
					"top": {
						Type:        "number",
						Description: "Top N files (default: 20)",
					},
					"where": {
						Type:        "string",
						Description: "Optional filter expression",
					},
				}),
			},
		},
		{
			Name:        "cleanup_temp_files",
			Description: "Remove old temporary MCP files. Default scope: none.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"max_age_days": {
						Type:        "number",
						Description: "Max file age in days (default: 7)",
					},
				},
			},
		},
	}
}

type Tool struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema ToolSchema `json:"inputSchema"`
}

type ToolSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
