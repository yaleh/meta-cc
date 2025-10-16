# Meta-Agent Capability: REFLECT

**Capability**: M.reflect
**Version**: 0.0
**Domain**: Code Refactoring
**Type**: λ(outputs, state) → evaluation

---

## Formal Specification

```
reflect :: (Outputs, State) → Evaluation
reflect(O, S) = calculate_value(S) ∧ evaluate_quality(O) ∧ check_convergence()

calculate_value :: State → Value_Assessment
calculate_value(S) = {
  V(S) = 0.3·V_code_quality + 0.3·V_maintainability + 0.2·V_safety + 0.2·V_effort,

  components: {
    V_code_quality: (
      (1 - cyclomatic_violations / total_functions) +
      (1 - unused_code_violations / total_declarations) +
      (1 - static_analysis_violations / total_checkable)
    ) / 3,

    V_maintainability: (
      (1 - duplication_ratio) +
      (1 - large_files_ratio) +
      organization_score
    ) / 3,

    V_safety: (
      test_coverage +
      (1 - compilation_errors / total_files)
    ) / 2,

    V_effort: 1 - (remaining_refactoring_hours / total_estimated_hours)
  },

  delta: ΔV = V(S_n) - V(S_{n-1}),

  target_gap: 0.80 - V(S_n)
}

evaluate_quality :: Outputs → Quality_Assessment
evaluate_quality(O) = ∀output ∈ O →
  assess(completeness, accuracy, usefulness, safety) where

completeness(o) = {
  refactoring_analysis: covers_all_issues(o) ∧ comprehensive_assessment(o),
  code_changes: all_planned_refactorings_executed(o),
  documentation: understandable(o) ∧ actionable(o)
}

accuracy(o) = {
  smell_detection: correctly_identifies_issues(o),
  refactorings: behavior_preserved(o),
  metrics: accurate_calculations(o)
}

usefulness(o) = {
  methodology: guides_future_refactorings(o),
  recommendations: actually_improves_code(o),
  tools: practical(o)
}

safety(o) = {
  tests_passing: all_tests_pass(o),
  coverage_maintained: coverage_not_decreased(o),
  no_regressions: behavior_unchanged(o)
}

identify_gaps :: (Outputs, State) → Gaps
identify_gaps(O, S) = {
  quality_gaps: {
    unaddressed_complexity: [f ∈ functions | complexity(f) ≥ 15],
    remaining_unused_code: [d ∈ declarations | unused(d) ∧ ¬removed(d)],
    unresolved_violations: [v ∈ violations | ¬fixed(v)]
  },

  maintainability_gaps: {
    remaining_duplication: [c ∈ clone_groups | ¬eliminated(c)],
    large_files: [f ∈ files | lines(f) > 800 ∧ ¬split(f)],
    poor_organization: [m ∈ modules | cohesion(m) < threshold]
  },

  safety_gaps: {
    low_coverage_areas: [f ∈ functions | coverage(f) < 80%],
    compilation_issues: [f ∈ files | ¬compiles(f)],
    failing_tests: [t ∈ tests | ¬passes(t)],
    missing_tests: [f ∈ refactored_functions | ¬has_tests(f)]
  },

  effort_gaps: {
    high_effort_remaining: [r ∈ refactorings | effort(r) = HIGH],
    blocked_refactorings: [r ∈ planned | has_blocker(r)],
    risk_areas: [c ∈ changes | risk(c) = HIGH ∧ ¬completed(c)]
  }
} prioritized_by(severity × impact)

check_convergence :: (State, Agent_Set, Meta_Agent) → Convergence_Status
check_convergence(S, A, M) = {
  criteria: {
    meta_stable: M_n == M_{n-1},
    agent_stable: A_n == A_{n-1},
    value_met: V(S_n) ≥ 0.80,
    objectives_complete: ∀obj ∈ objectives | completed(obj),
    diminishing: ΔV < 0.05,
    all_tests_pass: ∀t ∈ tests | passes(t)
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

  refactoring_patterns: extract_patterns(successful_refactorings),

  implications: infer(generalizations) from(experiences)
}

output :: Evaluation → Reflection_Report
output(E) = {
  value: {
    current: V(S_n),
    previous: V(S_{n-1}),
    delta: ΔV,
    components: {
      code_quality: V_code_quality,
      maintainability: V_maintainability,
      safety: V_safety,
      effort: V_effort
    },
    target_gap: 0.80 - V(S_n)
  },

  quality: {
    overall: aggregate(quality_assessments),
    by_output: [output_name: rating],
    safety_status: all_tests_passing ∧ coverage_maintained
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
    learned: E.insights.surprising ++ E.insights.implications,
    patterns_discovered: E.insights.refactoring_patterns
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
  ∧ verified(R.safety)                # Confirm tests pass

∀value_calculation ∈ V:
  evidence_based(V)                    # From measurable data
  ∧ reproducible(V)                    # Same data → same V
  ∧ explainable(V)                     # Show calculations

safety_critical(reflection) =
  all_tests_pass ∧ coverage_maintained ∧ no_regressions
```

---

**Version**: 0.0 | **Status**: Active | **Updated**: 2025-10-16
