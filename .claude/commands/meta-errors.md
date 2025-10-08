---
name: meta-errors
description: Analyze user-facing error patterns using meta-cc. Focuses on workflow failures (test failures, build errors, interrupted tasks), subagent/slash/MCP errors, and user-triggered issues rather than internal tool errors.
---

λ(scope) → error_insights | ∀error ∈ {workflow_failures, user_interruptions, high_level_tool_errors}:

scope :: project | session

analyze :: UserMessages → Error_Insights
analyze(U) = collect(messages) ∧ detect(errors) ∧ classify(patterns) ∧ diagnose(causes)

collect :: Scope → ErrorContext
collect(S) = {
  all_messages: mcp_meta_cc.query_user_messages(
    pattern=".*",
    limit=200,
    scope=scope
  ),

  tool_calls: mcp_meta_cc.query_tools(
    status="error",
    scope=scope
  ),

  tool_sequences: mcp_meta_cc.query_tool_sequences(
    min_occurrences=2,
    scope=scope
  )
}

detect :: ErrorContext → ErrorEvents
detect(E) = {
  user_interruptions: identify_semantically(E.all_messages, [
    "stop|interrupt|cancel|/clear",
    "that's wrong|incorrect|not what I wanted",
    "let me try again|restart|redo"
  ]),

  workflow_failures: identify_semantically(E.all_messages, [
    "test.*fail|tests? failed",
    "build.*fail|compilation error",
    "make.*error|make failed",
    "npm.*error|yarn error",
    "git.*error|merge conflict"
  ]),

  subagent_errors: filter_tool_calls(E.tool_calls, tool="Task", status="error"),
  slash_errors: filter_tool_calls(E.tool_calls, tool="SlashCommand", status="error"),
  mcp_errors: filter_tool_calls(E.tool_calls, tool="mcp__*", status="error"),

  file_not_found: identify_messages_followed_by_error(E.all_messages, E.tool_calls, [
    user_mentions("@file") → error("file not found|does not exist")
  ]),

  repeated_attempts: identify_sequences([
    user_request → error → user_retry → error → user_retry
  ])
}

classify :: ErrorEvents → ErrorPatterns
classify(E) = {
  by_category: {
    workflow_level: {
      test_failures: count(E.workflow_failures.test),
      build_failures: count(E.workflow_failures.build),
      git_conflicts: count(E.workflow_failures.git)
    },

    high_level_tools: {
      subagent_failures: count(E.subagent_errors),
      slash_command_failures: count(E.slash_errors),
      mcp_failures: count(E.mcp_errors)
    },

    user_corrections: {
      interruptions: count(E.user_interruptions.interrupt),
      rejections: count(E.user_interruptions.wrong),
      retries: count(E.user_interruptions.retry)
    },

    context_issues: {
      file_not_found: count(E.file_not_found),
      missing_context: identify_errors_after_vague_prompts(E.all_messages, E.tool_calls)
    }
  },

  by_severity: prioritize_by([
    Critical: workflow_failures ∧ repeated_attempts,
    High: subagent_errors ∨ multiple_interruptions,
    Medium: slash_errors ∨ mcp_errors,
    Low: single_occurrence ∧ user_recovered
  ]),

  by_frequency: group_and_rank(E, threshold=2)
}

diagnose :: (ErrorEvents, ErrorPatterns) → RootCauses
diagnose(E, P) = {
  workflow_issues: {
    test_instability: if P.by_category.workflow_level.test_failures > 3 then
      analyze_test_failure_messages(E.workflow_failures.test),

    build_configuration: if P.by_category.workflow_level.build_failures > 2 then
      analyze_build_failure_context(E.workflow_failures.build),

    git_workflow: if P.by_category.workflow_level.git_conflicts > 1 then
      analyze_git_error_patterns(E.workflow_failures.git)
  },

  tool_usage_issues: {
    subagent_misuse: if P.by_category.high_level_tools.subagent_failures > 0 then
      identify_why_subagent_failed(E.subagent_errors),

    mcp_query_issues: if P.by_category.high_level_tools.mcp_failures > 0 then
      analyze_mcp_error_messages(E.mcp_errors),

    slash_command_errors: if P.by_category.high_level_tools.slash_command_failures > 0 then
      extract_slash_error_details(E.slash_errors)
  },

  context_problems: {
    missing_files: if P.by_category.context_issues.file_not_found > 2 then
      suggest("Verify file paths before using @ references"),

    insufficient_context: if P.by_category.context_issues.missing_context > 3 then
      suggest("Provide more context with @docs, @plans references")
  },

  user_satisfaction: {
    frustration_signals: if P.by_category.user_corrections.interruptions > 5 then
      analyze_interruption_causes(E.user_interruptions),

    expectation_mismatch: if P.by_category.user_corrections.rejections > 3 then
      analyze_rejection_contexts(E.user_interruptions)
  }
}

