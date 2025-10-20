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
			Description: "jq expression for filtering (default: '.[]'). IMPORTANT: Do NOT wrap in quotes - use raw jq expression like: .[] | {field: .field}",
		},
		"stats_only": {
			Type:        "boolean",
			Description: "Return only statistics (default: false)",
		},
		"stats_first": {
			Type:        "boolean",
			Description: "Return stats first, then details (default: false)",
		},
		"inline_threshold_bytes": {
			Type:        "number",
			Description: "Threshold for inline vs file_ref mode in bytes (default: 8192). Can also set META_CC_INLINE_THRESHOLD env var",
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

// buildToolSchema creates a ToolSchema with merged parameters
func buildToolSchema(properties map[string]Property, required ...string) ToolSchema {
	schema := ToolSchema{
		Type:       "object",
		Properties: MergeParameters(properties),
	}
	if len(required) > 0 {
		schema.Required = required
	}
	return schema
}

// buildTool creates a Tool with the given name, description, and schema
func buildTool(name, description string, properties map[string]Property, required ...string) Tool {
	return Tool{
		Name:        name,
		Description: description,
		InputSchema: buildToolSchema(properties, required...),
	}
}

func getToolDefinitions() []Tool {
	return []Tool{
		buildTool("get_session_stats", "Get session statistics. Default scope: session.", map[string]Property{}),
		buildTool("query_tools", "Query tool calls with filters. Default scope: project.", map[string]Property{
			// Tier 2: Filtering
			"tool": {
				Type:        "string",
				Description: "Filter by tool name",
			},
			"status": {
				Type:        "string",
				Description: "Filter by status (error/success)",
			},
			// Tier 4: Output Control
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}),
		buildTool("query_user_messages", "Search user messages with regex. May contain large outputs. Default scope: project.", map[string]Property{
			// Tier 1: Required
			"pattern": {
				Type:        "string",
				Description: "Regex pattern to match (required)",
			},
			// Tier 3: Range
			"max_message_length": {
				Type:        "number",
				Description: "Max chars per message content (default: 0 = no truncation, rely on hybrid mode for large results)",
			},
			// Tier 4: Output Control
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
			"content_summary": {
				Type:        "boolean",
				Description: "Return only turn/timestamp/preview (100 chars), skip full content. Use hybrid mode instead for better information preservation.",
			},
		}, "pattern"),
		buildTool("query_context", "Query error context. Default scope: project.", map[string]Property{
			"error_signature": {
				Type:        "string",
				Description: "Error pattern ID (required)",
			},
			"window": {
				Type:        "number",
				Description: "Context window size (default: 3)",
			},
		}, "error_signature"),
		buildTool("query_tool_sequences", "Query workflow patterns. Default scope: project.", map[string]Property{
			// Tier 2: Filtering
			"pattern": {
				Type:        "string",
				Description: "Sequence pattern to match",
			},
			"include_builtin_tools": {
				Type:        "boolean",
				Description: "Include built-in tools (Bash, Read, Edit, etc.). Default: false (cleaner workflow patterns, 35x faster)",
			},
			// Tier 3: Range
			"min_occurrences": {
				Type:        "number",
				Description: "Min occurrences (default: 3)",
			},
		}),
		buildTool("query_file_access", "Query file operation history. Default scope: project.", map[string]Property{
			"file": {
				Type:        "string",
				Description: "File path (required)",
			},
		}, "file"),
		buildTool("query_project_state", "Query project state evolution. Default scope: project.", map[string]Property{}),
		buildTool("query_successful_prompts", "Query successful prompt patterns. Default scope: project.", map[string]Property{
			// Tier 3: Range
			"min_quality_score": {
				Type:        "number",
				Description: "Min quality score (default: 0.8)",
			},
			// Tier 4: Output Control
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}),
		buildTool("query_tools_advanced", "Query tools with SQL-like filters. Default scope: project.", map[string]Property{
			"where": {
				Type:        "string",
				Description: "SQL-like filter expression (required)",
			},
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}, "where"),
		buildTool("query_time_series", "Analyze metrics over time. Default scope: project.", map[string]Property{
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
		buildTool("query_assistant_messages", "Query assistant messages with pattern matching and filtering. Default scope: project.", map[string]Property{
			"pattern": {
				Type:        "string",
				Description: "Regex pattern to match text content",
			},
			"min_tools": {
				Type:        "number",
				Description: "Minimum tool use count",
			},
			"max_tools": {
				Type:        "number",
				Description: "Maximum tool use count",
			},
			"min_tokens_output": {
				Type:        "number",
				Description: "Minimum output tokens",
			},
			"min_length": {
				Type:        "number",
				Description: "Minimum text length",
			},
			"max_length": {
				Type:        "number",
				Description: "Maximum text length",
			},
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}),
		buildTool("query_conversation", "Query conversation turns (user+assistant pairs). Default scope: project.", map[string]Property{
			// Tier 2: Filtering
			"pattern": {
				Type:        "string",
				Description: "Regex pattern (user or assistant content)",
			},
			"pattern_target": {
				Type:        "string",
				Description: "Pattern target: user, assistant, any (default: any)",
			},
			// Tier 3: Range (turn ranges, then duration ranges)
			"start_turn": {
				Type:        "number",
				Description: "Starting turn sequence",
			},
			"end_turn": {
				Type:        "number",
				Description: "Ending turn sequence",
			},
			"min_duration": {
				Type:        "number",
				Description: "Minimum response duration (ms)",
			},
			"max_duration": {
				Type:        "number",
				Description: "Maximum response duration (ms)",
			},
			// Tier 4: Output Control
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}),
		buildTool("query_files", "File operation stats (returns array). Use jq_filter for filtering. Default scope: project.", map[string]Property{
			"threshold": {
				Type:        "number",
				Description: "Minimum access count to report (default: 5)",
			},
			// Note: Output is JSONL array (one file object per line)
			// Fields: file, total_accesses, read_count, edit_count, write_count, time_span_minutes, first_access, last_access
			// Example jq_filter: "select(.file | test(\"\\.go$\")) | select(.total_accesses > 10)"
		}),
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
		{
			Name:        "list_capabilities",
			Description: "List all available capabilities from configured sources. Returns compact capability index.",
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: map[string]Property{
					// No public parameters
					// Hidden test parameters (_sources, _disable_cache) are not exposed in schema
				},
			},
		},
		{
			Name:        "get_capability",
			Description: "Retrieve complete capability content by name from configured sources.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"name": {
						Type:        "string",
						Description: "Name of the capability to retrieve (without .md extension)",
					},
					// Hidden test parameters (_sources) are not exposed in schema
				},
				Required: []string{"name"},
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
