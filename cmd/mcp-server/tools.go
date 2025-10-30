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
		// Phase 27 Stage 27.1: query and query_raw tools removed
		// Use the 10 shortcut query tools instead

		// Layer 1: Convenience Tools (10 high-frequency queries)
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
		buildTool("get_session_directory", "Get session directory metadata. Default scope: project.", map[string]Property{
			"scope": {
				Type:        "string",
				Description: "Query scope: 'session' (current session only) or 'project' (all sessions)",
			},
		}, "scope"),
		buildTool("inspect_session_files", "Inspect session files for metadata (record types, time ranges, size).", map[string]Property{
			"files": {
				Type:        "array",
				Description: "Array of absolute file paths to inspect",
				Items: &Property{
					Type: "string",
				},
			},
			"include_samples": {
				Type:        "boolean",
				Description: "If true, include 1-2 sample records per type (default: false)",
			},
		}, "files"),
		{
			Name:        "execute_stage2_query",
			Description: "Execute Stage 2 query on selected files with filtering, sorting, and limits.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"files": {
						Type:        "array",
						Description: "Array of absolute file paths to query (from Stage 1 inspection)",
						Items: &Property{
							Type: "string",
						},
					},
					"filter": {
						Type:        "string",
						Description: "jq filter expression (e.g., 'select(.type == \"user\")'). Required.",
					},
					"sort": {
						Type:        "string",
						Description: "jq sort expression (e.g., 'sort_by(.timestamp)'). Optional.",
					},
					"transform": {
						Type:        "string",
						Description: "jq transform expression (e.g., '{type, timestamp}'). Optional.",
					},
					"limit": {
						Type:        "number",
						Description: "Maximum number of results to return. Optional (default: no limit).",
					},
				},
				Required: []string{"files", "filter"},
			},
		},
		buildTool("get_session_metadata", "Get session metadata including JSONL schema, file info, and query templates. Default scope: project.", map[string]Property{
			"scope": {
				Type:        "string",
				Description: "Query scope: 'project' (default) or 'session'",
			},
		}),
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
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Items       *Property `json:"items,omitempty"` // For array types
}
