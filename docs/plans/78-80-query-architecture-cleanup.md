# Plan 78–80: Query Architecture Cleanup

**Status**: Completed (internal/query/jq/ deleted; turnindex/, sequences/, assistant/ sub-packages created)
**Proposal**: [docs/proposals/proposal-query-architecture-cleanup.md](../proposals/proposal-query-architecture-cleanup.md)

---

## Overview

Three independent structural issues have accumulated in the `internal/query` area since the
architecture-hygiene phases (58–59) completed in March 2026. Each issue is addressed in its
own phase without affecting the others.

| Phase | Scope | Key deliverable |
|---|---|---|
| 78 | Delete dead code in `internal/query/jq/` | Directory removed; build and tests pass without modification to any other file |
| 79 | Fix permanently-skipped tests | Zero `t.Skip(...)` calls remain in the two affected test files |
| 80 | Extract sub-packages from `internal/query` mega-package | `internal/query/turnindex/`, `internal/query/sequences/`, and `internal/query/assistant/` exist as independent packages |

**Phase dependencies**: Phases 78 and 79 are independent and may proceed in any order.
Phase 80 must begin only after Phases 78 and 79 are complete (clean `make commit` baseline).

```text
Phase 78 ──┐
Phase 79 ──┤→ Phase 80
```

---

## Phase 78: Delete Dead Code in `internal/query/jq/`

### Objectives

Remove the `internal/query/jq/` sub-package entirely. Its three files are exact duplicates of
canonical implementations in the parent `internal/query` package and have zero production
importers. The `.golangci.yml` `depguard` rule that references the deleted path must also be
removed to avoid a dangling lint exception.

### Acceptance Criteria

- `internal/query/jq/` directory no longer exists in the repository.
- `go build ./...` and `go test ./...` pass without modification to any other file.
- No `golangci.yml` exclude rule references `internal/query/jq`.
- `make commit` passes.

### Stages

#### Stage 78-A: Confirm zero importers and delete `internal/query/jq/`

- Change budget: ≤50 lines (3 file deletions; no production code modified)
- Pre-flight check:
  - Run `grep -r '"github.com/yaleh/meta-cc/internal/query/jq"' .` to confirm zero production
    importers before deletion.
- Tasks:
  - Delete `internal/query/jq/jq.go` (436 lines — duplicate of `internal/query/jq.go`)
  - Delete `internal/query/jq/stage2_executor.go` (212 lines — duplicate of
    `internal/query/stage2_executor.go`)
  - Delete `internal/query/jq/stage2_executor_test.go` (97 lines — tests for the duplicate
    executor; coverage already provided by `internal/query/stage2_executor_test.go`)
  - Run `go build ./...` to confirm no missing references.
- Files:
  - `internal/query/jq/jq.go` — **delete**
  - `internal/query/jq/stage2_executor.go` — **delete**
  - `internal/query/jq/stage2_executor_test.go` — **delete**
- Tests:
  - `go test ./...` — must pass with no modification to any remaining file
- Exit criteria:
  - `internal/query/jq/` directory does not exist
  - `go build ./...` is clean
  - `go test ./...` passes

#### Stage 78-B: Remove dangling `depguard` rule from `.golangci.yml`

- Change budget: ≤10 lines
- Tasks:
  - Open `.golangci.yml` and locate the `depguard` rule named
    `no-query-jq-imports-query` that matches `**/internal/query/jq/**`.
  - Delete the entire rule block (rule name, path pattern, and any associated message or
    deny entries).
  - Run `golangci-lint run` (or `make lint`) to confirm no linter error about unmatched
    patterns and no remaining reference to `internal/query/jq`.
- Files:
  - `.golangci.yml` — delete `no-query-jq-imports-query` rule block
- Tests:
  - `make commit` — full validation including lint
- Exit criteria:
  - `golangci-lint run` produces no warning or error about `internal/query/jq`
  - `make commit` is green

### Phase 78 Validation

