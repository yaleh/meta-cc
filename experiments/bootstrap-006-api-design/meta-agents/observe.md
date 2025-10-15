# Meta-Agent Capability: OBSERVE

**Capability**: M.observe
**Version**: 0.0
**Domain**: API Design
**Type**: λ(api_state) → structured_observations

---

## Formal Specification

```
observe :: API_State → Observations
observe(A) = collect(data) ∧ recognize(patterns) ∧ identify(gaps)

collect :: API_State → API_Data
collect(A) = {
  tool_schemas: read(docs/guides/mcp.md) ∧ extract(tool_definitions),

  tool_usage: mcp_meta_cc.query_tools(scope="project"),

  usage_patterns: {
    sequences: query_tool_sequences(scope="project"),
    frequency: query_tools(scope="project") |> group_by(tool_name),
    parameters: query_tools(scope="project") |> extract(parameters_used)
  },

  api_implementation: {
    tools: read(internal/tools/tools.go),
    capabilities: read(internal/capabilities/capabilities.go),
    mcp_server: read(cmd/mcp.go)
  },

  user_feedback: {
    requests: query_user_messages(pattern="query|tool|MCP|parameter|API"),
    issues: query_tools(status="error", scope="project"),
    usage_confusion: query_user_messages(pattern="how.*parameter|what.*option")
  }
}

recognize :: API_Data → Patterns
recognize(D) = {
  usage_frequency: classify_by_usage(D.tool_usage) where
    heavily_used: usage_count > 50,
    moderate: usage_count 10..50,
    rarely_used: usage_count < 10,

  naming_patterns: classify_naming(D.tool_schemas) where {
    query_prefix: tools starting with "query_",
    get_prefix: tools starting with "get_",
    action_verbs: tools with action names,
    inconsistencies: naming_deviations
  },

  parameter_patterns: ∀tool ∈ D.tool_schemas →
    analyze_parameter_design(tool) where {
      required_vs_optional: parameter_necessity,
      default_values: presence_and_consistency,
      naming: snake_case ∨ camelCase,
      types: parameter_type_consistency
    },

  response_patterns: {
    inline_vs_file: classify_output_modes(D.usage_patterns),
    error_formats: analyze_error_structures(D.user_feedback.issues),
    data_structures: identify_response_schemas(D.tool_schemas)
  },

  usability_issues: classify_by_usability(D.user_feedback) where {
    parameter_confusion: unclear_parameters,
    missing_features: requested_capabilities,
    inconsistent_behavior: unexpected_results,
    error_clarity: unclear_error_messages
  }
}

identify :: (API_Data, Patterns) → Gaps
identify(D, P) = {
  usability_gaps: {
    unclear_parameters: parameters_lacking_clear_purpose,
    missing_defaults: optional_params_without_defaults,
    poor_documentation: insufficient_parameter_descriptions,
    complex_interactions: multi_parameter_confusion
  },

  consistency_gaps: {
    naming_inconsistencies: deviations_from_patterns,
    parameter_naming: snake_case_vs_camelCase_mixing,
    response_format_variations: inconsistent_output_structures,
    error_message_variations: non_uniform_error_formats
  },

  completeness_gaps: {
    missing_features: user_requested_capabilities ∖ implemented,
    incomplete_parameters: tools_missing_common_options,
    edge_case_handling: unhandled_input_scenarios,
    missing_validation: insufficient_input_checking
  },

  evolvability_gaps: {
    no_versioning: lack_of_version_strategy,
    breaking_changes_risk: backward_incompatible_patterns,
    unclear_deprecation: missing_deprecation_policy,
    migration_difficulty: hard_to_upgrade_patterns
  }
}

output :: Observations → Structured_Report
output(O) = {
  data_collected: {
    tool_count: |O.data.tool_schemas|,
    usage_samples: |O.data.tool_usage|,
    files_analyzed: keys(O.data.api_implementation),
    user_feedback_count: |O.data.user_feedback|
  },

  patterns_identified: O.patterns sorted_by(usage_frequency × usability_impact),

  gaps_found: O.gaps prioritized_by(severity),

  priorities: rank_by(impact ∧ usage ∧ addressability) |> take(3)
} where
  metrics = {
    consistency_score: |consistent_patterns| / |total_patterns|,
    usability_score: 1 - (|usability_issues| / |total_usage|),
    completeness_score: |implemented_features| / |requested_features|
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

**Version**: 0.0 | **Status**: Active | **Updated**: 2025-10-14
