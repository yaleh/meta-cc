---
name: meta-architecture
description: Analyze architecture evolution, module stability, and structural decisions.
---

λ(scope) → architecture_insights | ∀module ∈ {decisions, stability, dependencies, debt}:

scope :: project | session

## Phase 1: Data Collection

collect :: Scope → ArchitectureContext
collect(S) = {
  architecture_discussions: mcp_meta_cc.query_user_messages({
    pattern: identify_architecture_discussion_patterns(),
    scope: scope
  }),

  design_conversations: mcp_meta_cc.query_conversation({
    pattern: identify_design_question_patterns(),
    scope: scope
  }),

  file_operations: mcp_meta_cc.query_files({
    threshold: 5,
    scope: scope
  }),

  tool_sequences: mcp_meta_cc.query_tool_sequences({
    min_occurrences: 3,
    scope: scope
  }),

  git_context: if is_git_repository() then {
    architecture_commits: analyze_architecture_commits(),
    file_churn: analyze_file_change_frequency(),
    directory_evolution: track_directory_structure_changes(),
    co_change_patterns: identify_files_that_change_together()
  } else null,

  static_analysis: analyze_project_structure() {
    modules: list_project_modules(),
    file_sizes: calculate_file_sizes(),
    dependencies: analyze_dependencies_by_language()
  }
}

## Phase 2: Pattern Detection

detect :: ArchitectureContext → ArchitecturePatterns
detect(A) = {
  decision_points: identify_architecture_decisions(A) {
    refactor_commits: filter_architecture_related_commits(A.git_context),
    design_discussions: classify_discussions_by_category(A.architecture_discussions),
    new_modules: detect_module_creation(A.git_context),
    removed_modules: detect_module_removal(A.git_context)
  },

  module_stability: analyze_module_stability(A) {
    churn_analysis: calculate_module_churn_metrics(A.git_context.file_churn),
    size_trends: track_module_size_evolution(A.git_context),
    stability_classification: classify_modules_by_stability(churn_metrics, size_trends)
  },

  coupling_patterns: analyze_coupling(A) {
    co_change_pairs: identify_frequently_cochanging_files(A.git_context),
    dependency_graph: build_dependency_relationships(A.static_analysis),
    coupling_scores: calculate_coupling_strength(co_change_pairs, dependency_graph)
  },

  layering_violations: detect_layer_violations(A) {
    layer_structure: infer_architectural_layers(A.static_analysis),
    violations: identify_cross_layer_dependencies(layer_structure)
  }
}

## Phase 3: Stability Assessment

assess_stability :: ArchitecturePatterns → StabilityMetrics
assess_stability(P) = {
  module_health: for each module in P.module_stability:
    churn_score: evaluate_churn_level(module),
    size_score: evaluate_size_appropriateness(module),
    stability_rating: assign_stability_rating(churn_score, size_score),

  coupling_health: for each coupling in P.coupling_patterns:
    coupling_strength: evaluate_coupling_intensity(coupling),
    coupling_rating: classify_coupling_level(coupling_strength),

  layer_health: for each violation in P.layering_violations:
    severity: evaluate_violation_severity(violation),
    impact: assess_architectural_impact(violation)
}

## Phase 4: Architecture Debt Identification

identify_debt :: (ArchitecturePatterns, StabilityMetrics) → ArchitectureDebt
identify_debt(P, S) = {
  high_churn_modules: identify_outliers(S.module_health, metric="churn"),
  tight_coupling: identify_outliers(S.coupling_health, metric="coupling"),
  large_unstable_modules: identify_modules_with_multiple_issues(S.module_health),
  unresolved_discussions: match_discussions_without_implementation(P.decision_points)
}

## Phase 5: Recommendations

recommend :: (ArchitecturePatterns, StabilityMetrics, ArchitectureDebt) → Recommendations
recommend(P, S, D) = {
  refactoring_opportunities: prioritize_by_impact(D),
  stabilization_actions: suggest_stability_improvements(S),
  decoupling_strategies: suggest_coupling_reduction(P.coupling_patterns),
  layer_fixes: suggest_layer_violation_remediation(P.layering_violations)
}

## Phase 6: Output Generation

