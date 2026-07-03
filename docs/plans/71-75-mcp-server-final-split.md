# Plan 71-75: Finish Splitting `cmd/mcp-server`

**Status**: Completed (internal/mcp/{executor,query,filters,observability,tools,response,...} all exist; cmd/mcp-server reduced to ~814 LOC wiring)
**Proposal**: [docs/proposals/proposal-mcp-server-package-split.md](../proposals/proposal-mcp-server-package-split.md)

## Scope

Finish the command-package split based on the repository’s current state, not the older Phase 31 assumptions. This plan assumes:

- capability-loading code is already gone
- `internal/mcp/metrics`, `internal/mcp/pipeline`, `internal/mcp/filters`, and `internal/mcp/schema` already exist
- `cmd/mcp-server` still contains 3,593 LOC of implementation logic that must move into `internal/mcp/*`

## Dependencies and Sequencing

- Phase 71 establishes `internal/mcp/observability` and `internal/mcp/tools`; later phases depend on those APIs.
- Phase 72 (`internal/mcp/response`) is independent of Phase 71 once imports are stable.
- Phase 73 (`internal/mcp/query`) depends only on existing domain packages and may proceed after Phase 71 if desired.
- Phase 74 (`internal/mcp/executor`) depends on Phases 71-73.
- Phase 75 runs last and validates the final package shape.

```text
Phase 71 -> Phase 74 -> Phase 75
Phase 72 ---^
Phase 73 ---^
```

## Phase 71 - Extract Observability and Tool Catalog

### Objectives

Remove two remaining cross-cutting concerns from `cmd/mcp-server`: observability bootstrap and tool catalog construction.

### Acceptance Criteria

- `logging.go` and `tracing.go` logic live under `internal/mcp/observability`
- tool definitions and schema-index glue live under `internal/mcp/tools`
- `cmd/mcp-server` imports these packages instead of owning the logic directly

### Stages

#### Stage 71.1 - Create `internal/mcp/observability`

- Change budget: <=200 lines of logic change
- Tasks:
  - Move tracing and logging helpers from `cmd/mcp-server/logging.go` and `cmd/mcp-server/tracing.go`
  - Export `InitLogger`, `NewRequestLogger`, `WithLogger`, `LoggerFromContext`, `InitTracing`, `GetTracer`, `GetTraceID`, `GetSpanID`, `ClassifyError`
  - Update `main.go`, `server.go`, and `executor.go` imports
- Tests:
  - Move/adapt `logging_test.go` and `tracing_test.go`
  - Run `go test ./cmd/mcp-server ./internal/mcp/observability`
- Exit criteria:
  - `cmd/mcp-server` no longer contains observability implementations

#### Stage 71.2 - Create `internal/mcp/tools` for parameter builders and catalog

- Change budget: <=200 lines of logic change
- Tasks:
  - Move `StandardToolParameters`, `MergeParameters`, `jqFilterWithSchema`, `buildToolSchema`, and `buildTool`
  - Move `getToolDefinitions()` and related catalog helpers into `internal/mcp/tools`
  - Keep tool behavior and schema text unchanged
- Tests:
  - Move/adapt `tools_test.go` coverage for builders and catalog
  - Run `go test ./cmd/mcp-server ./internal/mcp/tools`
- Exit criteria:
  - `cmd/mcp-server/tools.go` is reduced to a thin compatibility shim or deleted

#### Stage 71.3 - Move schema-index lookup glue out of `cmd/mcp-server`

- Change budget: <=200 lines
- Tasks:
  - Move `buildToolSchemaIndex` and `getToolSchemaByName` to `internal/mcp/tools`
  - Update `executor.go`, `server.go`, and related tests to call package APIs
- Tests:
  - Move/adapt `tools_schema_test.go`
  - Run `go test ./cmd/mcp-server ./internal/mcp/tools ./internal/mcp/schema`
- Exit criteria:
  - Tool lookup from `cmd/mcp-server` goes through `internal/mcp/tools`

## Phase 72 - Extract Response and Temp File Infrastructure

### Objectives

Move hybrid output mode, file-ref metadata, temp file writing, and cleanup behavior into a dedicated response package.

### Acceptance Criteria

