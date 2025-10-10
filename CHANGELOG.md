# Changelog

All notable changes to the meta-cc project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.20.0] - 2025-10-10

### Added
- **Phase 21: Self-Hosted Marketplace**
  - Stage 21.1: Plugin marketplace configuration (.claude-plugin/marketplace.json)
  - Stage 21.2: Marketing documentation (docs/marketplace-listing.md)
  - Stage 21.3: Visual demonstration structure (docs/screenshots/)
  - Stage 21.4: Marketplace validation and testing
- Plugin marketplace configuration:
  - `.claude-plugin/marketplace.json` with rich metadata and component inventory
  - Comprehensive plugin description with feature highlights
  - GitHub Release asset references for all platforms
  - Installation command documentation (/plugin install yaleh/meta-cc)

## [Unreleased]

### Added
- **Phase 21: Self-Hosted Marketplace**
  - Stage 21.1: Plugin marketplace configuration (.claude-plugin/marketplace.json)
  - Stage 21.2: Marketing documentation (docs/marketplace-listing.md)
  - Stage 21.3: Visual demonstration structure (docs/screenshots/)
  - Stage 21.4: Marketplace validation and testing
- Plugin marketplace configuration:
  - `.claude-plugin/marketplace.json` with rich metadata and component inventory
  - Comprehensive plugin description with feature highlights
  - GitHub Release asset references for all platforms
  - Installation command documentation (/plugin install yaleh/meta-cc)
- Marketing documentation:
  - `docs/marketplace-listing.md` with compelling feature showcase
  - Visual asset structure in `docs/screenshots/` directory
  - Installation badges for marketplace and GitHub releases
- Marketplace validation:
  - `tests/marketplace_validation_test.sh` for format and consistency checks
  - Version synchronization verification across plugin.json and marketplace.json
  - Documentation cross-reference validation

### Changed
- README.md now prioritizes marketplace installation as recommended method
- Installation guide expanded with /plugin install workflow
- Documentation references updated to include marketplace installation
- Component inventory documented (10 slash commands, 3 subagents, 14 MCP tools)

### Improved
- Plugin discoverability via Claude Code marketplace
- One-command installation experience
- Professional visual documentation structure
- Installation workflow streamlined

- **Phase 20: Plugin Packaging & Release**
  - Stage 20.1: Plugin structure definition with plugin.json manifest
  - Stage 20.2: Automated installation script with platform detection and MCP configuration merging
  - Stage 20.3: Multi-platform GitHub Release workflow enhancements
  - Stage 20.4: Comprehensive installation documentation and testing checklists
- Plugin packaging structure:
  - `plugin.json` with metadata, dependencies, and platform definitions
  - `.claude/lib/mcp-config.json` template for MCP server configuration
  - Enhanced `install.sh` with platform detection (Linux, macOS, Windows)
  - New `uninstall.sh` for clean component removal
- Multi-platform plugin packages (5 platforms):
  - linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
  - Each package includes binaries, Claude Code integration files, and install/uninstall scripts
- Installation documentation:
  - `docs/installation.md` with platform-specific instructions and troubleshooting
  - `tests/PLUGIN_VERIFICATION.md` for comprehensive pre-release testing
- GitHub Release workflow enhancements:
  - Automated plugin package creation for all platforms
  - Checksum generation and verification
  - Auto-generated release notes

### Changed
- **Slash Command Enhancements**: Integrated new message query capabilities
  - `/meta-timeline` now includes assistant response analysis and conversation latency metrics
  - `/meta-coach` now includes interaction quality analysis with response time, tool efficiency, and satisfaction metrics
- **Installation Process**: Enhanced with platform detection and safe MCP configuration merging
  - `install.sh` now detects OS and architecture automatically
  - MCP configuration merged safely without overwriting existing servers
  - Post-install verification ensures correct setup
