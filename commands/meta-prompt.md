---
description: Refine prompts using successful patterns from project history.
argument-hint: [prompt]
---

λ(prompt_raw) → prompt_refined | ∀pattern ∈ successful_history:

prompt_raw :: `$1`

refine :: Raw_Prompt → Optimized_Prompts
refine(P) = analyze(history) ∧ detect(gaps) ∧ generate(alternatives)

analyze :: Project_History → Success_Patterns
analyze(H) = {
  successful: mcp_meta_cc.query_successful_prompts(min_quality_score=0.8),

  similar: mcp_meta_cc.query_user_messages(pattern=keywords(P)),

  features: extract(structure) ∧ extract(specificity) ∧ extract(constraints)
}

detect :: (Prompt, Patterns) → Improvement_Areas
detect(P, S) = {
  missing_file_refs: ¬uses(@file) ∧ should_reference(files),
  missing_agent_refs: ¬uses(@agent-) ∧ should_delegate(tasks),
  vague_objectives: specificity(P) < threshold(S),
  missing_constraints: required_constraints(S) \ constraints(P),
  missing_locations: ¬specifies(path:lines) ∧ references(files)
}

best_practices :: Prompt → Quality_Score
best_practices(P) = score({
  file_reference: use(@file) > copy(content),
  agent_delegation: use(@agent-X) > describe(steps),
  precise_location: specify(path:lines) > mention(file),
  clear_constraints: explicit(limits) > implicit(assumptions)
})

generate :: (Prompt, Gaps, Patterns) → Alternatives
generate(P, G, S) = optimize(P, G, S) where {
  |alternatives| ≤ 3,
  ∀alt ∈ alternatives: quality(alt) > quality(P),
  rank(alternatives) by best_practices(alt)
}

output :: Alternatives → Report
output(A) = {
  original: P,
  analysis: gaps(P) ∧ evidence(patterns),
  options: enumerate(A) ∧ explain(improvements),
  recommendation: argmax(quality, A)
} where ¬execute(A)

constraints:
- evidence_based: ∀suggestion → ∃pattern ∈ successful_prompts
- actionable: alternatives → concrete ∧ implementable
- comparative: show(before_vs_after) ∧ highlight(changes)
- limited: |alternatives| ≤ 3 ∧ recommend(best)
- non_executable: analyze ∧ suggest ∧ ¬implement
