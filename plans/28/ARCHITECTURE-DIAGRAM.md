# Phase 28: Architecture Diagram

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         User Interaction                         │
│                    /meta Refine prompt: XXX                      │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│              Public Capability (Visible in List)                 │
│                                                                   │
│  capabilities/commands/meta-prompt.md                            │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ 1. Pre-Optimization: Search History                    │    │
│  │    └─> get_capability("prompts/meta-prompt-search")    │    │
│  │                                                          │    │
│  │ 2. Normal Optimization: Generate Alternatives           │    │
│  │    └─> analyze() → detect() → generate()              │    │
│  │                                                          │    │
│  │ 3. Post-Optimization: Save Prompt                       │    │
│  │    └─> get_capability("prompts/meta-prompt-save")      │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                   │
└───────────┬───────────────────────────────────┬─────────────────┘
            │                                   │
            ▼ (search)                          ▼ (save)
┌───────────────────────────┐     ┌────────────────────────────────┐
│  Internal Capabilities    │     │  Internal Capabilities         │
│  (Not Visible in List)    │     │  (Not Visible in List)         │
│                           │     │                                │
│  prompts/                 │     │  prompts/                      │
│  meta-prompt-search.md    │     │  meta-prompt-save.md           │
│  ┌─────────────────────┐ │     │  ┌──────────────────────────┐ │
│  │ - Search library    │ │     │  │ - Initialize storage     │ │
│  │ - Calculate similarity│     │  │ - Generate ID            │ │
│  │ - Rank by score     │ │     │  │ - Create frontmatter     │ │
│  │ - Format results    │ │     │  │ - Format content         │ │
│  │ - Handle selection  │ │     │  │ - Write file             │ │
│  │ - Update usage      │ │     │  └──────────────────────────┘ │
│  └─────────────────────┘ │     │                                │
│           │               │     │                                │
│           └──uses─────────┼─────┼────────────────────┐           │
│                           │     │                    │           │
└───────────────────────────┘     └────────────────────┼───────────┘
                                                       │
                                                       ▼
                                  ┌────────────────────────────────┐
                                  │  Utility Capabilities          │
                                  │                                │
                                  │  prompts/                      │
                                  │  meta-prompt-utils.md          │
                                  │  ┌──────────────────────────┐ │
                                  │  │ - extract_keywords()     │ │
                                  │  │ - jaccard_similarity()   │ │
                                  │  │ - parse_frontmatter()    │ │
                                  │  │ - update_usage_count()   │ │
                                  │  └──────────────────────────┘ │
                                  │                                │
                                  └────────────┬───────────────────┘
                                               │
                                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Storage Layer                                │
│                                                                   │
│  .meta-cc/prompts/                                               │
│  ├── library/                                                    │
│  │   ├── release-full-ci-001.md                                 │
│  │   ├── debug-error-002.md                                     │
│  │   └── refactor-logic-003.md                                  │
│  └── metadata/                                                   │
│      └── usage.jsonl (future)                                    │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

## Workflow Diagrams

### Save Workflow

```
User: /meta Refine prompt: 发布新版本
│
├─> meta-prompt.md: analyze() → detect() → generate()
│   │
│   └─> Display alternatives
│
└─> Ask: "Save this prompt? (y/N)"
    │
    ├─> User: "y"
    │   │
    │   └─> meta-prompt-save.md
    │       │
    │       ├─> Initialize: create .meta-cc/prompts/library/
    │       ├─> Ask: category, keywords, description
    │       ├─> Generate ID: release-simple-release-001
    │       ├─> Create frontmatter (YAML)
    │       ├─> Format content (Markdown)
    │       └─> Write file
    │           │
    │           └─> Confirm: "Saved to .meta-cc/prompts/library/release-simple-release-001.md"
    │
    └─> User: "n" or Enter
        │
        └─> Skip (no file created)
```

### Search and Reuse Workflow