- `make commit` passes
- `internal/query/jq/` directory is absent from the repository
- `grep -r 'internal/query/jq' .` returns only changelog, git history, and this plan (no
  `.go` or `.golangci.yml` references)
- `go test ./...` passes with no modification to any file outside the deleted directory

---

## Phase 79: Fix Permanently-Skipped Tests

### Objectives

Bring both affected test files to a defined, permanently-passing state. Delete the 10 stub
functions in `handlers_convenience_test.go` that skip unconditionally and assert nothing.
Replace the 6 conditionally-skipped tests in `internal/analysis/service_test.go` with
stub-based equivalents that pass in any environment.

### Acceptance Criteria

- `cmd/mcp-server/handlers_convenience_test.go` contains zero `t.Skip(...)` calls.
- `internal/analysis/service_test.go` contains zero `t.Skip(...)` calls.
- New stub-based tests in `service_test.go` cover all six service methods
  (`AnalyzeBugs`, `AnalyzeErrors`, `QualityScan`, `GetWorkPatterns`, `GetTimeline`,
  `GetTechDebt`) via the `NewWithAnalyzers` injection path.
- `go test ./...` passes; test count is either unchanged or reduced only by deleted stubs.
- `make commit` passes.

### Stages

#### Stage 79-A: Delete 10 permanently-skipped stubs in `handlers_convenience_test.go`

- Change budget: ≤150 lines (deletions only; no new code)
- Background:
  The following 10 functions call `setupConvenienceToolTest` to build a complete fixture
  and then immediately call `t.Skip("underlying handleQuery() is already tested")` without
  executing any assertion. They are dead fixture-setup overhead and should be removed.

  | Function |
  |---|
  | `TestHandleQueryUserMessages` |
  | `TestHandleQueryTools` |
  | `TestHandleQueryToolErrors` |
  | `TestHandleQueryTokenUsage` |
  | `TestHandleQueryConversationFlow` |
  | `TestHandleQuerySystemErrors` |
  | `TestHandleQueryFileSnapshots` |
  | `TestHandleQueryTimestamps` |
  | `TestHandleQuerySummaries` |
  | `TestHandleQueryToolBlocks` |

  The 3 non-skipped tests in the same file
  (`TestHandleQueryUserMessagesContentLengthFiltering`,
  `TestQueryUserMessagesSchemaHasContentLengthParams`,
  `TestHandleQueryTools_ToolParamFilters`) must be preserved exactly as-is.
- Tasks:
  - Delete the 10 skipped test functions listed above from
    `cmd/mcp-server/handlers_convenience_test.go`.
  - Leave `setupConvenienceToolTest` in place if it is still referenced by the 3 remaining
    tests; delete it only if it becomes unused.
  - Run `go test ./cmd/mcp-server/...` to confirm the 3 non-skipped tests still pass.
- Files:
  - `cmd/mcp-server/handlers_convenience_test.go` — delete 10 skipped test functions
- Tests:
  - `go test ./cmd/mcp-server/...` — the 3 non-skipped tests must pass
  - `grep -c 't\.Skip' cmd/mcp-server/handlers_convenience_test.go` must return `0`
- Exit criteria:
  - Zero `t.Skip` calls in `handlers_convenience_test.go`
  - `go test ./cmd/mcp-server/...` passes
  - `make dev` passes

#### Stage 79-B: Replace 6 skipped tests in `internal/analysis/service_test.go` with stub-based tests

