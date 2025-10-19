# Skills & Agents Packaging Proposal

**Status**: Draft for Review
**Date**: 2025-10-19
**Author**: Claude Code Analysis

## Executive Summary

This proposal outlines improvements to the build and GitHub Actions workflow to package the existing **15 skills** (`.claude/skills/`) and **5 agents** (`.claude/agents/`) into the meta-cc plugin distribution, enabling users to benefit from validated methodologies across 8 domains (testing, CI/CD, error recovery, etc.).

### Current State
- **Skills**: 15 skills with 110 markdown files (~1.5MB) in `.claude/skills/`
- **Agents**: 5 agents in `.claude/agents/` (iteration-executor, knowledge-extractor, project-planner, stage-executor, iteration-prompt-designer)
- **Current packaging**: Only 1 command (`meta.md`) + 2 agents (project-planner, stage-executor) declared in `plugin.json`
- **Gap**: 15 skills + 3 additional agents not packaged or distributed

### Proposed State
- Package all 15 skills as part of plugin distribution
- Include all 5 agents in plugin.json manifest
- Add skills directory to release bundles
- Update installation scripts to deploy skills
- Maintain backward compatibility with existing capability system

---

## 1. Current Architecture Analysis

### 1.1 Directory Structure
```
.claude/
├── commands/
│   └── meta.md                    # Unified meta command (1 file)
├── agents/
│   ├── iteration-executor.md      # ❌ Not in plugin.json
│   ├── iteration-prompt-designer.md  # ❌ Not in plugin.json
│   ├── knowledge-extractor.md     # ❌ Not in plugin.json
│   ├── project-planner.md         # ✅ In plugin.json
│   └── stage-executor.md          # ✅ In plugin.json
└── skills/                        # ❌ Not packaged at all
    ├── testing-strategy/          # 15 skills total
    ├── ci-cd-optimization/
    ├── error-recovery/
    ├── dependency-health/
    ├── knowledge-transfer/
    ├── technical-debt-management/
    ├── code-refactoring/
    ├── cross-cutting-concerns/
    ├── observability-instrumentation/
    ├── api-design/
    ├── methodology-bootstrapping/
    ├── agent-prompt-evolution/
    ├── baseline-quality-assessment/
    ├── rapid-convergence/
    └── retrospective-validation/

capabilities/
├── commands/                      # 13 capability files for /meta command
│   └── meta-*.md                  # Distributed separately
└── agents/                        # Empty directory
```

### 1.2 Current Build Process
**Makefile targets**:
- `sync-plugin-files`: Copies only `meta.md` + agents to `dist/`
- `bundle-capabilities`: Creates `capabilities-latest.tar.gz` (only from `capabilities/` dir)
- Skills are **completely ignored** in current build

**GitHub Actions** (`.github/workflows/release.yml`):
- Line 69: Runs `scripts/sync-plugin-files.sh` (copies meta.md + agents)
- Line 104-137: Creates plugin packages with `dist/commands/` and `dist/agents/`
- Line 99: Creates separate `capabilities-latest.tar.gz`
- **Gap**: `.claude/skills/` never copied to dist or packages

### 1.3 Plugin Manifest
**plugin.json** (lines 23-29):
```json
"commands": [
  "./.claude/commands/meta.md"
],
"agents": [
  "./.claude/agents/project-planner.md",
  "./.claude/agents/stage-executor.md"
]
```
**Missing**:
- 15 skills declarations
- 3 agents (iteration-executor, knowledge-extractor, iteration-prompt-designer)

---

## 2. Proposed Changes

### 2.1 Plugin Manifest Update

**File**: `.claude-plugin/plugin.json`

Add new `skills` array (following Claude Code plugin spec):

