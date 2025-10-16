# Extracted Methodology Usage Guide

**Version**: 1.0
**Date**: 2025-10-16
**Status**: Validated

---

## Overview

This guide demonstrates how to use the extracted agents and meta-agents from Bootstrap-004 (Refactoring) and Bootstrap-006 (API Design) experiments. Each methodology has been decomposed into independently reusable components following the `(M, A, K)` System State framework.

---

## Directory Structure

```
experiments/
├── bootstrap-004-refactoring-guide/
│   ├── agents/                    # 4 specialized refactoring agents
│   │   ├── agent-verify-before-remove.md
│   │   ├── agent-builder-extractor.md
│   │   ├── agent-risk-prioritizer.md
│   │   └── agent-test-adder.md
│   ├── meta-agents/               # Refactoring orchestrator
│   │   └── refactoring-orchestrator.md
│   └── knowledge/                 # Principles and patterns
│       ├── principles.md          # 5 core principles
│       └── patterns.md            # 4 refactoring patterns
│
├── bootstrap-006-api-design/
│   ├── agents/                    # 6 specialized API design agents
│   │   ├── agent-parameter-categorizer.md
│   │   ├── agent-schema-refactorer.md
│   │   ├── agent-audit-executor.md
│   │   ├── agent-validation-builder.md
│   │   ├── agent-quality-gate-installer.md
│   │   └── agent-documentation-enhancer.md
│   ├── meta-agents/               # API design orchestrator
│   │   └── api-design-orchestrator.md
│   └── knowledge/                 # Principles and patterns
│       ├── principles.md          # 6 core principles
│       └── patterns.md            # 6 API design patterns
│
└── USAGE-GUIDE.md                 # This file
```

---

## Using Agents (A)

Agents are specialized, independently invocable execution units. Each agent has:
- **Role**: What it does
- **Input Schema**: YAML-formatted inputs
- **Execution Process**: Step-by-step HOW-TO
- **Output Schema**: YAML-formatted outputs
- **Success Criteria**: Validation metrics

### Example 1: Verify Before Removing Code

**Agent**: `bootstrap-004-refactoring-guide/agents/agent-verify-before-remove.md`

**Scenario**: You want to remove `validateToolInput` function but aren't sure if it's used.

**Usage as Subagent**:
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-verify-before-remove.md \
  target_code.file="internal/validation/validation.go" \
  target_code.function="validateToolInput" \
  scope="package"
```

**Expected Output**:
```yaml
verification_result:
  safe_to_remove: false
  reason: "Function is actively used"
  usage_found:
    - file: "internal/handlers/handlers.go"
      line: 123
      context: "Called in request handler"
  recommendation: "DO NOT REMOVE"
```

**Action**: Don't remove the function (it's actively used).

---

### Example 2: Extract Builder Helpers

**Agent**: `bootstrap-004-refactoring-guide/agents/agent-builder-extractor.md`

**Scenario**: You have 15 API tool definitions with duplicate parameter structures.

**Usage as Subagent**:
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-builder-extractor.md \
  target_file.path="cmd/mcp-server/tools.go" \
  target_file.type="api_schema" \
  analysis.duplication_threshold=0.15
```

**Expected Output**:
```yaml
extraction_summary:
  lines_reduced: 75 (18.9%)
  duplication_eliminated: 69 lines (100%)
  tools_refactored: 12/15 (80%)
  exceptions: 3 (documented)
  test_pass_rate: 100%
```

**Action**: 75 lines removed, 100% duplication eliminated, all tests pass.

---

### Example 3: Prioritize Refactoring Tasks

**Agent**: `bootstrap-004-refactoring-guide/agents/agent-risk-prioritizer.md`

**Scenario**: You have 3 refactoring tasks and need to prioritize them objectively.

**Usage as Subagent**:
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-risk-prioritizer.md \
  tasks='[
    {"name": "Extract helpers", "description": "Reduce duplication"},
    {"name": "Split file", "description": "Improve organization"},
    {"name": "Add tests", "description": "Increase coverage"}
  ]' \
  constraints.max_time_available=8 \
  constraints.risk_tolerance="low"
```

**Expected Output**:
```yaml
prioritized_tasks:
  - task: "Extract helpers"
    priority: 1.57 (P1)
    recommendation: "EXECUTE"

  - task: "Add tests"
    priority: 0.78 (P2)
    recommendation: "EXECUTE if time permits"

  - task: "Split file"
    priority: 0.28 (P3)
    recommendation: "SKIP (risky, low ROI)"
