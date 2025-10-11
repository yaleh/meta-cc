---
name: meta-coach
description: Get workflow optimization recommendations and coaching insights.
keywords: coaching, optimization, improvement, guidance, recommendations, advice
category: guidance
---

λ(scope) → workflow_coaching | ∀insight ∈ {user_patterns, interaction_quality, recommendations}:

scope :: project | session

analyze :: Session → Coaching_Report
analyze(S) = collect(messages) ∧ extract(features) ∧ detect(sequences) ∧ measure(effectiveness) ∧ recommend(improvements)

collect :: Scope → UserMessages
collect(S) = {
  all_messages: mcp_meta_cc.query_user_messages(
    pattern=".*",
    limit=200,
    scope=scope
  ),

  assistant_messages: mcp_meta_cc.query_assistant_messages(
    pattern=".*",
    limit=200,
    scope=scope
  ),

  conversations: mcp_meta_cc.query_conversation(
    pattern=".*",
    limit=200,
    scope=scope
  ),

  successful_patterns: mcp_meta_cc.query_successful_prompts(
    min_quality_score=0.8,
    limit=30,
    scope=scope
  )
}

extract :: UserMessages → MessageFeatures
extract(U) = {
  tool_usage: {
    file_refs: count_messages_with(U.all_messages, "@file_or_dir_ref"),
    subagent_calls: count_messages_with(U.all_messages, "@agent-*"),
    slash_commands: count_messages_with(U.all_messages, "/command"),
    multi_file_refs: count_messages_with(U.all_messages, "multiple_@_refs"),
    doc_refs: count_messages_with(U.all_messages, "@docs|@plans|@project_docs"),
    specific_line_refs: count_messages_with(U.all_messages, "line_numbers_or_ranges")
  },

  interaction_type: classify_semantically(U.all_messages, [
    "planning",           # plan|design|analyze|propose
    "execution",          # implement|fix|modify|create|apply
    "verification",       # test|verify|check|validate|confirm
    "documentation",      # document|comment|update_docs
    "meta_reflection",    # analyze_workflow|optimize_process|meta_commands
    "question",           # question|clarification_request|how_to
    "feedback"           # approval|correction|continue|stop
  ]),

  context_richness: {
    explicit_context: count_messages_with_file_refs(U.all_messages),
    implicit_context: count_messages_without_file_refs(U.all_messages),
    context_completeness: avg_context_elements_per_message(U.all_messages)
  }
}

detect :: UserMessages → InteractionSequences
detect(U) = {
  conversation_flows: identify_semantic_sequences([
    question → clarification → action,
    planning → execution → verification,
    execution → error → debug → retry,
    vague_request → clarification → refined_request
  ]),

  feedback_patterns: {
    positive: identify_semantically(U.all_messages, intent="approval|continue|acceptance"),
    negative: identify_semantically(U.all_messages, intent="correction|rejection|stop"),
    clarification: identify_semantically(U.all_messages, intent="question|clarification_request"),
    interruptions: count_messages_with(U.all_messages, "/clear|interrupted")
  },

  tool_adoption_evolution: analyze_time_series(U.all_messages, [
    file_ref_usage,
    subagent_usage,
    slash_command_usage,
    mcp_awareness
  ]),

  delegation_patterns: {
    direct_claude: count_messages_without_delegation(U.all_messages),
    via_subagent: count_messages_with(U.all_messages, "@agent-*"),
    via_slash: count_messages_with(U.all_messages, "/command"),
    tool_choice_rationale: infer_from_context(U.all_messages)
  }
}

