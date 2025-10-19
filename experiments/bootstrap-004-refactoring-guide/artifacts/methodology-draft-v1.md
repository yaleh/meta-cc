# Refactoring Methodology (Draft v1)

**Version**: 1.0
**Created**: 2025-10-19
**Status**: Initial Draft (Based on Iteration 1 observations)
**Completeness**: ~25% (4/15 checklist items)

---

## Overview

This methodology provides a systematic approach to code refactoring based on empirical observations from refactoring the `internal/query/` package. It emphasizes safety, incrementality, and measurable improvement.

**Core Principles**:
1. **Test First**: Ensure ≥85% test coverage before refactoring
2. **Incremental Steps**: Make small, reversible changes
3. **Verify Always**: Run tests after each change
4. **Measure Progress**: Track complexity, duplication, and coverage

---

## Process Steps

### 1. Identification Phase

**Goal**: Find refactoring targets systematically

#### 1.1 Run Automated Analysis

```bash
# Cyclomatic complexity
gocyclo -over 10 <package>

# Code duplication (15+ tokens)
dupl -threshold 15 <package>

# Static analysis
staticcheck <package>
go vet <package>

# Test coverage
go test -coverprofile=coverage.out <package>
go tool cover -func=coverage.out
```

#### 1.2 Identify Code Smells

**High Priority Smells**:
- Cyclomatic complexity >10 (production code)
- Code duplication >15 tokens (production code)
- Test coverage <85%
- Static analysis errors/warnings

**Medium Priority Smells**:
- Long functions (>50 lines)
- Large parameter lists (>5 parameters)
- Magic numbers/strings
- Unclear naming

**Low Priority Smells**:
- Test code duplication (acceptable in Go)
- Test function complexity (acceptable for comprehensive tests)

#### 1.3 Prioritize Targets

**Prioritization Framework**:

```
Priority = Impact × Urgency / Effort

Where:
  Impact = Code_Quality_Benefit (0-10)
  Urgency = Technical_Debt_Severity (0-10)
  Effort = Estimated_Time_Hours (1-10)
```

**High Priority**:
- High complexity + low test coverage
- Critical path code with duplication
- Functions changed frequently (git history)

**Medium Priority**:
- Moderate complexity
- Non-critical duplication
- Isolated modules

**Low Priority**:
- Naming clarity issues
- Test code improvements
- Minor optimizations

---

### 2. Planning Phase

**Goal**: Plan safe, incremental refactoring steps

#### 2.1 Select Refactoring Technique

**For Code Duplication** (CS-001):
- **If >95% identical**: Extract Method with Parameter
- **If 70-95% identical**: Extract Common Logic + Wrappers
- **If <70% identical**: Identify common sub-patterns

**For High Complexity** (CS-002):
- **If nested loops**: Extract Helper + Restructure
- **If deep nesting**: Guard Clauses + Early Returns
- **If long function**: Extract Method for each responsibility

**For Naming Issues** (CS-005):
- **If ambiguous**: Rename to reflect actual behavior
- **If too generic**: Add specificity
- **If misleading**: Rename + add documentation

#### 2.2 Create Incremental Steps Plan

**Template**:
```markdown
## Refactoring Plan: <Target Name>

### Code Smell: <Type> (Severity: HIGH/MEDIUM/LOW)

### Technique: <Chosen Technique>

### Incremental Steps:
1. [ ] Verify test coverage for target function
2. [ ] Extract <helper function name>
3. [ ] Refactor <main function> to use helper
4. [ ] Run tests after each step
5. [ ] Verify metrics improvement

### Safety Checkpoints:
- [ ] Tests pass before starting
- [ ] Tests pass after each step
- [ ] Complexity reduced (gocyclo)
- [ ] No behavior changes

### Expected Metrics:
- Complexity: <baseline> → <target>
- Duplication: <baseline lines> → <target lines>
- Coverage: maintained ≥<baseline>%
```

---

