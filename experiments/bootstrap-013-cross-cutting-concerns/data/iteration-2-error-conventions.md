# Error Handling Conventions for meta-cc

**Version**: 1.0
**Status**: PROPOSED
**Created**: 2025-10-17 (Iteration 2)
**Applies To**: All meta-cc Go code

---

## 1. Standard Library Approach

**Decision**: Use Go 1.13+ `errors` package and `fmt.Errorf` with `%w` for error wrapping

**Rationale**:
- ✅ Standard library (zero dependencies)
- ✅ Error wrapping and unwrapping built-in
- ✅ errors.Is/As for error type checking
- ✅ Future-proof (maintained by Go team)
- ✅ Idiomatic Go approach

**Alternatives Considered**:
- **pkg/errors**: Deprecated (functionality moved to stdlib in Go 1.13+)
- **Custom error library**: Unnecessary complexity

---

## 2. Error Wrapping Pattern

### 2.1 Basic Wrapping

**Convention**: Always wrap errors with context using `fmt.Errorf` with `%w` verb

**Pattern**:
```go
// ✓ Good: Wraps error with context
if err := operation(); err != nil {
    return fmt.Errorf("failed to process user data: %w", err)
}

// ✗ Bad: Returns raw error (loses context)
if err := operation(); err != nil {
    return err
}

// ✗ Bad: Uses %v instead of %w (breaks unwrapping)
if err := operation(); err != nil {
    return fmt.Errorf("operation failed: %v", err)
}
```

### 2.2 Context Requirements

**Convention**: Error messages must include:
1. **What operation failed** ("failed to process user data")
2. **Relevant identifiers** (user_id, file_path, session_id)
3. **Original error** (via %w)

**Examples**:
```go
// ✓ Good: Complete context
return fmt.Errorf("failed to parse session file %s at line %d: %w",
    filePath, lineNum, err)

// ✓ Good: Entity identifiers
return fmt.Errorf("failed to load user %s from database: %w",
    userID, err)

// ✗ Bad: Insufficient context
return fmt.Errorf("operation failed: %w", err)

// ✗ Bad: Too verbose (includes irrelevant details)
return fmt.Errorf("failed to open file %s with mode %s at %s using function %s: %w",
    path, mode, timestamp, funcName, err)
```

### 2.3 Multi-Level Wrapping

**Convention**: Each layer adds its own context, building a stack of information

**Pattern**:
```go
// Layer 1: Low-level operation
func readFile(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read file %s: %w", path, err)
    }
    return data, nil
}

// Layer 2: Business logic
func loadConfig(path string) (*Config, error) {
    data, err := readFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    // ... parse config ...
}

// Layer 3: CLI command
func runCommand() error {
    cfg, err := loadConfig("/etc/meta-cc/config.yaml")
    if err != nil {
        return fmt.Errorf("command initialization failed: %w", err)
    }
    // ... use config ...
}

// Result: Rich error context
// "command initialization failed: failed to load config: failed to read file /etc/meta-cc/config.yaml: open /etc/meta-cc/config.yaml: no such file or directory"
```

---

## 3. Sentinel Errors

### 3.1 Definition

**Convention**: Define package-level sentinel errors for expected error conditions

**Pattern**:
```go
package query

import "errors"

// Sentinel errors for common conditions
var (
    ErrNoResults       = errors.New("no results found")
    ErrInvalidQuery    = errors.New("invalid query syntax")
    ErrSessionNotFound = errors.New("session file not found")
)

// Usage
func ExecuteQuery(q string) ([]Result, error) {
    if !isValid(q) {
        return nil, ErrInvalidQuery
    }

    results := doQuery(q)
    if len(results) == 0 {
        return nil, ErrNoResults
    }

    return results, nil
}

// Checking
if err := ExecuteQuery(query); errors.Is(err, query.ErrNoResults) {
    // Handle no results case
}
```

### 3.2 When to Use Sentinel Errors

**Use For**:
- Expected error conditions that callers may handle differently
- Error conditions that need to be distinguished programmatically
- Common errors across multiple functions

**Don't Use For**:
- Errors with dynamic context (use wrapping instead)
- Errors unlikely to be checked by callers
- One-off error conditions

