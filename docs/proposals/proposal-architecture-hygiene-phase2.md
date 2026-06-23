# Proposal: Architecture Hygiene Phase 2 — Five Structural Improvements

**Status**: Implemented (Phases 60–64; TimeRange in types, pkg/ removed, cmd/mcp-server split; see audit 2026-06-23)
**Date**: 2026-03-11
**Author**: Yale Huang

---

> **架构师审查注记** (2026-03-11)
>
> 本文档经过逐条代码核实，发现以下需要修正的问题：
>
> **P1 — 关键数据错误**：
> - 原文说 "18 files, 14,820 LOC" 并以此为基准设置完成目标。实际 `cmd/mcp-server` 有 **47 个 `.go` 文件**，其中实现文件（非测试）约 **4,343 LOC**，测试文件约 10,477 LOC，合计 14,820 LOC。Phase 60 完成目标 "under 3,000 LOC (implementation files only)" 实际是从 4,343 降到 3,000，而非从 14,820 降到 3,000，目标应修正。
> - Stage 60.4 单独搬移 `tools.go`（495 LOC）已超出每 Stage ≤200 行限制，需拆分。
> - Stage 60.1 估算 "~150 lines moved" 严重低估：仅 `expandContextTurns` 一个函数就有约 130 行，加上三个 `build*Response` 方法和 `applyMessageFiltersToData`，实际搬移超过 350 行，需重新规划。
>
> **P5 — 命名与内容不符**：
> - 原标题 "internal/errors Is Structurally Sound" 但实际问题是 `TimeRange` 重复定义（与 `internal/errors` 无关）。
> - 经代码核实，`TimeRange` 存在 **3 处**定义（不是 2 处）：`internal/query/unified_types.go`、`internal/analyzer/errors_analysis.go`、**以及 `cmd/mcp-server/handlers_query.go`**。第三处定义在原文中被完全遗漏，Phase 61 的修复范围不完整。
>
> **P4 — 过时描述**：
> - 原文说 analyzer 函数签名是 `[]parser.SessionEntry`，但 Phase 59 后已迁移为 `[]types.SessionEntry`（已由 `internal/analysis/service.go` 导入 `internal/types` 验证）。代码示例需更新。
>
> **Stage 60.3 — 遗漏的循环依赖风险**：
> - `cmd/mcp-server/query_executor.go` 使用的 `TimeRange` 来自 `cmd/mcp-server/handlers_query.go` 的**本地定义**（第 4 处）。若先搬移 `QueryExecutor` 至 `internal/query/executor`，而 `handlers_query.go` 的 `TimeRange` 尚未迁移，会产生编译依赖问题。建议将 Stage 60.3 与 Phase 61 合并或明确顺序约束。
>
> **Phase 60 Stage 行数重新规划**：见 §3.1 修订版。

---

## 1. Problem Statement

A follow-up archguard analysis (2026-03-11) identified five structural issues that persist after the completion of Phases 58–59 (the original Architecture Hygiene initiative). Unlike Phase 58–59 which addressed `cmd/mcp-server`'s direct `internal/parser` import, the issues below represent deeper structural smells: bloated command packages, circular dependency between a subpackage and its parent, a semantic mismatch in the `pkg/` layer, an untestable analyzer package, and fragmented shared types.

---

### 1.1 P1 — `cmd/mcp-server` is Bloated (47 Files, 4,343 Implementation LOC)

`cmd/mcp-server` has accumulated significant business logic that belongs in `internal/`:

| File | LOC | Responsibilities |
|---|---|---|
| `executor.go` | 781 | Tool dispatch, pipeline orchestration, response building, context expansion |
| `tools.go` | 495 | Tool schema registry, parameter validation, schema definitions |
| `metrics.go` | 388 | Prometheus metrics, RED metrics, resource tracking |
| `query_executor.go` | 361 | jq caching, JSONL streaming, expression compilation |
| `handlers_stage1.go` | 361 | Session directory lookup, file inspection |
| `server.go` | 291 | MCP server wiring |
| `handlers_query.go` | 236 | Query handler dispatch, TimeRange definition |
| `handlers_convenience.go` | 229 | Convenience tool wiring |
| (other implementation files) | ~200 | filters, output_mode, file_reference, response_adapter, etc. |

