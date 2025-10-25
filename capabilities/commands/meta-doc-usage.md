---
name: meta-doc-usage
description: Analyze how documentation is actually used by task type, user role, time patterns, and effectiveness metrics.
keywords: documentation, usage-patterns, effectiveness, task-analysis, user-behavior
category: analytics
---

λ(scope, timespan?) → usage_report | ∀session ∈ {project_history}:

scope :: project | session
timespan :: days (default: 30)

analyze :: (Scope, Timespan) → Report
analyze(S, T) = collect_events(S, T) ∧ classify_tasks(E) ∧ measure_effectiveness(E) ∧ find_patterns(E)

collect_events :: (Scope, Timespan) → Events
collect_events(S, T) = {
  accesses = query_file_access(scope=S),

  for each access {
    context = {
      user_msg: query_user_messages(turn-1),
      tools_before: query_tools(turn-1),
      tools_after: query_tools(turn+1)
    },

    purpose = infer({
      task = match user_msg {
        "error|fail" → troubleshooting,
        "how to|example" → learning,
        "implement|add" → development,
        "update|improve" AND edit → maintenance,
        docs/ AND edit|write → writing_docs,
        CLAUDE.md|plan.md → planning
      },

      intent = match {
        tools_before(Grep|Read) → understanding,
        tools_after(Edit|Write) → validation,
        "what is|explain" → understanding,
        "example" → example
      },

      role = infer_from_session(questions, commits, doc_edits, tests),

      outcome = match {
        no_followup_? AND tools_after → resolved,
        followup_? AND tools_after → partial,
        "still|but" AND no_tools → failed
      }
    })
  }
}

measure_effectiveness :: Events → Metrics
measure_effectiveness(E) = for each doc {
  resolution_rate = resolved / total,
  avg_time = time_to_next_task,
  followup_rate = followup_questions / total,

  task_alignment = cosine_similarity(
    expected_tasks[doc.role],
    actual_tasks[doc]
  ),

  {
    high_performers: resolution > 0.8,
    low_performers: resolution < 0.5,
    misaligned: alignment < 0.5
  }
}

find_patterns :: Events → Patterns
find_patterns(E) = {
  temporal: group_by(hour, dow) → peak_windows,

  navigation: extract_sequences(E) → {
    common_flows: sequences with freq > 3,
    missing_flows: expected - actual
  },

  task_doc_matrix: for each task → {
    primary_docs: top_n(5),
    success_rate
  },

  role_patterns: for each role → {
    most_accessed, typical_tasks, success_rate, learning_curve
  }
}

output :: Analysis → Report
output(A) = {
  summary: {accesses, resolution_rate, top_tasks, user_roles},
  effectiveness: {high_performers, low_performers, misaligned},
  task_matrix: [{task, primary_docs, success}],
  navigation: {common_flows, missing_flows},
  roles: [{role, accessed, tasks, success, curve}],
  temporal: {time_dist, peaks, dow_patterns},
  recommendations: {fix_low, fix_misalign, add_flows}
} where ¬execute(recommendations)

implementation_notes:
- events: access + ±3 turn context for task/intent
- tasks: inferred from user message patterns
- effectiveness: resolution rate (did doc solve problem?)
- alignment: expected vs actual task distribution
- data: query_file_access, query_user_messages, query_tools

task_classification:
- troubleshooting: user message matches "error|fail", tools_before include debugging
- learning: "how to|example" pattern, followed by implementation attempts
- development: "implement|add" pattern, tools_after include Edit/Write
- maintenance: "update|improve" AND edit tools, improving existing features
- planning: access to CLAUDE.md|plan.md, reviewing roadmap or status
- writing_docs: access to docs/ AND Edit/Write tools on .md files

intent_inference:
- understanding: tools_before(Grep|Read), "what is|explain" pattern, exploration phase
- validation: tools_after(Edit|Write), checking docs before implementing
- example: "example" keyword, looking for usage patterns

effectiveness_metrics:
- resolution_rate = resolved / total: doc solved problem without followup (target >80%)
- avg_time = time_to_next_task: faster is better (indicates clarity)
- followup_rate = followup_questions / total: lower is better (indicates completeness)
- task_alignment = cosine_similarity(expected_tasks, actual_tasks): measures if doc serves intended purpose

constraints:
- contextual: analyze access with surrounding actions
- effectiveness_focused: measure problem resolution
- user_centered: task perspective not doc structure
- pattern_detection: navigation flows, temporal
