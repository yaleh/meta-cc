---
name: meta-prompt-save
description: Internal capability to save optimized prompts to project library
category: internal
---

λ(prompt_original, prompt_optimized) → saved_file | ∀field ∈ required_metadata:

initialize :: Project_Root → Storage_Path
initialize(P) = {
  storage: P + "/.meta-cc/prompts/library/",

  if (¬exists(storage)):
    create: mkdir -p storage,
    gitignore: write(storage + ".gitignore", "# Local prompt library\n# Commit selectively to share with team\n*.md\n"),
    confirm: "Initialized prompt library at " + storage,

  else:
    confirm: "Using existing library at " + storage,

  return: storage
}

generate_id :: (Storage_Path, Category, Description) → Unique_ID
generate_id(S, C, D) = {
  pattern: C + "-" + D + "-*.md",
  existing: glob(S + "/" + pattern),

  if (empty(existing)):
    next_id: "001",

  else:
    ids: extract_numbers(existing, /(\d{3})\.md$/),
    max_id: max(ids),
    next_id: sprintf("%03d", max_id + 1),

  return: next_id
}

collect_metadata :: (Original, Optimized) → User_Input
collect_metadata(O, P) = {
  ask_category: "What category is this prompt? (release/debug/refactor/test/docs/feature/other): ",
  category: read_input() ∨ "other",
  validate: category ∈ allowed_categories,

  ask_keywords: "Enter keywords (comma-separated, mix of languages): ",
  keywords_raw: read_input(),
  keywords: split(keywords_raw, /\s*,\s*/),
  validate: |keywords| ≥ 2,

  ask_description: "Short description (2-4 words, kebab-case): ",
  description: read_input(),
  normalize: lowercase(replace(description, /\s+/, "-")),
  validate: description =~ /^[a-z][a-z0-9-]*$/,

  return: {category, keywords, description}
}

infer_title :: (Optimized_Prompt, Description) → Title_String
infer_title(P, D) = {
  # Extract first line or heading if present
  first_line: extract_first_meaningful_line(P),

  # Convert description to title case
  title_from_desc: title_case(replace(D, /-/, " ")),

  # Prefer inferred from content, fallback to description
  title: first_line ∨ title_from_desc,
  truncate: substring(title, 0, 80),

  return: truncate
}

extract_variables :: Optimized_Prompt → [Variable_Names]
extract_variables(P) = {
  # Detect placeholders like {VERSION}, {BRANCH}, {FILE}
  pattern: /\{([A-Z_]+)\}/g,
  matches: find_all(P, pattern),
  variables: unique(matches),

  return: variables
}

create_frontmatter :: (ID, Title, Category, Keywords, Vars, Timestamp) → YAML_String
create_frontmatter(id, title, cat, kw, vars, ts) = {
  yaml: format_yaml({
    id: id,
    title: title,
    category: cat,
    keywords: kw,
    created: ts,
    updated: ts,
    usage_count: 0,
    effectiveness: 1.0,
    variables: vars,
    status: "active"
  }),

  return: yaml
}

format_content :: (Original, Optimized) → Markdown_String
format_content(O, P) = {
  sections: [
    "## Original Prompts",
    format_original(O),
    "",
    "## Optimized Prompt",
    "",
    P
  ],

  markdown: join(sections, "\n"),
  return: markdown
}

format_original :: Original_Prompt → Markdown_List
format_original(O) = {
  # Handle both single string and array
  items: is_array(O) ? O : [O],

  list: map(items, item → "- " + item),
  return: join(list, "\n")
}

save_file :: (Storage, Frontmatter, Content, Filename) → Result
save_file(S, F, C, N) = {
  filepath: S + "/" + N,

  file_content: "---\n" + F + "\n---\n\n" + C + "\n",

  write: atomic_write(filepath, file_content),

  confirm: {
    success: "✓ Saved prompt to: " + filepath,
    usage: "Reuse it next time with '/meta Refine prompt: ...'",
    manage: "Browse your library with '/meta prompts/meta-prompt-list'"
  },

  return: {success: true, filepath: filepath}
}

## Complete Workflow

workflow :: (Original, Optimized) → Saved_File
workflow(O, P) = {
  # Step 1: Initialize storage
  display: "Initializing prompt library...",
  storage: initialize(project_root()),

  # Step 2: Collect metadata from user
  display: "\nTo save this prompt, we need some metadata:",
  metadata: collect_metadata(O, P),

  # Step 3: Generate unique ID
  id: generate_id(storage, metadata.category, metadata.description),
  full_id: metadata.category + "-" + metadata.description + "-" + id,

  # Step 4: Infer title and extract variables
  title: infer_title(P, metadata.description),
  variables: extract_variables(P),

  # Step 5: Create frontmatter
  timestamp: now_iso8601(),
  frontmatter: create_frontmatter(
    full_id,
    title,
    metadata.category,
    metadata.keywords,
    variables,
    timestamp
  ),

  # Step 6: Format content
  content: format_content(O, P),

  # Step 7: Save file
  filename: full_id + ".md",
  result: save_file(storage, frontmatter, content, filename),

  # Step 8: Confirm to user
  display: result.confirm,

  return: result
}

## User Interaction

user_prompts :: String → String
user_prompts(prompt) = {
  display: prompt,
  input: read_line(),
  trim: strip_whitespace(input),

  return: trim
}

## Validation

validate_category :: String → Bool
validate_category(C) = {
  allowed: ["release", "debug", "refactor", "test", "docs", "feature", "hotfix", "optimization", "security", "other"],

  valid: C ∈ allowed,

  if (¬valid):
    display: "Invalid category. Using 'other'.",
    return: false,

  return: true
}

validate_description :: String → Bool
validate_description(D) = {
  valid: D =~ /^[a-z][a-z0-9-]*$/ ∧ length(D) ≥ 3 ∧ length(D) ≤ 40,

  if (¬valid):
    display: "Invalid description format. Use lowercase, hyphens, 3-40 chars.",
    return: false,

  return: true
}

validate_keywords :: [String] → Bool
validate_keywords(K) = {
  valid: |K| ≥ 2 ∧ |K| ≤ 10,

  if (¬valid):
    display: "Please provide 2-10 keywords.",
    return: false,

  return: true
}

## Error Handling

error_handling :: Operation → Result
error_handling(Op) = {
  try:
    result: Op(),
    return: result,

  catch (PermissionError):
    display: "Error: Cannot write to .meta-cc/ directory. Check permissions.",
    suggest: "Try: chmod 755 .meta-cc/",
    return: {success: false, error: "permission_denied"},

  catch (DiskFullError):
    display: "Error: Insufficient disk space.",
    return: {success: false, error: "disk_full"},

  catch (InvalidInput):
    display: "Error: Invalid input provided. Please try again.",
    return: {success: false, error: "invalid_input"}
}

## Constraints

constraints:
- non_intrusive: save is optional, called only when user confirms
- idempotent: can retry without side effects
- atomic: writes use temp file + rename to prevent corruption
- validated: all inputs validated before writing
- informative: clear confirmation messages with usage tips
