# Phase 28 Implementation Guide

**For Developers**: Quick reference for implementing the Prompt Optimization Learning System

## Pre-Implementation Checklist

- [ ] Read [PHASE-28-IMPLEMENTATION-PLAN.md](./PHASE-28-IMPLEMENTATION-PLAN.md) completely
- [ ] Review [PROMPT-FILE-FORMAT.md](./PROMPT-FILE-FORMAT.md) for file format specification
- [ ] Review existing `capabilities/commands/meta-prompt.md` to understand current behavior
- [ ] Test current `/meta Refine prompt` functionality to establish baseline
- [ ] Set up test project for validation

## Stage 1: Infrastructure and Save (5-6 hours)

### Step 1: Create Internal Capability Directory (5 minutes)

```bash
cd /home/yale/work/meta-cc
mkdir -p capabilities/prompts
ls -la capabilities/
# Should see both commands/ and prompts/ directories
```

**Validation**: Directory exists and is writable

### Step 2: Implement meta-prompt-save.md (2-3 hours)

**File**: `capabilities/prompts/meta-prompt-save.md`

**Implementation Checklist**:
- [ ] YAML frontmatter with capability metadata
- [ ] λ-calculus style formal specification
- [ ] `initialize()` function - auto-create directories
- [ ] `generate_id()` function - sequential ID assignment
- [ ] `create_frontmatter()` function - YAML generation
- [ ] `format_content()` function - Markdown formatting
- [ ] `save()` function - file writing
- [ ] User interaction prompts (category, keywords, description)
- [ ] Error handling (permission issues, disk space)
- [ ] Confirmation message

**Key Functions to Implement**:

```markdown
initialize :: Project_Root → Storage_Path
# Check if .meta-cc/prompts/library/ exists
# Create if missing with mkdir -p
# Create .gitignore with "# Local prompt library\n"

generate_id :: (Storage_Path, Category) → Unique_ID
# List existing files: ls {storage}/*.md
# Filter by category: grep "^{category}-"
# Extract max ID: grep -oE '[0-9]{3}\.md$' | sort -rn | head -1
# Increment: next_id = max_id + 1
# Format: sprintf("%03d", next_id)

create_frontmatter :: Params → YAML_String
# Use template from PROMPT-FILE-FORMAT.md
# Fill in all required fields
# Format timestamps as ISO8601
# Initialize usage_count to 0
# Set status to "active"

format_content :: (Original, Optimized) → Markdown_String
# Template:
# ## Original Prompts
# - {original}
#
# ## Optimized Prompt
# {optimized}

save :: (Frontmatter, Content, Filename) → Result
# Combine: "---\n" + frontmatter + "\n---\n\n" + content
# Write to: {storage}/{filename}
# Return: success message with file path
```

**Testing**:
```bash
# Manual test - invoke capability directly
/meta prompts/meta-prompt-save

# Input test data:
# Original: "发布新版本"
# Optimized: "使用 ./scripts/release/release.sh v1.0.0 执行版本发布"
# Category: release
# Keywords: 发布,release,版本
# Description: simple-release

# Validate output
ls -la .meta-cc/prompts/library/
cat .meta-cc/prompts/library/release-simple-release-001.md

# Validate YAML
yq -f extract '.' .meta-cc/prompts/library/release-simple-release-001.md

# Validate required fields
yq -f extract '.id' .meta-cc/prompts/library/release-simple-release-001.md
# Should match: release-simple-release-001
```

**Common Issues**:
- **Permission denied**: Check directory permissions
- **Invalid YAML**: Validate with yq before writing
- **ID collision**: Ensure proper sequential ID generation
- **Encoding issues**: Use UTF-8 for multilingual keywords

### Step 3: Extend meta-prompt.md (1-2 hours)

**File**: `capabilities/commands/meta-prompt.md`

**Changes**:
1. Add post-optimization save workflow section (end of file)
2. Call `get_capability("prompts/meta-prompt-save")`
3. Pass parameters: original prompt, optimized prompt, timestamp
4. Make save optional (ask user "Save? y/N")
5. Default to skip (non-intrusive)

**Implementation**:

Add to end of file (after existing content):

