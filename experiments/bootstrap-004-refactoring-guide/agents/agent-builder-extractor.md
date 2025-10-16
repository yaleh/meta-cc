---
name: agent-builder-extractor
description: Extracts helper functions from repetitive structure definitions (API schemas, forms, configs) to reduce duplication while preserving behavioral equivalence in Bootstrap-004.
---

λ(target_file, duplication_threshold) → extraction_result | ∀pattern ∈ duplicated_structures:

extract :: (File, Threshold) → Extraction_Result
extract(F, T) = analyze_duplication(F) → categorize_parameters(F) → extract_helpers(F) → refactor_usages(F) → verify_equivalence(F)

analyze_duplication :: File → Duplication_Analysis
analyze_duplication(F) = {
  total_lines: wc(F),

  patterns: grep_repeated_structures(F),

  duplication: {
    lines: count(duplicated_lines),
    ratio: lines / total_lines,
    occurrences: count(pattern_occurrences)
  },

  decision:
    if ratio < threshold then
      "SKIP"
    else if ratio ≥ 0.30 then
      "HIGH_PRIORITY"
    else
      "MEDIUM_PRIORITY"
}

categorize_parameters :: File → Parameter_Categories
categorize_parameters(F) = {
  common: [p | occurrences(p) ≥ 0.50 * |definitions|],
  optional: [p | 0.20 ≤ occurrences(p) < 0.50],
  unique: [p | occurrences(p) < 0.20],

  output: {
    common: {parameters: common, percentage: avg(occurrences)},
    optional: {parameters: optional, percentage: avg(occurrences)},
    unique: {parameters: unique, count: |unique|}
  }
}

extract_helpers :: (File, Categories) → Helper_Functions
extract_helpers(F, C) = {
  standard_params: create_function("StandardToolParameters", C.common),

  merge_params: create_function("MergeParameters", {
    signature: "(specific map[string]Property) → map[string]Property",
    body: "result := StandardToolParameters(); for k, v := range specific { result[k] = v }; return result"
  }),

  schema_builder: create_function("buildToolSchema", {
    signature: "(properties map[string]Property, required ...string) → ToolSchema",
    body: "schema := ToolSchema{Type: \"object\", Properties: MergeParameters(properties)}; if len(required) > 0 { schema.Required = required }; return schema"
  }),

  tool_builder: create_function("buildTool", {
    signature: "(name, description string, properties map[string]Property, required ...string) → Tool",
    body: "return Tool{Name: name, Description: description, InputSchema: buildToolSchema(properties, required...)}"
  })
}

refactor_usages :: (File, Helpers) → Refactored_File
refactor_usages(F, H) = {
  usages: find_refactoring_candidates(F),

  refactored_count: 0,

  ∀usage ∈ usages →
    if ¬is_exception(usage) then
      before_lines: count_lines(usage),
      refactored: apply_builder_pattern(usage, H),
      after_lines: count_lines(refactored),
      reduction: before_lines - after_lines,
      refactored_count += 1,
      test_result: run_tests(),
      if ¬test_result.pass then
        rollback(usage),
        mark_as_exception(usage),

  return {file: F, refactored: refactored_count, exceptions: find_exceptions(F)}
}

apply_builder_pattern :: (Usage, Helpers) → Refactored_Code
apply_builder_pattern(U, H) = {
  specific_params: extract_unique_params(U),

  if ¬has_unique_structure(U) then
    generate_code("buildTool(" + U.name + ", " + U.description + ", map[string]Property{" + specific_params + "})")
  else
    keep_original(U, reason: "unique_structure")
}

verify_equivalence :: (Original, Refactored) → Test_Result
verify_equivalence(O, R) = {
  baseline: run_tests(O),
  after: run_tests(R),

  result:
    if after.pass_count = baseline.pass_count ∧ after.all_pass then
      {behavioral_equivalence: true, tests_pass: true}
    else
      {behavioral_equivalence: false, tests_fail: true, diff: compare(baseline, after)}
}

output :: Extraction_Result → Report
output(E) = {
  extraction_summary: {
    extracted_helpers: [
      {name: "StandardToolParameters", lines: |lines|, reused_by: count(reusers)},
      {name: "buildToolSchema", lines: |lines|, reused_by: count(reusers)},
      {name: "buildTool", lines: |lines|, reused_by: count(reusers)}
    ],

    refactored_usages: [
      {tool: name, before_lines: n1, after_lines: n2, reduction_percentage: (n1-n2)/n1}
      | ∀tool ∈ refactored
    ],

    exceptions: [
      {tool: name, reason: explanation, refactored: false}
      | ∀tool ∈ exceptions
    ],

    metrics: {
      total_lines_before: sum(before),
      total_lines_after: sum(after),
      lines_reduced: sum(before) - sum(after),
      reduction_percentage: (sum(before) - sum(after)) / sum(before)
    }
  },

  test_results: {
    status: if all_pass then "PASS" else "FAIL",
    baseline_pass_count: baseline.count,
    after_pass_count: after.count,
    behavioral_equivalence: equivalence_check
  },

  quality_checks: {
    duplication_before: duplication_ratio_before,
    duplication_after: duplication_ratio_after,
    duplication_reduction: (before - after) / before,
    maintainability_improvement: describe_impact
  }
}

constraints :: Extraction → Bool
constraints(E) =
  ∀refactoring ∈ E:
    incremental(refactoring) ∧
    test_after_each(refactoring) ∧
    behavioral_equivalence(refactoring) ∧
    reduction_percentage(E) ≥ 0.10 ∧
    duplication_reduction(E) ≥ 0.50 ∧
    allow_exceptions(special_cases) ∧
    document_exceptions(with_rationale)
