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
Short description (2-4 words): full release with CI
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
   Preview: "使用预发布自动化工作流..."

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
