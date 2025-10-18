# Logging Standards for meta-cc

**Version**: 1.0
**Date**: 2025-10-17
**Framework**: Go log/slog
**Applies To**: meta-cc MCP server and internal modules

---

## Table of Contents

1. [When to Log](#when-to-log)
2. [Log Levels](#log-levels)
3. [Structured Field Naming](#structured-field-naming)
4. [Log Statement Patterns](#log-statement-patterns)
5. [Context Propagation](#context-propagation)
6. [Performance Guidelines](#performance-guidelines)
7. [Common Pitfalls](#common-pitfalls)

---

## When to Log

### DO Log:

✅ **Error Conditions** (ERROR level):
- All error returns (`if err != nil`)
- Unexpected states or conditions
- I/O failures (file, network)
- Parsing failures (invalid JSON, malformed input)

✅ **Significant Operations** (INFO level):
- Tool invocation start/completion
- Query execution start/completion
- Server lifecycle events (startup, shutdown)
- Configuration loading

✅ **Diagnostic Flow** (DEBUG level):
- Function entry/exit (in complex paths)
- State transitions
- Variable values at key decision points

✅ **Warning Conditions** (WARN level):
- Fallback to defaults
- Deprecated feature usage
- Performance degradation

### DON'T Log:

❌ **Avoid Excessive Logging**:
- Inside tight loops (log aggregated results instead)
- Every function call (log significant ones only)
- Sensitive data (passwords, tokens, API keys)
- Redundant information (already logged upstream)

❌ **Avoid Unstructured Logs**:
- String formatting with fmt.Sprintf() (use structured fields)
- Concatenated messages (use key-value pairs)
- Missing context (always include request_id when available)

---

## Log Levels

### DEBUG

**When to Use**: Detailed diagnostic information for development and debugging

**Use Cases**:
- Function entry/exit for complex flows
- Variable values at key points
- Internal state transitions
- Detailed error context (stack traces)

**Production**: Disabled (set `LOG_LEVEL=INFO`)

**Example**:
```go
logger.Debug("parsing session file",
    "file_path", filePath,
    "file_size", fileSize,
    "session_id", sessionID,
)
```

### INFO

**When to Use**: General informational messages about normal operations

**Use Cases**:
- Tool invocation start/success
- Query execution start/success
- Server startup/configuration loaded
- Significant milestones

**Production**: Enabled (default level)

**Example**:
```go
logger.Info("tool execution started",
    "request_id", requestID,
    "tool_name", "query_tools",
    "scope", "project",
)
```

### WARN

**When to Use**: Warning conditions that don't prevent operation but indicate potential issues

**Use Cases**:
- Fallback to defaults (missing config)
- Performance degradation
- Deprecated feature usage
- Resource usage approaching limits

**Production**: Enabled

**Example**:
```go
logger.Warn("missing configuration, using default",
    "config_key", "log_level",
    "default_value", "INFO",
)
```

### ERROR

**When to Use**: Error conditions that prevent operation or require attention

**Use Cases**:
- Tool execution failures
- Query failures
- Parsing errors
- I/O errors
- Unexpected conditions

**Production**: Enabled (always log errors)

**Example**:
```go
logger.Error("tool execution failed",
    "request_id", requestID,
    "tool_name", toolName,
    "error", err.Error(),
    "error_type", "execution_error",
)
```

---

## Structured Field Naming

### Naming Conventions

**Format**: `snake_case` (lowercase with underscores)

**Good Examples**:
- `request_id` ✅
- `tool_name` ✅
- `duration_ms` ✅
- `error_type` ✅

**Bad Examples**:
- `requestID` ❌ (camelCase)
- `RequestID` ❌ (PascalCase)
- `request-id` ❌ (kebab-case)
- `reqId` ❌ (abbreviated)

### Standard Fields

| Field Name | Type | Purpose | When to Use |
|------------|------|---------|-------------|
| `request_id` | string | Unique request identifier | All tool invocations |
| `tool_name` | string | MCP tool being executed | Tool execution paths |
| `duration_ms` | int64 | Operation duration (ms) | Operation completion |
| `status` | string | Operation outcome | Operation completion |
| `error` | string | Error message | ERROR level logs |
| `error_type` | string | Error classification | ERROR level logs |
| `query_type` | string | Type of query | Query execution |
| `record_count` | int | Records processed/returned | Query completion |
| `scope` | string | Query scope | Tool execution |
| `file_path` | string | File being processed | File operations |

### Custom Field Guidelines

- Use descriptive names (not abbreviations)
- Be consistent across codebase
- Document new fields in this file
- Group related fields (prefix with common name)

**Example**:
```go
// Query-related fields: query_type, query_duration_ms, query_record_count
logger.Info("query executed",
    "query_type", "tools",
    "query_duration_ms", elapsed.Milliseconds(),
    "query_record_count", recordCount,
)
```

---

## Log Statement Patterns

### Tool Invocation Pattern

**Start**:
```go
requestID := uuid.New().String()
logger := slog.Default().With("request_id", requestID, "tool_name", toolName)

logger.Info("tool execution started",
    "scope", scope,
)
```

**Success**:
```go
logger.Info("tool execution completed",
    "status", "success",
    "duration_ms", elapsed.Milliseconds(),
    "record_count", recordCount,
)
```

**Failure**:
```go
logger.Error("tool execution failed",
    "error", err.Error(),
    "error_type", classifyError(err),
    "duration_ms", elapsed.Milliseconds(),
)
```

### Error Handling Pattern

**Basic Error**:
```go
if err != nil {
    logger.Error("operation failed",
        "operation", "parse_session_file",
        "error", err.Error(),
        "error_type", "parse_error",
    )
    return err
}
```

**Error with Context**:
```go
if err != nil {
    logger.Error("query execution failed",
        "query_type", queryType,
        "file_path", filePath,
        "error", err.Error(),
        "error_type", "execution_error",
    )
    return fmt.Errorf("query failed: %w", err)
}
```

### Query Execution Pattern

**Start**:
```go
logger.Debug("executing query",
    "query_type", queryType,
    "filter_applied", filterApplied,
)
```

**Success**:
```go
logger.Info("query executed successfully",
    "query_type", queryType,
    "record_count", len(results),
    "duration_ms", elapsed.Milliseconds(),
)
```

### DEBUG Logging (Conditional)

**When DEBUG is Expensive**:
```go
if logger.Enabled(context.Background(), slog.LevelDebug) {
    logger.Debug("detailed state",
        "complex_object", fmt.Sprintf("%+v", obj),
        "stack_trace", string(debug.Stack()),
    )
}
```

---

## Context Propagation

### Pattern: Logger in Context

**Setup**:
```go
// Create request-scoped logger
requestLogger := slog.Default().With(
    "request_id", requestID,
    "tool_name", toolName,
)

// Attach to context
ctx = context.WithValue(ctx, loggerKey, requestLogger)
```

**Usage in Functions**:
```go
func executeQuery(ctx context.Context, queryType string) error {
    logger := ctx.Value(loggerKey).(*slog.Logger)

    logger.Info("executing query",
        "query_type", queryType,
    )

    // ... execution ...

    return nil
}
```

**Helper Function**:
```go
func loggerFromContext(ctx context.Context) *slog.Logger {
    if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return logger
    }
    return slog.Default()
}
```

---

## Performance Guidelines

### Target: < 5% Overhead

**Techniques**:

1. **Use INFO Level in Production**:
   - Set `LOG_LEVEL=INFO` (disable DEBUG verbosity)
   - DEBUG logs have zero cost when disabled (lazy evaluation)

2. **Avoid Logging in Tight Loops**:
   ```go
   // ❌ BAD: Log per record
   for _, record := range records {
       logger.Debug("processing record", "id", record.ID)
       process(record)
   }

   // ✅ GOOD: Log aggregated result
   recordCount := 0
   for _, record := range records {
       process(record)
       recordCount++
   }
   logger.Info("processed records", "count", recordCount)
   ```

3. **Lazy Evaluation for DEBUG**:
   ```go
   // ✅ GOOD: Expensive DEBUG only if enabled
   if logger.Enabled(ctx, slog.LevelDebug) {
       logger.Debug("expensive debug",
           "stack_trace", string(debug.Stack()),
       )
   }
   ```

4. **Structured Fields over String Formatting**:
   ```go
   // ❌ BAD: String formatting
   logger.Info(fmt.Sprintf("tool %s completed in %dms", toolName, duration))

   // ✅ GOOD: Structured fields
   logger.Info("tool completed",
       "tool_name", toolName,
       "duration_ms", duration,
   )
   ```

### Measurement

**Benchmark**:
```bash
go test -bench=. -benchmem
```

**Production Monitoring**:
- Monitor CPU usage increase after logging deployment
- Target: < 5% CPU increase
- If exceeded: Review log density, reduce DEBUG logging

---

## Common Pitfalls

### ❌ Pitfall 1: Unstructured Logging

**Bad**:
```go
logger.Info(fmt.Sprintf("Tool %s completed in %dms with %d records", toolName, duration, count))
```

**Good**:
```go
logger.Info("tool completed",
    "tool_name", toolName,
    "duration_ms", duration,
    "record_count", count,
)
```

### ❌ Pitfall 2: Missing Context

**Bad**:
```go
logger.Error("query failed", "error", err.Error())
```

**Good**:
```go
logger.Error("query failed",
    "request_id", requestID,
    "query_type", queryType,
    "error", err.Error(),
    "error_type", "execution_error",
)
```

### ❌ Pitfall 3: Logging Sensitive Data

**Bad**:
```go
logger.Debug("user credentials", "password", password)
```

**Good**:
```go
logger.Debug("user authenticated", "user_id", userID)
```

### ❌ Pitfall 4: Inconsistent Field Names

**Bad**:
```go
logger.Info("start", "requestID", id)
logger.Info("end", "request_id", id)
```

**Good**:
```go
logger.Info("start", "request_id", id)
logger.Info("end", "request_id", id)
```

---

## Example: Complete Instrumentation

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "time"

    "github.com/google/uuid"
)

func main() {
    // Initialize logger
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
        AddSource: true,
    }))
    slog.SetDefault(logger)

    // Execute tool with logging
    requestID := uuid.New().String()
    executeTool(context.Background(), requestID, "query_tools")
}

func executeTool(ctx context.Context, requestID, toolName string) error {
    // Create request-scoped logger
    logger := slog.Default().With("request_id", requestID, "tool_name", toolName)
    ctx = context.WithValue(ctx, "logger", logger)

    start := time.Now()

    // Log start
    logger.Info("tool execution started", "scope", "project")

    // Execute query
    results, err := executeQuery(ctx, "tools")
    elapsed := time.Since(start)

    if err != nil {
        logger.Error("tool execution failed",
            "error", err.Error(),
            "error_type", "execution_error",
            "duration_ms", elapsed.Milliseconds(),
        )
        return err
    }

    // Log success
    logger.Info("tool execution completed",
        "status", "success",
        "duration_ms", elapsed.Milliseconds(),
        "record_count", len(results),
    )

    return nil
}

func executeQuery(ctx context.Context, queryType string) ([]interface{}, error) {
    logger := ctx.Value("logger").(*slog.Logger)

    logger.Debug("executing query", "query_type", queryType)

    // Query execution...
    results := []interface{}{} // placeholder

    logger.Info("query executed", "query_type", queryType, "record_count", len(results))

    return results, nil
}
```

---

**Version**: 1.0
**Last Updated**: 2025-10-17
**Maintained By**: observability-methodology experiment (Iteration 1)
