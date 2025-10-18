# BAIME Experiment Template

**Version**: 1.0
**Based on**: Bootstrap-002 Test Strategy Development (successful BAIME application)
**Purpose**: Standardized template for BAIME (Bootstrapped AI Methodology Engineering) experiments
**Framework**: OCA cycle + Empirical Methodology + Value Optimization

---

## Quick Start Guide

### 1. Create Experiment Directory

```bash
mkdir -p experiments/bootstrap-NNN-[domain-name]/{data,knowledge,agents}
cd experiments/bootstrap-NNN-[domain-name]
```

### 2. Copy This Template

```bash
# Create three files:
cp experiments/BAIME-EXPERIMENT-TEMPLATE.md README.md
cp experiments/BAIME-ITERATION-PROMPTS-TEMPLATE.md ITERATION-PROMPTS.md
# Edit both files with domain-specific content
```

### 3. Fill in Domain-Specific Content

Replace all `[DOMAIN]`, `[TARGET]`, `[SCOPE]` placeholders with actual values.

### 4. Execute

Call iteration-executor agent for each iteration until convergence.

---

## README.md Template

```markdown
# Bootstrap-NNN: [Domain] Methodology

**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: [NOT STARTED / IN PROGRESS / CONVERGED]
**Created**: YYYY-MM-DD
**Domain**: [Software Testing / Documentation / Performance / etc.]

---

## Overview

This experiment applies the BAIME framework to develop a systematic [domain] methodology through observation of agent [domain] patterns in the [project-name] project. The experiment uses the complete three-layer architecture (bootstrapped-se + empirical-methodology + value-optimization) to achieve both task completion and methodology development.

**Previous executions**: [None / List previous attempts with convergence results]
**This execution**: [First execution / Re-implementing with BAIME framework]

---

## Objectives

### Meta-Objective (Methodology Development Layer)

**Develop systematic [domain] methodology through observation of agent [domain] patterns**

Apply BAIME's OCA cycle:
- **Observe**: Collect data on [domain] patterns, [metric1] evolution, [metric2] execution
- **Codify**: Extract patterns into reusable [domain] methodology
- **Automate**: Create tools and CI checks for [domain] enforcement

**Expected Output**: (O, Aₙ, Mₙ)
- O = [Domain] artifacts and improved [target metrics]
- Aₙ = Converged agent set (generic or specialized)
- Mₙ = Meta-Agent M₀ (expected to remain stable)

### Instance Objective (Task Execution Layer)

**Improve [specific target metric/quality] for [project-name] project**

**Target**: [Specific files/packages] (~[N] lines of code)

**Scope**:
- Achieve [criterion 1]: [e.g., ≥80% coverage, <10ms latency, etc.]
- Implement [criterion 2]: [e.g., systematic workflow, quality gates]
- Create [criterion 3]: [e.g., automation tools, integration patterns]
- Establish quality gates (≥[N]/[M] criteria)

**Deliverables**:
- [Deliverable 1]: [e.g., Improved test suite with ≥80% coverage]
- [Deliverable 2]: [e.g., Pattern library with N patterns]
- [Deliverable 3]: [e.g., Coverage-driven gap closure workflow]
- [Deliverable 4]: [e.g., Automated quality gates]

---

## Value Functions

### V_instance: [Domain] Implementation Quality

```
V_instance(s) = w₁·V_[component1] +    -- [Description of component1]
                w₂·V_[component2] +    -- [Description of component2]
                w₃·V_[component3] +    -- [Description of component3]
                w₄·V_[component4]      -- [Description of component4]
```

**Component Definitions**:

1. **V_[component1]** ([Description]):
   - 1.0: [Excellent criteria]
   - 0.8: [Good criteria]
   - 0.6: [Acceptable criteria]
   - 0.4: [Poor criteria]
   - <0.4: [Inadequate criteria]

2. **V_[component2]** ([Description]):
   - 1.0: [Excellent criteria]
   - 0.8: [Good criteria]
   - 0.6: [Acceptable criteria]
   - 0.4: [Poor criteria]
   - <0.4: [Inadequate criteria]

3. **V_[component3]** ([Description]):
   - 1.0: [Excellent criteria]
   - 0.8: [Good criteria]
   - 0.6: [Acceptable criteria]
   - 0.4: [Poor criteria]
   - <0.4: [Inadequate criteria]

4. **V_[component4]** ([Description]):
   - 1.0: [Excellent criteria]
   - 0.8: [Good criteria]
   - 0.6: [Acceptable criteria]
   - 0.4: [Poor criteria]
   - <0.4: [Inadequate criteria]

### V_meta: [Domain] Methodology Quality

```
V_meta(s) = 0.40·V_methodology_completeness +    -- Documentation quality
            0.30·V_methodology_effectiveness +   -- Practical impact
            0.30·V_methodology_reusability       -- Transferability
