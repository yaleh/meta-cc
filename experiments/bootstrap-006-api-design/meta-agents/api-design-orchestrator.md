# Meta-Agent: API Design Orchestrator

**Version**: 1.0
**Source**: Bootstrap-006 API Design Methodology
**Success Rate**: 100% compliance achieved through systematic orchestration

---

## Role

Coordinate 6 specialized agents to enforce API consistency, from parameter ordering to automated quality gates and documentation, ensuring comprehensive API design improvements.

## Agents Managed

```yaml
agents:
  A₁: agent-parameter-categorizer
    role: "Deterministic parameter categorization using tier system"
    when: "Parameter ordering needs consistency"

  A₂: agent-schema-refactorer
    role: "Safe API schema refactoring via JSON property guarantee"
    when: "Schema readability needs improvement"

  A₃: agent-audit-executor
    role: "Audit-first refactoring to identify actual work needed"
    when: "Multiple targets need refactoring"

  A₄: agent-validation-builder
    role: "Build automated validation tools for convention enforcement"
    when: "Need automated consistency checking"

  A₅: agent-quality-gate-installer
    role: "Install pre-commit hooks to prevent violations"
    when: "Need to prevent violations from entering repository"

  A₆: agent-documentation-enhancer
    role: "Example-driven documentation for better adoption"
    when: "Low-usage tools need better docs"
```

## Input Schema

```yaml
api_design_goal:
  target: string                  # "consistency" | "automation" | "documentation" | "complete"
  files: [string]                 # API definition files to process
  conventions: [string]           # Conventions to enforce

current_state:
  compliance_rate: number         # Current % compliance (0.0-1.0)
  automation_level: string        # "none" | "partial" | "full"
  documentation_quality: string   # "poor" | "adequate" | "excellent"

constraints:
  max_time_hours: number          # Time budget
  priority: string                # "safety" | "speed" | "completeness"
  breaking_changes_allowed: boolean # Default: false

execution_config:
  incremental: boolean            # Default: true
  test_after_each: boolean        # Default: true
  document_changes: boolean       # Default: true
```

## Decision Logic

### Phase 1: Assessment

**Goal**: Understand current state, identify gaps

```python
def assess_current_state(api_files):
    state = {
        "consistency": {
            "tools_audited": 0,
            "compliant": 0,
            "violations": [],
            "compliance_rate": 0.0
        },
        "automation": {
            "validation_tool_exists": False,
            "pre_commit_hook_exists": False,
            "ci_integration": False
        },
        "documentation": {
            "conventions_explained": False,
            "low_usage_tools": [],
            "examples_count": 0
        }
    }

    # A₃: Audit all tools
    audit_results = A₃.execute({
        "targets": extract_tools(api_files),
        "compliance_criteria": load_conventions()
    })

    state["consistency"]["tools_audited"] = audit_results.total
    state["consistency"]["compliant"] = audit_results.compliant_count
    state["consistency"]["compliance_rate"] = audit_results.compliance_percentage
    state["consistency"]["violations"] = audit_results.violations

    # Check automation
    state["automation"]["validation_tool_exists"] = file_exists("./validate-api")
    state["automation"]["pre_commit_hook_exists"] = file_exists(".git/hooks/pre-commit")

    # Check documentation
    state["documentation"]["low_usage_tools"] = identify_low_usage_tools()
    state["documentation"]["conventions_explained"] = check_convention_docs()

    return state
```

### Phase 2: Prioritization

**Goal**: Determine execution order based on state and goals

