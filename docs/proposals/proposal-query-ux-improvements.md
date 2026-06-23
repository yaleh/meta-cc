# Proposal: Query UX Improvements — group_by_session, preview_length, context_turns, session-level stats

**Status**: Implemented (Phases 47–55; since/until/preview_length/group_by_session/stats_level/context_turns all in internal/mcp/)
**Phase**: 52–55
**Priority**: P1
**Date**: 2026-03-09

---

## Background

After the Phase 50–51 noise-reduction work, `query_user_messages` reliably returns genuine user
intent messages. The next friction layer is output structure: large multi-session result sets
arrive as flat interleaved JSONL, Chinese content previews are truncated to English-tuned
defaults, and there is no way to see the turns surrounding a match without a second query.

Four concrete pain points drove this proposal:

| # | Pain point | Current workaround |
|---|------------|--------------------|
| 1 | Multi-session results interleaved, hard to correlate | Manual jq `group_by(.sessionId)` or Read loop |
| 2 | `content_summary` preview truncates at 100 chars — ~30 Chinese chars lose intent | Use full content (large output) |
| 3 | Matched turn has no surrounding context; second query required | Run `query_conversation_flow` and join manually |
| 4 | `stats_only` outputs hourly turn buckets; no per-session match count or duration | Manual jq aggregation |

---

## Root Cause Analysis

### C1 — Flat JSONL for multi-session output

`buildStandardResponse` → `adaptResponse` serialises `parsedData []interface{}` as flat JSONL.
No grouping step exists in the pipeline. Session boundaries are implicit in the `sessionId`
field only.

### C2 — Hardcoded 100-char preview

`ApplyContentSummary` in `filters.go` calls `content[:DefaultPreviewLength]` where
`DefaultPreviewLength = 100` (package-level const). The constant is not exposed as a parameter.
Go's `len()` is byte-counted; a 100-byte limit yields ~33 Chinese characters, typically
truncating mid-thought.

### C3 — No context window

`handleQueryUserMessages` applies a jq `select(...)` filter and returns only matching turns.
There is no mechanism to load adjacent turns from the same session. The session file path
is available via workingDir/scope resolution but is not re-queried after the initial filter pass.

### C4 — No session-level stats

`GenerateTimestampStats` (Phase 49) counts turns per hour and computes `session_count`.
It does not aggregate per session: match count per session, first/last match timestamp,
or session duration. `stats_only=true` callers get a useful overview but cannot compare
activity across sessions without reading all detail records.

---

## Proposed Changes

### Change 1 — `group_by_session` parameter

**New parameter**: `group_by_session bool` (default `false`), `query_user_messages` only.

When `true`, the flat `parsedData []interface{}` slice is transformed before serialisation:
entries are grouped by their `sessionId` (or `session_id` if `content_summary` was applied),
producing one object per session:

```json
{
  "session_id": "d3ea683a-...",
  "match_count": 7,
  "first_match": "2026-03-08T09:14:00Z",
  "last_match":  "2026-03-09T11:30:00Z",
  "turns": [ /* matched turn objects */ ]
}
```

**Implementation location**: new function `GroupBySession(entries []interface{}) []interface{}`
in `internal/query/jq.go`. The function:
1. Iterates entries in order. Extracts session key by checking `session_id` (snake_case, post-summary)
   first, then falling back to `sessionId` (camelCase, raw). This handles both content_summary and
   non-summary pipelines.
2. Builds an ordered slice of session structs, preserving first-seen order.
3. For each session, records `min(timestamp)`, `max(timestamp)`, `len(turns)`, and the turns slice.
4. Returns `[]interface{}` of session objects, using `session_id` (snake_case) as the key name
   in output consistently.

Called from `buildResponse` in `executor.go`, after `applyMessageFiltersToData`, when
`pipeline.groupBySession == true`.

**Mutual exclusion check**: at the top of `buildResponse`, before any stats or grouping logic:
```go
if pipeline.groupBySession && pipeline.statsOnly {
    return "", fmt.Errorf("group_by_session and stats_only are mutually exclusive")
}
```

**Compatibility**:
- Default `false` — no change to existing output.
- Compatible with `content_summary` (uses `session_id` field from summary output).
- Compatible with `exclude_system_messages`.
- Incompatible with `stats_only` — when both are set, return error:
  `"group_by_session and stats_only are mutually exclusive"`.
- Compatible with `stats_first`: stats computed from rawData (before grouping), grouped
  detail appended after `---`.

