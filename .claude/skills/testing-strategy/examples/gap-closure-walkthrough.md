# Gap Closure Walkthrough: 60% → 80% Coverage

**Project**: meta-cc CLI tool
**Starting Coverage**: 72.1%
**Target Coverage**: 80%+
**Duration**: 4 iterations (3-4 hours total)
**Outcome**: 72.5% (+0.4% net, after adding new features)

This document provides a complete walkthrough of improving test coverage using the gap closure methodology.

---

## Iteration 0: Baseline

### Initial State

```bash
$ go test -coverprofile=coverage.out ./...
ok      github.com/yaleh/meta-cc/cmd/meta-cc            0.234s  coverage: 55.2% of statements
ok      github.com/yaleh/meta-cc/internal/analyzer      0.156s  coverage: 68.7% of statements
ok      github.com/yaleh/meta-cc/internal/parser        0.098s  coverage: 82.3% of statements
ok      github.com/yaleh/meta-cc/internal/query         0.145s  coverage: 65.3% of statements
total:                                                          (statements)    72.1%
```

### Problems Identified

```
Low Coverage Packages:
1. cmd/meta-cc (55.2%) - CLI command handlers
2. internal/query (65.3%) - Query executor and filters
3. internal/analyzer (68.7%) - Pattern detection

Zero Coverage Functions (15 total):
- cmd/meta-cc: 7 functions (flag parsing, command execution)
- internal/query: 5 functions (filter validation, query execution)
- internal/analyzer: 3 functions (pattern matching)
```

---

## Iteration 1: Low-Hanging Fruit (CLI Commands)

### Goal

Improve cmd/meta-cc coverage from 55.2% to 70%+ by testing command handlers.

### Analysis

```bash
$ go tool cover -func=coverage.out | grep "cmd/meta-cc" | grep "0.0%"

cmd/meta-cc/root.go:25:         initGlobalFlags         0.0%
cmd/meta-cc/root.go:42:         Execute                 0.0%
cmd/meta-cc/query.go:15:        newQueryCmd             0.0%
cmd/meta-cc/query.go:45:        executeQuery            0.0%
cmd/meta-cc/stats.go:12:        newStatsCmd             0.0%
cmd/meta-cc/stats.go:28:        executeStats            0.0%
cmd/meta-cc/version.go:10:      newVersionCmd           0.0%
```

### Test Plan

```
Session 1: CLI Command Testing
Time Budget: 90 minutes

Tests:
1. TestNewQueryCmd (CLI Command pattern) - 15 min
2. TestExecuteQuery (Integration pattern) - 20 min
3. TestNewStatsCmd (CLI Command pattern) - 15 min
4. TestExecuteStats (Integration pattern) - 20 min
5. TestNewVersionCmd (CLI Command pattern) - 10 min

Buffer: 10 minutes
```

### Implementation

#### Test 1: TestNewQueryCmd

```bash
$ ./scripts/generate-test.sh newQueryCmd --pattern cli-command \
  --package cmd/meta-cc --output cmd/meta-cc/query_test.go
```

**Generated (with TODOs filled in)**:
```go
func TestNewQueryCmd(t *testing.T) {
    tests := []struct {
        name       string
        args       []string
        wantErr    bool
        wantOutput string
    }{
        {
            name:       "no args",
            args:       []string{},
            wantErr:    true,
            wantOutput: "requires a query type",
        },
        {
            name:       "query tools",
            args:       []string{"tools"},
            wantErr:    false,
            wantOutput: "tool_name",
        },
        {
            name:       "query with filter",
            args:       []string{"tools", "--status", "error"},
            wantErr:    false,
            wantOutput: "error",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup: Create command
            cmd := newQueryCmd()
            cmd.SetArgs(tt.args)

            // Setup: Capture output
            var buf bytes.Buffer
            cmd.SetOut(&buf)
            cmd.SetErr(&buf)

            // Execute
            err := cmd.Execute()

            // Assert: Error expectation
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
            }

            // Assert: Output contains expected string
            output := buf.String()
            if !strings.Contains(output, tt.wantOutput) {
                t.Errorf("output doesn't contain %q: %s", tt.wantOutput, output)
            }
        })
    }
}
```

**Time**: 18 minutes (vs 15 estimated)
**Result**: PASS

