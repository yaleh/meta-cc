---
name: meta-prompt-list
description: List and manage saved prompts from library
category: internal
---

λ(category, sort, detail) → formatted_list | ∀prompt ∈ library:

category :: Optional[String]  # Filter by category (e.g., "release", "debug")
sort :: Optional[String]       # Sort order: "usage" | "date" | "alpha" (default: usage)
detail :: Optional[String]     # Show full content for specific prompt ID

## Library Location

get_library_path :: Project_Root → Storage_Path
get_library_path(P) = {
  path: P + "/.meta-cc/prompts/library/",

  if (¬exists(path)):
    return: null,

  return: path
}

## List All Prompts

list_prompts :: (Storage_Path, Category_Filter) → [Prompt_Metadata]
list_prompts(S, C) = {
  # Get all prompt files
  files: glob(S + "*.md"),

  if (empty(files)):
    return: [],

  # Parse metadata from each file
  prompts: ∀file ∈ files: {
    meta: parse_frontmatter(file),  # From meta-prompt-utils

    # Apply category filter if specified
    if (C ≠ null ∧ meta.category ≠ C):
      skip: true,

    # Extract data for display
    result: {
      id: meta.id,
      title: meta.title,
      category: meta.category,
      keywords: meta.keywords,
      usage_count: meta.usage_count,
      created: meta.created,
      updated: meta.updated,
      status: meta.status,
      file: file
    }
  },

  return: filter(prompts, p → p.skip ≠ true)
}

## Sorting

apply_sort :: ([Prompt_Metadata], Sort_Method) → [Prompt_Metadata]
apply_sort(P, M) = {
  sort_method: M ∨ "usage",  # Default to usage

  sorted: case sort_method of {
    "usage": sort_desc(P, p → p.usage_count),
    "date": sort_desc(P, p → p.updated),
    "alpha": sort_asc(P, p → lowercase(p.title)),
    _: sort_desc(P, p → p.usage_count)  # Default fallback
  },

  return: sorted
}

## Summary Statistics

calculate_stats :: [Prompt_Metadata] → Statistics
calculate_stats(P) = {
  total_prompts: |P|,

  # Count unique categories
  unique_categories: unique(map(P, p → p.category)),
  category_count: |unique_categories|,

  # Total usage across all prompts
  total_usage: sum(map(P, p → p.usage_count)),

  # Find most used prompt
  most_used: argmax(P, p → p.usage_count),

  return: {
    total_prompts: total_prompts,
    categories: category_count,
    total_usage: total_usage,
    most_used: most_used ∨ null
  }
}

## Formatted Output

format_table :: [Prompt_Metadata] → Display_String
format_table(P) = {
  # Table header
  header: sprintf("%-40s %-15s %-10s %-10s", "Title", "Category", "Usage", "Updated"),
  separator: repeat("-", 80),

  # Table rows
  rows: ∀p ∈ P: {
    # Truncate title if too long
    title_display: truncate(p.title, 40),

    # Format date (show only date part)
    date_display: format_date(p.updated, "%Y-%m-%d"),

    # Format row
    row: sprintf(
      "%-40s %-15s %-10d %-10s",
      title_display,
      p.category,
      p.usage_count,
      date_display
    )
  },

  # Combine all parts
  table: join([header, separator] + rows, "\n"),

  return: table
}

format_summary :: Statistics → Display_String
format_summary(S) = {
  lines: [
    sprintf("Total prompts: %d", S.total_prompts),
    sprintf("Categories: %d", S.categories),
    sprintf("Total usage: %d", S.total_usage)
  ],

  # Add most used prompt if available
  if (S.most_used ≠ null):
    lines: lines + [
      sprintf("Most used: %s (%d uses)",
        S.most_used.title,
        S.most_used.usage_count)
    ],

  summary: join(lines, "\n"),
  return: summary
}

## Detail View

show_detail :: (Storage_Path, Prompt_ID) → Display_String
show_detail(S, ID) = {
  # Find file matching ID
  pattern: S + ID + ".md",
  files: glob(pattern),

  if (empty(files)):
    return: sprintf("Error: Prompt '%s' not found in library.", ID),

  file: first(files),

  # Read entire file (frontmatter + content)
  content: read_file(file),

  # Format with header
  display: [
    sprintf("=== Prompt Details: %s ===", ID),
    "",
    content
  ],

  return: join(display, "\n")
}

## Empty Library Handling

