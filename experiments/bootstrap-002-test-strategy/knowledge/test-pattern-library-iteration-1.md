# Test Pattern Library - Iteration 1

**Created**: 2025-10-18
**Experiment**: Bootstrap-002 Test Strategy Development
**Status**: Foundation Established

---

## Overview

This document codifies test patterns observed and created during Iteration 0-1 of the test strategy experiment. These patterns provide reusable templates for systematic test coverage improvement.

---

## Pattern 1: Unit Test Pattern

**Purpose**: Test a single function or method in isolation

**Structure**:
```go
func TestFunctionName_Scenario(t *testing.T) {
    // Setup
    input := createTestInput()

    // Execute
    result, err := FunctionUnderTest(input)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

**Key Characteristics**:
- Clear test name: `TestFunctionName_Scenario`
- Single assertion path
- Focused on one behavior
- Clear error messages with context

**Example from Codebase**:
```go
func TestIsStandardParameter(t *testing.T) {
    tests := []struct {
        param    string
        expected bool
    }{
        {"scope", true},
        {"custom_param", false},
    }

    for _, tt := range tests {
        t.Run(tt.param, func(t *testing.T) {
            result := isStandardParameter(tt.param)
            if result != tt.expected {
                t.Errorf("isStandardParameter(%s) = %v, expected %v",
                    tt.param, result, tt.expected)
            }
        })
    }
}
```

**When to Use**:
- Testing pure functions (no side effects)
- Simple input/output validation
- Error path testing
- Parameter validation

---

## Pattern 2: Table-Driven Test Pattern

**Purpose**: Test multiple scenarios with the same test logic

**Structure**:
```go
func TestFunction_MultipleScenarios(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    validInput,
            expected: validOutput,
            wantErr:  false,
        },
        {
            name:     "invalid input",
            input:    invalidInput,
            expected: zeroValue,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr && result != tt.expected {
                t.Errorf("Function() = %v, expected %v", result, tt.expected)
            }
        })
    }
}
```

**Key Characteristics**:
- Test table with named scenarios
- Subtests with `t.Run()`
- DRY (Don't Repeat Yourself) principle
- Comprehensive coverage through parameterization

**Example from Codebase**:
```go
func TestFindClosingBrace_Simple(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int
    }{
        {
            name:     "simple brace",
            input:    "{abc}",
            expected: 4,
        },
        {
            name:     "nested braces",
            input:    "{a{b}c}",
            expected: 6,
        },
        {
            name:     "no closing",
            input:    "{abc",
            expected: -1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := findClosingBrace(tt.input)
            if result != tt.expected {
                t.Errorf("Expected %d, got %d for input '%s'",
                    tt.expected, result, tt.input)
            }
        })
    }
}
```

**When to Use**:
- Testing boundary conditions
- Multiple input variations
- Comprehensive error path coverage
- Regression test suites

---

## Pattern 3: Integration Test Pattern (MCP Server)

**Purpose**: Test complete request/response flow through handlers

**Structure**:
```go
func TestHandlerName(t *testing.T) {
    // Setup: Create request
    req := JSONRPCRequest{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "method/name",
        Params:  map[string]interface{}{
            "param": "value",
        },
    }

    // Setup: Capture output
    var buf bytes.Buffer
    origStdout := outputWriter
    outputWriter = &buf
    defer func() { outputWriter = origStdout }()

    // Execute: Call handler
    handleMethod(context.Background(), req)

    // Assert: Parse response
    var resp JSONRPCResponse
    if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
        t.Fatalf("failed to parse response: %v", err)
    }

    // Assert: Validate response structure
    if resp.JSONRPC != "2.0" {
        t.Errorf("expected jsonrpc=2.0, got %s", resp.JSONRPC)
    }

    if resp.Error != nil {
        t.Errorf("expected no error, got %v", resp.Error)
    }

    // Assert: Validate response content
    result, ok := resp.Result.(map[string]interface{})
    if !ok {
        t.Fatal("expected result to be a map")
    }

    // Verify specific fields
    if value, ok := result["field"]; !ok || value != expected {
        t.Errorf("expected field=%v, got %v", expected, value)
    }
}
```

**Key Characteristics**:
- Uses `bytes.Buffer` to capture stdout
- Tests complete request/response cycle
- Validates JSON-RPC protocol compliance
- Checks both structure and content

**Example from Codebase**:
```go
func TestHandleInitialize(t *testing.T) {
    req := JSONRPCRequest{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "initialize",
        Params:  map[string]interface{}{},
    }

    var buf bytes.Buffer
    origStdout := outputWriter
    outputWriter = &buf
    defer func() { outputWriter = origStdout }()

    handleInitialize(context.Background(), req)

    var resp JSONRPCResponse
    if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
        t.Fatalf("failed to parse response: %v", err)
    }

    if resp.JSONRPC != "2.0" {
        t.Errorf("expected jsonrpc=2.0, got %s", resp.JSONRPC)
    }

    result, ok := resp.Result.(map[string]interface{})
    if !ok {
        t.Fatal("expected result to be a map")
    }

    if _, hasVersion := result["protocolVersion"]; !hasVersion {
        t.Error("expected protocolVersion in result")
    }
}
```

**When to Use**:
- Testing MCP server handlers
- HTTP endpoint testing (with httptest)
- End-to-end request/response validation
- Protocol compliance testing

---

## Pattern 4: Error Path Test Pattern

**Purpose**: Systematically test error handling and edge cases

**Structure**:
```go
func TestFunction_ErrorCases(t *testing.T) {
    tests := []struct {
        name      string
        input     InputType
        wantErr   bool
        errMsg    string
        errCode   int  // For structured errors
    }{
        {
            name:    "nil input",
            input:   nil,
            wantErr: true,
            errMsg:  "input cannot be nil",
        },
        {
            name:    "empty input",
            input:   InputType{},
            wantErr: true,
            errMsg:  "input cannot be empty",
        },
        {
            name:    "invalid format",
            input:   invalidFormatInput,
            wantErr: true,
            errMsg:  "invalid format",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := Function(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("expected error containing '%s', got '%s'",
                    tt.errMsg, err.Error())
            }
        })
    }
}
```

**Key Characteristics**:
- Focuses on error conditions
- Tests error messages
- Validates error types/codes
- Covers edge cases

**Example from Codebase**:
```go
func TestHandleRequest_UnknownMethod(t *testing.T) {
    req := JSONRPCRequest{
        JSONRPC: "2.0",
        ID:      3,
        Method:  "unknown/method",
        Params:  map[string]interface{}{},
    }

    var buf bytes.Buffer
    origStdout := outputWriter
    outputWriter = &buf
    defer func() { outputWriter = origStdout }()

    handleRequest(req)

    var resp JSONRPCResponse
    if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
        t.Fatalf("failed to parse response: %v", err)
    }

    if resp.Error == nil {
        t.Error("expected error for unknown method")
    }

    if resp.Error.Code != -32601 {
        t.Errorf("expected error code -32601, got %d", resp.Error.Code)
    }
}
```

**When to Use**:
- Testing validation logic
- Boundary condition testing
- Error recovery mechanisms
- Input sanitization verification

---

## Pattern 5: Test Helper Pattern

**Purpose**: Reduce duplication and improve test maintainability

**Structure**:
```go
// Test helper function
func createTestInput(t *testing.T, options ...Option) *InputType {
    t.Helper()  // Mark as helper for better error reporting

    input := &InputType{
        // Default values
        Field1: "default",
        Field2: 42,
    }

    // Apply options
    for _, opt := range options {
        opt(input)
    }

    return input
}

