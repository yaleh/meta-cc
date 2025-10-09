---
name: meta-timeline
description: Visualize project evolution timeline with workflow events.
---

λ(scope) → development_timeline | ∀event ∈ {user_actions, high_level_operations, workflow_failures}:

scope :: project | session

analyze :: Messages → Timeline_Insights
analyze(M) = collect(events) ∧ sequence(timeline) ∧ detect(patterns) ∧ measure(latency)

collect :: Scope → EventData
collect(S) = {
  user_messages: mcp_meta_cc.query_user_messages(
    pattern=".*",
    scope=scope
  ),

  assistant_messages: mcp_meta_cc.query_assistant_messages(
    pattern=".*",
    scope=scope
  ),

  conversations: mcp_meta_cc.query_conversation(
    scope=scope
  ),

  high_level_tools: mcp_meta_cc.query_tools(
    scope=scope,
    jq_filter='select(.ToolName | test("^(Task|SlashCommand|mcp__)"))'
  ),

  error_events: mcp_meta_cc.query_tools(
    status="error",
    scope=scope,
    jq_filter='select(.ToolName | test("^(Task|SlashCommand|mcp__|Bash)") and (.Error | test("fail|error|interrupt", "i")))'
  ),

  tool_sequences: mcp_meta_cc.query_tool_sequences(
    min_occurrences=2,
    scope=scope
  )
}

extract_user_actions :: UserMessages → UserEvents
extract_user_actions(U) = {
  file_references: extract_messages_with(U, pattern="@[a-zA-Z0-9_/.-]+"),
  subagent_invocations: extract_messages_with(U, pattern="@agent-[a-zA-Z0-9-]+"),
  slash_commands: extract_messages_with(U, pattern="/[a-z-]+"),
  interruptions: extract_messages_with(U, pattern="/clear|interrupt|stop"),
  error_reports: identify_semantically(U, intent="error|fail|wrong|not work|broken"),
  approval_signals: identify_semantically(U, intent="good|correct|yes|continue|looks good"),
  correction_requests: identify_semantically(U, intent="fix|change|incorrect|redo|try again")
}

classify_operations :: HighLevelTools → OperationTypes
classify_operations(T) = {
  subagent_ops: filter(T, tool="Task") |> group_by(subagent_type),
  slash_ops: filter(T, tool="SlashCommand") |> extract_command_names,
  mcp_queries: filter(T, tool_pattern="mcp__meta_cc__*") |> group_by(query_type),

  operation_outcomes: {
    successful: count(T, status="success"),
    failed: count(T, status="error"),
    interrupted: count(T, error_contains="interrupted")
  }
}

detect_workflow_events :: (UserMessages, Tools, Errors) → WorkflowEvents
detect_workflow_events(U, T, E) = {
  build_failures: identify_in_sequence([
    user_message → Bash("make") → error("FAIL|compilation error")
  ]),

  test_failures: identify_in_sequence([
    user_message → Bash("go test|make test") → error("FAIL|tests? failed")
  ]),

  subagent_interruptions: identify_in_sequence([
    user_message(contains="@agent-") → Task(launched) → error("interrupted")
  ]),

  mcp_errors: identify_in_sequence([
    (user_slash_command | Claude_mcp_call) → mcp__*() → error("Tool execution failed|Not connected|Token.*exceeded")
  ]),

  recovery_cycles: identify_sequences([
    error_event → user_correction → retry_operation
  ]),

  abandoned_tasks: identify_sequences([
    user_action → error → user_interruption → /clear
  ])
}