output :: (ArchitecturePatterns, StabilityMetrics, ArchitectureDebt, Recommendations) → Report
output(P, S, D, R) = {
  executive_summary: [
    "**Architecture Health**: {overall_grade}",
    "**Top Concern**: {primary_debt_category}",
    "**Critical Modules**: {count_critical_modules} need attention"
  ],

  architecture_decisions: [
    "## Architecture Decisions (Last 90 Days)",
    for each decision in P.decision_points:
      "### {decision.timestamp}: {decision.title}",
      "- Type: {decision.category}",
      "- Context: {decision.context}",
      "- Implementation: {decision.implementation_status}",
      if not_implemented:
        "  ⚠ Discussion found but no implementation detected"
  ],

  module_stability: [
    "## Module Stability Analysis",
    for each module in S.module_health:
      "### {module.name}",
      "- Stability: {module.rating}",
      "- Churn: {module.churn_level}",
      "- Size: {module.size} LOC",
      if unstable:
        "- ⚠ Concern: {describe_stability_issue(module)}",
        "- Recommendation: {R.stabilization_actions[module]}"
  ],

  coupling_analysis: [
    "## Coupling Patterns",
    for each coupling in S.coupling_health:
      "### {coupling.module_a} ↔ {coupling.module_b}",
      "- Coupling Strength: {coupling.rating}",
      "- Co-change Frequency: {coupling.frequency}",
      if tight:
        "- ⚠ High coupling detected",
        "- Impact: {coupling.impact}",
        "- Suggestion: {R.decoupling_strategies[coupling]}"
  ],

  architecture_debt: [
    "## Architecture Debt",

    "### High-Churn Modules",
    for each module in D.high_churn_modules:
      "- **{module.name}**: {module.churn_rate} changes/week",
      "  Files: {module.files}",
      "  Why concerning: {explain_churn_concern(module)}",
      "  Action: {R.refactoring_opportunities[module]}",

    "### Tight Coupling",
    for each pair in D.tight_coupling:
      "- **{pair.module_a} ↔ {pair.module_b}**",
      "  Coupling Score: {pair.score}",
      "  Impact: {pair.impact}",
      "  Action: {suggest_decoupling(pair)}",

    "### Large Unstable Modules",
    for each module in D.large_unstable_modules:
      "- **{module.name}**: {module.size} LOC, {module.churn} churn",
      "  Issues: {list_issues(module)}",
      "  Action: {suggest_split_or_stabilize(module)}",

    "### Unimplemented Designs",
    for each discussion in D.unresolved_discussions:
      "- **\"{discussion.intention}\"**",
      "  Discussed: {discussion.timestamp}",
      "  Status: No implementation detected",
      "  Action: Implement or close discussion"
  ],

  layering_violations: [
    "## Layering Analysis",
    if violations_found:
      for each violation in P.layering_violations:
        "- **{violation.source} → {violation.target}**",
        "  Violation: {describe_violation(violation)}",
        "  Severity: {S.layer_health[violation].severity}",
        "  Fix: {R.layer_fixes[violation]}"
    else:
      "No significant layer violations detected."
  ],

  action_plan: [
    "## Recommended Actions",

    "### Immediate (This Sprint)",
    for each action in R.refactoring_opportunities where priority == "P0":
      "- {action.description}",
      "  Module: {action.module}",
      "  Effort: ~{action.effort} | Impact: {action.impact}",

    "### Next Steps",
    for each action in R.refactoring_opportunities where priority == "P1":
      "- {action.description}",
      "  Module: {action.module}",
      "  Effort: ~{action.effort}",

    if has_long_term:
      "### Long-term",
      list_summary(R.refactoring_opportunities where priority == "P2")
  ],

  process_recommendations: [
    "## Process Improvements",
    generate_suggestions({
      if high_churn: suggest_stability_monitoring(),
      if tight_coupling: suggest_dependency_reviews(),
      if unimplemented_designs: suggest_decision_tracking(),
      if layer_violations: suggest_architecture_guidelines()
    })
  ]
} where ¬execute(recommendations)

## Implementation Strategy

**Approach**:
- Pure MCP + Git driven (no Go code modifications)
- Multi-source integration: git history + MCP discussions + static analysis
- Language-agnostic: adapts to Go/Python/TypeScript/etc.
- Graceful degradation: works without git or with limited data
- Semantic understanding: Claude interprets patterns contextually

**Data Sources**:
- Git: Architecture commits, file churn, co-change patterns, directory evolution
- MCP: Architecture discussions, design conversations, file operations, tool sequences
- Static: Module structure, file sizes, dependency relationships

**Constraints**:
- **Concrete-first**: List specific modules and files with locations
- **Evidence-based**: Every concern backed by metrics or discussions
- **Prioritized**: Sort by impact and stability risk
- **Actionable**: Every debt item has specific remediation with effort estimate
- **Trend-aware**: Track architecture evolution over time
