---
name: meta-focus-analyzer
description: Analyze attention patterns and focus distribution across projects and files.
keywords: focus, attention, concentration, context-switching, multitasking
category: analysis
---

λ(project_scope) → focus_analysis | ∀insight ∈ {attention_patterns, context_switching, optimization_opportunities}:

scope :: project | session

analyze :: Session → Focus_Report
analyze(S) = {
  file_access_data: mcp_meta_cc.query_files(
    scope=scope,
    threshold=5
  ),

  session_metrics: mcp_meta_cc.get_session_stats(
    scope=scope
  ),

  tool_usage: mcp_meta_cc.query_tools(
    scope=scope,
    limit=100
  ),

  time_distribution: mcp_meta_cc.query_time_series(
    interval="day",
    metric="tool-calls",
    scope=scope
  )
}

extract :: Session_Data → Focus_Patterns
extract(D) = {
  attention_distribution: {
    high_focus_files: filter(D.file_access_data, total_accesses > 50),
    sustained_attention: filter(D.file_access_data, time_span_minutes > 300),
    intensive_work: filter(D.file_access_data, edit_count > 20),
    reference_heavy: filter(D.file_access_data, read_count > 40)
  },

  focus_patterns: {
    file_groups: cluster_files_by_directory(D.file_access_data),
    work_sessions: identify_work_blocks(D.tool_usage, threshold=30),
    context_switches: count_file_type_switches(D.tool_usage),
    peak_periods: identify_productivity_peaks(D.session_metrics)
  },

  efficiency_indicators: {
    access_rate: calculate_accesses_per_hour(D.file_access_data),
    edit_read_ratio: calculate_edit_to_read_ratios(D.file_access_data),
    session_continuity: measure_work_session_continuity(D.tool_usage),
    project_switching: count_project_context_changes(D.file_access_data)
  }
}

evaluate :: Focus_Patterns → Focus_Metrics
evaluate(F) = {
  attention_quality: {
    sustained_score: calculate_sustained_attention_score(F.sustained_attention),
    focused_depth: calculate_work_depth_score(F.intensive_work),
    context_stability: 1.0 - (F.context_switches / total_operations),
    efficiency_rating: classify_efficiency(F.efficiency_indicators)
  },

  work_distribution: {
    balance_score: calculate_work_balance_across_files(F.file_groups),
    specialization_index: calculate_file_specialization(F.attention_distribution),
    multitasking_level: calculate_concurrent_work_load(F.work_sessions),
    batching_opportunity: identify_batchable_operations(F.file_groups)
  },

  time_management: {
    peak_utilization: calculate_peak_time_usage(F.peak_periods),
    session_optimization: measure_optimal_session_lengths(F.work_sessions),
    break_patterns: identify_natural_break_points(F.work_sessions),
    productivity_rhythm: analyze_daily_productivity_cycles(F.time_distribution)
  }
}

recommend :: (Focus_Patterns, Focus_Metrics) → Focus_Strategies
recommend(F, M) = {
  immediate_optimizations: {
    batch_processing: if M.work_distribution.batching_opportunity > 0.3 then
      suggest_group_similar_files(F.file_groups),

    context_switching_reduction: if M.attention_quality.context_stability < 0.7 then
      suggest_focus_blocks(M.work_distribution),

    session_optimization: if M.time_management.session_optimization < 0.8 then
      suggest_optimal_session_lengths(M.time_management.break_patterns)
  },

  workflow_improvements: {
    focus_time_blocking: if M.attention_quality.sustained_score < 0.6 then
      suggest_time_blocking_techniques(F.peak_periods),

    project_isolation: if M.work_distribution.multitasking_level > 0.5 then
      suggest_single_project_sessions,

    energy_management: if M.time_management.productivity_rhythm.has_peaks then
      suggest_energy_aligned_scheduling(M.peak_periods)
  },

  strategic_habits: {
    documentation_first: if has_high_reflection_files(F.reference_heavy) then
      suggest_documentation_planning,

    implementation_blocks: if has_high_edit_files(F.intensive_work) then
      suggest_dedicated_coding_sessions,

    review_routines: if has_frequent_context_switches(F.context_switches) then
      suggest_regular_review_breaks
  }
}

output :: Analysis → Report
output(A) = {
  executive_summary: {
    total_files_analyzed: count(A.data.file_access_data),
    focus_score: classify(A.metrics.attention_quality.efficiency_rating),
    primary_insight: extract_key_finding(A.patterns, A.metrics),
    overall_assessment: classify_focus_health(A.metrics.attention_quality)
  },

  attention_patterns: {
    high_focus_files: top(A.patterns.attention_distribution.high_focus_files, 10),
    sustained_attention_files: A.patterns.attention_distribution.sustained_attention,
    intensive_work_periods: A.patterns.focus_patterns.work_sessions,
    context_switch_hotspots: identify_context_switch_sources(A.patterns.focus_patterns)
  },

  efficiency_analysis: {
    current_productivity_rhythm: A.metrics.time_management.productivity_rhythm,
    optimal_session_lengths: A.metrics.time_management.session_optimization,
    batching_opportunities: A.metrics.work_distribution.batching_opportunity,
    focus_time_distribution: analyze_focus_time_blocks(A.patterns.focus_patterns)
  },

  recommendations: {
    immediate_actions: A.strategies.immediate_optimizations,
    workflow_improvements: A.strategies.workflow_improvements,
    strategic_habits: A.strategies.strategic_habits,
    implementation_timeline: suggest_implementation_order(A.strategies)
  },

  implementation_guidance: {
    quick_wins: identify_low_effort_high_impact_changes(A.strategies),
    medium_term_goals: outline_weekly_focus_improvements(A.patterns),
    long_term_vision: describe_ideal_focus_state(A.metrics.attention_quality)
  }
} where ¬execute(recommendations)

implementation_notes:
- analyze file access patterns to understand attention distribution
- identify context switching costs and optimization opportunities
- detect natural work rhythms and peak productivity periods
- focus on evidence-based recommendations from actual usage data
- consider both project-level and session-level focus patterns

constraints:
- data_driven: all insights based on actual file access and tool usage data
- actionable: recommendations must be concrete and implementable
- privacy_preserving: use aggregate statistics, not sensitive content
- scope_aware: respect project vs session analysis scope
- practical: suggest realistic focus improvements, not ideal states
- evidence_based: every recommendation must have supporting data