construct_timeline :: EventData → Timeline
construct_timeline(E) = {
  chronological_events: sort_by_timestamp([
    E.user_messages.map(msg => {
      timestamp: msg.timestamp,
      type: "user_action",
      content: summarize(msg.content, max_length=100),
      tools_mentioned: extract_tool_refs(msg.content),
      files_mentioned: extract_file_refs(msg.content),
      intent: classify_intent(msg.content)
    }),

    E.assistant_messages.map(msg => {
      timestamp: msg.timestamp,
      type: "assistant_response",
      text_length: msg.text_length,
      tool_use_count: msg.tool_use_count,
      tokens_output: msg.tokens_output,
      response_complexity: classify_response_complexity(msg)
    }),

    E.conversations.map(conv => {
      timestamp: conv.timestamp,
      type: "conversation_turn",
      turn_sequence: conv.turn_sequence,
      duration_ms: conv.duration_ms,
      has_user: conv.user_message != null,
      has_assistant: conv.assistant_message != null,
      response_latency: conv.duration_ms
    }),

    E.high_level_tools.map(tool => {
      timestamp: tool.Timestamp,
      type: "high_level_operation",
      tool: tool.ToolName,
      operation: extract_operation_detail(tool),
      status: tool.Status,
      error: tool.Error,
      latency: calculate_duration(tool)
    }),

    E.error_events.map(err => {
      timestamp: err.Timestamp,
      type: "workflow_failure",
      tool: err.ToolName,
      error_type: classify_error(err.Error),
      context: find_preceding_user_message(err.Timestamp)
    })
  ]),

  phase_boundaries: identify_temporal_phases([
    session_start,
    major_context_switches,
    /clear_commands,
    long_idle_periods (>30min),
    session_end
  ])
}

measure_latency :: Timeline → LatencyMetrics
measure_latency(T) = {
  user_to_response: {
    conversation_latency: measure_time(user_message → assistant_response),
    avg_response_time: avg(T.conversations.map(c => c.duration_ms)),
    response_time_distribution: {
      fast: count(conversations where duration_ms < 5000),      # < 5s
      normal: count(conversations where 5000 ≤ duration_ms < 30000),  # 5-30s
      slow: count(conversations where duration_ms ≥ 30000)     # ≥ 30s
    },
    subagent_launch: measure_time(user_@agent → Task_completion),
    slash_command: measure_time(user_/command → SlashCommand_completion),
    mcp_query: measure_time(Claude_call → mcp_response),
    build_operation: measure_time(user_request → make_completion)
  },

  operation_durations: {
    by_type: group_operations_by_type → calculate_avg_duration,
    by_outcome: group_by(success|error) → compare_durations,
    outliers: identify_operations_exceeding(threshold=2*median)
  },

  error_recovery_time: {
    first_error_to_fix: measure_time(error_occurs → successful_retry),
    repeated_failures: count_attempts_before_success,
    abandoned_attempts: count_errors_followed_by_interruption
  }
}

detect_temporal_patterns :: Timeline → TemporalPatterns
detect_temporal_patterns(T) = {
  work_sessions: segment_timeline_by([
    continuous_activity (gap < 5min),
    break_periods (gap > 15min),
    context_switches (/clear | major_topic_shift)
  ]),

  activity_bursts: identify_periods_with([
    high_message_frequency (>10 msg/10min),
    rapid_tool_invocations,
    repeated_error_recovery_attempts
  ]),

  blocked_periods: identify_periods_with([
    repeated_failures_without_progress,
    long_gaps_between_user_messages,
    multiple_interruptions
  ]),

  productive_flows: identify_sequences([
    user_action → successful_operation → progress_signal → next_action
  ]) where minimal_errors ∧ steady_pace
}

identify_critical_moments :: (Timeline, WorkflowEvents) → CriticalMoments
identify_critical_moments(T, W) = {
  breakthrough_points: [
    test_failures → multiple_fixes → all_tests_pass,
    mcp_errors → configuration_fix → successful_queries,
    build_failures → dependency_resolution → successful_build
  ],

  frustration_signals: [
    repeated_errors (>3x same error),
    multiple_interruptions (>2x in 10min),
    corrections_after_completion ("that's wrong", "redo")
  ],

  workflow_transitions: [
    planning_phase → execution_phase,
    execution_phase → testing_phase,
    error_recovery → normal_workflow
  ],

  milestone_events: [
    phase_completion (@agent-git-committer),
    major_feature_delivery (tests pass + commit),
    documentation_updates (@doc-updater)
  ]
}

