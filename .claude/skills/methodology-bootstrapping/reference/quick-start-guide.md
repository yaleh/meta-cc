# BAIME Quick Start Guide

**Version**: 1.0
**Framework**: Bootstrapped AI Methodology Engineering
**Time to First Iteration**: 45-90 minutes

Quick start guide for applying BAIME to create project-specific methodologies.

---

## What is BAIME?

**BAIME** = Bootstrapped AI Methodology Engineering

A meta-framework for systematically developing project-specific development methodologies through Observe-Codify-Automate (OCA) cycles.

**Use when**: Creating testing strategy, CI/CD pipeline, error handling patterns, documentation systems, or any reusable development methodology.

---

## 30-Minute Quick Start

### Step 1: Define Objective (10 min)

**Template**:
```markdown
## Objective
Create [methodology name] for [project] to achieve [goals]

## Success Criteria (Dual-Layer)
**Instance Layer** (V_instance ≥ 0.80):
- Metric 1: [e.g., coverage ≥ 75%]
- Metric 2: [e.g., tests pass 100%]

**Meta Layer** (V_meta ≥ 0.80):
- Patterns documented: [target count]
- Tools created: [target count]
- Transferability: [≥ 85%]
```

**Example** (Testing Strategy):
```markdown
## Objective
Create systematic testing methodology for meta-cc to achieve 75%+ coverage

## Success Criteria
Instance: coverage ≥ 75%, 100% pass rate
Meta: 8 patterns documented, 3 tools created, 90% transferable
```

### Step 2: Iteration 0 - Observe (20 min)

**Actions**:
1. Analyze current state
2. Identify pain points
3. Measure baseline metrics
4. Document problems

**Commands**:
```bash
# Example: Testing
go test -cover ./...  # Baseline coverage
grep -r "TODO.*test" .  # Find gaps

# Example: CI/CD
cat .github/workflows/*.yml  # Current pipeline
# Measure: build time, failure rate
```

**Output**: Baseline document with metrics and problems

### Step 3: Iteration 1 - Codify (30 min)

**Actions**:
1. Create 2-3 initial patterns
2. Document with examples
3. Apply to project
4. Measure improvement

**Template**:
```markdown
## Pattern 1: [Name]
**When**: [Use case]
**How**: [Steps]
**Example**: [Code snippet]
**Time**: [Minutes]
```

**Output**: Initial patterns document, applied examples

### Step 4: Iteration 2 - Automate (30 min)

**Actions**:
1. Identify repetitive tasks
2. Create automation scripts/tools
3. Measure speedup
4. Document tool usage

**Example**:
```bash
# Coverage gap analyzer
./scripts/analyze-coverage.sh coverage.out

# Test generator
./scripts/generate-test.sh FunctionName
```

**Output**: Working automation tools, usage docs

---

## Iteration Structure

### Standard Iteration (60-90 min)

```
ITERATION N:
├─ Observe (20 min)
│  ├─ Apply patterns from iteration N-1
│  ├─ Measure results
│  └─ Identify gaps
├─ Codify (25 min)
│  ├─ Refine existing patterns
│  ├─ Add new patterns for gaps
│  └─ Document improvements
└─ Automate (15 min)
   ├─ Create/improve tools
   ├─ Measure speedup
   └─ Update documentation
```

### Convergence Criteria

**Instance Layer** (V_instance ≥ 0.80):
- Primary metrics met (e.g., coverage, quality)
- Stable across iterations
- No critical gaps

**Meta Layer** (V_meta ≥ 0.80):
- Patterns documented and validated
- Tools created and effective
- Transferability demonstrated

**Stop when**: Both layers ≥ 0.80 for 2 consecutive iterations

---

## Value Function Calculation

### V_instance (Instance Quality)

```
V_instance = weighted_average(metrics)

Example (Testing):
V_instance = 0.5 × (coverage/target) + 0.3 × (pass_rate) + 0.2 × (speed)
          = 0.5 × (75/75) + 0.3 × (1.0) + 0.2 × (0.9)
          = 0.5 + 0.3 + 0.18
          = 0.98 ✓
```

### V_meta (Methodology Quality)

```
V_meta = 0.4 × completeness + 0.3 × reusability + 0.3 × automation

Where:
- completeness = patterns_documented / patterns_needed
- reusability = transferability_score (0-1)
- automation = time_saved / time_manual

Example:
V_meta = 0.4 × (8/8) + 0.3 × (0.90) + 0.3 × (0.75)
       = 0.4 + 0.27 + 0.225
       = 0.895 ✓
```

