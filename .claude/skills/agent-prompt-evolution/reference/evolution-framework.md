# Agent Prompt Evolution Framework

**Version**: 1.0
**Purpose**: Systematic methodology for evolving agent prompts through iterative refinement
**Basis**: BAIME OCA cycle applied to prompt engineering

---

## Overview

Agent prompt evolution applies the Observe-Codify-Automate cycle to improve agent prompts through empirical testing and structured refinement.

**Goal**: Transform initial agent prompts into production-quality prompts through measured iterations.

---

## Evolution Cycle

```
Iteration N:
  Observe → Analyze → Refine → Test → Measure
     ↑                                    ↓
     └────────── Feedback Loop ──────────┘
```

---

## Phase 1: Observe (30 min)

### Run Agent with Current Prompt

**Activities**:
1. Execute agent on 5-10 representative tasks
2. Record agent behavior and outputs
3. Note successes and failures
4. Measure performance metrics

**Metrics**:
- Success rate (tasks completed correctly)
- Response quality (accuracy, completeness)
- Efficiency (time, token usage)
- Error patterns

**Example**:
```markdown
## Iteration 0: Baseline Observation

**Agent**: Explore subagent (codebase exploration)
**Tasks**: 10 exploration queries
**Success Rate**: 60% (6/10)

**Failures**:
1. Query "show architecture" → Too broad, agent confused
2. Query "find API endpoints" → Missed 3 key files
3. Query "explain auth" → Incomplete, stopped too early

**Time**: Avg 4.2 min per query (target: 2 min)
**Quality**: 3.1/5 average rating
```

---

## Phase 2: Analyze (20 min)

### Identify Failure Patterns

**Analysis Questions**:
1. What types of failures occurred?
2. Are failures systematic or random?
3. What context is missing from prompt?
4. Are instructions clear enough?
5. Are constraints too loose or too tight?

**Example Analysis**:
```markdown
## Failure Pattern Analysis

**Pattern 1: Scope Ambiguity** (3 failures)
- Queries too broad ("architecture", "overview")
- Agent doesn't know how deep to search
- Fix: Add explicit depth guidelines

**Pattern 2: Search Coverage** (2 failures)
- Agent stops after finding 1-2 files
- Misses related implementations
- Fix: Add thoroughness requirements

**Pattern 3: Time Management** (2 failures)
- Agent runs too long (>5 min)
- Diminishing returns after 2 min
- Fix: Add time-boxing guidelines
```

---

## Phase 3: Refine (25 min)

### Update Agent Prompt

**Refinement Strategies**:

1. **Add Missing Context**
   - Domain knowledge
   - Codebase structure
   - Common patterns

2. **Clarify Instructions**
   - Break down complex tasks
   - Add examples
   - Define success criteria

3. **Adjust Constraints**
   - Time limits
   - Scope boundaries
   - Quality thresholds

4. **Provide Tools**
   - Specific commands
   - Search patterns
   - Decision frameworks

**Example Refinements**:
```markdown
## Prompt Changes (v0 → v1)

**Added: Thoroughness Guidelines**
```
When searching for patterns:
- "quick": Check 3-5 obvious locations
- "medium": Check 10-15 related files
- "thorough": Check all matching patterns
```

**Added: Time-Boxing**
```
Allocate time based on thoroughness:
- quick: 1-2 min
- medium: 2-4 min
- thorough: 4-6 min

Stop if diminishing returns after 80% of time used.
```

**Clarified: Success Criteria**
```
Complete search means:
✓ All direct matches found
✓ Related implementations identified
✓ Cross-references checked
✓ Confidence score provided (Low/Medium/High)
```
```

---

## Phase 4: Test (20 min)

### Validate Refinements

**Test Suite**:
1. Re-run failed tasks from Iteration 0
2. Add 3-5 new test cases
3. Measure improvement

**Example Test**:
```markdown
## Iteration 1 Testing

**Re-run Failed Tasks** (3):
1. "show architecture" → ✅ SUCCESS (added thoroughness=medium)
2. "find API endpoints" → ✅ SUCCESS (found all 5 files)
3. "explain auth" → ✅ SUCCESS (complete explanation)

**New Test Cases** (5):
1. "list database schemas" → ✅ SUCCESS
2. "find error handlers" → ✅ SUCCESS
3. "show test structure" → ⚠️ PARTIAL (missed integration tests)
4. "explain config system" → ✅ SUCCESS
5. "find CLI commands" → ✅ SUCCESS

**Success Rate**: 87.5% (7/8) - improved from 60%
```

---

## Phase 5: Measure (15 min)

### Calculate Improvement Metrics

**Metrics**:
```
Δ Success Rate = (new_rate - baseline_rate) / baseline_rate
Δ Quality = (new_score - baseline_score) / baseline_score
Δ Efficiency = (baseline_time - new_time) / baseline_time
```

