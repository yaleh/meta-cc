# CI/CD Quality Gates Methodology

**Domain**: Continuous Integration / Continuous Deployment
**Version**: 1.0 (Bootstrap-007, Iteration 1)
**Date**: 2025-10-16
**Status**: Production-Ready

---

## Overview

This methodology defines how to implement and enforce quality gates in CI/CD pipelines to prevent code quality violations from entering the codebase. Quality gates act as automated checkpoints that block merges or deployments when quality standards are not met.

**Core Principle**: **Shift-Left Quality** - Catch violations as early as possible (pre-commit > CI > production)

---

## Quality Gate Categories

### 1. Coverage Threshold Gates

**Purpose**: Ensure adequate test coverage to prevent regressions

**When to Use**:
- Projects with ≥70% coverage (to maintain or improve)
- Critical codebases requiring high reliability
- After achieving initial coverage baseline

**Implementation Pattern**:

```yaml
# .github/workflows/ci.yml
- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    THRESHOLD=80.0

    if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
      echo "❌ ERROR: Coverage ${COVERAGE}% < ${THRESHOLD}%"
      exit 1
    fi
```

**Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Threshold: 80% | Industry standard for production code |
| Enforcement: Hard block | Prevents coverage regression |
| Scope: Total coverage | Simpler than per-package enforcement |
| Platform: Single (ubuntu-latest) | Avoid redundant checks on 6 platforms |
| Trigger: After test + coverage run | Reuse coverage.out from tests |

**Success Criteria**:
- ✓ CI fails if coverage < threshold
- ✓ Clear error message with actual vs expected
- ✓ Instructions on how to fix (run `make test-coverage`)

---

### 2. Lint Violation Gates

**Purpose**: Enforce code style and catch common bugs automatically

**When to Use**:
- All projects (linting is foundational quality)
- After establishing project style guide
- When team agrees on lint rules

**Implementation Pattern**:

```yaml
# .github/workflows/ci.yml
lint:
  name: Lint
  runs-on: ubuntu-latest
  steps:
    - uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        args: --timeout=5m
```

**Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Tool: golangci-lint | Comprehensive Go linter aggregator |
| Version: latest | Stay current with new checks |
| Timeout: 5m | Balance thoroughness vs CI speed |
| Enforcement: Action default | Fails on violations automatically |
| Separate job | Parallel with tests for faster feedback |

**Success Criteria**:
- ✓ CI fails on any lint violation
- ✓ Violations shown in CI logs
- ✓ Developers can run locally (`make lint`)

---

### 3. CHANGELOG Validation Gates

**Purpose**: Ensure user-facing changes are documented

**When to Use**:
- Projects with public releases
- When using Keep a Changelog format
- PRs with code changes (not docs-only)

**Implementation Pattern**:

```bash
# scripts/check-changelog-updated.sh
CHANGED_FILES=$(git diff --name-only origin/$GITHUB_BASE_REF...HEAD)
CHANGELOG_UPDATED=$(echo "$CHANGED_FILES" | grep -c "^CHANGELOG.md$")
CODE_FILES=$(echo "$CHANGED_FILES" | grep -E '\.(go|mod)$' | wc -l)

if [ $CODE_FILES -gt 0 ] && [ $CHANGELOG_UPDATED -eq 0 ]; then
  echo "⚠️ WARNING: CHANGELOG.md not updated"
  # Currently warning, can make blocking
fi
```

**Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Enforcement: Warning only | Allow WIP PRs, block before merge manually |
| Exceptions: Docs-only, tests-only | Reduce false positives |
| Bypass: `[skip changelog]` in PR title | Emergency escape hatch |
| Trigger: PR events only | Not needed for direct commits |
| Script: Separate bash file | Reusable, testable logic |

**Future Enhancement**: Convert to hard block after team adopts workflow

**Success Criteria**:
- ✓ Warning shown when CHANGELOG not updated
- ✓ No warning for docs-only changes
- ✓ Clear instructions on what to add

---

## Enforcement Levels

Quality gates should be implemented with appropriate enforcement levels:

