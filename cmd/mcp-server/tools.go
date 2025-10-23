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
		buildTool("query", "Unified query interface for session data. Default scope: project.", map[string]Property{
			// Tier 1: Resource Selection (required)
			"resource": {
				Type:        "string",
				Description: "Resource type: 'entries' (raw JSONL), 'messages' (user/assistant), 'tools' (tool executions). Default: 'entries'",
			},
			// Tier 3: Filtering
			"filter": {
				Type:        "object",
				Description: "Filter conditions: tool_name, tool_status, role, content_match, session_id, git_branch, time_range, etc.",
			},
			// Tier 4: Transformation
			"transform": {
				Type:        "object",
				Description: "Transform operations: extract (JSONPath), group_by (field name), join (type + on)",
			},
			// Tier 5: Aggregation
			"aggregate": {
				Type:        "object",
				Description: "Aggregation: function ('count'|'sum'|'avg'|'min'|'max'|'group'), field (optional)",
			},
		}),
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
		buildTool("query_tools_advanced", "Query assistant's tools with SQL. Large output, not for user analysis. Default scope: project.", map[string]Property{
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
			// Override jq_filter with schema
			"jq_filter": jqFilterWithSchema(map[string]string{
				"turn_sequence":  "number - Turn sequence number",
				"timestamp":      "string - ISO8601 timestamp",
				"uuid":           "string - Message UUID",
				"model":          "string - Model name",
				"content_blocks": "array - Content blocks (text and tool_use)",
				"text_length":    "number - Text content length",
				"tool_use_count": "number - Number of tool uses",
				"tokens_input":   "number - Input token count",
				"tokens_output":  "number - Output token count",
				"stop_reason":    "string - Stop reason (optional)",
			}, ".[] | select(.tool_use_count > 5)"),
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
			// Override jq_filter with schema
			"jq_filter": jqFilterWithSchema(map[string]string{
				"turn":              "number - Turn sequence number",
				"user_message":      "object - User message data",
				"assistant_message": "object - Assistant message data",
				"duration_ms":       "number - Response duration in milliseconds",
			}, ".[] | select(.duration_ms > 5000)"),
		}),
		buildTool("query_files", "File operation stats (returns array). Use jq_filter for filtering. Default scope: project.", map[string]Property{
			"threshold": {
				Type:        "number",
				Description: "Minimum access count to report (default: 5)",
			},
			// Override jq_filter with schema
			"jq_filter": jqFilterWithSchema(map[string]string{
				"file_path":   "string - File path",
				"read_count":  "number - Number of Read operations",
				"write_count": "number - Number of Write operations",
				"edit_count":  "number - Number of Edit operations",
				"error_count": "number - Number of errors",
				"total_ops":   "number - Total operation count",
				"error_rate":  "number - Error rate (0.0-1.0)",
			}, ".[] | select(.file_path | test(\"\\\\.go$\")) | select(.total_ops > 10)"),
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