### 3. Testing Phase

**Goal**: Ensure tests provide safety net before refactoring

#### 3.1 Verify Test Coverage

```bash
go test -cover <package>

# Ensure ≥85% coverage
# If <85%, add tests before refactoring
```

#### 3.2 Run Baseline Tests

```bash
go test -v <package> > tests-before.txt
# All tests must pass (100% pass rate)
```

#### 3.3 Create Behavior Preservation Tests (If Needed)

**When to Add**:
- Complex functions with edge cases
- Functions without unit tests
- Integration points with external dependencies

**How to Add**:
```go
func TestExistingBehavior(t *testing.T) {
    // Document current behavior even if not ideal
    // Ensures refactoring preserves behavior
}
```

---

### 4. Execution Phase

**Goal**: Execute refactoring incrementally with continuous verification

#### 4.1 Setup Version Control

```bash
# Create feature branch
git checkout -b refactor/<feature>-iteration-<n>

# Ensure clean status
git status
```

#### 4.2 Incremental Transformation

**Protocol**:

```
FOR each_step IN refactoring_plan:
    1. Make ONE small change
    2. Run tests immediately: go test <package>
    3. IF tests pass:
         git add <files>
         git commit -m "refactor: <description>"
    4. ELSE:
         git restore <files>
         Investigate failure
         Retry with smaller step
    5. NEVER proceed with failing tests
```

**Commit Message Template**:
```
refactor(<module>): <brief description>

- <Change 1>
- <Change 2>
- <Metrics improvement>
- Maintain 100% test pass rate

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

#### 4.3 Safety Checkpoints

After **each change**:
- [ ] Run tests: `go test <package>`
- [ ] Verify no compiler errors
- [ ] Check test output for warnings

After **each refactoring target**:
- [ ] Run full test suite: `go test ./...`
- [ ] Measure complexity: `gocyclo <file>`
- [ ] Check duplication: `dupl <file>`
- [ ] Verify coverage: `go test -cover <package>`

---

### 5. Verification Phase

**Goal**: Confirm refactoring improved quality without breaking behavior

#### 5.1 Test Verification

```bash
# Run full test suite
go test -v <package> > tests-after.txt

# Verify 100% pass rate
# Compare with baseline (diff tests-before.txt tests-after.txt)
```

#### 5.2 Metrics Verification

```bash
# Complexity
gocyclo <package> > complexity-after.txt
# Verify reduction

# Duplication
dupl -threshold 15 <package> > duplication-after.txt
# Verify reduction

# Coverage
go test -coverprofile=coverage-after.out <package>
go tool cover -func=coverage-after.out > coverage-after-summary.txt
# Verify maintained or improved
```

#### 5.3 Behavior Preservation Verification

**Automated**:
- All existing tests pass ✅
- Coverage maintained ✅

**Manual** (for critical paths):
- Code review of changes
- Validate edge cases still handled
- Verify error handling unchanged

---

## Code Smell Catalog

### CS-001: Duplicated Code

**Detection**:
- `dupl -threshold 15 <package>`
- Manual inspection for semantic duplication

**Severity Levels**:
- **HIGH**: Production code, >90% identical, >15 lines
- **MEDIUM**: Production code, 70-90% identical, >10 lines
- **LOW**: Test code, any duplication

**Refactoring Techniques**:
1. **Extract Method with Parameter** (for >95% identical)
2. **Extract Common Logic** (for 70-95% identical)
3. **Template Method Pattern** (for algorithmic similarity)

**Example** (from Iteration 1):
```go
// BEFORE: Two 95% identical functions
func buildContextBefore(...) { /* 18 lines */ }
func buildContextAfter(...) { /* 18 lines, 95% identical */ }