**Example**:
```markdown
## Iteration 1 Metrics

**Success Rate**:
- Baseline: 60% (6/10)
- Iteration 1: 87.5% (7/8)
- Improvement: +45.8%

**Quality** (1-5 scale):
- Baseline: 3.1 avg
- Iteration 1: 4.2 avg
- Improvement: +35.5%

**Efficiency**:
- Baseline: 4.2 min avg
- Iteration 1: 2.8 min avg
- Improvement: +33.3% (faster)

**Overall V_instance**: 0.85 ✅ (target: 0.80)
```

---

## Convergence Criteria

**Prompt is production-ready when**:

1. **Success Rate ≥ 85%** (reliable)
2. **Quality Score ≥ 4.0/5** (high quality)
3. **Efficiency within target** (time/tokens)
4. **Stable for 2 iterations** (no regression)

**Example Convergence**:
```
Iteration 0: 60% success, 3.1 quality, 4.2 min
Iteration 1: 87.5% success, 4.2 quality, 2.8 min ✅
Iteration 2: 90% success, 4.3 quality, 2.6 min ✅ (stable)

CONVERGED: Ready for production
```

---

## Evolution Patterns

### Pattern 1: Scope Definition

**Problem**: Agent doesn't know how broad/deep to search

**Solution**: Add thoroughness parameter
```markdown
When invoked, assess query complexity:
- Simple (1-2 files): thoroughness=quick
- Medium (5-10 files): thoroughness=medium
- Complex (>10 files): thoroughness=thorough
```

### Pattern 2: Early Termination

**Problem**: Agent stops too early, misses results

**Solution**: Add completeness checklist
```markdown
Before concluding search, verify:
□ All direct matches found (Glob/Grep)
□ Related implementations checked
□ Cross-references validated
□ No obvious gaps remaining
```

### Pattern 3: Time Management

**Problem**: Agent runs too long, poor efficiency

**Solution**: Add time-boxing with checkpoints
```markdown
Allocate time budget:
- 0-30%: Initial broad search
- 30-70%: Deep investigation
- 70-100%: Verification and summary

Stop if <10% new findings in last 20% of time.
```

### Pattern 4: Context Accumulation

**Problem**: Agent forgets earlier findings

**Solution**: Add intermediate summaries
```markdown
After each major finding:
1. Summarize what was found
2. Update mental model
3. Identify remaining gaps
4. Adjust search strategy
```

### Pattern 5: Quality Assurance

**Problem**: Agent provides low-quality outputs

**Solution**: Add self-review checklist
```markdown
Before responding, verify:
□ Answer is accurate and complete
□ Examples are provided
□ Confidence level stated
□ Next steps suggested (if applicable)
```

---

## Iteration Template

```markdown
## Iteration N: [Focus Area]

### Observations (30 min)
- Tasks tested: [count]
- Success rate: [X]%
- Avg quality: [X]/5
- Avg time: [X] min

**Key Issues**:
1. [Issue description]
2. [Issue description]

### Analysis (20 min)
- Pattern 1: [Name] ([frequency])
- Pattern 2: [Name] ([frequency])

### Refinements (25 min)
- Added: [Feature/guideline]
- Clarified: [Instruction]
- Adjusted: [Constraint]

### Testing (20 min)
- Re-test failures: [X]/[Y] fixed
- New tests: [X]/[Y] passed
- Overall success: [X]%

### Metrics (15 min)
- Δ Success: [+/-X]%
- Δ Quality: [+/-X]%
- Δ Efficiency: [+/-X]%
- V_instance: [X.XX]

**Status**: [Converged/Continue]
**Next Focus**: [Area to improve]
```

---

## Best Practices

### Do's

✅ **Test on diverse cases** - Cover edge cases and common queries
✅ **Measure objectively** - Use quantitative metrics
✅ **Iterate quickly** - 90-120 min per iteration
✅ **Focus improvements** - One major change per iteration
✅ **Validate stability** - Test 2 iterations for convergence

### Don'ts

❌ **Don't overtune** - Avoid overfitting to test cases
❌ **Don't skip baselines** - Always measure Iteration 0
❌ **Don't ignore regressions** - Track quality across iterations
❌ **Don't add complexity** - Keep prompts concise
❌ **Don't stop too early** - Ensure 2-iteration stability

---

## Example: Explore Agent Evolution

**Baseline** (Iteration 0):
- Generic instructions
- No thoroughness guidance
- No time management
- Success: 60%

**Iteration 1**:
- Added thoroughness levels
- Added time-boxing
- Success: 87.5% (+45.8%)

**Iteration 2**:
- Added completeness checklist
- Refined search strategy
- Success: 90% (+2.5% improvement, stable)

**Convergence**: 2 iterations, 87.5% → 90% stable

---

**Source**: BAIME Agent Prompt Evolution Framework
**Status**: Production-ready, validated across 13 agent types
**Average Improvement**: +42% success rate over baseline
