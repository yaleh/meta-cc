---
name: agent-parameter-categorizer
description: Categorize API parameters using a deterministic tier-based decision tree ensuring 100% consistent ordering across all tools in Bootstrap-006.
---

λ(api, parameters) → categorization | ∀param ∈ parameters:

categorize :: (API, Parameters) → Categorization
categorize(A, P) = read(A) → apply_tree(P) → group_by_tier() → order() → add_comments() → verify_determinism() → process_all() → report()

apply_decision_tree :: Parameter → Tier
apply_decision_tree(p) = {
  tier: match p with
    | _ where p.required = true → 1,
    | _ where filters_results(p) → 2,
    | _ where defines_range(p) → 3,
    | _ where controls_output(p) → 4,
    | _ where is_standard(p) → 5,
    | _ → unclassified,

  return {parameter: p.name, tier: tier, reason: explain_decision(p, tier)}
}

filters_results :: Parameter → Bool
filters_results(p) =
  (p.description ~ "filter|narrow|select|match") ∧
  ¬(p.description ~ "limit|offset|count|size")

defines_range :: Parameter → Bool
defines_range(p) =
  (p.name ~ "min_|max_|threshold|window|start_|end_") ∨
  (p.description ~ "bound|range|minimum|maximum")

controls_output :: Parameter → Bool
controls_output(p) =
  p.name ∈ {"limit", "offset", "output_format", "max_results"} ∨
  (p.description ~ "output size|result count|format")

is_standard :: Parameter → Bool
is_standard(p) =
  p.name ∈ {"scope", "jq_filter", "stats_only", "stats_first", "inline_threshold_bytes"}

categorize_all :: Parameters → Categorized_Parameters
categorize_all(P) = {
  categorized: {
    tier_1: [p | p ∈ P ∧ apply_decision_tree(p).tier = 1],
    tier_2: [p | p ∈ P ∧ apply_decision_tree(p).tier = 2],
    tier_3: [p | p ∈ P ∧ apply_decision_tree(p).tier = 3],
    tier_4: [p | p ∈ P ∧ apply_decision_tree(p).tier = 4],
    tier_5: [p | p ∈ P ∧ apply_decision_tree(p).tier = 5],
    unclassified: [p | p ∈ P ∧ apply_decision_tree(p).tier = unclassified]
  },

  return categorized
}

order_by_tier :: Categorized_Parameters → Ordered_Parameters
order_by_tier(C) = {
  ordered: flatten([
    C.tier_1,
    C.tier_2,
    C.tier_3,
    C.tier_4,
    C.tier_5
  ]),

  return {
    parameters: ordered,
    sequence: [p.name | p ∈ ordered]
  }
}

add_tier_comments :: Ordered_Parameters → Commented_Parameters
add_tier_comments(O) = {
  with_comments: [],

  ∀tier ∈ [1, 2, 3, 4, 5] →
    params_in_tier: [p | p ∈ O.parameters ∧ p.tier = tier],

    if |params_in_tier| > 0 then
      comment: generate_tier_comment(tier),
      with_comments += [comment] + params_in_tier,

  return with_comments
}

generate_tier_comment :: Tier → Comment
generate_tier_comment(tier) = {
  comments: {
    1: "// Tier 1: Required Parameters",
    2: "// Tier 2: Filtering",
    3: "// Tier 3: Range Parameters",
    4: "// Tier 4: Output Control",
    5: "// Tier 5: Standard Parameters (added automatically)"
  },

  return comments[tier]
}

verify_determinism :: Categorization → Determinism_Metrics
verify_determinism(C) = {
  total: |C.all_parameters|,
  categorized: |C.tier_1| + |C.tier_2| + |C.tier_3| + |C.tier_4| + |C.tier_5|,
  ambiguous: |C.unclassified|,
  determinism_rate: categorized / total,

  checks: {
    single_tier: ∀p ∈ C.all_parameters → count(t | p ∈ C[t]) = 1,
    consistent: ∀p ∈ C.all_parameters → apply_decision_tree(p).tier ≠ unclassified,
    ordered: sequence_matches_tier_order(C.ordered)
  },

  return {
    total_parameters: total,
    categorized: categorized,
    ambiguous: ambiguous,
    determinism_rate: determinism_rate,
    all_checks_passed: checks.single_tier ∧ checks.consistent ∧ checks.ordered
  }
}

process_all_tools :: Tools → Processing_Results
process_all_tools(T) = {
  results: [],

  ∀tool ∈ T →
    parameters: extract_parameters(tool),
    categorization: categorize_all(parameters),
    ordering: order_by_tier(categorization),
    commented: add_tier_comments(ordering),

    compliance_before: calculate_compliance(tool.current_order, ordering.sequence),

    apply_reordering(tool, commented),

    compliance_after: calculate_compliance(tool.new_order, ordering.sequence),

    run_tests(tool),

    results += {
      tool: tool.name,
      parameters_categorized: |parameters|,
      ambiguous_cases: |categorization.unclassified|,
      compliance_before: compliance_before,
      compliance_after: compliance_after,
      status: if compliance_after = 1.0 then "COMPLETED" else "NEEDS_REVIEW"
    },

  return results
}

