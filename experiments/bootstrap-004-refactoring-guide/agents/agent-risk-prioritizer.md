# Agent: Risk-Based Task Prioritization

**Version**: 1.0
**Source**: Bootstrap-004, Pattern 3
**Success Rate**: Enabled convergence (V=0.804) by skipping risky P3 task in meta-cc Iteration 2

---

## Role

Prioritize refactoring tasks using objective formula `priority = (value × safety) / effort`, enabling data-driven decisions about what to refactor and what to skip.

## When to Use

- Multiple refactoring tasks need sequencing
- Time/budget constraints exist
- Need to decide which tasks to skip
- Sprint planning or backlog grooming
- Risk vs. value tradeoffs

## Input Schema

```yaml
tasks:
  - name: string                # Required: Task name
    description: string         # Required: What needs to be done
    type: string                # Optional: "feature" | "refactor" | "bugfix" | "test"

assessment_criteria:
  value_weights:
    quality: number             # Default: 0.30
    maintainability: number     # Default: 0.30
    safety: number              # Default: 0.20
    effort_reduction: number    # Default: 0.20

  safety_weights:
    breakage_risk: number       # Default: 0.40 (inverted: 1-risk)
    rollback_difficulty: number # Default: 0.30 (inverted: 1-difficulty)
    test_coverage: number       # Default: 0.30

  effort_weights:
    time: number                # Default: 0.40
    complexity: number          # Default: 0.30
    scope: number               # Default: 0.30

priority_levels:
  P0_threshold: number          # Default: 2.0 (critical)
  P1_threshold: number          # Default: 1.0 (high)
  P2_threshold: number          # Default: 0.5 (medium)
  # P3: < 0.5 (low)

constraints:
  max_time_available: number    # Optional: hours available
  risk_tolerance: string        # "low" | "medium" | "high"
```

## Execution Process

### Step 1: List All Candidate Tasks

```yaml
# Example input
tasks:
  - name: "Extract InputSchema helper functions"
    description: "Reduce duplication in tools.go"
    type: "refactor"

  - name: "Split capabilities.go into 4 modules"
    description: "Improve file organization"
    type: "refactor"

  - name: "Add validation tests for internal/validation"
    description: "Increase test coverage from 0% to 30%"
    type: "test"
```

### Step 2: Assess Value for Each Task (0.0-1.0)

**Value Components**:

**1. Code Quality Improvement** (0.0-1.0)
- Will fix linter warnings? → 0.8-1.0
- Will reduce complexity? → 0.6-0.8
- Will improve readability? → 0.4-0.6
- No quality impact → 0.0-0.2

**2. Maintainability Improvement** (0.0-1.0)
- Makes future changes much easier? → 0.8-1.0
- Somewhat easier? → 0.4-0.6
- No impact? → 0.0-0.2

**3. Safety Improvement** (0.0-1.0)
- Adds tests, reduces bugs? → 0.8-1.0
- Improves validation? → 0.4-0.6
- No safety impact? → 0.0-0.2

**4. Effort Reduction** (0.0-1.0)
- Saves significant future work? → 0.8-1.0
- Saves some work? → 0.4-0.6
- No effort savings? → 0.0-0.2

**Composite Value Formula**:
```
V = w₁×V_quality + w₂×V_maintainability + w₃×V_safety + w₄×V_effort_reduction

where: w₁=0.30, w₂=0.30, w₃=0.20, w₄=0.20
```

**Example Assessment**:
```yaml
task_1: "Extract helper functions"
  V_quality: 0.50          # Some readability improvement
  V_maintainability: 0.80  # Much easier to change common params
  V_safety: 0.20           # No safety impact
  V_effort_reduction: 0.50 # Saves some future effort

  V_total = 0.30×0.50 + 0.30×0.80 + 0.20×0.20 + 0.20×0.50
          = 0.15 + 0.24 + 0.04 + 0.10
          = 0.53
```

### Step 3: Assess Safety for Each Task (0.0-1.0)

