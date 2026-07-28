# MCP Query Tools Reference

meta-cc exposes MCP tools for Claude Code and Codex session analysis. Claude Code transcripts are read from `~/.claude/projects/`; Codex history is read through the provider layer from Codex local state and rollout JSONL files. Query results are normalized into a shared message/tool schema before filtering and response rendering.

## Provider Support

Convenience query tools accept a standard `provider` parameter:

| Provider | Meaning |
|----------|---------|
| `claude` | Query Claude Code sessions. This is the default for backward compatibility. |
| `codex` | Query Codex local history and rollout JSONL files. |
| `all` | Query both providers and include a `provider` field on returned records. |

Use `provider: "all"` only when your filters can handle records from both providers.

```javascript
query_session_signals({
  type: "tool_stats",
  provider: "codex",
  limit: 20
})

query_session_content({
  role: "user",
  provider: "all",
  pattern: "refactor",
  limit: 20
})
```

## `jq_filter` (custom post-filter, DIR-041)

`query_session_content`, `query_session_signals`, `query_sessions`, and `query_file_activity`
all accept an optional `jq_filter` parameter. It is applied as a **second, caller-controlled
filtering stage on top of** each tool's own built-in semantics (e.g. `role`/`pattern` on
`query_session_content`, `type` on `query_session_signals`/`query_file_activity`, or the
Go-native session filters on `query_sessions`) — not instead of them. The tool's own
parameters narrow the query first; `jq_filter` then narrows (or reshapes) whatever that
query already produced.

The expression runs against the **full result array as one JSON value** — exactly like
piping that array into `jq` on the command line — so the default, `.[]`, unpacks it back
into individual records (a no-op). Write filters accordingly, e.g. `.[] | select(.status ==
"error")` or `.[] | {tool: .tool_name}`, not a bare `select(...)` intended to run per-record.
Do not wrap the expression in quotes.

```javascript
// Only tool_stats records with status == "error"
query_session_signals({
  type: "tool_stats",
  jq_filter: '.[] | select(.status == "error")'
})
```

