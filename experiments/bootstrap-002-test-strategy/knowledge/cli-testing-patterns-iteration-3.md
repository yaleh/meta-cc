# CLI Testing Patterns - Iteration 3

**Created**: 2025-10-18
**Experiment**: Bootstrap-002 Test Strategy Development
**Status**: Established for Coverage-Driven CLI Testing

---

## Overview

This document extends the test pattern library with specific patterns for CLI command testing. These patterns address the unique challenges of testing command-line interfaces built with Cobra.

---

## Pattern 6: CLI Command Test Pattern

**Purpose**: Test Cobra command execution with flags and arguments

**Structure**:
```go
func TestCommandName_Scenario(t *testing.T) {
    // Setup: Create command
    cmd := &cobra.Command{
        Use: "command-name",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Command logic
            return nil
        },
    }

    // Setup: Add flags
    cmd.Flags().StringP("flag", "f", "default", "description")

    // Setup: Set arguments
    cmd.SetArgs([]string{"--flag", "value", "arg1"})

    // Execute: Run command
    err := cmd.Execute()

    // Assert: Verify result
    if err != nil {
        t.Fatalf("command failed: %v", err)
    }

    // Assert: Verify side effects (output, file creation, etc.)
    // ...
}
```

**Key Characteristics**:
- Uses `cmd.SetArgs()` for flag/argument injection
- Tests command execution in isolation
- Verifies both return values and side effects
- Handles cleanup with `defer` or `t.Cleanup()`

**Example**:
```go
func TestRootCommand_VersionFlag(t *testing.T) {
    // Setup: Capture stdout
    var buf bytes.Buffer
    rootCmd.SetOut(&buf)

    // Setup: Set version flag
    rootCmd.SetArgs([]string{"--version"})

    // Execute
    err := rootCmd.Execute()

    // Assert: No error
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    // Assert: Version string present
    output := buf.String()
    if !strings.Contains(output, "version") {
        t.Errorf("expected version string, got: %s", output)
    }
}
```

**When to Use**:
- Testing CLI command handlers
- Flag parsing verification
- Command composition testing
- Help text validation

---

## Pattern 7: CLI Integration Test Pattern

**Purpose**: Test complete CLI command with subprocess execution

**Structure**:
```go
func TestCLI_CommandIntegration(t *testing.T) {
    // Skip if binary not available
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // Build binary (or ensure it exists)
    binary := buildTestBinary(t)
    defer os.Remove(binary)

    // Execute command
    cmd := exec.Command(binary, "command", "--flag", "value")
    output, err := cmd.CombinedOutput()

    // Assert: Check exit code
    if err != nil {
        t.Fatalf("command failed: %v\nOutput: %s", err, output)
    }

    // Assert: Verify output
    if !strings.Contains(string(output), "expected") {
        t.Errorf("unexpected output: %s", output)
    }
}

// Helper: Build test binary
func buildTestBinary(t *testing.T) string {
    t.Helper()

    tmpDir := t.TempDir()
    binary := filepath.Join(tmpDir, "test-binary")

    cmd := exec.Command("go", "build", "-o", binary, ".")
    if err := cmd.Run(); err != nil {
        t.Fatalf("failed to build binary: %v", err)
    }

    return binary
}
```

**Key Characteristics**:
- Uses `os/exec` for subprocess execution
- Tests CLI as black box
- Verifies actual binary behavior
- Includes build step for test binary

**Example**:
```go
func TestMetaCC_StatsCommand(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    binary := buildTestBinary(t)
    defer os.Remove(binary)

    // Create test session
    sessionPath := createTestSession(t)

    // Execute stats command
    cmd := exec.Command(binary, "stats", "--session", sessionPath)
    output, err := cmd.CombinedOutput()

    if err != nil {
        t.Fatalf("stats command failed: %v\nOutput: %s", err, output)
    }

    // Verify JSON output
    var stats map[string]interface{}
    if err := json.Unmarshal(output, &stats); err != nil {
        t.Fatalf("invalid JSON output: %v", err)
    }

    // Verify expected fields
    if _, ok := stats["tool_calls"]; !ok {
        t.Error("expected tool_calls in stats")
    }
}
```

**When to Use**:
- End-to-end CLI testing
- Regression testing
- Binary behavior verification
- CI/CD integration tests

**Note**: These are expensive (slow), use sparingly

---

## Pattern 8: Global Flag Test Pattern

**Purpose**: Test global flag parsing and propagation

