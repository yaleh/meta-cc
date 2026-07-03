# MCP Server Guide

meta-cc provides a Model Context Protocol (MCP) server for local coding-agent history analysis. It supports Claude Code and Codex through a provider-aware conversation layer, then exposes normalized records to convenience query tools, two-stage jq queries, and higher-level analysis tools.

## Provider Support

| Provider | Local source | Notes |
|----------|--------------|-------|
| `claude` | `~/.claude/projects/<project-hash>/*.jsonl` | Default provider for backward compatibility. |
| `codex` | `${META_CC_CODEX_ROOT:-~/.codex}/state_5.sqlite` plus rollout JSONL paths from the `threads` table | `~/.codex/history.jsonl` is intentionally excluded. |
| `all` | Claude Code and Codex | Returned records include a `provider` field. |

Convenience query and analysis tools accept the standard `provider` parameter:

```javascript
query_user_messages({
  provider: "codex",
  pattern: "migration",
  limit: 20
})

get_work_patterns({
  provider: "all",
  working_dir: "/path/to/project"
})
```

If `provider` is omitted, meta-cc uses `claude` to preserve existing behavior.

## Configuration

### Claude Code

The release archive includes `.mcp.json`; the installer merges it into `~/.claude/mcp.json`.

Manual configuration:

```json
{
  "mcpServers": {
    "meta-cc": {
      "command": "meta-cc-mcp",
      "args": []
    }
  }
}
```

### Codex

The release archive includes `.codex-plugin/plugin.json` and `.codex-mcp.json`. The installer copies them under `~/.codex/plugins/meta-cc/`.

Use `CODEX_HOME` for installation targets and `META_CC_CODEX_ROOT` when you need the MCP provider to read a non-default Codex state directory:

```bash
CODEX_HOME=/tmp/codex ./install.sh
META_CC_CODEX_ROOT=/tmp/codex meta-cc-mcp
```

## Tool Catalog

### Convenience Queries

| Tool | Purpose | Claude Code | Codex |
|------|---------|-------------|-------|
| `query_user_messages` | Search user messages by regex | Yes | Yes |
| `query_tools` | Query assistant tool calls | Yes | Yes |
| `query_tool_errors` | Query failed tool results | Yes | Yes |
| `query_token_usage` | Query assistant token usage | Yes | Yes |
| `query_conversation_flow` | Query user/assistant flow | Yes | Yes |
| `query_tool_blocks` | Query `tool_use` or `tool_result` blocks | Yes | Yes |
| `query_timestamps` | Query timestamped records | Yes | Yes |
| `query_system_errors` | Query Claude Code API system errors | Yes | Host-specific empty |
| `query_file_snapshots` | Query Claude Code file snapshots | Yes | Host-specific empty |
| `query_summaries` | Query Claude Code summaries | Yes | Host-specific empty |

Examples:

```javascript
query_tools({
  provider: "codex",
  tool: "exec_command",
  working_dir: "/path/to/project",
  limit: 50
})

query_token_usage({
  provider: "codex",
  stats_first: true,
  limit: 20
})
```

### Analysis Tools

| Tool | Purpose | Claude Code | Codex |
|------|---------|-------------|-------|
| `analyze_errors` | Aggregate tool errors by tool and type | Yes | Yes |
| `analyze_bugs` | Detect error-fix pairs and recurring bug patterns | Yes | Yes |
| `quality_scan` | Compute quality dimensions | Yes | Yes |
| `get_work_patterns` | Tool frequency, hourly activity, and context switches | Yes | Yes |
| `get_timeline` | Chronological session events | Yes | Yes |
| `get_tech_debt` | TODO/FIXME/HACK markers and unresolved errors | Yes | Yes |

