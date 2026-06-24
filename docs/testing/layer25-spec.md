# Layer 2.5 Oracle Fixture Specification

## Overview

This document specifies the λ-spec (lambda-spec) for meta-cc MCP tool behavioral testing using
the BAIME Exp-D oracle framework. Go unit tests verify internal logic; Layer 2.5 fills the gap
by testing tool behavior from Claude's calling perspective using fixture-based assertions evaluated
by a Haiku oracle.

**Oracle target**: ≥0.85 accuracy on Class A decisions across all fixtures.

---

## Decision Classes

### Class A — Field Presence (binary, Haiku oracle)

**Question**: Does the output contain all required fields for this query type?

**Oracle prompt template** (see oracle-runner.md for full template):
> Given the following MCP tool output JSON, answer YES or NO: does every returned record contain
> the required fields listed below? Answer with YES or NO followed by a confidence score 0.0–1.0.

**Pass criterion**: Oracle answers YES with confidence ≥ 0.7.

### Class B — Semantic Correctness (nuanced, Haiku oracle)

**Question**: Does the content of filtered results actually match the filter criteria?

**Pass criterion**: Oracle answers YES with confidence ≥ 0.7.

### Class C — Ordering / Structural (deterministic, no oracle needed)

**Question**: Are results ordered correctly (e.g., descending by timestamp)?

**Pass criterion**: Programmatic check on returned timestamps.

---

## λ-spec: `query_session_signals`

### Tool Identity

```
tool: query_session_signals
MCP server: meta-cc
```

### Parameter Schema

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `type` | string | YES | — | Signal type: `errors`, `tokens`, `system_errors`, `timestamps`, `tool_stats` |
| `scope` | string | NO | `project` | `project` or `session` |
| `provider` | string | NO | `claude` | `claude`, `codex`, or `all` |
| `working_dir` | string | NO | cwd | Override project path |
| `limit` | number | NO | none | Max results |
| `stats_only` | boolean | NO | false | Return aggregate stats only |
| `stats_first` | boolean | NO | false | Stats followed by details |
| `since` | string | NO | — | RFC3339 start time |
| `until` | string | NO | — | RFC3339 end time |
| `tool` | string | NO | — | Filter by tool name (type=tool_stats only) |
| `status` | string | NO | — | `error` or `success` (type=tool_stats only) |

### Sub-type: `type=errors`

**Maps to**: JSONL filter `select(.type == "user" and (.message.content | type == "array")) | select(.message.content[] | select(.type == "tool_result" and .is_error == true))`

**Required output fields per record**:
- `type` — always `"user"`
- `timestamp` — ISO8601 string
- `sessionId` — UUID string
- `message.content` — array containing at least one tool_result block
- `message.content[].type` — `"tool_result"`
- `message.content[].is_error` — `true`

**Ordering rule**: Records are returned in file-read order (chronological ascending by default).

**Class A assertion**: Each record contains `timestamp`, `sessionId`, and at least one `tool_result` block with `is_error: true`.

### Sub-type: `type=tokens`

**Maps to**: JSONL filter `select(.type == "assistant" and has("message")) | select(.message | has("usage"))`

**Required output fields per record**:
- `type` — always `"assistant"`
- `timestamp` — ISO8601 string
- `sessionId` — UUID string
- `message.usage` — object with token counts
- `message.usage.input_tokens` — integer ≥ 0
- `message.usage.output_tokens` — integer ≥ 0

**Class A assertion**: Each record contains `timestamp`, `sessionId`, and `message.usage` with `input_tokens` and `output_tokens`.

### Sub-type: `type=system_errors`

**Maps to**: JSONL filter `select(.type == "system" and .subtype == "api_error")`

**Required output fields per record**:
- `type` — always `"system"`
- `subtype` — always `"api_error"`
- `timestamp` — ISO8601 string
- `sessionId` — UUID string

**Note**: Returns empty for Codex provider (host-specific Claude Code record type).

**Class A assertion**: Each record contains `type: "system"`, `subtype: "api_error"`, `timestamp`, and `sessionId`.

### Sub-type: `type=timestamps`

**Maps to**: JSONL filter `select(.timestamp != null)`

**Required output fields per record**:
- `timestamp` — ISO8601 string (non-null)
- `type` — record type string
- `sessionId` — UUID string

**Class A assertion**: Each record contains a non-null `timestamp`, `type`, and `sessionId`.

### Sub-type: `type=tool_stats`

**Maps to**: JSONL filter `select(.type == "assistant") | select(.message.content[] | .type == "tool_use")`

**Required output fields per record**:
- `type` — always `"assistant"`
- `timestamp` — ISO8601 string
- `sessionId` — UUID string
- `message.content` — array containing at least one `tool_use` block
- `message.content[].type` — `"tool_use"`
- `message.content[].name` — tool name string
- `message.content[].id` — tool call ID string

**Class A assertion**: Each record contains `timestamp`, `sessionId`, and at least one `tool_use` block with `name` and `id`.

---

## λ-spec: `query_session_content`

### Tool Identity

```
tool: query_session_content
MCP server: meta-cc
```

