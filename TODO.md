# TODO

Project-wide task tracking and future improvements.

## Documentation

### High Priority

- [ ] **Create documentation link checker tool**
  - **Context**: After docs reorganization (Phase 1+2), 18 files have 70+ broken internal links
  - **Requirement**: Automated tool to detect and report broken markdown links
  - **Scope**:
    - Check all `*.md` files in `docs/` recursively
    - Validate relative links (`../`, `./`, and same-directory)
    - Validate anchor links (`#section`)
    - Report broken links with file location and line number
    - Support CI/CD integration (exit code on failure)
  - **Output format**:
    ```
    ✗ docs/core/principles.md:45 - Broken link: adr/README.md
    ✗ docs/guides/mcp.md:120 - Broken link: capabilities-guide.md
    ```
  - **Implementation options**:
    - Shell script with `grep` and `find`
    - Go tool (integrate with existing meta-cc codebase)
    - GitHub Action using existing tools (e.g., markdown-link-check)
  - **Related**: docs restructuring completed in branch `docs/restructure-directories`
  - **Files affected**: See [Link Test Report](#link-test-report) below

### Medium Priority

- [ ] **Fix internal documentation links**
  - Fix 70+ broken internal links across 18 files
  - Prioritize: core/ > guides/ > reference/ > tutorials/
  - Can be done incrementally or with batch script

### Low Priority

- [ ] **Plans directory restructuring**
  - Add descriptive names to plans directories: `N/` → `NN-descriptive-name/`
  - Example: `8/` → `08-mcp-integration/`
  - Improves readability and discoverability

## Features

### Planned

- [ ] TBD (track feature requests here)

## Infrastructure

### Planned

- [ ] Add markdown linting to CI/CD pipeline
- [ ] Automate documentation structure validation

---

## Link Test Report

**Date**: 2025-10-12
**Branch**: `docs/restructure-directories`

### Files with Broken Internal Links

```
docs/core/plan.md                       (8 broken links)
docs/core/principles.md                 (14 broken links)
docs/guides/capabilities.md             (3 broken links)
docs/guides/integration.md              (4 broken links)
docs/guides/mcp.md                      (4 broken links)
docs/guides/plugin-development.md       (5 broken links)
docs/guides/release-process.md          (1 broken link)
docs/guides/troubleshooting.md          (1 broken link)
docs/reference/cli.md                   (4 broken links)
docs/reference/features.md              (7 broken links)
docs/reference/jsonl.md                 (3 broken links)
docs/reference/repository-structure.md  (5 broken links)
docs/reference/unified-meta-command.md  (3 broken links)
docs/tutorials/cli-composability.md     (1 broken link)
docs/tutorials/cookbook.md              (2 broken links)
docs/tutorials/examples.md              (3 broken links)
docs/tutorials/github-setup.md          (1 broken link)
docs/tutorials/installation.md          (2 broken links)
```

**Total**: 18 files, 70+ broken links

**Note**: Main entry points (CLAUDE.md, README.md, DOCUMENTATION_MAP.md) have all links working correctly.

---

## Maintenance

### Regular Tasks

- [ ] Review and update TODO.md quarterly
- [ ] Archive completed tasks to preserve history
- [ ] Link TODO items to GitHub Issues when applicable

---

**Last Updated**: 2025-10-12
**Maintainers**: Project team
