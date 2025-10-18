# Bootstrap-003: Error Recovery Methodology - Iteration Execution Prompts

**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Experiment**: Bootstrap-003 Error Recovery Methodology
**Meta-Agent**: M₀ (5 capabilities: observe, plan, execute, reflect, evolve)

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

**Note**: Error recovery may require Meta-Agent evolution for error-specific coordination. This is acceptable if justified by domain complexity.

---

## Phase 1: OBSERVE (Data Collection)

**Objective**: Gather empirical data about error patterns, frequencies, and impact

### Observation Tasks

```
observe_phase :: Agents → Observations
observe_phase(A) = sequential_execution(
  error_rate_measurement,
  error_pattern_analysis,
  error_impact_assessment,
  existing_error_handling_review
)
```

### Specific Actions

**1. Measure Current Error Rate**

```bash
# Session-level error statistics
cd /home/yale/work/meta-cc/experiments/bootstrap-003-error-recovery
meta-cc get-session-stats --scope project > data/session-stats-iteration-{n}.jsonl

# All error tool calls
meta-cc query-tools --status error --scope project \
  > data/error-tool-calls-iteration-{n}.jsonl

# Count errors
jq -s 'length' data/error-tool-calls-iteration-{n}.jsonl > data/error-count-iteration-{n}.txt
```

**Save outputs** to `data/error-*.jsonl` and `data/*-iteration-{n}.txt`

**2. Analyze Error Patterns by Tool**

```bash
# Errors by tool type
meta-cc query-tools --status error --scope project | \
  jq -r '.tool' | sort | uniq -c | sort -rn > data/errors-by-tool-iteration-{n}.txt

# Edit tool errors (likely code errors)
meta-cc query-tools --status error --tool Edit --scope project \
  > data/edit-errors-iteration-{n}.jsonl

# Bash tool errors (likely command errors)
meta-cc query-tools --status error --tool Bash --scope project \
  > data/bash-errors-iteration-{n}.jsonl

# Read tool errors (likely file not found)
meta-cc query-tools --status error --tool Read --scope project \
  > data/read-errors-iteration-{n}.jsonl
```

**3. Identify Error-Prone Files**

```bash
# Files frequently involved in errors
meta-cc query-tools --status error --tool Edit --scope project | \
  jq -r '.parameters.file_path // empty' | sort | uniq -c | sort -rn | head -20 \
  > data/error-prone-files-iteration-{n}.txt

# Error-prone tool sequences
meta-cc query-tool-sequences --scope project | \
  grep -i error > data/error-sequences-iteration-{n}.txt || true
```

**4. Analyze Error Context**

For high-frequency error signatures, get context:

```bash
# Example: Get context for a specific error pattern
# meta-cc query-context --error-signature "<pattern>" --window 3 \
#   > data/error-context-<pattern>-iteration-{n}.jsonl
```

**5. Categorize Errors (Manual Analysis)**

Review error samples and categorize:
- **Syntax errors**: Code syntax issues (Go compilation, linting)
- **Semantic errors**: Logic errors, type mismatches
- **Runtime errors**: File not found, permission denied, network errors
- **Integration errors**: API failures, tool integration issues
- **User errors**: Invalid parameters, incorrect usage
- **System errors**: Resource exhaustion, timeouts

**Save** categorization to `data/error-categorization-iteration-{n}.md`

### Observation Deliverables

- [ ] Error rate measured (total errors, error percentage)
- [ ] Errors categorized by tool type
- [ ] Error-prone files identified
- [ ] Error categories defined (initial taxonomy)
- [ ] Error impact assessed (blocking vs recoverable)
- [ ] Existing error handling patterns documented

---

## Phase 2: CODIFY (Pattern Extraction & Methodology)

**Objective**: Extract error handling patterns and document error recovery methodology

### Codification Tasks