- Change budget: ≤200 lines
- Background:
  The following 6 tests currently skip when `cmd/mcp-server/test.jsonl` is absent:

  | Function | Skip guard |
  |---|---|
  | `TestService_AnalyzeBugs` | `test.jsonl not available` |
  | `TestService_AnalyzeErrors` | same |
  | `TestService_QualityScan` | same |
  | `TestService_GetWorkPatterns` | same |
  | `TestService_GetTimeline` | same |
  | `TestService_GetTechDebt` | same |

  The `stubErrorAnalyzer` pattern is already present in the same file and used by the 2
  non-skipped tests (`TestService_WithStubErrorAnalyzer`,
  `TestService_WithStubErrorAnalyzer_Error`). All six analyzer interfaces are already
  defined in `internal/analyzer/interfaces.go` (`BugAnalyzer`, `ErrorAnalyzer`,
  `QualityScanner`, `WorkPatternsAnalyzer`, `TimelineAnalyzer`, `TechDebtAnalyzer`).
  `analysis.Analyzers` already has fields for all six; `NewWithAnalyzers` already accepts
  all six. No changes to `internal/analysis` package code are required.
- Tasks (TDD order):
  1. In `internal/analysis/service_test.go`, add five new stub structs following the
     `stubErrorAnalyzer` example — one for each remaining analyzer interface:
     `stubBugAnalyzer`, `stubQualityScanner`, `stubWorkPatternsAnalyzer`,
     `stubTimelineAnalyzer`, `stubTechDebtAnalyzer`. Each stub returns minimal valid data
     and no error on its single method.
  2. Add five new test functions — one per remaining service method — using
     `analysis.NewWithAnalyzers` to inject the corresponding stub:
     `TestService_WithStubBugAnalyzer`, `TestService_WithStubQualityScanner`,
     `TestService_WithStubWorkPatternsAnalyzer`, `TestService_WithStubTimelineAnalyzer`,
     `TestService_WithStubTechDebtAnalyzer`.
  3. Run `go test ./internal/analysis/...` to confirm the new stub-based tests pass.
  4. Delete the 6 conditionally-skipped test functions.
  5. Run `go test ./internal/analysis/...` again to confirm nothing regressed.
- Files:
  - `internal/analysis/service_test.go` — add 5 stub structs + 5 test functions; delete
    6 skipped test functions
- Tests:
  - `go test ./internal/analysis/...`
  - `grep -c 't\.Skip' internal/analysis/service_test.go` must return `0`
  - `make commit`
- Exit criteria:
  - Zero `t.Skip` calls in `internal/analysis/service_test.go`
  - All six service methods exercised by stub-based tests via `NewWithAnalyzers`
  - `make commit` is green

### Phase 79 Validation

- `make commit` passes
- `grep -rn 't\.Skip' cmd/mcp-server/handlers_convenience_test.go` returns no results
- `grep -rn 't\.Skip' internal/analysis/service_test.go` returns no results
- All six service methods (`AnalyzeBugs`, `AnalyzeErrors`, `QualityScan`,
  `GetWorkPatterns`, `GetTimeline`, `GetTechDebt`) have a passing stub-based test

---

## Phase 80: Extract Sub-packages from `internal/query` Mega-package

### Objectives

Decompose two self-contained sub-domains out of the `internal/query` mega-package (21
production files, ~3,200 implementation lines, zero declared interfaces). Promote the shared
helper `buildTurnIndex` (and `getToolCallTimestamp`) to a new neutral sub-package
`internal/query/turnindex/` first, then extract `sequences.go` and `assistant_messages.go`
into their own packages. Interfaces are deferred until a concrete test-double substitution
need arises in a calling package.

### Acceptance Criteria

- `internal/query/turnindex/`, `internal/query/sequences/`, and
  `internal/query/assistant/` exist as independent packages.
- `internal/query` production file count drops from 21 to 19 or fewer.
- Test coverage across all affected packages is ≥ 80%.
- All MCP tool behaviors are preserved (verified by `make commit`).
- No new `depguard` rule violations are introduced.

### Stages

#### Stage 80-A: Promote shared helpers to `internal/query/turnindex/`

