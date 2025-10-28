---
name: meta-prompt-search
description: Search for existing prompts in the library that match user query.
category: internal
internal: true
---

search_library :: Query_Prompt → Search_Result
search_library(Q) = {
  # Initialize library path
  library_path: get_library_root(),

  if (not exists(library_path)):
    return: {action: "continue", matches: [], message: "No library found. Will create on first save."},

  # Extract keywords from query
  keywords: extract_keywords(Q),

  # Search for matching prompts
  candidates: ∀file ∈ glob(library_path + "*.md"): {
    metadata: parse_frontmatter(file),
    content: read_file(file),
    original_prompts: extract_section(content, "Original Prompts"),
    similarity: calculate_similarity(keywords, metadata.keywords ∪ extract_keywords(original_prompts)),
    usage_score: normalize_usage(metadata.usage_count),
    combined_score: (similarity * 0.7) + (usage_score * 0.3)
  },

  # Filter and sort matches
  matches: filter(candidates, c → c.similarity > 0.2)
           |> sort_desc(combined_score)
           |> take(5),

  if (|matches| > 0):
    display: format_matches(matches, Q),
    return: {action: "select", matches: matches},
  else:
    return: {action: "continue", matches: [], message: "No similar prompts found."}
}

query_prompt :: String  # User's current prompt to optimize

## Search Algorithm

search :: Query_Prompt → [Ranked_Matches]
search(Q) = {
  # Step 1: Locate library
  library_path: project_root() + "/.meta-cc/prompts/library/",
  files: glob(library_path + "*.md"),

  # Step 2: Handle empty library
  if (¬exists(library_path) ∨ empty(files)):
    return: [],

  # Step 3: Extract keywords from query
  query_keywords: extract_keywords(Q),  # From meta-prompt-utils

  if (empty(query_keywords)):
    return: [],  # No meaningful keywords

  # Step 4: Calculate similarity for each candidate
  candidates: ∀file ∈ files: {
    meta: parse_frontmatter(file),  # From meta-prompt-utils

    # Combine metadata keywords + keywords from original prompts
    original_text: extract_original_prompts(file),
    original_keywords: extract_keywords(original_text),
    all_keywords: meta.keywords ∪ original_keywords,

    # Calculate Jaccard similarity
    similarity: jaccard_similarity(query_keywords, all_keywords),

    # Calculate usage score (logarithmic scaling, normalized to 0-1)
    # log(100) ≈ 4.6, so we divide by 5 to normalize
    usage_score: log(meta.usage_count + 1) / 5.0,

    # Combined score: 70% similarity, 30% usage frequency
    combined_score: (similarity * 0.7) + (usage_score * 0.3),

    # Store metadata for display
    result: {
      file: file,
      meta: meta,
      optimized: extract_optimized_prompt(file),
      similarity: similarity,
      usage_score: usage_score,
      combined_score: combined_score
    }
  },

  # Step 5: Filter by similarity threshold
  threshold: 0.2,  # 20% keyword overlap minimum
  matches: filter(candidates, c → c.similarity > threshold),

  # Step 6: Sort by combined score (descending)
  ranked: sort_desc(matches, m → m.combined_score),

  # Step 7: Take top N matches
  top_n: 5,
  results: take(ranked, top_n),

  return: results
}

## Content Extraction

extract_original_prompts :: FilePath → String
extract_original_prompts(F) = {
  # Read file content
  content: read_file(F),

  # Extract section between "## Original Prompts" and next "##" heading
  pattern: /## Original Prompts\n(.*?)\n##/s,
  original_section: regex_extract(content, pattern, group=1),

  # Parse list items (format: "- {prompt}")
  lines: split(original_section, "\n"),
  prompts: map(lines, line → strip_prefix(line, "- ")),

  # Combine all prompts into single text
  combined: join(prompts, " "),

  return: combined
}

