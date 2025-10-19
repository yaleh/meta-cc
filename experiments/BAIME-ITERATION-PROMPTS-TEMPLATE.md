# BAIME Iteration Execution Prompts Template

**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Experiment**: Bootstrap-NNN: [Domain] Methodology
**Meta-Agent**: M₀ (5 capabilities: observe, plan, execute, reflect, evolve)

---

## Overview

This document provides lambda-expression-based execution guidance for the iteration-executor agent. Each iteration follows the BAIME OCA cycle (Observe → Codify → Automate) with systematic value evaluation and convergence checking.

**Usage**: Call iteration-executor agent with this file as guidance for each iteration (0, 1, 2, ..., N) until convergence.

---

## Iteration Execution Protocol

λ(iteration_n, s_{n-1}) → (s_n, V(s_n), convergence_status):

```
iteration_cycle :: (M_{n-1}, A_{n-1}, s_{n-1}) → (M_n, A_n, s_n, V(s_n))
iteration_cycle(M, A, s) =
  pre_execution(s_{n-1}) →
  meta_agent_coordination(M) →
  observe_phase(A) →
  codify_phase(A) →
  automate_phase(A) →
  evaluate_phase(V) →
  convergence_check(V, M, A) →
  if converged then finalize else plan_next_iteration
```

---

## Pre-Execution Protocol

**Before each iteration**:

```
pre_execution :: State_{n-1} → Context
pre_execution(s) =
  read(iteration-{n-1}.md) ∧
  extract(V_instance, V_meta, gaps, learnings) ∧
  load(meta-agents/meta-agent-m0.md) ∧
  load(agents/*.md | if exists) ∧
  identify(focus_areas, priorities)
```

**Checklist**:
- [ ] Read previous iteration file (`iteration-{n-1}.md`)
- [ ] Review previous V_instance and V_meta scores
- [ ] Identify gaps from previous iteration
- [ ] Load Meta-Agent M₀ definition
- [ ] Load any existing agent definitions
- [ ] Determine focus for this iteration

---

## Meta-Agent M₀ Coordination

**Capabilities** (5 modular, always re-read from files):

1. **observe**: Pattern observation and data collection
2. **plan**: Iteration planning and objective setting
3. **execute**: Agent orchestration and task coordination
4. **reflect**: Value assessment and gap identification
5. **evolve**: System evolution decisions

**Coordination Pattern**:

```
meta_agent_protocol :: Capability → Execution
meta_agent_protocol(cap) =
  read(meta-agents/observe.md | plan.md | execute.md | reflect.md | evolve.md) ∧
  apply(guidance) ∧
  coordinate(agents) ∧
  ¬assume ∧ ¬cache
```

**Critical**: Always read capability files fresh, never cache instructions.

---

## Phase 1: OBSERVE (Data Collection)

**Objective**: Gather empirical data about current [domain] state and gaps

### Observation Tasks

```
observe_phase :: Agents → Observations
observe_phase(A) = sequential_execution(
  [measurement_task_1],
  [gap_identification_task],
  [pattern_analysis_task],
  [quality_assessment_task]
)
```

### Specific Actions

**1. Measure Current [Metric] State**

```bash
# [Domain-specific measurement commands]
[command 1]
[command 2] > data/[metric]-summary-iteration-{n}.txt

# [Per-component analysis]
[command 3]
[command 4]
```

**Save outputs** to `data/[metric]-*.txt`

**2. Identify High-Value [Domain] Targets**

```bash
# Files with high change frequency (need [domain work])
meta-cc query-files --threshold 10 > data/high-change-files-iteration-{n}.jsonl

# Files with error patterns (need defensive [domain work])
meta-cc query-tools --status error --tool Edit > data/error-prone-files-iteration-{n}.jsonl

# [Domain]-related conversations (understand pain points)
meta-cc query-user-messages --pattern "[keyword1]|[keyword2]|[keyword3]" \
  > data/[domain]-conversations-iteration-{n}.jsonl
```

**3. Analyze Existing [Domain] Patterns**

