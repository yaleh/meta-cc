---
name: meta-viz
description: Visualize meta-analysis outputs (meta-habits, meta-coach, meta-guide, meta-errors, meta-timeline) with ASCII dashboards, charts, and actionable insights. Transforms text data into visual summaries with health scores, trend indicators, and priority rankings.
---

λ(analysis_data) → visual_dashboard | ∀visualization ∈ {dashboard, charts, recommendations}:

input_sources :: [slash_output | mcp_file_ref | explicit_file]

visualize :: Analysis_Data → Visual_Report
visualize(D) = detect(type) ∧ parse(structure) ∧ render(visuals) ∧ prioritize(insights)

detect :: Data → Data_Type
detect(D) = {
  source_type: identify_source([
    slash_command_output,    # piped from /meta-* command
    mcp_file_ref,           # file_ref from MCP query result
    explicit_file           # @path/to/analysis.jsonl
  ]),

  content_type: classify_content([
    habits_analysis,        # from /meta-habits
    coaching_report,        # from /meta-coach
    guidance_suggestions,   # from /meta-guide
    error_patterns,        # from /meta-errors
    timeline_events,       # from /meta-timeline
    generic_metrics        # fallback: auto-detect structure
  ]),

  data_structures: extract_structures([
    percentages,           # N% patterns
    counts,               # N items, N occurrences
    distributions,        # category: count pairs
    sequences,            # ordered events/patterns
    scores,               # ratings, health metrics
    recommendations,      # actionable items
    trends                # temporal changes
  ])
}

parse :: (Data, Data_Type) → Structured_Data
parse(D, T) = {
  metrics: extract_numerical([
    percentages: pattern="(\d+)%",
    counts: pattern="(\d+)\s+(messages|items|times|occurrences)",
    scores: pattern="(\d+\.?\d*)/(\d+)",
    rates: pattern="(\d+\.?\d*)%?\s+rate"
  ]),

  distributions: extract_categorical([
    type_breakdown: "Type:\s+(\d+)%",
    tool_usage: "Tool.*:\s+(\d+)%",
    pattern_frequency: "Pattern.*:\s+(\d+)\s+times"
  ]),

  sequences: extract_ordered([
    workflows: "Step 1 → Step 2 → Step 3",
    evolution: "Day 1.*Day 2.*Day 3",
    causality: "A → B → C"
  ]),

  recommendations: extract_actionable([
    immediate: priority="high",
    short_term: priority="medium",
    long_term: priority="low"
  ]),

  trends: extract_temporal([
    increasing: indicators="↗|increasing|growing|up",
    decreasing: indicators="↘|decreasing|declining|down",
    stable: indicators="→|stable|steady|constant"
  ])
}

render :: Structured_Data → Visual_Elements
render(S) = {
  dashboard: render_executive_dashboard(S),
  charts: render_detailed_charts(S),
  recommendations: render_actionable_items(S),
  narratives: render_insights(S)
}

render_executive_dashboard :: Metrics → Dashboard
render_executive_dashboard(M) = {
  header: ═══ box with title and key metrics,

  health_scores: {
    overall: render_score(M.overall_score, format="●/100"),
    dimensions: render_multi_gauge([
      {name: dimension.name, value: dimension.score, symbol: "▓"}
      for dimension in M.key_dimensions
    ])
  },

  quick_insights: {
    strengths: extract_top(M.metrics, where="value > 80%", symbol="✓"),
    concerns: extract_top(M.metrics, where="value < 40%", symbol="⚠"),
    trends: summarize_directions(M.trends, symbols="↗→↘")
  },

  layout: ╔═══╗ box drawing, 80 columns width
}

render_detailed_charts :: Distributions → Charts
render_detailed_charts(D) = {
  horizontal_bars: for each distribution where type="percentage" {
    label: left_align(name, 20_chars),
    bar: render_bar(value, symbols="░▒▓█", width=30),
    value: right_align(percentage, 6_chars),
    trend: append_indicator(trend_direction, symbols="↗→↘")
  },

  progress_indicators: for each metric where type="progress" {
    label: metric.name,
    bar: render_progress(current/target, symbols="░▒▓█"),
    status: classify_status(current, target, symbols="✓⚠✗")
  },

  flow_diagrams: for each sequence where type="workflow" {
    nodes: render_boxes(steps, style="┌─┐│└┘"),
    arrows: connect_nodes(symbol="→"),
    annotations: add_metrics(success_rate, avg_time)
  },

  comparison_charts: for each pair where type="before_after" {
    side_by_side: render_dual_bars(before, after),
    delta: calculate_change(after - before, format="±N%"),
    significance: indicator(change_magnitude, symbols="░▒▓█")
  },

  radar_charts: for each profile where dimensions >= 3 {
    axes: render_ascii_radar(dimensions, max_value=100),
    plot: overlay_values(current_scores),
    labels: annotate_axes(dimension_names)
  }
}