### Parameter Schema

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `role` | string | YES | — | Message role: `user`, `assistant`, `tool`, `all` |
| `scope` | string | NO | `project` | `project` or `session` |
| `provider` | string | NO | `claude` | `claude`, `codex`, or `all` |
| `working_dir` | string | NO | cwd | Override project path |
| `limit` | number | NO | none | Max results |
| `stats_only` | boolean | NO | false | Return aggregate stats only |
| `stats_first` | boolean | NO | false | Stats followed by details |
| `pattern` | string | NO | `".*"` | Regex filter on message content (role=user) |
| `content_type` | string | NO | `string` | `string` or `array` (role=user) |
| `exclude_system_messages` | boolean | NO | false | Exclude system-injected messages (role=user) |
| `max_message_length` | number | NO | — | Truncate content at N chars (role=user) |
| `min_content_length` | number | NO | — | Min content length filter (role=user) |
| `max_content_length` | number | NO | — | Max content length filter (role=user) |
| `content_summary` | boolean | NO | false | Return summary only (role=user) |
| `group_by_session` | boolean | NO | false | Group results by session (role=user) |
| `context_turns` | number | NO | — | Include N turns before/after match (role=user) |
| `since` | string | NO | — | RFC3339 start time |
| `until` | string | NO | — | RFC3339 end time |
| `contains` | string | NO | — | Substring filter (role=assistant) |
| `block_type` | string | NO | `tool_use` | `tool_use` or `tool_result` (role=tool) |

### Sub-role: `role=user`

**Maps to**: `handleQueryUserMessages` — JSONL filter `select(.type == "user" and (.message.content | type == "string"))` plus optional pattern test.

**Required output fields per record**:
- `type` — always `"user"`
- `timestamp` — ISO8601 string
- `sessionId` — UUID string
- `message.role` — always `"user"`
- `message.content` — string matching the pattern (if pattern provided)

**Filter semantics**: Pattern is a Go/jq-compatible regex applied via `test(pattern)` on `.message.content`. Default pattern `".*"` matches all string messages.

**Class A assertion**: Each record contains `timestamp`, `sessionId`, `message.role: "user"`, and string `message.content`.

**Class B assertion** (when pattern given): `message.content` matches the provided regex pattern.

### Sub-role: `role=assistant`

**Maps to**: JSONL filter `select(.type == "assistant")` plus optional `contains` substring test.

**Required output fields per record**:
- `type` — always `"assistant"`
- `timestamp` — ISO8601 string
- `sessionId` — UUID string
- `message.content` — array of content blocks

**Filter semantics**: `contains` parameter performs case-insensitive substring match on the stringified content.

**Class A assertion**: Each record contains `timestamp`, `sessionId`, and `type: "assistant"`.

**Class B assertion** (when contains given): Stringified `message.content` contains the substring (case-insensitive).

### Sub-role: `role=tool`

**Maps to**: `handleQueryToolBlocks` — extracts individual tool blocks from assistant/user records.

When `block_type=tool_use`:
- Filter: `select(.type == "assistant")` then expand `message.content[]` where `.type == "tool_use"`
- Output includes outer context fields injected: `{timestamp, sessionId, turn}` merged with tool block fields

When `block_type=tool_result`:
- Filter: `select(.type == "user" and (.message.content | type == "array"))` then expand `message.content[]` where `.type == "tool_result"`
- Output includes outer context fields injected: `{timestamp, sessionId, turn}` merged with tool result block fields

**Required output fields per record** (block_type=tool_use):
- `type` — always `"tool_use"`
- `id` — tool call ID string
- `name` — tool name string
- `input` — tool input object
- `timestamp` — ISO8601 string (from outer record)
- `sessionId` — UUID string (from outer record)
- `turn` — integer (from outer record)

**Required output fields per record** (block_type=tool_result):
- `type` — always `"tool_result"`
- `tool_use_id` — matching tool call ID
- `content` — result content (string or array)
- `timestamp` — ISO8601 string (from outer record)
- `sessionId` — UUID string (from outer record)
- `turn` — integer (from outer record)

**Class A assertion**: Each record contains `type`, `id`/`tool_use_id`, `name`/`content`, `timestamp`, `sessionId`, and `turn`.

### Sub-role: `role=all`

**Maps to**: `handleQueryConversationFlow` — JSONL filter `select(.type == "user" or .type == "assistant")`

**Required output fields per record**:
- `type` — `"user"` or `"assistant"`
- `timestamp` — ISO8601 string
- `sessionId` — UUID string
- `message.role` — `"user"` or `"assistant"`

**Class A assertion**: Each record contains `timestamp`, `sessionId`, `type`, and `message.role`.

---

## Hybrid Output Mode

When results exceed 8KB inline threshold, the tool returns a `file_ref` instead of inline JSON:

```json
{
  "file_ref": "/tmp/meta-cc-output-XXXXXX.jsonl",
  "record_count": 150,
  "bytes": 45000
}
```

**Class A assertion for file_ref**: Response contains `file_ref` key pointing to a readable file path.

---

## Empty Results

When no records match the filter, the tool returns an empty result set:

```json
{
  "records": [],
  "total": 0
}
```

**Class A assertion for empty**: Response is valid JSON with `records` array (possibly empty) or `total: 0`. Tool does NOT return an error for empty results.

---

## See Also

- [Fixture Collection](fixtures/) — Individual fixture files
- [Oracle Runner](oracle-runner.md) — How to run fixtures against live MCP
- [MCP Query Tools Reference](../guides/mcp-query-tools.md) — Parameter reference