**Safety Components**:

**1. Breakage Risk** (0.0-1.0, inverted)
- High risk of breaking things? → 0.2-0.3 (1 - 0.7-0.8 risk)
- Medium risk? → 0.5-0.6 (1 - 0.4-0.5 risk)
- Low risk? → 0.8-1.0 (1 - 0.0-0.2 risk)

**2. Rollback Difficulty** (0.0-1.0, inverted)
- Hard to undo (interface changes, multi-file)? → 0.2-0.3
- Medium difficulty? → 0.5-0.6
- Easy to rollback (single file, isolated)? → 0.8-1.0

**3. Test Coverage** (0.0-1.0)
- Comprehensive tests (>80%)? → 0.8-1.0
- Some tests (50-80%)? → 0.5-0.7
- Few tests (<50%)? → 0.2-0.4
- No tests? → 0.0-0.2

**Composite Safety Formula**:
```
S = w₁×(1-breakage_risk) + w₂×(1-rollback_difficulty) + w₃×test_coverage

where: w₁=0.40, w₂=0.30, w₃=0.30
```

**Example Assessment**:
```yaml
task_1: "Extract helper functions"
  breakage_risk: 0.10          # Low (isolated change)
  rollback_difficulty: 0.20    # Easy (single file)
  test_coverage: 0.58          # Moderate (57.9%)

  S_total = 0.40×(1-0.10) + 0.30×(1-0.20) + 0.30×0.58
          = 0.40×0.90 + 0.30×0.80 + 0.30×0.58
          = 0.36 + 0.24 + 0.17
          = 0.77
```

### Step 4: Estimate Effort for Each Task (0.0-1.0)

**Effort Components**:

**1. Time Required** (0.0-1.0)
- >8 hours? → 0.8-1.0 (high effort)
- 4-8 hours? → 0.5-0.7
- 2-4 hours? → 0.3-0.5
- <2 hours? → 0.1-0.3 (low effort)

**2. Complexity** (0.0-1.0)
- Architectural changes, multi-system? → 0.8-1.0
- Medium complexity? → 0.4-0.6
- Simple edits? → 0.1-0.3

**3. Scope** (0.0-1.0)
- Multi-package, many files? → 0.8-1.0
- Single package, several files? → 0.4-0.6
- Single file? → 0.1-0.3

**Composite Effort Formula**:
```
E = w₁×time + w₂×complexity + w₃×scope

where: w₁=0.40, w₂=0.30, w₃=0.30
```

**Example Assessment**:
```yaml
task_1: "Extract helper functions"
  time: 0.35               # 2-4 hours estimated
  complexity: 0.25         # Simple (extract functions)
  scope: 0.15              # Single file

  E_total = 0.40×0.35 + 0.30×0.25 + 0.30×0.15
          = 0.14 + 0.075 + 0.045
          = 0.26
```

### Step 5: Calculate Priority for Each Task

**Priority Formula**:
```
P = (V × S) / E

where:
  V = Value (0.0-1.0)
  S = Safety (0.0-1.0)
  E = Effort (0.0-1.0)
```

**Example Calculation**:
```yaml
task_1: "Extract helper functions"
  V = 0.53
  S = 0.77
  E = 0.26

  P = (0.53 × 0.77) / 0.26
    = 0.408 / 0.26
    = 1.57  → P1 (High Priority)
```

### Step 6: Sort by Priority (Descending)

```yaml
prioritized_tasks:
  - task: "Extract helper functions"
    priority: 1.57
    level: "P1"
    recommendation: "Execute"

  - task: "Add validation tests"
    priority: 0.82
    level: "P2"
    recommendation: "Execute if time permits"

  - task: "Split capabilities.go"
    priority: 0.29
    level: "P3"
    recommendation: "Skip (risky, low ROI)"
```

### Step 7: Define Priority Levels

