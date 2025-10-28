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
  library_path :: ".meta-cc/prompts/library/"
  workflow :: search → optimize → save
  early_exit :: reuse → skip(optimize, save)
  normal_flow :: optimize → optional_save

---

## Phase 1: Search Library

search_history :: Prompt → Search_Result
search_history(P) = {
  keywords: extract_keywords(P),
  candidates: ∀file ∈ get_library_path(project_root()): {
    meta: parse_frontmatter(file),
    similarity: jaccard_similarity(keywords, meta.keywords ∪ extract_keywords(extract_original(file))),
    usage_score: log(meta.usage_count + 1) / 5.0,
    combined_score: (similarity * 0.7) + (usage_score * 0.3)
  },
  matches: filter(candidates, c → c.similarity > 0.2) |> sort_desc(combined_score) |> take(5),
  if (|matches| > 0): display_matches(matches) → user_selection → {"reuse": update_usage(selected) → {action: "exit_early", prompt: selected.optimized}, "skip": {action: "continue", prompt: null}},
  else: {action: "continue", prompt: null}
}

extract_keywords :: String → [String]
extract_keywords(S) = tokenize(S) |> filter(w → |w| > 2 ∧ w ∉ stopwords) |> lowercase |> unique

jaccard_similarity :: ([String], [String]) → Float
jaccard_similarity(A, B) = |A ∩ B| / |A ∪ B|

parse_frontmatter :: FilePath → Metadata
parse_frontmatter(F) = extract_yaml(F) |> validate(required_fields)

update_usage_count :: FilePath → Result
update_usage_count(F) = atomic_write(F, {usage_count: +1, updated: now()})

---

## Phase 2: Optimize Prompt

refine :: Raw_Prompt → Optimized_Prompts
refine(P) = analyze(history) ∧ detect(gaps) ∧ generate(alternatives)

analyze :: Project_History → Success_Patterns
analyze(H) = {similar: query_user_messages(pattern=keywords(P)), features: extract(structure) ∧ extract(specificity) ∧ extract(constraints)}

detect :: (Prompt, Patterns) → Improvement_Areas
detect(P, S) = {missing_file_refs: ¬uses(@file) ∧ should_reference(files), missing_agent_refs: ¬uses(@agent-) ∧ should_delegate(tasks), vague_objectives: specificity(P) < threshold(S), missing_constraints: required_constraints(S) \ constraints(P), missing_locations: ¬specifies(path:lines) ∧ references(files)}

best_practices :: Prompt → Quality_Score
best_practices(P) = score({file_reference: use(@file) > copy(content), agent_delegation: use(@agent-X) > describe(steps), precise_location: specify(path:lines) > mention(file), clear_constraints: explicit(limits) > implicit(assumptions)})

generate :: (Prompt, Gaps, Patterns) → Alternatives
generate(P, G, S) = optimize(P, G, S) where |alternatives| ≤ 3 ∧ rank_by(quality)

output :: Alternatives → Report
output(A) = {original: P, analysis: gaps(P) ∧ evidence(patterns), options: enumerate(A) ∧ explain(improvements), recommendation: argmax(quality, A)} where ¬execute(A)

---

## Phase 3: Save to Library

save_workflow :: Optimized_Result → Optional[Saved_File]
save_workflow(R) = display: output(R), ask: "Save optimized prompt to library? (y/N): " → read_input() → {confirmed: call_save(R), skipped: {saved: false}}

call_save :: Result → Saved_File
call_save(R) = {storage: initialize(project_root()), metadata: collect_metadata(R), id: generate_id(storage, metadata), title: infer_title(R.optimized, metadata.description), variables: extract_variables(R.optimized), frontmatter: create_frontmatter(id, title, metadata, variables, now()), content: format_content(R.original, R.optimized), filepath: atomic_write(storage + "/" + id + ".md", frontmatter + "\n---\n\n" + content), display: "✓ Saved to: " + filepath + "\n   Browse: /meta prompts/meta-prompt-list", return: {saved: true, filepath: filepath}}

