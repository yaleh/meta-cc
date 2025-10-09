---
name: meta-quality-scan
description: Quick quality assessment of recent work with scorecard and improvement recommendations.
---

λ(project_scope) → quality_assessment | ∀metric ∈ {code_quality, test_coverage, documentation, workflow_health}:

scope :: project | session

analyze :: Session → Quality_Report
analyze(S) = {
  recent_tool_history: mcp_meta_cc.query_tools(
    scope=scope,
    limit=50,
    status="success"
  ),

  error_history: mcp_meta_cc.query_tools(
    scope=scope,
    status="error",
    limit=30
  ),

  session_statistics: mcp_meta_cc.get_session_stats(
    scope=scope
  ),

  project_files: mcp_meta_cc.query_files(
    scope=scope,
    threshold=10
  )
}

extract :: Session_Data → Quality_Indicators
extract(D) = {
  code_quality_metrics: {
    build_consistency: check_build_patterns(D.recent_tool_history),
    test_execution_rate: calculate_test_frequency(D.recent_tool_history),
    error_recovery_patterns: analyze_error_recovery(D.error_history),
    verification_habits: measure_verification_frequency(D.recent_tool_history)
  },

  test_coverage_indicators: {
    test_frequency: count_test_executions(D.recent_tool_history),
    test_failure_rate: calculate_test_failure_rate(D.error_history),
    coverage_patterns: analyze_test_distribution(D.project_files),
    test_driven_adherence: measure_tdd_compliance(D.recent_tool_history)
  },

  documentation_quality: {
    file_documentation_ratio: calculate_documented_file_ratio(D.project_files),
    readme_completeness: assess_readme_quality(D.project_files),
    code_comment_density: estimate_comment_coverage(D.project_files),
    documentation_maintenance: check_doc_updates(D.recent_tool_history)
  },

  workflow_health: {
    error_rate: D.session_statistics.ErrorRate,
    tool_efficiency: calculate_tool_success_rate(D.recent_tool_history),
    session_completion: measure_task_completion_rate(D.recent_tool_history),
    iteration_patterns: analyze_refinement_cycles(D.recent_tool_history)
  }
}

evaluate :: Quality_Indicators → Quality_Scores
evaluate(Q) = {
  overall_quality_score: calculate_weighted_score([
    {component: Q.code_quality_metrics.build_consistency, weight: 0.25},
    {component: Q.test_coverage_indicators.test_frequency, weight: 0.30},
    {component: Q.documentation_quality.file_documentation_ratio, weight: 0.20},
    {component: Q.workflow_health.error_rate, weight: 0.25}
  ]),

  component_scores: {
    code_quality: classify_score(Q.code_quality_metrics.build_consistency),
    test_coverage: classify_score(Q.test_coverage_indicators.test_frequency),
    documentation: classify_score(Q.documentation_quality.file_documentation_ratio),
    workflow_health: classify_score(1.0 - Q.workflow_health.error_rate)
  },

  quality_gates: {
    build_success: Q.code_quality_metrics.build_consistency > 0.9,
    test_reliability: Q.test_coverage_indicators.test_failure_rate < 0.1,
    documentation_adequate: Q.documentation_quality.file_documentation_ratio > 0.7,
    error_rate_acceptable: Q.workflow_health.error_rate < 0.05
  },

  trend_analysis: {
    quality_trajectory: calculate_quality_trend(Q.code_quality_metrics),
    test_coverage_evolution: analyze_test_trend(Q.test_coverage_indicators),
    error_rate_trend: analyze_error_trend(Q.workflow_health.error_rate),
    improvement_velocity: measure_quality_improvement_rate(Q)
  }
}