```
User: /meta Refine prompt: release new version
│
├─> meta-prompt.md: Pre-optimization search
│   │
│   └─> meta-prompt-search.md
│       │
│       ├─> Extract query keywords: [release, new, version]
│       │
│       ├─> Read library: .meta-cc/prompts/library/*.md
│       │   │
│       │   └─> For each file:
│       │       ├─> Parse frontmatter
│       │       ├─> Extract keywords
│       │       ├─> Calculate similarity (Jaccard)
│       │       └─> Calculate combined score (0.7*sim + 0.3*usage)
│       │
│       ├─> Filter: similarity > 0.2
│       ├─> Sort: by combined_score desc
│       └─> Take top 5
│           │
│           ├─> Found matches?
│           │   │
│           │   ├─> Yes: Display ranked results
│           │   │   │
│           │   │   └─> Ask: "Select 1-5 or Enter to generate new"
│           │   │       │
│           │   │       ├─> User selects number
│           │   │       │   │
│           │   │       │   ├─> meta-prompt-utils.md: update_usage_count()
│           │   │       │   └─> Return optimized prompt
│           │   │       │       │
│           │   │       │       └─> SKIP normal optimization (early exit)
│           │   │       │
│           │   │       └─> User presses Enter
│           │   │           │
│           │   │           └─> Continue to normal optimization
│           │   │
│           │   └─> No: Continue to normal optimization
│           │
│           └─> Continue: analyze() → detect() → generate()
```

### List and Browse Workflow

```
User: /meta prompts/meta-prompt-list
│
└─> meta-prompt-list.md
    │
    ├─> Parse options: --category, --sort, --detail
    │
    ├─> Read library: .meta-cc/prompts/library/*.md
    │   │
    │   └─> For each file:
    │       └─> meta-prompt-utils.md: parse_frontmatter()
    │
    ├─> Apply filters (category)
    ├─> Apply sorting (usage/date/alpha)
    │
    └─> Format output
        │
        ├─> Table view (default)
        │   │
        │   └─> Display:
        │       ┌────┬──────────────────┬──────────┬───────┬──────────────┐
        │       │ #  │ Title            │ Category │ Usage │ Last Updated │
        │       ├────┼──────────────────┼──────────┼───────┼──────────────┤
        │       │ 1  │ Full Release...  │ release  │    5  │ 2 days ago   │
        │       │ 2  │ Error Analysis.. │ debug    │    3  │ 1 week ago   │
        │       └────┴──────────────────┴──────────┴───────┴──────────────┘
        │       Total: 2 prompts across 2 categories
        │
        └─> Detail view (--detail=ID)
            │
            └─> Display:
                ┌──────────────────────────────────────┐
                │ ## Metadata                          │
                │ id: release-full-ci-001              │
                │ title: Full Release with CI          │
                │ category: release                    │
                │ usage_count: 5                       │
                │ ...                                  │
                │                                      │
                │ ## Content                           │
                │ ## Original Prompts                  │
                │ - 发布新版本                         │
                │                                      │
                │ ## Optimized Prompt                  │
                │ 使用预发布自动化工作流...             │
                └──────────────────────────────────────┘
```

## Data Flow

### File Format Structure

```
release-full-ci-001.md
├── YAML Frontmatter (lines 1-12)
│   ├── id: release-full-ci-001
│   ├── title: Full Release with CI Monitoring
│   ├── category: release
│   ├── keywords: [发布, release, 新版本, ci, 监控]
│   ├── created: 2025-10-27T09:00:00Z
│   ├── updated: 2025-10-27T14:22:45Z
│   ├── usage_count: 5
│   ├── effectiveness: 1.0
│   ├── variables: ["{VERSION}"]
│   └── status: active
│
├── Section Separator (line 13)
│   └── ---
│
└── Markdown Content (lines 14+)
    ├── ## Original Prompts
    │   ├── - 提交和发布新版本
    │   └── - 发布新版本
    │
    ├── ## Optimized Prompt
    │   └── 使用预发布自动化工作流完成完整的版本发布：
    │       1. 验证所有测试通过：`make test`
    │       2. 更新版本号和 CHANGELOG
    │       ...
    │
    └── ## Notes (optional)
        └── Usage tips and related prompts
```

