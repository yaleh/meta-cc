# Technical Debt Prevention Guidelines - Iteration 3

**Created**: 2025-10-17
**Purpose**: Prevent new technical debt accumulation through proactive practices

---

## Prevention Strategy Framework

### 1. Pre-Commit Prevention

**Complexity Budget Enforcement**:
```yaml
complexity_gates:
  function_complexity:
    warning: 10
    error: 15
    action: "Refuse commit if >15"

  file_complexity:
    warning: 150
    error: 200
    action: "Block PR if >200"
```

**Implementation**:
- Git pre-commit hook: Run `gocyclo -over 15 .`
- CI/CD check: Fail build if complexity threshold exceeded
- PR review: Flag functions >10 complexity for review

**Prevention Value**: Stops bloaters before they enter codebase

---

### 2. Test Coverage Gates

**Coverage Requirements**:
```yaml
coverage_gates:
  overall:
    minimum: 80%
    target: 85%
    action: "Block PR if <80%"

  new_code:
    minimum: 90%
    action: "New code must have 90% coverage"

  critical_paths:
    minimum: 95%
    modules: ["cmd/", "internal/mcp/"]
```

**Implementation**:
- Pre-push hook: Run `go test -coverprofile=coverage.out ./...`
- CI/CD: Fail if coverage drops below 80%
- Code review: Require tests for new functions

**Prevention Value**: Prevents reliability debt

---

### 3. Static Analysis Enforcement

**Zero-Tolerance Policy**:
```yaml
static_analysis:
  staticcheck:
    errors: 0
    warnings: 0
    action: "Fail build on any issue"

  go_vet:
    errors: 0
    action: "Fail build on vet errors"
```

**Implementation**:
- Pre-commit: Run `staticcheck ./...` and `go vet ./...`
- CI/CD: Lint check in pipeline
- Auto-fix: Run `gofmt`, `goimports` on save

**Prevention Value**: Prevents style and reliability debt

---

### 4. Code Review Checklist

**Debt Prevention Review**:
- [ ] Function complexity <15 (check with gocyclo)
- [ ] No duplicate code blocks >50 tokens
- [ ] Test coverage ≥90% for new code
- [ ] No staticcheck warnings
- [ ] Clear function responsibilities (SRP)
- [ ] No shotgun surgery patterns (logic in multiple files)

**Reviewer Guidance**:
- Request refactoring if complexity >10
- Suggest abstraction if duplication detected
- Require tests before approval
- Check for architectural debt (coupling, shotgun surgery)

**Prevention Value**: Human oversight for architectural debt

---

### 5. Refactoring Time Budget

**Continuous Paydown**:
```yaml
debt_paydown_budget:
  per_sprint:
    percentage: 20
    description: "20% of sprint capacity for debt paydown"

  trigger:
    td_ratio_threshold: 12%
    action: "Dedicate sprint to debt reduction if >12%"
```

**Implementation**:
- Track TD ratio monthly
- Allocate 1 day per week for refactoring
- Prioritize high-value low-effort debt

**Prevention Value**: Prevents debt accumulation over time

---

### 6. Architecture Review

**Periodic Health Checks**:
```yaml
architecture_review:
  frequency: "Quarterly"
  focus:
    - "Identify shotgun surgery patterns"
    - "Review module coupling"
    - "Assess code smell trends"

  actions:
    - "Plan architectural refactorings"
    - "Update complexity budgets"
    - "Adjust prevention thresholds"
```

**Prevention Value**: Prevents architectural debt

---

## Prevention Best Practices

### For Developers

1. **Write Tests First (TDD)**: Prevents coverage debt
2. **Refactor Continuously**: Boy Scout Rule (leave code cleaner than found)
3. **Keep Functions Small**: Target <30 lines, complexity <10
4. **Avoid Copy-Paste**: Extract reusable functions
5. **Review Own Code**: Use `gocyclo` before committing

### For Teams

1. **Establish Complexity Budgets**: Team agreement on thresholds
2. **Automate Prevention**: Git hooks, CI/CD gates
3. **Track Debt Trends**: Monthly TD ratio tracking
4. **Celebrate Paydown**: Recognize debt reduction contributions
5. **Share Knowledge**: Code review as learning opportunity

### For Projects

1. **Set Quality Gates**: Automated checks in pipeline
2. **Allocate Paydown Time**: Budget for continuous refactoring
3. **Monitor Trends**: Dashboard showing TD ratio over time
4. **Plan Strategically**: Architectural reviews quarterly
5. **Document Standards**: Complexity guidelines in CONTRIBUTING.md

---

## Implementation Roadmap

### Phase 1: Immediate (1 day)
- Install pre-commit hooks (complexity, coverage, staticcheck)
- Add CI/CD quality gates
- Document prevention guidelines

### Phase 2: Short-term (1 week)
- Establish complexity budgets
- Create code review checklist
- Set up debt tracking dashboard

### Phase 3: Medium-term (1 month)
- Allocate refactoring time budget
- Conduct first architecture review
- Train team on prevention practices

### Phase 4: Long-term (Ongoing)
- Monthly TD ratio tracking
- Quarterly architecture reviews
- Continuous process improvement

---

## Expected Impact

**Baseline** (current):
- TD ratio: 15.52%
- New debt accumulation: ~2% per quarter

**With Prevention** (projected):
- TD ratio: <10% maintained
- New debt accumulation: <0.5% per quarter
- Net debt reduction: Yes (paydown > accumulation)

**ROI**:
- Prevention time: 1 day setup + 1 day/month monitoring
- Paydown time saved: 5+ days per quarter
- Net savings: 4 days per quarter