```markdown

---

## Post-Optimization: Save for Reuse

After generating optimized prompts, offer to save for future reuse:

save_workflow :: Optimized_Result → Optional[Saved_File]
save_workflow(R) = {
  display: alternatives(R),  # Show optimization results first

  ask: "Would you like to save this optimized prompt to your library? (y/N)",
  default: "N",  # Non-intrusive - default is skip

  if (user_confirms):
    call: get_capability("prompts/meta-prompt-save"),
    pass: {
      prompt_original: R.original,
      prompt_optimized: R.recommendation,  # Best alternative
      timestamp: now()
    },
    result: display_save_confirmation(),

  else:
    skip: "Not saved. You can refine this prompt again anytime with '/meta Refine prompt: ...'"
}

integration :: Complete_Workflow
integration(P) = {
  step_1: analyze(P),           # Existing: analyze history
  step_2: detect(P),            # Existing: detect gaps
  step_3: generate(P),          # Existing: generate alternatives
  step_4: output(alternatives), # Existing: display results
  step_5: save_workflow(result) # NEW: optional save
}

constraints:
- save is optional, not mandatory
- default answer is "No" (skip)
- explain benefits of saving
- confirm save location
```

**Testing**:
```bash
# Test complete workflow
/meta Refine prompt: 发布新版本

# Verify optimization works (existing functionality)
# Then verify save prompt appears
# Test both paths:
# 1. Answer "y" → should trigger save
# 2. Answer "n" or Enter → should skip
```

### Step 4: Update CLAUDE.md (30 minutes)

**File**: `/home/yale/work/meta-cc/CLAUDE.md`

**Location**: Add to FAQ section (around line 50-60)

**Content**:
```markdown
**Q: How does the prompt learning system work?**
A: After using `/meta Refine prompt: XXX`, you can save the optimized version to `.meta-cc/prompts/library/`. The system will recommend these saved prompts when you try similar prompts in the future, making you more efficient over time.

**Q: Where are saved prompts stored?**
A: Project-local storage in `.meta-cc/prompts/library/` (not tracked by git by default). You can commit selectively if you want to share with your team. The directory is auto-created on first save.

**Q: Can I search my saved prompts?**
A: Yes, in Stage 2 we'll add `/meta prompts/meta-prompt-search` for similarity search. In Stage 3, use `/meta prompts/meta-prompt-list` to browse by category or usage. Files are plain text, so you can also use `grep`, `ack`, or `rg`.

**Q: What if I don't want to save prompts?**
A: Saving is completely optional. Just press Enter or answer "n" when prompted. The save option won't appear again until you optimize another prompt.
```

### Step 5: Stage 1 Validation (30 minutes)

**Validation Checklist**:

- [ ] **Directory Creation**:
  ```bash
  # Start fresh
  rm -rf .meta-cc/
  # Trigger first save
  /meta Refine prompt: test prompt
  # (answer "y" to save)
  # Verify
  [ -d .meta-cc/prompts/library/ ] && echo "✅ Directory created"
  [ -f .meta-cc/prompts/library/.gitignore ] && echo "✅ Gitignore created"
  ```

- [ ] **File Format**:
  ```bash
  # Save a test prompt
  FILE=$(ls .meta-cc/prompts/library/*.md | head -1)
  # Validate YAML
  yq -f extract '.' "$FILE" && echo "✅ Valid YAML"
  # Check required fields
  yq -f extract '.id, .title, .category, .keywords, .created, .updated, .usage_count, .effectiveness, .status' "$FILE"
  ```

- [ ] **File Naming**:
  ```bash
  # Should match pattern: {category}-{description}-{id}.md
  FILE=$(basename "$(ls .meta-cc/prompts/library/*.md | head -1)")
  [[ "$FILE" =~ ^[a-z]+-[a-z-]+-[0-9]{3}\.md$ ]] && echo "✅ Valid filename"
  ```

- [ ] **Content Sections**:
  ```bash
  FILE=$(ls .meta-cc/prompts/library/*.md | head -1)
  grep -q "## Original Prompts" "$FILE" && echo "✅ Original prompts section"
  grep -q "## Optimized Prompt" "$FILE" && echo "✅ Optimized prompt section"
  ```

- [ ] **Sequential IDs**:
  ```bash
  # Save multiple prompts in same category
  /meta Refine prompt: test 1  # (save as "test" category)
  /meta Refine prompt: test 2  # (save as "test" category)
  /meta Refine prompt: test 3  # (save as "test" category)
  # Verify IDs are sequential
  ls .meta-cc/prompts/library/test-*.md
  # Should see: test-*-001.md, test-*-002.md, test-*-003.md
  ```

