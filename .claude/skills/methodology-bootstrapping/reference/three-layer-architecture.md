# Three-Layer OCA Architecture

**Version**: 1.0
**Framework**: BAIME - Observe-Codify-Automate
**Layers**: 3 (Observe, Codify, Automate)

Complete architectural reference for the OCA cycle.

---

## Overview

The OCA (Observe-Codify-Automate) cycle is the core of BAIME, consisting of three iterative layers that transform ad-hoc development into systematic, reusable methodologies.

```
ITERATION N:
  Observe → Codify → Automate → [Next Iteration]
     ↑                             ↓
     └──────────── Feedback ───────┘
```

---

## Layer 1: Observe

**Purpose**: Gather empirical data through hands-on work

**Duration**: 30-40% of iteration time (~20-30 min)

**Activities**:
1. **Apply** existing patterns/tools (if any)
2. **Execute** actual work on project
3. **Measure** results and effectiveness
4. **Identify** problems and gaps
5. **Document** observations

**Outputs**:
- Baseline metrics
- Problem list (prioritized)
- Pattern usage data
- Time measurements
- Quality metrics

**Example** (Testing Strategy, Iteration 1):
```markdown
## Observations

**Applied**:
- Wrote 5 unit tests manually
- Tried different test structures

**Measured**:
- Time per test: 15-20 min
- Coverage increase: +2.3%
- Tests passing: 5/5 (100%)

**Problems Identified**:
1. Setup code duplicated across tests
2. Unclear which functions to test first
3. No standard test structure
4. Coverage analysis manual and slow

**Time Spent**: 90 min (5 tests × 18 min avg)
```

### Observation Techniques

#### 1. Baseline Measurement

**What to measure**:
- Current state metrics (coverage, build time, error rate)
- Time spent on tasks
- Pain points and blockers
- Quality indicators

**Tools**:
```bash
# Testing
go test -cover ./...
go tool cover -func=coverage.out

# CI/CD
time make build
grep "FAIL" ci-logs.txt | wc -l

# Errors
grep "error" session.jsonl | wc -l
```

#### 2. Work Sampling

**Technique**: Track time on representative tasks

**Example**:
```markdown
Task: Write 5 unit tests

Sample 1: TestFunction1 - 18 min
Sample 2: TestFunction2 - 15 min
Sample 3: TestFunction3 - 22 min (complex)
Sample 4: TestFunction4 - 12 min (simple)
Sample 5: TestFunction5 - 16 min

Average: 16.6 min per test
Range: 12-22 min
Variance: High (complexity-dependent)
```

#### 3. Problem Taxonomy

**Classify problems**:
- **High frequency, high impact**: Urgent patterns needed
- **High frequency, low impact**: Automation candidates
- **Low frequency, high impact**: Document workarounds
- **Low frequency, low impact**: Ignore

---

## Layer 2: Codify

**Purpose**: Transform observations into documented patterns

**Duration**: 35-45% of iteration time (~25-35 min)

**Activities**:
1. **Analyze** observations for patterns
2. **Design** reusable solutions
3. **Document** patterns with examples
4. **Test** patterns on 2-3 cases
5. **Refine** based on feedback

**Outputs**:
- Pattern documents (problem-solution pairs)
- Code examples
- Usage guidelines
- Time/quality metrics per pattern

**Example** (Testing Strategy, Iteration 1):
```markdown
## Pattern: Table-Driven Tests

**Problem**: Writing multiple similar test cases is repetitive

**Solution**: Use table-driven pattern with test struct

**Structure**:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    Type
        expected Type
    }{
        {"case1", input1, output1},
        {"case2", input2, output2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Function(tt.input)
            assert.Equal(t, tt.expected, got)
        })
    }
}
```

**Time**: 12 min per test (vs 18 min manual)
**Savings**: 33% time reduction
**Validated**: 3 test functions, all passed
```

### Codification Techniques

#### 1. Pattern Template

```markdown
## Pattern: [Name]

**Category**: [Testing/CI/Error/etc.]

**Problem**:
[What problem does this solve?]

**Context**:
[When is this applicable?]

**Solution**:
[How to solve it? Step-by-step]

**Structure**:
[Code template or procedure]

**Example**:
[Real working example]

