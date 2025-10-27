# Phase 28: Prompt Optimization Learning System - Implementation Plan

## Overview

**Goal**: Implement a pure Capability-driven Prompt learning system that enables users to save, search, and reuse optimized prompts with progressive intelligence.

**Core Philosophy**: Zero intrusion - no new MCP tools, no modifications to `/meta` command, no Go code changes. Pure capability implementation leveraging existing infrastructure.

**Estimated Effort**: 12-15 hours across 3 stages
**Code Volume**: ~450 lines (Markdown capabilities + documentation)
**Test Coverage**: Capability validation through user testing

---

## Architecture Summary

### Data Structure

```
<project-root>/.meta-cc/
├── prompts/
│   ├── library/                    # Optimized prompts (flat storage)
│   │   ├── release-full-ci-001.md
│   │   ├── debug-error-002.md
│   │   └── refactor-logic-003.md
│   └── metadata/                   # Usage statistics (optional)
│       └── usage.jsonl
└── config.json                     # Project-level config (optional)
```

### Capability Structure

```
capabilities/
├── commands/                       # Public capabilities (visible in list_capabilities)
│   ├── meta-prompt.md             # Extended version with auto-init and save
│   └── ...
└── prompts/                        # Internal capabilities (NOT visible in list_capabilities)
    ├── meta-prompt-search.md      # Search historical prompts
    ├── meta-prompt-save.md        # Save optimized prompt
    ├── meta-prompt-list.md        # List prompts by category/usage
    └── meta-prompt-utils.md       # Utility functions
```

**Key Discovery**: Existing MCP capability loading mechanism natively supports differentiation:
- `list_capabilities()` only scans top-level `*.md` files (no subdirectory recursion)
- `get_capability("prompts/xxx")` can load subdirectory files
- Zero configuration needed for internal capabilities

---

## Stage 1: Infrastructure and Save Functionality (MVP)

**Objective**: Implement prompt saving with auto-initialization

**Effort**: 5-6 hours
**Code Volume**: ~180 lines

### Files to Create/Modify

1. **Create `capabilities/prompts/` directory**
   - New subdirectory for internal capabilities
   - Not visible in `list_capabilities` output

2. **Create `capabilities/prompts/meta-prompt-save.md`** (~100 lines)
   - Implements save logic
   - Auto-creates `.meta-cc/prompts/library/` directory
   - Generates YAML frontmatter + Markdown file
   - Handles file naming convention
   - Validates input and confirms with user

3. **Extend `capabilities/commands/meta-prompt.md`** (~50 lines)
   - Add post-optimization save workflow
   - Integrate auto-initialization
   - Add "Save this prompt?" confirmation
   - Call internal `prompts/meta-prompt-save` capability

4. **Create `.meta-cc/prompts/library/.gitignore`** (~10 lines)
   - Ignore pattern for local prompt library
   - Allow users to commit selectively

5. **Update `CLAUDE.md`** (~20 lines)
   - Add FAQ entry for prompt learning system
   - Document usage workflow
   - Explain `.meta-cc/` directory structure

### TDD Iteration

**Test 1: Auto-initialization**
- **Validation**: Check if `.meta-cc/prompts/library/` is created on first save
- **Approach**: Manual test with user confirmation
- **Acceptance**: Directory structure created silently

**Test 2: Save workflow**
- **Validation**: Generate valid prompt file with proper frontmatter
- **Approach**: Save a test prompt and inspect output file
- **Acceptance Criteria**:
  - File naming follows convention: `{category}-{description}-{id}.md`
  - YAML frontmatter includes all required fields
  - Markdown content preserves original and optimized prompts
  - Timestamps are ISO8601 format

**Test 3: File format validation**
- **Validation**: Ensure YAML frontmatter is valid
- **Approach**: Parse with `yq` or manual inspection
- **Acceptance**: All fields present and correctly formatted

### Implementation Steps

#### Step 1.1: Create internal capability infrastructure
```bash
mkdir -p capabilities/prompts
```

#### Step 1.2: Implement meta-prompt-save.md

