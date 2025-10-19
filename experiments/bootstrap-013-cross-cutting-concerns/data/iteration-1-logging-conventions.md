# Logging Conventions for meta-cc

**Version**: 1.0
**Created**: 2025-10-17 (Iteration 1)
**Status**: PROPOSED
**Agent**: convention-definer

---

## Executive Summary

This document defines comprehensive logging conventions for the meta-cc codebase. After analyzing the current state (0.7% logging coverage) and researching Go logging best practices, we recommend **`log/slog`** (Go 1.21+ standard library) for structured logging.

**Key Decision**: Use `log/slog` for all production logging
**Rationale**: Standard library, structured, performant, zero dependencies, future-proof

---

## Standard Logging Approach

### Chosen Technology: `log/slog`

**Comparison Analysis**:

| Library | Pros | Cons | Verdict |
|---------|------|------|---------|
| **log/slog** | Standard library, structured, good performance, zero deps | Requires Go 1.21+ | ✅ **RECOMMENDED** |
| zerolog | Fastest, zero-allocation | Third-party dependency, more complex | Alternative (if max perf needed) |
| zap | Very fast, mature | Third-party dependency, complex API | Alternative (for advanced use) |
| log (stdlib) | Simple, standard | Unstructured, basic | ❌ Insufficient |

**Decision**: **log/slog**
- ✅ Standard library (no external dependencies)
- ✅ Structured logging with key-value pairs
- ✅ Multiple output formats (JSON for production, text for development)
- ✅ Configurable log levels
- ✅ Good performance (optimized in Go 1.21+)
- ✅ Maintained by Go team (future-proof)

---

## Logger Initialization Pattern

### Package-Level Logger

**Standard Pattern**: Initialize logger at package level

```go
package parser

import (
    "log/slog"
    "os"
)

var log *slog.Logger

func init() {
    // Determine log level from environment
    level := slog.LevelInfo
    if envLevel := os.Getenv("META_CC_LOG_LEVEL"); envLevel != "" {
        switch envLevel {
        case "DEBUG":
            level = slog.LevelDebug
        case "INFO":
            level = slog.LevelInfo
        case "WARN":
            level = slog.LevelWarn
        case "ERROR":
            level = slog.LevelError
        }
    }

    // Determine output format from environment
    handler := createHandler(level)
    log = slog.New(handler)
}

func createHandler(level slog.Level) slog.Handler {
    opts := &slog.HandlerOptions{
        Level: level,
    }

    // JSON for production, text for development
    if os.Getenv("META_CC_LOG_FORMAT") == "json" {
        return slog.NewJSONHandler(os.Stderr, opts)
    }
    return slog.NewTextHandler(os.Stderr, opts)
}
```

**Rationale**:
- Package-level logger avoids passing logger through all function calls
- Configured once at initialization
- Environment variables allow runtime configuration
- Logs to stderr (standard practice for logs, stdout for data)

---

## Log Level Guidelines

### When to Use Each Level

#### 1. DEBUG

**Purpose**: Detailed diagnostic information for development and troubleshooting

**When to use**:
- Internal function calls and state transitions
- Detailed parsing/processing steps
- Variable values during debugging
- Performance timing details

**Example**:
```go
log.Debug("parsing JSONL line",
    "line_number", lineNum,
    "content_length", len(line),
    "timestamp", timestamp)
```

**Anti-pattern**: Don't use DEBUG for production-critical information

#### 2. INFO

**Purpose**: Informational messages about normal operation

**When to use**:
- Operation start/complete (query execution, file processing)
- Milestones and progress indicators
- Configuration loaded successfully
- Important state changes

**Example**:
```go
log.Info("query completed successfully",
    "query_type", "tools",
    "results_count", len(results),
    "duration_ms", duration.Milliseconds(),
    "session_id", sessionID)
```

**Anti-pattern**: Don't log every function call at INFO level

#### 3. WARN

**Purpose**: Warning about potentially problematic situations that don't prevent operation

**When to use**:
- Configuration values missing (using defaults)
- Deprecated usage detected
- Performance degradation detected
- Recoverable errors or retries

**Example**:
```go
log.Warn("environment variable not set, using default",
    "key", "META_CC_CAPABILITY_SOURCES",
    "default", defaultSources,
    "suggestion", "Set META_CC_CAPABILITY_SOURCES for custom capabilities")
```

