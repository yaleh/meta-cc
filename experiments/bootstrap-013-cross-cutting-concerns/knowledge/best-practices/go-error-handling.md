# Go Error Handling Best Practices

**Domain**: Cross-Cutting Concerns
**Language**: Go
**Topic**: Error Handling
**Status**: VALIDATED (Iteration 2)

---

## Overview

This document captures best practices for error handling in Go applications, based on Go 1.13+ standard library features and community best practices.

---

## 1. Always Wrap Errors with Context

**Practice**: Use `fmt.Errorf` with `%w` to wrap errors while adding context

**Rationale**:
- Preserves error chain (unwrappable with `errors.Unwrap`)
- Adds context at each layer
- Enables `errors.Is` and `errors.As` checking
- Provides rich debugging information

**Example**:
```go
// ✓ Good: Wraps error with context
if err := operation(); err != nil {
    return fmt.Errorf("failed to process user %s: %w", userID, err)
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

---

## 2. Include Sufficient Context in Errors

**Practice**: Error messages should answer "what failed", "where", and "why"

**Essential Context**:
- **Operation**: What was being attempted
- **Entity**: What was being operated on
- **Identifiers**: Relevant IDs, paths, names
- **Original Error**: Via %w

**Example**:
```go
// ✓ Good: Complete context
return fmt.Errorf("failed to parse session file %s at line %d: %w",
    filePath, lineNum, err)

// ✗ Bad: Insufficient context
return fmt.Errorf("operation failed: %w", err)

// ✗ Bad: Too verbose
return fmt.Errorf("failed in function %s at %s line %d timestamp %s: %w",
    funcName, fileName, lineNum, time.Now(), err)
```

---

## 3. Use Sentinel Errors for Expected Conditions

**Practice**: Define package-level sentinel errors for conditions callers may handle

**When to Use**:
- Expected error conditions
- Errors that need programmatic checking
- Common errors across functions

**Example**:
```go
package query

import "errors"

// Sentinel errors
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
if errors.Is(err, query.ErrNoResults) {
    // Handle no results case
}
```

---

## 4. Use Custom Error Types for Structured Data

**Practice**: Create custom error types when you need to attach metadata

**When to Use**:
- Errors need structured data (codes, fields, metadata)
- Special formatting or serialization required
- Method-based error behavior needed

**Example**:
```go
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

// Checking
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Printf("Invalid field: %s\n", valErr.Field)
}
```

---

## 5. Never Panic in Library Code

**Practice**: Libraries should ALWAYS return errors, NEVER panic

**Rationale**:
- Libraries can't know how calling code wants to handle errors
- Panics force recovery or crash
- Errors allow flexible handling

**Example**:
```go
// ✓ Good: Library returns error
func ParseSession(path string) ([]Entry, error) {
    if path == "" {
        return nil, errors.New("path cannot be empty")
    }
    // ... parse ...
    return entries, nil
}

// ✗ Bad: Library panics
func ParseSession(path string) []Entry {
    if path == "" {
        panic("path cannot be empty") // NEVER!
    }
    // ... parse ...
}

// ✓ OK: main() may panic for unrecoverable errors
func main() {
    if os.Getenv("REQUIRED_VAR") == "" {
        panic("REQUIRED_VAR must be set") // OK in main
    }
}
```

---

## 6. Log Errors at Top Level Only

**Practice**: Return errors from helpers, log only at top-level handlers

**Rationale**:
- Prevents duplicate logging
- Keeps helpers reusable
- Centralizes error handling

**Anti-Pattern** (Log-and-Throw):
```go
// ✗ Bad: Helper logs error before returning
func helper() error {
    if err := operation(); err != nil {
        log.Error("operation failed", "error", err) // Logged here
        return err // Will be logged again!
    }
    return nil
}

func caller() error {
    if err := helper(); err != nil {
        log.Error("helper failed", "error", err) // Duplicate!
        return err
    }
    return nil
}
```

**Correct Pattern**:
```go
// ✓ Good: Helper just wraps and returns
func helper() error {
    if err := operation(); err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }
    return nil
}

// ✓ Good: Top-level logs once
func topLevel() {
    if err := helper(); err != nil {
        log.Error("request failed", "error", err) // Log ONCE
    }
}
```

---

## 7. Use Lowercase Error Messages

**Practice**: Error messages start with lowercase (Go convention)

**Rationale**:
- Go convention
- Errors are often composed into larger messages
- Consistent style

**Example**:
```go
// ✓ Good
errors.New("failed to connect to database")
errors.New("session file not found")
fmt.Errorf("invalid query syntax: expected SELECT, got %s", token)

// ✗ Bad: Capitalized
errors.New("Failed to connect")

// ✗ Bad: Trailing punctuation
errors.New("failed to connect.")
```

---

## 8. Classify Errors for Structured Logging

**Practice**: Categorize errors to enable filtering and analysis

**Example**:
```go
func classifyError(err error) string {
    if err == nil {
        return ""
    }

    // Check sentinel errors
    if errors.Is(err, ErrNotFound) {
        return "not_found"
    }
    if errors.Is(err, ErrTimeout) {
        return "timeout_error"
    }

    // Check custom types
    var valErr *ValidationError
    if errors.As(err, &valErr) {
        return "validation_error"
    }

    // Pattern matching
    errMsg := err.Error()
    if strings.Contains(errMsg, "parse") {
        return "parse_error"
    }
    if strings.Contains(errMsg, "network") {
        return "network_error"
    }

    return "general_error"
}

