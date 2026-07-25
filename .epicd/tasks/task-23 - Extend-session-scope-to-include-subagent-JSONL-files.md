---
id: TASK-23
title: Extend session scope to include subagent JSONL files
assignee: []
created_date: '2026-07-15 11:00'
updated_date: '2026-07-15 11:30'
labels: []
dependencies: []
ordinal: 14000
pipeline_id: execution
phase: done
dod:
  - text: make commit
    checked: false
  - text: >-
      go test ./internal/mcp/query/... -run
      TestGetQueryFiles_SessionScope_ReturnsSingleFile
    checked: false
  - text: >-
      go test ./internal/mcp/query/... -run
      TestGetQueryFiles_ProjectScope_IncludeSubagents
    checked: false
  - text: >-
      go test ./cmd/mcp-server/... -run
      TestHandleGetSessionDirectory_SubagentFileCount
    checked: false
entry_phase: authoring/backlog
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Background

MCP query tools rely on `GetQueryBaseDir` (internal/mcp/query/query.go:446) to resolve which JSONL files to load. Two defects exist:

1. **`scope: "session"` returns the wrong directory.** The function calls `loc.FromProjectPath` to get the current session file path, then returns `filepath.Dir(sessionFile)` — the project directory — which is identical to the project scope. Querying with `scope: "session"` silently scans all sessions rather than the current one.

2. **Subagent session files are not reachable via any scope.** Claude Code stores subagent JSONL files under `<project_dir>/<session_uuid>/subagents/`. The private `getJSONLFiles` helper (internal/mcp/filters/filters.go:401) explicitly skips all directories (`if entry.IsDir() { continue }`), so subagent sessions are invisible to every MCP query and analysis tool.

Users discovered this when trying to query the tool calls made by an independent search agent: the only way to reach those files was to hardcode the path and pass it directly to `execute_stage2_query`.

## Goals

