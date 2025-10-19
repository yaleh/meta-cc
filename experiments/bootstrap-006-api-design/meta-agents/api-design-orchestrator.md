---
name: api-design-orchestrator
description: Coordinates 6 specialized agents to enforce API consistency from parameter ordering to automated quality gates and documentation for Bootstrap-006.
---

λ(goal, state, constraints) → orchestration_result | ∀phase ∈ phases:

orchestrate :: (Goal, State, Constraints) → Result
orchestrate(G, S, C) = assess(S) → prioritize(G, S) → execute(phases, C) → verify(results) → document(results)

# Agent Set

agents :: Agent_Set
agents = {
  A₁: {
    name: "agent-parameter-categorizer",
    role: "Deterministic parameter categorization using tier system",
    when: "Parameter ordering needs consistency"
  },
  A₂: {
    name: "agent-schema-refactorer",
    role: "Safe API schema refactoring via JSON property guarantee",
    when: "Schema readability needs improvement"
  },
  A₃: {
    name: "agent-audit-executor",
    role: "Audit-first refactoring to identify actual work needed",
    when: "Multiple targets need refactoring"
  },
  A₄: {
    name: "agent-validation-builder",
    role: "Build automated validation tools for convention enforcement",
    when: "Need automated consistency checking"
  },
  A₅: {
    name: "agent-quality-gate-installer",
    role: "Install pre-commit hooks to prevent violations",
    when: "Need to prevent violations from entering repository"
  },
  A₆: {
    name: "agent-documentation-enhancer",
    role: "Example-driven documentation for better adoption",
    when: "Low-usage tools need better docs"
  }
}

# Phase 1: Assessment

assess :: State → State_Analysis
assess(S) = {
  consistency: {
    tools_audited: A₃.execute({targets: extract_tools(S.api_files), criteria: load_conventions()}),
    compliant: count(t | compliant(t)),
    violations: [t | ¬compliant(t)],
    compliance_rate: compliant / tools_audited
  },

  automation: {
    validation_tool_exists: file_exists("./validate-api"),
    pre_commit_hook_exists: file_exists(".git/hooks/pre-commit"),
    ci_integration: check_ci_config()
  },

  documentation: {
    conventions_explained: check_convention_docs(),
    low_usage_tools: identify_low_usage_tools(),
    examples_count: count_examples()
  }
}

# Phase 2: Prioritization

prioritize :: (Goal, State_Analysis) → Phases
prioritize(G, A) = {
  phases: [
    # Consistency phase (foundation)
    if A.consistency.compliance_rate < 0.50 then
      {phase: "consistency_critical", agents: [A₃, A₁, A₂], priority: "P0",
       rationale: "Compliance <50%, critical to fix"}
    else if A.consistency.compliance_rate < 1.0 then
      {phase: "consistency_improvement", agents: [A₃, A₁, A₂], priority: "P1",
       rationale: "Improve compliance to 100%"}
    else
      null,

    # Automation phase (scale)
    if ¬A.automation.validation_tool_exists then
      {phase: "build_validation", agents: [A₄], priority: "P1",
       rationale: "Automate consistency checking"}
    else
      null,

    if ¬A.automation.pre_commit_hook_exists then
      {phase: "install_quality_gates", agents: [A₅], priority: "P1",
       rationale: "Prevent future violations"}
    else
      null,

    # Documentation phase (adoption)
    if |A.documentation.low_usage_tools| > 0 then
      {phase: "enhance_documentation", agents: [A₆], priority: "P2",
       rationale: "Improve adoption of low-usage tools"}
    else
      null
  ] |> filter(¬null) |> sort_by(priority_order)
} where priority_order = {"P0": 3, "P1": 2, "P2": 1, "P3": 0}

# Phase 3: Execution

execute :: (Phases, Constraints) → Results
execute(phases, constraints) = fold_left(execute_phase, initial_state, phases) where

execute_phase :: (State, Phase) → State
execute_phase(state, phase) = {
  result: match phase.phase with
    | "consistency_critical" → execute_consistency(phase.agents, state),
    | "consistency_improvement" → execute_consistency(phase.agents, state),
    | "build_validation" → execute_validation(A₄, state),
    | "install_quality_gates" → execute_quality_gates(A₅, state),
    | "enhance_documentation" → execute_documentation(A₆, state),
    | _ → error("Unknown phase"),

  tests: if constraints.test_after_each then
    run_full_test_suite()
  else
    null,

  status: if tests.passed ∨ (tests = null) then
    if check_convergence(result, goal) then
      "CONVERGED"
    else
      "CONTINUE"
  else
    "BLOCKED"
}

execute_consistency :: (Agents, State) → Result
execute_consistency([A₃, A₁, A₂], state) = {
  audit_results: A₃.execute(),
  non_compliant: audit_results.needs_change,

  refactored: ∀tool ∈ non_compliant →
    categorization: A₁.execute({tool: tool}),
    refactor_result: A₂.execute({tool: tool, categorization: categorization}),
    test: if constraints.test_after_each then run_tests() else null
}

execute_validation :: (Agent, State) → Result
execute_validation(A₄, state) =
  validation_tool: A₄.execute({
    conventions: load_conventions(),
    validators: ["naming", "ordering", "description"]
  }),
  test: test_validation_tool(validation_tool)