// AFTER: Unified function with direction parameter
func buildContextWindow(..., direction string) { /* unified logic */ }
func buildContextBefore(...) { return buildContextWindow(..., "before") }
func buildContextAfter(...) { return buildContextWindow(..., "after") }
```

---

### CS-002: Complex Function

**Detection**:
- `gocyclo -over 10 <package>`
- Cyclomatic complexity >10 for production code

**Severity Levels**:
- **HIGH**: Complexity >15, critical path code
- **MEDIUM**: Complexity 11-15, non-critical code
- **LOW**: Complexity >10 for test functions (acceptable)

**Refactoring Techniques**:
1. **Extract Helper Function** (for nested loops)
2. **Guard Clauses** (for deep nesting)
3. **Separate Concerns** (for multi-responsibility functions)

**Example** (from Iteration 1):
```go
// BEFORE: Nested loops, complexity 11
for _, occ := range occurrences {
    for _, tc := range toolCalls {
        if tc.turn == occ.StartTurn || tc.turn == occ.EndTurn {
            ts := getToolCallTimestamp(entries, tc.uuid)
            // complex min/max logic
        }
    }
}

// AFTER: Helper extraction, complexity 10
func findTimestampForTurn(...) int64 { /* isolated lookup */ }

for _, occ := range occurrences {
    startTs := findTimestampForTurn(entries, toolCalls, occ.StartTurn)
    endTs := findTimestampForTurn(entries, toolCalls, occ.EndTurn)
    timestamps = append(timestamps, startTs, endTs)
}
// Simple min/max finding (separate loop)
```

---

## Refactoring Techniques

### Technique 1: Extract Method with Parameter

**When to Use**:
- Two functions are >90% identical
- Difference is a single conditional or parameter value

**Steps**:
1. Identify the single difference point
2. Design parameter to control difference (string, bool, enum)
3. Extract unified function with parameter
4. Replace originals with thin wrappers
5. Run tests

**Benefits**:
- Eliminates duplication (DRY principle)
- Single source of truth
- Preserves API compatibility

**Transferability**: 🌍 Universal (all languages)

---

### Technique 2: Extract Helper for Nested Loops

**When to Use**:
- Function has nested loops (complexity contributor)
- Inner loop performs independent operation

**Steps**:
1. Extract inner loop logic to helper function
2. Restructure main function to collect data first
3. Process collected data in single loop
4. Verify complexity reduction

**Benefits**:
- Reduces cyclomatic complexity
- Improves time complexity (eliminates O(n*m))
- Helper is independently testable

**Transferability**: 🌍 Universal (all languages)

---

## Safety Checklist

### Before Starting
- [ ] Test coverage ≥85%
- [ ] All tests pass (100% pass rate)
- [ ] Git status is clean
- [ ] Feature branch created
- [ ] Baseline metrics collected

### During Refactoring
- [ ] Make ONE change at a time
- [ ] Run tests after EACH change
- [ ] Commit after successful step
- [ ] Rollback if tests fail
- [ ] NEVER proceed with failing tests

### After Completing
- [ ] All tests pass (100% pass rate)
- [ ] Metrics improved (or maintained)
- [ ] Coverage maintained (≥baseline)
- [ ] No behavior changes
- [ ] Git history is clean

---

## Rollback Procedures

### If a Step Fails

```bash
# Review changes
git diff

# Restore specific file
git restore <file>

# OR reset all changes (last resort)
git reset --hard HEAD
```

### If Tests Fail After Commit

```bash
# Revert last commit
git revert HEAD

