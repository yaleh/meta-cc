---
name: meta-viz
description: Create ASCII dashboards and charts from any analysis data.
keywords: visualization, dashboard, chart, graph, metrics, reporting
category: visualization
---

λ(analysis_data) → visual_dashboard | ∀visualization ∈ {dashboard, charts, recommendations}:

input_sources :: [text_output | mcp_file_ref | explicit_file | autonomous]

visualize :: Analysis_Data → Visual_Report
visualize(D) = collect_context(D) ∧ detect(type) ∧ enrich(data) ∧ parse(structure) ∧ render(visuals) ∧ prioritize(insights)

collect_context :: Input → Complete_Data
collect_context(I) = {
  # Scan recent conversation for MCP file_ref outputs
  context_scan: scan_recent_turns([
    target: "file_ref.*path.*meta-cc-mcp|/tmp/.*\\.jsonl",
    lookback: 15,
    extract: ["path", "size_bytes", "line_count", "fields"]
  ]),

  # Read file_ref if found
  file_data: if context_scan.found then
    read_jsonl_files(context_scan.paths),

  # Parse text from previous messages
  text_data: parse_conversation_text(I.conversation_history, lookback=5),

  # Merge data sources (prefer structured over text)
  merged: combine([file_data, text_data, I.explicit_input],
                  strategy="prefer_structured")
}