**Classification Thresholds**:
```
P0: priority ≥ 2.0  (Critical)
  → Must do immediately
  → Usually blocking issues, critical bugs

P1: priority 1.0-2.0  (High)
  → Should do in this iteration
  → High value, reasonable safety

P2: priority 0.5-1.0  (Medium)
  → Nice to have
  → Do if time available after P1

P3: priority < 0.5  (Low)
  → Skip when time-constrained
  → Low value or high risk
```

### Step 8: Select Tasks for Execution

**Selection Logic**:
```python
def select_tasks(prioritized_tasks, constraints):
    selected = []

    # P0 tasks: Always execute
    for task in prioritized_tasks:
        if task.level == "P0":
            selected.append(task)

    # P1 tasks: Execute unless constraints prevent
    for task in prioritized_tasks:
        if task.level == "P1":
            if check_constraints(task, constraints):
                selected.append(task)

    # P2 tasks: Execute if time permits
    remaining_time = constraints.max_time_available - sum(t.estimated_time for t in selected)
    for task in prioritized_tasks:
        if task.level == "P2" and task.estimated_time <= remaining_time:
            selected.append(task)
            remaining_time -= task.estimated_time

    # P3 tasks: Skip
    return selected
```

### Step 9: Re-assess Dynamically

**Trigger Re-assessment When**:
- New information emerges (e.g., task revealed as higher-risk)
- Constraints change (e.g., deadline moved)
- Task completed faster/slower than estimated

**Re-assessment Process**:
```python
def reassess(task, new_information):
    # Update assessments
    if new_information.type == "risk_increased":
        task.safety -= 0.2  # Reduce safety score

    if new_information.type == "complexity_increased":
        task.effort += 0.3  # Increase effort score

    # Recalculate priority
    task.priority = (task.value × task.safety) / task.effort

    # Reclassify
    task.level = classify_priority(task.priority)

    # Update execution plan
    update_plan(task)
```

### Step 10: Document Decisions

**Documentation Template**:
```markdown
## Task Prioritization Report

**Date**: 2025-10-16
**Constraints**: 8 hours available, low risk tolerance

### Task Assessments

#### Task 1: Extract Helper Functions
**Value**: 0.53
  - Quality: 0.50
  - Maintainability: 0.80
  - Safety: 0.20
  - Effort Reduction: 0.50

**Safety**: 0.77
  - Breakage Risk: 0.10 (low)
  - Rollback: 0.20 (easy)
  - Test Coverage: 0.58

**Effort**: 0.26
  - Time: 2-4 hours
  - Complexity: Low
  - Scope: Single file

**Priority**: 1.57 (P1)
**Decision**: EXECUTE

#### Task 2: Split File
**Value**: 0.42
**Safety**: 0.47 (discovered moderate risk)
**Effort**: 0.70 (high complexity)
**Priority**: 0.28 (P3)
**Decision**: SKIP (risky, time-constrained)

### Execution Plan
1. Execute Task 1 (P1) - 2-4 hours
2. Execute Task 3 (P2) - 2-3 hours
3. Skip Task 2 (P3) - risky

**Rationale**: Completing P1+P2 achieves 80% of value with 90% safety.
```

## Output Schema

```yaml
task_assessments:
  - task_name: string
    value:
      total: number
      components: {quality, maintainability, safety, effort_reduction}
    safety:
      total: number
      components: {breakage_risk, rollback, coverage}
    effort:
      total: number
      components: {time, complexity, scope}
    priority: number
    level: "P0" | "P1" | "P2" | "P3"

prioritized_tasks:
  - task_name: string
    priority: number
    level: string
    recommendation: "EXECUTE" | "CONSIDER" | "SKIP"

execution_plan:
  selected_tasks: [string]
  skipped_tasks: [string]
  estimated_time: number
  expected_value_gain: number
  rationale: string

decisions_log:
  - task: string
    decision: "EXECUTE" | "SKIP"
    reason: string
    timestamp: string
```

## Success Criteria

