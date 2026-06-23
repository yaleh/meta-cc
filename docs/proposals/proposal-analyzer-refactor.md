# Proposal: Analyzer Layer Clarification and Dead-Code Removal

**Status**: Implemented (Phase 65; internal/pipeline deleted, see audit 2026-06-23)
**Date**: 2026-03-12
**Author**: Yale Huang
**Reviewed**: 2026-03-12 (architectural review pass — corrections and additions inline)

---

## 1. Background

### 1.1 Context

ArchGuard analysis of the current codebase identified two distinct structural issues:

1. **Dual analyzer implementation**: `internal/analysis.Service` and `internal/analyzer.DefaultAnalyzer` both participate in implementing the six analysis interfaces (`BugAnalyzer`, `ErrorAnalyzer`, `QualityScanner`, `WorkPatternsAnalyzer`, `TimelineAnalyzer`, `TechDebtAnalyzer`), creating apparent responsibility overlap that confuses contributors about which layer owns what.

2. **Dead pipeline package**: `internal/pipeline` (2 files, `SessionPipeline` + `GlobalOptions`/`LoadOptions`) is not imported by any production code. Only its own self-contained `_test.go` references it. A comment in `cmd/mcp-server/handlers_query.go` line 125 mentions it by name, but only in a prose comment — there is no import.

This proposal argues that both issues are **misdiagnoses** when read as requiring new abstractions, but that one concrete cleanup action (removing `internal/pipeline`) is warranted.

---

### 1.2 The Two-Layer Analyzer Architecture

The actual call chain from MCP tool invocation to analysis result is:

```
cmd/mcp-server/executor.go  (ToolExecutor)
  → internal/analysis.Service          [facade: load + dispatch]
      → internal/analyzer.DefaultAnalyzer  [interface adapter]
          → internal/analyzer.<function>() [business logic]
```

Each layer has a distinct and non-overlapping role:

| Layer | File(s) | Role |
|---|---|---|
| `internal/analyzer/<func>.go` | `bugs_analysis.go`, `errors_analysis.go`, `quality_analysis.go`, `work_patterns.go`, `tech_debt.go`, `timeline.go` | Pure functions: receive `[]parser.SessionEntry` + `[]parser.ToolCall`, return typed result structs. No I/O, no file access. Note: `parser.SessionEntry` and `parser.ToolCall` are type aliases for `types.SessionEntry` and `types.ToolCall` (defined in `internal/parser/aliases.go`). |
| `internal/analyzer.DefaultAnalyzer` | `interfaces.go` | Struct that implements the six interfaces by delegating to the package-level functions above. Enables interface-based mocking. |
| `internal/analysis.Service` | `service.go` | Facade: resolves `working_dir`, calls `locator` + `parser` to load session files, invokes the injected analyzer interfaces, marshals results to JSON. Uses `[]types.SessionEntry` and `[]types.ToolCall` directly. |
| `cmd/mcp-server.ToolExecutor` | `executor.go` | Wires `analysis.New()` into the MCP tool dispatch switch; calls `e.analysisSvc.AnalyzeBugs(args)` etc. |

**Type alias note**: The domain interfaces in `internal/analyzer` are declared using `[]parser.SessionEntry` and `[]parser.ToolCall`. The facade in `internal/analysis` uses `[]types.SessionEntry` and `[]types.ToolCall`. These are the same types at compile time — `internal/parser/aliases.go` defines `type SessionEntry = types.SessionEntry` and `type ToolCall = types.ToolCall` (Go type aliases, not new types). There is no type conversion at the boundary. The proposal's earlier draft incorrectly described both layers as uniformly using `types.*` — the actual declarations use `parser.*` in `internal/analyzer/interfaces.go`.

