# Plan 50–51: Query Noise Reduction and Stats Fix

**Status**: Completed
**Date**: 2026-03-09
**Proposal**: [docs/proposals/proposal-query-noise-reduction.md](../proposals/proposal-query-noise-reduction.md)

---

## Overview

Two independent fixes to `query_user_messages`:

| # | Change | Severity | Phase |
|---|--------|----------|-------|
| P50 | `exclude_system_messages` parameter — filter Claude Code injected messages | High |
| P51 | Fix `session_count: 0` in `stats_first`/`stats_only` when `content_summary=true` | High |

**Development methodology**: TDD throughout. Each stage begins with a failing test; implementation follows.

**Code limits**: Phase ≤500 lines, Stage ≤200 lines.

**Phase independence**: P50 and P51 touch different code paths and can be executed in parallel (worktree isolation).

---

## Phase 50: `exclude_system_messages` Parameter

**Goal**: Add `exclude_system_messages bool` (default `false`) to `query_user_messages`. When
`true` and `content_type == "string"`, appends a compound jq `select | not` clause that
rejects entries whose `message.content` starts with any Claude Code system-injection prefix.

**Estimated code**: ~65 lines across 3 files.

**Files touched**:
- `cmd/mcp-server/handlers_convenience.go` — extract param, build jq clause
- `cmd/mcp-server/tools.go` — schema property
- `cmd/mcp-server/handlers_query_test.go` — integration tests

### Stage 50.1 — Tests + Implementation

**TDD sequence**:

1. Add `TestExcludeSystemMessages` in `handlers_query_test.go`:
   - Follow the `setupTimeFilterFixture()` pattern (set `META_CC_PROJECTS_ROOT`, write JSONL
     to a hash-named subdirectory, return cleanup function).
   - Fixture JSONL: 6 entries, all `"type":"user"`, all with `"message":{"content":"<string>"}`:
     - 2 real user intent messages (content: `"hello"`, `"fix the bug"`)
     - 1 content starting with `<local-command-caveat>Caveat: The messages...`
     - 1 content starting with `<command-name>/plugin</command-name>`
     - 1 content starting with `<local-command-stdout>some output</local-command-stdout>`
     - 1 content starting with `<task-notification>task done</task-notification>`
   - Assert `executeQuery(exclude_system_messages=false)` (via `handleQueryUserMessages`)
     returns 6 entries.
   - Assert same call with `exclude_system_messages=true` returns exactly 2 entries.

2. Add `TestExcludeSystemMessages_NoErrorOnArrayType` in `handlers_query_test.go`:
   - Fixture: 2 entries with `"message":{"content":[{"type":"tool_result",...}]}` (array type).
   - Call `handleQueryUserMessages` with `content_type="array"`, `exclude_system_messages=true`.
   - Assert no error is returned and both array-type entries are returned
     (the system-tag filter is silently skipped for array content type).

3. **Run tests → confirm FAIL**.

4. **Implement** in `handlers_convenience.go`:

```go
excludeSystem := getBoolParam(args, "exclude_system_messages", false)
// ... existing filter construction ...
if excludeSystem && (contentType == "string" || contentType == "") {
    jqFilter += ` | select(
        .message.content | (
            startswith("<local-command-caveat>") or
            startswith("<command-name>") or
            startswith("<local-command-stdout>") or
            startswith("<task-notification>")
        ) | not
    )`
}
```

5. Add schema property in `tools.go`:
```json
"exclude_system_messages": {
    "type": "boolean",
    "description": "If true, exclude Claude Code system-injected messages
        (<local-command-caveat>, <command-name>, <local-command-stdout>,
        <task-notification>). Only applies to string content type.
        Default: false."
}
```

6. **Run tests → confirm PASS**.
7. **Run `make commit`** to validate.

**Acceptance criteria**:
- `exclude_system_messages=true` removes all 4 system-tag prefixes from results.
- `exclude_system_messages=false` (or omitted) is identical to current behavior.
- `content_type=array` with `exclude_system_messages=true` produces no error.
- `tools.go` schema documents the new parameter.

---

## Phase 51: Fix `session_count: 0` Stats Bug

**Goal**: Restructure `buildResponse` so stats are always computed from raw (untransformed)
data, fixing `session_count=0` in both `stats_first` and `stats_only` modes when
`content_summary=true`.

**Estimated code**: ~70 lines across 2 files.

**Files touched**:
- `cmd/mcp-server/executor.go` — restructure `buildResponse`, update `buildStatsFirstResponse`
- `cmd/mcp-server/executor_test.go` — regression + new tests

### Stage 51.1 — Tests

**Important**: Tests must exercise the full `ExecuteTool("query_user_messages", ...)` path,
not call `buildStatsOnlyResponse` or `buildResponse` directly. The bug lives at line 299 in
`Execute`, which fires before `buildResponse`. Direct calls to response builders bypass line
299 and would not catch (or verify fixing) the actual bug.

Write failing tests in `executor_test.go`, following `setupLibraryFixture` pattern but with
a **two-session custom fixture** (the shared helper uses a single `sessionId`):

