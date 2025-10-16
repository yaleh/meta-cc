# Experiment Output Template

**Version**: 1.0
**Purpose**: Standardize experiment outputs for maximum reusability aligned with `(M, A, K)` System State model

---

## Output Structure

Every experiment should produce the following artifacts:

```
experiments/bootstrap-NNN-topic-name/
├── README.md                   # Experiment overview and results
├── METHODOLOGY.md              # K: Knowledge base (patterns, principles)
├── AGENTS.yaml                 # A: Agent set definitions
├── META-AGENT.yaml             # M: Meta-agent capabilities
├── tools/                      # A: Executable implementations
│   ├── agent-1.sh
│   ├── agent-2.py
│   └── README.md
├── validation/                 # Test and validation scripts
│   ├── test-pattern-1.sh
│   ├── test-pattern-2.sh
│   └── README.md
└── meta-value-trajectory.yaml  # Value function trajectory data
```

---

## 1. README.md (Experiment Overview)

```markdown
# Bootstrap-NNN: [Topic Name]

## Overview
- **Experiment Goal**: [What methodology to extract]
- **Approach**: [How extraction was performed]
- **Iterations**: [Number of iterations performed]
- **Convergence**: [Convergence status and value]

## Outputs

| Output Type | File | Description | Reusability |
|-------------|------|-------------|-------------|
| Knowledge (K) | METHODOLOGY.md | [Patterns extracted] | ⭐⭐⭐⭐⭐ |
| Agents (A) | AGENTS.yaml + tools/ | [Agents defined] | ⭐⭐⭐⭐ |
| Meta-Agent (M) | META-AGENT.yaml | [Meta-capabilities] | ⭐⭐⭐ |

## Validation Evidence
- **Iterations**: N iterations, converged at iteration X
- **Success Rate**: Y% of tasks completed successfully
- **Reusability Test**: Applied to [list of projects]

## Related Documents
- [Link to methodology in docs/methodology/]
- [Link to related experiments]
```

---

## 2. METHODOLOGY.md (Knowledge Base)

### Structure

```markdown
# [Topic] Methodology

## Overview
[High-level description]

## Pattern Catalog

### Pattern 1: [Name]

#### Context (When to Use)
[Description]

#### Problem
[Problem statement]

#### Solution
[Solution description]

#### Agent Specification (Machine-Readable)
```yaml
agent:
  id: pattern-1-agent
  type: [verification/transformation/analysis]
  inputs:
    - name: input1
      type: type1
  process:
    - step: step1
      tool: tool_name
  outputs:
    - name: output1
      type: type1
```

#### Decision Tree (Machine-Readable)
```yaml
decision_tree:
  root:
    question: "Question text?"
    yes: node_yes
    no: node_no
  node_yes:
    action: "Action to take"
```

#### Implementation
[Detailed steps or reference to tools/pattern-1-agent.sh]

#### Evidence
- **Observed in**: [Experiment/Iteration]
- **Success Rate**: X%
- **Example**: [Concrete example]

#### Reusability
```yaml
reusability:
  transferability:
    languages: [list]
    domains: [list]
  prerequisites:
    tools: [list]
  evidence:
    validated_on:
      - project: name
        success_rate: X%
```

[Repeat for Pattern 2, 3, ...]

## Pattern Application Framework
[Decision trees for pattern selection]

## Decision Trees
[High-level decision guidance]

## Success Metrics
[Quantitative and qualitative metrics]
```

---

## 3. AGENTS.yaml (Agent Set Definitions)

```yaml
# Agent Set for [Topic] Methodology
# Extracted from Bootstrap-NNN experiment

version: 1.0
agents:
  - id: agent-1
    name: "[Descriptive Name]"
    pattern: "Pattern 1"
    role: "[What this agent does]"

    input_schema:
      - name: input1
        type: type1
        description: "[Description]"

    process:
      - step: "[Step description]"
        tool: "[Tool name or 'manual']"
        args: [list of args]

    output_schema:
      - name: output1
        type: type1
        description: "[Description]"

    implementation:
      type: [script/interactive/algorithm/workflow]
      path: "tools/agent-1.sh"  # or null if interactive
      guidance: "METHODOLOGY.md#pattern-1"  # for interactive agents

    reusability:
      languages: [Go, Python, JavaScript]
      domains: [refactoring, testing, documentation]
      requires:
        tools: [tool1, tool2]
        knowledge: [prerequisite concepts]
        data: [required data types]

    evidence:
      extracted_from:
        experiment: "Bootstrap-NNN"
        iteration: 2
      validated_on:
        - project: "meta-cc"
          iterations: 3
          success_rate: 0.95

  - id: agent-2
    name: "[Descriptive Name]"
    # ... same structure ...
```

---

