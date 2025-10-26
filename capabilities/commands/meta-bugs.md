---
name: meta-bugs
description: Analyze project-level bugs, workflow failures, and fix patterns using meta-cognitive analysis.
keywords: bugs, defects, issues, failures, root-cause, fix-patterns
category: diagnostics
---

λ(scope) → bug_insights | ∀bug ∈ {workflow_failures, user_corrections, fix_patterns}:

scope :: project | session

## Phase 1: Data Collection

collect :: Scope → BugContext
collect(S) = {
  workflow_signals: mcp_meta_cc.query_user_messages(
    pattern="(test.*fail|build.*fail|make.*error|lint.*error|git.*error|git.*conflict|compilation.*error|tests?.*failed)",
    scope=scope
  ),

  # query_conversation does not exist - use query_user_messages to detect resolution signals
  resolution_signals: mcp_meta_cc.query_user_messages(
    pattern="(test.*pass|build.*success|all.*pass|fixed|resolved|working|problem.*solved)",
    scope=scope
  ),

  correction_signals: mcp_meta_cc.query_user_messages(
    pattern="(stop|interrupt|cancel|clear|wrong|incorrect|not what|mistake|try again|redo|restart)",
    scope=scope
  ),

  error_context: mcp_meta_cc.query_tools(
    status="error",
    scope=scope
  ),

  # query_tool_sequences does not exist - not implemented
  # tool_patterns: null,

  # get_session_stats does not exist - use query_timestamps or query_summaries for overview
  # session_stats: null
}

## Phase 2: Pattern Detection

detect :: BugContext → BugEvents
detect(B) = {
  workflow_failures: classify_workflow_failures(B.workflow_signals) {
    test_failures: filter_by_pattern(B.workflow_signals, "test.*fail|tests?.*failed"),
    build_failures: filter_by_pattern(B.workflow_signals, "build.*fail|make.*error|compilation.*error"),
    lint_failures: filter_by_pattern(B.workflow_signals, "lint.*fail|golangci-lint.*error|gofmt.*error"),
    git_failures: filter_by_pattern(B.workflow_signals, "git.*error|merge.*conflict|rebase.*fail")
  },

  user_corrections: classify_user_corrections(B.correction_signals) {
    interrupts: filter_by_pattern(B.correction_signals, "stop|interrupt|cancel|clear"),
    rejects: filter_by_pattern(B.correction_signals, "wrong|incorrect|not what|mistake"),
    retries: filter_by_pattern(B.correction_signals, "try again|redo|restart")
  },

  fix_cycles: identify_fix_cycles(B.workflow_signals, B.resolution_signals) {
    // Match workflow failures with subsequent resolutions
    // Calculate turns between failure and resolution
    // Identify repeated failure patterns
  }
}

## Phase 3: Pattern Analysis

analyze :: BugEvents → BugPatterns
analyze(E) = {
  workflow_breakdown: {
    test: {
      count: count(E.workflow_failures.test_failures),
      rate: count(test_failures) / B.session_stats.total_turns,
      examples: sample(E.workflow_failures.test_failures, n=3)
    },
    build: {
      count: count(E.workflow_failures.build_failures),
      rate: count(build_failures) / B.session_stats.total_turns,
      examples: sample(E.workflow_failures.build_failures, n=3)
    },
    lint: {
      count: count(E.workflow_failures.lint_failures),
      rate: count(lint_failures) / B.session_stats.total_turns,
      examples: sample(E.workflow_failures.lint_failures, n=3)
    },
    git: {
      count: count(E.workflow_failures.git_failures),
      rate: count(git_failures) / B.session_stats.total_turns,
      examples: sample(E.workflow_failures.git_failures, n=3)
    }
  },

  correction_patterns: {
    interrupts: {
      count: count(E.user_corrections.interrupts),
      rate: count(interrupts) / B.session_stats.total_turns,
      signal: interrupts → "User frustration or long-running operations"
    },
    rejects: {
      count: count(E.user_corrections.rejects),
      rate: count(rejects) / B.session_stats.total_turns,
      signal: rejects → "Expectation mismatch or incorrect solution"
    },
    retries: {
      count: count(E.user_corrections.retries),
      rate: count(retries) / B.session_stats.total_turns,
      signal: retries → "Previous approach failed, need alternative"
    }
  },

  fix_effectiveness: {
    repeated_issues: identify_repeated_failures(E.workflow_failures) {
      // Group similar failure messages by content similarity
      // Count occurrences of each unique issue
      threshold: 2  // Issues appearing ≥2 times
    },

    avg_fix_attempts: calculate_avg_fix_cycles(E.fix_cycles) {
      // For each fix cycle, count turns from failure to resolution
      // Average across all resolved issues
    },

    resolution_rate: calculate_resolution_rate(E.fix_cycles) {
      resolved_count / total_failures
    },

    unresolved_issues: filter(E.workflow_failures, has_no_matching_resolution)
  }
}