extract_optimized_prompt :: FilePath → String
extract_optimized_prompt(F) = {
  # Read file content
  content: read_file(F),

  # Extract section after "## Optimized Prompt" to end of file
  pattern: /## Optimized Prompt\n+(.*)/s,
  optimized_section: regex_extract(content, pattern, group=1),

  # Trim whitespace
  trimmed: strip(optimized_section),

  return: trimmed
}

## Results Formatting

format_results :: [Matches] → Display_String
format_results(M) = {
  # Case 1: No matches found
  if (|M| == 0):
    message: "No similar prompts found in your library. Generating fresh recommendations...\n",
    return: message,

  # Case 2: Display matches
  else:
    header: sprintf("Found %d similar prompt(s) in your library:\n\n", |M|),

    list: ∀m ∈ M (indexed from 1): {
      # Format similarity as percentage
      similarity_pct: sprintf("%.0f%%", m.similarity * 100),

      # Format usage count
      usage_str: sprintf("%d use%s", m.meta.usage_count, m.meta.usage_count == 1 ? "" : "s"),

      # Truncate optimized prompt for preview
      preview: truncate(m.optimized, 100) + (length(m.optimized) > 100 ? "..." : ""),

      # Format line
      line: sprintf(
        "%d. %s [%s] (%s match, %s)\n   Preview: %s\n",
        index,
        m.meta.title,
        m.meta.category,
        similarity_pct,
        usage_str,
        preview
      )
    },

    footer: "\nSelect a number (1-5) to reuse, or press Enter to generate new optimization: ",

    output: header + join(list, "\n") + footer,

    return: output
}

## Reuse Workflow

reuse :: (User_Input, [Matches]) → Reuse_Result
reuse(I, M) = {
  # Normalize input
  input: strip(I),

  # Case 1: Skip (empty, 0, or 'n')
  if (input == "" ∨ input == "0" ∨ lowercase(input) == "n"):
    return: {
      action: "generate_new",
      prompt: null,
      message: "Generating new optimization..."
    },

  # Case 2: Valid selection (1-5 and within bounds)
  else if (is_integer(input) ∧ 1 <= int(input) <= |M|):
    # Get selected match (1-indexed)
    selected: M[int(input) - 1],

    # Update usage count and timestamp
    update_result: update_usage_count(selected.file),  # From meta-prompt-utils

    # Return reused prompt
    return: {
      action: "reuse",
      prompt: selected.optimized,
      file: selected.file,
      original_usage: selected.meta.usage_count,
      new_usage: update_result.new_count,
      message: sprintf(
        "✓ Reused prompt from: %s (usage count: %d → %d)",
        basename(selected.file),
        selected.meta.usage_count,
        update_result.new_count
      )
    },

  # Case 3: Invalid selection
  else:
    return: {
      action: "generate_new",
      prompt: null,
      message: sprintf("Invalid selection '%s'. Generating new optimization...", input)
    }
}

## Complete Workflow

workflow :: Query_Prompt → Search_Result
workflow(Q) = {
  # Step 1: Search library
  display: "Searching prompt library for similar patterns...",
  matches: search(Q),

  # Step 2: Format and display results
  results_display: format_results(matches),
  display: results_display,

  # Step 3: Handle user selection (if matches found)
  if (|matches| > 0):
    user_input: read_input(),
    result: reuse(user_input, matches),
    display: result.message,
    return: result,

  else:
    # No matches, continue to normal optimization
    return: {
      action: "generate_new",
      prompt: null,
      message: null
    }
}

## Bash Implementation Helpers

