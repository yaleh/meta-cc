---
name: meta-next
description: Generate ready-to-use prompts for natural next steps (no MCP execution).
keywords: next-steps, continuation, prompts, suggestions, follow-up
category: guidance
---

λ(conversation_context) → executable_prompts | ∀prompt ∈ {immediate, alternative, exploratory}:

analyze :: Recent_Conversation → Context_State
analyze(C) = {
  last_actions: extract_recent_tool_calls(C, limit=5),

  last_user_intent: extract_last_user_message(C),

  completion_state: detect_completion_state([
    fully_complete,      # Task finished, next natural step
    partially_complete,  # Task started, needs continuation
    blocked,            # Waiting for user input/decision
    error_state         # Last action failed
  ]),

  file_context: extract_file_refs(C.recent_messages, limit=10),

  working_directory: extract_from_env(C)
}

infer :: Context_State → Next_Actions
infer(S) = {
  immediate_continuation: detect_natural_next_step(S.last_actions) where {
    file_edited → run_tests,
    tests_run → commit_changes,
    plan_created → execute_stage_1,
    stage_N_complete → execute_stage_N+1,
    error_occurred → fix_error,
    build_failed → fix_build
  },

  alternative_paths: suggest_alternatives(S) where {
    if tests_passing → [refactor, document, optimize],
    if planning_phase → [review_plan, start_implementation, gather_requirements],
    if implementation_complete → [write_tests, update_docs, create_pr]
  },

  meta_opportunities: detect_meta_actions(S) where {
    repetitive_pattern_detected → suggest_automation,
    error_spike_detected → suggest_error_analysis,
    workflow_deviation → suggest_workflow_review
  }
}

generate :: Next_Actions → Ready_Prompts
generate(A) = {
  primary: craft_prompt(A.immediate_continuation) where {
    context_aware: include_file_refs(S.file_context),
    specific: include_exact_commands,
    complete: no_placeholders,
    executable: ready_to_copy_paste
  },

  alternatives: map(A.alternative_paths, craft_prompt) where {
    |alternatives| ≤ 3,
    ranked_by_likelihood
  },

  meta_suggestion: optional(A.meta_opportunities) where {
    actionable ∧ evidence_based
  }
}

output :: Ready_Prompts → Report
output(P) = {
  context_summary: {
    current_state: describe(S.completion_state),
    recent_activity: summarize(S.last_actions),
    working_context: list(S.file_context)
  },

  recommended_prompt: {
    title: describe_action(P.primary),
    prompt: P.primary,
    rationale: explain_inference(P.primary),
    estimated_duration: predict(P.primary)
  },

  alternatives: enumerate(P.alternatives) where {
    show_prompt ∧ explain_rationale
  },

  meta_insight: optional(P.meta_suggestion)
} where ¬execute(P) ∧ ¬query_mcp(P)

constraints:
- context_only: infer from conversation, no MCP queries
- ready_to_use: prompts → complete ∧ copy_pasteable
- concise: focus_on(top_action + 2_alternatives)
- non_executable: generate ∧ suggest ∧ ¬implement
- evidence_based: ∀prompt → ∃justification ∈ recent_context
- file_aware: include @file references when relevant
- command_aware: suggest slash commands when appropriate