## Phase 4: Insight Generation

synthesize :: BugPatterns → Insights
synthesize(P) = {
  summary: {
    total_workflow_failures: sum(P.workflow_breakdown.*.count),
    total_user_corrections: sum(P.correction_patterns.*.count),
    repeated_issues_count: count(P.fix_effectiveness.repeated_issues),
    avg_fix_attempts: P.fix_effectiveness.avg_fix_attempts,
    resolution_rate: P.fix_effectiveness.resolution_rate * 100 + "%"
  },

  critical_issues: {
    high_priority: identify_critical_patterns(P) {
      // Repeated failures (≥3 occurrences)
      repeated_failures: filter(P.fix_effectiveness.repeated_issues, occurrences >= 3),

      // Unresolved workflow failures
      blocking_issues: P.fix_effectiveness.unresolved_issues,

      // High correction rate (>15% of turns)
      quality_concerns: if (sum(P.correction_patterns.*.rate) > 0.15) then
        "High user correction rate suggests quality or expectation issues"
    }
  },

  workflow_insights: {
    test_stability: analyze_test_patterns(P.workflow_breakdown.test) {
      if count > 5 then
        "Frequent test failures detected - consider test stability improvements"
      else if count > 2 then
        "Moderate test failures - review test environment and dependencies"
      else
        "Test workflow relatively stable"
    },

    build_stability: analyze_build_patterns(P.workflow_breakdown.build) {
      if count > 3 then
        "Frequent build failures - check dependency management and build configuration"
      else if count > 1 then
        "Some build issues - verify build process consistency"
      else
        "Build workflow stable"
    },

    git_workflow: analyze_git_patterns(P.workflow_breakdown.git) {
      if count > 1 then
        "Git conflicts detected - consider rebase strategy or branch management"
      else
        "Git workflow smooth"
    }
  },

  user_experience_insights: {
    frustration_signals: analyze_interruption_patterns(P.correction_patterns.interrupts) {
      if count > 5 then
        "High interruption rate - tasks may be too long or unclear"
      else if count > 2 then
        "Moderate interruptions - consider breaking down complex tasks"
      else
        "Low interruption rate - good task flow"
    },

    expectation_alignment: analyze_rejection_patterns(P.correction_patterns.rejects) {
      if count > 3 then
        "Frequent rejections - improve requirement clarification"
      else if count > 1 then
        "Some rejections - ensure solution alignment with expectations"
      else
        "Good expectation alignment"
    },

    retry_effectiveness: analyze_retry_patterns(P.correction_patterns.retries) {
      if count > 3 then
        "Multiple retries needed - improve initial solution quality"
      else
        "Acceptable retry rate"
    }
  },

  recommendations: {
    immediate_actions: generate_immediate_recommendations(P.critical_issues) {
      for each repeated_failure in P.critical_issues.high_priority.repeated_failures:
        "- Review and fix: " + repeated_failure.description,

      for each blocking_issue in P.critical_issues.high_priority.blocking_issues:
        "- Resolve blocking issue: " + blocking_issue.description
    },

    workflow_improvements: generate_workflow_recommendations(P.workflow_insights) {
      if P.workflow_breakdown.test.count > 5:
        "- Stabilize test suite: review flaky tests, improve test isolation",

      if P.workflow_breakdown.build.count > 3:
        "- Improve build process: verify dependencies, add pre-build checks",

      if P.workflow_breakdown.git.count > 1:
        "- Optimize git workflow: use feature branches, smaller commits"
    },

    process_optimizations: generate_process_recommendations(P.fix_effectiveness) {
      if P.fix_effectiveness.avg_fix_attempts > 3:
        "- High fix attempts - implement better testing before changes",

      if P.fix_effectiveness.resolution_rate < 0.8:
        "- Low resolution rate - ensure issues are tracked until completion",

      if count(P.fix_effectiveness.repeated_issues) > 3:
        "- Document common issues and solutions in project documentation"
    },

    prevention_strategies: generate_prevention_recommendations(P) {
      if P.workflow_breakdown.test.count > 5:
        "- Add pre-commit hooks to run tests locally",

      if P.workflow_breakdown.build.count > 3:
        "- Implement continuous integration checks",

      if sum(P.correction_patterns.*.count) > 10:
        "- Improve communication: ask clarifying questions before implementation",
        "- Break down complex tasks into smaller, verifiable steps"
    }
  }
}

