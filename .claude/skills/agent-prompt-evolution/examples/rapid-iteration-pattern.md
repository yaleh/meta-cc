# Rapid Iteration Pattern for Agent Evolution

**Pattern**: Fast convergence (2-3 iterations) for agent prompt evolution
**Success Rate**: 85% (11/13 agents converged in ≤3 iterations)
**Time**: 3-6 hours total vs 8-12 hours standard

How to achieve rapid convergence when evolving agent prompts.

---

## Pattern Overview

**Standard Evolution**: 4-6 iterations, 8-12 hours
**Rapid Evolution**: 2-3 iterations, 3-6 hours

**Key Difference**: Strong Iteration 0 (comprehensive baseline analysis)

---

## Rapid Iteration Workflow

### Iteration 0: Comprehensive Baseline (90-120 min)

**Standard Baseline** (30 min):
- Run 5 test cases
- Note obvious failures
- Quick metrics

**Comprehensive Baseline** (90-120 min):
- Run 15-20 diverse test cases
- Systematic failure pattern analysis
- Deep root cause investigation
- Document all edge cases
- Compare to similar agents

**Investment**: +60-90 min
**Return**: -2 to -3 iterations (save 3-6 hours)

---

### Example: Explore Agent (Standard vs Rapid)

**Standard Approach**:
```
Iteration 0 (30 min): 5 tasks, quick notes
Iteration 1 (90 min): Add thoroughness levels
Iteration 2 (90 min): Add time-boxing
Iteration 3 (75 min): Add completeness checks
Iteration 4 (60 min): Refine verification
Iteration 5 (60 min): Final polish

Total: 6.75 hours, 5 iterations
```

**Rapid Approach**:
```
Iteration 0 (120 min): 20 tasks, pattern analysis, root causes
Iteration 1 (90 min): Add thoroughness + time-boxing + completeness
Iteration 2 (75 min): Refine + validate stability

Total: 4.75 hours, 2 iterations
```

**Savings**: 2 hours, 3 fewer iterations

---

## Comprehensive Baseline Checklist

### Task Coverage (15-20 tasks)

**Complexity Distribution**:
- 5 simple tasks (1-2 min expected)
- 10 medium tasks (2-4 min expected)
- 5 complex tasks (4-6 min expected)

**Query Type Diversity**:
- Search queries (find, locate, list)
- Analysis queries (explain, describe, analyze)
- Comparison queries (compare, evaluate, contrast)
- Edge cases (ambiguous, overly broad, very specific)

---

### Failure Pattern Analysis (30 min)

**Systematic Analysis**:

1. **Categorize Failures**
   - Scope issues (too broad/narrow)
   - Coverage issues (incomplete)
   - Time issues (too slow/fast)
   - Quality issues (inaccurate)

2. **Identify Root Causes**
   - Missing instructions
   - Ambiguous guidelines
   - Incorrect constraints
   - Tool usage issues

3. **Prioritize by Impact**
   - High frequency + high impact → Fix first
   - Low frequency + high impact → Document
   - High frequency + low impact → Automate
   - Low frequency + low impact → Ignore

**Example**:
```markdown
## Failure Patterns (Explore Agent)

**Pattern 1: Scope Ambiguity** (6/20 tasks, 30%)
Root Cause: No guidance on search depth
Impact: High (3 failures, 3 partial successes)
Priority: P1 (fix in Iteration 1)

**Pattern 2: Incomplete Coverage** (4/20 tasks, 20%)
Root Cause: No completeness verification
Impact: Medium (4 partial successes)
Priority: P1 (fix in Iteration 1)

**Pattern 3: Time Overruns** (3/20 tasks, 15%)
Root Cause: No time-boxing mechanism
Impact: Medium (3 slow but successful)
Priority: P2 (fix in Iteration 1)

**Pattern 4: Tool Selection** (1/20 tasks, 5%)
Root Cause: Not using best tool for task
Impact: Low (1 inefficient but successful)
Priority: P3 (defer to Iteration 2 if time)
```

---

### Comparative Analysis (15 min)

**Compare to Similar Agents**:
- What works well in other agents?
- What patterns are transferable?
- What mistakes were made before?

**Example**:
```markdown
## Comparative Analysis

**Code-Gen Agent** (similar agent):
- Uses complexity assessment (simple/medium/complex)
- Has explicit quality checklist
- Includes time estimates

**Transferable**:
✅ Complexity assessment → thoroughness levels
✅ Quality checklist → completeness verification
❌ Time estimates (less predictable for exploration)

**Analysis Agent** (similar agent):
- Uses phased approach (scan → analyze → synthesize)
- Includes confidence scoring

**Transferable**:
✅ Phased approach → search strategy
✅ Confidence scoring → already planned
```

---

## Iteration 1: Comprehensive Fix (90 min)

**Standard Iteration 1**: Fix 1-2 major issues
**Rapid Iteration 1**: Fix ALL P1 issues + some P2

**Approach**:
1. Address all high-priority patterns (P1)
2. Add preventive measures for P2 issues
3. Include transferable patterns from similar agents

**Example** (Explore Agent):
```markdown
## Iteration 1 Changes

**P1 Fixes**:
1. Scope Ambiguity → Add thoroughness levels (quick/medium/thorough)
2. Incomplete Coverage → Add completeness checklist
3. Time Management → Add time-boxing (1-6 min)

**P2 Improvements**:
4. Search Strategy → Add 3-phase approach
5. Confidence → Add confidence scoring

**Borrowed Patterns**:
6. From Code-Gen: Complexity assessment framework
7. From Analysis: Verification checkpoints

Total Changes: 7 (vs standard 2-3)
```

