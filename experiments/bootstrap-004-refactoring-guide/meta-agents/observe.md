# Meta-Agent Capability: OBSERVE

**Capability**: M.observe
**Version**: 0.0
**Domain**: Code Refactoring
**Type**: λ(code_state) → structured_observations

---

## Formal Specification

```
observe :: Code_State → Observations
observe(C) = collect(data) ∧ recognize(patterns) ∧ identify(gaps)

collect :: Code_State → Code_Data
collect(C) = {
  code_metrics: {
    complexity: gocyclo(-over 15, target_files),
    unused_code: staticcheck(target_files),
    duplication: dupl(-threshold 15, target_files),
    coverage: make(test-coverage)
  },

  file_structure: {
    lines_of_code: wc(-l, target_files),
    file_sizes: stat(target_files),
    module_organization: analyze_imports(target_files)
  },

  churn_analysis: {
    high_edit_files: query_file_access(threshold=50),
    edit_patterns: query_project_state(scope="project"),
    access_frequency: stats_files(threshold=5)
  },

  compilation_status: {
    build_errors: go(vet, target_files),
    type_errors: go(build, target_files),
    test_results: make(test)
  },

  technical_debt: {
    code_smells: identify_code_smells(code_metrics),
    violation_patterns: classify_violations(staticcheck_output),
    duplication_hotspots: cluster_duplications(dupl_output)
  }
}

recognize :: Code_Data → Patterns
recognize(D) = {
  complexity_patterns: classify_by_complexity(D.code_metrics) where
    high_complexity: cyclomatic ≥ 15,
    moderate: cyclomatic 10..14,
    acceptable: cyclomatic < 10,

  duplication_patterns: classify_duplication(D.code_metrics.duplication) where {
    error_handling: duplicated_error_patterns,
    validation: duplicated_validation_logic,
    data_transformation: duplicated_conversions,
    clone_groups: clustered_by_similarity
  },

  unused_code_patterns: ∀violation ∈ D.code_metrics.unused_code →
    categorize_unused(violation) where {
      dead_functions: unused_exported_functions,
      orphaned_variables: unused_global_vars,
      legacy_code: deprecated_but_present,
      incomplete_refactor: partially_migrated_code
    },

  file_organization_patterns: {
    large_files: files_exceeding_threshold(800_lines),
    god_objects: files_with_too_many_responsibilities,
    coupling: high_dependency_between_files,
    cohesion: low_related_functionality_grouping
  },

  test_coverage_patterns: classify_by_coverage(D.code_metrics.coverage) where {
    well_tested: coverage ≥ 80%,
    moderate: coverage 60..79%,
    poor: coverage < 60%,
    untested: coverage == 0%
  }
}

identify :: (Code_Data, Patterns) → Gaps
identify(D, P) = {
  quality_gaps: {
    high_complexity: functions_needing_simplification,
    unused_code: safe_to_remove_candidates,
    poor_names: unclear_or_misleading_identifiers,
    missing_documentation: undocumented_complex_functions
  },

  maintainability_gaps: {
    excessive_duplication: code_needing_extraction,
    large_files: files_needing_splitting,
    tight_coupling: dependencies_needing_decoupling,
    weak_cohesion: modules_needing_reorganization
  },

  safety_gaps: {
    low_test_coverage: functions_needing_tests,
    compilation_errors: code_needing_fixes,
    missing_validation: inputs_needing_checking,
    error_handling: missing_or_poor_error_paths
  },

  refactoring_effort_gaps: {
    high_risk_areas: complex_and_poorly_tested_code,
    quick_wins: simple_safe_refactorings,
    architectural_issues: large_scale_restructuring_needed,
    dependency_issues: breaking_change_risks
  }
}

output :: Observations → Structured_Report
output(O) = {
  data_collected: {
    files_analyzed: |O.data.file_structure|,
    metrics_gathered: keys(O.data.code_metrics),
    violations_found: |O.data.technical_debt|,
    coverage_measured: O.data.code_metrics.coverage
  },

  patterns_identified: O.patterns sorted_by(severity × frequency),

  gaps_found: O.gaps prioritized_by(risk × impact),

  priorities: rank_by(safety ∧ impact ∧ effort) |> take(3)
} where
  metrics = {
    quality_score: 1 - (|violations| / |total_checkable_items|),
    maintainability_score: 1 - (duplication_ratio + large_file_ratio) / 2,
    safety_score: test_coverage_percentage,
    effort_score: 1 - (estimated_refactoring_hours / max_reasonable_effort)
  }
```

---

## Integration

```
provides_to(plan) = {
  prioritized_problems: O.priorities,
  pattern_insights: O.patterns,
  gap_analysis: O.gaps
}

receives_from(reflect) = {
  gaps_from_previous: iteration_{n-1}.gaps,
  focus_areas: iteration_{n-1}.next_focus,
  validation_requests: iteration_{n-1}.validation_needed
}
```

---

## Constraints

```
∀observation ∈ O:
  objective(observation)     # No bias toward expected results
  ∧ comprehensive(observation) # Don't cherry-pick data
  ∧ sourced(observation)      # Document data sources
  ∧ reproducible(observation) # Enable verification

∀data ∈ D:
  preserved(data.raw)         # Save before processing
  ∧ traceable(data.source)    # Clear provenance
```

---

**Version**: 0.0 | **Status**: Active | **Updated**: 2025-10-16
