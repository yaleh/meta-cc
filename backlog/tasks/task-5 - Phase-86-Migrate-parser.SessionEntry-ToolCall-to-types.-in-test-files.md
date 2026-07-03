---
id: TASK-5
title: 'Phase 86: Migrate parser.SessionEntry/ToolCall to types.* in test files'
status: 'Basic: Backlog'
assignee: []
created_date: '2026-06-23 06:24'
updated_date: '2026-06-23 06:24'
labels:
  - 'kind:basic'
dependencies: []
ordinal: 3000
---

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Proposal: Phase 86 — Migrate parser.* to types.* in Test Files

## Background

During Phase 56-57, the canonical type definitions for `SessionEntry` and `ToolCall` were moved from `internal/parser` to `internal/types`. Type aliases were left in `internal/parser` for backward compatibility with production code that was updated incrementally. However, test files were never updated and still import these types through the `parser` package alias path.

Referencing types through `parser.SessionEntry` / `parser.ToolCall` in tests creates unnecessary coupling to the `internal/parser` package even when the test logic has nothing to do with parsing. It also obscures the canonical location of the types for future readers. The goal is to eliminate all such references in `*_test.go` files, replacing them with `types.SessionEntry` and `types.ToolCall` from `internal/types`.

## Goals

1. Remove all `parser.SessionEntry` and `parser.ToolCall` usages from `*_test.go` files across the codebase (36 files identified).
2. Update `import` blocks in affected files to reference `internal/types` instead of (or in addition to) `internal/parser`, removing the `parser` import where it is no longer needed after the substitution.
3. Ensure all existing tests continue to pass without any logic changes.

## Proposed Approach

- Mechanical find-and-replace: swap `parser.SessionEntry` → `types.SessionEntry` and `parser.ToolCall` → `types.ToolCall` in every `*_test.go` file.
- Adjust `import` blocks: add the `internal/types` import where missing; remove the `internal/parser` import from files where it is only used for the two type names and no other symbols.
- No changes to production (non-test) files.
- No deletion of the type aliases in `internal/parser` — those remain for any production callers that have not yet migrated.

## Trade-offs and Risks

- **Risk: minimal.** Only test files are touched; type aliases in `parser` still exist so no compilation breakage in production code.
- **Risk: hidden parser usage.** Some test files may import `parser` for reasons beyond the two type names (e.g., calling parser functions). Those files will retain the `parser` import and only swap the type references. A follow-up grep after the change confirms no residual `parser.SessionEntry` / `parser.ToolCall`.
- **Trade-off: scope vs. 500-line limit.** 36 files × ~5 lines average change = ~180 lines, well within the 500-line phase limit.

---

# Plan: Phase 86 — Migrate parser.* to types.* in Test Files

## Phase A: Replace type references in all 36 affected test files

### Tests (write first / green-light criterion)

The existing tests themselves are the acceptance signal. The approach is:

1. Run `go test ./...` **before** making changes to capture the current passing baseline.
2. Apply all substitutions.
3. Run `go test ./...` again — all tests must pass (same count, no new failures).
4. Run the negative grep to confirm zero residual `parser.SessionEntry` / `parser.ToolCall` in test files.

No new test code needs to be written; the migration is purely mechanical.

### Implementation

**Files to modify** (36 total, grouped by package):

*internal/analysis/*
- `service_test.go`

*internal/analyzer/*
- `bugs_analysis_test.go`
- `errors_analysis_test.go`
- `patterns_test.go`
- `quality_analysis_test.go`
- `stats_test.go`
- `tech_debt_test.go`
- `testutil_test.go`
- `timeline_test.go`
- `work_patterns_test.go`
- `workflow_test.go`

*internal/filter/*
- `filter_test.go`
- `pagination_test.go`
- `time_test.go`

*internal/output/*
- `chunker_test.go`
- `estimator_test.go`
- `json_test.go`
- `projection_test.go`
- `sort_test.go`
- `summary_test.go`
- `tsv_test.go`

*internal/query/*
- `context_test.go`
- `file_access_test.go`
- `messages_test.go`
- `project_state_test.go`
- `prompts_test.go`
- `stats_helpers_test.go`
- `tools_test.go`

*internal/query/assistant/*
- `assistant_test.go`

*internal/query/resources/*
- `resources_test.go`

*internal/query/sequences/*
- `sequences_test.go`

*internal/query/turnindex/*
- `turnindex_test.go`

*internal/stats/*
- `aggregator_test.go`
- `files_test.go`
- `metrics_test.go`
- `timeseries_test.go`

**Per-file change pattern:**

1. In the `import` block: add `"<module>/internal/types"` if not already present; remove `"<module>/internal/parser"` if it is only used for `SessionEntry` / `ToolCall`.
2. In the body: replace every `parser.SessionEntry` with `types.SessionEntry` and every `parser.ToolCall` with `types.ToolCall`.

### DoD

- [ ] `go test ./internal/analyzer/...` passes
- [ ] `go test ./internal/filter/...` passes
- [ ] `go test ./internal/output/...` passes
- [ ] `go test ./internal/query/...` passes
- [ ] `go test ./internal/stats/...` passes
- [ ] `go test ./internal/analysis/...` passes
- [ ] `! grep -r "parser\.SessionEntry\|parser\.ToolCall" internal/ --include="*_test.go"` returns empty

## Constraints

- Only modify `*_test.go` files — production code is not touched.
- The type aliases in `internal/parser` are **not** removed; they remain for any callers not yet migrated.
- Each file change must leave test logic byte-for-byte identical aside from import paths and type-name prefixes.

## Acceptance Gate

- [ ] `go test ./...` passes with zero new failures
- [ ] `! grep -r "parser\.SessionEntry\|parser\.ToolCall" internal/ --include="*_test.go"` returns empty (no residual references)
<!-- SECTION:PLAN:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 `go test ./...` passes with zero new failures
- [ ] #2 `! grep -r 'parser\.SessionEntry\|parser\.ToolCall' internal/ --include='*_test.go'` returns empty (no residual references in test files)
<!-- DOD:END -->