**toolPipelineConfig change**: add `groupBySession bool` field.

---

### Change 2 — `preview_length` parameter

**New parameter**: `preview_length int` (default `100`), applies only when `content_summary=true`.

**Implementation location**: `cmd/mcp-server/filters.go`.

Current signature:
```go
func ApplyContentSummary(messages []interface{}) []interface{}
```

New signature:
```go
func ApplyContentSummary(messages []interface{}, previewLength int) []interface{}
```

Internal change: replace byte-indexed truncation with rune-safe truncation:
```go
// Current (UNSAFE for CJK):
preview = content[:DefaultPreviewLength]

// New (rune-safe):
runes := []rune(content)
if len(runes) > previewLength {
    preview = string(runes[:previewLength]) + "..."
}
```
Add guard: if `previewLength <= 0`, use `DefaultPreviewLength` (100) as fallback.

`applyMessageFiltersToData` in `executor.go` currently calls `ApplyContentSummary(messages)`.
Change to `ApplyContentSummary(messages, pipeline.previewLength)`.

**toolPipelineConfig change**: add `previewLength int` field, populated from
`getIntParam(args, "preview_length", DefaultPreviewLength)`.

**Schema note**: Chinese/Japanese/Korean characters occupy 3 bytes each in UTF-8, so callers
should set `preview_length=90` for ~30 readable CJK characters.

**Compatibility**: default `100` matches existing behaviour. `preview_length` without
`content_summary=true` is silently ignored.

---

### Change 3 — `context_turns` parameter

**New parameter**: `context_turns int` (default `0` = disabled), `query_user_messages` only,
string `content_type` only.

When `N > 0`, for each matched turn the response includes up to N turns before and N turns
after from the same session file. Context turns are marked `"context": true`; matched turns
`"context": false`.

**Implementation** (new method in `executor.go`):

```
expandContextTurns(rawData []interface{}, N int, baseDir string) ([]interface{}, error)
```

`baseDir` is extracted from `args["working_dir"]` (or resolved via `getQueryBaseDir`) in
`buildResponse` before calling this method — same resolution used in `Execute()`.

Steps:
1. Collect distinct session IDs from `rawData` (matched turns), building a `matchedUUIDs` set.
2. For each session ID, scan files in `baseDir` matching `*.jsonl`, reading each file to find
   turns with matching `sessionId`. This is necessary because no sessionId→filename helper
   currently exists; a new helper `loadTurnsForSession(baseDir, sessionID string)
   ([]interface{}, error)` will be added to `handlers_query.go`.
3. For each matched turn (identified by `uuid`), find its index in the full session turn list.
4. Collect indices `[max(0, idx-N) .. min(len-1, idx+N)]`.
5. Mark each turn: `"context": false` if its uuid is in `matchedUUIDs`, else `"context": true`.
6. Deduplicate: if two matched turns' windows overlap, shared turns appear once; a turn that
   is itself a match keeps `"context": false`.
7. Maintain chronological order within each session.

**New helper** `loadTurnsForSession(baseDir, sessionID string) ([]interface{}, error)` in
`handlers_query.go`: reads all JSONL files in `baseDir`, returns turns where
`obj["sessionId"] == sessionID`. Stops after first file that yields results (sessions are
typically in a single file).

**Only applies to string `content_type`**. When `content_type="array"`, `context_turns` is
silently ignored.

**Interaction with `group_by_session`**: when both are set, context turns are included in the
session's `turns` array marked with `"context": true`. The session-level `match_count` counts
only `"context": false` turns.

**toolPipelineConfig change**: add `contextTurns int` field.

---

### Change 4 — Session-level stats aggregation

**New parameter**: `stats_level string` (values: `"turn"` default, `"session"`).
Applies to `stats_only` and `stats_first` modes of `query_user_messages`.

When `stats_level="session"`, replace per-hour bucket output with per-session lines:

```jsonl
{"total_sessions": 4, "total_matches": 38, "time_range": {"from": "...", "to": "..."}}
{"session_id": "d3ea683a-...", "match_count": 12, "first_match": "2026-03-08T09:14Z", "last_match": "2026-03-09T11:30Z", "duration_minutes": 1576}
{"session_id": "a1b2c3d4-...", "match_count": 8, "first_match": "2026-03-08T14:02Z", "last_match": "2026-03-08T18:44Z", "duration_minutes": 282}
```

**Implementation**: new function in `internal/query/jq.go`:

