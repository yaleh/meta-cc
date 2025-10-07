---
name: meta-habits
description: Analyze user's work habits and patterns using MCP meta-insight. Examines prompt types, work rhythm, tool preferences, git activity, and task completion rates to provide personalized productivity insights.
---

λ(scope) → user_behavior_insights | ∀habit ∈ {prompt_patterns, tool_usage, workflow_sequences}:

scope :: project | session

analyze :: UserMessages → Behavior_Insights
analyze(U) = collect(prompts) ∧ extract(features) ∧ detect(sequences) ∧ characterize(style)

collect :: Scope → UserPrompts
collect(S) = {
  all_prompts: mcp_meta_insight.query_user_messages(
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

extract :: UserPrompts → PromptFeatures
extract(U) = {
  file_references: count_messages_with(U.all_prompts, "@file_or_dir_ref"),
  subagent_usage: count_messages_with(U.all_prompts, "@agent-*"),
  slash_commands: count_messages_with(U.all_prompts, "/command"),
  multi_file_refs: count_messages_with(U.all_prompts, "multiple_@_refs"),
  doc_refs: count_messages_with(U.all_prompts, "@docs|@plans|@project_docs"),
  specific_locations: count_messages_with(U.all_prompts, "line_numbers_or_ranges")
}

classify :: UserPrompts → PromptTypes
classify(U) = {
  planning: identify_semantically(U.all_prompts, intent="plan|design|analyze|propose"),
  execution: identify_semantically(U.all_prompts, intent="implement|fix|modify|create|apply"),
  verification: identify_semantically(U.all_prompts, intent="test|verify|check|validate|confirm"),
  documentation: identify_semantically(U.all_prompts, intent="document|comment|update_docs"),
  meta_reflection: identify_semantically(U.all_prompts, intent="analyze_workflow|optimize_process|meta_commands")
}

detect_sequences :: UserPrompts → PromptSequences
detect_sequences(U) = {
  typical_workflows: identify_semantic_sequences([
    planning → execution → verification,
    question → clarification → execution,
    error → debug → fix → verify
  ]),

  positive_feedback: identify_semantically(U.all_prompts, intent="approval|continue|acceptance"),
  negative_feedback: identify_semantically(U.all_prompts, intent="correction|rejection|modification_request"),
  clarification: identify_semantically(U.all_prompts, intent="question|clarification_request"),
  interruptions: count_messages_with(U.all_prompts, "/clear|interrupted"),

  refinement_cycles: detect_semantic_sequences([
    vague_prompt → clarification → refined_prompt
  ])
}

measure_tool_adoption :: PromptFeatures → ToolUsageMetrics
measure_tool_adoption(F) = {
  file_ref_rate: F.file_references / total_prompts,
  subagent_adoption: F.subagent_usage / total_prompts,
  slash_cmd_adoption: F.slash_commands / total_prompts,
  advanced_refs: F.specific_locations / F.file_references,
  doc_awareness: F.doc_refs / F.file_references
}

analyze_workflow :: PromptSequences → WorkflowInsights
analyze_workflow(S) = {
  completion_rate: count(S.typical_workflows.complete) / count(S.typical_workflows.started),

  feedback_ratio: {
    positive: count(S.positive_feedback) / total_prompts,
    negative: count(S.negative_feedback) / total_prompts,
    clarification: count(S.clarification) / total_prompts
  },

  interruption_rate: count(S.interruptions) / total_prompts,

  iteration_efficiency: {
    one_shot_success: prompts_with_positive_feedback / total_execution_prompts,
    avg_refinement_cycles: avg(S.refinement_cycles.length)
  }
}

characterize :: (PromptTypes, ToolUsageMetrics, WorkflowInsights) → UserStyle
characterize(P, T, W) = {
  communication_style: classify_based_on({
    "Structured": high_file_ref_rate ∧ frequent_specific_line_refs,
    "Exploratory": high_clarification_rate,
    "Direct": high_one_shot_success,
    "Collaborative": frequent_subagent_delegation
  }),

  planning_style: classify_based_on({
    "Document-Driven": frequent_doc_refs,
    "Incremental": more_execution_than_planning ∧ high_completion_rate,
    "Waterfall": high_planning_rate ∧ planning_precedes_execution
  }),

  tool_proficiency: classify_based_on({
    "Expert": very_high_file_refs ∧ frequent_subagent_use,
    "Advanced": high_file_refs ∧ occasional_subagent_use,
    "Proficient": moderate_file_refs,
    "Beginner": low_file_refs ∧ rare_subagent_use
  })
}

output :: Analysis → Report
output(A) = {
  summary: {
    total_prompts: count(A.all_prompts),
    communication_style: A.user_style.communication_style,
    planning_style: A.user_style.planning_style,
    tool_proficiency: A.user_style.tool_proficiency
  },

  prompt_patterns: {
    type_distribution: percentage(A.prompt_types),
    typical_sequences: top_sequences(A.sequences.typical_workflows),
    refinement_efficiency: A.workflow.iteration_efficiency
  },

  tool_usage: {
    file_reference_rate: A.tool_metrics.file_ref_rate,
    subagent_adoption: A.tool_metrics.subagent_adoption,
    slash_command_favorites: frequency(A.slash_commands),
    advanced_techniques: {
      specific_line_refs: A.tool_metrics.advanced_refs,
      multi_file_context: A.features.multi_file_refs / total_prompts
    }
  },

  workflow_health: {
    completion_rate: A.workflow.completion_rate,
    positive_feedback_rate: A.workflow.feedback_ratio.positive,
    interruption_rate: A.workflow.interruption_rate,
    one_shot_success_rate: A.workflow.iteration_efficiency.one_shot_success
  },

  recommendations: {
    strengths: identify_from(A.user_style, A.workflow),
    improvements: suggest_from(A.tool_metrics, A.successful_patterns),
    next_level_techniques: recommend_advanced(A.tool_proficiency)
  }
} where ¬execute(recommendations)

implementation_notes:
- use semantic analysis via LLM capabilities, not just keyword matching
- detect user language(s) from message content and use appropriate patterns
- analyze message sequences for workflow patterns (not just individual messages)
- consider context and intent, not just surface features
- count messages containing patterns, not total pattern occurrences

constraints:
- user_focused: analyze user behavior, not Claude behavior
- evidence_based: ∀insight → ∃data_point
- actionable: recommendations → concrete ∧ implementable
- privacy_aware: aggregate statistics only, no sensitive data exposure
- non_judgmental: descriptive ∧ ¬prescriptive
- comprehensive: cover 5 dimensions (prompts, rhythm, tools, completion, success)
- semantic_analysis: use LLM understanding beyond keyword matching