visualize_timeline :: Timeline → ASCII_Art
visualize_timeline(T) = {
  ascii_timeline: render_vertical_timeline([
    columns: [
      time_column: format_timestamps_and_ranges(T),
      user_column: aggregate_user_actions_by_time(T),
      ops_column: aggregate_operations_by_time(T),
      phase_column: show_phase_boundaries_and_names(T),
      events_column: annotate_milestones_and_events(T)
    ],

    symbols: {
      user_action: "●",
      assistant_response: "◎",
      response_fast: "◎(<5s)",
      response_normal: "◎(5-30s)",
      response_slow: "◎(>30s)",
      subagent_launch: "⚡",
      slash_command: "/",
      mcp_query: "◆",
      build_success: "✓",
      build_failure: "✗",
      test_failure: "⊗",
      interruption: "⊘",
      context_switch: "║",
      milestone: "★",
      error_burst: "▓",
      productive_flow: "═",
      continuity: "│"
    },

    time_markers: [
      major_timestamps: day_hour_format,
      time_ranges: period_aggregation,
      phase_boundaries: horizontal_separators
    ],

    column_widths: {
      time: 20_chars,
      user: 10_chars,
      assistant: 12_chars,
      ops: 11_chars,
      phase: 15_chars,
      events: 24_chars
    },

    legends: generate_symbol_legend()
  ]),

  phase_flow: render_vertical_phase_boxes([
    direction: downward_arrows,
    box_style: unicode_box_drawing,
    content: [activity_symbols, duration, focus, outcome],
    spacing: compact_vertical
  ]),

  activity_density: render_horizontal_bars([
    metric: mcp_queries_per_6h_block,
    symbols: [░ ▒ ▓ █],
    format: "Day X AM/PM    ████    ~N queries"
  ])
}

