# Meta-Agent Capability: PLAN

**Capability**: M.plan
**Version**: 0.0
**Domain**: Error Recovery
**Type**: λ(observations, state) → strategy

---

## Formal Specification

```
plan :: (Observations, State) → Strategy
plan(O, S) = assess(S) ∧ prioritize(O) ∧ select_agents(goal)

assess :: State → State_Analysis
assess(S) = {
  value: V(S) = 0.4·V_detection + 0.3·V_diagnosis + 0.2·V_recovery + 0.1·V_prevention,

  weakest: arg_min{V_detection, V_diagnosis, V_recovery, V_prevention},

  gap_to_target: 0.80 - V(S),

  critical_errors: O.patterns where impact = blocking ∨ frequency > 10%
}

prioritize :: Observations → Priority_Queue
prioritize(O) = rank_by(urgency ∧ impact ∧ addressability) where {
  critical: blocking_errors ∨ safety_critical ∨ frequency > 10%,
  high: significant_impact ∨ recurring ∨ foundation_for_others,
  medium: moderate_frequency ∨ quality_degrading,
  low: rare_edge_cases ∨ minor_inconvenience
} |> take(3)

define_goal :: (State_Analysis, Priority_Queue) → Iteration_Goal
define_goal(A, P) = {
  primary: address(head(P)),

  success_criteria: measurable ∧ specific ∧ achievable,

  expected_ΔV: estimate_improvement(primary),

  constraints: {
    focused: single_primary_objective,
    achievable: completable_in_iteration,
    measurable: has_clear_metrics
  }
}

select_agents :: Iteration_Goal → Agent_Plan
select_agents(G) = decision_tree(G) where

decision_tree(G) =
  if straightforward(G) then
    use_generic_agents([data-analyst, doc-writer, coder])
  else if requires_specialization(G) then
    if ∃agent ∈ A_{n-1} | can_handle(agent, G) then
      use_existing(agent)
    else
      trigger_evolve(new_specialized_agent(G))
  else
    use_generic_with_monitoring()

requires_specialization(G) =
  complex_domain_knowledge(G)
  ∧ expected_ΔV(G) ≥ 0.05
  ∧ reusable(G)
  ∧ (generic_agents_failed(G) ∨ inefficient(G))

new_specialized_agent(G) = {
  error-classifier: G requires error_taxonomy,
  root-cause-analyzer: G requires systematic_diagnosis,
  recovery-advisor: G requires recovery_strategies,
  error-pattern-learner: G requires pattern_recognition
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

  risks: [{risk: description, mitigation: strategy}]
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

¬predetermined(agent_evolution)  # Let needs drive specialization
¬force_convergence(iterations)   # Let data drive completion
```

---

**Version**: 0.0 | **Status**: Active | **Updated**: 2025-10-14
