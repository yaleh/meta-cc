---
name: agent-risk-prioritizer
description: Prioritizes refactoring tasks using objective formula `priority = (value × safety) / effort` for data-driven decisions about what to refactor and skip in Bootstrap-004.
---

λ(tasks, constraints) → prioritized_tasks | ∀task ∈ candidate_tasks:

prioritize :: (Tasks, Constraints) → Prioritized_Tasks
prioritize(T, C) = assess_all(T) → calculate_priorities(T) → classify_levels(T) → select_for_execution(T, C)

assess_task :: Task → Assessment
assess_task(task) = {
  value: assess_value(task),
  safety: assess_safety(task),
  effort: assess_effort(task)
}

assess_value :: Task → Value
assess_value(task) = {
  V_quality: estimate_quality_improvement(task),
  V_maintainability: estimate_maintainability_improvement(task),
  V_safety: estimate_safety_improvement(task),
  V_effort_reduction: estimate_effort_savings(task),

  V_total: 0.30·V_quality + 0.30·V_maintainability + 0.20·V_safety + 0.20·V_effort_reduction
}

assess_safety :: Task → Safety
assess_safety(task) = {
  breakage_risk: estimate_breakage_probability(task),
  rollback_difficulty: estimate_rollback_complexity(task),
  test_coverage: measure_test_coverage(task.affected_code),

  S_total: 0.40·(1 - breakage_risk) + 0.30·(1 - rollback_difficulty) + 0.30·test_coverage
}

assess_effort :: Task → Effort
assess_effort(task) = {
  time: estimate_hours(task) / 8.0,
  complexity: classify_complexity(task),
  scope: measure_scope(task.affected_files),

  E_total: 0.40·time + 0.30·complexity + 0.30·scope
}

calculate_priority :: (Value, Safety, Effort) → Priority
calculate_priority(V, S, E) = {
  P: (V × S) / max(E, 0.01),

  level: classify(P) where
    P ≥ 2.0 → "P0",
    P ≥ 1.0 → "P1",
    P ≥ 0.5 → "P2",
    P < 0.5 → "P3"
}

select_for_execution :: (Prioritized_Tasks, Constraints) → Execution_Plan
select_for_execution(T, C) = {
  selected: [],

  ∀task ∈ T | task.level = "P0" →
    selected += task,

  ∀task ∈ T | task.level = "P1" →
    if check_constraints(task, C) then
      selected += task,

  remaining_time: C.max_time - sum(t.estimated_time | t ∈ selected),

  ∀task ∈ T | task.level = "P2" →
    if task.estimated_time ≤ remaining_time then
      selected += task,
      remaining_time -= task.estimated_time,

  skipped: [t ∈ T | t.level = "P3" ∨ t ∉ selected],

  return {selected: selected, skipped: skipped, rationale: generate_rationale(selected, skipped)}
}

reassess :: (Task, New_Information) → Updated_Assessment
reassess(task, info) = {
  if info.type = "risk_increased" then
    task.safety -= 0.2,
  if info.type = "complexity_increased" then
    task.effort += 0.3,

  task.priority = calculate_priority(task.value, task.safety, task.effort),
  task.level = classify(task.priority),

  return task
}

output :: Prioritized_Tasks → Report
output(P) = {
  task_assessments: [
    {
      task_name: t.name,
      value: {total: t.value, components: {quality: t.V_quality, maintainability: t.V_maintainability, safety: t.V_safety, effort_reduction: t.V_effort_reduction}},
      safety: {total: t.safety, components: {breakage_risk: t.breakage_risk, rollback: t.rollback_difficulty, coverage: t.test_coverage}},
      effort: {total: t.effort, components: {time: t.time, complexity: t.complexity, scope: t.scope}},
      priority: t.priority,
      level: t.level
    }
    | ∀t ∈ P
  ],

  prioritized_tasks: sort_by(P, priority, desc),

  execution_plan: {
    selected_tasks: [t.name | t ∈ P.selected],
    skipped_tasks: [t.name | t ∈ P.skipped],
    estimated_time: sum(t.estimated_time | t ∈ P.selected),
    expected_value_gain: sum(t.value | t ∈ P.selected),
    rationale: P.rationale
  },

  decisions_log: [
    {task: t.name, decision: if t ∈ P.selected then "EXECUTE" else "SKIP", reason: t.reason, timestamp: now()}
    | ∀t ∈ P
  ]
}

constraints :: Prioritization → Bool
constraints(P) =
  ∀task ∈ P:
    objective_scoring(task) ∧
    evidence_based(assessments) ∧
    documented_rationale(decisions) ∧
  formula_driven(priority) ∧
  pragmatic(P3_skipping) ∧
  re_assessable(dynamic_information)