- Change budget: ≤80 lines
- Background:
  `buildTurnIndex` is defined in `context.go` and called by 6 production files
  (`assistant_messages.go`, `context.go`, `file_access.go`, `project_state.go`,
  `prompts.go`, `sequences.go`). `getToolCallTimestamp` is defined in `file_access.go`
  and called by `file_access.go` itself and `sequences.go`. Both are unexported, creating
  hidden coupling that prevents sub-package extraction.

  The existing `internal/query/files/` sub-package and its governing
  `no-query-files-imports-query` depguard rule establish the architectural constraint:
  sub-packages must not import the parent `internal/query` package. Promoted helpers must
  therefore live in a *new neutral sub-package* (`internal/query/turnindex/`) rather than
  remain in the parent package.
- Tasks:
  1. Create `internal/query/turnindex/turnindex.go` with package `turnindex`. Move (and
     export) `buildTurnIndex` as `BuildTurnIndex` and `getToolCallTimestamp` as
     `GetToolCallTimestamp`. Keep function signatures identical except for the case change.
  2. Create `internal/query/turnindex/turnindex_test.go` with at least one test per
     exported function (TDD: write the test file before moving the implementation).
  3. In `internal/query/context.go`: replace the `buildTurnIndex` definition with a
     thin forwarding call to `turnindex.BuildTurnIndex`; update all 6 callers in the
     parent package to use `turnindex.BuildTurnIndex` directly (or keep the forwarder and
     update only the forwarding stub — whichever stays within the line budget).
  4. In `internal/query/file_access.go`: replace `getToolCallTimestamp` with a call to
     `turnindex.GetToolCallTimestamp`; update the caller in `sequences.go`.
  5. Run `go build ./internal/query/...` and `go test ./internal/query/...`.
- Files:
  - `internal/query/turnindex/turnindex.go` — **create** (exported `BuildTurnIndex`,
    `GetToolCallTimestamp`)
  - `internal/query/turnindex/turnindex_test.go` — **create** (unit tests)
  - `internal/query/context.go` — replace `buildTurnIndex` definition; add
    `internal/query/turnindex` import
  - `internal/query/file_access.go` — replace `getToolCallTimestamp` definition; add
    `internal/query/turnindex` import
  - `internal/query/sequences.go` — update `getToolCallTimestamp` call site to
    `turnindex.GetToolCallTimestamp`
  - `internal/query/assistant_messages.go` — update `buildTurnIndex` call site to
    `turnindex.BuildTurnIndex`
  - `internal/query/project_state.go` — update `buildTurnIndex` call site
  - `internal/query/prompts.go` — update `buildTurnIndex` call site
- Tests:
  - `go test ./internal/query/turnindex/...`
  - `go test ./internal/query/...` — all existing tests must pass
  - `make commit`
- Exit criteria:
  - `internal/query/turnindex/` package exists with exported `BuildTurnIndex` and
    `GetToolCallTimestamp`
  - No unexported `buildTurnIndex` or `getToolCallTimestamp` remain in the parent package
  - `make commit` is green

#### Stage 80-B: Extract `sequences.go` to `internal/query/sequences/`

- Change budget: ≤150 lines
- Background:
  `sequences.go` (277 lines) implements `BuildToolSequenceQuery` and the
  sequence-detection algorithm. After Stage 80-A it depends only on
  `turnindex.BuildTurnIndex` and `turnindex.GetToolCallTimestamp` — no remaining dependency
  on unexported parent-package symbols. It is now safe to extract.
- Tasks (TDD order):
  1. Create `internal/query/sequences/sequences.go` with package `sequences`. Copy
     the implementation of `BuildToolSequenceQuery` and related unexported helpers from
     `internal/query/sequences.go`.
  2. Create `internal/query/sequences/sequences_test.go`. Migrate the existing tests for
     `BuildToolSequenceQuery` from `internal/query/` to the new package (adjust package
     name; update imports).
  3. Run `go test ./internal/query/sequences/...` — new tests must pass.
  4. Update callers: search `cmd/mcp-server/` for uses of `query.BuildToolSequenceQuery`;
     update the import to `sequences "github.com/yaleh/meta-cc/internal/query/sequences"`
     and update call sites.
  5. Delete `internal/query/sequences.go` from the parent package. Delete the
     corresponding test lines from `internal/query/` test files that tested
     `BuildToolSequenceQuery` (the coverage now lives in the new package).
  6. Run `go build ./...` and `go test ./...`.
