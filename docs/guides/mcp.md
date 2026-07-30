# MCP Server Guide

meta-cc provides a Model Context Protocol (MCP) server for local coding-agent history analysis. It supports Claude Code and Codex through a provider-aware conversation layer, then exposes normalized records to convenience query tools, two-stage jq queries, and higher-level analysis tools.

## Provider Support

| Provider | Local source | Notes |
|----------|--------------|-------|
| `claude` | `~/.claude/projects/<project-hash>/*.jsonl` | Host default under Claude Code and for standalone installations. |
| `codex` | Highest-compatible `state_N.sqlite` under the canonical Codex root (`META_CC_CODEX_ROOT` → `CODEX_HOME` → `~/.codex`), plus rollout JSONL paths from the `threads` table; rollout-only fallback when no compatible database exists | Host default under Codex. `~/.codex/history.jsonl` is intentionally excluded. See [Discovery roots](../reference/codex-history-model.md#discovery-roots-and-database-compatibility-dir-069). |
| `all` | Claude Code and Codex | Returned records include a `provider` field. |

Convenience query and analysis tools accept the standard `provider` parameter:

```javascript
query_session_content({
  role: "user",
  provider: "codex",
  pattern: "migration",
  limit: 20
})

get_work_patterns({
  provider: "all",
  working_dir: "/path/to/project"
})
```

If `provider` is omitted, meta-cc resolves it to the host that launched the MCP
server (DIR-073): `claude` under Claude Code, `codex` under Codex. Explicit
`provider` values always override this host default. Responses preserve
provider provenance (a `provider` field on sessions/records) so you can verify
which corpus was searched.

### Default provider for manual/standalone installs (`META_CC_HOST`)

Packaged plugin manifests inject `META_CC_HOST` (`claude` for the Claude Code
plugin, `codex` for the Codex plugin) into the MCP server environment; this is
the single source of truth for the omitted-provider default. Manual or
standalone installations have a deterministic fallback: when `META_CC_HOST` is
unset the default is `claude` (pre-DIR-073 behavior). To change the default
without touching query arguments, set it explicitly in your MCP client
configuration:

```json
{
  "mcpServers": {
    "meta-cc": {
      "command": "meta-cc-mcp",
      "args": [],
      "env": { "META_CC_HOST": "codex" }
    }
  }
}
```

An invalid `META_CC_HOST` (anything other than `claude` or `codex`) fails
server startup with an actionable error instead of silently reading the wrong
history.

## Configuration

### Claude Code

The release archive includes `.mcp.json`; the installer merges it into `~/.claude/mcp.json`.

Manual configuration (an unset `META_CC_HOST` deterministically defaults to
`claude`; set it to `codex` when this server is launched from Codex):

```json
{
  "mcpServers": {
    "meta-cc": {
      "command": "meta-cc-mcp",
      "args": [],
      "env": { "META_CC_HOST": "claude" }
    }
  }
}
```

### Codex

The release archive includes `.codex-plugin/plugin.json` and `.codex-mcp.json`. The installer copies them under `~/.codex/plugins/meta-cc/`.

`CODEX_HOME` targets a non-default Codex home for both installation and
history reads; `META_CC_CODEX_ROOT` is a meta-cc-only override that takes
precedence over `CODEX_HOME` when both are set (precedence:
`META_CC_CODEX_ROOT` → `CODEX_HOME` → `~/.codex` — see
[Discovery roots](../reference/codex-history-model.md#discovery-roots-and-database-compatibility-dir-069)):

```bash
CODEX_HOME=/tmp/codex ./install.sh
META_CC_CODEX_ROOT=/tmp/codex meta-cc-mcp
```

## Tool Catalog

The MCP server currently registers **16 tools**, derived at test time from
`internal/mcp/tools.GetToolDefinitions()` (see `internal/release` for the
consistency check that keeps this document's tool-count claims from
drifting from the live `tools/list` response).

### Session Discovery

| Tool | Purpose | Claude Code | Codex |
|------|---------|-------------|-------|
| `query_sessions` | List session/thread metadata (id, cwd, title, status, source_kind, archived, ...) without loading turn content. Supports an exact `session_id` lookup. | Yes | Yes (full filter set: archived, source_kind, model_provider, parent_thread_id) |

`query_sessions` is metadata-first (DIR-030): it never parses Codex rollout content or Claude turn content, so it stays fast even against a project with many sessions. See [MCP Query Tools Reference](mcp-query-tools.md#query_sessions-session-discovery) for the full filter list and the `scope="session"` vs exact `session_id` distinction.

### Consolidated Query Tools

These 3 tools replace the older set of individual `query_*` tools (see
[MCP Query Tools Reference](mcp-query-tools.md) for full parameter docs):

| Tool | Purpose | Claude Code | Codex |
|------|---------|-------------|-------|
| `query_session_content` | Query messages by role: `user`, `assistant`, `tool`, or `all` | Yes | Yes |
| `query_session_signals` | Query signals: `errors`, `tokens`, `system_errors`, `timestamps`, `tool_stats` | Yes | Yes (`system_errors` is Claude-specific) |
| `query_file_activity` | Query file history (`type: "snapshots"`) | Yes | Host-specific empty |

All three of `query_session_content`, `query_session_signals`, and `query_file_activity` also accept an exact `session_id` parameter (DIR-030): when set, the query reads only that one session instead of listing/loading every session in scope — see [`scope` vs exact `session_id`](mcp-query-tools.md#scope-vs-exact-session_id-dir-030).

Examples:

```javascript
query_session_signals({
  type: "tool_stats",
  provider: "codex",
  tool: "exec_command",
  working_dir: "/path/to/project",
  limit: 50
})

query_session_signals({
  type: "tokens",
  provider: "codex",
  stats_first: true,
  limit: 20
})
```

### Utility Tools

| Tool | Purpose |
|------|---------|
| `cleanup_temp_files` | Remove old temporary MCP output files (`file_ref` mode artifacts) |

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
| `analyze_errors` | `measured` | TotalErrors/ByTool are direct counts; `by_type` uses normalized signature and label classification heuristics |
| `analyze_bugs` | `measured` | `patterns`/`total_pairs` infer causal fixes from a same-tool success within three positions |
| `quality_scan` | `measured` | Error/diversity/completion dimensions aggregate observations; retry rate infers retries from same-tool proximity (see [Quality scan dimensions](#quality-scan-dimensions)) |
| `get_work_patterns` | `measured` | ToolFrequency/HourlyActivity/PeakHour are direct counts; `context_switches` uses a heuristic (file changes within 5 min) |
| `get_timeline` | `measured` | Events parsed directly from session entries |
| `get_tech_debt` | `measured` | Session `markers`/`hotspot_files` scan observed tool output; source-dir scans list both fields as estimated because comment detection is lexical; `open_issues` uses an error→success heuristic |

#### Quality scan dimensions

`quality_scan` reports four dimensions, each a score in `0.0–1.0` (higher =
better quality) with a `raw_value` count:

| Dimension | Semantics |
|-----------|-----------|
| `error_rate` | `1 − errors/total`; a tool call is an error when its status is `error` or it carries an error message |
| `retry_rate` | `1 − retries/total`; a retry is an errored call followed by a call to the same tool within the next 5 positions |
| `tool_diversity` | Size-independent *evenness* of the per-tool call distribution: Shannon entropy `H = −Σ pᵢ·ln(pᵢ)` normalized by `ln(unique tools)`. `1.0` = perfectly even usage across the tools actually used; approaches `0` as usage concentrates on one tool; a single tool is trivially even (`1.0`). Does not decay as call volume grows (`raw_value` is `unique/total`) |
| `completion_rate` | Fraction of calls that produced an observed `tool_result` at all (status present). Distinct from `error_rate`: an errored call still completed (it returned a result), while a `tool_use` with no observed `tool_result` is incomplete without being an error |

### Two-Stage Query Tools

Use these when you need file selection control or custom jq over selected JSONL files:

| Tool | Purpose |
|------|---------|
| `get_session_directory` | Locate a transcript directory and aggregate file metadata |
| `inspect_session_files` | Inspect selected JSONL files for record counts, time ranges, and samples |
| `execute_stage2_query` | Run jq-style filter/sort/transform on selected files |
| `get_session_metadata` | Return schema hints, file info, and query templates |
| `query_edit_sequences` | Analyze file edit/read patterns: docRole, co-accessed docs, DocVoid |

Two-stage tools operate on selected files. They retain raw-file compatibility, including normalized Codex JSONL records when a Codex rollout/session file is selected directly. Provider-aware cross-host querying is handled by the convenience query and analysis tools through the `provider` parameter.

`execute_stage2_query` responses include a bounded `diagnostics` envelope (`backend`, `provider`/`provider_effective`, `files_considered`/`files_loaded`/`files_skipped`, `records_scanned`, `matches_returned`, `truncated`, `degraded`, `skip_warnings`) so an empty result can be told apart from an incomplete or degraded search. Unreadable files are skipped with bounded warnings (degraded mode); if no file loads, the call fails closed. A type preflight runs your jq on a small representative sample before the full scan and fails fast — naming the observed input type and a correction — on common mismatches such as `test()` applied to an object. `inspect_session_files` expects file paths; passing a directory returns a structured correction pointing to the `get_session_directory(...).files` workflow. See the [MCP Query Tools Reference](mcp-query-tools.md) for field detail.

## Codex Normalization

The Codex provider reads session metadata from the highest-compatible `state_N.sqlite` under the canonical Codex root and follows each thread's `rollout_path` (with a cwd-enforced rollout-only fallback when no compatible database exists). It normalizes:

- legacy `response_item` messages with `input_text` / `output_text`
- `function_call` and `function_call_output`
- `custom_tool_call` and `custom_tool_call_output`
- `event_msg` `token_count` usage
- newer dotted schema events such as `turn.started`, `item.message`, `item.tool_call`, and `item.tool_result`

Codex `tokens_used` from SQLite is retained as session metadata, but `query_session_signals({type: "tokens", provider: "codex"})` reports per-turn usage only when the rollout contains a `token_count` event.

## Standard Parameters

Most query and analysis tools accept:

| Parameter | Type | Description |
|-----------|------|-------------|
| `scope` | string | `project` (default) or `session` (= most recently modified session) |
| `provider` | string | `claude`, `codex`, or `all`; omitted resolves to the active host (`META_CC_HOST`, standalone fallback `claude` — see [Default provider](#default-provider-for-manualstandalone-installs-meta_cc_host)) |
| `working_dir` | string | Override project path used for session lookup |
| `session_id` | string | Exact session ID (query/analysis tools and `query_sessions`). Distinct from `scope="session"` — see [MCP Query Tools Reference](mcp-query-tools.md#scope-vs-exact-session_id-dir-030). |
| `limit` | number | Maximum results; default is no limit |
| `stats_only` | boolean | Return aggregate statistics only |
| `stats_first` | boolean | Return stats followed by details. Only declared on the 4 tools routed through the response pipeline (see below) — not on the six analysis tools. |
| `inline_threshold_bytes` | number | Threshold for inline vs file reference output |

RFC3339 `since`/`until` time filters are declared only on `query_session_content`, `query_session_signals`, and `get_timeline` — not on every query tool. `query_sessions` filters session metadata with `created_since`/`created_until` (plus Codex-only `updated_since`/`updated_until`) instead.

### `jq_filter`, `stats_first`, `offset`, `page_size` (DIR-048: scoped to pipeline-routed tools)

Like `output_format` below, `jq_filter`, `stats_first`, `offset`, and `page_size` are only
meaningful against the flat record array produced by the shared response pipeline
(`internal/mcp/pipeline.BuildResponse`): `jq_filter` post-filters that array, `stats_first`
returns a stats header followed by the record details, and `offset`/`page_size` paginate over
it. They are declared only on `query_sessions`, `query_session_content`,
`query_session_signals`, and `query_file_activity`.

The six analysis tools (`analyze_errors`, `analyze_bugs`, `quality_scan`,
`get_work_patterns`, `get_timeline`, `get_tech_debt`) do not declare these four parameters:
they dispatch straight to `internal/analysis.Service` and return before the response pipeline
is ever reached, and each returns its own typed, non-flat-record analyzer result rather than a
record array these parameters could operate on. `stats_only` (genuinely wired per-tool since
DIR-042) and `include_subagents` are unaffected and remain available on all six.

### `output_format` (DIR-044: scoped to 4 tools, not universal)

`output_format` (`jsonl` default, or `tsv`) is declared only on the four consolidated query
tools whose results are built via the shared response pipeline: `query_sessions`,
`query_session_content`, `query_session_signals`, `query_file_activity`. It is genuinely
implemented — `output_format: "tsv"` returns a header row (the sorted union of field names
across all returned records) followed by one tab-separated row per record; nested
object/array field values are JSON-encoded inline within their own cell rather than
flattened, and any literal tab/newline/backslash inside a scalar string value is
backslash-escaped. Envelope metadata (`mode`, `pagination`, `warnings`) is emitted as leading
`#`-prefixed comment lines before the header. TSV only applies to inline-mode responses —
large (`file_ref`-mode) result sets remain JSON regardless of `output_format`.

The other 12 tools (`analyze_errors`, `analyze_bugs`, `quality_scan`, `get_work_patterns`,
`get_timeline`, `get_tech_debt`, `query_edit_sequences`, `cleanup_temp_files`,
`get_session_directory`, `inspect_session_files`, `execute_stage2_query`,
`get_session_metadata`) do not declare `output_format` — they marshal their own typed,
non-flat-record result structs directly and never route through the pipeline that
implements TSV serialization, so the parameter isn't advertised on them (previously it was
declared on all 16 tools but silently ignored everywhere).

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
  },
  "data_source": "measured"
}
```

Use this as a **scout step** before deciding how to query:

```text
1. get_timeline(scope=project, stats_only=true)   → assess scale and time range
2a. entries are small  → get_timeline(scope=project, limit=50)
2b. entries are large  → get_timeline(scope=session)           # current session only
                       → get_timeline(scope=project, limit=N)  # explicit cap
                       → get_timeline(since="2026-06-01T00:00:00Z")  # focused time range (native since/until, not jq_filter)
```

When `limit` truncates the result, the response includes `truncated: true` and
`total_events: N` so you know more data exists.

## Output Modes

The MCP server uses hybrid output:

- Small responses are returned inline.
- Large responses are written to a temporary file and returned as `file_ref`.

`file_ref` is a transport mode only — the inline `data` array and the JSONL file behind a `file_ref` hold the same logical records. Since DIR-080 the reference is self-describing: alongside `path`/`size_bytes`/`line_count`/`fields`/`summary`, it carries a versioned nested `shape` (typed jq paths with `optional`/`nullable` flags, heterogeneous `variants`, array `elements`, and provider-provenance `values`), a bounded redacted `sample`, and server-validated jq `recipes` generated from the emitted shape. Use the shipped recipes (scope `"record"`) directly as `execute_stage2_query` transforms or as `.[] | <recipe>` `jq_filter` expressions instead of reverse-engineering result structure. See [MCP Query Tools Reference](mcp-query-tools.md#self-describing-file_ref-results-dir-080) and [Two-Stage Query Guide](two-stage-query-guide.md) for details.

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
- For Codex, verify the highest `state_N.sqlite` under the canonical root (`${META_CC_CODEX_ROOT:-${CODEX_HOME:-$HOME/.codex}}`) has a `threads` row whose `cwd` matches the project and whose `rollout_path` points to a readable JSONL file. If no compatible database exists, sessions come from the rollout trees directly — check `${META_CC_CODEX_ROOT:-${CODEX_HOME:-$HOME/.codex}}/sessions/` for `*.jsonl` files whose `session_meta` records the project `cwd`.

### Tool returns empty on Codex

Some queries target Claude Code-only record types:

- `query_file_activity` (`type: "snapshots"`)
- `query_session_signals` (`type: "system_errors"`)
- Compact summaries via `query_session_content` (`role: "assistant"`, `contains: "## Summary"`)

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
for the current analysis tools), and the response includes an
`estimated_fields` list naming the JSON fields derived heuristically:

- **`TechDebtResult`** — session scans list `"open_issues"`, based on the
  heuristic "an error call with no subsequent success for the same tool".
  Source-directory scans list `"markers"` and `"hotspot_files"`: their
  comment-context detection uses the language-aware lexical `markerScanner`,
  which deliberately under-counts unsupported syntax rather than claiming
  parser-level certainty. Merged results union the contributing lists.

- **`WorkPatternsResult.context_switches`** — listed as `"context_switches"`:
  based on the heuristic "file-path changes between consecutive tool calls
  within a 5-minute window". The rest of the struct (ToolFrequency,
  HourlyActivity, PeakHour) is measured.

- **`QualityScanResult.dimensions`** — listed as `"dimensions"` because the
  serialized dimensions array combines the estimated `retry_rate` element with
  directly aggregated elements; a same-tool call within five positions after
  an error is inferred to be a retry.

- **`BugAnalysisResult` / `BugAnalysisStats`** — pair and pattern fields are
  listed because a same-tool success within three positions after an error is
  inferred to be that error's fix.

- **`ErrorAnalysisResult.by_type` / `ErrorAnalysisStats.by_type`** — listed as
  `"by_type"`: human-readable labels and signatures use normalization and
  classification heuristics.

- **`EditSequencesResult`** — lists heuristic descendant paths for file/document
  roles, pattern hints, documentation-gap signals, and pattern distribution.
  Wildcards such as `files.*.docVoid` identify fields on every map entry.

When building automated pipelines that depend on precise values, treat
fields with known estimated provenance with appropriate uncertainty margins.

## See Also

- [MCP Query Tools Reference](mcp-query-tools.md)
- [Two-Stage Query Guide](two-stage-query-guide.md)
- [GCL Gate Annotation Format](gcl-annotation.md)
- [JSONL Schema Reference](../reference/jsonl-schema.md)
- [Installation Guide](../tutorials/installation.md)