**Metrics**:
- Time: [X min]
- Quality: [metric]
- Reusability: [X%]

**Variations**:
[Alternative approaches]

**Anti-patterns**:
[Common mistakes]
```

#### 2. Pattern Hierarchy

**Level 1: Core Patterns** (6-8)
- Universal, high frequency
- Foundation for other patterns
- Example: Table-driven tests, Error classification

**Level 2: Composite Patterns** (2-4)
- Combine multiple core patterns
- Domain-specific
- Example: Coverage-driven gap closure (table-driven + prioritization)

**Level 3: Specialized Patterns** (0-2)
- Rare, specific use cases
- Optional extensions
- Example: Golden file testing for large outputs

#### 3. Progressive Refinement

**Iteration 0**: Observe only (no patterns yet)
**Iteration 1**: 2-3 core patterns (basics)
**Iteration 2**: 4-6 patterns (expanded)
**Iteration 3**: 6-8 patterns (refined)
**Iteration 4+**: Consolidate, no new patterns

---

## Layer 3: Automate

**Purpose**: Create tools to accelerate pattern application

**Duration**: 20-30% of iteration time (~15-20 min)

**Activities**:
1. **Identify** repetitive tasks (>3 times)
2. **Design** automation approach
3. **Implement** scripts/tools
4. **Test** on real examples
5. **Measure** speedup

**Outputs**:
- Automation scripts
- Tool documentation
- Speedup metrics (Nx faster)
- ROI calculations

**Example** (Testing Strategy, Iteration 2):
```markdown
## Tool: Coverage Gap Analyzer

**Purpose**: Identify which functions need tests (automated)

**Implementation**:
```bash
#!/bin/bash
# scripts/analyze-coverage-gaps.sh

go tool cover -func=coverage.out |
  grep "0.0%" |
  awk '{print $1, $2}' |
  while read file func; do
    # Categorize function type
    if grep -q "Error\|Valid" <<< "$func"; then
      echo "P1: $file:$func (error handling)"
    elif grep -q "Parse\|Process" <<< "$func"; then
      echo "P2: $file:$func (business logic)"
    else
      echo "P3: $file:$func (utility)"
    fi
  done | sort
```

**Speedup**: 15 min manual → 5 sec automated (180x)
**ROI**: 30 min investment, 10 uses = 150 min saved = 5x ROI
**Validated**: Used in iterations 2-4, always accurate
```

### Automation Techniques

#### 1. ROI Calculation

```
ROI = (time_saved × uses) / time_invested

Example:
- Manual task: 10 min
- Automation time: 1 hour
- Break-even: 6 uses
- Expected uses: 20
- ROI = (10 × 20) / 60 = 3.3x
```

**Rules**:
- ROI < 2x: Don't automate (not worth it)
- ROI 2-5x: Automate if frequently used
- ROI > 5x: Always automate

#### 2. Automation Tiers

**Tier 1: Simple Scripts** (15-30 min)
- Bash/Python scripts
- Parse existing tool output
- Generate boilerplate
- Example: Coverage gap analyzer

**Tier 2: Workflow Tools** (1-2 hours)
- Multi-step automation
- Integrate multiple tools
- Smart suggestions
- Example: Test generator with pattern detection

**Tier 3: Full Integration** (>2 hours)
- IDE/editor plugins
- CI/CD integration
- Pre-commit hooks
- Example: Automated methodology guide

**Start with Tier 1**, only progress to Tier 2/3 if ROI justifies

#### 3. Incremental Automation

**Phase 1**: Manual process documented
**Phase 2**: Script to assist (not fully automated)
**Phase 3**: Fully automated with validation
**Phase 4**: Integrated into workflow (hooks, CI)

**Example** (Test generation):
```
Phase 1: Copy-paste test template manually
Phase 2: Script generates template, manual fill-in
Phase 3: Script generates with smart defaults
Phase 4: Pre-commit hook suggests tests for new functions
```

---

## Dual-Layer Value Functions

### V_instance (Instance Quality)

**Measures**: Quality of work produced using methodology

**Formula**:
```
V_instance = Σ(w_i × metric_i)