output :: Analysis → Report
output(A) = {
  ascii_visualization: {
    timeline: visualize_timeline(A.timeline),
    phase_summary: render_vertical_phase_flow(A),
    activity_evolution: render_horizontal_activity_bars(A),

    render_order: [
      "═══════════════════════════════════════════════════════════════════════════════",
      "                        PROJECT DEVELOPMENT TIMELINE                           ",
      "                          Project Name (Date Range)                            ",
      "                          Duration: X days Y hours Z minutes                   ",
      "═══════════════════════════════════════════════════════════════════════════════",
      "",
      "TIME                USER      OPS        PHASE          EVENTS & MILESTONES    ",
      "────────────────────────────────────────────────────────────────────────────────",
      "",
      timeline.vertical_view,
      "",
      "────────────────────────────────────────────────────────────────────────────────",
      "",
      "LEGEND:",
      "  [symbol definitions]",
      "",
      "═══════════════════════════════════════════════════════════════════════════════",
      "",
      "",
      "PHASE FLOW SUMMARY",
      "───────────────────────────────────────────────────────────────────────────────",
      "",
      phase_summary.vertical_boxes,
      "",
      "═══════════════════════════════════════════════════════════════════════════════",
      "",
      "",
      "ACTIVITY DENSITY EVOLUTION",
      "───────────────────────────────────────────────────────────────────────────────",
      "",
      activity_evolution.horizontal_bars,
      "",
      "Growth Pattern: [analysis]",
      "",
      "───────────────────────────────────────────────────────────────────────────────",
      "",
      "WORKFLOW PATTERN SEQUENCES (Top 5)",
      "───────────────────────────────────────────────────────────────────────────────",
      "",
      workflow_sequences.formatted_list,
      "",
      "═══════════════════════════════════════════════════════════════════════════════"
    ]
  },

  timeline_overview: {
    total_duration: A.timeline.duration,
    total_events: count(A.timeline.chronological_events),
    work_sessions: count(A.temporal_patterns.work_sessions),
    context_switches: count(A.timeline.phase_boundaries),

    event_breakdown: {
      user_actions: count(events, type="user_action"),
      high_level_operations: count(events, type="high_level_operation"),
      workflow_failures: count(events, type="workflow_failure")
    },

    positioned_on_timeline: anchor_to_ascii_timeline_segments
  },

  user_activity_summary: {
    file_references: frequency(A.user_actions.file_references),
    subagent_usage: frequency(A.user_actions.subagent_invocations),
    slash_commands: frequency(A.user_actions.slash_commands),
    interruption_count: count(A.user_actions.interruptions),
    correction_rate: count(A.user_actions.correction_requests) / total_user_messages,

    timeline_annotation: annotate_timeline_with([
      symbol: "●",
      positions: map_events_to_timeline_positions(user_actions),
      legend_entry: "● User Actions"
    ])
  },

  high_level_operations: {
    subagent_invocations: {
      total: count(A.operations.subagent_ops),
      by_type: distribution(A.operations.subagent_ops),
      success_rate: A.operations.operation_outcomes.successful / total_ops,
      interruption_rate: A.operations.operation_outcomes.interrupted / total_ops,

      timeline_annotation: annotate_timeline_with([
        symbol: "⚡",
        positions: map_events_to_timeline_positions(subagent_ops),
        legend_entry: "⚡ Subagent Launches"
      ])
    },

    slash_command_usage: {
      total: count(A.operations.slash_ops),
      commands_used: unique(A.operations.slash_ops),
      failure_count: count(A.operations.slash_ops, status="error"),

      timeline_annotation: annotate_timeline_with([
        symbol: "/",
        positions: map_events_to_timeline_positions(slash_ops),
        legend_entry: "/ Slash Commands"
      ])
    },

    mcp_query_activity: {
      total: count(A.operations.mcp_queries),
      query_types: distribution(A.operations.mcp_queries),
      error_rate: count(A.operations.mcp_queries, status="error") / total_mcp,

      timeline_annotation: annotate_timeline_with([
        symbol: "◆",
        positions: map_events_to_timeline_positions(mcp_queries),
        legend_entry: "◆ MCP Queries"
      ])
    }
  },

  workflow_failures: {
    build_failures: {
      count: count(A.workflow_events.build_failures),
      examples: sample(A.workflow_events.build_failures, limit=3),

      timeline_annotation: annotate_timeline_with([
        symbol: "✗",
        positions: map_events_to_timeline_positions(build_failures),
        legend_entry: "✗ Build Failures"
      ])
    },

    test_failures: {
      count: count(A.workflow_events.test_failures),
      examples: sample(A.workflow_events.test_failures, limit=3),

      timeline_annotation: annotate_timeline_with([
        symbol: "⊗",
        positions: map_events_to_timeline_positions(test_failures),
        legend_entry: "⊗ Test Failures"
      ])
    },

    subagent_interruptions: {
      count: count(A.workflow_events.subagent_interruptions),
      which_agents: extract_agent_types(A.workflow_events.subagent_interruptions),

      timeline_annotation: annotate_timeline_with([
        symbol: "⊘",
        positions: map_events_to_timeline_positions(interruptions),
        legend_entry: "⊘ Interruptions"
      ])
    },

    mcp_errors: {
      count: count(A.workflow_events.mcp_errors),
      error_types: classify_errors(A.workflow_events.mcp_errors),

      timeline_annotation: render_error_density_heatmap([
        positions: map_events_to_timeline_positions(mcp_errors),
        density_symbol: "▓"
      ])
    },

    recovery_success: {
      recovered: count(A.workflow_events.recovery_cycles),
      abandoned: count(A.workflow_events.abandoned_tasks),
      avg_recovery_time: avg(A.latency.error_recovery_time),

      timeline_annotation: highlight_recovery_cycles_on_timeline
    }
  },

  latency_analysis: {
    operation_durations: {
      subagent_avg: avg(A.latency.user_to_response.subagent_launch),
      slash_command_avg: avg(A.latency.user_to_response.slash_command),
      mcp_query_avg: avg(A.latency.user_to_response.mcp_query),
      build_avg: avg(A.latency.user_to_response.build_operation)
    },

    slowest_operations: top_n(A.latency.operation_durations.outliers, n=5),

    error_impact: {
      error_durations: avg(A.latency.operation_durations.by_outcome.error),
      success_durations: avg(A.latency.operation_durations.by_outcome.success),
      error_overhead: ratio(error_durations, success_durations)
    }
  },

  temporal_patterns: {
    work_sessions: {
      count: count(A.temporal_patterns.work_sessions),
      avg_duration: avg(session.duration),
      longest_session: max(session.duration)
    },

    activity_bursts: {
      count: count(A.temporal_patterns.activity_bursts),
      characteristics: describe(A.temporal_patterns.activity_bursts)
    },

    blocked_periods: {
      count: count(A.temporal_patterns.blocked_periods),
      total_blocked_time: sum(period.duration),
      causes: identify_blocking_causes(A.temporal_patterns.blocked_periods)
    },

    productive_flows: {
      count: count(A.temporal_patterns.productive_flows),
      avg_flow_duration: avg(flow.duration),
      flow_characteristics: describe(A.temporal_patterns.productive_flows)
    }
  },

  critical_moments: {
    breakthroughs: {
      count: count(A.critical_moments.breakthrough_points),
      details: describe(A.critical_moments.breakthrough_points),

      timeline_annotation: annotate_timeline_with([
        symbol: "★",
        positions: map_events_to_timeline_positions(breakthrough_points),
        legend_entry: "★ Breakthroughs"
      ])
    },

    frustration_signals: {
      count: count(A.critical_moments.frustration_signals),
      severity: classify_severity(A.critical_moments.frustration_signals),

      timeline_annotation: render_frustration_heatmap([
        positions: map_events_to_timeline_positions(frustration_signals),
        intensity_levels: [low="░", medium="▒", high="▓"]
      ])
    },

    workflow_transitions: {
      count: count(A.critical_moments.workflow_transitions),
      sequence: visualize_transitions(A.critical_moments.workflow_transitions),

      timeline_annotation: annotate_timeline_with([
        symbol: "║",
        positions: map_events_to_timeline_positions(workflow_transitions),
        legend_entry: "║ Phase Boundaries"
      ])
    },

    milestones: {
      count: count(A.critical_moments.milestone_events),
      achievements: list(A.critical_moments.milestone_events),

      timeline_annotation: annotate_timeline_with([
        symbol: "★",
        positions: map_events_to_timeline_positions(milestone_events),
        legend_entry: "★ Milestones"
      ])
    }
  },

  chronological_narrative: {
    session_start: A.timeline.chronological_events[0],
    key_events: select_significant_events(A.timeline.chronological_events, threshold=0.8),
    phase_summaries: summarize_each_phase(A.timeline.phase_boundaries),
    session_end: A.timeline.chronological_events[-1],

    timeline_annotation: {
      render_narrative_as_ascii_flow: true,
      connect_events_with_arrows: "→",
      annotate_time_elapsed: true,
      highlight_causal_chains: render_causal_chain_graph
    }
  }
} where ¬execute(recommendations)