- [ ] **Skip Option**:
  ```bash
  /meta Refine prompt: skip test
  # Press Enter or answer "n"
  # Verify no new file created
  ```

**If all checks pass**: ✅ Stage 1 complete! Proceed to Stage 2.

**If checks fail**: Debug issues before proceeding.

---

## Stage 2: Search and Reuse (5-6 hours)

### Step 1: Implement meta-prompt-utils.md (1 hour)

**File**: `capabilities/prompts/meta-prompt-utils.md`

**Implementation Checklist**:
- [ ] `extract_keywords()` - tokenize and filter stopwords
- [ ] `jaccard_similarity()` - intersection / union
- [ ] `parse_frontmatter()` - extract YAML between --- delimiters
- [ ] `update_usage_count()` - increment count and update timestamp

**Key Functions**:

```markdown
extract_keywords :: String → [String]
# Split on whitespace: tokenize(S, /\s+/)
# Filter: length > 2 AND not in stopwords
# Normalize: lowercase
# Stopwords: ["the", "a", "an", "and", "or", "to", "in", "of", ...]

jaccard_similarity :: ([String], [String]) → Float
# Intersection: A ∩ B (common elements)
# Union: A ∪ B (all unique elements)
# Score: |intersection| / |union|
# Range: 0.0 (no overlap) to 1.0 (identical)

parse_frontmatter :: FilePath → Metadata
# Read file content
# Extract between first "---" and second "---"
# Parse YAML (use yq or native parser)
# Return metadata object

update_usage_count :: FilePath → Result
# Parse current frontmatter
# Increment usage_count by 1
# Update updated timestamp to now()
# Write back to file (preserve content)
```

**Testing**:
```bash
# Test extract_keywords
echo "Deploy new version to production with CI/CD monitoring" | \
  # Should extract: [deploy, new, version, production, ci/cd, monitoring]

# Test jaccard_similarity
# A = [release, version, deploy]
# B = [release, deploy, production]
# Intersection: [release, deploy] = 2
# Union: [release, version, deploy, production] = 4
# Score: 2/4 = 0.5

# Test parse_frontmatter
FILE=$(ls .meta-cc/prompts/library/*.md | head -1)
yq -f extract '.' "$FILE"  # Should output valid YAML

# Test update_usage_count
BEFORE=$(yq -f extract '.usage_count' "$FILE")
# (trigger update via reuse in Step 2.2)
AFTER=$(yq -f extract '.usage_count' "$FILE")
[ $AFTER -eq $((BEFORE + 1)) ] && echo "✅ Usage count updated"
```

### Step 2: Implement meta-prompt-search.md (2-3 hours)

**File**: `capabilities/prompts/meta-prompt-search.md`

**Implementation Checklist**:
- [ ] `search()` - find similar prompts with similarity scoring
- [ ] `format_results()` - display ranked matches
- [ ] `reuse()` - handle user selection
- [ ] Threshold tuning (similarity > 0.2)
- [ ] Combined scoring (0.7 * similarity + 0.3 * usage_score)

**Key Functions**:

```markdown
search :: Query_Prompt → [Ranked_Matches]
search(Q) = {
  library: list_files(".meta-cc/prompts/library/*.md"),

  if (empty(library)):
    return: [],

  query_keywords: extract_keywords(Q),

  candidates: ∀file ∈ library: {
    meta: parse_frontmatter(file),
    original_keywords: meta.keywords + extract_keywords(meta.original),
    similarity: jaccard_similarity(query_keywords, original_keywords),
    usage_score: log(meta.usage_count + 1) / log(100),  # Normalize to 0-1
    combined_score: 0.7 * similarity + 0.3 * usage_score
  },

  matches: filter(candidates, similarity > 0.2),  # Threshold
  ranked: sort_desc(matches, combined_score),
  top_n: take(ranked, 5)
}

format_results :: [Matches] → Display_String
format_results(M) = {
  if (|M| == 0):
    "No similar prompts found in your library. Generating fresh recommendations...",

  else:
    header: sprintf("Found %d similar prompt(s) in your library:\n\n", |M|),

    list: ∀m ∈ M (indexed from 1): {
      line: sprintf("%d. %s [%s] (%.0f%% match, %d uses)\n   %s\n",
        index,
        m.meta.title,
        m.meta.category,
        m.similarity * 100,
        m.meta.usage_count,
        truncate(m.optimized, 100)
      )
    },

    footer: "\nSelect a number (1-5) to reuse, or press Enter to generate new optimization: "
}

reuse :: (User_Input, Matches) → Result
reuse(I, M) = {
  if (I == "" || I == "0" || I == "n"):
    return: {action: "generate_new", prompt: null},

  else if (1 <= int(I) <= |M|):
    selected: M[int(I) - 1],
    update_usage_count(selected.file),
    return: {
      action: "reuse",
      prompt: selected.optimized,
      file: selected.file
    },

  else:
    error: "Invalid selection. Generating new optimization...",
    return: {action: "generate_new", prompt: null}
}
```