**Examples**:
```go
// ✓ Good: Expected condition, caller may handle specially
var ErrConfigNotFound = errors.New("config file not found")

// ✗ Bad: Dynamic context (use wrapping)
// var ErrParseLineN = errors.New("parse error at line N") // Wrong!
// Instead: return fmt.Errorf("parse error at line %d: %w", lineNum, err)

// ✗ Bad: One-off error
var ErrThisSpecificFileIsMissing = errors.New("missing specific.json") // Wrong!
// Instead: return fmt.Errorf("file not found: %s", path)
```

---

## 4. Custom Error Types

### 4.1 When to Create Custom Types

**Use Custom Error Types When**:
1. Errors need structured data (error codes, metadata)
2. Errors require special formatting or serialization
3. Errors need method-based behavior

**Pattern**:
```go
// Custom error type with structured data
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field %s: %s (got: %v)",
        e.Field, e.Message, e.Value)
}

// Usage
func ValidateUser(u *User) error {
    if u.Age < 0 {
        return &ValidationError{
            Field:   "age",
            Value:   u.Age,
            Message: "must be non-negative",
        }
    }
    return nil
}

// Checking with errors.As
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Printf("Invalid field: %s\n", valErr.Field)
}
```

### 4.2 Error Codes Pattern

**Use For**: CLI tools, APIs, structured error output

**Pattern** (Following `internal/output/error.go`):
```go
package output

type ErrorCode string

const (
    ErrInvalidArgument ErrorCode = "INVALID_ARGUMENT"
    ErrSessionNotFound ErrorCode = "SESSION_NOT_FOUND"
    ErrParseError      ErrorCode = "PARSE_ERROR"
    ErrFilterError     ErrorCode = "FILTER_ERROR"
    ErrNoResults       ErrorCode = "NO_RESULTS"
    ErrInternalError   ErrorCode = "INTERNAL_ERROR"
)

type ErrorOutput struct {
    Error   string    `json:"error"`
    Code    ErrorCode `json:"code"`
    Message string    `json:"message,omitempty"` // Suggestions
}

func OutputError(err error, code ErrorCode, format string) error {
    // ... output formatted error ...
    return NewExitCodeError(exitCode, err.Error())
}
```

---

## 5. Error Recovery and Retry

### 5.1 Panic vs Error Return

**Convention**:
- **NEVER panic in libraries or packages**
- **ONLY panic in main() for unrecoverable errors**
- **ALWAYS return errors** for expected failure modes

**Rationale**: Libraries cannot know how calling code wants to handle errors

**Examples**:
```go
// ✓ Good: Library code returns errors
func ParseSession(path string) ([]Entry, error) {
    if path == "" {
        return nil, errors.New("path cannot be empty")
    }
    // ... parse ...
    return entries, nil
}

// ✗ Bad: Library code panics
func ParseSession(path string) []Entry {
    if path == "" {
        panic("path cannot be empty") // NEVER!
    }
    // ... parse ...
}

// ✓ Good: main() may panic for unrecoverable errors
func main() {
    if os.Getenv("REQUIRED_VAR") == "" {
        panic("REQUIRED_VAR must be set") // OK in main
    }
}
```

### 5.2 Retry Pattern

**Convention**: Implement retry logic for transient errors only

**Pattern**:
```go
func retryOperation(maxAttempts int) error {
    var lastErr error

    for attempt := 0; attempt < maxAttempts; attempt++ {
        err := operation()
        if err == nil {
            return nil // Success
        }

        lastErr = err

        // Don't retry permanent errors
        if !isTransient(err) {
            return err
        }

        // Exponential backoff
        if attempt < maxAttempts-1 {
            delay := time.Duration(1<<attempt) * time.Second
            time.Sleep(delay)
        }
    }

    return fmt.Errorf("operation failed after %d attempts: %w",
        maxAttempts, lastErr)
}

func isTransient(err error) bool {
    // Network errors, 5xx server errors, timeouts
    return strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "connection refused") ||
           strings.Contains(err.Error(), "status 5")
}
```

**Transient Errors** (retry):
- Network timeouts
- Connection refused
- Server errors (5xx)
- Temporary resource unavailability

