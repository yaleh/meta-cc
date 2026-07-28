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
			Type: "string",
			Description: "jq expression applied as a post-filter over this tool's own already-produced result set (i.e. AFTER its built-in role/type/pattern semantics, such as role=assistant or type=tool_stats, have already selected records). " +
				"Defaults to '.[]' when omitted, which is a no-op (returns the result set unchanged). IMPORTANT: Do NOT wrap in quotes - use a raw jq expression that operates on the full result array, e.g.: .[] | select(.status == \"error\") or .[] | {field: .field}",
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

// SessionIDProperty is the "session_id" property shared by tools that
// support an exact-thread fast path (DIR-030): unlike scope="session"
// (which means "the most recently modified session"), session_id selects
// one specific session/thread by ID and reads only that session — see
// docs/guides/mcp-query-tools.md's "scope vs session_id" section for the
// full distinction. Declared once so every tool that wires it through
// (see internal/mcp/executor's dispatchProviderQuery/loadData callers)
// documents it identically.
func SessionIDProperty() Property {
	return Property{
		Type: "string",
		Description: `Exact session/thread ID to query, reading only that one session (never "most recent" and never every session in scope). ` +
			`Distinct from scope="session" (which means "the most recently modified session" and is unrelated to a specific ID). ` +
			`When set, "scope" and "cwd"/project-boundary filters still apply to the fetched session's metadata (e.g. a session outside the current project is excluded), but no other session is listed or loaded.`,
	}
}

