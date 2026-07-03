# Plan 31: Split cmd/mcp-server God Package

> **Status: Superseded** — This plan's goals were achieved via Plans 71–75 (Phases 71–75, 2026-03-12).
> Package names differ from what was originally proposed (e.g., `internal/mcp/observability` instead of `internal/observability`),
> but the structural split is complete. See [plans/71-75-mcp-server-final-split.md](71-75-mcp-server-final-split.md).

## Overview

Refactor `cmd/mcp-server` from a 43-file, ~15,500-line god package into six focused `internal/` packages. Each package owns a single concern and is independently testable.

Reference proposal: [docs/proposals/proposal-mcp-server-package-split.md](../proposals/proposal-mcp-server-package-split.md)

**Dependency flow (layered, no cycles):**
```
cmd/mcp-server
  → internal/executor
      → internal/mcptools
      → internal/mcpquery → internal/locator, internal/parser
      → internal/fileref  → internal/config, internal/errors
      → internal/capabilities → internal/config, internal/errors
      → internal/observability → internal/config
```

---

## Phase 31 — `internal/observability`

**Objectives:** Extract tracing, structured logging, and Prometheus metrics from `cmd/mcp-server` into `internal/observability`. This package has no dependencies on other new packages, making it the natural first step.

**Source files being moved:**
- `cmd/mcp-server/tracing.go` (102 lines)
- `cmd/mcp-server/logging.go` (91 lines)
- `cmd/mcp-server/metrics.go` (388 lines)

**Total source lines in scope:** ~581

### Stage 31.1 — Create `internal/observability` with tracing and logging

**Objectives:** Create the new package with `tracing.go` and `logging.go` content. Update `cmd/mcp-server` to import from the new package.

**Criteria:**
- `internal/observability` package compiles
- All functions exported: `InitTracing`, `GetTracer`, `GetTraceID`, `GetSpanID`, `InitLogger`, `NewRequestLogger`, `WithLogger`, `LoggerFromContext`, `ClassifyError`
- `cmd/mcp-server` uses the new package (no duplicate code)
- `make commit` passes

**Stages (TDD):**
1. Write `internal/observability/tracing_test.go` and `internal/observability/logging_test.go` (adapt from existing `tracing_test.go` and relevant sections of `logging.go` tests)
2. Create `internal/observability/tracing.go` (copy + rename package + export)
3. Create `internal/observability/logging.go` (copy + rename + export)
4. Update `cmd/mcp-server/tracing.go` → thin wrapper or direct deletion with import updates
5. Run `make commit`

**Line budget:** ≤200 lines modified

---

### Stage 31.2 — Move `metrics.go` to `internal/observability`

**Objectives:** Move the Prometheus metrics and USE/RED recording functions into `internal/observability`. Keep `init()` registration behaviour.

**Criteria:**
- All metric recording functions exported and callable from `cmd/mcp-server`
- Prometheus `init()` side effect preserved
- `GetErrorSeverity`, `ClassifyResourceError`, `ClassifyTimeoutError` exported
- `StartResourceMonitoring` exported
- `make commit` passes

**Steps:**
1. Write `internal/observability/metrics_test.go`
2. Create `internal/observability/metrics.go` (copy + rename package + export)
3. Update `cmd/mcp-server` files that call recording functions to import from `internal/observability`
4. Delete `cmd/mcp-server/metrics.go` and `cmd/mcp-server/tracing.go` and `cmd/mcp-server/logging.go`
5. Run `make commit`

**Line budget:** ≤200 lines modified

---

## Phase 32 — `internal/mcptools`

**Objectives:** Extract tool schema definitions and registry from `cmd/mcp-server/tools.go` into `internal/mcptools`.

**Source files:**
- `cmd/mcp-server/tools.go` (428 lines)

### Stage 32.1 — Create `internal/mcptools` package

**Criteria:**
- `Tool`, `ToolSchema`, `Property` types exported
- `GetToolDefinitions() []Tool` exported
- `GetToolSchemaByName(name string) (ToolSchema, error)` exported
- Helper builders `StandardToolParameters`, `MergeParameters`, `buildToolSchema`, `buildTool`, `jqFilterWithSchema` exported (prefix with capital letter)
- `cmd/mcp-server/tools.go` replaced with a thin file that re-uses `internal/mcptools`
- Existing tests in `tools_test.go` and `tools_schema_test.go` pass (adapted to new import)
- `make commit` passes

**Steps:**
1. Write `internal/mcptools/tools_test.go` (adapt from `tools_test.go` and `tools_schema_test.go`)
2. Create `internal/mcptools/tools.go`
3. Update `cmd/mcp-server/tools.go` to delegate to `internal/mcptools`
4. Update `cmd/mcp-server/executor.go` imports
5. Run `make commit`

**Line budget:** ≤200 lines modified

---

## Phase 33 — `internal/fileref`

**Objectives:** Extract output mode selection, FileReference generation, temp file management, and content filters into `internal/fileref`.