- **Release Workflow**: Now creates plugin packages instead of bare binaries
  - Each release includes 5 platform-specific plugin packages
  - One-command installation: `curl -L <url> | tar xz && ./install.sh`
  - Checksums provided for integrity verification

## [v0.14.0] - 2025-10-09

### Added
- **Phase 19: Assistant Message and Conversation Query**
  - Stage 19.1: AssistantMessage serialization support
  - Stage 19.2: `query-assistant-messages` CLI command with regex search
  - Stage 19.3: `query-conversation` CLI command with role filter support
  - Stage 19.4: MCP integration with `query_assistant_messages` and `query_conversation` tools
  - Stage 19.5: Comprehensive documentation updates
- New MCP tools (14 total, up from 12):
  - `mcp__meta_cc__query_assistant_messages` - Search assistant responses with regex
  - `mcp__meta_cc__query_conversation` - Search conversation messages with optional role filter
- CLI commands:
  - `meta-cc query assistant-messages --pattern <regex>` - Search assistant messages
  - `meta-cc query conversation --pattern <regex> [--role user|assistant]` - Search conversation

### Changed
- Documentation updated to reflect 14 MCP tools (previously 12)
- All message query tools now use hybrid output mode for large results
- Enhanced conversation analysis capabilities with role-based filtering

### Technical Details
- Supports regex pattern matching across assistant responses
- Conversation query can filter by role (user, assistant, or both)
- Full hybrid output mode support with configurable thresholds
- Consistent parameter naming with existing message query tools

## [v0.13.0] - 2025-10-09

### Added
- **Bundled Release Artifacts**: Platform-specific bundles (.tar.gz) containing binaries, slash commands, subagents, and installation script
  - 5 platform bundles: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
  - Each bundle includes: meta-cc CLI, meta-cc-mcp server, 8 slash commands, 3 subagents
  - One-command installation: `curl -L <bundle-url> | tar xz && ./install.sh`
  - Total release artifacts: 16 files (10 binaries + 5 bundles + checksums)
- **Installation Script**: `scripts/install.sh` for automated setup of binaries and Claude Code integration files
- **Makefile Target**: `bundle-release` for creating platform bundles (requires VERSION=vX.Y.Z)
- **Slash Command**: `/meta-next` for workflow continuation suggestions

### Changed
- **GitHub Actions**: Updated to build Linux ARM64 MCP server binary and generate platform bundles
- **Documentation**: Updated installation guides with bundle installation instructions
- **Phase 16 Migration Guide**: Clarified `inline_threshold_bytes` vs `max_output_bytes` differences and migration strategy

### Fixed
- Improved Phase 16 migration documentation with clear comparison table and usage examples

## [v0.12.1] - 2025-10-08

### Changed
- **Meta-insight to Meta-cc Renaming**: Comprehensive refactor to update all references from the old `mcp_meta_insight` / `meta‑insight` namespace to the new `mcp_meta_cc` / `meta‑cc` across documentation, agents, commands, and tests. This ensures consistency with the renamed meta-cc component and prevents confusion from stale meta-insight references.

## [v0.12.0] - 2025-10-08

### Added
- Configured remote repository and release process.

## Release Process

To create a new release:

1. Update CHANGELOG.md with version and release notes
2. Run `./scripts/release.sh v1.0.0`
3. Monitor GitHub Actions for build completion
4. Verify binaries on GitHub Releases page

## Versioning Strategy

- **v0.x.x**: Beta releases (pre-1.0)
- **v1.0.0**: First stable release
- **v1.x.0**: Minor version (new features, backward compatible)
- **v1.0.x**: Patch version (bug fixes only)
- **v1.0.0-beta.1**: Pre-release tags

---

## [v0.11.1-formalization] - 2025-10-03