Total: 47 `.go` files. Implementation files (non-test): ~4,343 LOC. Test files: ~10,477 LOC. Grand total: ~14,820 LOC.

The core issue is that `cmd/` should be a thin entry point. Go's conventional wisdom is that `cmd/` packages wire up dependencies and call `internal/` packages — they do not contain logic. Functions such as `buildResponse`, `expandContextTurns`, `applyMessageFiltersToData`, and `executeSpecialTool` in `executor.go` are business logic, not wiring logic. `QueryExecutor` in `query_executor.go` is a reusable data-access component, not a server-specific concern. `metrics.go` is a cross-cutting concern that should be a standalone `internal/metrics` or `internal/telemetry` package.

**Consequence**: The package is difficult to unit-test in isolation (all logic depends on MCP wiring), violates single responsibility, and makes future extraction into separate binaries harder.

---

### 1.2 P2 — `internal/query/files` Has a Reverse Dependency on `internal/query`

The subpackage `internal/query/files` imports its parent package:

```go
// internal/query/files/file_inspector.go:10
import "github.com/yaleh/meta-cc/internal/query"

// internal/query/files/file_inspector.go:26
TimeRange   query.TimeRange `json:"time_range"`
```

The sole reason for this import is to reuse the `query.TimeRange` struct. A child package importing its parent is an architectural smell: it creates tight coupling, makes the subpackage non-independently-importable, and constrains future refactoring of the parent.

The contrast is instructive: `internal/query/jq` was explicitly designed to avoid this dependency (its package comment states: "It has no dependency on the parent query package, making it independently importable"). The `files` subpackage should follow the same principle.

**Root cause**: `TimeRange` is defined in `internal/query/unified_types.go` rather than in a shared location such as `internal/types`. Duplicate `TimeRange` definitions exist in:
- `internal/analyzer/errors_analysis.go:11` — a local copy used within the analyzer
- `cmd/mcp-server/handlers_query.go:18` — a third copy used in MCP handler time-range filtering

All three duplicates confirm that the type has no canonical home. `internal/types` (which already contains `constants.go`, `session.go`, `toolcall.go`, `query_options.go`, etc.) is the correct location.

---

### 1.3 P3 — `pkg/` Packages Depend on `internal/` Packages

`pkg/output` and `pkg/pipeline` import from `internal/`:

```
pkg/output/chunker.go     → internal/parser
pkg/output/estimator.go   → internal/parser
pkg/output/projection.go  → internal/parser
pkg/output/sort.go        → internal/parser
pkg/output/summary.go     → internal/parser
pkg/output/tsv.go         → internal/parser

pkg/pipeline/session.go   → internal/locator
pkg/pipeline/session.go   → internal/parser
```

The Go module system's `internal/` visibility rule means these `pkg/` packages cannot be used by any external module — they are not actually "public". The `pkg/` convention exists for code intended to be importable by third parties. When `pkg/` depends on `internal/`, that intent is broken: a third-party importing `pkg/output` would fail at build time because it cannot resolve `internal/parser`.

Furthermore, no code outside of `pkg/` itself currently imports these packages (verified by searching for `"github.com/yaleh/meta-cc/pkg"` across all non-`pkg/` `.go` files — zero matches). This raises the question of whether `pkg/` should exist at all in its current form, or whether these packages should be relocated to `internal/`.

**Consequence**: False advertisement — `pkg/` implies external usability, but the dependency on `internal/` makes it impossible.

---

### 1.4 P4 — `internal/analyzer` Has No Interfaces (Zero Testability Seams)

`internal/analyzer` contains 10+ source files and 30+ types, but zero interfaces and zero methods on exported types — all exported functionality is bare functions:

