# Convergence Criteria

**How to know when your methodology development is complete.**

## Standard Dual Convergence

The most common pattern (used in 6/8 experiments):

### Criteria

```
Converged when ALL of:
1. M_n == M_{n-1}        (Meta-Agent stable)
2. A_n == A_{n-1}        (Agent set stable)
3. V_instance(s_n) ≥ 0.80
4. V_meta(s_n) ≥ 0.80
5. Objectives complete
6. ΔV < 0.02 for 2+ iterations (diminishing returns)
```

### Example: Bootstrap-009 (Observability)

```
Iteration 6:
  V_instance(s₆) = 0.87 (target: 0.80) ✅
  V_meta(s₆) = 0.83 (target: 0.80) ✅
  M₆ == M₅ ✅
  A₆ == A₅ ✅
  Objectives: All 3 pillars implemented ✅
  ΔV: 0.01 (< 0.02) ✅

→ CONVERGED (Standard Dual Convergence)
```

**Use when**: Both task and methodology are equally important.

---

## Meta-Focused Convergence

Alternative pattern when methodology is primary goal (used in 1/8 experiments):

### Criteria

```
Converged when ALL of:
1. M_n == M_{n-1}        (Meta-Agent stable)
2. A_n == A_{n-1}        (Agent set stable)
3. V_meta(s_n) ≥ 0.80    (Methodology excellent)
4. V_instance(s_n) ≥ 0.55 (Instance practically sufficient)
5. Instance gap is infrastructure, NOT methodology
6. System stable for 2+ iterations
```

### Example: Bootstrap-011 (Knowledge Transfer)

```
Iteration 3:
  V_instance(s₃) = 0.585 (practically sufficient)
  V_meta(s₃) = 0.877 (excellent, +9.6% above target) ✅
  M₃ == M₂ == M₁ ✅
  A₃ == A₂ == A₁ ✅

  Instance gap analysis:
  - Missing: Knowledge graph, semantic search (infrastructure)
  - Present: ALL 3 learning paths complete (methodology)
  - Value: 3-8x onboarding speedup already achieved

  Meta convergence:
  - Completeness: 0.80 (all templates complete)
  - Effectiveness: 0.95 (3-8x validated)
  - Reusability: 0.88 (95%+ transferable)

→ CONVERGED (Meta-Focused Convergence)
```

**Use when**:
- Experiment explicitly prioritizes meta-objective
- Instance gap is tooling/infrastructure, not methodology
- Methodology has reached complete transferability (≥90%)
- Further instance work would not improve methodology quality

**Validation checklist**:
- [ ] Primary objective is methodology (stated in README)
- [ ] Instance gap is infrastructure (not methodology gaps)
- [ ] V_meta_reusability ≥ 0.90
- [ ] Practical value delivered (speedup demonstrated)

---

## Practical Convergence

Alternative pattern when quality exceeds metrics (used in 1/8 experiments):

### Criteria

```
Converged when ALL of:
1. M_n == M_{n-1}        (Meta-Agent stable)
2. A_n == A_{n-1}        (Agent set stable)
3. V_instance + V_meta ≥ 1.60 (combined threshold)
4. Quality evidence exceeds raw metric scores
5. Justified partial criteria
6. ΔV < 0.02 for 2+ iterations
```

### Example: Bootstrap-002 (Testing)

```
Iteration 5:
  V_instance(s₅) = 0.848 (target: 0.80, +6% margin) ✅
  V_meta(s₅) ≈ 0.85 (estimated)
  Combined: 1.698 (> 1.60) ✅

  Quality evidence:
  - Coverage: 75% overall BUT 86-94% in core packages
  - Sub-package excellence > aggregate metric
  - Quality gates: 8/10 met consistently
  - Test quality: Fixtures, mocks, zero flaky tests
  - 15x speedup validated
  - 89% methodology reusability

  M₅ == M₄ ✅
  A₅ == A₄ ✅
  ΔV: 0.01 (< 0.02) ✅

→ CONVERGED (Practical Convergence)
```

**Use when**:
- Some components don't reach target but overall quality is excellent
- Sub-system excellence compensates for aggregate metrics
- Diminishing returns demonstrated
- Honest assessment shows methodology complete

**Validation checklist**:
- [ ] Combined V_instance + V_meta ≥ 1.60
- [ ] Quality evidence documented (not just metrics)
- [ ] Honest gap analysis (no inflation)
- [ ] Diminishing returns proven (ΔV trend)

---

## System Stability

All convergence patterns require system stability:

### Agent Set Stability (A_n == A_{n-1})

**Stable when**:
- Same agents used in iteration n and n-1
- No new specialized agents created
- No agent capabilities expanded

**Example**:
```
Iteration 5: {coder, doc-writer, data-analyst, log-analyzer}
Iteration 6: {coder, doc-writer, data-analyst, log-analyzer}
→ A₆ == A₅ ✅ STABLE
```

### Meta-Agent Stability (M_n == M_{n-1})

