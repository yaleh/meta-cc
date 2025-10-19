# Troubleshooting Guide Template

**Purpose**: Template for creating systematic troubleshooting documentation using Problem-Cause-Solution pattern

**Version**: 1.0
**Status**: Ready for use
**Validation**: Applied to BAIME troubleshooting section

---

## When to Use This Template

### Use For

✅ **Error diagnosis guides** (common errors, root causes, fixes)
✅ **Performance troubleshooting** (slow operations, bottlenecks, optimizations)
✅ **Configuration issues** (setup problems, misconfigurations, validation)
✅ **Integration problems** (API failures, connection issues, compatibility)
✅ **User workflow issues** (stuck states, unexpected behavior, workarounds)
✅ **Debug guides** (systematic debugging, diagnostic tools, log analysis)

### Don't Use For

❌ **FAQ** (use FAQ format for common questions)
❌ **Feature documentation** (use tutorial or reference)
❌ **Conceptual explanations** (use concept-explanation.md)
❌ **Step-by-step tutorials** (use tutorial-structure.md)

---

## Template Structure

### 1. Title and Scope

**Purpose**: Set expectations for what troubleshooting is covered

**Structure**:
```markdown
# Troubleshooting [System/Feature/Tool]

**Purpose**: Diagnose and resolve common issues with [system/feature]
**Scope**: Covers [what's included], see [other guide] for [what's excluded]
**Last Updated**: [Date]

## How to Use This Guide

1. Find your symptom in the issue list
2. Verify symptoms match your situation
3. Follow diagnosis steps to identify root cause
4. Apply recommended solution
5. If unresolved, see [escalation path]
```

**Example**:
```markdown
# Troubleshooting BAIME Methodology Development

**Purpose**: Diagnose and resolve common issues during BAIME experiments
**Scope**: Covers iteration execution, value scoring, convergence issues. See [BAIME Usage Guide] for workflow questions.
**Last Updated**: 2025-10-19

## How to Use This Guide

1. Find your symptom in the issue list below
2. Read the diagnosis section to identify root cause
3. Follow step-by-step solution
4. Verify fix worked by checking "Success Indicators"
5. If still stuck, see [Getting Help](#getting-help) section
```

---

### 2. Issue Index

**Purpose**: Help users quickly navigate to their problem

**Structure**:
```markdown
## Common Issues

**[Category 1]**:
- [Issue 1: Symptom summary](#issue-1-details)
- [Issue 2: Symptom summary](#issue-2-details)

**[Category 2]**:
- [Issue 3: Symptom summary](#issue-3-details)
- [Issue 4: Symptom summary](#issue-4-details)

**Quick Diagnosis**:
| If you see... | Likely issue | Jump to |
|---------------|--------------|---------|
| [Symptom] | [Issue name] | [Link] |
| [Symptom] | [Issue name] | [Link] |
```

**Example**:
```markdown
## Common Issues

**Iteration Execution Problems**:
- [Value scores not improving](#value-scores-not-improving)
- [Iterations taking too long](#iterations-taking-too-long)
- [Can't reach convergence](#cant-reach-convergence)

**Methodology Quality Issues**:
- [Low V_meta Reusability](#low-reusability)
- [Patterns not transferring](#patterns-not-transferring)

**Quick Diagnosis**:
| If you see... | Likely issue | Jump to |
|---------------|--------------|---------|
| V_instance/V_meta stuck or decreasing | Value scores not improving | [Link](#value-scores-not-improving) |
| V_meta Reusability < 0.60 | Patterns too project-specific | [Link](#low-reusability) |
| >7 iterations without convergence | Unrealistic targets or missing patterns | [Link](#cant-reach-convergence) |
```

---

### 3. Issue Template (Repeat for Each Issue)

**Purpose**: Systematic problem-diagnosis-solution structure

**Structure**:

```markdown
### Issue: [Issue Name]

#### Symptoms

**What you observe**:
- [Observable symptom 1]
- [Observable symptom 2]
- [Observable symptom 3]

**Example**:
```[format]
[Concrete example showing the problem]
```

**Not this issue if**:
- [Condition that rules out this issue]
- [Alternative explanation]

---

#### Diagnosis

**Root Causes** (one or more):

**Cause 1: [Root cause name]**

**How to verify**:
1. [Check step 1]
2. [Check step 2]
3. [Expected finding if this is the cause]

**Evidence**:
```[format]
[What evidence looks like for this cause]
```

**Cause 2: [Root cause name]**
[Same structure]

**Diagnostic Decision Tree**:
→ If [condition]: Likely Cause 1
→ Else if [condition]: Likely Cause 2
→ Otherwise: See [related issue]

---

#### Solutions

**Solution for Cause 1**:

**Step-by-step fix**:
1. [Action step 1]
   ```[language]
   [Code or command if applicable]
   ```
2. [Action step 2]
3. [Action step 3]

**Why this works**: [Explanation of solution mechanism]

**Time estimate**: [How long solution takes]

**Success indicators**:
- ✅ [How to verify fix worked]
- ✅ [Expected outcome]

**If solution doesn't work**:
- Check [alternative cause]
- See [related issue]

---

**Solution for Cause 2**:
[Same structure]

---

#### Prevention

**How to avoid this issue**:
- [Preventive measure 1]
- [Preventive measure 2]

**Early warning signs**:
- [Sign that issue is developing]
- [Metric to monitor]

**Best practices**:
- [Practice that prevents this issue]

---

#### Related Issues

- [Related issue 1] - [When to check]
- [Related issue 2] - [When to check]

**See also**:
- [Related documentation]
```

---

### 4. Full Example

```markdown
### Issue: Value Scores Not Improving

#### Symptoms

**What you observe**:
- V_instance or V_meta stuck across iterations (ΔV < 0.05)
- Value scores decreasing instead of increasing
- Multiple iterations (3+) without meaningful progress

**Example**:
```
Iteration 0: V_instance = 0.35, V_meta = 0.25
Iteration 1: V_instance = 0.37, V_meta = 0.28  (minimal Δ)
Iteration 2: V_instance = 0.34, V_meta = 0.30  (instance decreased!)
Iteration 3: V_instance = 0.36, V_meta = 0.31  (still stuck)
```

**Not this issue if**:
- Only 1-2 iterations completed (need more data)
- Scores are improving but slowly (ΔV = 0.05-0.10 is normal)
- Just hit temporary plateau (common at 0.60-0.70)

---

#### Diagnosis

**Root Causes**:

**Cause 1: Solving symptoms, not root problems**

**How to verify**:
1. Review problem identification from iteration-N.md "Problems" section
2. Check if problems describe symptoms (e.g., "low coverage") vs root causes (e.g., "no testing strategy")
3. Review solutions attempted - do they address why problem exists?

**Evidence**:
```markdown
❌ Symptom-based problem: "Test coverage is only 65%"
❌ Symptom-based solution: "Write more tests"
❌ Result: Coverage increased but tests brittle, V_instance stagnant

✅ Root-cause problem: "No systematic testing strategy"
✅ Root-cause solution: "Create TDD workflow, extract test patterns"
✅ Result: Better tests, sustainable coverage, V_instance improved
```

**Cause 2: Strategy not evidence-based**

**How to verify**:
1. Check if iteration-N-strategy.md references data artifacts
2. Look for phrases like "seems like", "probably", "might" (speculation)
3. Verify each planned improvement has supporting evidence

**Evidence**:
```markdown
❌ Speculative strategy: "Let's add integration tests because they seem useful"
❌ No supporting data

✅ Evidence-based strategy: "Data shows 80% of bugs in API layer (see data/bug-analysis.md), prioritize API tests"
✅ Clear data reference
```

**Cause 3: Scope too broad**

**How to verify**:
1. Count problems being addressed in current iteration
2. Check if all problems fully solved vs partially addressed
3. Review time spent per problem

**Evidence**:
```markdown
❌ Iteration 2 plan: Fix 7 problems (coverage, CI/CD, docs, errors, deps, perf, security)
❌ Result: All partially done, none complete, scores barely moved

