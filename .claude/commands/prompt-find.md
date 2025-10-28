---
name: prompt-find
description: Search for saved prompts in the library by keywords
---

# Prompt Library Search

λ(keywords) → search_results | keywords := "$@"

## Execution

search :: Keywords → Results
search(K) = {
  library: get_library_path(),

  if (not exists(library)):
    display: "No prompt library found. Save your first prompt with '/meta Refine prompt: <text>'",
    return: empty,

  files: glob(library + "*.md"),

  if (empty(files)):
    display: "Library is empty. Save your first prompt with '/meta Refine prompt: <text>'",
    return: empty,

  matches: ∀file ∈ files: {
    metadata: parse_frontmatter(file),
    content: read_file(file),
    score: calculate_match_score(K, metadata, content)
  } where score > 0,

  sorted: sort_desc(matches, m → m.score),

  display: format_results(sorted, K),

  if (|sorted| > 0):
    offer: "Show full prompt? Enter ID or 'q' to quit"
}

get_library_path :: () → Path
get_library_path() = project_root() + "/.meta-cc/prompts/library/"

parse_frontmatter :: FilePath → Metadata
parse_frontmatter(F) = {
  # Extract YAML frontmatter (lines 2-12 typically)
  # Fields: id, title, category, keywords, usage_count, updated
}

calculate_match_score :: (Keywords, Metadata, Content) → Score
calculate_match_score(K, M, C) = {
  keyword_matches: count_matches(K, M.keywords ∪ extract_text(M.title)),
  content_matches: count_matches(K, C.original_prompts),
  category_matches: count_matches(K, M.category),

  score: (keyword_matches * 3) + (content_matches * 2) + (category_matches * 1)
}

format_results :: ([Matches], Keywords) → Display
format_results(M, K) = {
  header: "Found " + |M| + " prompts matching: " + K,
  separator: "─" * 80,

  table: ∀match ∈ M: format_row(match),

  footer: "\nUse '/prompt-show <id>' to view full prompt"
}

format_row :: Match → String
format_row(M) = sprintf(
  "%-40s %-15s %-8s %s",
  M.metadata.id,
  M.metadata.category,
  "★" * min(5, M.score / 2),  # Star rating
  truncate(M.metadata.title, 50)
)

## Implementation

Execute the following steps:

1. **Check library exists**:
   ```bash
   if [ ! -d .meta-cc/prompts/library/ ]; then
     echo "No prompt library found."
     echo "Save your first prompt with: /meta Refine prompt: <your-prompt>"
     exit 0
   fi
   ```

2. **Search prompts**:
   - Read all .md files in library
   - For each file:
     - Parse frontmatter (id, title, category, keywords)
     - Extract original prompts section
     - Calculate match score based on keyword overlap
   - Sort by score descending
   - Display top matches

3. **Display results**:
   - Show table: ID | Category | Score | Title
   - Provide option to view full prompt with `/prompt-show <id>`

4. **Example output**:
   ```
   Found 3 prompts matching: phase execute plan
   ────────────────────────────────────────────────────────────────────────────────
   ID                                       Category        Score    Title
   ────────────────────────────────────────────────────────────────────────────────
   phase-execution-systematic-plan-001      phase-exec      ★★★★★    Systematic Phase Planning and Execution
   debug-error-analysis-001                 debug           ★★       Error Analysis and Debugging
   commit-workflow-001                      git             ★        Complete Git Commit Workflow

   Use '/prompt-show <id>' to view full prompt
   ```

## Keywords

Search keywords: $@

If empty, display usage:
```
Usage: /prompt-find <keywords>
Example: /prompt-find phase plan execute
```
