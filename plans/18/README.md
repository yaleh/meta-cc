# Phase 18: GitHub Release Preparation

**Status**: Planning Complete ✅ | Ready for Execution

## Quick Reference

### Objectives

Establish complete open-source release infrastructure for meta-cc v1.0:
- ✅ Legal compliance (MIT LICENSE, SECURITY.md)
- ✅ Community guidelines (CONTRIBUTING.md, CODE_OF_CONDUCT.md)
- ✅ Automated CI/CD (GitHub Actions)
- ✅ Cross-platform binary distribution (5 platforms)
- ✅ Professional documentation (badges, installation guides)

### Key Deliverables

| Category | Files | Lines |
|----------|-------|-------|
| Legal | LICENSE, SECURITY.md, NOTICE | ~60 |
| Community | CONTRIBUTING.md, CODE_OF_CONDUCT.md | ~300 |
| Templates | Issue/PR templates (.github/) | ~170 |
| CI/CD | ci.yml, release.yml, Makefile updates | ~300 |
| Release | scripts/release.sh, docs/release-process.md | ~200 |
| Docs | README.md updates, badges | ~90 |
| Config | docs/github-setup.md | ~50 |
| **Total** | **15 new files, 4 modified** | **~1,180** |

### Estimated Effort

| Stage | Description | Time | Priority |
|-------|-------------|------|----------|
| 18.1 | Legal & Licensing | 30 min | 🔴 Critical |
| 18.2 | Contributing Guidelines | 1 hour | 🟡 High |
| 18.3 | GitHub Templates | 45 min | 🟡 High |
| 18.4 | CI/CD Pipeline | 2 hours | 🔴 Critical |
| 18.5 | Release Automation | 1.5 hours | 🔴 Critical |
| 18.6 | Documentation | 45 min | 🟢 Medium |
| 18.7 | Repository Config | 30 min | 🟢 Medium |
| **Total** | | **~6.5 hours** | |

## Stage Checklist

### Stage 1: Legal & Licensing ⬜
- [ ] Create LICENSE (MIT)
- [ ] Create SECURITY.md
- [ ] Create NOTICE (if needed)
- [ ] Verify GitHub recognizes license
- [ ] Run: `ls -la LICENSE SECURITY.md`

### Stage 2: Contributing Guidelines ⬜
- [ ] Create CONTRIBUTING.md
- [ ] Create CODE_OF_CONDUCT.md
- [ ] Update README.md with contributing section
- [ ] Run: `grep "CONTRIBUTING.md" README.md`

### Stage 3: GitHub Templates ⬜
- [ ] Create bug_report.yml
- [ ] Create feature_request.yml
- [ ] Create PULL_REQUEST_TEMPLATE.md
- [ ] Create config.yml
- [ ] Run: `ls -la .github/ISSUE_TEMPLATE/`

### Stage 4: CI/CD Pipeline ⬜
- [ ] Create .github/workflows/ci.yml
- [ ] Create .github/workflows/release.yml
- [ ] Update Makefile with cross-compile target
- [ ] Test: `make cross-compile && ls build/`
- [ ] Verify: Push to GitHub, check Actions tab

### Stage 5: Release Automation ⬜
- [ ] Create scripts/release.sh
- [ ] Make executable: `chmod +x scripts/release.sh`
- [ ] Update CHANGELOG.md with release guidelines
- [ ] Create docs/release-process.md
- [ ] Test: `./scripts/release.sh` (expect error for validation)

### Stage 6: Documentation Enhancement ⬜
- [ ] Add badges to README.md
- [ ] Update installation section with binary downloads
- [ ] Add verification section
- [ ] Verify: Render README on GitHub

### Stage 7: Repository Configuration ⬜
- [ ] Set repository description and topics (manual)
- [ ] Configure branch protection on main (manual)
- [ ] Enable GitHub Actions (manual)
- [ ] Create docs/github-setup.md
- [ ] Update docs/plan.md (mark Phase 18 complete)
- [ ] Run: `make all` (final verification)

## Critical Files to Create

### Must Create (Legal)
```
LICENSE                           # MIT License (required for GitHub)
SECURITY.md                       # Vulnerability reporting
```

### Must Create (Community)
```
CONTRIBUTING.md                   # Development guidelines
CODE_OF_CONDUCT.md                # Community standards
```

### Must Create (CI/CD)
```
.github/workflows/ci.yml          # Automated testing
.github/workflows/release.yml     # Automated releases
scripts/release.sh                # Release helper script
```

### Must Create (Templates)
```
.github/ISSUE_TEMPLATE/bug_report.yml
.github/ISSUE_TEMPLATE/feature_request.yml
.github/PULL_REQUEST_TEMPLATE.md
```

## Key Files to Modify

```
README.md                         # Add badges + installation
CHANGELOG.md                      # Add release guidelines
Makefile                          # Add cross-compile target
docs/plan.md                      # Mark Phase 18 complete
```

## Binary Release Platforms

Phase 18 enables automated builds for:

1. **Linux amd64** (x86_64)
2. **Linux arm64** (ARM 64-bit)
3. **macOS amd64** (Intel)
4. **macOS arm64** (Apple Silicon)
5. **Windows amd64** (64-bit)

Plus MCP server binaries for each platform.

## Success Criteria

### Technical
- [ ] All 5 platform binaries build successfully
- [ ] CI workflow passes on Linux/macOS/Windows
- [ ] Release workflow uploads binaries to GitHub Release
- [ ] `make all` passes after all changes
- [ ] Test coverage maintained ≥80%

### Documentation
- [ ] LICENSE recognized by GitHub
- [ ] All badges functional and clickable
- [ ] Installation instructions clear for 5 platforms
- [ ] CONTRIBUTING.md complete with workflow
- [ ] CODE_OF_CONDUCT.md present

### Repository
- [ ] Repository description and topics set
- [ ] Branch protection active on main
- [ ] Required status checks configured
- [ ] Issue/PR templates functional

## Dependencies

- **None** - Phase 18 can start immediately
- All infrastructure is additive (no breaking changes)

## Testing Protocol

After each stage:
1. Run `make all` to verify no regressions
2. Verify stage deliverables exist
3. Check file syntax (YAML, Markdown)
4. Test functionality (e.g., template loading, workflow execution)

After Phase 18 complete:
1. Create test release: `./scripts/release.sh v0.99.0-test`
2. Verify GitHub Actions runs
3. Verify binaries uploaded
4. Download and test binary on local platform
5. Delete test release and tag

## References

- **Detailed Plan**: [plan.md](plan.md) - Complete stage-by-stage implementation
- **Project Roadmap**: [docs/plan.md](../../docs/plan.md) - Overall project status
- **Contributing**: Will be created in Stage 2
- **Release Process**: Will be created in Stage 5

## Next Step

**Execute Stage 18.1**: Create LICENSE, SECURITY.md, and NOTICE files

```bash
# Recommended execution approach:
# Option 1: Manual execution (stage by stage)
# Follow plan.md stages sequentially, run make all after each

# Option 2: Use stage-executor agent (automated)
# Delegate to @agent-stage-executor for systematic execution
```

---

**Phase Version**: 1.0
**Created**: 2025-10-07
**Dependencies**: None
**Estimated Duration**: 6.5 hours
**Status**: Ready for execution ✅
