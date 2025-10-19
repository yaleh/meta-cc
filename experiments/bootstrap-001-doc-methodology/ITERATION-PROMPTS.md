# Iteration Execution Prompts

This document provides prompt templates for executing each iteration of the bootstrap-001-doc-methodology experiment.

---

## Iteration 0: Baseline Establishment

```markdown
# Execute Iteration 0: Baseline Establishment

## Context
I'm starting the bootstrap-001-doc-methodology experiment. I've reviewed:
- experiments/bootstrap-001-doc-methodology/plan.md
- experiments/bootstrap-001-doc-methodology/README.md
- The three methodology frameworks (OCA, Bootstrapped SE, Value Space Optimization)

## Current State
- Meta-Agent: M₀ (5 core capabilities: observe, plan, execute, reflect, evolve)
- Agent Set: A₀ (3 generic agents: data-analyst, doc-writer, coder)

## Meta-Agent and Agent Prompt Files

All Meta-Agents and Agents MUST have corresponding prompt files:

**Meta-Agent prompt file:**
- experiments/bootstrap-001-doc-methodology/meta-agents/meta-agent-m0.md

**Agent prompt files:**
- experiments/bootstrap-001-doc-methodology/agents/data-analyst.md
- experiments/bootstrap-001-doc-methodology/agents/doc-writer.md
- experiments/bootstrap-001-doc-methodology/agents/coder.md

**CRITICAL EXECUTION PROTOCOL**:
- Before embodying the Meta-Agent role, ALWAYS read the Meta-Agent prompt file first
- Before invoking ANY agent, ALWAYS read its prompt file first
- This ensures correct execution context and captures all details from the files
- Never assume capabilities - always read from the source files

## Iteration 0 Objectives

Execute baseline establishment:

0. **Setup** (Before starting iteration):
   - **CREATE META-AGENT PROMPT FILE**: Write experiments/bootstrap-001-doc-methodology/meta-agents/meta-agent-m0.md
     - Document M₀'s 5 core capabilities (observe, plan, execute, reflect, evolve)
     - Define how M₀ coordinates agents
     - Specify decision-making process for agent selection
     - Include value function optimization strategy
     - Define convergence criteria evaluation process
   - **CREATE INITIAL AGENT PROMPT FILES**: Write agents/{data-analyst,doc-writer,coder}.md
     - Define each agent's role, capabilities, constraints
     - Specify input/output formats
     - Include task-specific instructions for baseline iteration

1. **Data Collection** (M₀.observe):
   - **READ** experiments/bootstrap-001-doc-methodology/meta-agents/meta-agent-m0.md (embody M₀ role)
   - Use git log to collect commits from 2025-10-10 to 2025-10-14
   - Use meta-cc CLI to query file access patterns (query-files --scope project)
   - Read current documentation structure (docs/ directory)
   - Identify key documentation files and their characteristics

2. **Baseline Analysis** (M₀.plan + generic-data-analyst):
   - **READ** experiments/bootstrap-001-doc-methodology/meta-agents/meta-agent-m0.md (for planning strategy)
   - **READ** experiments/bootstrap-001-doc-methodology/agents/data-analyst.md
   - Invoke data-analyst agent to:
     - Analyze current documentation state
     - Calculate value function components:
       - V_completeness: How complete is documentation? (features covered / total features)
       - V_accessibility: How easy to find info? (estimate based on structure)
       - V_maintainability: How easy to maintain? (estimate based on organization)
       - V_efficiency: Token cost (CLAUDE.md line count / target)
     - Calculate V(s₀) = 0.3·V_completeness + 0.3·V_accessibility + 0.2·V_maintainability + 0.2·V_efficiency

3. **Problem Identification** (M₀.reflect):
   - **READ** experiments/bootstrap-001-doc-methodology/meta-agents/meta-agent-m0.md (for reflection process)
   - What are the main documentation problems?
   - What patterns exist in the data?
   - What should be the focus of improvement?

4. **Documentation** (M₀.execute + generic-doc-writer):
   - **READ** experiments/bootstrap-001-doc-methodology/meta-agents/meta-agent-m0.md (for execution coordination)
   - **READ** experiments/bootstrap-001-doc-methodology/agents/doc-writer.md
   - Invoke doc-writer agent to:
     - Create experiments/bootstrap-001-doc-methodology/iteration-0.md with:
       - M₀ state (unchanged)
       - A₀ state (unchanged)
       - Data collection results (summary + references to data files)
       - Calculated V(s₀) with breakdown
       - Problem statement
       - Reflection on what's needed next
     - Save raw data to data/ directory:
       - data/s0-metrics.yaml (calculated metrics)
       - data/git-history.txt (relevant git log excerpts)
       - Any other collected data

5. **Reflection** (M₀.reflect):
   - **READ** experiments/bootstrap-001-doc-methodology/meta-agents/meta-agent-m0.md (for reflection and evolution assessment)
   - Is data collection complete?
   - Are M₀ capabilities sufficient for baseline establishment?
   - What should be the focus of Iteration 1?

## Constraints
- Do NOT pre-decide what agents to create next
- Do NOT assume the evolution path
- Let the data and problems guide next steps
- Be honest about what the data shows
- Calculate V(s₀) based on actual observations, not target values

## Output Format
Create iteration-0.md following this structure:
- Iteration metadata (number, date, duration)
- M₀ state documentation
- A₀ state documentation
- Data collection summary
- Value calculation (V(s₀))
- Problem identification
- Reflection and next steps consideration
```

