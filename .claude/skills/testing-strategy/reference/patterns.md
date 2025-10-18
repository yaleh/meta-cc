# Test Pattern Library

**Version**: 2.0
**Source**: Bootstrap-002 Test Strategy Development
**Last Updated**: 2025-10-18

This document provides 8 proven test patterns for Go testing with practical examples and usage guidance.

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

**When to Use**:
- Testing pure functions (no side effects)
- Simple input/output validation
- Single test scenario

**Time**: ~8-10 minutes per test

---

## Pattern 2: Table-Driven Test Pattern

**Purpose**: Test multiple scenarios with the same test logic

**Structure**:
```go
func TestFunction(t *testing.T) {
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

**When to Use**:
- Testing boundary conditions
- Multiple input variations
- Comprehensive coverage

**Time**: ~10-15 minutes for 3-5 scenarios

---

## Pattern 3: Integration Test Pattern

**Purpose**: Test complete request/response flow through handlers

**Structure**:
```go
func TestHandler(t *testing.T) {
    // Setup: Create request
    req := createTestRequest()

    // Setup: Capture output
    var buf bytes.Buffer
    outputWriter = &buf
    defer func() { outputWriter = originalWriter }()

    // Execute
    handleRequest(req)

    // Assert: Parse response
    var resp Response
    if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
        t.Fatalf("failed to parse response: %v", err)
    }

    // Assert: Validate response
    if resp.Error != nil {
        t.Errorf("unexpected error: %v", resp.Error)
    }
}
```

**When to Use**:
- Testing MCP server handlers
- HTTP endpoint testing
- End-to-end flows

**Time**: ~15-20 minutes per test

---

## Pattern 4: Error Path Test Pattern

**Purpose**: Systematically test error handling and edge cases

**Structure**:
```go
func TestFunction_ErrorCases(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        wantErr bool
        errMsg  string
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
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := Function(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
            }
        })
    }
}
```

**When to Use**:
- Testing validation logic
- Boundary condition testing
- Error recovery

**Time**: ~12-15 minutes for 3-4 error cases

---

## Pattern 5: Test Helper Pattern

**Purpose**: Reduce duplication and improve maintainability

**Structure**:
```go
// Test helper function
func createTestInput(t *testing.T, options ...Option) *InputType {
    t.Helper()  // Mark as helper for better error reporting

    input := &InputType{
        Field1: "default",
        Field2: 42,
    }

    for _, opt := range options {
        opt(input)
    }

    return input
}

// Usage
func TestFunction(t *testing.T) {
    input := createTestInput(t, WithField1("custom"))
    result, err := Function(input)
    // ...
}
```

**When to Use**:
- Complex test setup
- Repeated fixture creation
- Test data builders

**Time**: ~5 minutes to create, saves 2-3 min per test using it

---

## Pattern 6: Dependency Injection Pattern

**Purpose**: Test components that depend on external systems

**Structure**:
```go
// 1. Define interface
type Executor interface {
    Execute(args Args) (Result, error)
}

// 2. Production implementation
type RealExecutor struct{}
func (e *RealExecutor) Execute(args Args) (Result, error) {
    // Real implementation
}

// 3. Mock implementation
type MockExecutor struct {
    Results map[string]Result
    Errors  map[string]error
}

func (m *MockExecutor) Execute(args Args) (Result, error) {
    if err, ok := m.Errors[args.Key]; ok {
        return Result{}, err
    }
    return m.Results[args.Key], nil
}