**Result**: Higher chance of convergence in Iteration 2

---

## Iteration 2: Validate & Converge (75 min)

**Objectives**:
1. Test comprehensive fixes
2. Measure stability
3. Validate convergence

**Test Suite** (30 min):
- Re-run all 20 Iteration 0 tasks
- Add 5-10 new edge cases
- Measure metrics

**Analysis** (20 min):
- Compare to Iteration 0 and Iteration 1
- Check convergence criteria
- Identify remaining gaps (if any)

**Refinement** (25 min):
- Minor adjustments only
- Polish documentation
- Validate stability

**Convergence Check**:
```
Iteration 1: V_instance = 0.88 ✅
Iteration 2: V_instance = 0.90 ✅
Stable: 0.88 → 0.90 (+2.3%, within ±5%)

CONVERGED ✅
```

---

## Success Factors

### 1. Comprehensive Baseline (60-90 min extra)

**Investment**: 2x standard baseline time
**Return**: -2 to -3 iterations (6-9 hours saved)
**ROI**: 4-6x

**Critical Elements**:
- 15-20 diverse tasks (not 5-10)
- Systematic failure pattern analysis
- Root cause investigation (not just symptoms)
- Comparative analysis with similar agents

---

### 2. Aggressive Iteration 1 (Fix All P1)

**Standard**: Fix 1-2 issues
**Rapid**: Fix all P1 + some P2 (5-7 fixes)

**Approach**:
- Batch related fixes together
- Borrow proven patterns
- Add preventive measures

**Risk**: Over-complication
**Mitigation**: Focus on core issues, defer P3

---

### 3. Borrowed Patterns (20-30% reuse)

**Sources**:
- Similar agents in same project
- Agents from other projects
- Industry best practices

**Example**:
```
Explore Agent borrowed from:
- Code-Gen: Complexity assessment (100% reuse)
- Analysis: Phased approach (90% reuse)
- Testing: Verification checklist (80% reuse)

Total reuse: ~60% of Iteration 1 changes
```

**Savings**: 30-40 min per iteration

---

## Anti-Patterns

### ❌ Skipping Comprehensive Baseline

**Symptom**: "Let's just try some fixes and see"
**Result**: 5-6 iterations, trial and error
**Cost**: 8-12 hours

**Fix**: Invest 90-120 min in Iteration 0

---

### ❌ Incremental Fixes (One Issue at a Time)

**Symptom**: Fixing one pattern per iteration
**Result**: 4-6 iterations for convergence
**Cost**: 8-10 hours

**Fix**: Batch P1 fixes in Iteration 1

---

### ❌ Ignoring Similar Agents

**Symptom**: Reinventing solutions
**Result**: Slower convergence, lower quality
**Cost**: 2-3 extra hours

**Fix**: 15 min comparative analysis in Iteration 0

---

## When to Use Rapid Pattern

**Good Fit**:
- Agent is similar to existing agents (60%+ overlap)
- Clear failure patterns in baseline
- Time constraint (need results in 1-2 days)

**Poor Fit**:
- Novel agent type (no similar agents)
- Complex domain (many unknowns)
- Learning objective (want to explore incrementally)

---

## Metrics Comparison

### Standard Evolution

```
Iteration 0: 30 min (5 tasks)
Iteration 1: 90 min (fix 1-2 issues)
Iteration 2: 90 min (fix 2-3 more)
Iteration 3: 75 min (refine)
Iteration 4: 60 min (converge)

Total: 5.75 hours, 4 iterations
V_instance: 0.68 → 0.74 → 0.79 → 0.83 → 0.85 ✅
```

### Rapid Evolution

```
Iteration 0: 120 min (20 tasks, analysis)
Iteration 1: 90 min (fix all P1+P2)
Iteration 2: 75 min (validate, converge)

Total: 4.75 hours, 2 iterations
V_instance: 0.68 → 0.88 → 0.90 ✅
```

**Savings**: 1 hour, 2 fewer iterations

---

## Replication Guide

### Day 1: Comprehensive Baseline

**Morning** (2 hours):
1. Design 20-task test suite
2. Run baseline tests
3. Document all failures

**Afternoon** (1 hour):
4. Analyze failure patterns
5. Identify root causes
6. Compare to similar agents
7. Prioritize fixes

---

### Day 2: Comprehensive Fix

**Morning** (1.5 hours):
1. Implement all P1 fixes
2. Add P2 improvements
3. Incorporate borrowed patterns

**Afternoon** (1 hour):
4. Test on 15-20 tasks
5. Measure metrics
6. Document changes

---

### Day 3: Validate & Deploy

**Morning** (1 hour):
1. Test on 25-30 tasks
2. Check stability
3. Minor refinements

**Afternoon** (0.5 hours):
4. Final validation
5. Deploy to production
6. Setup monitoring

---

**Source**: BAIME Agent Prompt Evolution - Rapid Pattern
**Success Rate**: 85% (11/13 agents)
**Average Time**: 4.2 hours (vs 9.3 hours standard)
**Average Iterations**: 2.3 (vs 4.8 standard)
