# Plan: Architecture Cleanup — Phases 81–86

**Created**: 2026-03-12
**Proposal**: [proposal-arch-cleanup-phases-81-86.md](../proposals/proposal-arch-cleanup-phases-81-86.md)
**Preceding plan**: [78-80-query-architecture-cleanup.md](78-80-query-architecture-cleanup.md)
**Status**: Partially Completed (Phase 81 ✅ ParseTimestamp exported; Phase 82 ✅ workflow.go 52 LOC; Phase 83 ✅ interfaces.go uses types.*; Phase 84 ✅ no injected func fields in pipeline; Phase 85 ❌ ExecuteSpecialTool still exists; Phase 86 ❌ parser.* in test files)

---

## Overview

Six architectural issues identified after Phase 80 completion, ordered for safe incremental resolution:

| Phase | Issue(s) | Description | Est. Lines |
|-------|----------|-------------|-----------|
| 81 | H-2 | Deduplicate `buildTurnIndex` / `parseTimestamp` / `getActionType` | ~120 |
| 82 | H-3 | Split `workflow.go` god file (479 lines) | ~200 |
| 83 | C-1 | Fix analyzer interfaces: `parser` → `types` | ~80 |
| 84 | H-4 | Remove injected function types from `pipeline` | ~60 |
| 85 | C-2 | Refactor `ExecuteSpecialTool` to handler registry | ~200 |
| 86 | M-6 | Complete `parser` → `types` migration (production files) | ~180 |

**Total estimated modifications**: ~840 lines across 6 phases.

---

## Dependency Order

```
H-2 (Phase 81) → H-3 (Phase 82)   [H-3 split is cleaner after duplicates removed]
C-1 (Phase 83)                      [independent; clears parser from interfaces]
H-4 (Phase 84)                      [independent; simplifies executor call sites]
C-2 (Phase 85)                      [independent; OCP refactor of executor]
M-6 (Phase 86) ← C-1 (Phase 83)   [C-1 removes interface imports first]
```

---

## Phase 81 — Deduplicate Utility Functions (H-2)

**Goal**: Export `ParseTimestamp` from `internal/query/turnindex`; remove private duplicates from `internal/analyzer/workflow.go` and `internal/query/context.go`; move `getActionType` to `internal/types`.

**Acceptance Criteria**:
- `turnindex.ParseTimestamp` is exported (uppercase)
- `workflow.go` no longer defines `buildTurnIndex` or `parseTimestamp`
- `internal/query/context.go` no longer defines `parseTimestamp`
- `getActionType` exists in exactly one location
- `make commit` passes

### Stage 81.1 — Export `ParseTimestamp` from `turnindex` and update `turnindex` internal callers

**Files**:
- `internal/query/turnindex/turnindex.go` — rename `parseTimestamp` → `ParseTimestamp`; update internal call on line 31

**Estimated changes**: ~5 lines
**Test**: `go test ./internal/query/turnindex/...` must pass

### Stage 81.2 — Remove private duplicates from `workflow.go` and `work_patterns.go`

**Files**:
- `internal/analyzer/workflow.go` — remove `buildTurnIndex` (lines 248–258), remove `parseTimestamp` (lines 443–449), replace call sites with `turnindex.BuildTurnIndex` and `turnindex.ParseTimestamp`; add import
- `internal/analyzer/work_patterns.go` — replace `parseTimestamp(tc.Timestamp)` with `turnindex.ParseTimestamp(tc.Timestamp)`; add import

**Estimated changes**: ~40 lines removed, ~10 lines added (~50 total)
**Test**: `go test ./internal/analyzer/...` must pass

### Stage 81.3 — Remove private `parseTimestamp` from `query/context.go`

**Files**:
- `internal/query/context.go` — remove `parseTimestamp` (lines 181–187), replace 2 call sites with `turnindex.ParseTimestamp`; add import

**Estimated changes**: ~10 lines
**Test**: `go test ./internal/query/...` must pass

### Stage 81.4 — Move `getActionType` to `internal/types`; update callers