✅ Iteration 2 plan: Fix top 2 problems (test strategy + coverage analysis)
✅ Result: Both fully solved, V_instance +0.15
```

**Diagnostic Decision Tree**:
→ If problem statements describe symptoms: Cause 1 (symptoms not root causes)
→ Else if strategy lacks data references: Cause 2 (not evidence-based)
→ Else if >4 problems in iteration plan: Cause 3 (scope too broad)
→ Otherwise: Check value function definitions (may be miscalibrated)

---

#### Solutions

**Solution for Cause 1: Root Cause Analysis**

**Step-by-step fix**:
1. **For each identified problem, ask "Why?" 3 times**:
   ```
   Problem: "Test coverage is low"
   Why? → "We don't have enough tests"
   Why? → "Writing tests is slow and unclear"
   Why? → "No systematic testing strategy or patterns"
   ✅ Root cause: "No testing strategy"
   ```

2. **Reframe problems as root causes**:
   - Before: "Coverage is 65%" (symptom)
   - After: "No systematic testing strategy prevents sustainable coverage" (root cause)

3. **Design solutions that address root causes**:
   ```markdown
   Root cause: No testing strategy
   Solution: Create TDD workflow, extract test patterns
   Outcome: Strategy enables sustainable testing
   ```

4. **Update iteration-N.md "Problems" section with reframed problems**

**Why this works**: Addressing root causes creates sustainable improvement. Symptom fixes are temporary.

**Time estimate**: 30-60 minutes to reframe problems and redesign strategy

**Success indicators**:
- ✅ Problems describe "why" things aren't working, not just "what" is broken
- ✅ Solutions create systems/patterns that prevent problem recurrence
- ✅ Next iteration shows measurable V_instance/V_meta improvement (ΔV ≥ 0.10)

**If solution doesn't work**:
- Check if root cause analysis went deep enough (may need 5 "why"s instead of 3)
- Verify solutions actually address identified root cause
- See [Can't reach convergence](#cant-reach-convergence) if problem persists

---

**Solution for Cause 2: Evidence-Based Strategy**

**Step-by-step fix**:
1. **For each planned improvement, identify supporting evidence**:
   ```markdown
   Planned: "Improve test coverage"
   Evidence needed: "Which areas lack coverage? Why? What's the impact?"
   ```

2. **Collect data to support or refute each improvement**:
   ```bash
   # Example: Collect coverage data
   go test -coverprofile=coverage.out ./...
   go tool cover -func=coverage.out | sort -k3 -n

   # Document findings
   echo "Analysis: 80% of uncovered code is in pkg/api/" > data/coverage-analysis.md
   ```

3. **Reference data artifacts in strategy**:
   ```markdown
   Improvement: Prioritize API test coverage
   Evidence: coverage-analysis.md shows 80% of gaps in pkg/api/
   Expected impact: Coverage +15%, V_instance +0.10
   ```

4. **Review strategy.md - should have ≥2 data references per improvement**

**Why this works**: Evidence-based decisions have higher success rate than speculation.

**Time estimate**: 1-2 hours for data collection and analysis

**Success indicators**:
- ✅ iteration-N-strategy.md references data artifacts (≥2 per improvement)
- ✅ Can show "before" data that motivated improvement
- ✅ Improvements address measured gaps, not hypothetical issues

---

**Solution for Cause 3: Narrow Scope**

**Step-by-step fix**:
1. **List all identified problems with estimated impact**:
   ```markdown
   Problems:
   1. No testing strategy - Impact: +0.20 V_instance
   2. Low coverage - Impact: +0.10 V_instance
   3. No CI/CD - Impact: +0.05 V_instance
   4. Docs incomplete - Impact: +0.03 V_instance
   [7 more...]
   ```

2. **Sort by impact, select top 2-3**:
   ```markdown
   Iteration N priorities:
   1. Create testing strategy (+0.20 impact) ✅
   2. Improve coverage (+0.10 impact) ✅
   3. [Defer remaining 9 problems]
   ```

3. **Allocate time: 80% to top 2, 20% to #3**:
   ```
   Testing strategy: 3 hours
   Coverage improvement: 2 hours
   Other: 1 hour
   ```

4. **Update iteration-N.md "Priorities" section with focused list**

**Why this works**: Better to solve 2 problems completely than 5 problems partially. Depth > breadth.

**Time estimate**: 15-30 minutes to prioritize and revise plan

**Success indicators**:
- ✅ Iteration plan addresses 2-3 problems maximum
- ✅ Each problem has 1+ hours allocated
- ✅ Problems are fully resolved (not partially addressed)

---

#### Prevention

**How to avoid this issue**:
- **Honest baseline assessment** (Iteration 0): Low scores are expected, they're measurement not failure
- **Problem root cause analysis**: Always ask "why" 3-5 times
- **Evidence-driven planning**: Collect data before deciding what to fix
- **Narrow focus per iteration**: 2-3 high-impact problems, fully solved

**Early warning signs**:
- ΔV < 0.05 for first time (investigate immediately)
- Problem list growing instead of shrinking (scope creep)
- Strategy document lacks data references (speculation)

**Best practices**:
- Spend 20% of iteration time on data collection
- Document evidence in data/ artifacts
- Review previous iteration to understand what worked
- Prioritize ruthlessly (defer ≥50% of identified problems)

---

#### Related Issues

- [Can't reach convergence](#cant-reach-convergence) - If stuck after 7+ iterations
- [Iterations taking too long](#iterations-taking-too-long) - If time is constraint
- [Low V_meta Reusability](#low-reusability) - If methodology not transferring

**See also**:
- [BAIME Usage Guide: When value scores don't improve](../baime-usage.md#faq)
- [Evidence collection patterns](../patterns/evidence-collection.md)
```

---

## Quality Checklist

Before publishing, verify:

### Content Quality

- [ ] **Completeness**: All common issues covered?
- [ ] **Accuracy**: Solutions tested and verified working?
- [ ] **Diagnosis depth**: Root causes identified, not just symptoms?
- [ ] **Evidence**: Concrete examples for each symptom/cause/solution?

### Structure Quality

- [ ] **Issue index**: Easy to find relevant issue?
- [ ] **Consistent format**: All issues follow same structure?
- [ ] **Progressive detail**: Symptoms → Diagnosis → Solutions flow?
- [ ] **Cross-references**: Links to related issues and docs?

### Solution Quality

- [ ] **Actionable**: Step-by-step instructions clear?
- [ ] **Verifiable**: Success indicators defined?
- [ ] **Complete**: Handles "if doesn't work" scenarios?
- [ ] **Realistic**: Time estimates provided?

### User Experience

- [ ] **Quick navigation**: Can find issue in <1 minute?
- [ ] **Self-service**: Can solve without external help?
- [ ] **Escalation path**: Clear what to do if stuck?
- [ ] **Prevention guidance**: Helps avoid issue in future?

---

## Adaptation Guide

### For Different Domains

**Error Troubleshooting** (HTTP errors, exceptions):
- Focus on error codes, stack traces, log analysis
- Include common error messages verbatim
- Add debugging tool usage (debuggers, profilers)

**Performance Issues** (slow queries, memory leaks):
- Focus on metrics, profiling, bottleneck identification
- Include before/after performance data
- Add monitoring and alerting guidance

**Configuration Problems** (startup failures, invalid config):
- Focus on configuration validation, common misconfigurations
- Include example correct configs
- Add validation tools and commands

**Integration Issues** (API failures, auth problems):
- Focus on request/response analysis, credential validation
- Include curl/Postman examples
- Add network debugging tools

### Depth Guidelines

**Issue coverage**:
- **Essential**: Top 10 most common issues (80% of user problems)
- **Important**: Next 20 issues (15% of problems)
- **Reference**: Remaining issues (5% of problems)

**Solution depth**:
- **Common issues**: Full diagnosis + multiple solutions + examples
- **Rare issues**: Brief description + link to external resources
- **Edge cases**: Acknowledge existence + escalation path

---

## Common Mistakes to Avoid

### ❌ Mistake 1: Vague Symptoms

**Bad**:
```markdown
### Issue: Things aren't working

**Symptoms**: Tool doesn't work correctly
```

**Good**:
```markdown
### Issue: Build Fails with "Module not found" Error

**Symptoms**:
- Build command exits with error code 1
- Error message: "Error: Cannot find module './config'"
- Occurs after npm install, before npm start
```

### ❌ Mistake 2: Solutions Without Diagnosis

**Bad**:
```markdown
### Issue: Slow performance

**Solution**: Try turning it off and on again
```

**Good**:
```markdown
### Issue: Slow API Responses (>2s)

#### Diagnosis
**Cause: Database query N+1 problem**
- Check: Log shows 100+ queries per request
- Check: Each query takes <10ms but total >2s
- Evidence: ORM lazy loading on collection

#### Solution
1. Add eager loading: .include('relations')
2. Verify with query count (should be 2-3 queries)
```

### ❌ Mistake 3: Missing Success Indicators

**Bad**:
```markdown
### Solution
1. Run this command
2. Restart the server
3. Hope it works
```

**Good**:
```markdown
### Solution
1. Run: `npm cache clean --force`
2. Restart server: `npm start`

**Success indicators**:
- ✅ Server starts without errors
- ✅ Module found in node_modules/
- ✅ App loads at http://localhost:3000
```

---

## Template Variables

Customize these for your domain:

- `[System/Feature/Tool]` - What's being troubleshot
- `[Issue Name]` - Descriptive issue title
- `[Category]` - Logical grouping of issues
- `[Symptom]` - Observable problem
- `[Root Cause]` - Underlying reason
- `[Solution]` - Fix steps
- `[Time Estimate]` - How long fix takes

---

## Validation Checklist

Test your troubleshooting guide:

1. **Coverage test**: Are 80%+ of common issues documented?
2. **Navigation test**: Can user find their issue in <1 minute?
3. **Solution test**: Can user apply solution successfully?
4. **Completeness test**: Are all 4 sections (symptoms, diagnosis, solution, prevention) present for each issue?
5. **Accuracy test**: Have solutions been tested and verified?

**If any test fails, revise before publishing.**

---

## Version History

- **1.0** (2025-10-19): Initial template created from documentation methodology iteration 2

---

**Ready to use**: Apply this template to create systematic, effective troubleshooting documentation for any system or tool.