```
codify_phase :: Observations → Methodology
codify_phase(obs) = pattern_extraction(
  classify_error_types,
  document_diagnostic_procedures,
  create_recovery_strategies,
  define_prevention_guidelines
)
```

### Specific Actions

**1. Create Error Classification Taxonomy**

Based on observed errors, create structured taxonomy:

**Document** in `knowledge/error-taxonomy-iteration-{n}.md`

**Taxonomy Template**:
```markdown
# Error Classification Taxonomy

## Error Categories

### Category 1: [Error Type Name]

**Definition**: [What this error type means]

**Examples**:
- [Example error 1]
- [Example error 2]

**Frequency**: [% of total errors]

**Impact**: [Blocking / Recoverable / Ignorable]

**Common Causes**:
- [Cause 1]
- [Cause 2]

**Detection**: [How to detect this error type]

### Category 2: [Error Type Name]
...
```

**Goal**: ≥10 error categories covering ≥90% of observed errors

**2. Document Root Cause Diagnosis Procedures**

For each major error category, create diagnostic workflow:

**Document** in `knowledge/diagnostic-workflows-iteration-{n}.md`

**Workflow Template**:
```markdown
## Diagnostic Workflow: [Error Category]

### Step 1: Identify Error Symptoms
[What to look for]

### Step 2: Gather Context
[What information to collect]

### Step 3: Analyze Root Cause
[How to determine the underlying issue]

### Step 4: Verify Diagnosis
[How to confirm the root cause]

**Estimated Time**: ~[X] minutes
**Tools Needed**: [List of tools/commands]
**Success Criteria**: [How to know diagnosis is correct]
```

**Goal**: ≥5 diagnostic workflows for most common error categories

**3. Create Recovery Strategy Patterns**

For each error category, document recovery approaches:

**Document** in `knowledge/recovery-patterns-iteration-{n}.md`

**Pattern Template**:
```markdown
### Recovery Pattern: [Pattern Name]

**Applicable to**: [Error categories]

**Strategy**: [High-level recovery approach]

**Steps**:
1. [Recovery step 1]
2. [Recovery step 2]
3. [Recovery step 3]

**Automation Potential**: [Manual / Semi-automated / Fully automated]

**Success Rate**: [Expected recovery success percentage]

**Time to Recovery**: ~[X] minutes
```

**Goal**: ≥5 recovery patterns covering major error categories

**4. Define Prevention Guidelines**

Based on error analysis, create prevention practices:

**Document** in `knowledge/prevention-guidelines-iteration-{n}.md`

**Guidelines Template**:
```markdown
# Error Prevention Guidelines

## Guideline 1: [Prevention Practice Name]

**Purpose**: Prevent [error category]

**Practice**:
[What to do / what to avoid]

**Example**:
[Good example]
[Bad example - what not to do]

**Enforcement**: [Manual / Linter / CI check]

## Guideline 2: [Prevention Practice Name]
...
```

**Goal**: ≥8 prevention guidelines

### Codification Deliverables

- [ ] Error classification taxonomy created (≥10 categories)
- [ ] Diagnostic workflows documented (≥5 workflows)
- [ ] Recovery strategy patterns defined (≥5 patterns)
- [ ] Prevention guidelines established (≥8 guidelines)
- [ ] All methodology artifacts in `knowledge/`

---

## Phase 3: AUTOMATE (Tool Creation & CI Integration)

**Objective**: Create automation tools for error detection, diagnosis, and recovery

### Automation Tasks

```
automate_phase :: Methodology → Tools
automate_phase(method) = tool_creation(
  build_error_detector,
  create_recovery_scripts,
  integrate_ci,
  implement_monitoring
)
```

### Specific Actions

**1. Build Error Pattern Detector**

Create script to detect error patterns proactively:

**Tool**: `scripts/detect-error-patterns.sh`

```bash
#!/bin/bash
# Purpose: Detect common error patterns in code/logs
# Usage: ./scripts/detect-error-patterns.sh [file_or_directory]

# Check for error-prone patterns:
# - Missing error checks (Go: functions returning error without check)
# - Ignored errors
# - Common antipatterns

# Output: Report of detected issues
```