render_actionable_items :: Recommendations → Action_List
render_actionable_items(R) = {
  priority_structure: {
    high: {
      symbol: "🔴",
      urgency_bar: "████████████████████",
      presentation: prominent_box_with_ready_prompts
    },
    medium: {
      symbol: "🟡",
      urgency_bar: "██████████░░░░░░░░░░",
      presentation: structured_list_with_examples
    },
    low: {
      symbol: "🟢",
      urgency_bar: "████░░░░░░░░░░░░░░░░",
      presentation: compact_list_with_references
    }
  },

  item_format: {
    title: recommendation.title,
    separator: "─────────────────────────────────────",
    rationale: "Rationale: " + evidence,
    evidence: "Evidence: " + data_source,
    success_probability: "Success: " + percentage + "% (historical)",
    urgency: "Urgency: " + render_bar(urgency_score),
    ready_prompt: render_copy_paste_prompt(template),
    expected_workflow: "Expected: " + step_sequence,
    estimated_time: "Time: ~" + minutes + " minutes"
  }
}

symbols :: Symbol_System
symbols = {
  progress: {
    empty: "░",
    low: "▒",
    medium: "▓",
    full: "█"
  },

  health: {
    critical: "⚫",
    poor: "⚪",
    fair: "◐",
    good: "◑",
    excellent: "◕"
  },

  trend: {
    sharp_up: "↗",
    up: "↑",
    stable: "→",
    down: "↘",
    sharp_down: "↓"
  },

  rating: {
    filled: "★",
    empty: "☆"
  },

  status: {
    pass: "✓",
    fail: "✗",
    warning: "⚠",
    info: "ℹ"
  },

  priority: {
    critical: "🔴",
    high: "🟠",
    medium: "🟡",
    low: "🟢"
  },

  intensity: {
    low: "░░░",
    medium: "▒▒▒",
    high: "▓▓▓",
    peak: "████"
  },

  box_drawing: {
    horizontal: "─",
    vertical: "│",
    corners: "┌┐└┘",
    double_line: "═║╔╗╚╝",
    intersections: "├┤┬┴┼"
  }
}

visualization_types :: Type_Mapping
visualization_types = {
  percentage_data: {
    visual: horizontal_bar,
    symbols: "░▒▓█",
    width: 30,
    annotations: [value, trend]
  },

  count_data: {
    visual: vertical_bar,
    symbols: "▁▂▃▄▅▆▇█",
    annotations: [count, percentage_of_total]
  },

  trend_data: {
    visual: line_chart_ascii,
    symbols: "━ ╱╲",
    annotations: [direction_arrow, delta_value]
  },

  distribution_data: {
    visual: stacked_bar,
    symbols: "░▒▓█" + color_coding,
    annotations: [category, percentage]
  },

  sequence_data: {
    visual: flow_diagram,
    symbols: "→↓" + box_drawing,
    annotations: [step_name, metrics]
  },

  score_data: {
    visual: gauge,
    symbols: "⚫⚪◐◑◕" or "★☆",
    annotations: [score, threshold, status]
  },

  comparison_data: {
    visual: side_by_side_bars,
    symbols: "▓▓▓" vs "░░░",
    annotations: [before, after, delta]
  },

  relationship_data: {
    visual: node_edge_graph,
    symbols: "●──○" + arrows,
    annotations: [node_labels, edge_weights]
  }
}

output :: Visual_Report → Formatted_Output
output(V) = {
  structure: [
    "═══════════════════════════════════════════════════════════════════════════════",
    title_section(V.title, V.context),
    "═══════════════════════════════════════════════════════════════════════════════",
    "",
    executive_dashboard(V.dashboard),
    "",
    "",
    detailed_sections: [
      for each section in V.charts {
        section_title(section.name),
        "───────────────────────────────────────────────────────────────────────────────",
        "",
        section.visuals,
        "",
        section.insights,
        ""
      }
    ],
    "",
    recommendations_section(V.recommendations),
    "",
    "═══════════════════════════════════════════════════════════════════════════════"
  ],

  formatting: {
    width: 80_columns,
    alignment: {
      title: center,
      headers: left,
      metrics: right_align_numbers,
      labels: left_align_text
    },
    spacing: {
      between_sections: 2_blank_lines,
      within_sections: 1_blank_line,
      compact_lists: 0_blank_lines
    }
  }
}

