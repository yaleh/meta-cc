# Proposal: Finish Splitting `cmd/mcp-server` into a Thin MCP Entry Point

**Status**: Implemented via Phases 71–75 (internal/mcp/* packages created; cmd/mcp-server reduced to thin wiring)
**Date**: 2026-03-12  
**Supersedes**: the original package-split framing in `docs/plans/31-mcp-server-package-split.md`

## Background

The repository already executed part of the original `cmd/mcp-server` cleanup:

- capability loading was removed in Phases 41–46
- cross-cutting helpers were extracted to `internal/mcp/metrics`, `internal/mcp/pipeline`, `internal/mcp/filters`, and `internal/mcp/schema`
- an `internal/analysis` facade now isolates the analysis tools from direct parser/analyzer coupling

That means the original split proposal is no longer an accurate description of the current codebase. It still plans around `capabilities.go`, packages that no longer exist, and an earlier pre-extraction dependency graph.

The remaining problem is still real. `cmd/mcp-server` is the second-largest package in the repository and continues to own too much business logic.

## Goals

- Reduce `cmd/mcp-server` to MCP bootstrap and JSON-RPC wiring only.
- Move remaining reusable logic into focused `internal/mcp/*` packages that match the current namespace direction.
- Preserve MCP behavior: tool names, parameters, output shapes, and transport semantics must not change.
- Make the remaining server logic easier to unit-test without relying on `package main`.
- Refresh the architecture contract so the proposal matches the repository as it exists today.

## Non-Goals

- Changing the MCP protocol, JSON-RPC framing, or binary names.
- Reworking `internal/query` in this proposal.
- Changing analysis tool behavior or adding new tools.
- Rewriting tool semantics while extracting code.
- Deleting historical plans or proposals that document earlier phases.

## Current State

- `cmd/mcp-server` currently contains 17 implementation `.go` files totaling 3,593 LOC.
- ArchGuard reports `cmd/mcp-server` as the second-largest package in the repository with 17 files, 18 entities, and 69 functions.
- ArchGuard package dependencies show `cmd/mcp-server` still imports `internal/analysis`, `internal/config`, `internal/errors`, `internal/mcp/filters`, `internal/mcp/metrics`, `internal/mcp/pipeline`, `internal/mcp/schema`, `internal/query`, `internal/locator`, `internal/parser`, and `internal/query/files`.

The remaining responsibilities inside `cmd/mcp-server` are still too broad:

1. MCP bootstrap and request dispatch: `main.go`, `server.go`
2. Tool catalog and schema lookup glue: `tools.go`
3. Tool execution orchestration: `executor.go`
4. Query runtime and time-range filtering: `query_executor.go`, `handlers_query.go`
5. Convenience and stage handlers: `handlers_convenience.go`, `handlers_stage1.go`, `handlers_stage2.go`
6. Hybrid response/file-ref/temp-file logic: `output_mode.go`, `file_reference.go`, `response_adapter.go`, `temp_file_manager.go`, `filters.go`
7. Logging and tracing bootstrap: `logging.go`, `tracing.go`

This violates the intended `cmd/` role. The command package should wire dependencies and speak the MCP protocol, not host the bulk of operational logic.

## Proposed Design

### Summary

Finish the split using the current `internal/mcp/*` namespace rather than the older `internal/{executor,mcpquery,fileref,...}` layout. Keep `cmd/mcp-server` limited to startup, shutdown, JSON-RPC types, request parsing, and delegating to internal packages.

### Target Package Layout

```text
cmd/mcp-server
  main.go
  server.go

internal/mcp/observability
  tracing, logging, error classification helpers

internal/mcp/tools
  tool catalog, parameter builders, schema index helpers

internal/mcp/response
  output mode selection, file references, temp JSONL files, response adaptation

internal/mcp/query
  query executor, expression cache, base-dir resolution, time-range parsing/execution,
  stage-1 metadata helpers, stage-2 query wrapper

internal/mcp/executor
  ToolExecutor, pipeline config, special-tool routing, convenience handlers

existing packages retained
  internal/mcp/metrics
  internal/mcp/pipeline
  internal/mcp/filters
  internal/mcp/schema
  internal/analysis
  internal/query
```

### Detailed Changes

#### 1. Move observability bootstrap out of `cmd/mcp-server`

`logging.go` and `tracing.go` are reusable server-support code, not JSON-RPC wiring. They should live beside `internal/mcp/metrics` under `internal/mcp/observability`.

#### 2. Move the tool catalog into `internal/mcp/tools`

`tools.go` still owns standard parameters, tool builders, the tool definition catalog, and schema index glue. `internal/mcp/schema` already owns types and validation; the remaining registry/catalog logic belongs in a dedicated package.

#### 3. Move hybrid response and temp-file behavior into `internal/mcp/response`

`output_mode.go`, `file_reference.go`, `response_adapter.go`, and `temp_file_manager.go` implement output transport strategy, not command wiring. `cleanup_temp_files` should also be routed through that package.

#### 4. Move query runtime into `internal/mcp/query`

`query_executor.go` and the query-specific parts of `handlers_query.go` should move into a package that owns:

- jq compilation and caching
- file streaming
- session base-dir resolution
- RFC3339 time-range parsing and filtering
- stage-1 metadata helpers and stage-2 wrapper functions

This keeps MCP-specific query orchestration out of `cmd/` without forcing it into the broader `internal/query` domain package.

#### 5. Move executor and convenience routing into `internal/mcp/executor`

`executor.go` and the convenience tool handlers still embed most of the business workflow. They should move into a dedicated executor package that depends on `internal/mcp/query`, `internal/mcp/response`, `internal/mcp/tools`, `internal/mcp/schema`, `internal/mcp/pipeline`, `internal/mcp/filters`, and `internal/analysis`.

### Dependency Flow

```text
cmd/mcp-server
  -> internal/mcp/observability
  -> internal/mcp/tools
  -> internal/mcp/executor

internal/mcp/executor
  -> internal/analysis
  -> internal/mcp/query
  -> internal/mcp/response
  -> internal/mcp/schema
  -> internal/mcp/pipeline
  -> internal/mcp/filters
  -> internal/mcp/metrics

internal/mcp/query
  -> internal/locator
  -> internal/parser
  -> internal/query
  -> internal/query/files
```

This proposal intentionally keeps `internal/query` intact. ArchGuard shows it is the largest remaining package, but that is a separate refactor and should not be conflated with finishing the command-package split.

## Alternatives Considered

### Alternative A: Leave `cmd/mcp-server` as the current mixed package

Rejected. The package is still structurally heavy and remains a hotspot in ArchGuard output. This would accept the current drift between intended and actual layering.

### Alternative B: Re-run the original Phase 31 plan literally

Rejected. The old plan assumes `capabilities.go` still exists and proposes package names that no longer match the repository’s `internal/mcp/*` direction. Replaying it now would reintroduce conceptual churn.

### Alternative C: Push all remaining server logic into `internal/query`

Rejected. `internal/query` is already the largest package in the repository. Folding command-specific runtime and server response behavior into it would worsen cohesion rather than improve it.

## Risks

- Test coupling risk: many tests currently target `package main` helpers and will need careful migration.
- Diff-size risk: pure relocations create large file diffs even when logic changes are small.
- Boundary drift risk: if the split stops halfway, the codebase ends up with more wrappers but not thinner ownership.
- Documentation drift risk: old plans and proposals may continue to be read without noticing they are historical.

## Testing and Validation

- Keep extraction stages behavior-preserving: move code first, then simplify.
- Run focused package tests after each stage and `go test ./...` at the end of each phase.
- Require new `internal/mcp/*` packages created by this work to have strong package-local coverage; target >=80% for moved logic even though the repository-wide gate currently enforces a lower threshold.
- Re-run ArchGuard after the final phase and confirm `cmd/mcp-server` is no longer one of the dominant logic containers.

## Open Questions

1. Should `cmd/mcp-server` end with exactly `main.go` and `server.go`, or is one thin adapter file acceptable if it materially simplifies test migration?
2. Should `internal/mcp/query` own the stage-1 and stage-2 helpers, or should those wrappers live under `internal/mcp/executor` after extraction?
3. Should `internal/mcp/response` keep session-cache cleanup, or should temp-file/session-cache concerns split again after the main extraction is complete?
