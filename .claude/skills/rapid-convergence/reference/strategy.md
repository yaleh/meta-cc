# Rapid Convergence Strategy Guide

**Purpose**: Iteration-by-iteration tactics for 3-4 iteration convergence
**Time**: 10-15 hours total (vs 20-30 standard)

---

## Pre-Iteration 0: Planning (1-2 hours)

### Objectives

1. Confirm rapid convergence feasible
2. Establish measurement infrastructure
3. Define scope boundaries
4. Plan validation approach

### Tasks

**1. Baseline Assessment** (30 min):
```bash
# Query existing data
meta-cc query-tools --status=error
meta-cc query-user-messages --pattern="test|coverage"

# Calculate baseline metrics
# Estimate V_meta(s₀)
```

**2. Scope Definition** (20 min):
```markdown
## Domain: [1-sentence definition]

**In-Scope**: [3-5 items]
**Out-of-Scope**: [3-5 items]
**Edge Cases**: [Handling approach]
```

**3. Success Criteria** (20 min):
```markdown
## Convergence Targets

**V_instance ≥ 0.80**:
- Metric 1: [Target]
- Metric 2: [Target]

**V_meta ≥ 0.80**:
- Patterns: [8-10 documented]
- Tools: [3-5 created]
- Transferability: [≥80%]
```

**4. Prediction** (10 min):
```
Use prediction model:
Base(4) + penalties = [X] iterations expected
```

**Deliverable**: `README.md` with scope, targets, prediction

---

## Iteration 0: Comprehensive Baseline (3-5 hours)

### Objectives

- Achieve V_meta(s₀) ≥ 0.40
- Initial taxonomy: 70-80% coverage
- Identify top 3 automations

### Time Allocation

- Data analysis: 60-90 min (40%)
- Taxonomy creation: 45-75 min (30%)
- Pattern research: 30-45 min (20%)
- Automation planning: 15-30 min (10%)

### Tasks

**1. Comprehensive Data Analysis** (60-90 min):
```bash
# Extract ALL available data
meta-cc query-tools --scope=project > tools.jsonl
meta-cc query-user-messages --pattern=".*" > messages.jsonl

# Analyze patterns
cat tools.jsonl | jq -r '.error' | sort | uniq -c | sort -rn | head -20

# Calculate frequencies
total=$(cat tools.jsonl | wc -l)
# For each pattern: count / total
```

**2. Initial Taxonomy** (45-75 min):
```markdown
## Taxonomy v0

### Category 1: [Name] ([frequency]%, [count])
**Pattern**: [Description]
**Examples**: [3-5 examples]
**Root Cause**: [Analysis]

### Category 2: ...
[Repeat for 10-15 categories]

**Coverage**: [X]% ([classified]/[total])
```

**3. Pattern Research** (30-45 min):
```markdown
## Prior Art

**Source 1**: [Industry taxonomy/framework]
- Borrowed: [Pattern A, Pattern B, ...]
- Transferability: [X]%

**Source 2**: [Similar project]
- Borrowed: [Pattern C, Pattern D, ...]
- Adaptations needed: [List]

**Total Borrowable**: [X]/[Y] patterns = [Z]%
```

**4. Automation Planning** (15-30 min):
```markdown
## Top Automation Candidates

**1. [Tool Name]**
- Frequency: [X]% of cases
- Prevention: [Y]% of pattern
- ROI estimate: [Z]x
- Feasibility: [High/Medium/Low]

**2. [Tool Name]**
[Same structure]

**3. [Tool Name]**
[Same structure]
```

### Metrics

Calculate V_meta(s₀):
```
Completeness: [initial_categories] / [estimated_final] = [X]
Transferability: [borrowed] / [total_needed] = [Y]
Automation: [identified] / [expected] = [Z]

V_meta(s₀) = 0.4×[X] + 0.3×[Y] + 0.3×[Z] = [RESULT]

Target: ≥ 0.40 ✅/❌
```

**Deliverables**:
- `taxonomy-v0.md` (10-15 categories, ≥70% coverage)
- `baseline-metrics.md` (V_meta(s₀), frequencies)
- `automation-plan.md` (top 3 tools, ROI estimates)

---

## Iteration 1: High-Impact Automation (3-4 hours)

### Objectives

- V_instance ≥ 0.60 (significant improvement)
- Implement top 2-3 tools
- Expand taxonomy to 90%+ coverage

### Time Allocation

- Tool implementation: 90-120 min (50%)
- Taxonomy expansion: 45-60 min (25%)
- Testing & validation: 45-60 min (25%)

### Tasks

**1. Build Automation Tools** (90-120 min):
```bash
# Tool 1: validate-path.sh (30-40 min)
#!/bin/bash
# Fuzzy path matching, typo correction
# Target: 150-200 LOC

# Tool 2: check-file-size.sh (20-30 min)
#!/bin/bash
# File size check, auto-pagination
# Target: 100-150 LOC

# Tool 3: check-read-before-write.sh (40-50 min)
#!/bin/bash
# Workflow validation
# Target: 150-200 LOC
```

**2. Expand Taxonomy** (45-60 min):
```markdown
## Taxonomy v1

### [New Category 11]: [Name]
[Analysis of remaining 10-20% of cases]

### [New Category 12]: [Name]
[Continue until ≥90% coverage]

**Coverage**: [X]% ([classified]/[total])
**Gap Analysis**: [Remaining uncategorized patterns]
```

