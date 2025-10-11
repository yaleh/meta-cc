---
name: meta
description: Unified meta-cognition command with semantic capability matching. Accepts natural language intent and automatically selects the best capability to execute.
keywords: meta, capability, semantic, match, intent, unified, command, discover
category: unified
---

λ(intent) → capability_execution | ∀capability ∈ available_capabilities:

# Discover and match
discover :: void → CapabilityIndex
discover() = mcp_meta_cc.list_capabilities()

match :: (intent, CapabilityIndex) → ScoredCapabilities
match(I, C) = {
  # Score capabilities: name(+3), description(+2), keywords(+1), category(+1)
  # Return sorted by score descending, threshold > 0
  score_and_rank(I, C.capabilities)
}

# Composite detection
detect_composite :: (ScoredCapabilities) → CompositeIntent | null
detect_composite(scored) = {
  # Detect ≥2 capabilities with score ≥ max(3, best * 0.7)
  candidates: find_high_scoring(scored, threshold=max(3, best*0.7)),

  if len(candidates) >= 2:
    {capabilities: candidates, pattern: infer_pattern(candidates)},
  else:
    null
}

infer_pattern :: (ScoredCapabilities) → PipelinePattern
infer_pattern(caps) = {
  # Infer from categories:
  # - data_to_viz: diagnostics/analysis → visualization
  # - analysis_to_guidance: diagnostics → guidance/coaching
  # - multi_analysis: multiple diagnostics
  # - sequential: default fallback
  detect_pattern_from_categories(caps)
}

# Execution
execute :: (capability_name) → output
execute(name) = {
  content: mcp_meta_cc.get_capability(name),
  display_capability_info(content.frontmatter, content.source),
  interpret_and_execute(content.body)
}

execute_pipeline :: (CompositeIntent) → output
execute_pipeline(composite) = {
  # Order by pattern, execute sequentially
  # Error handling: first failure aborts, subsequent show partial results
  ordered_caps: order_by_pattern(composite),
  execute_sequential_with_error_handling(ordered_caps)
}

order_by_pattern :: (CompositeIntent) → [Capability]
order_by_pattern(composite) = {
  pattern: composite.pattern.type,

  if pattern == "data_to_viz":
    [find_non_viz(composite.capabilities), find_viz(composite.capabilities)],
  else if pattern == "analysis_to_guidance":
    [find_analysis(composite.capabilities), find_guidance(composite.capabilities)],
  else:
    composite.capabilities
}

# Main workflow
main :: intent → void
main(I) = {
  # Step 1: Discover and match
  index: discover(),
  scored: match(I, index),

  # Step 2: Report matching results
  report_matching_results(I, index, scored),

  # Step 3: Determine execution plan
  if empty(scored):
    display_available_capabilities(index),
    return,

  composite: detect_composite(scored),

  if composite:
    report_composite_plan(composite),
    execute_best_match(scored[0]),  # User can request full pipeline
  else:
    report_single_execution_plan(scored),
    execute(scored[0].capability.name)
}

report_matching_results :: (intent, CapabilityIndex, ScoredCapabilities) → void
report_matching_results(I, index, scored) = {
  # Display: total capabilities loaded, intent, match count, top matches
  display_discovery_summary(index),
  display_intent(I),
  display_match_summary(scored)
}

report_composite_plan :: (CompositeIntent) → void
report_composite_plan(composite) = {
  # Display: composite detection, pipeline pattern, execution plan
  display_composite_detection(composite),
  display_pipeline_pattern(composite.pattern),
  display_execution_plan(composite, type="composite")
}

report_single_execution_plan :: (ScoredCapabilities) → void
report_single_execution_plan(scored) = {
  # Display: best match, score, alternatives if any, execution plan
  display_best_match(scored[0]),
  display_alternatives_if_close(scored),
  display_execution_plan(scored[0], type="single")
}

main($1)

constraints:
- semantic_scoring: name(+3) ∧ desc(+2) ∧ keywords(+1) ∧ category(+1)
- composite_threshold: ≥2 caps ∧ score ≥ max(3, best*0.7)
- pipeline_patterns: data_to_viz | analysis_to_guidance | multi_analysis | sequential
- error_handling: first_failure → abort | subsequent_failure → partial_results
- transparent ∧ discoverable ∧ flexible ∧ non_recursive