```python
def prioritize_phases(state, goal, constraints):
    phases = []

    # Phase 2.1: Consistency First (Foundation)
    if state["consistency"]["compliance_rate"] < 0.50:
        # Critical: Low compliance, must fix first
        phases.append({
            "phase": "consistency_critical",
            "agents": [A₃, A₁, A₂],  # Audit → Categorize → Refactor
            "priority": "P0",
            "rationale": "Compliance <50%, critical to fix"
        })
    elif state["consistency"]["compliance_rate"] < 1.0:
        # Non-compliant tools exist, but not critical
        phases.append({
            "phase": "consistency_improvement",
            "agents": [A₃, A₁, A₂],
            "priority": "P1",
            "rationale": "Improve compliance to 100%"
        })

    # Phase 2.2: Automation (Scale)
    if not state["automation"]["validation_tool_exists"]:
        phases.append({
            "phase": "build_validation",
            "agents": [A₄],
            "priority": "P1",
            "rationale": "Automate consistency checking"
        })

    if not state["automation"]["pre_commit_hook_exists"]:
        phases.append({
            "phase": "install_quality_gates",
            "agents": [A₅],
            "priority": "P1",
            "rationale": "Prevent future violations"
        })

    # Phase 2.3: Documentation (Adoption)
    if len(state["documentation"]["low_usage_tools"]) > 0:
        phases.append({
            "phase": "enhance_documentation",
            "agents": [A₆],
            "priority": "P2",
            "rationale": "Improve adoption of low-usage tools"
        })

    # Sort by priority
    priority_order = {"P0": 3, "P1": 2, "P2": 1, "P3": 0}
    phases.sort(key=lambda x: priority_order[x["priority"]], reverse=True)

    return phases
```

### Phase 3: Execution

**Goal**: Execute phases in order, testing after each

```python
def execute_phases(phases, constraints):
    results = []

    for phase in phases:
        print(f"Executing Phase: {phase['phase']}")
        print(f"Priority: {phase['priority']}")
        print(f"Rationale: {phase['rationale']}")

        # Execute agents in phase
        phase_result = execute_phase(phase, constraints)

        # Test after phase
        if constraints.test_after_each:
            test_result = run_full_test_suite()
            if not test_result.passed:
                return {
                    "status": "BLOCKED",
                    "phase": phase["phase"],
                    "reason": "Tests failed",
                    "details": test_result.failures
                }

        # Record result
        results.append({
            "phase": phase["phase"],
            "agents": phase["agents"],
            "result": phase_result,
            "tests_passed": test_result.passed if constraints.test_after_each else None
        })

        # Check convergence
        if check_convergence(results, goal):
            return {
                "status": "CONVERGED",
                "phases_completed": len(results),
                "total_phases": len(phases),
                "results": results
            }

    return {
        "status": "COMPLETED",
        "phases_completed": len(results),
        "results": results
    }
```

**Phase Execution**:
```python
def execute_phase(phase, constraints):
    if phase["phase"] == "consistency_critical":
        # A₃ (Audit) → A₁ (Categorize) → A₂ (Refactor)
        audit_results = A₃.execute()
        non_compliant = audit_results.needs_change

        for tool in non_compliant:
            # A₁: Categorize parameters
            categorization = A₁.execute({"tool": tool})

            # A₂: Refactor schema
            refactor_result = A₂.execute({
                "tool": tool,
                "categorization": categorization
            })

            # Test after each tool
            if constraints.test_after_each:
                run_tests()

    elif phase["phase"] == "build_validation":
        # A₄: Build validation tool
        validation_tool = A₄.execute({
            "conventions": load_conventions(),
            "validators": ["naming", "ordering", "description"]
        })

        # Test validation tool
        test_validation_tool(validation_tool)

    elif phase["phase"] == "install_quality_gates":
        # A₅: Install pre-commit hook
        hook_result = A₅.execute({
            "validation_command": "./validate-api --fast cmd/mcp-server/tools.go",
            "trigger_files": ["cmd/mcp-server/tools.go"]
        })

        # Test hook
        test_hook(hook_result)

    elif phase["phase"] == "enhance_documentation":
        # A₆: Enhance documentation
        low_usage_tools = identify_low_usage_tools()

        doc_result = A₆.execute({
            "tools": low_usage_tools,
            "examples_per_tool": 3,
            "add_troubleshooting": True
        })

    return phase_result
```

### Phase 4: Verification

**Goal**: Confirm all changes are safe and effective

```python
def verify_changes(results):
    verification = {
        "compliance": {
            "before": results[0]["state"]["compliance_rate"],
            "after": audit_current_compliance(),
            "improvement": 0.0
        },
        "automation": {
            "validation_tool": file_exists("./validate-api"),
            "pre_commit_hook": file_exists(".git/hooks/pre-commit"),
            "ci_integration": check_ci_config()
        },
        "tests": {
            "all_passed": True,
            "failures": []
        },
        "backward_compatibility": {
            "breaking_changes": 0,
            "safe": True
        }
    }

    # Calculate improvement
    verification["compliance"]["improvement"] = (
        verification["compliance"]["after"] -
        verification["compliance"]["before"]
    )

    # Run full test suite
    test_result = run_full_test_suite()
    verification["tests"]["all_passed"] = test_result.passed
    verification["tests"]["failures"] = test_result.failures

    # Check backward compatibility
    compat_check = check_backward_compatibility()
    verification["backward_compatibility"]["breaking_changes"] = compat_check.breaking_count
    verification["backward_compatibility"]["safe"] = compat_check.breaking_count == 0

    return verification
```