- output-mode selection and file-ref generation are no longer implemented in `cmd/mcp-server`
- temp-file and session-cache cleanup logic is owned by `internal/mcp/response`
- `cleanup_temp_files` still returns the same result shape

### Stages

#### Stage 72.1 - Create `internal/mcp/response` with output mode and file references

- Change budget: <=200 lines of logic change
- Tasks:
  - Move `output_mode.go` and `file_reference.go` logic
  - Export output-mode configuration and file-reference helpers
  - Update call sites without changing behavior
- Tests:
  - Move/adapt `output_mode_test.go` and `file_reference_test.go`
  - Run `go test ./cmd/mcp-server ./internal/mcp/response`
- Exit criteria:
  - `cmd/mcp-server` no longer owns output-mode selection or file-reference construction

#### Stage 72.2 - Move response adaptation and serialization

- Change budget: <=200 lines
- Tasks:
  - Move `adaptResponse`, `buildInlineResponse`, `buildFileRefResponse`, `serializeResponse`, and `getSessionHash`
  - Keep the hybrid inline/file_ref contract unchanged
- Tests:
  - Move/adapt `response_adapter_test.go`
  - Run `go test ./cmd/mcp-server ./internal/mcp/response`
- Exit criteria:
  - Tool execution builds responses through `internal/mcp/response`

#### Stage 72.3 - Move temp file manager and cleanup tool

- Change budget: <=200 lines of logic change
- Tasks:
  - Move `TempFileManager`, `createTempFilePath`, `writeJSONLFile`, `cleanupOldFiles`, `executeCleanupTool`, `CleanupSessionCache`
  - Update `main.go` shutdown cleanup and `executor.go` special-tool path
- Tests:
  - Move/adapt `temp_file_manager_test.go`
  - Run `go test ./cmd/mcp-server ./internal/mcp/response`
- Exit criteria:
  - `cmd/mcp-server/temp_file_manager.go` is gone or reduced to a thin wrapper pending final cleanup

## Phase 73 - Extract Query Runtime and Stage Query Services

### Objectives

Move query execution, time-range filtering, and stage query services into a dedicated MCP query package.

### Acceptance Criteria

- `QueryExecutor` and expression cache live in `internal/mcp/query`
- time-range parsing and query-base-dir helpers no longer live in `cmd/mcp-server`
- stage-1 metadata helpers and stage-2 query wrapper are callable outside `package main`

### Stages

#### Stage 73.1 - Create `internal/mcp/query` for executor and cache

- Change budget: <=200 lines of logic change
- Tasks:
  - Move `QueryExecutor`, `ExpressionCache`, `QueryRequest`, `QueryResponse`, `QueryResult`
  - Move jq compilation, file streaming, and base-dir helper logic from `query_executor.go`
  - Keep behavior unchanged
- Tests:
  - Move/adapt `query_executor_test.go`
  - Run `go test ./cmd/mcp-server ./internal/mcp/query`
- Exit criteria:
  - `cmd/mcp-server/query_executor.go` is deleted or reduced to a wrapper

#### Stage 73.2 - Move time-range and handler-query support code

- Change budget: <=200 lines
- Tasks:
  - Move `parsedTimeRange`, `parseTimeRange`, and remaining query helper logic from `handlers_query.go`
  - Ensure the query package owns time-bound execution paths
- Tests:
  - Move/adapt `handlers_query_test.go`, `handlers_query_session_scope_test.go`, `handlers_query_workingdir_test.go`
  - Run `go test ./cmd/mcp-server ./internal/mcp/query`
- Exit criteria:
  - `cmd/mcp-server/handlers_query.go` no longer owns query runtime behavior

#### Stage 73.3 - Move stage-1 and stage-2 query services

- Change budget: <=200 lines of logic change
- Tasks:
  - Move `handleGetSessionDirectory`, `handleInspectSessionFiles`, `handleGetSessionMetadata`, `handleExecuteStage2Query`
  - Export replacements from `internal/mcp/query`
- Tests:
  - Move/adapt `handlers_stage1_test.go` and relevant `integration_test.go` coverage
  - Run `go test ./cmd/mcp-server ./internal/mcp/query`
- Exit criteria:
  - stage query helpers are no longer implemented in `cmd/mcp-server`

## Phase 74 - Extract Tool Executor and Convenience Handlers

### Objectives

Move the remaining orchestration logic into `internal/mcp/executor`.