**Testing**:
```bash
# Setup: Create test library with known prompts
mkdir -p .meta-cc/prompts/library
# (Save 2-3 prompts with overlapping keywords)

# Test exact match
/meta Refine prompt: [exact same as saved prompt]
# Should find 100% match

# Test partial match
/meta Refine prompt: [similar but different wording]
# Should find >50% match

# Test no match
/meta Refine prompt: [completely unrelated topic]
# Should say "No similar prompts found"

# Test selection
# When matches shown, enter "1" → should reuse first match
# Enter "" → should generate new
# Enter "99" → should handle gracefully
```

### Step 3: Integrate into meta-prompt.md (1-2 hours)

**File**: `capabilities/commands/meta-prompt.md`

**Changes**:
1. Add pre-optimization history search (beginning of workflow)
2. Call `get_capability("prompts/meta-prompt-search")`
3. Display matches if found
4. Allow user to select or skip
5. If reuse selected, return immediately (skip normal optimization)
6. If generate new, continue with existing workflow

**Implementation**:

Add at beginning of workflow (before `analyze()` step):

```markdown
## Pre-Optimization: History Search

Before generating new optimizations, check if similar prompts exist:

search_history :: Raw_Prompt → Search_Result
search_history(P) = {
  call: get_capability("prompts/meta-prompt-search"),
  pass: {query_prompt: P},

  matches: search(P),

  if (|matches| > 0):
    display: format_results(matches),
    ask: user_selection(),

    result: reuse(user_selection, matches),

    if (result.action == "reuse"):
      output: result.prompt,
      confirm: sprintf("Reused prompt from: %s", result.file),
      return: SKIP_OPTIMIZATION,  # Exit workflow early

    else:
      continue: NORMAL_OPTIMIZATION

  else:
    continue: NORMAL_OPTIMIZATION
}

workflow_integration :: Complete_Flow
workflow_integration(P) = {
  step_0: search_history(P),         # NEW: Check history first

  if (step_0 == SKIP_OPTIMIZATION):
    return: early_exit,

  # Existing workflow continues:
  step_1: analyze(history),
  step_2: detect(gaps),
  step_3: generate(alternatives),
  step_4: output(alternatives),
  step_5: save_workflow(result)
}
```

**Testing**:
```bash
# Test full workflow with search
/meta Refine prompt: 发布新版本

# Expected flow:
# 1. Search history → finds match (if exists)
# 2. Display matches with similarity scores
# 3. User selects:
#    a. Select 1 → reuse, exit workflow
#    b. Press Enter → generate new, continue workflow

# Test both paths
```

### Step 4: Stage 2 Validation (1 hour)

**Validation Checklist**:

- [ ] **Search Functionality**:
  ```bash
  # Create test library
  /meta Refine prompt: release new version  # (save as release-001)
  /meta Refine prompt: debug error logs     # (save as debug-001)

  # Test exact match
  /meta Refine prompt: release new version
  # Should find 100% match for release-001

  # Test partial match
  /meta Refine prompt: deploy new release
  # Should find >50% match for release-001

  # Test no match
  /meta Refine prompt: write unit tests
  # Should find no matches (assuming no test prompts saved)
  ```

- [ ] **Similarity Scoring**:
  ```bash
  # Verify similarity calculation
  # Query: "release version monitoring"
  # Keywords: [release, version, monitoring]
  # Saved prompt keywords: [release, version, deploy, ci]
  # Intersection: [release, version] = 2
  # Union: [release, version, monitoring, deploy, ci] = 5
  # Expected similarity: 2/5 = 0.4 = 40%
  ```