```go
func GenerateSessionStats(jsonlData string) (string, error)
```

Logic mirrors `GenerateTimestampStats` but groups by `sessionId` instead of hour. For each
session: collect all timestamps, compute `min` (first_match), `max` (last_match), count
(match_count), and `duration_minutes = int(max.Sub(min).Minutes())`. Summary line:
`total_sessions`, `total_matches`, overall `time_range`.

**Critical**: `GenerateSessionStats` must always receive raw JSONL with camelCase `sessionId`
(i.e., from `rawData`, not `parsedData`). This mirrors the existing pattern in
`buildStatsFirstResponse` (phase 51 fix) where stats always operate on `rawData`.
Field lookup: `obj["sessionId"]` — same as `GenerateTimestampStats`.

Called from `buildStatsOnlyResponse` and `buildStatsFirstResponse` when
`pipeline.statsLevel == "session"` (and `toolName == "query_user_messages"`).

**toolPipelineConfig change**: add `statsLevel string` populated from
`getStringParam(args, "stats_level", "turn")`.

**Validation**: if `stats_level` is neither `"turn"` nor `"session"`, return error
`"invalid stats_level: must be 'turn' or 'session'"`.

**Backward compatibility**: default `"turn"` produces identical output to current behaviour.

---

## Files Affected

| File | Changes |
|------|---------|
| `cmd/mcp-server/filters.go` | `ApplyContentSummary` signature: add `previewLength int`; replace hardcoded constant |
| `cmd/mcp-server/tools.go` | Add 4 schema properties to `query_user_messages` |
| `internal/query/jq.go` | Add `GroupBySession()`, `GenerateSessionStats()` |
| `cmd/mcp-server/executor.go` | `toolPipelineConfig`: 4 new fields; `buildResponse`: grouping + context expansion; stats dispatch for `stats_level` |

---

## Estimated Impact

| Change | Key files | Est. LoC |
|--------|-----------|----------|
| Phase 52: preview_length | filters.go, executor.go, tools.go | ~25 |
| Phase 53: group_by_session | jq.go, executor.go, tools.go | ~80 |
| Phase 54: session-level stats | jq.go, executor.go, tools.go | ~70 |
| Phase 55: context_turns | executor.go, handlers_query.go, tools.go | ~200 |
| Tests | *_test.go | ~200 |
| **Total** | 6 files | **~495** |

---

## Backward Compatibility

All parameters are optional with defaults reproducing current behaviour:

| Parameter | Default | Current behaviour reproduced |
|-----------|---------|------------------------------|
| `group_by_session` | `false` | Yes — flat JSONL unchanged |
| `preview_length` | `100` | Yes — `DefaultPreviewLength` const unchanged |
| `context_turns` | `0` | Yes — no context expansion |
| `stats_level` | `"turn"` | Yes — `GenerateTimestampStats` called as before |

No existing tests should break.

---

## Non-Goals

- **Sequence detection** (ordered multi-pattern matching across turns) — deferred.
- **Applying `group_by_session` to tools other than `query_user_messages`** — deferred.
- **Configurable stats granularity** (`stats_interval=15m`) — deferred.
- **LLM-assisted context summarisation** — meta-cc is LLM-free; context is raw turns only.

---

## Acceptance Criteria

### preview_length
- `content_summary=true, preview_length=30` → `content_preview` ≤ 30 bytes.
- Omitting `preview_length` with `content_summary=true` → `content_preview` ≤ 100 bytes (unchanged).
- `preview_length` without `content_summary=true` → no effect, no error.

### group_by_session
- `group_by_session=true` → one object per distinct `sessionId`, each with `session_id`,
  `match_count`, `first_match`, `last_match`, `turns`.
- `match_count == len(turns)` for each session.
- `group_by_session=true` + `stats_only=true` → explicit error.
- `group_by_session=true` + `content_summary=true` → each turn object in `turns` is a summary object.

### context_turns
- `context_turns=2` → matched turns plus up to 2 preceding and 2 following turns per match,
  each annotated `"context": true/false`.
- Overlapping windows produce no duplicate turns.
- `context_turns` + `content_type="array"` → silently ignored, no error.

### stats_level="session"
- `stats_only=true, stats_level="session"` → summary line with `total_sessions`/`total_matches`,
  then one line per session with `session_id`, `match_count`, `first_match`, `last_match`,
  `duration_minutes`.
- `stats_level="turn"` (or omitted) → identical to current `stats_only=true` output.
- `stats_level="invalid"` → error.