```

**Action**: Execute P1, consider P2, skip P3 (risky).

---

### Example 4: Add Tests Incrementally

**Agent**: `bootstrap-004-refactoring-guide/agents/agent-test-adder.md`

**Scenario**: Package has 0% test coverage, need to improve.

**Usage as Subagent**:
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-test-adder.md \
  target_package.path="internal/validation" \
  target_metrics.target_coverage=0.75 \
  test_strategy.focus="exported"
```

**Expected Output**:
```yaml
test_results:
  coverage_before: 0%
  coverage_after: 32.5%
  improvement: +32.5 percentage points
  tests_added: 18
  all_pass: true
```

**Action**: Coverage improved from 0% to 32.5%, all tests passing.

---

### Example 5: Categorize API Parameters

**Agent**: `bootstrap-006-api-design/agents/agent-parameter-categorizer.md`

**Scenario**: Need to order parameters consistently across 8 API tools.

**Usage as Subagent**:
```bash
/subagent @experiments/bootstrap-006-api-design/agents/agent-parameter-categorizer.md \
  target_api.file="cmd/mcp-server/tools.go" \
  target_api.type="json_schema" \
  categorization.add_comments=true
```

**Expected Output**:
```yaml
categorization_report:
  parameters_categorized: 60
  ambiguous_cases: 0
  determinism_rate: 100%
  tools_reordered: 5
  compliance: 67.5% → 100%
```

**Action**: 60 parameters categorized, 100% determinism, 100% compliance achieved.

---

### Example 6: Audit Before Refactoring

**Agent**: `bootstrap-006-api-design/agents/agent-audit-executor.md`

**Scenario**: Need to refactor 8 tools, but unclear which ones actually need changes.

**Usage as Subagent**:
```bash
/subagent @experiments/bootstrap-006-api-design/agents/agent-audit-executor.md \
  audit_scope.targets='["tool1", "tool2", "tool3", ..., "tool8"]' \
  compliance_criteria.convention="tier_based_ordering" \
  compliance_criteria.threshold=1.0
```

**Expected Output**:
```yaml
audit_report:
  targets_audited: 8
  already_compliant: 3 (37.5%)
  needs_change: 5 (62.5%)
  efficiency_gain: 37.5% (avoided 3/8 changes)
  time_saved: 90 minutes
```

**Action**: Refactor 5 tools (skip 3 already-compliant), saved 90 minutes.

---

## Using Meta-Agents (M)

Meta-agents orchestrate multiple agents with decision logic and convergence criteria.

### Example 7: Full Refactoring Workflow

**Meta-Agent**: `bootstrap-004-refactoring-guide/meta-agents/refactoring-orchestrator.md`

**Scenario**: Large refactoring project (improve code quality, reduce duplication, add tests).

**Usage as Subagent**:
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/meta-agents/refactoring-orchestrator.md \
  refactoring_goal.target="complete" \
  current_state.test_coverage=0.579 \
  current_state.duplication=0.174 \
  constraints.max_time_hours=8
```

**Orchestration Flow**:
```yaml
Phase 1: Assessment
  - Measure current state (coverage, duplication, complexity)
  - V(s₀) = 0.77 (below 0.80 threshold)

Phase 2: Task Prioritization
  - A₃ (Risk Prioritizer): Calculate priorities
    - Task 1 (Extract builders): P=1.57 (P1)
    - Task 2 (Split file): P=0.28 (P3, SKIP)
    - Task 3 (Add tests): P=0.78 (P2)

Phase 3: Execution
  - Execute A₂ (Builder Extractor):
    - 75 lines reduced, ΔV = +0.020
  - Execute A₄ (Test Adder):
    - Coverage 0% → 32.5%, ΔV = +0.014
  - Skip Task 2 (P3, risky)

Phase 4: Convergence Check
  - V(s₁) = 0.804 ≥ 0.80 → CONVERGED

Phase 5: Documentation
  - Generated report with metrics and rationale
```

**Expected Output**:
```yaml
orchestration_report:
  status: "CONVERGED"
  phases_executed: 3
  initial_value: 0.77
  final_value: 0.804
  tasks_completed: 2/3
  tasks_skipped: 1 (P3, risky)
  convergence_achieved: true
```

**Action**: Refactoring complete, converged at V=0.804 (above threshold).

---

### Example 8: Full API Design Workflow

**Meta-Agent**: `bootstrap-006-api-design/meta-agents/api-design-orchestrator.md`

**Scenario**: Improve API consistency, add automation, enhance documentation.

**Usage as Subagent**:
```bash
/subagent @experiments/bootstrap-006-api-design/meta-agents/api-design-orchestrator.md \
  api_design_goal.target="complete" \
  current_state.compliance_rate=0.675 \
  current_state.automation_level="none" \
  constraints.max_time_hours=12