```go
// All these are standalone functions, no receiver.
// Note: signatures use types.SessionEntry and types.ToolCall (post-Phase-59 migration):
func AnalyzeBugs(entries []types.SessionEntry, toolCalls []types.ToolCall, limit int) (*BugAnalysisResult, error)
func AnalyzeErrors(entries []types.SessionEntry, toolCalls []types.ToolCall, limit int) (*ErrorAnalysisResult, error)
func DetectErrorPatterns(entries []types.SessionEntry, toolCalls []types.ToolCall) []ErrorPattern
func QualityScan(entries []types.SessionEntry, toolCalls []types.ToolCall) (*QualityScanResult, error)
func CalculateStats(entries []types.SessionEntry, toolCalls []types.ToolCall) SessionStats
func GetTechDebt(entries []types.SessionEntry, toolCalls []types.ToolCall) (*TechDebtResult, error)
func GetTimeline(entries []types.SessionEntry, limit int) (*TimelineResult, error)
func GetWorkPatterns(entries []types.SessionEntry, toolCalls []types.ToolCall) (*WorkPatternsResult, error)
```

The higher-level `internal/analysis.Service` already wraps these functions, and `internal/analysis.AnalysisService` interface was added in Phase 59.2. However, the underlying analyzer functions remain untestable in isolation: callers cannot inject a mock analyzer or substitute behavior without importing the real implementation.

**Consequence**: Integration tests for `internal/analysis.Service` must instantiate real parsers and session data. There is no way to test higher-level logic with a stub analyzer.

---

### 1.5 P5 — Fragmented Shared Types (`TimeRange` Has No Canonical Home)

`internal/types` exists and already serves as the canonical location for shared data types (`constants.go`, `session.go`, `toolcall.go`, `query_options.go`). However, `TimeRange` was not placed there, resulting in 3 independent definitions:

1. `internal/query/unified_types.go` — used by query handlers
2. `internal/analyzer/errors_analysis.go` — a local copy in the analyzer
3. `cmd/mcp-server/handlers_query.go` — used by MCP handler time-range filtering, with `time.Time` fields instead of strings

Note that the `cmd/mcp-server/handlers_query.go` version uses `time.Time` fields (parsed from RFC3339 strings), while the other two use `string` fields. These are semantically distinct types despite sharing the name. The `internal/types` canonical version should use `string` fields (ISO8601) consistent with the JSON wire format, while `cmd/mcp-server/handlers_query.go`'s `time.Time`-based version is an internal parse-time struct that need not be shared.

**Note on `internal/errors`**: This package is structurally sound — 9 sentinel errors following Go error-wrapping conventions. No changes required to `internal/errors` itself.

---

## 2. Goals

1. **Thin `cmd/mcp-server`**: Move business logic from `cmd/` into `internal/` packages, making `cmd/` a pure wiring layer (target: implementation files <2,500 LOC, excluding tests).
2. **Break the `files` → `query` reverse dependency**: Move `TimeRange` to `internal/types` so `internal/query/files` can import it without touching the parent.
3. **Resolve `pkg/` semantic contradiction**: Either remove `internal/` imports from `pkg/`, or move `pkg/output` and `pkg/pipeline` to `internal/`.
4. **Add testability seams to `internal/analyzer`**: Introduce focused interfaces so callers can substitute behavior in tests.
5. **Canonicalize `TimeRange` in `internal/types`**: Eliminate the 2 string-field duplicates in `internal/query/unified_types.go` and `internal/analyzer/errors_analysis.go`; leave `cmd/mcp-server`'s `time.Time`-based local struct in place (it is a parse-time artifact with a different field type, not the wire-format type).

**Out of scope**: MCP protocol changes, new tool functionality, performance tuning.

---

## 3. Proposed Solution

### 3.1 Phase 60: Extract Business Logic from `cmd/mcp-server`

Split `cmd/mcp-server` responsibilities into focused `internal/` packages. The extraction must proceed in dependency order and respect the ≤200 LOC/stage constraint. Given the actual file sizes, Stage 60.1 and 60.4 in the original plan each exceed the stage limit and are revised here.

#### Stage 60.1 — Extract `internal/mcp/pipeline` (Response Building)

Move the `build*Response` family from `executor.go` into `internal/mcp/pipeline`. These functions have no MCP wire-protocol dependency — they operate on `[]interface{}` data and return strings.

Functions to move:
- `buildStatsOnlyResponse` (~30 lines)
- `buildStatsFirstResponse` (~45 lines)
- `buildStandardResponse` (~25 lines)
- `injectWarnings` (~20 lines)
- `dataToJSONL` (~20 lines)

`buildResponse` itself stays in `executor.go` for now (it coordinates the others and references `toolPipelineConfig`). After Stage 60.1, the `build*` helpers are in `internal/mcp/pipeline` and `buildResponse` calls them via the new package.

