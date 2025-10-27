# Phase 28: Prompt Optimization Learning System

## Quick Reference

**Status**: 📋 Planned
**Estimated Effort**: 12-15 hours
**Code Volume**: ~450 lines (Markdown capabilities + documentation)
**Stages**: 3 (MVP → Search → Management)

## Overview

Implement a pure Capability-driven Prompt learning system that enables users to save, search, and reuse optimized prompts with progressive intelligence.

**Core Philosophy**: Zero intrusion - no new MCP tools, no modifications to `/meta` command, no Go code changes.

## Quick Links

- **[Implementation Plan](./PHASE-28-IMPLEMENTATION-PLAN.md)** - Detailed stage-by-stage plan
- **[Phase 28 in plan.md](../../docs/core/plan.md#phase-28-prompt-优化学习系统详细)** - Context and rationale

## Architecture Summary

### Data Directory
```
<project-root>/.meta-cc/
└── prompts/
    ├── library/              # Saved prompts (flat storage)
    │   ├── release-full-ci-001.md
    │   ├── debug-error-002.md
    │   └── refactor-logic-003.md
    └── metadata/             # Usage statistics
        └── usage.jsonl
```

### Capability Structure
```
capabilities/
├── commands/
│   └── meta-prompt.md        # Extended with save/search
└── prompts/                  # Internal (not listed)
    ├── meta-prompt-save.md
    ├── meta-prompt-search.md
    ├── meta-prompt-list.md
    └── meta-prompt-utils.md
```

## Key Design Decisions

1. **Differentiated Capability Loading**: Leverage native MCP behavior - `list_capabilities()` only shows top-level files, but `get_capability("prompts/xxx")` can load subdirectory files
2. **Flat Storage**: Simple directory structure for CLI-friendly browsing
3. **YAML Frontmatter**: Structured metadata with plain-text content
4. **Progressive Intelligence**: System gets smarter with usage (similarity + frequency)
5. **Optional Save**: Non-intrusive workflow (user confirms)

## Stage Breakdown

### Stage 1: Infrastructure and Save (5-6h)
**Goal**: Implement basic save functionality with auto-initialization

**Deliverables**:
- Auto-create `.meta-cc/prompts/library/` directory
- Save workflow in `prompts/meta-prompt-save.md`
- Extend `commands/meta-prompt.md` with save integration
- YAML frontmatter + Markdown format
- Documentation in CLAUDE.md

**Validation**:
- Directory auto-created on first save
- Valid YAML frontmatter generated
- Files follow naming convention
- Optional save (can skip)

### Stage 2: Search and Reuse (5-6h)
**Goal**: Implement historical search and smart recommendations

**Deliverables**:
- Search capability in `prompts/meta-prompt-search.md`
- Similarity matching (Jaccard + usage weighting)
- Usage tracking (increment usage_count)
- Integration in `commands/meta-prompt.md`

**Validation**:
- Search finds similar prompts (>30% match)
- Top 5 results ranked correctly
- User can select or skip
- Usage count increments on reuse

### Stage 3: Management and Listing (3-4h)
**Goal**: Provide prompt management capabilities

**Deliverables**:
- List capability in `prompts/meta-prompt-list.md`
- Filter by category, sort by usage/date
- Summary statistics
- Detail view
- Complete user guide

**Validation**:
- List all prompts correctly
- Filter and sort work as expected
- Statistics are accurate
- Documentation complete

## Testing Approach

**User Validation** (no unit tests for capabilities):

1. **Functional Testing**: Manual workflow validation
2. **Edge Case Testing**: Empty library, invalid YAML, special characters
3. **Performance Testing**: 20+ prompts library
4. **UX Testing**: Real-world usage scenarios

**Validation Checklist**: See [Implementation Plan](./PHASE-28-IMPLEMENTATION-PLAN.md#testing-strategy)

## File Size Budget

| Component | Estimate | Actual |
|-----------|----------|--------|
| meta-prompt-save.md | 100 lines | TBD |
| meta-prompt-search.md | 120 lines | TBD |
| meta-prompt-list.md | 70 lines | TBD |
| meta-prompt-utils.md | 20 lines | TBD |
| meta-prompt.md updates | 90 lines | TBD |
| Documentation | 230 lines | TBD |
| **Total** | **630 lines** | **TBD** |

⚠️ **Note**: Initial estimate was 450 lines, current plan is ~630 lines. Consider splitting Stage 3 if needed to stay under 500 lines per phase.

## Success Criteria

### MVP (After Stage 1)
- ✅ Users can save optimized prompts
- ✅ Storage directory auto-created
- ✅ Valid file format generated
- ✅ Documentation updated

### Complete (After Stage 3)
- ✅ Users can search and reuse prompts
- ✅ System recommends similar prompts
- ✅ Usage tracking works correctly
- ✅ Management tools available
- ✅ Complete user guide published

## Rollout Strategy

1. **Week 1**: Stage 1 implementation + internal testing
2. **Week 2**: Stage 2 implementation + beta testing (2-3 users)
3. **Week 3**: Stage 3 implementation + public release

## Future Enhancements

- **Phase 28.4**: Performance optimization (indexing, caching)
- **Phase 28.5**: Cross-project sharing (global library)
- **Phase 28.6**: Intelligence improvements (effectiveness scoring)
- **Phase 28.7**: Community library (public repository)

## Questions?

- See [Implementation Plan](./PHASE-28-IMPLEMENTATION-PLAN.md) for detailed breakdowns
- Check [Phase 28 in plan.md](../../docs/core/plan.md) for context
- Review existing `/meta Refine prompt` capability in `capabilities/commands/meta-prompt.md`

## Getting Started

Ready to implement? Start with Stage 1:

```bash
# 1. Create capability directory
mkdir -p capabilities/prompts

# 2. Create meta-prompt-save.md
vim capabilities/prompts/meta-prompt-save.md

# 3. Test with real prompt
/meta Refine prompt: 发布新版本

# 4. Validate save workflow
ls -la .meta-cc/prompts/library/
```

See [PHASE-28-IMPLEMENTATION-PLAN.md](./PHASE-28-IMPLEMENTATION-PLAN.md) for complete implementation details.