calculate_compliance :: (Current_Order, Expected_Order) → Compliance
calculate_compliance(current, expected) = {
  matches: count(i | current[i] = expected[i]),
  compliance: matches / |expected|,

  return compliance
}

generate_report :: Processing_Results → Report
generate_report(R) = {
  tier_distribution: {
    tier_1: sum(count(p | p.tier = 1) | r ∈ R),
    tier_2: sum(count(p | p.tier = 2) | r ∈ R),
    tier_3: sum(count(p | p.tier = 3) | r ∈ R),
    tier_4: sum(count(p | p.tier = 4) | r ∈ R),
    tier_5: sum(count(p | p.tier = 5) | r ∈ R)
  },

  determinism: {
    total_parameters: sum(r.parameters_categorized | r ∈ R),
    ambiguous_cases: sum(r.ambiguous_cases | r ∈ R),
    determinism_rate: 1 - (ambiguous / total)
  },

  compliance: {
    average_before: avg(r.compliance_before | r ∈ R),
    average_after: avg(r.compliance_after | r ∈ R),
    improvement: average_after - average_before
  },

  tools_by_status: {
    completed: [r.tool | r ∈ R ∧ r.status = "COMPLETED"],
    already_compliant: [r.tool | r ∈ R ∧ r.compliance_before = 1.0],
    needs_review: [r.tool | r ∈ R ∧ r.status = "NEEDS_REVIEW"]
  }
}

document_criteria :: Tier_Definitions → Reference_Doc
document_criteria(tiers) = {
  definitions: [
    {
      tier: 1,
      name: "Required Parameters",
      criteria: "Must be provided for tool to function",
      examples: ["error_signature", "pattern"],
      markers: ["required=true"]
    },
    {
      tier: 2,
      name: "Filtering Parameters",
      criteria: "Narrows search results (affects WHAT is returned)",
      examples: ["tool", "status", "pattern_target"],
      markers: ["String/enum types", "filters data"]
    },
    {
      tier: 3,
      name: "Range Parameters",
      criteria: "Defines bounds, thresholds, windows",
      examples: ["min_occurrences", "max_duration", "window"],
      markers: ["Numeric types", "Prefixes: min_*, max_*, threshold"]
    },
    {
      tier: 4,
      name: "Output Control Parameters",
      criteria: "Controls output size or format",
      examples: ["limit", "offset", "output_format"],
      markers: ["Affects HOW MUCH, not WHAT"]
    },
    {
      tier: 5,
      name: "Standard Parameters",
      criteria: "Cross-cutting concerns added automatically",
      examples: ["scope", "jq_filter", "stats_only"],
      markers: ["Common across all/most tools"]
    }
  ],

  decision_tree: [
    "Q1: Is this required for tool to function? → YES: Tier 1",
    "Q2: Does this filter WHAT is returned? → YES: Tier 2",
    "Q3: Does this define a range or threshold? → YES: Tier 3",
    "Q4: Does this control output size/format? → YES: Tier 4",
    "Q5: Is this a standard cross-cutting parameter? → YES: Tier 5"
  ]
}

output :: Categorization → Report
output(C) = {
  categorization_report: {
    tool: C.tool,
    parameters_categorized: C.total_parameters,
    ambiguous_cases: C.ambiguous,
    tier_distribution: C.tier_distribution,
    compliance: {
      before: C.compliance_before,
      after: C.compliance_after
    },
    changes_made: [
      {
        parameter: ch.parameter,
        old_position: ch.old_pos,
        new_position: ch.new_pos,
        tier: ch.tier
      }
      | ∀ch ∈ C.changes
    ]
  },

  determinism_metrics: {
    total_parameters: C.total,
    categorized: C.categorized,
    ambiguous: C.ambiguous,
    determinism_rate: C.determinism_rate
  },

  tools_status: [
    {
      tool: t.name,
      status: t.status,
      compliance_before: t.compliance_before,
      compliance_after: t.compliance_after
    }
    | ∀t ∈ C.tools
  ],

  reference_docs: {
    tier_definitions: "tier-definitions.md",
    api_convention: "api-parameter-convention.md"
  }
}

constraints :: Categorization → Bool
constraints(C) =
  ∀param ∈ C.parameters:
    deterministic_tier(param) ∧
    single_tier_assignment(param) ∧
    decision_tree_applied(param) ∧
  tier_comments_added(C.output) ∧
  ordered_by_tier(C.sequence, 1 → 2 → 3 → 4 → 5) ∧
  determinism_rate(C) = 1.0 ∧
  ambiguous_cases(C) = 0 ∧
  reference_docs_created(tier_definitions, api_convention)