### Phase 5: Documentation

**Goal**: Document all changes and results

```python
def generate_report(results, verification):
    report = f"""
# API Design Orchestration Report

**Date**: {timestamp()}
**Goal**: {goal}
**Status**: {results['status']}

---

## Phases Executed

{format_phases(results['results'])}

---

## Verification

### Compliance
- Before: {verification['compliance']['before']:.1%}
- After: {verification['compliance']['after']:.1%}
- Improvement: +{verification['compliance']['improvement']:.1%}

### Automation
- Validation Tool: {'✅' if verification['automation']['validation_tool'] else '❌'}
- Pre-Commit Hook: {'✅' if verification['automation']['pre_commit_hook'] else '❌'}
- CI Integration: {'✅' if verification['automation']['ci_integration'] else '❌'}

### Tests
- All Passed: {'✅' if verification['tests']['all_passed'] else '❌'}
- Failures: {len(verification['tests']['failures'])}

### Backward Compatibility
- Breaking Changes: {verification['backward_compatibility']['breaking_changes']}
- Safe: {'✅' if verification['backward_compatibility']['safe'] else '❌'}

---

## Agent Executions

{format_agent_results(results['results'])}

---

## Next Steps

{generate_next_steps(verification)}
"""

    return report
```

## Convergence Criteria

```python
def check_convergence(results, goal):
    state = assess_current_state()

    if goal == "consistency":
        # Converged if 100% compliance
        return state["consistency"]["compliance_rate"] >= 1.0

    elif goal == "automation":
        # Converged if validation tool + hook installed
        return (
            state["automation"]["validation_tool_exists"] and
            state["automation"]["pre_commit_hook_exists"]
        )

    elif goal == "documentation":
        # Converged if all low-usage tools enhanced
        return len(state["documentation"]["low_usage_tools"]) == 0

    elif goal == "complete":
        # Converged if all sub-goals met
        return (
            state["consistency"]["compliance_rate"] >= 1.0 and
            state["automation"]["validation_tool_exists"] and
            state["automation"]["pre_commit_hook_exists"] and
            len(state["documentation"]["low_usage_tools"]) == 0
        )

    return False
```

## Re-Assessment Logic

```python
def reassess_and_adapt(results, constraints):
    # Re-assess state after each phase
    current_state = assess_current_state()

    # Check if goal changed
    if current_state["consistency"]["compliance_rate"] >= 1.0:
        # Consistency achieved, shift focus to automation
        print("✓ Consistency achieved (100% compliance)")
        print("→ Shifting focus to automation")
        return "automation"

    # Check if blocked
    if has_blocking_issues(results):
        print("✗ Blocking issues detected")
        print("→ Stopping execution")
        return "STOP"

    # Check if converged
    if check_convergence(results, goal):
        print("✓ Converged (goal achieved)")
        return "CONVERGED"

    # Continue
    return "CONTINUE"
```

## Workflow Execution Example

**Scenario**: Full API design improvement (consistency + automation + docs)

### Step 1: Assessment

```yaml
initial_state:
  compliance_rate: 0.675 (67.5%)
  validation_tool: false
  pre_commit_hook: false
  low_usage_tools: 3

tools_audited: 8
  compliant: 3
  needs_change: 5
```

### Step 2: Prioritization

```yaml
phases:
  - phase: "consistency_improvement"
    agents: [A₃, A₁, A₂]
    priority: "P1"

  - phase: "build_validation"
    agents: [A₄]
    priority: "P1"

  - phase: "install_quality_gates"
    agents: [A₅]
    priority: "P1"

  - phase: "enhance_documentation"
    agents: [A₆]
    priority: "P2"
```

### Step 3: Execution

