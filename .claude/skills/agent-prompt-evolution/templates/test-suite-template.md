# Agent Test Suite Template

**Purpose**: Standardized test suite for agent prompt validation
**Usage**: Copy and customize for your agent type

---

## Test Suite: [Agent Name]

**Agent Type**: [Explore/Code-Gen/Analysis/etc.]
**Version**: [X.Y]
**Test Date**: [YYYY-MM-DD]
**Tester**: [Name]

---

## Test Configuration

**Test Environment**:
- Claude Code Version: [version]
- Model: [model-id]
- Session ID: [session-id]

**Test Parameters**:
- Number of tasks: [20 recommended]
- Task diversity: [Low/Medium/High]
- Complexity distribution:
  - Simple: [N] tasks
  - Medium: [N] tasks
  - Complex: [N] tasks

---

## Test Cases

### Task 1: [Brief Description]

**Type**: [Simple/Medium/Complex]
**Category**: [Search/Analysis/Generation/etc.]

**Input**:
```
[Exact prompt or command given to agent]
```

**Expected Outcome**:
```
[What a successful completion looks like]
```

**Actual Result**:
- Status: ✅ Success / ⚠️ Partial / ❌ Failed
- Quality Rating: [1-5]
- Time: [X.X] min
- Tokens: [X]k

**Notes**:
```
[Any observations, issues, or improvements identified]
```

---

### Task 2: [Brief Description]

**Type**: [Simple/Medium/Complex]
**Category**: [Search/Analysis/Generation/etc.]

**Input**:
```
[Exact prompt or command given to agent]
```

**Expected Outcome**:
```
[What a successful completion looks like]
```

**Actual Result**:
- Status: ✅ Success / ⚠️ Partial / ❌ Failed
- Quality Rating: [1-5]
- Time: [X.X] min
- Tokens: [X]k

**Notes**:
```
[Any observations, issues, or improvements identified]
```

---

[Repeat for all 20 tasks]

---

## Summary Statistics

### Overall Performance

**Success Rate**:
```
Total Tasks: [N]
Successful: [N] (✅)
Partial: [N] (⚠️)
Failed: [N] (❌)

Success Rate: [X]% ([successful] / [total])
```

**Quality Score**:
```
Task Quality Ratings: [4, 5, 3, 4, 5, ...]
Average Quality: [X.X] / 5
```

**Efficiency**:
```
Total Time: [X.X] min
Average Time: [X.X] min/task
Total Tokens: [X]k
Average Tokens: [X.X]k/task
```

**Reliability**:
```
Success by Complexity:
- Simple: [X]% ([Y]/[Z])
- Medium: [X]% ([Y]/[Z])
- Complex: [X]% ([Y]/[Z])

Reliability Score: [X.XX]
```

---

## Composite Metrics

### V_instance Calculation

```
Success Rate: [X]% = [0.XX]
Quality Score: [X.X]/5 = [0.XX]
Efficiency Score: [target - actual] / target = [0.XX]
Reliability: [0.XX]

V_instance = 0.40 × [success_rate] +
             0.30 × [quality_normalized] +
             0.20 × [efficiency_score] +
             0.10 × [reliability]

           = [0.XX] + [0.XX] + [0.XX] + [0.XX]
           = [0.XX]

Target: ≥ 0.80
Status: ✅ / ⚠️ / ❌
```

---

## Failure Analysis

### Failed Tasks

| Task ID | Description | Failure Reason | Pattern |
|---------|-------------|----------------|---------|
| [N]     | [Brief]     | [Why failed]   | [Type]  |
| [N]     | [Brief]     | [Why failed]   | [Type]  |

### Failure Patterns

**Pattern 1: [Name]** ([N] occurrences)
- Description: [What went wrong]
- Root Cause: [Why it happened]
- Proposed Fix: [How to address]

**Pattern 2: [Name]** ([N] occurrences)
- Description: [What went wrong]
- Root Cause: [Why it happened]
- Proposed Fix: [How to address]

---

## Quality Issues

### Tasks with Quality < 4