**Expected speedup**: 10x (vs manual code review)

**2. Create Recovery Script Library**

Create scripts for automated recovery of common errors:

**Tool**: `scripts/auto-recover.sh`

```bash
#!/bin/bash
# Purpose: Attempt automated recovery for known error types
# Usage: ./scripts/auto-recover.sh [error_type]

case "$1" in
  "file-not-found")
    # Recovery logic for file not found
    ;;
  "permission-denied")
    # Recovery logic for permission issues
    ;;
  "timeout")
    # Recovery logic for timeouts (retry with backoff)
    ;;
  *)
    echo "No automated recovery available for: $1"
    ;;
esac
```

**Goal**: Automate recovery for ≥3 common error types

**3. Integrate CI/CD Error Detection**

Add error detection to CI pipeline:

**File**: `.github/workflows/error-detection.yml` (or update existing)

```yaml
name: Error Detection

on: [push, pull_request]

jobs:
  detect-errors:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Detect error patterns
        run: ./scripts/detect-error-patterns.sh .
      - name: Check error handling coverage
        run: |
          # Verify error checks present
          # Fail if critical errors unhandled
```

**4. Implement Error Monitoring**

Create error analytics:

**Tool**: `scripts/error-analytics.sh`

```bash
#!/bin/bash
# Purpose: Generate error analytics report
# Usage: ./scripts/error-analytics.sh

# Analyze error trends:
# - Error rate over time
# - Most common error types
# - Error-prone files/tools
# - MTTD (Mean Time To Diagnosis)
# - MTTR (Mean Time To Recovery)

# Output: Error analytics dashboard
```

**5. Create Comprehensive Error Recovery Guide** (Final Iteration)

Consolidate all patterns, workflows, and tools into single production-ready guide:

**File**: `knowledge/error-recovery-methodology-complete.md`

**Structure**:
- Overview
- Error Classification Taxonomy (≥10 categories)
- Diagnostic Workflows (≥5 workflows)
- Recovery Strategy Patterns (≥5 patterns)
- Prevention Guidelines (≥8 guidelines)
- Automation Tools (≥3 tools)
- Effectiveness Metrics
- Reusability Guide
- Troubleshooting

**Expected lines**: ~800-1,200 lines (comprehensive)

### Automation Deliverables

- [ ] Error pattern detector created (script in `scripts/`)
- [ ] Recovery script library created (≥3 recovery scripts)
- [ ] CI/CD integration complete
- [ ] Error monitoring implemented
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
V_instance(s_n) = 0.35·V_detection(s_n) +
                  0.30·V_diagnosis(s_n) +
                  0.20·V_recovery(s_n) +
                  0.15·V_prevention(s_n)
```

**Calculate each component**:

1. **V_detection**: [Current value and evidence]
   - Error detection coverage: [%]
   - Monitoring completeness: [assessment]
   - Score: [0.0-1.0]
   - Evidence: [Concrete measurements]

2. **V_diagnosis**: [Current value and evidence]
   - Root cause identification rate: [%]
   - Mean time to diagnosis: [minutes]
   - Score: [0.0-1.0]
   - Evidence: [Diagnostic success rate, time measurements]

3. **V_recovery**: [Current value and evidence]
   - Recovery success rate: [%]
   - Mean time to recovery: [minutes]
   - Automation level: [% automated]
   - Score: [0.0-1.0]
   - Evidence: [Recovery metrics]

4. **V_prevention**: [Current value and evidence]
   - Error rate reduction: [%]
   - Prevention practices in place: [count]
   - Score: [0.0-1.0]
   - Evidence: [Error rate change, guidelines implemented]

**Final V_instance(s_n)**: 0.35·[score1] + 0.30·[score2] + 0.20·[score3] + 0.15·[score4] = **[total]**

### V_meta Calculation

**Components** (universal BAIME rubrics):

```
V_meta(s_n) = 0.40·V_methodology_completeness(s_n) +
              0.30·V_methodology_effectiveness(s_n) +
              0.30·V_methodology_reusability(s_n)