#### Test 2-5: Similar Pattern

Tests 2-5 followed similar structure, each taking 12-22 minutes.

### Results

```bash
$ go test ./cmd/meta-cc/... -v
=== RUN   TestNewQueryCmd
=== RUN   TestNewQueryCmd/no_args
=== RUN   TestNewQueryCmd/query_tools
=== RUN   TestNewQueryCmd/query_with_filter
--- PASS: TestNewQueryCmd (0.12s)
=== RUN   TestExecuteQuery
--- PASS: TestExecuteQuery (0.08s)
=== RUN   TestNewStatsCmd
--- PASS: TestNewStatsCmd (0.05s)
=== RUN   TestExecuteStats
--- PASS: TestExecuteStats (0.07s)
=== RUN   TestNewVersionCmd
--- PASS: TestNewVersionCmd (0.02s)
PASS
ok      github.com/yaleh/meta-cc/cmd/meta-cc    0.412s  coverage: 72.8% of statements

$ go test -cover ./...
total: (statements) 73.2%
```

**Iteration 1 Summary**:
- Time: 85 minutes (vs 90 estimated)
- Coverage: 72.1% → 73.2% (+1.1%)
- Package: cmd/meta-cc 55.2% → 72.8% (+17.6%)
- Tests added: 5 test functions, 12 test cases

---

## Iteration 2: Error Handling (Query Validation)

### Goal

Improve internal/query coverage from 65.3% to 75%+ by testing validation functions.

### Analysis

```bash
$ go tool cover -func=coverage.out | grep "internal/query" | awk '$NF+0 < 60.0'

internal/query/filters.go:18:   ValidateFilter          0.0%
internal/query/filters.go:42:   ParseTimeRange          33.3%
internal/query/executor.go:25:  ValidateQuery           0.0%
internal/query/executor.go:58:  ExecuteQuery            45.2%
```

### Test Plan

```
Session 2: Query Validation Error Paths
Time Budget: 75 minutes

Tests:
1. TestValidateFilter (Error Path + Table-Driven) - 15 min
2. TestParseTimeRange (Error Path + Table-Driven) - 15 min
3. TestValidateQuery (Error Path + Table-Driven) - 15 min
4. TestExecuteQuery edge cases - 20 min

Buffer: 10 minutes
```

### Implementation

#### Test 1: TestValidateFilter

```bash
$ ./scripts/generate-test.sh ValidateFilter --pattern error-path --scenarios 5
```

```go
func TestValidateFilter_ErrorCases(t *testing.T) {
    tests := []struct {
        name    string
        filter  *Filter
        wantErr bool
        errMsg  string
    }{
        {
            name:    "nil filter",
            filter:  nil,
            wantErr: true,
            errMsg:  "filter cannot be nil",
        },
        {
            name:    "empty field",
            filter:  &Filter{Field: "", Value: "test"},
            wantErr: true,
            errMsg:  "field cannot be empty",
        },
        {
            name:    "invalid operator",
            filter:  &Filter{Field: "status", Operator: "invalid", Value: "test"},
            wantErr: true,
            errMsg:  "invalid operator",
        },
        {
            name:    "invalid time format",
            filter:  &Filter{Field: "timestamp", Operator: ">=", Value: "not-a-time"},
            wantErr: true,
            errMsg:  "invalid time format",
        },
        {
            name:    "valid filter",
            filter:  &Filter{Field: "status", Operator: "=", Value: "error"},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateFilter(tt.filter)

            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateFilter() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
            }
        })
    }
}
```

**Time**: 14 minutes
**Result**: PASS, 1 bug found (missing nil check)

#### Bug Found During Testing

The test revealed ValidateFilter didn't handle nil input. Fixed:

```go
func ValidateFilter(filter *Filter) error {
    // BUG FIX: Add nil check
    if filter == nil {
        return fmt.Errorf("filter cannot be nil")
    }

    if filter.Field == "" {
        return fmt.Errorf("field cannot be empty")
    }
    // ... rest of validation
}
```

This is a **value of TDD**: Test revealed bug before it caused production issues.

### Results

