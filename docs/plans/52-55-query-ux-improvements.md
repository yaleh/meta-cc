# Plan: Query UX Improvements (Phases 52–55)

**Proposal**: docs/proposals/proposal-query-ux-improvements.md
**Status**: Completed (preview_length, group_by_session, stats_level, context_turns all implemented; see audit 2026-06-23)
**Date**: 2026-03-09

---

## Overview

Four independent enhancements to `query_user_messages`, each as its own phase. Ordered from
simplest to most complex. Each phase = Stage 1 (failing tests) + Stage 2 (implementation).

| Phase | Feature | Key files | Est. LoC |
|-------|---------|-----------|----------|
| 52 | `preview_length` | filters.go, executor.go, tools.go | ~40 |
| 53 | `group_by_session` | jq.go, executor.go, tools.go | ~110 |
| 54 | `stats_level="session"` | jq.go, executor.go, tools.go | ~90 |
| 55 | `context_turns` | executor.go, handlers_query.go, tools.go | ~220 |

---

## Phase 52: `preview_length` parameter

### Stage 52.1 — Failing tests

**Note on existing tests**: `ApplyContentSummary` has 4 existing test call sites in
`filters_test.go` (lines ~273, ~343, ~420, ~527). When the signature changes in Stage 52.2,
these tests will also need updating to pass `previewLength=100` (preserving current behaviour).
Update them as part of Stage 52.2, not here.

**File**: `cmd/mcp-server/filters_test.go`

Add tests before implementation:

```
TestApplyContentSummaryPreviewLength
  - previewLength=30: content_preview ≤ 30 runes
  - previewLength=0: falls back to DefaultPreviewLength (100)
  - previewLength=-1: falls back to DefaultPreviewLength
  - ASCII content, previewLength=5: truncates correctly
  - CJK content (Chinese): 3-byte chars, previewLength=10 → exactly 10 Chinese chars, no mid-rune cut
  - content shorter than previewLength: returned in full, no "..."

TestApplyContentSummaryPreviewLength_Default (regression)
  - calling ApplyContentSummary with previewLength=100 → identical output to current behaviour
```

**File**: `cmd/mcp-server/executor_test.go` (integration)

```
TestPreviewLengthParameter
  - ExecuteTool("query_user_messages", {pattern: ".", content_summary: true, preview_length: 20})
  - Assert all content_preview fields ≤ 20 runes
  - ExecuteTool with content_summary=false, preview_length=20: assert no error, full content returned
```

Run `make dev` — tests must FAIL at this point.

### Stage 52.2 — Implementation

**`cmd/mcp-server/filters.go`**:

1. Change `ApplyContentSummary` signature:
   ```go
   // Before:
   func ApplyContentSummary(messages []interface{}) []interface{}
   // After:
   func ApplyContentSummary(messages []interface{}, previewLength int) []interface{}
   ```

2. Inside `ApplyContentSummary`, the current truncation is at lines 106–107:
   ```go
   // Current (byte-indexed, UNSAFE for CJK):
   if len(content) > DefaultPreviewLength {
       preview = content[:DefaultPreviewLength] + "..."
   }
   ```
   Replace with rune-safe truncation:
   ```go
   if previewLength <= 0 {
       previewLength = DefaultPreviewLength
   }
   runes := []rune(content)
   if len(runes) > previewLength {
       preview = string(runes[:previewLength]) + "..."
   } else {
       preview = content
   }
   ```
   Note: `DefaultPreviewLength` const at line 8 remains unchanged.

**`cmd/mcp-server/executor.go`**:

1. Add field to `toolPipelineConfig` struct (after `contentSummary bool`):
   ```go
   previewLength    int
   ```

2. In `newToolPipelineConfig`, add:
   ```go
   previewLength: getIntParam(args, "preview_length", DefaultPreviewLength),
   ```

3. In `applyMessageFiltersToData`, change call:
   ```go
   // Before:
   ApplyContentSummary(messages)
   // After:
   ApplyContentSummary(messages, c.previewLength)
   ```
   Note: `applyMessageFiltersToData` is a method on `ToolExecutor` — find the call to
   `filters.ApplyContentSummary` and update its signature.