```

**Calculate each component**:

1. **V_methodology_completeness**: [Current value and evidence]
   - Taxonomy complete: [Y/N, category count]
   - Workflows documented: [Y/N, workflow count]
   - Patterns defined: [Y/N, pattern count]
   - Guidelines established: [Y/N, guideline count]
   - Score: [0.0-1.0]
   - Evidence: [Artifact counts]

2. **V_methodology_effectiveness**: [Current value and evidence]
   - Speedup vs ad-hoc: [X]x
   - Error rate reduction: [%]
   - MTTD improvement: [%]
   - MTTR improvement: [%]
   - Score: [0.0-1.0]
   - Evidence: [Before/after measurements]

3. **V_methodology_reusability**: [Current value and evidence]
   - Transferability assessment: [% modification for other projects]
   - Domain independence: [% universal vs domain-specific]
   - Score: [0.0-1.0]
   - Evidence: [Transfer test if available]

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
3. ✅ / ❌ M_n == M_{n-1}: [comparison] [status] (Note: Mₙ ≠ M₀ acceptable for error domain)
4. ✅ / ❌ A_n == A_{n-1}: [comparison] [status]
5. ✅ / ❌ ΔV_instance < 0.02: [change] [status]
6. ✅ / ❌ ΔV_meta < 0.02: [change] [status]

**Convergence Status**: [CONVERGED / NOT CONVERGED]

**If NOT CONVERGED**: Continue to Phase 5 (Plan Next Iteration)
**If CONVERGED**: Proceed to Finalization (create results.md)

### Evaluation Deliverables

- [ ] V_instance(s_n) calculated with component breakdown
- [ ] V_meta(s_n) calculated with component breakdown
- [ ] Convergence criteria checked (all 6 criteria)
- [ ] Convergence status determined
- [ ] Value trajectory updated

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

Identify what's missing:

**Instance gaps** (V_instance < 0.80):
- Detection gap: [What error types not detected]
- Diagnosis gap: [What diagnostic procedures missing]
- Recovery gap: [What recovery strategies missing]
- Prevention gap: [What prevention practices missing]

**Meta gaps** (V_meta < 0.80):
- Completeness gap: [What documentation missing]
- Effectiveness gap: [What measurements/validations missing]
- Reusability gap: [What transfer evidence missing]

**2. Meta-Agent Evolution Decision**

```
meta_evolution :: M_{n} → M_{n+1}
meta_evolution(M) =
  if error_coordination_capability_needed
  then M' = M ∪ {error_coordination}
  else M' = M
```

**Decision**: ✅ Keep M₀ unchanged / ⚠️ Add capability: [capability name]

**Justification**: [Why M₀ sufficient / Why error-specific coordination needed]

**Note**: Error domain is complex; evolution to Mₙ ≠ M₀ is acceptable if justified.

**3. Agent Set Evolution Decision**

```
agent_evolution :: A_{n} → A_{n+1}
agent_evolution(A) =
  if error_specialization_valuable ∧ efficiency_gain > 2x
  then A' = A ∪ {error_specialized_agent}
  else A' = A