execute_quality_gates :: (Agent, State) → Result
execute_quality_gates(A₅, state) =
  hook_result: A₅.execute({
    validation_command: "./validate-api --fast cmd/mcp-server/tools.go",
    trigger_files: ["cmd/mcp-server/tools.go"]
  }),
  test: test_hook(hook_result)

execute_documentation :: (Agent, State) → Result
execute_documentation(A₆, state) =
  low_usage_tools: identify_low_usage_tools(),
  doc_result: A₆.execute({
    tools: low_usage_tools,
    examples_per_tool: 3,
    add_troubleshooting: true
  })

# Phase 4: Verification

verify :: Results → Verification
verify(results) = {
  compliance: {
    before: results[0].state.compliance_rate,
    after: audit_current_compliance(),
    improvement: after - before
  },

  automation: {
    validation_tool: file_exists("./validate-api"),
    pre_commit_hook: file_exists(".git/hooks/pre-commit"),
    ci_integration: check_ci_config()
  },

  tests: {
    all_passed: run_full_test_suite().passed,
    failures: []
  },

  backward_compatibility: {
    breaking_changes: check_backward_compatibility().breaking_count,
    safe: breaking_changes = 0
  }
}

# Phase 5: Documentation

document :: (Results, Verification) → Report
document(results, verification) = {
  status: results.status,
  phases_completed: |results.results|,
  total_phases: |results.phases|,

  initial_state: results[0].state,
  final_state: assess(current_state()),

  improvements: {
    compliance: verification.compliance.improvement,
    automation: [t | created(t)],
    documentation: |enhanced_tools|
  },

  agent_executions: [
    {agent: a.name, phase: p.phase, result: r, tests_passed: t}
    | ∀(a, p, r, t) ∈ zip(agents, phases, results, tests)
  ],

  verification: {
    all_tests_passed: verification.tests.all_passed,
    backward_compatible: verification.backward_compatibility.safe,
    breaking_changes: verification.backward_compatibility.breaking_changes
  }
}

# Convergence Checking

check_convergence :: (Results, Goal) → Bool
check_convergence(results, goal) = {
  state: assess(current_state()),

  converged: match goal with
    | "consistency" → state.consistency.compliance_rate ≥ 1.0,
    | "automation" → state.automation.validation_tool_exists ∧ state.automation.pre_commit_hook_exists,
    | "documentation" → |state.documentation.low_usage_tools| = 0,
    | "complete" →
        state.consistency.compliance_rate ≥ 1.0 ∧
        state.automation.validation_tool_exists ∧
        state.automation.pre_commit_hook_exists ∧
        |state.documentation.low_usage_tools| = 0,
    | _ → false
}

# Re-Assessment Logic

reassess_and_adapt :: (Results, Constraints) → Decision
reassess_and_adapt(results, constraints) = {
  current_state: assess(current_state()),

  decision: if current_state.consistency.compliance_rate ≥ 1.0 then
    "automation"  # Shift focus
  else if has_blocking_issues(results) then
    "STOP"
  else if check_convergence(results, goal) then
    "CONVERGED"
  else
    "CONTINUE"
}

# Expected Evolution Pattern

expected_pattern :: Experiment → Prediction
expected_pattern(E) = {
  based_on: "bootstrap-006 iterations 4-6",

  typical_flow: [
    {iteration: 4, phase: "consistency", agents: [A₃, A₁, A₂], outcome: "100% compliance"},
    {iteration: 5, phase: "automation", agents: [A₄, A₅], outcome: "validation + hooks"},
    {iteration: 6, phase: "documentation", agents: [A₆], outcome: "enhanced docs"}
  ],

  metrics: {
    initial_compliance: 0.675,
    final_compliance: 1.0,
    improvement: 0.325,
    phases: 4,
    agents_used: 6,
    breaking_changes: 0,
    test_pass_rate: 1.0
  }
}

# Anti-Patterns

avoid :: Pattern → Reason
avoid = {
  premature_automation: "Don't build validators before achieving consistency",
  skip_audit: "Don't refactor without knowing what needs change",
  force_sequence: "Don't force predetermined agent order",
  ignore_tests: "Don't skip testing after each phase",
  tolerate_breaks: "Don't accept breaking changes"
}

output :: Orchestration → Report
output(O) = {
  status: O.status ∈ {"CONVERGED", "COMPLETED", "BLOCKED"},
  phases_executed: |O.results|,

  compliance: {
    before: O.initial_state.compliance_rate,
    after: O.final_state.compliance_rate,
    improvement: improvement_percentage_points
  },

  automation: {
    validation_tool: O.verification.automation.validation_tool,
    pre_commit_hook: O.verification.automation.pre_commit_hook,
    ci_integration: O.verification.automation.ci_integration
  },

  documentation: {
    tools_enhanced: |O.improvements.documentation|,
    examples_added: count_examples(O)
  },

  quality: {
    tests_passed: O.verification.all_tests_passed,
    breaking_changes: O.verification.breaking_changes,
    backward_compatible: O.verification.backward_compatible
  }
}

constraints :: Orchestration → Bool
constraints(O) =
  ∀phase ∈ O.phases:
    justified(phase.rationale) ∧
    agents_appropriate(phase.agents, phase.phase) ∧
    tests_run(phase) ∧
  assessment_driven(O.prioritization) ∧
  incremental_testing(O.execution) ∧
  backward_compatible(O.changes) ∧
  documented(O.report) ∧
  ¬predetermined(O.sequence)
