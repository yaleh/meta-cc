---
name: agent-schema-refactorer
description: Safely refactor API schema ordering and structure leveraging JSON's unordered object property guarantee to ensure readability improvements without breaking existing clients in Bootstrap-006.
---

λ(schema, convention) → refactoring | ∀param ∈ parameters:

refactor :: (Schema, Convention) → Refactoring_Result
refactor(S, C) = verify_json_property(S) → identify_targets(S, C) → plan_changes() → make_changes() → run_tests() → verify_compilation() → confirm_backward_compat() → document() → update_docs()

verify_json_property :: Schema → JSON_Verification
verify_json_property(S) = {
  format_check: {
    is_json: S.format ∈ {"json_schema", "openapi", "graphql"},
    spec: "RFC 8259",
    guarantee: "Object properties are unordered",
    implication: "Reordering parameters is safe"
  },

  valid_for: ["JSON", "OpenAPI/Swagger", "GraphQL arguments"],
  invalid_for: ["Positional arguments", "Array ordering", "CSV"],

  decision: if format_check.is_json then
             {safe_to_reorder: true, reason: "JSON property order irrelevant"}
           else
             {safe_to_reorder: false, reason: "Order matters for format"},

  return {
    format: S.format,
    safe_to_reorder: decision.safe_to_reorder,
    guarantee: format_check.guarantee,
    verification: ∀call → call(old_order) = call(new_order)
  }
}

identify_targets :: (Schema, Convention) → Reordering_Targets
identify_targets(S, C) = {
  targets: [],

  ∀tool ∈ S.tools →
    current: extract_parameter_order(tool),
    expected: apply_convention(tool, C),

    if current ≠ expected then
      compliance: count(i | current[i] = expected[i]) / |expected|,
      changes: calculate_changes(current, expected),

      targets += {
        tool: tool.name,
        compliance_before: compliance,
        changes: changes
      },

  return targets
}

calculate_changes :: (Current, Expected) → Changes
calculate_changes(current, expected) = {
  changes: [
    {
      param: p,
      from: find_index(p, current),
      to: find_index(p, expected),
      tier: get_tier(p, expected)
    }
    | ∀p ∈ current where find_index(p, current) ≠ find_index(p, expected)
  ],

  return changes
}

plan_changes :: Targets → Change_Plan
plan_changes(T) = {
  plans: [
    {
      tool: t.tool,
      current_order: t.current,
      target_order: t.expected,
      changes: t.changes,
      tier_comments: identify_tier_comments(t.expected)
    }
    | ∀t ∈ T
  ],

  total_lines: sum(|p.changes| | p ∈ plans),

  return plans
}

make_changes :: Change_Plan → Modified_Schema
make_changes(P) = {
  modified: [],

  ∀plan ∈ P →
    tool: plan.tool,

    reordered: [
      {tier_comment: tier.comment} + tier.parameters
      | ∀tier ∈ group_by_tier(plan.target_order)
    ],

    modified += {
      tool: tool,
      before: plan.current_order,
      after: reordered,
      lines_changed: |plan.changes|
    },

  return modified
}

run_tests :: () → Test_Results
run_tests() = {
  unit_tests: {
    command: "go test ./...",
    result: execute(command),
    expected: "PASS"
  },

  integration_tests: {
    command: "go test -tags=integration ./...",
    result: execute(command),
    expected: "PASS"
  },

  api_tests: {
    command: "./test-api-calls.sh",
    result: execute(command),
    expected: "All calls successful"
  },

  all_passed: unit_tests.result = "PASS" ∧
              integration_tests.result = "PASS" ∧
              api_tests.result = "SUCCESS",

  decision: if all_passed then
              "Backward compatibility confirmed"
            else
              "Rollback changes, investigate failures",

  return {
    unit: unit_tests.result,
    integration: integration_tests.result,
    api: api_tests.result,
    all_passed: all_passed
  }
}

verify_compilation :: () → Build_Verification
verify_compilation() = {
  compile: {
    command: "go build ./...",
    expected: exit_code = 0
  },

  lint: {
    command: "staticcheck ./...",
    expected: "no errors"
  },

  type_check: {
    command: "go vet ./...",
    expected: exit_code = 0
  },

  all_passed: compile.result = 0 ∧
              lint.result = "no errors" ∧
              type_check.result = 0,

  return {
    compile_success: compile.result = 0,
    lint_clean: lint.result = "no errors",
    type_check_pass: type_check.result = 0,
    all_passed: all_passed
  }
}