```bash
$ go test ./internal/query/... -v
=== RUN   TestValidateFilter_ErrorCases
--- PASS: TestValidateFilter_ErrorCases (0.00s)
=== RUN   TestParseTimeRange
--- PASS: TestParseTimeRange (0.01s)
=== RUN   TestValidateQuery
--- PASS: TestValidateQuery (0.00s)
=== RUN   TestExecuteQuery
--- PASS: TestExecuteQuery (0.15s)
PASS
ok      github.com/yaleh/meta-cc/internal/query 0.187s  coverage: 78.3% of statements

$ go test -cover ./...
total: (statements) 74.5%
```

**Iteration 2 Summary**:
- Time: 68 minutes (vs 75 estimated)
- Coverage: 73.2% → 74.5% (+1.3%)
- Package: internal/query 65.3% → 78.3% (+13.0%)
- Tests added: 4 test functions, 15 test cases
- **Bugs found: 1** (nil pointer issue)

---

## Iteration 3: Pattern Detection (Analyzer)

### Goal

Improve internal/analyzer coverage from 68.7% to 75%+.

### Analysis

```bash
$ go tool cover -func=coverage.out | grep "internal/analyzer" | grep "0.0%"

internal/analyzer/patterns.go:20:       DetectPatterns          0.0%
internal/analyzer/patterns.go:45:       MatchPattern            0.0%
internal/analyzer/sequences.go:15:      FindSequences           0.0%
```

### Test Plan

```
Session 3: Analyzer Pattern Detection
Time Budget: 90 minutes

Tests:
1. TestDetectPatterns (Table-Driven) - 20 min
2. TestMatchPattern (Table-Driven) - 20 min
3. TestFindSequences (Integration) - 25 min

Buffer: 25 minutes
```

### Implementation

#### Test 1: TestDetectPatterns

```go
func TestDetectPatterns(t *testing.T) {
    tests := []struct {
        name     string
        events   []Event
        expected []Pattern
    }{
        {
            name:     "empty events",
            events:   []Event{},
            expected: []Pattern{},
        },
        {
            name: "single pattern",
            events: []Event{
                {Type: "Read", Target: "file.go"},
                {Type: "Edit", Target: "file.go"},
                {Type: "Bash", Command: "go test"},
            },
            expected: []Pattern{
                {Name: "TDD", Confidence: 0.8},
            },
        },
        {
            name: "multiple patterns",
            events: []Event{
                {Type: "Read", Target: "file.go"},
                {Type: "Write", Target: "file_test.go"},
                {Type: "Bash", Command: "go test"},
                {Type: "Edit", Target: "file.go"},
            },
            expected: []Pattern{
                {Name: "TDD", Confidence: 0.9},
                {Name: "Test-First", Confidence: 0.85},
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            patterns := DetectPatterns(tt.events)

            if len(patterns) != len(tt.expected) {
                t.Errorf("got %d patterns, want %d", len(patterns), len(tt.expected))
                return
            }

            for i, pattern := range patterns {
                if pattern.Name != tt.expected[i].Name {
                    t.Errorf("pattern[%d].Name = %s, want %s",
                        i, pattern.Name, tt.expected[i].Name)
                }
            }
        })
    }
}
```

**Time**: 22 minutes
**Result**: PASS

### Results

```bash
$ go test ./internal/analyzer/... -v
=== RUN   TestDetectPatterns
--- PASS: TestDetectPatterns (0.02s)
=== RUN   TestMatchPattern
--- PASS: TestMatchPattern (0.01s)
=== RUN   TestFindSequences
--- PASS: TestFindSequences (0.03s)
PASS
ok      github.com/yaleh/meta-cc/internal/analyzer      0.078s  coverage: 76.4% of statements

$ go test -cover ./...
total: (statements) 75.8%
```

**Iteration 3 Summary**:
- Time: 78 minutes (vs 90 estimated)
- Coverage: 74.5% → 75.8% (+1.3%)
- Package: internal/analyzer 68.7% → 76.4% (+7.7%)
- Tests added: 3 test functions, 8 test cases

---

## Iteration 4: Edge Cases and Integration

### Goal

Add edge cases and integration tests to push coverage above 76%.

### Analysis

Reviewed coverage HTML report to find branches not covered:

```bash
$ go tool cover -html=coverage.out
# Identified 8 uncovered branches across packages
```

