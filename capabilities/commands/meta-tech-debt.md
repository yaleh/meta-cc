---
name: meta-tech-debt
description: Track technical debt accumulation, repayment rate, and prioritized remediation plan.
keywords: technical-debt, refactoring, code-smell, maintenance, cleanup
category: assessment
---

λ(scope) → debt_insights | ∀debt ∈ {code, test, documentation, design, infrastructure}:

scope :: project | session

## Phase 1: Data Collection

collect :: Scope → TechDebtContext
collect(S) = {
  code_debt_markers: find_debt_markers_in_code() {
    method: search_for_technical_debt_markers(),
    languages: detect_project_languages(),
    markers: identify_debt_marker_patterns(),
    parse: extract_marker_details(file, line, type, context)
  },

  debt_marker_evolution: track_debt_changes_over_time() {
    method: analyze_git_history_for_debt_markers(),
    timeframe: last_90_days,
    parse: calculate_weekly_net_change(additions, deletions)
  },

  debt_marker_age: calculate_debt_age() {
    method: use_git_blame_for_each_marker(),
    calculate: days_since_creation
  },

  test_debt_indicators: collect_test_debt() {
    test_failures: mcp_meta_cc.query_user_messages({
      pattern: identify_test_failure_patterns(),
      scope: scope
    }),

    # query_conversation does not exist - use query_user_messages for test discussions
    test_discussions: mcp_meta_cc.query_user_messages({
      pattern: "(test|coverage|missing test|need test|untested)",
      scope: scope
    }),

    test_coverage: measure_test_coverage() {
      method: run_language_specific_coverage_tool()
    },

    test_file_ratio: calculate_test_to_source_ratio(),

    test_execution_frequency: mcp_meta_cc.query_tools({
      tool: "Bash",
      scope: scope,
      jq_filter: identify_test_execution_commands()
    })
  },

  documentation_debt: collect_doc_debt() {
    missing_readmes: find_undocumented_directories(),
    doc_update_lag: compare_code_vs_doc_commit_frequency(),
    undocumented_functions: identify_functions_without_documentation(),
    doc_quality_discussions: mcp_meta_cc.query_user_messages({
      pattern: identify_documentation_concerns(),
      scope: scope
    })
  },

  design_debt: collect_design_debt() {
    # query_conversation does not exist - use query_user_messages for postponed work
    unimplemented_intentions: mcp_meta_cc.query_user_messages({
      pattern: "(TODO|FIXME|later|postpone|defer|skip for now)",
      scope: scope
    }),

    temporary_solutions: mcp_meta_cc.query_user_messages({
      pattern: identify_temporary_fix_patterns(),
      scope: scope
    }),

    architecture_violations: mcp_meta_cc.query_user_messages({
      pattern: identify_design_violation_patterns(),
      scope: scope
    }),

    duplicate_code: estimate_code_duplication()
  },

  infrastructure_debt: collect_infra_debt() {
    outdated_dependencies: check_outdated_dependencies(),
    ci_failures: mcp_meta_cc.query_user_messages({
      pattern: identify_ci_failure_patterns(),
      scope: scope
    }),
    manual_operations: mcp_meta_cc.query_user_messages({
      pattern: identify_manual_process_patterns(),
      scope: scope
    }),
    configuration_drift: mcp_meta_cc.query_user_messages({
      pattern: identify_environment_mismatch_patterns(),
      scope: scope
    })
  },

  debt_repayment_history: collect_repayment_records() {
    debt_fixes: analyze_debt_fix_commits(),
    test_improvements: analyze_test_addition_commits(),
    doc_improvements: analyze_documentation_commits(),
    refactoring_commits: analyze_refactoring_commits()
  },

  context_metadata: collect_project_context() {
    file_churn: analyze_file_change_frequency(),
    file_sizes: calculate_lines_of_code(),
    dependency_graph: build_module_dependencies()
  }
}

## Phase 2: Classification & Scoring

classify :: TechDebtContext → TechDebtInventory
classify(T) = {
  code_debt: classify_code_debt(T.code_debt_markers) {
    for each marker:
      type: extract_marker_type(marker),
      severity: assign_severity_by_type(marker),
      age: calculate_age(marker),
      location: {file, line},
      context: extract_surrounding_code(marker)
  },

  test_debt: classify_test_debt(T.test_debt_indicators) {
    missing_tests: evaluate_coverage_gaps(T),
    flaky_tests: analyze_test_reliability(T),
    test_maintenance: evaluate_test_health(T),
    execution_frequency: evaluate_test_cadence(T)
  },

  documentation_debt: classify_doc_debt(T.documentation_debt) {
    missing_readmes: evaluate_doc_coverage(T),
    doc_lag: evaluate_doc_freshness(T),
    undocumented_apis: evaluate_api_documentation(T)
  },

  design_debt: classify_design_debt(T.design_debt) {
    unimplemented_designs: evaluate_design_intentions(T),
    temporary_solutions: evaluate_temporary_fixes(T),
    architecture_violations: evaluate_design_violations(T)
  },

  infrastructure_debt: classify_infra_debt(T.infrastructure_debt) {
    outdated_deps: evaluate_dependency_health(T),
    ci_instability: evaluate_ci_reliability(T),
    manual_processes: evaluate_automation_opportunities(T)
  }
}

