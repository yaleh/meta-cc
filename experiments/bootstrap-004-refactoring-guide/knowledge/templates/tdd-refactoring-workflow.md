# TDD Refactoring Workflow

**Purpose**: Enforce test-driven discipline during refactoring to ensure behavior preservation and quality

**When to Use**: During ALL refactoring work

**Origin**: Iteration 1 - Problem E1 (No TDD Enforcement)

---

## TDD Principle for Refactoring

**Red-Green-Refactor Cycle** (adapted for refactoring existing code):

1. **Green** (Baseline): Ensure existing tests pass
2. **Red** (Add Tests): Write tests for uncovered behavior (tests should pass immediately since code exists)
3. **Refactor**: Restructure code while maintaining green tests
4. **Green** (Verify): Confirm all tests still pass after refactoring

**Key Difference from New Development TDD**:
- **New Development**: Write failing test → Make it pass → Refactor
- **Refactoring**: Ensure passing tests → Add missing tests (passing) → Refactor → Keep tests passing

---

## Workflow Steps

### Phase 1: Baseline Green (Ensure Safety Net)

**Goal**: Verify existing tests provide safety net for refactoring

#### Step 1: Run Existing Tests

```bash
go test -v ./internal/query/... > tests-baseline.txt
```

**Checklist**:
- [ ] All existing tests pass: YES / NO
- [ ] Test count: ___ tests
- [ ] Duration: ___s
- [ ] If any fail: FIX BEFORE PROCEEDING

#### Step 2: Check Coverage

```bash
go test -cover ./internal/query/...
go test -coverprofile=coverage.out ./internal/query/...
go tool cover -html=coverage.out -o coverage.html
```

**Checklist**:
- [ ] Overall coverage: ___%
- [ ] Target function coverage: ___%
- [ ] Uncovered lines identified: YES / NO
- [ ] Coverage file: `coverage.html` (review in browser)

#### Step 3: Identify Coverage Gaps

**Review `coverage.html` and identify**:
- [ ] Uncovered branches: _______________
- [ ] Uncovered error paths: _______________
- [ ] Uncovered edge cases: _______________
- [ ] Missing edge case examples:
  - Empty inputs: ___ (e.g., empty slice, nil, zero)
  - Boundary conditions: ___ (e.g., single element, max value)
  - Error conditions: ___ (e.g., invalid input, out of range)

**Decision Point**:
- If coverage ≥95% on target code: Proceed to Phase 2 (Refactor)
- If coverage <95%: Proceed to Phase 1b (Write Missing Tests)

---

### Phase 1b: Write Missing Tests (Red → Immediate Green)

**Goal**: Add tests for uncovered code paths BEFORE refactoring

#### For Each Coverage Gap:

**1. Write Characterization Test** (documents current behavior):

```go
func TestCalculateSequenceTimeSpan_<EdgeCase>(t *testing.T) {
    // Setup: Create input that triggers uncovered path
    // ...

    // Execute: Call function
    result := calculateSequenceTimeSpan(occurrences, entries, toolCalls)

    // Verify: Document current behavior (even if it's wrong)
    assert.Equal(t, <expected>, result, "current behavior")
}
```

**Test Naming Convention**:
- `Test<FunctionName>_<EdgeCase>` (e.g., `TestCalculateTimeSpan_EmptyOccurrences`)
- `Test<FunctionName>_<Scenario>` (e.g., `TestCalculateTimeSpan_SingleOccurrence`)

**2. Verify Test Passes** (should pass immediately since code exists):

```bash
go test -v -run Test<FunctionName>_<EdgeCase> ./...
```

**Checklist**:
- [ ] Test written: `Test<FunctionName>_<EdgeCase>`
- [ ] Test passes immediately: YES / NO
- [ ] If NO: Bug in test or unexpected current behavior → Fix test
- [ ] Coverage increased: __% → ___%

**3. Commit Test**:

```bash
git add <test_file>
git commit -m "test: add <edge-case> test for <function>"
```

**Repeat for all coverage gaps until target coverage ≥95%**

#### Coverage Target

- [ ] **Overall coverage**: ≥85% (project minimum)
- [ ] **Target function coverage**: ≥95% (refactoring requirement)
- [ ] **New test coverage**: ≥100% (all new tests pass)

**Checkpoint**: Before proceeding to refactoring:
- [ ] All tests pass: PASS
- [ ] Target function coverage: ≥95%
- [ ] Coverage gaps documented if <95%: _______________

---

### Phase 2: Refactor (Maintain Green)

**Goal**: Restructure code while keeping all tests passing

#### For Each Refactoring Step:

**1. Plan Single Refactoring Transformation**:

- [ ] Transformation type: _______________ (Extract Method, Inline, Rename, etc.)
- [ ] Target code: _______________ (function, lines, scope)
- [ ] Expected outcome: _______________ (complexity reduction, clarity, etc.)
- [ ] Estimated time: ___ minutes

**2. Make Minimal Change**:

**Examples**:
- Extract Method: Move lines X-Y to new function `<name>`
- Simplify Conditional: Replace nested if with guard clause
- Rename: Change `<oldName>` to `<newName>`

