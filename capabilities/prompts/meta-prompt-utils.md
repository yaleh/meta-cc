---
name: meta-prompt-utils
description: Internal utility functions for prompt library management
category: internal
---

λ(operation) → utility_result | ∀function ∈ shared_utilities:

## Keyword Extraction

extract_keywords :: String → [String]
extract_keywords(S) = {
  # Tokenize on whitespace and punctuation
  raw_words: split(S, /[\s\p{P}]+/),

  # Filter stopwords and short words
  stopwords: ["the", "a", "an", "and", "or", "to", "in", "of", "for", "on", "with", "is", "are", "was", "were", "be", "been", "have", "has", "had", "do", "does", "did", "will", "would", "should", "could", "can", "may", "might", "this", "that", "these", "those", "i", "you", "he", "she", "it", "we", "they"],

  filtered: ∀w ∈ raw_words: {
    valid: length(w) > 2 ∧ lowercase(w) ∉ stopwords
  },

  # Normalize to lowercase
  normalized: map(filtered, w → lowercase(w)),

  # Remove duplicates
  unique_keywords: unique(normalized),

  return: unique_keywords
}

## Similarity Calculation

jaccard_similarity :: ([String], [String]) → Float
jaccard_similarity(A, B) = {
  # Handle empty sets
  if (|A| == 0 ∨ |B| == 0):
    return: 0.0,

  # Convert to sets for comparison
  set_a: unique(A),
  set_b: unique(B),

  # Calculate intersection (common elements)
  intersection: set_a ∩ set_b,

  # Calculate union (all unique elements)
  union: set_a ∪ set_b,

  # Jaccard coefficient: |A ∩ B| / |A ∪ B|
  score: float(|intersection|) / float(|union|),

  # Range: 0.0 (no overlap) to 1.0 (identical sets)
  return: score
}

## Frontmatter Parsing

parse_frontmatter :: FilePath → Metadata
parse_frontmatter(F) = {
  # Read file content
  content: read_file(F),

  # Extract YAML between first two "---" delimiters
  # Pattern: ^---\n(.*?)\n---\n (dotall mode)
  yaml_block: extract_between_delimiters(content, "---", "---"),

  # Parse YAML (using yq or native parser)
  metadata: parse_yaml(yaml_block),

  # Validate required fields
  required: ["id", "title", "category", "keywords", "created", "updated", "usage_count"],
  validate: ∀field ∈ required: field ∈ metadata,

  return: metadata
}

## Usage Count Update

update_usage_count :: FilePath → Result
update_usage_count(F) = {
  # Parse current frontmatter
  current_meta: parse_frontmatter(F),

  # Increment usage count
  new_count: current_meta.usage_count + 1,

  # Update timestamp
  new_timestamp: now_iso8601(),

  # Update metadata fields
  updated_meta: {
    ...current_meta,
    usage_count: new_count,
    updated: new_timestamp
  },

  # Read original content (after frontmatter)
  full_content: read_file(F),
  body_content: extract_after_frontmatter(full_content),

  # Reconstruct file: frontmatter + body
  new_frontmatter: format_yaml(updated_meta),
  new_file_content: "---\n" + new_frontmatter + "\n---\n\n" + body_content,

  # Write atomically (temp file + rename)
  temp_file: F + ".tmp",
  write: atomic_write(temp_file, new_file_content),
  rename: move(temp_file, F),

  return: {
    success: true,
    new_count: new_count,
    new_timestamp: new_timestamp
  }
}

## Bash Implementation Hints

bash_extract_keywords :: String → String
bash_extract_keywords(S) = {
  # Using tr, grep, and awk
  pipeline: "
    echo '$S' |
    tr '[:upper:]' '[:lower:]' |           # Lowercase
    tr -cs '[:alnum:]' '\n' |               # Split on non-alphanumeric
    grep -E '^.{3,}$' |                     # Length > 2
    grep -vwFf <(echo -e '${stopwords}') | # Filter stopwords
    sort -u                                 # Unique
  "
}

bash_jaccard_similarity :: (FilePath, FilePath) → Float
bash_jaccard_similarity(F1, F2) = {
  # Using comm to calculate set operations
  pipeline: "
    INTERSECTION=$(comm -12 <(sort '$F1') <(sort '$F2') | wc -l)
    UNION=$(cat '$F1' '$F2' | sort -u | wc -l)
    echo \"scale=4; $INTERSECTION / $UNION\" | bc
  "
}

bash_parse_frontmatter :: FilePath → Metadata
bash_parse_frontmatter(F) = {
  # Using yq (recommended) or awk
  using_yq: "yq eval '.' <(sed -n '/^---$/,/^---$/p' '$F' | sed '1d;$d')",

  using_awk: "awk '/^---$/{f=!f;next}f' '$F' | head -n -1"
}

bash_update_usage_count :: FilePath → Result
bash_update_usage_count(F) = {
  # Read current count
  current_count: "yq eval '.usage_count' '$F'",

  # Increment
  new_count: $((current_count + 1)),

  # Update in-place
  update_cmd: "yq eval -i '.usage_count = $new_count | .updated = \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"' '$F'"
}

## Constraints

constraints:
- pure_functions: no side effects except update_usage_count()
- efficient: O(n) for most operations
- robust: handle edge cases (empty files, malformed YAML)
- language_agnostic: support multilingual keywords
- atomic_updates: use temp files for writes to prevent corruption