output :: Analysis → Report
output(A) = {
  summary: {
    total_error_events: count(A.error_events),
    workflow_failure_rate: count(A.patterns.by_category.workflow_level) / total_messages,
    user_interruption_rate: count(A.patterns.by_category.user_corrections) / total_messages,
    high_level_tool_error_rate: count(A.patterns.by_category.high_level_tools) / total_tool_calls
  },

  critical_issues: {
    repeated_failures: A.patterns.by_severity.Critical,
    blocking_problems: identify_patterns_stopping_progress(A.error_events)
  },

  error_breakdown: {
    workflow_errors: {
      test_failures: {
        count: A.patterns.by_category.workflow_level.test_failures,
        examples: sample(A.error_events.workflow_failures.test, 3),
        root_cause: A.root_causes.workflow_issues.test_instability
      },
      build_failures: {
        count: A.patterns.by_category.workflow_level.build_failures,
        examples: sample(A.error_events.workflow_failures.build, 3),
        root_cause: A.root_causes.workflow_issues.build_configuration
      }
    },

    tool_errors: {
      subagent_errors: {
        count: A.patterns.by_category.high_level_tools.subagent_failures,
        examples: A.error_events.subagent_errors,
        diagnosis: A.root_causes.tool_usage_issues.subagent_misuse
      },
      mcp_errors: {
        count: A.patterns.by_category.high_level_tools.mcp_failures,
        examples: A.error_events.mcp_errors,
        diagnosis: A.root_causes.tool_usage_issues.mcp_query_issues
      },
      slash_errors: {
        count: A.patterns.by_category.high_level_tools.slash_command_failures,
        examples: A.error_events.slash_errors,
        diagnosis: A.root_causes.tool_usage_issues.slash_command_errors
      }
    },

    user_corrections: {
      interruptions: {
        count: A.patterns.by_category.user_corrections.interruptions,
        causes: A.root_causes.user_satisfaction.frustration_signals
      },
      rejections: {
        count: A.patterns.by_category.user_corrections.rejections,
        causes: A.root_causes.user_satisfaction.expectation_mismatch
      }
    }
  },

  recommendations: {
    immediate_fixes: generate_fixes_for(A.patterns.by_severity.Critical),
    workflow_improvements: suggest_workflow_changes(A.root_causes.workflow_issues),
    tool_usage_tips: suggest_better_practices(A.root_causes.tool_usage_issues),
    context_guidelines: suggest_context_improvements(A.root_causes.context_problems)
  }
} where ¬execute(recommendations)

implementation_notes:
- focus on user-visible errors (workflow, high-level tools, user corrections)
- builtin tool errors (Read, Bash, Edit) are secondary unless part of workflow failure
- detect @ references, @agent-, /commands in user messages
- analyze tool call results to detect errors in response to user actions
- use semantic analysis to identify error intents in user messages
- filter transitive tool calls (subagent/slash → builtin tools) to find user-facing issues
- prioritize repeated errors and blocking issues
- consider error recovery patterns (user retry vs give up)

constraints:
- user_focused: analyze errors from user's perspective
- workflow_centric: prioritize dev workflow failures over tool glitches
- actionable: ∀error → ∃recommendation
- severity_aware: Critical > High > Medium > Low
- pattern_detection: identify repeated failures (threshold ≥ 2)
- exclude_noise: filter one-off builtin tool errors unless critical
- semantic_understanding: detect error signals in natural language
