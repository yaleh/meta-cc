# Bootstrap-XXX: [Domain] Methodology

**Status**: [Not Started / In Progress / Completed]
**Start Date**: YYYY-MM-DD
**Priority**: [HIGH / MEDIUM / LOW]

---

## Experiment Objectives (Two-Layer Architecture)

### Meta-Objective (Meta-Agent Layer)

**Goal**: Develop [domain] methodology through iterative observation, codification, and automation of agent work patterns.

**Deliverables**:
- Methodology documents capturing observed patterns
- Best practices extracted from agent work
- Specialized agents created based on observed needs
- Reusable three-tuple (M, A, methodology artifacts)

**Success Criteria**:
- Convergence achieved: V(s) ≥ 0.80
- Meta-agent stable: M_n = M_{n-1}
- Agent set stable: A_n = A_{n-1}
- Methodology artifacts complete and transferable

### Instance Objective (Agent Layer)

**Goal**: Execute [concrete domain task] on [specific target].

**Concrete Scope**:
- Target files/modules: [specific paths]
- Quantifiable scope: [e.g., "100 functions", "5 API tools", "10 modules"]
- Clear deliverables: [actual work products]

**Success Criteria**:
- [Measurable outcome 1]: [e.g., "80% documentation coverage"]
- [Measurable outcome 2]: [e.g., "Reduce complexity by 30%"]
- [Measurable outcome 3]: [e.g., "All APIs follow naming convention"]

**Expected Agent Work**:
Agents execute concrete tasks on the target scope. Meta-agent observes patterns, codifies methodology, creates specialized agents as needed.

**Example Tasks**:
- Task 1: [Specific, actionable task]
- Task 2: [Specific, actionable task]
- Task 3: [Specific, actionable task]

---

## Experiment Overview

This experiment applies the bootstrapped software engineering methodology to develop systematic [domain] practices. The goal is to improve [target system] across [N dimensions]: [dimension 1], [dimension 2], etc.

### Architectural Separation

**Critical Principle**: This experiment maintains strict separation between:

1. **Meta-Agent Work** (Methodology Development):
   - Observes agent execution patterns
   - Codifies observations into methodology artifacts
   - Creates specialized agents based on observed needs
   - Produces: `METHODOLOGY.md`, `PATTERNS.md`, specialized agent prompts

2. **Agent Work** (Concrete Task Execution):
   - Executes specific, measurable tasks on real codebase
   - Produces: Code changes, documentation, tests, refactorings
   - Does NOT create methodology documents (unless specialized doc-agent)
   - Focus: Deliver concrete value on instance objective

**Anti-Pattern to Avoid**: Agents should NOT be tasked with "develop [domain] methodology" - this is the meta-agent's job through observation.

### Three-Methodology Framework

This experiment integrates three complementary methodologies:

1. **Empirical Methodology Development (OCA Framework)**
   - **Observe**: Meta-agent collects data about agent work patterns
   - **Codify**: Meta-agent extracts design principles from observations
   - **Automate**: Meta-agent creates specialized agents for recurring patterns

2. **Bootstrapped Software Engineering**
   - **Three-Tuple**: (Output, Agent Set, Meta-Agent) evolves iteratively
   - **Convergence**: Iterations continue until M_n = M_{n-1}, A_n = A_{n-1}, V(s_n) ≥ 0.80
   - **Reusability**: Final three-tuple transferable to other [domain] tasks

3. **Value Space Optimization**
   - **Value Function**: V: S → ℝ maps system state to quality score
   - **Agent as Gradient**: Agents optimize ∇V(s) to improve system quality
   - **Meta-Agent as Hessian**: Meta-agent optimizes ∇²V(s) for agent effectiveness

---

## Value Function

### Components

