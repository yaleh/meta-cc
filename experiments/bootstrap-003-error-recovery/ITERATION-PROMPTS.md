# Iteration Execution Prompts

This document provides prompt templates for executing each iteration of the bootstrap-003-error-recovery experiment.

---

## Iteration 0: Baseline Establishment

```markdown
# Execute Iteration 0: Baseline Establishment

## Context
I'm starting the bootstrap-003-error-recovery experiment. I've reviewed:
- experiments/bootstrap-003-error-recovery/plan.md
- experiments/bootstrap-003-error-recovery/README.md
- The three methodology frameworks (OCA, Bootstrapped SE, Value Space Optimization)

## Current State
- Meta-Agent: M₀ (5 core capabilities: observe, plan, execute, reflect, evolve)
- Agent Set: A₀ (3 generic agents: data-analyst, doc-writer, coder)

## Meta-Agent and Agent Prompt Files

All Meta-Agents and Agents MUST have corresponding prompt files:

**Meta-Agent capability files (M₀):**
- experiments/bootstrap-003-error-recovery/meta-agents/observe.md
- experiments/bootstrap-003-error-recovery/meta-agents/plan.md
- experiments/bootstrap-003-error-recovery/meta-agents/execute.md
- experiments/bootstrap-003-error-recovery/meta-agents/reflect.md
- experiments/bootstrap-003-error-recovery/meta-agents/evolve.md

**Agent prompt files:**
- experiments/bootstrap-003-error-recovery/agents/data-analyst.md
- experiments/bootstrap-003-error-recovery/agents/doc-writer.md
- experiments/bootstrap-003-error-recovery/agents/coder.md

**CRITICAL EXECUTION PROTOCOL**:
- Before using any Meta-Agent capability, ALWAYS read its capability file first
- Before invoking ANY agent, ALWAYS read its prompt file first
- Each Meta-Agent capability is documented in a separate file
- This ensures correct execution context and captures all details from the files
- Never assume capabilities - always read from the source files

## Iteration 0 Objectives

Execute baseline establishment:

0. **Setup** (Before starting iteration):
   - **CREATE META-AGENT CAPABILITY FILES**: Write separate files for each M₀ capability:
     - **meta-agents/observe.md**: Data collection and pattern recognition for errors
       - How to query error history
       - What error patterns to look for
       - Data sources and collection strategies
     - **meta-agents/plan.md**: Strategy formulation and agent selection for error handling
       - How to prioritize error types
       - When to use generic vs specialized agents
       - Decision-making process for error triage
     - **meta-agents/execute.md**: Agent coordination for error analysis tasks
       - How to coordinate agents for error work
       - Handoff protocols between agents
       - Task execution patterns
     - **meta-agents/reflect.md**: Evaluation and learning from error handling work
       - How to calculate V(s) for error handling
       - Gap identification in error coverage
       - Convergence criteria evaluation
     - **meta-agents/evolve.md**: System adaptation for error handling needs
       - When to create specialized error agents
       - How to identify new capability needs
       - Evolution triggers and processes
   - **CREATE INITIAL AGENT PROMPT FILES**: Write agents/{data-analyst,doc-writer,coder}.md
     - Define each agent's role, capabilities, constraints
     - Specify input/output formats for error analysis
     - Include task-specific instructions for error baseline iteration

1. **Error Data Collection** (M₀.observe):
   - **READ** experiments/bootstrap-003-error-recovery/meta-agents/observe.md (for observation strategies)
   - Query error history: `meta-cc query-tools --status error --scope project`
   - Analyze tool-specific error rates (Bash, Read, Edit, Write)
   - Query error patterns: `meta-cc query-tool-sequences --scope project`
   - Collect error messages and stack traces
   - Identify high-frequency error scenarios

2. **Error Baseline Analysis** (M₀.plan + generic-data-analyst):
   - **READ** experiments/bootstrap-003-error-recovery/meta-agents/plan.md (for planning strategy)
   - **READ** experiments/bootstrap-003-error-recovery/agents/data-analyst.md
   - Invoke data-analyst agent to:
     - Calculate error statistics:
       - Total errors: 1,137
       - Error rate: 6.06%
       - Error distribution by tool type
     - Analyze error patterns and frequencies
     - Identify error categories (initial hypothesis)
     - Calculate value function components:
       - V_detection: Current error detection coverage (0-1)
       - V_diagnosis: Root cause identification capability (0-1)
       - V_recovery: Availability of recovery procedures (0-1)
       - V_prevention: Proactive error prevention measures (0-1)
     - Calculate V(s₀) = 0.4·V_detection + 0.3·V_diagnosis + 0.2·V_recovery + 0.1·V_prevention

3. **Problem Identification** (M₀.reflect):
   - **READ** experiments/bootstrap-003-error-recovery/meta-agents/reflect.md (for reflection process)
   - What are the most common error types?
   - Which errors cause the most disruption?
   - What error patterns exist in the data?
   - What gaps exist in current error handling?
   - What should be the focus of improvement?

4. **Documentation** (M₀.execute + generic-doc-writer):
   - **READ** experiments/bootstrap-003-error-recovery/meta-agents/execute.md (for execution coordination)
   - **READ** experiments/bootstrap-003-error-recovery/agents/doc-writer.md
   - Invoke doc-writer agent to:
     - Create experiments/bootstrap-003-error-recovery/iteration-0.md with:
       - M₀ state (unchanged)
       - A₀ state (unchanged)
       - Error data collection results
       - Error distribution analysis
       - Calculated V(s₀) with breakdown
       - Problem statement
       - Reflection on what's needed next
     - Save raw data to data/ directory:
       - data/s0-metrics.yaml (calculated metrics)
       - data/error-history.jsonl (error records)
       - data/error-distribution.yaml (by tool type)
       - data/error-patterns.txt (identified patterns)

5. **Reflection** (M₀.reflect):
   - **READ** experiments/bootstrap-003-error-recovery/meta-agents/reflect.md (for reflection process)
   - **READ** experiments/bootstrap-003-error-recovery/meta-agents/evolve.md (for evolution assessment)
   - Is error data collection complete?
   - Are M₀ capabilities sufficient for error baseline establishment?
   - What error categories emerged from the data?
   - What should be the focus of Iteration 1?

## Constraints
- Do NOT pre-decide what error categories exist
- Do NOT assume error recovery strategies
- Let the error data guide the taxonomy
- Be honest about what the data shows
- Calculate V(s₀) based on actual error handling state, not target values

## Output Format
Create iteration-0.md following this structure:
- Iteration metadata (number, date, duration)
- M₀ state documentation
- A₀ state documentation
- Error data collection summary
- Error distribution analysis
- Value calculation (V(s₀))
- Problem identification
- Reflection and next steps consideration
```

