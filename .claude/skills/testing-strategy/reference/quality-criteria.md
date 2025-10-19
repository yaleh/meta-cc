# Test Quality Standards

**Version**: 2.0
**Source**: Bootstrap-002 Test Strategy Development
**Last Updated**: 2025-10-18

This document defines quality criteria, coverage targets, and best practices for test development.

---

## Test Quality Checklist

For every test, ensure compliance with these quality standards:

### Structure

- [ ] Test name clearly describes scenario
- [ ] Setup is minimal and focused
- [ ] Single concept tested per test
- [ ] Clear error messages with context

### Execution

- [ ] Cleanup handled (defer, t.Cleanup)
- [ ] No hard-coded paths or values
- [ ] Deterministic (no randomness)
- [ ] Fast execution (<100ms for unit tests)

### Coverage

- [ ] Tests both happy and error paths
- [ ] Uses test helpers where appropriate
- [ ] Follows documented patterns
- [ ] Includes edge cases

---

## CLI Test Additional Checklist

When testing CLI commands, also ensure:

- [ ] Command flags reset between tests
- [ ] Output captured properly (stdout/stderr)
- [ ] Environment variables reset (if used)
- [ ] Working directory restored (if changed)
- [ ] Temporary files cleaned up
- [ ] No dependency on external binaries (unless integration test)
- [ ] Tests both happy path and error cases
- [ ] Help text validated (if command has help)

---

## Coverage Target Goals

### By Category

Different code categories require different coverage levels based on criticality:

| Category | Target Coverage | Priority | Rationale |
|----------|----------------|----------|-----------|
| Error Handling | 80-90% | P1 | Critical for reliability |
| Business Logic | 75-85% | P2 | Core functionality |
| CLI Handlers | 70-80% | P2 | User-facing behavior |
| Integration | 70-80% | P3 | End-to-end validation |
| Utilities | 60-70% | P3 | Supporting functions |
| Infrastructure | 40-60% | P4 | Best effort |

**Overall Project Target**: 75-80%

### Priority Decision Tree

```
Is function critical to core functionality?
├─ YES: Is it error handling or validation?
│  ├─ YES: Priority 1 (80%+ coverage target)
│  └─ NO: Is it business logic?
│     ├─ YES: Priority 2 (75%+ coverage)
│     └─ NO: Priority 3 (60%+ coverage)
└─ NO: Is it infrastructure/initialization?
   ├─ YES: Priority 4 (test if easy, skip if hard)
   └─ NO: Priority 5 (skip)
```

---

## Test Naming Conventions

### Unit Tests

```go
// Format: TestFunctionName_Scenario
TestValidateInput_NilInput
TestValidateInput_EmptyInput
TestProcessData_ValidFormat
```

### Table-Driven Tests

```go
// Format: TestFunctionName (scenarios in table)
TestValidateInput  // Table contains: "nil input", "empty input", etc.
TestProcessData    // Table contains: "valid format", "invalid format", etc.
```

### Integration Tests

```go
// Format: TestHandler_Scenario or TestIntegration_Feature
TestQueryTools_SuccessfulQuery
TestGetSessionStats_ErrorHandling
TestIntegration_CompleteWorkflow
```

---

## Test Structure Best Practices

### Setup-Execute-Assert Pattern

```go
func TestFunction(t *testing.T) {
    // Setup: Create test data and dependencies
    input := createTestInput()
    mock := createMockDependency()

    // Execute: Call the function under test
    result, err := Function(input, mock)

    // Assert: Verify expected behavior
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

### Cleanup Handling

```go
func TestFunction(t *testing.T) {
    // Using defer for cleanup
    originalValue := globalVar
    defer func() { globalVar = originalValue }()

    // Or using t.Cleanup (preferred)
    t.Cleanup(func() {
        globalVar = originalValue
    })

    // Test logic...
}
```

### Helper Functions

```go
// Mark as helper for better error reporting
func createTestInput(t *testing.T) *Input {
    t.Helper()  // Errors will point to caller, not this line

    return &Input{
        Field1: "test",
        Field2: 42,
    }
}
```

---

## Error Message Guidelines

### Good Error Messages

```go
// Include context and actual values
if result != expected {
    t.Errorf("Function() = %v, expected %v", result, expected)
}

// Include relevant state
if len(results) != expectedCount {
    t.Errorf("got %d results, expected %d: %+v",
        len(results), expectedCount, results)
}
```

### Poor Error Messages

```go
// Avoid: No context
if err != nil {
    t.Fatal("error occurred")
}

