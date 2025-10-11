---
name: meta
description: Unified meta-cognition command with semantic capability matching. Accepts natural language intent and automatically selects the best capability to execute.
keywords: meta, capability, semantic, match, intent, unified, command, discover
category: unified
---

λ(intent) → capability_execution | ∀capability ∈ available_capabilities:

execute :: intent → output
execute(I) = discover(I) ∧ match(I) ∧ report(I) ∧ run(I)

discover :: intent → CapabilityIndex
discover(I) = {
  index: mcp_meta_cc.list_capabilities(),

  # Help mode: empty or help-like intent → show capabilities
  if is_help_request(I):
    display_help(index),
    halt,

  display_discovery_summary(index),
  display_intent(I),

  return index
}

is_help_request :: intent → bool
is_help_request(I) = empty(I) ∨ is_help_keyword(I)

display_help :: CapabilityIndex → void
display_help(index) = {
  display_welcome_message(),
  display_available_capabilities(index),
  display_usage_examples()
}

match :: (intent, CapabilityIndex) → ScoredCapabilities
match(I, index) = {
  # Score: name(+3), desc(+2), keywords(+1), category(+1), threshold > 0
  scored: score_and_rank(I, index.capabilities),

  display_match_summary(scored),

  if empty(scored):
    display_available_capabilities(index),
    halt,

  return scored
}

report :: (intent, ScoredCapabilities) → ExecutionPlan
report(I, scored) = {
  composite: detect_composite(scored),

  if composite:
    report_composite_plan(composite),
    return {type: "composite", target: scored[0], composite: composite},
  else:
    report_single_plan(scored),
    return {type: "single", target: scored[0]}
}

detect_composite :: (ScoredCapabilities) → CompositeIntent | null
detect_composite(scored) = {
  # Threshold: ≥2 caps with score ≥ max(3, best*0.7)
  candidates: find_high_scoring(scored, threshold=max(3, best*0.7)),

  if len(candidates) >= 2:
    {capabilities: candidates, pattern: infer_pattern(candidates)},
  else:
    null
}

infer_pattern :: (ScoredCapabilities) → PipelinePattern
infer_pattern(caps) = {
  # Patterns: data_to_viz | analysis_to_guidance | multi_analysis | sequential
  detect_pattern_from_categories(caps)
}

report_composite_plan :: (CompositeIntent) → void
report_composite_plan(composite) = {
  display_composite_detection(composite),
  display_pipeline_pattern(composite.pattern),
  display_execution_plan(composite, type="composite")
}

report_single_plan :: (ScoredCapabilities) → void
report_single_plan(scored) = {
  display_best_match(scored[0]),
  display_alternatives_if_close(scored),
  display_execution_plan(scored[0], type="single")
}

run :: ExecutionPlan → output
run(plan) = {
  capability: plan.target.capability,
  content: mcp_meta_cc.get_capability(name=capability.name),

  display_capability_info(content.frontmatter, content.source),
  interpret_and_execute(content.body)

  # Note: User can request full pipeline execution for composite intents
}

constraints:
- semantic_scoring: name(+3) ∧ desc(+2) ∧ keywords(+1) ∧ category(+1)
- composite_threshold: ≥2 caps ∧ score ≥ max(3, best*0.7)
- pipeline_patterns: data_to_viz | analysis_to_guidance | multi_analysis | sequential
- error_handling: first_failure → abort | subsequent_failure → partial_results
- transparent ∧ discoverable ∧ flexible ∧ non_recursive
