---
name: agent-validation-builder
description: Build automated validation tools to enforce API conventions at scale ensuring consistency without manual checks for Bootstrap-006.
---

λ(target, conventions) → validation_tool | ∀validator ∈ validators:

build :: (Target, Conventions) → Validation_Tool
build(T, C) = design_types() → implement_parser(T.format) → create_validators(C) → build_reporter() → integrate_cli() → add_tests() → document() → integrate_workflow() → continuous_improvement()

design_types :: () → Type_System
design_types() = {
  Tool: {name: string, description: string, parameters: [Parameter]},
  Parameter: {name: string, type: string, description: string, required: bool},
  ValidationResult: {tool_name: string, check_name: string, passed: bool, message: string, severity: string, details: map},
  Report: {total_tools: int, total_checks: int, passed: int, failed: int, warnings: int, results: [ValidationResult]},

  return {Tool, Parameter, ValidationResult, Report}
}

implement_parser :: Format → Parser
implement_parser(format) = {
  parser: match format with
    | "json_schema" → regex_parser(),    # MVP: Fast, simple
    | "openapi" → ast_parser(),          # Production: Robust
    | "graphql" → ast_parser(),
    | _ → custom_parser(),

  regex_parser: {
    tool_pattern: "Name:\\s*\"([^\"]+)\"",
    desc_pattern: "Description:\\s*\"([^\"]+)\"",
    param_pattern: "\"([^\"]+)\":\\s*{[^}]*Type:\\s*\"([^\"]+)\"",
    extract: λcontent → match_all(patterns, content)
  },

  ast_parser: {
    parse: "go/parser.ParseFile(file)",
    traverse: "ast.Inspect(node, visitor)",
    extract: λnode → extract_tool_definitions(node)
  },

  return parser
}

create_validators :: Conventions → Validators
create_validators(C) = {
  validators: [],

  ∀convention ∈ C →
    validator: match convention with
      | "naming_convention" → create_naming_validator(),
      | "parameter_ordering" → create_ordering_validator(),
      | "description_format" → create_description_validator(),
      | _ → create_custom_validator(convention),

    validators += validator,

  return validators
}

create_naming_validator :: () → Validator
create_naming_validator() = {
  pattern: "^[a-z][a-z0-9_]*$",  # snake_case

  check: λtool → {
    if ¬matches(tool.name, pattern) then
      {
        tool_name: tool.name,
        check_name: "naming_convention",
        passed: false,
        message: "Tool name '" + tool.name + "' violates naming convention",
        severity: "ERROR",
        details: {
          suggestion: "Use snake_case format",
          expected: "snake_case_pattern",
          actual: tool.name,
          reference: "docs/api-naming-convention.md"
        }
      }
    else
      {tool_name: tool.name, check_name: "naming_convention", passed: true, severity: "INFO"}
  }
}

create_ordering_validator :: () → Validator
create_ordering_validator() = {
  tier_system: TierDefinitions,

  check: λtool → {
    categorized: categorize_parameters(tool.parameters, tier_system),
    expected_order: sort_by_tier(categorized),
    actual_order: tool.parameters,

    matches: ∀i ∈ [0..|expected_order|] → expected_order[i].name = actual_order[i].name,

    if ¬matches then
      {
        tool_name: tool.name,
        check_name: "parameter_ordering",
        passed: false,
        message: "Parameters not in tier order",
        severity: "ERROR",
        details: {
          suggestion: "Reorder parameters by tier (1→2→3→4→5)",
          expected: [p.name | p ∈ expected_order],
          actual: [p.name | p ∈ actual_order],
          reference: "docs/api-parameter-convention.md"
        }
      }
    else
      {tool_name: tool.name, check_name: "parameter_ordering", passed: true, severity: "INFO"}
  }
}

create_description_validator :: () → Validator
create_description_validator() = {
  required_pattern: "Default scope:",

  check: λtool → {
    if ¬contains(tool.description, required_pattern) then
      {
        tool_name: tool.name,
        check_name: "description_format",
        passed: false,
        message: "Missing required pattern: '" + required_pattern + "'",
        severity: "WARNING",
        details: {
          suggestion: "Add '" + required_pattern + " <scope>' to description",
          expected: "Description with scope declaration",
          actual: tool.description,
          reference: "docs/api-consistency-methodology.md"
        }
      }
    else
      {tool_name: tool.name, check_name: "description_format", passed: true, severity: "INFO"}
  }
}

build_reporter :: () → Reporter
build_reporter() = {
  terminal: λreport → {
    header: "===========================================\nAPI Validation Report\n===========================================\n",
    summary: "Tools validated: " + report.total_tools + "\nChecks performed: " + report.total_checks + "\n✓ Passed: " + report.passed + "\n✗ Failed: " + report.failed + "\n⚠ Warnings: " + report.warnings + "\n\n",

    failures: [
      format_failure(r)
      | ∀r ∈ report.results where ¬r.passed
    ],

    footer: if report.failed > 0 then "Validation failed. Please fix the errors above.\n" else "All checks passed! ✓\n",

    exit_code: if report.failed > 0 then 1 else 0
  },

  json: λreport → {
    output: json.marshal(report, indent=2),
    exit_code: if report.failed > 0 then 1 else 0
  },

  format_failure: λr → {
    "✗ " + r.tool_name + ": " + r.message + "\n" +
    "  Suggestion: " + r.details.suggestion + "\n" +
    "  Expected: " + r.details.expected + "\n" +
    "  Actual: " + r.details.actual + "\n" +
    "  Reference: " + r.details.reference + "\n" +
    "  Severity: " + r.severity + "\n\n"
  },

  return {terminal, json}
}