The apparent "dual implementation" arises because `internal/analysis.Service` holds an `Analyzers` struct containing the six interface fields, and `internal/analyzer.DefaultAnalyzer` satisfies all six interfaces. This is **correct composition**: `Service` is the caller, `DefaultAnalyzer` is the callee. The ArchGuard observation that both "implement" the interfaces is accurate but misleading — `Service` does not implement any of the six analyzer interfaces; it owns them as injected dependencies.

Evidence:
- `internal/analysis/service.go:210`: `AnalysisService` interface lists `AnalyzeBugs`, `AnalyzeErrors`, etc. — these are **MCP-level facade methods** that take `map[string]interface{}` args and return `(string, error)`, matching the `cmd/mcp-server` contract.
- `internal/analyzer/interfaces.go:6–33`: The six domain interfaces take typed Go structs (`[]parser.SessionEntry`, `[]parser.ToolCall`) and return typed result structs (`*BugAnalysisResult`, etc.).

The two interface sets have entirely different signatures. There is no duplication.

---

### 1.3 `internal/pipeline` Is Dead Code

`internal/pipeline` was introduced to abstract the "locate → load → extract" pipeline for what was likely an earlier CLI or a planned CLI command interface. It contains:

- `session.go` (119 LOC): `SessionPipeline` struct with `Load`, `ExtractToolCalls`, `BuildTurnIndex`, `Entries`, `EntryCount`, `SessionPath` methods.
- `options.go` (14 LOC): `GlobalOptions` and `LoadOptions` structs carrying CLI-flag values (`SessionID`, `ProjectPath`, `SessionOnly`, `AutoDetect`, `Validate`).
- `session_test.go` (438 LOC, 16 test functions): self-contained tests of the above.

**Evidence of non-use**:
- A full import-graph search across all `*.go` files in the repository finds **zero non-test imports** of `github.com/yaleh/meta-cc/internal/pipeline`.
- `session_test.go` is self-contained (tests only `internal/pipeline` itself).
- The comment in `cmd/mcp-server/handlers_query.go:125` that reads "This matches the behavior of `buildPipelineOptions + SessionPipeline.Load`" is a historical prose note, not an active import.
- The actual session loading in production code is performed by `internal/analysis/service.go:loadData()`, which calls `locator.NewSessionLocator()` and `parser.NewSessionParser()` directly without going through `SessionPipeline`.

The `GlobalOptions` field names (`SessionID`, `ProjectPath`, `SessionOnly`) map to CLI flag semantics that do not exist in the current MCP-only architecture. The `LoadOptions.AutoDetect` and `LoadOptions.Validate` fields are not exercised by any caller outside the package's own tests.

**Behavioral divergence from `loadData()`**: `SessionPipeline.Load` fails fast — it returns an error if any individual session file fails JSONL parsing. `internal/analysis/service.go:loadData()` silently skips malformed files (`continue` on parse error). This divergence makes `SessionPipeline` a subtly different (and stricter) behavior than what production code implements. Reusing `SessionPipeline` in the future would require reconciling this difference.

**`types.SessionLoader` interface**: `internal/types/loader.go` defines a `SessionLoader` interface with three methods: `Entries() []SessionEntry`, `ExtractToolCalls() []ToolCall`, and `BuildTurnIndex() map[string]int`. `SessionPipeline` satisfies this interface structurally. However, no production caller in `internal/query/resources` or elsewhere imports `internal/pipeline` to obtain a `SessionPipeline` as a `SessionLoader` — mock implementations are used in tests instead. This confirms the dead-code status: the interface exists, `SessionPipeline` would satisfy it, but no code wires them together.

**Note**: `internal/mcp/pipeline` is a different package with an entirely different purpose (response-building helpers: `BuildStatsOnlyResponse`, `BuildStatsFirstResponse`, `BuildStandardResponse`, `InjectWarnings`). It is actively imported by `cmd/mcp-server/executor.go` as `pipelinepkg`. This package must not be confused with `internal/pipeline` and must not be touched.

---

## 2. Goals