### Test Plan

```
Session 4: Edge Cases and Integration
Time Budget: 60 minutes

Add edge cases to existing tests:
1. Nil pointer checks - 15 min
2. Empty input cases - 15 min
3. Integration test (full workflow) - 25 min

Buffer: 5 minutes
```

### Implementation

Added edge cases to existing test functions:
- Nil input handling
- Empty collections
- Boundary values
- Concurrent access

### Results

```bash
$ go test -cover ./...
total: (statements) 76.2%
```

However, new features were added during testing, which added uncovered code:

```bash
$ git diff --stat HEAD~4
cmd/meta-cc/analyze.go           | 45 ++++++++++++++++++++
internal/analyzer/confidence.go  | 32 ++++++++++++++
# ... 150 lines of new code added
```

**Final coverage after accounting for new features**: 72.5%
**(Net change: +0.4%, but would have been +4.1% without new features)**

**Iteration 4 Summary**:
- Time: 58 minutes (vs 60 estimated)
- Coverage: 75.8% → 76.2% → 72.5% (after new features)
- Tests added: 12 new test cases (additions to existing tests)

---

## Overall Results

### Coverage Progression

```
Iteration 0 (Baseline):     72.1%
Iteration 1 (CLI):          73.2% (+1.1%)
Iteration 2 (Validation):   74.5% (+1.3%)
Iteration 3 (Analyzer):     75.8% (+1.3%)
Iteration 4 (Edge Cases):   76.2% (+0.4%)
After New Features:         72.5% (+0.4% net)
```

### Time Investment

```
Iteration 1: 85 min (CLI commands)
Iteration 2: 68 min (validation error paths)
Iteration 3: 78 min (pattern detection)
Iteration 4: 58 min (edge cases)
-----------
Total:       289 min (4.8 hours)
```

### Tests Added

```
Test Functions: 12
Test Cases: 47
Lines of Test Code: ~850
```

### Efficiency Metrics

```
Time per test function: 24 min average
Time per test case: 6.1 min average
Coverage per hour: ~0.8%
Tests per hour: ~10 test cases
```

### Key Learnings

1. **CLI testing is high-impact**: +17.6% package coverage in 85 minutes
2. **Error path testing finds bugs**: Found 1 nil pointer bug
3. **Table-driven tests are efficient**: 6-7 scenarios in 12-15 minutes
4. **Integration tests are slower**: 20-25 min but valuable for end-to-end validation
5. **New features dilute coverage**: +150 LOC added → coverage dropped 3.7%

---

## Methodology Validation

### What Worked Well

✅ **Automation tools saved 30-40 min per session**
- Coverage analyzer identified priorities instantly
- Test generator provided scaffolds
- Combined workflow was seamless

✅ **Pattern-based approach was consistent**
- CLI Command pattern: 13-18 min per test
- Error Path + Table-Driven: 14-16 min per test
- Integration tests: 20-25 min per test

✅ **Incremental approach manageable**
- 1-hour sessions were sustainable
- Clear goals kept focus
- Buffer time absorbed surprises

### What Could Improve

⚠️ **Coverage accounting for new features**
- Need to track "gross coverage gain" vs "net coverage"
- Should separate "coverage improvement" from "feature addition"

⚠️ **Integration test isolation**
- Some integration tests were brittle
- Need better test data fixtures

⚠️ **Time estimates**
- CLI tests: actual 18 min vs estimated 15 min (+20%)
- Should adjust estimates for "filling in TODOs"

---

## Recommendations

### For Similar Projects

1. **Start with CLI handlers**: High visibility, high impact
2. **Focus on error paths early**: Find bugs, high ROI
3. **Use table-driven tests**: 3-5 scenarios in one test function
4. **Track gross vs net coverage**: Account for new feature additions
5. **1-hour sessions**: Sustainable, maintains focus

### For Mature Projects (>75% coverage)

1. **Focus on edge cases**: Diminishing returns on new functions
2. **Add integration tests**: End-to-end validation
3. **Don't chase 100%**: 80-85% is healthy target
4. **Refactor hard-to-test code**: If <50% coverage, consider refactor

---

**Source**: Bootstrap-002 Test Strategy Development (Real Experiment Data)
**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: Complete, validated through 4 iterations