### Acceptance Criteria

- `ToolExecutor` and pipeline config live in `internal/mcp/executor`
- convenience tool handlers no longer live in `cmd/mcp-server`
- `server.go` executes tools through a package API rather than local orchestration logic

### Stages

#### Stage 74.1 - Create `internal/mcp/executor` with ToolExecutor core

- Change budget: <=200 lines of logic change
- Tasks:
  - Move `ToolExecutor`, `toolPipelineConfig`, parameter helpers, scope helpers, success/failure metric wrappers, and special-tool dispatch
  - Update imports to use `internal/mcp/query`, `internal/mcp/response`, `internal/mcp/tools`, `internal/mcp/schema`, `internal/mcp/pipeline`, `internal/mcp/filters`, `internal/mcp/metrics`, and `internal/analysis`
- Tests:
  - Move/adapt `executor_test.go`, `executor_jq_filter_test.go`, `executor_no_cli_test.go`, `executor_phase25_cleanup_test.go`
  - Run `go test ./cmd/mcp-server ./internal/mcp/executor`
- Exit criteria:
  - `cmd/mcp-server/executor.go` is deleted or reduced to a wrapper

#### Stage 74.2 - Move convenience handlers into `internal/mcp/executor`

- Change budget: <=200 lines
- Tasks:
  - Move `handleQueryUserMessages`, `handleQueryTools`, `handleQueryToolErrors`, and the remaining convenience handlers
  - Keep jq expressions and parameter validation unchanged
- Tests:
  - Move/adapt `handlers_convenience_test.go`
  - Run `go test ./cmd/mcp-server ./internal/mcp/executor`
- Exit criteria:
  - convenience handlers are owned by `internal/mcp/executor`

#### Stage 74.3 - Rewire `server.go` and `main.go`

- Change budget: <=200 lines
- Tasks:
  - Replace direct local calls with package calls from `internal/mcp/executor`, `internal/mcp/tools`, and `internal/mcp/observability`
  - Remove obsolete compatibility shims if possible
- Tests:
  - Run `go test ./cmd/mcp-server`
  - Run focused integration cases for `tools/list` and `tools/call`
- Exit criteria:
  - MCP request handling uses only internal package APIs for business logic

## Phase 75 - Final Cleanup and Verification

### Objectives

Remove residual wrappers, stabilize tests, and confirm the command package is thin.

### Acceptance Criteria

- `cmd/mcp-server` implementation files are limited to MCP bootstrap and wiring
- no duplicated implementations remain between `cmd/mcp-server` and `internal/mcp/*`
- ArchGuard no longer reports `cmd/mcp-server` as one of the dominant logic packages

### Stages

#### Stage 75.1 - Remove leftover wrappers and dead files

- Change budget: <=200 lines of logic change
- Tasks:
  - Delete stale wrapper files that exist only for migration convenience
  - Keep only the minimal files needed for startup and JSON-RPC handling
  - Fold `query_runner.go` into its final owner if it still exists
- Tests:
  - Run `go test ./cmd/mcp-server ./internal/mcp/...`
- Exit criteria:
  - no redundant forwarding implementations remain

#### Stage 75.2 - Final verification pass

- Change budget: <=200 lines
- Tasks:
  - Run `go test ./...`
  - Run coverage for affected packages
  - Re-run ArchGuard summary/package graph
  - Update any architecture docs that still describe the pre-split state
- Tests:
  - `go test ./...`
  - `go test ./... -coverprofile=coverage.out`
- Exit criteria:
  - tests pass, moved packages are covered, and docs match the code

## Test Strategy

- Use move-first TDD: migrate tests with the code so each extracted package gets direct package-local coverage.
- Preserve behavior before simplification; no semantic changes in extraction commits.
- Keep integration tests around `cmd/mcp-server` for end-to-end MCP behavior.
- Require strong coverage for each new `internal/mcp/*` package created by this plan, targeting >=80% for moved logic.

## Risks and Mitigations

- Large relocation diffs -> keep logic changes small and isolate pure moves from cleanup.
- Test breakage from package visibility changes -> migrate tests stage-by-stage with compile checkpoints.
- Wrapper creep -> Phase 75 explicitly removes migration shims.
- Stale architectural docs -> final phase includes documentation reconciliation.