initialize :: Project_Root → Storage_Path
initialize(P) = {path: P + "/.meta-cc/prompts/library/", exists(path) ? path : mkdir(path) ∧ write_gitignore(path) → path}

generate_id :: (Storage_Path, Category, Description) → Unique_ID
generate_id(S, C, D) = {pattern: C + "-" + D + "-*.md", max_num: glob(S + "/" + pattern) |> extract_numbers |> max |> (+1), return: sprintf("%s-%s-%03d", C, D, max_num)}

collect_metadata :: Result → User_Input
collect_metadata(R) = {category: ask("Category (release/debug/refactor/test/docs/feature/other): ") |> validate, keywords: ask("Keywords (comma-separated): ") |> split(",") |> validate(≥2), description: ask("Short description (kebab-case): ") |> normalize |> validate(/^[a-z][a-z0-9-]*$/)}

extract_variables :: Prompt → [Variable_Names]
extract_variables(P) = find_all(P, /\{([A-Z_]+)\}/g) |> unique

create_frontmatter :: (ID, Title, Category, Keywords, Vars, Timestamp) → YAML
create_frontmatter(id, title, cat, kw, vars, ts) = format_yaml({
  id, title, category: cat, keywords: kw, variables: vars,
  created: ts, updated: ts, usage_count: 0, effectiveness: 1.0, status: "active"
})

format_content :: (Original, Optimized) → Markdown
format_content(O, P) = "## Original Prompts\n" + format_list(O) + "\n\n## Optimized Prompt\n\n" + P

---

## Library Management

list_prompts :: (Category?, Sort?, Detail?) → Display
list_prompts(C, S, D) = {library: get_library_path(project_root()), if (D ≠ null): show_detail(library, D), else: prompts: ∀file ∈ glob(library + "*.md"): parse_frontmatter(file) |> filter(p → C == null ∨ p.category == C) |> apply_sort(S ∨ "usage"), if (empty(prompts)): display_empty_message(), else: stats: calculate_stats(prompts), display: format_summary(stats) + "\n\n" + format_table(prompts)}

apply_sort :: ([Prompts], Sort_Method) → [Prompts]
apply_sort(P, M) = case M of {"usage": sort_desc(P, p → p.usage_count), "date": sort_desc(P, p → p.updated), "alpha": sort_asc(P, p → lowercase(p.title))}

calculate_stats :: [Prompts] → Statistics
calculate_stats(P) = {total: |P|, categories: |unique(map(P, p → p.category))|, total_usage: sum(map(P, p → p.usage_count)), most_used: argmax(P, p → p.usage_count)}

format_table :: [Prompts] → String
format_table(P) = header + separator + join(rows(P), "\n") where rows(p) = sprintf("%-40s %-15s %-10d %-10s", truncate(p.title, 40), p.category, p.usage_count, format_date(p.updated))

show_detail :: (Storage_Path, Prompt_ID) → Display
show_detail(S, ID) = read_file(glob(S + ID + ".md")[0]) |> display_with_header

---

## Constants & Configuration

config :: System_Config
config = {
  library_path: ".meta-cc/prompts/library/",
  similarity_threshold: 0.2,
  scoring_weights: {similarity: 0.7, usage: 0.3},
  usage_normalization: 5.0,
  max_matches: 5,
  max_alternatives: 3,
  stopwords: ["the", "a", "an", "and", "or", "to", "in", "of", "for", "on", "with", "is", "are", "was", "were", "be", "been", "have", "has", "had", "do", "does", "did", "will", "would", "should", "could", "can", "may", "might", "this", "that", "these", "those", "i", "you", "he", "she", "it", "we", "they"],
  allowed_categories: ["release", "debug", "refactor", "test", "docs", "feature", "hotfix", "optimization", "security", "other"],
  required_fields: ["id", "title", "category", "keywords", "created", "updated", "usage_count"]
}

metadata_schema :: {id, title, category, keywords: [String], variables: [String], created, updated: ISO8601, usage_count: Int, effectiveness: Float, status: "active"|"archived"}