```bash
# Find existing [domain] artifacts
find . -name "*[pattern]*" -type f > data/existing-[artifacts]-iteration-{n}.txt

# Analyze [domain] structure
[analysis command 1]
[analysis command 2]
[analysis command 3]
```

**4. Identify [Domain] Gaps**

Use data-analyst agent or manual analysis:
- Which [components] have [metric] <[threshold]?
- Which critical [elements] are [missing/inadequate]?
- Which [error paths / edge cases] lack [domain work]?

**Save** gap analysis to `data/[domain]-gaps-iteration-{n}.md`

### Observation Deliverables

- [ ] [Metric] summary (`data/[metric]-summary-iteration-{n}.txt`)
- [ ] High-value [domain] targets identified
- [ ] Existing [domain] patterns documented
- [ ] [Domain] gaps prioritized
- [ ] Quality issues noted ([specific issues])

---

## Phase 2: CODIFY (Pattern Extraction & Methodology)

**Objective**: Extract patterns and document [domain] methodology

### Codification Tasks

```
codify_phase :: Observations → Methodology
codify_phase(obs) = pattern_extraction(
  identify_successful_patterns,
  document_[domain]_strategies,
  create_reusable_templates,
  define_quality_criteria
)
```

### Specific Actions

**1. Extract Successful [Domain] Patterns**

Analyze existing high-quality [domain work]:
- What makes good [domain work] in this codebase?
- Which [domain] patterns are most effective?
- How are [resources / fixtures / components] structured?

**Document** patterns in `knowledge/[domain]-patterns-iteration-{n}.md`

**Pattern Template**:
```markdown
### Pattern N: [Pattern Name]

**Purpose**: [What problem it solves]

**When to Use**:
- [Scenario 1]
- [Scenario 2]

**Structure**:
[Code/template example]

**Benefits**:
- [Benefit 1]
- [Benefit 2]

**Time Estimate**: ~[X]-[Y] minutes per [unit of work]
```

**2. Design [Metric]-Driven Workflow**