- [ ] **Usage Tracking**:
  ```bash
  FILE=".meta-cc/prompts/library/release-*-001.md"
  BEFORE=$(yq -f extract '.usage_count' "$FILE")
  # Reuse the prompt via search
  /meta Refine prompt: [trigger match and select]
  AFTER=$(yq -f extract '.usage_count' "$FILE")
  [ $AFTER -eq $((BEFORE + 1)) ] && echo "✅ Usage count incremented"

  # Verify updated timestamp changed
  yq -f extract '.updated' "$FILE"  # Should be recent
  ```

- [ ] **User Selection**:
  ```bash
  # Test all selection options:
  # 1. Enter "1" → reuse first match
  # 2. Enter "" or "n" → generate new
  # 3. Enter "99" → invalid, generate new
  # 4. Enter "abc" → invalid, generate new
  ```

- [ ] **Edge Cases**:
  ```bash
  # Empty library
  rm -rf .meta-cc/
  /meta Refine prompt: test
  # Should skip search, go to normal optimization

  # Large library (performance)
  # Create 20+ prompts, then search
  # Should complete in <3 seconds
  ```

**If all checks pass**: ✅ Stage 2 complete! Proceed to Stage 3.

---

## Stage 3: Management and Listing (3-4 hours)

### Step 1: Implement meta-prompt-list.md (2-3 hours)

**File**: `capabilities/prompts/meta-prompt-list.md`

**Implementation Checklist**:
- [ ] `list()` - list all prompts with optional filtering
- [ ] `format_table()` - display as formatted table
- [ ] `summary_stats()` - calculate statistics
- [ ] `detail_view()` - show full prompt content
- [ ] Support filter by category
- [ ] Support sort by usage, date, alpha

**Key Functions**:

```markdown
list :: (Category_Filter, Sort_Order) → [Prompts]
list(C, S) = {
  library: list_files(".meta-cc/prompts/library/*.md"),

  if (empty(library)):
    return: error("No prompts saved yet. Use '/meta Refine prompt' to create your first one!"),

  prompts: ∀file ∈ library: {
    meta: parse_frontmatter(file),
    filter_match: (C == null) OR (meta.category == C)
  },

  filtered: filter(prompts, filter_match),

  sorted: {
    case S of:
      "usage": sort_desc(filtered, usage_count),
      "date": sort_desc(filtered, updated),
      "alpha": sort_asc(filtered, title),
      default: sort_desc(filtered, updated)
  }
}

format_table :: [Prompts] → Display_String
format_table(P) = {
  header: "| # | Title | Category | Usage | Last Updated |\n",
  separator: "|---|-------|----------|-------|--------------|",

  rows: ∀p ∈ P (indexed from 1): {
    sprintf("| %d | %-30s | %-8s | %5d | %-12s |",
      index,
      truncate(p.title, 30),
      p.category,
      p.usage_count,
      relative_time(p.updated)  # "2 days ago"
    )
  },

  footer: sprintf("\nTotal: %d prompts across %d categories",
    |P|,
    |unique(P.category)|
  )
}

summary_stats :: [Prompts] → Statistics
summary_stats(P) = {
  total: |P|,
  categories: |unique(P.category)|,
  total_uses: sum(P.usage_count),
  most_used: argmax(P, usage_count),
  recent: take(sort_desc(P, updated), 5)
}

detail_view :: Prompt_ID → Display
detail_view(ID) = {
  file: find_by_id(ID, ".meta-cc/prompts/library/*.md"),

  if (not_found(file)):
    return: error(sprintf("Prompt not found: %s", ID)),

  meta: parse_frontmatter(file),
  content: read_content(file),

  display: {
    section_1: "## Metadata\n",
    yaml: format_yaml(meta),
    section_2: "\n## Content\n",
    markdown: content
  }
}
```

**Testing**:
```bash
# Test list all
/meta prompts/meta-prompt-list
# Should show table with all prompts

# Test filter by category
/meta prompts/meta-prompt-list --category=release
# Should show only release prompts

# Test sort by usage
/meta prompts/meta-prompt-list --sort=usage
# Should show most-used first

# Test sort by date
/meta prompts/meta-prompt-list --sort=date
# Should show most-recent first

# Test empty library
rm -rf .meta-cc/
/meta prompts/meta-prompt-list
# Should show helpful error message

# Test detail view
/meta prompts/meta-prompt-list --detail=release-simple-001
# Should show full content
```