handle_empty :: Unit → Display_String
handle_empty() = {
  message: [
    "Your prompt library is empty.",
    "",
    "To get started:",
    "1. Use '/meta Refine prompt: XXX' to optimize a prompt",
    "2. When prompted, choose to save it to your library",
    "3. The system will suggest saved prompts for similar queries",
    "",
    "Learn more: /meta prompts/meta-prompt-save"
  ],

  return: join(message, "\n")
}

## Complete Workflow

workflow :: (Category, Sort, Detail) → Display_Result
workflow(C, S, D) = {
  # Step 1: Locate library
  library_path: get_library_path(project_root()),

  # Case 1: Library doesn't exist
  if (library_path == null):
    display: handle_empty(),
    return: {success: true},

  # Case 2: Show detail view for specific prompt
  if (D ≠ null):
    display: show_detail(library_path, D),
    return: {success: true},

  # Case 3: List prompts (default)
  # Step 2: List prompts with optional category filter
  prompts: list_prompts(library_path, C),

  # Case 3a: No prompts (or no matches after filtering)
  if (empty(prompts)):
    if (C ≠ null):
      display: sprintf("No prompts found in category '%s'.", C),
    else:
      display: handle_empty(),
    return: {success: true},

  # Step 3: Sort prompts
  sorted: apply_sort(prompts, S),

  # Step 4: Calculate statistics
  stats: calculate_stats(sorted),

  # Step 5: Format output
  summary: format_summary(stats),
  table: format_table(sorted),

  # Step 6: Display results
  display: [
    "=== Prompt Library ===",
    "",
    summary,
    "",
    table,
    "",
    "Usage:",
    "  View details: /meta prompts/meta-prompt-list detail=<id>",
    "  Filter: /meta prompts/meta-prompt-list category=<category>",
    "  Sort: /meta prompts/meta-prompt-list sort=usage|date|alpha"
  ],

  output: join(display, "\n"),
  display: output,

  return: {success: true}
}

## Bash Implementation Helpers

bash_list_prompts :: (Storage_Path, Category_Filter) → JSONL
bash_list_prompts(S, C) = {
  script: '
    LIBRARY_PATH="$1"
    CATEGORY_FILTER="$2"

    # Check if library exists
    if [ ! -d "$LIBRARY_PATH" ]; then
      exit 0
    fi

    # Process each file
    for file in "$LIBRARY_PATH"/*.md; do
      [ -f "$file" ] || continue

      # Parse frontmatter using yq (requires yq to be installed)
      # Alternative: use awk/sed if yq not available
      META=$(sed -n "/^---$/,/^---$/p" "$file" | sed "1d;\$d")

      # Extract fields (using grep/awk for simple parsing)
      ID=$(echo "$META" | grep "^id:" | cut -d: -f2- | xargs)
      TITLE=$(echo "$META" | grep "^title:" | cut -d: -f2- | xargs)
      CATEGORY=$(echo "$META" | grep "^category:" | cut -d: -f2- | xargs)
      USAGE=$(echo "$META" | grep "^usage_count:" | cut -d: -f2- | xargs)
      UPDATED=$(echo "$META" | grep "^updated:" | cut -d: -f2- | xargs)
      STATUS=$(echo "$META" | grep "^status:" | cut -d: -f2- | xargs)

      # Apply category filter
      if [ -n "$CATEGORY_FILTER" ] && [ "$CATEGORY" != "$CATEGORY_FILTER" ]; then
        continue
      fi

      # Output JSONL
      printf "{\"id\":\"%s\",\"title\":\"%s\",\"category\":\"%s\",\"usage_count\":%d,\"updated\":\"%s\",\"status\":\"%s\",\"file\":\"%s\"}\n" \
        "$ID" "$TITLE" "$CATEGORY" "$USAGE" "$UPDATED" "$STATUS" "$file"
    done
  '
}

bash_sort :: (JSONL, Sort_Method) → JSONL
bash_sort(J, M) = {
  # Using jq for sorting
  usage_sort: "jq -s 'sort_by(.usage_count) | reverse | .[]'",
  date_sort: "jq -s 'sort_by(.updated) | reverse | .[]'",
  alpha_sort: "jq -s 'sort_by(.title | ascii_downcase) | .[]'",

  command: case M of {
    "usage": usage_sort,
    "date": date_sort,
    "alpha": alpha_sort,
    _: usage_sort  # Default
  }
}

## Constraints

constraints:
- efficient: O(n log n) for sorting, acceptable for <1000 prompts
- graceful: handle empty library, no matches, missing files
- informative: show statistics and usage hints
- flexible: support filtering, sorting, detail view
- non_intrusive: read-only operations, no modifications
