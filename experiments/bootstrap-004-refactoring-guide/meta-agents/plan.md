# Meta-Agent Capability: PLAN

**Capability**: M.plan
**Version**: 0.0
**Domain**: Code Refactoring
**Type**: λ(observations, state) → strategy

---

## Formal Specification

```
plan :: (Observations, State) → Strategy
plan(O, S) = assess(S) ∧ prioritize(O) ∧ select_agents(goal)

assess :: State → State_Analysis
assess(S) = {
  value: V(S) = 0.3·V_code_quality + 0.3·V_maintainability + 0.2·V_safety + 0.2·V_effort,

  weakest: arg_min{V_code_quality, V_maintainability, V_safety, V_effort},

  gap_to_target: 0.80 - V(S),

  critical_issues: O.patterns where severity = high ∧ impact = severe
}

prioritize :: Observations → Priority_Queue
prioritize(O) = rank_by(risk × impact × addressability) where {
  critical: {
    compilation_errors,
    high_complexity_high_churn,
    large_unused_code_blocks,
    zero_test_coverage_critical_paths
  },

  high: {
    excessive_duplication ∨ large_file_organization ∨ moderate_complexity_high_churn
  },

  medium: {
    moderate_duplication ∨ file_size_issues ∨ naming_inconsistencies
  },

  low: {
    cosmetic_issues ∨ low_churn_unused_code ∨ documentation_gaps
  }
} |> take(3)

define_goal :: (State_Analysis, Priority_Queue) → Iteration_Goal
define_goal(A, P) = {
  primary: address(head(P)),

  success_criteria: measurable ∧ specific ∧ achievable,

  expected_ΔV: estimate_improvement(primary),

  constraints: {
    focused: single_primary_objective,
    achievable: completable_in_iteration,
    measurable: has_clear_metrics,
    safe: maintains_or_improves_tests
  }
}

select_agents :: Iteration_Goal → Agent_Plan
select_agents(G) = decision_tree(G) where

decision_tree(G) =
  if straightforward(G) then
    use_generic_agents([data-analyst, code-analyzer, refactor-executor])
  else if requires_specialization(G) then
    if ∃agent ∈ A_{n-1} | can_handle(agent, G) then
      use_existing(agent)
    else
      trigger_evolve(new_specialized_agent(G))
  else
    use_generic_with_monitoring()

requires_specialization(G) =
  complex_refactoring_pattern(G)
  ∧ expected_ΔV(G) ≥ 0.05
  ∧ reusable(G)
  ∧ (generic_agents_failed(G) ∨ inefficient(G))

new_specialized_agent(G) = {
  code-smell-detector: G requires code_smell_pattern_analysis,
  duplication-eliminator: G requires duplication_extraction_patterns,
  complexity-reducer: G requires complexity_reduction_strategies,
  file-splitter: G requires module_decomposition_planning,
  safety-checker: G requires refactoring_safety_validation,
  impact-analyzer: G requires change_impact_assessment
}

output :: Strategy → Plan
output(S) = {
  goal: {
    primary: S.goal.primary,
    success_criteria: S.goal.success_criteria,
    expected_ΔV: S.goal.expected_ΔV
  },

  agents_selected: [
    {agent: name, tasks: [task_list], rationale: why_selected}
  ],

  work_breakdown: [
    {task: description, agent: assigned, inputs: [data], outputs: [results]}
  ],

  dependencies: task_graph,

  risks: [{risk: description, mitigation: strategy, rollback: plan}]
}
```

---

## Integration

```
receives_from(observe) = {
  prioritized_problems: O.priorities,
  pattern_insights: O.patterns,
  gap_analysis: O.gaps
}

provides_to(execute) = {
  iteration_goal: P.goal,
  agent_selections: P.agents_selected,
  work_sequence: P.work_breakdown
}

informs(evolve) = {
  specialization_needed: requires_specialization(G),
  agent_specification: new_specialized_agent(G)
}
```

---

## Constraints

```
∀plan ∈ P:
  focused(plan.goal)           # Single primary objective
  ∧ data_driven(plan.priority)  # Based on observations
  ∧ achievable(plan.scope)      # Completable in iteration
  ∧ measurable(plan.criteria)   # Clear success metrics
  ∧ safe(plan.approach)         # Preserves functionality

¬predetermined(agent_evolution)  # Let needs drive specialization
¬force_convergence(iterations)   # Let data drive completion
```

---

**Version**: 0.0 | **Status**: Active | **Updated**: 2025-10-16