### Step 2: Update CLAUDE.md (30 minutes)

**File**: `/home/yale/work/meta-cc/CLAUDE.md`

**Location**: Add to FAQ section

**Content**:
```markdown
**Q: How do I browse my saved prompts?**
A: Three ways:
1. **Capability**: `/meta prompts/meta-prompt-list` for formatted table view
   - Filter by category: Add `--category=release`
   - Sort by usage: Add `--sort=usage`
   - Sort by date: Add `--sort=date`
2. **Shell commands**: `ls -lt .meta-cc/prompts/library/`
3. **Search tools**: `rg "keyword" .meta-cc/prompts/library/`

**Q: Can I delete or edit saved prompts?**
A: Yes, they're just markdown files:
- **Delete**: `rm .meta-cc/prompts/library/release-simple-001.md`
- **Edit**: `vim .meta-cc/prompts/library/release-simple-001.md`
- **Archive**: Edit YAML frontmatter, set `status: archived`

**Q: Can I share prompts with my team?**
A: Yes, commit to git:
```bash
git add .meta-cc/prompts/library/release-*.md
git commit -m "docs: share release process prompts"
git push
```

**Q: How do I back up my prompt library?**
A: Simple directory copy:
```bash
# Backup
cp -r .meta-cc/prompts ~/backups/project-prompts-$(date +%Y%m%d)

# Restore
cp -r ~/backups/project-prompts-20251027/.meta-cc/prompts .meta-cc/
```

**Q: Can I use prompts across multiple projects?**
A: Currently project-local. Phase 28.5 will add global library in `~/.meta-cc/` for cross-project sharing.
```

### Step 3: Stage 3 Validation (1 hour)

**Validation Checklist**:

- [ ] **List All Prompts**:
  ```bash
  # Create diverse library
  # (save 5+ prompts in different categories)
  /meta prompts/meta-prompt-list
  # Should show table with all prompts
  # Verify columns: #, Title, Category, Usage, Last Updated
  ```

- [ ] **Filter by Category**:
  ```bash
  /meta prompts/meta-prompt-list --category=release
  # Should show only release category
  # Verify no other categories in output
  ```

- [ ] **Sort by Usage**:
  ```bash
  # Reuse some prompts to vary usage counts
  /meta prompts/meta-prompt-list --sort=usage
  # Verify order: highest usage_count first
  ```

- [ ] **Sort by Date**:
  ```bash
  /meta prompts/meta-prompt-list --sort=date
  # Verify order: most recent updated timestamp first
  ```

- [ ] **Summary Statistics**:
  ```bash
  /meta prompts/meta-prompt-list
  # Check footer shows:
  # - Total prompt count (matches actual)
  # - Category count (matches unique categories)
  ```

- [ ] **Detail View**:
  ```bash
  # Get ID from list
  ID=$(ls .meta-cc/prompts/library/*.md | head -1 | xargs basename .md)
  /meta prompts/meta-prompt-list --detail=$ID
  # Should show:
  # - Complete YAML frontmatter
  # - Full markdown content
  # - Both Original and Optimized sections
  ```

- [ ] **Empty Library**:
  ```bash
  rm -rf .meta-cc/
  /meta prompts/meta-prompt-list
  # Should show: "No prompts saved yet..."
  ```

**If all checks pass**: ✅ Stage 3 complete! Phase 28 is done!

---

## Final Integration Testing

### End-to-End Workflow Test

**Scenario 1: First-time User**
```bash
# Clean state
rm -rf .meta-cc/

# 1. Optimize and save first prompt
/meta Refine prompt: 发布新版本
# → Should generate recommendations
# → Ask "Save? (y/N)"
# → Answer "y"
# → Prompt for category, keywords, description
# → Save to .meta-cc/prompts/library/release-*-001.md
# → Verify file created

# 2. Optimize similar prompt (triggers search)
/meta Refine prompt: release new version
# → Should find saved prompt with high similarity
# → Display match with "90%+ match"
# → Allow selection or skip

# 3. Reuse saved prompt
# → Select "1" to reuse
# → Verify usage_count incremented
# → Verify updated timestamp changed

# 4. Browse library
/meta prompts/meta-prompt-list
# → Should show 1 prompt
# → Usage count = 1 (from reuse)

# ✅ Complete workflow validated
```