For arbitrary custom jq over the raw session JSONL files themselves (rather than one of
these four tools' already-produced, semantically-filtered result), use
`execute_stage2_query` instead — see the
[Two-Stage Query Guide](two-stage-query-guide.md).

## Host Storage

| Host | Session source | Notes |
|------|----------------|-------|
| Claude Code | `~/.claude/projects/<project-hash>/` | Native Claude Code JSONL schema |
| Codex | `~/.codex/state_5.sqlite` plus rollout JSONL files | Project sessions are filtered by `cwd`; `~/.codex/history.jsonl` is intentionally excluded |

Codex rollout normalization covers:

- user and assistant messages
- function/custom tool calls
- function/custom tool outputs
- token usage exposed through `query_session_signals(type="tokens")`

Claude-specific records without Codex equivalents remain host-specific: `file-history-snapshot`, top-level `summary`, and `system` records with `subtype: "api_error"`.

## `scope` vs exact `session_id` (DIR-030)

`scope="session"` (on `query_session_content`, `query_session_signals`, `query_file_activity`, and the analysis tools) means **"the most recently modified session in the current project"** — it is a convenience default, not a way to target one specific session by ID. It has meant this since before DIR-030 and that meaning is unchanged.

`session_id` is a separate, optional parameter on those same tools (and on `query_sessions`, see below) that selects **exactly one session/thread by its ID**, regardless of how recent it is or how many other sessions exist in the project:

- `scope="session"` — "give me whatever I was just working on" (most-recent, no ID needed).
- `session_id="<id>"` — "give me exactly this session" (id required, works for any session in the project, not just the most recent).

When `session_id` is set, the query reads only that one session — it never lists or loads any other session, which is both faster and precise for a project with many sessions. Combine `query_sessions` (to discover which session you want) with `session_id` on a content/signal tool (to query it):

```javascript
// 1. Discover: find the session about "migration"
const sessions = await query_sessions({
  provider: "codex",
  title_contains: "migration"
})

// 2. Query: read exactly that session's tool errors
query_session_signals({
  type: "errors",
  provider: "codex",
  session_id: sessions[0].session_id
})
```

## Tool Catalog

### `query_sessions` (session discovery)

Lists session/thread **metadata** — id, cwd, title, status, source_kind, model, model_provider, archived, parent_thread_id, is_subagent, created_at, updated_at — **without loading any turn/message content**. Use it to discover which session to target before querying it with `session_id` on another tool (see above). Default scope: project.

| Parameter | Type | Notes |
|-----------|------|-------|
| `session_id` | string | Exact session/thread ID; returns metadata for just that one session (still no turn content loaded). |
| `cwd` | string | Exact cwd to scope the listing to. Defaults to the project boundary derived from `working_dir`. |
| `archived` | boolean | Codex only. `true` = archived threads, `false` = active threads. **Archived threads are excluded by default** (DIR-032) — omitting both this and `status` is equivalent to `archived: false`. |
| `status` | string | Codex only. `"active"` or `"archived"` — alias for `archived`. Conflicting `status`/`archived` values are a validation error. Also defaults to active-only when omitted. |
| `source_kind` | array of string | Codex only. One or more of `cli`, `vscode`, `exec`, `appServer`, `subAgent`, `subAgentReview`, `subAgentCompact`, `subAgentThreadSpawn`, `subAgentOther`, `unknown`. An unrecognized value is a validation error. |
| `model_provider` | string | Codex only (e.g. `"openai"`). |
| `model` | string | Substring match against the model name. |
| `parent_thread_id` | string | Codex only. Exact parent/ancestor thread ID — lists that thread's direct **children**. |
| `ancestors_of` | string | Codex only (DIR-032). Given a thread ID, returns its **ancestor chain** (parent, grandparent, ...) instead of the normal listing. Each entry carries `lineage` (`"root"`, `"child"`, or `"unknown"` when spawn metadata wasn't available). Traversal stops — with a warning, never silently — at a confirmed root, an explicit `unknown`, a project/cwd boundary crossing, a cycle, or a depth limit. |
| `title_contains` | string | Case-insensitive substring match against the session title. |
| `created_since` / `created_until` | string (RFC3339) | Filter by creation time range. |
| `updated_since` / `updated_until` | string (RFC3339) | Codex only. Filter by last-updated time range. |
| `include_subagents` | boolean | Default `true`; `false` excludes subagent-sourced sessions. |
| `limit` | number | Max sessions to return. Combine with the standard `offset`/`page_size` parameters for pagination over a large project. |

Codex-only filters (`source_kind`, `model_provider`, `parent_thread_id`, `archived`, `status`, `ancestors_of`) fail with an actionable error if used while `provider` is `"claude"` (the default) — Claude sessions don't carry this metadata, so the filter could never match and a silent empty result would be indistinguishable from "no sessions found".

A session whose metadata can't be read (e.g. a corrupted `threads` row) is skipped with a warning rather than aborting the whole listing; other sessions still return normally.

Examples:

```javascript
// List active (non-archived) Codex sessions in the current project, most recent first
query_sessions({ provider: "codex" })

// Include archived subagent threads under a given parent (must opt in)
query_sessions({
  provider: "codex",
  archived: true,
  source_kind: ["subAgent"],
  parent_thread_id: "thread-abc"
})

// Exact lookup (metadata only, no turns)
query_sessions({ provider: "codex", session_id: "thread-abc" })

// Walk a subagent thread's ancestor chain
query_sessions({ provider: "codex", ancestors_of: "thread-abc" })
```

### Consolidated Query Tools

These three tools replace the previous 10 individual `query_*` tools. They scan the current project by default and accept `scope`, `provider`, `working_dir`, `limit`, `stats_only`, `stats_first`, and output parameters.

#### `query_session_content`

Query session messages by role. The `role` parameter is required.

| Role | Purpose | Claude Code | Codex |
|------|---------|-------------|-------|
| `user` | Search user messages by regex pattern | Yes | Yes |
| `assistant` | Query assistant messages (optionally filter by `contains`) | Yes | Yes |
| `tool` | Query `tool_use` or `tool_result` blocks | Yes | Yes |
| `all` | Query user/assistant conversation flow | Yes | Yes |

Additional parameters when `role=user`:
- `pattern` — regex filter applied to message content
- `content_type` — `string` (default) or `array` (tool results)
- `exclude_system_messages` — exclude Claude Code system-injected messages
- `max_message_length` — truncate content at N characters
- `min_content_length` / `max_content_length` — filter by content length
- `content_summary` — return summary only (session_id/turn/timestamp/preview)
- `group_by_session` — group results by session
- `context_turns` — include up to N adjacent records before/after each match (see "Context window semantics (`context_turns`)" below); supported for every provider (`claude`, `codex`, `all`)
- `since` / `until` — RFC3339 time range filter

Additional parameters when `role=assistant`:
- `contains` — substring filter (case-insensitive); use `"## Summary"` to retrieve summaries

Additional parameters when `role=tool`:
- `block_type` — `tool_use` (default) or `tool_result`
- `tool_name` — substring or regex filter on the tool's `name` field (only applies to `block_type=tool_use`); e.g. `"Dispatch"` or `"Read|Write"`

Examples:

```javascript
// User messages about migration
query_session_content({
  role: "user",
  provider: "codex",
  pattern: "migration",
  limit: 20
})

// Session summaries
query_session_content({
  role: "assistant",
  contains: "## Summary",
  stats_first: true
})

// Conversation flow
query_session_content({
  role: "all",
  scope: "session"
})
```

##### Context window semantics (`context_turns`)

`context_turns=N` expands each matched `role=user` record to include up to `N`
adjacent records on each side, from the same session only:

- **Ordering/counting is by record position**, not turn count or timestamp:
  each provider assigns every record a stable position in that session's
  normalized record stream when the query first loads it. Windows are
  bounded at the session's start/end, and overlapping windows (from two
  nearby matches) are merged so no record repeats. Records sharing one
  timestamp still get distinct positions, so `context_turns` never conflates
  them.
- **Matched vs. added records**: a matched record is returned with
  `context: false`; a record added purely for context is returned with
  `context: true`.
- **Provider support is uniform**: `context_turns` works identically for
  `provider: "claude"`, `"codex"` (both the rollout and app-server Codex
  backends), and `"all"`. A match returned with `context_turns: 0` is never
  silently erased by adding `context_turns > 0` — this was a defect (DIR-036)
  fixed by routing context loading through the same provider/session
  abstraction the original query used, instead of assuming Claude's on-disk
  JSONL layout for every provider.
- **Failure handling**: if a session's full history cannot be reloaded for
  context (e.g. a transient backend failure), the session's original matches
  are still returned (with `context: false` and no added context) and a
  warning is included in the response — never a silent, unexplained empty
  result for a query that had real matches.
- `context_turns` composes with `group_by_session`, pagination,
  `exclude_compact_summaries`, `scope: "session"`, and an explicit
  `session_id` without erasing matches. It also composes with
  `content_summary`: the match is preserved, but (both for Claude and Codex)
  the expanded window's records reflect the freshly reloaded canonical
  records rather than the `content_summary` preview projection — this is
  pre-existing, provider-symmetric behavior, not something DIR-036 changed.

#### `query_session_signals`

Query structured session signals. The `type` parameter is required.

| Type | Purpose | Claude Code | Codex |
|------|---------|-------------|-------|
| `errors` | Failed tool results | Yes | Yes |
| `tokens` | Assistant token usage | Yes | Yes |
| `system_errors` | Claude API system errors | Yes | Host-specific empty |
| `timestamps` | All timestamped entries | Yes | Yes |
| `tool_stats` | Assistant tool calls | Yes | Yes |

Additional parameters when `type=tool_stats`:
- `tool` — filter by tool name
- `status` — filter by outcome: `"error"` or `"success"` (DIR-043). Each
  returned record is a `tool_use` whose paired `tool_result` (correlated by
  `tool_use_id`, since the outcome lives on a separate, later JSONL record)
  had `is_error` matching the requested value. A `tool_use` with no observed
  `tool_result` yet (e.g. a call still in flight) is excluded from both
  `status: "error"` and `status: "success"` results — so `error` + `success`
  counts can sum to less than, but never more than, the unfiltered
  `tool_stats` total. Any value other than `error`/`success` is a validation
  error, not a silently-ignored filter.

**`provider: "all"` and outcome filtering (`status`, and `type: "errors"`)**
(DIR-046): `provider: "all"` (and `provider: "codex"`) route Claude session
data through a normalized, cross-provider record shape rather than raw
Claude JSONL. A join-direction bug in that normalization path used to
correlate each Claude `tool_use` with the WRONG neighboring `tool_result`
(the one answering the *previous* tool call, not this one), which silently
dropped almost all Claude-sourced outcome/error signal once records passed
through it — `status: "error"`/`status: "success"` under `provider: "all"`
would collapse to near-zero or a single session, and `type: "errors"` would
do the same. This is now fixed at the source (Claude turn/tool-result
pairing), so `status` filtering and `type: "errors"` return complete,
multi-session, multi-provider results under `provider: "all"` exactly like
`provider: "claude"` does alone — no separate workaround or warning is
needed.

All types support `since` / `until` RFC3339 time range filters.

Examples:

```javascript
// Tool errors with stats
query_session_signals({
  type: "errors",
  provider: "claude",
  scope: "session",
  stats_first: true
})

// Token usage
query_session_signals({
  type: "tokens",
  provider: "codex",
  stats_first: true,
  limit: 20
})

// Tool calls filtered by name
query_session_signals({
  type: "tool_stats",
  tool: "exec_command",
  provider: "all",
  working_dir: "/path/to/project",
  limit: 50
})
```

#### `query_file_activity`

Query file history and activity. The `type` parameter is required.

| Type | Purpose | Claude Code | Codex |
|------|---------|-------------|-------|
| `snapshots` | File history snapshots with messageId | Yes | Host-specific empty |

Example:

```javascript
query_file_activity({
  type: "snapshots",
  scope: "project",
  limit: 20
})
```

### Two-Stage Query Tools

Use these when you need file selection control or custom jq over selected session files.

| Tool | Purpose | Claude Code | Codex |
|------|---------|-------------|-------|
| `get_session_directory` | Locate session directory and aggregate metadata | Yes | Yes |
| `inspect_session_files` | Inspect selected JSONL files for record counts, time ranges, and samples | Yes | Yes (files pre-selected by `get_session_directory`) |
| `execute_stage2_query` | Run jq-style filter/sort/transform on selected files | Yes | Yes (jq expression must match the selected provider's raw schema) |
| `get_session_metadata` | Return schema hints, file info, and query templates | Yes | Yes |

`get_session_directory` and `get_session_metadata` accept `provider` (`claude` default, `codex`, or `all`) and `working_dir` just like the convenience query tools (see [Standard Parameters](#standard-parameters)). They never mix providers into one response:

- `provider: "claude"` (default): unchanged response shape — a single `directory` plus aggregate counts.
- `provider: "codex"`: resolves only Codex rollout files, returned as an explicit `files` list (Codex sessions are not guaranteed to share one directory) with Codex-specific `jsonl_schema`/`query_templates`.
- `provider: "all"`: an explicit `{ "providers": { "claude": {...}, "codex": {...} } }` breakdown — Claude and Codex files/schemas are never merged. A provider with no data appears in `warnings` instead of silently vanishing; the call only fails if *no* provider has data. An invalid or unavailable single provider (e.g. `codex` with no Codex session state) fails closed with an error rather than silently falling back to Claude.

Example workflow (Claude, default):

```javascript
const dir = await get_session_directory({
  scope: "project"
})

const inspection = await inspect_session_files({
  files: ["/path/to/session.jsonl"],
  include_samples: true
})

const results = await execute_stage2_query({
  files: ["/path/to/session.jsonl"],
  filter: 'select(.type == "assistant") | select(.message | has("usage"))',
  transform: '{timestamp, usage: .message.usage}',
  limit: 20
})
```

Example workflow (Codex):

```javascript
const dir = await get_session_directory({
  scope: "project",
  provider: "codex",
  working_dir: "/path/to/project"
})
// dir.files: ["/home/user/.codex/sessions/2026/06/14/rollout-abc.jsonl", ...]

const inspection = await inspect_session_files({
  files: dir.files,
  include_samples: true
})

const results = await execute_stage2_query({
  files: dir.files,
  filter: 'select(.type == "item.tool_result" and .payload.is_error == true)',
  limit: 20
})
```

### Analysis Tools

| Tool | Purpose | Claude Code | Codex |
|------|---------|-------------|-------|
| `analyze_errors` | Aggregate tool errors by tool and type | Yes | Yes |
| `analyze_bugs` | Detect error-fix pairs and recurring bug patterns | Yes | Yes |
| `quality_scan` | Compute error, retry, diversity, and completion dimensions | Yes | Yes |
| `get_work_patterns` | Tool frequency, hourly activity, context switches | Yes | Yes |
| `get_timeline` | Chronological session events | Yes | Yes |
| `get_tech_debt` | TODO/FIXME/HACK markers and unresolved errors | Yes | Yes |

Example:

```javascript
get_work_patterns({
  provider: "codex",
  working_dir: "/path/to/project"
})

analyze_errors({
  provider: "all",
  scope: "project",
  limit: 10
})
```

### Cleanup

| Tool | Purpose |
|------|---------|
| `cleanup_temp_files` | Remove old temporary MCP output files |

## Standard Parameters

Most query tools accept:

| Parameter | Type | Description |
|-----------|------|-------------|
| `scope` | string | `project` (default) or `session` (= most recently modified session — see [`scope` vs exact `session_id`](#scope-vs-exact-session_id-dir-030)) |
| `provider` | string | `claude` (default), `codex`, or `all` |
| `working_dir` | string | Override project path used for session lookup |
| `session_id` | string | Exact session/thread ID (content/signal/file-activity/analysis tools, and `query_sessions`). Reads only that one session; distinct from `scope="session"`. |
| `limit` | number | Maximum results; default is no limit |
| `stats_only` | boolean | Return aggregate statistics only — no per-item `examples`/detail arrays. Honored consistently by all six `analysis.Service` tools (`analyze_errors`, `analyze_bugs`, `quality_scan`, `get_work_patterns`, `get_tech_debt`, `get_timeline`); `analyze_errors`/`analyze_bugs` return per-tool/per-pattern *counts* with no `examples` text, and `get_tech_debt` returns a `hotspot_file_count` in place of the full per-file `hotspot_files` list. `quality_scan` and `get_work_patterns` are already aggregate-only, so `stats_only` returns their normal (unchanged) shape. |
| `stats_first` | boolean | Return stats followed by details |
| `inline_threshold_bytes` | number | Threshold for inline vs file reference output |

### `output_format` (DIR-044: scoped to 4 tools, not universal)

`output_format` (`jsonl` default, or `tsv`) is declared only on the four consolidated query
tools whose results are built via `internal/mcp/pipeline.BuildResponse`: `query_sessions`,
`query_session_content`, `query_session_signals`, `query_file_activity`. It is genuinely
implemented (not just schema-declared): `output_format: "tsv"` returns a header row (the
sorted union of field names across all returned records) followed by one tab-separated row
per record. Nested object/array field values are JSON-encoded inline within their own cell
rather than flattened, and any literal tab/newline/backslash inside a scalar string value is
backslash-escaped so row structure (one record per line, fields split on tab) stays intact.
Envelope metadata (`mode`, `pagination`, `warnings`) is emitted as leading `#`-prefixed
comment lines before the header rather than being silently dropped. TSV only applies to
inline-mode responses — large result sets returned in `file_ref` mode are unaffected by
`output_format` and remain JSON (the data already lives in a separate JSONL temp file).

The other 12 tools — the six `analysis.Service` tools (`analyze_errors`, `analyze_bugs`,
`quality_scan`, `get_work_patterns`, `get_timeline`, `get_tech_debt`), plus
`query_edit_sequences`, `cleanup_temp_files`, `get_session_directory`,
`inspect_session_files`, `execute_stage2_query`, and `get_session_metadata` — do not declare
`output_format` in their MCP schema. Each of them marshals its own typed, non-flat-record
result struct to JSON directly and never touches `PipelineConfig`, so retrofitting TSV there
would mean special-casing each analyzer's distinct result shape rather than a single bounded
change. Previously `output_format` was declared (via `StandardToolParameters()`) on all 16
tools uniformly, but `pipeline.BuildResponse` never branched on it, so every tool silently
ignored the parameter regardless of what was passed.

## Hybrid Output Mode

meta-cc uses hybrid output:

- Small responses are returned inline.
- Large responses are written to a temporary file and returned as `file_ref`.

This keeps MCP responses usable for both short interactive questions and large project-level scans.

## Common Recipes

Find Codex user prompts about a topic:

```javascript
query_session_content({
  role: "user",
  provider: "codex",
  pattern: "release|deploy",
  limit: 20
})
```

Count tool usage across providers:

```javascript
get_work_patterns({
  provider: "all"
})
```

Inspect token usage:

```javascript
query_session_signals({
  type: "tokens",
  provider: "codex",
  stats_first: true
})
```

Run custom jq on selected files:

```javascript
execute_stage2_query({
  files: ["/path/to/session.jsonl"],
  filter: 'select(.type == "user" and (.message.content | type == "string"))',
  transform: '{timestamp, content: .message.content}',
  limit: 100
})
```

## Migration from Previous Tools

The 10 previous individual query tools have been replaced:

| Old tool | New equivalent |
|----------|---------------|
| `query_user_messages` | `query_session_content(role="user")` |
| `query_tool_blocks` | `query_session_content(role="tool")` |
| `query_summaries` | `query_session_content(role="assistant", contains="## Summary")` |
| `query_conversation_flow` | `query_session_content(role="all")` |
| `query_tool_errors` | `query_session_signals(type="errors")` |
| `query_token_usage` | `query_session_signals(type="tokens")` |
| `query_system_errors` | `query_session_signals(type="system_errors")` |
| `query_timestamps` | `query_session_signals(type="timestamps")` |
| `query_tools` | `query_session_signals(type="tool_stats")` |
| `query_file_snapshots` | `query_file_activity(type="snapshots")` |

## See Also

- [MCP Server Guide](mcp.md)
- [Two-Stage Query Guide](two-stage-query-guide.md)
- [JSONL Schema Reference](../reference/jsonl-schema.md)