ascii_art_examples:
  example_timeline_output = """
  ═══════════════════════════════════════════════════════════════════════════════
                          PROJECT DEVELOPMENT TIMELINE
                            Project Name (Date Range)
                            Duration: X days Y hours Z minutes
  ═══════════════════════════════════════════════════════════════════════════════

  TIME                USER      OPS        PHASE          EVENTS & MILESTONES
  ────────────────────────────────────────────────────────────────────────────────

  Day 1               ║                    ║
  00:00 ────────────  ●         ⚡         ║ Setup        Session start
  01:30               ⊘                    ║              Interrupted
  02:00               ●                    ║              Recovery
                      │                    ║
  06:00-12:00         ░░                   ║              Low activity
                      │                    ║
                      │                    ║
  Day 2               ║                    ║────────────  PHASE BOUNDARY
  00:00-06:00         ●●        ⚡⚡/       ║ Phase N      Development begins
                      ●●        ◆◆         ║              Planning
                      │         │          ║
  12:00               ●●●       ⚡⚡⚡       ║              ★ Milestone:
                      ●●        ◆◆◆        ║              Feature complete
  14:00-23:00         ●●●       ⚡⚡        ║              Implementation peak
                      ●●        ◆◆         ║
                      ═════════════════════              Productive flow
                      │                    ║
  Day 3               ║                    ║────────────  PHASE BOUNDARY
  00:00-06:00         ●●●       ⚡⚡        ║ Testing      Continuous work
                      ●●        ◆◆◆        ║
  08:00               /                    ║              /slash-command
                      │         │          ║
  12:00               ●●●●      ⚡⚡        ║              ★ Milestone:
                      ●●●       ◆◆◆◆       ║              Tests passing
  14:00-23:00         ●●●●      ⚡⚡        ║              Verification
                      ●●        ◆◆◆        ║
                      │         ▓▓▓▓       ║              High MCP activity
                      ═════════════════════              Flow continues
                      │                    ║
  Day 4 NOW ────────  ●         /         ║              Current analysis
                      │         ██████     ║              Peak activity

  ────────────────────────────────────────────────────────────────────────────────

  LEGEND:
    ● User Message (1-4 msgs/hour)    ⚡ Subagent Launch    / Slash Command
    ◆ MCP Query Activity              ⊘ Interruption        ★ Milestone
    ║ Phase Boundary                  ═ Productive Flow     █ Peak Activity
    ░ Low   ▒ Medium   ▓ High   █ Peak (activity density)

  ═══════════════════════════════════════════════════════════════════════════════


  PHASE FLOW SUMMARY
  ───────────────────────────────────────────────────────────────────────────────

    Phase Setup (Day 1, ~24h)
    ┌─────────────────────────────────────┐
    │ Activity: ⊘●  (interrupted start)   │
    │ Duration: 24 hours                  │
    │ Focus:    Project initialization    │
    │ Outcome:  Context recovery          │
    └─────────────────────────────────────┘
                      ↓
    Phase N Development (Day 2, ~24h)
    ┌─────────────────────────────────────┐
    │ Activity: ⚡⚡⚡ ◆◆◆ /               │
    │ Duration: 24 hours                  │
    │ Focus:    Core implementation       │
    │ Outcome:  ★ Feature complete        │
    └─────────────────────────────────────┘
                      ↓
    Phase M Testing (Day 3, ~20h)
    ┌─────────────────────────────────────┐
    │ Activity: ⚡ ◆◆◆◆ /                 │
    │ Duration: 20 hours                  │
    │ Focus:    Verification, integration │
    │ Outcome:  ★ Tests passing           │
    └─────────────────────────────────────┘

  ═══════════════════════════════════════════════════════════════════════════════


  ACTIVITY DENSITY EVOLUTION
  ───────────────────────────────────────────────────────────────────────────────

  MCP Query Frequency (queries per 6-hour block):

    Day 1 AM    ░░░           ~5 queries
    Day 1 PM    ░░            ~3 queries
    Day 2 AM    ▒▒▒▒          ~15 queries
    Day 2 PM    ▒▒▒▒▒▒        ~20 queries
    Day 3 AM    ▓▓▓▓▓▓▓       ~35 queries
    Day 3 PM    ▓▓▓▓▓▓▓▓      ~40 queries
    Day 4 AM    ████████████  ~70 queries

  Growth Pattern: Exponential / Linear / Steady

  ───────────────────────────────────────────────────────────────────────────────

  WORKFLOW PATTERN SEQUENCES (Top 5)
  ───────────────────────────────────────────────────────────────────────────────

    1. query_user_messages → query_tools
       ──────────────────────────────────
       Occurrences: 30 times
       Span:        5,847 minutes (~97h)
       Purpose:     User intent → tool correlation

    2. get_session_stats → query_user_messages
       ──────────────────────────────────────────
       Occurrences: 25 times
       Span:        6,926 minutes (~115h)
       Purpose:     Overview → detail drill-down

  ═══════════════════════════════════════════════════════════════════════════════
  """

