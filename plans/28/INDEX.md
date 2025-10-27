# Phase 28: Prompt Optimization Learning System - Documentation Index

## Quick Navigation

| Document | Purpose | Audience | Length |
|----------|---------|----------|--------|
| **[README.md](./README.md)** | Phase overview and quick reference | All | ~180 lines |
| **[PHASE-28-IMPLEMENTATION-PLAN.md](./PHASE-28-IMPLEMENTATION-PLAN.md)** | Complete implementation plan with TDD iterations | Developers | ~1,050 lines |
| **[IMPLEMENTATION-GUIDE.md](./IMPLEMENTATION-GUIDE.md)** | Step-by-step implementation guide | Developers | ~1,150 lines |
| **[PROMPT-FILE-FORMAT.md](./PROMPT-FILE-FORMAT.md)** | File format specification and examples | Developers/Users | ~640 lines |
| **[ARCHITECTURE-DIAGRAM.md](./ARCHITECTURE-DIAGRAM.md)** | Visual architecture and workflow diagrams | All | ~570 lines |

**Total Documentation**: ~3,590 lines

---

## For Different Audiences

### 👤 Project Managers / Stakeholders

**Start here**:
1. [README.md](./README.md) - Phase overview, goals, and success criteria
2. [ARCHITECTURE-DIAGRAM.md](./ARCHITECTURE-DIAGRAM.md) - Visual overview
3. [PHASE-28-IMPLEMENTATION-PLAN.md](./PHASE-28-IMPLEMENTATION-PLAN.md) - Stage breakdown and timeline

**Key Information**:
- **Estimated Effort**: 12-15 hours across 3 stages
- **Code Volume**: ~450 lines (Markdown capabilities + documentation)
- **Risk Level**: Low (zero intrusion, pure capability implementation)
- **Dependencies**: None (leverages existing infrastructure)

### 👨‍💻 Developers (Implementation)

**Start here**:
1. [IMPLEMENTATION-GUIDE.md](./IMPLEMENTATION-GUIDE.md) - Complete step-by-step guide
2. [PROMPT-FILE-FORMAT.md](./PROMPT-FILE-FORMAT.md) - File format specification
3. [PHASE-28-IMPLEMENTATION-PLAN.md](./PHASE-28-IMPLEMENTATION-PLAN.md) - Detailed requirements

**Implementation Path**:
```bash
# 1. Read implementation guide
cat plans/28/IMPLEMENTATION-GUIDE.md

# 2. Start Stage 1
mkdir -p capabilities/prompts
vim capabilities/prompts/meta-prompt-save.md

# 3. Follow validation checklist
# (See IMPLEMENTATION-GUIDE.md Stage 1 Validation)
```

### 🧪 QA / Testers

**Start here**:
1. [IMPLEMENTATION-GUIDE.md](./IMPLEMENTATION-GUIDE.md) - Validation checklists
2. [PHASE-28-IMPLEMENTATION-PLAN.md](./PHASE-28-IMPLEMENTATION-PLAN.md) - Acceptance criteria

**Testing Focus**:
- **Stage 1**: Save workflow, file format validation
- **Stage 2**: Search accuracy, similarity matching, usage tracking
- **Stage 3**: List/filter/sort functionality, edge cases
- **Integration**: End-to-end workflows, performance, cross-project isolation

### 📚 Documentation Writers