## 4. META-AGENT.yaml (Meta-Cognitive Capabilities)

```yaml
# Meta-Agent for [Topic] Methodology
# Defines decision-making capabilities for agent selection and task orchestration

version: 1.0
meta_agent:
  id: "[topic]-meta"
  name: "[Topic] Meta-Agent"

  capabilities:
    - name: select_pattern
      description: "Select appropriate pattern(s) for given context"

      input_schema:
        context:
          [context fields]: type
        goals: list[string]

      decision_tree:
        path: "METHODOLOGY.md#pattern-application-framework"
        # or inline:
        root:
          question: "[Question]"
          yes: node1
          no: node2

      output_schema:
        recommended_patterns: list[pattern_id]
        rationale: string
        confidence: number[0.0-1.0]

    - name: assess_risk
      description: "Assess risk level of proposed task"

      input_schema:
        task_description: string
        code_metrics:
          complexity: number
          coverage: number

      formula:
        safety: "[Mathematical formula]"

      output_schema:
        risk_level: enum[low, medium, high]
        safety_score: number[0.0-1.0]
        recommendations: list[string]

    - name: prioritize_tasks
      description: "Prioritize multiple tasks based on value, safety, effort"

      input_schema:
        tasks: list[task]
        constraints:
          time_budget: number
          quality_threshold: number

      formula:
        priority: "[Mathematical formula]"

      output_schema:
        sorted_tasks: list[task]
        skip_recommendations: list[task_id]
        rationale: string

    - name: orchestrate_workflow
      description: "Coordinate multiple agents to complete complex task"

      input_schema:
        task: complex_task
        available_agents: list[agent_id]

      process:
        - step: "Decompose task into subtasks"
        - step: "Assign agents to subtasks"
        - step: "Monitor execution and adapt"

      output_schema:
        execution_plan: list[subtask_assignment]
        coordination_rules: list[rule]

  reusability:
    transferability:
      domains: [list of domains]
      projects: [list of project types]

    adaptation_guide:
      "[language/domain]": "[Adaptation notes]"

  evidence:
    extracted_from:
      experiment: "Bootstrap-NNN"
      observation_method: "[Two-layer architecture / other]"

    validated_on:
      - project: "meta-cc"
        tasks: [list of tasks]
        success_rate: 0.89
```

---

## 5. tools/ (Executable Implementations)

Each agent with `implementation.type: script` should have a corresponding file:

```bash
#!/bin/bash
# tools/agent-1.sh
# Agent: [Name]
# Pattern: Pattern 1
# Description: [What this script does]

set -euo pipefail

# Input validation
if [ $# -lt 1 ]; then
    echo "Usage: $0 <input1> [input2]"
    exit 1
fi

# Main logic
# [Implementation]

# Output (JSON format for machine parsing)
echo '{"output1": "value1", "output2": "value2"}'
```

### tools/README.md

```markdown
# Tools Directory

## Available Agents

| Agent ID | Script | Language | Description |
|----------|--------|----------|-------------|
| agent-1 | agent-1.sh | Bash | [Description] |
| agent-2 | agent-2.py | Python | [Description] |

## Usage

```bash
# Agent 1 example
./agent-1.sh input_value

# Agent 2 example
python agent-2.py --input input_value
```

## Dependencies

- agent-1: requires `staticcheck`, `ripgrep`
- agent-2: requires `pytest`, `coverage`
```

---

## 6. validation/ (Test and Validation Scripts)

```bash
#!/bin/bash
# validation/test-pattern-1.sh
# Test validation for Pattern 1

set -euo pipefail

echo "Testing Pattern 1: [Name]"

# Setup test environment
# [Create test fixtures]

# Run agent
result=$(../tools/agent-1.sh test_input)

# Validate output
expected='{"output1": "expected_value"}'
if [ "$result" == "$expected" ]; then
    echo "✓ Test passed"
    exit 0
else
    echo "✗ Test failed"
    echo "  Expected: $expected"
    echo "  Got: $result"
    exit 1
fi
```

### validation/README.md

```markdown
# Validation Scripts

## Running Tests

```bash
# Run all tests
./run-all-tests.sh

# Run specific test
./test-pattern-1.sh
```

## Test Coverage

| Pattern | Test Script | Status |
|---------|-------------|--------|
| Pattern 1 | test-pattern-1.sh | ✓ Pass |
| Pattern 2 | test-pattern-2.sh | ✓ Pass |
```

---

## 7. meta-value-trajectory.yaml (Value Function Data)