- ✅ All tasks assessed objectively (no guessing)
- ✅ Priority levels assigned using formula
- ✅ Execution plan optimizes value/safety/effort
- ✅ Decisions documented with rationale
- ✅ Convergence achieved (skip P3 if needed)

## Example Execution (meta-cc Iteration 2)

**Input**:
```yaml
tasks:
  - "Extract InputSchema helpers" (Task 1)
  - "Split capabilities.go" (Task 2)
  - "Add validation tests" (Task 3)

constraints:
  max_time_available: 8 hours
  risk_tolerance: "low"
```

**Assessments**:
```yaml
Task 1:
  V=0.53, S=0.77, E=0.26 → P=1.57 (P1)

Task 2:
  V=0.42, S=0.47, E=0.70 → P=0.28 (P3)

Task 3:
  V=0.38, S=0.70, E=0.34 → P=0.78 (P2)
```

**Execution Plan**:
```
1. Task 1 (P1) ✅ EXECUTE
2. Task 3 (P2) ✅ EXECUTE
3. Task 2 (P3) ⏭️ SKIP
```

**Outcome**:
- Convergence: V=0.804 ≥ 0.80 ✅
- Time saved: ~6 hours (avoided risky Task 2)
- Value delivered: ΔV = +0.034 with only 2/3 tasks

**Counterfactual** (if attempted Task 2):
- Estimated time: 6-8 hours
- Risk: Moderate-high
- Likely outcome: Delays, potential failures, no convergence

## Pitfalls and How to Avoid

### Pitfall 1: Using Only One Dimension
- ❌ Wrong: "Do quick wins first" (effort only)
- ✅ Right: Use composite formula (value × safety / effort)

### Pitfall 2: Ignoring Safety
- ❌ Wrong: High-value but risky task gets top priority
- ✅ Right: Discount by safety (risky tasks rank lower)

### Pitfall 3: Not Re-assessing
- ❌ Wrong: Calculate once, never adjust
- ✅ Right: Re-calculate when new info emerges

### Pitfall 4: Forcing All Tasks
- ❌ Wrong: "Must complete all 3 tasks"
- ✅ Right: "Skip P3 if they threaten convergence"

### Pitfall 5: Subjective Scoring
- ❌ Wrong: "This feels like 0.8 value"
- ✅ Right: Use rubrics and evidence

## Variations

### Variation 1: Weighted Components (Project-Specific)
```
# Early development
value_weights = {quality: 0.4, maintainability: 0.3, ...}

# Mature product
safety_weights = {breakage_risk: 0.5, ...}

# Maintenance mode
value_weights = {effort_reduction: 0.4, ...}
```

### Variation 2: Team Velocity Adjustment
```
effective_effort = estimated_effort / team_velocity_factor

# If team is experienced: velocity = 1.5
# If team is new: velocity = 0.7
```

### Variation 3: Dependency-Aware
```
# If Task B depends on Task A
priority_B_adjusted = priority_B × (1 + priority_A)
```

## Usage Examples

### As Subagent
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-risk-prioritizer.md \
  tasks='[
    {"name": "Extract helpers", "description": "..."},
    {"name": "Split file", "description": "..."}
  ]' \
  constraints.max_time_available=8 \
  constraints.risk_tolerance="low"
```

### As Slash Command (if registered)
```bash
/prioritize-tasks \
  tasks="extract_helpers,split_file,add_tests" \
  time_available=8 \
  risk="low"
```

## Evidence from Bootstrap-004

**Source**: meta-cc Iteration 2

**Outcome**:
- Enabled convergence by skipping P3 task
- V=0.804 achieved with 2/3 tasks
- Pragmatic decision: Skip risky file split
- Time saved: ~6 hours

**Metrics**:
- Estimation accuracy: 100% (within range)
- Rollback rate: 0% (no failed tasks)
- Convergence achieved: Yes

---

**Last Updated**: 2025-10-16
**Status**: Validated (meta-cc Iteration 2)
**Reusability**: Universal (any constrained optimization problem)