| Task ID | Quality | Issues Identified | Improvements Needed |
|---------|---------|-------------------|---------------------|
| [N]     | [1-3]   | [Description]     | [Actions]           |
| [N]     | [1-3]   | [Description]     | [Actions]           |

---

## Efficiency Analysis

### Tasks Exceeding Time Budget

| Task ID | Actual Time | Target Time | Δ    | Reason |
|---------|-------------|-------------|------|--------|
| [N]     | [X.X] min   | [Y] min     | [+Z] | [Why]  |
| [N]     | [X.X] min   | [Y] min     | [+Z] | [Why]  |

### Token Usage Analysis

```
Tokens per task: [min-max] range
High-usage tasks: [list]
Optimization opportunities: [suggestions]
```

---

## Recommendations

### Priority 1 (Critical)

1. **[Issue]**: [Description]
   - Impact: [High/Medium/Low]
   - Frequency: [X] occurrences
   - Proposed Fix: [Action]
   - Expected Improvement: [X]% success rate

2. **[Issue]**: [Description]
   - Impact: [High/Medium/Low]
   - Frequency: [X] occurrences
   - Proposed Fix: [Action]
   - Expected Improvement: [X]% quality

### Priority 2 (Important)

1. **[Issue]**: [Description]
   - Impact: [High/Medium/Low]
   - Frequency: [X] occurrences
   - Proposed Fix: [Action]

### Priority 3 (Nice to Have)

1. **[Improvement]**: [Description]
   - Benefit: [What improves]
   - Effort: [Low/Medium/High]

---

## Next Iteration Plan

### Focus Areas

1. **[Area 1]**: [Why focus here]
   - Baseline: [Current metric]
   - Target: [Goal metric]
   - Approach: [How to improve]

2. **[Area 2]**: [Why focus here]
   - Baseline: [Current metric]
   - Target: [Goal metric]
   - Approach: [How to improve]

### Prompt Changes

**Planned Additions**:
- [ ] [Guideline/instruction to add]
- [ ] [Constraint to add]
- [ ] [Example to add]

**Planned Clarifications**:
- [ ] [Instruction to clarify]
- [ ] [Constraint to adjust]

**Planned Removals**:
- [ ] [Unnecessary instruction]
- [ ] [Redundant constraint]

---

## Test Suite Evolution

### Version History

| Version | Date | Success | Quality | V_inst | Changes |
|---------|------|---------|---------|--------|---------|
| 0.1     | [D]  | [X]%    | [X.X]   | [0.XX] | Baseline|
| 0.2     | [D]  | [X]%    | [X.X]   | [0.XX] | [Changes]|
| [curr]  | [D]  | [X]%    | [X.X]   | [0.XX] | [Changes]|

### Convergence Tracking

```
Iteration 0: V_instance = [0.XX] (baseline)
Iteration 1: V_instance = [0.XX] ([+/-]%)
Iteration 2: V_instance = [0.XX] ([+/-]%)
Current:     V_instance = [0.XX] ([+/-]%)

Converged: ✅ / ❌
(Requires V_instance ≥ 0.80 for 2 consecutive iterations)
```

---

## Appendix: Task Catalog

### Task Templates by Category

**Search Tasks**:
- "Find all [pattern] in [scope]"
- "Locate [functionality] implementation"
- "Show [architecture aspect]"

**Analysis Tasks**:
- "Explain how [feature] works"
- "Identify [issue type] in [code]"
- "Compare [approach A] vs [approach B]"

**Generation Tasks**:
- "Create [artifact type] for [purpose]"
- "Generate [code/docs] following [pattern]"
- "Refactor [code] to [goal]"

### Complexity Guidelines

**Simple** (1-2 min, 1-3k tokens):
- Single-file search
- Direct lookup
- Straightforward generation

**Medium** (2-4 min, 3-7k tokens):
- Multi-file search
- Pattern analysis
- Moderate generation

**Complex** (4-6 min, 7-15k tokens):
- Cross-codebase search
- Deep analysis
- Complex generation

---

**Template Version**: 1.0
**Source**: BAIME Agent Prompt Evolution
**Usage**: Copy to `agent-test-suite-[name]-[version].md`