---

## Iteration 1+: Subsequent Iterations (General Template)

```markdown
# Execute Iteration N: [To be determined by Meta-Agent]

## Context from Previous Iteration

Review the previous iteration file: experiments/bootstrap-001-doc-methodology/iteration-[N-1].md

Extract:
- Current Meta-Agent state: M_{N-1}
- Current Agent Set: A_{N-1}
- Current state metrics: V(s_{N-1})
- Problems identified
- Reflection notes on what's needed next

## Meta-Agent Decision Process

**BEFORE STARTING**: Read the Meta-Agent prompt file to embody M_{N-1} role:
- **READ** experiments/bootstrap-001-doc-methodology/meta-agents/meta-agent-m{N-1}.md
- Understand current capabilities, decision processes, and strategies
- Load coordination patterns and agent selection policy

As M_{N-1}, follow the five-capability process:

### 1. OBSERVE (M.observe)
- **READ** meta-agent file for observation strategies
- Review previous iteration outputs
- Examine data collected so far
- Identify gaps or new data needs
- Query additional data if needed (git log, meta-cc CLI, file reads)

### 2. PLAN (M.plan)
- **READ** meta-agent file for planning and decision-making process
- Based on observations, what is the primary goal for this iteration?
- What capabilities are needed to achieve this goal?
- Are current agents (A_{N-1}) sufficient?
- If not, what kind of specialized agent is needed?

### 3. EXECUTE (M.execute)
- **READ** meta-agent file for execution coordination strategies
- Decision point: Should I create a new specialized agent?

**IF current agents are insufficient:**
- **EVOLVE** (M.evolve): Create new specialized agent
  - Define agent name and specialization domain
  - Document capabilities the new agent provides
  - Explain why generic agents are insufficient
  - **CREATE AGENT PROMPT FILE**: Write experiments/bootstrap-001-doc-methodology/agents/{agent-name}.md
    - Include: agent role, capabilities, input/output format, constraints
    - Include: specific instructions for this iteration's task
  - Add to agent set: A_N = A_{N-1} ∪ {new_agent}
  - **UPDATE M**: Add new meta-agent capability if needed
    - Did this iteration reveal need for new coordination pattern?
    - Example: "manage_shared_references" if agents need to share data
    - If M_N ≠ M_{N-1}:
      - **CREATE NEW META-AGENT PROMPT FILE**: Write meta-agents/meta-agent-m{N}.md
      - Document new capabilities and their rationale
      - Preserve all existing capabilities
      - Update decision-making and coordination strategies

- **READ agent prompt file** before invocation
- Invoke the new specialized agent (or existing agents) to execute work
- Produce iteration outputs

**ELSE use existing agents:**
- **READ agent prompt file** from experiments/bootstrap-001-doc-methodology/agents/{agent-name}.md
- Invoke appropriate agents from A_{N-1}
- Execute planned work
- Produce iteration outputs

**CRITICAL EXECUTION PROTOCOL**:
1. ALWAYS read Meta-Agent prompt file before embodying the Meta-Agent role
2. ALWAYS read agent prompt file before each agent invocation
3. Do NOT cache instructions across iterations - always read from files
4. Prompt files may be updated between iterations - get latest details from files
5. Never assume capabilities - always verify from source files

### 4. REFLECT (M.reflect)
- **READ** meta-agent file for reflection and evaluation processes
- Evaluate output quality
- Calculate new value: V(s_N)
- Calculate value change: ΔV = V(s_N) - V(s_{N-1})
- Are task objectives met?
- Are there still gaps or problems?
- What should be the focus of next iteration?

### 5. CHECK CONVERGENCE
Evaluate convergence criteria:

```yaml
convergence_check:
  meta_agent_stable:
    question: "Did M gain new capabilities this iteration?"
    M_N == M_{N-1}: [Yes/No]

  agent_set_stable:
    question: "Were new agents created this iteration?"
    A_N == A_{N-1}: [Yes/No]

  value_threshold:
    question: "Is V(s_N) ≥ 0.80 (target)?"
    V(s_N): [calculated value]
    threshold_met: [Yes/No]

  task_objectives:
    methodology_documented: [Yes/No]
    automation_implemented: [Yes/No]
    all_objectives_met: [Yes/No]

  diminishing_returns:
    ΔV_current: [current value change]
    interpretation: "Is improvement marginal?"