**Checklist**:
- [ ] Single, focused change: YES / NO
- [ ] No behavioral changes: Only structural / organizational
- [ ] Files modified: _______________
- [ ] Lines changed: ~___

**3. Run Tests Immediately**:

```bash
go test -v ./internal/query/... | tee test-results-step-N.txt
```

**Checklist**:
- [ ] All tests pass: PASS / FAIL
- [ ] Duration: ___s (should be quick, <10s)
- [ ] If FAIL: **ROLLBACK IMMEDIATELY**

**4. Verify Coverage Maintained**:

```bash
go test -cover ./internal/query/...
```

**Checklist**:
- [ ] Coverage: Before __% → After ___%
- [ ] Change: +/- ___%
- [ ] If decreased >1%: Investigate (might need to update tests)
- [ ] If decreased >5%: **ROLLBACK**

**5. Verify Complexity**:

```bash
gocyclo -over 10 internal/query/<target-file>.go
```

**Checklist**:
- [ ] Target function complexity: ___
- [ ] Change from previous: +/- ___
- [ ] If increased: Not a valid refactoring step → ROLLBACK

**6. Commit Incremental Progress**:

```bash
git add .
git commit -m "refactor(<file>): <pattern> - <what changed>"
```

**Example Commit Messages**:
- `refactor(sequences): extract collectTimestamps helper`
- `refactor(sequences): simplify min/max calculation`
- `refactor(sequences): rename ts to timestamp for clarity`

**Checklist**:
- [ ] Commit hash: _______________
- [ ] Message follows convention: YES / NO
- [ ] Commit is small, focused: YES / NO

**Repeat refactoring steps until refactoring complete or target achieved**

---

### Phase 3: Final Verification (Confirm Green)

**Goal**: Comprehensive verification that refactoring succeeded

#### 1. Run Full Test Suite

```bash
go test -v ./... | tee test-results-final.txt
```

**Checklist**:
- [ ] All tests pass: PASS / FAIL
- [ ] Test count: ___ (should match baseline or increase)
- [ ] Duration: ___s
- [ ] No flaky tests: All consistent

#### 2. Verify Coverage Improved or Maintained

```bash
go test -cover ./internal/query/...
go test -coverprofile=coverage-final.out ./internal/query/...
go tool cover -func=coverage-final.out | grep total
```

**Checklist**:
- [ ] Baseline coverage: ___%
- [ ] Final coverage: ___%
- [ ] Change: +___%
- [ ] Target met (≥85% overall, ≥95% refactored code): YES / NO

#### 3. Compare Baseline and Final Metrics

| Metric | Baseline | Final | Change | Target Met |
|--------|----------|-------|--------|------------|
| **Complexity** | ___ | ___ | ___% | YES / NO |
| **Coverage** | ___% | ___% | +___% | YES / NO |
| **Test count** | ___ | ___ | +___ | N/A |
| **Test duration** | ___s | ___s | ___s | N/A |

**Checklist**:
- [ ] All targets met: YES / NO
- [ ] If NO: Document gaps and plan next iteration

#### 4. Update Documentation

```bash
# Add/update GoDoc comments for refactored code
# Example:
// calculateSequenceTimeSpan calculates the time span in minutes between
// the first and last occurrence of a sequence pattern across turns.
// Returns 0 if no valid timestamps found.
```

**Checklist**:
- [ ] GoDoc added/updated: YES / NO
- [ ] Public functions documented: ___ / ___ (100%)
- [ ] Parameter descriptions clear: YES / NO
- [ ] Return value documented: YES / NO

---

## TDD Metrics (Track Over Time)

**Refactoring Session**: ___ (e.g., calculateSequenceTimeSpan - 2025-10-19)

| Metric | Value |
|--------|-------|
| **Baseline coverage** | ___% |
| **Final coverage** | ___% |
| **Coverage improvement** | +___% |
| **Tests added** | ___ |
| **Test failures during refactoring** | ___ |
| **Rollbacks due to test failures** | ___ |
| **Time spent writing tests** | ___ min |
| **Time spent refactoring** | ___ min |
| **Test writing : Refactoring ratio** | ___:1 |

**TDD Discipline Score**: (Tests passing after each step) / (Total steps) × 100% = ___%

**Target**: 100% TDD discipline (tests pass after EVERY step)

---

## Common TDD Refactoring Patterns

### Pattern 1: Extract Method with Tests

**Scenario**: Function too complex, need to extract helper

**Steps**:
1. ✅ Ensure tests pass
2. ✅ Write test for behavior to be extracted (if not covered)
3. ✅ Extract method
4. ✅ Tests still pass
5. ✅ Write direct test for new extracted method
6. ✅ Tests pass
7. ✅ Commit

**Example**:
```go
// Before:
func calculate() {
    // ... 20 lines of timestamp collection
    // ... 15 lines of min/max finding
}

// After:
func calculate() {
    timestamps := collectTimestamps()
    return findMinMax(timestamps)
}

func collectTimestamps() []int64 { /* extracted */ }
func findMinMax([]int64) int { /* extracted */ }
```