Estimated: ~140 lines moved (5 functions) + ~30 lines of new wiring. Stays within ≤200 net modification limit.

#### Stage 60.2 — Extract `internal/mcp/metrics`

Move `metrics.go` to `internal/mcp/metrics` (or `internal/telemetry`). Metrics registration and recording are cross-cutting concerns and must not reside in `cmd/`.

Estimated: ~388 lines moved, ~15 lines of updated imports. This stage moves one file without logic changes — split as "delete + add" in a single commit. The 388 lines of moved code counts as ~0 net new lines (pure relocation), but the git diff will be large. Verify `make commit` passes after this stage before proceeding.

#### Stage 60.3 — Extract `internal/mcp/filters` (Message Processing)

Move message-processing functions from `executor.go` and `filters.go` to `internal/mcp/filters`:
- `applyMessageFiltersToData` (~5 lines, delegates to `ApplyContentSummary` and `TruncateMessageContent`)
- `expandContextTurns` (~130 lines)

**Pre-condition**: Phase 61 (TimeRange canonicalization) must be fully merged before this stage. `internal/query/files/file_inspector.go` imports `internal/query` solely for `TimeRange`; until that import is removed by Phase 61, any code movement that touches `filters.go` or `executor.go` referencing session types risks cascading import errors. Run `go list -f '{{.Imports}}' ./internal/query/files/...` and confirm no `internal/query` reference before starting.

Estimated: ~135 lines moved + ~20 lines of new wiring. Stays within ≤200 net.

#### Stage 60.4a — Extract Tool Schema to `internal/mcp/schema` (Part 1: Types and Validation)

`tools.go` (495 LOC) contains ToolSchema definitions and validation logic. At 495 lines, it exceeds the ≤200 line stage limit and must be split.

Stage 60.4a: Move type definitions and the `validateArgKeys` function (~120 lines of types + ~35 lines of validation = ~155 lines) to `internal/mcp/schema`.

Estimated: ~155 lines moved + ~20 lines of wiring.

#### Stage 60.4b — Extract Tool Schema to `internal/mcp/schema` (Part 2: Schema Registry)

Move the remaining schema registration and `getToolSchemaByName` (~340 lines) to `internal/mcp/schema`. After this stage, `tools.go` in `cmd/mcp-server` becomes a thin shim that calls `internal/mcp/schema`.

Estimated: ~340 lines moved + ~20 lines of wiring.

**Phase 60 completion target**: `cmd/mcp-server` implementation files (excluding tests) reduced from ~4,343 LOC to under 2,500 LOC. All moved packages have ≥80% test coverage. `go build ./...` passes after every stage.

---

### 3.2 Phase 61: Break the `files` → `query` Reverse Dependency and Canonicalize `TimeRange`

Phase 61 is the prerequisite for Stage 60.3. It should be executed before or in parallel with Phase 60.

#### Stage 61.1 — Canonicalize `TimeRange` in `internal/types`

Add `TimeRange` to `internal/types/query_options.go` (which already contains shared query-related types):

```go
// internal/types/query_options.go (addition)
// TimeRange specifies an optional inclusive time window for timestamp filtering.
// Both fields use ISO8601/RFC3339 string format to preserve JSON round-trip fidelity.
type TimeRange struct {
    Start string `json:"start,omitempty"` // ISO8601 timestamp, inclusive lower bound
    End   string `json:"end,omitempty"`   // ISO8601 timestamp, inclusive upper bound
}
```

Note: `cmd/mcp-server/handlers_query.go` defines a distinct `TimeRange` with `time.Time` fields used internally after RFC3339 parsing. That struct is a parse-time artifact and should remain local to `cmd/mcp-server` under a clearer name (e.g., `parsedTimeRange`) to avoid confusion with the wire-format `types.TimeRange`.

#### Stage 61.2 — Update `TimeRange` Consumers

Update the 2 string-field `TimeRange` consumers to import from `internal/types`:

- `internal/query/unified_types.go`: Replace local `TimeRange` definition with a type alias or embedded struct pointing to `types.TimeRange`.
- `internal/query/files/file_inspector.go`: Import `internal/types.TimeRange` instead of `internal/query.TimeRange`.
- `internal/analyzer/errors_analysis.go`: Replace local `TimeRange` struct with `internal/types.TimeRange`.