**Content Structure**:
```markdown
---
name: meta-prompt-save
description: Internal capability to save optimized prompts to project library
category: internal
---

λ(prompt_original, prompt_optimized, category, keywords) → saved_file

initialize :: Project_Root → Storage_Path
initialize(P) = {
  storage: P + "/.meta-cc/prompts/library/",
  create_if_missing: mkdir -p storage,
  gitignore: create(storage + ".gitignore", "# Local prompt library\n")
}

generate_id :: Storage_Path → Unique_ID
generate_id(S) = {
  existing: ls S | grep -E "^{category}-.*-[0-9]{3}.md$",
  max_id: max(extract_number(existing)),
  next_id: sprintf("%03d", max_id + 1)
}

create_frontmatter :: (ID, Category, Keywords, Timestamp) → YAML
create_frontmatter(id, cat, kw, ts) = yaml({
  id: cat + "-" + description + "-" + id,
  title: infer_title(prompt_optimized),
  category: cat,
  keywords: kw,
  created: ts,
  updated: ts,
  usage_count: 0,
  effectiveness: 1.0,
  variables: extract_variables(prompt_optimized),
  status: "active"
})

format_content :: (Original, Optimized) → Markdown
format_content(O, P) = markdown({
  section_1: "## Original Prompts",
  original: O,
  section_2: "## Optimized Prompt",
  optimized: P
})

save :: (Frontmatter, Content, Filename) → File
save(F, C, N) = {
  file: storage_path + "/" + N,
  write: "---\n" + F + "\n---\n\n" + C,
  confirm: "Saved to: " + file
}

workflow:
1. Auto-initialize storage directory
2. Generate unique ID
3. Prompt user for category and keywords
4. Create frontmatter
5. Format markdown content
6. Save file
7. Confirm to user

user_interaction:
- "What category is this prompt? (e.g., release, debug, refactor)"
- "Enter keywords (comma-separated): "
- "Short description (2-4 words): "
```

#### Step 1.3: Extend meta-prompt.md

**Add to end of existing meta-prompt.md**:
```markdown

## Post-Optimization Workflow

After generating optimized prompts, offer to save for future reuse:

save_prompt :: Optimized_Prompts → Optional[Saved_File]
save_prompt(P) = {
  ask: "Would you like to save this optimized prompt for future reuse? (y/N)",

  if (user_confirms):
    call: get_capability("prompts/meta-prompt-save"),
    pass: {
      prompt_original: original_prompt,
      prompt_optimized: P.recommendation,
      timestamp: now()
    },

  else:
    skip: "Prompt not saved. You can refine it again anytime with /meta Refine prompt: ..."
}

constraints:
- non_intrusive: save is optional, default is skip
- seamless: integrate into existing workflow
- educational: explain benefits of saving prompts
```

#### Step 1.4: Update CLAUDE.md

**Add to FAQ section**:
```markdown
**Q: How does the prompt learning system work?**
A: After using `/meta Refine prompt: XXX`, you can save the optimized version to `.meta-cc/prompts/library/`. The system will recommend these saved prompts when you try similar prompts in the future, making you more efficient over time.

**Q: Where are saved prompts stored?**
A: Project-local storage in `.meta-cc/prompts/library/` (not tracked by git by default). You can commit selectively if you want to share with your team.

**Q: Can I search my saved prompts?**
A: Yes, use `/meta prompts/meta-prompt-list` to browse by category or usage frequency. Files are plain text, so you can also use `grep`, `ack`, or `rg` to search.
```

### Acceptance Criteria

- ✅ `.meta-cc/prompts/library/` directory is auto-created on first save
- ✅ Users can save optimized prompts after refinement
- ✅ Generated files follow naming convention: `{category}-{description}-{id}.md`
- ✅ YAML frontmatter includes all required fields (id, title, category, keywords, timestamps, usage_count, effectiveness, variables, status)
- ✅ Markdown content includes both original and optimized prompts
- ✅ Save workflow is optional (user can skip)
- ✅ `.gitignore` is created automatically
- ✅ Documentation updated in CLAUDE.md

### Dependencies

None - this is the foundation stage.

---

## Stage 2: Search and Reuse Functionality

**Objective**: Implement historical search and smart recommendations

**Effort**: 5-6 hours
**Code Volume**: ~180 lines

### Files to Create/Modify

1. **Create `capabilities/prompts/meta-prompt-search.md`** (~120 lines)
   - Implements similarity matching algorithm
   - Searches historical prompts
   - Ranks results by relevance and usage frequency
   - Returns top N matches

2. **Update `capabilities/commands/meta-prompt.md`** (~40 lines)
   - Add pre-optimization history search
   - Display matched historical prompts
   - Allow user to select historical version or generate new
   - Update usage_count after selection

