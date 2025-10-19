# Bootstrap-004: Refactoring Guide

**Status**: 🔄 IN PROGRESS - Iteration 1 Complete
**Started**: 2025-10-18
**Last Updated**: 2025-10-19
**Methodology**: BAIME v2.0 (Bootstrapped AI Methodology Engineering)

---

## Quick Links

- [Detailed Plan](plan.md) - Complete experiment design
- [Iteration Prompts](ITERATION-PROMPTS.md) - Execution templates
- [Iteration Logs](iterations/) - Execution history

---

## Experiment Overview

### Two-Layer Architecture

#### Meta-Objective (Meta-Agent Layer)
Develop **systematic code refactoring methodology** through observation of agent refactoring patterns.

**Goal**: Create reusable refactoring methodology that can be applied to any codebase.

#### Instance Objective (Agent Layer)
Refactor `internal/query/` package to improve code quality.

**Target**: ~500 lines across query engine modules
**Scope**: Reduce cyclomatic complexity by 30%, improve test coverage to 85%
**Deliverables**: Refactored code, improved module structure, enhanced tests

---

## Value Functions

### Instance Value Function (Refactoring Quality)

```
V_instance(s) = 0.3·V_code_quality +     # Quality metrics
                0.3·V_maintainability +  # Maintenance ease
                0.2·V_safety +           # Refactoring safety
                0.2·V_effort             # Efficiency
```

**Components**:

1. **V_code_quality** (0.0-1.0): Code quality improvement
   - Cyclomatic complexity reduction (target: 30%)
   - Code duplication elimination
   - Static analysis issue resolution
   - Naming and structure clarity

2. **V_maintainability** (0.0-1.0): Maintenance ease
   - Test coverage improvement (target: 85%)
   - Module cohesion and coupling
   - Documentation quality
   - Code organization

3. **V_safety** (0.0-1.0): Refactoring safety
   - All tests pass after refactoring
   - No behavior changes (verified)
   - Incremental, reversible steps
   - Version control discipline

4. **V_effort** (0.0-1.0): Efficiency
   - Time per refactoring unit
   - Automated vs manual work ratio
   - Rework needed (lower is better)

### Meta Value Function (Methodology Quality)

```
V_meta(s) = 0.4·V_methodology_completeness +   # Methodology documentation
            0.3·V_methodology_effectiveness +  # Efficiency improvement
            0.3·V_methodology_reusability      # Transferability
```

**Components**:

1. **V_methodology_completeness** (0.0-1.0): Documentation quality
   - Process steps documented
   - Decision criteria defined
   - Examples and edge cases covered
   - Rationale explained

2. **V_methodology_effectiveness** (0.0-1.0): Practical impact
   - Efficiency gain vs ad-hoc (target: 5-10x speedup)
   - Quality improvement (code metrics)
   - Error reduction

3. **V_methodology_reusability** (0.0-1.0): Transferability
   - Cross-language applicability (target: 80%)
   - Domain independence
   - Minimal adaptation needed

---

## Convergence Criteria

The experiment converges when **ALL** conditions are met:

```
CONVERGED iff:
  V_instance(s_N) ≥ 0.80 ∧               # Refactoring quality达标
  V_meta(s_N) ≥ 0.80 ∧                   # Methodology成熟
  M_N == M_{N-1} ∧                       # Meta-Agent stable
  A_N == A_{N-1} ∧                       # Agent set stable
  ΔV_instance < 0.02 (for 2+ iterations) ∧  # Instance value stable
  ΔV_meta < 0.02 (for 2+ iterations)     # Meta value stable
```

**Expected Iterations**: 5-7 (based on medium complexity domain)

---

## Data Sources

### Primary Data
- **High-edit files**: `plan.md` (183 edits), `tools.go` (115 accesses)
- **Code complexity**: `gocyclo`, `dupl` metrics
- **Static analysis**: `staticcheck`, `go vet` output
- **Test coverage**: Current coverage baseline

### Meta-CC Queries
```bash
# File access patterns (refactoring candidates)
meta-cc query-files --threshold 20

# Error-prone edits (need refactoring)
meta-cc query-tools --status error --tool Edit

# High-change areas (technical debt)
meta-cc query-conversation --pattern "fix|bug|issue|problem"
```

---

## Expected Agents

Based on completed experiments and refactoring domain analysis:

- **code-smell-detector**: Identify code quality issues
- **refactoring-planner**: Plan safe refactoring steps
- **safety-checker**: Verify refactoring safety (tests pass, no behavior change)
- **impact-analyzer**: Analyze change impact and risks
- *Additional agents may emerge during execution*

---

## BAIME Framework Application

### Observe-Codify-Automate (OCA) Cycle

**Iteration Structure**:
1. **Observe**: Execute refactoring, collect patterns
2. **Codify**: Document successful patterns as methodology
3. **Automate**: Create tools/scripts for repeated patterns

### Dual Value Tracking

- **V_instance**: Measures concrete refactoring quality (code improvements)
- **V_meta**: Measures methodology development quality (reusable knowledge)

Both must converge to ≥0.80 for experiment success.

### Context Allocation (30/40/20/10)

