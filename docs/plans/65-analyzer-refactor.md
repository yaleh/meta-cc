# Plan 65: Analyzer Clarity and Dead-Code Removal

**Status**: Completed (internal/pipeline deleted; see audit 2026-06-23)
**Proposal**: [docs/proposals/proposal-analyzer-refactor.md](../proposals/proposal-analyzer-refactor.md)

---

## Overview

One phase addressing structural issues identified by ArchGuard analysis after Phases 60–64:

| Phase | Scope | Key deliverable |
|---|---|---|
| 65 | Remove `internal/pipeline` dead code; document two-layer analyzer architecture; update stale comment | `internal/pipeline` directory deleted; `internal/analyzer/doc.go` added; `handlers_query.go:125` comment corrected |

**Execution order constraint**: Stage 65.1 (deletion) must complete before Stage 65.2 (comment update and documentation), because Stage 65.2 updates the stale comment that references the deleted package — deleting first ensures the comment update is accurate.

```
Stage 65.1 (delete internal/pipeline) → Stage 65.2 (doc.go + comment fix)
```

**Pre-condition**: None. This phase is independent of all Phases 60–64 and may be executed on any branch that has Phases 60–64 merged.

---

## Phase 65: Analyzer Clarity and Dead-Code Removal

**Goal**: Eliminate the dead `internal/pipeline` package (~571 LOC removed, 16 test functions), add a `doc.go` that makes the two-layer analyzer architecture explicit, and update a stale prose comment in `cmd/mcp-server/handlers_query.go`. After this phase, the codebase has exactly one session-loading path (`internal/analysis/service.go:loadData()`), and contributors reading `go doc internal/analyzer` immediately see the intended layering.

**Pre-condition**: None.

**Estimated LOC**: ~571 lines removed (3 deleted files + directory), ~19 lines added (`doc.go`) plus −1 line (`handlers_query.go` stale comment deleted). Net: ~552 lines removed. Well within the ≤500-line phase limit for additions; deletions are excluded from the line-count constraint.

---

### Stage 65.1 — Verify and Delete `internal/pipeline`

**Goal**: Confirm `internal/pipeline` has no external callers, then delete the package entirely.

**Files to delete**:
- `internal/pipeline/session.go` (119 LOC — `SessionPipeline` struct with `Load`, `ExtractToolCalls`, `BuildTurnIndex`, `Entries`, `EntryCount`, `SessionPath` methods)
- `internal/pipeline/options.go` (14 LOC — `GlobalOptions` and `LoadOptions` structs carrying CLI-flag values)
- `internal/pipeline/session_test.go` (438 LOC — 16 self-contained test functions)
- `internal/pipeline/` directory (now empty)

**Step-by-step procedure**:

1. Verify the build is clean before any deletion:
   ```bash
   go build ./...
   ```
   Expected: no errors.

2. Confirm no external imports exist:
   ```bash
   grep -r '"github.com/yaleh/meta-cc/internal/pipeline"' . --include="*.go"
   ```
   Expected: empty output (or only matches within `internal/pipeline/` itself, which are self-imports — acceptable).

   **Important**: Do not confuse `internal/pipeline` with `internal/mcp/pipeline`. The latter is actively imported by `cmd/mcp-server/executor.go` as `pipelinepkg` and must not be touched.

3. Delete the three files and the directory:
   ```bash
   rm internal/pipeline/session.go
   rm internal/pipeline/options.go
   rm internal/pipeline/session_test.go
   rmdir internal/pipeline/
   ```

4. Verify compilation after deletion:
   ```bash
   go build ./...
   ```
   Expected: no errors (the package had no external callers).

5. Run the full validation:
   ```bash
   make commit
   ```
   Expected: all tests pass. The 16 deleted test functions will no longer appear in test output — this is the intended outcome.

**Estimated LOC**: ~571 lines removed, 0 lines added.

**Stage 65.1 acceptance criteria**:
- `internal/pipeline/` directory does not exist
- `grep -r '"github.com/yaleh/meta-cc/internal/pipeline"' . --include="*.go"` returns no output
- `internal/mcp/pipeline/` directory is unaffected (verify: `ls internal/mcp/pipeline/` lists `pipeline.go` and `pipeline_test.go`)
- `go build ./...` passes
- `make commit` passes

---

### Stage 65.2 — Add `doc.go` and Update Stale Comment

**Goal**: Document the two-layer analyzer architecture in a canonical `doc.go`, and correct the stale prose comment in `cmd/mcp-server/handlers_query.go` that references the now-deleted `internal/pipeline` package.

**TDD approach**: The `doc.go` is documentation-only and does not affect behavior. Verification is via `go doc internal/analyzer` output and `go build ./...`. No new test file is required for this stage — the documentation change has no testable behavior beyond compilation.

**Step 1 — Create `internal/analyzer/doc.go`**

Create the file with the following content:

```go
// Package analyzer provides the business-logic layer of analysis.
//
// Architecture:
//
//	internal/analysis  (facade)
//	   ↓ injects via Analyzers struct
//	internal/analyzer.DefaultAnalyzer  (interface adapter)
//	   ↓ delegates to
//	internal/analyzer.<function>()     (pure functions, no I/O)
//
// Domain interfaces (BugAnalyzer, ErrorAnalyzer, QualityScanner,
// WorkPatternsAnalyzer, TimelineAnalyzer, TechDebtAnalyzer) use
// []parser.SessionEntry and []parser.ToolCall — these are type aliases
// for []types.SessionEntry and []types.ToolCall defined in
// internal/parser/aliases.go. No type conversion occurs at the boundary.
//
// DefaultAnalyzer is a thin adapter that allows cmd/mcp-server tests
// to substitute mock implementations via the six interface types.
// Business logic lives exclusively in the package-level functions;
// DefaultAnalyzer adds no logic of its own.
package analyzer
```