---

## Iteration 1+: Subsequent Iterations (General Template)

```markdown
# Execute Iteration N: [To be determined by Meta-Agent]

## Context from Previous Iteration

Review the previous iteration file: experiments/bootstrap-003-error-recovery/iteration-[N-1].md

Extract:
- Current Meta-Agent state: M_{N-1}
- Current Agent Set: A_{N-1}
- Current state metrics: V(s_{N-1})
- Error problems identified
- Reflection notes on what's needed next

## Meta-Agent Decision Process

**BEFORE STARTING**: Read the Meta-Agent capability files to understand M_{N-1} role:
- **READ** experiments/bootstrap-003-error-recovery/meta-agents/observe.md
- **READ** experiments/bootstrap-003-error-recovery/meta-agents/plan.md
- **READ** experiments/bootstrap-003-error-recovery/meta-agents/execute.md
- **READ** experiments/bootstrap-003-error-recovery/meta-agents/reflect.md
- **READ** experiments/bootstrap-003-error-recovery/meta-agents/evolve.md
- Understand current capabilities, decision processes, and error handling strategies
- Load coordination patterns and agent selection policy

As M_{N-1}, follow the five-capability process:

### 1. OBSERVE (M.observe)
- **READ** meta-agents/observe.md for observation strategies
- Review previous iteration outputs
- Examine error data collected so far
- Identify gaps in error understanding
- Query additional error data if needed (meta-cc queries, error logs)

### 2. PLAN (M.plan)
- **READ** meta-agents/plan.md for planning and decision-making process
- Based on observations, what is the primary goal for this iteration?
- What capabilities are needed to achieve this goal?
- Are current agents (A_{N-1}) sufficient for error analysis needs?
- If not, what kind of specialized agent is needed?

### 3. EXECUTE (M.execute)
- **READ** meta-agents/execute.md for execution coordination strategies
- Decision point: Should I create a new specialized agent?

**IF current agents are insufficient:**
- **EVOLVE** (M.evolve): Create new specialized agent
  - **READ** meta-agents/evolve.md for evolution process and criteria
  - Define agent name and specialization domain
  - Document capabilities the new agent provides
  - Explain why generic agents are insufficient for this error analysis task
  - **CREATE AGENT PROMPT FILE**: Write experiments/bootstrap-003-error-recovery/agents/{agent-name}.md
    - Include: agent role, capabilities, input/output format, constraints
    - Include: specific instructions for error analysis in this iteration
    - Examples: error-classifier, root-cause-analyzer, recovery-advisor, error-pattern-learner
  - Add to agent set: A_N = A_{N-1} ∪ {new_agent}
  - **UPDATE M**: Add new meta-agent capability if needed
    - Did this iteration reveal need for new coordination pattern?
    - Example: "prioritize_critical_errors" if severity-based triage needed
    - If M_N ≠ M_{N-1}:
      - **CREATE NEW CAPABILITY FILE**: Write meta-agents/{new-capability}.md
      - Document the new capability and its rationale
      - Explain when and how to use this capability
      - Include decision-making patterns and strategies
      - Update other capability files if they reference this new capability

- **READ agent prompt file** before invocation
- Invoke the new specialized agent (or existing agents) to execute error analysis work
- Produce iteration outputs

**ELSE use existing agents:**
- **READ agent prompt file** from experiments/bootstrap-003-error-recovery/agents/{agent-name}.md
- Invoke appropriate agents from A_{N-1}
- Execute planned error analysis work
- Produce iteration outputs

**CRITICAL EXECUTION PROTOCOL**:
1. ALWAYS read ALL Meta-Agent capability files before starting iteration work
2. ALWAYS read the specific capability file before using that capability
3. ALWAYS read agent prompt file before each agent invocation
4. Do NOT cache instructions across iterations - always read from files
5. Capability files may be updated between iterations - get latest details from files
6. Never assume capabilities - always verify from source files

### 4. REFLECT (M.reflect)
- **READ** meta-agents/reflect.md for reflection and evaluation processes
- Evaluate output quality
- Calculate new value: V(s_N)
- Calculate value change: ΔV = V(s_N) - V(s_{N-1})
- Are error handling objectives met?
- Are there still gaps in error coverage?
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
    error_taxonomy_complete: [Yes/No]
    diagnostic_tools_implemented: [Yes/No]
    recovery_procedures_documented: [Yes/No]
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

Create experiments/bootstrap-003-error-recovery/iteration-N.md with:

### 1. Iteration Metadata
```yaml
iteration: N
date: YYYY-MM-DD
duration: ~X hours
status: [completed/converged]
focus: [error_classification / diagnosis / recovery / prevention]
```

### 2. Meta-Agent Evolution (if applicable)
```yaml
M_{N-1} → M_N:
  version: N.0
  new_capabilities: [list any new capabilities added]
  evolution_reason: "Why was this capability needed for error handling?"
  evolution_trigger: "What error problem triggered this?"

  OR

