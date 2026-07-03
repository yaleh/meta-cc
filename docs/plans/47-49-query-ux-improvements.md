# Plan 47–49: Query UX Improvements

**Status**: Completed (since/until, group_by_session, stats_level all implemented in internal/mcp/; see audit 2026-06-23)
**Date**: 2026-03-09
**Proposal**: [docs/proposals/proposal-query-ux-improvements.md](../proposals/proposal-query-ux-improvements.md)

---

## Overview

Three independent improvements to `query_user_messages` and related tools, each self-contained and additive (no breaking changes).

| # | Change | Severity | Phase |
|---|--------|----------|-------|
| C1 | `since`/`until` Go-level time filter | High | 47 |
| C2 | `session_id` in `content_summary` output | Medium | 48 |
| C3 | Tool-aware stats for `stats_first`/`stats_only` | Medium | 49 |

**Development methodology**: TDD throughout. Each stage begins with failing tests; implementation follows.

**Code limits**: Phase ≤500 lines, Stage ≤200 lines.

**Stage independence**: All three phases are independent and can be executed in parallel.

---

## Phase 47: Native `since` / `until` Time Parameters

**Goal**: Add `since` and `until` parameters to `query_user_messages` (and `query_timestamps`, `query_conversation_flow`) that filter by `timestamp` in Go before jq execution, replacing the unreliable jq string comparison workaround.

**Estimated code**: ~120–160 lines (tests + implementation)

**Files touched**:
- `cmd/mcp-server/handlers_query.go` — add `TimeRange` struct; add `executeQueryWithTimeRange`
- `cmd/mcp-server/query_executor.go` — propagate `TimeRange` through `streamFiles()` and `processFile()`; apply Go-level filter there
- `cmd/mcp-server/handlers_convenience.go` — extract and parse `since`/`until` in affected handlers
- `cmd/mcp-server/tools.go` — declare new parameters in schema for affected tools
- `cmd/mcp-server/handlers_query_test.go` — new time-filter tests (do **not** use `handlers_convenience_test.go`: that file currently skips all tests with `t.Skip()`)

### Stage Dependencies

```
47.1 (core time filter in executeQuery + query_executor.go)
  └─► 47.2 (wire into handlers + schema + error handling)
```

47.1 is the foundation; 47.2 depends on it. Error handling for malformed values is part of 47.2 (TDD: write error tests before implementing the parse step).

---

### Stage 47.1 — Time Filter in `executeQuery`

**Problem**: `executeQuery()` has no time-awareness; callers cannot filter by time without jq string comparison.

**TDD sequence**:

1. Write `TestExecuteQueryTimeFilter` in `handlers_query_test.go`:
   - Create fixture with 3 JSONL entries: timestamps `T-2h`, `T-1h`, `T+0`.
   - Assert `executeQueryWithTimeRange(scope, filter, limit, workingDir, since=T-1.5h, until=nil)` returns exactly 2 entries (`T-1h`, `T+0`).
   - Assert `executeQueryWithTimeRange(..., since=nil, until=T-0.5h)` returns exactly 2 entries (`T-2h`, `T-1h`).
   - Assert both `since` and `until` set returns only the matching window.
   - Run — expect FAIL.
2. In `handlers_query.go`, define:
   ```go
   type TimeRange struct {
       Since *time.Time
       Until *time.Time
   }
   ```
3. In `handlers_query.go`, implement `executeQueryWithTimeRange(scope, jqFilter string, limit int, workingDir string, tr TimeRange)`. This function passes `tr` into a new `streamFilesWithTimeRange()` call (see step 4).
4. In `query_executor.go`, extend `streamFiles()` to accept `TimeRange` (rename to `streamFilesWithTimeRange` or add an options struct). Inside `processFile()`, after parsing each JSONL line to a `map[string]interface{}` (current line ~175), read the `timestamp` field, parse with `time.Parse(time.RFC3339, ts)`, and skip the entry if outside the time range. Unparseable timestamps are non-fatal: include the entry.
5. Keep existing `executeQuery()` as a thin wrapper calling `executeQueryWithTimeRange` with zero `TimeRange`.
5. Run tests — expect PASS.
6. Run `make commit`.

**Acceptance criteria**:
- `executeQueryWithTimeRange` filters entries by timestamp in Go (not jq).
- Entries whose `timestamp` cannot be parsed are included (not silently dropped) — parse errors in individual entries are non-fatal.
- `executeQuery()` signature unchanged.

---

### Stage 47.2 — Wire Parameters into Handlers and Schema

