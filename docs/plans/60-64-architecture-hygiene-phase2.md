# Plan 60–64: Architecture Hygiene Phase 2

**Status**: Mostly Completed (Phase 61 ✅ TimeRange in types; Phase 62 ✅ pkg/ removed; Phase 60 ✅ cmd/mcp-server <900 LOC; Phase 63/64 — verify testability seams and depguard; see audit note 2026-06-23)
**Proposal**: [docs/proposals/proposal-architecture-hygiene-phase2.md](../proposals/proposal-architecture-hygiene-phase2.md)

---

## Overview

Five phases addressing structural issues identified by archguard analysis after Phases 58–59:

| Phase | Scope | Key deliverable |
|---|---|---|
| 61 | Canonicalize `TimeRange` in `internal/types`; break `files` → `query` reverse dependency | `internal/query/files` has zero imports of `internal/query` |
| 60 | Extract business logic from `cmd/mcp-server` into focused `internal/mcp/*` packages | `cmd/mcp-server` implementation files reduced to <2,500 LOC |
| 62 | Move `pkg/output` and `pkg/pipeline` to `internal/` | `pkg/` directory removed; no dependency violation |
| 63 | Add testability seams to `internal/analyzer` | `internal/analysis/service_test.go` can test with stub analyzer |
| 64 | Linting enforcement and guard rails | `depguard` rules prevent recurrence of subpackage coupling violations |

**Execution order constraint**: Phase 61 must complete before Phase 60 Stage 60.3. Stages 60.1 and 60.2 have no blocking pre-condition and may run in parallel with Phase 61. Phases 62 and 63 are independent of each other and may proceed in any order after Phase 60 completes. Phase 64 runs last.

```
Phase 61 ─────────────────────────────────────────────────┐
                                                           ↓ (required by 60.3)
Phase 60.1 → Phase 60.2 → Phase 60.3 → Phase 60.4a → Phase 60.4b
                                                           ↓
                          Phase 62 (independent, after Phase 60)
                          Phase 63 (independent, after Phase 60)
                                                           ↓
                          Phase 64 (verification pass, after Phases 61–63)
```

---

## Phase 61: Canonicalize `TimeRange` and Break Reverse Dependency

**Goal**: Add `TimeRange` to `internal/types` as the single canonical definition. Update all string-field consumers to import from `internal/types`. After this phase, `internal/query/files` has zero imports of `internal/query`, matching the pattern established by `internal/query/jq`.

**Pre-condition**: None. This is the earliest phase to execute; it unblocks Phase 60 Stage 60.3.

**Estimated LOC**: ~40 lines across 5 files.

### Stage 61.1 — Add `TimeRange` to `internal/types`

Add the canonical `TimeRange` struct to `internal/types/query_options.go`.

Files:
- `internal/types/query_options.go` — append `TimeRange` struct with `string` fields

```go
// TimeRange specifies an optional inclusive time window for timestamp filtering.
// Both fields use ISO8601/RFC3339 string format to preserve JSON round-trip fidelity.
type TimeRange struct {
    Start string `json:"start,omitempty"` // ISO8601 timestamp, inclusive lower bound
    End   string `json:"end,omitempty"`   // ISO8601 timestamp, inclusive upper bound
}
```

Add a compile-time assertion in a new `internal/types/types_test.go` (outside the `types` package, in package `types_test`) to verify the type and field names are correct:
```go
// Compile-time assertion: TimeRange exists with correct string fields
var _ types.TimeRange = types.TimeRange{Start: "", End: ""}
```

Run `make dev` to verify compilation.

Estimated: ~15 lines

### Stage 61.2 — Update `TimeRange` Consumers

Update the 2 string-field `TimeRange` definitions to use `internal/types.TimeRange`. Rename `cmd/mcp-server/handlers_query.go`'s `time.Time`-based local struct to `parsedTimeRange` to eliminate naming confusion — that struct is a parse-time artifact and must remain local.

Files:
- `internal/query/unified_types.go` — remove local `TimeRange` definition; replace with type alias `type TimeRange = types.TimeRange` or direct reference `types.TimeRange`; add `internal/types` import
- `internal/query/files/file_inspector.go` — replace `query.TimeRange` reference with `types.TimeRange`; remove `internal/query` import; add `internal/types` import
- `internal/analyzer/errors_analysis.go` — remove local `TimeRange` struct; add `internal/types` import; update all usages to `types.TimeRange`
- `cmd/mcp-server/handlers_query.go` — rename local `TimeRange` struct to `parsedTimeRange`; update all references to `parsedTimeRange` within the file