**3. Test & Measure** (45-60 min):
```bash
# Test tools on historical data
./scripts/validate-path.sh "path/to/file" # Expect suggestions
./scripts/check-file-size.sh "large-file.json" # Expect warning

# Calculate impact
prevented=$(estimate_prevention_rate)
time_saved=$(calculate_time_savings)
roi=$(calculate_roi)

# Update metrics
```

### Metrics

```
V_instance calculation:
- Success rate: [X]%
- Quality: [Y]/5
- Efficiency: [Z] min/task

V_instance = 0.4×[success] + 0.3×[quality/5] + 0.2×[efficiency] + 0.1×[reliability]
           = [RESULT]

Target: ≥ 0.60 (progress toward 0.80)
```

**Deliverables**:
- `scripts/tool1.sh`, `scripts/tool2.sh`, `scripts/tool3.sh`
- `taxonomy-v1.md` (≥90% coverage)
- `iteration-1-results.md` (V_instance, V_meta, gaps)

---

## Iteration 2: Validation & Refinement (3-4 hours)

### Objectives

- V_instance ≥ 0.80 ✅
- V_meta ≥ 0.80 ✅
- Validate stability (2 consecutive iterations)

### Time Allocation

- Retrospective validation: 60-90 min (40%)
- Taxonomy completion: 30-45 min (20%)
- Tool refinement: 45-60 min (25%)
- Documentation: 30-45 min (15%)

### Tasks

**1. Retrospective Validation** (60-90 min):
```bash
# Apply methodology to historical data
meta-cc validate \
  --methodology error-recovery \
  --history .claude/sessions/*.jsonl

# Measure:
# - Coverage: [X]% of historical cases handled
# - Time savings: [Y] hours saved
# - Prevention: [Z]% errors prevented
# - Confidence: [Score]
```

**2. Complete Taxonomy** (30-45 min):
```markdown
## Taxonomy v2 (Final)

[Review all categories]
[Add final 1-2 categories if needed]
[Refine existing categories]

**Final Coverage**: [X]% ≥ 95% ✅
**Uncategorized**: [Y]% (acceptable edge cases)
```

**3. Refine Tools** (45-60 min):
```bash
# Based on validation feedback
# - Fix bugs discovered
# - Improve accuracy
# - Add edge case handling
# - Optimize performance

# Re-test
# Re-measure ROI
```

**4. Documentation** (30-45 min):
```markdown
## Complete Methodology

### Patterns: [8-10 documented]
### Tools: [3-5 with usage]
### Transferability: [≥80%]
### Validation: [Results]
```

### Metrics

```
V_instance: [X] (≥0.80? ✅/❌)
V_meta: [Y] (≥0.80? ✅/❌)

Stability check:
- Iteration 1: V_instance = [A]
- Iteration 2: V_instance = [B]
- Change: [|B-A|] < 0.05? ✅/❌

Convergence: ✅/❌
```

**Decision**:
- ✅ Converged → Deploy
- ❌ Not converged → Iteration 3 (gap analysis)

**Deliverables**:
- `validation-report.md` (confidence, coverage, ROI)
- `methodology-complete.md` (production-ready)
- `transferability-guide.md` (80%+ reuse documentation)

---

## Iteration 3 (If Needed): Gap Closure (2-3 hours)

### Objectives

- Close specific gaps preventing convergence
- Reach dual-layer convergence (V_instance ≥ 0.80, V_meta ≥ 0.80)

### Gap Analysis

```markdown
## Why Not Converged?

**V_instance gaps** ([X] < 0.80):
- Metric A: [current] vs [target] = gap [Z]
- Root cause: [Analysis]
- Fix: [Action]

**V_meta gaps** ([Y] < 0.80):
- Component: [completeness/transferability/automation]
- Current: [X]
- Target: [Y]
- Fix: [Action]
```

### Focused Improvements

**Time**: 2-3 hours (targeted, not comprehensive)

**Tasks**:
- Address 1-2 major gaps only
- Refine existing work (no new patterns)
- Validate fixes

**Re-measure**:
```
V_instance: [X] ≥ 0.80? ✅/❌
V_meta: [Y] ≥ 0.80? ✅/❌
Stable for 2 iterations? ✅/❌
```

---

## Timeline Summary

### Rapid Convergence (3 iterations)

```
Pre-Iteration 0: 1-2h
Iteration 0:     3-5h (comprehensive baseline)
Iteration 1:     3-4h (automation + expansion)
Iteration 2:     3-4h (validation + convergence)
---
Total:          10-15h ✅
```

### Standard (If Iteration 3 Needed)

```
Pre-Iteration 0: 1-2h
Iteration 0:     3-5h
Iteration 1:     3-4h
Iteration 2:     3-4h
Iteration 3:     2-3h (gap closure)
---
Total:          12-18h (still faster than standard 20-30h)
```

---

## Anti-Patterns

### ❌ Rushing Iteration 0

**Symptom**: Spending 1-2 hours (vs 3-5)
**Impact**: Low V_meta(s₀), requires more iterations
**Fix**: Invest 3-5 hours for comprehensive baseline

### ❌ Over-Engineering Tools

**Symptom**: Spending 4+ hours per tool
**Impact**: Delays convergence
**Fix**: Simple tools (150-200 LOC, 30-60 min each)

### ❌ Premature Convergence

**Symptom**: Declaring done at V = 0.75
**Impact**: Quality issues in production
**Fix**: Respect 0.80 threshold, ensure 2-iteration stability

---

**Source**: BAIME Rapid Convergence Strategy
**Validation**: Bootstrap-003 (3 iterations, 10 hours)
**Success Rate**: 85% (11/13 experiments)
