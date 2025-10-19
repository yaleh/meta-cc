# Go Logging Best Practices

**Domain**: Cross-Cutting Concerns
**Language**: Go
**Topic**: Logging
**Status**: VALIDATED (Iteration 1)

---

## Overview

This document captures best practices for logging in Go applications, derived from industry standards and the Go community.

---

## 1. Use Structured Logging

**Practice**: Always use structured logging with key-value pairs

**Rationale**:
- Machine-parseable
- Easy to filter and search
- Consistent format
- Better for log aggregation tools (ELK, Splunk, etc.)

**Example**:
```go
// ✓ Good: Structured
log.Info("user login",
    "user_id", userID,
    "ip_address", ip,
    "timestamp", time.Now())

// ✗ Bad: Unstructured
log.Info(fmt.Sprintf("User %s logged in from %s", userID, ip))
```

---

## 2. Choose Log Levels Carefully

**Practice**: Use log levels to indicate severity and purpose

**Levels**:
- **DEBUG**: Development debugging, detailed diagnostic info
- **INFO**: Normal operation milestones, important events
- **WARN**: Potentially problematic situations, degraded performance
- **ERROR**: Operation failures, errors returned to user

**Example**:
```go
// DEBUG: Internal details
log.Debug("cache hit", "key", cacheKey, "ttl", ttl)

// INFO: Important milestones
log.Info("request completed", "endpoint", "/api/users", "duration_ms", duration)

// WARN: Recoverable issues
log.Warn("retry attempt", "attempt", 2, "max_attempts", 3)

// ERROR: Failures
log.Error("database query failed", "query", sql, "error", err)
```

---

## 3. Include Sufficient Context

**Practice**: Always provide enough context to understand what happened

**Essential context**:
- What operation was being performed
- What entity was involved (file, user, request)
- Relevant identifiers (IDs, paths, names)
- Error details (if applicable)

**Example**:
```go
// ✗ Bad: Insufficient context
log.Error("operation failed", "error", err)

// ✓ Good: Sufficient context
log.Error("failed to process user payment",
    "user_id", userID,
    "order_id", orderID,
    "amount", amount,
    "payment_method", method,
    "error", err)
```

---

## 4. Use Package-Level Loggers

**Practice**: Initialize logger at package level, not global

**Rationale**:
- Avoids global state pollution
- Each package can have its own logger configuration
- Easier to test (can override per-package)

**Example**:
```go
package parser

import "log/slog"

var log *slog.Logger

func init() {
    log = createLogger() // Package-specific initialization
}
```

---

## 5. Log to stderr (Not stdout)

**Practice**: Always log to stderr, reserve stdout for data output

**Rationale**:
- Follows Unix conventions (data on stdout, logs/errors on stderr)
- Allows separating logs from data in pipelines
- Standard practice in CLI tools

**Example**:
```go
// ✓ Good: Logs to stderr
handler := slog.NewTextHandler(os.Stderr, opts)

// ✗ Bad: Logs to stdout (conflicts with data output)
handler := slog.NewTextHandler(os.Stdout, opts)
```

---

## 6. Avoid Logging Sensitive Data

**Practice**: Never log passwords, API keys, tokens, or PII

**Sensitive data**:
- Passwords, API keys, tokens
- Personally Identifiable Information (PII): email, phone, SSN
- Financial data: credit card numbers, account balances
- Health information

**Example**:
```go
// ✗ Bad: Logging sensitive data
log.Info("authentication attempt",
    "username", username,
    "password", password) // NEVER!

// ✓ Good: Log non-sensitive identifiers
log.Info("authentication attempt",
    "user_id", userID,
    "ip_address", ip)
```

---

## 7. Use Appropriate Sampling for High-Volume Logs

**Practice**: Sample or batch high-volume logs to avoid performance degradation

**Example**:
```go
// ✗ Bad: Log every item in loop (10,000 log lines!)
for _, item := range items {
    log.Debug("processing item", "item", item)
}

// ✓ Good: Log summary
log.Info("processing items batch",
    "count", len(items),
    "batch_id", batchID)

// ... process items ...

log.Info("batch processed",
    "count", len(items),
    "duration_ms", duration.Milliseconds(),
    "errors", errorCount)
```

---

## 8. Make Logs Actionable

**Practice**: Logs should help diagnose and fix problems