---

## Common Patterns

### Pattern 1: Gap Closure

**When**: Improving metrics systematically (coverage, quality, etc.)

**Steps**:
1. Measure baseline
2. Identify gaps (prioritized)
3. Create pattern to address top gap
4. Apply pattern
5. Re-measure

**Example**: Test coverage 60% → 75%
- Identify 10 uncovered functions
- Create table-driven test pattern
- Apply to top 5 functions
- Coverage increases to 68%
- Repeat

### Pattern 2: Problem-Pattern-Solution

**When**: Documenting reusable solutions

**Template**:
```markdown
## Problem
[What problem does this solve?]

## Context
[When does this problem occur?]

## Solution
[How to solve it?]

## Example
[Concrete code example]

## Results
[Measured improvements]
```

### Pattern 3: Automation-First

**When**: Task done >3 times

**Steps**:
1. Identify repetitive task
2. Measure time manually
3. Create script/tool
4. Measure time with automation
5. Calculate ROI = time_saved / time_invested

**Example**:
- Manual coverage analysis: 15 min
- Script creation: 30 min
- Script execution: 30 sec
- ROI: (15 min × 20 uses) / 30 min = 10x

---

## Rapid Convergence Tips

### Achieve 3-4 Iteration Convergence

**1. Strong Iteration 0**
- Comprehensive baseline analysis
- Clear problem taxonomy
- Initial pattern seeds

**2. Focus on High-Impact**
- Address top 20% problems (80% impact)
- Create patterns for frequent tasks
- Automate high-ROI tasks first

**3. Parallel Pattern Development**
- Work on 2-3 patterns simultaneously
- Test on multiple examples
- Iterate quickly

**4. Borrow from Prior Work**
- Reuse patterns from similar projects
- Adapt proven solutions
- 70-90% transferable

---

## Anti-Patterns

### ❌ Don't Do

1. **No baseline measurement**
   - Can't measure progress without baseline
   - Always start with Iteration 0

2. **Premature automation**
   - Automate before understanding problem
   - Manual first, automate once stable

3. **Pattern bloat**
   - Too many patterns (>12)
   - Keep it focused and actionable

4. **Ignoring transferability**
   - Project-specific hacks
   - Aim for 80%+ transferability

5. **Skipping validation**
   - Patterns not tested on real examples
   - Always validate with actual usage

### ✅ Do Instead

1. Start with baseline metrics
2. Manual → Pattern → Automate
3. 6-8 core patterns maximum
4. Design for reusability
5. Test patterns immediately

---

## Success Indicators

### After Iteration 1

- [ ] 2-3 patterns documented
- [ ] Baseline metrics improved 10-20%
- [ ] Patterns applied to 3+ examples
- [ ] Clear next steps identified

### After Iteration 3

- [ ] 6-8 patterns documented
- [ ] Instance metrics at 70-80% of target
- [ ] 1-2 automation tools created
- [ ] Patterns validated across contexts

### Convergence (Iteration 4-6)

- [ ] V_instance ≥ 0.80 (2 consecutive)
- [ ] V_meta ≥ 0.80 (2 consecutive)
- [ ] No critical gaps remaining
- [ ] Transferability ≥ 85%

---

## Examples by Domain

### Testing Methodology
- **Iterations**: 6
- **Patterns**: 8 (table-driven, fixture, CLI, etc.)
- **Tools**: 3 (coverage analyzer, test generator, guide)
- **Result**: 72.5% coverage, 5x speedup

### Error Recovery
- **Iterations**: 3
- **Patterns**: 13 error categories, 10 recovery patterns
- **Tools**: 3 (path validator, size checker, read-before-write)
- **Result**: 95.4% error classification, 23.7% automated prevention

### CI/CD Pipeline
- **Iterations**: 5
- **Patterns**: 7 pipeline stages, 4 optimization patterns
- **Tools**: 2 (pipeline analyzer, config generator)
- **Result**: Build time 8min → 3min, 100% reliability

---

## Getting Help

**Stuck on**:
- **Iteration 0**: Read baseline-quality-assessment skill
- **Slow convergence**: Read rapid-convergence skill
- **Validation**: Read retrospective-validation skill
- **Agent prompts**: Read agent-prompt-evolution skill

---

**Source**: BAIME Framework (Bootstrap experiments 001-013)
**Status**: Production-ready, validated across 13 methodologies
**Success Rate**: 100% convergence, 3.1x average speedup