```

**Component Definitions** (Universal BAIME rubrics):

1. **V_methodology_completeness**:
   - 1.0: Complete process + criteria + examples + edge cases + rationale
   - 0.8: Complete workflow + criteria, missing examples/edge cases
   - 0.6: Step-by-step procedures, missing decision criteria
   - <0.6: Observational notes only, no structured process

2. **V_methodology_effectiveness**:
   - 1.0: >10x speedup vs ad-hoc, >50% quality improvement
   - 0.8: 5-10x speedup, 20-50% quality improvement
   - 0.6: 2-5x speedup, 10-20% quality improvement
   - <0.6: <2x speedup, no measurable quality gain

3. **V_methodology_reusability**:
   - 1.0: <15% modification needed, nearly universal
   - 0.8: 15-40% modification, minor tweaks
   - 0.6: 40-70% modification, some adaptation
   - <0.6: >70% modification, highly specialized

---

## Convergence Criteria

**Standard Dual Convergence** (expected pattern):

```
CONVERGED iff:
  1. V_instance(s_n) ≥ 0.80        -- Task quality threshold
  2. V_meta(s_n) ≥ 0.80            -- Methodology quality threshold
  3. M_n == M_{n-1}                -- Meta-Agent stable
  4. A_n == A_{n-1}                -- Agent set stable
  5. ΔV_instance < 0.02 (2+ iters) -- Instance convergence
  6. ΔV_meta < 0.02 (2+ iters)     -- Meta convergence
```

**Expected Iterations**: [N-M] (based on [low/medium/high] domain complexity)

**Alternative Convergence Patterns**:
- **Practical Convergence**: If quality evidence exceeds metrics (justified partial criteria)
- **Meta-Focused Convergence**: If V_meta ≥ 0.80 and methodology is primary value

---

## Data Sources

### Observation Tools (meta-cc MCP)

```bash
# [Domain]-specific analysis
[observation command 1]

# File access patterns (identify high-value targets)
meta-cc query-files --threshold 10

# Error patterns (guide [domain] design)
meta-cc query-tools --status error

# [Domain]-related conversations
meta-cc query-user-messages --pattern "[domain-keyword1]|[domain-keyword2]"