3. **Create `capabilities/prompts/meta-prompt-utils.md`** (~20 lines)
   - Shared utility functions
   - Keyword extraction
   - Similarity calculation
   - File metadata parsing

### TDD Iteration

**Test 1: Similarity matching**
- **Validation**: Correctly identify similar prompts based on keywords
- **Approach**: Create test prompts with known overlapping keywords
- **Acceptance**: Jaccard similarity score >0.3 returns match

**Test 2: Search workflow**
- **Validation**: Search finds relevant historical prompts
- **Approach**: Save 2-3 prompts, then search with similar query
- **Acceptance**: Relevant prompts appear in top 3 results

**Test 3: Usage tracking**
- **Validation**: usage_count increments after reuse
- **Approach**: Reuse a saved prompt and check metadata update
- **Acceptance**: YAML frontmatter shows incremented count

**Test 4: No history case**
- **Validation**: Gracefully handle no matches
- **Approach**: Search with completely unrelated keywords
- **Acceptance**: Falls back to normal optimization flow

### Implementation Steps

#### Step 2.1: Implement meta-prompt-utils.md

**Content Structure**:
```markdown
---
name: meta-prompt-utils
description: Utility functions for prompt library management
category: internal
---

## Utility Functions

extract_keywords :: String → [String]
extract_keywords(S) = {
  # Extract significant words from prompt
  words: tokenize(S, /\s+/),
  filter: ¬stopwords ∧ length > 2,
  normalize: lowercase(words)
}

jaccard_similarity :: ([String], [String]) → Float
jaccard_similarity(A, B) = {
  intersection: |A ∩ B|,
  union: |A ∪ B|,
  score: intersection / union  # Range: 0.0 to 1.0
}

parse_frontmatter :: File → Metadata
parse_frontmatter(F) = {
  content: read(F),
  yaml_block: extract_between("---", "---", content),
  metadata: parse_yaml(yaml_block)
}

update_usage_count :: File → File
update_usage_count(F) = {
  meta: parse_frontmatter(F),
  meta.usage_count += 1,
  meta.updated = now(),
  write_frontmatter(F, meta)
}
```

#### Step 2.2: Implement meta-prompt-search.md

**Content Structure**:
```markdown
---
name: meta-prompt-search
description: Search historical prompts with similarity matching
category: internal
---

λ(query_prompt) → ranked_matches | ∀prompt ∈ library

search :: Query_Prompt → [Matched_Prompts]
search(Q) = {
  library: ".meta-cc/prompts/library/*.md",

  if (¬exists(library)):
    return: [],

  query_keywords: extract_keywords(Q),

  candidates: ∀file ∈ library: {
    meta: parse_frontmatter(file),
    original_keywords: meta.keywords + extract_keywords(meta.original),
    similarity: jaccard_similarity(query_keywords, original_keywords),
    usage_score: log(meta.usage_count + 1),  # Logarithmic scaling
    combined_score: similarity * 0.7 + usage_score * 0.3
  },

  matches: filter(candidates, similarity > 0.2),  # Threshold
  ranked: sort_desc(matches, combined_score),
  top_n: take(ranked, 5)
}

format_results :: [Matches] → Display
format_results(M) = {
  if (|M| == 0):
    "No similar prompts found in history. Generating fresh recommendations...",

  else:
    header: "Found {|M|} similar prompts in your library:",
    list: ∀m ∈ M: {
      index: idx + 1,
      title: m.meta.title,
      category: m.meta.category,
      usage: m.meta.usage_count + " uses",
      similarity: sprintf("%.0f%% match", m.similarity * 100),
      preview: truncate(m.optimized, 100)
    },
    footer: "Select a number to reuse, or press Enter to generate new optimization."
}

reuse :: (User_Selection, Matches) → Prompt
reuse(S, M) = {
  if (S == "" || S == "0"):
    return: generate_new(),

  else:
    selected: M[S - 1],
    update_usage_count(selected.file),
    return: selected.optimized,
    confirm: "Reusing prompt from library (updated usage count)."
}

workflow:
1. Extract keywords from user's query prompt
2. Search library for similar prompts
3. Calculate similarity + usage scores
4. Rank and display top 5 matches
5. Allow user to select or generate new
6. Update usage_count if reused
```

#### Step 2.3: Update meta-prompt.md

