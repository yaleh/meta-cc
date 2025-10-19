# Refactoring Safety Checklist

**Purpose**: Ensure safe, behavior-preserving refactoring through systematic verification

**When to Use**: Before starting ANY refactoring work

**Origin**: Iteration 1 - Problem P1 (No Refactoring Safety Checklist)

---

## Pre-Refactoring Checklist

### 1. Baseline Verification

- [ ] **All tests passing**: Run full test suite (`go test ./...`)
  - Status: PASS / FAIL
  - If FAIL: Fix failing tests BEFORE refactoring

- [ ] **No uncommitted changes**: Check git status
  - Status: CLEAN / DIRTY
  - If DIRTY: Commit or stash before refactoring

- [ ] **Baseline metrics recorded**: Capture current complexity, coverage, duplication
  - Complexity: `gocyclo -over 1 <target-package>/`
  - Coverage: `go test -cover <target-package>/...`
  - Duplication: `dupl -threshold 15 <target-package>/`
  - Saved to: `data/iteration-N/baseline-<target>.txt`

### 2. Test Coverage Verification

- [ ] **Target code has tests**: Verify tests exist for code being refactored
  - Test file: `<target>_test.go`
  - Coverage: ___% (from `go test -cover`)
  - If <75%: Write tests FIRST (TDD)

- [ ] **Tests cover current behavior**: Run tests and verify they pass
  - Characterization tests: Tests that document current behavior
  - Edge cases covered: Empty inputs, nil checks, error conditions
  - If gaps found: Write additional tests FIRST

### 3. Refactoring Plan

- [ ] **Refactoring pattern selected**: Choose appropriate pattern
  - Pattern: _______________ (e.g., Extract Method, Simplify Conditionals)
  - Reference: `knowledge/patterns/<pattern>.md`

- [ ] **Incremental steps defined**: Break into small, verifiable steps
  - Step 1: _______________
  - Step 2: _______________
  - Step 3: _______________
  - (Each step should take <10 minutes, pass tests)

- [ ] **Rollback plan documented**: Define how to undo if problems occur
  - Rollback method: Git revert / Git reset / Manual
  - Rollback triggers: Tests fail, complexity increases, coverage decreases >5%

---

## During Refactoring Checklist (Per Step)

### Step N: <Step Description>

#### Before Making Changes

- [ ] **Tests pass**: `go test ./...`
  - Status: PASS / FAIL
  - Time: ___s

#### Making Changes

- [ ] **One change at a time**: Make minimal, focused change
  - Files modified: _______________
  - Lines changed: ___
  - Scope: Single function / Multiple functions / Cross-file

- [ ] **No behavioral changes**: Only restructure, don't change logic
  - Verified: Code does same thing, just organized differently

#### After Making Changes

- [ ] **Tests still pass**: `go test ./...`
  - Status: PASS / FAIL
  - Time: ___s
  - If FAIL: Rollback immediately

- [ ] **Coverage maintained or improved**: `go test -cover ./...`
  - Before: ___%
  - After: ___%
  - Change: +/- ___%
  - If decreased >1%: Investigate and add tests

- [ ] **No new complexity**: `gocyclo -over 10 <target-file>`
  - Functions >10: ___
  - If increased: Rollback or simplify further

- [ ] **Commit incremental progress**: `git add . && git commit -m "refactor: <description>"`
  - Commit hash: _______________
  - Message: "refactor: <pattern> - <what changed>"
  - Safe rollback point: Can revert this specific change

---

## Post-Refactoring Checklist

### 1. Final Verification

- [ ] **All tests pass**: `go test ./...`
  - Status: PASS
  - Duration: ___s

- [ ] **Coverage improved or maintained**: `go test -cover ./...`
  - Baseline: ___%
  - Final: ___%
  - Change: +___%
  - Target: ≥85% overall, ≥95% for refactored code

- [ ] **Complexity reduced**: `gocyclo -avg <target-package>/`
  - Baseline: ___
  - Final: ___
  - Reduction: ___%
  - Target function: <10 complexity

- [ ] **No duplication introduced**: `dupl -threshold 15 <target-package>/`
  - Baseline groups: ___
  - Final groups: ___
  - Change: -___ groups

- [ ] **No new static warnings**: `go vet <target-package>/...`
  - Warnings: 0
  - If >0: Fix before finalizing

### 2. Behavior Preservation

- [ ] **Integration tests pass** (if applicable)
  - Status: PASS / N/A

- [ ] **Manual verification** (for critical code)
  - Test scenario 1: _______________
  - Test scenario 2: _______________
  - Result: Behavior unchanged

- [ ] **Performance not regressed** (if applicable)
  - Benchmark: `go test -bench . <target-package>/...`
  - Change: +/- ___%
  - Acceptable: <10% regression