**Tests**:
- Existing: `TestCalculate` (still passes)
- New: `TestCollectTimestamps` (covers extracted logic)
- New: `TestFindMinMax` (covers min/max logic)

---

### Pattern 2: Simplify Conditionals with Tests

**Scenario**: Nested conditionals hard to read, need to simplify

**Steps**:
1. ✅ Ensure tests pass (covering all branches)
2. ✅ If branches uncovered: Add tests for all paths
3. ✅ Simplify conditionals (guard clauses, early returns)
4. ✅ Tests still pass
5. ✅ Commit

**Example**:
```go
// Before: Nested conditionals
if len(timestamps) > 0 {
    minTs := timestamps[0]
    maxTs := timestamps[0]
    for _, ts := range timestamps[1:] {
        if ts < minTs {
            minTs = ts
        }
        if ts > maxTs {
            maxTs = ts
        }
    }
    return int((maxTs - minTs) / 60)
} else {
    return 0
}

// After: Guard clause
if len(timestamps) == 0 {
    return 0
}
minTs := timestamps[0]
maxTs := timestamps[0]
for _, ts := range timestamps[1:] {
    if ts < minTs {
        minTs = ts
    }
    if ts > maxTs {
        maxTs = ts
    }
}
return int((maxTs - minTs) / 60)
```

**Tests**: No new tests needed (behavior unchanged), existing tests verify correctness

---

### Pattern 3: Remove Duplication with Tests

**Scenario**: Duplicated code blocks, need to extract to shared helper

**Steps**:
1. ✅ Ensure tests pass
2. ✅ Identify duplication: Lines X-Y in File A same as Lines M-N in File B
3. ✅ Extract to shared helper
4. ✅ Replace first occurrence with helper call
5. ✅ Tests pass
6. ✅ Replace second occurrence
7. ✅ Tests pass
8. ✅ Commit

**Example**:
```go
// Before: Duplication
// File A:
if startTs > 0 {
    timestamps = append(timestamps, startTs)
}

// File B:
if endTs > 0 {
    timestamps = append(timestamps, endTs)
}

// After: Shared helper
func appendIfValid(timestamps []int64, ts int64) []int64 {
    if ts > 0 {
        return append(timestamps, ts)
    }
    return timestamps
}

// File A: timestamps = appendIfValid(timestamps, startTs)
// File B: timestamps = appendIfValid(timestamps, endTs)
```

**Tests**:
- Existing tests for Files A and B (still pass)
- New: `TestAppendIfValid` (covers helper)

---

## TDD Anti-Patterns (Avoid These)

### ❌ Anti-Pattern 1: "Skip Tests, Code Seems Fine"

**Problem**: Refactor without running tests
**Risk**: Break behavior without noticing
**Fix**: ALWAYS run tests after each change

### ❌ Anti-Pattern 2: "Write Tests After Refactoring"

**Problem**: Tests written to match new code (not verify behavior)
**Risk**: Tests pass but behavior changed
**Fix**: Write tests BEFORE refactoring (characterization tests)

### ❌ Anti-Pattern 3: "Batch Multiple Changes Before Testing"

**Problem**: Make 3-4 changes, then run tests
**Risk**: If tests fail, hard to identify which change broke it
**Fix**: Test after EACH change

### ❌ Anti-Pattern 4: "Update Tests to Match New Code"

**Problem**: Tests fail after refactoring, so "fix" tests
**Risk**: Masking behavioral changes
**Fix**: If tests fail, rollback refactoring → Fix code, not tests

### ❌ Anti-Pattern 5: "Low Coverage is OK for Refactoring"

**Problem**: Refactor code with <75% coverage
**Risk**: Behavioral changes not caught by tests
**Fix**: Achieve ≥95% coverage BEFORE refactoring

---

## Automation Support

**Continuous Testing** (automatically run tests on file save):

### Option 1: File Watcher (entr)

```bash
# Install entr
go install github.com/eradman/entr@latest

# Auto-run tests on file change
find internal/query -name '*.go' | entr -c go test ./internal/query/...
```

### Option 2: IDE Integration

- **VS Code**: Go extension auto-runs tests on save
- **GoLand**: Configure test auto-run in settings
- **Vim**: Use vim-go with `:GoTestFunc` on save

### Option 3: Pre-Commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
go test ./... || exit 1
go test -cover ./... | grep -E 'coverage: [0-9]+' || exit 1
```

**Checklist**:
- [ ] Automation setup: YES / NO
- [ ] Tests run automatically: YES / NO
- [ ] Feedback time: ___s (target <5s)

---

## Notes

- **TDD Discipline**: Tests must pass after EVERY single change
- **Small Steps**: Each refactoring step should take <10 minutes
- **Fast Tests**: Test suite should run in <10 seconds for fast feedback
- **No Guessing**: If unsure about behavior, write test to document it
- **Coverage Goal**: ≥95% for code being refactored, ≥85% overall

---

**Version**: 1.0 (Iteration 1)
**Next Review**: Iteration 2 (refine based on usage data)
**Automation**: See Problem V1 for automated complexity checking integration
