# Meta-Agent Capability: EXECUTE

**Capability**: M.execute
**Version**: 0.0
**Domain**: Code Refactoring
**Type**: λ(plan, agents) → outputs

---

## Formal Specification

```
execute :: (Plan, Agent_Set) → Outputs
execute(P, A) = prepare(context) ∧ coordinate(agents) ∧ collect(results)

prepare :: (Plan, Agent_Set) → Execution_Context
prepare(P, A) = ∀agent ∈ P.agents_selected →
  read(agent.prompt_file) ∧ build_context(agent, P) where

build_context(agent, P) = {
  task: specific ∧ actionable,
  inputs: {
    code_metrics: [complexity, duplication, coverage],
    target_files: [files_to_refactor],
    current_state: V(S_{n-1}),
    iteration_goals: P.goal,
    safety_constraints: [test_requirements, backward_compatibility]
  },
  outputs: {
    format: structured ∧ machine_readable,
    location: data/{agent_name}-iteration-{n}.{ext},
    requirements: quality_criteria
  },
  constraints: {
    scope: P.goal.scope,
    dependencies: task_graph,
    quality: acceptance_criteria,
    safety: must_pass_tests ∧ maintain_coverage
  }
}

coordinate :: (Execution_Context, Agent_Set) → Execution
coordinate(C, A) = pattern_match(C.dependencies) where

sequential(tasks) =
  fold_left(λ(state, task) → execute_agent(task) >>= pass_output, initial, tasks)

parallel(tasks) =
  map(execute_agent, tasks) |> collect_all

iterative(task) =
  fix(λrecurse → λstate →
    if meets_criteria(state) then
      state
    else
      execute_agent(task, state) >>= recurse)

execute_agent :: (Agent, Context) → Result
execute_agent(agent, ctx) =
  read(agents/{agent.name}.md)
  >>= invoke(agent, ctx.task, ctx.inputs)
  >>= validate(ctx.outputs.requirements)
  >>= run_safety_checks()
  >>= save(ctx.outputs.location)

agent_coordination :: Agent_Name → Task_Pattern
agent_coordination = {
  code-analyzer: {
    inputs: [code_metrics, target_files, patterns],
    tasks: ["detect_smells", "identify_refactoring_opportunities", "prioritize_changes"],
    outputs: [analysis_report, refactoring_candidates, priority_list]
  },

  refactor-executor: {
    inputs: [refactoring_plan, target_files, test_suite],
    tasks: ["apply_refactorings", "run_tests", "verify_behavior"],
    outputs: [refactored_code, test_results, diff_summary]
  },

  doc-writer: {
    inputs: [refactoring_results, metrics, decisions],
    tasks: ["document_changes", "update_methodology", "write_iteration_report"],
    outputs: [iteration_docs, methodology_updates, decision_log]
  },

  # Specialized agents (created during execution if needed)
  code-smell-detector: {
    inputs: [code_structure, patterns, metrics],
    tasks: ["identify_smells", "classify_severity", "recommend_fixes"],
    outputs: [smell_catalog, severity_assessment, fix_recommendations]
  },

  duplication-eliminator: {
    inputs: [clone_groups, context_analysis],
    tasks: ["extract_common_logic", "create_abstractions", "refactor_call_sites"],
    outputs: [extracted_functions, refactored_code, duplication_metrics]
  },

  complexity-reducer: {
    inputs: [complex_functions, control_flow_graphs],
    tasks: ["decompose_functions", "extract_methods", "simplify_logic"],
    outputs: [simplified_code, complexity_metrics, test_updates]
  }
}

monitor :: Execution → Status
monitor(E) = {
  progress: {
    completed: [task | done(task)],
    in_progress: [task | running(task)],
    blocked: [task | has_blocker(task)]
  },

  quality: {
    meets_requirements: ∀output | satisfies(output, requirements),
    has_gaps: ∃output | incomplete(output),
    needs_refinement: ∃output | quality_below_threshold(output)
  },

  safety: {
    tests_passing: all_tests_pass(),
    coverage_maintained: coverage(S_n) ≥ coverage(S_{n-1}),
    no_regressions: behavior_preserved()
  },

  adjustment: if quality.needs_refinement ∨ progress.blocked ∨ ¬safety then
    provide_guidance(agent) ∨ pivot_strategy() ∨ rollback()
  else
    continue()
}

collect :: Execution → Outputs
collect(E) = {
  refactoring_analysis: [code_smell_reports, duplication_analysis, complexity_metrics],
  refactored_code: [file_changes, diffs, commit_logs],
  documentation: [iteration_reports, methodology_updates, decision_rationale],
  data: [metrics_before_after, improvement_measurements, safety_validation]
} where ∀output | validated(output) ∧ saved(output) ∧ tests_pass(output)
```

---

## Integration

```
receives_from(plan) = {
  iteration_goal: P.goal,
  agent_selections: P.agents_selected,
  work_sequence: P.work_breakdown
}

provides_to(reflect) = {
  completed_outputs: E.outputs,
  work_performed: E.execution_log,
  observations: E.monitor.insights
}

coordinates_with(evolve) = {
  invokes: newly_created_agents,
  tests: agent_effectiveness
}
```

---

## Constraints

```
∀execution ∈ E:
  follows_plan(execution)               # Don't deviate without re-planning
  ∧ reads_prompts(execution.agents)    # Always read agent files
  ∧ systematic(execution.collection)   # Collect all outputs
  ∧ documented(execution.decisions)    # Track execution choices
  ∧ safe(execution.changes)            # All tests must pass

∀agent_invocation ∈ E:
  read(agent.prompt_file) → invoke(agent) # Protocol enforcement
  ∧ clear_task(agent.context)            # Unambiguous instructions
  ∧ validated(agent.output)              # Quality checking
  ∧ tests_pass(agent.changes)            # Safety validation
```

---

**Version**: 0.0 | **Status**: Active | **Updated**: 2025-10-16