**Anti-pattern**: Don't use WARN for fatal errors

#### 4. ERROR

**Purpose**: Error conditions that prevent specific operations but don't crash the program

**When to use**:
- Operation failures (file not found, parse error)
- External service failures (API calls)
- Data validation failures
- Any error returned to the user

**Example**:
```go
log.Error("failed to parse session file",
    "file", filepath,
    "line_number", lineNum,
    "error", err.Error(),
    "suggestion", "Check file format and permissions")
```

**Anti-pattern**: Don't log and swallow errors; log at point of handling

---

## Structured Logging Format

### Standard: Key-Value Pairs

**Pattern**: Always use key-value pairs for context

```go
// ✓ GOOD: Structured logging
log.Info("processing batch",
    "batch_id", batchID,
    "files_count", len(files),
    "start_time", startTime,
    "user_id", userID)

// ✗ BAD: String interpolation
log.Info(fmt.Sprintf("Processing batch %s with %d files", batchID, len(files)))
```

**Rationale**:
- Structured logs are machine-parseable
- Easy to filter and search
- Consistent format across codebase
- Better for log aggregation tools

### Context Fields Naming Convention

**Standard field names**:
- `file`: File path
- `line_number`: Line number in file
- `error`: Error message (use `err.Error()`)
- `duration_ms`: Duration in milliseconds
- `count`: Generic count
- `{entity}_id`: ID of entity (e.g., `session_id`, `user_id`)
- `{entity}_count`: Count of entities (e.g., `files_count`, `results_count`)

**Naming rules**:
- Use snake_case for field names
- Be descriptive but concise
- Use consistent names across codebase
- Avoid abbreviations unless very common

---

## Context Propagation

### Adding Operation Context

**Pattern**: Add operation-specific context at function entry

```go
func ProcessSession(ctx context.Context, sessionID string, filepath string) error {
    // Add operation context
    logger := log.With(
        "operation", "process_session",
        "session_id", sessionID,
        "file", filepath,
    )

    logger.Info("starting session processing")

    // Use logger throughout function
    if err := parseFile(filepath); err != nil {
        logger.Error("failed to parse file", "error", err)
        return err
    }

    logger.Info("session processing completed",
        "lines_processed", lineCount,
        "duration_ms", duration.Milliseconds())
    return nil
}
```

**Rationale**:
- Context automatically included in all logs
- Reduces repetition
- Easy to trace related log messages

---

## Configuration

### Environment Variables

**Standard configuration**:

```bash
# Log level: DEBUG, INFO, WARN, ERROR (default: INFO)
export META_CC_LOG_LEVEL=INFO

# Log format: text, json (default: text)
export META_CC_LOG_FORMAT=text

# Enable/disable logging: true, false (default: true)
export META_CC_LOGGING_ENABLED=true
```

**Implementation**:
```go
func init() {
    // Check if logging disabled
    if os.Getenv("META_CC_LOGGING_ENABLED") == "false" {
        log = slog.New(slog.NewTextHandler(io.Discard, nil))
        return
    }

    level := parseLogLevel(os.Getenv("META_CC_LOG_LEVEL"))
    handler := createHandler(level, os.Getenv("META_CC_LOG_FORMAT"))
    log = slog.New(handler)
}
```

---

## Logging Insertion Points

### Where to Add Logging

**Critical insertion points in meta-cc**:

1. **Parser (internal/parser/)**
   - INFO: Start/complete parsing session file
   - DEBUG: Each JSONL line parsed
   - ERROR: Parse errors, invalid JSON
   - Example: "Parsing session file" → "Parsed 1,234 records in 45ms"

2. **Analyzer (internal/analyzer/)**
   - INFO: Start/complete analysis operation
   - DEBUG: Pattern detection steps
   - ERROR: Analysis failures
   - Example: "Analyzing error patterns" → "Found 15 patterns in 500 errors"

3. **Query Engine (internal/query/)**
   - INFO: Query execution start/complete
   - DEBUG: Filter application, data processing
   - WARN: No results found
   - ERROR: Query failures
   - Example: "Executing query" → "Query returned 42 results in 12ms"

4. **MCP Server (cmd/mcp-server/)**
   - INFO: Server start/stop, request handling
   - DEBUG: Request/response details
   - WARN: Client errors, rate limiting
   - ERROR: Server errors
   - Example: "MCP server started on :8080" → "Handled request in 23ms"

