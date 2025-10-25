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

### High Priority

- [ ] **Create documentation structure validator (Capability)**
  - **Phase**: Documentation Management Phase 3
  - **Purpose**: Final component of documentation management toolchain
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
    - CI/CD automated validation
  - **Requirements**: LLM-powered semantic analysis
  - **Priority**: Completes documentation management toolchain (last missing component)

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

- [x] **Create comprehensive documentation health and analysis capabilities** ✅ **COMPLETED: 2025-10-25**
  - **Phase**: Documentation Management Phase 3 Completion
  - **Implementation**: Created 5 additional documentation analysis capabilities:
    - `capabilities/commands/meta-doc-evolution.md` (3,426 bytes) - Documentation evolution tracking
    - `capabilities/commands/meta-doc-gaps.md` (3,954 bytes) - Content gap analysis
    - `capabilities/commands/meta-doc-health.md` (3,738 bytes) - Overall health assessment
    - `capabilities/commands/meta-doc-usage.md` (4,501 bytes) - Usage pattern analysis
  - **Access**: Via `/meta doc-evolution`, `/meta doc-gaps`, `/meta doc-health`, `/meta doc-usage`
  - **Features**:
    - Complete documentation management toolchain now available
    - Complements existing link validation, sync checking, and bootstrap capabilities
    - Provides comprehensive documentation lifecycle management
  - **Usage**: `/meta "analyze documentation health"` or specific capability names
  - **Impact**: Full documentation methodology implementation across all phases


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

### Completed

- [x] **Add comprehensive documentation management capabilities** ✅ **COMPLETED: 2025-10-25**
  - Complete documentation management toolchain now available
  - 9 documentation-related capabilities covering full lifecycle:
    - Link validation (`meta-doc-links`)
    - Sync checking (`meta-doc-sync`)
    - Project bootstrap (`meta-project-bootstrap`)
    - Health analysis (`meta-doc-health`, `meta-doc-evolution`, `meta-doc-gaps`, `meta-doc-usage`)
  - Only `meta-doc-structure` remains to be implemented

### Planned

- [ ] Add markdown linting to CI/CD pipeline
- [ ] Automate documentation structure validation (add meta-doc-structure to CI)
- [ ] Remove CLI-related code (MCP now operates independently)

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

- [x] Review and update TODO.md ✅ **COMPLETED: 2025-10-25**
- [ ] Archive completed tasks to preserve history
- [ ] Link TODO items to GitHub Issues when applicable

### Code Modernization

- [ ] **Remove CLI-related code (MCP Independence)**
  - **Rationale**: MCP server now operates independently without CLI dependency
  - **Scope**:
    - Remove obsolete CLI code from `cmd/` directory
    - Update build scripts to exclude CLI components
    - Update documentation to reflect MCP-only architecture
    - Preserve MCP server functionality (16 tools)
  - **Impact**:
    - Reduced codebase complexity
    - Cleaner architecture (MCP-only)
    - Simplified build and deployment
  - **Priority**: Medium (architectural cleanup)

---

**Last Updated**: 2025-10-25
**Maintainers**: Project team