convergence_status: [CONVERGED / NOT_CONVERGED]
```

**IF CONVERGED:**
- Stop iteration process
- Proceed to results analysis
- Document three-tuple: (O, A_N, M_N)

**IF NOT CONVERGED:**
- Identify what's needed for next iteration
- Continue to Iteration N+1

## Documentation Requirements

Create experiments/bootstrap-001-doc-methodology/iteration-N.md with:

### 1. Iteration Metadata
```yaml
iteration: N
date: YYYY-MM-DD
duration: ~X hours
status: [completed/converged]
```

### 2. Meta-Agent Evolution (if applicable)
```yaml
M_{N-1} → M_N:
  version: N.0
  new_capabilities: [list any new capabilities added]
  evolution_reason: "Why was this capability needed?"
  evolution_trigger: "What problem triggered this?"

  OR

M_N = M_{N-1}: "No evolution (unchanged)"
```

### 3. Agent Set Evolution (if applicable)
```yaml
A_{N-1} → A_N:
  new_agents:
    - name: agent_name
      specialization: domain
      capabilities: [list]
      creation_reason: "Why was generic agent insufficient?"

  OR

A_N = A_{N-1}: "No new agents (using existing agents)"

agents_invoked_this_iteration:
  - agent_name: task_performed
```

### 4. Work Executed
- What was done this iteration?
- What outputs were produced?
- Summary of deliverables

### 5. State Transition
```yaml
s_{N-1} → s_N:
  changes:
    - description of state changes

  metrics:
    V_completeness: [value]
    V_accessibility: [value]
    V_maintainability: [value]
    V_efficiency: [value]

  value_function:
    V(s_N): [calculated]
    V(s_{N-1}): [previous]
    ΔV: [change]
    percentage: +X.X%
```

### 6. Reflection
- What was learned this iteration?
- What worked well?
- What challenges were encountered?
- What is needed next?

### 7. Convergence Check
[Use the convergence criteria structure above]

### 8. Data Artifacts
- Save any data generated to data/ directory
- Reference data files in iteration document
- Examples:
  - data/iteration-N-metrics.yaml
  - data/iteration-N-outputs.jsonl
  - data/iteration-N-agent-definitions.yaml

## Key Principles

1. **Be Honest**: Calculate V(s_N) based on actual state, not desired values
2. **Let System Evolve**: Don't force predetermined agent names or evolution paths
3. **Justify Specialization**: Only create new agents when generic agents are truly insufficient
4. **Document Evolution**: Clearly explain WHY M or A evolved
5. **Check Convergence**: Rigorously evaluate convergence criteria each iteration
6. **Stop When Done**: If converged, don't force more iterations
7. **No Token Limits**: There are NO token budget constraints in this experiment
   - Do NOT skip steps due to perceived token limits
   - Do NOT abbreviate data collection or analysis
   - Do NOT summarize when full details are needed
   - Complete ALL steps thoroughly regardless of length
8. **Meta-Agent and Agent Prompt Files Required**: Every Meta-Agent and agent must have a prompt file
   - Meta-Agent files: experiments/bootstrap-001-doc-methodology/meta-agents/meta-agent-m{N}.md
   - Agent files: experiments/bootstrap-001-doc-methodology/agents/{agent-name}.md
   - Read: ALWAYS read Meta-Agent file before embodying M role
   - Read: ALWAYS read agent prompt file before agent invocation
   - Update: Create new Meta-Agent file when M evolves (M_N ≠ M_{N-1})
   - Update: Modify agent prompt files as agents evolve or requirements change

## Common Iteration Patterns

Based on OCA framework, iterations may follow:

- **Observe Phase** (Iterations 0-1): Data collection, pattern discovery
- **Codify Phase** (Iteration 2-3): Extract principles, write methodology
- **Automate Phase** (Iteration 3-4): Create validation tools, implement capabilities

But let the actual needs drive the sequence, not this expected pattern.
```