M_N = M_{N-1}: "No evolution (unchanged)"
```

### 3. Agent Set Evolution (if applicable)
```yaml
A_{N-1} → A_N:
  new_agents:
    - name: agent_name
      specialization: error_analysis_domain
      capabilities: [list]
      creation_reason: "Why was generic agent insufficient for this error task?"

  OR

A_N = A_{N-1}: "No new agents (using existing agents)"

agents_invoked_this_iteration:
  - agent_name: error_task_performed
```

### 4. Work Executed
- What error analysis was done this iteration?
- What error handling outputs were produced?
- Summary of error-related deliverables

### 5. State Transition
```yaml
s_{N-1} → s_N:
  changes:
    - description of error handling improvements

  metrics:
    V_detection: [value]     # Error detection coverage
    V_diagnosis: [value]     # Root cause accuracy
    V_recovery: [value]      # Recovery effectiveness
    V_prevention: [value]    # Prevention quality

  value_function:
    V(s_N): [calculated]
    V(s_{N-1}): [previous]
    ΔV: [change]
    percentage: +X.X%
```

### 6. Reflection
- What was learned about errors this iteration?
- What error handling strategies worked well?
- What challenges were encountered in error analysis?
- What error handling improvements are needed next?

### 7. Convergence Check
[Use the convergence criteria structure above]

### 8. Data Artifacts
- Save any error data generated to data/ directory
- Reference data files in iteration document
- Examples:
  - data/iteration-N-error-metrics.yaml
  - data/iteration-N-error-taxonomy.yaml
  - data/iteration-N-recovery-procedures.md
  - data/iteration-N-agent-definitions.yaml

## Key Principles

1. **Be Honest**: Calculate V(s_N) based on actual error handling state, not desired values
2. **Let System Evolve**: Don't force predetermined agent names or error categories
3. **Justify Specialization**: Only create new agents when generic agents can't handle error analysis
4. **Document Evolution**: Clearly explain WHY M or A evolved for error handling
5. **Check Convergence**: Rigorously evaluate convergence criteria each iteration
6. **Stop When Done**: If converged, don't force more iterations
7. **No Token Limits**: There are NO token budget constraints in this experiment
   - Do NOT skip steps due to perceived token limits
   - Do NOT abbreviate error data collection or analysis
   - Do NOT summarize when full details are needed
   - Complete ALL steps thoroughly regardless of length
8. **Meta-Agent and Agent Prompt Files Required**: Every Meta-Agent and agent must have a prompt file
   - Meta-Agent files: experiments/bootstrap-003-error-recovery/meta-agents/meta-agent-m{N}.md
   - Agent files: experiments/bootstrap-003-error-recovery/agents/{agent-name}.md
   - Read: ALWAYS read Meta-Agent file before embodying M role
   - Read: ALWAYS read agent prompt file before agent invocation
   - Update: Create new Meta-Agent file when M evolves (M_N ≠ M_{N-1})
   - Update: Modify agent prompt files as agents evolve or requirements change

## Common Iteration Patterns

Based on OCA framework for error handling, iterations may follow:

- **Observe Phase** (Iterations 0-1): Error data collection, pattern discovery
- **Codify Phase** (Iteration 2-3): Error taxonomy, recovery procedures
- **Automate Phase** (Iteration 3-4): Diagnostic tools, prevention mechanisms

But let the actual error handling needs drive the sequence, not this expected pattern.
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
- All error handling objectives completed

## Objectives

Create experiments/bootstrap-003-error-recovery/results.md analyzing:

### 1. Three-Tuple Output Analysis

**Output O**:
- List all error handling deliverables produced
  - Error taxonomy
  - Diagnostic tools
  - Recovery procedures
  - Prevention mechanisms
- Calculate total lines of code/documentation
- Assess quality and completeness
- Validate against error handling objectives

**Agent Set A_N**:
- List all agents in converged set (specialized + generic)
- Calculate specialization ratio: specialized_count / total_count
- Analyze utilization: How much was each agent used?
- Assess reusability: Which agents are transferable to other projects?

**Meta-Agent M_N**:
- List all capabilities (core + evolved)
- Document learned policy (error triage strategy)
- Analyze evolution timeline: M₀ → M₁ → ... → M_N
- Assess transferability: Which capabilities are domain-independent?

### 2. Convergence Validation

- Formally verify convergence criteria
- Analyze convergence speed (iterations needed)
- Plot value trajectory: V(s₀) → V(s₁) → ... → V(s_N)
- Identify S-curve pattern and diminishing returns

### 3. Value Space Analysis

- Plot V(s) over iterations
- Analyze component contributions (detection, diagnosis, recovery, prevention)
- Calculate total value improvement: ΔV_total = V(s_N) - V(s₀)
- Identify which iterations provided largest error handling gains

### 4. Error Analysis

- Error rate improvement: 6.06% → [final rate]
- Error detection coverage achieved
- Diagnosis accuracy achieved
- Recovery success rate achieved
- Prevention effectiveness achieved

### 5. Reusability Validation

Simulate transfer tests:

**Transfer Test 1** (Similar domain):
- Task: Apply (A_N, M_N) to another Go project's errors
- Estimate: How many iterations needed? (should be fewer than N)
- Estimate: Which agents can be reused directly?
- Estimate: Speedup factor

**Transfer Test 2** (Different domain):
- Task: Apply (A_N, M_N) to web service error handling
- Analyze: Which agents transfer? Which need adaptation?
- Analyze: Which M_N capabilities transfer?
- Estimate: Effort savings vs. from-scratch

### 6. Comparison with Actual History

Compare experiment results with actual meta-cc error handling:
- Baseline error rate: 6.06%
- Error handling improvements over project lifetime
- Process: Iteration pattern vs. actual development
- Validate: Does experiment accurately simulate reality?

### 7. Methodology Validation

Validate alignment with three frameworks:

**Empirical Methodology Development (OCA)**:
- ✓ Observe phase: Error data collection and analysis
- ✓ Codify phase: Error taxonomy and recovery procedures
- ✓ Automate phase: Diagnostic tool creation

**Bootstrapped Software Engineering**:
- ✓ Three-tuple iteration: (M_i, A_i) = M_{i-1}(T, A_{i-1})
- ✓ Convergence: Formal criteria met
- ✓ Reusability: Three-tuple transferable

**Value Space Optimization**:
- ✓ Value function: V: S → ℝ defined and tracked
- ✓ Agent as gradient: A(s) ≈ ∇V(s) demonstrated
- ✓ Meta-Agent as Hessian: M(s, A) ≈ ∇²V(s) demonstrated

### 8. Key Learnings

Synthesize insights:
- What error handling strategies worked well?
- What was surprising about error patterns?
- What would you do differently?
- What are the implications for reliability engineering?

### 9. Scientific Contribution

Articulate what this experiment demonstrates:
- Does it validate the bootstrapping hypothesis for error handling?
- Does it demonstrate value space optimization for reliability?
- Does it show specialized error agent emergence?
- What is novel or significant?

### 10. Future Work

Suggest extensions:
- Additional error handling validation experiments
- Multi-project error analysis
- Automated error recovery systems
- Predictive error prevention models

## Output Format

Structure results.md with:
- Executive Summary
- Three-Tuple Output Analysis (detailed)
- Convergence Validation
- Value Space Trajectory
- Error Analysis Results
- Reusability Validation
- Comparison with Actual History
- Methodology Validation
- Key Learnings
- Scientific Contribution
- Future Work
- Conclusion

Include visualizations where helpful:
- Value trajectory plot (ASCII art or description)
- Error rate improvement over iterations
- Agent evolution timeline
- Convergence metrics table
```