Run `make commit` after this stage.

Estimated: ~25 lines across 4 files

**Phase 61 acceptance criteria**:
- `grep -rn "type TimeRange" ./internal/` returns exactly one result (in `internal/types/query_options.go`)
- `go list -f '{{.Imports}}' ./internal/query/files/...` contains no reference to `github.com/yaleh/meta-cc/internal/query`
- `make commit` passes

---

## Phase 60: Extract Business Logic from `cmd/mcp-server`

**Goal**: Move business logic out of `cmd/mcp-server` into focused `internal/mcp/*` packages. After this phase, `cmd/mcp-server` contains only wiring (`main.go`, `server.go`), thin adapters, and MCP-protocol-specific glue. Implementation files (non-test) reduced from ~4,343 LOC to under 2,500 LOC.

**Extraction rule**: Move code unchanged first. No refactoring during extraction. Update imports and confirm `make commit` passes before proceeding to the next stage.

**Pre-conditions**:
- No blocking pre-condition for Stages 60.1 and 60.2.
- Stage 60.3 requires Phase 61 to be fully merged (verify import graph before starting).

**Estimated LOC**: ~500 lines net across all stages (under the ≤500 LOC/phase limit).

### Stage 60.1 — Extract Response Building to `internal/mcp/pipeline`

Move the `build*Response` helper family from `executor.go` to a new package `internal/mcp/pipeline`. These functions have no MCP wire-protocol dependency — they operate on `[]interface{}` data and return strings.

Functions to move:
- `buildStatsOnlyResponse` (~30 lines)
- `buildStatsFirstResponse` (~45 lines)
- `buildStandardResponse` (~25 lines)
- `injectWarnings` (~20 lines)
- `dataToJSONL` (~20 lines)

Note: `buildResponse` itself stays in `executor.go` — it coordinates the helpers and references `toolPipelineConfig`. After this stage it calls `internal/mcp/pipeline` for the helpers.

Files to create:
- `internal/mcp/pipeline/pipeline.go` — new package, moved functions
- `internal/mcp/pipeline/pipeline_test.go` — unit tests for moved functions

Files to update:
- `cmd/mcp-server/executor.go` — add `internal/mcp/pipeline` import; update `buildResponse` to call `pipeline.*`; remove the 5 helper function bodies

Run `make commit`.

Estimated: ~140 lines moved + ~30 lines wiring = ~170 net lines

### Stage 60.2 — Extract Metrics to `internal/mcp/metrics`

Move `cmd/mcp-server/metrics.go` (388 LOC) to `internal/mcp/metrics`. Metrics registration and recording are cross-cutting concerns and must not reside in `cmd/`.

This stage is a pure relocation — no logic changes. The git diff will be large (388 lines deleted + 388 lines added), but net new code is near zero.

Files to create:
- `internal/mcp/metrics/metrics.go` — moved content, package renamed to `metrics`
- `internal/mcp/metrics/metrics_test.go` — unit tests if not already present

Files to update:
- `cmd/mcp-server/metrics.go` — delete (replaced by subpackage)
- All files in `cmd/mcp-server/` that reference metrics types — update import path

Run `make commit`. Verify `make commit` passes before proceeding to Stage 60.3.

Estimated: ~388 lines moved (~0 net new logic) + ~15 lines import updates

### Stage 60.3 — Extract Message Processing to `internal/mcp/filters`

**Pre-condition**: Phase 61 must be merged. Verify: `go list -f '{{.Imports}}' ./internal/query/files/...` shows no `internal/query` before starting.

Move message-processing functions from `executor.go` (and any existing `filters.go`) to `internal/mcp/filters`:
- `applyMessageFiltersToData` (~5 lines, delegates to `ApplyContentSummary` and `TruncateMessageContent`)
- `expandContextTurns` (~130 lines)

Files to create:
- `internal/mcp/filters/filters.go` — new package, moved functions
- `internal/mcp/filters/filters_test.go` — unit tests, including table-driven tests for `expandContextTurns`

Files to update:
- `cmd/mcp-server/executor.go` — add `internal/mcp/filters` import; replace inline function bodies with calls to `filters.*`