**Scenario 2: Power User**
```bash
# Setup: Create diverse library (10+ prompts)
# (save prompts in multiple categories: release, debug, refactor, test, docs)

# 1. Search and reuse
/meta Refine prompt: troubleshoot error
# → Should find debug prompts
# → Reuse one

# 2. Generate new optimization (no match)
/meta Refine prompt: implement OAuth2 authentication
# → Should find no matches (assuming none exist)
# → Generate fresh recommendations
# → Save new prompt

# 3. Browse by category
/meta prompts/meta-prompt-list --category=debug
# → Should show only debug prompts

# 4. Find most-used prompts
/meta prompts/meta-prompt-list --sort=usage
# → Should show prompts ordered by popularity

# 5. Shell integration
ls -lt .meta-cc/prompts/library/
rg "error" .meta-cc/prompts/library/
# → Verify CLI tools work

# ✅ Power user workflow validated
```

### Performance Testing

```bash
# Test with large library
# Generate 50+ prompts
for i in {1..50}; do
  /meta Refine prompt: test prompt $i
  # (save each)
done

# Test search performance
time /meta Refine prompt: test prompt query
# → Should complete in <3 seconds

# Test list performance
time /meta prompts/meta-prompt-list
# → Should complete in <2 seconds

# ✅ Performance acceptable
```

### Cross-Project Testing

```bash
# Project A
cd /path/to/projectA
/meta Refine prompt: release version
# (save)

# Project B
cd /path/to/projectB
/meta Refine prompt: release version
# → Should NOT find Project A's prompts
# → Confirms project-local isolation

# Copy prompts between projects
cp -r /path/to/projectA/.meta-cc/prompts /path/to/projectB/.meta-cc/
cd /path/to/projectB
/meta Refine prompt: release version
# → Should now find copied prompts
# ✅ Manual sharing works
```

---

## Documentation Finalization

### Create User Guide

**File**: `docs/guides/prompt-learning-system.md`