// Usage with logging
log.Error("operation failed",
    "error", err.Error(),
    "error_type", classifyError(err))
```

---

## 9. Implement Retry Logic for Transient Errors

**Practice**: Retry only transient errors, fail fast on permanent errors

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

**Example**:
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
    if errors.Is(err, ErrTimeout) {
        return true
    }

    errMsg := err.Error()
    return strings.Contains(errMsg, "timeout") ||
           strings.Contains(errMsg, "connection refused") ||
           strings.Contains(errMsg, "status 5")
}
```

---

## 10. Test Error Paths, Not Just Happy Paths

**Practice**: Write tests for both success and error cases

**Example**:
```go
func TestParseSession(t *testing.T) {
    // Test success
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

func TestSentinelErrors(t *testing.T) {
    err := ExecuteQuery("")

    // Use errors.Is for checking
    if !errors.Is(err, ErrInvalidQuery) {
        t.Errorf("expected ErrInvalidQuery, got: %v", err)
    }
}

func TestCustomErrorTypes(t *testing.T) {
    err := ValidateUser(&User{Age: -1})

    // Use errors.As for custom types
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

## 11. Avoid Sensitive Data in Errors

**Practice**: Never include passwords, tokens, PII in error messages

**Sensitive Data**:
- Passwords, API keys, tokens
- Personally Identifiable Information (email, phone, SSN)
- Financial data (credit card numbers)
- Health information

**Example**:
```go
// ✗ Bad: Sensitive data exposed
return fmt.Errorf("failed to authenticate user %s with password %s: %w",
    username, password, err) // NEVER!

// ✓ Good: Non-sensitive identifiers only
return fmt.Errorf("failed to authenticate user %s: %w", userID, err)
```

---

## 12. Make Error Messages Actionable

**Practice**: When possible, suggest how to fix the problem

**Example**:
```go
// ✓ Good: Suggests remediation
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

## 13. Use errors.Is and errors.As Correctly

**Practice**: Use `errors.Is` for sentinel errors, `errors.As` for types

**errors.Is**:
```go
// Check if err is or wraps a specific sentinel error
if errors.Is(err, ErrNotFound) {
    // Handle not found
}

if errors.Is(err, os.ErrNotExist) {
    // Handle file not exist
}
```

**errors.As**:
```go
// Extract specific error type from chain
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Printf("Invalid field: %s\n", valErr.Field)
}

var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Printf("Path operation failed: %s\n", pathErr.Path)
}
```

---

## Common Anti-Patterns

### 1. Returning Raw Errors

❌ Problem: Loses context

```go
// Bad
func helper() error {
    if err := operation(); err != nil {
        return err // No context!
    }
    return nil
}

// Good
func helper() error {
    if err := operation(); err != nil {
        return fmt.Errorf("helper operation failed: %w", err)
    }
    return nil
}
```

### 2. Using %v Instead of %w

❌ Problem: Breaks error unwrapping

```go
// Bad: Can't use errors.Is/As
return fmt.Errorf("failed: %v", err)

// Good: Preserves error chain
return fmt.Errorf("failed: %w", err)
```

### 3. Log-and-Throw

❌ Problem: Duplicate logging

```go
// Bad: Logged twice
func helper() error {
    if err := operation(); err != nil {
        log.Error("failed", "error", err) // Log 1
        return err // Will be logged again!
    }
    return nil
}

// Good: Log at top level only
func helper() error {
    if err := operation(); err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }
    return nil
}
```

### 4. Panicking in Libraries

❌ Problem: Forces recovery or crash

```go
// Bad: Library panics
func Parse(input string) Data {
    if input == "" {
        panic("empty input") // Wrong!
    }
    // ...
}

// Good: Library returns error
func Parse(input string) (Data, error) {
    if input == "" {
        return Data{}, errors.New("empty input")
    }
    // ...
}
```

### 5. Insufficient Error Context

❌ Problem: Can't debug issues

```go
// Bad: Too vague
return fmt.Errorf("error: %w", err)

// Good: Sufficient context
return fmt.Errorf("failed to parse session file %s at line %d: %w",
    filePath, lineNum, err)
```

### 6. Retrying Permanent Errors

❌ Problem: Wastes time and resources

```go
// Bad: Retries 404 (permanent)
for i := 0; i < 10; i++ {
    err := fetch(url)
    if err != nil {
        time.Sleep(time.Second)
        continue // Retries even for 404!
    }
    return nil
}

// Good: Check if error is transient
for i := 0; i < 10; i++ {
    err := fetch(url)
    if err == nil {
        return nil
    }
    if !isTransient(err) {
        return err // Don't retry permanent errors
    }
    time.Sleep(time.Second)
}
```

---

## References

- [Go Error Handling Blog](https://go.dev/blog/error-handling-and-go)
- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Uber Go Style Guide - Errors](https://github.com/uber-go/guide/blob/master/style.md#errors)
- [Dave Cheney - Don't just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
- [pkg/errors (deprecated, but good concepts)](https://github.com/pkg/errors)

---

**Status**: VALIDATED
**Source**: Iteration 2 (Bootstrap-013)
**Validation**: Derived from Go 1.13+ standard library and community best practices