```yaml
V(s) = w₁·V_[component1] + w₂·V_[component2] + w₃·V_[component3] + w₄·V_[component4]

Component Definitions:
  V_[component1]:
    description: "[What this measures]"
    weight: w₁ = [0.0-1.0]
    measurement: "[How to calculate]"

  V_[component2]:
    description: "[What this measures]"
    weight: w₂ = [0.0-1.0]
    measurement: "[How to calculate]"

  V_[component3]:
    description: "[What this measures]"
    weight: w₃ = [0.0-1.0]
    measurement: "[How to calculate]"

  V_[component4]:
    description: "[What this measures]"
    weight: w₄ = [0.0-1.0]
    measurement: "[How to calculate]"

Target: V(s) ≥ 0.80
```

### Baseline Metrics

```yaml
V(s₀): [calculate] / 0.80 (target)

Components:
  V_[component1]: [score] ([description])
  V_[component2]: [score] ([description])
  V_[component3]: [score] ([description])
  V_[component4]: [score] ([description])

Gap to Target: [0.80 - V(s₀)]

Priority Issue: [Identify weakest component]
```

---

## Initial System Configuration

### Meta-Agent M₀

**Capabilities** (5 standard):
- `observe.md` - Pattern observation from agent work
- `plan.md` - Improvement prioritization and agent selection
- `execute.md` - Agent coordination
- `reflect.md` - Quality evaluation and value calculation
- `evolve.md` - Agent specialization triggers

**Location**: `experiments/bootstrap-XXX-[domain]/meta-agents/`

### Agent Set A₀

**Discovery Method**: [Union search across experiments / Manual selection / Baseline]

**Generic Agents**:
- `coder.md` - General-purpose code implementation
- `data-analyst.md` - Data analysis and metrics
- `doc-writer.md` - Documentation creation
- `doc-generator.md` - Automated documentation generation

**Specialized Agents** (from other experiments):
- [List any pre-existing specialized agents available]

**Location**: `experiments/bootstrap-XXX-[domain]/agents/`

---

## Iteration Progress

### Iteration 0: Baseline Establishment

**Status**: [Planned / In Progress / Completed]
**Date**: YYYY-MM-DD

**Objectives**:
- [ ] Discover/adapt M₀ capabilities (5 files)
- [ ] Discover/identify A₀ agents
- [ ] Collect baseline data for value function
- [ ] Calculate V(s₀)
- [ ] Identify primary problems/gaps
- [ ] Define concrete tasks for Iteration 1 (instance objective)

**Expected Deliverables**:
- `meta-agents/` - 5 capability files
- `agents/` - Initial agent prompt files
- `data/s0-[metrics].yaml` - Baseline measurements
- `iteration-0.md` - Baseline documentation

### Iteration 1: [Planned Focus]

**Status**: Planned
**Expected Focus**: [Address highest-priority gap]

**Expected Work** (Instance Tasks):
- [ ] [Concrete task 1 on target scope]
- [ ] [Concrete task 2 on target scope]
- [ ] [Concrete task 3 on target scope]
- [ ] Calculate V(s₁)

**Expected Meta-Agent Work**:
- Observe patterns in how agents execute tasks
- Begin codifying initial patterns
- Evaluate if specialized agent is needed

**Expected ΔV**: [projection] (if [component] improved from [X] to [Y])

---

## Directory Structure

```
experiments/bootstrap-XXX-[domain]/
├── README.md                    (this file)
├── ITERATION-PROMPTS.md         (execution templates for each iteration)
├── iteration-0.md               (baseline establishment documentation)
├── iteration-1.md               (iteration 1 documentation)
├── meta-agents/                 (M_n capability files)
│   ├── observe.md
│   ├── plan.md
│   ├── execute.md
│   ├── reflect.md
│   └── evolve.md
├── agents/                      (A_n agent prompt files)
│   ├── coder.md
│   ├── data-analyst.md
│   ├── doc-writer.md
│   └── doc-generator.md
└── data/                        (iteration data artifacts)
    ├── s0-[metrics].yaml
    └── [iteration-specific data files]
```

---

## Key Principles

