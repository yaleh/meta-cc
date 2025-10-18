# Testing Strategy: Minimal Baseline Example

**Experiment**: bootstrap-002-test-strategy
**Baseline Investment**: 60 min
**V_meta(s₀)**: 0.04 (Poor)
**Result**: 6 iterations (standard convergence)

---

## Activities (60 min)

### 1. Coverage Measurement (30 min)

```bash
go test -cover ./...
# Result: 72.1% coverage, 590 tests
```

### 2. Ad-hoc Testing (20 min)

Wrote 3 tests manually, noted duplication issues

### 3. No Prior Art Research (0 min)

Started from scratch

### 4. Vague Automation Ideas (10 min)

"Maybe coverage analysis tools..." (not concrete)

---

## V_meta(s₀) Calculation

```
Completeness: 0/8 = 0.00 (no patterns documented)
Transferability: 0/8 = 0.00 (no research)
Automation: 0/3 = 0.00 (not identified)

V_meta(s₀) = 0.4×0.00 + 0.3×0.00 + 0.3×0.00 = 0.00
```

---

## Outcome

- Iterations: 6 (standard convergence)
- Total time: 25.5 hours
- Patterns emerged gradually over 6 iterations

---

## What Could Have Been Different

**If invested 2 more hours in iteration 0**:
- Research test patterns (borrow 5-6)
- Analyze codebase for test opportunities
- Identify coverage tools

**Estimated result**:
- V_meta(s₀) = 0.30-0.40
- 4-5 iterations (vs 6)
- Time saved: 3-6 hours

**ROI**: 2-3x

---

**Source**: Bootstrap-002, minimal baseline approach