### 3. Documentation

- [ ] **Code documented**: Add/update GoDoc comments
  - Public functions: ___ documented / ___ total
  - Target: 100% of public APIs

- [ ] **Refactoring logged**: Document refactoring in session log
  - File: `data/iteration-N/refactoring-log.md`
  - Logged: Pattern, time, issues, lessons

### 4. Final Commit

- [ ] **Clean git history**: All incremental commits made
  - Total commits: ___
  - Clean messages: YES / NO
  - Revertible: YES / NO

- [ ] **Final metrics recorded**: Save post-refactoring metrics
  - File: `data/iteration-N/final-<target>.txt`
  - Metrics: Complexity, coverage, duplication saved

---

## Rollback Protocol

**When to Rollback**:
- Tests fail after a refactoring step
- Coverage decreases >5%
- Complexity increases
- New static analysis errors
- Refactoring taking >2x estimated time
- Uncertainty about correctness

**How to Rollback**:
1. **Immediate**: Stop making changes
2. **Assess**: Identify which commit introduced problem
3. **Revert**: `git revert <commit-hash>` or `git reset --hard <last-good-commit>`
4. **Verify**: Run tests to confirm rollback successful
5. **Document**: Log why rollback was needed
6. **Re-plan**: Choose different approach or break into smaller steps

**Rollback Checklist**:
- [ ] Identified problem commit: _______________
- [ ] Reverted changes: `git revert _______________`
- [ ] Tests pass after rollback: PASS / FAIL
- [ ] Documented rollback reason: _______________
- [ ] New plan documented: _______________

---

## Safety Statistics (Track Over Time)

**Refactoring Session**: ___ (e.g., calculateSequenceTimeSpan - 2025-10-19)

| Metric | Value |
|--------|-------|
| **Steps completed** | ___ |
| **Rollbacks needed** | ___ |
| **Tests failed** | ___ times |
| **Coverage regression** | YES / NO |
| **Complexity regression** | YES / NO |
| **Total time** | ___ minutes |
| **Average time per step** | ___ minutes |
| **Safety incidents** | ___ (breaking changes, lost work, etc.) |

**Safety Score**: (Steps completed - Rollbacks - Safety incidents) / Steps completed × 100% = ___%

**Target**: ≥95% safety score (≤5% incidents)

---

## Checklist Usage Example

**Refactoring**: `calculateSequenceTimeSpan` (Complexity 10 → <8)
**Pattern**: Extract Method (collectOccurrenceTimestamps, findMinMaxTimestamps)
**Date**: 2025-10-19

### Pre-Refactoring
- [x] All tests passing: PASS (0.008s)
- [x] No uncommitted changes: CLEAN
- [x] Baseline metrics: Saved to `data/iteration-1/baseline-sequences.txt`
  - Complexity: 10
  - Coverage: 85%
  - Duplication: 0 groups in this file
- [x] Target has tests: `sequences_test.go` exists
- [x] Coverage: 85% (need to add edge case tests)
- [x] Pattern: Extract Method
- [x] Steps: 1) Write edge case tests, 2) Extract collectTimestamps, 3) Extract findMinMax
- [x] Rollback: Git revert if tests fail

### During Refactoring - Step 1: Write Edge Case Tests
- [x] Tests pass before: PASS
- [x] Added tests for empty timestamps, single timestamp
- [x] Tests pass after: PASS
- [x] Coverage: 85% → 95%
- [x] Commit: `git commit -m "test: add edge cases for calculateSequenceTimeSpan"`

### During Refactoring - Step 2: Extract collectTimestamps
- [x] Tests pass before: PASS
- [x] Extracted helper, updated main function
- [x] Tests pass after: PASS
- [x] Coverage: 95% (maintained)
- [x] Complexity: 10 → 7
- [x] Commit: `git commit -m "refactor: extract collectTimestamps helper"`

### Post-Refactoring
- [x] All tests pass: PASS
- [x] Coverage: 85% → 95% (+10%)
- [x] Complexity: 10 → 6 (-40%)
- [x] Duplication: 0 (no change)
- [x] Documentation: Added GoDoc to calculateSequenceTimeSpan
- [x] Logged: `data/iteration-1/refactoring-log.md`

**Safety Score**: 3 steps, 0 rollbacks, 0 incidents = 100%

---

## Notes

- **Honesty**: Mark actual status, not desired status
- **Discipline**: Don't skip checks "because it seems fine"
- **Speed**: Checks should be quick (<1 minute total per step)
- **Automation**: Use scripts to automate metric collection (see Problem V1)
- **Adaptation**: Adjust checklist based on project needs, but maintain core safety principles

---

**Version**: 1.0 (Iteration 1)
**Next Review**: Iteration 2 (refine based on usage data)