// Option pattern for flexibility
type Option func(*InputType)

func WithField1(value string) Option {
    return func(i *InputType) {
        i.Field1 = value
    }
}

// Usage in tests
func TestFunction_WithHelper(t *testing.T) {
    input := createTestInput(t, WithField1("custom"))

    result, err := Function(input)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // assertions...
}
```

**Key Characteristics**:
- Marked with `t.Helper()`
- Centralized test data creation
- Flexible with options pattern
- Reduces duplication

**Example from Codebase**:
```go
// Helper function to create temporary test file
func createTempFile(t *testing.T, content string) string {
    t.Helper()

    tmpDir := t.TempDir()
    tmpFile := filepath.Join(tmpDir, "test_tools.go")

    err := os.WriteFile(tmpFile, []byte(content), 0644)
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }

    return tmpFile
}

// Usage
func TestParseTools_ValidFile(t *testing.T) {
    content := `package tools

func getToolDefinitions() []ToolDefinition {
    return []ToolDefinition{...}
}`

    tmpFile := createTempFile(t, content)
    defer os.Remove(tmpFile)

    tools, err := ParseTools(tmpFile)
    // assertions...
}
```

**When to Use**:
- Complex test setup
- Repeated fixture creation
- Test data builders
- Cleanup operations

---

## Coverage-Driven Workflow

### Step 1: Identify Gaps
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# Find low-coverage functions
go tool cover -func=coverage.out | grep -E "^[^:]+:[^:]+:[[:space:]]+[0-9]+\.[0-9]%" | awk '$NF < 80.0'
```

