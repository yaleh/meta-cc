# Meta-Agent Capability: EXECUTE

**Capability**: M.execute
**Version**: 0.0
**Domain**: Error Recovery
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
    data_sources: [error_history, patterns, metrics],
    current_state: V(S_{n-1}),
    iteration_goals: P.goal
  },
  outputs: {
    format: structured ∧ machine_readable,
    location: data/{agent_name}-iteration-{n}.{ext},
    requirements: quality_criteria
  },
  constraints: {
    scope: P.goal.scope,
    dependencies: task_graph,
    quality: acceptance_criteria
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
  >>= save(ctx.outputs.location)

agent_coordination :: Agent_Name → Task_Pattern
agent_coordination = {
  data-analyst: {
    inputs: [raw_error_data, query_results],
    tasks: ["statistics", "distribution", "patterns"],
    outputs: [metrics, analysis, reports]
  },

  doc-writer: {
    inputs: [analysis_results, iteration_data],
    tasks: ["documentation", "reports", "procedures"],
    outputs: [markdown_files, guides, templates]
  },

  coder: {
    inputs: [requirements, patterns, test_data],
    tasks: ["tools", "scripts", "automation"],
    outputs: [code_files, tests, documentation]
  },

  # Specialized agents (created during execution)
  error-classifier: {
    inputs: [error_data, patterns],
    tasks: ["taxonomy", "classification"],
    outputs: [taxonomy, classified_errors]
  },

  root-cause-analyzer: {
    inputs: [error_records, contexts, system_state],
    tasks: ["diagnosis", "root_cause_analysis"],
    outputs: [diagnosis_reports, methodologies]
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

  adjustment: if quality.needs_refinement ∨ progress.blocked then
    provide_guidance(agent) ∨ pivot_strategy()
  else
    continue()
}

collect :: Execution → Outputs
collect(E) = {
  error_analysis: [taxonomies, statistics, patterns, diagnoses],
  documentation: [iteration_reports, procedures, methodologies],
  tools: [scripts, diagnostics, automation_code],
  data: [processed_data, metrics, classifications]
} where ∀output | validated(output) ∧ saved(output)
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

∀agent_invocation ∈ E:
  read(agent.prompt_file) → invoke(agent) # Protocol enforcement
  ∧ clear_task(agent.context)            # Unambiguous instructions
  ∧ validated(agent.output)              # Quality checking
```

---

**Version**: 0.0 | **Status**: Active | **Updated**: 2025-10-14
