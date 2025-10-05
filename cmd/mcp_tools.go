package cmd

// getConsolidatedToolsList returns the MCP tool definitions with scope parameter
// Phase 12 Revision: Consolidate _session tools into scope parameter
func getConsolidatedToolsList() []map[string]interface{} {
	// Common scope property for all tools (except get_session_stats for backward compat)
	scopeProperty := map[string]interface{}{
		"type":        "string",
		"enum":        []string{"session", "project"},
		"default":     "session",
		"description": "Query scope: 'session' for current session only, 'project' for all sessions in project",
	}

	outputFormatProperty := map[string]interface{}{
		"type":    "string",
		"enum":    []string{"jsonl", "tsv"},
		"default": "jsonl",
	}

	return []map[string]interface{}{
		// Backward compatibility: get_session_stats remains session-only
		{
			"name":        "get_session_stats",
			"description": "Get session statistics (turn count, tool usage, error rate). Use when user asks about session performance or workflow efficiency. Always operates on current session only.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"output_format": outputFormatProperty,
				},
			},
		},

		// Core analysis tools with scope parameter
		{
			"name":        "analyze_errors",
			"description": "Analyze error patterns (repeated failures, tool-specific errors). Use for investigating error trends or debugging recurring issues. Supports project-wide analysis or session-only via scope parameter.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope":         scopeProperty,
					"output_format": outputFormatProperty,
				},
			},
		},
		{
			"name":        "extract_tools",
			"description": "Extract tool call history with pagination. Use for bulk data export or analyzing tool usage patterns. Supports project-wide or session-only scope.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": scopeProperty,
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     100,
						"description": "Maximum number of tools to extract (default 100, prevents overflow)",
					},
					"output_format": outputFormatProperty,
				},
			},
		},

		// Query tools with scope parameter
		{
			"name":        "query_tools",
			"description": "Query tool call history with filters (tool name, status, time range). Use for investigating specific tool usage patterns or debugging errors. Supports filtering by tool name and execution status.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": scopeProperty,
					"tool": map[string]interface{}{
						"type":        "string",
						"description": "Filter by tool name (e.g., 'Bash', 'Read', 'Edit')",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"error", "success"},
						"description": "Filter by execution status",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     20,
						"description": "Maximum number of results (default 20)",
					},
					"output_format": outputFormatProperty,
				},
			},
		},
		{
			"name":        "query_user_messages",
			"description": "Search user messages using regex patterns. Use for finding specific prompts, reviewing conversation history, or analyzing prompt patterns. Returns user messages matching the pattern with context.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": scopeProperty,
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Regex pattern to match in message content (required)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     10,
						"description": "Maximum number of results (default 10)",
					},
					"output_format": outputFormatProperty,
				},
				"required": []string{"pattern"},
			},
		},
		{
			"name":        "query_context",
			"description": "Query context around specific errors (turns before/after error occurrence). Use for understanding error context and debugging specific failure patterns.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": scopeProperty,
					"error_signature": map[string]interface{}{
						"type":        "string",
						"description": "Error pattern ID to query (required)",
					},
					"window": map[string]interface{}{
						"type":        "integer",
						"default":     3,
						"description": "Context window size in turns before/after (default 3)",
					},
					"output_format": outputFormatProperty,
				},
				"required": []string{"error_signature"},
			},
		},
		{
			"name":        "query_tool_sequences",
			"description": "Query repeated tool call sequences (workflow patterns). Use for identifying common workflows, automation opportunities, or inefficient tool usage patterns.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": scopeProperty,
					"min_occurrences": map[string]interface{}{
						"type":        "integer",
						"default":     3,
						"description": "Minimum occurrences to report (default 3)",
					},
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Specific sequence pattern to match (e.g., 'Read -> Edit -> Bash')",
					},
					"output_format": outputFormatProperty,
				},
			},
		},
		{
			"name":        "query_file_access",
			"description": "Query file operation history (read/edit/write operations on specific files). Use for tracking file modification history or understanding file-level workflows.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": scopeProperty,
					"file": map[string]interface{}{
						"type":        "string",
						"description": "File path to query (required)",
					},
					"output_format": outputFormatProperty,
				},
				"required": []string{"file"},
			},
		},
		{
			"name":        "query_project_state",
			"description": "Query current project state from session (active files, pending tasks, recent changes). Use for understanding project status or generating contextual recommendations.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope":         scopeProperty,
					"output_format": outputFormatProperty,
				},
			},
		},
		{
			"name":        "query_successful_prompts",
			"description": "Query historically successful prompt patterns (prompts that led to successful outcomes). Use for learning effective prompting strategies or generating prompt suggestions.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": scopeProperty,
					"min_quality_score": map[string]interface{}{
						"type":        "number",
						"default":     0.8,
						"description": "Minimum quality score (0.0-1.0, default 0.8)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     10,
						"description": "Maximum number of results (default 10)",
					},
					"output_format": outputFormatProperty,
				},
			},
		},

		// Phase 10: Advanced query tools with scope
		{
			"name":        "query_tools_advanced",
			"description": "Query tool calls with SQL-like filter expressions. Use for complex filtering scenarios that simple filters cannot handle (e.g., combining multiple conditions).",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": scopeProperty,
					"where": map[string]interface{}{
						"type":        "string",
						"description": "SQL-like filter expression (e.g., \"tool='Bash' AND status='error'\")",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     20,
						"description": "Maximum number of results (default 20)",
					},
					"output_format": outputFormatProperty,
				},
				"required": []string{"where"},
			},
		},
		{
			"name":        "aggregate_stats",
			"description": "Aggregate statistics grouped by field (tool, status, or uuid). Use for generating summary reports or identifying high-level trends in tool usage.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": scopeProperty,
					"group_by": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"tool", "status", "uuid"},
						"default":     "tool",
						"description": "Field to group by (tool, status, or uuid)",
					},
					"metrics": map[string]interface{}{
						"type":        "string",
						"default":     "count,error_rate",
						"description": "Comma-separated metrics (count, error_rate)",
					},
					"where": map[string]interface{}{
						"type":        "string",
						"description": "Optional filter expression",
					},
					"output_format": outputFormatProperty,
				},
			},
		},
		{
			"name":        "query_time_series",
			"description": "Analyze metrics over time (tool call frequency, error rates by hour/day/week). Use for identifying temporal patterns or workflow trends over time.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": scopeProperty,
					"metric": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"tool-calls", "error-rate"},
						"default":     "tool-calls",
						"description": "Metric to analyze (tool-calls or error-rate)",
					},
					"interval": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"hour", "day", "week"},
						"default":     "hour",
						"description": "Time interval for bucketing",
					},
					"where": map[string]interface{}{
						"type":        "string",
						"description": "Optional filter expression",
					},
					"output_format": outputFormatProperty,
				},
			},
		},
		{
			"name":        "query_files",
			"description": "File-level operation statistics (total operations, edit/read/write counts, error rates by file). Use for identifying frequently modified files or file-level hotspots.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": scopeProperty,
					"sort_by": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"total_ops", "edit_count", "read_count", "write_count", "error_count", "error_rate"},
						"default":     "total_ops",
						"description": "Sort field",
					},
					"top": map[string]interface{}{
						"type":        "integer",
						"default":     20,
						"description": "Limit results to top N files",
					},
					"where": map[string]interface{}{
						"type":        "string",
						"description": "Optional filter expression",
					},
					"output_format": outputFormatProperty,
				},
			},
		},
	}
}