# [Domain] execution patterns
meta-cc query-tool-sequences --pattern "[pattern]"
```

### Baseline Metrics

Current state (to be measured in Iteration 0):
- [Metric 1]: TBD
- [Metric 2]: TBD
- [Metric 3]: TBD
- [Metric 4]: TBD

---

## Expected Agents

Based on BAIME principles (let specialization emerge from data):

**Initial agents** (generic, from M₀):
- coder: Write [domain] code
- data-analyst: Analyze [domain] data
- doc-writer: Document [domain] patterns

**Potential specialized agents** (create only if needed):
- [specialized-agent-1]: [Description of need]
- [specialized-agent-2]: [Description of need]
- [specialized-agent-3]: [Description of need]

**Decision criteria**: Create specialized agent only when:
- Generic agents insufficient (demonstrated over 2+ iterations)
- Specialization provides >2x efficiency gain
- Pattern will be reused across multiple files/modules

---

## BAIME Framework Application

### Phase 1: Observe (Empirical Foundation)

**Iteration 0-1**: Baseline establishment
- Measure current [metric1]
- Analyze [metric1] gaps
- Identify high-value [domain] targets
- Document existing [domain] patterns

**Data Collection**:
- [Data collection command 1]
- File access patterns (high-change files need [domain work])
- Error patterns (guide defensive [domain work])
- Existing [domain] structure analysis

### Phase 2: Codify (Pattern Extraction)

**Iteration 1-3**: Methodology development
- Extract successful [domain] patterns
- Document [domain]-driven workflow
- Create [resource] patterns
- Define quality criteria

**Artifacts**:
- [Domain] methodology documentation
- [Domain] pattern library
- [Metric]-driven workflow
- Quality gate definitions

### Phase 3: Automate (Tool Creation)

**Iteration 2-4**: Automation implementation
- CI/CD integration
- [Metric] gates (block if <[threshold])
- Automated [domain] templates
- [Metric] reporting dashboard

**Tools**:
- GitHub Actions workflow
- Makefile [domain] targets
- [Domain] scripts
- [Resource] generators

### Phase 4: Evaluate (Value Optimization)

**Every Iteration**:
- Calculate V_instance ([domain] quality)
- Calculate V_meta (methodology quality)
- Check convergence criteria
- Decide: continue or converge

### Phase 5: Evolve (Self-Improvement)

**If not converged**:
- Analyze gaps in current state
- Identify improvement opportunities
- Decide agent evolution (if needed)
- Plan next iteration focus

---

## Success Criteria

### Instance Success (Task Completion)

- [ ] [Metric 1] ≥[threshold] (target: [range])
- [ ] All critical paths covered
- [ ] [Resource type] with [quality criteria]
- [ ] [Quality metric] <[threshold]%
- [ ] CI integration complete
- [ ] Quality gates operational (≥[N]/[M])
- [ ] [Performance metric] acceptable (<[threshold])
- [ ] [Workflow name] documented

### Meta Success (Methodology Development)

- [ ] Complete [domain] methodology documented
- [ ] Reusable [domain] patterns created
- [ ] [Metric]-driven workflow codified
- [ ] Automation tools implemented
- [ ] Transferability ≥85% (to other [similar projects])
- [ ] Efficiency gain ≥10x vs ad-hoc [domain work]
- [ ] Methodology validated through application

---

## Context Management

**Estimated Context Allocation** (BAIME framework):
- Observation: 30% ([metric] analysis, gap identification)
- Codification: 40% (methodology documentation, pattern extraction)
- Automation: 20% (CI integration, tool creation)
- Reflection: 10% (value calculation, convergence assessment)

**Context Pressure Handling**:
- If usage >80%: Serialize state to `knowledge/`, split session
- If usage >50%: Use reference compression, link to files
- Target: Meta-Focused Convergence if context constrained

---

## Experiment Timeline

**Phase 1: Setup and Baseline** (Iteration 0)
- Duration: ~2 hours
- Deliverable: Baseline metrics, initial observation

**Phase 2: Iterative Improvement** (Iterations 1-[N])
- Duration: ~[M]-[P] hours (1.5-2.5 hours per iteration)
- Deliverable: Incremental [metric] improvements, methodology refinement

**Phase 3: Convergence and Documentation** (Final iteration)
- Duration: ~2 hours
- Deliverable: Final methodology, results analysis

**Total Estimated Duration**: [X]-[Y] hours ([N]-[M] iterations)

---

## Transferability Plan

**Expected Reusability**: [XX-YY]% (based on domain complexity)

**What Transfers**:
- [Universal component 1] (100%)
- [Universal component 2] ([XX]%)
- [Universal component 3] ([XX]%)
- [Universal component 4] ([XX]%)

**What Needs Adaptation**:
- Language-specific [component] ([Source Lang] → [Target Lang])
- [Framework]-specific details ([Source Framework] → [Target Framework])
- [Tool] commands ([source tool] → [target tool])

**Adaptation Effort**:
- Same language ([Source]): [X]% modification
- Similar language ([Source] → [Similar]): [Y]% modification
- Different paradigm ([Source] → [Different]): [Z]% modification

---

## Risk Assessment

**Technical Risks**:
- [Risk 1]: [Description and likelihood]
- [Risk 2]: [Description and likelihood]
- [Risk 3]: [Description and likelihood]

**Mitigation**:
- [Mitigation 1]
- [Mitigation 2]
- [Mitigation 3]

**Methodology Risks**:
- Over-[domain work] (diminishing returns)
- Under-[domain work] (gaps in critical paths)

**Mitigation**:
- Use value function to guide decisions
- Focus on high-change, high-value files first
- Quality over quantity (effective work, not just metrics)

---

## Related Experiments

**Synergies**:
- **Bootstrap-[NNN] ([Related Domain 1])**: [How they relate]
- **Bootstrap-[MMM] ([Related Domain 2])**: [How they relate]
- **Bootstrap-[PPP] ([Related Domain 3])**: [How they relate]

**Dependencies**:
- [None / List dependencies]

**Enables**:
- [Future experiment 1]
- [Future experiment 2]
- [Future experiment 3]

---

## References

**BAIME Framework**:
- [bootstrapped-ai-methodology-engineering.md](../../.claude/skills/bootstrapped-ai-methodology-engineering.md)
- [bootstrapped-se.md](../../.claude/skills/bootstrapped-se.md)
- [value-optimization.md](../../.claude/skills/value-optimization.md)

**Experiment Templates**:
- [EXPERIMENTS-OVERVIEW.md](../EXPERIMENTS-OVERVIEW.md)

**Domain Resources**:
- [Domain-specific resource 1]
- [Domain-specific resource 2]
- [Domain-specific resource 3]

---

**Version**: 1.0
**Status**: Ready to execute
**Next Step**: Create ITERATION-PROMPTS.md and begin execution
```