---

## Final Iteration: Results Analysis

```markdown
# Create Final Results Analysis

## Context
Convergence has been achieved at Iteration N.

Previous iteration showed:
- M_N = M_{N-1} (no new meta-agent capabilities)
- A_N = A_{N-1} (no new agents created)
- V(s_N) ≥ 0.80 (target threshold met)
- All task objectives completed

## Objectives

Create experiments/bootstrap-001-doc-methodology/results.md analyzing:

### 1. Three-Tuple Output Analysis

**Output O**:
- List all deliverables produced
- Calculate total lines of code/documentation
- Assess quality and completeness
- Validate against task objectives

**Agent Set A_N**:
- List all agents in converged set (specialized + generic)
- Calculate specialization ratio: specialized_count / total_count
- Analyze utilization: How much was each agent used?
- Assess reusability: Which agents are transferable?

**Meta-Agent M_N**:
- List all capabilities (core + evolved)
- Document learned policy (agent selection strategy)
- Analyze evolution timeline: M₀ → M₁ → ... → M_N
- Assess transferability: Which capabilities are domain-independent?

### 2. Convergence Validation

- Formally verify convergence criteria
- Analyze convergence speed (iterations needed)
- Plot value trajectory: V(s₀) → V(s₁) → ... → V(s_N)
- Identify S-curve pattern and diminishing returns

### 3. Value Space Analysis

- Plot V(s) over iterations
- Analyze component contributions (completeness, accessibility, maintainability, efficiency)
- Calculate total value improvement: ΔV_total = V(s_N) - V(s₀)
- Identify which iterations provided largest gains

### 4. Reusability Validation

Simulate transfer tests:

**Transfer Test 1** (Similar domain):
- Task: Apply (A_N, M_N) to new documentation project
- Estimate: How many iterations needed? (should be fewer than N)
- Estimate: Which agents can be reused directly?
- Estimate: Speedup factor

**Transfer Test 2** (Different domain):
- Task: Apply (A_N, M_N) to different task (e.g., testing methodology)
- Analyze: Which agents transfer? Which need adaptation?
- Analyze: Which M_N capabilities transfer?
- Estimate: Effort savings vs. from-scratch

### 5. Comparison with Actual History

Compare experiment results with actual meta-cc development:
- Timeline: Experiment vs. actual duration
- Outputs: Lines of code, deliverables
- Process: Iteration pattern vs. actual development phases
- Validate: Does experiment accurately simulate reality?

### 6. Methodology Validation

Validate alignment with three frameworks:

**Empirical Methodology Development (OCA)**:
- ✓ Observe phase: Data collection and analysis
- ✓ Codify phase: Methodology extraction
- ✓ Automate phase: Tool creation

**Bootstrapped Software Engineering**:
- ✓ Three-tuple iteration: (M_i, A_i) = M_{i-1}(T, A_{i-1})
- ✓ Convergence: Formal criteria met
- ✓ Reusability: Three-tuple transferable

**Value Space Optimization**:
- ✓ Value function: V: S → ℝ defined and tracked
- ✓ Agent as gradient: A(s) ≈ ∇V(s) demonstrated
- ✓ Meta-Agent as Hessian: M(s, A) ≈ ∇²V(s) demonstrated

### 7. Key Learnings

Synthesize insights:
- What worked well?
- What was surprising?
- What would you do differently?
- What are the implications for future work?

### 8. Scientific Contribution

Articulate what this experiment demonstrates:
- Does it validate the bootstrapping hypothesis?
- Does it demonstrate value space optimization?
- Does it show agent specialization emergence?
- What is novel or significant?

### 9. Future Work

Suggest extensions:
- Additional validation experiments
- Multi-domain meta-agent training
- Automated agent creation
- Convergence prediction models

## Output Format

Structure results.md with:
- Executive Summary
- Three-Tuple Output Analysis (detailed)
- Convergence Validation
- Value Space Trajectory
- Reusability Validation
- Comparison with Actual History
- Methodology Validation
- Key Learnings
- Scientific Contribution
- Future Work
- Conclusion

Include visualizations where helpful:
- Value trajectory plot (ASCII art or description)
- Agent evolution timeline
- Convergence metrics table
```