implementation_notes:
- prioritize workflow-level events over low-level tool calls
- filter builtin tool errors (Read, Write, Edit, Bash) unless part of workflow failure
- use temporal proximity to link user actions with system responses
- calculate latency by matching tool invocation timestamps with completion timestamps
- detect causal chains: user_message → operation → outcome → user_reaction
- identify transitive tool calls: slash/subagent → mcp → builtin tools
- focus on user-facing operations: direct @ refs, /commands, @agent- invocations
- use semantic analysis to understand user intent and satisfaction
- detect context switches via topic changes, /clear, or long gaps
- measure productivity via completion rates, not just activity volume

ascii_visualization_requirements:
- ALL output MUST start with ASCII art timeline visualization
- Timeline orientation: VERTICAL (time flows from top to bottom)
- Timeline width: 80 columns (terminal-friendly)
- Timeline structure: Multi-column layout (TIME | USER | OPS | PHASE | EVENTS)
- Use Unicode box-drawing characters: ─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼ ═ ║ ╔ ╗ ╚ ╝
- Activity density symbols: ░ (low) → ▒ (medium) → ▓ (high) → █ (peak)
- Phase flow diagram: vertical boxes with downward arrows (↓)
- Time markers: precise alignment in leftmost column
- Event symbols aligned in respective columns (USER, OPS, PHASE, EVENTS)
- Legend: always include after timeline visualization
- All subsequent analysis sections reference timeline positions
- Use arrow symbols to connect related events: → ↓ ⇒ ↦
- Annotate key moments directly on timeline with symbols (★ ⊘ ║ ═)
- Vertical continuity lines: use │ for ongoing periods