**Files**:
- `internal/types/toolcall.go` — add exported `FileActionType(toolName string) string` function
- `internal/analyzer/workflow.go` — remove private `getActionType` (lines 419–432); update caller on line 94
- `internal/query/file_access.go` — remove private `getActionType` (lines 120–134); update caller on line 37

**Estimated changes**: ~35 lines removed, ~20 lines added (~55 total)
**Test**: `go test ./internal/...` must pass

**Phase 81 total**: ~115 lines

---

## Phase 82 — Split `workflow.go` God File (H-3)

**Goal**: Decompose `internal/analyzer/workflow.go` (479 lines, 18 functions) into focused files. All new files remain in `package analyzer`.

**Acceptance Criteria**:
- `workflow.go` is either deleted or reduced to ≤50 lines (shared types only)
- Each new file is ≤200 lines
- All exported public functions remain accessible as `analyzer.DetectToolSequences`, `analyzer.DetectFileChurn`, `analyzer.DetectIdlePeriods`
- `make commit` passes

### Stage 82.1 — Extract sequence detection to `sequences.go`

**Functions to move**: `DetectToolSequences`, `extractToolCallsWithTurns`, `findAllSequences`, `calculateSequenceTimeSpan`, `toolCallWithTurn` type

**Files**:
- Create `internal/analyzer/sequences.go` with the 5 items above
- Remove those items from `internal/analyzer/workflow.go`

**Estimated changes**: ~140 lines moved (net zero, but 140 lines of new file + 140 deleted from workflow)
**Test**: `go test ./internal/analyzer/...` must pass

### Stage 82.2 — Extract churn and idle analysis to `churn.go` and `idle.go`

**Functions to move**:
- `churn.go`: `DetectFileChurn`, `fileAccessStats` type, `extractFileFromToolCall`, `extractCommandFromToolCall`
- `idle.go`: `DetectIdlePeriods`, `extractTurnContext`

**Files**:
- Create `internal/analyzer/churn.go` (~100 lines)
- Create `internal/analyzer/idle.go` (~80 lines)
- Remove those items from `internal/analyzer/workflow.go`

**Estimated changes**: ~180 lines moved
**Test**: `go test ./internal/analyzer/...` must pass

**Note**: After Stages 82.1–82.2, `workflow.go` should contain only the helper types (`IdlePeriod`, `TurnContext`, `FileChurnDetail`, `SequenceAnalysis`, `FileChurnAnalysis`, `IdlePeriodAnalysis`) and `getToolCallTimestamp`. The latter will be replaced by `turnindex.GetToolCallTimestamp` in a cleanup step within 82.2.

**Phase 82 total**: ~200 lines modified (plus moved code)

---

## Phase 83 — Fix Analyzer Interfaces: `parser` → `types` (C-1)

**Goal**: All public interfaces in `internal/analyzer/interfaces.go` use `[]types.SessionEntry` and `[]types.ToolCall`; the `import "internal/parser"` in that file is removed.

**Acceptance Criteria**:
- `internal/analyzer/interfaces.go` imports `internal/types`, not `internal/parser`
- `DefaultAnalyzer` method signatures updated to match
- All callers compile without change (type aliases are identical)
- `make commit` passes

### Stage 83.1 — Update interface declarations

**Files**:
- `internal/analyzer/interfaces.go` — change all `parser.SessionEntry` → `types.SessionEntry`, `parser.ToolCall` → `types.ToolCall`; replace `import "parser"` with `import "types"`

**Estimated changes**: ~15 lines
**Test**: `go test ./internal/analyzer/...` must pass

### Stage 83.2 — Update `DefaultAnalyzer` method receivers

**Files**:
- `internal/analyzer/interfaces.go` — update all 6 method receivers on `DefaultAnalyzer` to match updated interface parameter types

**Estimated changes**: ~20 lines (already part of the same file; may be combined with 83.1)
**Test**: `make commit` must pass

**Phase 83 total**: ~35 lines

---

## Phase 84 — Remove Injected Function Types from `pipeline` (H-4)

**Goal**: `internal/mcp/pipeline/pipeline.go` directly imports `internal/mcp/response` instead of accepting `AdaptResponseFunc`/`SerializeResponseFunc` as parameters.

