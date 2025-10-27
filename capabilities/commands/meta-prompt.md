---
name: meta-prompt
description: Refine prompts using successful patterns from project history.
argument-hint: [prompt]
keywords: prompt, refinement, optimization, effectiveness, clarity
category: guidance
---

λ(prompt_raw) → prompt_refined | workflow:
  search_history(prompt_raw) →ᵉˣⁱᵗ reused_prompt
  ∨ (analyze(history) ∧ detect(gaps) ∧ generate(alternatives) ∧ output(alternatives))
  →ᵒᵖᵗ save_workflow(result) → saved_prompt

where:
  prompt_raw :: `$1`
  workflow :: search → optimize → save
  early_exit :: reuse → skip(optimize, save)
  normal_flow :: optimize → optional_save

---

## Workflow Phase 1: History Search

search_history(P) = {
  call: get_capability(name="meta-prompt-search", type="prompts"),
  pass: {query_prompt: P},
  result: workflow(P),

  if (result.action == "reuse"):
    display: result.message,
    display: "\n" + result.prompt,
    return: {action: "exit_early", prompt: result.prompt, reused: true},

  else:
    continue: phase_2_optimization,
    return: {action: "continue", prompt: null, reused: false}
}

---

## Workflow Phase 2: Generate Alternatives

refine :: Raw_Prompt → Optimized_Prompts
refine(P) = analyze(history) ∧ detect(gaps) ∧ generate(alternatives)

analyze :: Project_History → Success_Patterns
analyze(H) = {
  # query_successful_prompts does not exist - analyze user messages manually
  # successful: null,

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

---

## Workflow Phase 3: Save for Reuse

save_workflow :: Optimized_Result → Optional[Saved_File]
save_workflow(R) = {
  # First display the optimization results
  display: output(R),

  # Then offer to save
  ask: "\nWould you like to save this optimized prompt to your library? (y/N): ",
  default: "N",  # Non-intrusive - default is skip

  user_response: read_input(),

  if (user_confirms(user_response)):
    call: get_capability(name="meta-prompt-save", type="prompts"),
    pass: {
      prompt_original: R.original,
      prompt_optimized: R.recommendation,  # Best alternative
      timestamp: now()
    },
    result: save_result,

    display: "\n" + save_result.confirm,
    return: {saved: true, file: save_result.filepath},

  else:
    skip: "\nPrompt not saved. You can refine it again anytime with '/meta Refine prompt: ...'",
    return: {saved: false}
}

user_confirms :: String → Bool
user_confirms(R) = {
  normalized: lowercase(trim(R)),
  return: normalized ∈ ["y", "yes", "1", "true"]
}

---

## Workflow Integration

integrated_workflow :: Raw_Prompt → Final_Result
integrated_workflow(P) = {
  step_0: search_history(P),          # NEW (Stage 2): Check history first

  # If user selected historical prompt, exit early
  if (step_0.action == "exit_early"):
    return: {
      result: "reused",
      prompt: step_0.prompt,
      source: "library"
    },

  # Otherwise, continue with normal optimization
  step_1: analyze(history),           # Existing: analyze history patterns
  step_2: detect(P, patterns),        # Existing: detect improvement areas
  step_3: generate(P, gaps, patterns),# Existing: generate alternatives
  step_4: output(alternatives),       # Existing: display results
  step_5: save_workflow(result),      # Stage 1: optional save

  return: {
    result: "optimized",
    prompt: alternatives.recommendation,
    source: "generated"
  }
}

workflow_order :: Execution_Steps
workflow_order() = {
  phase_1: "Search History (Stage 2)",
  description_1: "Check for similar prompts in library, allow reuse or skip",

  phase_2: "Generate Alternatives (Original)",
  description_2: "Analyze patterns, detect gaps, generate optimizations",

  phase_3: "Save to Library (Stage 1)",
  description_3: "Optionally save optimized prompt for future reuse",

  exit_points: {
    early: "After reusing historical prompt (skip phases 2-3)",
    normal: "After generating and optionally saving new prompt"
  }
}

## Benefits of Saving Prompts

benefits :: Why_Save → Value_Proposition
benefits() = {
  efficiency: "Reuse proven prompts instead of recreating",
  consistency: "Maintain consistent approach across similar tasks",
  learning: "Build project-specific knowledge base over time",
  sharing: "Collaborate by committing useful prompts to git",
  discovery: "System will suggest saved prompts for similar queries"
}

## Post-Save Actions

post_save :: Saved_File → Next_Steps
post_save(F) = {
  immediate: {
    confirm: "Prompt saved successfully",
    location: F.filepath,
    id: F.id
  },

  future: {
    reuse: "System will automatically suggest this prompt for similar queries",
    browse: "View all saved prompts: '/meta prompts/meta-prompt-list'",
    share: "Commit to git to share with team: 'git add .meta-cc/prompts/library/'"
  }
}