```

**Orchestration Flow**:
```yaml
Phase 1: Assessment
  - A₃ (Audit): 8 tools audited
    - Compliant: 3 (37.5%)
    - Needs change: 5 (62.5%)

Phase 2: Consistency Improvement
  - A₁ (Categorize): 60 parameters categorized
  - A₂ (Refactor): 5 tools reordered
  - Compliance: 67.5% → 100%

Phase 3: Build Automation
  - A₄ (Build Validator): 3 validators created
  - A₅ (Install Quality Gate): Pre-commit hook installed

Phase 4: Enhance Documentation
  - A₆ (Enhance Docs): 3 low-usage tools documented (11 examples)

Phase 5: Verification
  - Compliance: 100%
  - Automation: Validation tool + hook installed
  - Documentation: 25 tested examples (100% accuracy)
```

**Expected Output**:
```yaml
orchestration_report:
  status: "COMPLETE"
  phases_executed: 4
  compliance: 67.5% → 100%
  automation: Validation tool + pre-commit hook
  documentation: 3 tools enhanced, 25 examples
  tests_passed: 100%
  breaking_changes: 0
```

**Action**: API design complete, 100% compliance, full automation, enhanced docs.

---

## Using Knowledge Base (K)

Knowledge base provides theoretical understanding (WHY) separated from execution (HOW).

### Example 9: Learn Refactoring Principles

**Knowledge**: `bootstrap-004-refactoring-guide/knowledge/principles.md`

**Use Case**: Understand WHY to verify before removing code.

**Content Structure**:
```markdown
# Principle 1: Verify Before Changing

### Statement
Never trust assumptions about code usage. Always verify objectively.

### Why This Matters
- Intuition about "unused code" is frequently wrong
- Removing used code causes production failures
- Hours wasted debugging

### Evidence
- Scenario: Developer claimed "validation logic unused"
- Verification: `rg "validateToolInput"` found usage
- Outcome: Prevented removal, saved 2-4 hours

### How to Apply
1. Run static analyzer + grep
2. Check test coverage
3. Verify at appropriate scope
```

**Learning Outcome**: Understand principle, see evidence, apply to your context.

---

### Example 10: Learn API Design Patterns

**Knowledge**: `bootstrap-006-api-design/knowledge/patterns.md`

**Use Case**: Understand deterministic parameter categorization pattern.

**Content Structure**:
```markdown
# Pattern 1: Deterministic Parameter Categorization

### Context
Need consistent parameter ordering across all tools.

### Problem
Arbitrary ordering, inconsistent schemas, cognitive load.

### Solution
5-tier decision tree (Required → Filtering → Range → Output → Standard)

### Implementation
Agent: agent-parameter-categorizer.md

### Consequences
- Benefits: 100% determinism, consistent
- Tradeoffs: Need convention definition

### Evidence
- 60 parameters categorized, 0 ambiguous cases
```

**Learning Outcome**: Understand pattern, see context/problem/solution, apply agent.

---

## Validation Results

### Bootstrap-004 (Refactoring) Validation

**Reusability Metrics**:
```yaml
agents:
  - agent-verify-before-remove:
      language_agnostic: true
      domain_agnostic: true
      tool_agnostic: true
      evidence: Saved 2-4 hours (24-48x ROI)

  - agent-builder-extractor:
      language_agnostic: true (adapt syntax)
      domain_agnostic: true (any structure definitions)
      evidence: 75 lines reduced (18.9%)

  - agent-risk-prioritizer:
      language_agnostic: true
      domain_agnostic: true (any constrained optimization)
      evidence: Enabled convergence (skipped P3)

  - agent-test-adder:
      language_agnostic: true (adapt frameworks)
      domain_agnostic: true (any code with tests)
      evidence: 0% → 32.5% coverage improvement

meta_agent:
  - refactoring-orchestrator:
      coordination_logic: Universal (applicable to any refactoring workflow)
      convergence_formula: V(s) ≥ 0.80
      evidence: V=0.804 achieved with 2/3 tasks

knowledge_base:
  - principles.md: 5 principles, language/domain agnostic
  - patterns.md: 4 patterns, reusability matrix provided