Run `make commit`.

Estimated: ~135 lines moved + ~20 lines wiring = ~155 net lines

### Stage 60.4a — Extract Tool Schema Types and Validation to `internal/mcp/schema` (Part 1)

`tools.go` (495 LOC) contains ToolSchema definitions and validation logic. At 495 lines it exceeds the ≤200-line stage limit and is split into two stages.

Stage 60.4a: Move type definitions and validation to `internal/mcp/schema`.

Content to move:
- ToolSchema type definitions (~120 lines)
- `validateArgKeys` function (~35 lines)

Files to create:
- `internal/mcp/schema/schema.go` — types and validation
- `internal/mcp/schema/schema_test.go` — unit tests for `validateArgKeys`

Files to update:
- `cmd/mcp-server/tools.go` — add `internal/mcp/schema` import; remove moved type definitions; update references to `schema.*`

Run `make commit`.

Estimated: ~155 lines moved + ~20 lines wiring = ~175 net lines

### Stage 60.4b — Extract Tool Schema Registry to `internal/mcp/schema` (Part 2)

Move the remaining schema registration and `getToolSchemaByName` to `internal/mcp/schema`. After this stage, `tools.go` in `cmd/mcp-server` becomes a thin shim.

Content to move:
- Schema registry map and registration (~220 lines)
- `getToolSchemaByName` function (~120 lines)

Files to update:
- `internal/mcp/schema/schema.go` — append registry and lookup function
- `internal/mcp/schema/schema_test.go` — add tests for `getToolSchemaByName`
- `cmd/mcp-server/tools.go` — remove moved content; retain only thin wrapper that calls `schema.*`

Run `make commit`.

Estimated: ~340 lines moved + ~20 lines wiring (pure relocation, ~0 net new logic)

**Phase 60 acceptance criteria**:
- `cmd/mcp-server` implementation files (non-test): fewer than 2,500 LOC total
- New packages `internal/mcp/pipeline`, `internal/mcp/metrics`, `internal/mcp/filters`, `internal/mcp/schema` exist with ≥80% test coverage each
- `go build ./...` passes after every stage
- `make commit` passes at end of phase

---

## Phase 62: Resolve `pkg/` Semantic Contradiction

**Goal**: Remove the structural violation where `pkg/output` and `pkg/pipeline` import from `internal/`. Implement Option A from the proposal: move both packages to `internal/`.

**Rationale**: No external consumers exist (zero imports of `github.com/yaleh/meta-cc/pkg` outside the module). Moving to `internal/` removes the false public-API promise and eliminates the dependency violation. Future re-publication should be done with proper API stabilization.

**Pre-condition**: Phase 60 complete (no blocking dep, but sequential for safety).

**Estimated LOC**: ~40 lines (import path updates across ~16 files).

### Stage 62.1 — Move `pkg/` Packages to `internal/` (Atomic Rename)

Execute as a single atomic commit to avoid a half-renamed state that breaks compilation.

Renames:
- `pkg/output/` → `internal/output/`
- `pkg/pipeline/` → `internal/pipeline/`

Use `git mv` for the directory renames to preserve history.

Files to update (import path changes only, no logic):
- All files in the moved packages: change `package` declaration if needed; update any self-referential imports
- All files outside the moved packages that import `github.com/yaleh/meta-cc/pkg/output` or `github.com/yaleh/meta-cc/pkg/pipeline` — update import paths to `internal/output` and `internal/pipeline`
- Search: `grep -rn '"github.com/yaleh/meta-cc/pkg/' .` to find all consumers

After the rename, verify `pkg/` directory is empty and remove it.

Run `make commit`.

Estimated: ~40 lines across ~16 files (import path updates only)

**Phase 62 acceptance criteria**:
- `pkg/` directory does not exist (or is empty with no `.go` files)
- `go list -f '{{.Imports}}' ./internal/output/... ./internal/pipeline/...` shows no `internal/parser` or `internal/locator` violations (they are now peer `internal/` packages, which is acceptable)
- `make commit` passes

---

## Phase 63: Add Testability Seams to `internal/analyzer`

**Goal**: Introduce focused interfaces to `internal/analyzer` so callers (primarily `internal/analysis.Service`) can substitute behavior in tests without loading real session files.

**Pre-condition**: Phase 60 complete (no blocking dep, but sequential for safety).