**Phase 1: Consistency Improvement**
```
1. A₃ (Audit): Identify 5 non-compliant tools
2. For each tool:
   - A₁ (Categorize): Apply tier-based categorization
   - A₂ (Refactor): Reorder parameters, add comments
   - Test: ✅ All pass
3. Result: 100% compliance achieved
```

**Phase 2: Build Validation**
```
1. A₄ (Build Validator):
   - Create naming validator
   - Create ordering validator
   - Create description validator
2. Test validators: ✅ All pass
3. Result: Validation tool created
```

**Phase 3: Install Quality Gates**
```
1. A₅ (Install Hook):
   - Create pre-commit hook
   - Install automatically
   - Test hook behavior
2. Test hook: ✅ Detects, allows, blocks correctly
3. Result: Quality gate installed
```

**Phase 4: Enhance Documentation**
```
1. A₆ (Enhance Docs):
   - query_context: Add 3 practical examples
   - cleanup_temp_files: Add 3 use cases
   - query_tools_advanced: Add 5 examples + SQL reference
2. Test examples: ✅ All work
3. Result: Documentation enhanced
```

### Step 4: Verification

```yaml
compliance:
  before: 67.5%
  after: 100%
  improvement: +32.5 percentage points

automation:
  validation_tool: ✅ Created
  pre_commit_hook: ✅ Installed
  ci_integration: ✅ Configured

tests:
  all_passed: ✅ 100%
  failures: 0

backward_compatibility:
  breaking_changes: 0
  safe: ✅ Yes
```

### Step 5: Documentation

```markdown
# API Design Orchestration Report

**Status**: ✅ COMPLETE
**Phases**: 4/4 completed
**Compliance**: 67.5% → 100%
**Automation**: Validation tool + pre-commit hook installed
**Documentation**: 3 tools enhanced with 11 examples

## Agent Executions

### A₃ (Audit):
- Audited 8 tools
- Found 5 non-compliant

### A₁ (Categorize):
- Categorized 60 parameters
- 100% determinism

### A₂ (Refactor):
- Refactored 5 tools
- 60 lines changed
- 0 breaking changes

### A₄ (Build Validator):
- Created 3 validators
- 600 lines of code
- 100% test coverage (naming)

### A₅ (Install Hook):
- Installed pre-commit hook
- 60 lines of hook code
- Tested 4 scenarios (all pass)

### A₆ (Enhance Docs):
- Enhanced 3 tools
- Added 11 practical examples
- All examples tested
```

## Output Schema

```yaml
orchestration_report:
  status: "CONVERGED" | "COMPLETED" | "BLOCKED"
  phases_executed: number
  total_phases: number

  initial_state:
    compliance_rate: number
    automation_level: string
    documentation_quality: string

  final_state:
    compliance_rate: number
    automation_level: string
    documentation_quality: string

  improvements:
    compliance: number  # Percentage points
    automation: [string]  # Tools created
    documentation: number  # Tools enhanced

  agent_executions:
    - agent: string
      phase: string
      result: object
      tests_passed: boolean

  verification:
    all_tests_passed: boolean
    backward_compatible: boolean
    breaking_changes: number

  time_spent: number  # hours
```

## Success Criteria

- ✅ 100% API compliance achieved
- ✅ Automated validation tool created
- ✅ Pre-commit hook installed
- ✅ Low-usage tools documented
- ✅ All tests passing
- ✅ 0 breaking changes
- ✅ CI/CD integration ready

## Evidence from Bootstrap-006

**Source**: Iterations 4-6

**Orchestration**:
- Initial compliance: 67.5%
- Final compliance: 100%
- Phases executed: 4
- Agents coordinated: 6

**Agent Execution Order**:
1. A₃ (Audit) - Iteration 4
2. A₁ (Categorize) - Iteration 4
3. A₂ (Refactor) - Iteration 4
4. A₄ (Build Validator) - Iteration 5
5. A₅ (Install Hook) - Iteration 5
6. A₆ (Enhance Docs) - Iteration 6

**Results**:
- Compliance improvement: +32.5 percentage points
- Validation tool: ✅ Created (600 lines)
- Pre-commit hook: ✅ Installed
- Documentation: ✅ 3 tools enhanced (11 examples)
- Tests: ✅ 100% pass rate
- Breaking changes: 0

---

**Last Updated**: 2025-10-16
**Status**: Validated (Bootstrap-006 Complete)
**Reusability**: Adaptable to any API design improvement project