5. **CLI Commands (cmd/)**
   - INFO: Command execution start/complete
   - DEBUG: Flag values, intermediate steps
   - ERROR: Command failures
   - Example: "Executing query-tools command" → "Command completed successfully"

---

## Anti-Patterns

### 1. Using fmt.Printf for Logging

❌ **Bad**:
```go
fmt.Printf("DEBUG: processing file %s\n", filename)
fmt.Fprintf(os.Stderr, "Error: failed to parse: %v\n", err)
```

✅ **Good**:
```go
log.Debug("processing file", "filename", filename)
log.Error("failed to parse", "error", err)
```

**Why**: `fmt.Printf` cannot be:
- Filtered by level
- Captured for analysis
- Structured for machine parsing
- Configured centrally

**Migration**: Replace all `fmt.Printf` in internal/ with appropriate `log.*()` calls

---

### 2. Missing Context in Logs

❌ **Bad**:
```go
if err != nil {
    log.Error("operation failed", "error", err)
}
```

✅ **Good**:
```go
if err != nil {
    log.Error("failed to parse JSONL file",
        "file", filepath,
        "line_number", lineNum,
        "operation", "parse_session",
        "error", err)
}
```

**Why**: Insufficient context makes debugging nearly impossible

---

### 3. Logging Sensitive Data

❌ **Bad**:
```go
log.Info("processing user request",
    "api_key", apiKey,
    "password", password)
```

✅ **Good**:
```go
log.Info("processing user request",
    "user_id", userID,
    "request_id", requestID)
```

**Why**: Logs may be stored, transmitted, or accessed by unauthorized parties

---

### 4. Incorrect Log Levels

❌ **Bad**:
```go
// Using ERROR for non-errors
log.Error("processing complete") // Should be INFO

// Using INFO for debugging details
log.Info("variable x =", x, "variable y =", y) // Should be DEBUG

// Using DEBUG for critical errors
log.Debug("fatal database error", "error", err) // Should be ERROR
```

✅ **Good**:
```go
log.Info("processing complete")
log.Debug("variable values", "x", x, "y", y)
log.Error("database connection failed", "error", err)
```

**Why**: Incorrect levels make it hard to filter and prioritize logs

---

### 5. Over-Logging

❌ **Bad**:
```go
for _, item := range items {
    log.Debug("processing item", "item", item) // 10,000 log lines!
}
```

✅ **Good**:
```go
log.Info("processing items", "count", len(items))
// ... process items ...
log.Info("items processed",
    "count", len(items),
    "duration_ms", duration.Milliseconds())
```

**Why**: Excessive logging degrades performance and obscures important information

---

### 6. Under-Logging

❌ **Bad**:
```go
func ParseSessionFile(filepath string) (*Session, error) {
    // ... 200 lines of parsing logic ...
    // No logging at all!
    return session, nil
}
```

✅ **Good**:
```go
func ParseSessionFile(filepath string) (*Session, error) {
    log.Info("parsing session file", "file", filepath)

    // ... parsing logic with strategic DEBUG logs ...

    if err != nil {
        log.Error("failed to parse session",
            "file", filepath,
            "error", err)
        return nil, err
    }

    log.Info("session parsed successfully",
        "file", filepath,
        "records", len(session.Records),
        "duration_ms", duration.Milliseconds())
    return session, nil
}
```

**Why**: No visibility into operation status, failures, or performance

---

## Performance Considerations

### When Performance Matters

**Logging overhead**:
- `slog` is highly optimized (minimal allocations)
- DEBUG logs are skipped entirely if level > DEBUG
- Structured logging has minimal overhead vs unstructured

**Best practices**:
1. **Use appropriate levels**: DEBUG logs are free when level = INFO
2. **Avoid expensive operations in log calls**:
   ```go
   // ✗ Bad: computeExpensive() called even if DEBUG disabled
   log.Debug("expensive result", "result", computeExpensive())

   // ✓ Good: Check level before expensive operation
   if log.Enabled(context.Background(), slog.LevelDebug) {
       log.Debug("expensive result", "result", computeExpensive())
   }
   ```
3. **Don't log in tight loops**: Batch and summarize instead

---

## Migration Guide

### Step 1: Add log/slog Import

```go
import (
    "log/slog"
    "os"
)
```

### Step 2: Initialize Package Logger