**Estimated LOC**: ~180 lines across two stages.

### Stage 63.1 — Define Focused Analyzer Interfaces

Create `internal/analyzer/interfaces.go` with one interface per analysis concern. Use the function-type approach for simplicity unless a two-method interface is needed:

```go
// internal/analyzer/interfaces.go

type BugAnalyzer interface {
    AnalyzeBugs(entries []types.SessionEntry, toolCalls []types.ToolCall, limit int) (*BugAnalysisResult, error)
}

type ErrorAnalyzer interface {
    AnalyzeErrors(entries []types.SessionEntry, toolCalls []types.ToolCall, limit int) (*ErrorAnalysisResult, error)
}

type QualityScanner interface {
    QualityScan(entries []types.SessionEntry, toolCalls []types.ToolCall) (*QualityScanResult, error)
}

type WorkPatternsAnalyzer interface {
    GetWorkPatterns(entries []types.SessionEntry, toolCalls []types.ToolCall) (*WorkPatternsResult, error)
}

type TimelineAnalyzer interface {
    GetTimeline(entries []types.SessionEntry, limit int) (*TimelineResult, error)
}

type TechDebtAnalyzer interface {
    GetTechDebt(entries []types.SessionEntry, toolCalls []types.ToolCall) (*TechDebtResult, error)
}
```

Scope note: `internal/analyzer` also exports `DetectErrorPatterns` and `CalculateStats` as free functions. These are not wrapped into interfaces here because `internal/analysis.Service` does not call them directly (they are used internally or by other callers). If future callers need mockable versions, additional interfaces can be added without breaking existing code.

Add a default implementation struct that wraps the existing free functions:

```go
// DefaultAnalyzer implements all interfaces by delegating to package-level functions.
type DefaultAnalyzer struct{}
```

Add compile-time assertions for all interfaces:
```go
var _ BugAnalyzer          = (*DefaultAnalyzer)(nil)
var _ ErrorAnalyzer        = (*DefaultAnalyzer)(nil)
// ... etc.
```

Files to create:
- `internal/analyzer/interfaces.go` — interface definitions + `DefaultAnalyzer` type
- `internal/analyzer/interfaces_test.go` — compile-time assertions

Run `make dev`.

Estimated: ~100 lines

### Stage 63.2 — Update `internal/analysis.Service` to Accept Interfaces

Refactor `internal/analysis.Service` to accept analyzer interfaces via constructor injection rather than calling package-level functions directly. Default constructor uses `analyzer.DefaultAnalyzer{}`.

Files to update:
- `internal/analysis/service.go` — add interface fields to `Service` struct; update `New()` to accept options or use `NewWithAnalyzers(...)` constructor with defaults; update method bodies to call injected interfaces instead of `analyzer.*` free functions
- `internal/analysis/service_test.go` — add at least one test using a stub `ErrorAnalyzer` (inline struct implementing the interface) without loading any real session files; this is the key acceptance test for this phase

Run `make commit`.

Estimated: ~80 lines

**Phase 63 acceptance criteria**:
- `internal/analyzer/interfaces.go` exists and exports at least 6 interfaces (one per analysis concern)
- `var _ <Interface> = (*DefaultAnalyzer)(nil)` compile-time assertions pass for all interfaces
- `internal/analysis/service_test.go` contains at least one test using a stub `ErrorAnalyzer` (no real session files loaded)
- `make commit` passes

---

## Phase 64: Linting Enforcement

**Goal**: Add linting rules that prevent recurrence of the subpackage coupling violations fixed in Phases 60–63. Verify all structural fixes hold.

**Pre-condition**: Phases 61, 62, and 63 all complete.

**Estimated LOC**: ~10 lines (linting configuration only, no production code changes).

### Stage 64.1 — Add `depguard` Rules and Run Final Verification

Add `depguard` configuration to `.golangci.yml` (or equivalent linting config) to enforce:
- `internal/query/files` must not import `internal/query`
- `internal/query/jq` must not import `internal/query`
- `internal/errors` must not import `internal/query` or `internal/analyzer`

Run the full verification checklist:

```bash
# TimeRange canonicalization
grep -rn "type TimeRange" ./internal/
# Expected: exactly one result in internal/types/query_options.go

# files subpackage independence
go list -f '{{.Imports}}' ./internal/query/files/...
# Expected: no reference to github.com/yaleh/meta-cc/internal/query

# pkg/ removed
ls pkg/ 2>/dev/null || echo "pkg/ not present"

# Full build and test
make commit
```

