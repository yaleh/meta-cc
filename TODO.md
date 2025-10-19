# TODO

Project-wide task tracking and future improvements.

## Documentation

### High Priority

- [x] **Create documentation link checker tool** ✅ **COMPLETED: 2025-10-12**
  - **Implementation**: Created `capabilities/commands/meta-doc-links.md` (208 lines)
  - **Access**: Via `/meta doc-links` or `/meta "check documentation links"`
  - **Features**:
    - Validates all internal markdown links in docs/, plans/, root .md files
    - Checks file existence and anchor validity
    - Severity-based classification (Critical/High/Medium/Low)
    - Pre-commit safety check (blocks on critical issues only)
    - Actionable output with file:line:target format
  - **Usage**: `/meta doc-links` before committing documentation changes
  - **Related**: docs restructuring completed in branch `docs/restructure-directories`
  - **Files affected**: See [Link Test Report](#link-test-report) below

### Medium Priority

- [x] **Create documentation synchronization checker (Capability)** ✅ **COMPLETED: 2025-10-12**
  - **Phase**: Documentation Management Phase 2
  - **Implementation**: Created `capabilities/commands/meta-doc-sync.md` (344 lines)
  - **Access**: Via `/meta doc-sync` or `/meta "check documentation sync"`
  - **Features**:
    - Cross-reference validation between CLAUDE.md, plan.md, principles.md, README.md
    - Phase status consistency check (plan.md vs plans/N/ directories)
    - DOCUMENTATION_MAP.md completeness validation
    - Code limits and constraints consistency verification
    - Severity-based classification (Critical/High/Medium/Low)
    - Pre-merge safety check with actionable fix recommendations
  - **Use cases**:
    - Pre-merge check before merging documentation changes
    - Pre-release validation before creating releases
    - Post-phase review after completing project phases
  - **Usage**: `/meta doc-sync` before merging or releasing

- [x] **Create project bootstrap capability based on documentation methodology** ✅ **COMPLETED: 2025-10-12**
  - **Phase**: Documentation Management Phase 0 Automation
  - **Implementation**: Created `capabilities/commands/meta-project-bootstrap.md` (1072 lines)
  - **Access**: Via `/meta project-bootstrap` or `/meta "bootstrap new project"`
  - **Features**:
    - Fully implements Documentation Management Methodology v5.0 requirements
    - Self-contained: Includes all necessary knowledge without external references
    - Context-aware: Adapts to detected project type and language stack
    - Creates all Phase 0 essential documents: plan.md, principles.md, CLAUDE.md, README.md
    - Generates complete directory structure (docs/, core/, guides/, reference/, tutorials/, architecture/, methodology/, archive/)
    - Provides language-specific guidance (Go, Rust, Python, JavaScript, C++, Ruby, C#)
    - Includes verification checklist and quality assurance measures
  - **Key Capabilities**:
    - Project type detection (web/frontend, cli, api, library, system, data, embedded)
    - Language stack analysis and appropriate template selection
    - Existing project context preservation (backup to docs/archive/)
    - Industry best practices integration
    - Claude Code optimization (all documents structured for AI assistance)
    - Interactive setup workflow with error handling
  - **Use Cases**:
    - New project initialization with proven documentation methodology
    - Existing project documentation audit and restructuring
    - Standardizing documentation across multiple projects
    - Educational tool for learning documentation best practices
  - **Methodology Compliance**:
    - Follows Documentation Management Methodology v5.0 completely
    - Implements all Phase 0 requirements from docs/methodology/documentation-management.md
    - Avoids all documented anti-patterns (mega-README, redundant docs, premature complexity)
    - Applies core principles: DRY, Progressive Disclosure, Task-Oriented Organization
  - **Content Highlights**:
    - Complete templates for all Phase 0 documents with language-specific customizations
    - Detailed exit criteria checklist (7 verification points)
    - Best practices and anti-patterns reference
    - Multi-language support (8 major languages with specific guidance)
    - Project-type specific examples (CLI, library, web app)
  - **Usage**: `/meta project-bootstrap` in new project directory
  - **Output**: Complete Phase 0 documentation structure ready for Claude Code assistance

- [ ] **Create documentation structure validator (Capability)**
  - **Phase**: Documentation Management Phase 3
  - **Purpose**: Long-term documentation health monitoring
  - **Implementation**: Capability `meta-doc-structure` (via `/meta`)
  - **Scope**:
    - DRY principle validation (detect duplicate content)
    - Progressive Disclosure compliance (README → guides → reference hierarchy)
    - Task-Oriented organization check (docs/guides/ structure)
    - Document size validation (per methodology guidelines)
    - Core document existence check (plan.md, principles.md)
  - **Use cases**:
    - Quarterly documentation health assessment
    - New project methodology compliance
    - Documentation refactoring guidance
  - **Requirements**: LLM-powered semantic analysis
  - **Priority**: Completes documentation management toolchain

- [x] **Fix internal documentation links** ✅ **COMPLETED: 2025-10-12**
  - Fixed 70+ broken internal links across 18 files
  - Updated cross-directory references and architecture paths
  - All documentation links now working correctly
  - Verified via documentation link checker capability

### Low Priority

- [x] **Plans directory restructuring** ✅ **COMPLETED: 2025-10-12**
  - Renamed all 21 phase directories with descriptive names
  - Format: NN-descriptive-name/ (e.g., 08-mcp-integration/)
  - Created README.md quick references for 12 phase directories
  - Improves readability and discoverability

## Features

### Planned

- [ ] TBD (track feature requests here)

## Infrastructure

### Planned

- [ ] Add markdown linting to CI/CD pipeline
- [ ] Automate documentation structure validation

---

## Documentation Management Tools

**Available Capabilities** (2025-10-12):

### Link Validation Tool
- **Capability**: `meta-doc-links` via `/meta doc-links`
- **Purpose**: Automated checking of internal markdown links
- **Features**: Severity classification, anchor validation, pre-commit safety
- **Status**: ✅ COMPLETED - Use for ongoing link validation

### Documentation Sync Checker
- **Capability**: `meta-doc-sync` via `/meta doc-sync`
- **Purpose**: Cross-reference validation between core documents
- **Features**: Phase status consistency, constraints verification, merge safety
- **Status**: ✅ COMPLETED - Use for pre-merge validation

### Project Bootstrap
- **Capability**: `meta-project-bootstrap` via `/meta project-bootstrap`
- **Purpose**: Automated Phase 0 documentation setup
- **Features**: Complete methodology implementation, language-specific templates
- **Status**: ✅ COMPLETED - Use for new project initialization

---

## Maintenance

### Regular Tasks

- [ ] Review and update TODO.md quarterly
- [ ] Archive completed tasks to preserve history
- [ ] Link TODO items to GitHub Issues when applicable

---

**Last Updated**: 2025-10-12
**Maintainers**: Project team