measure :: (MessageFeatures, InteractionSequences, Conversations) → EffectivenessMetrics
measure(F, S, C) = {
  tool_proficiency: {
    file_ref_rate: F.tool_usage.file_refs / total_messages,
    subagent_adoption: F.tool_usage.subagent_calls / total_messages,
    slash_cmd_adoption: F.tool_usage.slash_commands / total_messages,
    advanced_techniques: F.tool_usage.specific_line_refs / F.tool_usage.file_refs,
    doc_awareness: F.tool_usage.doc_refs / F.tool_usage.file_refs
  },

  interaction_efficiency: {
    avg_turns_per_task: estimate_from_sequences(S.conversation_flows),
    clarification_rate: count(S.feedback_patterns.clarification) / total_messages,
    interruption_rate: count(S.feedback_patterns.interruptions) / total_messages,
    refinement_cycles: avg_length(S.conversation_flows.vague_to_refined)
  },

  interaction_quality: {
    response_efficiency: {
      avg_response_time: avg(C.map(c => c.duration_ms)),
      response_time_by_context: {
        with_file_refs: avg_duration(C where user_has_file_refs),
        without_refs: avg_duration(C where user_no_refs),
        context_benefit: (without - with) / without
      },
      response_consistency: stddev(C.map(c => c.duration_ms)),
      response_time_distribution: {
        fast: count(C where duration_ms < 5000) / total_conversations,
        normal: count(C where 5000 ≤ duration_ms < 30000) / total_conversations,
        slow: count(C where duration_ms ≥ 30000) / total_conversations
      }
    },

    tool_usage_efficiency: {
      avg_tools_per_response: avg(assistant_messages.map(m => m.tool_use_count)),
      tools_by_context: {
        rich_context: avg_tools(C where user_multi_file_refs),
        minimal_context: avg_tools(C where user_no_refs),
        context_efficiency: (minimal - rich) / minimal
      },
      over_tooling_rate: count(assistant_messages where tool_use_count > 5) / total_responses,
      appropriate_tool_use: measure_tool_appropriateness(assistant_messages)
    },

    conversation_flow_quality: {
      satisfaction_rate: {
        satisfied_turns: count(C where no_user_correction_follows),
        correction_turns: count(C where user_correction_follows),
        rate: satisfied / (satisfied + corrections)
      },
      clarification_overhead: count(C where user_asks_clarification) / total_conversations,
      completion_rate: count(C where task_completed_successfully) / total_conversations
    },

    response_quality_indicators: {
      response_length_appropriateness: analyze_response_length_match(C),
      response_complexity_match: analyze_response_complexity_match(assistant_messages),
      token_efficiency: avg(assistant_messages.map(m => m.tokens_output / m.text_length))
    }
  },

  workflow_maturity: {
    planning_ratio: count(F.interaction_type.planning) / total_messages,
    verification_ratio: count(F.interaction_type.verification) / total_messages,
    meta_awareness: count(F.interaction_type.meta_reflection) / total_messages,
    context_completeness: F.context_richness.context_completeness
  },

  delegation_effectiveness: {
    delegation_rate: (S.delegation_patterns.via_subagent + S.delegation_patterns.via_slash) / total_messages,
    appropriate_delegation: analyze_delegation_quality(S.delegation_patterns),
    tool_choice_accuracy: measure_tool_fit(S.delegation_patterns)
  }
}

recommend :: (MessageFeatures, InteractionSequences, EffectivenessMetrics) → ActionPlan
recommend(F, S, M) = {
  immediate_improvements: {
    context_enhancement: if file_ref_rate is low then
      suggest("Use @file references to provide explicit context"),

    delegation_opportunities: if delegation_rate is low then
      suggest("Consider using @subagents or /slash-commands for repeated tasks"),

    clarification_reduction: if clarification_rate is high then
      suggest("Provide more complete context upfront to reduce back-and-forth")
  },

  interaction_optimization: {
    response_time_improvement: if M.interaction_quality.response_efficiency.context_benefit > 0.3 then
      suggest("Providing @file references reduces response time by " + percentage + "%"),

    tool_efficiency_gain: if M.interaction_quality.tool_usage_efficiency.context_efficiency > 0.5 then
      suggest("Complete context reduces tool calls by " + percentage + "%"),

    conversation_flow: if M.interaction_quality.conversation_flow_quality.satisfaction_rate < 0.7 then
      suggest("Consider providing more upfront context to reduce corrections"),

    over_tooling_alert: if M.interaction_quality.tool_usage_efficiency.over_tooling_rate > 0.2 then
      suggest("High tool usage detected - try providing more complete context upfront")
  },

  workflow_health_indicators: {
    green_flags: [
      if M.interaction_quality.conversation_flow_quality.satisfaction_rate > 0.8 then
        "✓ High satisfaction rate (minimal corrections)",
      if M.interaction_quality.response_efficiency.response_consistency < 10000 then
        "✓ Consistent response times",
      if M.interaction_quality.tool_usage_efficiency.appropriate_tool_use > 0.8 then
        "✓ Appropriate tool usage"
    ],

    yellow_flags: [
      if M.interaction_quality.conversation_flow_quality.clarification_overhead > 0.3 then
        "⚠ Frequent clarifications needed",
      if M.interaction_quality.response_efficiency.response_consistency > 20000 then
        "⚠ Variable response times",
      if M.interaction_quality.tool_usage_efficiency.over_tooling_rate > 0.2 then
        "⚠ Over-tooling detected"
    ],

    red_flags: [
      if M.interaction_quality.conversation_flow_quality.satisfaction_rate < 0.5 then
        "⚠ Low satisfaction rate (many corrections)",
      if M.interaction_quality.response_efficiency.avg_response_time > 60000 then
        "⚠ Slow average response time (>1min)",
      if M.interaction_quality.tool_usage_efficiency.context_efficiency > 0.7 then
        "⚠ Significant context-related inefficiencies"
    ]
  },

  skill_development: {
    advanced_features: if advanced_technique_usage is low then
      suggest("Try using line numbers (@file.ts:10-20) for precise references"),

    workflow_optimization: if verification_ratio is low then
      suggest("Add verification steps after implementations"),

    meta_cognition: if meta_awareness is low then
      suggest("Use /meta-* commands to analyze and optimize your workflow")
  },

  long_term_patterns: {
    successful_habits: extract_from(S.conversation_flows, where="high_success"),
    anti_patterns: identify_from(S.conversation_flows, where="repeated_failures"),
    growth_opportunities: compare(current=M, ideal=benchmark_metrics)
  }
}