**Add at beginning of optimization workflow**:
```markdown

## Pre-Optimization: History Search

Before generating new optimizations, search for similar historical prompts:

search_history :: Raw_Prompt → Optional[Historical_Matches]
search_history(P) = {
  call: get_capability("prompts/meta-prompt-search"),
  pass: {query_prompt: P},

  matches: search(P),

  if (|matches| > 0):
    display: format_results(matches),
    ask: "Select a prompt to reuse (1-5), or press Enter to generate new:",

    if (user_selects_existing):
      return: reuse(selection, matches),
      skip: generate(alternatives),  # Skip normal optimization

    else:
      continue: generate(alternatives)  # Normal optimization flow

  else:
    continue: generate(alternatives)  # No matches, normal flow
}

integration :: Workflow_Order
integration() = {
  step_1: search_history(prompt_raw),  # NEW: Check history first
  step_2: analyze(history),             # Existing step
  step_3: detect(gaps),                 # Existing step
  step_4: generate(alternatives),       # Existing step
  step_5: save_prompt(result)           # From Stage 1
}
```

### Acceptance Criteria

- ✅ Search finds similar prompts based on keyword overlap
- ✅ Similarity scoring works correctly (Jaccard + usage weighting)
- ✅ Top N matches displayed with relevance score
- ✅ User can select historical prompt to reuse
- ✅ User can skip and generate new optimization
- ✅ Usage count increments after reuse
- ✅ Updated timestamp reflects reuse
- ✅ Graceful handling when no history exists
- ✅ Graceful handling when no matches found

### Dependencies

- **Stage 1 complete**: Requires save functionality and data structure

---

## Stage 3: Management and Listing

**Objective**: Provide prompt management and browsing capabilities

**Effort**: 3-4 hours
**Code Volume**: ~90 lines

### Files to Create/Modify

1. **Create `capabilities/prompts/meta-prompt-list.md`** (~70 lines)
   - List all saved prompts
   - Filter by category
   - Sort by usage frequency or date
   - Display summary statistics

2. **Update `CLAUDE.md`** (~20 lines)
   - Document listing and management features
   - Add usage examples
   - Update FAQ

### TDD Iteration

**Test 1: List all prompts**
- **Validation**: Display all saved prompts
- **Approach**: Save 3-5 prompts, then list them
- **Acceptance**: All prompts appear with metadata

**Test 2: Filter by category**
- **Validation**: Only show prompts in specified category
- **Approach**: List with category filter
- **Acceptance**: Filtered results match category

**Test 3: Sort by usage**
- **Validation**: Prompts sorted by usage_count descending
- **Approach**: Reuse some prompts to vary counts, then sort
- **Acceptance**: Order matches usage frequency

**Test 4: Summary statistics**
- **Validation**: Show total prompts, categories, most used
- **Approach**: Display summary with multiple prompts
- **Acceptance**: Stats are accurate

### Implementation Steps

#### Step 3.1: Implement meta-prompt-list.md

**Content Structure**:
```markdown
---
name: meta-prompt-list
description: List and manage saved prompts in library
category: internal
---

λ(filter_category, sort_by) → prompt_listing

list :: (Category_Filter, Sort_Order) → [Prompts]
list(C, S) = {
  library: ".meta-cc/prompts/library/*.md",

  if (¬exists(library)):
    return: "No prompts saved yet. Use /meta Refine prompt to create your first one!",

  prompts: ∀file ∈ library: {
    meta: parse_frontmatter(file),
    filter: (C == null) ∨ (meta.category == C)
  },

  sorted: {
    if (S == "usage"): sort_desc(prompts, usage_count),
    if (S == "date"): sort_desc(prompts, updated),
    if (S == "alpha"): sort_asc(prompts, title),
    default: sort_desc(prompts, updated)  # Most recent first
  }
}

format_table :: [Prompts] → Display
format_table(P) = {
  header: "| # | Title | Category | Usage | Last Updated |",
  separator: "|---|-------|----------|-------|--------------|",
  rows: ∀p ∈ P: {
    index: idx + 1,
    title: truncate(p.title, 30),
    category: p.category,
    usage: p.usage_count,
    updated: relative_time(p.updated)  # "2 days ago"
  },
  footer: sprintf("Total: %d prompts across %d categories", |P|, |unique(P.category)|)
}

summary_stats :: [Prompts] → Statistics
summary_stats(P) = {
  total: |P|,
  categories: |unique(P.category)|,
  total_uses: sum(P.usage_count),
  most_used: argmax(P, usage_count),
  recent: take(sort_desc(P, updated), 5)
}

detail_view :: Prompt_ID → Full_Content
detail_view(ID) = {
  file: find_by_id(ID),
  meta: parse_frontmatter(file),
  content: read_content(file),

  display: {
    section_1: "## Metadata",
    yaml: format_yaml(meta),
    section_2: "## Content",
    markdown: content
  }
}

workflow:
1. Read all files from library
2. Apply category filter (if specified)
3. Sort by specified order
4. Format as table
5. Display summary statistics
6. Allow user to view details of specific prompt

usage_examples:
- "List all prompts": list(null, "date")
- "List release prompts": list("release", "usage")
- "Show most used": list(null, "usage")
- "View details": detail_view("release-full-ci-001")
```