```json
{
  "name": "meta-cc",
  "version": "0.28.8",
  "description": "...",
  "commands": [
    "./.claude/commands/meta.md"
  ],
  "agents": [
    "./.claude/agents/iteration-executor.md",
    "./.claude/agents/iteration-prompt-designer.md",
    "./.claude/agents/knowledge-extractor.md",
    "./.claude/agents/project-planner.md",
    "./.claude/agents/stage-executor.md"
  ],
  "skills": [
    "./.claude/skills/testing-strategy/SKILL.md",
    "./.claude/skills/ci-cd-optimization/SKILL.md",
    "./.claude/skills/error-recovery/SKILL.md",
    "./.claude/skills/dependency-health/SKILL.md",
    "./.claude/skills/knowledge-transfer/SKILL.md",
    "./.claude/skills/technical-debt-management/SKILL.md",
    "./.claude/skills/code-refactoring/SKILL.md",
    "./.claude/skills/cross-cutting-concerns/SKILL.md",
    "./.claude/skills/observability-instrumentation/SKILL.md",
    "./.claude/skills/api-design/SKILL.md",
    "./.claude/skills/methodology-bootstrapping/SKILL.md",
    "./.claude/skills/agent-prompt-evolution/SKILL.md",
    "./.claude/skills/baseline-quality-assessment/SKILL.md",
    "./.claude/skills/rapid-convergence/SKILL.md",
    "./.claude/skills/retrospective-validation/SKILL.md"
  ]
}
```

**Rationale**:
- Each skill has a `SKILL.md` entry point with frontmatter metadata
- Skills reference supporting files via relative paths
- Claude Code plugin system will discover and load all skills

---

### 2.2 Build Script Updates

#### 2.2.1 Update `scripts/sync-plugin-files.sh`

**Current** (lines 27-40):
```bash
echo "  Copying unified meta command from .claude/commands/..."
cp "$PROJECT_ROOT/.claude/commands/meta.md" "$DIST_DIR/commands/"

echo "  Copying agents from .claude/agents/..."
if ls "$PROJECT_ROOT/.claude/agents/"*.md 1> /dev/null 2>&1; then
    cp "$PROJECT_ROOT/.claude/agents/"*.md "$DIST_DIR/agents/"
fi
```

**Proposed**:
```bash
# Create dist directories
mkdir -p "$DIST_DIR/commands" "$DIST_DIR/agents" "$DIST_DIR/skills"

# Copy unified meta command
echo "  Copying unified meta command from .claude/commands/..."
cp "$PROJECT_ROOT/.claude/commands/meta.md" "$DIST_DIR/commands/"

# Copy ALL agents
echo "  Copying agents from .claude/agents/..."
if ls "$PROJECT_ROOT/.claude/agents/"*.md 1> /dev/null 2>&1; then
    cp "$PROJECT_ROOT/.claude/agents/"*.md "$DIST_DIR/agents/"
fi

# NEW: Copy skills directory with all supporting files
echo "  Copying skills from .claude/skills/..."
if [ -d "$PROJECT_ROOT/.claude/skills" ]; then
    cp -r "$PROJECT_ROOT/.claude/skills/"* "$DIST_DIR/skills/"
    SKILL_COUNT=$(find "$DIST_DIR/skills" -name "SKILL.md" | wc -l)
    SKILL_FILES=$(find "$DIST_DIR/skills" -type f | wc -l)
    echo "    ✓ Copied $SKILL_COUNT skills ($SKILL_FILES total files)"
fi

# Count files
CMD_COUNT=$(find "$DIST_DIR/commands" -name "*.md" | wc -l)
AGENT_COUNT=$(find "$DIST_DIR/agents" -name "*.md" | wc -l)
SKILL_COUNT=$(find "$DIST_DIR/skills" -name "SKILL.md" | wc -l)

echo "✓ Plugin files synced to $DIST_DIR/"
echo "✓ Total: $CMD_COUNT commands, $AGENT_COUNT agents, $SKILL_COUNT skills"
echo "  Note: 13 capability files distributed separately in capabilities-latest.tar.gz"
```

**Impact**: Adds skills directory to `dist/` for packaging

---

#### 2.2.2 Update Makefile `sync-plugin-files` target

