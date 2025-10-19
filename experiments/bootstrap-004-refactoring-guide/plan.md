# Bootstrap-004: Refactoring Guide - Detailed Plan

**Created**: 2025-10-18
**Status**: 🔄 IN PROGRESS - Iteration 0
**Methodology**: BAIME v2.0

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Two-Layer Architecture](#two-layer-architecture)
3. [Value Functions](#value-functions)
4. [Data Sources](#data-sources)
5. [Iteration Plan](#iteration-plan)
6. [Agent Architecture](#agent-architecture)
7. [Success Criteria](#success-criteria)
8. [Risks and Mitigations](#risks-and-mitigations)

---

## Executive Summary

### Problem Statement

Code refactoring is a critical but risky activity in software development. Without systematic methodology, refactoring can:
- Introduce bugs (behavior changes)
- Take excessive time (inefficient approach)
- Miss critical issues (incomplete coverage)
- Lack safety guarantees (insufficient testing)

**Current State**: Meta-CC has high-edit files (`plan.md`: 183 edits, `tools.go`: 115 accesses) indicating refactoring needs, but lacks systematic refactoring methodology.

### Experiment Goals

**Meta-Objective** (Meta-Agent Layer):
Develop systematic code refactoring methodology through observation of agent refactoring patterns.

**Instance Objective** (Agent Layer):
Refactor `internal/query/` package to improve code quality.

### Expected Outcomes

1. **Instance Layer**: Refactored `internal/query/` with 30% complexity reduction, 85% test coverage
2. **Meta Layer**: Reusable refactoring methodology with 80%+ transferability, 5-10x efficiency gain
3. **Deliverables**: Methodology guide, automation tools, validated patterns

---

## Two-Layer Architecture

### Meta-Agent Layer (Methodology Development)

**Role**: Observe agent refactoring patterns, codify methodology, automate repeated patterns

**Activities**:
1. **Observe**: Watch agents execute refactoring steps
2. **Codify**: Document successful patterns as methodology
3. **Automate**: Create tools for repeated refactoring tasks

**Output**: Refactoring methodology (process + tools + documentation)

### Agent Layer (Concrete Refactoring)

**Role**: Execute refactoring on `internal/query/` package

**Activities**:
1. Analyze code complexity and quality issues
2. Plan safe refactoring steps
3. Execute refactoring incrementally
4. Verify safety (tests pass, no behavior changes)

**Output**: Refactored code with improved quality metrics

---

## Value Functions

### Instance Value Function: V_instance(s)

Measures concrete refactoring quality on `internal/query/` package.

```
V_instance(s) = 0.3·V_code_quality +     # Quality metrics
                0.3·V_maintainability +  # Maintenance ease
                0.2·V_safety +           # Refactoring safety
                0.2·V_effort             # Efficiency
```

#### Component Definitions

##### V_code_quality (0.0-1.0)

**Definition**: Code quality improvement measured by static analysis metrics.

**Measurement**:
```
V_code_quality = 0.4·complexity_reduction +
                 0.3·duplication_reduction +
                 0.2·static_analysis_improvement +
                 0.1·naming_clarity

Where:
  complexity_reduction = (baseline_complexity - current_complexity) / baseline_complexity
    Target: 30% reduction → score 1.0

  duplication_reduction = (baseline_duplication - current_duplication) / baseline_duplication
    Target: 80% reduction → score 1.0

  static_analysis_improvement = (baseline_issues - current_issues) / baseline_issues
    Target: 90% resolution → score 1.0

  naming_clarity = subjective assessment (0.0-1.0)
    0.0-0.3: Unclear names, poor structure
    0.3-0.6: Some improvements, inconsistent
    0.6-0.8: Good naming, mostly clear
    0.8-1.0: Excellent naming, very clear
```

**Tools**:
- `gocyclo` for cyclomatic complexity
- `dupl` for code duplication
- `staticcheck`, `go vet` for static analysis

##### V_maintainability (0.0-1.0)

**Definition**: Long-term maintenance ease.

**Measurement**:
```
V_maintainability = 0.4·test_coverage +
                    0.3·module_cohesion +
                    0.2·documentation_quality +
                    0.1·code_organization

Where:
  test_coverage = current_coverage / 0.85
    Target: 85% → score 1.0

  module_cohesion = cohesion_score (from static analysis)
    High cohesion, low coupling → 1.0
    Low cohesion, high coupling → 0.0

  documentation_quality = (documented_functions / total_functions) · clarity_factor
    All functions documented clearly → 1.0

  code_organization = subjective assessment (0.0-1.0)
    0.0-0.3: Poor organization, hard to navigate
    0.3-0.6: Some structure, but inconsistent
    0.6-0.8: Good organization, logical structure
    0.8-1.0: Excellent organization, very intuitive
```

##### V_safety (0.0-1.0)

**Definition**: Refactoring safety (no bugs introduced).

**Measurement**:
```
V_safety = 0.5·test_pass_rate +
           0.3·behavior_preservation +
           0.2·incremental_discipline

Where:
  test_pass_rate = passing_tests / total_tests
    All tests pass → 1.0

  behavior_preservation = verified through:
    - Integration tests pass
    - Manual verification of edge cases
    - No API contract changes
    1.0 = fully verified, 0.0 = unverified

  incremental_discipline = assessment of refactoring process
    1.0 = small, safe, reversible steps
    0.0 = large, risky, irreversible changes
```

##### V_effort (0.0-1.0)

**Definition**: Refactoring efficiency.

**Measurement**:
```
V_effort = 1.0 - (actual_time / expected_time)

Where:
  expected_time = baseline_time (ad-hoc refactoring)
  actual_time = systematic_refactoring_time

  Score interpretation:
    1.0 = 10x faster than ad-hoc
    0.8 = 5x faster
    0.6 = 2.5x faster
    0.4 = 1.7x faster
    0.0 = same speed as ad-hoc
```

**Baseline**: Estimate ad-hoc refactoring time for `internal/query/` package.

#### Baseline Establishment (Iteration 0)

**Objectives**:
1. Measure current metrics for `internal/query/`:
   - Cyclomatic complexity (per function)
   - Code duplication percentage
   - Static analysis issues count
   - Test coverage percentage
   - Module coupling metrics

2. Set baseline values:
   - V_code_quality(s₀)
   - V_maintainability(s₀)
   - V_safety(s₀) = 1.0 (tests currently pass)
   - V_effort(s₀) = 0.0 (no refactoring yet)

3. Calculate V_instance(s₀)

**Expected V_instance(s₀)**: 0.35-0.45 (medium quality baseline)

### Meta Value Function: V_meta(s)

Measures methodology development quality (universal across experiments).

```
V_meta(s) = 0.4·V_methodology_completeness +   # Documentation completeness
            0.3·V_methodology_effectiveness +  # Practical effectiveness
            0.3·V_methodology_reusability      # Cross-domain transferability
```

#### Component Definitions

##### V_methodology_completeness (0.0-1.0)

**Definition**: How completely is the refactoring methodology documented?

**Measurement Rubric**:

| Score | Level | Criteria |
|-------|-------|----------|
| 0.0-0.3 | **Basic** | Observational notes only, no structured process |
| 0.3-0.6 | **Structured** | Step-by-step procedures, but missing decision criteria |
| 0.6-0.8 | **Comprehensive** | Complete workflow + decision criteria, but lacking examples/edge cases |
| 0.8-1.0 | **Fully Codified** | Complete documentation: process + criteria + examples + edge cases + rationale |

**Checklist** (each item = ~0.067 points):
- [ ] Refactoring process steps documented
- [ ] Code smell detection criteria defined
- [ ] Refactoring technique catalog created
- [ ] Safety verification procedures documented
- [ ] Risk assessment framework defined
- [ ] Examples for each refactoring type provided
- [ ] Edge cases and failure modes documented
- [ ] Decision trees for refactoring choices
- [ ] Rollback procedures documented
- [ ] Testing strategy for refactoring defined
- [ ] Automation opportunities identified
- [ ] Tool usage guidelines created
- [ ] Cross-language adaptation notes
- [ ] Common pitfalls documented
- [ ] Success patterns identified

##### V_methodology_effectiveness (0.0-1.0)

**Definition**: How much does the methodology improve outcomes vs. ad-hoc refactoring?

**Measurement**:
```
V_effectiveness = 0.5·efficiency_gain + 0.5·quality_improvement

Where:
  efficiency_gain based on speedup:
    <2x speedup   → 0.0-0.3
    2-5x speedup  → 0.3-0.6
    5-10x speedup → 0.6-0.8
    >10x speedup  → 0.8-1.0

  quality_improvement based on:
    - Error rate reduction (bugs introduced)
    - Code quality metric improvements
    - Safety guarantee improvements

    <10% improvement  → 0.0-0.3
    10-20% improvement → 0.3-0.6
    20-50% improvement → 0.6-0.8
    >50% improvement   → 0.8-1.0
```

**Validation**:
- Compare systematic vs ad-hoc refactoring time (estimate or actual)
- Measure quality metric deltas
- Track bug introduction rate

##### V_methodology_reusability (0.0-1.0)

**Definition**: How easily can the methodology transfer to other domains/languages?

**Measurement Rubric**:

| Score | Level | Transfer Effort | Modification Needed |
|-------|-------|----------------|---------------------|
| 0.0-0.3 | **Domain-Specific** | High effort | >70% modification |
| 0.3-0.6 | **Partially Portable** | Moderate effort | 40-70% modification |
| 0.6-0.8 | **Largely Portable** | Low effort | 15-40% modification |
| 0.8-1.0 | **Highly Portable** | Minimal effort | <15% modification |

**Assessment**:
- Estimate % of methodology that is language-agnostic
- Identify Go-specific vs universal patterns
- Test transfer to Python/Rust/Java (simulation)

**Target**: 80%+ reusability (score ≥ 0.80)

#### Baseline Establishment (Iteration 0)

**Expected V_meta(s₀)**: 0.15-0.25 (minimal methodology at start)

**Components**:
- V_methodology_completeness(s₀) ≈ 0.15 (basic observational notes)
- V_methodology_effectiveness(s₀) ≈ 0.0 (no methodology yet)
- V_methodology_reusability(s₀) ≈ 0.0 (nothing to reuse yet)

---

## Data Sources

### Primary Code Targets

**`internal/query/` Package Analysis**:
```bash
# Files to refactor
internal/query/executor.go
internal/query/filters.go
internal/query/parser.go
internal/query/tools.go
internal/query/conversation.go
```

### Complexity Analysis

**Tools and Commands**:
```bash
# Cyclomatic complexity
gocyclo -over 10 internal/query/

# Code duplication
dupl -threshold 15 internal/query/

# Static analysis
staticcheck ./internal/query/...
go vet ./internal/query/...

# Test coverage
go test -coverprofile=coverage.out ./internal/query/...
go tool cover -func=coverage.out
```

### Meta-CC Queries

**Refactoring Target Identification**:
```bash
# High-change files (likely need refactoring)
meta-cc query-files --threshold 20

# Error-prone edits (quality issues)
meta-cc query-tools --status error --tool Edit

# Bug-related conversations (problem areas)
meta-cc query-conversation --pattern "fix|bug|issue|problem"

# File access history for internal/query/
meta-cc query-file-access --file internal/query/executor.go
```

### Historical Data

From EXPERIMENTS-OVERVIEW.md:
- `plan.md`: 183 edits (documentation churn)
- `tools.go`: 115 accesses (high-touch file)
- Edit operations: 2,476 total (refactoring opportunity)
- Error rate: 6.06% (quality improvement needed)

---

## Iteration Plan

### Iteration 0: Baseline Establishment

**Duration**: 2-3 hours

**Objectives**:
1. Analyze `internal/query/` package complexity
2. Establish baseline metrics
3. Identify refactoring targets
4. Calculate V_instance(s₀) and V_meta(s₀)
5. Document baseline state

**Tasks**:
- [ ] Run complexity analysis tools
- [ ] Measure test coverage
- [ ] Identify code smells
- [ ] Prioritize refactoring targets
- [ ] Calculate baseline value functions
- [ ] Document findings in `iterations/iteration-0.md`

**Expected Outputs**:
- Baseline metrics report
- Refactoring target list (prioritized)
- Initial value assessment: V_instance(s₀) ≈ 0.40, V_meta(s₀) ≈ 0.20

### Iteration 1: Initial Refactoring + Pattern Observation

**Duration**: 3-4 hours

**Objectives**:
1. Execute first refactoring (highest priority target)
2. Observe refactoring patterns
3. Document successful techniques
4. Begin methodology codification

**BAIME Phase**: Observe (70%), Codify (20%), Automate (10%)

**Tasks**:
- [ ] Select highest priority refactoring target
- [ ] Plan refactoring steps (incremental)
- [ ] Execute refactoring (with tests)
- [ ] Verify safety (all tests pass)
- [ ] Measure new complexity/coverage metrics
- [ ] Document patterns observed
- [ ] Begin methodology draft

**Expected Outputs**:
- Refactored module (1-2 files)
- Pattern observation notes
- Initial methodology draft (process steps)
- V_instance(s₁) ≈ 0.50-0.55, V_meta(s₁) ≈ 0.35-0.45

### Iteration 2: Methodology Codification + More Refactoring

**Duration**: 3-4 hours

**Objectives**:
1. Continue refactoring (2-3 more targets)
2. Codify emerging patterns
3. Create decision framework
4. Refine methodology documentation

**BAIME Phase**: Observe (30%), Codify (50%), Automate (20%)

**Tasks**:
- [ ] Execute 2-3 more refactoring tasks
- [ ] Identify common patterns across refactorings
- [ ] Create refactoring decision tree
- [ ] Document code smell catalog
- [ ] Define safety verification checklist
- [ ] Draft tool automation opportunities

**Expected Outputs**:
- Additional refactored modules (2-3 files)
- Refactoring methodology guide (draft)
- Code smell catalog
- Decision framework
- V_instance(s₂) ≈ 0.60-0.65, V_meta(s₂) ≈ 0.55-0.65

### Iteration 3: Automation Introduction

**Duration**: 3-4 hours

**Objectives**:
1. Create automation tools for repeated patterns
2. Complete remaining refactoring
3. Finalize methodology documentation
4. Begin multi-context validation planning

**BAIME Phase**: Observe (20%), Codify (30%), Automate (50%)

**Tasks**:
- [ ] Create code smell detection script
- [ ] Create refactoring safety checker
- [ ] Create complexity reporter
- [ ] Complete refactoring of remaining files
- [ ] Finalize methodology documentation
- [ ] Plan multi-context validation

**Expected Outputs**:
- Automation tools (3-5 scripts)
- Complete refactoring of `internal/query/`
- Comprehensive methodology guide
- Multi-context validation plan
- V_instance(s₃) ≈ 0.70-0.75, V_meta(s₃) ≈ 0.70-0.75

### Iteration 4: Multi-Context Validation

**Duration**: 2-3 hours

**Objectives**:
1. Apply methodology to different context (different package)
2. Measure transferability
3. Refine methodology based on transfer learnings
4. Assess convergence

**BAIME Phase**: Validation (70%), Refinement (30%)

**Tasks**:
- [ ] Select validation context (e.g., `internal/parser/`)
- [ ] Apply methodology to new context
- [ ] Measure adaptation effort
- [ ] Document transfer challenges
- [ ] Refine methodology for universality
- [ ] Calculate final value functions

**Expected Outputs**:
- Validation refactoring (different package)
- Transferability assessment
- Refined methodology
- Convergence assessment
- V_instance(s₄) ≈ 0.80-0.85, V_meta(s₄) ≈ 0.75-0.80

### Iteration 5+: Convergence Attempts

**Duration**: 2-3 hours each

**Objectives**:
1. Achieve V_instance ≥ 0.80 and V_meta ≥ 0.80
2. Stabilize Meta-Agent and Agent set
3. Finalize all deliverables

**Convergence Criteria**:
```
CONVERGED iff:
  V_instance(s_N) ≥ 0.80 ∧
  V_meta(s_N) ≥ 0.80 ∧
  M_N == M_{N-1} ∧
  A_N == A_{N-1} ∧
  ΔV_instance < 0.02 (for 2+ iterations) ∧
  ΔV_meta < 0.02 (for 2+ iterations)
```

**Tasks**:
- [ ] Address any remaining gaps
- [ ] Polish methodology documentation
- [ ] Complete automation tools
- [ ] Verify all metrics
- [ ] Create final results report

---

## Agent Architecture

### Meta-Agent M₀ (Stable)

**Capabilities** (from Bootstrap-001, 002, 003):
1. Orchestrate iteration execution
2. Calculate value functions
3. Assess convergence
4. Guide agent creation
5. Document methodology evolution

**Expected Stability**: M₀ sufficient (no evolution needed based on 001, 002, 003)

### Initial Agent Set A₀ (Generic Agents)

Based on completed experiments, start with generic agents:

**General-Purpose Agent**:
- Execute refactoring tasks
- Run analysis tools
- Write documentation
- Implement automation scripts

**Expected Evolution**: Specialized agents likely to emerge in Iteration 1-2:

### Expected Specialized Agents (Emerge as Needed)

#### code-smell-detector
**Trigger**: When code analysis becomes repetitive (Iteration 1-2)
**Responsibilities**:
- Identify code smells systematically
- Prioritize issues by severity
- Suggest refactoring techniques

#### refactoring-planner
**Trigger**: When refactoring planning becomes complex (Iteration 1-2)
**Responsibilities**:
- Plan safe refactoring steps
- Create incremental transformation plans
- Estimate refactoring effort

#### safety-checker
**Trigger**: When safety verification becomes critical (Iteration 2)
**Responsibilities**:
- Verify all tests pass
- Check behavior preservation
- Validate incremental steps

#### impact-analyzer
**Trigger**: When change impact analysis needed (Iteration 2-3)
**Responsibilities**:
- Analyze change ripple effects
- Identify affected modules
- Assess risk levels

**Total Expected Agents**: 3-5 (including 3-4 specialized)

**Agent Specialization Ratio**: ~40-60% specialized (based on Bootstrap-001, 003)

---

## Success Criteria

### Instance Layer Success (V_instance ≥ 0.80)

**Code Quality**:
- ✅ Cyclomatic complexity reduced by 30% (from baseline)
- ✅ Code duplication reduced by 80%
- ✅ Static analysis issues reduced by 90%
- ✅ Naming clarity score ≥ 0.80

**Maintainability**:
- ✅ Test coverage ≥ 85%
- ✅ Module cohesion score ≥ 0.80
- ✅ Documentation coverage ≥ 90%
- ✅ Code organization score ≥ 0.80

**Safety**:
- ✅ All tests pass (100% pass rate)
- ✅ No behavior changes (verified)
- ✅ Incremental discipline ≥ 0.80

**Effort**:
- ✅ Efficiency score ≥ 0.60 (2.5x speedup over ad-hoc)

### Meta Layer Success (V_meta ≥ 0.80)

**Completeness**:
- ✅ All 15 checklist items documented
- ✅ Complete methodology guide created
- ✅ Examples and edge cases covered

**Effectiveness**:
- ✅ 5-10x efficiency gain demonstrated
- ✅ 20-50% quality improvement shown
- ✅ Safety guarantees validated

**Reusability**:
- ✅ 80%+ cross-language transferability
- ✅ Successful transfer to 2+ contexts
- ✅ <20% adaptation effort

### System Stability

**Agent Set Stability**:
- M_N == M_{N-1} (Meta-Agent unchanged)
- A_N == A_{N-1} (Agent set unchanged)

**Value Stability**:
- ΔV_instance < 0.02 for 2+ consecutive iterations
- ΔV_meta < 0.02 for 2+ consecutive iterations

---

## Risks and Mitigations

### Risk 1: Safety Issues (Bugs Introduced)

**Probability**: MEDIUM
**Impact**: HIGH
**Mitigation**:
- Always run full test suite after each refactoring step
- Maintain incremental commits (easy rollback)
- Use behavior preservation tests
- Manual verification of edge cases

### Risk 2: Scope Creep (Too Many Files)

**Probability**: MEDIUM
**Impact**: MEDIUM
**Mitigation**:
- Limit scope to `internal/query/` package (~500 lines)
- Prioritize high-impact refactorings
- Use 30/40/20/10 context allocation discipline

### Risk 3: Methodology Over-Specificity (Low Transferability)

**Probability**: LOW (based on Bootstrap-001, 002, 003)
**Impact**: HIGH
**Mitigation**:
- Multi-context validation (Iteration 4)
- Document universal vs Go-specific patterns
- Test transfer to simulated Python/Rust contexts

### Risk 4: Convergence Delay (>7 Iterations)

**Probability**: LOW-MEDIUM
**Impact**: MEDIUM
**Mitigation**:
- Clear baseline metrics guide progress
- Value function feedback loop
- Learn from Bootstrap-003's rapid convergence (3 iterations)

### Risk 5: Test Coverage Challenges

**Probability**: LOW
**Impact**: MEDIUM
**Mitigation**:
- Leverage Bootstrap-002 test strategy methodology
- Create tests before refactoring (TDD approach)
- Use coverage-driven gap closure

---

## Deliverables

### Instance Layer Deliverables

1. **Refactored Code**:
   - `internal/query/` package fully refactored
   - All tests passing
   - Metrics improved (complexity, coverage, duplication)

2. **Test Suite**:
   - Test coverage ≥ 85%
   - Behavior preservation tests
   - Edge case coverage

3. **Documentation**:
   - Updated module documentation
   - Code comments for complex logic
   - Architecture diagrams (if needed)

### Meta Layer Deliverables

1. **Refactoring Methodology Guide**:
   - Process steps (detect → plan → execute → verify)
   - Decision framework (when to refactor, which technique)
   - Code smell catalog (identification + resolution)
   - Safety verification procedures

2. **Automation Tools**:
   - Code smell detector script
   - Refactoring safety checker
   - Complexity reporter
   - Refactoring planner helper

3. **Validation Evidence**:
   - Multi-context application results
   - Transferability assessment
   - Efficiency gain measurements
   - Quality improvement metrics

4. **Comprehensive Documentation**:
   - Methodology.md (complete refactoring guide)
   - Examples for each refactoring type
   - Common pitfalls and solutions
   - Cross-language adaptation notes

---

## Timeline and Milestones

### Expected Timeline

| Week | Iterations | Milestones |
|------|------------|------------|
| 1 | 0-1 | Baseline + Initial refactoring |
| 2 | 2-3 | Methodology codification + Automation |
| 3 | 4-5+ | Multi-context validation + Convergence |

**Total Duration**: 15-20 hours (5-7 iterations)

### Key Milestones

- **M1** (End of Iteration 0): Baseline established, V_instance(s₀) ≈ 0.40, V_meta(s₀) ≈ 0.20
- **M2** (End of Iteration 1): First refactoring complete, patterns observed
- **M3** (End of Iteration 2): Methodology draft complete, V_meta(s₂) ≈ 0.60
- **M4** (End of Iteration 3): Automation tools created, V_instance(s₃) ≈ 0.75
- **M5** (End of Iteration 4): Multi-context validation complete
- **M6** (End of Iteration 5+): Convergence achieved, V_instance ≥ 0.80, V_meta ≥ 0.80

---

## References

### Completed BAIME Experiments

**Bootstrap-002: Test Strategy Development**:
- 6 iterations, V_instance=0.80, V_meta=0.80
- 3.1x average speedup, 94.2% reusability
- Generic agents sufficient (0 specialized)
- [Link](../bootstrap-002-test-strategy/)

**Bootstrap-003: Error Recovery Methodology**:
- 3 iterations, V_instance=0.83, V_meta=0.85
- 2x faster than Bootstrap-002 (clear metrics → rapid convergence)
- 23.7% error reduction, 20.9x speedup
- Generic agents sufficient (0 specialized)
- [Link](../bootstrap-003-error-recovery/)

**Bootstrap-001: Documentation Methodology**:
- 3 iterations, V_instance=0.808
- 85% reusability
- 2 specialized agents (40% ratio)
- [Link](../bootstrap-001-doc-methodology/)

### Methodology Documents

- [BAIME Framework](../../docs/methodology/bootstrapped-software-engineering.md)
- [OCA Cycle](../../docs/methodology/empirical-methodology-development.md)
- [Value Functions](../../docs/methodology/value-space-optimization.md)
- [EXPERIMENTS-OVERVIEW.md](../EXPERIMENTS-OVERVIEW.md)

---

**Document Version**: 1.0
**Created**: 2025-10-18
**Last Updated**: 2025-10-18