**Acceptance Criteria**:
- `AdaptResponseFunc` and `SerializeResponseFunc` type declarations removed from `pipeline.go`
- `BuildStatsFirstResponse` and `BuildStandardResponse` take no function parameters
- `internal/mcp/executor/executor.go` no longer creates closure wrappers for injection
- No import cycle introduced
- `make commit` passes

### Stage 84.1 — Update `pipeline.go` to import `response` directly

**Files**:
- `internal/mcp/pipeline/pipeline.go` — remove `AdaptResponseFunc` and `SerializeResponseFunc` types; update `BuildStatsFirstResponse` and `BuildStandardResponse` to call `responsepkg.AdaptResponse` and `responsepkg.SerializeResponse` directly; add `import responsepkg "github.com/yaleh/meta-cc/internal/mcp/response"`

**Estimated changes**: ~30 lines removed/changed
**Test**: `go build ./internal/mcp/pipeline/...` must pass (no cycle)

### Stage 84.2 — Update `executor.go` call sites

**Files**:
- `internal/mcp/executor/executor.go` — remove the `adaptFn` and `serializeFn` closure definitions (lines 372–377); update calls to `pipelinepkg.BuildStatsFirstResponse` and `pipelinepkg.BuildStandardResponse` to remove the injected parameters

**Estimated changes**: ~20 lines removed
**Test**: `make commit` must pass

**Phase 84 total**: ~50 lines

---

## Phase 85 — Refactor `ExecuteSpecialTool` to Handler Registry (C-2)

**Goal**: Replace the 21-case `switch` in `ExecuteSpecialTool` with a registry lookup. New tools require no changes to `executor.go`.

**Acceptance Criteria**:
- `ExecuteSpecialTool` is ≤15 lines
- Each handler group is in its own file (analysis handlers, query handlers, cleanup handlers)
- `executeWithMetrics` helper eliminates repeated error-classify/record pattern
- All 11 special tools remain functional
- `make commit` passes

### Stage 85.1 — Define handler type and `executeWithMetrics`; create registry skeleton

**Files**:
- Create `internal/mcp/executor/handler.go` — define `SpecialToolHandler` function type; implement `executeWithMetrics` wrapper
- Create `internal/mcp/executor/registry.go` — define `SpecialToolRegistry` type; `RegisterHandler`, `Lookup` methods; global `defaultRegistry`

**Estimated changes**: ~60 lines
**Test**: Unit tests for registry lookup

### Stage 85.2 — Migrate analysis tool handlers to `analysis_handlers.go`

**Functions** (6 tools): `analyze_bugs`, `analyze_errors`, `quality_scan`, `get_work_patterns`, `get_timeline`, `get_tech_debt`

**Files**:
- Create `internal/mcp/executor/analysis_handlers.go` — 6 handler functions wrapping `e.AnalysisSvc.*`; register via `RegisterHandler` calls

**Estimated changes**: ~60 lines
**Test**: Existing analysis tool tests must pass

### Stage 85.3 — Migrate query and cleanup handlers; replace `ExecuteSpecialTool` switch

**Functions** (5 tools): `get_session_directory`, `inspect_session_files`, `execute_stage2_query`, `get_session_metadata`, `cleanup_temp_files`

**Files**:
- Create `internal/mcp/executor/query_handlers.go` — 5 handler functions; register via `RegisterHandler` calls
- Update `internal/mcp/executor/executor.go` — replace the 11-arm `ExecuteSpecialTool` switch body with registry lookup + `executeWithMetrics` call (~10 lines)

**Estimated changes**: ~80 lines
**Test**: `make commit` must pass; all 11 special tools reachable via registry; all 21 tools in the file remain functional

**Phase 85 total**: ~200 lines

---

## Phase 86 — Complete `parser` → `types` Migration for Production Files (M-6)

**Goal**: All 44 production (non-test) Go files import `internal/types` directly instead of `internal/parser` (42 test files deferred to a follow-on phase). The `internal/parser` package retains `aliases.go` for backward compatibility but is marked deprecated.