### Changed
- **Agent Formalization**: Replaced verbose natural language content with compact lambda calculus formal specifications
  - 5 agents formalized: `doc-updater`, `prompt-suggester`, `pattern-analyzer`, `prompt-refiner`, `meta-coach`
  - **92% overall size reduction** (3074 → 244 lines) while preserving **100% behavioral semantics**
  - Individual reductions:
    - `meta-coach.md`: 1092 → 47 lines (96% reduction)
    - `prompt-refiner.md`: 604 → 51 lines (92% reduction)
    - `pattern-analyzer.md`: 543 → 50 lines (91% reduction)
    - `prompt-suggester.md`: 441 → 50 lines (89% reduction)
    - `doc-updater.md`: 394 → 46 lines (88% reduction)
  - All frontmatter YAML headers preserved exactly
  - Zero regressions (all tests pass)

### Added
- Formal specification documentation in `plans/11/`:
  - `agent-formalization-inventory.md`: Stage 1 inventory analysis
  - `agent-formalization-design.md`: Stage 2 design with semantic mapping
  - `agent-formalization-replacement-prompt.md`: Refined replacement strategy
- `.claude/agents/FORMALIZATION_SUMMARY.md`: Complete formalization results and methodology

## [v0.11.0] - 2025-10-03

### Added
- **Phase 11: Unix Composability** - Comprehensive MCP server piping and CLI integration
  - Stage 11.1: Tool output format standardization (JSON Lines, CSV, JSON)
  - Stage 11.2: Unix pipeline integration (`meta-cc ... | jq`, `| grep`)
  - Stage 11.3: stderr/stdout separation for clean piping
  - Stage 11.4: Cookbook documentation with 15+ practical recipes
- `docs/cookbook.md`: Real-world usage patterns and workflows
- `--format` flag support across all commands (json, jsonl, csv, markdown, table)

### Changed
- All CLI commands now support structured output for pipeline composition
- MCP server responses now use standardized schemas
- Error messages redirected to stderr (data to stdout)

## [v0.10.0] - 2025-10-02

### Added
- **Phase 10: Advanced Query & Aggregation** - SQL-like filtering and time-series analysis
  - Stage 10.1: `query-tools-advanced` with WHERE clause support
  - Stage 10.2: Aggregation engine (`aggregate-stats` by tool/status/uuid)
  - Stage 10.3: Time-series analysis (`query-time-series` with hourly/daily/weekly buckets)
  - Stage 10.4: File-level statistics (`query-files` with operation counts and error rates)
- Advanced MCP tools:
  - `mcp__meta_cc__query_tools_advanced`
  - `mcp__meta_cc__aggregate_stats`
  - `mcp__meta_cc__query_time_series`
  - `mcp__meta_cc__query_files`

### Changed
- Enhanced query engine with SQL-like expression parsing
- Aggregation pipeline with GROUP BY and metric calculation

## [v0.9.0] - 2025-10-01

### Added
- **Phase 9: Context Query & Workflow Patterns**
  - Stage 9.1: Error context extraction (`query-context` with temporal windows)
  - Stage 9.2: Tool sequence detection (`query-tool-sequences` for workflow patterns)
  - Stage 9.3: File access history (`query-file-access` for read/edit/write tracking)
- MCP server tools:
  - `mcp__meta_cc__query_context`
  - `mcp__meta_cc__query_tool_sequences`
  - `mcp__meta_cc__query_file_access`

### Changed
- Enhanced pattern detection with sequence analysis
- File operation tracking across session history

## [v0.8.0] - 2025-09-30

### Added
- **Phase 8: Enhanced Query & Pagination**
  - Stage 8.1-8.3: Tool extraction with pagination (`extract-tools --limit`)
  - Stage 8.4-8.6: Advanced tool querying (`query-tools --tool=X --status=Y`)
  - Stage 8.7-8.9: Message search with regex (`query-user-messages --pattern`)
  - Stage 8.10-8.12: Project state and successful prompt analysis