---

## Quick Reference: Iteration Checklist

For each iteration N ≥ 1, ensure you:

- [ ] Review previous iteration (iteration-[N-1].md)
- [ ] Extract current state (M_{N-1}, A_{N-1}, V(s_{N-1}))
- [ ] **READ ALL META-AGENT CAPABILITY FILES**: Read all files in meta-agents/ directory
  - [ ] Read meta-agents/observe.md
  - [ ] Read meta-agents/plan.md
  - [ ] Read meta-agents/execute.md
  - [ ] Read meta-agents/reflect.md
  - [ ] Read meta-agents/evolve.md
- [ ] OBSERVE: Identify error handling needs and gaps
  - [ ] Read meta-agents/observe.md before observing
- [ ] PLAN: Define iteration goal for error improvements
  - [ ] Read meta-agents/plan.md before planning
- [ ] DECIDE: Create new agent? Add M capability?
- [ ] **IF NEW AGENT**: Create agent prompt file in agents/{agent-name}.md
- [ ] **IF M EVOLVES**: Create new capability file in meta-agents/{new-capability}.md
- [ ] **BEFORE AGENT EXECUTION**: Read agent prompt file(s) for agents to be invoked
- [ ] EXECUTE: Invoke agents, produce error handling outputs
  - [ ] Read meta-agents/execute.md before executing