1. **Two-Layer Architecture**: Meta-agent develops methodology by observing agents execute concrete tasks
2. **Discovery-Driven Evolution**: Initial state discovered through union search or empirical observation
3. **Honest Value Assessment**: V(s) calculated from actual observations, not target values
4. **Modular Meta-Agent**: Separate capability files, not versioned monoliths
5. **Agent Prompt Files**: Every agent has explicit prompt file, read before invocation
6. **No Token Limits**: Complete all analysis steps thoroughly without abbreviation
7. **Needs-Driven Specialization**: Create specialized agents only when generic agents insufficient (ΔV ≥ 0.05)
8. **Concrete Instance Work**: Agents work on real code/docs/tests, not methodology documents

---

## Convergence Criteria

Experiment concludes when ALL criteria met:

1. **Meta-Agent Stable**: M_n = M_{n-1} (no new capabilities)
2. **Agent Set Stable**: A_n = A_{n-1} (no new agents created)
3. **Value Threshold**: V(s_n) ≥ 0.80 (target quality achieved)
4. **Objectives Complete**: All [domain] methodology tasks finished
5. **Diminishing Returns**: ΔV < 0.05 (minimal improvement per iteration)

**Current Status**: [NOT CONVERGED / CONVERGED]

---

## Specialization Threshold

New specialized agents created ONLY when:

```yaml
specialization_criteria:
  domain_complexity: HIGH (requires deep domain expertise)
  expected_ΔV: ≥ 0.05 (significant value contribution)
  reusability: HIGH (applicable to multiple future tasks)
  generic_insufficient: TRUE (generic agents cannot achieve quality)

decision: CREATE_SPECIALIZED_AGENT([name])
```

If ΔV < 0.05, use existing generic agents even for complex work.

---

## Iteration Execution Pattern

Each iteration follows this pattern:

### 1. OBSERVE Phase (Meta-Agent)
- Read previous iteration results
- Analyze current state V(s_n)
- Identify gaps and patterns
- Review agent work from previous iteration
- Extract patterns worth codifying

### 2. PLAN Phase (Meta-Agent)
- Prioritize concrete tasks for agents (instance objective)
- Select agents for this iteration
- Evaluate specialization need
- Project expected ΔV
- Define success criteria

### 3. EXECUTE Phase (Agents)
- Agents work on concrete tasks
- Produce actual deliverables (code, docs, tests)
- Do NOT produce methodology documents
- Focus on instance objective

### 4. REFLECT Phase (Meta-Agent)
- Calculate V(s_{n+1})
- Evaluate agent effectiveness
- Codify observed patterns into methodology artifacts
- Check convergence criteria

### 5. EVOLVE Phase (Meta-Agent)
- Decide if M_{n+1} needs new capabilities
- Decide if A_{n+1} needs specialized agents
- Update three-tuple for next iteration

---

## Success Metrics

**Instance Objective Success**:
- [Metric 1]: [Target value]
- [Metric 2]: [Target value]
- [Metric 3]: [Target value]

**Meta-Objective Success**:
- Methodology artifacts complete and documented
- Patterns codified and reusable
- Specialized agents (if any) well-defined
- Three-tuple transferable to new [domain] projects

**Overall Success**:
- V(s_final) ≥ 0.80
- Convergence achieved in 3-5 iterations
- Methodology reusability ≥ 80%
- Specialized agents (if created) demonstrate clear value

---

## Expected Timeline

**Baseline** (Iteration 0): 1-2 days
**Iteration 1**: 1-2 days (first concrete work + observation)
**Iteration 2**: 1-2 days (refinement + pattern codification)
**Iteration 3**: 1-2 days (convergence expected)
**Total**: 5-8 days

---

## Next Steps

1. Execute Iteration 0: Establish baseline
2. Define concrete tasks for Iteration 1 (focus on instance objective)
3. Begin agent work on real target scope
4. Meta-agent observes and codifies patterns
5. Iterate until convergence

---

**Experiment Type**: Bootstrapped Software Engineering
**Framework**: OCA + BSE + Value Space Optimization
**Architecture**: Two-Layer (Meta-Agent observes Agent work)
**Expected Iterations**: 3-5
**Success Metric**: V(s_final) ≥ 0.80 with stable (M, A)
**Reusability Target**: ≥ 80%