## Capability Loading Mechanism

```
MCP Server: list_capabilities()
│
├─> Scans: capabilities/commands/*.md
│   │
│   └─> Returns:
│       ├── meta-prompt (VISIBLE ✓)
│       ├── meta-errors
│       ├── meta-quality-scan
│       └── ...
│
└─> Does NOT scan: capabilities/prompts/*.md (subdirectory, skipped)


MCP Server: get_capability("prompts/meta-prompt-search")
│
├─> Looks for: capabilities/prompts/meta-prompt-search.md
│   │
│   └─> File exists?
│       │
│       ├─> Yes: Load and return content
│       │
│       └─> No: Return error

Key Insight: Subdirectory files are NOT listed but CAN be loaded!
             Perfect for internal capabilities.
```

## Similarity Matching Algorithm

```
Query: "release new version with monitoring"
├─> extract_keywords()
│   └─> [release, new, version, monitoring]
│
Library: .meta-cc/prompts/library/
├─> release-full-ci-001.md
│   ├─> keywords: [发布, release, 新版本, ci, 监控]
│   ├─> usage_count: 5
│   │
│   └─> Calculate similarity:
│       │
│       ├─> Query keywords: [release, new, version, monitoring]
│       ├─> File keywords: [发布, release, 新版本, ci, 监控]
│       │   (translated: [release, release, new-version, ci, monitoring])
│       │
│       ├─> Intersection: [release, monitoring] = 2
│       ├─> Union: [release, new, version, monitoring, 发布, 新版本, ci, 监控] = 8
│       ├─> Jaccard similarity: 2/8 = 0.25
│       │
│       ├─> Usage score: log(5+1) / log(100) = 0.39
│       └─> Combined score: 0.7*0.25 + 0.3*0.39 = 0.175 + 0.117 = 0.292
│
├─> debug-error-002.md
│   ├─> keywords: [debug, error, troubleshoot, 调试]
│   ├─> usage_count: 3
│   │
│   └─> Calculate similarity:
│       ├─> Intersection: [] = 0
│       ├─> Union: [release, new, version, monitoring, debug, error, troubleshoot, 调试] = 8
│       ├─> Jaccard similarity: 0/8 = 0.0
│       └─> Combined score: 0.7*0.0 + 0.3*0.27 = 0.081
│       └─> FILTERED OUT (similarity < 0.2 threshold)
│
└─> Results:
    ├─> Filter: similarity > 0.2
    ├─> Sort: by combined_score desc
    └─> Top 5:
        1. release-full-ci-001 (29% match, 5 uses)
```

## Directory Structure Evolution

### Stage 1: Basic Save

```
project/
├── .meta-cc/
│   └── prompts/
│       └── library/
│           ├── .gitignore
│           └── release-simple-001.md
└── ...
```

### Stage 2: After Reuse

```
project/
├── .meta-cc/
│   └── prompts/
│       └── library/
│           ├── .gitignore
│           ├── release-simple-001.md (usage_count: 0 → 1)
│           ├── debug-error-001.md
│           └── refactor-logic-001.md
└── ...
```

### Stage 3: Mature Library

```
project/
├── .meta-cc/
│   └── prompts/
│       ├── library/
│       │   ├── .gitignore
│       │   ├── release-full-ci-001.md (usage: 5)
│       │   ├── release-simple-002.md (usage: 2)
│       │   ├── release-hotfix-003.md (usage: 1)
│       │   ├── debug-error-001.md (usage: 3)
│       │   ├── debug-performance-002.md (usage: 1)
│       │   ├── refactor-extract-001.md (usage: 2)
│       │   ├── test-coverage-001.md (usage: 4)
│       │   └── docs-api-001.md (usage: 1)
│       └── metadata/ (future Phase 28.4+)
│           └── usage.jsonl
└── ...
```