// 4. Tests use mock
func TestProcess(t *testing.T) {
    mock := &MockExecutor{
        Results: map[string]Result{"key": {Value: "expected"}},
    }
    err := ProcessData(mock, testData)
    // ...
}
```

**When to Use**:
- Testing components that execute commands
- Testing HTTP clients
- Testing database operations

**Time**: ~20-25 minutes (includes refactoring)

---

## Pattern 7: CLI Command Test Pattern

**Purpose**: Test Cobra command execution with flags

**Structure**:
```go
func TestCommand(t *testing.T) {
    // Setup: Create command
    cmd := &cobra.Command{
        Use: "command",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Command logic
            return nil
        },
    }

    // Setup: Add flags
    cmd.Flags().StringP("flag", "f", "default", "description")

    // Setup: Set arguments
    cmd.SetArgs([]string{"--flag", "value"})

    // Setup: Capture output
    var buf bytes.Buffer
    cmd.SetOut(&buf)

    // Execute
    err := cmd.Execute()

    // Assert
    if err != nil {
        t.Fatalf("command failed: %v", err)
    }

    // Verify output
    if !strings.Contains(buf.String(), "expected") {
        t.Errorf("unexpected output: %s", buf.String())
    }
}
```

**When to Use**:
- Testing CLI command handlers
- Flag parsing verification
- Command composition testing

**Time**: ~12-15 minutes per test

---

## Pattern 8: Global Flag Test Pattern

**Purpose**: Test global flag parsing and propagation

**Structure**:
```go
func TestGlobalFlags(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        expected GlobalOptions
    }{
        {
            name: "default",
            args: []string{},
            expected: GlobalOptions{ProjectPath: getCwd()},
        },
        {
            name: "with flag",
            args: []string{"--session", "abc"},
            expected: GlobalOptions{SessionID: "abc"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resetGlobalFlags()  // Important: reset state
            rootCmd.SetArgs(tt.args)
            rootCmd.ParseFlags(tt.args)
            opts := getGlobalOptions()

            if opts.SessionID != tt.expected.SessionID {
                t.Errorf("SessionID = %v, expected %v", opts.SessionID, tt.expected.SessionID)
            }
        })
    }
}
```

**When to Use**:
- Testing global flag parsing
- Flag interaction testing
- Option struct population

**Time**: ~10-12 minutes (table-driven, high efficiency)

---

## Pattern Selection Decision Tree

```
What are you testing?
├─ CLI command with flags?
│  ├─ Multiple flag combinations? → Pattern 8 (Global Flag)
│  ├─ Integration test needed? → Pattern 7 (CLI Command)
│  └─ Command execution? → Pattern 7 (CLI Command)
├─ Error paths?
│  ├─ Multiple error scenarios? → Pattern 4 (Error Path) + Pattern 2 (Table-Driven)
│  └─ Single error case? → Pattern 4 (Error Path)
├─ Unit function?
│  ├─ Multiple inputs? → Pattern 2 (Table-Driven)
│  └─ Single input? → Pattern 1 (Unit Test)
├─ External dependency?
│  └─ → Pattern 6 (Dependency Injection)
└─ Integration flow?
   └─ → Pattern 3 (Integration Test)
```

---

## Pattern Efficiency Metrics

**Time per Test** (measured):
- Unit Test (Pattern 1): ~8 min
- Table-Driven (Pattern 2): ~12 min (3-4 scenarios)
- Integration Test (Pattern 3): ~18 min
- Error Path (Pattern 4): ~14 min (4 scenarios)
- Test Helper (Pattern 5): ~5 min to create
- Dependency Injection (Pattern 6): ~22 min (includes refactoring)
- CLI Command (Pattern 7): ~13 min
- Global Flag (Pattern 8): ~11 min

**Coverage Impact per Test**:
- Table-Driven: 0.20-0.30% total coverage (high impact)
- Error Path: 0.10-0.15% total coverage
- CLI Command: 0.15-0.25% total coverage
- Unit Test: 0.10-0.20% total coverage

**Best ROI Patterns**:
1. Global Flag Tests (Pattern 8): High coverage, fast execution
2. Table-Driven Tests (Pattern 2): Multiple scenarios, efficient
3. Error Path Tests (Pattern 4): Critical coverage, systematic

---

**Source**: Bootstrap-002 Test Strategy Development
**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: Production-ready, validated through 4 iterations