### Step 2: Prioritize
1. **Critical paths** (error handling, validation): Must have 80%+ coverage
2. **Business logic** (core functionality): Target 75-85% coverage
3. **Integration points** (handlers, I/O): Target 70-80% coverage
4. **Utilities** (helpers, formatters): Target 60-70% coverage

### Step 3: Select Pattern
- **Unit tests**: Pure functions, simple logic
- **Table-driven**: Multiple scenarios, edge cases
- **Integration**: Handlers, endpoints, complete flows
- **Error path**: Validation, boundary conditions

### Step 4: Write Test
- Follow selected pattern template
- Include happy path AND error paths
- Use clear naming and assertions
- Add comments for complex scenarios

### Step 5: Verify
```bash
# Run tests
go test ./...

# Check coverage improved
go test -coverprofile=new_coverage.out ./...
go tool cover -func=new_coverage.out | tail -1
```

---

## Quality Checklist

For each test, verify:

- [ ] Test name clearly describes scenario
- [ ] Setup is minimal and focused
- [ ] Single concept tested per test
- [ ] Error messages include context
- [ ] Cleanup handled (defer, t.Cleanup)
- [ ] No hard-coded paths or values
- [ ] Deterministic (no randomness)
- [ ] Fast execution (<100ms for unit tests)
- [ ] Tests both happy and error paths
- [ ] Uses test helpers where appropriate

---

## Common Pitfalls

### 1. Overly Complex Tests
**Problem**: Tests with too many assertions or complex setup
**Solution**: Break into multiple focused tests

### 2. Missing Error Path Coverage
**Problem**: Only testing happy paths
**Solution**: Add table-driven error tests for each function

### 3. Brittle Tests
**Problem**: Tests break on minor implementation changes
**Solution**: Test behavior, not implementation details

### 4. Slow Tests
**Problem**: Integration tests that take seconds
**Solution**: Mock external dependencies, use httptest

### 5. Unclear Failure Messages
**Problem**: "Expected true, got false"
**Solution**: Include context: "Expected valid email, got invalid format"

---

## Metrics

### Pattern Distribution (Iteration 1)
- **Unit tests**: ~60% (stable)
- **Table-driven tests**: ~30% (stable)
- **Integration tests**: ~8% (↑ from ~10%, added MCP server tests)
- **Error path tests**: ~17% (↑ from ~15%, needs more work)

### Coverage Impact
- **Iteration 0 baseline**: 72.1%
- **Iteration 1 baseline** (after parser fix): 71.3%
- **Target**: 80%
- **Gap**: -8.7 percentage points

### Test Count
- **Total test functions**: 590+ (baseline)
- **Tests added Iteration 1**: 5 (handleToolsCall variants)
- **Subtests**: 136+ (baseline)

---

## Next Steps

### Iteration 2 Priorities
1. **Fix integration test mocking**: Make handleToolsCall tests pass by improving mocking
2. **Add more MCP server tests**: Target ExecuteTool, capability handlers
3. **Systematic error path coverage**: Add error tests to all validation functions
4. **CLI command tests**: cmd/ package needs significant work (57.9% → 75%+)

### Methodology Improvements Needed
1. **Fixture generator**: Automate test data creation
2. **Coverage-driven test selection**: Tool to suggest next test to write
3. **Mock library**: Reusable mocks for HTTP, file I/O
4. **Test template generator**: CLI tool to scaffold tests from function signatures

---

## References

- Go testing documentation: https://golang.org/pkg/testing/
- Table-driven tests: https://github.com/golang/go/wiki/TableDrivenTests
- httptest package: https://golang.org/pkg/net/http/httptest/
- Test helpers: https://golang.org/pkg/testing/#T.Helper

---

**Version**: 1.0
**Last Updated**: 2025-10-18
**Status**: Foundation Established - Ready for Iteration 2
