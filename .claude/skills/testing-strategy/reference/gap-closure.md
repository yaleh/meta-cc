# Coverage Gap Closure Methodology

**Version**: 2.0
**Source**: Bootstrap-002 Test Strategy Development
**Last Updated**: 2025-10-18

This document describes the systematic approach to closing coverage gaps through prioritization, pattern selection, and continuous verification.

---

## Overview

Coverage gap closure is a systematic process for improving test coverage by:

1. Identifying functions with low/zero coverage
2. Prioritizing based on criticality
3. Selecting appropriate test patterns
4. Implementing tests efficiently
5. Verifying coverage improvements
6. Tracking progress

---

## Step-by-Step Gap Closure Process

### Step 1: Baseline Coverage Analysis

Generate current coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out > coverage-baseline.txt
```

**Extract key metrics**:
```bash
# Overall coverage
go tool cover -func=coverage.out | tail -1
# total: (statements) 72.1%

# Per-package coverage
go tool cover -func=coverage.out | grep "^github.com" | awk '{print $1, $NF}' | sort -t: -k1,1 -k2,2n
```

**Document baseline**:
```
Date: 2025-10-18
Total Coverage: 72.1%
Packages Below Target (<75%):
- internal/query: 65.3%
- internal/analyzer: 68.7%
- cmd/meta-cc: 55.2%
```

### Step 2: Identify Coverage Gaps

**Automated approach** (recommended):
```bash
./scripts/analyze-coverage-gaps.sh coverage.out --top 20 --threshold 70
```

**Manual approach**:
```bash
# Find zero-coverage functions
go tool cover -func=coverage.out | grep "0.0%" > zero-coverage.txt

# Find low-coverage functions (<60%)
go tool cover -func=coverage.out | awk '$NF+0 < 60.0' > low-coverage.txt

# Group by package
cat zero-coverage.txt | awk -F: '{print $1}' | sort | uniq -c
```

**Output example**:
```
Zero Coverage Functions (42 total):
  12 internal/query/filters.go
   8 internal/analyzer/patterns.go
   6 cmd/meta-cc/server.go
   ...

Low Coverage Functions (<60%, 23 total):
   7 internal/query/executor.go (45-55% coverage)
   5 internal/parser/jsonl.go (50-58% coverage)
   ...
```

### Step 3: Categorize and Prioritize

**Categorization criteria**:

| Category | Characteristics | Priority |
|----------|----------------|----------|
| **Error Handling** | Validation, error paths, edge cases | P1 |
| **Business Logic** | Core algorithms, data processing | P2 |
| **CLI Handlers** | Command execution, flag parsing | P2 |
| **Integration** | End-to-end flows, handlers | P3 |
| **Utilities** | Helpers, formatters | P3 |
| **Infrastructure** | Init, setup, configuration | P4 |

**Prioritization algorithm**:

```
For each function with <target coverage:
  1. Identify category (error-handling, business-logic, etc.)
  2. Assign priority (P1-P4)
  3. Estimate time (based on pattern + complexity)
  4. Estimate coverage impact (+0.1% to +0.3%)
  5. Calculate ROI = impact / time
  6. Sort by priority, then ROI
```

**Example prioritized list**:
```
P1 (Critical - Error Handling):
1. ValidateInput (0%) - Error Path + Table → 15 min, +0.25%
2. CheckFormat (25%) - Error Path → 12 min, +0.18%
3. ParseQuery (33%) - Error Path + Table → 15 min, +0.20%

P2 (High - Business Logic):
4. ProcessData (45%) - Table-Driven → 12 min, +0.20%
5. ApplyFilters (52%) - Table-Driven → 10 min, +0.15%

P2 (High - CLI):
6. ExecuteCommand (0%) - CLI Command → 13 min, +0.22%
7. ParseFlags (38%) - Global Flag → 11 min, +0.18%
```

### Step 4: Create Test Plan

For each testing session (target: 2-3 hours):

**Plan template**:
```
Session: Validation Error Paths
Date: 2025-10-18
Target: +5% package coverage, +1.5% total coverage
Time Budget: 2 hours (120 min)

Tests Planned:
1. ValidateInput - Error Path + Table (15 min) → +0.25%
2. CheckFormat - Error Path (12 min) → +0.18%
3. ParseQuery - Error Path + Table (15 min) → +0.20%
4. ProcessData - Table-Driven (12 min) → +0.20%
5. ApplyFilters - Table-Driven (10 min) → +0.15%
6. Buffer time: 56 min (for debugging, refactoring)

Expected Outcome:
- 5 new test functions
- Coverage: 72.1% → 73.1% (+1.0%)
```

### Step 5: Implement Tests

For each test in the plan:

**Workflow**:
```bash
# 1. Generate test scaffold
./scripts/generate-test.sh FunctionName --pattern PATTERN

# 2. Fill in test details
vim path/to/test_file.go

