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
		"include_subagents": {
			Type:        "boolean",
			Description: "Include subagent session files (default: true). Pass false to query only top-level sessions.",
		},
		"offset": {
			Type:        "number",
			Description: "Number of records to skip before returning results (for pagination). Default: 0",
		},
		"page_size": {
			Type:        "number",
			Description: "Maximum number of records to return per page (for pagination). Default: 0 = no limit, return all results",
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
			"provider": {
				Type:        "string",
				Description: `Provider filter: "claude" (default) returns the Claude session directory unchanged; "codex" resolves only Codex rollout files (returned as an explicit "files" list plus "directory" when they share one parent); "all" returns a per-provider breakdown ({"providers": {"claude": ..., "codex": ...}}) since Claude and Codex files never share one directory. Unavailable/invalid providers fail with an error instead of silently falling back to Claude.`,
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
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
			Description: "Execute Stage 2 query on selected files with filtering, sorting, and limits. Align transform field paths with inspect_session_files(include_samples=true) samples. Empty results trigger a warning.",
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
			"provider": {
				Type:        "string",
				Description: `Provider filter: "claude" (default) returns the Claude JSONL schema/templates unchanged; "codex" returns Codex's own raw rollout schema and jq templates (never a lossy projection onto the Claude schema); "all" returns a per-provider breakdown ({"providers": {"claude": ..., "codex": ...}}). Unavailable/invalid providers fail with an error instead of silently falling back to Claude.`,
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
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
			"since": {
				Type:        "string",
				Description: `Include only entries with timestamp >= this value (ISO 8601 / RFC3339, e.g. "2026-01-01T00:00:00Z"). Enables full event stream for a focused time range even in large projects.`,
			},
			"until": {
				Type:        "string",
				Description: `Include only entries with timestamp < this value (ISO 8601 / RFC3339, e.g. "2026-06-01T00:00:00Z").`,
			},
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
		}),
		BuildTool("get_tech_debt", "Detect TODO/FIXME/HACK/XXX markers and unresolved errors as tech debt. Default scope: project.", map[string]Property{
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
			"source_dir": {
				Type:        "string",
				Description: "Optional path to source code directory to scan for TODO/FIXME/HACK/XXX markers on disk. Results merged with session-transcript markers.",
			},
		}),
		// ─── New consolidated query tools (replacing the 10 legacy query_* tools) ───

		BuildTool("query_session_content",
			"Query session messages by role (user/assistant/tool/all). Default scope: project. role=tool outputs: {timestamp, session_id, turn, ...block fields}. For custom transform, first call inspect_session_files(include_samples=true).",
			map[string]Property{
				"role": {
					Type:        "string",
					Description: "Message role to query: 'user' (user messages), 'assistant' (assistant messages), 'tool' (tool use/result blocks), or 'all' (user+assistant conversation flow)",
				},
				"contains": {
					Type:        "string",
					Description: "Optional substring filter applied to message content (case-insensitive). When role=assistant, use '## Summary' to retrieve summaries.",
				},
				"pattern": {
					Type:        "string",
					Description: "Regex pattern to match in message content (only applies when role=user). If omitted with role=user, matches all messages.",
				},
				"block_type": {
					Type:        "string",
					Description: "When role=tool: 'tool_use' or 'tool_result' (default: 'tool_use'). Each result includes outer context fields: timestamp, sessionId, turn.",
				},
				"tool_name": {
					Type:        "string",
					Description: "When role=tool and block_type=tool_use: filter by tool name substring or regex (e.g. 'Dispatch' or 'Read|Write'). Omit to return all.",
				},
				"content_type": {
					Type:        "string",
					Description: "When role=user: 'string' (default) or 'array' (tool results)",
				},
				"exclude_system_messages": {
					Type:        "boolean",
					Description: "When role=user: exclude Claude Code system-injected messages. Default: false.",
				},
				"max_message_length": {
					Type:        "number",
					Description: "When role=user: max chars per message content (default: 0 = no truncation). Truncates content, does not filter.",
				},
				"min_content_length": {
					Type:        "number",
					Description: "When role=user: min content length. Only messages with content at least this long are returned.",
				},
				"max_content_length": {
					Type:        "number",
					Description: "When role=user: max content length. Filters out messages longer than this (unlike max_message_length which truncates).",
				},
				"content_summary": {
					Type:        "boolean",
					Description: "Return only session_id/turn/timestamp/preview (100 chars), skip full content.",
				},
				"preview_length": {
					Type:        "integer",
					Description: "Max characters per content_preview when content_summary=true. Default: 100.",
				},
				"group_by_session": {
					Type:        "boolean",
					Description: "When role=user: group results by session. Mutually exclusive with stats_only.",
				},
				"stats_level": {
					Type:        "string",
					Description: "Aggregation level for stats_only/stats_first: 'turn' (default) or 'session'.",
				},
				"context_turns": {
					Type:        "integer",
					Description: "When role=user: number of turns to include before/after each matched turn. Default: 0.",
				},
				"exclude_compact_summaries": {
					Type:        "boolean",
					Description: "Exclude compact summary messages (isCompactSummary=true) from results and context_turns. Default: true. Pass false to include them (e.g. to search the summaries themselves).",
				},
				"since": {
					Type:        "string",
					Description: `Include only records with timestamp >= this value (RFC3339, e.g. "2026-03-07T00:00:00Z")`,
				},
				"until": {
					Type:        "string",
					Description: `Include only records with timestamp < this value (RFC3339)`,
				},
				"limit": {
					Type:        "number",
					Description: "Max results (no limit by default, rely on hybrid output mode)",
				},
				"working_dir": {
					Type:        "string",
					Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
				},
			}, "role"),

		BuildTool("query_session_signals",
			"Query session signals: errors/tokens/system_errors/timestamps/tool_stats. Default scope: project. Time filtering: use since/until params directly—no need for jq_filter time conditions.",
			map[string]Property{
				"type": {
					Type:        "string",
					Description: "Signal type: 'errors' (tool execution errors), 'tokens' (assistant token usage), 'system_errors' (API errors), 'timestamps' (all timestamped entries), or 'tool_stats' (assistant tool calls)",
				},
				"tool": {
					Type:        "string",
					Description: "When type=tool_stats: filter by tool name",
				},
				"status": {
					Type:        "string",
					Description: "When type=tool_stats: filter by status (error/success)",
				},
				"since": {
					Type:        "string",
					Description: `Include only records with timestamp >= this value (RFC3339, e.g. "2026-03-07T00:00:00Z")`,
				},
				"until": {
					Type:        "string",
					Description: `Include only records with timestamp < this value (RFC3339)`,
				},
				"limit": {
					Type:        "number",
					Description: "Max results (no limit by default, rely on hybrid output mode)",
				},
				"working_dir": {
					Type:        "string",
					Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
				},
			}, "type"),

		BuildTool("query_file_activity",
			"Query file history and activity (type=snapshots). Default scope: project.",
			map[string]Property{
				"type": {
					Type:        "string",
					Description: "Activity type: 'snapshots' (file history snapshots with messageId)",
				},
				"limit": {
					Type:        "number",
					Description: "Max results (no limit by default, rely on hybrid output mode)",
				},
				"working_dir": {
					Type:        "string",
					Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
				},
			}, "type"),

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