**Stable when**:
- Same 5 capabilities in iteration n and n-1
- No new coordination patterns
- No Meta-Agent prompt evolution

**Standard M₀ capabilities**:
1. observe - Pattern observation
2. plan - Iteration planning
3. execute - Agent orchestration
4. reflect - Value assessment
5. evolve - System evolution

**Finding**: M₀ was sufficient in ALL 8 experiments (no evolution needed)

---

## Diminishing Returns

**Definition**: ΔV < epsilon for k consecutive iterations

**Standard threshold**: epsilon = 0.02, k = 2

**Calculation**:
```
ΔV_n = |V_total(s_n) - V_total(s_{n-1})|

If ΔV_n < 0.02 AND ΔV_{n-1} < 0.02:
  → Diminishing returns detected
```

**Example**:
```
Iteration 4: V_total = 0.82, ΔV = 0.05 (significant)
Iteration 5: V_total = 0.84, ΔV = 0.02 (small)
Iteration 6: V_total = 0.85, ΔV = 0.01 (small)
→ Diminishing returns since Iteration 5
```

**Interpretation**:
- Large ΔV (>0.05): Significant progress, continue
- Medium ΔV (0.02-0.05): Steady progress, continue
- Small ΔV (<0.02): Diminishing returns, consider converging

---

## Decision Tree

```
Start with iteration n:

1. Calculate V_instance(s_n) and V_meta(s_n)

2. Check system stability:
   M_n == M_{n-1}? → YES/NO
   A_n == A_{n-1}? → YES/NO

   If NO to either → Continue iteration n+1

3. Check convergence pattern:

   Pattern A: Standard Dual Convergence
   ├─ V_instance ≥ 0.80? → YES
   ├─ V_meta ≥ 0.80? → YES
   ├─ Objectives complete? → YES
   ├─ ΔV < 0.02 for 2 iterations? → YES
   └─→ CONVERGED ✅

   Pattern B: Meta-Focused Convergence
   ├─ V_meta ≥ 0.80? → YES
   ├─ V_instance ≥ 0.55? → YES
   ├─ Primary objective is methodology? → YES
   ├─ Instance gap is infrastructure? → YES
   ├─ V_meta_reusability ≥ 0.90? → YES
   └─→ CONVERGED ✅

   Pattern C: Practical Convergence
   ├─ V_instance + V_meta ≥ 1.60? → YES
   ├─ Quality evidence strong? → YES
   ├─ Justified partial criteria? → YES
   ├─ ΔV < 0.02 for 2 iterations? → YES
   └─→ CONVERGED ✅

4. If no pattern matches → Continue iteration n+1
```

---

## Common Mistakes

### Mistake 1: Premature Convergence

**Symptom**: Declaring convergence before system stable

**Example**:
```
Iteration 3:
  V_instance = 0.82 ✅
  V_meta = 0.81 ✅
  BUT M₃ ≠ M₂ (new Meta-Agent capability added)

→ NOT CONVERGED (system unstable)
```

**Fix**: Wait until M_n == M_{n-1} and A_n == A_{n-1}

### Mistake 2: Inflated Values

**Symptom**: V scores mysteriously jump to exactly 0.80

**Example**:
```
Iteration 4: V_instance = 0.77
Iteration 5: V_instance = 0.80 (claimed)
BUT no substantial work done!
```

**Fix**: Honest assessment, gap enumeration, evidence-based scoring

### Mistake 3: Moving Goalposts

**Symptom**: Changing criteria mid-experiment

**Example**:
```
Initial plan: V_instance ≥ 0.80
Final state: V_instance = 0.65
Conclusion: "Actually, 0.65 is sufficient" ❌ WRONG
```

**Fix**: Either reach 0.80 OR use Meta-Focused/Practical with explicit justification

### Mistake 4: Ignoring System Instability

**Symptom**: Declaring convergence while agents still evolving

**Example**:
```
Iteration 5:
  Both V scores ≥ 0.80 ✅
  BUT new specialized agent created in Iteration 5
  A₅ ≠ A₄

→ NOT CONVERGED (agent set unstable)
```

**Fix**: Run Iteration 6 to confirm A₆ == A₅

---

## Convergence Prediction

Based on 8 experiments, you can predict iteration count:

**Base estimate**: 5 iterations

**Adjustments**:
- Well-defined domain: -2 iterations
- Existing tools available: -1 iteration
- High interdependency: +2 iterations
- Novel patterns needed: +1 iteration
- Large codebase scope: +1 iteration
- Multiple competing goals: +1 iteration

**Examples**:
- Dependency Health: 5 - 2 - 1 = 2 → actual 3 ✓
- Observability: 5 + 0 + 1 = 6 → actual 6 ✓
- Cross-Cutting: 5 + 2 + 1 = 8 → actual 8 ✓

---

**Next**: Read [dual-value-functions.md](dual-value-functions.md) for V_instance and V_meta calculation.
