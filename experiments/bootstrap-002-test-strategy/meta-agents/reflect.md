# Reflect Capability

## Purpose
Calculate V(s) for testing state, assess test quality honestly, identify coverage gaps, and determine iteration progress toward convergence.

## Value Calculation V(s)

### Component: V_coverage (Weight: 0.3)

**Line Coverage**:
```
line_coverage = (lines_covered / total_lines) for non-test code
V_line = min(line_coverage / 0.80, 1.0)  # Target: 80%
```

**Branch Coverage**:
```
branch_coverage = (branches_covered / total_branches)
V_branch = min(branch_coverage / 0.70, 1.0)  # Target: 70%
```

**Combined Coverage**:
```
V_coverage = 0.6·V_line + 0.4·V_branch
```

**Data Sources**:
- `go test -cover ./...` for line coverage
- `go tool cover -func=coverage.out` for detailed coverage
- Coverage reports from data/test-coverage.json

**Honest Assessment**:
- Do NOT assume coverage without running tests
- Exclude test files from coverage calculations
- Report actual percentages, not rounded
- Identify specific functions/packages below threshold

### Component: V_reliability (Weight: 0.3)

**Test Pass Rate**:
```
pass_rate = (tests_passing / total_tests)
```

**Critical Path Coverage**:
```
critical_coverage = (critical_paths_tested / critical_paths_total)
critical_paths = error_handling + boundary_conditions + core_logic
```

**Test Stability** (no flaky tests):
```
stability = 1 - (flaky_tests / total_tests)
flaky_test = test that fails intermittently without code changes
```

**Combined Reliability**:
```
V_reliability = 0.4·pass_rate + 0.4·critical_coverage + 0.2·stability
```

**Data Sources**:
- Test execution logs: `go test -v ./...`
- Critical path analysis: manual code review + error history
- Flakiness detection: run tests multiple times (5+ runs)

**Honest Assessment**:
- Test 100 passing ≠ 100% reliability if critical paths untested
- Identify specific critical paths without tests
- Measure actual flakiness (run tests multiple times)
- Report error case coverage separately

### Component: V_maintainability (Weight: 0.2)

**Test Complexity**:
```
avg_test_complexity = sum(cyclomatic_complexity) / num_tests
max_acceptable = 10  # Cyclomatic complexity threshold
V_complexity = 1 - min(avg_test_complexity / max_acceptable, 1.0)
```

**Test Clarity**:
```
clarity_score = (tests_with_clear_names / total_tests) ×
                (tests_with_good_assertions / total_tests)
clear_name = descriptive, follows convention (TestFunction_Condition_Expectation)
good_assertion = specific (assert.Equal vs assert.NoError)
```

**DRY Principle**:
```
duplication_score = 1 - (duplicate_setup_lines / total_test_lines)
```

**Combined Maintainability**:
```
V_maintainability = 0.4·V_complexity + 0.3·clarity_score + 0.3·duplication_score
```

**Data Sources**:
- Test files analysis: count functions, assertions, patterns
- Cyclomatic complexity: gocyclo tool
- Code review: identify duplicate setup, unclear names

**Honest Assessment**:
- Count actual duplicate setup code lines
- Review test names for clarity (manual inspection)
- Calculate cyclomatic complexity objectively
- Report specific maintainability issues

### Component: V_speed (Weight: 0.2)

**Execution Time**:
```
execution_time = sum(test_duration) for all tests
baseline_time = initial measurement from Iteration 0
V_time = 1 - min((execution_time - baseline_time) / baseline_time, 1.0)
      = 1.0 if execution_time <= baseline_time (no slowdown)
```

**Parallel Efficiency**:
```
parallel_ratio = (tests_marked_parallel / parallelizable_tests)
parallelizable = tests without shared state
V_parallel = parallel_ratio
```

**Combined Speed**:
```
V_speed = 0.7·V_time + 0.3·V_parallel
```

**Data Sources**:
- Test timing: `go test -v ./...` output parsing
- Parallel test count: `grep "t.Parallel()" *_test.go | wc -l`
- Baseline timing from Iteration 0 data/test-execution.json

**Honest Assessment**:
- Measure actual test execution time (run multiple times)
- Identify slow tests (>100ms for unit tests)
- Calculate parallel potential (manual analysis)
- Report specific slow tests

## Overall Value Calculation

```
V(s) = 0.3·V_coverage + 0.3·V_reliability + 0.2·V_maintainability + 0.2·V_speed

Target: V(sₙ) ≥ 0.80
```

## Gap Identification

### Coverage Gaps
- Functions with <80% line coverage
- Packages with <70% branch coverage
- Error paths without tests
- Boundary conditions without tests

### Reliability Gaps
- Critical paths without tests
- Flaky tests (intermittent failures)
- Missing error case tests
- Insufficient edge case coverage

### Maintainability Gaps
- Tests with unclear names
- Duplicate setup code (no fixtures)
- High complexity tests (cyclomatic > 10)
- Poor assertion specificity

### Speed Gaps
- Slow unit tests (>100ms)
- Tests not using t.Parallel() when possible
- Inefficient test setup/teardown
- Redundant test operations

## Quality Assessment

### Test Quality Checklist
- [ ] Tests pass consistently (5+ runs)
- [ ] Coverage targets met (80% line, 70% branch)
- [ ] Critical paths tested (error handling, edge cases)
- [ ] Clear test names (TestFunction_Condition_Expectation)
- [ ] Specific assertions (assert.Equal, not just assert.NoError)
- [ ] Table-driven tests for multiple scenarios
- [ ] Subtests for related cases
- [ ] Proper test isolation (mocks for external deps)
- [ ] Fast execution (<100ms per unit test)
- [ ] No flaky tests

## Convergence Progress

### Iteration-to-Iteration Comparison
```
ΔV = V(sₙ) - V(sₙ₋₁)

Progress Assessment:
- ΔV > 0.05: Good progress
- ΔV = 0.02-0.05: Moderate progress
- ΔV < 0.02: Slow progress, may be converging
```

### Component-Level Progress
Track each component separately:
- ΔV_coverage: Coverage improvement
- ΔV_reliability: Reliability improvement
- ΔV_maintainability: Maintainability improvement
- ΔV_speed: Speed improvement

Identify which component is blocking convergence.

## Reflection Questions

### After Each Iteration
1. **Coverage**: Did we increase coverage? Which gaps remain?
2. **Reliability**: Are tests more reliable? Any flaky tests?
3. **Maintainability**: Are tests easier to understand and modify?
4. **Speed**: Did test execution time increase? Can we parallelize?
5. **Agent Effectiveness**: Did specialized agents improve outcomes?
6. **Strategy**: Is our testing strategy effective? What to adjust?

### Honest Self-Assessment
- What worked well in this iteration?
- What didn't work as expected?
- Are we generating high-quality tests or just increasing numbers?
- Is test maintainability degrading as we add more tests?
- Are we testing the right things (critical paths vs trivial code)?

## Output Format

**Reflection Document**:
- Current V(s) calculation with all components
- Component-level breakdown and evidence
- Identified gaps (specific functions, packages)
- Quality assessment against checklist
- Progress since last iteration (ΔV)
- Recommendations for next iteration