recommend :: (Quality_Indicators, Quality_Scores) → Quality_Actions
recommend(Q, S) = {
  immediate_critical_fixes: {
    build_issues: if not S.quality_gates.build_success then
      diagnose_build_problems(Q.code_quality_metrics),

    test_failures: if not S.quality_gates.test_reliability then
      identify_test_failure_patterns(Q.test_coverage_indicators),

    error_spike: if not S.quality_gates.error_rate_acceptable then
      analyze_error_causes(Q.workflow_health),

    documentation_gaps: if not S.quality_gates.documentation_adequate then
      prioritize_documentation_needs(Q.documentation_quality)
  },

  quality_improvements: {
    test_enhancement: if S.component_scores.test_coverage in ["C", "D"] then
      suggest_test_improvements(Q.test_coverage_indicators),

    code_refactoring: if S.component_scores.code_quality in ["C", "D"] then
      identify_refactoring_opportunities(Q.code_quality_metrics),

    doc_enhancement: if S.component_scores.documentation in ["C", "D"] then
      recommend_documentation_improvements(Q.documentation_quality),

    workflow_optimization: if S.component_scores.workflow_health in ["C", "D"] then
      suggest_workflow_improvements(Q.workflow_health)
  },

  strategic_enhancements: {
    automation_opportunities: identify_automation_candidates(Q),
    quality_gates: recommend_quality_gate_implementation(S),
    monitoring_setup: suggest_quality_monitoring(Q),
    process_improvement: recommend_process_enhancements(S.trend_analysis)
  }
}

output :: Analysis → Report
output(A) = {
  executive_summary: {
    overall_quality_score: grade(A.metrics.overall_quality_score),
    quality_grade: assign_letter_grade(A.metrics.overall_quality_score),
    key_strengths: identify_strengths(A.metrics.component_scores),
    critical_issues: identify_critical_gates(A.metrics.quality_gates),
    quality_trend: A.metrics.trend_analysis.quality_trajectory
  },

  quality_scorecard: {
    code_quality: {
      score: A.metrics.component_scores.code_quality,
      grade: assign_letter_grade(A.metrics.component_scores.code_quality),
      key_metrics: summarize_code_quality(A.indicators.code_quality_metrics)
    },

    test_coverage: {
      score: A.metrics.component_scores.test_coverage,
      grade: assign_letter_grade(A.metrics.component_scores.test_coverage),
      key_metrics: summarize_test_coverage(A.indicators.test_coverage_indicators)
    },

    documentation: {
      score: A.metrics.component_scores.documentation,
      grade: assign_letter_grade(A.metrics.component_scores.documentation),
      key_metrics: summarize_documentation(A.indicators.documentation_quality)
    },

    workflow_health: {
      score: A.metrics.component_scores.workflow_health,
      grade: assign_letter_grade(A.metrics.component_scores.workflow_health),
      key_metrics: summarize_workflow_health(A.indicators.workflow_health)
    }
  },

  quality_gates_status: {
    passed: count_true(A.metrics.quality_gates),
    failed: count_false(A.metrics.quality_gates),
    critical_failures: identify_critical_gate_failures(A.metrics.quality_gates),
    next_check: recommend_next_gate_check(A.metrics.quality_gates)
  },

  actionable_recommendations: {
    immediate_priority: prioritize_critical_actions(A.actions.immediate_critical_fixes),
    short_term_improvements: A.actions.quality_improvements,
    strategic_initiatives: A.actions.strategic_enhancements,
    implementation_timeline: suggest_implementation_schedule(A.actions)
  },

  quality_metrics_dashboard: {
    current_state: visualize_current_quality(A.metrics),
    historical_trends: display_quality_trends(A.metrics.trend_analysis),
    benchmark_comparison: compare_to_benchmarks(A.metrics),
    improvement_targets: set_quality_targets(A.metrics)
  }
} where ¬execute(recommendations)

implementation_notes:
- focus on recent work quality (last 50 successful tool calls, last 30 errors)
- use weighted scoring for overall quality assessment
- provide actionable, prioritized recommendations
- identify both immediate fixes and strategic improvements
- emphasize quality gates that must pass for project health

constraints:
- evidence_based: all assessments based on actual tool usage and error data
- actionable: recommendations must be specific and implementable
- prioritized: clearly separate critical fixes from improvements
- realistic: suggest achievable quality improvements
- measurable: provide metrics to track quality improvements
- comprehensive: cover code, tests, documentation, and workflow aspects