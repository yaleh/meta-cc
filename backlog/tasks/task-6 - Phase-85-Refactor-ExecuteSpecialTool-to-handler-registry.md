---
id: TASK-6
title: 'Phase 85: Refactor ExecuteSpecialTool to handler registry'
status: 'Basic: Done'
assignee: []
created_date: '2026-06-23 06:24'
updated_date: '2026-06-23 06:31'
labels:
  - 'kind:basic'
dependencies: []
references:
  - internal/mcp/executor/executor.go
  - internal/mcp/executor/registry.go
  - internal/mcp/executor/handler.go
  - internal/mcp/executor/query_handlers.go
  - internal/mcp/executor/analysis_handlers.go
  - internal/mcp/executor/edit_sequences_handler.go
  - internal/mcp/executor/registry_test.go
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Phase 85: ExecuteSpecialTool 重构为 handler registry。

`internal/mcp/executor/executor.go` 中的 `ExecuteSpecialTool` 方法原本用一个大 if-else/switch 分发特殊工具（如 `cleanup_temp_files`、`get_session_directory` 等）。改成 map-based handler registry 以提升可测试性和可扩展性。约 200 行改动。

目标文件:
- `internal/mcp/executor/executor.go`
- `internal/mcp/executor/registry.go`
- `internal/mcp/executor/query_handlers.go`
- `internal/mcp/executor/analysis_handlers.go`
- `internal/mcp/executor/registry_test.go`

注意：经代码调研，该重构已在 commit 248d80c 完成。
<!-- SECTION:DESCRIPTION:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Proposal: Phase 85 — Refactor ExecuteSpecialTool to Handler Registry

## Background

**WHY — the original problem:**

Prior to this refactoring, `ExecuteSpecialTool` in `internal/mcp/executor/executor.go` used a large `if-else` / `switch` chain to dispatch special tools (`cleanup_temp_files`, `get_session_directory`, `inspect_session_files`, `execute_stage2_query`, `get_session_metadata`, `analyze_bugs`, `analyze_errors`, `quality_scan`, `get_work_patterns`, `get_timeline`, `get_tech_debt`). This created several problems:

1. **Testability**: Adding a test for a single handler required constructing a full executor and triggering the entire dispatch chain; there was no way to unit-test handlers in isolation.
2. **Extensibility**: Every new special tool required editing `executor.go`, increasing the risk of merge conflicts and violating the Open/Closed Principle.
3. **Inconsistency**: Convenience (query) tools already used `registerQueryHandler()` + `init()` in `handlers.go`; special tools were the only remaining exception.

**Current state (post-refactoring):**

Commit `248d80c` completed the migration. `ExecuteSpecialTool` is now a pure registry lookup (lines 76-88 of `executor.go`); all special tool logic lives in `init()` blocks inside dedicated files (`query_handlers.go`, `analysis_handlers.go`, `edit_sequences_handler.go`).

## Goals

1. Replace the `if-else` / `switch` in `ExecuteSpecialTool` with a `map[string]SpecialToolHandler` registry lookup — zero branching on tool name.
2. Register all special handlers via `registerHandler()` + `init()`, mirroring the existing `registerQueryHandler()` pattern.
3. Maintain full backward-compatibility: no change to MCP tool signatures or behaviour.
4. Achieve ≥80 % test coverage on the executor package, with each handler individually testable.
5. Keep `executor.go` free of per-tool `import` paths; each handler file owns its own imports.

## Proposed Approach

**High-level design:**

- Define `SpecialToolHandler` type in `handler.go` (already done: `func(ctx, *ToolExecutor, map[string]interface{}) (string, error)`).
- Declare `specialToolRegistry` and `registerHandler()` in `registry.go` (already done).
- Split handlers by concern into dedicated files that each call `registerHandler()` inside `init()`:
  - `query_handlers.go` — file-system / session-directory tools
  - `analysis_handlers.go` — analysis service tools
  - `edit_sequences_handler.go` — edit sequence tool
- `ExecuteSpecialTool` performs a single map lookup; on miss it returns `(_, false, nil)` to signal "not a special tool".
- `registry_test.go` verifies all expected tool names appear in `specialToolRegistry` at startup, and unit-tests `registerHandler` in isolation.

## Trade-offs and Risks

**What is NOT done:**
- No changes to MCP server wiring (`cmd/mcp-server/`) — that layer is out of scope.
- No changes to `queryHandlerRegistry` or convenience tool pipeline — only `specialToolRegistry` is affected.
- No extraction of handler logic into separate packages; handlers remain in the `executor` package to avoid circular imports.

**Known risks:**
- `init()` registration order is non-deterministic across files; mitigated by using a map (order-independent) and ensuring no two files register the same key.
- Accidental double-registration silently overwrites; mitigated by the registry test enumerating every expected key.

# Plan: Phase 85 — Refactor ExecuteSpecialTool to Handler Registry