**Source files:**
- `cmd/mcp-server/output_mode.go` (146 lines)
- `cmd/mcp-server/response_adapter.go` (120 lines)
- `cmd/mcp-server/file_reference.go` (134 lines)
- `cmd/mcp-server/temp_file_manager.go` (167 lines)
- `cmd/mcp-server/filters.go` (current file, ~160 lines)

**Total source lines in scope:** ~727 — split across two stages.

### Stage 33.1 — Create `internal/fileref` with output mode and FileReference

**Criteria:**
- `OutputModeConfig`, `DefaultOutputModeConfig`, `CalculateOutputSize`, `SelectOutputMode`, `SelectOutputModeWithConfig`, `GetOutputModeConfig` exported
- `FileReference`, `GenerateFileReference`, `ExtractFields`, `GenerateSummary` exported
- `cmd/mcp-server/output_mode.go` and `file_reference.go` replaced with thin delegating wrappers
- Existing tests in `output_mode_test.go` and `file_reference_test.go` pass
- `make commit` passes

**Line budget:** ≤200 lines

### Stage 33.2 — Add temp file manager and filters to `internal/fileref`

**Criteria:**
- `WriteJSONLFile`, `CleanupOldFiles`, `CreateTempFilePath`, `ExecuteCleanupTool` exported from `internal/fileref`
- `TruncateMessageContent`, `ApplyContentSummary` exported from `internal/fileref`
- `AdaptResponse`, `BuildInlineResponse`, `BuildFileRefResponse`, `SerializeResponse`, `GetSessionHash` exported
- `cmd/mcp-server/temp_file_manager.go`, `filters.go`, `response_adapter.go` replaced/updated
- Existing tests pass
- `make commit` passes

**Line budget:** ≤200 lines

---

## Phase 34 — `internal/mcpquery`

**Objectives:** Extract the jq query executor, expression cache, streaming file processing, and session directory resolution into `internal/mcpquery`.

**Source files:**
- `cmd/mcp-server/query_executor.go` (248 lines)
- `cmd/mcp-server/handlers_query.go` (155 lines)

**Total source lines:** ~403

### Stage 34.1 — Create `internal/mcpquery` package

**Criteria:**
- `QueryExecutor`, `ExpressionCache`, `QueryResult`, `QueryRequest` exported
- `NewQueryExecutor(baseDir string) *QueryExecutor` exported
- `(*QueryExecutor).ExecuteQuery(scope, jqFilter string, limit int, workingDir string) (QueryResult, error)` exported
- `GetQueryBaseDir(scope, workingDir string) (string, error)` exported
- `GetJSONLFiles(dir string) ([]string, error)` exported
- `cmd/mcp-server/query_executor.go` and `handlers_query.go` updated to delegate
- Existing tests in `query_executor_test.go` and `handlers_query_test.go` pass
- `make commit` passes

**Line budget:** ≤200 lines

### Stage 34.2 — Move convenience handlers to use `internal/mcpquery`

**Criteria:**
- `handlers_convenience.go` updated to call `internal/mcpquery.GetQueryBaseDir` / `(*QueryExecutor).ExecuteQuery`
- `handlers_stage1.go` updated to call `internal/mcpquery.GetQueryBaseDir`, `GetJSONLFiles`
- `handlers_stage2.go` uses updated imports as needed
- Existing tests in `handlers_convenience_test.go` and `handlers_query_session_scope_test.go` pass
- `make commit` passes

**Line budget:** ≤200 lines

---

## Phase 35 — `internal/capabilities`

**Objectives:** Extract capability loading, session cache, GitHub fetch, and package download into `internal/capabilities`. This is the largest source file (~1,171 lines) so it is split into three stages.

**Source file:** `cmd/mcp-server/capabilities.go` (1,171 lines)

### Stage 35.1 — Create `internal/capabilities` with core types and local source loading

**Criteria:**
- `SourceType`, `CapabilityType`, `CapabilitySource`, `CapabilityMetadata`, `CapabilityIndex`, `GitHubSource` exported
- `ParseCapabilitySources`, `DetectSourceType`, `ParseFrontmatter`, `LoadLocalCapabilities` exported
- `ValidateCapabilityType`, `ParseCapabilityReference` exported
- Tests for local source loading pass
- `make commit` passes

**Line budget:** ≤200 lines

### Stage 35.2 — Add GitHub and package source loading to `internal/capabilities`

**Criteria:**
- `LoadGitHubCapabilities`, `ReadGitHubCapability`, `ParseGitHubSource`, `BuildJsDelivrURL` exported
- `LoadPackageCapabilities`, `DownloadPackage`, `ExtractPackage`, `DownloadAndExtractPackage` exported
- `GetPackageCacheDir`, `GetSessionCacheDir`, `CleanupSessionCache` exported
- `RetryWithBackoff`, `IsNotFoundError`, `IsServerError`, `IsNetworkUnreachableError` exported (or unexported if only used internally)
- Package download and GitHub tests pass
- `make commit` passes