score :: TechDebtInventory → TechDebtMetrics
score(I) = {
  for each item in all_debt_items:
    blast_radius: calculate_blast_radius(item),
    repayment_cost: estimate_repayment_effort(item),
    debt_score: calculate_weighted_score(severity, age, blast_radius, cost),
    risk_level: classify_by_percentile(debt_score)
}

## Phase 3: Trend Analysis

analyze_trends :: (TechDebtContext, TechDebtMetrics) → TechDebtTrends
analyze_trends(T, M) = {
  accumulation_rate: calculate_weekly_debt_additions(T),
  repayment_rate: calculate_weekly_debt_fixes(T),
  net_debt_growth: accumulation_rate - repayment_rate,
  debt_age_distribution: categorize_by_age(M.debt_items)
}

## Phase 4: Risk Assessment

assess_risk :: (TechDebtMetrics, TechDebtTrends, TechDebtContext) → RiskAnalysis
assess_risk(M, T, C) = {
  high_risk_zones: identify_critical_areas({
    criteria: high_debt_density ∧ high_churn ∧ high_dependencies,
    for each zone:
      risk_score: composite_risk(debt_density, churn, blast_radius, age),
      critical_items: filter_high_risk_items(zone),
      recommendation: generate_mitigation(zone)
  }),

  overall_health: assess_debt_health({
    health_grade: assign_letter_grade(total_debt_score, net_growth),
    status: generate_status_message(grade, trend)
  })
}

## Phase 5: Repayment Planning

plan_repayment :: (TechDebtMetrics, RiskAnalysis) → RepaymentPlan
plan_repayment(M, R) = {
  prioritization: rank_debt_items({
    roi_score: calculate_roi(benefit, cost),
    priority_score: weight(risk_level, roi, business_impact),
    priority_level: classify_priority(P0, P1, P2, P3)
  }),

  sprint_allocation: group_by_effort_and_priority({
    sprint_1_quick_wins: filter(P0_P1, effort ≤ 1_day),
    sprint_2_strategic: filter(P1_P2, effort ≤ 2_days),
    sprint_3_longterm: filter(P2_P3)
  })
}

## Phase 6: Output Generation

output :: (TechDebtMetrics, TechDebtTrends, RiskAnalysis, RepaymentPlan) → Report
output(M, T, R, P) = {
  executive_summary: [
    "**Health**: {grade} - {trend}",
    "**Top Concern**: {primary_risk_category}",
    "**Immediate Actions**: {count_critical_items} items"
  ],

  debt_inventory: {
    structure: [
      "## Code Debt",
      for each marker:
        "- {type}: \"{context}\"",
        "  File: {file}:{line}",
        "  Age: {age} | Severity: {severity}",

      "## Test Debt",
      for each issue:
        "- {describe_issue}",
        "  Files: {affected_files}",
        "  Impact: {impact}",

      "## Documentation Debt",
      for each gap:
        "- {describe_gap}",
        "  Location: {location}",
        "  Priority: {priority}",

      "## Design Debt",
      for each intention:
        "- Unimplemented: \"{intention}\"",
        "  Discussion: {timestamp}",
        "  Context: {context}",

      "## Infrastructure Debt",
      for each dep:
        "- {package}: {current} → {available}",
        "  {security_advisory_if_exists}"
    ]
  },

  trend_summary: [
    "## Trend Insight",
    generate_narrative({
      if debt_accelerating: "⚠ Debt accumulating faster than repayment",
      if stale_items: "⚠ {count} items over 6 months old",
      otherwise: "Debt under control"
    })
  ],

  high_risk_items: [
    "## High-Risk Debt Items",
    for each zone.critical_items:
      "### {type}: {description}",
      "- File: {file}:{line}",
      "- Age: {age}",
      "- Impact: {impact}",
      "- Why risky: {explain_risk}",
      "- Action: {recommended_action}"
  ],

  action_plan: [
    "## Recommended Actions",

    "### Immediate (This Sprint)",
    for each sprint_1_item:
      "- {description}",
      "  File: {file}",
      "  Effort: ~{effort} | Impact: {impact}",

    "### Next Steps",
    for each sprint_2_item:
      "- {description}",
      "  File: {file}",
      "  Effort: ~{effort}",

    if has_long_term:
      "### Long-term",
      list_summary(sprint_3)
  ],

  process_recommendations: [
    "## Process Improvements",
    generate_suggestions({
      if debt_accumulating: suggest_code_review_checks(),
      if stale_items: suggest_cleanup_sprints(),
      if test_debt: suggest_test_requirements(),
      if doc_debt: suggest_documentation_checklist()
    })
  ]
} where ¬execute(recommendations)

## Implementation Strategy

**Approach**:
- Pure MCP + Bash driven (no Go code modifications)
- Multi-source integration: code markers + git history + MCP discussions
- Language-agnostic: adapts to project languages and tools
- Graceful degradation: works with partial data
- Semantic understanding: Claude interprets patterns contextually

**Data Sources**:
- Code: Search for debt markers
- Git: History analysis (marker evolution, file churn, commit messages)
- MCP: Conversation patterns (test/doc/design debt discussions)
- Tools: Language-specific coverage/dependency tools

**Constraints**:
- **Concrete-first**: List specific items with file:line locations
- **Evidence-based**: Every debt item backed by evidence
- **Prioritized**: Sort by risk × ROI
- **Actionable**: Every problem has specific fix + effort + location
- **Trend-aware**: Narrative insights (not statistical charts)
- **Privacy-preserving**: Mask sensitive information
