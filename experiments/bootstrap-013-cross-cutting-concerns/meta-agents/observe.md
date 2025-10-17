# Meta-Agent Capability: OBSERVE

**Capability**: M.observe
**Version**: 0.0
**Domain**: Error Recovery
**Type**: λ(error_state) → structured_observations

---

## Formal Specification

```
observe :: Error_State → Observations
observe(E) = collect(data) ∧ recognize(patterns) ∧ identify(gaps)

collect :: Error_State → Error_Data
collect(E) = {
  history: mcp_meta_cc.query_tools(status="error", scope="project"),

  tool_errors: {
    bash: query_tools(tool="Bash", status="error"),
    read: query_tools(tool="Read", status="error"),
    edit: query_tools(tool="Edit", status="error"),
    write: query_tools(tool="Write", status="error")
  },

  patterns: mcp_meta_cc.query_tool_sequences(scope="project"),

  context: {
    messages: query_user_messages(pattern=".*error.*"),
    sequences: query_tool_sequences(pattern=".*error.*"),
    files: query_file_access(errors_present=true)
  }
}

recognize :: Error_Data → Patterns
recognize(D) = {
  frequency: classify_by_frequency(D.history) where
    high: count(error) > 10,
    rare: count(error) ≤ 10 ∧ critical,

  category: classify_by_type(D.history) where {
    environment: file_access ∨ permission ∨ network,
    syntax: invalid_command ∨ malformed_input,
    logic: state_inconsistency ∨ wrong_assumption,
    integration: external_service ∨ dependency
  },

  tool_specific: ∀tool ∈ D.tool_errors →
    analyze_failure_modes(tool),

  temporal: {
    trends: ∂(error_rate)/∂t,
    clusters: identify_time_clusters(D.history),
    cascades: detect_error_chains(D.patterns)
  },

  impact: classify_by_impact(D.history) where {
    blocking: stops_workflow,
    degrading: reduces_quality,
    silent: unnoticed_failure,
    cascading: triggers_more_errors
  }
}

identify :: (Error_Data, Patterns) → Gaps
identify(D, P) = {
  detection: {
    undetected_types: find_silent_failures(D),
    late_detection: find_delayed_catches(D),
    missing_reporting: find_unreported(D)
  },

  diagnosis: {
    unknown_causes: ∀e ∈ D.history | ¬has_root_cause(e),
    inaccurate: ∀e ∈ D.history | misdiagnosed(e),
    missing_tools: required_diagnostics ∖ available_tools
  },

  recovery: {
    no_procedure: ∀e ∈ D.history | ¬has_recovery(e),
    unclear_steps: ∀e ∈ D.history | ambiguous_recovery(e),
    no_automation: manual_only_recoveries(D)
  },

  prevention: {
    recurring: ∀e ∈ D.history | count(e) > 1 ∧ preventable(e),
    missing_checks: required_validations ∖ implemented_checks,
    insufficient_guards: weak_safeguards(D)
  }
}

output :: Observations → Structured_Report
output(O) = {
  data_collected: {
    error_count: |O.data.history|,
    time_span: [min(timestamp), max(timestamp)],
    tools_analyzed: keys(O.data.tool_errors)
  },

  patterns_identified: O.patterns sorted_by(frequency × impact),

  gaps_found: O.gaps prioritized_by(severity),

  priorities: rank_by(impact ∧ frequency ∧ addressability) |> take(3)
} where
  metrics = {
    error_rate: |errors| / |total_operations|,
    coverage: |detected_types| / |total_types|,
    trend: Δ(error_rate) / Δt
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
