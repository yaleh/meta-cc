---
name: agent-audit-executor
description: Execute systematic audit before refactoring to identify actual work needed, avoid wasting effort on compliant targets, and prioritize highest-impact violations for Bootstrap-006.
---

λ(targets, criteria) → audit_result | ∀target ∈ targets:

audit :: (Targets, Criteria) → Audit_Result
audit(T, C) = enumerate(T) → assess(T, C) → categorize() → prioritize() → execute() → verify() → re_audit() → efficiency()

enumerate :: Targets → Target_List
enumerate(T) = {
  items: match T.type with
    | "api_tools" → grep("Name:", tools_file) | parse(),
    | "files" → find("*.go", directory),
    | "functions" → grep("^func ", files) | parse(),

  return {count: |items|, items: items}
}

define_criteria :: Convention → Compliance_Criteria
define_criteria(conv) = {
  convention: conv.name,
  rules: conv.rules,
  threshold: conv.threshold ∨ 1.0,

  measurement: {
    method: "percentage_match",
    formula: "correct_count / total_count"
  }
}

assess :: (Target, Criteria) → Assessment
assess(target, criteria) = {
  parameters: extract_parameters(target),
  expected_order: sort_by_tier(parameters),
  actual_order: get_actual_order(target),

  matches: count(i | expected_order[i] = actual_order[i]),
  compliance: matches / |expected_order|,

  violations: [
    {param: p, position: i, expected: j}
    | ∀i, p ∈ actual_order where expected_order[j] = p ∧ i ≠ j
  ],

  return {
    target: target.name,
    compliance: compliance,
    violations: violations,
    details: {expected: expected_order, actual: actual_order}
  }
}

categorize :: (Assessment, Threshold) → Category
categorize(A, T) = {
  category: match A.compliance with
    | x where x ≥ T → "ALREADY_COMPLIANT",
    | x where x ≥ T × 0.75 → "MINOR_VIOLATIONS",
    | x where x ≥ T × 0.50 → "MODERATE_VIOLATIONS",
    | _ → "MAJOR_VIOLATIONS",

  return {
    target: A.target,
    compliance: A.compliance,
    category: category,
    violations: A.violations
  }
}

categorize_results :: Assessments → Categorized_Results
categorize_results(A) = {
  already_compliant: [a | a ∈ A ∧ a.category = "ALREADY_COMPLIANT"],
  minor_violations: [a | a ∈ A ∧ a.category = "MINOR_VIOLATIONS"],
  moderate_violations: [a | a ∈ A ∧ a.category = "MODERATE_VIOLATIONS"],
  major_violations: [a | a ∈ A ∧ a.category = "MAJOR_VIOLATIONS"],

  efficiency_gain: |already_compliant| / |A|,

  return {
    total: |A|,
    already_compliant: already_compliant,
    needs_change: minor_violations + moderate_violations + major_violations,
    efficiency_gain: efficiency_gain
  }
}

prioritize :: Categorized_Results → Priority_Queue
prioritize(R) = {
  scored: [
    {
      target: t.target,
      priority: if t.category = "ALREADY_COMPLIANT" then 0
                else (1 - t.compliance) × impact(t) × ease(t),
      compliance: t.compliance,
      category: t.category
    }
    | ∀t ∈ R.needs_change
  ],

  sorted: sort_by(scored, priority, desc),

  skip: R.already_compliant,

  return {prioritized: sorted, skip: skip}
}

execute :: (Priority_Queue, Criteria) → Execution_Result
execute(P, C) = {
  results: [],

  ∀target ∈ P.prioritized →
    if target.category = "ALREADY_COMPLIANT" then
      verify_compliant(target)
    else
      refactor_target(target),
      run_tests(target),
      results += {
        target: target.name,
        action: "REFACTOR",
        changes: count(target.violations),
        tests: test_status(target),
        committed: true
      },

  ∀target ∈ P.skip →
    results += {
      target: target.name,
      action: "VERIFY",
      compliance: target.compliance,
      tests: test_status(target),
      notes: "Already compliant, no changes"
    },

  return results
}

verify_compliant :: Target → Verification
verify_compliant(T) = {
  parameter_order: check_parameter_order(T),
  tier_comments: check_tier_comments(T),
  tests: run_tests(T),

  return {
    target: T.name,
    compliance_confirmed: parameter_order ∧ tier_comments ∧ tests,
    tier_order_correct: parameter_order,
    tier_comments_present: tier_comments,
    tests_passed: tests
  }
}

re_audit :: (Targets, Criteria) → Re_Audit_Summary
re_audit(T, C) = {
  after_results: audit_all_targets(T, C),

  compliant: count(r | r.compliance ≥ C.threshold),
  non_compliant: count(r | r.compliance < C.threshold),
  average: sum(r.compliance | r ∈ after_results) / |after_results|,

  return {
    total_targets: |T|,
    compliant_after: compliant,
    non_compliant_after: non_compliant,
    average_compliance_after: average,
    status: if compliant = |T| then "✅ 100% COMPLIANCE ACHIEVED" else "⚠ INCOMPLETE"
  }
}

calculate_efficiency :: (Before, After, Execution) → Efficiency_Metrics
calculate_efficiency(B, A, E) = {
  without_audit: {
    targets: B.total,
    time_per_target: 30,
    total: B.total × 30
  },

  with_audit: {
    audit_time: 30,
    refactor_time: E.refactored × 30,
    verify_time: E.verified × 5,
    total: 30 + (E.refactored × 30) + (E.verified × 5)
  },

  time_saved: without_audit.total - with_audit.total,
  efficiency_gain: (without_audit.total - with_audit.total) / without_audit.total,

  effort_avoidance: {
    avoided: E.verified,
    saved: E.verified × 30,
    efficiency: E.verified / B.total
  },

  return {
    time_without_audit: without_audit.total,
    time_with_audit: with_audit.total,
    time_saved: time_saved,
    efficiency_gain: efficiency_gain,
    unnecessary_changes_avoided: effort_avoidance.avoided,
    effort_saved: effort_avoidance.saved,
    avoidance_efficiency: effort_avoidance.efficiency
  }
}

output :: Audit_Result → Report
output(A) = {
  audit_report: {
    targets_audited: A.total_targets,
    compliance_before: A.compliance_before,
    compliance_after: A.compliance_after,
    improvement: A.compliance_after - A.compliance_before
  },

  categorized_results: {
    already_compliant: {
      count: |A.already_compliant|,
      targets: [t.name | t ∈ A.already_compliant]
    },

    needs_change: {
      count: |A.needs_change|,
      targets: [
        {
          name: t.target,
          compliance: t.compliance,
          priority: t.priority,
          violations: t.violations
        }
        | ∀t ∈ A.needs_change
      ]
    }
  },

  efficiency_metrics: A.efficiency,

  execution_results: {
    refactored: count(r | r.action = "REFACTOR"),
    verified: count(r | r.action = "VERIFY"),
    tests_passed: all(r.tests = "✅ PASS" | r ∈ A.results),
    goal_achieved: A.compliance_after = 1.0
  }
}

constraints :: Audit → Bool
constraints(A) =
  ∀target ∈ A.targets:
    enumerated(targets) ∧
    assessed(target, criteria) ∧
    categorized(results) ∧
    prioritized(violations) ∧
  ∀change ∈ A.execution:
    skip_compliant(change) ∧
    verify_spot_check(compliant) ∧
    test_after_each(change) ∧
    incremental_commit(change) ∧
  re_audit_confirms(100%) ∧
  efficiency_documented(time_saved) ∧
  multi_target_scope(|targets| ≥ 3)