#### Step 3.2: Update CLAUDE.md

**Add to FAQ section**:
```markdown
**Q: How do I browse my saved prompts?**
A: Three ways:
1. Use `/meta prompts/meta-prompt-list` for formatted table view
2. Use shell commands: `ls .meta-cc/prompts/library/`
3. Use grep/ripgrep: `rg "keyword" .meta-cc/prompts/library/`

**Q: Can I delete or edit saved prompts?**
A: Yes, they're just markdown files. You can:
- Delete: `rm .meta-cc/prompts/library/release-full-ci-001.md`
- Edit: `vim .meta-cc/prompts/library/release-full-ci-001.md`
- Archive: Set `status: archived` in YAML frontmatter

**Q: Can I share prompts with my team?**
A: Yes, commit `.meta-cc/prompts/library/` to git:
```bash
git add .meta-cc/prompts/library/release-*.md
git commit -m "docs: share release process prompts"
```

**Q: How do I back up my prompt library?**
A: The library is in `.meta-cc/prompts/library/` - just copy the directory or commit to git.
```

### Acceptance Criteria

- ✅ List all saved prompts in tabular format
- ✅ Filter by category works correctly
- ✅ Sort by usage frequency works correctly
- ✅ Sort by date works correctly
- ✅ Summary statistics are accurate (total, categories, most used)
- ✅ Detail view shows complete prompt content
- ✅ Graceful handling when library is empty
- ✅ Documentation includes usage examples and management tips

### Dependencies

- **Stage 1 complete**: Requires save functionality and data structure
- **Stage 2 complete**: Reuses utility functions

---

## Documentation Updates

### Files to Update

1. **CLAUDE.md** (covered in each stage)
   - FAQ entries for prompt learning system
   - Usage examples
   - Management tips

2. **Create `docs/guides/prompt-learning-system.md`** (~150 lines)
   - Complete user guide
   - Architecture overview
   - Workflow examples
   - Best practices
   - CLI integration tips

3. **Update `docs/reference/unified-meta-command.md`** (~20 lines)
   - Add prompt learning capabilities
   - Document internal vs. public capabilities

4. **Update `README.md`** (~30 lines)
   - Mention prompt learning system in features
   - Link to guide

### Content for docs/guides/prompt-learning-system.md

```markdown
# Prompt Learning System

## Overview

The Prompt Learning System enables you to save, search, and reuse optimized prompts, creating a project-specific knowledge base that becomes more intelligent over time.

## Quick Start

### 1. Optimize a prompt
```
/meta Refine prompt: 发布新版本
```

### 2. Save the optimized version
After seeing the recommendations, choose to save:
```
Would you like to save this optimized prompt? (y/N): y
Category (e.g., release, debug, refactor): release
Keywords (comma-separated): 发布, release, 新版本, ci
Short description: full release with CI
```

### 3. Reuse saved prompts
Next time you use a similar prompt:
```
/meta Refine prompt: release new version
```

The system will automatically suggest your saved prompt:
```
Found 1 similar prompt in your library:
1. Full Release with CI [release] (90% match, 2 uses)
   "使用预发布自动化工作流..."

Select 1 to reuse, or press Enter to generate new:
```

## Architecture

### Storage Structure
```
.meta-cc/
└── prompts/
    ├── library/              # Your saved prompts
    │   ├── release-*.md
    │   ├── debug-*.md
    │   └── refactor-*.md
    └── metadata/             # Usage statistics
        └── usage.jsonl
