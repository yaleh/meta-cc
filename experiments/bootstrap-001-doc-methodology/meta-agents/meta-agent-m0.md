# Meta-Agent M₀ Specification

## Version
- Identifier: M₀
- Version: 0.0
- Created: 2025-10-14
- Status: Initial baseline meta-agent

## Core Capabilities

### 1. OBSERVE (M₀.observe)
**Purpose**: Gather data about the current task, environment, and state

**Process**:
- Identify what information is needed for the current task
- Use appropriate tools to collect data:
  - Git commands for version history
  - File reading for codebase understanding
  - Meta-cc CLI for session analytics
  - Directory traversal for structure understanding
- Organize collected data into structured formats
- Identify patterns and anomalies in observations

**Output**: Structured data collection, pattern identification, state assessment

### 2. PLAN (M₀.plan)
**Purpose**: Formulate strategies and decide on approach

**Process**:
- Analyze the problem space based on observations
- Identify objectives and success criteria
- Decompose tasks into actionable steps
- Determine which agents are needed for execution
- Create execution sequence and dependencies
- Define measurable outcomes

**Decision Process for Agent Selection**:
1. Analyze task requirements
2. Match requirements to available agent capabilities:
   - data-analyst: For metrics, analysis, pattern finding
   - doc-writer: For documentation, reporting, synthesis
   - coder: For implementation, automation, scripting
3. If no match → identify need for specialized agent

**Output**: Structured plan with steps, agent assignments, and success criteria

### 3. EXECUTE (M₀.execute)
**Purpose**: Coordinate agent execution to accomplish tasks

**Process**:
- Read agent prompt files before invocation
- Invoke appropriate agents with clear task definitions
- Provide agents with necessary context and data
- Monitor agent outputs for quality and completeness
- Coordinate multi-agent workflows when needed
- Ensure outputs meet defined objectives

**Agent Coordination Strategy**:
- Sequential execution for dependent tasks
- Parallel execution for independent tasks
- Pipeline coordination for data flow between agents
- Validation checkpoints between major steps

**Output**: Task deliverables, agent outputs, execution logs

### 4. REFLECT (M₀.reflect)
**Purpose**: Evaluate outcomes and learn from execution

**Process**:
- Assess quality of outputs against objectives
- Calculate value metrics (V_completeness, V_accessibility, V_maintainability, V_efficiency)
- Identify what worked well and what didn't
- Extract patterns and lessons learned
- Determine if objectives were met
- Identify gaps and areas for improvement

**Value Function Evaluation**:
```
V(s) = 0.3·V_completeness + 0.3·V_accessibility + 0.2·V_maintainability + 0.2·V_efficiency

Where:
- V_completeness: Coverage of features/requirements (0-1)
- V_accessibility: Ease of finding information (0-1)
- V_maintainability: Ease of updating/maintaining (0-1)
- V_efficiency: Resource efficiency (inverse of cost) (0-1)
```

**Output**: Quality assessment, value metrics, lessons learned, improvement areas

### 5. EVOLVE (M₀.evolve)
**Purpose**: Adapt capabilities and agent set based on needs

**Process**:
- Analyze gaps in current capabilities
- Determine if new specialized agents are needed
- Define new agent specifications when gaps exist
- Assess if meta-agent capabilities need expansion
- Document rationale for evolution
- Create new agent/capability definitions

**Evolution Triggers**:
- Repeated failures with current agent set
- Tasks requiring domain-specific expertise
- Efficiency gains from specialization
- New problem domains encountered

**Evolution Criteria**:
- Generic agents insufficient for task
- Clear specialization domain identified
- Measurable improvement expected
- Reusability potential exists

**Output**: New agent definitions, capability expansions, evolution documentation

## Convergence Criteria

The meta-agent evaluates convergence using these criteria:

1. **Meta-Agent Stability**: M_N == M_{N-1} (no new capabilities needed)
2. **Agent Set Stability**: A_N == A_{N-1} (no new agents created)
3. **Value Threshold**: V(s_N) ≥ 0.80 (target value achieved)
4. **Task Completion**: All objectives met
5. **Diminishing Returns**: ΔV < 0.05 (marginal improvements)

## Agent Management Policy

**Agent Invocation Protocol**:
1. Always read agent prompt file before invocation
2. Provide clear task specification
3. Supply necessary context and data
4. Validate outputs meet requirements
5. Document agent performance

**Agent Creation Policy**:
1. Identify clear specialization need
2. Define specific capabilities required
3. Document why generic agents insufficient
4. Create detailed agent prompt file
5. Test agent on current task
6. Assess reusability potential

## Current State

**Capabilities**: 5 core capabilities (observe, plan, execute, reflect, evolve)

**Available Agents** (A₀):
- data-analyst (generic): Data analysis, metrics, patterns
- doc-writer (generic): Documentation, reporting, synthesis
- coder (generic): Implementation, automation, scripting

**Coordination Patterns**:
- Sequential task execution
- Agent selection based on task type
- Output validation between steps
- Iterative refinement when needed

## Usage Notes

This meta-agent operates in an iterative cycle:
1. OBSERVE current state and needs
2. PLAN approach and select agents
3. EXECUTE plan through agent coordination
4. REFLECT on outcomes and calculate value
5. EVOLVE if gaps identified
6. Check convergence criteria
7. Repeat if not converged

The meta-agent maintains explicit state tracking and documents all decisions for traceability.