**Permanent Errors** (don't retry):
- Not found (404)
- Invalid input (400)
- Authentication failures
- Parse errors
- Logic errors

---

## 6. Error Logging Strategy

### 6.1 Log-and-Throw Anti-Pattern

**Convention**: Log errors at the point of handling, NOT at creation or wrapping

**Anti-Pattern**:
```go
// ✗ Bad: Log-and-throw (error logged twice)
func helper() error {
    if err := operation(); err != nil {
        log.Error("operation failed", "error", err) // Logged here
        return err // Will be logged again by caller!
    }
    return nil
}

func caller() error {
    if err := helper(); err != nil {
        log.Error("helper failed", "error", err) // Logged again!
        return err
    }
    return nil
}
```

**Correct Pattern**:
```go
// ✓ Good: Return errors, let top-level handler log
func helper() error {
    if err := operation(); err != nil {
        return fmt.Errorf("operation failed: %w", err) // Just wrap
    }
    return nil
}

func caller() error {
    if err := helper(); err != nil {
        return fmt.Errorf("helper failed: %w", err) // Just wrap
    }
    return nil
}

func topLevel() {
    if err := caller(); err != nil {
        log.Error("request failed", "error", err) // Log ONCE at top
    }
}
```

### 6.2 When to Log Errors

**Log At**:
- Top-level handlers (HTTP handlers, CLI command entry points)
- Goroutine boundaries (errors in concurrent code)
- Error recovery points (after retries, fallback logic)

**Don't Log At**:
- Library functions (just wrap and return)
- Helper functions (just wrap and return)
- Every error occurrence (creates duplicate logs)

**Log Level Guidelines**:
- **ERROR**: Operation failures that affect functionality
- **WARN**: Recoverable issues, degraded performance, retries
- **INFO**: Not for errors (use for success milestones)
- **DEBUG**: Detailed error context (stack traces)

---

## 7. Error Context Best Practices

### 7.1 Essential Context

**Always Include**:
1. **Operation**: What was being attempted
2. **Entity**: What was being operated on (file, user, session)
3. **Identifiers**: Relevant IDs, paths, names
4. **Original Error**: Via %w

**Example**:
```go
return fmt.Errorf("failed to parse session file %s at line %d: %w",
    filePath, lineNum, err)
//                  ↑              ↑         ↑        ↑
//              operation       entity  identifier  original
```

### 7.2 Context Guidelines

**Do**:
- Include file paths (for file operations)
- Include line numbers (for parsing)
- Include entity IDs (user_id, session_id)
- Include operation stage ("during validation", "while saving")

**Don't**:
- Include sensitive data (passwords, tokens, PII)
- Include excessive detail (function names, timestamps)
- Include redundant information (already in wrapped error)

**Examples**:
```go
// ✓ Good: Relevant context
return fmt.Errorf("failed to save user %s to database: %w", userID, err)

// ✗ Bad: Sensitive data
return fmt.Errorf("failed to authenticate user %s with password %s: %w",
    username, password, err) // NEVER!

// ✗ Bad: Excessive detail
return fmt.Errorf("failed in function %s at %s line %d timestamp %s: %w",
    funcName, fileName, lineNum, time.Now(), err) // Too much!

// ✗ Bad: Redundant info
return fmt.Errorf("file not found: %w", os.ErrNotExist) // Already says "file not found"
// Better: return fmt.Errorf("config file %s not found: %w", path, os.ErrNotExist)
```

---

## 8. Error Message Style

### 8.1 Message Format

**Convention**:
- Use lowercase for error messages (Go convention)
- Start with verb ("failed to", "cannot", "unable to")
- Be specific and actionable
- Don't end with punctuation (unless multiple sentences)

**Examples**:
```go
// ✓ Good
errors.New("failed to connect to database")
errors.New("session file not found")
fmt.Errorf("invalid query syntax: expected SELECT, got %s", token)

// ✗ Bad: Capitalized
errors.New("Failed to connect") // Wrong

// ✗ Bad: Vague
errors.New("error") // Too vague

// ✗ Bad: Unnecessary punctuation
errors.New("failed to connect.") // No trailing period
```

### 8.2 Actionable Messages

**Convention**: When possible, suggest remediation

**Pattern**:
```go
// Using custom error type for suggestions
type ErrorOutput struct {
    Error   string
    Message string // Suggestion
}

switch code {
case ErrSessionNotFound:
    errOutput.Message = "Try specifying --session or --project flags"
case ErrInvalidArgument:
    errOutput.Message = "Check command syntax with --help"
case ErrFilterError:
    errOutput.Message = "Verify filter syntax (e.g., tool=Bash status=error)"
}
```

---

## 9. Integration with Logging

### 9.1 Error Classification

**Pattern** (Following `cmd/mcp-server/logging.go`):
```go
func classifyError(err error) string {
    if err == nil {
        return ""
    }

    errMsg := err.Error()

    // Parse errors
    if strings.Contains(errMsg, "parse") || strings.Contains(errMsg, "unmarshal") {
        return "parse_error"
    }

    // Validation errors
    if strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "invalid") {
        return "validation_error"
    }

    // I/O errors
    if strings.Contains(errMsg, "no such file") || strings.Contains(errMsg, "permission denied") {
        return "io_error"
    }

    // Network errors
    if strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection") {
        return "network_error"
    }

    return "general_error"
}

// Usage with logging
log.Error("operation failed",
    "error", err.Error(),
    "error_type", classifyError(err))
```

### 9.2 Logging Wrapped Errors

**Convention**: Log the complete error chain (auto-handled by %w)

**Pattern**:
```go
// Error chain automatically includes all context
err := fmt.Errorf("command failed: %w",
    fmt.Errorf("config load failed: %w",
        fmt.Errorf("file not found: %s", path)))

// Logging outputs full chain
log.Error("execution failed", "error", err)
// Output: "command failed: config load failed: file not found: /etc/config.yaml"
```

---

## 10. Testing Error Handling

### 10.1 Error Testing Pattern

**Convention**: Test both happy path and error cases

**Pattern**:
```go
func TestParseSession(t *testing.T) {
    // Test success case
    entries, err := ParseSession("valid.jsonl")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(entries) != 10 {
        t.Errorf("expected 10 entries, got %d", len(entries))
    }

    // Test error case
    _, err = ParseSession("nonexistent.jsonl")
    if err == nil {
        t.Fatal("expected error for nonexistent file")
    }

    // Test error wrapping
    if !strings.Contains(err.Error(), "nonexistent.jsonl") {
        t.Errorf("error should contain file path: %v", err)
    }
}
```

### 10.2 Testing Sentinel Errors

**Pattern**:
```go
func TestSentinelErrors(t *testing.T) {
    err := ExecuteQuery("")

    // Use errors.Is for sentinel error checking
    if !errors.Is(err, ErrInvalidQuery) {
        t.Errorf("expected ErrInvalidQuery, got: %v", err)
    }
}
```

### 10.3 Testing Custom Error Types

**Pattern**:
```go
func TestCustomErrors(t *testing.T) {
    err := ValidateUser(&User{Age: -1})

    var valErr *ValidationError
    if !errors.As(err, &valErr) {
        t.Fatal("expected ValidationError")
    }

    if valErr.Field != "age" {
        t.Errorf("expected field 'age', got '%s'", valErr.Field)
    }
}
```

---

## Summary

### Must-Follow Patterns:

1. ✅ **Always wrap errors** with fmt.Errorf and %w
2. ✅ **Include context** (operation, entity, identifiers)
3. ✅ **Never panic in libraries** (only in main for unrecoverable errors)
4. ✅ **Log at top level** (avoid log-and-throw)
5. ✅ **Use lowercase messages** (Go convention)
6. ✅ **Define sentinel errors** for expected conditions
7. ✅ **Create custom types** for structured errors
8. ✅ **Test error paths** (not just happy path)

### Anti-Patterns to Avoid:

1. ❌ Returning raw errors (no context)
2. ❌ Using %v instead of %w (breaks unwrapping)
3. ❌ Panicking in library code
4. ❌ Log-and-throw (duplicate logging)
5. ❌ Insufficient error context
6. ❌ Sensitive data in errors
7. ❌ Capitalizing error messages
8. ❌ Retrying permanent errors

---

**Status**: PROPOSED (Iteration 2)
**Next**: Validate in implementation (Iteration 3)
**Generated By**: convention-definer (Bootstrap-013)