Rename `cmd/mcp-server/handlers_query.go`'s `TimeRange` struct to `parsedTimeRange` (or `timeFilter`) to eliminate naming ambiguity.

After this stage, `internal/query/files` will have zero imports from `internal/query`, matching the pattern already established by `internal/query/jq`.

Estimated: ~40 lines changed across 5 files (including the rename in `handlers_query.go`).

**Phase 61 completion target**: `go list -f '{{.Imports}}' ./internal/query/files/...` contains no reference to `github.com/yaleh/meta-cc/internal/query`. `grep -rn "type TimeRange" ./internal/` returns exactly one result (in `internal/types`).

---

### 3.3 Phase 62: Resolve the `pkg/` Semantic Contradiction

Two options, presented with tradeoffs:

#### Option A — Move `pkg/` to `internal/`

Rename `pkg/output` → `internal/output` and `pkg/pipeline` → `internal/pipeline`.

**Rationale**: Since no external consumers exist (zero imports outside the module), the `pkg/` designation is aspirational rather than functional. Moving to `internal/` removes the false promise and eliminates the dependency violation.

**Impact**: ~16 files touched (7 implementation + 7 test files in `pkg/output`, 2 files in `pkg/pipeline`), import path updates only. No logic changes. The `pkg/` directory is removed entirely.

#### Option B — Remove `internal/` Imports from `pkg/`

Promote the types that `pkg/output` and `pkg/pipeline` need from `internal/parser` and `internal/locator` into a new `pkg/types` package (or into `internal/types`, then re-export via a thin wrapper).

**Rationale**: Preserves the intent to eventually offer `pkg/output` and `pkg/pipeline` as public packages.

**Impact**: More invasive — requires redesigning type dependencies for `parser.SessionEntry`, `parser.ToolCall`, `locator.SessionLocator`, and related types.

**Recommendation**: Option A. The packages have no external consumers today. If external use becomes a real requirement in the future, a proper API stabilization effort should be done at that time, including semantic versioning and migration guides. Premature `pkg/` placement that violates Go conventions is more harmful than not having a `pkg/` layer at all.

Estimated (Option A): ~40 lines changed (import path updates across ~16 files), 2 directory renames via `git mv`.

**Phase 62 completion target**: `pkg/` directory either removed or contains zero imports of `internal/`. `go build ./...` passes.

---

### 3.4 Phase 63: Add Testability Seams to `internal/analyzer`

#### Stage 63.1 — Define Focused Analyzer Interfaces

Rather than wrapping all 8 exported functions into one monolithic interface (which would create an unimplementable mock), define focused interfaces aligned with the calling patterns in `internal/analysis.Service`. Six functions are called directly by `analysis.Service` and get interfaces: `AnalyzeBugs`, `AnalyzeErrors`, `QualityScan`, `GetTechDebt`, `GetTimeline`, `GetWorkPatterns`. The remaining two — `DetectErrorPatterns` and `CalculateStats` — are either internal helpers or called by other paths; they do not need interfaces in this phase:

```go
// internal/analyzer/interfaces.go

type BugAnalyzer interface {
    AnalyzeBugs(entries []types.SessionEntry, toolCalls []types.ToolCall, limit int) (*BugAnalysisResult, error)
}

type ErrorAnalyzer interface {
    AnalyzeErrors(entries []types.SessionEntry, toolCalls []types.ToolCall, limit int) (*ErrorAnalysisResult, error)
}

// ... one interface per analysis concern
```

Alternatively, define function types that can be substituted in tests:

```go
type BugAnalyzerFunc func(entries []types.SessionEntry, toolCalls []types.ToolCall, limit int) (*BugAnalysisResult, error)
```

The function-type approach has lower boilerplate and is idiomatic Go for simple one-method interfaces. Use it unless the interface approach is needed for a more complex test double.

#### Stage 63.2 — Update `internal/analysis.Service` to Accept Interfaces

Refactor `internal/analysis.Service` to accept analyzer interfaces via constructor injection, rather than calling package-level functions directly. Default implementations can use the existing free functions.

This makes `Service` mockable for higher-level tests without any real session data.