constraints:
- temporal_accuracy: preserve chronological order, calculate precise durations
- workflow_focused: business-level events > implementation details
- user_centric: analyze from user's perspective, not system internals
- causality_aware: link events via temporal and semantic relationships
- privacy_preserving: aggregate statistics, no sensitive content exposure
- actionable_insights: ∀pattern → ∃improvement_opportunity
- evidence_based: ∀conclusion → ∃supporting_data
- comprehensive: cover user actions, operations, failures, latency, patterns, critical moments
- semantic_understanding: use LLM to understand intent beyond keywords
- filter_noise: exclude transitive tool calls unless directly relevant
- visualization_first: ASCII art timeline MUST appear before any textual analysis
- reference_positions: all analysis sections MUST reference specific timeline positions

output_structure:
1. ASCII Art Timeline (VERTICAL, multi-column: TIME | USER | OPS | PHASE | EVENTS)
2. Legend (symbol definitions, immediately after timeline)
3. Phase Flow Summary (vertical boxes with downward arrows)
4. Activity Density Evolution (horizontal bars showing query frequency growth)
5. Workflow Pattern Sequences (top recurring patterns)
6. Timeline Overview (with timeline position references)
7. User Activity Summary (annotated with timeline symbols)
8. High-Level Operations (annotated with timeline symbols)
9. Workflow Failures (annotated with timeline symbols)
10. Latency Analysis (optional, if data available)
11. Temporal Patterns (with timeline segment references)
12. Critical Moments (annotated with timeline symbols)
13. Chronological Narrative (flowing from timeline visualization)

presentation_style:
- Visual-first: reader should understand project trajectory from ASCII art alone
- Symbol-driven: use consistent symbols throughout (● ⚡ / ◆ ✗ ⊗ ⊘ ║ ★ ▓ ═)
- Vertically organized: time flows downward, events align precisely in columns
- Column alignment: strict left-to-right structure (TIME | USER | OPS | PHASE | EVENTS)
- Compact: fit timeline in 80 columns for terminal viewing
- Self-documenting: legend always present, symbols always explained
- Narrative flow: ASCII visualization → phase summary → activity evolution → detailed analysis → insights