**Line budget:** ≤200 lines

### Stage 35.3 — Add capability index cache and tool handlers

**Criteria:**
- `GetCapabilityIndex`, `GetCapabilityContent`, `MergeSources`, `HasLocalSources` exported
- `ExecuteListCapabilitiesTool(cfg *config.Config, args map[string]interface{}) (string, error)` exported
- `ExecuteGetCapabilityTool(cfg *config.Config, args map[string]interface{}) (string, error)` exported
- `cmd/mcp-server/capabilities.go` replaced with thin delegation
- All existing capability tests in `capabilities_test.go`, `capabilities_cache_test.go`, `capabilities_http_test.go`, `capabilities_integration_test.go` pass
- `make commit` passes

**Line budget:** ≤200 lines

---

## Phase 36 — `internal/executor`

**Objectives:** Extract `ToolExecutor` orchestration (per-tool dispatch, pipeline config, jq helpers, metric recording wrappers) into `internal/executor`.

**Source files:**
- `cmd/mcp-server/executor.go` (524 lines)
- `cmd/mcp-server/handlers_convenience.go` (203 lines)
- `cmd/mcp-server/handlers_stage1.go` (361 lines)
- `cmd/mcp-server/handlers_stage2.go` (~100 lines)

**Total source lines:** ~1,188 — split across two stages.

### Stage 36.1 — Create `internal/executor` with ToolExecutor and pipeline helpers

**Criteria:**
- `ToolExecutor`, `NewToolExecutor()`, `(*ToolExecutor).ExecuteTool(cfg, toolName, args) (string, error)` exported
- `toolPipelineConfig`, `newToolPipelineConfig`, `getStringParam`, `getBoolParam`, `getIntParam`, `getFloatParam`, `validateArgKeys`, `injectWarnings` moved to `internal/executor` (some exported, some unexported)
- `buildResponse` and its variants moved
- `cmd/mcp-server/executor.go` reduced to a thin adapter importing `internal/executor`
- Existing executor tests pass
- `make commit` passes

**Line budget:** ≤200 lines

### Stage 36.2 — Move handler functions to `internal/executor`

**Criteria:**
- `handleQueryUserMessages`, `handleQueryTools`, … `handleQueryToolBlocks` moved into `internal/executor` as methods on `ToolExecutor` or as package-level functions
- `handleGetSessionDirectory`, `handleInspectSessionFiles`, `handleGetSessionMetadata`, `handleExecuteStage2Query` moved into `internal/executor`
- `cmd/mcp-server` reduced to `main.go`, `server.go`, and minimal glue
- All tests pass
- `make commit` passes

**Line budget:** ≤200 lines

---

## Phase 37 — Final cleanup and validation

**Objectives:** Remove all scaffolding, verify the final state matches the proposed architecture, and run full test + lint suite.

### Stage 37.1 — Remove redundant forwarding code and validate

**Criteria:**
- `cmd/mcp-server/` contains only: `main.go`, `server.go`, and test files that test integration
- No duplicate function definitions between `cmd/mcp-server` and `internal/` packages
- All `internal/` packages have ≥80% test coverage
- `make push` passes (includes lint)
- ArchGuard analysis shows `cmd/mcp-server` is no longer flagged as a god package

**Steps:**
1. Run `make push` — fix any lint or vet issues
2. Run `make test-coverage` — address coverage gaps
3. Manual review of package dependencies to confirm no cycles
4. Update `docs/core/plan.md` to mark Phase 31-37 complete

**Line budget:** ≤200 lines

---

## Implementation Notes

### Naming convention for exported symbols
When a symbol moves from `package main` to an `internal/X` package, it is exported (uppercase first letter) if called from outside the package. Internal helpers remain unexported.

### Test migration strategy
Each existing `*_test.go` file in `cmd/mcp-server/` tests functions that will move. The strategy is:
1. Copy the test to the new package directory, adjusting the `package` declaration and any import paths
2. Leave the original test in `cmd/mcp-server/` until the implementation move is complete, then either delete it (if the function moved) or update it to test through the new package

### `package main` constraint
`cmd/mcp-server/` files belong to `package main`. They cannot be imported by `internal/` packages. This means `internal/executor` will import `internal/mcpquery`, `internal/fileref`, etc., but never `cmd/mcp-server`.

### Global state
`metrics.go` uses `init()` for Prometheus registration and package-level vars for atomic counters. Moving to `internal/observability` preserves this pattern — the `init()` runs when the package is first imported (which happens at binary startup via `cmd/mcp-server/main.go`).

`capabilities.go` uses `sync.Once` for `sessionCacheDir` and `sync.RWMutex` for `sessionCapabilityCache`. These globals move with the package. Their behaviour is unchanged.

### Backward compatibility
No MCP tool names, parameters, or output formats change. The only observable difference is improved test coverage and package structure visible in `go doc` output.
