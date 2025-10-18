# ITERATION-PROMPTS.md Template

**Purpose**: Structure for agent iteration prompts in BAIME experiments
**Usage**: Copy this template to `ITERATION-PROMPTS.md` in your experiment directory

---

## ITERATION-PROMPTS.md

```markdown
# Iteration Prompts for [Methodology Name]

**Experiment**: [experiment-name]
**Objective**: [Clear objective statement]
**Target**: [Specific measurable goals]

---

## Iteration 0: Baseline & Observe

**Objective**: Establish baseline metrics and identify core problems

**Prompt**:
```
Analyze current [domain] state for [project]:

1. Measure baseline metrics:
   - [Metric 1]: Current value
   - [Metric 2]: Current value
   - [Metric 3]: Current value

2. Identify problems:
   - High frequency, high impact issues
   - Pain points in current workflow
   - Gaps in current approach

3. Document observations:
   - Time spent on tasks
   - Quality indicators
   - Blockers encountered

4. Deliverables:
   - baseline-metrics.md
   - problems-identified.md
   - iteration-0-summary.md

Target time: 60 minutes
```

**Expected Output**:
- Baseline metrics document
- Prioritized problem list
- Initial hypotheses for patterns

---

## Iteration 1: Core Patterns

**Objective**: Create 2-3 core patterns addressing top problems

**Prompt**:
```
Develop initial patterns for [domain]:

1. Select top 3 problems from Iteration 0

2. For each problem, create pattern:
   - Problem statement
   - Solution approach
   - Code/process template
   - Working example
   - Time/quality metrics

3. Apply patterns:
   - Test on 2-3 real examples
   - Measure time and quality
   - Document results

4. Calculate V_instance:
   - [Metric 1]: Target vs Actual
   - [Metric 2]: Target vs Actual
   - Overall: V_instance = ?

5. Deliverables:
   - pattern-1.md
   - pattern-2.md
   - pattern-3.md
   - iteration-1-results.md

Target time: 90 minutes
```

**Expected Output**:
- 2-3 documented patterns with examples
- V_instance ≥ 0.50 (initial progress)
- Identified gaps for Iteration 2

---

## Iteration 2: Expand & Automate

**Objective**: Add 2-3 more patterns, create first automation tool

**Prompt**:
```
Expand pattern library and begin automation:

1. Refine Iteration 1 patterns based on usage

2. Add 2-3 new patterns for remaining gaps

3. Create automation tool:
   - Identify repetitive task (done >3 times)
   - Design tool to automate it
   - Implement script/tool
   - Measure speedup (Nx faster)
   - Calculate ROI

4. Calculate metrics:
   - V_instance = ?
   - V_meta = patterns_documented / patterns_needed

5. Deliverables:
   - pattern-4.md, pattern-5.md, pattern-6.md
   - scripts/tool-name.sh
   - tool-documentation.md
   - iteration-2-results.md

Target time: 90 minutes
```

**Expected Output**:
- 5-6 total patterns
- 1 automation tool (ROI > 3x)
- V_instance ≥ 0.70, V_meta ≥ 0.60

---

## Iteration 3: Consolidate & Validate

**Objective**: Reach V_instance ≥ 0.80, validate transferability

**Prompt**:
```
Consolidate patterns and validate methodology:

1. Review all patterns:
   - Merge similar patterns
   - Remove unused patterns
   - Refine documentation

2. Add final patterns if gaps exist (target: 6-8 total)

3. Create additional automation tools if ROI > 3x

4. Validate transferability:
   - Can patterns apply to other projects?
   - What needs adaptation?
   - Estimate transferability %

5. Calculate convergence:
   - V_instance = ? (target ≥ 0.80)
   - V_meta = ? (target ≥ 0.60)

6. Deliverables:
   - consolidated-patterns.md
   - transferability-analysis.md
   - iteration-3-results.md

Target time: 90 minutes
```

**Expected Output**:
- 6-8 consolidated patterns
- V_instance ≥ 0.80 (target met)
- Transferability score (≥ 80%)

---

## Iteration 4: Meta-Layer Convergence

**Objective**: Reach V_meta ≥ 0.80, prepare for production

**Prompt**:
```
Achieve meta-layer convergence:

1. Complete methodology documentation:
   - All patterns with examples
   - All tools with usage guides
   - Transferability guide for other languages/projects

2. Measure automation effectiveness:
   - Time manual vs with tools
   - ROI for each tool
   - Overall speedup

3. Calculate final metrics:
   - V_instance = ? (maintain ≥ 0.80)
   - V_meta = 0.4×completeness + 0.3×transferability + 0.3×automation
   - Check: V_meta ≥ 0.80?

4. Create deliverables:
   - complete-methodology.md (production-ready)
   - tool-suite-documentation.md
   - transferability-guide.md
   - final-results.md

5. If not converged: Identify remaining gaps and plan Iteration 5

Target time: 90 minutes
```

**Expected Output**:
- Complete, production-ready methodology
- V_meta ≥ 0.80 (converged)
- Dual convergence (V_instance ≥ 0.80, V_meta ≥ 0.80)

---

## Iteration 5+ (If Needed): Gap Closure

**Objective**: Address remaining gaps to reach dual convergence

**Prompt**:
```
Close remaining gaps:

1. Analyze why convergence not reached:
   - V_instance gaps: [specific metrics below target]
   - V_meta gaps: [patterns missing, tools needed, transferability issues]

2. Targeted improvements:
   - Create patterns for specific gaps
   - Improve automation for low ROI areas
   - Enhance transferability documentation

3. Re-measure:
   - V_instance = ?
   - V_meta = ?
   - Check dual convergence

4. Deliverables:
   - gap-analysis.md
   - additional-patterns.md (if needed)
   - iteration-N-results.md

Repeat until dual convergence achieved

Target time: 60-90 minutes per iteration
```

**Stopping Criteria**:
- V_instance ≥ 0.80 for 2 consecutive iterations
- V_meta ≥ 0.80 for 2 consecutive iterations
- No critical gaps remaining

---

## Customization Guide

### For Different Domains

**Testing Methodology**:
- Replace metrics with: coverage%, pass rate, test count
- Patterns: Test patterns (table-driven, fixture, etc.)
- Tools: Coverage analyzer, test generator

**CI/CD Pipeline**:
- Replace metrics with: build time, failure rate, deployment frequency
- Patterns: Pipeline stages, optimization patterns
- Tools: Pipeline analyzer, config generator

**Error Recovery**:
- Replace metrics with: error classification coverage, MTTR, prevention rate
- Patterns: Error categories, recovery patterns
- Tools: Error classifier, diagnostic workflows

### Adjusting Iteration Count

**Rapid Convergence (3-4 iterations)**:
- Strong Iteration 0 (2 hours)
- Borrow patterns (70-90% reuse)
- Focus on high-impact only

**Standard Convergence (5-6 iterations)**:
- Normal Iteration 0 (1 hour)
- Create patterns from scratch
- Comprehensive coverage

---

**Template Version**: 1.0
**Source**: BAIME Framework
**Usage**: Copy and customize for your experiment
**Success Rate**: 100% across 13 experiments
```