File to create:
- `internal/analyzer/doc.go` (~20 lines)

**Step 2 — Update stale comment in `cmd/mcp-server/handlers_query.go`**

Line 125 of `cmd/mcp-server/handlers_query.go` contains a historical reference to `buildPipelineOptions + SessionPipeline.Load`. After deleting `internal/pipeline`, this comment is stale.

Context (lines 124–129 before the fix):
```go
// Project scope: use SessionLocator to find all session files
// This matches the behavior of buildPipelineOptions + SessionPipeline.Load

// AllSessionsFromProject returns the list of session files
// We need to return the directory containing those files
sessionFiles, err := loc.AllSessionsFromProject(projectPath)
```

Delete line 125 only. Line 124 is already correct and must not be changed. After deletion:
```go
// Project scope: use SessionLocator to find all session files

// AllSessionsFromProject returns the list of session files
// We need to return the directory containing those files
sessionFiles, err := loc.AllSessionsFromProject(projectPath)
```

This is a net-negative-one-line change (one comment line deleted, no new lines added).

File to update:
- `cmd/mcp-server/handlers_query.go` — delete line 125 only (net −1 line)

**Step 3 — Verify**:

```bash
go build ./...
go doc internal/analyzer
make commit
```

Expected:
- `go doc internal/analyzer` displays the architecture comment as the first output.
- `make commit` passes.

**Estimated LOC**: ~20 lines added (`doc.go`), −1 line in `handlers_query.go` (stale comment deleted). Total net addition: ~19 lines.

**Stage 65.2 acceptance criteria**:
- `internal/analyzer/doc.go` exists
- `go doc internal/analyzer` output begins with the architecture comment describing the three-layer call chain
- `cmd/mcp-server/handlers_query.go` line ~125 no longer references `buildPipelineOptions` or `SessionPipeline`
- `grep -r "SessionPipeline\|buildPipelineOptions" ./cmd/ --include="*.go"` returns no output
- `go build ./...` passes
- `make commit` passes

---

## Testing Strategy

This phase is primarily a deletion and documentation change. The testing approach is:

1. **No new production logic** — both stages add zero new business logic. TDD does not require writing tests before deletion or before a `doc.go` file.

2. **Regression prevention via `make commit`** — running `make commit` after each stage is the primary test gate. The test suite must pass with fewer test functions (16 fewer after Stage 65.1).

3. **Coverage**: Active test coverage is unchanged or improved — removing 16 self-contained tests for dead code does not reduce coverage of any production path. Run `make test-coverage` after Stage 65.2 to confirm ≥80% coverage is maintained across all active packages.

4. **Compile-time verification** — `go build ./...` before and after deletion confirms no import edge was missed.

5. **Test failure protocol**: If `make commit` fails after either stage, stop immediately. Document the failure and blockers. Do not proceed to the next stage until resolved.

---

## Dependency Map

```
Stage 65.1 (delete internal/pipeline)
    → No external dependencies. Safe to run on any branch.
    → Blocking for: Stage 65.2 (comment update references the deleted package)

Stage 65.2 (doc.go + comment fix)
    → Requires: Stage 65.1 complete
    → No external dependencies on Phases 60–64
```

`internal/mcp/pipeline` is **not affected** by either stage. It is a distinct package (`internal/mcp/pipeline/pipeline.go`, `pipeline_test.go`) imported by `cmd/mcp-server/executor.go` as `pipelinepkg`. Verify its existence before and after deletion with:
```bash
ls internal/mcp/pipeline/
```

---

## Execution Order Summary

| Order | Stage | Description | LOC change | Pre-condition |
|---|---|---|---|---|
| 1 | 65.1 | Verify and delete `internal/pipeline` | −571 (3 files + dir) | None |
| 2 | 65.2 | Add `internal/analyzer/doc.go`; delete stale comment in `handlers_query.go` | +~19 | Stage 65.1 complete |
| **Total** | | | **−~552 net** | |

Both stages respect ≤200 LOC/stage for additions. Phase 65 net addition (~20 lines) is well under the ≤500 LOC/phase limit. Deletions do not count against the addition limit.

---

## Validation Checklist

- [ ] `internal/pipeline/` directory does not exist
- [ ] `grep -r '"github.com/yaleh/meta-cc/internal/pipeline"' . --include="*.go"` returns no output
- [ ] `internal/mcp/pipeline/` directory is unaffected (contains `pipeline.go` and `pipeline_test.go`)
- [ ] `internal/analyzer/doc.go` exists with the three-layer architecture comment
- [ ] `go doc internal/analyzer` displays the architecture comment first
- [ ] `cmd/mcp-server/handlers_query.go` line ~125 contains no reference to `SessionPipeline` or `buildPipelineOptions`
- [ ] `grep -r "SessionPipeline\|buildPipelineOptions" ./cmd/ --include="*.go"` returns no output
- [ ] `go build ./...` passes after each stage
- [ ] `make commit` passes after each stage
- [ ] `make test-coverage` confirms ≥80% coverage across all active packages