- Files:
  - `internal/query/sequences/sequences.go` — **create**
  - `internal/query/sequences/sequences_test.go` — **create**
  - `internal/query/sequences.go` — **delete**
  - `cmd/mcp-server/` handler files referencing `query.BuildToolSequenceQuery` — update
    import and call site
  - `internal/query/` test files — remove test lines for `BuildToolSequenceQuery` that
    migrated to the new package
- Tests:
  - `go test ./internal/query/sequences/...`
  - `go test ./internal/query/...`
  - `go test ./cmd/mcp-server/...`
  - `make commit`
- Exit criteria:
  - `internal/query/sequences/` package exists
  - `internal/query/sequences.go` does not exist
  - `make commit` is green

#### Stage 80-C: Extract `assistant_messages.go` to `internal/query/assistant/`

- Change budget: ≤200 lines
- Background:
  `assistant_messages.go` (546 lines) implements `BuildAssistantMessages` and
  `BuildConversationTurns`. After Stage 80-A it depends only on
  `turnindex.BuildTurnIndex` — no remaining dependency on unexported parent-package
  symbols. Its test file adds roughly 400–500 lines; migrating the full test file in a
  single stage will be tight against the 200-line limit. Limit the migration to the
  *new* lines written (new package file header, import updates, and any adjustments)
  rather than counting lines that move verbatim.
- Tasks (TDD order):
  1. Create `internal/query/assistant/assistant.go` with package `assistant`. Copy the
     implementation of `BuildAssistantMessages` and `BuildConversationTurns` (and any
     unexported helpers used only by those two functions) from
     `internal/query/assistant_messages.go`.
  2. Create `internal/query/assistant/assistant_test.go`. Migrate the existing tests for
     `BuildAssistantMessages` and `BuildConversationTurns` from `internal/query/` to the
     new package (adjust package name; update imports).
  3. Run `go test ./internal/query/assistant/...` — migrated tests must pass.
  4. Update callers: search `cmd/mcp-server/` for uses of `query.BuildAssistantMessages`
     and `query.BuildConversationTurns`; update the import to
     `assistant "github.com/yaleh/meta-cc/internal/query/assistant"` and update call
     sites.
  5. Delete `internal/query/assistant_messages.go`. Delete the corresponding test lines
     from `internal/query/` test files that tested those two functions.
  6. Run `go build ./...` and `go test ./...`.
- Files:
  - `internal/query/assistant/assistant.go` — **create**
  - `internal/query/assistant/assistant_test.go` — **create**
  - `internal/query/assistant_messages.go` — **delete**
  - `cmd/mcp-server/` handler files referencing `query.BuildAssistantMessages` or
    `query.BuildConversationTurns` — update import and call sites
  - `internal/query/` test files — remove test lines for the migrated functions
- Tests:
  - `go test ./internal/query/assistant/...`
  - `go test ./internal/query/...`
  - `go test ./cmd/mcp-server/...`
  - `make commit`
- Exit criteria:
  - `internal/query/assistant/` package exists
  - `internal/query/assistant_messages.go` does not exist
  - `internal/query` production file count is ≤ 19
  - `make commit` is green

### Phase 80 Validation

- `make commit` passes
- `internal/query/turnindex/`, `internal/query/sequences/`, and
  `internal/query/assistant/` directories all exist
- `internal/query/sequences.go` and `internal/query/assistant_messages.go` do not exist
- `go test ./internal/query/...` coverage is ≥ 80%
- `grep -rn 'query\.BuildToolSequenceQuery\|query\.BuildAssistantMessages\|query\.BuildConversationTurns' cmd/` returns no results (callers use sub-package imports)

---

## Test Strategy

