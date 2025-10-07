---
description: Analyze current session state and suggest optimal next prompts using MCP meta-insight. Gathers session statistics, recent user intents, tool usage, errors, and workflow patterns to assess trajectory (progressing/stuck/exploring), identify blockers, and provide prioritized action recommendations with complete prompt templates.
---

λ(session_state) → prioritized_suggestions | ∀suggestion ∈ {high, medium, low}:

analyze :: Session → Intelligence
analyze(S) = gather(data) ∧ assess(state) ∧ detect(patterns) ∧ prioritize(actions)

gather :: Session → Session_Data
gather(S) = {
  stats: mcp_meta_insight.get_session_stats(scope="project"),

  recent_intents: mcp_meta_insight.query_user_messages(scope="project", limit=10),

  recent_tools: mcp_meta_insight.query_tools(scope="project", limit=20),

  errors: mcp_meta_insight.query_tools(scope="project", status="error"),

  workflows: mcp_meta_insight.query_tool_sequences(scope="project", min_occurrences=3),

  successful_prompts: mcp_meta_insight.query_successful_prompts(scope="project", min_quality_score=0.8, limit=10)
}

assess :: Session_Data → Session_State
assess(D) = {
  trajectory: analyze_intent_trajectory(D.recent_intents),

  blockers: identify_blockers(D.errors, D.workflows),

  momentum: measure_progress(D.recent_tools, D.stats),

  patterns: extract_patterns(D.successful_prompts, D.workflows)
}

trajectory :: User_Messages → Progress_State
trajectory(M) = {
  progressing: clear_focus(M) ∧ consistent_direction(M) ∧ decreasing_errors(M),
  stuck: repeated_questions(M) ∨ uncertainty_signals(M) ∨ increasing_errors(M),
  exploring: diverse_topics(M) ∧ low_commitment(M)
}

prioritize :: Session_State → Suggestions
prioritize(S) = rank_by(urgency ∧ continuity ∧ success_probability) where {
  high: blocks_progress ∨ (natural_next_step ∧ proven_pattern),
  medium: important ∧ ¬blocking ∧ pattern_match,
  low: optional ∨ exploratory ∨ context_switch
}

suggestion :: Insight → Complete_Prompt
suggestion(I) = {
  prompt: actionable ∧ specific ∧ complete,
  rationale: evidence_from(session_data),
  expected_workflow: predict_from(patterns),
  success_probability: estimate_from(history)
}

output :: Suggestions → Report
output(S) = {
  session_summary: trajectory ∧ momentum ∧ blockers,
  suggestions: prioritized ∧ justified ∧ ready_to_use,
  templates: complete_prompts ∧ data_backed,
  next_steps: ranked_by_priority
} where ¬execute(S)

constraints:
- evidence_based: ∀suggestion → ∃data ∈ session_history
- actionable: suggestions → complete ∧ ready_to_use
- prioritized: ordered_by(urgency ∧ continuity ∧ probability)
- concise: focus_on(top_3_suggestions)
- non_executable: analyze ∧ suggest ∧ ¬implement