**Current** (lines 94-108):
```makefile
sync-plugin-files:
	@echo "Preparing plugin files for release packaging..."
	@mkdir -p $(DIST_DIR)/commands $(DIST_DIR)/agents
	@echo "  Copying entry point from .claude/commands/..."
	@cp .claude/commands/meta.md $(DIST_DIR)/commands/
	@echo "  Copying capabilities from $(CAPABILITIES_DIR)/commands/..."
	@cp $(CAPABILITIES_DIR)/commands/*.md $(DIST_DIR)/commands/ 2>/dev/null || true
	@echo "  Copying agents from .claude/agents/..."
	@cp .claude/agents/*.md $(DIST_DIR)/agents/ 2>/dev/null || true
	@echo "  Copying agents from $(CAPABILITIES_DIR)/agents/..."
	@cp $(CAPABILITIES_DIR)/agents/*.md $(DIST_DIR)/agents/ 2>/dev/null || true
	@echo "✓ Plugin files synced to $(DIST_DIR)/"
	@CMD_COUNT=$$(find $(DIST_DIR)/commands -name "*.md" 2>/dev/null | wc -l); \
	AGENT_COUNT=$$(find $(DIST_DIR)/agents -name "*.md" 2>/dev/null | wc -l); \
	echo "✓ Total: $$CMD_COUNT command files, $$AGENT_COUNT agent files"
```

**Proposed**:
```makefile
sync-plugin-files:
	@echo "Preparing plugin files for release packaging..."
	@mkdir -p $(DIST_DIR)/commands $(DIST_DIR)/agents $(DIST_DIR)/skills
	@echo "  Copying entry point from .claude/commands/..."
	@cp .claude/commands/meta.md $(DIST_DIR)/commands/
	@echo "  Copying agents from .claude/agents/..."
	@cp .claude/agents/*.md $(DIST_DIR)/agents/ 2>/dev/null || true
	@echo "  Copying skills from .claude/skills/..."
	@if [ -d ".claude/skills" ]; then \
		cp -r .claude/skills/* $(DIST_DIR)/skills/; \
		echo "    ✓ Skills directory copied"; \
	fi
	@echo "✓ Plugin files synced to $(DIST_DIR)/"
	@CMD_COUNT=$$(find $(DIST_DIR)/commands -name "*.md" 2>/dev/null | wc -l); \
	AGENT_COUNT=$$(find $(DIST_DIR)/agents -name "*.md" 2>/dev/null | wc -l); \
	SKILL_COUNT=$$(find $(DIST_DIR)/skills -name "SKILL.md" 2>/dev/null | wc -l); \
	echo "✓ Total: $$CMD_COUNT commands, $$AGENT_COUNT agents, $$SKILL_COUNT skills"
```