auto_visualization_rules :: Detection_Rules
auto_visualization_rules = {
  habits_analysis: {
    dashboard_metrics: [
      "communication_style",
      "planning_style",
      "tool_proficiency"
    ],
    key_charts: [
      "prompt_type_distribution",
      "tool_adoption_metrics",
      "workflow_efficiency",
      "typical_sequences"
    ],
    emphasis: tool_usage_patterns + workflow_sequences
  },

  coaching_report: {
    dashboard_metrics: [
      "session_health_score",
      "context_quality",
      "delegation_effectiveness",
      "workflow_maturity"
    ],
    key_charts: [
      "interaction_type_breakdown",
      "delegation_patterns",
      "feedback_analysis",
      "conversation_flows"
    ],
    emphasis: effectiveness_metrics + actionable_recommendations
  },

  guidance_suggestions: {
    dashboard_metrics: [
      "trajectory_state",
      "momentum_indicator",
      "blocker_count"
    ],
    key_charts: [
      "recent_intent_trajectory",
      "pattern_matches",
      "success_probabilities"
    ],
    emphasis: prioritized_suggestions + ready_prompts
  },

  error_patterns: {
    dashboard_metrics: [
      "error_rate",
      "pattern_count",
      "recovery_success"
    ],
    key_charts: [
      "error_frequency_by_type",
      "error_clustering",
      "recovery_cycles"
    ],
    emphasis: critical_errors + fix_recommendations
  },

  timeline_events: {
    dashboard_metrics: [
      "duration",
      "total_events",
      "phase_count"
    ],
    key_charts: [
      "vertical_timeline_ascii",
      "phase_flow_diagram",
      "activity_density_evolution"
    ],
    emphasis: visual_timeline + critical_moments
  },

  generic_metrics: {
    auto_detect: true,
    fallback_rules: [
      if has_percentages → horizontal_bars,
      if has_counts → vertical_bars,
      if has_sequences → flow_diagrams,
      if has_scores → gauges,
      if has_trends → line_charts
    ]
  }
}

implementation_notes:
- prioritize visual clarity over data density
- use consistent symbol system across all visualizations
- provide executive dashboard for quick 3-second understanding
- layer information: overview → details → raw data
- make recommendations copy-paste ready
- auto-detect input source and content type
- gracefully handle missing or incomplete data
- preserve numerical precision in annotations
- use box-drawing for visual structure
- align elements for terminal rendering (80 columns)

usage_examples:
  # Visualize meta-habits output
  /meta-habits | /meta-viz

  # Visualize meta-coach output with explicit scope
  /meta-coach scope=project | /meta-viz

  # Visualize from MCP file_ref (after large query)
  /meta-viz source=@/tmp/meta-cc-mcp-*.jsonl

  # Visualize explicit analysis file
  /meta-viz @path/to/analysis-results.jsonl

  # Visualize with custom focus
  /meta-viz source=@analysis.jsonl focus=recommendations

constraints:
- visual_first: dashboard appears before detailed sections
- terminal_friendly: 80 columns, monospace compatible
- symbol_consistency: same symbols mean same things across all outputs
- actionable: recommendations include ready-to-use prompts
- evidence_based: all visualizations tied to actual data
- layered_detail: executive → detailed → raw progression
- auto_adaptive: detect content type and choose appropriate visuals
- accessibility: use both symbols and text labels
- performance: render in <2 seconds for typical inputs
- extensible: easy to add new visualization types

output_structure:
1. Title Header (═══ box with context)
2. Executive Dashboard (╔═══╗ box with key metrics + health scores)
3. Detailed Visualizations:
   - Distribution Charts (horizontal bars)
   - Progress Indicators (progress bars)
   - Flow Diagrams (boxes + arrows)
   - Comparison Charts (side-by-side)
   - Radar/Profile Charts (ASCII radar)
4. Actionable Recommendations (priority-ordered with ready prompts)
5. Quick Actions / Summary Footer

presentation_style:
- visual_hierarchy: size, spacing, symbols indicate importance
- scan_optimized: key insights jump out in 3-second glance
- terminal_native: works perfectly in command-line interface
- print_friendly: can be copy-pasted to documentation
- color_blind_safe: use symbols + text, not just color coding
- progressive_disclosure: summary → details → deep-dive