Estimated: ~100 lines of new interface definitions, ~80 lines of refactoring in `internal/analysis/service.go`.

**Phase 63 completion target**: `internal/analysis/service_test.go` can test at least one handler using a stub `ErrorAnalyzer` without loading real session files.

---

### 3.5 Phase 64: Linting Enforcement (Cleanup and Guard Rails)

Phase 64 is a verification and preventive pass after Phases 61–63 complete:

- Verify: `grep -rn "type TimeRange" ./internal/` — should return exactly one result (in `internal/types`).
- Verify: `go list -f '{{.Imports}}' ./internal/query/files/...` — no `internal/query` reference.
- Add a `depguard` linting rule to enforce that `internal/query/files` and `internal/query/jq` do not import `internal/query`.
- Verify `internal/errors` has no imports from `internal/query` or `internal/analyzer` (currently sound, guard against regression).

Estimated: ~10 lines of linting configuration, no production code changes.

---

## 4. Trade-off Analysis

### 4.1 Extraction Risk vs. Regression Risk

Extracting logic from `cmd/mcp-server` into new `internal/` packages (Phase 60) is the highest-risk change. The mitigation strategy is:

1. Move code unchanged first (no refactoring during extraction).
2. Update imports and confirm `make commit` passes.
3. Refactor within the new package in a subsequent stage.

This two-step approach means each stage stays under the 200-line modification limit. Phase 61 (TimeRange) should be completed before Phase 60 Stage 60.3 to avoid cascading import issues.

### 4.2 Option A vs. Option B for `pkg/` (Phase 62)

| Criterion | Option A (Move to `internal/`) | Option B (Remove `internal/` deps) |
|---|---|---|
| Effort | Low (~40 lines, ~16 files) | High (~200+ lines) |
| Risk | Low (renaming only) | Medium (type redesign) |
| External API promise | Abandoned | Preserved |
| Correctness | Accurate | Accurate |
| Future flexibility | Requires new work to re-publish | Ready to publish once stable |

Option A is preferred because: (a) there are zero external consumers today, (b) Option B requires a proper API stabilization effort that should not be done under time pressure, and (c) incorrect `pkg/` placement is worse than no `pkg/` placement.

### 4.3 Interface Granularity for `internal/analyzer` (Phase 63)

A single large `AnalyzerSuite` interface with 8+ methods would be an anti-pattern (tests would need to implement all methods even for single-concern tests). Fine-grained interfaces (one per analysis function) follow the Interface Segregation Principle and enable focused mocking. The tradeoff is more interface definitions, but Go's implicit interface satisfaction makes this low overhead.

Function-type aliases (`type BugAnalyzerFunc func(...)`) are an alternative that avoids even the single-method interface boilerplate. Both approaches are acceptable; choose based on test complexity.

### 4.4 Phase Ordering Constraint

Phase 61 creates a hard dependency for Phase 60 Stage 60.3. Stages 60.1 and 60.2 have no blocking pre-condition and may proceed in parallel with Phase 61. The recommended execution order is:

```
Phase 61 ─────────────────────────────────────────────────────────────────────────┐
Phase 60.1 → Phase 60.2 → (wait for Phase 61) → Phase 60.3 → Phase 60.4a → 60.4b ┘
                                                                                    ↓
                                                              Phase 62, 63 (independent) → Phase 64
```

Phases 62 and 63 are independent of each other and can be executed in parallel or in any order after Phase 60 completes.

---

## 5. Risks

| Risk | Probability | Severity | Mitigation |
|---|---|---|---|
| Phase 60 extraction introduces import cycles | Medium | High | Run `go build ./...` after each stage; use `internal/mcp/` namespace to avoid cycles |
| Stage 60.3 depends on Phase 61 TimeRange work | Medium | High | Complete Phase 61 before Stage 60.3; verify imports before each stage |
| Phase 61 `TimeRange` canonicalization breaks compile | Low | Medium | TDD: add compile-time assertion `var _ = types.TimeRange(query.TimeRange{})` before refactoring; confirm `cmd/mcp-server`'s local `TimeRange` (with `time.Time` fields) is renamed, not replaced |
| Phase 62 directory rename breaks IDE tooling temporarily | Low | Low | Rename via `git mv`, update all imports atomically in one commit |
| Phase 63 interface injection changes `Service` behavior | Low | High | Keep default implementations as thin wrappers over existing free functions; add integration test before refactoring |
| Stage 60 individual stages exceed 200-line limit | Low (with revised plan) | Medium | Stages 60.4a and 60.4b split `tools.go` explicitly; verify line counts before committing each stage |
| `make commit` coverage regression after extraction | Medium | Medium | Ensure moved packages have ≥80% test coverage before next stage; do not defer test writing |