- **30% Pattern Observation**: Execute refactoring, identify patterns
- **40% Methodology Codification**: Document systematic approach
- **20% Automation**: Create refactoring tools and scripts
- **10% Validation**: Multi-context testing (different codebases)

---

## Current Status

### Iteration 0: Baseline Establishment ✅ COMPLETE

**Completed**: 2025-10-19

**Accomplished**:
1. ✅ Analyzed `internal/query/` package complexity
2. ✅ Established baseline metrics (cyclomatic complexity, test coverage, code duplication)
3. ✅ Identified refactoring targets (5 prioritized targets)
4. ✅ Calculated initial V_instance(s₀) = 0.46 and V_meta(s₀) = 0.06
5. ✅ Documented comprehensive baseline state

**Key Findings**:
- **Test Coverage**: 92.2% (already exceeds 85% target)
- **Cyclomatic Complexity**: Only 1 production function >10 (complexity 11)
- **Code Duplication**: 32 clone groups (3 in production code, 29 in tests)
- **Static Analysis**: Clean (only version warning)
- **Challenge**: Codebase is already high quality, must focus on duplication/complexity/clarity

**Value Functions**:
- **V_instance(s₀) = 0.46** - Moderate baseline (high maintainability, low code quality component due to no reductions yet)
- **V_meta(s₀) = 0.06** - Minimal methodology (expected at start)

**Next**: Iteration 1 - Initial refactoring + pattern observation

### Iteration 1: Initial Refactoring + Pattern Observation ✅ COMPLETE

**Completed**: 2025-10-19
**Duration**: ~90 minutes (60 min refactoring + 30 min documentation)

**Accomplished**:
1. ✅ Eliminated `buildContextBefore` / `buildContextAfter` duplication (18 lines saved)
2. ✅ Simplified `calculateSequenceTimeSpan` complexity (11 → 10)
3. ✅ Pattern observations documented (4 universal patterns identified)
4. ✅ Methodology draft v1 created (25% completeness)
5. ✅ Achieved 1.4x efficiency gain vs. baseline estimate

**Value Functions**:
- **V_instance(s₁) = 0.54** (improved from 0.46, +17%)
- **V_meta(s₁) = 0.37** (improved from 0.06, +517%)

**Key Metrics**:
- Functions >10 complexity: 5 → 4 (-20%)
- Production clone groups: 3 → 2 (-33%)
- Test pass rate: 100% (maintained)
- Coverage: 92.0% (stable)

**Next**: Iteration 2 - Methodology codification + more refactoring

### Iteration 2: Methodology Codification + More Refactoring

**Status**: 🔜 READY TO START

**Planned Targets**:
1. Extract Sequence Pattern Builder (sequences.go duplication)
2. Extract Magic Number Constants
3. (Optional) Improve naming clarity

**Expected Outcomes**:
- Methodology completeness: 40% → 60-65%
- Code smell catalog: 2 → 5 smells
- Refactoring techniques: 2 → 5-7 techniques
- V_instance(s₂) ≈ 0.65-0.70
- V_meta(s₂) ≈ 0.55-0.60

---

## Experiment Timeline

| Iteration | Focus | Expected Duration |
|-----------|-------|-------------------|
| 0 | Baseline establishment | 2-3 hours |
| 1 | Initial refactoring + pattern observation | 3-4 hours |
| 2 | Methodology codification + automation | 3-4 hours |
| 3+ | Refinement until convergence | 2-3 hours each |

**Total Estimated**: 15-20 hours (5-7 iterations)

---

## Success Criteria

### Instance Layer Success (V_instance ≥ 0.80)
- ✅ Cyclomatic complexity reduced by 30%
- ✅ Test coverage ≥ 85%
- ✅ Code duplication eliminated
- ✅ All tests pass
- ✅ Static analysis issues resolved

### Meta Layer Success (V_meta ≥ 0.80)
- ✅ Complete refactoring methodology documented
- ✅ 5-10x efficiency gain demonstrated
- ✅ 80%+ cross-language transferability
- ✅ Automation tools created
- ✅ Multi-context validation completed

---

## Why This Experiment?

1. **Code maintainability is long-term critical**: Technical debt accumulates without systematic refactoring
2. **Refactoring requires careful planning**: Unsafe refactoring introduces bugs
3. **High-frequency edits indicate refactoring needs**: Data-driven target selection
4. **Methodology applicable to any codebase**: 80%+ reusability expected

---

## References

**Completed BAIME Experiments**:
- [Bootstrap-002: Test Strategy](../bootstrap-002-test-strategy/) - 6 iterations, V_instance=0.80, V_meta=0.80
- [Bootstrap-003: Error Recovery](../bootstrap-003-error-recovery/) - 3 iterations, V_instance=0.83, V_meta=0.85

**Methodology Documents**:
- [BAIME Framework](../../docs/methodology/bootstrapped-software-engineering.md)
- [OCA Cycle](../../docs/methodology/empirical-methodology-development.md)
- [Value Functions](../../docs/methodology/value-space-optimization.md)

**Experiment Overview**:
- [EXPERIMENTS-OVERVIEW.md](../EXPERIMENTS-OVERVIEW.md)

---

**Document Version**: 1.0
**Created**: 2025-10-18
**Last Updated**: 2025-10-18