```

### File Format
Each saved prompt is a markdown file with YAML frontmatter:

```markdown
---
id: release-full-ci-001
title: Full Release with CI Monitoring
category: release
keywords: [发布, release, 新版本, ci, 监控]
created: 2025-10-27T09:00:00Z
updated: 2025-10-27T09:10:00Z
usage_count: 2
effectiveness: 1.0
variables: [VERSION]
status: active
---

## Original Prompts
- 提交和发布新版本
- 发布新版本

## Optimized Prompt
使用预发布自动化工作流...
```

## Usage Patterns

### Browse your library
```bash
# Via capability
/meta prompts/meta-prompt-list

# Via shell
ls -lt .meta-cc/prompts/library/

# Search by keyword
rg "release" .meta-cc/prompts/library/
```

### Filter by category
```bash
# List only release prompts
ls .meta-cc/prompts/library/release-*.md

# Count by category
ls .meta-cc/prompts/library/ | cut -d'-' -f1 | sort | uniq -c
```

### View usage statistics
```bash
# Most used prompts
grep "usage_count:" .meta-cc/prompts/library/*.md | sort -t: -k3 -nr | head -5
```

## Best Practices

### Categorization
Use consistent categories:
- `release`: Version releases, deployments
- `debug`: Error investigation, troubleshooting
- `refactor`: Code restructuring, cleanup
- `test`: Test writing, coverage improvement
- `docs`: Documentation updates
- `feature`: New feature implementation

### Keyword Selection
Include:
- English and native language terms
- Common abbreviations (CI, API, DB)
- Domain-specific terminology
- Synonyms and related terms

### When to Save
Save prompts when:
- You'll likely use similar prompts again
- The optimization is significantly better
- It captures project-specific patterns
- It includes useful constraints or context

### Maintenance
- Review and archive unused prompts quarterly
- Update keywords based on search patterns
- Share team-wide prompts via git
- Clean up duplicates regularly

## Advanced Usage

### Share with team
```bash
# Add to git
git add .meta-cc/prompts/library/
git commit -m "docs: share prompt library"

# Cherry-pick specific prompts
git add .meta-cc/prompts/library/release-*.md
git commit -m "docs: share release prompts"
```

### Backup and restore
```bash
# Backup
cp -r .meta-cc/prompts ~/backups/project-prompts-$(date +%Y%m%d)

# Restore
cp -r ~/backups/project-prompts-20251027 .meta-cc/prompts
```

### Export to other projects
```bash
# Copy useful prompts to another project
cp .meta-cc/prompts/library/release-*.md \
   ../other-project/.meta-cc/prompts/library/
```

## CLI Integration

### Use with jq
```bash
# Extract all keywords
for f in .meta-cc/prompts/library/*.md; do
  yq -f extract '.keywords' "$f"
done | sort -u

# Find high-usage prompts
for f in .meta-cc/prompts/library/*.md; do
  echo "$(yq -f extract '.usage_count' "$f") $f"
done | sort -rn | head -5
```

### Use with fzf (fuzzy finder)
```bash
# Interactive prompt selection
fd -e md . .meta-cc/prompts/library/ | \
  fzf --preview 'bat --color=always {}'
```

## Troubleshooting

### "No prompts found"
- Check if `.meta-cc/prompts/library/` exists
- Verify you've saved at least one prompt
- Check file permissions

### "No similar prompts found"
- Your keywords might be too specific
- Try broader search terms
- Consider saving more diverse prompts

### "Can't update usage count"
- Check file write permissions
- Verify YAML frontmatter is valid
- Try manually editing the file

## Future Enhancements

Planned features (Phase 28.4+):
- Effectiveness scoring based on reuse patterns
- Cross-project prompt sharing
- Community prompt library
- Automatic prompt categorization
- Smart keyword extraction
```

---

## Testing Strategy

Since this is a pure capability implementation, testing focuses on user validation rather than automated unit tests.

### Validation Checklist

**Stage 1 Validation**:
- [ ] Create new project, trigger first save → directory auto-created
- [ ] Save 3 different prompts with different categories
- [ ] Verify YAML frontmatter is valid (use `yq`)
- [ ] Verify file naming follows convention
- [ ] Verify .gitignore is created
- [ ] Try saving without confirmation (skip works)

**Stage 2 Validation**:
- [ ] Reuse exact same prompt → finds 100% match
- [ ] Reuse similar prompt with overlapping keywords → finds >70% match
- [ ] Reuse completely different prompt → no matches, generates new
- [ ] Select historical prompt → usage_count increments
- [ ] Select "generate new" → normal optimization flow
- [ ] Verify updated timestamp after reuse

**Stage 3 Validation**:
- [ ] List all prompts → table displays correctly
- [ ] Filter by category → only matching category shown
- [ ] Sort by usage → order is correct
- [ ] Sort by date → most recent first
- [ ] View details → full content displayed
- [ ] Empty library → graceful message

### Edge Cases to Test

1. **Empty library**: First-time user experience
2. **Large library**: 20+ prompts (performance check)
3. **Invalid YAML**: Corrupted frontmatter handling
4. **Missing fields**: Graceful degradation
5. **Concurrent saves**: Two users saving simultaneously
6. **Special characters**: Unicode in keywords and content
7. **Long prompts**: >2000 characters

---

## Risk Assessment

### Low Risk
- ✅ No Go code changes → No compilation issues
- ✅ No MCP tool changes → No protocol issues
- ✅ Pure capability → Isolated from core functionality
- ✅ Optional feature → Can be disabled/ignored

### Medium Risk
- ⚠️ File I/O operations in capabilities → Potential permission issues
- ⚠️ YAML parsing in capabilities → Potential format errors
- ⚠️ User confirmation prompts → UX complexity

### Mitigation Strategies
- **Permission issues**: Document required permissions in troubleshooting
- **Format errors**: Provide validation script for users
- **UX complexity**: Clear prompts, sensible defaults, skip options

---

## Success Metrics

### Quantitative
- User saves ≥3 prompts within first week
- User reuses saved prompt at least once
- Search finds relevant match ≥50% of the time
- Zero file corruption issues

### Qualitative
- Users report time savings from reuse
- Users find search results relevant
- Users understand save/reuse workflow
- Documentation is clear and sufficient

---

## Rollout Plan

### Phase 1: Internal Testing (Week 1)
- Developer self-testing with real projects
- Fix critical bugs and UX issues
- Refine documentation

### Phase 2: Beta Testing (Week 2)
- Share with 2-3 beta users
- Collect feedback on workflow
- Adjust based on usage patterns

### Phase 3: Public Release (Week 3)
- Update CHANGELOG
- Announce in README
- Create example prompts library
- Monitor for issues

---

## Future Enhancements (Post-Phase 28)

### Phase 28.4: Performance Optimization
- Add index file for faster search
- Cache frontmatter metadata
- Optimize similarity calculation

### Phase 28.5: Cross-Project Sharing
- Global prompt library in `~/.meta-cc/`
- Merge project + global results
- Import/export functionality

### Phase 28.6: Intelligence Improvements
- Effectiveness scoring based on reuse patterns
- Automatic keyword extraction
- Context-aware recommendations
- Learning from edits

### Phase 28.7: Community Library
- Public prompt repository
- Contribution workflow
- Rating and reviews
- Moderation system

---

## Appendix: File Size Estimates

### Capability Files
- `capabilities/prompts/meta-prompt-save.md`: ~100 lines
- `capabilities/prompts/meta-prompt-search.md`: ~120 lines
- `capabilities/prompts/meta-prompt-list.md`: ~70 lines
- `capabilities/prompts/meta-prompt-utils.md`: ~20 lines
- Updates to `capabilities/commands/meta-prompt.md`: ~90 lines
- **Total capability code**: ~400 lines

### Documentation
- `docs/guides/prompt-learning-system.md`: ~150 lines
- Updates to `CLAUDE.md`: ~50 lines
- Updates to other docs: ~30 lines
- **Total documentation**: ~230 lines

### Grand Total: ~630 lines
**Exceeds initial estimate of 450 lines by ~40%**

**Recommendation**: Consider splitting Stage 3 into a separate optional phase if needed to stay under 500 lines.

---

## Conclusion

This implementation plan provides a clear path to building a pure-capability prompt learning system with zero intrusion to existing infrastructure. The staged approach ensures incremental value delivery while maintaining manageable scope for each iteration.

**Key Success Factors**:
1. Leverage existing capability loading mechanism
2. Focus on user experience and simplicity
3. Maintain CLI-friendly plain-text format
4. Provide clear documentation and examples
5. Validate through real-world usage

**Next Steps**:
1. Review and approve this plan
2. Create iteration 28 directory structure
3. Begin Stage 1 implementation
4. Validate with real usage before proceeding to next stage