1. **Eliminate `internal/pipeline`** (dead code): remove `session.go`, `options.go`, and `session_test.go`.
2. **Document the two-layer analyzer boundary** clearly so contributors do not introduce a new abstraction layer or merge the layers.
3. Achieve zero reduction in functionality or test coverage of active code paths.

**Non-goals**:
- Do not merge `internal/analysis.Service` and `internal/analyzer.DefaultAnalyzer` — the two-layer design is intentional.
- Do not change any interface signatures.
- Do not touch `internal/mcp/pipeline`.

---

## 3. Proposed Changes

### 3.1 Remove `internal/pipeline`

Delete the following files:
- `/home/yale/work/meta-cc/internal/pipeline/session.go`
- `/home/yale/work/meta-cc/internal/pipeline/options.go`
- `/home/yale/work/meta-cc/internal/pipeline/session_test.go`

Pre-deletion verification steps:
1. Run `go build ./...` to confirm no import errors exist before deletion.
2. Run `grep -r '"github.com/yaleh/meta-cc/internal/pipeline"' . --include="*.go"` to confirm the result is empty (or limited to the package itself).
3. Delete the three files and the now-empty `internal/pipeline/` directory.
4. Run `go build ./...` again to confirm no compilation errors.
5. Run `make commit` to verify all tests pass.

No source files outside `internal/pipeline/` need to be modified.

**LOC impact**: ~571 lines removed (119 + 14 + 438). No new code required for this stage.

### 3.2 Add Package-Level Documentation to `internal/analyzer`

Add a `doc.go` (or update the package comment in `interfaces.go`) to make the intended design explicit:

```go
// Package analyzer provides the business-logic layer of analysis.
//
// Architecture:
//
//   internal/analysis  (facade)
//      ↓ injects via Analyzers struct
//   internal/analyzer.DefaultAnalyzer  (interface adapter)
//      ↓ delegates to
//   internal/analyzer.<function>()     (pure functions, no I/O)
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
```

This is a documentation-only change that does not affect compilation.

**Rationale for `doc.go` over updating `interfaces.go`**: A dedicated `doc.go` ensures the architecture comment is the first thing `go doc internal/analyzer` displays, is not buried among interface declarations, and survives future refactoring of `interfaces.go`. Both approaches are valid; `doc.go` is preferred for architecture-level documentation.

### 3.3 Update the Stale Comment in `handlers_query.go`

Line 125 of `cmd/mcp-server/handlers_query.go` references `buildPipelineOptions + SessionPipeline.Load` in a comment. After deleting `internal/pipeline`, update this comment to remove the historical reference and describe the actual behavior (uses `locator.AllSessionsFromProject` directly).

Context (lines 124–129, current):
```go
// Project scope: use SessionLocator to find all session files
// This matches the behavior of buildPipelineOptions + SessionPipeline.Load

// AllSessionsFromProject returns the list of session files
// We need to return the directory containing those files
sessionFiles, err := loc.AllSessionsFromProject(projectPath)
```

Line 125 is the only line to remove. Line 124 is already correct and must not be changed. After deletion the block becomes:
```go
// Project scope: use SessionLocator to find all session files

// AllSessionsFromProject returns the list of session files
// We need to return the directory containing those files
sessionFiles, err := loc.AllSessionsFromProject(projectPath)
```

This is a net-negative-one-line change (one comment line deleted).

---

## 4. Design Rationale

### 4.1 Why Not Merge `analysis.Service` and `analyzer.DefaultAnalyzer`?

Merging would collapse two separate concerns into one package:

- **Loading** (I/O, locator, parser) — currently in `internal/analysis`
- **Computing** (pure algorithms over typed structs) — currently in `internal/analyzer`