### Hard Block (Fail CI)
**Use When**:
- Violation causes immediate problems (breaking changes, security)
- Standard is well-established and stable
- Team has buy-in and tooling

**Examples**:
- Coverage threshold (prevents regression)
- Lint violations (style agreed upon)
- Build failures (obviously blocking)

### Soft Warning (Log but Pass)
**Use When**:
- New standard being introduced
- Team learning new practice
- High false-positive rate

**Examples**:
- CHANGELOG validation (new workflow)
- Experimental lint rules
- Deprecated API usage

**Transition Path**: Warning → Blocking after 2-4 weeks adoption period

---

## Implementation Steps

### Step 1: Establish Baseline
```bash
# Measure current state
make test-coverage  # Check coverage
make lint           # Check violations
```

**Decision Point**: Only enforce if baseline already meets threshold

### Step 2: Add Quality Gate
```yaml
# Add step to CI workflow
- name: Check quality gate
  run: |
    # Measure metric
    # Compare to threshold
    # Fail if violation
```

### Step 3: Document Standard
```markdown
# In docs/methodology/
- What: Threshold value and why
- When: Enforcement trigger
- How: Fix violations locally
- Bypass: Emergency escape hatch
```

### Step 4: Monitor Effectiveness
```bash
# Track metrics over time
- False positive rate
- Developer friction
- Actual quality improvement
```

### Step 5: Iterate
- Adjust thresholds based on data
- Add exceptions for edge cases
- Improve error messages

---

## Common Pitfalls

### Pitfall 1: Too Strict Too Soon
**Problem**: Blocking on 80% coverage when current is 40%

**Solution**:
- Start with current baseline (40%)
- Increment by 5% per iteration
- Add "no regression" rule: don't go below baseline

### Pitfall 2: No Local Verification
**Problem**: Developers only find violations in CI

**Solution**:
- Provide `make lint`, `make test-coverage` commands
- Add pre-commit hooks (optional)
- Document local workflow

### Pitfall 3: Unclear Error Messages
**Problem**: "Coverage check failed" with no guidance

**Solution**:
```bash
echo "❌ Coverage ${COVERAGE}% < ${THRESHOLD}%"
echo ""
echo "To fix:"
echo "  1. Run: make test-coverage"
echo "  2. Open: coverage.html"
echo "  3. Add tests for uncovered code"
```

### Pitfall 4: No Escape Hatch
**Problem**: Emergency hotfix blocked by quality gate

**Solution**:
- Provide bypass mechanism (`[skip ci]`, `--no-verify`)
- Document when bypass is acceptable
- Require follow-up PR to fix properly

---

## Quality Gate Decision Framework

Use this decision tree to determine if a quality gate is appropriate:

```
Is the metric measurable automatically?
├─ No  → Don't add gate (use code review instead)
└─ Yes → Is the threshold clear and objective?
    ├─ No  → Define threshold first
    └─ Yes → Does current code meet threshold?
        ├─ No  → Improve code first OR lower threshold
        └─ Yes → Is there team buy-in?
            ├─ No  → Start with warning, get feedback
            └─ Yes → Implement as blocking gate ✓
```

---

## Platform-Specific Considerations

### GitHub Actions
- Use separate jobs for parallel execution
- Leverage matrix builds but avoid redundant checks
- Use `if:` conditions to run checks once (ubuntu-latest + latest Go)

### GitLab CI
- Use `rules:` for conditional execution
- Leverage `allow_failure: false` for blocking
- Use `artifacts:` to save reports

### Local Development
- Provide Makefile targets matching CI checks
- Consider git hooks for pre-commit validation
- Document how to run all checks locally

---

## Metrics and Monitoring

Track quality gate effectiveness:

### Leading Indicators (CI Behavior)
- **Gate trigger rate**: How often gates catch violations
- **False positive rate**: Gates triggering incorrectly
- **Average time to fix**: Developer friction metric
- **Bypass frequency**: How often gates are skipped

### Lagging Indicators (Code Quality)
- **Coverage trend**: Is coverage stable or improving?
- **Defect rate**: Fewer bugs reaching production?
- **Code review time**: Less time on style issues?
- **Developer satisfaction**: Team happiness with gates