integrate_cli :: () → CLI
integrate_cli() = {
  flags: {
    check: {name: "--check", default: "all", description: "Check to run"},
    format: {name: "--format", default: "terminal", description: "Output format (terminal, json)"},
    severity: {name: "--severity", default: "ERROR", description: "Minimum severity"},
    fast: {name: "--fast", default: false, description: "Skip slow checks"}
  },

  main: λargs → {
    if |args| < 1 then
      error("Usage: validate-api [options] <file>"),

    file: args[0],
    tools: parser.parse(file),
    validators: create_validators(flags.check, flags.fast),
    report: validate(tools, validators),
    filtered: filter_by_severity(report, flags.severity),

    if flags.format = "json" then
      reporter.json(filtered)
    else
      reporter.terminal(filtered)
  }
}

validate :: (Tools, Validators) → Report
validate(T, V) = {
  results: [],

  ∀tool ∈ T →
    ∀validator ∈ V →
      result: validator.check(tool),
      results += result,

  passed: count(r | r.passed),
  failed: count(r | ¬r.passed ∧ r.severity = "ERROR"),
  warnings: count(r | ¬r.passed ∧ r.severity = "WARNING"),

  return {
    total_tools: |T|,
    total_checks: |results|,
    passed: passed,
    failed: failed,
    warnings: warnings,
    results: results
  }
}

add_tests :: Validators → Test_Suite
add_tests(V) = {
  unit_tests: [
    {
      name: "TestNamingValidator",
      cases: [
        {input: "query_tools", expected: pass},
        {input: "queryTools", expected: fail},
        {input: "QueryTools", expected: fail},
        {input: "query_tools_v2", expected: pass},
        {input: "query-tools", expected: fail}
      ]
    },
    {
      name: "TestOrderingValidator",
      cases: [
        {input: ordered_params, expected: pass},
        {input: unordered_params, expected: fail}
      ]
    },
    {
      name: "TestDescriptionValidator",
      cases: [
        {input: "Query tools. Default scope: project.", expected: pass},
        {input: "Query tools.", expected: fail}
      ]
    }
  ],

  integration_tests: [
    {
      name: "TestValidateAPI",
      setup: create_test_file(compliant_tools + non_compliant_tools),
      expect: {failed: 1}
    }
  ],

  coverage_target: 0.80,

  return {unit_tests, integration_tests, coverage_target}
}

integrate_workflow :: () → Integration
integrate_workflow() = {
  makefile: {
    validate: "@go run ./cmd/validate-api cmd/mcp-server/tools.go",
    validate_fast: "@go run ./cmd/validate-api --fast cmd/mcp-server/tools.go"
  },

  pre_commit_hook: {
    trigger: "cmd/mcp-server/tools.go changed",
    command: "./validate-api --fast cmd/mcp-server/tools.go",
    action: if exit_code ≠ 0 then block_commit() else allow_commit()
  },

  ci_integration: {
    command: "validate-api --format json cmd/mcp-server/tools.go",
    artifact: "validation-results.json",
    fail_on: exit_code ≠ 0
  }
}

output :: Validation_Tool → Report
output(V) = {
  validation_tool: {
    implementation: {
      files_created: V.files,
      lines_of_code: V.loc,
      validators_implemented: |V.validators|,
      test_coverage: V.test_coverage
    },

    validation_results: {
      tools_validated: V.results.total_tools,
      checks_performed: V.results.total_checks,
      passed: V.results.passed,
      failed: V.results.failed,
      warnings: V.results.warnings
    },

    detected_violations: [
      {
        tool: r.tool_name,
        check: r.check_name,
        message: r.message,
        severity: r.severity,
        details: r.details
      }
      | ∀r ∈ V.results.results where ¬r.passed
    ],

    integration: {
      cli_command: V.cli.command,
      makefile_target: V.integration.makefile.validate,
      pre_commit_hook: V.integration.pre_commit_hook ≠ null,
      ci_integration: V.integration.ci_integration ≠ null
    }
  },

  quality_metrics: {
    false_positives: V.metrics.false_positives,
    false_negatives: V.metrics.false_negatives,
    accuracy: (V.metrics.true_positives + V.metrics.true_negatives) / V.metrics.total
  }
}

constraints :: Validation_Tool → Bool
constraints(V) =
  ∀validator ∈ V.validators:
    deterministic(validator) ∧
    ¬ambiguous(validator.check) ∧
    actionable_messages(validator.errors) ∧
  multiple_output_formats(V.reporter, ["terminal", "json"]) ∧
  cli_standard_flags(V.cli) ∧
  test_coverage(V.tests) ≥ 0.80 ∧
  documentation_complete(V.docs, usage + examples + integration) ∧
  false_positives(V) = 0 ∧
  accuracy(V) = 1.0