- [ ] REFLECT: Evaluate quality, calculate V(s_N)
  - [ ] Read meta-agents/reflect.md before reflecting
- [ ] CHECK CONVERGENCE: Apply formal criteria
- [ ] DOCUMENT: Create iteration-N.md
- [ ] SAVE DATA: Store error metrics and artifacts in data/
- [ ] **NO TOKEN LIMITS**: Verify all steps completed fully without abbreviation

If CONVERGED:
- [ ] Create results.md
- [ ] Perform reusability analysis
- [ ] Compare with actual error history
- [ ] Document three-tuple (O, A_N, M_N)

---

## Notes on Execution Style

**Be the Meta-Agent**: When executing iterations, embody M's perspective:
- Think through the observe-plan-execute-reflect-evolve cycle for error handling
- Make explicit decisions about error agent creation
- Justify why specialization is needed for error analysis
- Track your own capability evolution

**Be Rigorous**:
- Calculate V(s) honestly based on actual error handling state
- Don't force convergence prematurely
- Don't skip iterations to reach a predetermined end
- Let the error data and needs drive the process

**Be Thorough**:
- Document decisions and reasoning
- Save intermediate error data
- Show your work (calculations, error analysis)
- Make evolution path traceable
- **NO TOKEN LIMITS**: Complete all steps fully, never abbreviate due to length concerns

