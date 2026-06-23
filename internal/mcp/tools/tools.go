package tools

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yaleh/meta-cc/internal/mcp/schema"
)

// Type aliases so callers don't need to import schema directly.
type Tool = schema.Tool
type ToolSchema = schema.ToolSchema
type Property = schema.Property

// StandardToolParameters returns the standard set of parameters for all MCP tools
func StandardToolParameters() map[string]Property {
	return map[string]Property{
		"scope": {
			Type:        "string",
			Description: "Query scope: 'project' (default) or 'session'",
		},
		"provider": {
			Type:        "string",
			Description: `Provider filter: "claude" (default), "codex", or "all". Use "all" only when your jq filter is compatible with both providers.`,
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
			Description: "Threshold for inline vs file_ref mode in bytes (default: 32768). Can also set META_CC_INLINE_THRESHOLD env var",
		},
		"output_format": {
			Type:        "string",
			Description: "Output format: jsonl or tsv (default: jsonl)",
		},
	}
}

// JqFilterWithSchema creates a jq_filter property with output schema documentation
func JqFilterWithSchema(fields map[string]string, example string) Property {
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

// BuildToolSchema creates a ToolSchema with merged parameters
func BuildToolSchema(properties map[string]Property, required ...string) ToolSchema {
	s := ToolSchema{
		Type:       "object",
		Properties: MergeParameters(properties),
	}
	if len(required) > 0 {
		s.Required = required
	}
	return s
}

// BuildTool creates a Tool with the given name, description, and schema
func BuildTool(name, description string, properties map[string]Property, required ...string) Tool {
	return Tool{
		Name:        name,
		Description: description,
		InputSchema: BuildToolSchema(properties, required...),
	}
}

// GetToolDefinitions returns all tool definitions
func GetToolDefinitions() []Tool {
	return []Tool{
		// Phase 27 Stage 27.1: query and query_raw tools removed
		// Use the 10 shortcut query tools instead

		// Layer 1: Convenience Tools (10 high-frequency queries)
		// Note: query_user_messages and query_tools already exist above
		BuildTool("query_tool_errors", "Query tool execution errors. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("query_token_usage", "Query assistant messages with token usage stats. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("query_conversation_flow", "Query user and assistant conversation flow. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
			"transform": {
				Type:        "string",
				Description: "Optional jq transform for parent-child relationships",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
			"since": {
				Type:        "string",
				Description: `Include only records with timestamp >= this value (RFC3339, e.g. "2026-03-07T00:00:00Z")`,
			},
			"until": {
				Type:        "string",
				Description: `Include only records with timestamp < this value (RFC3339)`,
			},
		}),
		BuildTool("query_system_errors", "Query system API errors. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("query_file_snapshots", "Query file history snapshots. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("query_timestamps", "Query all entries with timestamps. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
			"since": {
				Type:        "string",
				Description: `Include only records with timestamp >= this value (RFC3339, e.g. "2026-03-07T00:00:00Z")`,
			},
			"until": {
				Type:        "string",
				Description: `Include only records with timestamp < this value (RFC3339)`,
			},
		}),
		BuildTool("query_summaries", "Query session summaries. Default scope: project.", map[string]Property{
			"keyword": {
				Type:        "string",
				Description: "Keyword to search in summary (case-insensitive)",
			},
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("query_tool_blocks", "Query tool use or tool result blocks. Default scope: project.", map[string]Property{
			"block_type": {
				Type:        "string",
				Description: "Block type: 'tool_use' or 'tool_result' (required)",
			},
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}, "block_type"),

		BuildTool("query_tools", "Query assistant's internal tool calls. Large output, not for user analysis. Default scope: project.", map[string]Property{
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
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
			// Override jq_filter with schema (snake_case fields)
			"jq_filter": JqFilterWithSchema(map[string]string{
				"tool_name": "string - Tool identifier (e.g., \"Bash\", \"Read\", \"mcp__meta-cc__query_tools\")",
				"status":    "string - Execution status (\"success\" or \"error\")",
				"timestamp": "string - ISO8601 timestamp",
				"error":     "string - Error message if status is \"error\"",
				"input":     "object - Tool input parameters",
				"output":    "object - Tool output/result",
				"uuid":      "string - Unique call identifier",
			}, ".[] | select(.tool_name == \"Bash\" and .status == \"error\")"),
		}),
		BuildTool("query_user_messages", "Search user messages with regex. May contain large outputs. Default scope: project.", map[string]Property{
			// Tier 1: Required
			"pattern": {
				Type:        "string",
				Description: "Regex pattern to match (required)",
			},
			// Tier 2: Content type
			"content_type": {
				Type:        "string",
				Description: "Content type filter: 'string' (default) or 'array' (tool results)",
			},
			// Tier 3: Range / Length filtering
			"max_message_length": {
				Type:        "number",
				Description: "Max chars per message content (default: 0 = no truncation, rely on hybrid mode for large results). Truncates content, does not filter.",
			},
			"min_content_length": {
				Type:        "number",
				Description: "Minimum content length in characters. Only messages with content at least this long are returned. Only applies to string content type.",
			},
			"max_content_length": {
				Type:        "number",
				Description: "Maximum content length in characters. Only messages with content at most this long are returned. Only applies to string content type. Unlike max_message_length which truncates, this filters out messages entirely.",
			},
			// Tier 5: Cross-project
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
			// Tier 4: Output Control
			"limit": {
				Type:        "number",
				Description: "Max results (no limit by default, rely on hybrid output mode)",
			},
			"content_summary": {
				Type:        "boolean",
				Description: "Return only session_id/turn/timestamp/preview (100 chars), skip full content. Use hybrid mode instead for better information preservation.",
			},
			"preview_length": {
				Type:        "integer",
				Description: "Max characters (runes) per content_preview when content_summary=true. Default: 100. For CJK content, set to 30 for ~30 readable characters (CJK chars are 3 bytes each in UTF-8).",
			},
			// Tier 3: Time range filtering
			"since": {
				Type:        "string",
				Description: `Include only records with timestamp >= this value (RFC3339, e.g. "2026-03-07T00:00:00Z")`,
			},
			"until": {
				Type:        "string",
				Description: `Include only records with timestamp < this value (RFC3339)`,
			},
			"exclude_system_messages": {
				Type:        "boolean",
				Description: `If true, exclude Claude Code system-injected messages whose content starts with <local-command-caveat>, <command-name>, <local-command-stdout>, or <task-notification>. Only applies to string content type. Default: false.`,
			},
			"group_by_session": {
				Type:        "boolean",
				Description: "Group results by session. Returns one object per session with session_id, match_count, first_match, last_match, and turns array. Mutually exclusive with stats_only.",
			},
			"stats_level": {
				Type:        "string",
				Description: "Aggregation level for stats_only/stats_first: 'turn' (default, hourly buckets) or 'session' (per-session match_count and duration).",
			},
			"context_turns": {
				Type:        "integer",
				Description: "Number of turns to include before and after each matched turn (same session). Context turns are marked with 'context': true; matched turns with 'context': false. Default: 0 (disabled). Only applies to string content_type.",
			},
			// Override jq_filter with schema
			"jq_filter": JqFilterWithSchema(map[string]string{
				"sessionId": "string - Session identifier",
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
		BuildTool("get_session_directory", "Get session directory metadata. Default scope: project.", map[string]Property{
			"scope": {
				Type:        "string",
				Description: "Query scope: 'session' (current session only) or 'project' (all sessions)",
			},
		}, "scope"),
		BuildTool("inspect_session_files", "Inspect session files for metadata (record types, time ranges, size).", map[string]Property{
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
		BuildTool("analyze_errors", "Aggregate tool errors by tool name and error type. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max examples per group (0 = unlimited)",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("analyze_bugs", "Detect error-fix pairs and recurring bug patterns. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max examples per pattern (0 = unlimited)",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("quality_scan", "Compute quality dimensions: error rate, retry rate, diversity, completion. Default scope: project.", map[string]Property{
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("get_work_patterns", "Get tool frequency, hourly activity, and context switches. Default scope: project.", map[string]Property{
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("get_session_metadata", "Get session metadata including JSONL schema, file info, and query templates. Default scope: project.", map[string]Property{
			"scope": {
				Type:        "string",
				Description: "Query scope: 'project' (default) or 'session'",
			},
		}),
		BuildTool("get_timeline", "Get chronological session events as JSON. Claude renders visualization. Default scope: project.", map[string]Property{
			"limit": {
				Type:        "number",
				Description: "Max events to return (0 = unlimited)",
			},
			"stats_only": {
				Type:        "boolean",
				Description: "Return summary statistics (total entries, time range, event type counts) instead of the full event list. Safe for large project scopes.",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("get_tech_debt", "Detect TODO/FIXME/HACK markers and unresolved errors as tech debt signals. Default scope: project.", map[string]Property{
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("query_edit_sequences", "Analyze file edit/read patterns: docRole, co-accessed docs, DocVoid. Default scope: project.", map[string]Property{
			"files": {
				Type:        "array",
				Description: "Array of absolute file paths to analyze (required)",
				Items: &Property{
					Type: "string",
				},
			},
			"include_content": {
				Type:        "boolean",
				Description: "If true, include full old/new string content in edit events (default: false)",
			},
			"limit_per_file": {
				Type:        "number",
				Description: "Maximum events to return per file (default: 50)",
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}, "files"),
	}
}

// ValidateToolArgs checks that toolName is known and all provided arg keys are declared in its schema.
func ValidateToolArgs(toolName string, args map[string]interface{}) error {
	index := BuildToolSchemaIndex()
	s, err := GetToolSchemaByName(index, toolName)
	if err != nil {
		return err
	}
	return schema.ValidateArgKeys(args, s)
}

// BuildToolSchemaIndex builds the index from tool definitions.
func BuildToolSchemaIndex() map[string]ToolSchema {
	return schema.BuildSchemaIndex(GetToolDefinitions())
}

// GetToolSchemaByName returns the ToolSchema for the named tool, or an error if not found.
func GetToolSchemaByName(index map[string]ToolSchema, name string) (ToolSchema, error) {
	return schema.GetByName(index, name)
}
