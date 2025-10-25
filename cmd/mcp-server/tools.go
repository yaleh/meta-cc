package main

import (
	"fmt"
	"sort"
	"strings"
)

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
			Description: "jq expression for filtering. Defaults to '.[]' when omitted. IMPORTANT: Do NOT wrap in quotes - use raw jq expression like: .[] | {field: .field}",
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

// jqFilterWithSchema creates a jq_filter property with output schema documentation
func jqFilterWithSchema(fields map[string]string, example string) Property {
	var fieldDocs []string
	for field, desc := range fields {
		fieldDocs = append(fieldDocs, fmt.Sprintf("    %s: %s", field, desc))
	}
	sort.Strings(fieldDocs)

	desc := fmt.Sprintf(`jq expression for filtering. Defaults to '.[]' when omitted. Do NOT wrap in quotes.

Output schema:
%s

Example: %s`, strings.Join(fieldDocs, "\n"), example)

	return Property{
		Type:        "string",
		Description: desc,
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
		// Phase 25: query tool uses jq-based interface (jq_filter, jq_transform)
		// All standard parameters (scope, jq_filter, stats_only, etc.) are provided by StandardToolParameters()
		// No tool-specific parameters needed beyond the standard ones
		buildTool("query", "Execute jq query on session data. Default scope: project.", map[string]Property{}),
		buildTool("query_raw", "Execute raw jq expression. For power users. Default scope: project.", map[string]Property{
			"jq_expression": {
				Type:        "string",
				Description: "Complete jq expression (required). Maximum flexibility.",
			},
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}, "jq_expression"),

		// Layer 1: NEW Convenience Tools (8 high-frequency queries)
		// Note: query_user_messages and query_tools already exist above
		buildTool("query_tool_errors", "Query tool execution errors. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}),
		buildTool("query_token_usage", "Query assistant messages with token usage stats. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}),
		buildTool("query_conversation_flow", "Query user and assistant conversation flow. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
			"transform": {
				Type:        "string",
				Description: "Optional jq transform for parent-child relationships",
			},
		}),
		buildTool("query_system_errors", "Query system API errors. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}),
		buildTool("query_file_snapshots", "Query file history snapshots. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}),
		buildTool("query_timestamps", "Query all entries with timestamps. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}),
		buildTool("query_summaries", "Query session summaries. Default scope: project.", map[string]Property{
			"keyword": {
				Type:        "string",
				Description: "Keyword to search in summary (case-insensitive)",
			},
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}),
		buildTool("query_tool_blocks", "Query tool use or tool result blocks. Default scope: project.", map[string]Property{
			"block_type": {
				Type:        "string",
				Description: "Block type: 'tool_use' or 'tool_result' (required)",
			},
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
		}, "block_type"),

		buildTool("get_session_stats", "Get session statistics. Default scope: session.", map[string]Property{}),
		buildTool("query_tools", "Query assistant's internal tool calls. Large output, not for user analysis. Default scope: project.", map[string]Property{
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
			// Override jq_filter with schema (snake_case fields)
			"jq_filter": jqFilterWithSchema(map[string]string{
				"tool_name": "string - Tool identifier (e.g., \"Bash\", \"Read\", \"mcp__meta-cc__query_tools\")",
				"status":    "string - Execution status (\"success\" or \"error\")",
				"timestamp": "string - ISO8601 timestamp",
				"error":     "string - Error message if status is \"error\"",
				"input":     "object - Tool input parameters",
				"output":    "object - Tool output/result",
				"uuid":      "string - Unique call identifier",
			}, ".[] | select(.tool_name == \"Bash\" and .status == \"error\")"),
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
			// Override jq_filter with schema
			"jq_filter": jqFilterWithSchema(map[string]string{
				"turn":      "number - Turn sequence number",
				"timestamp": "string - ISO8601 timestamp",
				"content":   "string - User message content",
			}, ".[] | select(.content | test(\"error|bug\"; \"i\"))"),
		}, "pattern"),
		buildTool("query_tool_sequences", "Query assistant's tool sequences. Large output, not for user analysis. Default scope: project.", map[string]Property{
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
			// Override jq_filter with schema
			"jq_filter": jqFilterWithSchema(map[string]string{
				"pattern":              "string - Tool sequence pattern",
				"count":                "number - Number of times pattern appeared",
				"occurrences":          "array - Individual occurrence details (each with start_turn, end_turn)",
				"time_span_minutes":    "number - Time span in minutes",
				"length":               "number - Number of tools in pattern (optional)",
				"success_rate":         "number - Success rate of sequence (optional)",
				"avg_duration_minutes": "number - Average duration in minutes (optional)",
			}, ".[] | select(.count > 5)"),
		}),
		buildTool("query_file_access", "Query file operation history. Default scope: project.", map[string]Property{
			"file": {
				Type:        "string",
				Description: "File path (required)",
			},
			// Override jq_filter with schema
			"jq_filter": jqFilterWithSchema(map[string]string{
				"file":              "string - File path",
				"total_accesses":    "number - Total access count",
				"operations":        "object - Operation counts by type (Read/Edit/Write)",
				"timeline":          "array - Chronological access events",
				"time_span_minutes": "number - Time span in minutes",
			}, "select(.total_accesses > 10)"),
		}, "file"),
		buildTool("query_project_state", "Query project state evolution. Default scope: project.", map[string]Property{
			// Override jq_filter with schema
			"jq_filter": jqFilterWithSchema(map[string]string{
				"timestamp": "string - ISO8601 timestamp",
				"type":      "string - State type (session_state, etc.)",
			}, ".[] | select(.type == \"session_state\")"),
		}),
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
			// Override jq_filter with schema
			"jq_filter": jqFilterWithSchema(map[string]string{
				"turn":          "number - Turn sequence number",
				"content":       "string - Prompt content",
				"quality_score": "number - Quality score (0.0-1.0)",
			}, ".[] | select(.quality_score > 0.9)"),
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