## Future Enhancements (Phase 28.4+)

### Phase 28.4: Indexing

```
.meta-cc/prompts/
├── library/
│   └── *.md (50+ files)
├── metadata/
│   ├── usage.jsonl
│   └── index.json  ← NEW: Fast lookup
│       ├── keywords: {"release": [1, 2, 3], "debug": [4, 5]}
│       ├── categories: {"release": [1, 2, 3], "debug": [4, 5]}
│       └── usage_ranking: [1, 4, 7, 2, ...]
└── ...

Performance gain: O(n) → O(log n) search
```

### Phase 28.5: Global Library

```
~/.meta-cc/
└── prompts/
    └── library/
        ├── common-release-001.md
        ├── common-debug-001.md
        └── ...

project/.meta-cc/
└── prompts/
    └── library/
        ├── project-specific-001.md
        └── ...

Search strategy:
1. Search project-local library first
2. Search global library second
3. Merge and rank results
```

### Phase 28.6: Effectiveness Scoring

```
Effectiveness calculation:
- Initial: 1.0 (optimistic default)
- After reuse: Track if user edits prompt
  - No edit → effectiveness * 1.1 (capped at 1.0)
  - Minor edit → effectiveness * 0.95
  - Major edit → effectiveness * 0.8
- Factor into ranking:
  combined_score = 0.6*similarity + 0.3*usage + 0.1*effectiveness
```

### Phase 28.7: Community Library

```
Remote: https://github.com/yaleh/meta-cc-prompts
├── library/
│   ├── release/
│   │   ├── full-ci-monitoring.md
│   │   └── simple-release.md
│   ├── debug/
│   └── ...
└── metadata/
    ├── ratings.json
    └── contributors.json

User workflow:
1. Browse: /meta prompts/browse-community
2. Preview: /meta prompts/preview release-full-ci
3. Install: /meta prompts/install release-full-ci-001
4. Contribute: /meta prompts/contribute release-simple-001
```

## Integration Points

### With Existing Meta Commands

```
/meta Refine prompt: XXX
├─> Uses: Prompt Learning System (Phase 28)
└─> Complements: None (standalone)

/meta errors
├─> Uses: MCP query tools (Phase 27)
└─> Could use: Saved error analysis prompts

/meta quality-scan
├─> Uses: MCP query tools + pattern analysis
└─> Could use: Saved quality check prompts

/meta coach
├─> Uses: MCP query tools + comprehensive analysis
└─> Could use: Saved coaching prompts for common issues
```

### With MCP Tools

```
Prompt Learning System (Phase 28)
├─> Does NOT use MCP query tools (zero intrusion)
├─> Pure capability implementation
└─> File system operations only

Future integration (Phase 28.6+):
└─> Could use MCP to analyze prompt effectiveness
    ├─> Query: How often was this prompt edited after reuse?
    └─> Adjust effectiveness score based on usage patterns
```

---

## Legend

```
┌─────┐
│ Box │  Component or module
└─────┘

├─>  Data flow or control flow
│
└─>  Alternative path or option

...  Continuation or omitted details

(VISIBLE ✓)  User-facing component
(usage: N)   Metadata annotation
```

---

For complete details, see:
- [PHASE-28-IMPLEMENTATION-PLAN.md](./PHASE-28-IMPLEMENTATION-PLAN.md)
- [PROMPT-FILE-FORMAT.md](./PROMPT-FILE-FORMAT.md)
- [IMPLEMENTATION-GUIDE.md](./IMPLEMENTATION-GUIDE.md)
- [README.md](./README.md)
