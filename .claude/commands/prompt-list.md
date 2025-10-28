---
name: prompt-list
description: List all saved prompts with optional filtering and sorting
---

# Prompt Library Listing

λ(args) → formatted_list | args := "$@"

## Execution

list :: Args → Display
list(A) = {
  library: get_library_path(),

  if (not exists(library)):
    display: "No prompt library found. Save your first prompt with '/meta Refine prompt: <text>'",
    return: empty,

  files: glob(library + "*.md"),

  if (empty(files)):
    display: "Library is empty. Save your first prompt with '/meta Refine prompt: <text>'",
    return: empty,

  prompts: ∀file ∈ files: parse_frontmatter(file),

  filters: parse_args(A),
  filtered: apply_filters(prompts, filters),
  sorted: apply_sort(filtered, filters.sort),

  stats: calculate_stats(sorted),

  display: format_output(stats, sorted)
}

parse_args :: String → Filters
parse_args(A) = {
  # Parse optional arguments:
  # category=<cat>  - Filter by category
  # sort=usage|date|alpha  - Sort order (default: usage)
  # detail=<id>  - Show detailed view of specific prompt
}

apply_filters :: ([Prompts], Filters) → [Prompts]
apply_filters(P, F) = {
  if (F.category ≠ null):
    filter(P, p → p.category == F.category),
  else:
    P
}

apply_sort :: ([Prompts], SortMethod) → [Prompts]
apply_sort(P, S) = case S of {
  "usage": sort_desc(P, p → p.usage_count),
  "date":  sort_desc(P, p → p.updated),
  "alpha": sort_asc(P, p → lowercase(p.title)),
  _:       sort_desc(P, p → p.usage_count)  # Default: usage
}

calculate_stats :: [Prompts] → Statistics
calculate_stats(P) = {
  total: |P|,
  categories: |unique(map(P, p → p.category))|,
  total_usage: sum(map(P, p → p.usage_count)),
  most_used: argmax(P, p → p.usage_count)
}

format_output :: (Statistics, [Prompts]) → Display
format_output(S, P) = {
  header: format_stats(S),
  separator: "═" * 100,
  table_header: format_table_header(),
  table_rows: ∀p ∈ P: format_row(p),
  footer: format_footer()
}

format_row :: Prompt → String
format_row(P) = sprintf(
  "%-45s %-15s %-8d %-12s %s",
  truncate(P.title, 45),
  P.category,
  P.usage_count,
  format_date(P.updated),
  P.id
)

get_library_path :: () → Path
get_library_path() = project_root() + "/.meta-cc/prompts/library/"

## Implementation

Execute the following steps:

1. **Check library exists**:
   ```bash
   if [ ! -d .meta-cc/prompts/library/ ]; then
     echo "No prompt library found."
     exit 0
   fi
   ```

2. **Parse arguments** (optional):
   - `category=<name>`: Filter by category
   - `sort=usage|date|alpha`: Sort order
   - Default: sort by usage (most used first)

3. **Read all prompts**:
   - Glob all .md files in library
   - Parse frontmatter for each
   - Apply filters if specified
   - Sort by chosen method

4. **Calculate statistics**:
   - Total prompts
   - Unique categories
   - Total usage count
   - Most used prompt

5. **Display formatted output**:
   ```
   Prompt Library Summary
   ══════════════════════════════════════════════════════════════════════════════════════════════════
   Total Prompts: 4  |  Categories: 4  |  Total Uses: 4  |  Most Used: debug-error-analysis-001 (2×)
   ──────────────────────────────────────────────────────────────────────────────────────────────────
   Title                                         Category        Uses     Updated      ID
   ──────────────────────────────────────────────────────────────────────────────────────────────────
   Error Analysis and Debugging                  debug              2     Oct 27       debug-error-analysis-001
   Systematic Phase Planning and Execution       phase-exec         1     Oct 28       phase-execution-systematic-001
   Complete Git Commit Workflow                  git                1     Oct 27       commit-workflow-001
   Simple Release Process                        release            0     Oct 27       test-simple-release-001
   ──────────────────────────────────────────────────────────────────────────────────────────────────

   Commands:
   - View prompt: /prompt-show <id>
   - Search: /prompt-find <keywords>
   - Filter: /prompt-list category=<name>
   - Sort: /prompt-list sort=usage|date|alpha
   ```

## Usage Examples

```bash
# List all prompts (sorted by usage)
/prompt-list

# Filter by category
/prompt-list category=debug

# Sort by date (most recent first)
/prompt-list sort=date

# Sort alphabetically
/prompt-list sort=alpha

# Combine filters
/prompt-list category=release sort=date
```

## Arguments

Arguments: $@

Parse as key=value pairs:
- `category=<name>`: Filter by category
- `sort=usage|date|alpha`: Sort order