bash_search :: Query_Prompt → JSONL
bash_search(Q) = {
  script: '
    LIBRARY_PATH="$(git rev-parse --show-toplevel 2>/dev/null || pwd)/.meta-cc/prompts/library"

    # Check if library exists
    if [ ! -d "$LIBRARY_PATH" ]; then
      exit 0  # Empty result
    fi

    # Extract query keywords
    QUERY_KEYWORDS=$(echo "$1" | tr "[:upper:]" "[:lower:]" | tr -cs "[:alnum:]" "\n" | grep -E "^.{3,}$" | sort -u)

    # Process each file
    for file in "$LIBRARY_PATH"/*.md; do
      [ -f "$file" ] || continue

      # Parse frontmatter
      META=$(yq eval -o=json "." <(sed -n "/^---$/,/^---$/p" "$file" | sed "1d;\$d"))

      # Extract keywords from original prompts
      ORIGINAL=$(sed -n "/## Original Prompts/,/^##/p" "$file" | grep "^- " | sed "s/^- //" | tr "[:upper:]" "[:lower:]" | tr -cs "[:alnum:]" "\n" | grep -E "^.{3,}$" | sort -u)

      # Combine metadata keywords + extracted keywords
      ALL_KEYWORDS=$(echo "$META" | jq -r ".keywords[]"; echo "$ORIGINAL") | sort -u

      # Calculate Jaccard similarity
      INTERSECTION=$(comm -12 <(echo "$QUERY_KEYWORDS") <(echo "$ALL_KEYWORDS") | wc -l)
      UNION=$(echo "$QUERY_KEYWORDS $ALL_KEYWORDS" | tr " " "\n" | sort -u | wc -l)
      SIMILARITY=$(echo "scale=4; $INTERSECTION / $UNION" | bc)

      # Filter by threshold
      if (( $(echo "$SIMILARITY > 0.2" | bc -l) )); then
        # Calculate usage score
        USAGE_COUNT=$(echo "$META" | jq -r ".usage_count")
        USAGE_SCORE=$(echo "scale=4; l($USAGE_COUNT + 1) / 5.0" | bc -l)

        # Combined score
        COMBINED=$(echo "scale=4; ($SIMILARITY * 0.7) + ($USAGE_SCORE * 0.3)" | bc)

        # Output JSONL
        echo "$META" | jq -c ". + {file: \"$file\", similarity: $SIMILARITY, combined_score: $COMBINED}"
      fi
    done | sort -t: -k2 -rn  # Sort by combined_score descending
  '
}

## Scoring Details

scoring :: Similarity_And_Usage → Combined_Score
scoring(S, U) = {
  # Similarity (Jaccard coefficient): 0.0-1.0
  # - Measures keyword overlap between query and saved prompt
  # - Weight: 70% (primary factor)

  # Usage score (logarithmic): 0.0-1.0 (normalized)
  # - log(1) = 0.0 (never used)
  # - log(100) ≈ 4.6, divided by 5 ≈ 0.92 (heavily used)
  # - Weight: 30% (secondary factor, prevents over-weighting popular prompts)

  combined: (S * 0.7) + (U * 0.3),

  # Result range: 0.0-1.0
  # - 0.0: No similarity, never used
  # - 1.0: Perfect match, heavily used

  return: combined
}

## Threshold Tuning

threshold_guide :: Similarity → Match_Quality
threshold_guide() = {
  thresholds: {
    0.8-1.0: "Excellent match - nearly identical prompts",
    0.6-0.8: "Good match - highly relevant prompts",
    0.4-0.6: "Moderate match - related prompts",
    0.2-0.4: "Weak match - loosely related prompts",
    0.0-0.2: "Poor match - filtered out (not shown)"
  },

  # Current threshold: 0.2 (20% keyword overlap)
  # Rationale: Inclusive enough to find related prompts, exclusive enough to avoid noise
  current: 0.2,

  tuning: {
    increase_to_0.3: "If too many irrelevant results",
    decrease_to_0.1: "If too few results (sparse library)",
    dynamic: "Consider adjusting based on library size"
  }
}

## Constraints

constraints:
- efficient: O(n * m) where n=library_size, m=avg_keywords (acceptable for <1000 prompts)
- threshold: similarity > 0.2 (tunable)
- ranking: combined_score = 0.7*similarity + 0.3*usage_score
- top_n: display max 5 matches
- atomic_updates: usage_count updates use temp files
- graceful_degradation: handle empty library, no matches, invalid selection
