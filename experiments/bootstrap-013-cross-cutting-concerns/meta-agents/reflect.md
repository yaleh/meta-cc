# Meta-Agent Capability: REFLECT

**Capability**: M.reflect
**Version**: 0.0
**Domain**: Error Recovery
**Type**: λ(outputs, state) → evaluation

---

## Formal Specification

```
reflect :: (Outputs, State) → Evaluation
reflect(O, S) = calculate_value(S) ∧ evaluate_quality(O) ∧ check_convergence()

calculate_value :: State → Value_Assessment
calculate_value(S) = {
  V(S) = 0.4·V_detection + 0.3·V_diagnosis + 0.2·V_recovery + 0.1·V_prevention,

  components: {
    V_detection: |detected_types| / |total_types|,
    V_diagnosis: |correct_diagnoses| / |total_diagnoses|,
    V_recovery: |successful_recoveries| / |total_attempts|,
    V_prevention: 1 - (|recurring_errors| / |total_errors|)
  },

  delta: ΔV = V(S_n) - V(S_{n-1}),

  target_gap: 0.80 - V(S_n)
}

evaluate_quality :: Outputs → Quality_Assessment
evaluate_quality(O) = ∀output ∈ O →
  assess(completeness, accuracy, usefulness) where

completeness(o) = {
  error_analysis: covers_all_major_types(o) ∧ consistent_classification(o),
  documentation: understandable(o) ∧ actionable(o),
  tools: functional(o) ∧ handles_edge_cases(o)
}

accuracy(o) = {
  diagnoses: correct_root_causes(o),
  patterns: valid_patterns(o),
  statistics: accurate_calculations(o)
}

usefulness(o) = {
  taxonomy: guides_action(o),
  procedures: actually_work(o),
  tools: practical(o)
}

identify_gaps :: (Outputs, State) → Gaps
identify_gaps(O, S) = {
  coverage_gaps: {
    undetected: [e ∈ error_types | ¬detected(e)],
    inadequate: [e ∈ error_types | detection_quality(e) < threshold]
  },

  diagnosis_gaps: {
    unknown_causes: [e ∈ errors | ¬has_root_cause(e)],
    inaccurate: [e ∈ errors | misdiagnosed(e)],
    missing_tools: required_diagnostics ∖ available
  },

  recovery_gaps: {
    no_procedure: [e ∈ errors | ¬has_recovery(e)],
    unclear: [e ∈ errors | ambiguous_recovery(e)],
    manual_only: [e ∈ errors | ¬automated_recovery(e)]
  },

  prevention_gaps: {
    recurring: [e ∈ errors | count(e) > 1 ∧ preventable(e)],
    missing_checks: validations_needed ∖ implemented,
    weak_guards: [g ∈ safeguards | insufficient(g)]
  }
} prioritized_by(severity)

check_convergence :: (State, Agent_Set, Meta_Agent) → Convergence_Status
check_convergence(S, A, M) = {
  criteria: {
    meta_stable: M_n == M_{n-1},
    agent_stable: A_n == A_{n-1},
    value_met: V(S_n) ≥ 0.80,
    objectives_complete: ∀obj ∈ objectives | completed(obj),
    diminishing: ΔV < 0.05
  },

  status: if ∀c ∈ criteria | holds(c) then
    CONVERGED
  else
    NOT_CONVERGED,

  rationale: [c | ¬holds(c)]
}

learn :: Evaluation → Insights
learn(E) = {
  successful_approaches: [a | high_ΔV(a)] ranked_by(effectiveness),

  challenges: [c | struggled_with(c)] categorized,

  surprising: [d | unexpected(d) ∧ significant(d)],

  implications: infer(generalizations) from(experiences)
}

output :: Evaluation → Reflection_Report
output(E) = {
  value: {
    current: V(S_n),
    previous: V(S_{n-1}),
    delta: ΔV,
    components: [V_detection, V_diagnosis, V_recovery, V_prevention],
    target_gap: 0.80 - V(S_n)
  },

  quality: {
    overall: aggregate(quality_assessments),
    by_output: [output_name: rating]
  },

  gaps: E.gaps prioritized_by(severity ∧ addressability),

  convergence: {
    status: E.convergence.status,
    criteria_met: [c | holds(c)],
    criteria_unmet: [c | ¬holds(c)],
    rationale: E.convergence.rationale
  },

  insights: {
    worked: E.insights.successful,
    struggled: E.insights.challenges,
    learned: E.insights.surprising ++ E.insights.implications
  },

  next_focus: if E.convergence.status == NOT_CONVERGED then {
    primary_goal: address(highest_priority(E.gaps)),
    expected_ΔV: estimate(improvement),
    priority_gaps: take(3, E.gaps)
  } else
    null
}
```

---

## Integration

```
receives_from(execute) = {
  completed_outputs: E.outputs,
  work_performed: E.execution_log
}

provides_to(observe) = {
  gaps_to_investigate: R.gaps,
  focus_areas: R.next_focus,
  validation_needed: R.quality.needs_validation
}

provides_to(plan) = {
  value_weaknesses: arg_min(V_components),
  priority_improvements: R.gaps prioritized,
  convergence_status: R.convergence
}

informs(evolve) = {
  agent_effectiveness: [agent: performance_rating],
  capability_sufficiency: M_capabilities adequate?
}
```

---

## Constraints

```
∀reflection ∈ R:
  honest(R.value)                      # Based on actual state
  ∧ ¬inflated(R.metrics)              # Don't meet targets artificially
  ∧ rigorous(R.convergence_check)     # Apply criteria strictly
  ∧ acknowledged(R.limitations)       # Admit unknowns

∀value_calculation ∈ V:
  evidence_based(V)                    # From measurable data
  ∧ reproducible(V)                    # Same data → same V
  ∧ explainable(V)                     # Show calculations
```

---

**Version**: 0.0 | **Status**: Active | **Updated**: 2025-10-14