- MCP server tools:
  - `mcp__meta_cc__extract_tools`
  - `mcp__meta_cc__query_tools`
  - `mcp__meta_cc__query_user_messages`
  - `mcp__meta_cc__query_project_state`
  - `mcp__meta_cc__query_successful_prompts`

### Changed
- All query commands now support pagination and filtering
- Enhanced Markdown output with section anchors

## [v0.7.0] - 2025-09-28

### Added
- **Phase 7: MCP Server Implementation**
  - Stage 7.1-7.3: MCP server foundation with stdio transport
  - Stage 7.4-7.6: Tool exposure (`get_session_stats`, `analyze_errors`, `extract_tools`)
  - Stage 7.7-7.9: Integration testing and error handling
- MCP server binary: `meta-cc-mcp`
- MCP configuration documentation in `docs/mcp-setup.md`

### Changed
- Refactored internal packages for MCP tool integration
- Enhanced JSON schema output for MCP compatibility

## [v0.6.0] - 2025-09-25

### Added
- **Phase 6: Slash Commands & Subagent Integration**
  - Stage 6.1-6.3: Slash command implementation (`/meta-stats`, `/meta-errors`, `/meta-timeline`)
  - Stage 6.4-6.6: `@meta-coach` subagent with conversational analysis
  - Stage 6.7-6.9: Integration testing and documentation
- Slash commands: `/meta-stats`, `/meta-errors`, `/meta-timeline`, `/meta-query-tools`
- Subagent: `@meta-coach` for guided workflow optimization

### Changed
- Enhanced CLI output formatting for slash command integration
- Subagent prompts with context-aware recommendations

## [v0.5.0] - 2025-09-20

### Added
- **Phase 5: Error Pattern Analysis**
  - Stage 5.1-5.3: Error signature extraction and clustering
  - Stage 5.4-5.6: Frequency analysis and recurrence detection
  - Stage 5.7-5.9: Root cause identification and recommendations
- `meta-cc analyze errors` command with pattern detection
- Error clustering by signature and temporal analysis

### Changed
- Enhanced error analysis with contextual tool usage correlation

## [v0.4.0] - 2025-09-15

### Added
- **Phase 4: Timeline Visualization**
  - Stage 4.1-4.3: Chronological tool execution timeline
  - Stage 4.4-4.6: Error event correlation and visualization
  - Stage 4.7-4.9: Markdown timeline rendering
- `meta-cc analyze timeline` command with temporal analysis
- ASCII timeline visualization with error markers

### Changed
- Timeline output includes tool durations and error context

## [v0.3.0] - 2025-09-10

### Added
- **Phase 3: Semantic Integration** (Optional)
  - Enhanced query capabilities with semantic filters
  - Cross-session pattern detection

### Changed
- Improved query performance with indexed lookups

## [v0.2.0] - 2025-09-05

### Added
- **Phase 2: Index Optimization** (Optional)
  - SQLite indexing for cross-session queries
  - Advanced query commands: `meta-cc query sessions`, `meta-cc query tools`
  - Session database with efficient lookups

### Changed
- Optimized large session file parsing with streaming

## [v0.1.0] - 2025-09-01

### Added
- **Phase 1: Core Parser (MVP)**
  - JSONL session history parser
  - Basic commands:
    - `meta-cc parse extract --session <uuid>`: Extract structured data
    - `meta-cc parse stats --session <uuid>`: Tool usage statistics
    - `meta-cc analyze errors --session <uuid>`: Error pattern detection
  - JSON, Markdown, and table output formats
  - Session auto-detection from current directory
  - Initial documentation and README

### Infrastructure
- Go project setup with Cobra CLI framework
- Unit test foundation with 70%+ coverage
- GitHub repository initialization

---

**Legend**:
- `[vX.Y.Z]` - Release version
- `Added` - New features
- `Changed` - Changes to existing functionality
- `Deprecated` - Soon-to-be removed features
- `Removed` - Removed features
- `Fixed` - Bug fixes
- `Security` - Security improvements