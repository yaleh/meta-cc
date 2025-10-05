package main

func getToolDefinitions() []Tool {
	return []Tool{
		{
			Name:        "get_session_stats",
			Description: "Get session statistics (turn count, tool usage, error rate). Always operates on current session only.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
				},
			},
		},
		{
			Name:        "analyze_errors",
			Description: "Analyze error patterns across project history (repeated failures, tool-specific errors, temporal trends). Default project-level scope enables discovery of persistent issues across sessions.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) enables cross-session pattern discovery, 'session' limits to current session only.",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results (default: '.[]')",
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
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
			},
		},
		{
			Name:        "extract_tools",
			Description: "Extract tool call history across all project sessions with pagination. Default project-level scope provides complete workflow timeline for pattern analysis.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"limit": {
						Type:        "number",
						Description: "Maximum number of tools to extract (default: 100)",
					},
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"stats_only": {
						Type:        "boolean",
						Description: "Return only statistics",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
			},
		},
		{
			Name:        "query_tools",
			Description: "Query tool call history across project with filters (tool name, status). Default project-level scope reveals cross-session usage patterns and trends.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"limit": {
						Type:        "number",
						Description: "Maximum number of results (default: 20)",
					},
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"tool": {
						Type:        "string",
						Description: "Filter by tool name (e.g., 'Bash', 'Read', 'Edit')",
					},
					"status": {
						Type:        "string",
						Description: "Filter by execution status (error or success)",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"stats_only": {
						Type:        "boolean",
						Description: "Return only statistics",
					},
					"stats_first": {
						Type:        "boolean",
						Description: "Return stats first, then details",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
			},
		},
		{
			Name:        "query_user_messages",
			Description: "Search user messages across all project sessions using regex patterns. Default project-level scope enables discovery of recurring prompt patterns and intent evolution.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"pattern": {
						Type:        "string",
						Description: "Regex pattern to match in message content (required)",
					},
					"limit": {
						Type:        "number",
						Description: "Maximum number of results (default: 10)",
					},
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
				Required: []string{"pattern"},
			},
		},
		{
			Name:        "query_context",
			Description: "Query context around specific errors across project history (turns before/after error occurrence). Default project-level scope helps identify if error patterns recur in similar contexts across sessions.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"error_signature": {
						Type:        "string",
						Description: "Error pattern ID to query (required)",
					},
					"window": {
						Type:        "number",
						Description: "Context window size in turns before/after (default: 3)",
					},
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
				Required: []string{"error_signature"},
			},
		},
		{
			Name:        "query_tool_sequences",
			Description: "Query repeated tool call sequences across project history (workflow patterns like 'Read->Edit->Bash'). Default project-level scope reveals your evolved workflow habits and automation opportunities.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"pattern": {
						Type:        "string",
						Description: "Specific sequence pattern to match (e.g., 'Read -> Edit -> Bash')",
					},
					"min_occurrences": {
						Type:        "number",
						Description: "Minimum occurrences to report (default: 3)",
					},
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
			},
		},
		{
			Name:        "query_file_access",
			Description: "Query file operation history across project (read/edit/write operations on specific files). Default project-level scope shows complete file evolution timeline.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"file": {
						Type:        "string",
						Description: "File path to query (required)",
					},
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
				Required: []string{"file"},
			},
		},
		{
			Name:        "query_project_state",
			Description: "Query project state evolution across all sessions (active files, task progression, change patterns). Default project-level scope provides comprehensive project timeline.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
			},
		},
		{
			Name:        "query_successful_prompts",
			Description: "Query historically successful prompt patterns across all project sessions (prompts that led to successful outcomes). Default project-level scope identifies your most effective prompting strategies over time.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"limit": {
						Type:        "number",
						Description: "Maximum number of results (default: 10)",
					},
					"min_quality_score": {
						Type:        "number",
						Description: "Minimum quality score (0.0-1.0, default: 0.8)",
					},
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
			},
		},
		{
			Name:        "query_tools_advanced",
			Description: "Query tool calls with SQL-like filter expressions across project sessions. Default project-level scope enables complex multi-condition analysis (e.g., 'tool=\"Bash\" AND status=\"error\" AND duration>5000').",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"where": {
						Type:        "string",
						Description: "SQL-like filter expression (e.g., \"tool='Bash' AND status='error'\")",
					},
					"limit": {
						Type:        "number",
						Description: "Maximum number of results (default: 20)",
					},
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
				Required: []string{"where"},
			},
		},
		{
			Name:        "aggregate_stats",
			Description: "Aggregate statistics grouped by field (tool, status, or uuid) across all project sessions. Default project-level scope provides comprehensive summary metrics (tool counts, error rates) for cross-session comparison.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"group_by": {
						Type:        "string",
						Description: "Field to group by: tool, status, or uuid (default: tool)",
					},
					"metrics": {
						Type:        "string",
						Description: "Comma-separated metrics (count, error_rate)",
					},
					"where": {
						Type:        "string",
						Description: "Optional filter expression",
					},
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
			},
		},
		{
			Name:        "query_time_series",
			Description: "Analyze metrics over time (tool call frequency, error rates) bucketed by hour/day/week across project history. Default project-level scope reveals temporal patterns and workflow evolution.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"interval": {
						Type:        "string",
						Description: "Time interval for bucketing: hour, day, or week (default: hour)",
					},
					"metric": {
						Type:        "string",
						Description: "Metric to analyze: tool-calls or error-rate (default: tool-calls)",
					},
					"where": {
						Type:        "string",
						Description: "Optional filter expression",
					},
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
					},
				},
			},
		},
		{
			Name:        "query_files",
			Description: "File-level operation statistics (total operations, edit/read/write counts, error rates) across all project sessions. Default project-level scope identifies files with persistent churn or error patterns.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"sort_by": {
						Type:        "string",
						Description: "Sort field: total_ops, edit_count, read_count, write_count, error_count, error_rate (default: total_ops)",
					},
					"top": {
						Type:        "number",
						Description: "Limit results to top N files (default: 20)",
					},
					"where": {
						Type:        "string",
						Description: "Optional filter expression",
					},
					"scope": {
						Type:        "string",
						Description: "Query scope: 'project' (default) or 'session'",
					},
					"output_format": {
						Type:        "string",
						Description: "Output format: jsonl or tsv (default: jsonl)",
					},
					"jq_filter": {
						Type:        "string",
						Description: "jq expression for filtering results",
					},
					"max_output_bytes": {
						Type:        "number",
						Description: "Maximum output size in bytes (default: 51200)",
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
