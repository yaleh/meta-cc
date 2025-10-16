---
name: agent-verify-before-remove
description: Verifies code is unused before removal using static analysis, test coverage, reference search, and runtime checks to prevent costly mistakes in Bootstrap-004.
---

λ(target_code, scope) → verification_result | ∀method ∈ verification_methods:

verify :: (Target_Code, Scope) → Verification_Result
verify(T, S) = analyze_static(T, S) ∧ check_coverage(T) ∧ search_references(T, S) ∧ check_runtime(T)

analyze_static :: (Target_Code, Scope) → Static_Analysis
analyze_static(T, S) = {
  tool: match T.language with
    | go → run("staticcheck ./... && go vet ./..."),
    | python → run("pylint && mypy && vulture"),
    | javascript → run("eslint && tsc --noEmit"),
    | java → run("spotbugs:check"),

  result: {
    warnings: count(unused_warnings),
    flagged_unused: contains(T.function, warnings),
    details: extract(warning_messages)
  },

  confidence:
    if flagged_unused then 0.90
    else if no_warnings then 0.30
    else 0.10
}

check_coverage :: Target_Code → Coverage_Analysis
check_coverage(T) = {
  tool: match T.language with
    | go → run("go test -cover ./..."),
    | python → run("pytest --cov"),
    | javascript → run("jest --coverage"),

  coverage: extract_percentage(T.function),
  dependent_tests: count(tests_using(T.function)),

  confidence:
    if coverage > 0 then 1.0  # DEFINITELY used
    else 0.40  # May be unused, but not conclusive
}

search_references :: (Target_Code, Scope) → Reference_Search
search_references(T, S) = {
  tool: "ripgrep",

  command: rg(T.function, type=T.language, scope=S),

  matches: parse_output(command) where {
    definition: [m | m.line = T.definition_line],
    usages: [m | m.line ≠ T.definition_line],
    check_patterns: [
      direct_calls: pattern(T.function + "("),
      interface_impl: pattern("implements " + T.function),
      type_assertions: pattern("x.(" + T.function + ")"),
      reflection: pattern("reflect.*" + T.function)
    ]
  },

  confidence:
    if |usages| > 0 then 1.0  # DEFINITELY used
    else if |matches| = 1 ∧ matches[0] = definition then 0.85  # Only definition found
    else 0.60
}

check_runtime :: Target_Code → Runtime_Check
check_runtime(T) = {
  methods: [
    log_analysis: grep(T.function, "/var/log/app.log"),
    api_analytics: query("/api/analytics?function=" + T.function),
    reflection_usage: rg("reflect.*" + T.function)
  ],

  invocations: count(log_hits + api_hits + reflection_hits),

  confidence:
    if invocations > 0 then 1.0  # DEFINITELY used
    else 0.20  # Logs may be incomplete
}

aggregate_evidence :: (Static, Coverage, References, Runtime) → Decision
aggregate_evidence(S, C, R, Rt) = {
  confidence: max(S.confidence, C.confidence, R.confidence, Rt.confidence),

  status:
    if R.usages > 0 ∨ C.coverage > 0 ∨ Rt.invocations > 0 then
      "IN_USE"
    else if S.flagged_unused ∧ R.usages = 0 ∧ C.coverage = 0 then
      "SAFE_TO_REMOVE"
    else
      "UNCERTAIN",

  recommendation:
    if status = "IN_USE" then
      {action: "KEEP", reason: describe_evidence(usage_locations)}
    else if status = "SAFE_TO_REMOVE" ∧ confidence ≥ 0.80 then
      {action: "REMOVE", reason: "no usage detected with high confidence"}
    else
      {action: "INVESTIGATE_FURTHER", reason: "inconclusive evidence"}
}

verify_after_removal :: Target_Code → Test_Result
verify_after_removal(T) = {
  baseline: run_tests(before_removal),
  remove_code(T),
  after: run_tests(after_removal),

  result:
    if after.pass_count = baseline.pass_count ∧ after.all_pass then
      {success: true, tests_pass: true}
    else
      rollback(T),
      {success: false, tests_pass: false, reason: "tests failed"}
}

output :: (Evidence, Decision) → Verification_Report
output(E, D) = {
  status: D.status,
  confidence: D.confidence,

  evidence: {
    static_analysis: {
      tool: E.static.tool,
      warnings: E.static.warnings,
      flagged_unused: E.static.flagged_unused
    },

    test_coverage: {
      percentage: E.coverage.coverage,
      dependent_tests: E.coverage.dependent_tests
    },

    reference_search: {
      tool: E.references.tool,
      matches: |E.references.matches|,
      locations: E.references.usages
    },

    runtime_check: {
      method: E.runtime.methods,
      invocations: E.runtime.invocations
    }
  },

  recommendation: D.recommendation,

  report: generate_markdown_report(E, D)
}

constraints :: Verification → Bool
constraints(V) =
  ∀method ∈ verification_methods:
    executed(method) ∧
    documented(method) ∧
    evidence_based(decision) ∧
  confidence(V) ≥ 0.80 → status(V) = "SAFE_TO_REMOVE" ∧
  ∀removal: verify_tests_pass(after_removal) ∧
  conservative_approach(uncertain_cases)
