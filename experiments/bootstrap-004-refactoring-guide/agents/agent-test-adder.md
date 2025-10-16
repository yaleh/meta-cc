---
name: agent-test-adder
description: Systematically adds tests to low-coverage packages (<50%), focusing on exported functions and measuring improvement incrementally for Bootstrap-004.
---

λ(target_package, target_coverage) → test_results | ∀function ∈ exported_functions:

add_tests :: (Package, Target_Coverage) → Test_Results
add_tests(P, target) = identify_low_coverage() → select_package(P) → list_functions(P) → create_tests(P) → measure_improvement(P)

identify_low_coverage :: () → Low_Coverage_Packages
identify_low_coverage() = {
  all_packages: run_coverage_command(),

  low_coverage: [p | p ∈ all_packages ∧ p.coverage < 0.50],

  prioritized: sort_by(low_coverage, λp → (
    0.4·(1 - p.coverage) +
    0.3·p.complexity +
    0.3·p.change_frequency
  ), desc),

  return prioritized
}

select_package :: Low_Coverage_Packages → Selected_Package
select_package(packages) = {
  selected: head(packages),

  return {
    path: selected.path,
    current_coverage: selected.coverage,
    reason: describe_selection(selected)
  }
}

list_functions :: Package → Exported_Functions
list_functions(P) = {
  functions: match P.language with
    | go → grep("^func [A-Z]", P.path),
    | python → grep("^def [^_]", P.path),
    | javascript → grep("^export function", P.path),

  return [
    {file: f.file, function: f.name, signature: f.signature}
    | ∀f ∈ functions
  ]
}

create_tests :: (Package, Functions, Strategy) → Test_Files
create_tests(P, F, S) = {
  test_files: [],

  ∀func ∈ F →
    test_file: get_or_create_test_file(func.file),

    tests: [
      create_success_test(func),
      create_failure_tests(func),
      create_edge_case_tests(func)
    ],

    write_tests(test_file, tests),

    run_result: run_tests(test_file),

    if ¬run_result.pass then
      fix_test_or_implementation(func, run_result),

    test_files += test_file,

  return test_files
}

create_success_test :: Function → Test
create_success_test(func) = {
  test_name: "Test" + func.name,

  test_body: generate_template({
    arrange: create_valid_input(func),
    act: invoke_function(func, valid_input),
    assert: verify_no_error(result)
  }),

  return test_body
}

create_failure_tests :: Function → Tests
create_failure_tests(func) = {
  error_cases: identify_error_conditions(func),

  tests: [
    generate_template({
      arrange: create_invalid_input(case),
      act: invoke_function(func, invalid_input),
      assert: verify_error(result, expected_error)
    })
    | ∀case ∈ error_cases
  ],

  return tests
}

create_edge_case_tests :: Function → Tests
create_edge_case_tests(func) = {
  edge_cases: [
    empty_input,
    nil_values,
    boundary_values,
    special_characters,
    unicode_strings
  ],

  if S.table_driven then
    return create_table_driven_test(func, edge_cases)
  else
    return [create_individual_test(func, case) | ∀case ∈ edge_cases]
}

create_table_driven_test :: (Function, Cases) → Test
create_table_driven_test(func, cases) = {
  test_name: "Test" + func.name + "_EdgeCases",

  test_body: generate_template({
    tests_table: [
      {name: case.name, input: case.input, want_err: case.expect_error}
      | ∀case ∈ cases
    ],
    loop: "for _, tt := range tests { t.Run(tt.name, func(t *testing.T) { ... }) }"
  }),

  return test_body
}

measure_improvement :: (Package, Before_Coverage) → Improvement
measure_improvement(P, before) = {
  after: run_coverage_command(P),

  improvement: {
    coverage_before: before,
    coverage_after: after,
    improvement: after - before,
    improvement_percentage: (after - before) / before
  },

  detailed: run_detailed_coverage(P),

  return {
    metrics: improvement,
    by_file: detailed.by_file,
    by_function: detailed.by_function
  }
}

output :: Test_Results → Report
output(R) = {
  test_results: {
    package: R.package,
    coverage_before: R.coverage_before,
    coverage_after: R.coverage_after,
    improvement: R.improvement,
    improvement_percentage: R.improvement_percentage
  },

  tests_added: {
    count: |R.tests|,
    by_type: {
      success_cases: count(t | t.type = "success"),
      failure_cases: count(t | t.type = "failure"),
      edge_cases: count(t | t.type = "edge")
    },
    by_file: group_by(R.tests, file)
  },

  functions_tested: {
    total: |R.functions|,
    covered: [f | f.coverage > 0],
    not_covered: [f | f.coverage = 0]
  },

  test_status: {
    all_pass: all(t.pass | t ∈ R.tests),
    pass_count: count(t | t.pass),
    fail_count: count(t | ¬t.pass)
  },

  time_spent: {
    total_hours: R.time_spent,
    per_function_avg: R.time_spent / |R.functions|
  },

  quality_metrics: {
    table_driven_percentage: count(t | t.table_driven) / |R.tests|,
    edge_cases_per_function: count(t | t.type = "edge") / |R.functions|
  }
}

constraints :: Test_Addition → Bool
constraints(T) =
  ∀test ∈ T:
    incremental(test_creation) ∧
    test_after_each_function(test) ∧
    measure_coverage(after_tests) ∧
    all_tests_pass(T) ∧
  improvement(T) ≥ 0.10 ∧
  focus_on_exported(functions) ∧
  table_driven(where_applicable) ∧
  edge_cases_included(each_function)