confirm_backward_compat :: Modified_Schema → Compatibility_Check
confirm_backward_compat(M) = {
  test_cases: [
    {
      name: "Old order still works",
      call: generate_call(old_order),
      expected: "200 OK"
    },
    {
      name: "New order works",
      call: generate_call(new_order),
      expected: "200 OK"
    },
    {
      name: "Mixed order works",
      call: generate_call(mixed_order),
      expected: "200 OK"
    }
  ],

  results: [
    {test: tc.name, passed: execute_call(tc.call) = tc.expected}
    | ∀tc ∈ test_cases
  ],

  all_passed: ∀r ∈ results → r.passed,

  return {
    old_order_works: results[0].passed,
    new_order_works: results[1].passed,
    mixed_order_works: results[2].passed,
    breaking_changes: if all_passed then 0 else count(r | ¬r.passed)
  }
}

document_changes :: (Modified_Schema, Tests, Compilation, Compatibility) → Report
document_changes(M, T, C, Compat) = {
  changes_made: {
    tools_refactored: [m.tool | m ∈ M],
    tools_verified: [t | t.compliance_before = 1.0],
    lines_changed: sum(m.lines_changed | m ∈ M)
  },

  safety_verification: {
    json_property_confirmed: true,
    tests_passed: T.all_passed,
    compilation_successful: C.all_passed,
    backward_compatible: Compat.all_passed,
    breaking_changes: Compat.breaking_changes
  },

  readability_improvements: {
    tier_comments_added: count_tier_comments(M),
    logical_grouping: "Filtering → Output Control → Standard",
    consistency_before: avg(m.compliance_before | m ∈ M),
    consistency_after: 1.0
  },

  return {
    report: format_markdown(changes_made, safety_verification, readability_improvements),
    status: if T.all_passed ∧ C.all_passed ∧ Compat.all_passed then "SUCCESS" else "FAILURE"
  }
}

create_migration_guide :: () → Migration_Guide
create_migration_guide() = {
  guide: {
    what_changed: "Parameters reordered to follow tier system. Tier comments added for clarity. No functional changes.",
    impact: "NONE - JSON property order doesn't affect API calls.",
    existing_code_works: true,
    action_required: "Nothing required - existing code continues to work.",
    optional: "Update your code to follow new convention for consistency."
  },

  faq: [
    {q: "Do I need to update my API calls?", a: "No, JSON property order doesn't matter."},
    {q: "Will my old code break?", a: "No, backward compatibility guaranteed."},
    {q: "Why was this changed?", a: "Improved readability and consistency (documentation only)."}
  ]
}

output :: Refactoring → Report
output(R) = {
  refactoring_report: {
    files_modified: R.files,
    lines_changed: R.lines_changed,
    tools_refactored: [
      {
        tool: t.name,
        parameters_reordered: |t.changes|,
        compliance_before: t.compliance_before,
        compliance_after: t.compliance_after
      }
      | ∀t ∈ R.targets
    ],
    tools_verified: [
      {tool: t.name, compliance: t.compliance, status: "ALREADY_COMPLIANT"}
      | ∀t ∈ R.compliant
    ]
  },

  safety_verification: {
    json_property_confirmed: R.json_verified,
    tests_passed: R.tests.all_passed,
    compilation_successful: R.compilation.all_passed,
    backward_compatible: R.compatibility.all_passed,
    breaking_changes: R.compatibility.breaking_changes
  },

  readability_improvements: {
    tier_comments_added: R.tier_comments,
    logical_grouping: R.grouping,
    consistency_before: R.consistency_before,
    consistency_after: R.consistency_after
  },

  backward_compatibility: {
    old_order_works: R.compatibility.old_order_works,
    new_order_works: R.compatibility.new_order_works,
    mixed_order_works: R.compatibility.mixed_order_works,
    clients_affected: 0
  }
}

constraints :: Refactoring → Bool
constraints(R) =
  json_property_verified(R.schema) ∧
  ∀tool ∈ R.targets:
    safe_to_reorder(tool) ∧
    changes_planned(tool) ∧
    tier_comments_added(tool) ∧
  tests_passed(R) ∧ test_pass_rate(R) = 1.0 ∧
  compilation_successful(R) ∧
  backward_compatible(R) ∧
  breaking_changes(R) = 0 ∧
  old_calls_work(R) ∧ new_calls_work(R) ∧ mixed_calls_work(R) ∧
  documentation_updated(R) ∧
  migration_guide_provided(R)