**Start here**:
1. [PHASE-28-IMPLEMENTATION-PLAN.md](./PHASE-28-IMPLEMENTATION-PLAN.md#documentation-updates) - Documentation requirements
2. [PROMPT-FILE-FORMAT.md](./PROMPT-FILE-FORMAT.md) - Technical specification
3. [ARCHITECTURE-DIAGRAM.md](./ARCHITECTURE-DIAGRAM.md) - Visual aids

**Content to Create**:
- User guide: `docs/guides/prompt-learning-system.md`
- Reference updates: `docs/reference/unified-meta-command.md`
- FAQ updates: `CLAUDE.md`
- Feature announcement: `README.md`

### 👥 Users (After Release)

**Start here**:
1. `docs/guides/prompt-learning-system.md` (to be created in Stage 3)
2. `CLAUDE.md` FAQ entries
3. [PROMPT-FILE-FORMAT.md](./PROMPT-FILE-FORMAT.md) - File format reference

**Quick Start**:
```bash
# 1. Optimize a prompt
/meta Refine prompt: 发布新版本

# 2. Save when prompted (optional)
# Answer "y" and provide category/keywords

# 3. Reuse next time
/meta Refine prompt: release new version
# Select from similar prompts or generate new
```

---

## Document Summaries

### README.md
**Purpose**: Quick reference and entry point

**Contents**:
- Phase overview and status
- Architecture summary
- Stage breakdown (3 stages)
- Quick links to detailed docs
- Success criteria
- Rollout strategy

**Read Time**: 5-10 minutes

---

### PHASE-28-IMPLEMENTATION-PLAN.md
**Purpose**: Complete implementation plan with TDD approach

**Contents**:
- Detailed stage-by-stage plan
- For each stage:
  - Objectives and deliverables
  - Files to create/modify
  - TDD iteration details
  - Acceptance criteria
  - Dependencies
- Documentation updates
- Testing strategy
- Risk assessment
- Success metrics
- Rollout plan
- Future enhancements

**Sections**:
1. Overview and architecture summary
2. Stage 1: Infrastructure and Save (5-6h, ~180 lines)
3. Stage 2: Search and Reuse (5-6h, ~180 lines)
4. Stage 3: Management and Listing (3-4h, ~90 lines)
5. Documentation updates
6. Testing strategy
7. Risk assessment and mitigation
8. Success metrics and rollout

**Read Time**: 30-45 minutes

**Use Cases**:
- Planning implementation
- Understanding requirements
- Estimating effort
- Defining acceptance criteria

---

### IMPLEMENTATION-GUIDE.md
**Purpose**: Practical, step-by-step implementation guide

**Contents**:
- Pre-implementation checklist
- Stage-by-stage implementation steps
- Code examples and templates
- Testing procedures
- Validation checklists
- Common issues and solutions
- Completion checklist
- Quick reference card

**Sections**:
1. Pre-implementation checklist
2. Stage 1: Step-by-step (2-3 hours per step)
   - Create directory structure
   - Implement meta-prompt-save.md
   - Extend meta-prompt.md
   - Update CLAUDE.md
   - Validation checklist
3. Stage 2: Step-by-step
   - Implement meta-prompt-utils.md
   - Implement meta-prompt-search.md
   - Integrate into meta-prompt.md
   - Validation checklist
4. Stage 3: Step-by-step
   - Implement meta-prompt-list.md
   - Update CLAUDE.md
   - Validation checklist
5. Final integration testing
6. Documentation finalization
7. Common issues and solutions
8. Completion checklist

**Read Time**: 60-90 minutes (reference during implementation)

**Use Cases**:
- Implementing each stage
- Following TDD workflow
- Debugging issues
- Validating completion

---

### PROMPT-FILE-FORMAT.md
**Purpose**: Technical specification for prompt file format

**Contents**:
- File location and naming convention
- YAML frontmatter schema
- Field specifications (required/optional)
- Content sections structure
- Validation rules
- Example files
- Tooling support (parsing, validation)
- Migration and versioning
- Best practices

**Sections**:
1. Overview
2. File naming convention
3. Format structure
4. YAML frontmatter schema (10 required, 6 optional fields)
5. Field specifications (detailed)
6. Content sections (Original, Optimized, Notes)
7. Validation rules (file, YAML, content, consistency)
8. Example files (3 complete examples)
9. Tooling support (yq, Python, validation script)
10. Migration and versioning
11. Best practices

**Read Time**: 20-30 minutes

**Use Cases**:
- Implementing save/load functionality
- Creating validation scripts
- Understanding file structure
- Writing documentation
- Debugging file issues

---

### ARCHITECTURE-DIAGRAM.md
**Purpose**: Visual architecture and workflow documentation

**Contents**:
- System overview diagram
- Workflow diagrams (save, search, list)
- Data flow diagrams
- File format structure
- Capability loading mechanism
- Similarity matching algorithm
- Directory structure evolution
- Future enhancements
- Integration points

**Sections**:
1. System overview (ASCII diagram)
2. Workflow diagrams (save, search, list)
3. Data flow (file format structure)
4. Capability loading mechanism
5. Similarity matching algorithm (detailed example)
6. Directory structure evolution (3 stages)
7. Future enhancements (Phase 28.4-28.7)
8. Integration points (meta commands, MCP tools)

**Read Time**: 15-25 minutes

**Use Cases**:
- Understanding system architecture
- Visualizing workflows
- Planning implementation
- Explaining to stakeholders
- Debugging data flow issues

---

## Implementation Roadmap

### Week 1: Stage 1 - Infrastructure and Save
**Goal**: Users can save optimized prompts

**Tasks**:
1. Create `capabilities/prompts/` directory
2. Implement `meta-prompt-save.md`
3. Extend `meta-prompt.md` with save workflow
4. Update `CLAUDE.md` FAQ
5. Validation testing

**Deliverables**:
- ✅ Auto-create `.meta-cc/prompts/library/` directory
- ✅ Users can save prompts with YAML frontmatter
- ✅ Files follow naming convention
- ✅ Documentation updated

**Validation**: See [IMPLEMENTATION-GUIDE.md Stage 1 Validation](./IMPLEMENTATION-GUIDE.md#step-5-stage-1-validation-30-minutes)

---

### Week 2: Stage 2 - Search and Reuse
**Goal**: System recommends similar prompts and tracks usage

**Tasks**:
1. Implement `meta-prompt-utils.md`
2. Implement `meta-prompt-search.md`
3. Integrate search into `meta-prompt.md`
4. Validation testing

**Deliverables**:
- ✅ Search finds similar prompts (Jaccard similarity)
- ✅ Top 5 matches ranked by relevance + usage
- ✅ Users can select or skip
- ✅ Usage count increments on reuse

**Validation**: See [IMPLEMENTATION-GUIDE.md Stage 2 Validation](./IMPLEMENTATION-GUIDE.md#step-4-stage-2-validation-1-hour)

---

### Week 3: Stage 3 - Management and Listing
**Goal**: Users can browse, filter, and manage saved prompts

**Tasks**:
1. Implement `meta-prompt-list.md`
2. Update `CLAUDE.md` with browsing FAQ
3. Create `docs/guides/prompt-learning-system.md`
4. Update reference documentation
5. Validation testing
6. Final integration testing

**Deliverables**:
- ✅ List all prompts in table format
- ✅ Filter by category
- ✅ Sort by usage/date/alpha
- ✅ Summary statistics
- ✅ Complete user guide

**Validation**: See [IMPLEMENTATION-GUIDE.md Stage 3 Validation](./IMPLEMENTATION-GUIDE.md#step-3-stage-3-validation-1-hour)

---

## Key Design Decisions

### 1. Zero Intrusion Philosophy
**Decision**: Pure capability implementation, no MCP tools, no Go code changes

**Rationale**:
- Minimizes risk (no core system changes)
- Faster implementation (Markdown only)
- Easy to maintain and extend
- Fully compatible with existing infrastructure

**Trade-offs**:
- Limited performance optimization (no native code)
- Relies on shell tools for file operations
- Cannot integrate deeply with MCP query system

---

### 2. Differentiated Capability Loading
**Decision**: Use `capabilities/prompts/` subdirectory for internal capabilities

**Rationale**:
- Leverages native MCP behavior (subdirectories not listed)
- Keeps `/meta` interface clean (only public capabilities visible)
- No configuration needed
- Clear separation of concerns

**Discovery**: Existing MCP capability loading already supports this pattern

---

### 3. Flat Storage Structure
**Decision**: Store all prompts in flat `.meta-cc/prompts/library/` directory

**Rationale**:
- Simple to implement and maintain
- CLI-friendly (easy to browse with ls, grep, etc.)
- No need for complex indexing (initially)
- Scales to 100+ prompts without issues

**Trade-offs**:
- May need indexing for 500+ prompts (Phase 28.4)
- No hierarchical organization (rely on categories)

---

### 4. Similarity Matching Algorithm
**Decision**: Jaccard similarity on keywords with usage weighting

**Formula**: `combined_score = 0.7 * jaccard_similarity + 0.3 * log(usage_count+1) / log(100)`

**Rationale**:
- Simple and interpretable
- Fast computation (no NLP required)
- Balances similarity and popularity
- Works well for short prompts

**Trade-offs**:
- Keyword-based (misses semantic similarity)
- Sensitive to keyword quality
- May improve with embedding-based search (Phase 28.6)

---

### 5. Project-Local Storage
**Decision**: Store prompts in `.meta-cc/` within each project

**Rationale**:
- Project-specific prompts stay local
- Easy to share via git (selective commit)
- No global configuration needed
- Matches existing `.claude/` pattern

**Future**: Phase 28.5 will add global library (`~/.meta-cc/`)

---

## Success Criteria

### MVP (After Stage 1)
- [ ] Users can save optimized prompts
- [ ] Storage directory auto-created
- [ ] Valid file format (YAML + Markdown)
- [ ] Optional save (non-intrusive)

### Complete (After Stage 3)
- [ ] Users can search and reuse prompts
- [ ] System recommends similar prompts
- [ ] Usage tracking works correctly
- [ ] Users can browse and filter prompts
- [ ] Complete documentation published

### Metrics
- **Adoption**: ≥3 prompts saved per active user
- **Reuse**: ≥50% reuse rate (select historical vs. generate new)
- **Relevance**: ≥70% user satisfaction with search results
- **Performance**: <3s search time for 50+ prompt library

---

## Common Questions

### Q: Why not use a database?
**A**: Flat files are simpler, CLI-friendly, git-compatible, and sufficient for 100+ prompts. Can add indexing later if needed (Phase 28.4).

### Q: Why YAML frontmatter instead of JSON?
**A**: YAML is more human-readable, widely used in markdown files (Jekyll, Hugo), and easier to edit manually.

### Q: Why not use embeddings for similarity?
**A**: Keyword-based matching is simpler, faster, and sufficient for MVP. Embeddings require external dependencies (models, APIs) and add complexity. Can be added in Phase 28.6 if needed.

### Q: Why project-local instead of global?
**A**: Project-specific prompts should stay with the project. Global library adds complexity and conflicts. Phase 28.5 will add optional global library for cross-project sharing.

### Q: How does this scale to 1000+ prompts?
**A**: Current design scales to ~100 prompts. For 1000+, Phase 28.4 will add indexing and caching. Indexing provides O(log n) search instead of O(n).

### Q: Can I use this without git?
**A**: Yes! The system works with or without git. Git is only needed for team sharing (optional).

### Q: Does this work with Claude Code Desktop and Web?
**A**: Yes, it's a plugin capability that works in both environments. The `.meta-cc/` directory is project-local, so it works wherever the project is accessed.

---

## Related Documentation

### Meta-CC Documentation
- [Core Principles](../../docs/core/principles.md) - Project constraints and methodology
- [Implementation Plan](../../docs/core/plan.md) - Overall project roadmap
- [Unified Meta Command](../../docs/reference/unified-meta-command.md) - /meta command reference
- [Plugin Development Guide](../../docs/guides/plugin-development.md) - Capability development

### Claude Code Documentation
- [Slash Commands](https://docs.claude.com/en/docs/claude-code/slash-commands)
- [Capabilities](https://docs.claude.com/en/docs/claude-code/capabilities)
- [MCP Integration](https://docs.claude.com/en/docs/claude-code/mcp)

---

## Feedback and Contributions

### Report Issues
- File issues in GitHub: [yaleh/meta-cc/issues](https://github.com/yaleh/meta-cc/issues)
- Tag with `phase-28` or `prompt-learning-system`

### Suggest Improvements
- Open discussions: [yaleh/meta-cc/discussions](https://github.com/yaleh/meta-cc/discussions)
- Share usage patterns and examples
- Propose new features for Phase 28.4+

### Contribute
- See [Plugin Development Guide](../../docs/guides/plugin-development.md)
- Follow [Core Principles](../../docs/core/principles.md)
- Submit PRs with tests and documentation

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-27 | Initial documentation package |

---

## Appendix: File Checklist

### Implementation Files (To Be Created)

**Capability Files** (~400 lines):
- [ ] `capabilities/prompts/meta-prompt-save.md` (~100 lines)
- [ ] `capabilities/prompts/meta-prompt-search.md` (~120 lines)
- [ ] `capabilities/prompts/meta-prompt-list.md` (~70 lines)
- [ ] `capabilities/prompts/meta-prompt-utils.md` (~20 lines)
- [ ] `capabilities/commands/meta-prompt.md` (extend ~90 lines)

**Documentation Files** (~230 lines):
- [ ] `docs/guides/prompt-learning-system.md` (~150 lines)
- [ ] `CLAUDE.md` (add ~50 lines to FAQ)
- [ ] `docs/reference/unified-meta-command.md` (add ~20 lines)
- [ ] `README.md` (add ~10 lines to features)

**Test/Validation Files**:
- [ ] `plans/28/validate-stage-1.sh` (optional)
- [ ] `plans/28/validate-stage-2.sh` (optional)
- [ ] `plans/28/validate-stage-3.sh` (optional)

### Documentation Files (Already Created)

- [x] `plans/28/README.md` (181 lines)
- [x] `plans/28/PHASE-28-IMPLEMENTATION-PLAN.md` (1,051 lines)
- [x] `plans/28/IMPLEMENTATION-GUIDE.md` (1,148 lines)
- [x] `plans/28/PROMPT-FILE-FORMAT.md` (640 lines)
- [x] `plans/28/ARCHITECTURE-DIAGRAM.md` (570 lines)
- [x] `plans/28/INDEX.md` (this file)

**Total Lines**: ~3,590 lines of planning documentation

---

**Ready to implement?** Start with [IMPLEMENTATION-GUIDE.md](./IMPLEMENTATION-GUIDE.md) and follow the step-by-step instructions.

**Questions?** Check the [Common Questions](#common-questions) section or refer to specific documents based on your role.

**Feedback?** Open an issue or discussion on GitHub.
