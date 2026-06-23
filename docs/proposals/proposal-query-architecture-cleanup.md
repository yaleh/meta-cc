# Proposal: Query Architecture Cleanup — Three Structural Improvements

**Status**: Implemented (Phases 78–80)
**Date**: 2026-03-12
**Author**: Yale Huang

---

## Background

Three independent structural issues have accumulated in the `internal/query` area since the
architecture-hygiene phases (58–59) completed in March 2026. Each issue is localized and can be
addressed in its own phase without affecting the others. This proposal describes all three issues,
their proposed resolutions, and the trade-offs involved.

---

## Current State

### File inventory

| Path | Lines | Role |
|---|---|---|
| `internal/query/jq.go` | 433 | Canonical jq/stats implementation (package `query`) |
| `internal/query/jq/jq.go` | 436 | Exact duplicate of above (package `jq`) |
| `internal/query/stage2_executor.go` | 212 | Canonical Stage-2 query executor (package `query`) |
| `internal/query/jq/stage2_executor.go` | 212 | Exact duplicate of above (package `jq`) |
| `internal/query/jq/stage2_executor_test.go` | 97 | Tests for the duplicate executor |
| `internal/query/assistant_messages.go` | 546 | Assistant-message + conversation-turn logic |
| `internal/query/sequences.go` | 277 | Tool-sequence pattern detection |
| `cmd/mcp-server/handlers_convenience_test.go` | 532 | Convenience-tool tests; 10 of 13 functions skipped |
| `internal/analysis/service_test.go` | 216 | Service tests; 6 of 8 functions skipped |

`internal/query` (non-test production files): **21 files**, 0 declared interfaces.

---

## Goals

1. Eliminate 744 lines of unreachable duplicate code in `internal/query/jq/`.
2. Bring permanently-skipped tests to a defined state: either exercising real assertions or
   explicitly deleted when the skip reason is "already covered elsewhere."
3. Begin decomposing the `internal/query` mega-package by extracting two self-contained
   sub-domains and introducing interfaces at their boundaries.

---

## Proposed Solutions

### Issue 1 — Dead code in `internal/query/jq/`

#### Current state

`internal/query/jq/` was created to provide a dependency-free sub-package for jq and Stage-2
query utilities. Both files are **character-for-character copies** of the parent package files:

- `internal/query/jq/jq.go` (436 lines) duplicates `internal/query/jq.go` (433 lines);
  the only difference is the `package` declaration (`jq` vs `query`).
- `internal/query/jq/stage2_executor.go` (212 lines) is byte-identical to
  `internal/query/stage2_executor.go` (212 lines).

A grep of all `.go` files confirms **zero production importers** of
`github.com/yaleh/meta-cc/internal/query/jq`. The only files that reference the path are
documentation, the `.golangci.yml` lint configuration, and changelog entries.

The test file `internal/query/jq/stage2_executor_test.go` (97 lines) tests the duplicate
executor directly; its coverage is already provided by
`internal/query/stage2_executor_test.go` in the canonical package.

#### Proposed action

Delete `internal/query/jq/` entirely:

- `internal/query/jq/jq.go`
- `internal/query/jq/stage2_executor.go`
- `internal/query/jq/stage2_executor_test.go`

No callers need to be updated. The `.golangci.yml` `exclude-rules` entry that references
`internal/query/jq` should be removed at the same time to avoid a dangling lint exception.

**Net reduction**: ~745 lines (3 files).

---

### Issue 2 — Skip accumulation in tests

#### Current state

Two test files contain tests that are permanently skipped with `t.Skip(...)`:

**`cmd/mcp-server/handlers_convenience_test.go`**

| Function | Skip reason |
|---|---|
| `TestHandleQueryUserMessages` | "underlying handleQuery() is already tested" |
| `TestHandleQueryTools` | same |
| `TestHandleQueryToolErrors` | same |
| `TestHandleQueryTokenUsage` | same |
| `TestHandleQueryConversationFlow` | same |
| `TestHandleQuerySystemErrors` | same |
| `TestHandleQueryFileSnapshots` | same |
| `TestHandleQueryTimestamps` | same |
| `TestHandleQuerySummaries` | same |
| `TestHandleQueryToolBlocks` | same |

10 of 13 test functions in this file are permanently skipped (77%).

The 3 non-skipped tests (`TestHandleQueryUserMessagesContentLengthFiltering`,
`TestQueryUserMessagesSchemaHasContentLengthParams`,
`TestHandleQueryTools_ToolParamFilters`) are fully implemented and pass. The skipped functions
call `setupConvenienceToolTest` (which builds a complete fixture) and then immediately skip,
meaning the fixture setup runs but the assertions never execute.