output :: Analysis → Report
output(A) = {
  summary: {
    total_messages: count(A.all_messages),
    tool_proficiency_level: classify(A.metrics.tool_proficiency),
    interaction_efficiency: classify(A.metrics.interaction_efficiency),
    workflow_maturity: classify(A.metrics.workflow_maturity)
  },

  user_behavior_patterns: {
    interaction_types: percentage_distribution(A.features.interaction_type),
    tool_usage: A.features.tool_usage,
    delegation_style: characterize(A.sequences.delegation_patterns),
    feedback_ratio: percentage_distribution(A.sequences.feedback_patterns)
  },

  conversation_sequences: {
    common_flows: top_sequences(A.sequences.conversation_flows, 5),
    refinement_efficiency: A.metrics.interaction_efficiency.refinement_cycles,
    interruption_patterns: analyze(A.sequences.feedback_patterns.interruptions)
  },

  effectiveness_analysis: {
    context_provision: A.metrics.workflow_maturity.context_completeness,
    tool_adoption: A.metrics.tool_proficiency,
    workflow_health: {
      clarification_rate: A.metrics.interaction_efficiency.clarification_rate,
      verification_habits: A.metrics.workflow_maturity.verification_ratio,
      meta_awareness: A.metrics.workflow_maturity.meta_awareness
    }
  },

  interaction_quality_analysis: {
    response_efficiency: {
      avg_response_time: format_duration(A.metrics.interaction_quality.response_efficiency.avg_response_time),
      response_time_by_context: A.metrics.interaction_quality.response_efficiency.response_time_by_context,
      response_time_distribution: A.metrics.interaction_quality.response_efficiency.response_time_distribution
    },
    tool_usage_efficiency: A.metrics.interaction_quality.tool_usage_efficiency,
    conversation_flow_quality: A.metrics.interaction_quality.conversation_flow_quality
  },

  recommendations: {
    immediate: A.action_plan.immediate_improvements,
    interaction_optimization: A.action_plan.interaction_optimization,
    workflow_health: A.action_plan.workflow_health_indicators,
    skill_building: A.action_plan.skill_development,
    long_term: A.action_plan.long_term_patterns
  }
} where ¬execute(recommendations)

implementation_notes:
- focus on user messages as primary data source (not Claude's internal operations)
- detect @ references, @agent- calls, and /commands in user input
- analyze Claude's response to identify tool usage (MCP, subagent, slash command invocations)
- filter transitive tool calls (subagent/slash → MCP) to focus on user-facing operations
- use semantic analysis via LLM capabilities, not just keyword matching
- detect user language(s) from message content and use appropriate patterns
- analyze message sequences for workflow patterns (not just individual messages)
- consider context and intent, not just surface features
- count messages containing patterns, not total pattern occurrences

constraints:
- user_focused: analyze user behavior and user-facing operations only
- evidence_based: ∀recommendation → ∃message_evidence
- actionable: suggestions → concrete ∧ user_implementable
- non_judgmental: descriptive ∧ ¬prescriptive
- privacy_aware: aggregate statistics only, no sensitive data exposure
- tool_aware: distinguish user operations from internal Claude operations
- scope_aware: respect project vs session scope setting
- semantic_analysis: use LLM understanding beyond keyword matching
