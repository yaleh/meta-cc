---
name: agent-documentation-enhancer
description: Enhance API documentation with practical, example-driven content that teaches both usage and rationale to improve tool adoption and reduce support burden in Bootstrap-006.
---

λ(tools, strategy) → enhancements | ∀tool ∈ low_usage_tools:

enhance :: (Tools, Strategy) → Enhancements
enhance(T, S) = prioritize(T) → add_conventions() → add_use_cases(T) → progressive_complexity(T) → document_automation() → add_troubleshooting() → test_examples() → schedule_maintenance()

prioritize_tools :: (Tools, Usage_Data) → Priority_Queue
prioritize_tools(T, U) = {
  scored: [
    {
      tool: t,
      score: 0.4·low_usage(t, U) + 0.3·support_requests(t, U) + 0.2·complexity(t) + 0.1·newness(t, U)
    }
    | ∀t ∈ T
  ],

  low_usage: λt → if U[t.name].usage_rate < 0.10 then 1 else 0,
  support_requests: λt → if U[t.name].support_requests > 5 then 1 else 0,
  complexity: λt → if |t.parameters| > 5 ∨ has_complex_params(t) then 1 else 0,
  newness: λt → if U[t.name].age_days < 30 then 1 else 0,

  return sort_by(scored, score, desc)
}

add_conventions :: () → Convention_Section
add_conventions() = {
  tier_system: {
    tier_1: {name: "Required Parameters", description: "Must be provided for tool to function"},
    tier_2: {name: "Filtering Parameters", description: "Narrow search results (affect WHAT is returned)"},
    tier_3: {name: "Range Parameters", description: "Define bounds, thresholds, windows"},
    tier_4: {name: "Output Control", description: "Control output size or format"},
    tier_5: {name: "Standard Parameters", description: "Cross-cutting concerns (scope, jq_filter, stats_only)"}
  },

  rationale: {
    consistency: "Learn pattern once, applies everywhere",
    predictability: "Required params first, output control last",
    readability: "Logical grouping makes schemas easier to understand"
  },

  clarifications: {
    json_ordering: "JSON object properties are unordered. Order is for documentation/readability only."
  }
}

add_use_cases :: Tool → Use_Cases
add_use_cases(tool) = {
  cases: [
    {
      scenario: case.name,
      problem: describe_user_problem(case),
      solution: generate_json_example(case),
      outcome: explain_result(case),
      analysis: explain_learning(case)
    }
    | ∀case ∈ identify_scenarios(tool)
  ],

  count: 3 ≤ |cases| ≤ 5,

  return cases
}

progressive_complexity :: (Tool, Examples) → Progressive_Examples
progressive_complexity(tool, examples) = {
  basic: [
    {
      name: ex.name,
      params: minimal_params(ex),
      annotation: "Basic usage with default settings"
    }
    | ∀ex ∈ examples where complexity(ex) = "low"
  ],

  practical: [
    {
      name: ex.name,
      params: common_params(ex),
      annotation: "Real-world scenario with typical configuration",
      use_when: describe_use_case(ex)
    }
    | ∀ex ∈ examples where complexity(ex) = "medium"
  ],

  advanced: [
    {
      name: ex.name,
      params: full_params(ex),
      annotation: "Complex scenario with multiple conditions",
      use_when: describe_advanced_case(ex)
    }
    | ∀ex ∈ examples where complexity(ex) = "high"
  ],

  structure: basic → practical → advanced,

  return {basic: basic, practical: practical, advanced: advanced}
}

add_sql_reference :: Tool → SQL_Reference
add_sql_reference(tool) = {
  operators: [
    {op: "=", example: "tool = 'Read'", description: "Exact match"},
    {op: "!=", example: "status != 'error'", description: "Not equal"},
    {op: ">", example: "duration_ms > 1000", description: "Numeric comparison"},
    {op: "LIKE", example: "tool LIKE 'query%'", description: "Pattern matching"},
    {op: "IN", example: "tool IN ('Read', 'Write')", description: "Multiple values"},
    {op: "AND", example: "tool = 'Read' AND status = 'error'", description: "Logical operators"}
  ],

  practical_cases: [
    {name: "Find slow commands", query: "duration_ms > 5000"},
    {name: "Error pattern analysis", query: "tool IN ('Read', 'Write') AND status = 'error'"},
    {name: "Tool usage comparison", query: "tool IN ('Bash', 'Read')"},
    {name: "Activity during time window", query: "timestamp >= '2025-10-16T14:00' AND timestamp < '2025-10-16T15:00'"},
    {name: "Multi-condition filtering", query: "tool = 'Bash' AND status = 'error' AND duration_ms > 10000"}
  ],

  return {operators: operators, practical_cases: practical_cases}
}

