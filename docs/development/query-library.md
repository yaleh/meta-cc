# Query Library Overview

The `internal/query` package exposes reusable helpers that power both the CLI and the MCP server. The package layers on top of the `pkg/pipeline` session loader and the existing filter/pagination utilities so that query behaviour remains consistent regardless of the entry point.

## Exposed APIs

- `RunToolsQuery` — returns filtered tool calls (`[]parser.ToolCall`) with sentinel errors for session loading and filter validation.
- `RunUserMessagesQuery` — returns enriched user messages with optional context windows.
- `ApplyJQFilter` / `GenerateStats` — shared jq utilities that were previously embedded in the MCP executor.

Both helpers accept strongly typed option structs that mirror CLI flags (`ToolsQueryOptions`, `UserMessagesQueryOptions`). Sentinel errors (`ErrSessionLoad`, `ErrFilterInvalid`, `ErrInvalidPattern`) make it easy for callers to map faults to user-facing error codes.

## Consuming the Library

```go
opts := query.ToolsQueryOptions{
    Pipeline: queryPipelineOptions, // derived from CLI flags or scope parameters
    Status:   "error",
    Limit:    10,
}
results, err := query.RunToolsQuery(opts)
if err != nil {
    // Handle ErrSessionLoad / ErrFilterInvalid etc.
}
formatted, err := internaloutput.FormatOutput(results, "jsonl")
```

When used inside the MCP executor, the same helpers replace the previous subprocess invocation, while the CLI reuses them via `cmd/query_tools.go` and `cmd/query_messages.go`.

## Regression Coverage

The integration tests in `cmd/query_library_compare_test.go` ensure the CLI output matches the library output for representative scenarios (tool queries with limits and user message queries with context). These tests use the shared fixture helpers under the `cmd` package to create temporary Claude sessions.

```
go test ./cmd -run QueryLibrary
```

Running `go test ./...` exercises the entire query stack (library + CLI + MCP) to guard against behavioural drift.
