package cmd

// getConsolidatedToolsList returns the MCP tool definitions with scope parameter
// Phase 12 Revision: Consolidate _session tools into scope parameter
func getConsolidatedToolsList() []map[string]interface{} {
	// Common scope property for all tools (except get_session_stats for backward compat)
	// Phase 12: Default to 'project' for cross-session meta-cognition analysis
	scopeProperty := map[string]interface{}{
		"type":        "string",
		"enum":        []string{"session", "project"},
		"default":     "project",
		"description": "Query scope: 'project' (default) enables cross-session pattern discovery, 'session' limits to current session only. Project-level analysis is recommended for meta-cognition insights.",
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
			"description": "Analyze error patterns across project history (repeated failures, tool-specific errors, temporal trends). Default project-level scope enables discovery of persistent issues across sessions. Use for meta-cognition: identifying systematic workflow problems, debugging recurring issues, or tracking error resolution over time.",
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
			"description": "Extract tool call history across all project sessions with pagination. Default project-level scope provides complete workflow timeline for pattern analysis. Use for meta-cognition: discovering tool usage evolution, identifying workflow optimization opportunities, or exporting data for external analysis.",
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
			"description": "Query tool call history across project with filters (tool name, status). Default project-level scope reveals cross-session usage patterns and trends. Use for meta-cognition: analyzing tool effectiveness over time, identifying frequently failing operations, or understanding workflow evolution. Supports filtering by tool name and execution status.",
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
			"description": "Search user messages across all project sessions using regex patterns. Default project-level scope enables discovery of recurring prompt patterns and intent evolution. Use for meta-cognition: analyzing how your questions change over time, identifying successful prompt strategies, or reviewing conversation context across sessions.",
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
			"description": "Query context around specific errors across project history (turns before/after error occurrence). Default project-level scope helps identify if error patterns recur in similar contexts across sessions. Use for meta-cognition: understanding systematic causes of errors, recognizing environmental triggers, or comparing error contexts over time.",
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
			"description": "Query repeated tool call sequences across project history (workflow patterns like 'Read->Edit->Bash'). Default project-level scope reveals your evolved workflow habits and automation opportunities. Use for meta-cognition: discovering repetitive patterns worth automating, identifying inefficient workflows, or understanding how your tool usage evolves.",
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
			"description": "Query file operation history across project (read/edit/write operations on specific files). Default project-level scope shows complete file evolution timeline. Use for meta-cognition: identifying frequently churned files, understanding refactoring patterns, or tracking file modification frequency over time.",
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
			"description": "Query project state evolution across all sessions (active files, task progression, change patterns). Default project-level scope provides comprehensive project timeline. Use for meta-cognition: understanding how project focus shifts over time, tracking long-term progress, or identifying stalled initiatives.",
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
			"description": "Query historically successful prompt patterns across all project sessions (prompts that led to successful outcomes). Default project-level scope identifies your most effective prompting strategies over time. Use for meta-cognition: learning what prompt patterns work best for you, improving future interactions, or generating prompt templates.",
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