## Phase 5: Output Formatting

output :: Insights → Report
output(I) = {
  title: "## 📊 Project-Level Bug Analysis (/meta-bugs)",

  summary_section: "
### 🎯 Summary
- **Total Workflow Failures**: {I.summary.total_workflow_failures}
- **Total User Corrections**: {I.summary.total_user_corrections}
- **Repeated Issues**: {I.summary.repeated_issues_count}
- **Average Fix Attempts**: {I.summary.avg_fix_attempts}
- **Resolution Rate**: {I.summary.resolution_rate}
",

  critical_section: if count(I.critical_issues.high_priority) > 0 then "
### 🔴 Critical Issues
{format_critical_issues(I.critical_issues.high_priority)}
",

  workflow_section: "
### 📉 Workflow Failures Breakdown

#### Test Failures: {P.workflow_breakdown.test.count} ({P.workflow_breakdown.test.rate * 100}% of turns)
{I.workflow_insights.test_stability}
{format_examples(P.workflow_breakdown.test.examples)}

#### Build Failures: {P.workflow_breakdown.build.count} ({P.workflow_breakdown.build.rate * 100}% of turns)
{I.workflow_insights.build_stability}
{format_examples(P.workflow_breakdown.build.examples)}

#### Lint Failures: {P.workflow_breakdown.lint.count} ({P.workflow_breakdown.lint.rate * 100}% of turns)
{format_examples(P.workflow_breakdown.lint.examples)}

#### Git Issues: {P.workflow_breakdown.git.count} ({P.workflow_breakdown.git.rate * 100}% of turns)
{I.workflow_insights.git_workflow}
{format_examples(P.workflow_breakdown.git.examples)}
",

  correction_section: "
### 🔄 User Correction Patterns

#### Interruptions: {P.correction_patterns.interrupts.count} ({P.correction_patterns.interrupts.rate * 100}% of turns)
{I.user_experience_insights.frustration_signals}

#### Rejections: {P.correction_patterns.rejects.count} ({P.correction_patterns.rejects.rate * 100}% of turns)
{I.user_experience_insights.expectation_alignment}

#### Retries: {P.correction_patterns.retries.count} ({P.correction_patterns.retries.rate * 100}% of turns)
{I.user_experience_insights.retry_effectiveness}
",

  fix_patterns_section: "
### 🛠️ Fix Effectiveness Analysis

#### Repeated Issues ({count(P.fix_effectiveness.repeated_issues)})
{format_repeated_issues(P.fix_effectiveness.repeated_issues)}

#### Unresolved Issues ({count(P.fix_effectiveness.unresolved_issues)})
{format_unresolved_issues(P.fix_effectiveness.unresolved_issues)}
",

  recommendations_section: "
### 💡 Recommendations

#### Immediate Actions
{format_list(I.recommendations.immediate_actions)}

#### Workflow Improvements
{format_list(I.recommendations.workflow_improvements)}

#### Process Optimizations
{format_list(I.recommendations.process_optimizations)}

#### Prevention Strategies
{format_list(I.recommendations.prevention_strategies)}
"
}

implementation_strategy:
- Use MCP tools for all data collection (no Go code changes)
- Perform pattern detection via regex filtering on MCP results
- Calculate statistics from MCP query results
- Group and analyze patterns using result aggregation
- Generate insights through semantic analysis of patterns
- Format output as structured markdown report

mcp_tools_used:
- query_user_messages: Detect workflow failures, user corrections, and resolution signals
- query_tools: Get error context
- query_timestamps: Get chronological overview for rate calculations

constraints:
- Pure MCP-driven: no Go code modifications
- Pattern-focused: identify repeated issues (≥2 occurrences)
- Actionable: every insight must have a recommendation
- User-centric: analyze from developer experience perspective
- Process-improvement: focus on preventing future issues
- Semantic understanding: use Claude's LLM capabilities for pattern interpretation