- **TDD**: For Phase 79 Stage 79-B, write the five stub structs and test functions before
  deleting the skipped tests. For Phase 80, create the new package's test file before
  moving the implementation.
- **Coverage requirement**: ≥ 80% across all affected packages. No coverage regression is
  acceptable. Run `make test-coverage` after each phase to verify.
- **Incremental validation**: Run `make commit` after every stage. Do not start the next
  stage until the current stage's `make commit` is green.
- **Pre-flight grep**: Before each deletion, run the relevant grep to confirm zero
  unexpected references to the item being deleted.

---

## Dependencies

- **Phase 78** depends on: nothing (Phase 76–77 are complete).
- **Phase 79** depends on: nothing; may run concurrently with Phase 78.
- **Phase 80** depends on: Phases 78 and 79 must both have a clean `make commit`
  baseline before Phase 80 begins. Stage 80-B and 80-C each depend on Stage 80-A.

```text
Phase 78 ──┐
Phase 79 ──┴→ Phase 80 (Stage 80-A → Stage 80-B → Stage 80-C)
```

- **Blocks**: A potential follow-on phase to define `AssistantQuerier` and
  `SequenceQuerier` interfaces in `cmd/mcp-server/` (the consumer), deferred until a
  concrete test-double substitution need arises.

---

## Non-Goals

- Defining `AssistantQuerier` or `SequenceQuerier` interfaces in the new sub-packages
  (deferred per architect review; interfaces belong at the consumer when needed)
- Modifying `BuildAssistantMessages`, `BuildConversationTurns`, or
  `BuildToolSequenceQuery` behavior
- Changing MCP wire behavior (tool names, parameters, response shapes)
- Decomposing files other than `sequences.go` and `assistant_messages.go`
- Closing the `cmd/mcp-server → internal/query` import entirely (other files in
  `internal/query` remain in the parent package)

---

## Total Estimates

| Stage | Description | Estimated LOC |
|---|---|---|
| 78-A | Delete 3 files in `internal/query/jq/` | ~0 (deletions) |
| 78-B | Remove dangling `depguard` rule from `.golangci.yml` | ~10 |
| 79-A | Delete 10 skipped stubs in `handlers_convenience_test.go` | ~120 deleted |
| 79-B | Add 5 stub structs + 5 test functions; delete 6 skipped tests in `service_test.go` | ~100 new + ~80 deleted |
| 80-A | Create `internal/query/turnindex/`; update 6 callers | ~80 |
| 80-B | Extract `sequences.go` → `internal/query/sequences/` | ~100–150 |
| 80-C | Extract `assistant_messages.go` → `internal/query/assistant/` | ~150–200 |
| **Total** | 7 stages across 3 phases | **~440–560 lines net new** |

Each stage is within the ≤200-line limit. Each phase is within the ≤500-line limit.

---

## Validation Checklist

- [ ] `internal/query/jq/` directory deleted
- [ ] `grep -r 'internal/query/jq' .` returns no `.go` or `.golangci.yml` matches
- [ ] `no-query-jq-imports-query` depguard rule removed from `.golangci.yml`
- [ ] `grep -c 't\.Skip' cmd/mcp-server/handlers_convenience_test.go` returns `0`
- [ ] `grep -c 't\.Skip' internal/analysis/service_test.go` returns `0`
- [ ] All six service methods covered by stub-based tests via `NewWithAnalyzers`
- [ ] `internal/query/turnindex/` package exists with exported `BuildTurnIndex` and `GetToolCallTimestamp`
- [ ] `internal/query/sequences/` package exists; `internal/query/sequences.go` deleted
- [ ] `internal/query/assistant/` package exists; `internal/query/assistant_messages.go` deleted
- [ ] `internal/query` production file count ≤ 19
- [ ] `go test ./internal/query/...` coverage ≥ 80%
- [ ] `make commit` passes after each stage
- [ ] All 21 MCP tool behaviors preserved (verified by `make commit` full test suite)