---

## 6. Implementation Order and Phase Sizing

| Phase | Title | Estimated LOC Changed | Key Constraint |
|---|---|---|---|
| 61 | Canonicalize `TimeRange`; break `files` → `query` dep | ~40 (across 5 files) | Low risk; prerequisite for Phase 60 Stage 60.3 |
| 60.1 | Extract response-building to `internal/mcp/pipeline` | ~140 (moved) + ~30 (wiring) = ~170 net | ≤200 net lines |
| 60.2 | Extract metrics to `internal/mcp/metrics` | ~388 (moved, ~0 net new) + ~15 (import updates) | Pure relocation; verify `make commit` passes |
| 60.3 | Extract message processing to `internal/mcp/filters` | ~135 (moved) + ~20 (wiring) = ~155 net | Requires Phase 61 first; after Stage 60.2 |
| 60.4a | Extract tool schema types and validation (Part 1) | ~155 (moved) + ~20 (wiring) = ~175 net | ≤200 net lines; after Stage 60.3 |
| 60.4b | Extract tool schema registry (Part 2) | ~340 (moved, ~0 net new) + ~20 (wiring) | Pure relocation; after Stage 60.4a |
| 62 | Resolve `pkg/` semantic contradiction (Option A) | ~40 (import path updates, ~16 files) | One commit, atomic rename |
| 63 | Add testability seams to `internal/analyzer` | ~180 (interface defs + injection wiring) | Two stages: 63.1 (interfaces) + 63.2 (injection) |
| 64 | Linting enforcement and cleanup | ~10 (config only) | Verification pass |

**Total estimated net code change**: ~770 LOC across 9 stages in 5 phases (see execution order table). All stages respect the ≤200 LOC/stage limit defined in `docs/core/principles.md`. Phase 60 alone (170+0+155+175+0=500 net, treating pure relocations as 0) sits exactly at the ≤500 LOC/phase limit. Phase 61 (40 net), Phase 62 (40 net), Phase 63 (180 net), and Phase 64 (10 net) are all well under the phase limit.

---

## 7. Success Criteria

- `cmd/mcp-server` implementation files (non-test): fewer than 2,500 LOC total.
- `cmd/mcp-server` contains no business logic: remaining files are wiring (`main.go`, `server.go`), thin adapters, and MCP-protocol-specific glue only.
- `go list -f '{{.Imports}}' ./internal/query/files/...` contains no `internal/query` reference.
- `go list -f '{{.Imports}}' ./pkg/...` contains no `internal/` reference (or `pkg/` is removed entirely).
- `internal/analyzer` has at least one exported interface per analysis concern.
- `grep -rn "type TimeRange" ./internal/` returns exactly one result (in `internal/types/query_options.go`).
- `internal/analysis/service_test.go` includes at least one test using a stub analyzer (no real session files loaded).
- `make commit` passes (all tests green, coverage ≥80%).

---

## 8. References

- `docs/proposals/proposal-architecture-hygiene.md` — Phase 58–59 predecessor proposal (implemented)
- `docs/core/principles.md` — Code limits and development methodology
- `internal/query/jq/jq.go` — Reference implementation of an independently-importable subpackage
- `internal/query/jq/stage2_executor.go` — Reference for independently-importable jq executor (overlapping concern with `query_executor.go`)
- `internal/analysis/service.go` — Existing facade and `AnalysisService` interface (Phase 59.2)
- `cmd/mcp-server/executor.go` — Primary extraction target (781 LOC)
- `cmd/mcp-server/handlers_query.go` — Contains local `TimeRange` struct with `time.Time` fields (distinct from the wire-format `TimeRange`)
- `internal/types/query_options.go` — Canonical home for shared query-related types
- `internal/types/` — Canonical home for shared data types