**Target Ranges**:
- Gate trigger rate: 5-15% (too low = gate not useful, too high = too strict)
- False positive rate: <5% (minimize developer frustration)
- Bypass frequency: <2% (emergency use only)

---

## Testing Quality Gates

Quality gates themselves should be tested:

### Unit Tests (Scripts)
```bash
# Test coverage threshold check
test_coverage_above_threshold() {
  COVERAGE=85.0
  THRESHOLD=80.0
  # Should pass
}

test_coverage_below_threshold() {
  COVERAGE=75.0
  THRESHOLD=80.0
  # Should fail (exit 1)
}
```

### Integration Tests (CI)
- Trigger quality gate on test PR
- Verify gate blocks/warns as expected
- Check error message clarity

### Regression Tests
- Ensure gates don't break on edge cases
- Test bypass mechanisms work
- Verify performance (gate should be fast)

---

## Evolution and Maintenance

Quality gates are not "set and forget":

### Quarterly Review
- Review gate effectiveness metrics
- Adjust thresholds based on trends
- Remove gates that don't provide value
- Add new gates for emerging needs

### When to Adjust Thresholds
**Increase**:
- Team consistently exceeds current threshold
- Quality standards need to improve
- Industry best practices evolve

**Decrease**:
- Gate causing excessive friction
- False positive rate too high
- Threshold unrealistic for codebase

**Remove**:
- Metric no longer relevant
- Better metric available
- Tool deprecation

---

## Case Study: meta-cc Coverage Gate

**Context**: Bootstrap-007 Iteration 1

**Baseline**:
- Current coverage: 71.7%
- Target: 80% (industry standard)
- Gap: 8.3%

**Decision**: Implement blocking coverage gate at 80%

**Rationale**:
1. **Prevents regression**: Even if current coverage is below target, gate prevents it from going lower
2. **Clear target**: Team knows to add tests to reach 80%
3. **Measurable**: Automated coverage reports

**Implementation**:
```yaml
- name: Check coverage threshold
  if: matrix.os == 'ubuntu-latest' && matrix.go == '1.22'
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    THRESHOLD=80.0
    if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
      echo "❌ ERROR: Coverage ${COVERAGE}% is below threshold ${THRESHOLD}%"
      exit 1
    fi
```

**Outcome**:
- Gate would fail CI at 71.7% (expected)
- Provides clear error message with instructions
- Incentivizes adding tests to reach 80%

**Next Steps**:
- Add tests to increase coverage above 80%
- Monitor developer friction
- Adjust threshold if needed

---

## Reusability

This methodology applies to any CI/CD pipeline:

### Language-Agnostic Patterns
1. **Measure**: Extract metric from tool output
2. **Compare**: Check against threshold
3. **Decide**: Exit 1 if violation, 0 if pass
4. **Report**: Clear error with fix instructions

### Transferable to:
- Python: pytest-cov for coverage
- JavaScript: nyc/istanbul for coverage, ESLint for linting
- Ruby: SimpleCov for coverage, RuboCop for linting
- Rust: cargo-tarpaulin for coverage, clippy for linting

### Adaptation Required:
- Tool names and commands
- Coverage output parsing
- Lint configuration files
- CI platform syntax

---

## References

**Standards**:
- [Keep a Changelog](https://keepachangelog.com/) - CHANGELOG format
- [Semantic Versioning](https://semver.org/) - Version numbering

**Tools**:
- [golangci-lint](https://golangci-lint.run/) - Go linter aggregator
- [Codecov](https://codecov.io/) - Coverage tracking service

**Related Methodologies**:
- docs/methodology/error-recovery.md - Error handling patterns
- experiments/bootstrap-006-api-consistency/iteration-5.md - Quality gate extraction

---

**Version History**:
- 1.0 (2025-10-16): Initial extraction from Bootstrap-007 Iteration 1

---

**Usage**: Copy this methodology when implementing quality gates in new projects, adapting tools and thresholds to project context.