**`cmd/mcp-server/tools.go`**:

Add `preview_length` property to `query_user_messages` schema (after `content_summary`):
```json
"preview_length": {
  "type": "integer",
  "description": "Max characters per content_preview when content_summary=true (default: 100). Uses rune (character) count, not bytes — for CJK content use preview_length=30 for ~30 readable characters."
}
```

Run `make commit` — all tests must pass.

---

## Phase 53: `group_by_session` parameter

### Stage 53.1 — Failing tests

**File**: `internal/query/jq_test.go`

```
TestGroupBySession_Basic
  - Input: 4 entries, 2 distinct sessionIds (sess-A x2, sess-B x2)
  - Output: 2 session objects, each with session_id, match_count, first_match, last_match, turns
  - sess-A.match_count == 2, sess-B.match_count == 2

TestGroupBySession_OrderPreserved
  - Input: entries interleaved [sess-A, sess-B, sess-A, sess-B]
  - Output: sess-A first (first-seen order), sess-B second

TestGroupBySession_SnakeCaseSessionId
  - Input: entries with snake_case "session_id" (post-content_summary)
  - Output: correctly grouped, session_id field in output

TestGroupBySession_CamelCaseSessionId
  - Input: entries with camelCase "sessionId" (raw)
  - Output: correctly grouped, session_id field in output (normalised to snake_case in output)

TestGroupBySession_SingleSession
  - Input: all entries have same sessionId
  - Output: 1 session object with match_count = len(input)
```

**File**: `cmd/mcp-server/executor_test.go`

```
TestGroupBySession_Integration
  - ExecuteTool("query_user_messages", {pattern: ".", group_by_session: true})
  - Assert output is session objects (has "session_id" key), not flat turns

TestGroupBySession_MutualExclusionWithStatsOnly
  - ExecuteTool with {group_by_session: true, stats_only: true}
  - Assert error message contains "mutually exclusive"

TestGroupBySession_WithContentSummary
  - ExecuteTool with {group_by_session: true, content_summary: true}
  - Assert session objects present; each "turns" array contains summary objects (has content_preview)

TestGroupBySession_WithStatsFirst
  - ExecuteTool with {group_by_session: true, stats_first: true}
  - Assert stats header present (session_count, total)
  - Assert grouped detail after "---" separator
```

Run `make dev` — tests must FAIL.

### Stage 53.2 — Implementation

**`internal/query/jq.go`**:

Add `GroupBySession` function:
```go
type sessionGroup struct {
    SessionID  string
    MatchCount int
    FirstMatch string
    LastMatch  string
    Turns      []interface{}
}

func GroupBySession(entries []interface{}) []interface{} {
    var order []string
    groups := make(map[string]*sessionGroup)

    for _, entry := range entries {
        obj, ok := entry.(map[string]interface{})
        if !ok { continue }

        // Support both camelCase (raw) and snake_case (post-summary)
        sessionID, _ := obj["session_id"].(string)
        if sessionID == "" {
            sessionID, _ = obj["sessionId"].(string)
        }
        if sessionID == "" { sessionID = "unknown" }

        ts, _ := obj["timestamp"].(string)

        if _, exists := groups[sessionID]; !exists {
            order = append(order, sessionID)
            groups[sessionID] = &sessionGroup{SessionID: sessionID, FirstMatch: ts, LastMatch: ts}
        }
        g := groups[sessionID]
        g.MatchCount++
        g.Turns = append(g.Turns, entry)
        if ts < g.FirstMatch { g.FirstMatch = ts }
        if ts > g.LastMatch  { g.LastMatch  = ts }
    }

    result := make([]interface{}, 0, len(order))
    for _, id := range order {
        g := groups[id]
        result = append(result, map[string]interface{}{
            "session_id":  g.SessionID,
            "match_count": g.MatchCount,
            "first_match": g.FirstMatch,
            "last_match":  g.LastMatch,
            "turns":       g.Turns,
        })
    }
    return result
}
```

**`cmd/mcp-server/executor.go`**:

1. Add field to `toolPipelineConfig`:
   ```go
   groupBySession bool
   ```

2. In `newToolPipelineConfig`:
   ```go
   groupBySession: getBoolParam(args, "group_by_session", false),
   ```

3. In `buildResponse` at line 321, after the existing `pipeline.statsOnly` branch, add mutual
   exclusion check:
   ```go
   if pipeline.groupBySession && pipeline.statsOnly {
       return "", fmt.Errorf("group_by_session and stats_only are mutually exclusive")
   }
   ```

4. In `buildResponse`, after `applyMessageFiltersToData` (line ~334) and before `buildStatsFirstResponse`/
   `buildStandardResponse`, add grouping:
   ```go
   if pipeline.groupBySession && toolName == "query_user_messages" {
       parsedData = jq.GroupBySession(parsedData)
   }
   ```

**`cmd/mcp-server/tools.go`**:

Add schema property:
```json
"group_by_session": {
  "type": "boolean",
  "description": "Group results by session. Returns one object per session with session_id, match_count, first_match, last_match, and turns array. Mutually exclusive with stats_only."
}
```

Run `make commit`.

---

## Phase 54: `stats_level="session"` parameter

### Stage 54.1 — Failing tests

**File**: `internal/query/jq_test.go`

```
TestGenerateSessionStats_Basic
  - Input: JSONL with 2 sessions (sess-A x3 turns, sess-B x2 turns), timestamps spanning 10min each
  - Output line 1: {"total_sessions":2,"total_matches":5,"time_range":{...}}
  - Output line 2: sess-A, match_count=3, duration_minutes=10
  - Output line 3: sess-B, match_count=2, duration_minutes=10

TestGenerateSessionStats_SingleTurnSession
  - Session with only 1 turn: duration_minutes=0

TestGenerateSessionStats_OrderByFirstMatch
  - Sessions ordered by first_match timestamp ascending
```

**File**: `cmd/mcp-server/executor_test.go`

```
TestStatsLevelSession_StatsOnly
  - ExecuteTool with {pattern: ".", stats_only: true, stats_level: "session"}
  - Assert output has "total_sessions" key (not hour buckets)
  - Assert per-session lines have session_id, match_count, duration_minutes

TestStatsLevelSession_StatsFirst
  - ExecuteTool with {pattern: ".", stats_first: true, stats_level: "session"}
  - Assert stats header uses session aggregation
  - Assert detail records follow after "---"

TestStatsLevelTurn_Regression
  - ExecuteTool with {pattern: ".", stats_only: true} (no stats_level)
  - Assert output identical to current: hour buckets, session_count/total/time_range header

TestStatsLevelInvalid
  - ExecuteTool with {stats_level: "invalid"}
  - Assert error contains "must be 'turn' or 'session'"
```

Run `make dev` — tests must FAIL.

### Stage 54.2 — Implementation

**`internal/query/jq.go`**:

Add `GenerateSessionStats` function (mirrors `GenerateTimestampStats`):
```go
func GenerateSessionStats(jsonlData string) (string, error) {
    type sessionAgg struct {
        SessionID string
        Count     int
        First     time.Time
        Last      time.Time
    }
    var order []string
    sessions := make(map[string]*sessionAgg)
    var overall struct{ First, Last time.Time }
    firstOverall := true

    for _, line := range splitJSONLLines(jsonlData) {
        var obj map[string]interface{}
        if err := json.Unmarshal([]byte(line), &obj); err != nil { continue }

        sessionID, _ := obj["sessionId"].(string)  // always camelCase (rawData)
        tsStr, _ := obj["timestamp"].(string)
        ts, err := time.Parse(time.RFC3339, tsStr)
        if err != nil { ts, err = time.Parse("2006-01-02T15:04:05.000Z", tsStr) }
        if err != nil { continue }

        if _, exists := sessions[sessionID]; !exists {
            order = append(order, sessionID)
            sessions[sessionID] = &sessionAgg{SessionID: sessionID, First: ts, Last: ts}
        }
        s := sessions[sessionID]
        s.Count++
        if ts.Before(s.First) { s.First = ts }
        if ts.After(s.Last)  { s.Last  = ts }

        if firstOverall || ts.Before(overall.First) { overall.First = ts; firstOverall = false }
        if ts.After(overall.Last) { overall.Last = ts }
    }
    // Serialise: summary line + per-session lines ordered by first_match
    // ... (sort order, json.Marshal)
}
```