---

## ITERATION-PROMPTS.md Template

See `BAIME-ITERATION-PROMPTS-TEMPLATE.md` for complete iteration execution guidance using lambda expressions.

---

## Directory Structure

```
experiments/bootstrap-NNN-[domain-name]/
├── README.md                          # This template (filled in)
├── ITERATION-PROMPTS.md               # Execution guidance (lambda expressions)
├── iteration-0.md                     # Baseline iteration
├── iteration-1.md                     # First improvement iteration
├── iteration-N.md                     # Final convergence iteration
├── results.md                         # Comprehensive experiment analysis
├── data/                              # Raw data and measurements
│   ├── baseline-metrics.txt
│   ├── [metric]-iteration-N.txt
│   └── effectiveness-measurements.yaml
├── knowledge/                         # Codified methodology
│   ├── [domain]-methodology.md
│   ├── [domain]-patterns.md
│   └── [domain]-workflow.md
├── agents/                            # Specialized agent prompts (if created)
│   ├── [specialized-agent-1].md
│   └── [specialized-agent-2].md
└── scripts/                           # Automation tools (if created)
    ├── analyze-[domain].sh
    ├── generate-[resource].sh
    └── comprehensive-guide.md
```

---

## Usage Instructions

### Step 1: Customize README.md

1. Replace `[Domain]` with your domain (e.g., "Test Strategy", "API Design")
2. Replace `[TARGET]` with specific target (e.g., "`internal/` packages")
3. Replace `[SCOPE]` with quantifiable scope (e.g., "~5,000 lines")
4. Define domain-specific V_instance components (4 components, weights sum to 1.0)
5. Fill in convergence criteria (iterations estimate, thresholds)
6. List data sources and observation commands
7. Specify success criteria (checkboxes)

### Step 2: Customize ITERATION-PROMPTS.md

1. Define domain-specific workflow using lambda expressions
2. Specify observation phase data collection
3. Define codification artifacts
4. List automation tool requirements
5. Add evaluation formulas
6. Include domain-specific edge cases

### Step 3: Execute Experiment

```bash
# For each iteration (0, 1, 2, ...):
# Call iteration-executor agent with:
# - Current state s_{n-1}
# - Iteration number n
# - README.md objectives
# - ITERATION-PROMPTS.md guidance

# Continue until convergence criteria met
```

### Step 4: Create results.md

After convergence, create comprehensive results.md documenting:
- Complete experiment summary
- Convergence analysis with trajectory
- Three-tuple output (O, Aₙ, Mₙ)
- Transferability validation
- Lessons learned

---

## Key Success Factors (from Bootstrap-002)

### What Worked

1. **Dual value functions**: Track instance and meta progress independently
2. **Multi-context validation**: Test methodology across 3+ project archetypes
3. **Lambda expressions**: Use compact functional notation for workflows
4. **Automation first**: Tools provide 100x+ speedup over documentation alone
5. **Context allocation**: 30/40/20/10 (Observe/Codify/Automate/Reflect) works well
6. **Generic agents**: Let specialization emerge; don't force it
7. **Stable equilibrium**: Require 2-3 iterations of stability before declaring convergence

### What to Avoid

1. **Single value function**: Conflates task completion with methodology quality
2. **Single-context validation**: Insufficient to prove reusability
3. **Premature specialization**: Create specialized agents only when generic insufficient
4. **Forcing convergence**: Let data guide; don't rush to meet target iterations
5. **Skipping multi-context validation**: Critical for V_meta convergence
6. **Over-documentation**: Automation tools > extensive docs

---

## Template Checklist

Before starting experiment, ensure:

- [ ] README.md customized with domain-specific content
- [ ] V_instance components defined (4 components, domain-specific rubrics)
- [ ] V_meta components use universal BAIME rubrics (completeness, effectiveness, reusability)
- [ ] Convergence criteria specified (6 criteria)
- [ ] Data sources identified (observation commands)
- [ ] Success criteria listed (instance + meta)
- [ ] ITERATION-PROMPTS.md created with lambda expressions
- [ ] Directory structure created (data/, knowledge/, agents/, scripts/)
- [ ] Related experiments identified (synergies, dependencies)

---

**Template Version**: 1.0
**Created**: 2025-10-18
**Based on**: Bootstrap-002 (Full Dual Convergence, V_instance=0.80, V_meta=0.80)
**Validation**: Proven through 6 iterations, 3.1x speedup, 94.2% reusability

**Next**: Create `BAIME-ITERATION-PROMPTS-TEMPLATE.md` for execution guidance