```

**Decision**: ✅ Keep generic agents / ⚠️ Create specialized agent: [agent name]

**Justification**: [Why generic sufficient / Why specialization needed]

**Potential Specialized Agents** (based on previous execution):
- error-classifier: Classify errors into taxonomy
- root-cause-analyzer: Perform root cause analysis
- recovery-strategist: Design recovery strategies
- prevention-advisor: Recommend preventive measures

**4. Next Iteration Focus**

Based on gap analysis:

**Primary objective**: [Focus area]
**Secondary objective**: [Focus area]
**Stretch goal**: [Focus area]

**Expected progress**:
- V_instance: [current] → [target]
- V_meta: [current] → [target]

### Evolution Deliverables

- [ ] Gap analysis completed
- [ ] Meta-agent evolution decision (with justification)
- [ ] Agent set evolution decision (with justification)
- [ ] Next iteration focus planned
- [ ] Expected progress targets set

---

## Domain-Specific Guidance

### Error Recovery Best Practices

1. **Start with high-frequency errors**: 80/20 rule - focus on most common errors first
2. **Prioritize by impact**: Blocking errors > Recoverable > Ignorable
3. **Measure everything**: MTTD, MTTR, recovery success rate, error rate
4. **Automate incrementally**: Start with detection, then diagnosis, then recovery
5. **Test recovery procedures**: Verify recovery works before deploying
6. **Document edge cases**: Unusual errors need special handling
7. **Learn from failures**: Every unrecovered error is learning opportunity

### Error Taxonomy Principles

- **MECE (Mutually Exclusive, Collectively Exhaustive)**: No overlap, no gaps
- **Actionable**: Each category should have clear recovery path
- **Observable**: Each category should have detectable symptoms
- **Hierarchical**: Top-level categories with sub-categories if needed

### Recovery Strategy Design

- **Fast path**: Automated recovery for common errors
- **Guided path**: Semi-automated recovery with human verification
- **Manual path**: Clear procedures for complex errors
- **Escalation**: When to give up and escalate

---

## Iteration Output Structure

Each iteration should produce `iteration-{n}.md` with:

```markdown
# Iteration {n}: [Focus Name]

**Date**: YYYY-MM-DD
**Duration**: ~[X] hours
**Status**: [Completed]

## Executive Summary
[What was accomplished]

## Pre-Execution Context
- Previous V_instance: [score]
- Previous V_meta: [score]
- Focus: [primary objective]

## OBSERVE Phase
[Error analysis, pattern identification]

**Deliverables**:
- [Data files created]

## CODIFY Phase
[Taxonomy, workflows, patterns created]

**Deliverables**:
- [Knowledge files created]

## AUTOMATE Phase
[Tools created, CI integration]

**Deliverables**:
- [Scripts/automation created]

## EVALUATE Phase

### V_instance Components
[Breakdown with scores]

**V_instance(s_{n})** = [calculation] = **[total]**

### V_meta Components
[Breakdown with scores]

**V_meta(s_{n})** = [calculation] = **[total]**

### Convergence Check
[6 criteria checked]

**Status**: [CONVERGED / NOT CONVERGED]

## EVOLVE Phase
[Gap analysis, evolution decisions]

## Iteration Summary
- V_instance: [score] ([change])
- V_meta: [score] ([change])
- Error rate: [rate] ([change])
- Key achievements: [list]
```

---

## Checklist: Before Starting Each Iteration

- [ ] Read previous iteration file
- [ ] Review V_instance and V_meta scores
- [ ] Review error rate trends
- [ ] Load Meta-Agent M₀ (or Mₙ if evolved)
- [ ] Load agent definitions
- [ ] Review README.md objectives
- [ ] Ensure data/ and knowledge/ directories ready

---

## Checklist: After Completing Each Iteration

- [ ] Created `iteration-{n}.md`
- [ ] Calculated V_instance with breakdown
- [ ] Calculated V_meta with breakdown
- [ ] Checked convergence (6 criteria)
- [ ] Saved error data to `data/`
- [ ] Saved methodology to `knowledge/`
- [ ] Saved scripts to `scripts/`
- [ ] Documented error rate change
- [ ] Planned next iteration (if not converged)

---

**Template Version**: 1.0
**Created**: 2025-10-18
**Based on**: BAIME-ITERATION-PROMPTS-TEMPLATE.md and Bootstrap-003 domain requirements
**Domain**: Error Detection, Diagnosis, and Recovery

**Usage**: Call iteration-executor agent for each iteration with this guidance