**Problem**: Handlers don't extract `since`/`until`; tool schema doesn't declare them.

**TDD sequence**:

1. Write `TestHandleQueryUserMessagesSince` in `handlers_query_test.go` (not `handlers_convenience_test.go` — that file skips all tests):
   - Call `handleQueryUserMessages` with `args["since"] = "2026-03-07T00:00:00Z"` and a fixture spanning 3 days.
   - Assert only entries `>= 2026-03-07T00:00:00Z` are returned.
   - Also write `TestHandleQueryUserMessagesBadSince`: pass `args["since"] = "2026-03-07"` (not RFC3339), assert error contains `"invalid since value"` and the original value. Pass `"not-a-date"`, assert same.
   - Run — expect FAIL.
2. In `handleQueryUserMessages()`, extract:
   ```go
   sinceStr := getStringParam(args, "since", "")
   untilStr := getStringParam(args, "until", "")
   ```
   Parse each non-empty value with `time.Parse(time.RFC3339, ...)`. On parse error, return `fmt.Errorf("invalid since value %q: must be RFC3339 (e.g. 2026-03-07T00:00:00Z)", sinceStr)` immediately.
   Build `TimeRange` and pass to `executeQueryWithTimeRange`.
3. Apply the same pattern to `handleQueryTimestamps` and `handleQueryConversationFlow`.
4. In `tools.go`, add `since` and `until` parameters to the schema of the three affected tools:
   ```go
   "since": {Type: "string", Description: "Include only records with timestamp >= this value (RFC3339, e.g. \"2026-03-07T00:00:00Z\")"},
   "until": {Type: "string", Description: "Include only records with timestamp < this value (RFC3339)"},
   ```
5. Run tests — expect PASS.
6. Run `make commit`.

**Acceptance criteria**:
- `query_user_messages(since: "2026-03-07T00:00:00Z")` returns only records with `timestamp >= 2026-03-07T00:00:00Z`.
- Schema declares `since` and `until` for the three affected tools.

---

## Phase 48: `session_id` in `content_summary` Output

**Goal**: Include `session_id` (sourced from raw `.sessionId`) in the `content_summary: true` output, enabling callers to group summary results by session without fetching full output.

**Estimated code**: ~30–50 lines (tests + implementation)

**Diagnosis**: `sessionId` is already present in every raw JSONL record. The field is stripped when building the `content_summary` projection. No parser or pipeline changes needed.

**Files touched**:
- `cmd/mcp-server/filters.go` — `ApplyContentSummary()` function (lines ~91–123): add `session_id` to the summary map
- `cmd/mcp-server/filters_test.go` — extend `TestApplyContentSummary` and `TestApplyContentSummary_TurnSequenceAndUUID`
- `cmd/mcp-server/tools.go` — update `content_summary` parameter description

### Stage 48.1 — Add `session_id` to `content_summary` Projection

**TDD sequence**:

1. The projection is built in `filters.go:ApplyContentSummary()`. The current summary map has exactly 4 keys: `turn_sequence` (array index), `uuid`, `timestamp`, `content_preview`. The raw entry is available as a `map[string]interface{}` at the point of construction, so `sessionId` is accessible as `entry["sessionId"]`.
2. Write `TestContentSummaryIncludesSessionID` in `filters_test.go`:
   - Provide fixture JSONL with `sessionId: "abc-123"` on each entry.
   - Call the tool with `content_summary: true`.
   - Assert every output record contains `"session_id": "abc-123"`.
   - Run — expect FAIL.
3. In the projection code, add `session_id` field sourced from the raw entry's `sessionId` key.
4. Run tests — expect PASS.
5. Update `content_summary` parameter description in `tools.go` to mention the included fields: `"Return only session_id/turn/timestamp/preview (100 chars), skip full content."`.
6. Run `make commit`.

**Acceptance criteria**:
- Every record in `content_summary: true` output includes `session_id`.
- Value matches the `sessionId` field from the raw JSONL entry.
- Existing 4 fields (`turn_sequence`, `uuid`, `timestamp`, `content_preview`) are unchanged.
- `TestApplyContentSummary_Immutability` (existing) continues to pass — the original entry must not be mutated.

**Note on documentation**: Update `query_user_messages` tool description to explicitly list `.sessionId` as an available field for use in `jq_filter` expressions.

---

## Phase 49: Tool-Aware Stats for `stats_first` / `stats_only`

**Goal**: Replace the meaningless `{"key": "unknown", "count": N}` stats output for `query_user_messages` with time-bucketed stats (by hour), while leaving stats for tool-call tools unchanged.