# 3. Run test
go test ./package/... -v -run TestFunctionName

# 4. Verify coverage improvement
go test -coverprofile=temp.out ./package/...
go tool cover -func=temp.out | grep FunctionName
```

**Example implementation**:
```go
// Generated scaffold
func TestValidateInput_ErrorCases(t *testing.T) {
    tests := []struct {
        name    string
        input   *Input  // TODO: Fill in
        wantErr bool
        errMsg  string
    }{
        {
            name:    "nil input",
            input:   nil,  // ← Fill in
            wantErr: true,
            errMsg:  "cannot be nil",  // ← Fill in
        },
        // TODO: Add more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := ValidateInput(tt.input)
            // Assertions...
        })
    }
}

// After filling TODOs (takes ~10-12 min per test)
```

### Step 6: Verify Coverage Impact

After implementing each test:

```bash
# Run new test
go test ./internal/validation/ -v -run TestValidateInput

# Generate coverage for package
go test -coverprofile=new_coverage.out ./internal/validation/

# Compare with baseline
echo "=== Before ==="
go tool cover -func=coverage.out | grep "internal/validation/"

echo "=== After ==="
go tool cover -func=new_coverage.out | grep "internal/validation/"

# Calculate improvement
echo "=== Change ==="
diff <(go tool cover -func=coverage.out | grep ValidateInput) \
     <(go tool cover -func=new_coverage.out | grep ValidateInput)
```

**Expected output**:
```
=== Before ===
internal/validation/validate.go:15:  ValidateInput  0.0%

=== After ===
internal/validation/validate.go:15:  ValidateInput  85.7%

=== Change ===
< internal/validation/validate.go:15:  ValidateInput  0.0%
> internal/validation/validate.go:15:  ValidateInput  85.7%
```

### Step 7: Track Progress

**Per-test tracking**:
```
Test: TestValidateInput_ErrorCases
Time: 12 min (estimated 15 min) → 20% faster
Pattern: Error Path + Table-Driven
Coverage Impact:
  - Function: 0% → 85.7% (+85.7%)
  - Package: 57.9% → 62.3% (+4.4%)
  - Total: 72.1% → 72.3% (+0.2%)
Issues: None
Notes: Table-driven very efficient for error cases
```

**Session summary**:
```
Session: Validation Error Paths
Date: 2025-10-18
Duration: 110 min (planned 120 min)

Tests Completed: 5/5
1. ValidateInput → +0.25% (actual: +0.2%)
2. CheckFormat → +0.18% (actual: +0.15%)
3. ParseQuery → +0.20% (actual: +0.22%)
4. ProcessData → +0.20% (actual: +0.18%)
5. ApplyFilters → +0.15% (actual: +0.12%)

Total Impact:
- Coverage: 72.1% → 72.97% (+0.87%)
- Tests added: 5 test functions, 18 test cases
- Time efficiency: 110 min / 5 tests = 22 min/test (vs 25 min/test ad-hoc)

Lessons:
- Error Path + Table-Driven pattern very effective
- Test generator saved ~40 min total
- Buffer time well-used for edge cases
```

### Step 8: Iterate

Repeat the process:

```bash
# Update baseline
mv new_coverage.out coverage.out

# Re-analyze gaps
./scripts/analyze-coverage-gaps.sh coverage.out --top 15