## Phase A: Define registry infrastructure and SpecialToolHandler type

### Tests (write first)

File: `internal/mcp/executor/registry_test.go`

Test cases (write and watch fail before implementation):
- `TestSpecialToolRegistry_AnalysisHandlers` — assert each analysis tool name is present in `specialToolRegistry` at process start
- `TestSpecialToolRegistry_QueryHandlers` — assert each query/session tool name is present in `specialToolRegistry`
- `TestSpecialToolRegistry_UnknownTool` — assert a random string is absent from the registry
- `TestRegisterHandler_AddsToRegistry` — call `registerHandler` with a test name + stub, verify presence and invocation
- `TestQueryHandlerRegistry_AllConvenienceTools` — existing test; must continue to pass
- `TestQueryHandlerRegistry_UnknownTool` — existing test; must continue to pass
- `TestRegisterQueryHandler_AddsToRegistry` — existing test; must continue to pass

### Implementation

Files to create / modify:
- `internal/mcp/executor/handler.go` — declare `SpecialToolHandler` and `QueryHandlerFunc` types
- `internal/mcp/executor/registry.go` — declare `specialToolRegistry` map and `registerHandler()` func; declare `queryHandlerRegistry` map and `registerQueryHandler()` func

### DoD
- [ ] `go test ./internal/mcp/executor/...`
- [ ] `specialToolRegistry` and `registerHandler` exported-shape unchanged from handler.go / registry.go
- [ ] No per-tool import in executor.go

## Phase B: Migrate special tool handlers to init()-registered files

### Tests (write first)

File: `internal/mcp/executor/registry_test.go` (add to existing)

Test cases:
- `TestSpecialToolRegistry_AnalysisHandlers` — covers all six analysis tools registered in `analysis_handlers.go` (already listed above; write before implementation)
- `TestSpecialToolRegistry_QueryHandlers` — covers five session tools registered in `query_handlers.go`
- (For `edit_sequences_handler.go`) extend `TestSpecialToolRegistry_QueryHandlers` or add a dedicated `TestSpecialToolRegistry_EditSequences` that asserts `query_edit_sequences` is present

File: `internal/mcp/executor/edit_sequences_handler_test.go`
- `TestHandleQueryEditSequences_EmptyFiles` — call handler directly with empty `files` param; expect error
- `TestHandleQueryEditSequences_MissingFiles` — call handler with missing `files` key; expect error

### Implementation

Files to create / modify:
- `internal/mcp/executor/analysis_handlers.go` — `init()` registers `analyze_bugs`, `analyze_errors`, `quality_scan`, `get_work_patterns`, `get_timeline`, `get_tech_debt` via `registerHandler()`
- `internal/mcp/executor/query_handlers.go` — `init()` registers `cleanup_temp_files`, `get_session_directory`, `inspect_session_files`, `execute_stage2_query`, `get_session_metadata` via `registerHandler()`
- `internal/mcp/executor/edit_sequences_handler.go` — `init()` registers `query_edit_sequences` via `registerHandler()`
- `internal/mcp/executor/executor.go` — `ExecuteSpecialTool` body becomes a single map lookup; remove all per-tool `if`/`switch` branches

### DoD
- [ ] `go test ./internal/mcp/executor/...`
- [ ] `executor.go` contains no `if toolName ==` or `switch toolName` branching
- [ ] Each handler file has its own `init()` with `registerHandler()` calls

## Phase C: Validate full suite and coverage

### Tests (write first)

No new test files; this phase runs the full suite to confirm nothing regressed.

### Implementation

No new code; fix any test failures surfaced by Phase B.

### DoD
- [ ] `go test ./internal/mcp/executor/...`
- [ ] Test coverage for `executor` package ≥ 80 % (`go test -cover ./internal/mcp/executor/...`)
- [ ] `go vet ./internal/mcp/executor/...` reports no issues

## Constraints

- Total code change ≤ 200 lines (this is a restructuring, not a feature addition)
- Do not change MCP tool names, parameter schemas, or return formats
- Do not add new packages; all changes remain within `internal/mcp/executor/`
- Handler files must not import `executor.go` symbols that would create cycles

## Acceptance Gate

- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] `go test -cover ./internal/mcp/executor/...` shows ≥ 80 % coverage
- [ ] `git diff HEAD~1 --stat` shows ≤ 200 lines changed
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
ExecuteSpecialTool 重构已在 commit 248d80c 完成。executor.go:76 现为纯 map 查找（specialToolRegistry），所有 handler 通过 init() 注册在 analysis_handlers.go / query_handlers.go / edit_sequences_handler.go，与 queryHandlerRegistry 完全对称。无需任何代码改动，任务关闭。
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 go test ./internal/mcp/executor/...
- [ ] #2 go test -cover ./internal/mcp/executor/... (≥80% coverage)
- [ ] #3 go vet ./internal/mcp/executor/...
- [ ] #4 go test ./...
<!-- DOD:END -->