document_automation :: Tool → Automation_Docs
document_automation(tool) = {
  purpose: describe_purpose(tool),

  installation: {
    automatic: generate_install_command(tool),
    manual: generate_manual_steps(tool)
  },

  options: [
    {flag: f.name, description: f.description, default: f.default}
    | ∀f ∈ tool.options
  ],

  examples: {
    passing: generate_passing_example(tool),
    failing: generate_failing_example(tool)
  },

  integration: {
    local: generate_local_usage(tool),
    pre_commit: generate_pre_commit_config(tool),
    ci_cd: {
      github_actions: generate_github_workflow(tool),
      gitlab_ci: generate_gitlab_config(tool)
    }
  }
}

add_troubleshooting :: Tool → Troubleshooting
add_troubleshooting(tool) = {
  common_issues: [
    {
      name: issue.name,
      symptom: describe_symptom(issue),
      cause: explain_cause(issue),
      fix: provide_solution(issue)
    }
    | ∀issue ∈ identify_common_issues(tool)
  ],

  constraint: |common_issues| ≥ 6,

  return common_issues
}

test_examples :: Examples → Verification
test_examples(E) = {
  results: [],

  ∀example ∈ E →
    result: run_example(example),
    expected: get_expected_output(example),

    results += {
      example: example.name,
      passed: result = expected,
      result: result,
      expected: expected
    },

  accuracy: count(r | r.passed) / |results|,

  return {
    total: |E|,
    passed: count(r | r.passed),
    failed: count(r | ¬r.passed),
    accuracy: accuracy
  }
}

schedule_maintenance :: () → Maintenance_Schedule
schedule_maintenance() = {
  frequency: "quarterly",

  tasks: [
    "Review low-usage tools (check if adoption improved)",
    "Update examples (ensure they still work)",
    "Add new use cases (based on user feedback)",
    "Fix broken examples (after API changes)",
    "Expand troubleshooting (new common issues)"
  ],

  metrics: [
    "Tool adoption rates (before/after enhancement)",
    "Support request frequency (should decrease)",
    "Example accuracy (should be 100%)"
  ]
}

output :: Enhancements → Report
output(E) = {
  documentation_enhancements: {
    tools_enhanced: [t.name | t ∈ E.tools],
    enhancements_per_tool: [
      {
        tool: t.name,
        convention_section: t.has_conventions,
        practical_cases_added: |t.practical_cases|,
        progressive_examples: t.has_progressive,
        troubleshooting_items: |t.troubleshooting|
      }
      | ∀t ∈ E.tools
    ]
  },

  example_structure: {
    total_examples: sum(|t.examples| | t ∈ E.tools),
    by_type: {
      basic: sum(|t.basic| | t ∈ E.tools),
      practical: sum(|t.practical| | t ∈ E.tools),
      advanced: sum(|t.advanced| | t ∈ E.tools)
    },
    annotations: {
      comments: count(e.has_comment | e ∈ all_examples),
      explanations: count(e.has_explanation | e ∈ all_examples)
    }
  },

  automation_docs: {
    installation_guide: E.has_installation,
    behavior_examples: E.passing_examples + E.failing_examples,
    ci_integrations: count(E.github_actions, E.gitlab_ci),
    troubleshooting_items: sum(|t.troubleshooting| | t ∈ E.tools)
  },

  quality_metrics: {
    examples_tested: E.verification.total,
    examples_passing: E.verification.passed,
    accuracy: E.verification.accuracy
  },

  adoption_impact: {
    usage_before: E.metrics.usage_before,
    usage_after: E.metrics.usage_after,
    improvement: E.metrics.usage_after - E.metrics.usage_before
  }
}

constraints :: Documentation → Bool
constraints(D) =
  ∀tool ∈ D.tools:
    conventions_before_catalog(D) ∧
    practical_examples(tool) ∧ 3 ≤ |tool.examples| ≤ 5 ∧
    problem_solution_outcome(tool.examples) ∧
    progressive_complexity(tool.examples) ∧
  ∀automation ∈ D.automation_tools:
    installation_guide(automation) ∧
    behavior_examples(automation, passing + failing) ∧
    ci_integration_examples(automation) ∧
    troubleshooting_items(automation) ∧ |automation.troubleshooting| ≥ 6 ∧
  all_examples_tested(D) ∧
  all_examples_passing(D) ∧
  accuracy(D) = 1.0