All six analysis tools include a `data_source` field in their response (see
[Data Source Provenance](#data-source-provenance-baime-layer-7) below).

| Tool | `data_source` value | Notes on mixed provenance |
|------|---------------------|---------------------------|
| `analyze_errors` | `measured` | TotalErrors/ByTool/ByType are direct counts |
| `analyze_bugs` | `measured` | Patterns/TotalPairs are direct error→success pair counts |
| `quality_scan` | `measured` | All four dimensions are ratios of direct counts |
| `get_work_patterns` | `measured` | ToolFrequency/HourlyActivity/PeakHour are direct counts; `context_switches` uses a heuristic (file changes within 5 min) |
| `get_timeline` | `measured` | Events parsed directly from session entries |
| `get_tech_debt` | `measured` | Markers/HotspotFiles scanned from tool output; `open_issues` uses a heuristic (error with no subsequent success) |

### Two-Stage Query Tools

Use these when you need file selection control or custom jq over selected JSONL files:

| Tool | Purpose |
|------|---------|
| `get_session_directory` | Locate a transcript directory and aggregate file metadata |
| `inspect_session_files` | Inspect selected JSONL files for record counts, time ranges, and samples |
| `execute_stage2_query` | Run jq-style filter/sort/transform on selected files |
| `get_session_metadata` | Return schema hints, file info, and query templates |

Two-stage tools operate on selected files. They retain raw-file compatibility, including normalized Codex JSONL records when a Codex rollout/session file is selected directly. Provider-aware cross-host querying is handled by the convenience query and analysis tools through the `provider` parameter.

## Codex Normalization

The Codex provider reads session metadata from `state_5.sqlite` and follows each thread's `rollout_path`. It normalizes:

- legacy `response_item` messages with `input_text` / `output_text`
- `function_call` and `function_call_output`
- `custom_tool_call` and `custom_tool_call_output`
- `event_msg` `token_count` usage
- newer dotted schema events such as `turn.started`, `item.message`, `item.tool_call`, and `item.tool_result`

Codex `tokens_used` from SQLite is retained as session metadata, but `query_token_usage(provider: "codex")` reports per-turn usage only when the rollout contains a `token_count` event.

## Standard Parameters

Most query and analysis tools accept:

| Parameter | Type | Description |
|-----------|------|-------------|
| `scope` | string | `project` (default) or `session` |
| `provider` | string | `claude` (default), `codex`, or `all` |
| `working_dir` | string | Override project path used for session lookup |
| `limit` | number | Maximum results; default is no limit |
| `stats_only` | boolean | Return aggregate statistics only |
| `stats_first` | boolean | Return stats followed by details |
| `output_format` | string | `jsonl` or `tsv` |
| `inline_threshold_bytes` | number | Threshold for inline vs file reference output |

Time-aware tools also accept RFC3339 `since` and `until`.

### `get_timeline` and `stats_only`

`get_timeline` has special behaviour for `stats_only` that differs from other tools.
Without it, a project with many sessions can return hundreds of kilobytes of event
data and overflow the context window.

**`stats_only: true`** returns a compact summary instead of the full event list:

```json
{
  "total_entries": 9368,
  "time_range": {
    "from": "2026-06-16T14:19:48Z",
    "to":   "2026-06-20T03:25:43Z",
    "span": "85h 5m"
  },
  "event_type_counts": {
    "assistant_message": 5600,
    "user_message": 3768
  }
}
```

Use this as a **scout step** before deciding how to query:

```text
1. get_timeline(scope=project, stats_only=true)   → assess scale and time range
2a. entries are small  → get_timeline(scope=project, limit=50)
2b. entries are large  → get_timeline(scope=session)           # current session only
                       → get_timeline(scope=project, limit=N)  # explicit cap
                       → get_timeline(jq_filter='select .timestamp > "..."')
```

When `limit` truncates the result, the response includes `truncated: true` and
`total_events: N` so you know more data exists.

## Output Modes

The MCP server uses hybrid output:

- Small responses are returned inline.
- Large responses are written to a temporary file and returned as `file_ref`.

## Verification

Ask either host:

```text
Which tools do I use most often?
Find user messages mentioning "refactor"
Show token usage for recent turns
```

For Codex-specific verification, `make test-e2e-codex` creates a temporary Codex home with a real `state_5.sqlite` and rollout JSONL, then calls the MCP server over JSON-RPC with `provider: "codex"`.

## Troubleshooting

### No sessions found

- Check `working_dir` points at the project whose history you want.
- For Claude Code, verify `~/.claude/projects/<project-hash>/` exists.
- For Codex, verify `${META_CC_CODEX_ROOT:-$HOME/.codex}/state_5.sqlite` has a `threads` row whose `cwd` matches the project and whose `rollout_path` points to a readable JSONL file.

### Tool returns empty on Codex

Some tools query Claude Code-only record types:

- `query_file_snapshots`
- `query_summaries`
- `query_system_errors`

Empty results are expected for Codex unless Codex adds equivalent local records.

## Data Source Provenance (BAIME Layer 7)

Every analysis tool response includes a top-level `data_source` field that
classifies how the result was produced. This implements the BAIME Layer 7
provenance standard and enables callers to detect and correct for systematic
self-observation bias (e.g., delta_H inflation observed in BAIME TASK-152).

### Values

| Value | Meaning |
|-------|---------|
| `"measured"` | The value was computed by directly counting or aggregating observable events from the session trace (tool calls, entries, timestamps). High confidence; no inferential leap. |
| `"estimated"` | The value was inferred via a heuristic rule rather than directly observed. Lower confidence; treat as approximate. |

### Mixed-Provenance Structs

Some result structs combine measured and estimated fields. In these cases the
top-level `data_source` reflects the **dominant** provenance (always `measured`
for the current analysis tools), and mixed fields are documented in code
comments:

- **`TechDebtResult.open_issues`** — estimated: based on the heuristic "an
  error call with no subsequent success for the same tool". The rest of the
  struct (Markers, HotspotFiles) is measured.

- **`WorkPatternsResult.context_switches`** — estimated: based on the heuristic
  "file-path changes between consecutive tool calls within a 5-minute window".
  The rest of the struct (ToolFrequency, HourlyActivity, PeakHour) is measured.

When building automated pipelines that depend on precise values, treat
fields with known estimated provenance with appropriate uncertainty margins.

## See Also

- [MCP Query Tools Reference](mcp-query-tools.md)
- [Two-Stage Query Guide](two-stage-query-guide.md)
- [GCL Gate Annotation Format](gcl-annotation.md)
- [JSONL Schema Reference](../reference/jsonl-schema.md)
- [Installation Guide](../tutorials/installation.md)