**`cmd/mcp-server/executor.go`**:

1. Add to `toolPipelineConfig`:
   ```go
   statsLevel string  // "turn" (default) or "session"
   ```

2. In `newToolPipelineConfig`:
   ```go
   statsLevel: getStringParam(args, "stats_level", "turn"),
   ```

3. In `buildResponse`, validate:
   ```go
   if pipeline.statsLevel != "" && pipeline.statsLevel != "turn" && pipeline.statsLevel != "session" {
       return "", fmt.Errorf("invalid stats_level: must be 'turn' or 'session'")
   }
   ```

4. In `buildStatsOnlyResponse`, dispatch:
   ```go
   // existing: timestampStatsTools dispatch
   // add: if pipeline.statsLevel == "session" && toolName == "query_user_messages"
   //      → call GenerateSessionStats(jsonlData)
   ```
   Note: `buildStatsOnlyResponse` currently takes `parsedData []interface{}`. It must use
   `rawData` for stats. Verify the call site passes `rawData` (confirmed in executor.go from
   phase 51 fix).

**`cmd/mcp-server/tools.go`**:

```json
"stats_level": {
  "type": "string",
  "description": "Aggregation level for stats_only/stats_first: 'turn' (default, hourly buckets) or 'session' (per-session match_count and duration)."
}
```

Run `make commit`.

---

## Phase 55: `context_turns` parameter

### Stage 55.1 — Failing tests

**File**: `cmd/mcp-server/handlers_query_test.go`

```
TestLoadTurnsForSession_Basic
  - Create temp JSONL file with 5 turns (3 for sess-A, 2 for sess-B)
  - loadTurnsForSession(dir, "sess-A") → 3 turns
  - loadTurnsForSession(dir, "sess-B") → 2 turns
  - loadTurnsForSession(dir, "sess-X") → 0 turns, nil error

TestLoadTurnsForSession_MultipleFiles
  - Create 2 temp JSONL files, each with different sessions
  - loadTurnsForSession finds correct session across files
```

**File**: `cmd/mcp-server/executor_test.go`

```
TestContextTurns_Basic
  - Session with 5 turns; match turn 3
  - context_turns=1 → turns 2, 3, 4 returned; turn 3 has "context":false, others "context":true

TestContextTurns_BoundaryStart
  - Match turn 0 (first in session), context_turns=2 → turns 0,1,2 returned (no negative index)

TestContextTurns_BoundaryEnd
  - Match turn 4 (last in session), context_turns=2 → turns 2,3,4 returned

TestContextTurns_OverlappingWindows
  - Session with 10 turns; matches at turns 2 and 4, context_turns=2
  - Window for turn 2: [0,1,2,3,4]; window for turn 4: [2,3,4,5,6]
  - Union: [0,1,2,3,4,5,6] — no duplicates
  - Turns 2 and 4 have "context":false; others "context":true

TestContextTurns_ArrayContentType
  - ExecuteTool with {content_type: "array", context_turns: 2}
  - Assert no error, context field NOT present (silently ignored)

TestContextTurns_WithGroupBySession
  - ExecuteTool with {context_turns: 1, group_by_session: true}
  - Assert session objects present; turns array contains context turns with "context" field
  - match_count counts only "context":false turns

TestContextTurns_Zero_NoEffect
  - ExecuteTool with {context_turns: 0}
  - Assert output identical to without context_turns
```

Run `make dev` — tests must FAIL.

### Stage 55.2 — Implementation

**`cmd/mcp-server/handlers_query.go`**:

Add `loadTurnsForSession` helper:
```go
func loadTurnsForSession(baseDir, sessionID string) ([]interface{}, error) {
    files, err := getJSONLFiles(baseDir)
    if err != nil { return nil, err }

    for _, file := range files {
        var turns []interface{}
        // read file line by line, parse JSON, filter by sessionId == sessionID
        // if any matching turns found, return them (session is in one file)
        if len(turns) > 0 { return turns, nil }
    }
    return nil, nil
}
```