**Be Authentic**:
- This is a real experiment, not a simulation
- Discover error patterns, don't assume them
- Create error agents based on need, not predetermined plan
- Stop when truly converged, not at a target iteration count

**Meta-Agent and Agent Execution Protocol**:
- **Meta-Agent capability files**: experiments/bootstrap-003-error-recovery/meta-agents/
  - **Structure**: Each capability in separate file (observe.md, plan.md, execute.md, reflect.md, evolve.md)
  - **ALWAYS** read ALL capability files before starting iteration work
  - **ALWAYS** read specific capability file before using that capability
  - Create new capability file when M evolves (M_N ≠ M_{N-1})
  - Each file captures complete capability specification, strategies, and decision processes
  - Never assume Meta-Agent behavior - always read from files for full details
- **Agent files**: experiments/bootstrap-003-error-recovery/agents/
  - **ALWAYS** read agent prompt file before invocation
  - Create prompt files for new agents immediately upon definition
  - Update prompt files as agents evolve or requirements change
  - Never assume agent instructions - always read from file for full details
- **Reading ensures**:
  - Complete context and all details are captured
  - No assumptions about capabilities or processes
  - Latest updates and refinements are incorporated
  - Explicit rather than implicit execution
  - Modular understanding of each capability

---

**Document Version**: 1.0
**Created**: 2025-10-14
**Purpose**: Guide authentic execution of bootstrap-003-error-recovery experiment

**Alignment**: Based on bootstrap-001-doc-methodology ITERATION-PROMPTS.md v1.2