**`internal/analysis/service_test.go`**

| Function | Skip reason |
|---|---|
| `TestService_AnalyzeBugs` | `test.jsonl not available` |
| `TestService_AnalyzeErrors` | same |
| `TestService_QualityScan` | same |
| `TestService_GetWorkPatterns` | same |
| `TestService_GetTimeline` | same |
| `TestService_GetTechDebt` | same |

6 of 8 test functions skip when `cmd/mcp-server/test.jsonl` is absent (75%). The 2
non-skipped tests (`TestService_WithStubErrorAnalyzer`,
`TestService_WithStubErrorAnalyzer_Error`) use the `stubErrorAnalyzer` pattern introduced in
the architecture-hygiene phases and pass unconditionally.

#### Proposed action

**For `handlers_convenience_test.go`** (10 skipped stubs):

The skip comment "underlying handleQuery() is already tested" accurately describes why the
stubs exist. The convenience handlers are thin wrappers that pre-configure a jq filter and
delegate to `handleQuery`. The correct resolution is to **delete the 10 skipped stub
functions** rather than implement them, for these reasons:

- Each skipped test calls `setupConvenienceToolTest` but never asserts anything; deleting
  them reduces misleading fixture overhead.
- The 3 non-skipped tests in the same file provide the correct model for per-handler
  parametric testing when real assertions are needed.
- If coverage of individual convenience handlers is later required, new tests should follow
  the `setupTestSessionDir` + `os.Chdir` pattern already established by the non-skipped tests.

**For `internal/analysis/service_test.go`** (6 conditionally-skipped tests):

These tests skip because `cmd/mcp-server/test.jsonl` is an optional real-data fixture that
may not be present in all developer environments or CI. The `stubErrorAnalyzer` pattern
(already present in the same file) is the correct long-term model. The proposed resolution is:

- **Extend the stub pattern** to cover the remaining five service methods
  (`AnalyzeBugs`, `QualityScan`, `GetWorkPatterns`, `GetTimeline`, `GetTechDebt`) by
  adding corresponding stub interfaces and stub structs following the `stubErrorAnalyzer`
  example.
- **Delete the 6 conditionally-skipped tests** once stub-based equivalents cover the same
  code paths.

This is a smaller change than it first appears. The `analysis.Analyzers` struct already
declares fields for all six analyzers (`BugAnalyzer`, `ErrorAnalyzer`, `QualityScanner`,
`WorkPatterns`, `Timeline`, `TechDebt`) and `NewWithAnalyzers` already accepts all six.
The five interfaces are likewise already defined in `internal/analyzer/interfaces.go`
(`BugAnalyzer`, `QualityScanner`, `WorkPatternsAnalyzer`, `TimelineAnalyzer`,
`TechDebtAnalyzer`). No new interface or struct field needs to be added; only five
stub structs (one per remaining analyzer interface) and five test functions need to be
written in `service_test.go`.

---

### Issue 3 — `internal/query` mega-package

#### Current state

`internal/query` contains **21 production `.go` files** (approximately 3,000 implementation
lines, 5,000+ test lines) with **zero declared interfaces**. All 21 files share a single
package namespace, making it impossible to inject test doubles at any internal boundary.

Two files are natural extraction candidates:

**`assistant_messages.go`** (546 lines): Implements `BuildAssistantMessages` and
`BuildConversationTurns`. Depends on `buildTurnIndex` (defined in `context.go`). Does NOT
directly call `getToolCallTimestamp`; that helper is only called from `file_access.go` and
`sequences.go`. `assistant_messages.go` is coupled to `buildTurnIndex` alone.

**`sequences.go`** (277 lines): Implements `BuildToolSequenceQuery` and the
sequence-detection algorithm. Depends on `buildTurnIndex` (from `context.go`) and
`getToolCallTimestamp` (from `file_access.go`). Both shared helpers must be resolved before
`sequences.go` can be extracted independently.

The root cause of the coupling is two unexported functions that are used across multiple files:

- `buildTurnIndex` in `context.go` (used by 6 production files: `assistant_messages.go`,
  `context.go` itself, `file_access.go`, `project_state.go`, `prompts.go`, `sequences.go`)
- `getToolCallTimestamp` in `file_access.go` (used by 2 production files: `file_access.go`
  itself, `sequences.go`)

#### Proposed action

A two-stage approach:

**Stage A — Promote shared helpers**

Move `buildTurnIndex` and `getToolCallTimestamp` to a new internal sub-package
`internal/query/turnindex` (or keep them in `context.go` but export them as package-level
functions). This breaks the hidden coupling that prevents extraction.

**Stage B — Extract sub-packages**

After Stage A:

1. Extract `assistant_messages.go` → `internal/query/assistant/` (new package).
2. Extract `sequences.go` → `internal/query/sequences/` (new package).
3. Define a minimal interface at each package boundary:

```go
// internal/query/assistant/service.go
type AssistantQuerier interface {
    BuildAssistantMessages(entries []parser.SessionEntry, opts Options) ([]AssistantMessage, error)
    BuildConversationTurns(entries []parser.SessionEntry, opts ConversationOptions) ([]ConversationTurn, error)
}
```

```go
// internal/query/sequences/service.go
type SequenceQuerier interface {
    BuildToolSequenceQuery(entries []parser.SessionEntry, minOccurrences int, pattern string, includeBuiltin bool) (*ToolSequenceQuery, error)
}
```

4. Update callers in `cmd/mcp-server/` to import the new sub-packages.
5. Retain stub-based tests that inject mock implementations through the new interfaces.

---

## Trade-off Analysis

### Issue 1 (delete `internal/query/jq/`)

| Trade-off | Analysis |
|---|---|
| Upside | Eliminates 744 lines of dead code; no caller migration needed |
| Risk | None — zero production importers confirmed by grep |
| Alternative | Migrate production code to use `internal/query/jq` instead of `internal/query`; rejected because it would require updating all existing callers for no functional benefit |

### Issue 2 (fix skipped tests)

| Trade-off | Analysis |
|---|---|
| Deleting skipped stubs in `handlers_convenience_test.go` | Reduces test count but removes misleading no-op fixtures; the covered code paths are tested by the 3 non-skipped tests and by `handlers_query_test.go` |
| Extending stub pattern in `service_test.go` | Requires adding 5 new analyzer interfaces + stubs; increases code but replaces environment-dependent skips with deterministic assertions |
| Keeping skips | Zero benefit; skip accumulation masks test gaps and makes `go test -v` output misleading |

### Issue 3 (decompose `internal/query`)

| Trade-off | Analysis |
|---|---|
| Sub-package extraction | Improves navigability and enables interface injection; requires promoting shared helpers first |
| Deferred decomposition | Lower immediate risk; leaves the mega-package intact and growing |
| Extracting only `sequences.go` first | Simpler scope; `sequences.go` has fewer lines and a narrower interface than `assistant_messages.go` |
| Full decomposition in one phase | Exceeds the 500-line phase limit; must be split into at least two stages |

---

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| `internal/query/jq` deletion breaks an undiscovered caller | Low | Confirm zero importers via `grep -r '"github.com/yaleh/meta-cc/internal/query/jq"' .` before deletion |
| Deleting skipped stubs removes future coverage intent | Low | Document the intended coverage model in a comment in `handlers_query_test.go` |
| Stub extension for `service_test.go` diverges from real behavior | Medium | Stub methods return minimal valid data; real-data tests remain optional via the existing skip pattern |
| `buildTurnIndex` promotion breaks internal encapsulation | Low | Exported helpers in a sub-package are still internal to the module; no external API change |
| Phase limit exceeded during `internal/query` decomposition | Medium | Stage A (promote helpers) and Stage B (extract sub-packages) must be separate stages; each stage must be validated by `make commit` before proceeding |

---

## Success Criteria

**Issue 1 resolved when:**
- `internal/query/jq/` directory no longer exists in the repository.
- `go build ./...` and `go test ./...` pass without modification to any other file.
- No `golangci-yml` exclude rule references `internal/query/jq`.

**Issue 2 resolved when:**
- `handlers_convenience_test.go` contains zero `t.Skip(...)` calls.
- `internal/analysis/service_test.go` contains zero `t.Skip(...)` calls.
- `go test ./...` passes; test count is either unchanged or reduced only by deleted stubs.
- New stub-based tests in `service_test.go` cover all six service methods
  (`AnalyzeBugs`, `AnalyzeErrors`, `QualityScan`, `GetWorkPatterns`, `GetTimeline`,
  `GetTechDebt`) via the `NewWithAnalyzers` injection path.

**Issue 3 resolved when:**
- `internal/query/assistant/` and `internal/query/sequences/` exist as independent packages.
- Each new package exports at least one interface.
- `internal/query` production file count drops from 21 to 19 or fewer.
- Test coverage across the affected packages is ≥ 80%.
- All 21 MCP tool behaviors are preserved (verified by `make commit`).

---

## Reviewer Notes

*Added 2026-03-12 during architect review. Corrections and additions follow.*

### Corrections to original text

**Issue 1 — line count discrepancy**: The Goals section originally stated "648 lines";
the Trade-off table stated "~745 lines". The actual count is **744 lines**
(436 + 212 + 96). Both figures have been corrected to 744 throughout the document.

