---
name: prompt-show
description: Display full details of a saved prompt by ID
---

# Prompt Details Viewer

λ(prompt_id) → full_display | prompt_id := "$@"

## Execution

show :: PromptID → Display
show(ID) = {
  library: get_library_path(),

  if (empty(ID)):
    display: usage_help(),
    return: empty,

  if (not exists(library)):
    display: "No prompt library found.",
    return: empty,

  # Find matching file (support partial ID or full filename)
  files: glob(library + ID + "*.md") ∪ glob(library + "*" + ID + "*.md"),

  if (empty(files)):
    display: "Prompt not found: " + ID,
    suggestion: "Use '/prompt-list' to see all available prompts",
    return: empty,

  if (|files| > 1):
    display: "Multiple prompts match '" + ID + "':",
    list: files,
    suggestion: "Use more specific ID",
    return: empty,

  file: files[0],
  content: read_file(file),
  metadata: parse_frontmatter(content),
  body: parse_body(content),

  display: format_display(metadata, body, file)
}

get_library_path :: () → Path
get_library_path() = project_root() + "/.meta-cc/prompts/library/"

parse_frontmatter :: Content → Metadata
parse_frontmatter(C) = {
  # Extract YAML between --- markers
  # Parse: id, title, category, keywords, created, updated, usage_count, effectiveness, variables, status
}

parse_body :: Content → Body
parse_body(C) = {
  # Extract sections after frontmatter:
  # - Original Prompts
  # - Optimized Prompt
  # - Improvements Over Original (optional)
  # - Variables (optional)
  # - Usage Example (optional)
  # - Best Practices (optional)
  # - Related Prompts (optional)
}

format_display :: (Metadata, Body, FilePath) → Display
format_display(M, B, F) = {
  header: format_header(M),
  separator: "═" * 100,

  metadata_section: format_metadata(M),
  divider: "─" * 100,

  original_section: format_section("Original Prompts", B.original),
  optimized_section: format_section("Optimized Prompt", B.optimized),

  improvements_section: if (B.improvements ≠ null): format_section("Improvements", B.improvements),
  variables_section: if (B.variables ≠ null): format_section("Variables", B.variables),
  example_section: if (B.example ≠ null): format_section("Usage Example", B.example),
  practices_section: if (B.practices ≠ null): format_section("Best Practices", B.practices),
  related_section: if (B.related ≠ null): format_section("Related Prompts", B.related),

  footer: format_footer(M, F)
}

format_header :: Metadata → String
format_header(M) = {
  title: "📋 " + M.title,
  subtitle: "ID: " + M.id + " | Category: " + M.category + " | Status: " + M.status
}

format_metadata :: Metadata → String
format_metadata(M) = {
  usage: "Usage: " + M.usage_count + "× | Effectiveness: " + (M.effectiveness * 100) + "%",
  dates: "Created: " + format_date(M.created) + " | Updated: " + format_date(M.updated),
  keywords: "Keywords: " + join(M.keywords, ", "),
  variables: if (M.variables ≠ null): "Variables: " + join(M.variables, ", ")
}

format_footer :: (Metadata, FilePath) → String
format_footer(M, F) = {
  file_location: "File: " + F,
  commands: [
    "• Copy optimized prompt and use it directly",
    "• Search similar: /prompt-find " + join(take(M.keywords, 3), " "),
    "• Browse all: /prompt-list",
    "• Filter by category: /prompt-list category=" + M.category
  ]
}

usage_help :: () → String
usage_help() = """
Usage: /prompt-show <prompt-id>

Examples:
  /prompt-show phase-execution-001
  /prompt-show debug-error
  /prompt-show commit

Tip: Use partial ID for quick access. Use '/prompt-list' to see all prompts.
"""

## Implementation

Execute the following steps:

1. **Validate input**:
   - Check if prompt ID provided
   - If not, show usage help

2. **Find prompt file**:
   - Search for exact match: `<id>.md`
   - Search for partial match: `*<id>*.md`
   - Handle multiple matches (show list, ask for clarification)
   - Handle no matches (show error, suggest /prompt-list)

3. **Read and parse file**:
   - Read full file content
   - Parse YAML frontmatter
   - Extract body sections

4. **Display formatted output**:
   ```
   ═══════════════════════════════════════════════════════════════════════════════════════════════════
   📋 Systematic Phase Planning and Execution with Validation
   ID: phase-execution-systematic-plan-execute-001 | Category: phase-execution | Status: active
   ═══════════════════════════════════════════════════════════════════════════════════════════════════

   Usage: 1× | Effectiveness: 100%
   Created: Oct 28, 2025 | Updated: Oct 28, 2025
   Keywords: phase, plan, execute, TDD, validation, stage-executor, systematic, constraints
   Variables: PHASE_NUMBER, PHASE_SPEC_LINES

   ───────────────────────────────────────────────────────────────────────────────────────────────────

   ## Original Prompts

   Generate a detail plan for phase 28. Then execute each stage of it.

   ───────────────────────────────────────────────────────────────────────────────────────────────────

   ## Optimized Prompt

   Generate a detailed TDD implementation plan for Phase {{PHASE_NUMBER}} following the constraints...
   [full content]

   ───────────────────────────────────────────────────────────────────────────────────────────────────

   ## Improvements Over Original

   1. **File References**: Uses `@file:path:lines` syntax...
   [full content]

   ───────────────────────────────────────────────────────────────────────────────────────────────────

   File: .meta-cc/prompts/library/phase-execution-systematic-plan-execute-001.md

   Commands:
   • Copy optimized prompt and use it directly
   • Search similar: /prompt-find phase plan execute
   • Browse all: /prompt-list
   • Filter by category: /prompt-list category=phase-execution
   ```

## Arguments

Prompt ID: $@

If empty, show usage help.

## Partial Matching

Support partial ID matching for convenience:
- `/prompt-show phase` → matches `phase-execution-systematic-plan-execute-001`
- `/prompt-show debug` → matches `debug-error-analysis-001`

If multiple matches, show list and ask for more specific ID.