```

---

### Bootstrap-006 (API Design) Validation

**Reusability Metrics**:
```yaml
agents:
  - agent-parameter-categorizer:
      language_agnostic: true
      domain_agnostic: true (parametric interfaces)
      evidence: 100% determinism (0 ambiguous cases)

  - agent-schema-refactorer:
      json_specific: true (JSON property guarantee)
      domain_agnostic: true (JSON-based APIs)
      evidence: 0 breaking changes

  - agent-audit-executor:
      language_agnostic: true
      domain_agnostic: true (any multi-target refactoring)
      evidence: 37.5% efficiency gain

  - agent-validation-builder:
      language_agnostic: true (adapt tools)
      domain_agnostic: true (any conventions)
      evidence: 0 false positives, 100% accuracy

  - agent-quality-gate-installer:
      language_agnostic: true (Git hooks universal)
      domain_agnostic: true (any quality check)
      evidence: 100% violation prevention

  - agent-documentation-enhancer:
      language_agnostic: true
      domain_agnostic: true (any documentation)
      evidence: 25 examples (100% accuracy)

meta_agent:
  - api-design-orchestrator:
      coordination_logic: Universal (API design workflow)
      convergence_criteria: 100% compliance + full automation
      evidence: Complete workflow (iterations 4-6)

knowledge_base:
  - principles.md: 6 principles, language/domain agnostic
  - patterns.md: 6 patterns, reusability matrix provided
```

---

## Integration Examples

### Integrate Bootstrap-004 into Your Project

**Step 1: Choose Agent**
```bash
# Need to verify before removing code?
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-verify-before-remove.md

# Need to reduce duplication?
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-builder-extractor.md
```

**Step 2: Provide Context-Specific Inputs**
```bash
# Your project-specific values
target_code.file="src/my-module/validation.ts"
target_code.function="validateInput"
scope="project"
```

**Step 3: Adapt Output to Your Workflow**
```bash
# Agent returns YAML output
# Parse and use in your automation
```

---

### Integrate Bootstrap-006 into Your Project

**Step 1: Choose Workflow Phase**
```bash
# Phase 1: Consistency
/subagent @.../agent-parameter-categorizer.md
/subagent @.../agent-schema-refactorer.md

# Phase 2: Automation
/subagent @.../agent-validation-builder.md
/subagent @.../agent-quality-gate-installer.md

# Phase 3: Documentation
/subagent @.../agent-documentation-enhancer.md
```

**Step 2: Use Meta-Agent for Full Workflow**
```bash
# Orchestrate all phases
/subagent @.../api-design-orchestrator.md \
  api_design_goal.target="complete"
```

---

## Success Metrics

### Extraction Quality

**Structure**:
- ✅ Separated M (Meta-Agent), A (Agents), K (Knowledge)
- ✅ Agents have YAML Input/Output schemas
- ✅ Meta-Agents have decision logic and convergence criteria
- ✅ Knowledge base has principles and patterns

**Reusability**:
- ✅ Agents independently invocable (as subagents)
- ✅ Cross-references between M, A, K
- ✅ Language/domain/tool adaptability documented
- ✅ Evidence and metrics provided

**Validation**:
- ✅ Bootstrap-004: 4 agents, 1 meta-agent, 2 knowledge docs
- ✅ Bootstrap-006: 6 agents, 1 meta-agent, 2 knowledge docs
- ✅ All agents have usage examples
- ✅ All agents have evidence from experiments

---

## Maintenance

### Updating Agents

**When to Update**:
- New evidence from usage
- Bug fixes in process
- Adaptations for new languages/tools
- Improved metrics or examples

**How to Update**:
1. Read current agent file
2. Identify section to update (e.g., Evidence, Variations)
3. Edit with new content
4. Update "Last Updated" timestamp
5. Document change in version history

---

### Adding New Agents

**Process**:
1. Observe agent execution patterns (TWO-LAYER ARCHITECTURE)
2. Extract decision-making process
3. Codify as agent.md with Input/Output schemas
4. Validate through execution
5. Document evidence and metrics
6. Add to methodology directory

---

## References

**Bootstrap-004 (Refactoring)**:
- Source: `experiments/bootstrap-004-refactoring-guide/REFACTORING-METHODOLOGY.md`
- Agents: 4 (verify, extract, prioritize, test)
- Meta-Agent: refactoring-orchestrator
- Knowledge: 5 principles, 4 patterns

**Bootstrap-006 (API Design)**:
- Source: `experiments/bootstrap-006-api-design/API-DESIGN-METHODOLOGY.md`
- Agents: 6 (categorize, refactor, audit, validate, install, document)
- Meta-Agent: api-design-orchestrator
- Knowledge: 6 principles, 6 patterns

**Framework**:
- System State: `Σ = (M, A, K)`
- Output Triple: `(O, Aₙ, Mₙ)`
- Convergence: `V(s) ≥ 0.80`

---

**Last Updated**: 2025-10-16
**Status**: Complete
**Next Steps**: Use agents in your projects, provide feedback, extend with new agents