```go
func setupTwoSessionFixture(t *testing.T) func() {
    projectDir := t.TempDir()
    projectsRoot := t.TempDir()
    t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

    fixture := `{"type":"user","timestamp":"2026-03-09T06:00:00Z","uuid":"u1",` +
        `"sessionId":"sess-A","message":{"role":"user","content":"hello"}}` + "\n" +
        `{"type":"user","timestamp":"2026-03-09T07:00:00Z","uuid":"u2",` +
        `"sessionId":"sess-B","message":{"role":"user","content":"world"}}` + "\n"

    writeSessionFixture(t, projectDir, "two-sessions", fixture)
    // chdir to projectDir so locator finds the session
    oldWd, _ := os.Getwd()
    _ = os.Chdir(projectDir)
    return func() { _ = os.Chdir(oldWd) }
}
```

Tests:

1. `TestStatsFirstWithContentSummary`:
   - Use `setupTwoSessionFixture`.
   - Call `ExecuteTool("query_user_messages", {pattern:".", stats_first:true, content_summary:true})`.
   - Split output on `"\n---\n"`. Parse first line (summary) as JSON.
   - Assert `session_count == 2` (not 0).
   - Assert detail section (after `---`) contains `"content_preview"` field.

2. `TestStatsOnlyWithContentSummary`:
   - Same fixture and call with `stats_only:true, content_summary:true`.
   - Parse first output line as JSON. Assert `session_count == 2`.

3. `TestStatsFirstWithoutContentSummary` (regression guard):
   - Same fixture, `stats_first:true`, no `content_summary`.
   - Assert `session_count == 2` (already works today; must not regress).

4. **Run tests → confirm FAIL** for tests 1 and 2; test 3 passes (bug only triggered when
   `content_summary=true` causes line 299 to transform entries).

### Stage 51.2 — Implementation

**Root cause recap**: `applyMessageFiltersToData` at executor.go line 299 renames
`sessionId`→`session_id` before `buildResponse` is called. Stats then see `session_id`
but check for `sessionId`.

**Fix**:

1. **Remove** the pre-`buildResponse` filter block (executor.go lines 299–301):
   ```go
   // DELETE:
   if toolName == "query_user_messages" && config.requiresMessageFilters() {
       queryResult.Entries = e.applyMessageFiltersToData(...)
   }
   ```

2. **Restructure `buildResponse`** — move the transform inside, after the stats path:
   ```go
   func (e *ToolExecutor) buildResponse(cfg *config.Config, result QueryResult,
       args map[string]interface{}, toolName string, pipeline toolPipelineConfig) (string, error) {

       rawData := result.Entries

       // stats_only: compute stats from raw data; no detail rendering
       if pipeline.statsOnly {
           output, err := e.buildStatsOnlyResponse(rawData, toolName)
           if err != nil { return "", err }
           return injectWarnings(output, result.Warnings)
       }

       // Apply message filters for detail rendering (AFTER stats-only path)
       parsedData := rawData
       if toolName == "query_user_messages" && pipeline.requiresMessageFilters() {
           parsedData = e.applyMessageFiltersToData(rawData,
               pipeline.maxMessageLength, pipeline.contentSummary)
       }

       var output string
       var err error
       if pipeline.statsFirst {
           output, err = e.buildStatsFirstResponse(cfg, rawData, parsedData, args, toolName)
       } else {
           output, err = e.buildStandardResponse(cfg, parsedData, args, toolName)
       }
       if err != nil { return "", err }
       return injectWarnings(output, result.Warnings)
   }
   ```

3. **Update `buildStatsFirstResponse`** — add `rawData` parameter for stats:
   ```go
   func (e *ToolExecutor) buildStatsFirstResponse(
       cfg *config.Config,
       rawData []interface{},    // used for stats (sessionId preserved)
       parsedData []interface{}, // used for detail (may be content_summary)
       args map[string]interface{},
       toolName string,
   ) (string, error) {
       jsonlData, err := e.dataToJSONL(rawData)  // ← was parsedData
       // ... stats computation unchanged ...
       response, err := adaptResponse(cfg, parsedData, args, toolName)  // ← unchanged
       // ...
   }
   ```

4. **Run tests → confirm PASS**.
5. **Run `make commit`** to validate.

### Stage Dependencies

```
51.1 (failing tests)
  └─► 51.2 (implementation + green)
```

**Acceptance criteria**:
- `stats_first=true, content_summary=true` → `session_count` equals actual distinct session count (verified via `ExecuteTool` path).
- `stats_only=true, content_summary=true` → same.
- `stats_first=true` without `content_summary` → no regression.
- `stats_only=true` without `content_summary` → no regression.
- Standard (non-stats) response with `content_summary=true` → no regression (verified by existing `TestApplyMessageFiltersToData`).

---

## Verification

After both phases, run:

```bash
make push
```

Manual smoke test (requires running MCP server):

```
# Should return ~75 records instead of ~222, with correct session_count
query_user_messages(
  pattern=".",
  since="2026-03-08T00:00:00Z",
  exclude_system_messages=true,
  stats_first=true,
  content_summary=true
)
# Verify: stats header has session_count > 0
# Verify: no records with content_preview starting with <local-command-caveat>
```

---

## Parallel Execution

Both phases are independent. Run in parallel using worktree isolation:

```
Phase 50 (worktree A): handlers_convenience.go, tools.go, handlers_query_test.go
Phase 51 (worktree B): executor.go, executor_test.go
```

Merge order: either phase first; no conflicts expected (disjoint file sets).