Files to update:
- `.golangci.yml` (or linting config file) — add `depguard` deny rules for the three package pairs above

Run `make push` (full verification including lint).

Estimated: ~10 lines in linting config

**Phase 64 acceptance criteria**:
- `grep -rn "type TimeRange" ./internal/` returns exactly one result
- `go list -f '{{.Imports}}' ./internal/query/files/...` contains no `internal/query`
- `depguard` rules are active and fail on any future violation
- `make push` passes (all tests, lint, coverage)

---

## Testing Strategy

All phases follow TDD:

1. **Write tests before implementation** for new packages (`internal/mcp/pipeline`, `internal/mcp/metrics`, `internal/mcp/filters`, `internal/mcp/schema`, `internal/analyzer/interfaces`)
2. **Coverage requirement**: ≥80% per new/modified package; verify with `make test-coverage` after each stage
3. **Compile-time assertions**: Add `var _ Interface = (*Impl)(nil)` patterns for all new interfaces (Stages 61.1, 63.1)
4. **Regression prevention**: `make commit` must pass after every stage; do not defer test writing to the end of a phase
5. **Integration tests**: Use `testing.Short()` guard for any tests requiring real session files; pure unit tests must work without session data

**Test failure protocol**: If `make commit` fails repeatedly after a stage, stop immediately. Document the failure state and blockers. Do not proceed to the next stage until resolved.

---

## Execution Order Summary

| Order | Phase / Stage | Description | LOC | Pre-condition |
|---|---|---|---|---|
| 1 | 61.1 | Add `TimeRange` to `internal/types` | ~15 | None |
| 2 | 61.2 | Update `TimeRange` consumers | ~25 | Stage 61.1 |
| 3 | 60.1 | Extract response building to `internal/mcp/pipeline` | ~170 | None |
| 4 | 60.2 | Extract metrics to `internal/mcp/metrics` | ~0 net | None |
| 5 | 60.3 | Extract message processing to `internal/mcp/filters` | ~155 | Phase 61 merged |
| 6 | 60.4a | Extract tool schema types/validation (Part 1) | ~175 | Stage 60.3 complete |
| 7 | 60.4b | Extract tool schema registry (Part 2) | ~0 net | Stage 60.4a |
| 8 | 62.1 | Move `pkg/` to `internal/` | ~40 | Phase 60 complete |
| 9 | 63.1 | Define focused analyzer interfaces | ~100 | Phase 60 complete |
| 10 | 63.2 | Inject interfaces into `analysis.Service` | ~80 | Stage 63.1 |
| 11 | 64.1 | Linting rules + final verification | ~10 | Phases 61–63 complete |
| **Total** | | | **~770 lines** | |

All stages respect ≤200 LOC/stage. Phase 60 net change (170+0+155+175+0=500) sits exactly at the ≤500 LOC/phase limit. Phase 61 (~40 net), Phase 62 (~40 net), Phase 63 (~180 net), and Phase 64 (~10 net) are all well under the phase limit.

---

## Validation Checklist

- [ ] `cmd/mcp-server` implementation files (non-test): fewer than 2,500 LOC total
- [ ] `internal/mcp/pipeline`, `internal/mcp/metrics`, `internal/mcp/filters`, `internal/mcp/schema` packages exist
- [ ] All new `internal/mcp/*` packages have ≥80% test coverage
- [ ] `grep -rn "type TimeRange" ./internal/` returns exactly one result (in `internal/types/query_options.go`)
- [ ] `go list -f '{{.Imports}}' ./internal/query/files/...` contains no `internal/query` reference
- [ ] `cmd/mcp-server/handlers_query.go` local struct renamed from `TimeRange` to `parsedTimeRange`
- [ ] `pkg/` directory removed (or contains zero `.go` files)
- [ ] `internal/analyzer/interfaces.go` exports at least 6 focused interfaces
- [ ] `var _ <Interface> = (*DefaultAnalyzer)(nil)` passes for all interfaces
- [ ] `internal/analysis/service_test.go` includes at least one test with a stub analyzer (no real session files)
- [ ] `depguard` rules active in linting config
- [ ] `make push` passes (all tests green, lint clean, coverage ≥80%)