Keeping them separate enables:
- Unit-testing `internal/analyzer` functions with in-memory data, with no filesystem setup.
- Mocking the six analyzer interfaces in `internal/analysis` tests (already demonstrated in `service_test.go` which injects mock analyzers via `NewWithAnalyzers`).
- Future substitution of the default implementations (e.g., a faster or ML-based analyzer) without changing the facade.

### 4.2 Why Not Repurpose `internal/pipeline` Instead of Deleting It?

`SessionPipeline` duplicates logic that `internal/analysis/service.go:loadData()` already performs. The `GlobalOptions` struct carries CLI-flag semantics (`SessionID`, `ProjectPath`, `SessionOnly`) that have no role in the current MCP server. Keeping the package in the tree creates ongoing maintenance obligation (the tests must continue to pass, the types must stay compatible) for code that serves no caller. The cost-benefit is negative.

Additionally, `SessionPipeline.Load` has a stricter error policy (fail on first bad file) than `loadData()` (skip bad files silently). Any future attempt to unify them would require a deliberate API decision. Deleting the package now removes this source of semantic confusion.

If a CLI command interface is added to meta-cc in the future, `SessionPipeline` can be re-introduced at that time with a design grounded in the new requirements and consistent with the then-current loading policy.

### 4.3 `internal/mcp/pipeline` is Not Affected

`internal/mcp/pipeline` (`pipeline.go`, `pipeline_test.go`) provides response-building utilities for the MCP server's output layer. It imports only `internal/query` and the standard library. It is actively used by `cmd/mcp-server/executor.go` (imported as `pipelinepkg`) and has strong test coverage. No changes to this package are proposed.

### 4.4 `types.SessionLoader` Interface Does Not Resurrect `internal/pipeline`

`internal/types/loader.go` defines a `SessionLoader` interface (`Entries`, `ExtractToolCalls`, `BuildTurnIndex`). `SessionPipeline` would satisfy this interface structurally, but no production caller wires them together. The `internal/query/resources` package uses mock implementations in tests, not `SessionPipeline`. The interface itself is not at risk; it will continue to be satisfied by the concrete loaders that production code already uses.

---

## 5. Trade-off Analysis

| Dimension | Keep `internal/pipeline` | Delete `internal/pipeline` |
|---|---|---|
| Code size | +~571 LOC dead code | Reduced by ~571 LOC |
| Test suite | 16 test functions maintaining code no caller uses | Tests eliminated along with the dead code |
| Import graph | Package exists but unreachable from `cmd/` | Graph shrinks by one node |
| Error policy | Stricter (fail-fast on bad JSONL) vs `loadData()` silently skips | Divergence eliminated |
| Future CLI risk | Package available but diverged from `loadData()` | Package must be re-created; less risk of stale API |
| Contributor clarity | Confusing: two loading abstractions with different error semantics | Clear: one loading path in `loadData()` |

Verdict: deletion is the correct action.

---

## 6. Risk Assessment

| Risk | Likelihood | Mitigation |
|---|---|---|
| `internal/pipeline` is imported somewhere not found by grep | Very low | Verified with import-graph search: zero external imports. Confirm again with `go build ./...` before and after deletion. |
| Deleting package breaks CI | None | Package is not imported by any production path; tests in the package are self-contained and will simply no longer exist |
| Confusion between `internal/pipeline` and `internal/mcp/pipeline` during deletion | Low | Explicitly list the three file paths and the directory to delete in §3.1 |
| Analyzer layer documentation becomes stale again | Medium | Encode the architecture in a `doc.go` file rather than prose, so Go tooling surfaces it |
| `types.SessionLoader` interface loses its only structural implementer | Low | The interface is defined for testability via mocks; removing `SessionPipeline` does not break the interface or any test that uses mocks. Concrete loaders used by production code (`query/resources` path) remain intact. |
| Type alias misunderstanding (`parser.*` vs `types.*`) causes future interface mismatch | Low | Document the alias relationship in `doc.go` (§3.2). The compiler enforces identity at the call site. |

---