**Acceptance Criteria**:
- 0 production files import `internal/parser` for type usage
- `internal/parser/aliases.go` has deprecation comment
- `make commit` passes after each stage

**File groups** (by package):

| Stage | Package(s) | Files | Est. Lines |
|-------|------------|-------|-----------|
| 86.1 | `internal/analyzer/` | 9 files | ~18 lines |
| 86.2 | `internal/query/` (sub-packages) | 12 files | ~24 lines |
| 86.3 | `internal/filter/`, `internal/output/`, `internal/stats/` | 10 files | ~20 lines |
| 86.4 | `internal/mcp/filters/`, `internal/mcp/query/`, `internal/analysis/` | 3 files | ~6 lines |

Each stage: replace `"github.com/yaleh/meta-cc/internal/parser"` with `"github.com/yaleh/meta-cc/internal/types"` in the import block; replace `parser.Xxx` references with `types.Xxx` (or bare `types.ExtractToolCalls` → `types.ExtractToolCalls`). For `var ExtractToolCalls = types.ExtractToolCalls` in parser, callers become `types.ExtractToolCalls` directly.

### Stage 86.1 — Migrate `internal/analyzer/` files

**Files**: `bugs_analysis.go`, `errors_analysis.go`, `interfaces.go` (already done in Phase 83), `patterns.go`, `quality_analysis.go`, `stats.go`, `tech_debt.go`, `timeline.go`, `work_patterns.go`, `workflow.go`

**Note**: `interfaces.go` is already done after Phase 83. Remaining: ~9 files.
**Estimated changes**: ~18 lines
**Test**: `go test ./internal/analyzer/...`

### Stage 86.2 — Migrate `internal/query/` and sub-packages

**Files**: `aggregate.go`, `assistant/assistant.go`, `context.go` (already updated in Phase 81), `file_access.go`, `file_churn.go`, `files/file_inspector.go`, `filter.go`, `project_state.go`, `prompts.go`, `resources.go`, `resources/messages.go`, `resources/tools.go`, `sequences/sequences.go`, `stage2_executor.go`, `stats_helpers.go`, `tools.go`, `turnindex/turnindex.go` (already updated in Phase 81), `unified.go`

**Note**: Some files updated in Phase 81; remaining ~12 files.
**Estimated changes**: ~24 lines
**Test**: `go test ./internal/query/...`

### Stage 86.3 — Migrate `internal/filter/`, `internal/output/`, `internal/stats/`

**Files**: `filter/filter.go`, `filter/pagination.go`, `filter/time.go`, `output/chunker.go`, `output/estimator.go`, `output/projection.go`, `output/sort.go`, `output/summary.go`, `output/tsv.go`, `stats/aggregator.go`, `stats/files.go`, `stats/metrics.go`, `stats/timeseries.go`

**Estimated changes**: ~26 lines
**Test**: `go test ./internal/filter/... ./internal/output/... ./internal/stats/...`

### Stage 86.4 — Migrate remaining files; add deprecation comment to `aliases.go`

**Files**: `internal/mcp/filters/filters.go`, `internal/mcp/query/query.go`, `internal/analysis/service.go`; update `internal/parser/aliases.go` with deprecation comment

**Estimated changes**: ~12 lines
**Test**: `make commit`

**Phase 86 total**: ~80 lines production code changes + deprecation comment

---

## Testing Strategy

**Methodology**: TDD where new code is introduced (C-2 registry, H-2 exported function). For refactors (H-3 split, H-4 injection removal, M-6 import substitution), existing tests verify behaviour is unchanged.

**Coverage target**: ≥80% maintained throughout. New code in Phase 85 (handler registry) must have dedicated unit tests.

**After each stage**: `make commit` (format + build + tests).
**After each phase**: `make push` (full lint + coverage check).

---

## Rollback Plan

Each phase is independently revertable via `git revert <phase-commit>`. Phases have no cross-phase runtime dependencies — only compilation dependencies (C-1 must precede the M-6 stages that touch `interfaces.go` callers, but M-6 stage 86.1 handles that by updating the remaining analyzer files).