**Estimated code**: ~80–110 lines (tests + implementation)

**Root cause**: `GenerateStats()` in `internal/query/jq.go` groups by `tool`/`ToolName` field. User message records have neither field, so all land in the `"unknown"` bucket. The `executor.go` `buildStatsFirstResponse()` calls `GenerateStats()` without knowing the tool type.

**Approach**: Add `GenerateTimestampStats()` in `jq.go`. In `executor.go`, dispatch to the correct stats function based on `toolName`.

**Files touched**:
- `internal/query/jq.go` — new `GenerateTimestampStats()` function
- `internal/query/jq_test.go` — tests for `GenerateTimestampStats`
- `cmd/mcp-server/executor.go` — conditional dispatch in `buildStatsFirstResponse` and `buildStatsOnlyResponse`

### Stage Dependencies

```
49.1 (GenerateTimestampStats)
  └─► 49.2 (dispatch in executor.go)
```

---

### Stage 49.1 — `GenerateTimestampStats()` in `jq.go`

**TDD sequence**:

1. Write `TestGenerateTimestampStats` in `jq_test.go`:
   - Input: 5 JSONL records with timestamps spanning 3 hours (2 in hour A, 2 in hour B, 1 in hour C).
   - Assert output contains: `{"total": 5, "session_count": N, "time_range": {"from": ..., "to": ...}}` followed by one line per non-empty hour.
   - Assert each hour line: `{"hour": "2026-03-09T06", "count": 2}`.
   - Run — expect FAIL.
2. Implement `GenerateTimestampStats(jsonlData string) (string, error)`:
   - Parse each line's `timestamp` with `time.Parse(time.RFC3339, ...)`. Skip lines with unparseable timestamps (non-fatal).
   - Bucket by hour: `ts.UTC().Format("2006-01-02T15")`.
   - Count distinct `sessionId` values.
   - Find min/max timestamps.
   - Output: first line is summary JSON `{total, session_count, time_range}`, subsequent lines are `{hour, count}` sorted chronologically.
3. Run tests — expect PASS.
4. Run `make commit`.

**Acceptance criteria**:
- `GenerateTimestampStats` produces time-bucketed output.
- Hours with zero records are omitted.
- Summary line (total, session_count, time_range) is the first output line.
- `GenerateStats()` is unchanged.

---

### Stage 49.2 — Dispatch in `executor.go`

**TDD sequence**:

1. Write `TestStatsOnlyDispatch` in `executor_test.go`:
   - For `toolName = "query_user_messages"` with user message fixture: assert stats output contains `"hour"` key (not `"key": "unknown"`).
   - For `toolName = "query_tool_errors"` with tool call fixture: assert stats output contains `"key"` field (existing behavior).
   - Run — expect FAIL.
2. Define the set of tools that should use timestamp-based stats:
   ```go
   var timestampStatsTools = map[string]bool{
       "query_user_messages":    true,
       "query_conversation_flow": true,
       "query_timestamps":       true,
       "query_summaries":        true,
   }
   ```
3. In `buildStatsOnlyResponse()` and `buildStatsFirstResponse()`, replace the unconditional `querypkg.GenerateStats(jsonlData)` call:
   ```go
   if timestampStatsTools[toolName] {
       output, err = querypkg.GenerateTimestampStats(jsonlData)
   } else {
       output, err = querypkg.GenerateStats(jsonlData)
   }
   ```
4. Run tests — expect PASS.
5. Run `make commit`.

**Acceptance criteria**:
- `query_user_messages(stats_only: true)` returns time-bucketed stats with `total`, `session_count`, `time_range`, and per-hour counts.
- `query_tool_errors(stats_only: true)` returns tool-name-grouped stats (existing behavior).
- The `timestampStatsTools` map is the single place to update when adding new tools.

---

## Verification

After all three phases, run the following to confirm no regressions:

```bash
make commit    # format + build + tests
make push      # full lint + coverage check
```

Manual smoke tests (using meta-cc MCP):

```javascript
// C1: time filter
query_user_messages({since: "2026-03-09T00:00:00Z"})         // should return only today's messages
query_user_messages({since: "not-a-date"})                   // should return error, not empty

// C2: session_id in summary
query_user_messages({content_summary: true, limit: 3})       // every record should have session_id

// C3: timestamp stats
query_user_messages({stats_only: true})                      // should show hourly buckets, not "unknown"
query_tool_errors({stats_only: true})                        // should still show tool-name grouping
```