// OutputFormatProperty is the "output_format" property shared by the four
// consolidated query tools (query_sessions, query_session_content,
// query_session_signals, query_file_activity) whose results are produced via
// internal/mcp/pipeline.BuildResponse. DIR-044: this was formerly declared on
// all 16 MCP tools via StandardToolParameters(), but pipeline.BuildResponse
// never branched on it, so every tool silently ignored it. It's now (a)
// genuinely implemented as real TSV serialization here, in
// pipeline.BuildResponse (see internal/mcp/response's SerializeResponseTSV /
// RecordsToTSV), and (b) scoped to only the tools whose data actually flows
// through that pipeline — the other 12 tools (analyze_errors, analyze_bugs,
// quality_scan, get_work_patterns, get_timeline, get_tech_debt,
// query_edit_sequences, cleanup_temp_files, get_session_directory,
// inspect_session_files, execute_stage2_query, get_session_metadata) marshal
// their own typed, non-flat analyzer-result structs directly and never
// touch PipelineConfig, so output_format is no longer advertised on them
// rather than left declared-but-inert.
func OutputFormatProperty() Property {
	return Property{
		Type: "string",
		Description: "Output format: \"jsonl\" (default) or \"tsv\". TSV output (inline-mode responses only) is a header row " +
			"(the sorted union of field names across all returned records) followed by one tab-separated row per record; " +
			"nested object/array field values are JSON-encoded inline within their own cell rather than flattened, and any " +
			"literal tab/newline/backslash inside a scalar string value is backslash-escaped. Envelope metadata (mode, " +
			"pagination, warnings) is emitted as leading \"#\"-prefixed comment lines before the header. file_ref-mode " +
			"responses (large result sets written to a temp file) are unaffected by this setting and remain JSON.",
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
			"session_id": SessionIDProperty(),
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
			"session_id": SessionIDProperty(),
		}),
		BuildTool("quality_scan", "Compute quality dimensions: error rate, retry rate, diversity, completion. Default scope: project.", map[string]Property{
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
			"session_id": SessionIDProperty(),
		}),
		BuildTool("get_work_patterns", "Get tool frequency, hourly activity, and context switches. Default scope: project.", map[string]Property{
			"working_dir": {
				Type:        "string",
				Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
			},
			"session_id": SessionIDProperty(),
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
			"session_id": SessionIDProperty(),
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
			"session_id": SessionIDProperty(),
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
					Type: "integer",
					Description: "When role=user: number of adjacent records to include before/after each matched record, bounded at session start/end and deduplicated across overlapping windows. Default: 0. " +
						"Supported for every provider (claude, codex, all): a match returned with context_turns=0 is never erased by setting context_turns>0. Ordering/counting is by each record's position in that session's normalized record stream (chronological order), not by turn count or timestamp equality — multiple records can share one timestamp. Matched records are marked context:false; added records are marked context:true. " +
						"For codex/all, context is loaded through the same provider/session backend (rollout or app-server) the query itself used; if that reload fails for a given session, its original matches are still returned (context:false, no added context) and a warning is included rather than silently returning no data.",
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
				"session_id":    SessionIDProperty(),
				"output_format": OutputFormatProperty(),
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
				"session_id":    SessionIDProperty(),
				"output_format": OutputFormatProperty(),
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
				"session_id":    SessionIDProperty(),
				"output_format": OutputFormatProperty(),
			}, "type"),

		BuildTool("query_sessions",
			"List session/thread metadata without loading turn content. Default scope: project.",
			map[string]Property{
				"cwd": {
					Type:        "string",
					Description: `Exact cwd to scope the listing to. Defaults to the project boundary derived from working_dir (or MCP server CWD) — sessions outside that cwd are never returned, matching every other tool's project-scope behavior.`,
				},
				"session_id": SessionIDProperty(),
				"archived": {
					Type:        "boolean",
					Description: `Codex only (provider="codex" or "all"). Filter by archived state: true = only archived threads, false = only active threads. Archived threads are excluded by default (omitting both this and status is equivalent to archived=false) — pass archived=true or status="archived" to discover them.`,
				},
				"status": {
					Type:        "string",
					Description: `Codex only. "active" or "archived" — an alias for the archived filter expressed as a status value. Conflicting status/archived values fail with an error. Like archived, omitting this defaults to active-only.`,
				},
				"ancestors_of": {
					Type:        "string",
					Description: `Codex only. Given a thread ID, return its ancestor chain (parent, grandparent, ...) instead of the normal listing, each entry annotated with "lineage" ("root", "child", or "unknown" when spawn metadata wasn't available). Traversal stops — and the last entry's "lineage_truncated" is set — if an ancestor lookup crosses the project/cwd boundary or a bounded depth limit is reached, rather than silently continuing into another project's data.`,
				},
				"source_kind": {
					Type:        "array",
					Description: `Codex only. Filter by one or more source kinds: "cli", "vscode", "exec", "appServer", "subAgent", "subAgentReview", "subAgentCompact", "subAgentThreadSpawn", "subAgentOther", "unknown". An unrecognized value is a validation error, not a silently-empty result.`,
					Items:       &Property{Type: "string"},
				},
				"model_provider": {
					Type:        "string",
					Description: `Codex only. Filter by model provider (e.g. "openai").`,
				},
				"model": {
					Type:        "string",
					Description: "Filter by model name substring (case-insensitive).",
				},
				"parent_thread_id": {
					Type:        "string",
					Description: "Codex only. Filter by exact parent/ancestor thread ID (e.g. to list a subagent thread's lineage).",
				},
				"title_contains": {
					Type:        "string",
					Description: "Filter by a case-insensitive substring match against the session title.",
				},
				"created_since": {
					Type:        "string",
					Description: `Include only sessions created at/after this value (RFC3339, e.g. "2026-03-07T00:00:00Z").`,
				},
				"created_until": {
					Type:        "string",
					Description: `Include only sessions created before this value (RFC3339).`,
				},
				"updated_since": {
					Type:        "string",
					Description: `Include only sessions updated at/after this value (RFC3339). Codex only — Claude sessions don't track a separate updated_at.`,
				},
				"updated_until": {
					Type:        "string",
					Description: `Include only sessions updated before this value (RFC3339). Codex only.`,
				},
				"limit": {
					Type:        "number",
					Description: "Max sessions to return (0 = unlimited). Combine with the standard offset/page_size parameters for cursor-style pagination over a large project.",
				},
				"working_dir": {
					Type:        "string",
					Description: "Override working directory for session lookup. Defaults to MCP server CWD.",
				},
				"output_format": OutputFormatProperty(),
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
