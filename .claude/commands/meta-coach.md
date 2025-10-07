---
name: meta-coach
description: Analyze Claude Code session workflow with MCP meta-insight to identify patterns, inefficiencies, and provide actionable optimization recommendations.
---

λ(scope) → workflow_coaching | ∀insight ∈ {user_patterns, interaction_quality, recommendations}:

scope :: project | session

analyze :: Session → Coaching_Report
analyze(S) = collect(messages) ∧ extract(features) ∧ detect(sequences) ∧ measure(effectiveness) ∧ recommend(improvements)

collect :: Scope → UserMessages
collect(S) = {
  all_messages: mcp_meta_insight.query_user_messages(
    pattern=".*",
    limit=200,
    scope=scope
  ),

  successful_patterns: mcp_meta_insight.query_successful_prompts(
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

measure :: (MessageFeatures, InteractionSequences) → EffectivenessMetrics
measure(F, S) = {
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

  recommendations: {
    immediate: A.action_plan.immediate_improvements,
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