1. Fix `scope: "session"` so it resolves to the single current-session JSONL file (and, with `include_subagents=true`, that session's `subagents/` files), not the whole project directory.

2. Add `include_subagents` parameter (default: `true`) to all query and analysis tools. When true, also scan `<project_dir>/<uuid>/subagents/*.jsonl` files alongside the top-level session files.

3. Surface subagent file count in `get_session_directory` output so callers can gauge scope before running expensive queries.

## Approach

Replace the `GetQueryBaseDir → dir → getJSONLFiles(dir)` chain with a unified `GetQueryFiles(scope, workingDir string, includeSubagents bool) ([]string, error)` function that directly returns the resolved file list for any scope/subagent combination. This eliminates the dir-as-intermediate-value pattern that caused the session-scope bug.

- `scope=session, includeSubagents=false` → `[<current_session>.jsonl]`
- `scope=session, includeSubagents=true`  → `[<current_session>.jsonl] + <uuid>/subagents/*.jsonl`
- `scope=project, includeSubagents=false` → all top-level `*.jsonl` (current behavior)
- `scope=project, includeSubagents=true`  → top-level + all `*/subagents/*.jsonl`

The subagent scan is fixed at two levels deep (never full recursion) to avoid picking up `tool-results/` or other subdirectories.

`get_session_directory` is updated to return `subagent_file_count` (always, even when `include_subagents=false`, so callers know what they are opting out of).

The `include_subagents` parameter is added to: `query_session_content`, `query_session_signals`, `query_file_activity`, `get_session_directory`, `get_timeline`, `analyze_errors`, `analyze_bugs`, `quality_scan`, `get_work_patterns`, `get_tech_debt`.

## Non-Goals

- Full recursive JSONL scanning (out of scope; only `*/subagents/` is added).
- Indexing or querying `tool-results/` directories.
- Changing how `execute_stage2_query` works — it already accepts explicit file paths, which is the correct workaround for ad-hoc subagent queries.
- CLI binary flag changes (the MCP server has no standalone CLI query interface beyond MCP tool calls).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 scope=session resolves to a single JSONL file: go test ./internal/mcp/query/... exits 0 and new TestGetQueryFiles_SessionScope test passes (added in Phase A)
- [ ] #2 include_subagents=false excludes subagent JSONL files: go test ./internal/mcp/filters/... -run TestGetJSONLFiles_ExcludeSubagents exits 0
- [ ] #3 get_session_directory response contains subagent_file_count >= 0 regardless of include_subagents value: go test ./internal/mcp/executor/... -run TestGetSessionDirectory_SubagentCount exits 0
- [ ] #4 All existing tests continue to pass: make commit exits 0
- [ ] #5 include_subagents=true (default) includes subagent JSONL files: go test ./internal/mcp/query/... -run TestGetQueryFiles_ProjectScope_IncludeSubagents exits 0
- [ ] #6 include_subagents=false excludes subagent JSONL files: go test ./internal/mcp/query/... -run TestGetQueryFiles_ProjectScope_ExcludeSubagents exits 0
- [ ] #7 get_session_directory response contains subagent_file_count >= 0: go test ./cmd/mcp-server/... -run TestHandleGetSessionDirectory_SubagentFileCount exits 0
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Plan: Extend session scope to include subagent JSONL files

## Phase A: Introduce `GetQueryFiles` — session scope fix + subagent scanning

### Tests (write first)
File: `internal/mcp/query/query_files_test.go` (new)

Test cases:
- `TestGetQueryFiles_SessionScope_ReturnsSingleFile` — session scope returns exactly the current session JSONL file (1 file, not all files in dir)
- `TestGetQueryFiles_SessionScope_IncludeSubagents_AppendsSubagentFiles` — session scope + includeSubagents=true also returns `<sessionUUID>/subagents/*.jsonl`
- `TestGetQueryFiles_ProjectScope_ExcludeSubagents` — project scope + includeSubagents=false returns only top-level JSONL files (existing behavior)
- `TestGetQueryFiles_ProjectScope_IncludeSubagents` — project scope + includeSubagents=true returns top-level + all `*/subagents/*.jsonl`
- `TestGetQueryFiles_SubagentDirScansOnlySubagentsNotToolResults` — confirms `tool-results/` subdirectory is not scanned

### Implementation
File: `internal/mcp/query/query.go`
- Add exported `GetQueryFiles(scope, workingDir string, includeSubagents bool) ([]string, error)`
  - session scope: call `loc.FromProjectPath` → get single file path; append that session's `<uuid>/subagents/*.jsonl` if includeSubagents=true
  - project scope: call existing `GetJSONLFiles(dir)` for top-level; scan `<dir>/<entry>/subagents/*.jsonl` if includeSubagents=true (fixed 2-level, not full recursion)
- `GetQueryBaseDir` and `GetJSONLFiles` stay unchanged (still used by `ExpandContextTurns` in pipeline.go)

### DoD
- [ ] `go test ./internal/mcp/query/... -run TestGetQueryFiles`
- [ ] `go test ./internal/mcp/query/...`

## Phase B: Wire `GetQueryFiles` into executor, pipeline, stage + `include_subagents` param

### Tests (write first)
Files:
- `internal/mcp/executor/handlers_test.go` — add `TestHandleQueryUserMessages_IncludeSubagents_Default`, `TestHandleQueryToolBlocks_IncludeSubagents_False`
- `internal/mcp/pipeline/pipeline_test.go` — add `TestPipelineConfig_IncludeSubagents`
- `internal/mcp/query/stage_test.go` — add `TestHandleGetSessionMetadata_SessionScope_SingleFile`

### Implementation
File: `internal/mcp/executor/executor.go`
- Add `includeSubagents bool` param to `ExecuteQueryWithTimeRange` and `ExecuteQuery`
- Replace `GetQueryBaseDir + GetJSONLFiles` (lines 164-176) with `GetQueryFiles(scope, workingDir, includeSubagents)`
- Update `ExecuteQueryForProvider` and `ExecuteQueryWithTimeRangeForProvider` to extract `includeSubagents` from `args` or accept it as param

File: `internal/mcp/executor/handlers.go`
- In each handler function: extract `includeSubagents := GetBoolParam(args, "include_subagents", true)` and pass to executor call

File: `internal/mcp/pipeline/pipeline.go`
- Add `IncludeSubagents bool` to `PipelineConfig`
- Note: `GetQueryBaseDir` call at line 71 (for `ExpandContextTurns`) stays unchanged

File: `internal/mcp/executor/executor.go` (`NewToolPipelineConfig`)
- Add `IncludeSubagents: GetBoolParam(args, "include_subagents", true)`

File: `internal/mcp/query/stage.go` (`HandleGetSessionMetadata`)
- Replace `GetQueryBaseDir + GetJSONLFiles` (lines 262-268) with `GetQueryFiles(scope, "", true)`

File: `internal/mcp/tools/tools.go`
- Add `"include_subagents": {Type: "boolean", Description: "Include subagent session files (default: true). Pass false to query only top-level sessions."}` to `StandardToolParameters()`

### DoD
- [ ] `go test ./internal/mcp/executor/... -run TestHandleQuery`
- [ ] `go test ./internal/mcp/executor/... ./internal/mcp/pipeline/... ./internal/mcp/query/...`

## Phase C: `get_session_directory` — add `subagent_file_count`

### Tests (write first)
File: `cmd/mcp-server/handlers_stage1_test.go`
- `TestHandleGetSessionDirectory_SubagentFileCount_WithSubagents` — creates dir with `<uuid>/subagents/` tree; asserts `subagent_file_count > 0`
- `TestHandleGetSessionDirectory_SubagentFileCount_NoSubagents` — no subagent subdirs; asserts `subagent_file_count == 0`
- `TestHandleGetSessionDirectory_SessionScope_FileCountIsOne` — session scope returns `file_count == 1` (update existing `TestHandleGetSessionDirectory_SessionScope`)

### Implementation
File: `internal/mcp/query/stage.go`
- `HandleGetSessionDirectory`: call `GetQueryFiles(scope, "", false)` for main files and `GetQueryFiles(scope, "", true)` for all; set `subagent_file_count = len(all) - len(main)` and `file_count = len(main)`
- Add `subagent_file_count` to returned map
- Note: `CollectDirectoryMetadata` is kept for size/timestamp metadata only

### DoD
- [ ] `go test ./cmd/mcp-server/... -run TestHandleGetSessionDirectory`
- [ ] `go test ./cmd/mcp-server/...`

## Constraints
- `getJSONLFiles` in `internal/mcp/filters/filters.go` is NOT modified — private, used only by `ExpandContextTurns` which correctly needs only top-level session files
- `GetQueryBaseDir` in `query.go` is NOT modified — used by `pipeline.go` for `ExpandContextTurns`, where returning the project directory is correct
- Subagent scanning is fixed at 2 levels: `<project_dir>/<uuid>/subagents/*.jsonl` only; `tool-results/` and other subdirs are not scanned
- No changes to `execute_stage2_query`

## Acceptance Gate
- [ ] `make commit`
- [ ] `go test ./internal/mcp/query/... -run TestGetQueryFiles_SessionScope_ReturnsSingleFile`
- [ ] `go test ./internal/mcp/query/... -run TestGetQueryFiles_ProjectScope_IncludeSubagents`
- [ ] `go test ./cmd/mcp-server/... -run TestHandleGetSessionDirectory_SubagentFileCount`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
authoring/draft self-review: Approved(1) after 1 round

authoring/refining review: APPROVED after 2 iteration(s)

claimed: 2026-07-15T11:10:40Z

audit round 1: 2 HIGH findings (include_subagents missing for role=assistant; file_count wrong for session scope). Fixed inline in 9dc5e67. Round 2: VERDICT done, all 5 ACs pass.
<!-- SECTION:NOTES:END -->