---

## Quick Reference: Iteration Checklist

For each iteration N ≥ 1, ensure you:

- [ ] Review previous iteration (iteration-[N-1].md)
- [ ] Extract current state (M_{N-1}, A_{N-1}, V(s_{N-1}))
- [ ] **READ META-AGENT FILE**: Read meta-agents/meta-agent-m{N-1}.md before embodying M role
- [ ] OBSERVE: Identify needs and gaps (using M's observe capability)
- [ ] PLAN: Define iteration goal (using M's plan capability)
- [ ] DECIDE: Create new agent? Add M capability?
- [ ] **IF NEW AGENT**: Create agent prompt file in agents/{agent-name}.md
- [ ] **IF M EVOLVES**: Create new meta-agents/meta-agent-m{N}.md with updated capabilities
- [ ] **BEFORE AGENT EXECUTION**: Read agent prompt file(s) for agents to be invoked
- [ ] EXECUTE: Invoke agents, produce outputs (using M's execute capability)
- [ ] REFLECT: Evaluate quality, calculate V(s_N) (using M's reflect capability)
- [ ] CHECK CONVERGENCE: Apply formal criteria
- [ ] DOCUMENT: Create iteration-N.md
- [ ] SAVE DATA: Store metrics and artifacts in data/
- [ ] **NO TOKEN LIMITS**: Verify all steps completed fully without abbreviation

If CONVERGED:
- [ ] Create results.md
- [ ] Perform reusability analysis
- [ ] Compare with actual history
- [ ] Document three-tuple (O, A_N, M_N)

---

## Notes on Execution Style

**Be the Meta-Agent**: When executing iterations, embody M's perspective:
- Think through the observe-plan-execute-reflect-evolve cycle
- Make explicit decisions about agent creation
- Justify why specialization is needed
- Track your own capability evolution

**Be Rigorous**:
- Calculate V(s) honestly based on actual state
- Don't force convergence prematurely
- Don't skip iterations to reach a predetermined end
- Let the data and needs drive the process

**Be Thorough**:
- Document decisions and reasoning
- Save intermediate data
- Show your work (calculations, analysis)
- Make evolution path traceable
- **NO TOKEN LIMITS**: Complete all steps fully, never abbreviate due to length concerns

**Be Authentic**:
- This is a real experiment, not a simulation
- Discover patterns, don't assume them
- Create agents based on need, not predetermined plan
- Stop when truly converged, not at a target iteration count

**Meta-Agent and Agent Execution Protocol**:
- **Meta-Agent files**: experiments/bootstrap-001-doc-methodology/meta-agents/
  - **ALWAYS** read Meta-Agent prompt file before embodying M role
  - Create new Meta-Agent file when M evolves (M_N ≠ M_{N-1})
  - File captures complete capabilities, strategies, and decision processes
  - Never assume Meta-Agent behavior - always read from file for full details
- **Agent files**: experiments/bootstrap-001-doc-methodology/agents/
  - **ALWAYS** read agent prompt file before invocation
  - Create prompt files for new agents immediately upon definition
  - Update prompt files as agents evolve or requirements change
  - Never assume agent instructions - always read from file for full details
- **Reading ensures**:
  - Complete context and all details are captured
  - No assumptions about capabilities or processes
  - Latest updates and refinements are incorporated
  - Explicit rather than implicit execution

---

**Document Version**: 1.2
**Created**: 2025-10-14
**Last Updated**: 2025-10-14
**Purpose**: Guide authentic execution of bootstrap-001-doc-methodology experiment

**Changelog**:
- v1.2 (2025-10-14): Added Meta-Agent prompt file requirements
  - All Meta-Agents must have corresponding .md prompt files
  - Meta-Agent files: meta-agents/meta-agent-m{N}.md
  - Must read Meta-Agent file before embodying M role
  - Must create new Meta-Agent file when M evolves
  - Updated all execution protocols to include Meta-Agent file reading
  - Added explicit READ instructions throughout iteration steps
  - Ensures complete context and all details are captured from files
- v1.1 (2025-10-14): Added agent prompt file requirements and token limit clarifications
  - All agents must have corresponding .md prompt files
  - Must read agent prompt file before each invocation
  - Emphasized NO token limits for thorough execution
- v1.0 (2025-10-14): Initial version