Create systematic approach:
1. Identify gap ([what's missing])
2. Prioritize (value × risk)
3. Execute [domain work] ([unit of work])
4. Verify [metric] improvement
5. Refactor for quality

**Document** workflow in `knowledge/[domain]-workflow-iteration-{n}.md`

**3. Create [Domain] Templates**

For common scenarios:
- [Template type 1]
- [Template type 2]
- [Template type 3]

**Save** templates in `knowledge/[domain]-templates-iteration-{n}.md`

**4. Define Quality Standards**

Establish criteria for high-quality [domain work]:
1. **[Criterion 1]**: [Description]
2. **[Criterion 2]**: [Description]
3. **[Criterion 3]**: [Description]
4. **[Criterion 4]**: [Description]
5. **[Criterion 5]**: [Description]
6. **[Criterion 6]**: [Description]
7. **[Criterion 7]**: [Description]
8. **[Criterion 8]**: [Description]

**Document** in `knowledge/quality-standards-iteration-{n}.md`

### Codification Deliverables

- [ ] [Domain] patterns extracted (≥[N] patterns)
- [ ] [Metric]-driven workflow documented
- [ ] [Domain] templates created
- [ ] Quality standards defined (≥[M] criteria)
- [ ] Methodology artifacts in `knowledge/`

---

## Phase 3: AUTOMATE (Tool Creation & CI Integration)

**Objective**: Create automation tools and integrate with CI/CD

### Automation Tasks

```
automate_phase :: Methodology → Tools
automate_phase(method) = tool_creation(
  build_analysis_tools,
  create_generators,
  integrate_ci,
  implement_gates
)
```

### Specific Actions

**1. Build [Domain] Analysis Tool**

Create script to automate gap identification:

**Tool**: `scripts/analyze-[domain]-gaps.sh`

```bash
#!/bin/bash
# Purpose: Automatically identify [domain] gaps and prioritize work
# Usage: ./scripts/analyze-[domain]-gaps.sh

# [Step 1: Collect data]
# [Step 2: Parse data]
# [Step 3: Identify gaps]
# [Step 4: Prioritize]
# [Step 5: Output report]
```

**Expected speedup**: [XX]x (vs manual analysis)

**2. Create [Resource] Generator**

Create script to generate [resource] scaffolds:

**Tool**: `scripts/generate-[resource].sh`

```bash
#!/bin/bash
# Purpose: Generate [resource] scaffold from pattern templates
# Usage: ./scripts/generate-[resource].sh [parameters]

# [Step 1: Select pattern]
# [Step 2: Generate scaffold]
# [Step 3: Save to file]
```

**Expected speedup**: [YY]x (vs manual writing)

**3. Integrate CI/CD**

Add [domain] automation to CI pipeline:

**File**: `.github/workflows/[domain].yml` (or update existing workflow)

```yaml
name: [Domain] Quality Gates

on: [push, pull_request]

jobs:
  [domain]-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run [domain] analysis
        run: ./scripts/analyze-[domain]-gaps.sh
      - name: Check [metric] threshold
        run: |
          # Fail if [metric] < [threshold]
```

**4. Implement Quality Gates**

Add Makefile target:

```makefile
.PHONY: [domain]-check
[domain]-check:
	@echo "Running [domain] quality checks..."
	@./scripts/analyze-[domain]-gaps.sh
	@# Add checks for [metric] >= [threshold]
	@echo "[Domain] quality gates passed"
```

**5. Create Comprehensive [Domain] Guide** (Final Iteration)

Consolidate all patterns, workflow, and tools into single production-ready guide:

**File**: `knowledge/[domain]-methodology-complete.md`

**Structure**:
- Overview
- Pattern Library ([N] patterns)
- Automation Tools ([M] tools)
- [Metric]-Driven Workflow
- Quality Standards
- Effectiveness Metrics
- Reusability Guide
- Troubleshooting

**Expected lines**: ~[XXX]-[YYY] lines (comprehensive)

### Automation Deliverables

- [ ] [Domain] analysis tool created (script in `scripts/`)
- [ ] [Resource] generator created (script in `scripts/`)
- [ ] CI/CD integration complete
- [ ] Quality gates implemented (Makefile + GitHub Actions)
- [ ] Comprehensive methodology guide created (final iteration only)

---

## Phase 4: EVALUATE (Value Function Calculation)

**Objective**: Calculate V_instance and V_meta, assess convergence

### Evaluation Tasks

```
evaluate_phase :: State → (V_instance, V_meta, Convergence)
evaluate_phase(s_n) =
  calculate_v_instance(s_n) ∧
  calculate_v_meta(s_n) ∧
  check_convergence(V, M, A)
```

### V_instance Calculation

**Components** (from README.md):

```
V_instance(s_n) = w₁·V_[component1](s_n) +
                  w₂·V_[component2](s_n) +
                  w₃·V_[component3](s_n) +
                  w₄·V_[component4](s_n)
```

**Calculate each component using rubrics from README.md**:

1. **V_[component1]**: [Current value and evidence]
   - Current state: [Description]
   - Score: [0.0-1.0]
   - Evidence: [Concrete measurements]

2. **V_[component2]**: [Current value and evidence]
   - Current state: [Description]
   - Score: [0.0-1.0]
   - Evidence: [Concrete measurements]

3. **V_[component3]**: [Current value and evidence]
   - Current state: [Description]
   - Score: [0.0-1.0]
   - Evidence: [Concrete measurements]

4. **V_[component4]**: [Current value and evidence]
   - Current state: [Description]
   - Score: [0.0-1.0]
   - Evidence: [Concrete measurements]

**Final V_instance(s_n)**: w₁·[score1] + w₂·[score2] + w₃·[score3] + w₄·[score4] = **[total]**

### V_meta Calculation

**Components** (universal BAIME rubrics):

```
V_meta(s_n) = 0.40·V_methodology_completeness(s_n) +
              0.30·V_methodology_effectiveness(s_n) +
              0.30·V_methodology_reusability(s_n)
```

**Calculate each component**:

1. **V_methodology_completeness**: [Current value and evidence]
   - Checklist coverage: [Process steps / Decision criteria / Examples / Edge cases / Rationale]
   - Documentation artifacts: [List files in knowledge/]
   - Score: [0.0-1.0] ([Level: Basic/Structured/Comprehensive/Fully Codified])
   - Evidence: [Concrete documentation counts]

2. **V_methodology_effectiveness**: [Current value and evidence]
   - Speedup measurement: [X]x vs ad-hoc approach
   - Quality improvement: [Y]% improvement in [metric]
   - Score: [0.0-1.0] ([Level: Marginal/Moderate/Significant/Transformative])
   - Evidence: [Time comparisons, before/after metrics]

3. **V_methodology_reusability**: [Current value and evidence]
   - Transfer test: [% modification needed for similar/different domains]
   - Domain independence: [% of methodology that is domain-agnostic]
   - Score: [0.0-1.0] ([Level: Domain-Specific/Partially Portable/Largely Portable/Highly Portable])
   - Evidence: [Multi-context validation data if available]

**Final V_meta(s_n)**: 0.40·[score1] + 0.30·[score2] + 0.30·[score3] = **[total]**

### Convergence Check

**Standard Dual Convergence Criteria**:

```
converged :: (V_instance, V_meta, M, A, history) → Bool
converged(V_i, V_m, M_n, A_n, hist) =
  V_i ≥ 0.80 ∧
  V_m ≥ 0.80 ∧
  M_n == M_{n-1} ∧
  A_n == A_{n-1} ∧
  ΔV_instance < 0.02 (for 2+ iterations) ∧
  ΔV_meta < 0.02 (for 2+ iterations)
```

**Check each criterion**:

1. ✅ / ❌ V_instance(s_n) ≥ 0.80: [score] [status]
2. ✅ / ❌ V_meta(s_n) ≥ 0.80: [score] [status]
3. ✅ / ❌ M_n == M_{n-1}: [comparison] [status]
4. ✅ / ❌ A_n == A_{n-1}: [comparison] [status]
5. ✅ / ❌ ΔV_instance < 0.02: [change over last 2+ iterations] [status]
6. ✅ / ❌ ΔV_meta < 0.02: [change over last 2+ iterations] [status]

**Convergence Status**: [CONVERGED / NOT CONVERGED]

**If NOT CONVERGED**: Continue to Phase 5 (Plan Next Iteration)
**If CONVERGED**: Proceed to Finalization (create results.md)

### Evaluation Deliverables

- [ ] V_instance(s_n) calculated with component breakdown
- [ ] V_meta(s_n) calculated with component breakdown
- [ ] Convergence criteria checked (all 6 criteria)
- [ ] Convergence status determined
- [ ] Value trajectory updated (track across iterations)

---

## Phase 5: EVOLVE (Plan Next Iteration)

**Objective**: Identify gaps and plan focus for next iteration

**Only execute if NOT CONVERGED**

### Evolution Tasks

```
evolve_phase :: (V, M, A, gaps) → (M_{n+1}, A_{n+1}, focus_{n+1})
evolve_phase(V_i, V_m, M, A, gaps) =
  identify_gaps(V_i, V_m) ∧
  decide_meta_evolution(M | gaps) ∧
  decide_agent_evolution(A | gaps) ∧
  plan_next_focus(gaps)
```

### Specific Actions

**1. Gap Analysis**

Identify what's missing based on V_instance and V_meta scores:

**Instance gaps** (V_instance < 0.80):
- Component [1] gap: [Description of what's missing]
- Component [2] gap: [Description of what's missing]
- Component [3] gap: [Description of what's missing]
- Component [4] gap: [Description of what's missing]

**Meta gaps** (V_meta < 0.80):
- Completeness gap: [What documentation/artifacts are missing]
- Effectiveness gap: [What measurements/validations are missing]
- Reusability gap: [What transfer tests are missing]

**2. Meta-Agent Evolution Decision**

```
meta_evolution :: M_{n} → M_{n+1}
meta_evolution(M) =
  if new_capability_needed
  then M' = M ∪ {new_capability}
  else M' = M
```

**Decision**: ✅ Keep M₀ unchanged / ⚠️ Add new capability: [capability name]

**Justification**: [Why M₀ is sufficient / Why new capability needed]

**3. Agent Set Evolution Decision**

```
agent_evolution :: A_{n} → A_{n+1}
agent_evolution(A) =
  if specialization_valuable ∧ efficiency_gain > 2x
  then A' = A ∪ {specialized_agent}
  else A' = A
```

**Decision**: ✅ Keep generic agents / ⚠️ Create specialized agent: [agent name]

**Justification**: [Why generic agents sufficient / Why specialization needed]

**Criteria for Specialization**:
- [ ] Generic agents insufficient (demonstrated over 2+ iterations)
- [ ] Specialization provides >2x efficiency gain
- [ ] Pattern will be reused across multiple files/modules

**4. Next Iteration Focus**

Based on gap analysis, plan focus for iteration {n+1}:

**Primary objective**: [Focus area 1]
**Secondary objective**: [Focus area 2]
**Stretch goal**: [Focus area 3]

**Expected progress**:
- V_instance: [current] → [target] (+[delta])
- V_meta: [current] → [target] (+[delta])

### Evolution Deliverables

- [ ] Gap analysis completed (instance + meta)
- [ ] Meta-agent evolution decision (M₀ sufficient / evolve)
- [ ] Agent set evolution decision (generic sufficient / specialize)
- [ ] Next iteration focus planned
- [ ] Expected progress targets set

---

## Iteration Output Structure

Each iteration should produce `iteration-{n}.md` with this structure:

```markdown
# Iteration {n}: [Focus Name]

**Date**: YYYY-MM-DD
**Duration**: ~[X] hours
**Status**: [In Progress / Completed]

## Executive Summary
[2-3 sentence overview of what was accomplished]

## Pre-Execution Context
- Previous V_instance: [score]
- Previous V_meta: [score]
- Gaps identified: [list]
- Focus: [primary objective]

## OBSERVE Phase
[Observations made, data collected, patterns identified]

**Deliverables**:
- [List data files created]

## CODIFY Phase
[Patterns extracted, methodology documented]

**Deliverables**:
- [List knowledge files created]

## AUTOMATE Phase
[Tools created, CI integration, quality gates]

**Deliverables**:
- [List scripts/automation created]

## EVALUATE Phase

### V_instance Components
[Breakdown of each component with scores]

**V_instance(s_{n})** = [calculation] = **[total]**

### V_meta Components
[Breakdown of each component with scores]

**V_meta(s_{n})** = [calculation] = **[total]**

### Convergence Check
[Check of all 6 criteria]

**Status**: [CONVERGED / NOT CONVERGED]

## EVOLVE Phase
[Gap analysis, evolution decisions, next iteration plan]

**Next Iteration Focus**: [if not converged]

## Iteration Summary
- Duration: [X] hours
- V_instance: [score] ([change from previous])
- V_meta: [score] ([change from previous])
- Key achievements: [list]
- Gaps remaining: [list]
```

---

## Multi-Context Validation (Final Iterations)

**When to perform**: Iteration {N-1} or {N}, after V_instance converged and V_meta approaching 0.80

**Objective**: Validate methodology transferability across different project contexts

### Validation Protocol

```
multi_context_validation :: Methodology → Effectiveness_Data
multi_context_validation(method) =
  select_contexts(3+) ∧
  measure_adaptation_effort ∧
  measure_speedup ∧
  calculate_v_reusability
```

### Context Selection

Select 3+ project archetypes within the same codebase:
1. **Context A**: [Archetype 1] (e.g., HTTP service, parser, business logic)
2. **Context B**: [Archetype 2] (different complexity/characteristics)
3. **Context C**: [Archetype 3] (different complexity/characteristics)

### Measurements Per Context

For each context, measure:

1. **Time without methodology**:
   - [Metric] analysis: [X] min
   - Pattern selection: [Y] min
   - [Resource] scaffolding: [Z] min
   - First [unit of work]: [W] min total
   - Subsequent [unit of work]: [V] min avg

2. **Time with methodology**:
   - [Metric] analysis: [X'] min (using tool)
   - Pattern selection: [Y'] min (using guide)
   - [Resource] scaffolding: [Z'] min (using generator)
   - First [unit of work]: [W'] min total
   - Subsequent [unit of work]: [V'] min avg

3. **Speedup calculation**:
   - First [unit of work] speedup: [W] / [W'] = [ratio]x
   - Subsequent speedup: [V] / [V'] = [ratio]x
   - Session average speedup: [calculation]x

4. **Adaptation effort**:
   - Workflow changes: [%]
   - Pattern modifications: [%]
   - Tool modifications: [%]
   - Total adaptation: [%]

### Aggregate Analysis

Calculate across all contexts:
- Average speedup: [min, max, avg]
- Average adaptation: [min, max, avg]
- Workflow universality: [% unchanged]

**Update V_meta**:
- V_effectiveness: Based on average speedup
- V_reusability: Based on average adaptation effort

**Save** validation data to `data/cross-context-effectiveness-iteration-{n}.yaml`

---

## Edge Cases and Troubleshooting

### If V_instance Converges Early (iteration < N-2)

**Situation**: V_instance ≥ 0.80 but V_meta < 0.80

**Action**:
- Continue iterating with focus on meta layer
- Invest in multi-context validation
- Enhance documentation completeness
- Create additional automation tools
- Maintain instance quality (don't regress)

### If V_meta Stuck (<0.70 after N iterations)

**Diagnose**:
- Completeness: Is documentation incomplete?
- Effectiveness: Missing measurements/validation?
- Reusability: Lacking multi-context evidence?

**Action**:
- Add missing documentation artifacts
- Perform effectiveness measurements
- Execute multi-context validation
- Create cross-language transfer guides (if applicable)

### If Both Metrics Oscillating

**Diagnose**:
- Are changes beneficial or detrimental?
- Is there trade-off between instance and meta?

**Action**:
- Stabilize instance layer first
- Focus meta improvements that don't affect instance
- Check for measurement errors

### If Meta-Agent M₀ Seems Insufficient

**Diagnose**:
- What capability is missing?
- Is it truly a meta-agent issue or agent specialization need?

**Decision Criteria**:
- Create new meta-agent capability ONLY if:
  - Generic M₀ capabilities insufficient for coordination
  - New domain requires novel meta-cognition
  - Demonstrated over 3+ iterations

**Conservative Approach**: M₀ has proven stable across 8 experiments; resist evolution.

---

## Final Convergence and Results

### When Fully Converged

All 6 convergence criteria met → Create `results.md`

**Template**: See Bootstrap-002/results.md for comprehensive structure

**Required Sections**:
1. Executive Summary
2. Convergence Achievement
3. Experiment Timeline
4. Value Function Analysis
5. Three-Tuple Output (O, Aₙ, Mₙ)
6. Methodology Quality Assessment
7. Transferability Validation
8. BAIME Framework Validation
9. Lessons Learned
10. Future Work
11. Conclusion
12. Appendix: Data Summary

**Length**: Expect ~1,000 lines for comprehensive results

---

## Checklist: Before Starting Each Iteration

- [ ] Read previous iteration file (`iteration-{n-1}.md`)
- [ ] Review V_instance and V_meta scores
- [ ] Identify gaps and focus areas
- [ ] Load Meta-Agent M₀ definition
- [ ] Load agent definitions (if any specialized agents exist)
- [ ] Review README.md objectives
- [ ] Review this ITERATION-PROMPTS.md file
- [ ] Ensure data/ and knowledge/ directories exist
- [ ] Ready to execute OCA cycle

---

## Checklist: After Completing Each Iteration

- [ ] Created `iteration-{n}.md` with all sections
- [ ] Calculated V_instance with component breakdown
- [ ] Calculated V_meta with component breakdown
- [ ] Checked all 6 convergence criteria
- [ ] Saved all data files to `data/`
- [ ] Saved all knowledge files to `knowledge/`
- [ ] Saved any scripts to `scripts/`
- [ ] Documented gaps and learnings
- [ ] Planned next iteration focus (if not converged)
- [ ] OR created results.md (if converged)

---

**Template Version**: 1.0
**Created**: 2025-10-18
**Based on**: Bootstrap-002 ITERATION-PROMPTS.md (proven successful)
**Validation**: Applied across 6 iterations with full dual convergence

**Usage**: Customize [DOMAIN], [METRIC], [placeholders] with domain-specific content, then use as execution guidance for iteration-executor agent.