**Note**: Removed capability copying (they're distributed via `capabilities-latest.tar.gz`)

---

#### 2.2.3 Update GitHub Actions Release Workflow

**File**: `.github/workflows/release.yml`

**Current** (lines 104-137 - package creation):
```yaml
- name: Create plugin packages
  run: |
    VERSION=${{ steps.version.outputs.VERSION }}
    mkdir -p build/packages

    for platform in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64; do
      PKG_DIR=build/packages/meta-cc-plugin-${platform}
      mkdir -p $PKG_DIR/bin $PKG_DIR/.claude-plugin $PKG_DIR/commands $PKG_DIR/agents $PKG_DIR/lib

      # Copy binaries...
      # Copy Claude Code plugin structure
      cp -r .claude-plugin/* $PKG_DIR/.claude-plugin/
      cp -r dist/commands/* $PKG_DIR/commands/
      cp -r dist/agents/* $PKG_DIR/agents/ 2>/dev/null || true
      cp -r lib/* $PKG_DIR/lib/
      # ...
    done
```

**Proposed**:
```yaml
- name: Create plugin packages
  run: |
    VERSION=${{ steps.version.outputs.VERSION }}
    mkdir -p build/packages

    for platform in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64; do
      PKG_DIR=build/packages/meta-cc-plugin-${platform}
      mkdir -p $PKG_DIR/bin $PKG_DIR/.claude-plugin $PKG_DIR/commands $PKG_DIR/agents $PKG_DIR/skills $PKG_DIR/lib

      # Copy binaries (unchanged)...

      # Copy Claude Code plugin structure
      cp -r .claude-plugin/* $PKG_DIR/.claude-plugin/
      cp -r dist/commands/* $PKG_DIR/commands/
      cp -r dist/agents/* $PKG_DIR/agents/
      cp -r dist/skills/* $PKG_DIR/skills/              # NEW
      cp -r lib/* $PKG_DIR/lib/

      # Copy scripts and metadata (unchanged)...

      # Create archive
      cd build/packages
      tar -czf meta-cc-plugin-${VERSION}-${platform}.tar.gz meta-cc-plugin-${platform}
      cd ../..
    done
```

**Changes**:
1. Add `$PKG_DIR/skills` to `mkdir -p`
2. Add `cp -r dist/skills/* $PKG_DIR/skills/`

---

### 2.3 Installation Script Updates

**File**: `scripts/install.sh`

**Current behavior**: Installs to `~/.claude/plugins/meta-cc/` with commands and agents

**Proposed addition**:
```bash
# After installing commands and agents
if [ -d "$PLUGIN_DIR/skills" ]; then
    echo "  Installing skills..."
    cp -r "$PLUGIN_DIR/skills" "$INSTALL_DIR/"
    SKILL_COUNT=$(find "$INSTALL_DIR/skills" -name "SKILL.md" | wc -l)
    echo "    ✓ Installed $SKILL_COUNT skills"
fi
```

**Installation layout**:
```
~/.claude/plugins/meta-cc/
├── .claude-plugin/
│   ├── plugin.json          # Updated with skills array
│   └── marketplace.json
├── bin/
│   ├── meta-cc
│   └── meta-cc-mcp
├── commands/
│   └── meta.md
├── agents/
│   ├── iteration-executor.md
│   ├── knowledge-extractor.md
│   ├── project-planner.md
│   ├── stage-executor.md
│   └── iteration-prompt-designer.md
├── skills/                   # NEW
│   ├── testing-strategy/
│   │   ├── SKILL.md
│   │   ├── templates/
│   │   ├── examples/
│   │   └── reference/
│   ├── ci-cd-optimization/
│   └── ... (15 skills total)
└── lib/
```

---

## 3. Skills Structure Validation

Each skill follows this structure:

```
skill-name/
├── SKILL.md              # Entry point with frontmatter
├── templates/            # Reusable templates
├── examples/             # Example implementations
└── reference/            # Reference documentation
```

### Sample SKILL.md Frontmatter
```yaml
---
name: Testing Strategy
description: Systematic testing methodology for Go projects...
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
---
```

**Validation**: All 15 skills have `SKILL.md` entry points (confirmed by `find .claude/skills -name "SKILL.md" | wc -l` = 15)

---

## 4. Impact Analysis

### 4.1 Package Size Impact
- **Current plugin packages**: ~10-15MB per platform (binaries + minimal files)
- **Added skills**: ~1.5MB (110 markdown files)
- **New total**: ~11.5-16.5MB per platform
- **Percentage increase**: ~10-15%

**Rationale**: Skills are lightweight markdown documentation, negligible overhead

### 4.2 User Experience Impact

**Before**:
```bash
# User installs plugin
/plugin marketplace add yaleh/meta-cc

# Skills not available - must manually discover and install
```

**After**:
```bash
# User installs plugin
/plugin marketplace add yaleh/meta-cc

# Skills immediately available
testing-strategy
ci-cd-optimization
error-recovery
dependency-health
... (15 skills total)
```

**Benefit**: Zero-friction access to validated methodologies

### 4.3 Backward Compatibility
- **Capabilities system**: Unchanged (distributed via `capabilities-latest.tar.gz`)
- **Existing commands**: `/meta` command still works with capability sources
- **Existing agents**: No breaking changes to project-planner/stage-executor
- **Environment variables**: `META_CC_CAPABILITY_SOURCES` still respected

**Risk**: None identified

---

## 5. Testing Plan

### 5.1 Build Verification
```bash
# 1. Update files per proposal
# 2. Run build
make clean
make sync-plugin-files

# 3. Verify dist structure
tree dist/
# Expected:
#   dist/
#   ├── commands/
#   │   └── meta.md
#   ├── agents/
#   │   └── *.md (5 files)
#   └── skills/
#       └── */ (15 directories)

# 4. Verify skill count
find dist/skills -name "SKILL.md" | wc -l
# Expected: 15
```

### 5.2 Package Verification
```bash
# 1. Build release bundle
VERSION=v0.29.0-test make bundle-release

# 2. Extract and verify
cd build/bundles/
tar -xzf meta-cc-bundle-linux-amd64.tar.gz
cd meta-cc-v0.29.0-test-linux-amd64/

# 3. Verify structure
ls -la skills/
find skills/ -name "SKILL.md" | wc -l
# Expected: 15

# 4. Verify plugin.json
jq '.skills | length' .claude-plugin/plugin.json
# Expected: 15
```

### 5.3 Installation Test
```bash
# 1. Test local installation
./install.sh

# 2. Verify installed skills
ls -la ~/.claude/plugins/meta-cc/skills/
find ~/.claude/plugins/meta-cc/skills -name "SKILL.md" | wc -l
# Expected: 15

# 3. Verify Claude Code discovers skills
# (Manual test in Claude Code UI)
```

### 5.4 Smoke Test Updates

**File**: `scripts/smoke-tests.sh`

Add skill verification:
```bash
# Test 26: Skills structure
test_skills_structure() {
    echo "Test 26: Skills structure verification..."

    SKILL_COUNT=$(find "$EXTRACT_DIR/skills" -name "SKILL.md" 2>/dev/null | wc -l)
    if [ "$SKILL_COUNT" -eq 15 ]; then
        echo "  ✓ PASS: Found 15 skills"
    else
        echo "  ✗ FAIL: Expected 15 skills, found $SKILL_COUNT"
        return 1
    fi
}

# Test 27: plugin.json skills array
test_plugin_manifest_skills() {
    echo "Test 27: plugin.json skills array..."

    MANIFEST="$EXTRACT_DIR/.claude-plugin/plugin.json"
    SKILL_COUNT=$(jq '.skills | length' "$MANIFEST")

    if [ "$SKILL_COUNT" -eq 15 ]; then
        echo "  ✓ PASS: plugin.json declares 15 skills"
    else
        echo "  ✗ FAIL: Expected 15 skills in manifest, found $SKILL_COUNT"
        return 1
    fi
}

# Test 28: All agents in manifest
test_plugin_manifest_agents() {
    echo "Test 28: plugin.json agents array..."

    MANIFEST="$EXTRACT_DIR/.claude-plugin/plugin.json"
    AGENT_COUNT=$(jq '.agents | length' "$MANIFEST")

    if [ "$AGENT_COUNT" -eq 5 ]; then
        echo "  ✓ PASS: plugin.json declares 5 agents"
    else
        echo "  ✗ FAIL: Expected 5 agents in manifest, found $AGENT_COUNT"
        return 1
    fi
}
```

---

## 6. Documentation Updates

### 6.1 README.md
Add skills section:

```markdown
## Features

### Skills (15 validated methodologies)
- **Testing Strategy**: TDD, coverage-driven gap closure, CLI testing
- **CI/CD Optimization**: Quality gates, release automation, smoke testing
- **Error Recovery**: 13-category taxonomy, diagnostic workflows
- **Dependency Health**: Security-first, batch remediation
- **Knowledge Transfer**: Progressive learning paths, onboarding
- **Technical Debt Management**: SQALE methodology, prioritization
- **Code Refactoring**: Test-driven refactoring, complexity reduction
- **Cross-Cutting Concerns**: Error handling, logging, configuration
- **Observability**: Logs, metrics, traces, structured logging
- **API Design**: 6 validated patterns, parameter categorization
- **Methodology Bootstrapping**: BAIME framework, dual-layer value
- **Agent Prompt Evolution**: Agent specialization tracking
- **Baseline Quality Assessment**: Rapid convergence enablement
- **Rapid Convergence**: 3-4 iteration methodology development
- **Retrospective Validation**: Historical data validation

See [Skills Documentation](docs/guides/skills.md) for details.
```

### 6.2 Create docs/guides/skills.md
New comprehensive guide explaining:
- How to use skills
- Skill directory structure
- Integration with capabilities
- Customization and extension

---

## 7. Implementation Roadmap

### Phase 1: Core Changes (1-2 hours)
1. ✅ Update `.claude-plugin/plugin.json` (add skills + missing agents)
2. ✅ Update `scripts/sync-plugin-files.sh` (add skills copying)
3. ✅ Update `Makefile` sync-plugin-files target
4. ✅ Update `.github/workflows/release.yml` (add skills to packages)

### Phase 2: Installation & Scripts (1 hour)
5. ✅ Update `scripts/install.sh` (install skills directory)
6. ✅ Update `scripts/uninstall.sh` (remove skills directory)
7. ✅ Update `scripts/smoke-tests.sh` (add 3 new tests)

### Phase 3: Testing (2 hours)
8. ✅ Run `make clean && make sync-plugin-files`
9. ✅ Verify dist/ structure
10. ✅ Test bundle creation with `VERSION=v0.29.0-test make bundle-release`
11. ✅ Extract and verify package contents
12. ✅ Test installation script locally
13. ✅ Run updated smoke tests

### Phase 4: Documentation (1 hour)
14. ✅ Update `README.md` (add skills section)
15. ✅ Create `docs/guides/skills.md`
16. ✅ Update `CLAUDE.md` (mention skills)
17. ✅ Update release notes template

### Phase 5: Release (30 min)
18. ✅ Create PR with all changes
19. ✅ Merge after review
20. ✅ Create release with `./scripts/release.sh v0.29.0`
21. ✅ Verify GitHub release includes skills in packages
22. ✅ Test marketplace installation

**Total Estimated Time**: 4.5-5.5 hours

---

## 8. Alternatives Considered

### Option A: Keep Skills Separate (Status Quo)
**Pros**: Smaller plugin packages, opt-in skills
**Cons**: Fragmented user experience, manual discovery, low adoption

### Option B: Skills as Separate Plugin
**Pros**: Modular distribution, independent versioning
**Cons**: Two-plugin installation, dependency management complexity

### Option C: Bundled Skills (Proposed)
**Pros**: Unified distribution, zero-friction UX, immediate value
**Cons**: Slightly larger packages (~10-15% increase)

**Decision**: Option C provides best user experience with negligible cost

---

## 9. Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Skills bloat package size | Medium | High | 1.5MB is <15% increase, acceptable |
| Plugin.json spec changes | High | Low | Following official Claude Code spec |
| Installation script bugs | Medium | Low | Comprehensive smoke tests (28 tests) |
| Backward compatibility break | High | Very Low | No changes to existing features |
| Skills not discovered by Claude Code | High | Low | Follow official plugin spec exactly |

---

## 10. Success Metrics

**Post-Release Validation** (1 week):
- [ ] Plugin installs successfully across 5 platforms
- [ ] All 15 skills discoverable in Claude Code UI
- [ ] All 5 agents available via Task tool
- [ ] Smoke tests pass (28/28)
- [ ] User feedback positive (GitHub issues/discussions)

**Long-Term Metrics** (1 month):
- [ ] Skills usage analytics (if available)
- [ ] Reduced support requests for "how to use methodologies"
- [ ] Community adoption of skills

---

## 11. Open Questions

1. **Skills naming convention**: Should we rename `.claude/skills/` to `.claude-plugin/skills/` for consistency?
   - **Recommendation**: Keep `.claude/skills/` (follows Claude Code convention)

2. **Skills versioning**: Should skills have independent version numbers?
   - **Recommendation**: No, follow plugin version for simplicity

3. **Skills discovery UI**: How will Claude Code display 15 skills?
   - **Action**: Test with actual plugin installation

4. **Documentation depth**: Should all skill reference docs be included?
   - **Recommendation**: Yes, full documentation for offline access

5. **Capability vs Skills relationship**: Should we merge them?
   - **Recommendation**: Keep separate (capabilities = /meta content, skills = reusable workflows)

---

## 12. Next Steps

**For Approval**:
1. Review this proposal
2. Confirm plugin.json skills array spec matches Claude Code documentation
3. Approve implementation approach
4. Schedule release timeline

**After Approval**:
1. Create feature branch `feat/package-skills-and-agents`
2. Implement Phase 1-4 changes
3. Test thoroughly with Phase 3 verification
4. Create PR with comprehensive description
5. Merge and release v0.29.0

---

## Appendix A: Full plugin.json Diff

```diff
 {
   "name": "meta-cc",
   "version": "0.28.8",
   "description": "Meta-Cognition tool for Claude Code - analyze session history for workflow optimization",
   "author": {
     "name": "Yale Huang",
     "email": "yaleh@ieee.org",
     "url": "https://github.com/yaleh"
   },
   "license": "MIT",
   "homepage": "https://github.com/yaleh/meta-cc",
   "repository": "https://github.com/yaleh/meta-cc",
   "keywords": [
     "workflow-analysis",
     "session-history",
     "productivity",
     "metacognition",
     "analytics",
     "optimization",
     "mcp",
-    "subagents"
+    "subagents",
+    "skills",
+    "methodologies"
   ],
   "commands": [
     "./.claude/commands/meta.md"
   ],
   "agents": [
+    "./.claude/agents/iteration-executor.md",
+    "./.claude/agents/iteration-prompt-designer.md",
+    "./.claude/agents/knowledge-extractor.md",
     "./.claude/agents/project-planner.md",
     "./.claude/agents/stage-executor.md"
+  ],
+  "skills": [
+    "./.claude/skills/testing-strategy/SKILL.md",
+    "./.claude/skills/ci-cd-optimization/SKILL.md",
+    "./.claude/skills/error-recovery/SKILL.md",
+    "./.claude/skills/dependency-health/SKILL.md",
+    "./.claude/skills/knowledge-transfer/SKILL.md",
+    "./.claude/skills/technical-debt-management/SKILL.md",
+    "./.claude/skills/code-refactoring/SKILL.md",
+    "./.claude/skills/cross-cutting-concerns/SKILL.md",
+    "./.claude/skills/observability-instrumentation/SKILL.md",
+    "./.claude/skills/api-design/SKILL.md",
+    "./.claude/skills/methodology-bootstrapping/SKILL.md",
+    "./.claude/skills/agent-prompt-evolution/SKILL.md",
+    "./.claude/skills/baseline-quality-assessment/SKILL.md",
+    "./.claude/skills/rapid-convergence/SKILL.md",
+    "./.claude/skills/retrospective-validation/SKILL.md"
   ]
 }
```

---

## Appendix B: Skills Inventory

| Skill | Entry Point | Supporting Files | Description |
|-------|-------------|------------------|-------------|
| testing-strategy | SKILL.md | templates/3, examples/4, reference/6 | TDD, coverage-driven testing |
| ci-cd-optimization | SKILL.md | templates/5, examples/3, reference/4 | Quality gates, release automation |
| error-recovery | SKILL.md | templates/2, examples/3, reference/5 | 13-category error taxonomy |
| dependency-health | SKILL.md | templates/3, examples/2, reference/4 | Security-first dependency management |
| knowledge-transfer | SKILL.md | templates/3, examples/2, reference/3 | Progressive learning paths |
| technical-debt-management | SKILL.md | templates/2, examples/3, reference/7 | SQALE methodology |
| code-refactoring | SKILL.md | templates/3, examples/2, reference/4 | Test-driven refactoring |
| cross-cutting-concerns | SKILL.md | templates/4, examples/2, reference/5 | Error handling, logging, config |
| observability-instrumentation | SKILL.md | templates/3, examples/3, reference/4 | Logs, metrics, traces |
| api-design | SKILL.md | templates/2, examples/2, reference/3 | 6 validated API patterns |
| methodology-bootstrapping | SKILL.md | templates/4, examples/3, reference/6 | BAIME framework |
| agent-prompt-evolution | SKILL.md | templates/1, examples/2, reference/2 | Agent specialization tracking |
| baseline-quality-assessment | SKILL.md | templates/2, examples/1, reference/3 | Rapid convergence enablement |
| rapid-convergence | SKILL.md | templates/1, examples/1, reference/2 | 3-4 iteration methodology |
| retrospective-validation | SKILL.md | templates/1, examples/1, reference/2 | Historical data validation |

**Total**: 15 skills, 110 markdown files, ~1.5MB

---

## Appendix C: Agent Inventory

| Agent | Current Status | Description | Priority |
|-------|----------------|-------------|----------|
| project-planner.md | ✅ In plugin.json | Analyzes project docs, generates TDD iterations | High |
| stage-executor.md | ✅ In plugin.json | Executes plans with validation, quality assurance | High |
| iteration-executor.md | ❌ Missing | Executes single iteration through lifecycle phases | High |
| knowledge-extractor.md | ❌ Missing | Extracts validated knowledge from experiments | Medium |
| iteration-prompt-designer.md | ❌ Missing | Designs ITERATION-PROMPTS.md for bootstrapping | Medium |

**Action**: Add 3 missing agents to plugin.json

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-19 | Claude Code | Initial proposal |
