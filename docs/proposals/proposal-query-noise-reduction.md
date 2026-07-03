# Proposal: Query Noise Reduction and Stats Fix

**Status**: Implemented (Phases 50–51)
**Phase**: 50-51
**Priority**: P0
**Date**: 2026-03-09

---

## Background

Analysis of historical user messages via `query_user_messages` reveals that approximately 66% of
returned records are Claude Code system-injected messages with no user intent content:

| Record type | Approximate share |
|-------------|-------------------|
| `<local-command-caveat>` wrappers | ~25% |
| `<command-name>/plugin</command-name>` slash command bodies | ~15% |
| `<local-command-stdout>` command output echo | ~15% |
| `<task-notification>` Task agent completion | ~8% |
| Context continuation injections | ~3% |
| **Real user intent messages** | **~34%** |

A recent project-scope query (`since=2026-03-08`) returned 222 records; only ~75 were genuine
user messages. The noise forces multiple Read operations, inflates context usage, and obscures
intent analysis.

Separately, `stats_first` / `stats_only` mode always reports `session_count: 0`, even when
multiple sessions are present.

---

## Root Cause Analysis

### P0-B: session_count always 0

**Call chain** (executor.go, Execute method):

```
Execute()
  line 299: applyMessageFiltersToData(queryResult.Entries, ...)  ← transforms IN PLACE
               when content_summary=true: sessionId → session_id (via ApplyContentSummary)
  buildResponse(queryResult, ...)
    buildStatsFirstResponse(parsedData=already-transformed)
      dataToJSONL(parsedData)                ← JSONL now has session_id, not sessionId
      GenerateTimestampStats(jsonlData)       ← checks obj["sessionId"] → always empty
                                               → session_count = 0
```

`adaptResponse` (called later inside the response builders) does **not** apply any message
filters — it is a pure serialization wrapper (inline vs. file_ref). There is no double-
application concern.

The bug affects both `stats_first` and `stats_only` modes when `content_summary=true`.

**Fix**: move `applyMessageFiltersToData` from line 299 into `buildResponse`, after the
stats-computation path, so stats always see the raw camelCase `sessionId` field.

---

## Proposed Changes

### Change 1 — `exclude_system_messages` parameter (Phase 50)

Add an optional `exclude_system_messages bool` parameter to `query_user_messages`
(default `false`, backward compatible).

When `true`, and `content_type` is `"string"` (the default), append a single jq `select`
clause using compound `or` that rejects entries whose `message.content` starts with any
system-injection sentinel prefix:

```jq
| select(
    .message.content | (
      startswith("<local-command-caveat>") or
      startswith("<command-name>") or
      startswith("<local-command-stdout>") or
      startswith("<task-notification>")
    ) | not
  )
```

When `content_type == "array"`, system injection is impossible (array-type entries are tool
results, never string-tagged messages), so `exclude_system_messages` is silently ignored.

**Implementation**:
- Extract `excludeSystem := getBoolParam(args, "exclude_system_messages", false)`
- If `excludeSystem && contentType == "string"` (or `contentType == ""`), append the compound
  `select | not` clause to `jqFilter`.

**Files touched**:
- `cmd/mcp-server/handlers_convenience.go` — extract param, build jq clause (~12 LoC)
- `cmd/mcp-server/tools.go` — add `exclude_system_messages` schema property (~8 LoC)
- `cmd/mcp-server/handlers_query_test.go` — integration tests: with/without flag, verify
  system messages excluded, verify backward compat (~45 LoC)

**Expected impact**: 222 records → ~75 records (−66%) in the representative query above.

---

### Change 2 — Fix `session_count: 0` stats bug (Phase 51)

**Fix location**: `executor.go`, `buildResponse` function.

**Concrete change**:

1. **Remove** lines 299–301 (the pre-`buildResponse` `applyMessageFiltersToData` call).

2. **In `buildResponse`** (which already receives `toolPipelineConfig`), restructure as:

```go
func (e *ToolExecutor) buildResponse(cfg *config.Config, result QueryResult,
    args map[string]interface{}, toolName string, pipeline toolPipelineConfig) (string, error) {

    rawData := result.Entries

    // stats_only: stats from raw data only, no detail rendering needed
    if pipeline.statsOnly {
        output, err = e.buildStatsOnlyResponse(rawData, toolName)
        // ... error handling ...
        return injectWarnings(output, result.Warnings)
    }

    // Apply message filters for detail rendering (AFTER stats path)
    parsedData := rawData
    if toolName == "query_user_messages" && pipeline.requiresMessageFilters() {
        parsedData = e.applyMessageFiltersToData(rawData,
            pipeline.maxMessageLength, pipeline.contentSummary)
    }

    if pipeline.statsFirst {
        // rawData → stats (sessionId preserved), parsedData → detail
        output, err = e.buildStatsFirstResponse(cfg, rawData, parsedData, args, toolName)
    } else {
        output, err = e.buildStandardResponse(cfg, parsedData, args, toolName)
    }
    // ... error handling + injectWarnings ...
}
```

3. **Update `buildStatsFirstResponse` signature**: add `rawData []interface{}` parameter
   (used for stats computation); existing `parsedData` parameter becomes the detail data.

```go
func (e *ToolExecutor) buildStatsFirstResponse(
    cfg *config.Config,
    rawData []interface{},      // for stats (camelCase sessionId preserved)
    parsedData []interface{},   // for detail (may be content_summary transformed)
    args map[string]interface{},
    toolName string,
) (string, error) {
    jsonlData, _ := e.dataToJSONL(rawData)    // ← was parsedData
    // ... stats computation ...
    response, _ := adaptResponse(cfg, parsedData, args, toolName)  // ← unchanged
    // ...
}
```

**Files touched**:
- `cmd/mcp-server/executor.go` — restructure `buildResponse`, update
  `buildStatsFirstResponse` signature (~30 LoC change)
- `cmd/mcp-server/executor_test.go` — tests:
  - `stats_first + content_summary` → `session_count > 0`
  - `stats_only + content_summary` → `session_count > 0`
  - `stats_first` without `content_summary` → unchanged behavior
  (~40 LoC)

---

## Impact Summary

| Change | Files | Estimated LoC |
|--------|-------|---------------|
| P50: exclude_system_messages | handlers_convenience.go, tools.go, handlers_query_test.go | ~65 |
| P51: session_count fix | executor.go, executor_test.go | ~70 |
| **Total** | 4 files | **~135** |

Both phases are under the 200 LoC per-stage and 500 LoC per-phase limits.

---

## Non-Goals

- Semantic classification of system messages using LLM (meta-cc is LLM-free).
- Filtering system messages on other tools (`query_conversation_flow`, etc.) — those tools
  return both user and assistant entries; the XML-tag system-injection pattern is specific to
  `query_user_messages`.
- Changing the default behavior of `exclude_system_messages` (stays `false` for backward
  compatibility).
- Filtering context-continuation messages ("This session is being continued…") — these
  contain useful conversation context and are not XML-tagged injections.

---

## Acceptance Criteria

### Phase 50
- `query_user_messages(pattern=".", exclude_system_messages=true)` returns no records whose
  `content` starts with `<local-command-caveat>`, `<command-name>`, `<local-command-stdout>`,
  or `<task-notification>`.
- `query_user_messages(pattern=".", exclude_system_messages=false)` (or omitted) returns the
  same results as before (backward compatible).
- `exclude_system_messages=true` with `content_type="array"` silently ignored (no error).
- `tools.go` schema lists `exclude_system_messages` with clear description.

### Phase 51
- `query_user_messages(pattern=".", stats_first=true, content_summary=true)` returns a stats
  header where `session_count` equals the actual number of distinct session IDs in the result.
- `query_user_messages(pattern=".", stats_only=true, content_summary=true)` likewise reports
  correct non-zero `session_count`.
- `stats_first=true` without `content_summary` continues to work correctly (no regression).
- `stats_only=true` without `content_summary` continues to work correctly (no regression).