# OR reset to previous commit (if not pushed)
git reset --hard HEAD~1
```

---

## Automation Opportunities

### Identified in Iteration 1

**Opportunity 1**: Code Smell Detection Script
```bash
#!/bin/bash
# detect-code-smells.sh
gocyclo -over 10 $1
dupl -threshold 15 $1
staticcheck $1
go test -cover $1
```

**Opportunity 2**: Refactoring Safety Checker
```bash
#!/bin/bash
# refactoring-safety-check.sh
go test ./...
go tool cover -func=coverage.out | grep total
gocyclo <package>
```

**Opportunity 3**: Metrics Comparison Tool
```bash
# Compare baseline vs. current
diff complexity-baseline.txt complexity-current.txt
diff duplication-baseline.txt duplication-current.txt
```

---

## Common Pitfalls and Solutions

### Pitfall 1: Refactoring Without Tests

**Problem**: Changes break behavior, no way to detect
**Solution**: Add tests BEFORE refactoring (TDD approach)

### Pitfall 2: Big-Bang Refactoring

**Problem**: Large changes, hard to identify what broke
**Solution**: Incremental steps with testing after each

### Pitfall 3: Ignoring Test Failures

**Problem**: Proceeding with broken tests leads to cascading failures
**Solution**: NEVER proceed with failing tests, rollback immediately

### Pitfall 4: Over-Refactoring

**Problem**: Diminishing returns, wasted effort
**Solution**: Pragmatic assessment ("good enough" is often good enough)

---

## Cross-Language Adaptation Notes

### Go-Specific Patterns (from Iteration 1)

- Test code duplication is acceptable (table-driven tests)
- Exported functions must have doc comments
- `gocyclo` and `dupl` are Go-specific tools

### Universal Patterns

- Extract Method (all languages)
- Helper Extraction (all languages)
- Incremental Testing (all languages)
- Git Commit Strategy (all projects)

### Adapting to Python

- Use `pylint --max-complexity 10`
- Use `radon cc` for complexity
- Use `pytest --cov` for coverage

### Adapting to Rust

- Use `cargo clippy` for linting
- Use `cargo tarpaulin` for coverage
- Complexity tools less mature (manual inspection)

### Adapting to JavaScript/TypeScript

- Use `eslint --max-complexity 10`
- Use `jscpd` for duplication
- Use `jest --coverage` for coverage

---

## Methodology Completeness Checklist

**Status: 4/15 items complete**

- [x] Process steps documented (partial: 5 phases outlined)
- [ ] Code smell detection criteria fully defined (partial: 2/7 smells)
- [ ] Refactoring technique catalog created (partial: 2/10 techniques)
- [ ] Safety verification procedures documented (partial: basic checklist)
- [ ] Risk assessment framework defined (not started)
- [ ] Examples for each refactoring type provided (partial: 2 examples)
- [ ] Edge cases and failure modes documented (not started)
- [ ] Decision trees for refactoring choices (not started)
- [x] Rollback procedures documented (complete)
- [x] Testing strategy for refactoring defined (complete)
- [x] Automation opportunities identified (partial: 3 scripts outlined)
- [ ] Tool usage guidelines created (partial: basic commands)
- [ ] Cross-language adaptation notes (partial: 4 languages)
- [ ] Common pitfalls documented (partial: 4 pitfalls)
- [ ] Success patterns identified (partial: 2 patterns)

**Next Iteration Goal**: Increase to 10/15 items (67% completeness)

---

## References

- **BAIME Framework**: Observe → Codify → Automate cycle
- **Value Functions**: V_instance (refactoring quality), V_meta (methodology quality)
- **Bootstrap-002**: Test Strategy methodology (validated framework)
- **Bootstrap-003**: Error Recovery methodology (rapid convergence example)

---

## Version History

- **v1.0** (2025-10-19): Initial draft based on Iteration 1 observations
  - 5 process steps (Identify, Plan, Test, Execute, Verify)
  - 2 code smells documented (Duplication, Complexity)
  - 2 refactoring techniques (Extract Method with Parameter, Extract Helper)
  - Safety checklist and rollback procedures
  - 3 automation opportunities identified

**Next Version**: v2.0 (Iteration 2)
- Complete code smell catalog (7 smells)
- Expand refactoring technique guide (10 techniques)
- Add decision frameworks
- Create risk assessment matrix
- Provide more examples

---

**Document Status**: ✅ Draft Complete (25% methodology completeness)
**Ready for Iteration 2**: Yes (foundation established for further codification)