## 7. Implementation Plan

This proposal is small enough to be executed in a single phase.

### Phase N: Analyzer Clarity and Dead-Code Removal

**Stage N.1** — Verify and delete `internal/pipeline` (~571 LOC net reduction, no new code)

1. Run `go build ./...` and confirm clean.
2. Run `grep -r '"github.com/yaleh/meta-cc/internal/pipeline"' . --include="*.go"` and confirm no external imports.
3. Delete `internal/pipeline/session.go`, `options.go`, `session_test.go`, and the directory.
4. Run `make commit` and confirm all tests pass.

**Stage N.2** — Update comment and add documentation (~30 LOC net addition)

1. Update the stale comment in `cmd/mcp-server/handlers_query.go:125`.
2. Add `internal/analyzer/doc.go` with the architecture comment from §3.2.
3. Run `make commit`.

Total estimated net change: ~571 LOC removed (3 deleted files + directory), ~19 LOC added (doc.go only; the stale comment is deleted, not replaced). Well within phase limits.

---

## 8. Open Issues (Architectural Review Notes)

The following issues were identified during architectural review of this proposal against the actual codebase and must be resolved before implementation:

1. **LOC count corrected**: The original draft stated "~175 LOC dead code." The actual counts are: `session.go` 119 LOC, `options.go` 14 LOC, `session_test.go` 438 LOC — total 571 LOC. The trade-off table and implementation plan have been updated accordingly.

2. **Test function count corrected**: The original draft stated "17 tests." The actual count is 16 test functions (`grep -c "^func Test" session_test.go`). Updated in §5.

3. **Type alias clarification added**: The original draft described both `internal/analysis` and `internal/analyzer` as using `types.SessionEntry`. This is only true for `internal/analysis`. The `internal/analyzer` interfaces declare `parser.SessionEntry` and `parser.ToolCall`. These resolve to the same underlying type via the alias in `internal/parser/aliases.go`, but the declaration difference is architecturally significant — it means `internal/analyzer` has a dependency on `internal/parser` (for the alias names), not directly on `internal/types`. The `doc.go` proposed in §3.2 now explicitly documents this alias relationship.

4. **`types.SessionLoader` interface addressed**: The original draft did not mention `internal/types/loader.go`'s `SessionLoader` interface, which `SessionPipeline` structurally satisfies. §1.3 and §4.4 now address this to confirm it does not affect the deletion decision.

5. **Error policy divergence documented**: The original draft did not note that `SessionPipeline.Load` fails fast on bad JSONL while `loadData()` skips silently. This behavioral gap is now documented in §1.3 and §4.2 as an additional argument for deletion rather than reuse.

---

## 9. References

- `internal/analysis/service.go` — facade that owns the load + dispatch + marshal responsibility
- `internal/analyzer/interfaces.go` — six domain interfaces and `DefaultAnalyzer` adapter
- `internal/parser/aliases.go` — type aliases: `parser.SessionEntry = types.SessionEntry`, `parser.ToolCall = types.ToolCall`
- `internal/types/loader.go` — `SessionLoader` interface (`Entries`, `ExtractToolCalls`, `BuildTurnIndex`)
- `internal/pipeline/session.go` — dead `SessionPipeline` implementation (119 LOC)
- `internal/pipeline/options.go` — dead `GlobalOptions`/`LoadOptions` types (14 LOC)
- `internal/pipeline/session_test.go` — 16 self-contained tests (438 LOC)
- `internal/mcp/pipeline/pipeline.go` — active response-building helpers (not affected)
- `cmd/mcp-server/executor.go` — `ToolExecutor` wiring, imports `internal/analysis` and `internal/mcp/pipeline`
- `cmd/mcp-server/handlers_query.go:125` — stale comment referencing the deleted package
- Prior proposal: [Architecture Hygiene Phase 2](proposal-architecture-hygiene-phase2.md) — broader structural context