**Include**:
- What went wrong
- Why it went wrong (if known)
- What to do about it (suggestions)

**Example**:
```go
// ✗ Bad: Unhelpful
log.Error("error", "error", err)

// ✓ Good: Actionable
log.Error("failed to connect to database",
    "host", dbHost,
    "port", dbPort,
    "error", err,
    "suggestion", "Check database credentials and network connectivity")
```

---

## 9. Use Context for Request Tracing

**Practice**: Propagate context through call chains for distributed tracing

**Example**:
```go
func ProcessRequest(ctx context.Context, req *Request) error {
    // Add request context
    logger := log.With(
        "request_id", req.ID,
        "user_id", req.UserID,
        "operation", "process_request",
    )

    logger.Info("request received")

    // Use logger throughout function
    if err := validate(req); err != nil {
        logger.Error("validation failed", "error", err)
        return err
    }

    // Pass context to subfunctions
    result, err := processWithContext(ctx, req, logger)
    if err != nil {
        logger.Error("processing failed", "error", err)
        return err
    }

    logger.Info("request completed", "duration_ms", duration)
    return nil
}
```

---

## 10. Configure via Environment Variables

**Practice**: Make logging configurable without code changes

**Standard environment variables**:
- `LOG_LEVEL`: DEBUG, INFO, WARN, ERROR
- `LOG_FORMAT`: text, json
- `LOGGING_ENABLED`: true, false

**Example**:
```go
func init() {
    level := parseLogLevel(os.Getenv("LOG_LEVEL"))
    format := os.Getenv("LOG_FORMAT")
    log = createLogger(level, format)
}
```

---

## 11. Test Logging

**Practice**: Verify logging works in tests

**Example**:
```go
func TestLogging(t *testing.T) {
    var buf bytes.Buffer
    handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    })
    testLog := slog.New(handler)

    testLog.Info("test message", "key", "value")

    output := buf.String()
    if !strings.Contains(output, "test message") {
        t.Error("expected log message not found")
    }
}
```

---

## 12. Don't Log-and-Throw

**Practice**: Log at the point of error handling, not at error creation

**Example**:
```go
// ✗ Bad: Log-and-throw (logs error twice)
func helper() error {
    if err := operation(); err != nil {
        log.Error("operation failed", "error", err) // Logged here
        return err // And will be logged by caller
    }
    return nil
}

// ✓ Good: Return error, let caller decide to log
func helper() error {
    if err := operation(); err != nil {
        return fmt.Errorf("operation failed: %w", err) // Return wrapped error
    }
    return nil
}

func caller() {
    if err := helper(); err != nil {
        log.Error("helper failed", "error", err) // Log once, at handling point
    }
}
```

---

## 13. Use go/slog for New Projects

**Practice**: Prefer `log/slog` (Go 1.21+) for new projects

**Rationale**:
- Standard library (no dependencies)
- Structured logging built-in
- Good performance
- Future-proof (maintained by Go team)

**When to use alternatives**:
- **zerolog**: Maximum performance required
- **zap**: Advanced features (log rotation, async logging)
- **logrus**: Legacy projects already using it

---

## Common Anti-Patterns

### 1. Using fmt.Printf for Logging

❌ Problem: Cannot filter by level, not structured, not configurable

```go
// Bad
fmt.Printf("Processing file %s\n", filename)

// Good
log.Info("processing file", "filename", filename)
```

### 2. Logging Inside Loops

❌ Problem: Performance degradation, log spam

```go
// Bad
for _, item := range items {
    log.Info("processing", "item", item) // Thousands of logs!
}

// Good
log.Info("processing items", "count", len(items))
```

### 3. Missing Error Context

❌ Problem: Cannot diagnose issues

```go
// Bad
log.Error("error", "error", err)

// Good
log.Error("failed to parse config file",
    "file", configPath,
    "line", lineNum,
    "error", err)
```

---

## References

- [log/slog package](https://pkg.go.dev/log/slog)
- [Structured Logging in Go](https://go.dev/blog/slog)
- [Uber Go Style Guide - Logging](https://github.com/uber-go/guide/blob/master/style.md#logging)
- [12-Factor App - Logs](https://12factor.net/logs)

---

**Status**: VALIDATED
**Source**: Iteration 1 (Bootstrap-013)
**Validation**: Derived from Go community standards and meta-cc requirements