// Avoid: Missing actual values
if !valid {
    t.Error("validation failed")
}
```

---

## Test Performance Standards

### Unit Tests

- **Target**: <100ms per test
- **Maximum**: <500ms per test
- **If slower**: Consider mocking or refactoring

### Integration Tests

- **Target**: <1s per test
- **Maximum**: <5s per test
- **If slower**: Use `testing.Short()` to skip in short mode

```go
func TestIntegration_SlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping slow integration test in short mode")
    }
    // Test logic...
}
```

### Running Tests

```bash
# Fast tests only
go test -short ./...

# All tests with timeout
go test -timeout 5m ./...
```

---

## Test Data Management

### Inline Test Data

For small, simple data:

```go
tests := []struct {
    name  string
    input string
}{
    {"empty", ""},
    {"single", "a"},
    {"multiple", "abc"},
}
```

### Fixture Files

For complex data structures:

```go
func loadTestFixture(t *testing.T, name string) []byte {
    t.Helper()
    data, err := os.ReadFile(filepath.Join("testdata", name))
    if err != nil {
        t.Fatalf("failed to load fixture %s: %v", name, err)
    }
    return data
}
```

### Golden Files

For output validation:

```go
func TestFormatOutput(t *testing.T) {
    output := formatOutput(testData)

    goldenPath := filepath.Join("testdata", "expected_output.golden")

    if *update {
        os.WriteFile(goldenPath, []byte(output), 0644)
    }

    expected, _ := os.ReadFile(goldenPath)
    if string(expected) != output {
        t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", output, expected)
    }
}
```

---

## Common Anti-Patterns to Avoid

### 1. Testing Implementation Instead of Behavior

```go
// Bad: Tests internal implementation
func TestFunction(t *testing.T) {
    obj := New()
    if obj.internalField != "expected" {  // Don't test internals
        t.Error("internal field wrong")
    }
}

// Good: Tests observable behavior
func TestFunction(t *testing.T) {
    obj := New()
    result := obj.PublicMethod()  // Test public interface
    if result != expected {
        t.Error("unexpected result")
    }
}
```

### 2. Overly Complex Test Setup

```go
// Bad: Complex setup obscures test intent
func TestFunction(t *testing.T) {
    // 50 lines of setup...
    result := Function(complex, setup, params)
    // Assert...
}

// Good: Use helper functions
func TestFunction(t *testing.T) {
    setup := createTestSetup(t)  // Helper abstracts complexity
    result := Function(setup)
    // Assert...
}
```

### 3. Testing Multiple Concepts in One Test

```go
// Bad: Tests multiple unrelated things
func TestValidation(t *testing.T) {
    // Tests format validation
    // Tests length validation
    // Tests encoding validation
    // Tests error handling
}

// Good: Separate tests for each concept
func TestValidation_Format(t *testing.T) { /*...*/ }
func TestValidation_Length(t *testing.T) { /*...*/ }
func TestValidation_Encoding(t *testing.T) { /*...*/ }
func TestValidation_ErrorHandling(t *testing.T) { /*...*/ }
```

### 4. Shared State Between Tests

```go
// Bad: Tests depend on execution order
var sharedState string

func TestFirst(t *testing.T) {
    sharedState = "initialized"
}

func TestSecond(t *testing.T) {
    // Breaks if TestFirst doesn't run first
    if sharedState != "initialized" { /*...*/ }
}

// Good: Each test is independent
func TestFirst(t *testing.T) {
    state := "initialized"  // Local state
    // Test...
}

func TestSecond(t *testing.T) {
    state := setupState()  // Creates own state
    // Test...
}
```

---

## Code Review Checklist for Tests

When reviewing test code, verify:

- [ ] Tests are independent (can run in any order)
- [ ] Test names are descriptive
- [ ] Happy path and error paths both covered
- [ ] Edge cases included
- [ ] No magic numbers or strings (use constants)
- [ ] Cleanup handled properly
- [ ] Error messages provide context
- [ ] Tests are reasonably fast
- [ ] No commented-out test code
- [ ] Follows established patterns in codebase

---

## Continuous Improvement

### Track Test Metrics

Record for each test batch:

```
Date: 2025-10-18
Batch: Validation error paths (4 tests)
Pattern: Error Path + Table-Driven
Time: 50 min (estimated 60 min) → 17% faster
Coverage: internal/validation 57.9% → 75.2% (+17.3%)
Total coverage: 72.3% → 73.5% (+1.2%)
Efficiency: 0.3% per test
Issues: None
Lessons: Table-driven error tests very efficient
```

### Regular Coverage Analysis

```bash
# Weekly coverage review
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -20

# Identify degradation
diff coverage-last-week.txt coverage-this-week.txt
```

### Test Suite Health

Monitor:
- Total test count (growing)
- Test execution time (stable or decreasing)
- Coverage percentage (stable or increasing)
- Flaky test rate (near zero)
- Test maintenance time (decreasing)

---

**Source**: Bootstrap-002 Test Strategy Development
**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: Production-ready, validated through 4 iterations