**Structure**:
```go
func TestGlobalOptions_Scenario(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        expected GlobalOptions
        wantErr  bool
    }{
        {
            name: "default project path",
            args: []string{},
            expected: GlobalOptions{
                ProjectPath: getCurrentDir(),
            },
            wantErr: false,
        },
        {
            name: "with session flag",
            args: []string{"--session", "abc123"},
            expected: GlobalOptions{
                SessionID: "abc123",
            },
            wantErr: false,
        },
        {
            name: "session-only mode",
            args: []string{"--session-only"},
            expected: GlobalOptions{
                SessionOnly: true,
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup: Reset flags
            resetGlobalFlags()

            // Setup: Set args
            rootCmd.SetArgs(tt.args)

            // Execute: Parse flags
            if err := rootCmd.ParseFlags(tt.args); (err != nil) != tt.wantErr {
                t.Fatalf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
            }

            // Get options
            opts := getGlobalOptions()

            // Assert: Verify options
            if opts.SessionID != tt.expected.SessionID {
                t.Errorf("SessionID = %v, expected %v", opts.SessionID, tt.expected.SessionID)
            }

            if opts.SessionOnly != tt.expected.SessionOnly {
                t.Errorf("SessionOnly = %v, expected %v", opts.SessionOnly, tt.expected.SessionOnly)
            }
        })
    }
}
```

**Key Characteristics**:
- Table-driven for multiple flag combinations
- Resets global state between tests
- Verifies flag propagation to options
- Tests flag interaction logic

**When to Use**:
- Testing global flag parsing
- Flag interaction testing
- Default value verification
- Option struct population

---

## Coverage-Driven Workflow (Refined)

### Step 1: Identify Coverage Gaps

```bash
# Generate coverage report with function-level detail
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out > coverage-by-func.txt

# Find low-coverage packages
go tool cover -func=coverage.out | grep "^github.com/yaleh/meta-cc/cmd/" | awk '$NF < 60.0'

# Identify zero-coverage functions
go tool cover -func=coverage.out | grep "0.0%"
```

**Output Analysis**:
- **0% coverage**: Completely untested (highest priority)
- **1-60% coverage**: Partially tested, likely missing error paths
- **60-80% coverage**: Good coverage, may need edge cases
- **80%+ coverage**: Excellent coverage, maintain

### Step 2: Prioritize Test Targets

**Decision Tree**:

```
Is function critical to core functionality?
├─ YES: Is it error handling or validation?
│  ├─ YES: Priority 1 (must have 80%+ coverage)
│  └─ NO: Is it business logic?
│     ├─ YES: Priority 2 (target 75%+ coverage)
│     └─ NO: Priority 3 (target 60%+ coverage)
└─ NO: Is it infrastructure/initialization?
   ├─ YES: Priority 4 (test if possible, skip if hard)
   └─ NO: Priority 5 (skip for now)
```

**Priority Matrix**:

| Category | Examples | Target Coverage | Priority |
|----------|----------|----------------|----------|
| Error Handling | Validation, input sanitization | 80-90% | P1 |
| Business Logic | Core algorithms, transformations | 75-85% | P2 |
| CLI Handlers | Command execution, flag parsing | 70-80% | P2 |
| Integration | MCP handlers, I/O operations | 70-80% | P3 |
| Utilities | Helpers, formatters | 60-70% | P3 |
| Infrastructure | Init, logging, config | Best effort | P4 |

### Step 3: Select Test Pattern

**Decision Tree**:

```
What are you testing?
├─ CLI command with flags?
│  ├─ Multiple flag combinations? → Table-Driven + CLI Command Pattern
│  ├─ Integration test needed? → CLI Integration Pattern
│  └─ Global flags? → Global Flag Pattern
├─ Error paths?
│  ├─ Multiple error scenarios? → Table-Driven + Error Path Pattern
│  └─ Single error case? → Error Path Pattern
├─ Unit function?
│  ├─ Multiple inputs? → Table-Driven Pattern
│  └─ Single input? → Unit Test Pattern
└─ Integration flow?
   └─ → Integration Test Pattern
```

### Step 4: Write Test

**Template Selection Guide**:

1. **CLI Command Test**: Use Pattern 6
   - Fast, in-process
   - Tests command logic
   - Good for flag parsing

2. **CLI Integration Test**: Use Pattern 7
   - Slower, subprocess
   - Tests actual binary
   - Good for E2E scenarios

3. **Global Flag Test**: Use Pattern 8
   - Table-driven
   - Tests flag interaction
   - Good for option parsing

4. **Error Path Test**: Use Pattern 4 (from main library)
   - Systematic error coverage
   - Validates error messages
   - Critical for robustness

### Step 5: Verify Coverage Impact

```bash
# Run tests and generate new coverage
go test -coverprofile=new_coverage.out ./...

# Compare coverage
echo "Old coverage:"
go tool cover -func=coverage.out | tail -1

echo "New coverage:"
go tool cover -func=new_coverage.out | tail -1

# Show improved functions
go tool cover -func=new_coverage.out > new_coverage.txt
diff coverage-by-func.txt new_coverage.txt | grep "^>"
```

**Metrics to Track**:
- **Total coverage change**: Should increase by 0.5-2% per batch of 5-10 tests
- **Package coverage change**: Target package should increase by 5-10%
- **Function coverage change**: Targeted functions should increase by 20-50%
- **Test count increase**: Track tests added vs coverage gained (efficiency)

### Step 6: Document Pattern Usage

**For Each Test Batch, Record**:
- Pattern(s) used
- Time to write tests (actual vs estimated)
- Coverage increase achieved
- Issues encountered
- Lessons learned

