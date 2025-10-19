# Mocking Patterns for MCP Server Tests - Iteration 2

**Created**: 2025-10-18
**Experiment**: Bootstrap-002 Test Strategy Development
**Status**: Active

---

## Problem Statement

MCP server integration tests (`handle_tools_call_test.go`) fail because they attempt to execute real `meta-cc` commands:

```
ERROR meta-cc command failed exit_code=1
ERROR tool execution failed error="meta-cc command ... failed with exit code"
```

**Root Cause**: Tests call `ExecuteTool()` which runs real shell commands via `exec.Command()`, but:
1. The `meta-cc` binary may not be built or in PATH during tests
2. Tests should not depend on external binaries
3. Tests should be fast and isolated

---

## Solution: Dependency Injection Pattern

### Pattern: Executor Interface

**Purpose**: Allow tests to inject mock executors instead of using real command execution

**Implementation Strategy**:

```go
// 1. Define interface for tool execution
type ToolExecutor interface {
    ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (*ToolResult, error)
}

// 2. Default implementation (production)
type RealToolExecutor struct {
    cfg *config.Config
}

func (e *RealToolExecutor) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (*ToolResult, error) {
    // Current ExecuteTool() logic
    return &ToolResult{...}, nil
}

// 3. Mock implementation (tests)
type MockToolExecutor struct {
    Results map[string]*ToolResult  // Tool name -> result
    Errors  map[string]error         // Tool name -> error
    Calls   []MockCall               // Track calls for verification
}

type MockCall struct {
    ToolName  string
    Arguments map[string]interface{}
    Timestamp time.Time
}

func (m *MockToolExecutor) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (*ToolResult, error) {
    // Record call
    m.Calls = append(m.Calls, MockCall{
        ToolName:  toolName,
        Arguments: args,
        Timestamp: time.Now(),
    })

    // Return mock result or error
    if err, ok := m.Errors[toolName]; ok {
        return nil, err
    }

    if result, ok := m.Results[toolName]; ok {
        return result, nil
    }

    // Default: return empty success
    return &ToolResult{Content: []interface{}{}}, nil
}

// 4. Modify handleToolsCall to accept executor
var defaultExecutor ToolExecutor

func init() {
    defaultExecutor = &RealToolExecutor{cfg: cfg}
}

func handleToolsCall(ctx context.Context, req JSONRPCRequest) {
    handleToolsCallWithExecutor(ctx, req, defaultExecutor)
}

func handleToolsCallWithExecutor(ctx context.Context, req JSONRPCRequest, executor ToolExecutor) {
    // ... validation ...

    result, err := executor.ExecuteTool(ctx, toolName, arguments)

    // ... response handling ...
}

// 5. Tests use mock executor
func TestHandleToolsCall_Success(t *testing.T) {
    mock := &MockToolExecutor{
        Results: map[string]*ToolResult{
            "get_session_stats": {
                Content: []interface{}{
                    map[string]interface{}{
                        "type": "text",
                        "text": "Stats result",
                    },
                },
            },
        },
    }

    req := JSONRPCRequest{...}

    var buf bytes.Buffer
    outputWriter = &buf

    handleToolsCallWithExecutor(context.Background(), req, mock)

    // Verify mock was called
    if len(mock.Calls) != 1 {
        t.Errorf("expected 1 call, got %d", len(mock.Calls))
    }

    // Verify response
    var resp JSONRPCResponse
    json.Unmarshal(buf.Bytes(), &resp)

    if resp.Error != nil {
        t.Errorf("unexpected error: %v", resp.Error)
    }
}
```

---

## Alternative: Mock at Command Layer

**Purpose**: Mock `exec.Command` instead of executor

**Pros**:
- Less refactoring
- Tests command construction logic

**Cons**:
- More complex mocking
- Tests become brittle (depend on exact command strings)
- Harder to verify behavior

**Not Recommended**: Executor interface approach is cleaner and more maintainable.

---

## Pattern Library Integration

### Pattern 6: Dependency Injection Test Pattern

**Purpose**: Test components that depend on external systems (commands, HTTP, databases)

**Structure**:
```go
// 1. Define interface
type Dependency interface {
    DoWork(args Args) (Result, error)
}

// 2. Production implementation
type RealDependency struct{}
func (d *RealDependency) DoWork(args Args) (Result, error) {
    // Real implementation
}

// 3. Mock implementation
type MockDependency struct {
    ExpectedCalls []MockCall
    Results       map[string]Result
    Errors        map[string]error
}

func (m *MockDependency) DoWork(args Args) (Result, error) {
    // Record call and return mock data
}

// 4. Component uses interface
func ProcessData(dep Dependency, data Data) error {
    result, err := dep.DoWork(extractArgs(data))
    return handleResult(result, err)
}

// 5. Tests inject mock
func TestProcessData_Success(t *testing.T) {
    mock := &MockDependency{
        Results: map[string]Result{
            "key": {Value: "expected"},
        },
    }

    err := ProcessData(mock, testData)

    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }

    // Verify mock was called correctly
    if len(mock.ExpectedCalls) != 1 {
        t.Errorf("expected 1 call")
    }
}
```

**Key Characteristics**:
- Separates interface from implementation
- Production code uses default implementation
- Tests inject mock implementations
- Fast execution (no external dependencies)
- Deterministic results

**When to Use**:
- Testing components that execute commands
- Testing HTTP clients
- Testing database operations
- Testing file I/O operations
- Integration tests that need isolation

---

## Implementation Checklist

For MCP server test fixes:

- [x] Document problem (tests execute real commands)
- [x] Design solution (executor interface with dependency injection)
- [ ] Refactor executor.go to define ToolExecutor interface
- [ ] Create RealToolExecutor implementation
- [ ] Create MockToolExecutor implementation
- [ ] Modify handleToolsCall to accept executor parameter
- [ ] Add handleToolsCallWithExecutor test helper
- [ ] Update all integration tests to use mock executor
- [ ] Verify tests pass with mocks
- [ ] Measure coverage improvement

---

## Expected Benefits

1. **Reliability**: Tests don't depend on external binaries
2. **Speed**: Mock execution ~1000x faster than real command execution
3. **Coverage**: Can test error paths that are hard to trigger with real commands
4. **Isolation**: Tests don't affect filesystem or external systems
5. **Determinism**: Mock results are consistent across runs

---

## Metrics

### Before (Iteration 1)
- MCP integration tests: 3/5 failing
- Test execution time: ~140s total
- Coverage `handleToolsCall`: 67.3%
- Coverage `ExecuteTool`: 60.0%

### After (Iteration 2 Target)
- MCP integration tests: 5/5 passing
- Test execution time: ~120s total (faster without real commands)
- Coverage `handleToolsCall`: 80%+ (error paths tested)
- Coverage `ExecuteTool`: 75%+ (mock allows more scenarios)

---

## References

- Go interfaces: https://go.dev/tour/methods/9
- Dependency injection: https://dave.cheney.net/2016/08/20/solid-go-design
- Testing patterns: https://go.dev/blog/table-driven-tests
- Test pattern library: `test-pattern-library-iteration-1.md`

---

**Version**: 1.0
**Status**: Ready for Implementation
