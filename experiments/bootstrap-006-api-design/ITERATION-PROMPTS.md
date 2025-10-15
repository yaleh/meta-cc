# Iteration Execution Prompts

This document provides prompt templates for executing each iteration of the bootstrap-006-api-design experiment.

---

## Iteration 0: Baseline Establishment

```markdown
# Execute Iteration 0: Baseline Establishment

## Context
I'm starting the bootstrap-006-api-design experiment. I've reviewed:
- experiments/bootstrap-006-api-design/plan.md
- experiments/bootstrap-006-api-design/README.md
- The three methodology frameworks (OCA, Bootstrapped SE, Value Space Optimization)

## Current State (To Be Determined by Search)

**Initial state will be the UNION of all existing meta-agent and agent files from other experiments:**

- **Meta-Agent: M₀** = ∪(all meta-agent capability files found in experiments/bootstrap-*)
  - Capabilities determined by search results (e.g., observe, plan, execute, reflect, evolve, or others)
  - NOT predetermined - discover what capabilities exist across experiments

- **Agent Set: A₀** = ∪(all generic agent files found in experiments/bootstrap-*)
  - Agents determined by search results (e.g., data-analyst, doc-writer, coder, or others)
  - NOT predetermined - discover what agents exist and can be reused
  - Include both generic agents AND specialized agents that might be reusable

**Discovery-Driven Approach**:
- DO NOT assume M₀ has exactly 5 capabilities
- DO NOT assume A₀ has exactly 3 agents
- Let the search results define the initial state
- The union ensures maximum reusability from prior experiments

## Meta-Agent and Agent Prompt Files

All Meta-Agents and Agents will have corresponding prompt files copied from other experiments:

**Meta-Agent capability files** (location: experiments/bootstrap-006-api-design/meta-agents/):
- Files discovered and copied during setup (Step 0)
- May include: observe.md, plan.md, execute.md, reflect.md, evolve.md, or additional capabilities
- Full list determined by search across experiments/bootstrap-*

**Agent prompt files** (location: experiments/bootstrap-006-api-design/agents/):
- Files discovered and copied during setup (Step 0)
- May include: data-analyst.md, doc-writer.md, coder.md, or additional agents
- May include specialized agents that are reusable for API design
- Full list determined by search across experiments/bootstrap-*

**CRITICAL EXECUTION PROTOCOL**:
- Before using ANY capability, ALWAYS read the corresponding capability file first
- Before invoking ANY agent, ALWAYS read its prompt file first
- This ensures correct execution context and captures all details from the files
- Never assume capabilities - always read from the source files
- Never assume which files exist - verify from actual filesystem search

## Iteration 0 Objectives

Execute baseline establishment:

0. **Setup** (Before starting iteration):
   - **SEARCH AND COPY META-AGENT CAPABILITY FILES (UNION APPROACH)**:
     - Search for ALL meta-agent capability files from ALL experiments
     - Copy UNION of all discovered capabilities to create comprehensive M₀
     - Commands:
       ```bash
       # Search for ALL meta-agent capability files across experiments
       echo "=== Searching for all meta-agent capability files ==="
       find experiments/bootstrap-* -type f -path "*/meta-agents/*.md" | sort

       # Create union: copy all unique capability files
       mkdir -p experiments/bootstrap-006-api-design/meta-agents

       # Strategy 1: Copy from most recent experiment with modular architecture
       # (bootstrap-003-error-recovery as baseline)
       for file in experiments/bootstrap-003-error-recovery/meta-agents/*.md; do
         [ -f "$file" ] && cp "$file" experiments/bootstrap-006-api-design/meta-agents/
       done

       # Strategy 2: Add any additional capabilities from other experiments
       # (merge capabilities not in bootstrap-003)
       for exp in experiments/bootstrap-00{1,2,4}*; do
         if [ -d "$exp/meta-agents" ]; then
           for file in "$exp/meta-agents"/*.md; do
             basename_file=$(basename "$file")
             if [ ! -f "experiments/bootstrap-006-api-design/meta-agents/$basename_file" ]; then
               echo "Adding additional capability: $basename_file from $exp"
               cp "$file" experiments/bootstrap-006-api-design/meta-agents/
             fi
           done
         fi
       done

       # List final M₀ capability set (the union)
       echo "=== Final M₀ capabilities ==="
       ls -1 experiments/bootstrap-006-api-design/meta-agents/*.md | xargs -n1 basename
       ```
     - **DOCUMENT M₀ STATE**: After copying, document the discovered M₀:
       - Count and list all capability files
       - Note which experiments contributed which capabilities
       - This becomes the baseline M₀ definition (not predetermined)

     - **REVIEW AND ADAPT**: After copying, review and adapt each file for API design domain:
       - For EACH capability file found:
         - Replace domain-specific terminology (e.g., "error" → "API design")
         - Update data sources for API design context
         - Update specific commands for API-related queries
         - Adapt examples to API design scenarios
       - Common adaptations:
         - **observe**: API data collection, tool schemas, usage patterns
         - **plan**: API improvement prioritization, breaking vs non-breaking changes
         - **execute**: API analysis workflows, consistency checking coordination
         - **reflect**: API quality assessment (usability, consistency, completeness, evolvability)
         - **evolve**: API agent specialization triggers
         - **[other capabilities]**: Adapt based on capability purpose

   - **SEARCH AND COPY AGENT PROMPT FILES (UNION APPROACH)**:
     - Search for ALL agent files from ALL experiments (generic + specialized)
     - Copy UNION of all potentially reusable agents to create comprehensive A₀
     - Commands:
       ```bash
       # Search for ALL agent files across experiments
       echo "=== Searching for all agent files ==="
       find experiments/bootstrap-* -type f -path "*/agents/*.md" | sort

       # Show agent distribution across experiments
       for exp in experiments/bootstrap-00{1,2,3,4}*; do
         if [ -d "$exp/agents" ]; then
           echo "=== $exp ==="
           ls -1 "$exp/agents"/*.md 2>/dev/null | xargs -n1 basename
         fi
       done

       # Create union: copy all unique agent files
       mkdir -p experiments/bootstrap-006-api-design/agents

       # Copy all agents from all experiments (union)
       for exp in experiments/bootstrap-00{1,2,3,4}*; do
         if [ -d "$exp/agents" ]; then
           for file in "$exp/agents"/*.md; do
             basename_file=$(basename "$file")
             if [ ! -f "experiments/bootstrap-006-api-design/agents/$basename_file" ]; then
               echo "Adding agent: $basename_file from $exp"
               cp "$file" experiments/bootstrap-006-api-design/agents/
             elif [ "$exp" = "experiments/bootstrap-003-error-recovery" ]; then
               # Prefer bootstrap-003 version if duplicate
               echo "Using $basename_file from bootstrap-003 (preferred)"
               cp "$file" experiments/bootstrap-006-api-design/agents/
             fi
           done
         fi
       done

       # List final A₀ agent set (the union)
       echo "=== Final A₀ agents ==="
       ls -1 experiments/bootstrap-006-api-design/agents/*.md | xargs -n1 basename
       ```
     - **DOCUMENT A₀ STATE**: After copying, document the discovered A₀:
       - Count and list all agent files
       - Categorize: generic agents vs specialized agents
       - Note which experiments contributed which agents
       - Assess reusability: which specialized agents might be useful for API design
       - This becomes the baseline A₀ definition (not predetermined)

     - **REVIEW AND ADAPT**: After copying, review each agent file for API design context:
       - For EACH agent file found:
         - Verify capabilities match API design needs
         - Update examples to use API-specific terminology
         - Ensure constraints align with API design objectives
         - Decide if specialized agents need adaptation or can be used as-is
       - Generic agents (e.g., data-analyst, doc-writer, coder):
         - Usually reusable with minor terminology updates
       - Specialized agents (e.g., error-classifier, test-generator):
         - Evaluate reusability for API design domain
         - May inspire analogous API-specific agents
         - Keep if potentially useful, remove if irrelevant

1. **API Data Collection** (Use M₀'s observation capability):
   - **FIRST**: List and read available M₀ capability files:
     ```bash
     ls -1 experiments/bootstrap-006-api-design/meta-agents/*.md
     ```
   - **READ** the observation capability file (likely observe.md, but verify from search results)
   - Follow observation strategies defined in the capability file
   - Collect API-specific data:
     ```bash
     # MCP tool definitions
     cat docs/guides/mcp.md | grep -A 20 "^### [0-9]\\+\\."

     # Tool usage frequency
     meta-cc query-tools --scope project --stats-only

     # Tool call sequences
     meta-cc query-tool-sequences --min-occurrences 3 --scope project

     # API error patterns
     meta-cc query-tools --status error --scope project

     # User requests for API features
     meta-cc query-user-messages --pattern "query|tool|MCP|parameter|API" --scope project
     ```
   - Read API implementation files:
     - internal/tools/tools.go
     - internal/capabilities/capabilities.go
     - cmd/mcp.go

2. **Baseline API Analysis** (Use M₀'s planning capability + available agents):
   - **FIRST**: List and identify available agents:
     ```bash
     ls -1 experiments/bootstrap-006-api-design/agents/*.md
     ```
   - **READ** the planning capability file (for planning strategy)
   - **READ** appropriate agent file(s) for analysis work:
     - Look for data-analyst.md or equivalent analytical agent
     - If multiple analytical agents exist, choose most appropriate
   - Invoke chosen agent(s) to:
     - Analyze current API state across 16 MCP tools
     - Calculate value function components:
       - V_usability: Parameter clarity, default values, error messages
       - V_consistency: Naming conventions, parameter patterns, response formats
       - V_completeness: Feature coverage, parameter options, edge cases
       - V_evolvability: Versioning strategy, deprecation policy, backward compatibility
     - Calculate V(s₀) = 0.3·V_usability + 0.3·V_consistency + 0.2·V_completeness + 0.2·V_evolvability

3. **API Problem Identification** (Use M₀'s reflection capability):
   - **READ** the reflection capability file (for reflection process)
   - Apply reflection to identify:
     - Main API design inconsistencies
     - Usability issues in current tool parameters
     - Completeness gaps in API coverage
     - Evolvability concerns for future changes

4. **Documentation** (Use M₀'s execution capability + documentation agent):
   - **READ** the execution capability file (for coordination)
   - **READ** appropriate agent file for documentation:
     - Look for doc-writer.md or equivalent documentation agent
   - Invoke documentation agent to:
     - Create experiments/bootstrap-006-api-design/iteration-0.md with:
       - M₀ state (list ALL discovered capabilities, not predetermined count)
       - A₀ state (list ALL discovered agents, not predetermined count)
       - API data collection results (summary + references to data files)
       - Calculated V(s₀) with breakdown
       - API problem statement
       - Reflection on what's needed next
     - Save raw API data to data/ directory:
       - data/s0-api-metrics.yaml (calculated metrics)
       - data/s0-tool-definitions.jsonl (tool schemas)
       - data/s0-tool-usage.jsonl (frequency data)
       - data/s0-api-errors.jsonl (error patterns)
       - data/s0-agents-inventory.yaml (A₀ agent list and sources)
       - data/s0-capabilities-inventory.yaml (M₀ capability list and sources)

5. **Reflection** (Use M₀'s reflection capability):
   - **READ** the reflection capability file (for reflection and evolution assessment)
   - Reflect on baseline establishment:
     - Is API data collection complete?
     - Are M₀ capabilities (as discovered) sufficient for baseline establishment?
     - Are A₀ agents (as discovered) sufficient for initial work?
     - What should be the focus of Iteration 1?

## Constraints
- Do NOT pre-decide what agents to create next
- Do NOT assume the evolution path
- Let the API data and problems guide next steps
- Be honest about what the API data shows
- Calculate V(s₀) based on actual API observations, not target values

## Output Format
Create iteration-0.md following this structure:
- Iteration metadata (number, date, duration)
- M₀ state documentation (list ALL discovered capabilities with their sources)
  - Count: X capabilities (discovered, not predetermined)
  - List each capability file and its origin experiment
  - Example: "observe.md (from bootstrap-003-error-recovery)"
- A₀ state documentation (list ALL discovered agents with their sources)
  - Count: Y agents (discovered, not predetermined)
  - Categorize: generic agents, specialized agents (potentially reusable)
  - List each agent file and its origin experiment
  - Example: "data-analyst.md (from bootstrap-003-error-recovery)"
- API data collection summary
- Value calculation (V(s₀))
- API problem identification
- Reflection and next steps consideration
```

---

## Iteration 1+: Subsequent Iterations (General Template)

```markdown
# Execute Iteration N: [To be determined by Meta-Agent]

## Context from Previous Iteration

Review the previous iteration file: experiments/bootstrap-006-api-design/iteration-[N-1].md

Extract:
- Current Meta-Agent state: M_{N-1}
- Current Agent Set: A_{N-1}
- Current API state metrics: V(s_{N-1})
- API problems identified
- Reflection notes on what's needed next

## Meta-Agent Decision Process

**BEFORE STARTING**: Discover and read ALL Meta-Agent capability files to understand M_{N-1}:
- **LIST** all capability files:
  ```bash
  ls -1 experiments/bootstrap-006-api-design/meta-agents/*.md
  ```
- **READ** EACH capability file discovered (do not assume specific names)
  - May include: observe.md, plan.md, execute.md, reflect.md, evolve.md
  - May include additional capabilities discovered during setup
  - Read ALL files to understand complete M_{N-1} capabilities

As M_{N-1}, follow M_{N-1}'s capability process (typically observe-plan-execute-reflect-evolve, but verify from actual files):

### 1. OBSERVE (Use observation capability)
- **LIST** and identify observation capability file (typically observe.md or similar)
- **READ** the observation capability file for observation strategies
- Review previous iteration outputs
- Examine API data collected so far
- Identify gaps or new API data needs
- Query additional data if needed:
  ```bash
  # API consistency analysis
  grep -r "query_" internal/tools/ | grep -o 'func.*(' | sort

  # Parameter pattern analysis
  meta-cc query-tools --scope project | jq '.[] | .parameters | keys'

  # API error pattern analysis
  meta-cc query-tools --status error --scope project | \
    jq 'group_by(.error_type) | map({error: .[0].error_type, count: length})'

  # User API feature requests
  meta-cc query-user-messages --pattern "parameter|option|feature" --scope project
  ```

### 2. PLAN (Use planning capability)
- **LIST** and identify planning capability file (typically plan.md or similar)
- **READ** the planning capability file for planning and decision-making process
- Based on observations, what is the primary API design goal for this iteration?
- What capabilities are needed to achieve this goal?
- Are current agents (A_{N-1}) sufficient?
- If not, what kind of specialized API agent is needed?
- What is the prioritization of API issues (by usage frequency × severity)?

### 3. EXECUTE (Use execution capability)
- **LIST** and identify execution capability file (typically execute.md or similar)
- **READ** the execution capability file for execution coordination strategies
- Decision point: Should I create a new specialized agent?

**IF current agents are insufficient:**
- **EVOLVE** (Use evolution capability): Create new specialized agent
  - **LIST** and identify evolution capability file (typically evolve.md or similar)
  - **READ** the evolution capability file for evolution criteria
  - Define agent name and specialization domain
  - Document capabilities the new agent provides
  - Explain why generic agents are insufficient
  - **CREATE AGENT PROMPT FILE**: Write experiments/bootstrap-006-api-design/agents/{agent-name}.md
    - Include: agent role, capabilities, input/output format, constraints
    - Include: specific instructions for this iteration's API task
  - Add to agent set: A_N = A_{N-1} ∪ {new_agent}
  - **UPDATE M**: Add new meta-agent capability if needed
    - Did this iteration reveal need for new coordination pattern?
    - Example: "api_compatibility_validation" for checking backward compatibility
    - If M_N ≠ M_{N-1}:
      - **CREATE NEW CAPABILITY FILE**: Write meta-agents/{capability}.md
      - Document new capability and its rationale
      - Update other capabilities that reference this new capability

- **READ agent prompt file** before invocation
- Invoke the new specialized agent (or existing agents) to execute API work
- Produce iteration outputs

**ELSE use existing agents:**
- **READ agent prompt file** from experiments/bootstrap-006-api-design/agents/{agent-name}.md
- Invoke appropriate agents from A_{N-1}
- Execute planned API work
- Produce iteration outputs

**CRITICAL EXECUTION PROTOCOL**:
1. ALWAYS read ALL Meta-Agent capability files at start of iteration
2. ALWAYS read specific capability file before using that capability
3. ALWAYS read agent prompt file before each agent invocation
4. Do NOT cache instructions across iterations - always read from files
5. Prompt files may be updated between iterations - get latest details from files
6. Never assume capabilities - always verify from source files

### 4. REFLECT (Use reflection capability)
- **LIST** and identify reflection capability file (typically reflect.md or similar)
- **READ** the reflection capability file for reflection and evaluation processes
- Evaluate API output quality
- Calculate new value: V(s_N)
- Calculate value change: ΔV = V(s_N) - V(s_{N-1})
- Are API task objectives met?
- Are there still gaps or API problems?
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
    api_methodology_documented: [Yes/No]
    api_validation_implemented: [Yes/No]
    all_objectives_met: [Yes/No]

  diminishing_returns:
    ΔV_current: [current value change]
    interpretation: "Is API improvement marginal?"

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

Create experiments/bootstrap-006-api-design/iteration-N.md with:

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
  new_capabilities:
    - name: capability_name
      file: meta-agents/{capability}.md
      reason: "Why was this capability needed?"
      trigger: "What API problem triggered this?"

  OR

M_N = M_{N-1}: "No evolution (unchanged)"
```

### 3. Agent Set Evolution (if applicable)
```yaml
A_{N-1} → A_N:
  new_agents:
    - name: agent_name
      specialization: api_domain
      capabilities: [list]
      creation_reason: "Why was generic agent insufficient?"

  OR

A_N = A_{N-1}: "No new agents (using existing agents)"

agents_invoked_this_iteration:
  - agent_name: api_task_performed
```

### 4. Work Executed
- What API work was done this iteration?
- What API outputs were produced?
- Summary of API deliverables

### 5. State Transition
```yaml
s_{N-1} → s_N:
  changes:
    - description of API state changes

  metrics:
    V_usability: [value]
    V_consistency: [value]
    V_completeness: [value]
    V_evolvability: [value]

  value_function:
    V(s_N): [calculated]
    V(s_{N-1}): [previous]
    ΔV: [change]
    percentage: +X.X%
```

### 6. Reflection
- What was learned about API design this iteration?
- What worked well?
- What challenges were encountered?
- What is needed next?

### 7. Convergence Check
[Use the convergence criteria structure above]

### 8. Data Artifacts
- Save any API data generated to data/ directory
- Reference data files in iteration document
- Examples:
  - data/iteration-N-api-metrics.yaml
  - data/iteration-N-api-proposals.jsonl
  - data/iteration-N-consistency-report.yaml

## Key Principles

1. **Be Honest**: Calculate V(s_N) based on actual API state, not desired values
2. **Let System Evolve**: Don't force predetermined agent names or evolution paths
3. **Justify Specialization**: Only create new agents when generic agents are truly insufficient
4. **Document Evolution**: Clearly explain WHY M or A evolved
5. **Check Convergence**: Rigorously evaluate convergence criteria each iteration
6. **Stop When Done**: If converged, don't force more iterations
7. **No Token Limits**: There are NO token budget constraints in this experiment
   - Do NOT skip steps due to perceived token limits
   - Do NOT abbreviate API data collection or analysis
   - Do NOT summarize when full details are needed
   - Complete ALL steps thoroughly regardless of length
8. **Modular Meta-Agent Architecture**: Capability files, not versioned monoliths
   - Capability files: experiments/bootstrap-006-api-design/meta-agents/*.md (discovered via search, not predetermined)
   - Count: X capabilities (from union of experiments/bootstrap-*)
   - Read: ALWAYS read ALL capability files at start of iteration
   - Read: ALWAYS read specific capability file before using it
   - Update: Create new capability file when M evolves (new coordination needs)
   - Do NOT create meta-agent-m{N}.md files
9. **Agent Prompt Files Required**: Every agent must have a prompt file
   - Agent files: experiments/bootstrap-006-api-design/agents/{agent-name}.md
   - Read: ALWAYS read agent prompt file before agent invocation
   - Update: Create new agent file when agent is specialized
   - Update: Modify agent prompt files as agents evolve

## Common Iteration Patterns

Based on OCA framework, iterations may follow:

- **Observe Phase** (Iterations 0-1): API audit, usage pattern discovery, inconsistency identification
- **Codify Phase** (Iteration 2-3): Design principles extraction, naming conventions, consistency rules
- **Automate Phase** (Iteration 3-4): API validation tools, consistency checkers, migration helpers

But let the actual API design needs drive the sequence, not this expected pattern.
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
- All API task objectives completed

## Objectives

Create experiments/bootstrap-006-api-design/results.md analyzing:

### 1. Three-Tuple Output Analysis

**Output O**:
- List all API deliverables produced
  - API design methodology document
  - Consistency checker tools
  - API validation utilities
  - Design guideline documents
- Calculate total lines of code/documentation
- Assess quality and completeness
- Validate against task objectives

**Agent Set A_N**:
- List all agents in converged set (specialized + generic)
- Calculate specialization ratio: specialized_count / total_count
- Analyze utilization: How much was each agent used?
- Assess reusability: Which agents are transferable to other API design tasks?

**Meta-Agent M_N**:
- List all capabilities (M₀ discovered set + any evolved capabilities)
  - M₀ capabilities: X files (from union search)
  - Evolved capabilities: Y files (created during iterations)
  - Total M_N capabilities: X+Y files
- Document learned policy (agent selection strategy for API design)
- Analyze evolution timeline: M₀ → M_N
- Assess transferability: Which capabilities are domain-independent vs API-specific?

### 2. Convergence Validation

- Formally verify convergence criteria
- Analyze convergence speed (iterations needed)
- Plot value trajectory: V(s₀) → V(s₁) → ... → V(s_N)
- Identify S-curve pattern and diminishing returns

### 3. Value Space Analysis

- Plot V(s) over iterations
- Analyze component contributions (usability, consistency, completeness, evolvability)
- Calculate total value improvement: ΔV_total = V(s_N) - V(s₀)
- Identify which iterations provided largest gains
- Break down by component: Which improved most?

### 4. API Design Methodology Extraction

**Naming Conventions**:
- Extract standardized naming patterns (query_, get_, list_)
- Document parameter naming conventions (snake_case)
- Codify response format patterns

**Parameter Design Patterns**:
- Required vs optional parameter guidelines
- Default value strategy
- Parameter ordering conventions
- Common parameter patterns (scope, jq_filter, stats_only)

**Response Format Patterns**:
- Inline vs file_ref mode decision criteria
- Error message format standards
- Status code conventions

**Evolvability Strategy**:
- Breaking vs non-breaking change guidelines
- Deprecation policy
- Migration path patterns
- Backward compatibility testing

### 5. Reusability Validation

Simulate transfer tests:

**Transfer Test 1** (Similar domain):
- Task: Apply (A_N, M_N) to new API design project (e.g., REST API design)
- Estimate: How many iterations needed? (should be fewer than N)
- Estimate: Which agents can be reused directly?
- Estimate: Speedup factor

**Transfer Test 2** (Different domain):
- Task: Apply (A_N, M_N) to different task (e.g., CLI design)
- Analyze: Which agents transfer? Which need adaptation?
- Analyze: Which M_N capabilities transfer?
- Estimate: Effort savings vs. from-scratch

### 6. Comparison with Actual History

Compare experiment results with actual meta-cc API development:
- Timeline: Experiment vs. actual duration
- Outputs: API design decisions, parameter patterns
- Process: Iteration pattern vs. actual API evolution phases
- Validate: Does experiment accurately simulate reality?
- Lessons: What would have been done differently with this methodology?

### 7. Methodology Validation

Validate alignment with three frameworks:

**Empirical Methodology Development (OCA)**:
- ✓ Observe phase: API data collection and analysis
- ✓ Codify phase: Design principles extraction
- ✓ Automate phase: Validation tool creation

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
- What worked well in API design methodology development?
- What was surprising about API evolution patterns?
- What would you do differently?
- What are the implications for future API design work?

### 9. Scientific Contribution

Articulate what this experiment demonstrates:
- Does it validate the bootstrapping hypothesis for API design?
- Does it demonstrate value space optimization?
- Does it show agent specialization emergence?
- What is novel or significant about API design methodology?

### 10. Future Work

Suggest extensions:
- Apply methodology to other meta-cc APIs (CLI, Slash Commands)
- Multi-domain API design (REST, GraphQL, gRPC)
- Automated API consistency checking
- API evolution prediction models

## Output Format

Structure results.md with:
- Executive Summary
- Three-Tuple Output Analysis (detailed)
- Convergence Validation
- Value Space Trajectory
- API Design Methodology Extraction
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
- API consistency improvement chart
```

---

## Quick Reference: Iteration Checklist

For each iteration N ≥ 1, ensure you:

- [ ] Review previous iteration (iteration-[N-1].md)
- [ ] Extract current state (M_{N-1}, A_{N-1}, V(s_{N-1}))
- [ ] **LIST AND READ ALL META-AGENT CAPABILITY FILES**: `ls meta-agents/*.md` then read EACH file before starting
- [ ] **READ OBSERVATION CAPABILITY**: Identify and read observation capability file (e.g., observe.md) before observation phase
- [ ] OBSERVE: Identify API needs and gaps (using M's observation capability)
- [ ] **READ PLANNING CAPABILITY**: Identify and read planning capability file (e.g., plan.md) before planning phase
- [ ] PLAN: Define iteration goal for API design (using M's planning capability)
- [ ] **READ EVOLUTION CAPABILITY**: Identify and read evolution capability file (e.g., evolve.md) if considering agent creation
- [ ] DECIDE: Create new agent? Add M capability?
- [ ] **IF NEW AGENT**: Create agent prompt file in agents/{agent-name}.md
- [ ] **IF M EVOLVES**: Create new capability file in meta-agents/{capability}.md
- [ ] **READ EXECUTION CAPABILITY**: Identify and read execution capability file (e.g., execute.md) before execution phase
- [ ] **BEFORE AGENT EXECUTION**: Read agent prompt file(s) for agents to be invoked
- [ ] EXECUTE: Invoke agents, produce API outputs (using M's execution capability)
- [ ] **READ REFLECTION CAPABILITY**: Identify and read reflection capability file (e.g., reflect.md) before reflection phase
- [ ] REFLECT: Evaluate quality, calculate V(s_N) (using M's reflection capability)
- [ ] CHECK CONVERGENCE: Apply formal criteria
- [ ] DOCUMENT: Create iteration-N.md
- [ ] SAVE DATA: Store API metrics and artifacts in data/
- [ ] **NO TOKEN LIMITS**: Verify all steps completed fully without abbreviation

If CONVERGED:
- [ ] Create results.md
- [ ] Perform reusability analysis
- [ ] Compare with actual history
- [ ] Document three-tuple (O, A_N, M_N)

---

## Notes on Execution Style

**Be the Meta-Agent**: When executing iterations, embody M's perspective for API design:
- Think through the observe-plan-execute-reflect-evolve cycle
- Make explicit decisions about agent creation
- Justify why API specialization is needed
- Track your own capability evolution

**Be Rigorous**:
- Calculate V(s) honestly based on actual API state
- Don't force convergence prematurely
- Don't skip iterations to reach a predetermined end
- Let the API data and needs drive the process

**Be Thorough**:
- Document API decisions and reasoning
- Save intermediate API data
- Show your work (calculations, analysis)
- Make evolution path traceable
- **NO TOKEN LIMITS**: Complete all API analysis steps fully, never abbreviate due to length concerns

**Be Authentic**:
- This is a real API design experiment, not a simulation
- Discover API patterns, don't assume them
- Create agents based on need, not predetermined plan
- Stop when truly converged, not at a target iteration count

**Modular Meta-Agent Architecture**:
- **Capability files**: experiments/bootstrap-006-api-design/meta-agents/
  - **ALWAYS** read ALL capability files at start of iteration
  - **ALWAYS** read specific capability file before using it
  - Create new capability file when M evolves (new coordination needs)
  - File captures specific capability strategies and decision processes
  - Never assume Meta-Agent behavior - always read from files for full details
- **Agent files**: experiments/bootstrap-006-api-design/agents/
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

**Document Version**: 1.0
**Created**: 2025-10-15
**Purpose**: Guide authentic execution of bootstrap-006-api-design experiment
**Architecture**: Modular Meta-Agent (separate capability files, not versioned monolith)

**Key Architecture Decision**:
- **Modular Meta-Agent**: M₀ has X separate capability files (discovered from union of experiments/bootstrap-*)
  - Typical capabilities: observe.md, plan.md, execute.md, reflect.md, evolve.md
  - May include additional capabilities discovered during search
  - Count and composition determined by search, NOT predetermined
- **NOT Monolithic Meta-Agent**: Do NOT create meta-agent-m0.md, meta-agent-m1.md, etc.
- **Rationale**: Better understandability, maintainability, and evolvability
- **Evolution**: Add new capability files when needed, don't version entire Meta-Agent
- **Discovery-Driven**: Initial state (M₀, A₀) is union of all discovered files, maximizing reusability

**Alignment with Bootstrap-001**:
This document follows the same modular architecture pattern validated in bootstrap-001-doc-methodology, but extends it with discovery-driven initialization from prior experiment results.