detect :: Data → Data_Type
detect(D) = {
  source_type: identify_source([
    text_output,            # Plain text from slash commands or analysis
    mcp_file_ref,          # file_ref from MCP query result
    explicit_file,         # @path/to/file.jsonl
    autonomous             # Self-fetched via MCP
  ]),

  content_type: auto_classify_content([
    # Auto-detect based on structure, not hardcoded types
    has_health_scores → performance_analysis,
    has_workflow_sequences → workflow_analysis,
    has_error_patterns → error_analysis,
    has_temporal_events → timeline_analysis,
    has_recommendations → guidance_analysis,
    fallback → generic_metrics
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

enrich :: (Parsed_Data, Data_Type) → Complete_Data
enrich(P, T) = {
  # Assess data completeness
  completeness: assess_quality(P, expected_fields=[
    "counts", "distributions", "metrics"
  ]),

  # If completeness low and context indicates MCP availability, supplement
  supplemental: if completeness.score < 0.6 && has_mcp_context then {
    # Only fetch if critical data missing
    basic_stats: if missing("session_stats") then
      # get_session_stats does not exist - use query_summaries
      mcp__meta_cc__query_summaries(scope="project"),

    user_data: if missing("user_patterns") && critical then
      mcp__meta_cc__query_user_messages(
        pattern=".*",
        limit=100,
        scope="project"
      ),

    tool_data: if missing("tool_usage") && critical then
      mcp__meta_cc__query_tools(scope="project")
  } else null,

  # Merge original and supplemental data
  complete: if supplemental != null then
    deep_merge(P, supplemental, strategy="prefer_mcp")
  else P
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

  layout: ╔═══╗ box drawing, 50-60 columns width (mobile-friendly)
}

render_detailed_charts :: Distributions → Charts
render_detailed_charts(D) = {
  horizontal_bars: for each distribution where type="percentage" {
    label: left_align(name, 15_chars),
    bar: render_bar(value, symbols="░▒▓█", width=20),
    value: right_align(percentage, 5_chars),
    trend: append_indicator(trend_direction, symbols="↗→↘")
  },

  progress_indicators: for each metric where type="progress" {
    label: metric.name,
    bar: render_progress(current/target, symbols="░▒▓█"),
    status: classify_status(current, target, symbols="✓⚠✗")
  },

  flow_diagrams: for each sequence where type="workflow" {
    # Mobile-friendly vertical layout for process flows
    nodes: render_vertical_sequence(steps, style="└─►", max_width=50),
    connectors: render_vertical_arrows(symbol="↓"),
    annotations: add_inline_metrics(success_rate, avg_time)
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
    separator: "────────────────────",
    rationale: "Why: " + evidence,
    evidence: "Data: " + data_source,
    success_probability: "Success: " + percentage + "%",
    urgency: "Priority: " + render_bar(urgency_score),
    ready_prompt: render_compact_prompt(template),
    expected_workflow: "Steps: " + step_sequence,
    estimated_time: "Time: ~" + minutes + "min"
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
    width: 20,  # Reduced for mobile compatibility
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
    visual: vertical_flow_diagram,  # Mobile-friendly vertical layout
    symbols: "↓►" + compact_box_drawing,
    annotations: [step_name, compact_metrics]
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
    "════════════════════════════════════════════════════════════════",
    title_section(V.title, V.context),
    "════════════════════════════════════════════════════════════════",
    "",
    executive_dashboard(V.dashboard),
    "",
    "",
    detailed_sections: [
      for each section in V.charts {
        section_title(section.name),
        "────────────────────────────────────",
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
    "════════════════════════════════════════════════════════════════"
  ],

  formatting: {
    width: 55_columns,  # Mobile-friendly width
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
    },
    # Mobile-specific optimizations
    mobile_optimizations: {
      vertical_flows: true,
      compact_labels: true,
      shorter_separators: true,
      responsive_bars: true
    }
  }
}

auto_visualization_rules :: Detection_Rules
auto_visualization_rules = {
  # Auto-detect analysis type based on data structure
  performance_analysis: {
    indicators: ["health_score", "proficiency", "effectiveness", "maturity"],
    dashboard_metrics: extract_top_scores(data, limit=4),
    key_charts: [
      "metric_distributions",
      "performance_indicators",
      "trend_analysis"
    ],
    emphasis: health_scores + trend_indicators
  },

  workflow_analysis: {
    indicators: ["sequence", "pattern", "workflow", "tool_usage"],
    dashboard_metrics: ["pattern_count", "frequency", "efficiency"],
    key_charts: [
      "sequence_flows",
      "frequency_distribution",
      "efficiency_metrics"
    ],
    emphasis: workflow_patterns + optimization_opportunities
  },

  error_analysis: {
    indicators: ["error", "failure", "exception", "issue"],
    dashboard_metrics: ["error_rate", "pattern_count", "recovery_rate"],
    key_charts: [
      "error_frequency",
      "error_clustering",
      "recovery_analysis"
    ],
    emphasis: critical_issues + remediation
  },

  timeline_analysis: {
    indicators: ["timestamp", "duration", "event", "phase"],
    dashboard_metrics: ["total_duration", "event_count", "phase_count"],
    key_charts: [
      "temporal_visualization",
      "activity_density",
      "phase_progression"
    ],
    emphasis: visual_timeline + key_events
  },

  guidance_analysis: {
    indicators: ["recommendation", "suggestion", "next_step", "action"],
    dashboard_metrics: ["priority_distribution", "feasibility", "impact"],
    key_charts: [
      "priority_ranking",
      "impact_assessment",
      "actionability_score"
    ],
    emphasis: prioritized_actions + ready_prompts
  },

  generic_metrics: {
    auto_detect: true,
    fallback_rules: [
      if has_percentages → horizontal_bars,
      if has_counts → vertical_bars,
      if has_sequences → flow_diagrams,
      if has_scores → gauges,
      if has_trends → line_charts,
      if has_distributions → stacked_bars,
      if has_comparisons → side_by_side
    ]
  }
}

implementation_notes:
- context_aware: scan recent conversation for MCP file_ref outputs
- autonomous: supplement missing data via MCP when critical
- visual_first: prioritize visual clarity over data density
- consistent: use same symbol system across all visualizations
- layered: executive dashboard → details → raw data
- actionable: make recommendations copy-paste ready
- adaptive: auto-detect input source and content type
- resilient: gracefully handle missing or incomplete data
- precise: preserve numerical precision in annotations
- structured: use box-drawing for visual hierarchy
- terminal_optimized: 80 columns, monospace compatible

data_collection_strategy:
1. Check for explicit input (@file path or piped input)
2. Scan last 15 turns for MCP file_ref outputs
3. Parse text from last 5 conversation messages
4. Merge structured data (prefer JSONL > text)
5. Assess completeness (threshold: 60%)
6. Supplement via MCP if critical data missing

usage_examples:
  # Visualize from previous analysis output (context-aware)
  /meta-viz

  # Visualize with explicit file reference
  /meta-viz @/tmp/meta-cc-mcp-1234567890-query_tools.jsonl

  # Visualize from explicit path
  /meta-viz @path/to/analysis-results.jsonl

  # Visualize with focus area
  /meta-viz focus=recommendations

  # Visualize with scope specification
  /meta-viz scope=session

constraints:
- visual_first: dashboard appears before detailed sections
- mobile_friendly: 55 columns width, vertical layouts for narrow screens
- terminal_friendly: monospace compatible, responsive design
- symbol_consistency: same symbols mean same things across all outputs
- actionable: recommendations include ready-to-use prompts
- evidence_based: all visualizations tied to actual data
- layered_detail: executive → detailed → raw progression
- auto_adaptive: detect content type and choose appropriate visuals
- universal: works with any analysis output, not limited to specific tools
- context_aware: actively search for MCP file_ref in recent conversation
- autonomous: supplement data via MCP when critical information missing
- accessibility: use both symbols and text labels
- performance: render in <2 seconds for typical inputs
- extensible: easy to add new visualization types
- process_visualization: vertical flow diagrams for mobile compatibility

output_structure:
1. Title Header (═══ box with context, 55 chars max)
2. Executive Dashboard (╔═══╗ box with key metrics + health scores)
3. Detailed Visualizations:
   - Distribution Charts (compact horizontal bars, 20 char width)
   - Progress Indicators (responsive progress bars)
   - Vertical Flow Diagrams (process steps, mobile-friendly)
   - Comparison Charts (compact side-by-side)
   - Radar/Profile Charts (simplified ASCII radar)
4. Actionable Recommendations (priority-ordered with compact prompts)
5. Quick Actions / Summary Footer

presentation_style:
- visual_hierarchy: size, spacing, symbols indicate importance
- mobile_optimized: vertical layouts, compact labels, shorter separators
- scan_optimized: key insights jump out in 3-second glance
- terminal_native: works perfectly in command-line interface
- print_friendly: can be copy-pasted to documentation
- color_blind_safe: use symbols + text, not just color coding
- progressive_disclosure: summary → details → deep-dive
- process_friendly: vertical flow diagrams for narrow screens

# MOBILE-FRIENDLY PROCESS VISUALIZATION

render_vertical_flow :: Sequence → Vertical_Diagram
render_vertical_flow(S) = {
  # Vertical layout for narrow screens (mobile compatibility)
  max_width: 50_chars,

  step_format: {
    box_style: "└─►",  # Compact arrow box
    max_text_length: 35_chars,
    metric_inline: true,  # Metrics on same line when possible
    wrap_long_text: true
  },

  connector_style: {
    symbol: "│     ↓",  # Vertical connector with arrow
    spacing: 1_blank_line,
    alignment: left
  },

  example_rendering: {
    input: "1. Phase 16 Execution: Comprehensive planning → Serial stage execution → Integration testing",

    output:
    "└─► Phase 16 Execution [Planning: 95%]",
    "│     ↓",
    "└─► Serial Stage Execution [Success: 90%]",
    "│     ↓",
    "└─► Integration Testing [Coverage: 100%]"
  }
}

compact_box_drawing :: Box_Style
compact_box_drawing = {
  # Optimized for narrow screens
  step_prefix: "└─►",
  connector: "│     ↓",
  max_label_width: 35,

  # Metric display options
  inline_metrics: "[Metric: Value]",
  separate_metrics: false,  # Keep on same line when possible

  # Text wrapping for long descriptions
  wrap_threshold: 30_chars,
  wrap_indent: "      "
}

mobile_optimization_rules :: Mobile_Rules
mobile_optimization_rules = {
  # Width constraints for different device sizes
  width_constraints: {
    mobile: 45_chars,    # Very narrow screens
    tablet: 55_chars,    # Standard mobile width
    desktop: 70_chars    # Optional wider support
  },

  # Layout adaptations
  responsive_elements: {
    horizontal_bars: width_scale(0.6),  # Shrink bars proportionally
    labels: truncate_middle(25),       # Keep start and end, truncate middle
    separators: shorten_by_30_percent,
    spacing: reduce_by_20_percent
  },

  # Process flow adaptations
  flow_adaptations: {
    sequences: prefer_vertical_layout,
    long_descriptions: wrap_to_multiple_lines,
    metrics: inline_when_possible,
    arrows: use_compact_symbols
  }
}