**`cmd/mcp-server/executor.go`**:

1. Add to `toolPipelineConfig`:
   ```go
   contextTurns int
   ```

2. In `newToolPipelineConfig`:
   ```go
   contextTurns: getIntParam(args, "context_turns", 0),
   ```

3. Add method `expandContextTurns`:
   ```go
   func (e *ToolExecutor) expandContextTurns(
       rawData []interface{}, N int, baseDir string,
   ) ([]interface{}, error) {
       // 1. Build matchedUUIDs set
       // 2. For each distinct sessionId in rawData, call loadTurnsForSession
       // 3. For each matched turn, find index, expand window [idx-N..idx+N]
       // 4. Mark context field, deduplicate, preserve order
       // return expanded []interface{}
   }
   ```

4. In `buildResponse`, after `applyMessageFiltersToData`:
   ```go
   if pipeline.contextTurns > 0 &&
      toolName == "query_user_messages" &&
      getStringParam(args, "content_type", "string") != "array" {
       baseDir, err := getQueryBaseDir(
           getStringParam(args, "scope", "project"),
           getStringParam(args, "working_dir", ""),
       )
       if err != nil { return "", err }
       parsedData, err = e.expandContextTurns(parsedData, pipeline.contextTurns, baseDir)
       if err != nil { return "", err }
   }
   ```
   Note: `e.workingDir` — verify ToolExecutor has a workingDir field; if not, extract from
   MCP server config or pass via Execute args (already in `args` map).

5. In `GroupBySession` (phase 53), ensure `match_count` counts only `context!=true`:
   ```go
   // When counting, skip turns where context==true
   if ctx, _ := obj["context"].(bool); !ctx {
       g.MatchCount++
   }
   g.Turns = append(g.Turns, entry)
   ```

**`cmd/mcp-server/tools.go`**:

```json
"context_turns": {
  "type": "integer",
  "description": "Number of turns to include before and after each matched turn (same session). Context turns are marked with 'context': true. Default: 0 (disabled). Only applies to string content_type."
}
```

Run `make commit`.

---

## Cross-Phase Notes

### Field name consistency (all phases)

- **Raw data** (before `applyMessageFiltersToData`): always camelCase `sessionId`
- **Parsed data** (after `ApplyContentSummary`): snake_case `session_id`
- `GenerateTimestampStats` and `GenerateSessionStats` always receive **rawData** — use `obj["sessionId"]`
- `GroupBySession` receives **parsedData** — must check both field names (see Phase 53 implementation)

### workingDir availability in ToolExecutor

`ToolExecutor` is an empty struct (no fields). Use `getStringParam(args, "working_dir", "")`
to extract `workingDir` at the call site — already updated in Stage 55.2 pseudocode above.

### Phase sizing check

Each stage must stay ≤200 LoC. Stage 55.2 is estimated at ~230 LoC — split preemptively:
- **Stage 55.2a**: `loadTurnsForSession` helper in `handlers_query.go` (~70 LoC)
- **Stage 55.2b**: `expandContextTurns` method + integration in `buildResponse` (~160 LoC)

---

## Validation

After all 4 phases:

```bash
make push   # full validation: format + build + tests + lint + coverage
```

Then validate with real data:
```javascript
// Phase 52
query_user_messages({pattern: ".", content_summary: true, preview_length: 30, limit: 5})
// → content_preview ≤ 30 chars, CJK readable

// Phase 53
query_user_messages({pattern: "proposal|plan|implement", group_by_session: true, since: "2026-03-08T00:00:00Z"})
// → session objects with match_count, first_match, last_match, turns

// Phase 54
query_user_messages({pattern: ".", stats_only: true, stats_level: "session"})
// → per-session aggregation instead of hour buckets

// Phase 55
query_user_messages({pattern: "创建 proposal", context_turns: 2, content_summary: true})
// → each matched turn surrounded by 2 context turns
```
