# Meta-Agent Capability: REFLECT

**Capability**: M.reflect
**Version**: 0.0
**Domain**: API Design
**Type**: λ(outputs, state) → evaluation

---

## Formal Specification

```
reflect :: (Outputs, State) → Evaluation
reflect(O, S) = calculate_value(S) ∧ evaluate_quality(O) ∧ check_convergence()

calculate_value :: State → Value_Assessment
calculate_value(S) = {
  V(S) = 0.3·V_usability + 0.3·V_consistency + 0.2·V_completeness + 0.2·V_evolvability,

  components: {
    V_usability: 1 - (|usability_issues| / |total_usage_instances|),
    V_consistency: |consistent_patterns| / |total_patterns|,
    V_completeness: |implemented_features| / |required_features|,
    V_evolvability: has_versioning ∧ has_deprecation_policy ∧ backward_compatible
  },

  delta: ΔV = V(S_n) - V(S_{n-1}),

  target_gap: 0.80 - V(S_n)
}

evaluate_quality :: Outputs → Quality_Assessment
evaluate_quality(O) = ∀output ∈ O →
  assess(completeness, accuracy, usefulness) where

completeness(o) = {
  api_analysis: covers_all_tools(o) ∧ consistent_assessment(o),
  documentation: understandable(o) ∧ actionable(o),
  tools: functional(o) ∧ handles_edge_cases(o)
}

accuracy(o) = {
  consistency_analysis: correctly_identifies_violations(o),
  patterns: valid_api_patterns(o),
  statistics: accurate_calculations(o)
}

usefulness(o) = {
  guidelines: guides_api_design(o),
  recommendations: actually_improves_api(o),
  tools: practical(o)
}

identify_gaps :: (Outputs, State) → Gaps
identify_gaps(O, S) = {
  usability_gaps: {
    uncovered_tools: [t ∈ api_tools | ¬analyzed(t)],
    poor_usability: [t ∈ api_tools | usability_score(t) < threshold]
  },

  consistency_gaps: {
    naming_violations: [t ∈ api_tools | ¬follows_naming_convention(t)],
    parameter_inconsistencies: [t ∈ api_tools | inconsistent_parameters(t)],
    missing_standards: required_patterns ∖ documented_patterns
  },

  completeness_gaps: {
    missing_features: [f ∈ requested_features | ¬implemented(f)],
    incomplete_tools: [t ∈ api_tools | missing_key_parameters(t)],
    unhandled_cases: [c ∈ use_cases | ¬supported(c)]
  },

  evolvability_gaps: {
    no_versioning: ¬has_version_strategy,
    breaking_risks: [c ∈ changes | potentially_breaking(c)],
    missing_migration: required_migrations ∖ documented_migrations,
    weak_compatibility: [t ∈ api_tools | backward_compatibility_unclear(t)]
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
    components: [V_usability, V_consistency, V_completeness, V_evolvability],
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