Where:
- w_i = weight for metric i
- metric_i = normalized metric value (0-1)
- Σw_i = 1.0
```

**Example** (Testing):
```
V_instance = 0.5 × (coverage/target) +
             0.3 × (pass_rate) +
             0.2 × (maintainability)

Target: V_instance ≥ 0.80
```

**Convergence**: Stable for 2 consecutive iterations

### V_meta (Methodology Quality)

**Measures**: Quality and reusability of methodology itself

**Formula**:
```
V_meta = 0.4 × completeness +
         0.3 × transferability +
         0.3 × automation_effectiveness

Where:
- completeness = patterns_documented / patterns_needed
- transferability = cross_project_reuse_score (0-1)
- automation_effectiveness = time_with_tools / time_manual
```

**Example** (Testing):
```
V_meta = 0.4 × (8/8) +
         0.3 × (0.90) +
         0.3 × (4min/20min)

       = 0.4 + 0.27 + 0.06
       = 0.73

Target: V_meta ≥ 0.80
```

**Convergence**: Stable for 2 consecutive iterations

### Dual Convergence Criteria

**Both must be met**:
1. V_instance ≥ 0.80 for 2 consecutive iterations
2. V_meta ≥ 0.80 for 2 consecutive iterations

**Why dual-layer?**:
- V_instance alone: Could be good results with bad process
- V_meta alone: Could be great methodology with poor results
- Both together: Good results + reusable methodology

---

## Iteration Coordination

### Standard Flow

```
ITERATION N:
├─ Start (5 min)
│  ├─ Review previous iteration results
│  ├─ Set goals for this iteration
│  └─ Load context (patterns, tools, metrics)
│
├─ Observe (25 min)
│  ├─ Apply existing patterns
│  ├─ Work on project tasks
│  ├─ Measure results
│  └─ Document problems
│
├─ Codify (30 min)
│  ├─ Analyze observations
│  ├─ Create/refine patterns
│  ├─ Document with examples
│  └─ Validate on 2-3 cases
│
├─ Automate (20 min)
│  ├─ Identify automation opportunities
│  ├─ Create/improve tools
│  ├─ Measure speedup
│  └─ Calculate ROI
│
└─ Close (10 min)
   ├─ Calculate V_instance and V_meta
   ├─ Check convergence criteria
   ├─ Document iteration summary
   └─ Plan next iteration (if needed)
```

### Convergence Detection

```python
def check_convergence(history):
    if len(history) < 2:
        return False

    # Check last 2 iterations
    last_two = history[-2:]

    # Both V_instance and V_meta must be ≥ 0.80
    instance_converged = all(v.instance >= 0.80 for v in last_two)
    meta_converged = all(v.meta >= 0.80 for v in last_two)

    # No significant gaps remaining
    no_critical_gaps = last_two[-1].critical_gaps == 0

    return instance_converged and meta_converged and no_critical_gaps
```

---

## Best Practices

### Do's

✅ **Start with Observe** - Don't skip baseline
✅ **Validate patterns** - Test on 2-3 real examples
✅ **Measure everything** - Time, quality, speedup
✅ **Iterate quickly** - 60-90 min per iteration
✅ **Focus on ROI** - Automate high-value tasks
✅ **Document continuously** - Don't wait until end

### Don'ts

❌ **Don't skip Observe** - Patterns without data are guesses
❌ **Don't over-codify** - 6-8 patterns maximum
❌ **Don't premature automation** - Understand problem first
❌ **Don't ignore transferability** - Aim for 80%+ reuse
❌ **Don't continue past convergence** - Stop at dual 0.80

---

## Architecture Variations

### Rapid Convergence (3-4 iterations)

**Modifications**:
- Strong Iteration 0 (comprehensive baseline)
- Borrow patterns from similar projects (70-90% reuse)
- Parallel pattern development
- Focus on high-impact only

### Slow Convergence (>6 iterations)

**Causes**:
- Weak Iteration 0 (insufficient baseline)
- Too many patterns (>10)
- Complex domain
- Insufficient automation

**Fixes**:
- Strengthen baseline analysis
- Consolidate patterns
- Increase automation investment
- Focus on critical paths only

---

**Source**: BAIME Framework
**Status**: Production-ready, validated across 13 methodologies
**Convergence Rate**: 100% (all experiments converged)
**Average Iterations**: 4.9 (median 5)