```yaml
# Value Function Trajectory for Bootstrap-NNN Experiment
# Used for training Meta-Agent and validating convergence

experiment:
  id: "Bootstrap-NNN"
  topic: "[Topic Name]"
  start_date: "2025-10-16"
  end_date: "2025-10-16"
  total_iterations: N

value_function:
  components:
    - name: "code_quality"
      weight: 0.30
      metric: "lint_violations / LOC"
    - name: "maintainability"
      weight: 0.30
      metric: "subjective assessment"
    - name: "safety"
      weight: 0.20
      metric: "test_coverage"
    - name: "effort"
      weight: 0.20
      metric: "1 - (remaining_work / total_work)"

trajectory:
  - iteration: 0
    timestamp: "2025-10-16T10:00:00Z"
    state:
      code_quality: 0.65
      maintainability: 0.60
      safety: 0.70
      effort: 0.00
    value: 0.487
    action: "[Action description]"
    agent: "initial"

  - iteration: 1
    timestamp: "2025-10-16T11:30:00Z"
    state:
      code_quality: 0.70
      maintainability: 0.65
      safety: 0.75
      effort: 0.30
    value: 0.600
    delta_value: +0.113
    action: "[Action description]"
    agent: "agent-1"
    success: true

  # ... more iterations ...

convergence:
  achieved: true
  iteration: N
  final_value: 0.804
  threshold: 0.80
  criteria:
    - "V(s) ≥ 0.80"
    - "ΔV < 0.05 for 2 consecutive iterations"

  evidence:
    iterations_to_converge: N
    success_rate: 0.95
    value_improvement: "+0.317 (from 0.487 to 0.804)"

errors:
  - iteration: 3
    error_type: "test_failure"
    immediate_cost: -2.0  # hours
    knowledge_value: +3.5  # hours (prevented future errors)
    prevention_value: +5.0  # hours (automated check added)
    tooling_value: +2.0  # hours (linter rule)
    net_value: +8.5  # Total: 425% ROI

agent_evolution:
  - iteration: 1
    agents: ["generic-coder"]
    specialization: "low"

  - iteration: 3
    agents: ["agent-1", "agent-2", "generic-coder"]
    specialization: "medium"
    change: "Split generic-coder into agent-1 (verification) and agent-2 (transformation)"

  - iteration: N
    agents: ["agent-1", "agent-2", "agent-3", "agent-4"]
    specialization: "high"
    change: "Added agent-3 (prioritization) and agent-4 (testing)"

meta_agent_evolution:
  - iteration: 1
    capabilities: ["observe", "plan", "execute"]

  - iteration: 2
    capabilities: ["observe", "plan", "execute", "select_pattern"]
    change: "Added pattern selection capability"

  - iteration: N
    capabilities: ["observe", "plan", "execute", "select_pattern", "assess_risk", "prioritize_tasks"]
    change: "Added risk assessment and task prioritization"
```

---

## Usage Guide

### For Methodology Consumers

1. **Read**: Start with `METHODOLOGY.md` for human understanding
2. **Reference**: Use `AGENTS.yaml` to understand agent specifications
3. **Adapt**: Modify `tools/` scripts for your language/domain
4. **Validate**: Run `validation/` tests to verify adaptation

### For Meta-Agent Training

1. **Load**: Parse `AGENTS.yaml` and `META-AGENT.yaml`
2. **Train**: Use `meta-value-trajectory.yaml` for training data
3. **Validate**: Use experiment evidence to validate transferability

### For Experiment Reproducibility

1. **Context**: Read `README.md` for experiment context
2. **Data**: Use `meta-value-trajectory.yaml` for trajectory data
3. **Verify**: Run validation scripts to reproduce results

---

## Checklist for Experiment Completion

Before publishing experiment outputs, ensure:

- [ ] `README.md` provides clear overview and links
- [ ] `METHODOLOGY.md` contains all patterns with machine-readable specs
- [ ] `AGENTS.yaml` defines all agents with complete schemas
- [ ] `META-AGENT.yaml` defines meta-cognitive capabilities
- [ ] `tools/` contains executable implementations (or guidance for interactive)
- [ ] `validation/` contains test scripts for each pattern
- [ ] `meta-value-trajectory.yaml` documents value function and convergence
- [ ] All reusability metadata is complete (languages, domains, prerequisites)
- [ ] Evidence sections cite specific iterations and metrics
- [ ] Cross-references between files are correct

---

## Example: Ideal Experiment Output

See `experiments/bootstrap-006-api-design/` for a nearly-complete example that includes:
- ✅ Comprehensive `API-DESIGN-METHODOLOGY.md`
- ✅ Implemented tools (validate-api, pre-commit hooks)
- ⚠️ Missing `AGENTS.yaml` (should be added)
- ⚠️ Missing `META-AGENT.yaml` (should be added)
- ⚠️ Missing `meta-value-trajectory.yaml` (should be added)

---

**Last Updated**: 2025-10-16
**Status**: Template v1.0
**Next Review**: After next experiment completion