```go
var log *slog.Logger

func init() {
    level := slog.LevelInfo
    if envLevel := os.Getenv("META_CC_LOG_LEVEL"); envLevel != "" {
        // Parse level...
    }

    handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: level,
    })
    log = slog.New(handler)
}
```

### Step 3: Replace fmt.Printf with log.*()

```go
// Before
fmt.Printf("Processing file %s\n", filename)

// After
log.Info("processing file", "filename", filename)
```

### Step 4: Add Logging to Critical Points

Priority order:
1. **High Priority**: Error conditions, operation start/complete
2. **Medium Priority**: Important state changes, warnings
3. **Low Priority**: Debugging details (add as needed)

---

## Testing Logging

### Verify Logging Works

```go
func TestLogging(t *testing.T) {
    // Capture log output
    var buf bytes.Buffer
    handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    })
    testLog := slog.New(handler)

    // Log something
    testLog.Info("test message", "key", "value")

    // Verify output
    output := buf.String()
    if !strings.Contains(output, "test message") {
        t.Error("expected log message not found")
    }
    if !strings.Contains(output, "\"key\":\"value\"") {
        t.Error("expected structured field not found")
    }
}
```

---

## Examples

### Example 1: Parser Logging

```go
package parser

import (
    "context"
    "log/slog"
    "os"
    "time"
)

var log *slog.Logger

func init() {
    // Initialize logger (see initialization pattern above)
    level := slog.LevelInfo
    handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
    log = slog.New(handler)
}

func ParseSessionFile(filepath string) (*Session, error) {
    start := time.Now()
    log.Info("parsing session file", "file", filepath)

    data, err := os.ReadFile(filepath)
    if err != nil {
        log.Error("failed to read file",
            "file", filepath,
            "error", err)
        return nil, fmt.Errorf("read file: %w", err)
    }

    log.Debug("file read successfully",
        "file", filepath,
        "bytes", len(data))

    session := &Session{}
    lines := bytes.Split(data, []byte("\n"))

    for i, line := range lines {
        if len(line) == 0 {
            continue
        }

        log.Debug("parsing line", "line_number", i+1, "length", len(line))

        var record Record
        if err := json.Unmarshal(line, &record); err != nil {
            log.Error("failed to parse JSONL line",
                "file", filepath,
                "line_number", i+1,
                "error", err)
            return nil, fmt.Errorf("parse line %d: %w", i+1, err)
        }

        session.Records = append(session.Records, record)
    }

    duration := time.Since(start)
    log.Info("session parsed successfully",
        "file", filepath,
        "records", len(session.Records),
        "duration_ms", duration.Milliseconds())

    return session, nil
}
```

### Example 2: Query Logging

```go
package query

import (
    "context"
    "log/slog"
    "time"
)

var log *slog.Logger

func ExecuteQuery(ctx context.Context, query *Query) (*Results, error) {
    // Add operation context
    logger := log.With(
        "operation", "execute_query",
        "query_type", query.Type,
        "session_id", query.SessionID,
    )

    start := time.Now()
    logger.Info("executing query")

    results, err := processQuery(query)
    if err != nil {
        logger.Error("query execution failed", "error", err)
        return nil, err
    }

    if len(results.Items) == 0 {
        logger.Warn("query returned no results",
            "filters", query.Filters,
            "suggestion", "try broader filter criteria")
    }

    duration := time.Since(start)
    logger.Info("query completed",
        "results_count", len(results.Items),
        "duration_ms", duration.Milliseconds(),
        "filters_applied", len(query.Filters))

    return results, nil
}
```

---

## Summary

**Adopted Standard**: `log/slog` with structured logging

**Key Principles**:
1. ✅ Use structured logging (key-value pairs)
2. ✅ Choose appropriate log levels
3. ✅ Include sufficient context
4. ✅ Configure via environment variables
5. ✅ Avoid sensitive data in logs
6. ✅ Log to stderr (data to stdout)

**Next Steps**:
1. Implement logger initialization in each package
2. Add logging to critical operations (parser, query, analyzer, MCP server)
3. Replace fmt.Printf with log.*() calls
4. Create linters to enforce conventions (Iteration 3-4)
5. Measure logging coverage improvement

---

**Document Version**: 1.0
**Status**: PROPOSED (to be validated in implementation)
**Created By**: convention-definer agent
**Iteration**: 1