# Plan next session
# ...
```

---

## Coverage Improvement Patterns

### Pattern: Rapid Low-Hanging Fruit

**When**: Many zero-coverage functions, need quick wins

**Approach**:
1. Target P1/P2 zero-coverage functions
2. Use simple patterns (Unit, Table-Driven)
3. Skip complex infrastructure functions
4. Aim for 60-70% function coverage quickly

**Expected**: +5-10% total coverage in 3-4 hours

### Pattern: Systematic Package Closure

**When**: Specific package below target

**Approach**:
1. Focus on single package
2. Close all P1/P2 gaps in that package
3. Achieve 75-80% package coverage
4. Move to next package

**Expected**: +10-15% package coverage in 4-6 hours

### Pattern: Critical Path Hardening

**When**: Need high confidence in core functionality

**Approach**:
1. Identify critical business logic
2. Achieve 85-90% coverage on critical functions
3. Use Error Path + Integration patterns
4. Add edge case coverage

**Expected**: +0.5-1% total coverage per critical function

---

## Troubleshooting

### Issue: Coverage Not Increasing

**Symptoms**: Add tests, coverage stays same

**Diagnosis**:
```bash
# Check if function is actually being tested
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep FunctionName
```

**Causes**:
- Testing already-covered code (indirect coverage)
- Test not actually calling target function
- Function has unreachable code

**Solutions**:
- Focus on 0% coverage functions
- Verify test actually exercises target code path
- Use coverage visualization: `go tool cover -html=coverage.out`

### Issue: Coverage Decreasing

**Symptoms**: Coverage goes down after adding code

**Causes**:
- New code added without tests
- Refactoring exposed previously hidden code

**Solutions**:
- Always add tests for new code (TDD)
- Update coverage baseline after new features
- Set up pre-commit hooks to block coverage decreases

### Issue: Hard to Test Functions

**Symptoms**: Can't achieve good coverage on certain functions

**Causes**:
- Complex dependencies
- Infrastructure code (init, config)
- Difficult-to-mock external systems

**Solutions**:
- Use Dependency Injection (Pattern 6)
- Accept lower coverage for infrastructure (40-60%)
- Consider refactoring if truly untestable
- Extract testable business logic

### Issue: Slow Progress

**Symptoms**: Tests take much longer than estimated

**Causes**:
- Complex setup required
- Unclear function behavior
- Pattern mismatch

**Solutions**:
- Create test helpers (Pattern 5)
- Read function implementation first
- Adjust pattern selection
- Break into smaller tests

---

## Metrics and Goals

### Healthy Coverage Progression

**Typical trajectory** (starting from 60-70%):

```
Week 1: 62% → 68% (+6%)  - Low-hanging fruit
Week 2: 68% → 72% (+4%)  - Package-focused
Week 3: 72% → 75% (+3%)  - Critical paths
Week 4: 75% → 77% (+2%)  - Edge cases
Maintenance: 75-80%      - New code + decay prevention
```

**Time investment**:
- Initial ramp-up: 8-12 hours total
- Maintenance: 1-2 hours per week

### Coverage Targets by Project Phase

| Phase | Target | Focus |
|-------|--------|-------|
| **MVP** | 50-60% | Core happy paths |
| **Beta** | 65-75% | + Error handling |
| **Production** | 75-80% | + Edge cases, integration |
| **Mature** | 80-85% | + Documentation examples |

### When to Stop

**Diminishing returns** occur when:
- Coverage >80% total
- All P1/P2 functions >75%
- Remaining gaps are infrastructure/init code
- Time per 1% increase >3 hours

**Don't aim for 100%**:
- Infrastructure code hard to test (40-60% ok)
- Some code paths may be unreachable
- ROI drops significantly >85%

---

## Example: Complete Gap Closure Session

### Starting State

```
Package: internal/validation
Current Coverage: 57.9%
Target Coverage: 75%+
Gap: 17.1%

Zero Coverage Functions:
- ValidateInput (0%)
- CheckFormat (0%)
- ParseQuery (0%)

Low Coverage Functions:
- ValidateFilter (45%)
- NormalizeInput (52%)
```

### Plan

```
Session: Close validation coverage gaps
Time Budget: 2 hours
Target: 57.9% → 75%+ (+17.1%)

Tests:
1. ValidateInput (15 min) → +4.5%
2. CheckFormat (12 min) → +3.2%
3. ParseQuery (15 min) → +4.1%
4. ValidateFilter gaps (12 min) → +2.8%
5. NormalizeInput gaps (10 min) → +2.5%
Total: 64 min active, 56 min buffer
```

### Execution

```bash
# Test 1: ValidateInput
$ ./scripts/generate-test.sh ValidateInput --pattern error-path --scenarios 4
$ vim internal/validation/validate_test.go
# ... fill in TODOs (10 min) ...
$ go test ./internal/validation/ -run TestValidateInput -v
PASS (12 min actual)

# Test 2: CheckFormat
$ ./scripts/generate-test.sh CheckFormat --pattern error-path --scenarios 3
$ vim internal/validation/format_test.go
# ... fill in TODOs (8 min) ...
$ go test ./internal/validation/ -run TestCheckFormat -v
PASS (11 min actual)

# Test 3: ParseQuery
$ ./scripts/generate-test.sh ParseQuery --pattern table-driven --scenarios 5
$ vim internal/validation/query_test.go
# ... fill in TODOs (12 min) ...
$ go test ./internal/validation/ -run TestParseQuery -v
PASS (14 min actual)

# Test 4: ValidateFilter (add missing cases)
$ vim internal/validation/filter_test.go
# ... add 3 edge cases (8 min) ...
$ go test ./internal/validation/ -run TestValidateFilter -v
PASS (10 min actual)

# Test 5: NormalizeInput (add missing cases)
$ vim internal/validation/normalize_test.go
# ... add 2 edge cases (6 min) ...
$ go test ./internal/validation/ -run TestNormalizeInput -v
PASS (8 min actual)
```

### Result

```
Time: 55 min (vs 64 min estimated)
Coverage: 57.9% → 75.2% (+17.3%)
Tests Added: 5 functions, 17 test cases
Efficiency: 11 min per test (vs 15 min ad-hoc estimate)

SUCCESS: Target achieved (75%+)
```

---

**Source**: Bootstrap-002 Test Strategy Development
**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: Production-ready, validated through 4 iterations