**Issue 1 — `golangci.yml` scope**: The lint config at `.golangci.yml` contains a
`depguard` rule named `no-query-jq-imports-query` that matches `**/internal/query/jq/**`.
This rule must be deleted (not just the `internal/query/jq` path entry) when the
sub-package is removed; leaving a dangling rule would cause `golangci-lint` to flag
unmatched patterns depending on the linter version.

**Issue 2 — `service_test.go` scope is smaller than stated**: The original proposal said
"adding 5 new analyzer interfaces + stubs." This overstates the work. All six analyzer
interfaces (`BugAnalyzer`, `ErrorAnalyzer`, `QualityScanner`, `WorkPatternsAnalyzer`,
`TimelineAnalyzer`, `TechDebtAnalyzer`) are already defined in
`internal/analyzer/interfaces.go`. The `analysis.Analyzers` struct already has fields for
all six, and `NewWithAnalyzers` already accepts all six. Only five new *stub structs* and
five new test functions in `service_test.go` are required. The `internal/analysis` package
needs no modification.

**Issue 3 — `assistant_messages.go` dependency on `getToolCallTimestamp` is absent**:
The original text stated `assistant_messages.go` depends on both `buildTurnIndex` and
`getToolCallTimestamp`. Code inspection shows `assistant_messages.go` calls only
`buildTurnIndex`. The `getToolCallTimestamp` helper is called exclusively from
`file_access.go` (its own file) and `sequences.go`. The corrected dependency map:

| File | `buildTurnIndex` | `getToolCallTimestamp` |
|---|---|---|
| `assistant_messages.go` | yes | no |
| `sequences.go` | yes | yes |
| `file_access.go` | yes | yes (defined here) |
| `project_state.go` | yes | no |
| `prompts.go` | yes | no |
| `context.go` | yes (defined here) | no |

This means `assistant_messages.go` could be extracted after resolving only
`buildTurnIndex`. However, because `buildTurnIndex` is used by six files, the Stage A
promotion is still necessary before either extraction target can be moved.

### Hidden dependencies not in the original proposal

**`internal/query/files/` sub-package already exists** and is governed by a parallel
`depguard` rule (`no-query-files-imports-query`) in `.golangci.yml`. The Issue 3
sub-package extraction must not introduce a parallel anti-pattern that puts shared helpers
in the parent package while new sub-packages depend on them — that would reproduce the
same violation the `files` rule was written to prevent. Stage A must place promoted
helpers in a *new neutral sub-package* (e.g., `internal/query/turnindex/`) rather than
keeping them in `context.go`, to preserve the architectural constraint.

**`internal/query` total lines (~8,300)**: The production files alone total ~3,200 lines;
test files add ~5,100 lines. Any Stage B extraction that moves a file will require
proportional test migration. Estimated test-file lines for `assistant_messages.go`:
roughly 400–500 lines (the test file currently lives in the same package). This must be
budgeted within the 200-line-per-stage limit — the test migration alone may consume a full
stage.

### Issue 3 feasibility re-assessment

The two-stage approach (Stage A: promote helpers; Stage B: extract sub-packages) is
correct but undersized. A realistic breakdown:

| Stage | Work | Estimated delta |
|---|---|---|
| 3-A | Create `internal/query/turnindex/` with `buildTurnIndex` and `getToolCallTimestamp`; update all 6 callers | ~80 lines changed |
| 3-B | Extract `sequences.go` → `internal/query/sequences/`; migrate its tests | ~100–150 lines |
| 3-C | Extract `assistant_messages.go` → `internal/query/assistant/`; migrate its tests | ~150–200 lines |

Three stages rather than two. Each stage is within the 200-line limit, but Stage 3-C will
be tight. The interfaces proposed in Stage B of the original text (the `AssistantQuerier`
and `SequenceQuerier` snippets) are sound but should be defined in the *calling* package
(`cmd/mcp-server/`) per Go's interface-at-the-consumer idiom, not in the providing
sub-package. Defining them in the providing package is not wrong, but it creates an
unnecessary dependency direction and adds code the provider should not own.

### Over-engineering assessment for Issue 3 interfaces

The proposal recommends defining `AssistantQuerier` and `SequenceQuerier` interfaces as
part of the extraction. This is mildly over-engineered for the current codebase state:
`cmd/mcp-server` is the only caller of these functions, and it currently has no test
doubles for them. Interfaces add value when a caller needs substitution in tests; there is
no evidence that is currently required. A pragmatic alternative: extract the packages
without declaring interfaces first; add interfaces in a follow-on phase only when a
concrete test-double need arises. The extraction itself (moving files, adjusting import
paths) delivers the navigability benefit without the interface overhead.