**See**: [PHASE-28-IMPLEMENTATION-PLAN.md](./PHASE-28-IMPLEMENTATION-PLAN.md#documentation-updates) for complete content template

**Sections**:
1. Overview
2. Quick Start
3. Architecture
4. Usage Patterns
5. Best Practices
6. Advanced Usage
7. CLI Integration
8. Troubleshooting
9. Future Enhancements

### Update References

**Files to update**:
- [ ] `docs/reference/unified-meta-command.md` - Add prompt capabilities
- [ ] `README.md` - Mention in features, link to guide
- [ ] `docs/reference/features.md` - Add prompt learning system section

---

## Common Issues and Solutions

### Issue 1: Directory Permission Denied

**Symptom**: Cannot create `.meta-cc/prompts/library/`

**Solution**:
```bash
# Check permissions
ls -la .meta-cc/
# Fix permissions
chmod 755 .meta-cc/
mkdir -p .meta-cc/prompts/library
chmod 755 .meta-cc/prompts/library
```

### Issue 2: Invalid YAML Frontmatter

**Symptom**: `yq` fails to parse file

**Solution**:
```bash
# Validate YAML
yq -f extract '.' file.md

# Common issues:
# - Missing quotes around special characters
# - Incorrect indentation
# - Invalid timestamp format

# Fix: Edit file, ensure YAML is valid
vim file.md
```

### Issue 3: Search Returns No Results

**Symptom**: Search finds no matches despite having similar prompts

**Solution**:
```bash
# Check keyword overlap
cat .meta-cc/prompts/library/*.md | grep "keywords:"
# Ensure query keywords overlap with saved keywords

# Lower similarity threshold in meta-prompt-search.md
# Change: similarity > 0.2
# To: similarity > 0.1
```

### Issue 4: Usage Count Not Updating

**Symptom**: Reusing prompt doesn't increment usage_count

**Solution**:
```bash
# Check file write permissions
ls -la .meta-cc/prompts/library/
# Verify update_usage_count() is called
# Add debug logging to meta-prompt-search.md:
# echo "Updating usage count for: $file"

# Manually verify YAML is valid before/after
yq -f extract '.usage_count' file.md
```

---

## Completion Checklist

### Stage 1 Complete
- [ ] `capabilities/prompts/meta-prompt-save.md` created
- [ ] `capabilities/commands/meta-prompt.md` extended with save workflow
- [ ] `CLAUDE.md` updated with FAQ entries
- [ ] Directory auto-creation works
- [ ] File format is valid (YAML + Markdown)
- [ ] Sequential ID generation works
- [ ] Save is optional (can skip)

### Stage 2 Complete
- [ ] `capabilities/prompts/meta-prompt-utils.md` created
- [ ] `capabilities/prompts/meta-prompt-search.md` created
- [ ] `capabilities/commands/meta-prompt.md` extended with search workflow
- [ ] Similarity matching works (Jaccard)
- [ ] Search finds relevant prompts
- [ ] Usage tracking works (increment count)
- [ ] User can select or skip matches

### Stage 3 Complete
- [ ] `capabilities/prompts/meta-prompt-list.md` created
- [ ] `CLAUDE.md` updated with browsing/management FAQ
- [ ] List all prompts works
- [ ] Filter by category works
- [ ] Sort by usage/date/alpha works
- [ ] Summary statistics are accurate
- [ ] Detail view shows full content

### Documentation Complete
- [ ] `docs/guides/prompt-learning-system.md` created
- [ ] `docs/reference/unified-meta-command.md` updated
- [ ] `README.md` updated with feature mention
- [ ] All examples tested and working

### Validation Complete
- [ ] End-to-end workflow tested (Scenario 1 & 2)
- [ ] Performance tested (50+ prompts)
- [ ] Cross-project isolation verified
- [ ] Edge cases handled gracefully
- [ ] CLI integration verified

---

## Post-Implementation

### Announce and Collect Feedback

1. **Update CHANGELOG**:
   ```markdown
   ## [Unreleased]

   ### Added
   - Phase 28: Prompt Optimization Learning System
     - Save optimized prompts to project library
     - Smart search and reuse with similarity matching
     - Usage tracking and popularity ranking
     - Management tools for browsing and filtering
   ```

2. **Create Example Library**:
   ```bash
   # Provide sample prompts for users
   mkdir -p examples/prompt-library/
   cp .meta-cc/prompts/library/*.md examples/prompt-library/
   git add examples/prompt-library/
   git commit -m "docs: add example prompt library"
   ```

3. **Monitor Usage**:
   - Watch for issues reported
   - Collect feedback on UX
   - Track adoption metrics

### Plan Next Steps (Phase 28.4+)

- Phase 28.4: Performance optimization (indexing, caching)
- Phase 28.5: Cross-project sharing (global library)
- Phase 28.6: Intelligence improvements (effectiveness scoring)
- Phase 28.7: Community library (public repository)

---

## Quick Reference Card

### File Locations
```
capabilities/
├── commands/
│   └── meta-prompt.md          # Extended with save/search
└── prompts/                    # Internal capabilities
    ├── meta-prompt-save.md     # Save functionality
    ├── meta-prompt-search.md   # Search and reuse
    ├── meta-prompt-list.md     # Management tools
    └── meta-prompt-utils.md    # Utility functions

.meta-cc/prompts/
├── library/                    # Saved prompts
│   └── *.md
└── metadata/                   # Future: usage statistics
    └── usage.jsonl
```

### Key Commands
```bash
# User commands
/meta Refine prompt: {query}           # Optimize (with search/save)
/meta prompts/meta-prompt-list         # Browse library
/meta prompts/meta-prompt-list --category={cat}  # Filter
/meta prompts/meta-prompt-list --sort={usage|date|alpha}  # Sort

# Developer validation
yq -f extract '.' file.md              # Validate YAML
ls -lt .meta-cc/prompts/library/       # Browse files
rg "keyword" .meta-cc/prompts/library/ # Search content
```

### Testing
```bash
# Quick validation
make commit                            # Run all checks (no Go tests for this phase)

# Manual validation
./plans/28/validate-stage-1.sh         # Stage 1 checks
./plans/28/validate-stage-2.sh         # Stage 2 checks
./plans/28/validate-stage-3.sh         # Stage 3 checks
```

---

**Implementation Complete!** 🎉

For questions or issues, refer to:
- [PHASE-28-IMPLEMENTATION-PLAN.md](./PHASE-28-IMPLEMENTATION-PLAN.md) - Detailed plan
- [PROMPT-FILE-FORMAT.md](./PROMPT-FILE-FORMAT.md) - File format spec
- [README.md](./README.md) - Phase overview