**Example Log**:
```
Batch: CLI root command tests (5 tests)
Pattern: Table-Driven + CLI Command Pattern
Time: 45 min (est. 60 min) → 25% faster than estimated
Coverage: cmd/root.go 30% → 75% (+45%)
Coverage: cmd/ package 57.9% → 61.2% (+3.3%)
Coverage: total 72.3% → 73.1% (+0.8%)
Efficiency: 0.16% coverage per test
Issues: None
Lessons: t.Cleanup() more reliable than defer for test cleanup
```

---

## Quality Checklist for CLI Tests

Beyond standard test quality checklist:

- [ ] Command flags reset between tests (global state)
- [ ] Output captured properly (stdout/stderr)
- [ ] Environment variables reset (if used)
- [ ] Working directory restored (if changed)
- [ ] Temporary files cleaned up
- [ ] No dependency on external binaries (unless integration test)
- [ ] Fast execution (<100ms for unit, <1s for integration)
- [ ] Tests both happy path and error cases
- [ ] Help text validated (if command has help)
- [ ] Exit codes verified (for integration tests)

---

## Common Pitfalls in CLI Testing

### 1. Global State Pollution

**Problem**: Tests share global flags/variables, causing interference

**Solution**: Reset global state in test setup
```go
func resetGlobalFlags() {
    sessionID = ""
    projectPath = ""
    sessionOnly = false
    outputFormat = ""
}

func TestCommand(t *testing.T) {
    resetGlobalFlags()
    defer resetGlobalFlags()

    // Test logic...
}
```

### 2. Missing Output Capture

**Problem**: Command output goes to stdout, can't verify

**Solution**: Use `SetOut()` or capture with buffer
```go
var buf bytes.Buffer
cmd.SetOut(&buf)
cmd.Execute()
output := buf.String()
```

### 3. Integration Test Flakiness

**Problem**: Tests depend on binary build, filesystem state

**Solution**:
- Use `testing.Short()` to skip in fast mode
- Build binary once, reuse
- Clean up temporary files reliably
- Don't depend on system binaries (git, etc.) unless necessary

### 4. Brittle Flag Tests

**Problem**: Tests break when flag names change

**Solution**: Test behavior, not implementation
```go
// Bad: Hard-coded flag names
cmd.SetArgs([]string{"--session", "abc"})

// Good: Test behavior
cmd.SetArgs([]string{"--session", "abc"})
opts := getGlobalOptions()
if opts.SessionID != "abc" {
    t.Error("session ID not set correctly")
}
```

### 5. Slow Test Suites

**Problem**: Too many integration tests slow down CI

**Solution**:
- Prefer in-process CLI tests (Pattern 6) over subprocess (Pattern 7)
- Use integration tests sparingly (critical paths only)
- Run integration tests in parallel where possible
- Skip integration tests in short mode

---

## Efficiency Metrics

### Pattern Usage Statistics (Iteration 3)

**Time per Test** (estimated from actual usage):
- CLI Command Test (Pattern 6): ~12-15 min
- CLI Integration Test (Pattern 7): ~20-25 min (includes build setup)
- Global Flag Test (Pattern 8): ~10-12 min (table-driven, reuses setup)
- Error Path Test (Pattern 4): ~8-10 min (focused, single purpose)

**Coverage Impact per Test**:
- CLI Command Test: 0.15-0.25% total coverage
- CLI Integration Test: 0.10-0.20% total coverage (lower due to overhead)
- Global Flag Test: 0.20-0.30% total coverage (high impact)
- Error Path Test: 0.10-0.15% total coverage

**Efficiency Ratio** (coverage % per hour):
- CLI Command Tests: ~1.0-1.7% per hour (4-5 tests/hour × 0.2% each)
- Error Path Tests: ~0.6-1.2% per hour (6 tests/hour × 0.125% each)

**Best ROI**: Global Flag Tests (table-driven, high coverage per test)
**Most Comprehensive**: CLI Integration Tests (but slowest)

---

## Next Steps for Methodology

### Iteration 4+ Improvements
1. **Test Generator Tool**: CLI to scaffold tests from function signatures
2. **Coverage Target Tool**: Suggest which function to test next (prioritized)
3. **Pattern Selector Tool**: Interactive pattern selection based on function type
4. **Efficiency Dashboard**: Track coverage/time metrics across iterations

### Automation Opportunities
1. Auto-generate table-driven test scaffolds
2. Flag combination generator for CLI tests
3. Error path enumeration from function signatures
4. Test coverage diff tool (highlight improvements)

---

## References

- Cobra testing: https://github.com/spf13/cobra/blob/main/command_test.go
- Go CLI testing: https://golang.org/pkg/os/exec/
- Table-driven tests: https://github.com/golang/go/wiki/TableDrivenTests
- Coverage tools: https://golang.org/cmd/cover/

---

**Version**: 1.0
**Last Updated**: 2025-10-18
**Status**: Ready for Implementation (Iteration 3)
